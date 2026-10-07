package costing

import (
	"context"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WeightedAverageCalculator struct {
	repo ports.StockRepository
}

func NewWeightedAverageCalculator(repo ports.StockRepository) *WeightedAverageCalculator {
	return &WeightedAverageCalculator{repo: repo}
}

func (c *WeightedAverageCalculator) CalculateOutboundCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	issueQty decimal.Decimal,
) (domain.CostCalculationResult, error) {
	// Look up current balance
	bal, err := c.repo.GetBalance(ctx, domain.StockKey{
		LegalEntityID: entityID,
		ItemID:        itemID,
		WarehouseID:   warehouseID,
		BinID:         uuid.Nil,
	})
	if err != nil {
		return domain.CostCalculationResult{}, err
	}

	unitCost := decimal.Zero
	if bal != nil && bal.QuantityOnHand.IsPositive() {
		unitCost = bal.AverageCost
	} else {
		// Fallback to item cost price
		itemInfo, err := c.repo.GetItemCostingDetails(ctx, itemID)
		if err == nil && itemInfo != nil {
			unitCost = itemInfo.CostPrice
		}
	}

	totalCost := issueQty.Mul(unitCost).Round(6)
	return domain.CostCalculationResult{
		UnitCost:     unitCost,
		TotalCost:    totalCost,
		Consumptions: nil,
	}, nil
}

func (c *WeightedAverageCalculator) RecordInboundCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	inboundQty, unitCost decimal.Decimal,
	sourceMovementID uuid.UUID,
) error {
	// Weighted average cost is updated atomically in stock_balances via ApplyBalanceDelta.
	// For audit and hybrid reporting purposes, we also record a cost layer tracking the inbound lot.
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
