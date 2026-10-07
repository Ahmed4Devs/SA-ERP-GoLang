package costing

import (
	"context"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type FIFOCalculator struct {
	repo ports.StockRepository
}

func NewFIFOCalculator(repo ports.StockRepository) *FIFOCalculator {
	return &FIFOCalculator{repo: repo}
}

func (c *FIFOCalculator) CalculateOutboundCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	issueQty decimal.Decimal,
) (domain.CostCalculationResult, error) {
	if issueQty.LessThanOrEqual(decimal.Zero) {
		return domain.CostCalculationResult{}, nil
	}

	// Lock active layers in chronological order (FIFO)
	layers, err := c.repo.GetActiveCostLayersForUpdate(ctx, entityID, itemID, warehouseID)
	if err != nil {
		return domain.CostCalculationResult{}, err
	}

	remainingToIssue := issueQty
	totalCost := decimal.Zero
	consumptions := make([]domain.CostLayerConsumption, 0)

	for _, layer := range layers {
		if remainingToIssue.IsZero() {
			break
		}

		qtyFromLayer := decimal.Min(remainingToIssue, layer.RemainingQuantity)
		layerCost := qtyFromLayer.Mul(layer.UnitCost).Round(6)
		totalCost = totalCost.Add(layerCost)

		consumptions = append(consumptions, domain.CostLayerConsumption{
			LegalEntityID:    entityID,
			CostLayerID:      layer.ID,
			QuantityConsumed: qtyFromLayer,
			UnitCost:         layer.UnitCost,
			TotalCost:        layerCost,
		})

		newRemaining := layer.RemainingQuantity.Sub(qtyFromLayer)
		_, err = c.repo.UpdateCostLayerRemaining(ctx, layer.ID, newRemaining)
		if err != nil {
			return domain.CostCalculationResult{}, err
		}

		remainingToIssue = remainingToIssue.Sub(qtyFromLayer)
	}

	// If cost layers don't cover full quantity, fall back to item cost price for remainder
	if remainingToIssue.IsPositive() {
		itemInfo, err := c.repo.GetItemCostingDetails(ctx, itemID)
		fallbackUnitCost := decimal.Zero
		if err == nil && itemInfo != nil {
			fallbackUnitCost = itemInfo.CostPrice
		}
		fallbackCost := remainingToIssue.Mul(fallbackUnitCost).Round(6)
		totalCost = totalCost.Add(fallbackCost)
	}

	blendedUnitCost := decimal.Zero
	if issueQty.IsPositive() {
		blendedUnitCost = totalCost.DivRound(issueQty, 6)
	}

	return domain.CostCalculationResult{
		UnitCost:     blendedUnitCost,
		TotalCost:    totalCost,
		Consumptions: consumptions,
	}, nil
}

func (c *FIFOCalculator) RecordInboundCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	inboundQty, unitCost decimal.Decimal,
	sourceMovementID uuid.UUID,
) error {
	layer := &domain.CostLayer{
		LegalEntityID:     entityID,
		ItemID:            itemID,
		WarehouseID:       warehouseID,
		SourceMovementID:  sourceMovementID,
		OriginalQuantity:  inboundQty,
		RemainingQuantity: inboundQty,
		UnitCost:          unitCost,
		TotalCost:         inboundQty.Mul(unitCost).Round(6),
	}
	_, err := c.repo.InsertCostLayer(ctx, layer)
	return err
}
