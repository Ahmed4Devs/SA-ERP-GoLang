package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	vatRegex = regexp.MustCompile(`^3[0-9]{13}3$`)
)

// Domain Errors
var (
	ErrQuotationNotFound              = errors.New("sales quotation not found")
	ErrOrderNotFound                  = errors.New("sales order not found")
	ErrDeliveryNotFound               = errors.New("goods delivery not found")
	ErrInvoiceNotFound                = errors.New("sales invoice not found")
	ErrCreditNoteNotFound             = errors.New("sales credit note not found")
	ErrQuotationNotAccepted           = errors.New("quotation must be in ACCEPTED status to convert to order")
	ErrOrderAlreadyDelivered          = errors.New("order has already been fully delivered")
	ErrOrderAlreadyInvoiced           = errors.New("order has already been fully invoiced")
	ErrInvalidOrderStatus             = errors.New("order status does not permit this action")
	ErrInvalidInvoiceStatus           = errors.New("invoice status does not permit this action")
	ErrInvoiceAlreadyPosted           = errors.New("invoice is already posted and immutable")
	ErrCustomerCreditLimitExceeded    = errors.New("customer credit limit exceeded")
	ErrCustomerCreditHold             = errors.New("customer account is on credit hold")
	ErrInvalidTaxIDForStandardInvoice = errors.New("B2B standard tax invoice requires a valid 15-digit ZATCA VAT registration number")
	ErrEmptyLineItems                 = errors.New("document must contain at least one line item")
	ErrInsufficientDeliveryStock      = errors.New("insufficient inventory stock to fulfill delivery")
	ErrOriginalInvoiceNotPosted       = errors.New("credit note can only be issued against a posted invoice")
)

// QuotationStatus represents the lifecycle of a quotation
type QuotationStatus string

const (
	QuotationStatusDraft     QuotationStatus = "DRAFT"
	QuotationStatusSent      QuotationStatus = "SENT"
	QuotationStatusAccepted  QuotationStatus = "ACCEPTED"
	QuotationStatusRejected  QuotationStatus = "REJECTED"
	QuotationStatusExpired   QuotationStatus = "EXPIRED"
	QuotationStatusConverted QuotationStatus = "CONVERTED"
)

// OrderStatus represents the lifecycle of a sales order
type OrderStatus string

const (
	OrderStatusDraft              OrderStatus = "DRAFT"
	OrderStatusConfirmed          OrderStatus = "CONFIRMED"
	OrderStatusPartiallyDelivered OrderStatus = "PARTIALLY_DELIVERED"
	OrderStatusDelivered          OrderStatus = "DELIVERED"
	OrderStatusPartiallyInvoiced  OrderStatus = "PARTIALLY_INVOICED"
	OrderStatusInvoiced           OrderStatus = "INVOICED"
	OrderStatusCancelled          OrderStatus = "CANCELLED"
)

// DeliveryStatus represents the status of a goods shipment/delivery
type DeliveryStatus string

const (
	DeliveryStatusDraft     DeliveryStatus = "DRAFT"
	DeliveryStatusShipped   DeliveryStatus = "SHIPPED"
	DeliveryStatusDelivered DeliveryStatus = "DELIVERED"
	DeliveryStatusCancelled DeliveryStatus = "CANCELLED"
)

// InvoiceType differentiates ZATCA standard vs simplified tax invoices
type InvoiceType string

const (
	InvoiceTypeStandard   InvoiceType = "STANDARD"   // B2B / Government
	InvoiceTypeSimplified InvoiceType = "SIMPLIFIED" // B2C
)

// InvoiceStatus represents invoice processing stages
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "DRAFT"
	InvoiceStatusSubmitted InvoiceStatus = "SUBMITTED"
	InvoiceStatusApproved  InvoiceStatus = "APPROVED"
	InvoiceStatusPosted    InvoiceStatus = "POSTED"
	InvoiceStatusCancelled InvoiceStatus = "CANCELLED"
)

// CreditNoteReason represents reason for credit note
type CreditNoteReason string

const (
	CreditNoteReasonReturn     CreditNoteReason = "RETURN"
	CreditNoteReasonDiscount   CreditNoteReason = "DISCOUNT"
	CreditNoteReasonCorrection CreditNoteReason = "CORRECTION"
	CreditNoteReasonOther      CreditNoteReason = "OTHER"
)

