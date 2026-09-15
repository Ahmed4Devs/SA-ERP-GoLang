package ports

import (
	"context"
	"time"

	"erp/internal/tax/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TaxRepository interface {
	// Categories & Reasons
	ListTaxCategories(ctx context.Context) ([]domain.TaxCategory, error)
	GetTaxCategoryByCode(ctx context.Context, code string) (*domain.TaxCategory, error)
	ListExemptionReasons(ctx context.Context) ([]domain.ExemptionReason, error)
	ListZeroRateReasons(ctx context.Context) ([]domain.ZeroRateReason, error)

	// Tax Codes & Rates
	ListTaxCodesByEntity(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxCode, error)
	GetTaxCodeByID(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error)
	GetTaxCodeByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.TaxCode, error)
	CreateTaxCode(ctx context.Context, code *domain.TaxCode) (*domain.TaxCode, error)
	UpdateTaxCode(ctx context.Context, code *domain.TaxCode) (*domain.TaxCode, error)
	CreateTaxRateVersion(ctx context.Context, rate *domain.TaxRateVersion) (*domain.TaxRateVersion, error)
	GetActiveRateVersion(ctx context.Context, taxCodeID uuid.UUID, targetDate time.Time) (*domain.TaxRateVersion, error)
	ListRateVersions(ctx context.Context, taxCodeID uuid.UUID) ([]domain.TaxRateVersion, error)

	// Tax Reporting Periods
	ListTaxPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxPeriod, error)
	GetTaxPeriodByID(ctx context.Context, id uuid.UUID) (*domain.TaxPeriod, error)
	GetOpenPeriodForDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.TaxPeriod, error)
	CreateTaxPeriod(ctx context.Context, period *domain.TaxPeriod) (*domain.TaxPeriod, error)
	UpdateTaxPeriodStatus(ctx context.Context, id uuid.UUID, status domain.TaxPeriodStatus) (*domain.TaxPeriod, error)

	// Tax Transactions
	CreateTaxTransaction(ctx context.Context, tx *domain.TaxTransaction) (*domain.TaxTransaction, error)
	GetTaxTransactionByID(ctx context.Context, id uuid.UUID) (*domain.TaxTransaction, error)
	GetTaxTransactionBySource(ctx context.Context, legalEntityID uuid.UUID, sourceModule, sourceType string, sourceID uuid.UUID) (*domain.TaxTransaction, error)
	ListTaxTransactionsByEntity(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.TaxTransaction, error)
	ListTaxTransactionsByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.TaxTransaction, error)

	// Tax Returns (ZATCA VAT Return Form)
	CreateTaxReturn(ctx context.Context, ret *domain.TaxReturn) (*domain.TaxReturn, error)
	GetTaxReturnByID(ctx context.Context, id uuid.UUID) (*domain.TaxReturn, error)
	GetTaxReturnByPeriod(ctx context.Context, legalEntityID, periodID uuid.UUID) (*domain.TaxReturn, error)
	UpdateTaxReturn(ctx context.Context, ret *domain.TaxReturn) (*domain.TaxReturn, error)
	ListTaxReturnsByEntity(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxReturn, error)
	AggregatePeriodSummary(ctx context.Context, periodID uuid.UUID) ([]TaxPeriodAggregationRow, error)
}

type TaxPeriodAggregationRow struct {
	TransactionType     string
	TaxCategoryCode     string
	TaxRatePercent      decimal.Decimal
	ExemptionReasonCode *string
	ZeroRateReasonCode  *string
	TotalTaxable        decimal.Decimal
	TotalTax            decimal.Decimal
	TotalGross          decimal.Decimal
}
