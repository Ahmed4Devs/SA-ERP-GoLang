package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Domain Errors
var (
	ErrPurchaseRequestNotFound = errors.New("purchase request not found")
	ErrRFQNotFound             = errors.New("request for quotation not found")
	ErrSupplierQuoteNotFound   = errors.New("supplier quote not found")
	ErrPurchaseOrderNotFound   = errors.New("purchase order not found")
	ErrGoodsReceiptNotFound    = errors.New("goods receipt not found")
	ErrInvalidPOStatus         = errors.New("purchase order status does not permit this operation")
	ErrInvalidGRNStatus        = errors.New("goods receipt status does not permit this operation")
	ErrPONotApproved           = errors.New("purchase order must be approved before issuance or receipt")
	ErrPONotIssued             = errors.New("purchase order must be issued before goods can be received")
	ErrEmptyLineItems          = errors.New("purchasing document must contain at least one line item")
	ErrQuantityExceedsOrdered  = errors.New("received quantity exceeds remaining ordered quantity")
	ErrThreeWayMatchMismatch   = errors.New("three-way matching variance exceeds allowed tolerance")
)

// Purchase Request Status
type PRStatus string

const (
	PRStatusDraft            PRStatus = "DRAFT"
	PRStatusSubmitted        PRStatus = "SUBMITTED"
	PRStatusApproved         PRStatus = "APPROVED"
	PRStatusRejected         PRStatus = "REJECTED"
	PRStatusPartiallyOrdered PRStatus = "PARTIALLY_ORDERED"
	PRStatusOrdered          PRStatus = "ORDERED"
	PRStatusCancelled        PRStatus = "CANCELLED"
)

// RFQ Status
type RFQStatus string

const (
	RFQStatusDraft     RFQStatus = "DRAFT"
	RFQStatusSent      RFQStatus = "SENT"
	RFQStatusClosed    RFQStatus = "CLOSED"
	RFQStatusCancelled RFQStatus = "CANCELLED"
)

// Supplier Quote Status
type QuoteStatus string

const (
	QuoteStatusPending  QuoteStatus = "PENDING"
	QuoteStatusAccepted QuoteStatus = "ACCEPTED"
	QuoteStatusRejected QuoteStatus = "REJECTED"
	QuoteStatusExpired  QuoteStatus = "EXPIRED"
)

// Purchase Order Status
type POStatus string

const (
	POStatusDraft             POStatus = "DRAFT"
	POStatusApproved          POStatus = "APPROVED"
	POStatusIssued            POStatus = "ISSUED"
	POStatusPartiallyReceived POStatus = "PARTIALLY_RECEIVED"
	POStatusReceived          POStatus = "RECEIVED"
	POStatusClosed            POStatus = "CLOSED"
	POStatusCancelled         POStatus = "CANCELLED"
)

// Goods Receipt Status
type GRNStatus string

const (
	GRNStatusDraft     GRNStatus = "DRAFT"
	GRNStatusAccepted  GRNStatus = "ACCEPTED"
	GRNStatusRejected  GRNStatus = "REJECTED"
	GRNStatusCancelled GRNStatus = "CANCELLED"
)

// Three-Way Match Status
type MatchStatus string

const (
	MatchStatusPerfectMatch     MatchStatus = "PERFECT_MATCH"
	MatchStatusPriceVariance    MatchStatus = "PRICE_VARIANCE"
	MatchStatusQuantityVariance MatchStatus = "QUANTITY_VARIANCE"
	MatchStatusBothVariance     MatchStatus = "BOTH_VARIANCE"
	MatchStatusUnmatched        MatchStatus = "UNMATCHED"
)

// ----------------------------------------------------------------------------
// Domain Entities
// ----------------------------------------------------------------------------

