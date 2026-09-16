package http

import (
	"erp/internal/receivables/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateCustomerRequest struct {
	LegalEntityID    uuid.UUID           `json:"legal_entity_id"`
	CustomerType     domain.CustomerType `json:"customer_type"`
	NameEn           string              `json:"name_en"`
	NameAr           string              `json:"name_ar"`
	TaxID            *string             `json:"tax_id,omitempty"`
	CRNumber         *string             `json:"cr_number,omitempty"`
	NationalID       *string             `json:"national_id,omitempty"`
	BuildingNo       string              `json:"building_no,omitempty"`
	StreetName       string              `json:"street_name,omitempty"`
	District         string              `json:"district,omitempty"`
	City             string              `json:"city"`
	PostalCode       string              `json:"postal_code,omitempty"`
	AdditionalNo     string              `json:"additional_no,omitempty"`
	CountryCode      string              `json:"country_code,omitempty"`
	Email            *string             `json:"email,omitempty"`
	Phone            *string             `json:"phone,omitempty"`
	ContactPerson    *string             `json:"contact_person,omitempty"`
	PaymentTermsID   *uuid.UUID          `json:"payment_terms_id,omitempty"`
	CurrencyCode     string              `json:"currency_code,omitempty"`
	DefaultAccountID *uuid.UUID          `json:"default_account_id,omitempty"`
	CreditLimit      decimal.Decimal     `json:"credit_limit"`
	MaxOverdueDays   int32               `json:"max_overdue_days"`
	BlockAction      domain.BlockAction  `json:"block_action"`
}

type UpdateCustomerRequest struct {
	NameEn           string              `json:"name_en"`
	NameAr           string              `json:"name_ar"`
	TaxID            *string             `json:"tax_id,omitempty"`
	CRNumber         *string             `json:"cr_number,omitempty"`
	NationalID       *string             `json:"national_id,omitempty"`
	BuildingNo       string              `json:"building_no,omitempty"`
	StreetName       string              `json:"street_name,omitempty"`
	District         string              `json:"district,omitempty"`
	City             string              `json:"city"`
	PostalCode       string              `json:"postal_code,omitempty"`
	AdditionalNo     string              `json:"additional_no,omitempty"`
	CountryCode      string              `json:"country_code,omitempty"`
	Email            *string             `json:"email,omitempty"`
	Phone            *string             `json:"phone,omitempty"`
	ContactPerson    *string             `json:"contact_person,omitempty"`
	PaymentTermsID   *uuid.UUID          `json:"payment_terms_id,omitempty"`
	CurrencyCode     string              `json:"currency_code,omitempty"`
	DefaultAccountID *uuid.UUID          `json:"default_account_id,omitempty"`
}

type SetCustomerStatusRequest struct {
	Status domain.CustomerStatus `json:"status"`
}

type SetCreditHoldRequest struct {
	IsHold bool    `json:"is_hold"`
	Reason *string `json:"reason,omitempty"`
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

type AllocationItemRequest struct {
	OpenItemID      uuid.UUID       `json:"open_item_id"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
}

type CreateReceiptRequest struct {
	LegalEntityID    uuid.UUID               `json:"legal_entity_id"`
	CustomerID       uuid.UUID               `json:"customer_id"`
	ReceiptDate      string                  `json:"receipt_date"`
	PaymentMethod    domain.PaymentMethod    `json:"payment_method"`
	DepositAccountID uuid.UUID               `json:"deposit_account_id"`
	Amount           decimal.Decimal         `json:"amount"`
	CurrencyCode     string                  `json:"currency_code"`
	ReferenceNumber  *string                 `json:"reference_number,omitempty"`
	Notes            *string                 `json:"notes,omitempty"`
	Allocations      []AllocationItemRequest `json:"allocations,omitempty"`
}

type AllocateReceiptRequest struct {
	Allocations []AllocationItemRequest `json:"allocations"`
}

type CreateOpenItemRequest struct {
	LegalEntityID  uuid.UUID           `json:"legal_entity_id"`
	CustomerID     uuid.UUID           `json:"customer_id"`
	DocumentType   domain.DocumentType `json:"document_type"`
	DocumentNumber string              `json:"document_number"`
	DocumentID     *uuid.UUID          `json:"document_id,omitempty"`
	DocumentDate   string              `json:"document_date"`
	DueDate        string              `json:"due_date"`
	CurrencyCode   string              `json:"currency_code"`
	OriginalAmount decimal.Decimal     `json:"original_amount"`
}