// CreditNoteStatus represents status of credit note
type CreditNoteStatus string

const (
	CreditNoteStatusDraft     CreditNoteStatus = "DRAFT"
	CreditNoteStatusPosted    CreditNoteStatus = "POSTED"
	CreditNoteStatusCancelled CreditNoteStatus = "CANCELLED"
)

// Quotation represents a price quote
type Quotation struct {
	ID               uuid.UUID
	LegalEntityID    uuid.UUID
	QuotationNumber  string
	CustomerID       uuid.UUID
	CustomerNameEn   string
	CustomerNameAr   string
	QuoteDate        time.Time
	ValidUntil       time.Time
	CurrencyCode     string
	ExchangeRate     decimal.Decimal
	Subtotal         decimal.Decimal
	DiscountAmount   decimal.Decimal
	TaxableAmount    decimal.Decimal
	TaxAmount        decimal.Decimal
	TotalAmount      decimal.Decimal
	Status           QuotationStatus
	ConvertedOrderID *uuid.UUID
	Notes            *string
	CreatedByUserID  *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Lines            []QuotationLine
}

// QuotationLine represents an item in a quotation
type QuotationLine struct {
	ID              uuid.UUID
	QuotationID     uuid.UUID
	LineNumber      int
	ItemID          uuid.UUID
	DescriptionEn   string
	DescriptionAr   string
	UOM             string
	Quantity        decimal.Decimal
	UnitPrice       decimal.Decimal
	DiscountPercent decimal.Decimal
	DiscountAmount  decimal.Decimal
	TaxableAmount   decimal.Decimal
	TaxCode         string
	TaxRatePercent  decimal.Decimal
	TaxAmount       decimal.Decimal
	TotalAmount     decimal.Decimal
}

// SalesOrder represents a confirmed customer order
type SalesOrder struct {
	ID                   uuid.UUID
	LegalEntityID        uuid.UUID
	OrderNumber          string
	CustomerID           uuid.UUID
	QuotationID          *uuid.UUID
	CustomerPONumber     *string
	OrderDate            time.Time
	ExpectedDeliveryDate *time.Time
	PaymentTermsID       *uuid.UUID
	CurrencyCode         string
	ExchangeRate         decimal.Decimal
	Subtotal             decimal.Decimal
	DiscountAmount       decimal.Decimal
	TaxableAmount        decimal.Decimal
	TaxAmount            decimal.Decimal
	TotalAmount          decimal.Decimal
	Status               OrderStatus
	Notes                *string
	CreatedByUserID      *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Lines                []SalesOrderLine
}

// SalesOrderLine represents an item in an order
type SalesOrderLine struct {
	ID                uuid.UUID
	OrderID           uuid.UUID
	LineNumber        int
	ItemID            uuid.UUID
	DescriptionEn     string
	DescriptionAr     string
	UOM               string
	Quantity          decimal.Decimal
	DeliveredQuantity decimal.Decimal
	InvoicedQuantity  decimal.Decimal
	UnitPrice         decimal.Decimal
	DiscountPercent   decimal.Decimal
	DiscountAmount    decimal.Decimal
	TaxableAmount     decimal.Decimal
	TaxCode           string
	TaxRatePercent    decimal.Decimal
	TaxAmount         decimal.Decimal
	TotalAmount       decimal.Decimal
}

// Delivery represents a delivery voucher / shipment note
type Delivery struct {
	ID                uuid.UUID
	LegalEntityID     uuid.UUID
	DeliveryNumber    string
	OrderID           *uuid.UUID
	CustomerID        uuid.UUID
	WarehouseID       uuid.UUID
	DeliveryDate      time.Time
	Status            DeliveryStatus
	ShippingReference *string
	Notes             *string
	CreatedByUserID   *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Lines             []DeliveryLine
}

// DeliveryLine represents a dispatched item
type DeliveryLine struct {
	ID                  uuid.UUID
	DeliveryID          uuid.UUID
	LineNumber          int
	OrderLineID         *uuid.UUID
	ItemID              uuid.UUID
	UOM                 string
	Quantity            decimal.Decimal
	WarehouseID         uuid.UUID
	BinID               *uuid.UUID
	InventoryMovementID *uuid.UUID
	UnitCost            decimal.Decimal
	TotalCost           decimal.Decimal
}

