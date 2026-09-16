package ports

import (
	"context"
	"time"

	"erp/internal/sales/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateQuotationLineCommand struct {
	ItemID          uuid.UUID       `json:"item_id"`
	DescriptionEn   string          `json:"description_en"`
	DescriptionAr   string          `json:"description_ar"`
	UOM             string          `json:"uom"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	TaxCode         string          `json:"tax_code"`
	TaxRatePercent  decimal.Decimal `json:"tax_rate_percent"`
}

type CreateQuotationCommand struct {
	LegalEntityID   uuid.UUID                    `json:"legal_entity_id"`
	CustomerID      uuid.UUID                    `json:"customer_id"`
	CustomerNameEn  string                       `json:"customer_name_en"`
	CustomerNameAr  string                       `json:"customer_name_ar"`
	QuoteDate       time.Time                    `json:"quote_date"`
	ValidUntil      time.Time                    `json:"valid_until"`
	CurrencyCode    string                       `json:"currency_code"`
	ExchangeRate    decimal.Decimal              `json:"exchange_rate"`
	Notes           *string                      `json:"notes"`
	CreatedByUserID *uuid.UUID                   `json:"created_by_user_id"`
	Lines           []CreateQuotationLineCommand `json:"lines"`
}

type CreateOrderLineCommand struct {
	ItemID          uuid.UUID       `json:"item_id"`
	DescriptionEn   string          `json:"description_en"`
	DescriptionAr   string          `json:"description_ar"`
	UOM             string          `json:"uom"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	TaxCode         string          `json:"tax_code"`
	TaxRatePercent  decimal.Decimal `json:"tax_rate_percent"`
}

type CreateOrderCommand struct {
	LegalEntityID        uuid.UUID                `json:"legal_entity_id"`
	CustomerID           uuid.UUID                `json:"customer_id"`
	QuotationID          *uuid.UUID               `json:"quotation_id"`
	CustomerPONumber     *string                  `json:"customer_po_number"`
	OrderDate            time.Time                `json:"order_date"`
	ExpectedDeliveryDate *time.Time               `json:"expected_delivery_date"`
	PaymentTermsID       *uuid.UUID               `json:"payment_terms_id"`
	CurrencyCode         string                   `json:"currency_code"`
	ExchangeRate         decimal.Decimal          `json:"exchange_rate"`
	Notes                *string                  `json:"notes"`
	CreatedByUserID      *uuid.UUID               `json:"created_by_user_id"`
	Lines                []CreateOrderLineCommand `json:"lines"`
}

type CreateDeliveryLineCommand struct {
	OrderLineID *uuid.UUID      `json:"order_line_id"`
	ItemID      uuid.UUID       `json:"item_id"`
	UOM         string          `json:"uom"`
	Quantity    decimal.Decimal `json:"quantity"`
	WarehouseID uuid.UUID       `json:"warehouse_id"`
	BinID       *uuid.UUID      `json:"bin_id"`
}

type CreateDeliveryCommand struct {
	LegalEntityID     uuid.UUID                   `json:"legal_entity_id"`
	OrderID           *uuid.UUID                  `json:"order_id"`
	CustomerID        uuid.UUID                   `json:"customer_id"`
	WarehouseID       uuid.UUID                   `json:"warehouse_id"`
	DeliveryDate      time.Time                   `json:"delivery_date"`
	ShippingReference *string                     `json:"shipping_reference"`
	Notes             *string                     `json:"notes"`
	CreatedByUserID   *uuid.UUID                  `json:"created_by_user_id"`
	Lines             []CreateDeliveryLineCommand `json:"lines"`
}

