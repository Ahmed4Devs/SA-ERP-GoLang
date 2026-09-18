package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TaxCategory struct {
	Code          string    `json:"code"`
	NameEn        string    `json:"name_en"`
	NameAr        string    `json:"name_ar"`
	DescriptionEn string    `json:"description_en"`
	DescriptionAr string    `json:"description_ar"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type ExemptionReason struct {
	Code          string    `json:"code"`
	CategoryCode  string    `json:"category_code"`
	NameEn        string    `json:"name_en"`
	NameAr        string    `json:"name_ar"`
	DescriptionEn string    `json:"description_en"`
	DescriptionAr string    `json:"description_ar"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type ZeroRateReason struct {
	Code          string    `json:"code"`
	CategoryCode  string    `json:"category_code"`
	NameEn        string    `json:"name_en"`
	NameAr        string    `json:"name_ar"`
	DescriptionEn string    `json:"description_en"`
	DescriptionAr string    `json:"description_ar"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type TaxCode struct {
	ID                   uuid.UUID        `json:"id"`
	LegalEntityID        uuid.UUID        `json:"legal_entity_id"`
	Code                 string           `json:"code"`
	NameEn               string           `json:"name_en"`
	NameAr               string           `json:"name_ar"`
	TaxCategoryCode      string           `json:"tax_category_code"`
	TaxType              string           `json:"tax_type"` // OUTPUT, INPUT, BOTH
	SalesTaxAccountID    *uuid.UUID       `json:"sales_tax_account_id,omitempty"`
	PurchaseTaxAccountID *uuid.UUID       `json:"purchase_tax_account_id,omitempty"`
	IsRecoverable        bool             `json:"is_recoverable"`
	IsActive             bool             `json:"is_active"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	Version              int64            `json:"version"`
	CategoryNameEn       string           `json:"category_name_en,omitempty"`
	CategoryNameAr       string           `json:"category_name_ar,omitempty"`
	CurrentRatePercent   *decimal.Decimal `json:"current_rate_percent,omitempty"`
}

type TaxRateVersion struct {
	ID            uuid.UUID       `json:"id"`
	TaxCodeID     uuid.UUID       `json:"tax_code_id"`
	RatePercent   decimal.Decimal `json:"rate_percent"`
	EffectiveFrom time.Time       `json:"effective_from"`
	EffectiveTo   *time.Time      `json:"effective_to,omitempty"`
	Description   string          `json:"description"`
	CreatedBy     *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type TaxPeriodStatus string

const (
	TaxPeriodStatusOpen   TaxPeriodStatus = "OPEN"
	TaxPeriodStatusFiled  TaxPeriodStatus = "FILED"
	TaxPeriodStatusClosed TaxPeriodStatus = "CLOSED"
)

type TaxPeriod struct {
	ID            uuid.UUID       `json:"id"`
	LegalEntityID uuid.UUID       `json:"legal_entity_id"`
	PeriodNumber  int             `json:"period_number"`
	FiscalYear    int             `json:"fiscal_year"`
	PeriodType    string          `json:"period_type"` // MONTHLY, QUARTERLY
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	DueDate       time.Time       `json:"due_date"`
	Status        TaxPeriodStatus `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type TaxTransaction struct {
	ID                 uuid.UUID            `json:"id"`
	LegalEntityID      uuid.UUID            `json:"legal_entity_id"`
	TaxPeriodID        *uuid.UUID           `json:"tax_period_id,omitempty"`
	SourceModule       string               `json:"source_module"`
	SourceType         string               `json:"source_type"`
	SourceID           uuid.UUID            `json:"source_id"`
	TransactionNumber  string               `json:"transaction_number"`
	TransactionDate    time.Time            `json:"transaction_date"`
	TransactionType    string               `json:"transaction_type"` // OUTPUT, INPUT
	PartnerID          *uuid.UUID           `json:"partner_id,omitempty"`
	PartnerName        *string              `json:"partner_name,omitempty"`
	PartnerTaxNumber   *string              `json:"partner_tax_number,omitempty"`
	TotalTaxableAmount decimal.Decimal      `json:"total_taxable_amount"`
	TotalTaxAmount     decimal.Decimal      `json:"total_tax_amount"`
	TotalGrossAmount   decimal.Decimal      `json:"total_gross_amount"`
	JournalEntryID     *uuid.UUID           `json:"journal_entry_id,omitempty"`
	Status             string               `json:"status"` // POSTED, CANCELLED, REVERSED
	CreatedBy          *uuid.UUID           `json:"created_by,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	Lines              []TaxTransactionLine `json:"lines,omitempty"`
}

type TaxTransactionLine struct {
	ID                  uuid.UUID       `json:"id"`
	TaxTransactionID    uuid.UUID       `json:"tax_transaction_id"`
	LineNumber          int             `json:"line_number"`
	TaxCodeID           uuid.UUID       `json:"tax_code_id"`
	TaxCode             string          `json:"tax_code,omitempty"`
	TaxNameAr           string          `json:"tax_name_ar,omitempty"`
	TaxNameEn           string          `json:"tax_name_en,omitempty"`
	TaxCategoryCode     string          `json:"tax_category_code"`
	TaxRatePercent      decimal.Decimal `json:"tax_rate_percent"`
	TaxableAmount       decimal.Decimal `json:"taxable_amount"`
	TaxAmount           decimal.Decimal `json:"tax_amount"`
	GrossAmount         decimal.Decimal `json:"gross_amount"`
	ExemptionReasonCode *string         `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string         `json:"zero_rate_reason_code,omitempty"`
	IsRecoverable       bool            `json:"is_recoverable"`
	CreatedAt           time.Time       `json:"created_at"`
}

// ZATCA 16-Box Official VAT Return Declaration Model
type DeclarationBoxItem struct {
	TaxableAmount decimal.Decimal `json:"taxable_amount"`
	Adjustment    decimal.Decimal `json:"adjustment"`
	VATAmount     decimal.Decimal `json:"vat_amount"`
}

type DeclarationBoxes struct {
	// Sales (Output VAT)
	Box1StandardRatedSales   DeclarationBoxItem `json:"box_1_standard_rated_sales"`
	Box2PrivateHealthcareEdu DeclarationBoxItem `json:"box_2_private_healthcare_edu"`
	Box3ZeroRatedDomestic    DeclarationBoxItem `json:"box_3_zero_rated_domestic"`
	Box4Exports              DeclarationBoxItem `json:"box_4_exports"`
	Box5ExemptSales          DeclarationBoxItem `json:"box_5_exempt_sales"`
	Box6TotalSales           DeclarationBoxItem `json:"box_6_total_sales"`

	// Purchases (Input VAT)
	Box7StandardRatedPurchases DeclarationBoxItem `json:"box_7_standard_rated_purchases"`
	Box8ImportsSubjectToVAT    DeclarationBoxItem `json:"box_8_imports_subject_to_vat"`
	Box9ImportsReverseCharge   DeclarationBoxItem `json:"box_9_imports_reverse_charge"`
	Box10ZeroRatedPurchases    DeclarationBoxItem `json:"box_10_zero_rated_purchases"`
	Box11ExemptPurchases       DeclarationBoxItem `json:"box_11_exempt_purchases"`
	Box12TotalPurchases        DeclarationBoxItem `json:"box_12_total_purchases"`

	// Net VAT Due
	Box13TotalVATDueCurrentPeriod decimal.Decimal `json:"box_13_total_vat_due_current_period"`
	Box14CorrectionsFromPrevious  decimal.Decimal `json:"box_14_corrections_from_previous"`
	Box15VATCreditCarriedForward  decimal.Decimal `json:"box_15_vat_credit_carried_forward"`
	Box16NetVATDue                decimal.Decimal `json:"box_16_net_vat_due"`
}

type TaxReturn struct {
	ID                    uuid.UUID        `json:"id"`
	LegalEntityID         uuid.UUID        `json:"legal_entity_id"`
	TaxPeriodID           uuid.UUID        `json:"tax_period_id"`
	ReturnNumber          string           `json:"return_number"`
	FilingDate            *time.Time       `json:"filing_date,omitempty"`
	Status                string           `json:"status"` // DRAFT, SUBMITTED, PAID, AMENDED
	TotalSalesTaxable     decimal.Decimal  `json:"total_sales_taxable"`
	TotalSalesVAT         decimal.Decimal  `json:"total_sales_vat"`
	TotalPurchasesTaxable decimal.Decimal  `json:"total_purchases_taxable"`
	TotalPurchasesVAT     decimal.Decimal  `json:"total_purchases_vat"`
	NetVATDue             decimal.Decimal  `json:"net_vat_due"`
	DeclarationBoxes      DeclarationBoxes `json:"declaration_boxes"`
	SubmittedAt           *time.Time       `json:"submitted_at,omitempty"`
	SubmittedBy           *uuid.UUID       `json:"submitted_by,omitempty"`
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
	Version               int64            `json:"version"`
}
