package http

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatePRRequest struct {
	LegalEntityID  uuid.UUID             `json:"legal_entity_id"`
	DepartmentID   *uuid.UUID            `json:"department_id"`
	RequestDate    string                `json:"request_date"`
	RequiredByDate *string               `json:"required_by_date"`
	Purpose        *string               `json:"purpose"`
	Notes          *string               `json:"notes"`
	CurrencyCode   string                `json:"currency_code"`
	Lines          []CreatePRLineRequest `json:"lines"`
}

type CreatePRLineRequest struct {
	ItemID              *uuid.UUID      `json:"item_id"`
	Description         string          `json:"description"`
	UOM                 string          `json:"uom"`
	Quantity            decimal.Decimal `json:"quantity"`
	EstimatedUnitCost   decimal.Decimal `json:"estimated_unit_cost"`
	SuggestedSupplierID *uuid.UUID      `json:"suggested_supplier_id"`
}

type CreateRFQRequest struct {
	LegalEntityID     uuid.UUID              `json:"legal_entity_id"`
	PurchaseRequestID *uuid.UUID             `json:"purchase_request_id"`
	Title             string                 `json:"title"`
	RFQDate           string                 `json:"rfq_date"`
	ClosingDate       string                 `json:"closing_date"`
	Notes             *string                `json:"notes"`
	Lines             []CreateRFQLineRequest `json:"lines"`
}

type CreateRFQLineRequest struct {
	PRLineID        *uuid.UUID       `json:"pr_line_id"`
	ItemID          *uuid.UUID       `json:"item_id"`
	Description     string           `json:"description"`
	UOM             string           `json:"uom"`
	Quantity        decimal.Decimal  `json:"quantity"`
	TargetUnitPrice *decimal.Decimal `json:"target_unit_price"`
}

type RecordQuoteRequest struct {
	LegalEntityID          uuid.UUID                `json:"legal_entity_id"`
	RFQID                  *uuid.UUID               `json:"rfq_id"`
	SupplierID             uuid.UUID                `json:"supplier_id"`
	SupplierQuoteReference string                   `json:"supplier_quote_reference"`
	QuoteDate              string                   `json:"quote_date"`
	ValidUntil             string                   `json:"valid_until"`
	CurrencyCode           string                   `json:"currency_code"`
	ExchangeRate           decimal.Decimal          `json:"exchange_rate"`
	PaymentTermsID         *uuid.UUID               `json:"payment_terms_id"`
	Notes                  *string                  `json:"notes"`
	Lines                  []RecordQuoteLineRequest `json:"lines"`
}

type RecordQuoteLineRequest struct {
	RFQLineID       *uuid.UUID      `json:"rfq_line_id"`
	ItemID          *uuid.UUID      `json:"item_id"`
	Description     string          `json:"description"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	TaxRatePercent  decimal.Decimal `json:"tax_rate_percent"`
}

type CreatePORequest struct {
	LegalEntityID        uuid.UUID             `json:"legal_entity_id"`
	PurchaseRequestID    *uuid.UUID            `json:"purchase_request_id"`
	RFQID                *uuid.UUID            `json:"rfq_id"`
	SupplierQuoteID      *uuid.UUID            `json:"supplier_quote_id"`
	SupplierID           uuid.UUID             `json:"supplier_id"`
	OrderDate            string                `json:"order_date"`
	ExpectedDeliveryDate *string               `json:"expected_delivery_date"`
	PaymentTermsID       *uuid.UUID            `json:"payment_terms_id"`
	DeliveryWarehouseID  *uuid.UUID            `json:"delivery_warehouse_id"`
	CurrencyCode         string                `json:"currency_code"`
	ExchangeRate         decimal.Decimal       `json:"exchange_rate"`
	Notes                *string               `json:"notes"`
	Lines                []CreatePOLineRequest `json:"lines"`
}

type CreatePOLineRequest struct {
	PRLineID         *uuid.UUID      `json:"pr_line_id"`
	ItemID           uuid.UUID       `json:"item_id"`
	Description      string          `json:"description"`
	UOM              string          `json:"uom"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	DiscountPercent  decimal.Decimal `json:"discount_percent"`
	TaxCode          string          `json:"tax_code"`
	TaxRatePercent   decimal.Decimal `json:"tax_rate_percent"`
	ExpenseAccountID *uuid.UUID      `json:"expense_account_id"`
}

type CreateReceiptRequest struct {
	LegalEntityID      uuid.UUID                  `json:"legal_entity_id"`
	PurchaseOrderID    uuid.UUID                  `json:"purchase_order_id"`
	SupplierID         uuid.UUID                  `json:"supplier_id"`
	WarehouseID        uuid.UUID                  `json:"warehouse_id"`
	ReceiptDate        string                     `json:"receipt_date"`
	DeliveryNoteNumber *string                    `json:"delivery_note_number"`
	Notes              *string                    `json:"notes"`
	Lines              []CreateReceiptLineRequest `json:"lines"`
}

type CreateReceiptLineRequest struct {
	POLineID         uuid.UUID       `json:"po_line_id"`
	ItemID           uuid.UUID       `json:"item_id"`
	WarehouseID      uuid.UUID       `json:"warehouse_id"`
	BinID            *uuid.UUID      `json:"bin_id"`
	ReceivedQuantity decimal.Decimal `json:"received_quantity"`
	AcceptedQuantity decimal.Decimal `json:"accepted_quantity"`
	RejectedQuantity decimal.Decimal `json:"rejected_quantity"`
	UnitCost         decimal.Decimal `json:"unit_cost"`
	RejectionReason  *string         `json:"rejection_reason"`
}

type ThreeWayMatchRequest struct {
	POLineID              uuid.UUID       `json:"po_line_id"`
	GRNLineID             uuid.UUID       `json:"grn_line_id"`
	InvoiceLineID         uuid.UUID       `json:"invoice_line_id"`
	InvoicedQuantity      decimal.Decimal `json:"invoiced_quantity"`
	InvoicedUnitPrice     decimal.Decimal `json:"invoiced_unit_price"`
	PriceTolerancePercent decimal.Decimal `json:"price_tolerance_percent"`
	QtyTolerancePercent   decimal.Decimal `json:"qty_tolerance_percent"`
}
