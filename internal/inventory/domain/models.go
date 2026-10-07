package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WarehouseType string

const (
	WarehouseTypePhysical    WarehouseType = "PHYSICAL"
	WarehouseTypeVirtual     WarehouseType = "VIRTUAL"
	WarehouseTypeTransit     WarehouseType = "TRANSIT"
	WarehouseTypeConsignment WarehouseType = "CONSIGNMENT"
)

type Warehouse struct {
	ID            uuid.UUID     `json:"id"`
	LegalEntityID uuid.UUID     `json:"legal_entity_id"`
	Code          string        `json:"code"`
	NameEn        string        `json:"name_en"`
	NameAr        string        `json:"name_ar"`
	WarehouseType WarehouseType `json:"warehouse_type"`
	IsActive      bool          `json:"is_active"`
	AddressLine1  string        `json:"address_line1,omitempty"`
	City          string        `json:"city,omitempty"`
	PostalCode    string        `json:"postal_code,omitempty"`
	CountryCode   string        `json:"country_code,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type BinLocation struct {
	ID          uuid.UUID `json:"id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Code        string    `json:"code"`
	Aisle       string    `json:"aisle,omitempty"`
	Rack        string    `json:"rack,omitempty"`
	Shelf       string    `json:"shelf,omitempty"`
	Bin         string    `json:"bin,omitempty"`
	IsDefault   bool      `json:"is_default"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MovementType string

const (
	MovementTypeOpening         MovementType = "OPENING"
	MovementTypePurchaseReceipt MovementType = "PURCHASE_RECEIPT"
	MovementTypeSalesIssue      MovementType = "SALES_ISSUE"
	MovementTypeCustomerReturn  MovementType = "CUSTOMER_RETURN"
	MovementTypeSupplierReturn  MovementType = "SUPPLIER_RETURN"
	MovementTypeTransferOut     MovementType = "TRANSFER_OUT"
	MovementTypeTransferIn      MovementType = "TRANSFER_IN"
	MovementTypeAdjustmentIn    MovementType = "ADJUSTMENT_IN"
	MovementTypeAdjustmentOut   MovementType = "ADJUSTMENT_OUT"
	MovementTypeProductionIn    MovementType = "PRODUCTION_IN"
	MovementTypeProductionOut   MovementType = "PRODUCTION_OUT"
	MovementTypeReversal        MovementType = "REVERSAL"
)

type StockMovement struct {
	ID              uuid.UUID       `json:"id"`
	LegalEntityID   uuid.UUID       `json:"legal_entity_id"`
	MovementNumber  string          `json:"movement_number"`
	MovementType    MovementType    `json:"movement_type"`
	ItemID          uuid.UUID       `json:"item_id"`
	SKUID           *uuid.UUID      `json:"sku_id,omitempty"`
	WarehouseID     uuid.UUID       `json:"warehouse_id"`
	BinID           uuid.UUID       `json:"bin_id"`
	LotNumber       *string         `json:"lot_number,omitempty"`
	SerialNumber    *string         `json:"serial_number,omitempty"`
	Quantity        decimal.Decimal `json:"quantity"` // Positive for IN, Negative for OUT
	UOMID           uuid.UUID       `json:"uom_id"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	SourceDocType   *string         `json:"source_doc_type,omitempty"`
	SourceDocID     *uuid.UUID      `json:"source_doc_id,omitempty"`
	SourceDocLineID *uuid.UUID      `json:"source_doc_line_id,omitempty"`
	OccurredAt      time.Time       `json:"occurred_at"`
	CreatedBy       uuid.UUID       `json:"created_by"`
	CreatedAt       time.Time       `json:"created_at"`
	ItemCode        string          `json:"item_code,omitempty"`
	ItemNameEn      string          `json:"item_name_en,omitempty"`
	ItemNameAr      string          `json:"item_name_ar,omitempty"`
	WarehouseCode   string          `json:"warehouse_code,omitempty"`
	WarehouseNameAr string          `json:"warehouse_name_ar,omitempty"`
	BinCode         string          `json:"bin_code,omitempty"`
	UOMCode         string          `json:"uom_code,omitempty"`
}