type PurchaseRequest struct {
	ID                uuid.UUID             `json:"id"`
	LegalEntityID     uuid.UUID             `json:"legal_entity_id"`
	PRNumber          string                `json:"pr_number"`
	RequestedByUserID uuid.UUID             `json:"requested_by_user_id"`
	DepartmentID      *uuid.UUID            `json:"department_id,omitempty"`
	RequestDate       time.Time             `json:"request_date"`
	RequiredByDate    *time.Time            `json:"required_by_date,omitempty"`
	Status            PRStatus              `json:"status"`
	Purpose           *string               `json:"purpose,omitempty"`
	Notes             *string               `json:"notes,omitempty"`
	EstimatedTotal    decimal.Decimal       `json:"estimated_total"`
	CurrencyCode      string                `json:"currency_code"`
	Lines             []PurchaseRequestLine `json:"lines,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

type PurchaseRequestLine struct {
	ID                  uuid.UUID       `json:"id"`
	PurchaseRequestID   uuid.UUID       `json:"purchase_request_id"`
	LineNumber          int             `json:"line_number"`
	ItemID              *uuid.UUID      `json:"item_id,omitempty"`
	Description         string          `json:"description"`
	UOM                 string          `json:"uom"`
	Quantity            decimal.Decimal `json:"quantity"`
	OrderedQuantity     decimal.Decimal `json:"ordered_quantity"`
	EstimatedUnitCost   decimal.Decimal `json:"estimated_unit_cost"`
	EstimatedTotalCost  decimal.Decimal `json:"estimated_total_cost"`
	SuggestedSupplierID *uuid.UUID      `json:"suggested_supplier_id,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

type RFQ struct {
	ID                uuid.UUID  `json:"id"`
	LegalEntityID     uuid.UUID  `json:"legal_entity_id"`
	RFQNumber         string     `json:"rfq_number"`
	PurchaseRequestID *uuid.UUID `json:"purchase_request_id,omitempty"`
	Title             string     `json:"title"`
	RFQDate           time.Time  `json:"rfq_date"`
	ClosingDate       time.Time  `json:"closing_date"`
	Status            RFQStatus  `json:"status"`
	Notes             *string    `json:"notes,omitempty"`
	CreatedByUserID   *uuid.UUID `json:"created_by_user_id,omitempty"`
	Lines             []RFQLine  `json:"lines,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type RFQLine struct {
	ID              uuid.UUID        `json:"id"`
	RFQID           uuid.UUID        `json:"rfq_id"`
	LineNumber      int              `json:"line_number"`
	PRLineID        *uuid.UUID       `json:"pr_line_id,omitempty"`
	ItemID          *uuid.UUID       `json:"item_id,omitempty"`
	Description     string           `json:"description"`
	UOM             string           `json:"uom"`
	Quantity        decimal.Decimal  `json:"quantity"`
	TargetUnitPrice *decimal.Decimal `json:"target_unit_price,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}

type SupplierQuote struct {
	ID                     uuid.UUID           `json:"id"`
	LegalEntityID          uuid.UUID           `json:"legal_entity_id"`
	RFQID                  *uuid.UUID          `json:"rfq_id,omitempty"`
	SupplierID             uuid.UUID           `json:"supplier_id"`
	SupplierQuoteReference string              `json:"supplier_quote_reference"`
	QuoteDate              time.Time           `json:"quote_date"`
	ValidUntil             time.Time           `json:"valid_until"`
	CurrencyCode           string              `json:"currency_code"`
	ExchangeRate           decimal.Decimal     `json:"exchange_rate"`
	PaymentTermsID         *uuid.UUID          `json:"payment_terms_id,omitempty"`
	Subtotal               decimal.Decimal     `json:"subtotal"`
	TaxAmount              decimal.Decimal     `json:"tax_amount"`
	TotalAmount            decimal.Decimal     `json:"total_amount"`
	Status                 QuoteStatus         `json:"status"`
	Notes                  *string             `json:"notes,omitempty"`
	Lines                  []SupplierQuoteLine `json:"lines,omitempty"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
}

type SupplierQuoteLine struct {
	ID              uuid.UUID       `json:"id"`
	QuoteID         uuid.UUID       `json:"quote_id"`
	LineNumber      int             `json:"line_number"`
	RFQLineID       *uuid.UUID      `json:"rfq_line_id,omitempty"`
	ItemID          *uuid.UUID      `json:"item_id,omitempty"`
	Description     string          `json:"description"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	TaxRatePercent  decimal.Decimal `json:"tax_rate_percent"`
	TaxAmount       decimal.Decimal `json:"tax_amount"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	IsAwarded       bool            `json:"is_awarded"`
	CreatedAt       time.Time       `json:"created_at"`
}

type PurchaseOrder struct {
	ID                   uuid.UUID           `json:"id"`
	LegalEntityID        uuid.UUID           `json:"legal_entity_id"`
	PONumber             string              `json:"po_number"`
	PurchaseRequestID    *uuid.UUID          `json:"purchase_request_id,omitempty"`
	RFQID                *uuid.UUID          `json:"rfq_id,omitempty"`
	SupplierQuoteID      *uuid.UUID          `json:"supplier_quote_id,omitempty"`
	SupplierID           uuid.UUID           `json:"supplier_id"`
	OrderDate            time.Time           `json:"order_date"`
	ExpectedDeliveryDate *time.Time          `json:"expected_delivery_date,omitempty"`
	PaymentTermsID       *uuid.UUID          `json:"payment_terms_id,omitempty"`
	DeliveryWarehouseID  *uuid.UUID          `json:"delivery_warehouse_id,omitempty"`
	CurrencyCode         string              `json:"currency_code"`
	ExchangeRate         decimal.Decimal     `json:"exchange_rate"`
	Subtotal             decimal.Decimal     `json:"subtotal"`
	DiscountAmount       decimal.Decimal     `json:"discount_amount"`
	TaxableAmount        decimal.Decimal     `json:"taxable_amount"`
	TaxAmount            decimal.Decimal     `json:"tax_amount"`
	TotalAmount          decimal.Decimal     `json:"total_amount"`
	Status               POStatus            `json:"status"`
	Notes                *string             `json:"notes,omitempty"`
	CreatedByUserID      *uuid.UUID          `json:"created_by_user_id,omitempty"`
	ApprovedByUserID     *uuid.UUID          `json:"approved_by_user_id,omitempty"`
	ApprovedAt           *time.Time          `json:"approved_at,omitempty"`
	IssuedAt             *time.Time          `json:"issued_at,omitempty"`
	Lines                []PurchaseOrderLine `json:"lines,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

type PurchaseOrderLine struct {
	ID               uuid.UUID       `json:"id"`
	PurchaseOrderID  uuid.UUID       `json:"purchase_order_id"`
	LineNumber       int             `json:"line_number"`
	PRLineID         *uuid.UUID      `json:"pr_line_id,omitempty"`
	ItemID           uuid.UUID       `json:"item_id"`
	Description      string          `json:"description"`
	UOM              string          `json:"uom"`
	Quantity         decimal.Decimal `json:"quantity"`
	ReceivedQuantity decimal.Decimal `json:"received_quantity"`
	BilledQuantity   decimal.Decimal `json:"billed_quantity"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	DiscountPercent  decimal.Decimal `json:"discount_percent"`
	DiscountAmount   decimal.Decimal `json:"discount_amount"`
	TaxableAmount    decimal.Decimal `json:"taxable_amount"`
	TaxCode          string          `json:"tax_code"`
	TaxRatePercent   decimal.Decimal `json:"tax_rate_percent"`
	TaxAmount        decimal.Decimal `json:"tax_amount"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	ExpenseAccountID *uuid.UUID      `json:"expense_account_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

type GoodsReceipt struct {
	ID                 uuid.UUID          `json:"id"`
	LegalEntityID      uuid.UUID          `json:"legal_entity_id"`
	ReceiptNumber      string             `json:"receipt_number"`
	PurchaseOrderID    uuid.UUID          `json:"purchase_order_id"`
	SupplierID         uuid.UUID          `json:"supplier_id"`
	WarehouseID        uuid.UUID          `json:"warehouse_id"`
	ReceiptDate        time.Time          `json:"receipt_date"`
	DeliveryNoteNumber *string            `json:"delivery_note_number,omitempty"`
	Status             GRNStatus          `json:"status"`
	JournalEntryID     *uuid.UUID         `json:"journal_entry_id,omitempty"`
	Notes              *string            `json:"notes,omitempty"`
	ReceivedByUserID   *uuid.UUID         `json:"received_by_user_id,omitempty"`
	InspectedByUserID  *uuid.UUID         `json:"inspected_by_user_id,omitempty"`
	Lines              []GoodsReceiptLine `json:"lines,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type GoodsReceiptLine struct {
	ID                  uuid.UUID       `json:"id"`
	GoodsReceiptID      uuid.UUID       `json:"goods_receipt_id"`
	LineNumber          int             `json:"line_number"`
	POLineID            uuid.UUID       `json:"po_line_id"`
	ItemID              uuid.UUID       `json:"item_id"`
	WarehouseID         uuid.UUID       `json:"warehouse_id"`
	BinID               *uuid.UUID      `json:"bin_id,omitempty"`
	ReceivedQuantity    decimal.Decimal `json:"received_quantity"`
	AcceptedQuantity    decimal.Decimal `json:"accepted_quantity"`
	RejectedQuantity    decimal.Decimal `json:"rejected_quantity"`
	UnitCost            decimal.Decimal `json:"unit_cost"`
	TotalCost           decimal.Decimal `json:"total_cost"`
	InventoryMovementID *uuid.UUID      `json:"inventory_movement_id,omitempty"`
	RejectionReason     *string         `json:"rejection_reason,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
}

// ThreeWayMatchResult represents the outcome of comparing PO, Goods Receipt, and Supplier Invoice
type ThreeWayMatchResult struct {
	POLineID              uuid.UUID       `json:"po_line_id"`
	GRNLineID             uuid.UUID       `json:"grn_line_id"`
	InvoiceLineID         uuid.UUID       `json:"invoice_line_id"`
	POQuantity            decimal.Decimal `json:"po_quantity"`
	GRNAcceptedQuantity   decimal.Decimal `json:"grn_accepted_quantity"`
	InvoicedQuantity      decimal.Decimal `json:"invoiced_quantity"`
	POUnitPrice           decimal.Decimal `json:"po_unit_price"`
	InvoiceUnitPrice      decimal.Decimal `json:"invoice_unit_price"`
	QuantityVariance      decimal.Decimal `json:"quantity_variance"`
	PriceVariance         decimal.Decimal `json:"price_variance"`
	MatchStatus           MatchStatus     `json:"match_status"`
	PassedTolerance       bool            `json:"passed_tolerance"`
	Notes                 string          `json:"notes"`
}

// CalculateLineAmounts computes gross, discount, taxable, tax, and total
func CalculateLineAmounts(qty, unitPrice, discountPercent, taxRatePercent decimal.Decimal) (gross, disc, taxable, tax, total decimal.Decimal) {
	gross = qty.Mul(unitPrice).Round(4)
	hundred := decimal.NewFromInt(100)
	if discountPercent.GreaterThan(decimal.Zero) {
		disc = gross.Mul(discountPercent).Div(hundred).Round(4)
	} else {
		disc = decimal.Zero
	}
	taxable = gross.Sub(disc).Round(4)
	if taxRatePercent.GreaterThan(decimal.Zero) {
		tax = taxable.Mul(taxRatePercent).Div(hundred).Round(4)
	} else {
		tax = decimal.Zero
	}
	total = taxable.Add(tax).Round(4)
	return
}

// PerformThreeWayMatch evaluates PO line, GRN line, and invoice line
func PerformThreeWayMatch(
	poLine PurchaseOrderLine,
	grnLine GoodsReceiptLine,
	invQty, invPrice decimal.Decimal,
	priceTolerancePercent, qtyTolerancePercent decimal.Decimal,
) ThreeWayMatchResult {
	result := ThreeWayMatchResult{
		POLineID:            poLine.ID,
		GRNLineID:           grnLine.ID,
		POQuantity:          poLine.Quantity,
		GRNAcceptedQuantity: grnLine.AcceptedQuantity,
		InvoicedQuantity:    invQty,
		POUnitPrice:         poLine.UnitPrice,
		InvoiceUnitPrice:    invPrice,
	}

	result.QuantityVariance = invQty.Sub(grnLine.AcceptedQuantity)
	result.PriceVariance = invPrice.Sub(poLine.UnitPrice)

	// Evaluate tolerances
	qtyDiff := result.QuantityVariance.Abs()
	priceDiff := result.PriceVariance.Abs()

	qtyPassed := qtyDiff.IsZero()
	if !qtyPassed && grnLine.AcceptedQuantity.GreaterThan(decimal.Zero) && qtyTolerancePercent.GreaterThan(decimal.Zero) {
		maxQtyDiff := grnLine.AcceptedQuantity.Mul(qtyTolerancePercent).Div(decimal.NewFromInt(100))
		qtyPassed = qtyDiff.LessThanOrEqual(maxQtyDiff)
	}

	pricePassed := priceDiff.IsZero()
	if !pricePassed && poLine.UnitPrice.GreaterThan(decimal.Zero) && priceTolerancePercent.GreaterThan(decimal.Zero) {
		maxPriceDiff := poLine.UnitPrice.Mul(priceTolerancePercent).Div(decimal.NewFromInt(100))
		pricePassed = priceDiff.LessThanOrEqual(maxPriceDiff)
	}

	result.PassedTolerance = qtyPassed && pricePassed

	if qtyDiff.IsZero() && priceDiff.IsZero() {
		result.MatchStatus = MatchStatusPerfectMatch
		result.Notes = "Matched perfectly on quantity and price"
	} else if !qtyDiff.IsZero() && !priceDiff.IsZero() {
		result.MatchStatus = MatchStatusBothVariance
		result.Notes = "Variances detected in both quantity and unit price"
	} else if !qtyDiff.IsZero() {
		result.MatchStatus = MatchStatusQuantityVariance
		result.Notes = "Variance detected in invoiced quantity vs accepted goods receipt"
	} else {
		result.MatchStatus = MatchStatusPriceVariance
		result.Notes = "Variance detected in invoiced unit price vs purchase order"
	}

	return result
}
