package ports

import (
	"context"

	"erp/internal/inventory/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PostMovementRequest struct {
	LegalEntityID   uuid.UUID           `json:"legal_entity_id"`
	MovementType    domain.MovementType `json:"movement_type"`
	ItemID          uuid.UUID           `json:"item_id"`
	SKUID           *uuid.UUID          `json:"sku_id,omitempty"`
	WarehouseID     uuid.UUID           `json:"warehouse_id"`
	BinID           *uuid.UUID          `json:"bin_id,omitempty"` // If nil, defaults to system default bin
	LotNumber       *string             `json:"lot_number,omitempty"`
	SerialNumber    *string             `json:"serial_number,omitempty"`
	Quantity        decimal.Decimal     `json:"quantity"` // Positive for IN, Negative for OUT
	UOMID           uuid.UUID           `json:"uom_id"`
	UnitCost        decimal.Decimal     `json:"unit_cost"`
	SourceDocType   *string             `json:"source_doc_type,omitempty"`
	SourceDocID     *uuid.UUID          `json:"source_doc_id,omitempty"`
	SourceDocLineID *uuid.UUID          `json:"source_doc_line_id,omitempty"`
	CreatedBy       uuid.UUID           `json:"created_by"`
}

type ReserveStockRequest struct {
	LegalEntityID    uuid.UUID       `json:"legal_entity_id"`
	ItemID           uuid.UUID       `json:"item_id"`
	WarehouseID      uuid.UUID       `json:"warehouse_id"`
	BinID            *uuid.UUID      `json:"bin_id,omitempty"`
	Quantity         decimal.Decimal `json:"quantity"`
	DemandSourceType string          `json:"demand_source_type"`
	DemandSourceID   uuid.UUID       `json:"demand_source_id"`
}

type CreateTransferLineRequest struct {
	ItemID            uuid.UUID       `json:"item_id"`
	SKUID             *uuid.UUID      `json:"sku_id,omitempty"`
	RequestedQuantity decimal.Decimal `json:"requested_quantity"`
	UOMID             uuid.UUID       `json:"uom_id"`
	Notes             string          `json:"notes,omitempty"`
}

type CreateTransferRequest struct {
	LegalEntityID   uuid.UUID                   `json:"legal_entity_id"`
	FromWarehouseID uuid.UUID                   `json:"from_warehouse_id"`
	ToWarehouseID   uuid.UUID                   `json:"to_warehouse_id"`
	Notes           string                      `json:"notes,omitempty"`
	CreatedBy       uuid.UUID                   `json:"created_by"`
	Lines           []CreateTransferLineRequest `json:"lines"`
}

type CreateAdjustmentLineRequest struct {
	ItemID       uuid.UUID       `json:"item_id"`
	BinID        *uuid.UUID      `json:"bin_id,omitempty"`
	LotNumber    *string         `json:"lot_number,omitempty"`
	SerialNumber *string         `json:"serial_number,omitempty"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitCost     decimal.Decimal `json:"unit_cost"`
	Reason       string          `json:"reason,omitempty"`
}

type CreateAdjustmentRequest struct {
	LegalEntityID  uuid.UUID                     `json:"legal_entity_id"`
	WarehouseID    uuid.UUID                     `json:"warehouse_id"`
	AdjustmentType domain.AdjustmentType         `json:"adjustment_type"`
	ReasonCode     string                        `json:"reason_code"`
	Notes          string                        `json:"notes,omitempty"`
	CreatedBy      uuid.UUID                     `json:"created_by"`
	Lines          []CreateAdjustmentLineRequest `json:"lines"`
}

type CreateCountLineRequest struct {
	ItemID          uuid.UUID       `json:"item_id"`
	BinID           *uuid.UUID      `json:"bin_id,omitempty"`
	CountedQuantity decimal.Decimal `json:"counted_quantity"`
	Notes           string          `json:"notes,omitempty"`
}

type CreateCountRequest struct {
	LegalEntityID uuid.UUID                `json:"legal_entity_id"`
	WarehouseID   uuid.UUID                `json:"warehouse_id"`
	CountType     domain.CountType         `json:"count_type"`
	Lines         []CreateCountLineRequest `json:"lines"`
}

type InventoryService interface {
	// Warehouses & Bins
	ListWarehouses(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Warehouse, error)
	GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error)
	CreateWarehouse(ctx context.Context, wh domain.Warehouse) (*domain.Warehouse, error)
	ListBins(ctx context.Context, warehouseID uuid.UUID) ([]domain.BinLocation, error)
	CreateBin(ctx context.Context, bin domain.BinLocation) (*domain.BinLocation, error)

	// Movements & Ledger
	PostStockMovement(ctx context.Context, req PostMovementRequest) (*domain.StockMovement, error)
	ListMovements(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, movType *string, limit, offset int32) ([]domain.StockMovement, error)
	GetMovement(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error)

	// Balances
	GetBalance(ctx context.Context, key domain.StockKey) (*domain.StockBalance, error)
	ListBalances(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, limit, offset int32) ([]domain.StockBalance, error)

	// Reservations
	ReserveStock(ctx context.Context, req ReserveStockRequest) (*domain.StockReservation, error)
	ReleaseStock(ctx context.Context, reservationID uuid.UUID) error

	// Transfers
	CreateTransferOrder(ctx context.Context, req CreateTransferRequest) (*domain.TransferOrder, error)
	DispatchTransfer(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*domain.TransferOrder, error)
	ReceiveTransfer(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*domain.TransferOrder, error)
	ListTransferOrders(ctx context.Context, legalEntityID uuid.UUID, status *string, limit, offset int32) ([]domain.TransferOrder, error)
	GetTransferOrder(ctx context.Context, id uuid.UUID) (*domain.TransferOrder, error)

	// Adjustments
	CreateAdjustment(ctx context.Context, req CreateAdjustmentRequest) (*domain.StockAdjustment, error)
	PostAdjustment(ctx context.Context, adjustmentID uuid.UUID, userID uuid.UUID) (*domain.StockAdjustment, error)
	ListAdjustments(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockAdjustment, error)

	// Counts
	CreateStockCount(ctx context.Context, req CreateCountRequest) (*domain.StockCount, error)
	ReconcileStockCount(ctx context.Context, countID uuid.UUID, userID uuid.UUID) (*domain.StockCount, error)
	ListStockCounts(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockCount, error)

	// Costing & Valuation
	GetValuationReport(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID) ([]domain.ValuationSummaryItem, error)
	ListCostLayers(ctx context.Context, entityID uuid.UUID, itemID, warehouseID *uuid.UUID, onlyActive *bool, limit, offset int32) ([]domain.CostLayer, error)
	ExecuteRevaluation(ctx context.Context, req ExecuteRevaluationRequest) (*domain.InventoryRevaluation, error)
	ListRevaluations(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID, limit, offset int32) ([]domain.InventoryRevaluation, error)
}

type ExecuteRevaluationRequest struct {
	LegalEntityID  uuid.UUID       `json:"legal_entity_id"`
	WarehouseID    uuid.UUID       `json:"warehouse_id"`
	ItemID         uuid.UUID       `json:"item_id"`
	NewUnitCost    decimal.Decimal `json:"new_unit_cost"`
	Reason         string          `json:"reason"`
	JournalEntryID *uuid.UUID      `json:"journal_entry_id,omitempty"`
	CreatedBy      uuid.UUID       `json:"created_by"`
}
