package ports

import (
	"context"
	"time"

	"erp/internal/purchasing/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Command types
type CreatePRLineCmd struct {
	ItemID              *uuid.UUID      `json:"item_id,omitempty"`
	Description         string          `json:"description"`
	UOM                 string          `json:"uom"`
	Quantity            decimal.Decimal `json:"quantity"`
	EstimatedUnitCost   decimal.Decimal `json:"estimated_unit_cost"`
	SuggestedSupplierID *uuid.UUID      `json:"suggested_supplier_id,omitempty"`
}

type CreatePRCmd struct {
	LegalEntityID     uuid.UUID         `json:"legal_entity_id"`
	RequestedByUserID uuid.UUID         `json:"requested_by_user_id"`
	DepartmentID      *uuid.UUID        `json:"department_id,omitempty"`
	RequestDate       time.Time         `json:"request_date"`
	RequiredByDate    *time.Time        `json:"required_by_date,omitempty"`
	Purpose           *string           `json:"purpose,omitempty"`
	Notes             *string           `json:"notes,omitempty"`
	CurrencyCode      string            `json:"currency_code"`
	Lines             []CreatePRLineCmd `json:"lines"`
}

type CreateRFQLineCmd struct {
	PRLineID        *uuid.UUID       `json:"pr_line_id,omitempty"`
	ItemID          *uuid.UUID       `json:"item_id,omitempty"`
	Description     string           `json:"description"`
	UOM             string           `json:"uom"`
	Quantity        decimal.Decimal  `json:"quantity"`
	TargetUnitPrice *decimal.Decimal `json:"target_unit_price,omitempty"`
}

type CreateRFQCmd struct {
	LegalEntityID     uuid.UUID          `json:"legal_entity_id"`
	PurchaseRequestID *uuid.UUID         `json:"purchase_request_id,omitempty"`
	Title             string             `json:"title"`
	RFQDate           time.Time          `json:"rfq_date"`
	ClosingDate       time.Time          `json:"closing_date"`
	Notes             *string            `json:"notes,omitempty"`
	CreatedByUserID   *uuid.UUID         `json:"created_by_user_id,omitempty"`
	Lines             []CreateRFQLineCmd `json:"lines"`
}

type RecordQuoteLineCmd struct {
	RFQLineID       *uuid.UUID      `json:"rfq_line_id,omitempty"`
	ItemID          *uuid.UUID      `json:"item_id,omitempty"`
	Description     string          `json:"description"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	TaxRatePercent  decimal.Decimal `json:"tax_rate_percent"`
}

type RecordQuoteCmd struct {
	LegalEntityID          uuid.UUID            `json:"legal_entity_id"`
	RFQID                  *uuid.UUID           `json:"rfq_id,omitempty"`
	SupplierID             uuid.UUID            `json:"supplier_id"`
	SupplierQuoteReference string               `json:"supplier_quote_reference"`
	QuoteDate              time.Time            `json:"quote_date"`
	ValidUntil             time.Time            `json:"valid_until"`
	CurrencyCode           string               `json:"currency_code"`
	ExchangeRate           decimal.Decimal      `json:"exchange_rate"`
	PaymentTermsID         *uuid.UUID           `json:"payment_terms_id,omitempty"`
	Notes                  *string              `json:"notes,omitempty"`
	Lines                  []RecordQuoteLineCmd `json:"lines"`
}

type CreatePOLineCmd struct {
	PRLineID         *uuid.UUID      `json:"pr_line_id,omitempty"`
	ItemID           uuid.UUID       `json:"item_id"`
	Description      string          `json:"description"`
	UOM              string          `json:"uom"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	DiscountPercent  decimal.Decimal `json:"discount_percent"`
	TaxCode          string          `json:"tax_code"`
	TaxRatePercent   decimal.Decimal `json:"tax_rate_percent"`
	ExpenseAccountID *uuid.UUID      `json:"expense_account_id,omitempty"`
}

type CreatePOCmd struct {
	LegalEntityID        uuid.UUID         `json:"legal_entity_id"`
	PurchaseRequestID    *uuid.UUID        `json:"purchase_request_id,omitempty"`
	RFQID                *uuid.UUID        `json:"rfq_id,omitempty"`
	SupplierQuoteID      *uuid.UUID        `json:"supplier_quote_id,omitempty"`
	SupplierID           uuid.UUID         `json:"supplier_id"`
	OrderDate            time.Time         `json:"order_date"`
	ExpectedDeliveryDate *time.Time        `json:"expected_delivery_date,omitempty"`
	PaymentTermsID       *uuid.UUID        `json:"payment_terms_id,omitempty"`
	DeliveryWarehouseID  *uuid.UUID        `json:"delivery_warehouse_id,omitempty"`
	CurrencyCode         string            `json:"currency_code"`
	ExchangeRate         decimal.Decimal   `json:"exchange_rate"`
	Notes                *string           `json:"notes,omitempty"`
	CreatedByUserID      *uuid.UUID        `json:"created_by_user_id,omitempty"`
	Lines                []CreatePOLineCmd `json:"lines"`
}

type CreateReceiptLineCmd struct {
	POLineID         uuid.UUID       `json:"po_line_id"`
	ItemID           uuid.UUID       `json:"item_id"`
	WarehouseID      uuid.UUID       `json:"warehouse_id"`
	BinID            *uuid.UUID      `json:"bin_id,omitempty"`
	ReceivedQuantity decimal.Decimal `json:"received_quantity"`
	AcceptedQuantity decimal.Decimal `json:"accepted_quantity"`
	RejectedQuantity decimal.Decimal `json:"rejected_quantity"`
	UnitCost         decimal.Decimal `json:"unit_cost"`
	RejectionReason  *string         `json:"rejection_reason,omitempty"`
}

type CreateReceiptCmd struct {
	LegalEntityID      uuid.UUID              `json:"legal_entity_id"`
	PurchaseOrderID    uuid.UUID              `json:"purchase_order_id"`
	SupplierID         uuid.UUID              `json:"supplier_id"`
	WarehouseID        uuid.UUID              `json:"warehouse_id"`
	ReceiptDate        time.Time              `json:"receipt_date"`
	DeliveryNoteNumber *string                `json:"delivery_note_number,omitempty"`
	Notes              *string                `json:"notes,omitempty"`
	ReceivedByUserID   *uuid.UUID             `json:"received_by_user_id,omitempty"`
	Lines              []CreateReceiptLineCmd `json:"lines"`
}

type ThreeWayMatchReq struct {
	POLineID              uuid.UUID       `json:"po_line_id"`
	GRNLineID             uuid.UUID       `json:"grn_line_id"`
	InvoiceLineID         uuid.UUID       `json:"invoice_line_id"`
	InvoicedQuantity      decimal.Decimal `json:"invoiced_quantity"`
	InvoicedUnitPrice     decimal.Decimal `json:"invoiced_unit_price"`
	PriceTolerancePercent decimal.Decimal `json:"price_tolerance_percent"`
	QtyTolerancePercent   decimal.Decimal `json:"qty_tolerance_percent"`
}

// PurchasingService defines the business operations for P2P
type PurchasingService interface {
	// Purchase Requests
	CreatePurchaseRequest(ctx context.Context, cmd CreatePRCmd) (*domain.PurchaseRequest, error)
	GetPurchaseRequest(ctx context.Context, id uuid.UUID) (*domain.PurchaseRequest, error)
	ListPurchaseRequests(ctx context.Context, legalEntityID uuid.UUID, status *domain.PRStatus, limit, offset int) ([]*domain.PurchaseRequest, error)
	ApprovePurchaseRequest(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*domain.PurchaseRequest, error)

	// RFQs
	CreateRFQ(ctx context.Context, cmd CreateRFQCmd) (*domain.RFQ, error)
	GetRFQ(ctx context.Context, id uuid.UUID) (*domain.RFQ, error)
	ListRFQs(ctx context.Context, legalEntityID uuid.UUID, status *domain.RFQStatus, limit, offset int) ([]*domain.RFQ, error)

	// Supplier Quotes
	RecordSupplierQuote(ctx context.Context, cmd RecordQuoteCmd) (*domain.SupplierQuote, error)
	GetSupplierQuote(ctx context.Context, id uuid.UUID) (*domain.SupplierQuote, error)
	ListSupplierQuotes(ctx context.Context, legalEntityID uuid.UUID, rfqID, supplierID *uuid.UUID, status *domain.QuoteStatus, limit, offset int) ([]*domain.SupplierQuote, error)
	AwardSupplierQuote(ctx context.Context, quoteID uuid.UUID) (*domain.SupplierQuote, error)

	// Purchase Orders
	CreatePurchaseOrder(ctx context.Context, cmd CreatePOCmd) (*domain.PurchaseOrder, error)
	GetPurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error)
	ListPurchaseOrders(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID, status *domain.POStatus, limit, offset int) ([]*domain.PurchaseOrder, error)
	ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*domain.PurchaseOrder, error)
	IssuePurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error)

	// Goods Receipts & Inspection
	CreateGoodsReceipt(ctx context.Context, cmd CreateReceiptCmd) (*domain.GoodsReceipt, error)
	GetGoodsReceipt(ctx context.Context, id uuid.UUID) (*domain.GoodsReceipt, error)
	ListGoodsReceipts(ctx context.Context, legalEntityID uuid.UUID, poID, supplierID *uuid.UUID, status *domain.GRNStatus, limit, offset int) ([]*domain.GoodsReceipt, error)
	InspectAndAcceptGoodsReceipt(ctx context.Context, id uuid.UUID, inspectorID uuid.UUID) (*domain.GoodsReceipt, error)

	// Three-Way Matching
	VerifyThreeWayMatch(ctx context.Context, req ThreeWayMatchReq) (*domain.ThreeWayMatchResult, error)
}