type StockKey struct {
	LegalEntityID uuid.UUID
	ItemID        uuid.UUID
	WarehouseID   uuid.UUID
	BinID         uuid.UUID
}

type StockBalance struct {
	ID                uuid.UUID       `json:"id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	ItemID            uuid.UUID       `json:"item_id"`
	WarehouseID       uuid.UUID       `json:"warehouse_id"`
	BinID             uuid.UUID       `json:"bin_id"`
	QuantityOnHand    decimal.Decimal `json:"quantity_on_hand"`
	QuantityReserved  decimal.Decimal `json:"quantity_reserved"`
	QuantityAvailable decimal.Decimal `json:"quantity_available"`
	AverageCost       decimal.Decimal `json:"average_cost"`
	TotalValue        decimal.Decimal `json:"total_value"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Version           int64           `json:"version"`
	ItemCode          string          `json:"item_code,omitempty"`
	ItemNameEn        string          `json:"item_name_en,omitempty"`
	ItemNameAr        string          `json:"item_name_ar,omitempty"`
	BaseUOMCode       string          `json:"base_uom_code,omitempty"`
	WarehouseCode     string          `json:"warehouse_code,omitempty"`
	WarehouseNameAr   string          `json:"warehouse_name_ar,omitempty"`
	BinCode           string          `json:"bin_code,omitempty"`
}

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "ACTIVE"
	ReservationStatusReleased  ReservationStatus = "RELEASED"
	ReservationStatusFulfilled ReservationStatus = "FULFILLED"
)

type StockReservation struct {
	ID                uuid.UUID         `json:"id"`
	LegalEntityID     uuid.UUID         `json:"legal_entity_id"`
	ItemID            uuid.UUID         `json:"item_id"`
	WarehouseID       uuid.UUID         `json:"warehouse_id"`
	BinID             *uuid.UUID        `json:"bin_id,omitempty"`
	Quantity          decimal.Decimal   `json:"quantity"`
	ReservationStatus ReservationStatus `json:"reservation_status"`
	DemandSourceType  string            `json:"demand_source_type"`
	DemandSourceID    uuid.UUID         `json:"demand_source_id"`
	ExpiresAt         *time.Time        `json:"expires_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type TransferStatus string

const (
	TransferStatusDraft      TransferStatus = "DRAFT"
	TransferStatusRequested  TransferStatus = "REQUESTED"
	TransferStatusApproved   TransferStatus = "APPROVED"
	TransferStatusDispatched TransferStatus = "DISPATCHED"
	TransferStatusReceived   TransferStatus = "RECEIVED"
	TransferStatusCancelled  TransferStatus = "CANCELLED"
)

type TransferOrder struct {
	ID                   uuid.UUID           `json:"id"`
	LegalEntityID        uuid.UUID           `json:"legal_entity_id"`
	OrderNumber          string              `json:"order_number"`
	FromWarehouseID      uuid.UUID           `json:"from_warehouse_id"`
	ToWarehouseID        uuid.UUID           `json:"to_warehouse_id"`
	Status               TransferStatus      `json:"status"`
	ShippingDate         *time.Time          `json:"shipping_date,omitempty"`
	ExpectedDeliveryDate *time.Time          `json:"expected_delivery_date,omitempty"`
	ReceivedDate         *time.Time          `json:"received_date,omitempty"`
	Notes                string              `json:"notes,omitempty"`
	CreatedBy            uuid.UUID           `json:"created_by"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	FromWarehouseCode    string              `json:"from_warehouse_code,omitempty"`
	FromWarehouseNameAr  string              `json:"from_warehouse_name_ar,omitempty"`
	ToWarehouseCode      string              `json:"to_warehouse_code,omitempty"`
	ToWarehouseNameAr    string              `json:"to_warehouse_name_ar,omitempty"`
	Lines                []TransferOrderLine `json:"lines,omitempty"`
}

