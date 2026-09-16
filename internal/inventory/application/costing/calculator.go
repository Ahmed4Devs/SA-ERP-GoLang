package costing

import (
	"context"

	"erp/internal/inventory/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CostCalculator defines the contract for an inventory costing methodology.
type CostCalculator interface {
	// CalculateOutboundCost calculates the unit cost and total cost for an outbound stock movement.
	CalculateOutboundCost(
		ctx context.Context,
		entityID, itemID, warehouseID uuid.UUID,
		issueQty decimal.Decimal,
	) (domain.CostCalculationResult, error)

	// RecordInboundCost records the cost layer or updates the moving average upon receiving inventory.
	RecordInboundCost(
		ctx context.Context,
		entityID, itemID, warehouseID uuid.UUID,
		inboundQty, unitCost decimal.Decimal,
		sourceMovementID uuid.UUID,
	) error
}