// SalesInvoice represents a ZATCA-compliant Tax Invoice
type SalesInvoice struct {
	ID                  uuid.UUID
	LegalEntityID       uuid.UUID
	InvoiceNumber       string
	InvoiceType         InvoiceType
	OrderID             *uuid.UUID
	DeliveryID          *uuid.UUID
	CustomerID          uuid.UUID
	CustomerNameAr      string
	CustomerNameEn      string
	CustomerTaxID       *string
	CustomerCRNumber    *string
	CustomerAddressJSON *string
	IssueDate           time.Time
	DueDate             time.Time
	PaymentTermsID      *uuid.UUID
	CurrencyCode        string
	ExchangeRate        decimal.Decimal
	Subtotal            decimal.Decimal
	DiscountAmount      decimal.Decimal
	TaxableAmount       decimal.Decimal
	TaxAmount           decimal.Decimal
	TotalAmount         decimal.Decimal
	Status              InvoiceStatus
	JournalEntryID      *uuid.UUID
	OpenItemID          *uuid.UUID
	QRCodeData          *string
	Notes               *string
	CreatedByUserID     *uuid.UUID
	PostedAt            *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Lines               []SalesInvoiceLine
}

// SalesInvoiceLine represents an immutable invoice item snapshot per Section 40.14
type SalesInvoiceLine struct {
	ID              uuid.UUID
	InvoiceID       uuid.UUID
	LineNumber      int
	ItemID          uuid.UUID
	Description     string
	UOM             string
	Quantity        decimal.Decimal
	UnitPrice       decimal.Decimal
	GrossAmount     decimal.Decimal
	DiscountAmount  decimal.Decimal
	TaxableAmount   decimal.Decimal
	TaxCode         string
	TaxRatePercent  decimal.Decimal
	TaxAmount       decimal.Decimal
	NetAmount       decimal.Decimal
	ExemptionReason *string
}

// CreditNote represents a sales credit note
type CreditNote struct {
	ID                uuid.UUID
	LegalEntityID     uuid.UUID
	CreditNoteNumber  string
	OriginalInvoiceID uuid.UUID
	CustomerID        uuid.UUID
	IssueDate         time.Time
	ReasonCode        CreditNoteReason
	ReasonDetails     *string
	CurrencyCode      string
	ExchangeRate      decimal.Decimal
	Subtotal          decimal.Decimal
	TaxAmount         decimal.Decimal
	TotalAmount       decimal.Decimal
	Status            CreditNoteStatus
	JournalEntryID    *uuid.UUID
	OpenItemID        *uuid.UUID
	CreatedByUserID   *uuid.UUID
	PostedAt          *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Lines             []CreditNoteLine
}

// CreditNoteLine represents a credit note line
type CreditNoteLine struct {
	ID             uuid.UUID
	CreditNoteID   uuid.UUID
	LineNumber     int
	ItemID         uuid.UUID
	Description    string
	Quantity       decimal.Decimal
	UnitPrice      decimal.Decimal
	TaxableAmount  decimal.Decimal
	TaxCode        string
	TaxRatePercent decimal.Decimal
	TaxAmount      decimal.Decimal
	TotalAmount    decimal.Decimal
}

// ValidateStandardInvoice ensures mandatory ZATCA Phase 2 fields for B2B invoices
func ValidateStandardInvoice(taxID *string) error {
	if taxID == nil || !vatRegex.MatchString(*taxID) {
		return ErrInvalidTaxIDForStandardInvoice
	}
	return nil
}

// CalculateLineAmounts computes line discounts, taxable base, and VAT
func CalculateLineAmounts(qty, price, discountPercent, taxRate decimal.Decimal) (gross, discount, taxable, tax, total decimal.Decimal) {
	gross = qty.Mul(price).Round(6)
	if discountPercent.GreaterThan(decimal.Zero) {
		discount = gross.Mul(discountPercent).Div(decimal.NewFromInt(100)).Round(6)
	} else {
		discount = decimal.Zero
	}
	taxable = gross.Sub(discount).Round(6)
	if taxRate.GreaterThan(decimal.Zero) {
		tax = taxable.Mul(taxRate).Div(decimal.NewFromInt(100)).Round(6)
	} else {
		tax = decimal.Zero
	}
	total = taxable.Add(tax).Round(6)
	return
}