type TransferOrderLine struct {
	ID                 uuid.UUID       `json:"id"`
	TransferOrderID    uuid.UUID       `json:"transfer_order_id"`
	ItemID             uuid.UUID       `json:"item_id"`
	SKUID              *uuid.UUID      `json:"sku_id,omitempty"`
	RequestedQuantity  decimal.Decimal `json:"requested_quantity"`
	DispatchedQuantity decimal.Decimal `json:"dispatched_quantity"`
	ReceivedQuantity   decimal.Decimal `json:"received_quantity"`
	UOMID              uuid.UUID       `json:"uom_id"`
	Notes              string          `json:"notes,omitempty"`
	ItemCode           string          `json:"item_code,omitempty"`
	ItemNameEn         string          `json:"item_name_en,omitempty"`
	ItemNameAr         string          `json:"item_name_ar,omitempty"`
	UOMCode            string          `json:"uom_code,omitempty"`
}

type AdjustmentType string

const (
	AdjustmentTypeIncrease   AdjustmentType = "INCREASE"
	AdjustmentTypeDecrease   AdjustmentType = "DECREASE"
	AdjustmentTypeScrap      AdjustmentType = "SCRAP"
	AdjustmentTypeFound      AdjustmentType = "FOUND"
	AdjustmentTypeDamaged    AdjustmentType = "DAMAGED"
	AdjustmentTypeCycleCount AdjustmentType = "CYCLE_COUNT"
)

type AdjustmentStatus string

const (
	AdjustmentStatusDraft     AdjustmentStatus = "DRAFT"
	AdjustmentStatusApproved  AdjustmentStatus = "APPROVED"
	AdjustmentStatusPosted    AdjustmentStatus = "POSTED"
	AdjustmentStatusCancelled AdjustmentStatus = "CANCELLED"
)

type StockAdjustment struct {
	InventoryAccountID *uuid.UUID            `json:"inventory_account_id"`
	OffsetAccountID    *uuid.UUID            `json:"offset_account_id"`
	JournalEntryID     *uuid.UUID            `json:"journal_entry_id"`
	ID                 uuid.UUID             `json:"id"`
	LegalEntityID      uuid.UUID             `json:"legal_entity_id"`
	AdjustmentNumber   string                `json:"adjustment_number"`
	WarehouseID        uuid.UUID             `json:"warehouse_id"`
	AdjustmentType     AdjustmentType        `json:"adjustment_type"`
	ReasonCode         string                `json:"reason_code"`
	Status             AdjustmentStatus      `json:"status"`
	AdjustmentDate     time.Time             `json:"adjustment_date"`
	PostedAt           *time.Time            `json:"posted_at,omitempty"`
	PostedBy           *uuid.UUID            `json:"posted_by,omitempty"`
	Notes              string                `json:"notes,omitempty"`
	CreatedBy          uuid.UUID             `json:"created_by"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
	WarehouseCode      string                `json:"warehouse_code,omitempty"`
	WarehouseNameAr    string                `json:"warehouse_name_ar,omitempty"`
	Lines              []StockAdjustmentLine `json:"lines,omitempty"`
}

type StockAdjustmentLine struct {
	ID           uuid.UUID       `json:"id"`
	AdjustmentID uuid.UUID       `json:"adjustment_id"`
	ItemID       uuid.UUID       `json:"item_id"`
	UOMID        uuid.UUID       `json:"uom_id"`
	BinID        *uuid.UUID      `json:"bin_id,omitempty"`
	LotNumber    *string         `json:"lot_number,omitempty"`
	SerialNumber *string         `json:"serial_number,omitempty"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitCost     decimal.Decimal `json:"unit_cost"`
	TotalCost    decimal.Decimal `json:"total_cost"`
	Reason       string          `json:"reason,omitempty"`
	ItemCode     string          `json:"item_code,omitempty"`
	ItemNameEn   string          `json:"item_name_en,omitempty"`
	ItemNameAr   string          `json:"item_name_ar,omitempty"`
	BinCode      string          `json:"bin_code,omitempty"`
}

