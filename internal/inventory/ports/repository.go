package ports

import (
	"context"

	"erp/internal/inventory/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StockRepository interface {
	// 1. Canonical StockRepository Contract (Section 25)
	LockBalance(
		ctx context.Context,
		legalEntityID uuid.UUID,
		itemID uuid.UUID,
		warehouseID uuid.UUID,
		binID uuid.UUID,
	) (*domain.StockBalance, error)

	InsertMovement(
		ctx context.Context,
		movement *domain.StockMovement,
	) (*domain.StockMovement, error)

	GetItemBaseUOM(
		ctx context.Context,
		itemID uuid.UUID,
	) (uuid.UUID, error)

	ApplyBalanceDelta(
		ctx context.Context,
		key domain.StockKey,
		quantityDelta decimal.Decimal,
		costDelta decimal.Decimal,
	) (*domain.StockBalance, error)

	Reserve(
		ctx context.Context,
		reservation *domain.StockReservation,
	) (*domain.StockReservation, error)

	ReleaseReservation(
		ctx context.Context,
		reservationID uuid.UUID,
	) error

	// 2. Balances & Movements Queries
	GetBalance(ctx context.Context, key domain.StockKey) (*domain.StockBalance, error)
	ListBalances(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, limit, offset int32) ([]domain.StockBalance, error)
	ListMovements(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, movType *string, limit, offset int32) ([]domain.StockMovement, error)
	GetMovementByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error)

	// 3. Warehouses & Bins
	ListWarehouses(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Warehouse, error)
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	GetWarehouseByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Warehouse, error)
	CreateWarehouse(ctx context.Context, wh *domain.Warehouse) (*domain.Warehouse, error)
	UpdateWarehouse(ctx context.Context, wh *domain.Warehouse) (*domain.Warehouse, error)
	ListBins(ctx context.Context, warehouseID uuid.UUID) ([]domain.BinLocation, error)
	GetBinByID(ctx context.Context, id uuid.UUID) (*domain.BinLocation, error)
	GetDefaultBin(ctx context.Context, warehouseID uuid.UUID) (*domain.BinLocation, error)
	CreateBin(ctx context.Context, bin *domain.BinLocation) (*domain.BinLocation, error)

	// 4. Transfers
	CreateTransferOrder(ctx context.Context, order *domain.TransferOrder, lines []domain.TransferOrderLine) (*domain.TransferOrder, error)
	GetTransferOrder(ctx context.Context, id uuid.UUID) (*domain.TransferOrder, error)
	ListTransferOrders(ctx context.Context, legalEntityID uuid.UUID, status *string, limit, offset int32) ([]domain.TransferOrder, error)
	UpdateTransferOrderStatus(ctx context.Context, id uuid.UUID, status domain.TransferStatus) (*domain.TransferOrder, error)

	// 5. Adjustments
	CreateStockAdjustment(ctx context.Context, adj *domain.StockAdjustment, lines []domain.StockAdjustmentLine) (*domain.StockAdjustment, error)
	GetStockAdjustment(ctx context.Context, id uuid.UUID) (*domain.StockAdjustment, error)
	ListStockAdjustments(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockAdjustment, error)
	UpdateStockAdjustmentStatus(ctx context.Context, id uuid.UUID, status domain.AdjustmentStatus, postedBy *uuid.UUID) (*domain.StockAdjustment, error)

	// 6. Counts
	CreateStockCount(ctx context.Context, count *domain.StockCount, lines []domain.StockCountLine) (*domain.StockCount, error)
	GetStockCount(ctx context.Context, id uuid.UUID) (*domain.StockCount, error)
	ListStockCounts(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockCount, error)
	UpdateStockCountStatus(ctx context.Context, id uuid.UUID, status domain.CountStatus, reconciledBy *uuid.UUID) (*domain.StockCount, error)

	// 7. Costing & Valuation
	InsertCostLayer(ctx context.Context, layer *domain.CostLayer) (*domain.CostLayer, error)
	GetActiveCostLayersForUpdate(ctx context.Context, entityID, itemID, warehouseID uuid.UUID) ([]domain.CostLayer, error)
	UpdateCostLayerRemaining(ctx context.Context, layerID uuid.UUID, remainingQty decimal.Decimal) (*domain.CostLayer, error)
	InsertCostLayerConsumption(ctx context.Context, consumption *domain.CostLayerConsumption) (*domain.CostLayerConsumption, error)
	ListCostLayers(ctx context.Context, entityID uuid.UUID, itemID, warehouseID *uuid.UUID, onlyActive *bool, limit, offset int32) ([]domain.CostLayer, error)
	CreateInventoryRevaluation(ctx context.Context, rev *domain.InventoryRevaluation) (*domain.InventoryRevaluation, error)
	ListInventoryRevaluations(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID, limit, offset int32) ([]domain.InventoryRevaluation, error)
	GetValuationSummary(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID) ([]domain.ValuationSummaryItem, error)
	UpdateStockBalanceCostAndValue(ctx context.Context, entityID, itemID, warehouseID uuid.UUID, avgCost, totalValue decimal.Decimal) (*domain.StockBalance, error)
	GetItemCostingDetails(ctx context.Context, itemID uuid.UUID) (*domain.ItemCostingInfo, error)
}
