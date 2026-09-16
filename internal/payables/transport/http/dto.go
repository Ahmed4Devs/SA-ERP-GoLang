package http

import (
	"erp/internal/payables/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateSupplierRequest struct {
	LegalEntityID             uuid.UUID           `json:"legal_entity_id"`
	SupplierNumber            string              `json:"supplier_number,omitempty"`
	SupplierType              domain.SupplierType `json:"supplier_type"`
	NameEn                    string              `json:"name_en"`
	NameAr                    string              `json:"name_ar"`
	TaxID                     *string             `json:"tax_id,omitempty"`
	CRNumber                  *string             `json:"cr_number,omitempty"`
	NationalID                *string             `json:"national_id,omitempty"`
	IBAN                      *string             `json:"iban,omitempty"`
	BankName                  *string             `json:"bank_name,omitempty"`
	SwiftCode                 *string             `json:"swift_code,omitempty"`
	BuildingNo                string              `json:"building_no,omitempty"`
	StreetName                string              `json:"street_name,omitempty"`
	District                  string              `json:"district,omitempty"`
	City                      string              `json:"city"`
	PostalCode                string              `json:"postal_code,omitempty"`
	AdditionalNo              string              `json:"additional_no,omitempty"`
	CountryCode               string              `json:"country_code,omitempty"`
	Email                     *string             `json:"email,omitempty"`
	Phone                     *string             `json:"phone,omitempty"`
	ContactPerson             *string             `json:"contact_person,omitempty"`
	Website                   *string             `json:"website,omitempty"`
	PaymentTermsID            *uuid.UUID          `json:"payment_terms_id,omitempty"`
	CurrencyCode              string              `json:"currency_code,omitempty"`
	DefaultExpenseAccountID   *uuid.UUID          `json:"default_expense_account_id,omitempty"`
	DefaultLiabilityAccountID *uuid.UUID          `json:"default_liability_account_id,omitempty"`
}

type UpdateSupplierRequest struct {
	NameEn                    string     `json:"name_en"`
	NameAr                    string     `json:"name_ar"`
	TaxID                     *string    `json:"tax_id,omitempty"`
	CRNumber                  *string    `json:"cr_number,omitempty"`
	NationalID                *string    `json:"national_id,omitempty"`
	IBAN                      *string    `json:"iban,omitempty"`
	BankName                  *string    `json:"bank_name,omitempty"`
	SwiftCode                 *string    `json:"swift_code,omitempty"`
	BuildingNo                string     `json:"building_no,omitempty"`
	StreetName                string     `json:"street_name,omitempty"`
	District                  string     `json:"district,omitempty"`
	City                      string     `json:"city"`
	PostalCode                string     `json:"postal_code,omitempty"`
	AdditionalNo              string     `json:"additional_no,omitempty"`
	CountryCode               string     `json:"country_code,omitempty"`
	Email                     *string    `json:"email,omitempty"`
	Phone                     *string    `json:"phone,omitempty"`
	ContactPerson             *string    `json:"contact_person,omitempty"`
	Website                   *string    `json:"website,omitempty"`
	PaymentTermsID            *uuid.UUID `json:"payment_terms_id,omitempty"`
	DefaultExpenseAccountID   *uuid.UUID `json:"default_expense_account_id,omitempty"`
	DefaultLiabilityAccountID *uuid.UUID `json:"default_liability_account_id,omitempty"`
}

type SetSupplierStatusRequest struct {
	Status domain.SupplierStatus `json:"status"`
}

type CreatePaymentTermRequest struct {
	LegalEntityID   uuid.UUID       `json:"legal_entity_id"`
	Code            string          `json:"code"`
	NameEn          string          `json:"name_en"`
	NameAr          string          `json:"name_ar"`
	DueDays         int32           `json:"due_days"`
	DiscountDays    int32           `json:"discount_days"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	IsActive        bool            `json:"is_active"`
}

type CreateSupplierInvoiceRequest struct {
	LegalEntityID         uuid.UUID                           `json:"legal_entity_id"`
	InvoiceNumber         string                              `json:"invoice_number,omitempty"`
	SupplierInvoiceNumber string                              `json:"supplier_invoice_number"`
	SupplierID            uuid.UUID                           `json:"supplier_id"`
	InvoiceDate           string                              `json:"invoice_date"`
	DueDate               string                              `json:"due_date,omitempty"`
	CurrencyCode          string                              `json:"currency_code,omitempty"`
	ExchangeRate          decimal.Decimal                     `json:"exchange_rate,omitempty"`
	Notes                 *string                             `json:"notes,omitempty"`
	Lines                 []CreateSupplierInvoiceLineRequest  `json:"lines"`
}

type CreateSupplierInvoiceLineRequest struct {
	ItemID           *uuid.UUID      `json:"item_id,omitempty"`
	ExpenseAccountID uuid.UUID       `json:"expense_account_id"`
	Description      string          `json:"description"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	DiscountPercent  decimal.Decimal `json:"discount_percent,omitempty"`
	TaxRate          decimal.Decimal `json:"tax_rate,omitempty"`
}

type CreateSupplierPaymentRequest struct {
	LegalEntityID         uuid.UUID                    `json:"legal_entity_id"`
	PaymentNumber         string                       `json:"payment_number,omitempty"`
	SupplierID            uuid.UUID                    `json:"supplier_id"`
	PaymentDate           string                       `json:"payment_date"`
	PaymentMethod         domain.PaymentMethod         `json:"payment_method"`
	DisbursementAccountID uuid.UUID                    `json:"disbursement_account_id"`
	Amount                decimal.Decimal              `json:"amount"`
	CurrencyCode          string                       `json:"currency_code,omitempty"`
	ReferenceNumber       *string                      `json:"reference_number,omitempty"`
	Notes                 *string                      `json:"notes,omitempty"`
	Allocations           []PaymentAllocationRequest   `json:"allocations,omitempty"`
}

type PaymentAllocationRequest struct {
	OpenItemID uuid.UUID       `json:"open_item_id"`
	Amount     decimal.Decimal `json:"amount"`
}

type CreateCreditNoteRequest struct {
	LegalEntityID            uuid.UUID                    `json:"legal_entity_id"`
	CreditNoteNumber         string                       `json:"credit_note_number,omitempty"`
	SupplierCreditNoteNumber *string                      `json:"supplier_credit_note_number,omitempty"`
	SupplierID               uuid.UUID                    `json:"supplier_id"`
	OriginalInvoiceID        *uuid.UUID                   `json:"original_invoice_id,omitempty"`
	CreditNoteDate           string                       `json:"credit_note_date"`
	CurrencyCode             string                       `json:"currency_code,omitempty"`
	Reason                   *string                      `json:"reason,omitempty"`
	Lines                    []CreateCreditNoteLineRequest `json:"lines"`
}

type CreateCreditNoteLineRequest struct {
	ItemID           *uuid.UUID      `json:"item_id,omitempty"`
	ExpenseAccountID uuid.UUID       `json:"expense_account_id"`
	Description      string          `json:"description"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	TaxRate          decimal.Decimal `json:"tax_rate,omitempty"`
}
