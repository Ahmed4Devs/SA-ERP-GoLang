package costing

import (
	"context"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StandardCostCalculator struct {
	repo ports.StockRepository
}

func NewStandardCostCalculator(repo ports.StockRepository) *StandardCostCalculator {
	return &StandardCostCalculator{repo: repo}
}

func (c *StandardCostCalculator) CalculateOutboundCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	issueQty decimal.Decimal,
) (domain.CostCalculationResult, error) {
	itemInfo, err := c.repo.GetItemCostingDetails(ctx, itemID)
	if err != nil {
		return domain.CostCalculationResult{}, err
	}

	standardCost := itemInfo.StandardCost
	if standardCost.IsZero() {
		standardCost = itemInfo.CostPrice
	}

	totalCost := issueQty.Mul(standardCost).Round(4)
	return domain.CostCalculationResult{
		UnitCost:     standardCost,
		TotalCost:    totalCost,
		Consumptions: nil,
	}, nil
}

func (c *StandardCostCalculator) RecordInboundCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	inboundQty, unitCost decimal.Decimal,
	sourceMovementID uuid.UUID,
) error {
	// Under Standard Costing, inventory is booked at standard_cost.
	// The variance between unitCost and standardCost is Purchase Price Variance (PPV).
	layer := &domain.CostLayer{
		LegalEntityID:     entityID,
		ItemID:            itemID,
		WarehouseID:       warehouseID,
		SourceMovementID:  sourceMovementID,
		OriginalQuantity:  inboundQty,
		RemainingQuantity: inboundQty,
		UnitCost:          unitCost,
		TotalCost:         inboundQty.Mul(unitCost).Round(4),
	}
	_, err := c.repo.InsertCostLayer(ctx, layer)
	return err
}