type CreateInvoiceLineCommand struct {
	ItemID          uuid.UUID       `json:"item_id"`
	Description     string          `json:"description"`
	UOM             string          `json:"uom"`
	Quantity        decimal.Decimal `json:"quantity"`
	UnitPrice       decimal.Decimal `json:"unit_price"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	TaxCode         string          `json:"tax_code"`
	TaxRatePercent  decimal.Decimal `json:"tax_rate_percent"`
	ExemptionReason *string         `json:"exemption_reason"`
}

type CreateInvoiceCommand struct {
	LegalEntityID       uuid.UUID                  `json:"legal_entity_id"`
	InvoiceType         domain.InvoiceType         `json:"invoice_type"`
	OrderID             *uuid.UUID                 `json:"order_id"`
	DeliveryID          *uuid.UUID                 `json:"delivery_id"`
	CustomerID          uuid.UUID                  `json:"customer_id"`
	CustomerNameAr      string                     `json:"customer_name_ar"`
	CustomerNameEn      string                     `json:"customer_name_en"`
	CustomerTaxID       *string                    `json:"customer_tax_id"`
	CustomerCRNumber    *string                    `json:"customer_cr_number"`
	CustomerAddressJSON *string                    `json:"customer_address_json"`
	IssueDate           time.Time                  `json:"issue_date"`
	DueDate             time.Time                  `json:"due_date"`
	PaymentTermsID      *uuid.UUID                 `json:"payment_terms_id"`
	CurrencyCode        string                     `json:"currency_code"`
	ExchangeRate        decimal.Decimal            `json:"exchange_rate"`
	Notes               *string                    `json:"notes"`
	CreatedByUserID     *uuid.UUID                 `json:"created_by_user_id"`
	Lines               []CreateInvoiceLineCommand `json:"lines"`
}

type CreateCreditNoteLineCommand struct {
	ItemID         uuid.UUID       `json:"item_id"`
	Description    string          `json:"description"`
	Quantity       decimal.Decimal `json:"quantity"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	TaxCode        string          `json:"tax_code"`
	TaxRatePercent decimal.Decimal `json:"tax_rate_percent"`
}

type CreateCreditNoteCommand struct {
	LegalEntityID     uuid.UUID                     `json:"legal_entity_id"`
	OriginalInvoiceID uuid.UUID                     `json:"original_invoice_id"`
	CustomerID        uuid.UUID                     `json:"customer_id"`
	IssueDate         time.Time                     `json:"issue_date"`
	ReasonCode        domain.CreditNoteReason       `json:"reason_code"`
	ReasonDetails     *string                       `json:"reason_details"`
	CurrencyCode      string                        `json:"currency_code"`
	CreatedByUserID   *uuid.UUID                    `json:"created_by_user_id"`
	Lines             []CreateCreditNoteLineCommand `json:"lines"`
}

type SalesService interface {
	// Quotations
	CreateQuotation(ctx context.Context, cmd CreateQuotationCommand) (*domain.Quotation, error)
	GetQuotation(ctx context.Context, id uuid.UUID) (*domain.Quotation, error)
	ListQuotations(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.QuotationStatus, limit, offset int) ([]*domain.Quotation, error)
	ConvertQuotationToOrder(ctx context.Context, quoteID uuid.UUID, userID *uuid.UUID) (*domain.SalesOrder, error)

	// Orders
	CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.SalesOrder, error)
	ConfirmOrder(ctx context.Context, orderID uuid.UUID) (*domain.SalesOrder, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.SalesOrder, error)
	ListOrders(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.OrderStatus, limit, offset int) ([]*domain.SalesOrder, error)

	// Deliveries
	CreateDelivery(ctx context.Context, cmd CreateDeliveryCommand) (*domain.Delivery, error)
	DispatchDelivery(ctx context.Context, deliveryID uuid.UUID) (*domain.Delivery, error)
	GetDelivery(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	ListDeliveries(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, orderID *uuid.UUID, status *domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, error)

	// Invoices
	CreateInvoice(ctx context.Context, cmd CreateInvoiceCommand) (*domain.SalesInvoice, error)
	CreateInvoiceFromOrder(ctx context.Context, orderID uuid.UUID, invoiceType domain.InvoiceType, userID *uuid.UUID) (*domain.SalesInvoice, error)
	PostInvoice(ctx context.Context, invoiceID uuid.UUID) (*domain.SalesInvoice, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*domain.SalesInvoice, error)
	ListInvoices(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, invoiceType *domain.InvoiceType, status *domain.InvoiceStatus, limit, offset int) ([]*domain.SalesInvoice, error)

	// Credit Notes
	CreateCreditNote(ctx context.Context, cmd CreateCreditNoteCommand) (*domain.CreditNote, error)
	PostCreditNote(ctx context.Context, creditNoteID uuid.UUID) (*domain.CreditNote, error)
	GetCreditNote(ctx context.Context, id uuid.UUID) (*domain.CreditNote, error)
	ListCreditNotes(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, originalInvoiceID *uuid.UUID, status *domain.CreditNoteStatus, limit, offset int) ([]*domain.CreditNote, error)
}
