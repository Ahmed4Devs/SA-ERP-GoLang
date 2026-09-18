package http

import (
	"erp/internal/sales/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateQuotationRequest struct {
	LegalEntityID  uuid.UUID                    `json:"legal_entity_id"`
	CustomerID     uuid.UUID                    `json:"customer_id"`
	CustomerNameEn string                       `json:"customer_name_en"`
	CustomerNameAr string                       `json:"customer_name_ar"`
	QuoteDate      string                       `json:"quote_date"`
	ValidUntil     string                       `json:"valid_until"`
	CurrencyCode   string                       `json:"currency_code"`
	ExchangeRate   decimal.Decimal              `json:"exchange_rate"`
	Notes          *string                      `json:"notes"`
	Lines          []CreateQuotationLineRequest `json:"lines"`
}

type CreateQuotationLineRequest struct {
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

type CreateOrderRequest struct {
	LegalEntityID        uuid.UUID                `json:"legal_entity_id"`
	CustomerID           uuid.UUID                `json:"customer_id"`
	QuotationID          *uuid.UUID               `json:"quotation_id"`
	CustomerPONumber     *string                  `json:"customer_po_number"`
	OrderDate            string                   `json:"order_date"`
	ExpectedDeliveryDate *string                  `json:"expected_delivery_date"`
	PaymentTermsID       *uuid.UUID               `json:"payment_terms_id"`
	CurrencyCode         string                   `json:"currency_code"`
	ExchangeRate         decimal.Decimal          `json:"exchange_rate"`
	Notes                *string                  `json:"notes"`
	Lines                []CreateOrderLineRequest `json:"lines"`
}

type CreateOrderLineRequest struct {
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

type CreateDeliveryRequest struct {
	LegalEntityID     uuid.UUID                   `json:"legal_entity_id"`
	OrderID           *uuid.UUID                  `json:"order_id"`
	CustomerID        uuid.UUID                   `json:"customer_id"`
	WarehouseID       uuid.UUID                   `json:"warehouse_id"`
	DeliveryDate      string                      `json:"delivery_date"`
	ShippingReference *string                     `json:"shipping_reference"`
	Notes             *string                     `json:"notes"`
	Lines             []CreateDeliveryLineRequest `json:"lines"`
}

type CreateDeliveryLineRequest struct {
	OrderLineID *uuid.UUID      `json:"order_line_id"`
	ItemID      uuid.UUID       `json:"item_id"`
	UOM         string          `json:"uom"`
	Quantity    decimal.Decimal `json:"quantity"`
	WarehouseID uuid.UUID       `json:"warehouse_id"`
	BinID       *uuid.UUID      `json:"bin_id"`
}

type CreateInvoiceRequest struct {
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
	IssueDate           string                     `json:"issue_date"`
	DueDate             string                     `json:"due_date"`
	PaymentTermsID      *uuid.UUID                 `json:"payment_terms_id"`
	CurrencyCode        string                     `json:"currency_code"`
	ExchangeRate        decimal.Decimal            `json:"exchange_rate"`
	Notes               *string                    `json:"notes"`
	Lines               []CreateInvoiceLineRequest `json:"lines"`
}

type CreateInvoiceLineRequest struct {
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

type CreateCreditNoteRequest struct {
	LegalEntityID     uuid.UUID                     `json:"legal_entity_id"`
	OriginalInvoiceID uuid.UUID                     `json:"original_invoice_id"`
	CustomerID        uuid.UUID                     `json:"customer_id"`
	IssueDate         string                        `json:"issue_date"`
	ReasonCode        domain.CreditNoteReason       `json:"reason_code"`
	ReasonDetails     *string                       `json:"reason_details"`
	CurrencyCode      string                        `json:"currency_code"`
	Lines             []CreateCreditNoteLineRequest `json:"lines"`
}

type CreateCreditNoteLineRequest struct {
	ItemID         uuid.UUID       `json:"item_id"`
	Description    string          `json:"description"`
	Quantity       decimal.Decimal `json:"quantity"`
	UnitPrice      decimal.Decimal `json:"unit_price"`
	TaxCode        string          `json:"tax_code"`
	TaxRatePercent decimal.Decimal `json:"tax_rate_percent"`
}

type DocumentResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}
