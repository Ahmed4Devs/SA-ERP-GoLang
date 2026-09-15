package ports

import (
	"context"
	"time"

	"erp/internal/tax/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PostTaxTransactionRequest struct {
	LegalEntityID      uuid.UUID                  `json:"legal_entity_id"`
	SourceModule       string                     `json:"source_module"`
	SourceType         string                     `json:"source_type"`
	SourceID           uuid.UUID                  `json:"source_id"`
	TransactionNumber  string                     `json:"transaction_number"`
	TransactionDate    time.Time                  `json:"transaction_date"`
	TransactionType    string                     `json:"transaction_type"` // OUTPUT, INPUT
	PartnerID          *uuid.UUID                 `json:"partner_id,omitempty"`
	PartnerName        *string                    `json:"partner_name,omitempty"`
	PartnerTaxNumber   *string                    `json:"partner_tax_number,omitempty"`
	JournalEntryID     *uuid.UUID                 `json:"journal_entry_id,omitempty"`
	CreatedBy          *uuid.UUID                 `json:"created_by,omitempty"`
	Lines              []domain.TaxTransactionLine `json:"lines"`
}

type CalculateTaxRequest struct {
	LegalEntityID   uuid.UUID                        `json:"legal_entity_id"`
	TransactionDate time.Time                        `json:"transaction_date"`
	Lines           []domain.TaxLineCalculationInput `json:"lines"`
}

type TaxService interface {
	// Reference & Configuration
	ListCategories(ctx context.Context) ([]domain.TaxCategory, error)
	ListExemptionReasons(ctx context.Context) ([]domain.ExemptionReason, error)
	ListZeroRateReasons(ctx context.Context) ([]domain.ZeroRateReason, error)
	ListTaxCodes(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxCode, error)
	GetTaxCode(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error)
	CreateTaxCode(ctx context.Context, req domain.TaxCode, initialRate decimal.Decimal, effectiveFrom time.Time, createdBy *uuid.UUID) (*domain.TaxCode, error)
	CreateRateVersion(ctx context.Context, rate domain.TaxRateVersion) (*domain.TaxRateVersion, error)
	GetActiveRate(ctx context.Context, taxCodeID uuid.UUID, targetDate time.Time) (*domain.TaxRateVersion, error)

	// Financial Calculation Engine
	CalculateTax(ctx context.Context, req CalculateTaxRequest) (*domain.TaxCalculationResult, error)

	// Tax Transaction Engine
	PostTaxTransaction(ctx context.Context, req PostTaxTransactionRequest) (*domain.TaxTransaction, error)
	GetTaxTransaction(ctx context.Context, id uuid.UUID) (*domain.TaxTransaction, error)
	ListTaxTransactions(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.TaxTransaction, error)

	// Periods & VAT Returns (ZATCA)
	ListPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxPeriod, error)
	CreatePeriod(ctx context.Context, period domain.TaxPeriod) (*domain.TaxPeriod, error)
	ClosePeriod(ctx context.Context, periodID uuid.UUID) (*domain.TaxPeriod, error)
	GenerateVATReturnDraft(ctx context.Context, legalEntityID, periodID uuid.UUID) (*domain.TaxReturn, error)
	SubmitVATReturn(ctx context.Context, legalEntityID, returnID uuid.UUID, actorID uuid.UUID) (*domain.TaxReturn, error)
	ListVATReturns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxReturn, error)
}
