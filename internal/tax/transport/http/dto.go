package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateTaxCodeRequest struct {
	LegalEntityID        uuid.UUID       `json:"legal_entity_id"`
	Code                 string          `json:"code"`
	NameEn               string          `json:"name_en"`
	NameAr               string          `json:"name_ar"`
	TaxCategoryCode      string          `json:"tax_category_code"`
	TaxType              string          `json:"tax_type"`
	SalesTaxAccountID    *uuid.UUID      `json:"sales_tax_account_id,omitempty"`
	PurchaseTaxAccountID *uuid.UUID      `json:"purchase_tax_account_id,omitempty"`
	IsRecoverable        bool            `json:"is_recoverable"`
	InitialRate          decimal.Decimal `json:"initial_rate"`
	EffectiveFrom        *time.Time      `json:"effective_from,omitempty"`
}

type AddRateVersionRequest struct {
	RatePercent   decimal.Decimal `json:"rate_percent"`
	EffectiveFrom time.Time       `json:"effective_from"`
	EffectiveTo   *time.Time      `json:"effective_to,omitempty"`
	Description   string          `json:"description"`
}

type CalculateTaxRequestDTO struct {
	LegalEntityID   uuid.UUID             `json:"legal_entity_id"`
	TransactionDate *time.Time            `json:"transaction_date,omitempty"`
	Lines           []CalculateTaxLineDTO `json:"lines"`
}

type CalculateTaxLineDTO struct {
	LineNumber          int             `json:"line_number"`
	TaxCodeID           uuid.UUID       `json:"tax_code_id"`
	Quantity            decimal.Decimal `json:"quantity"`
	UnitPrice           decimal.Decimal `json:"unit_price"`
	DiscountAmount      decimal.Decimal `json:"discount_amount"`
	ExemptionReasonCode *string         `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string         `json:"zero_rate_reason_code,omitempty"`
}

type PostTaxTransactionRequestDTO struct {
	LegalEntityID     uuid.UUID                   `json:"legal_entity_id"`
	SourceModule      string                      `json:"source_module"`
	SourceType        string                      `json:"source_type"`
	SourceID          uuid.UUID                   `json:"source_id"`
	TransactionNumber string                      `json:"transaction_number"`
	TransactionDate   *time.Time                  `json:"transaction_date,omitempty"`
	TransactionType   string                      `json:"transaction_type"` // OUTPUT, INPUT
	PartnerID         *uuid.UUID                  `json:"partner_id,omitempty"`
	PartnerName       *string                     `json:"partner_name,omitempty"`
	PartnerTaxNumber  *string                     `json:"partner_tax_number,omitempty"`
	JournalEntryID    *uuid.UUID                  `json:"journal_entry_id,omitempty"`
	Lines             []PostTaxTransactionLineDTO `json:"lines"`
}

type PostTaxTransactionLineDTO struct {
	TaxCodeID           uuid.UUID       `json:"tax_code_id"`
	TaxableAmount       decimal.Decimal `json:"taxable_amount"`
	ExemptionReasonCode *string         `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string         `json:"zero_rate_reason_code,omitempty"`
}

type CreateTaxPeriodRequest struct {
	LegalEntityID uuid.UUID `json:"legal_entity_id"`
	PeriodNumber  int       `json:"period_number"`
	FiscalYear    int       `json:"fiscal_year"`
	PeriodType    string    `json:"period_type"` // MONTHLY, QUARTERLY
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	DueDate       time.Time `json:"due_date"`
}

type GenerateVATReturnRequest struct {
	LegalEntityID uuid.UUID `json:"legal_entity_id"`
	TaxPeriodID   uuid.UUID `json:"tax_period_id"`
}
