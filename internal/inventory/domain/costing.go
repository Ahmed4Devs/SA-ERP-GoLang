package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CostingMethod string

const (
	CostingMethodWeightedAverage CostingMethod = "WEIGHTED_AVERAGE"
	CostingMethodFIFO            CostingMethod = "FIFO"
	CostingMethodStandard        CostingMethod = "STANDARD"
)

var (
	ErrUnsupportedCostingMethod = errors.New("unsupported costing method")
	ErrInsufficientCostLayers   = errors.New("insufficient active cost layers to satisfy issue quantity")
	ErrNegativeUnitCost         = errors.New("unit cost cannot be negative")
	ErrZeroQuantityRevaluation  = errors.New("cannot revalue item with zero stock on hand")
	ErrRevaluationNotFound      = errors.New("inventory revaluation not found")
)

type CostLayer struct {
	ID                uuid.UUID       `json:"id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	ItemID            uuid.UUID       `json:"item_id"`
	WarehouseID       uuid.UUID       `json:"warehouse_id"`
	SourceMovementID  uuid.UUID       `json:"source_movement_id"`
	OriginalQuantity  decimal.Decimal `json:"original_quantity"`
	RemainingQuantity decimal.Decimal `json:"remaining_quantity"`
	UnitCost          decimal.Decimal `json:"unit_cost"`
	TotalCost         decimal.Decimal `json:"total_cost"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	ItemCode          string          `json:"item_code,omitempty"`
	ItemNameEn        string          `json:"item_name_en,omitempty"`
	ItemNameAr        string          `json:"item_name_ar,omitempty"`
	WarehouseCode     string          `json:"warehouse_code,omitempty"`
	WarehouseNameAr   string          `json:"warehouse_name_ar,omitempty"`
}

type CostLayerConsumption struct {
	ID               uuid.UUID       `json:"id"`
	LegalEntityID    uuid.UUID       `json:"legal_entity_id"`
	CostLayerID      uuid.UUID       `json:"cost_layer_id"`
	IssueMovementID  uuid.UUID       `json:"issue_movement_id"`
	QuantityConsumed decimal.Decimal `json:"quantity_consumed"`
	UnitCost         decimal.Decimal `json:"unit_cost"`
	TotalCost        decimal.Decimal `json:"total_cost"`
	ConsumedAt       time.Time       `json:"consumed_at"`
}

type InventoryRevaluation struct {
	ID                uuid.UUID       `json:"id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	RevaluationNumber string          `json:"revaluation_number"`
	WarehouseID       uuid.UUID       `json:"warehouse_id"`
	ItemID            uuid.UUID       `json:"item_id"`
	QuantityOnHand    decimal.Decimal `json:"quantity_on_hand"`
	OldUnitCost       decimal.Decimal `json:"old_unit_cost"`
	NewUnitCost       decimal.Decimal `json:"new_unit_cost"`
	OldTotalValue     decimal.Decimal `json:"old_total_value"`
	NewTotalValue     decimal.Decimal `json:"new_total_value"`
	VarianceAmount    decimal.Decimal `json:"variance_amount"`
	Reason            string          `json:"reason"`
	JournalEntryID    *uuid.UUID      `json:"journal_entry_id,omitempty"`
	CreatedBy         uuid.UUID       `json:"created_by"`
	CreatedAt         time.Time       `json:"created_at"`
	ItemCode          string          `json:"item_code,omitempty"`
	ItemNameEn        string          `json:"item_name_en,omitempty"`
	ItemNameAr        string          `json:"item_name_ar,omitempty"`
	WarehouseCode     string          `json:"warehouse_code,omitempty"`
	WarehouseNameAr   string          `json:"warehouse_name_ar,omitempty"`
}

type ValuationSummaryItem struct {
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	ItemID            uuid.UUID       `json:"item_id"`
	WarehouseID       uuid.UUID       `json:"warehouse_id"`
	QuantityOnHand    decimal.Decimal `json:"quantity_on_hand"`
	QuantityReserved  decimal.Decimal `json:"quantity_reserved"`
	QuantityAvailable decimal.Decimal `json:"quantity_available"`
	AverageCost       decimal.Decimal `json:"average_cost"`
	TotalValue        decimal.Decimal `json:"total_value"`
	ItemCode          string          `json:"item_code"`
	ItemNameEn        string          `json:"item_name_en"`
	ItemNameAr        string          `json:"item_name_ar"`
	CostingMethod     CostingMethod   `json:"costing_method"`
	StandardCost      decimal.Decimal `json:"standard_cost"`
	WarehouseCode     string          `json:"warehouse_code"`
	WarehouseNameAr   string          `json:"warehouse_name_ar"`
}

type CostCalculationResult struct {
	UnitCost     decimal.Decimal
	TotalCost    decimal.Decimal
	Consumptions []CostLayerConsumption
}

type ItemCostingInfo struct {
	ID            uuid.UUID
	CostingMethod CostingMethod
	CostPrice     decimal.Decimal
	StandardCost  decimal.Decimal
}
