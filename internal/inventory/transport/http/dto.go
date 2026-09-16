package http

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateWarehouseRequestDTO struct {
	Code          string `json:"code"`
	NameEn        string `json:"name_en"`
	NameAr        string `json:"name_ar"`
	WarehouseType string `json:"warehouse_type"`
	AddressLine1  string `json:"address_line1,omitempty"`
	City          string `json:"city,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	CountryCode   string `json:"country_code,omitempty"`
}

type CreateBinRequestDTO struct {
	Code      string `json:"code"`
	Aisle     string `json:"aisle,omitempty"`
	Rack      string `json:"rack,omitempty"`
	Shelf     string `json:"shelf,omitempty"`
	Bin       string `json:"bin,omitempty"`
	IsDefault bool   `json:"is_default"`
}

type PostMovementRequestDTO struct {
	MovementType    string          `json:"movement_type"`
	ItemID          uuid.UUID       `json:"item_id"`
	SKUID           *uuid.UUID      `json:"sku_id,omitempty"`
	WarehouseID     uuid.UUID       `json:"warehouse_id"`
	BinID           *uuid.UUID      `json:"bin_id,omitempty"`
	LotNumber       *string         `json:"lot_number,omitempty"`
	SerialNumber    *string         `json:"serial_number,omitempty"`
	Quantity        decimal.Decimal `json:"quantity"`
	UOMID           uuid.UUID       `json:"uom_id"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	SourceDocType   *string         `json:"source_doc_type,omitempty"`
	SourceDocID     *uuid.UUID      `json:"source_doc_id,omitempty"`
	SourceDocLineID *uuid.UUID      `json:"source_doc_line_id,omitempty"`
}

type ReserveStockRequestDTO struct {
	ItemID           uuid.UUID       `json:"item_id"`
	WarehouseID      uuid.UUID       `json:"warehouse_id"`
	BinID            *uuid.UUID      `json:"bin_id,omitempty"`
	Quantity         decimal.Decimal `json:"quantity"`
	DemandSourceType string          `json:"demand_source_type"`
	DemandSourceID   uuid.UUID       `json:"demand_source_id"`
}

type TransferLineDTO struct {
	ItemID            uuid.UUID       `json:"item_id"`
	SKUID             *uuid.UUID      `json:"sku_id,omitempty"`
	RequestedQuantity decimal.Decimal `json:"requested_quantity"`
	UOMID             uuid.UUID       `json:"uom_id"`
	Notes             string          `json:"notes,omitempty"`
}

type CreateTransferRequestDTO struct {
	FromWarehouseID uuid.UUID         `json:"from_warehouse_id"`
	ToWarehouseID   uuid.UUID         `json:"to_warehouse_id"`
	Notes           string            `json:"notes,omitempty"`
	Lines           []TransferLineDTO `json:"lines"`
}

type AdjustmentLineDTO struct {
	ItemID       uuid.UUID       `json:"item_id"`
	BinID        *uuid.UUID      `json:"bin_id,omitempty"`
	LotNumber    *string         `json:"lot_number,omitempty"`
	SerialNumber *string         `json:"serial_number,omitempty"`
	Quantity     decimal.Decimal `json:"quantity"`
	UnitCost     decimal.Decimal `json:"unit_cost"`
	Reason       string          `json:"reason,omitempty"`
}

type CreateAdjustmentRequestDTO struct {
	WarehouseID    uuid.UUID           `json:"warehouse_id"`
	AdjustmentType string              `json:"adjustment_type"`
	ReasonCode     string              `json:"reason_code"`
	Notes          string              `json:"notes,omitempty"`
	Lines          []AdjustmentLineDTO `json:"lines"`
}

type CountLineDTO struct {
	ItemID          uuid.UUID       `json:"item_id"`
	BinID           *uuid.UUID      `json:"bin_id,omitempty"`
	CountedQuantity decimal.Decimal `json:"counted_quantity"`
	Notes           string          `json:"notes,omitempty"`
}

type CreateCountRequestDTO struct {
	WarehouseID uuid.UUID      `json:"warehouse_id"`
	CountType   string         `json:"count_type"`
	Lines       []CountLineDTO `json:"lines"`
}

type ExecuteRevaluationRequestDTO struct {
	WarehouseID    uuid.UUID       `json:"warehouse_id"`
	ItemID         uuid.UUID       `json:"item_id"`
	NewUnitCost    decimal.Decimal `json:"new_unit_cost"`
	Reason         string          `json:"reason"`
	JournalEntryID *uuid.UUID      `json:"journal_entry_id,omitempty"`
}