type CountType string

const (
	CountTypeFull  CountType = "FULL"
	CountTypeCycle CountType = "CYCLE"
	CountTypeSpot  CountType = "SPOT"
)

type CountStatus string

const (
	CountStatusPlanned    CountStatus = "PLANNED"
	CountStatusInProgress CountStatus = "IN_PROGRESS"
	CountStatusCompleted  CountStatus = "COMPLETED"
	CountStatusPosted     CountStatus = "POSTED"
	CountStatusCancelled  CountStatus = "CANCELLED"
)

type StockCount struct {
	CreatedBy          *uuid.UUID       `json:"created_by"`
	InventoryAccountID *uuid.UUID       `json:"inventory_account_id"`
	OffsetAccountID    *uuid.UUID       `json:"offset_account_id"`
	JournalEntryID     *uuid.UUID       `json:"journal_entry_id"`
	ID                 uuid.UUID        `json:"id"`
	LegalEntityID      uuid.UUID        `json:"legal_entity_id"`
	CountNumber        string           `json:"count_number"`
	WarehouseID        uuid.UUID        `json:"warehouse_id"`
	CountType          CountType        `json:"count_type"`
	Status             CountStatus      `json:"status"`
	CountDate          time.Time        `json:"count_date"`
	ReconciledAt       *time.Time       `json:"reconciled_at,omitempty"`
	ReconciledBy       *uuid.UUID       `json:"reconciled_by,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	WarehouseCode      string           `json:"warehouse_code,omitempty"`
	WarehouseNameAr    string           `json:"warehouse_name_ar,omitempty"`
	Lines              []StockCountLine `json:"lines,omitempty"`
}

type StockCountLine struct {
	UnitCost        decimal.Decimal `json:"unit_cost"`
	ID              uuid.UUID       `json:"id"`
	CountID         uuid.UUID       `json:"count_id"`
	ItemID          uuid.UUID       `json:"item_id"`
	UOMID           uuid.UUID       `json:"uom_id"`
	BinID           *uuid.UUID      `json:"bin_id,omitempty"`
	SystemQuantity  decimal.Decimal `json:"system_quantity"`
	CountedQuantity decimal.Decimal `json:"counted_quantity"`
	Variance        decimal.Decimal `json:"variance"`
	Notes           string          `json:"notes,omitempty"`
	ItemCode        string          `json:"item_code,omitempty"`
	ItemNameEn      string          `json:"item_name_en,omitempty"`
	ItemNameAr      string          `json:"item_name_ar,omitempty"`
	BinCode         string          `json:"bin_code,omitempty"`
}

type Lot struct {
	ID              uuid.UUID  `json:"id"`
	LegalEntityID   uuid.UUID  `json:"legal_entity_id"`
	ItemID          uuid.UUID  `json:"item_id"`
	LotNumber       string     `json:"lot_number"`
	ManufactureDate *time.Time `json:"manufacture_date,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	IsBlocked       bool       `json:"is_blocked"`
	CreatedAt       time.Time  `json:"created_at"`
}

type SerialStatus string

const (
	SerialStatusInStock  SerialStatus = "IN_STOCK"
	SerialStatusReserved SerialStatus = "RESERVED"
	SerialStatusIssued   SerialStatus = "ISSUED"
	SerialStatusReturned SerialStatus = "RETURNED"
	SerialStatusScrapped SerialStatus = "SCRAPPED"
)

type SerialNumber struct {
	ID            uuid.UUID    `json:"id"`
	LegalEntityID uuid.UUID    `json:"legal_entity_id"`
	ItemID        uuid.UUID    `json:"item_id"`
	SerialNumber  string       `json:"serial_number"`
	WarehouseID   *uuid.UUID   `json:"warehouse_id,omitempty"`
	BinID         *uuid.UUID   `json:"bin_id,omitempty"`
	Status        SerialStatus `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
