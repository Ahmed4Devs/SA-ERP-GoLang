package ports

import (
	"context"
	"time"

	"erp/internal/accounting/domain"
	"github.com/google/uuid"
)

type AccountingService interface {
	ListChartOfAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ChartOfAccounts, error)
	GetChartOfAccountsByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error)

	CreateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error)
	UpdateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Account, error)
	ListAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Account, error)

	ListFiscalYears(ctx context.Context, legalEntityID uuid.UUID) ([]domain.FiscalYear, error)
	GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error)
	ListFiscalPeriods(ctx context.Context, fiscalYearID uuid.UUID) ([]domain.FiscalPeriod, error)
	GetFiscalPeriodByID(ctx context.Context, id uuid.UUID) (*domain.FiscalPeriod, error)
	GetFiscalPeriodByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.FiscalPeriod, error)
	UpdateFiscalPeriodStatus(ctx context.Context, id uuid.UUID, status domain.FiscalPeriodStatus) (*domain.FiscalPeriod, error)

	ListJournals(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Journal, error)
	GetJournalByID(ctx context.Context, id uuid.UUID) (*domain.Journal, error)
	CreateJournal(ctx context.Context, j domain.Journal) (*domain.Journal, error)
}

type LedgerService interface {
	CreateJournalEntry(ctx context.Context, entry domain.JournalEntry, lines []domain.JournalLine) (*domain.JournalEntry, error)
	GetJournalEntryByID(ctx context.Context, id uuid.UUID) (*domain.JournalEntry, error)
	GetJournalEntryByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.JournalEntry, error)
	ListJournalEntries(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.JournalEntry, int64, error)
	PostJournalEntry(ctx context.Context, id uuid.UUID, postedBy uuid.UUID) (*domain.JournalEntry, error)
	ReverseJournalEntry(ctx context.Context, id uuid.UUID, reason string, reversedBy uuid.UUID) (*domain.JournalEntry, error)
	GetTrialBalance(ctx context.Context, legalEntityID uuid.UUID, asOfDate *time.Time) (*domain.TrialBalanceReport, error)
}

type AccountingRepository interface {
	ListChartOfAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ChartOfAccounts, error)
	GetChartOfAccountsByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error)

	CreateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error)
	UpdateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error)
	GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Account, error)
	ListAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Account, error)

	ListFiscalYears(ctx context.Context, legalEntityID uuid.UUID) ([]domain.FiscalYear, error)
	GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error)
	GetFiscalYearByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.FiscalYear, error)
	ListFiscalPeriods(ctx context.Context, fiscalYearID uuid.UUID) ([]domain.FiscalPeriod, error)
	GetFiscalPeriodByID(ctx context.Context, id uuid.UUID) (*domain.FiscalPeriod, error)
	GetFiscalPeriodByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.FiscalPeriod, error)
	UpdateFiscalPeriodStatus(ctx context.Context, id uuid.UUID, status domain.FiscalPeriodStatus) (*domain.FiscalPeriod, error)

	ListJournals(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Journal, error)
	GetJournalByID(ctx context.Context, id uuid.UUID) (*domain.Journal, error)
	GetJournalByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Journal, error)
	CreateJournal(ctx context.Context, j domain.Journal) (*domain.Journal, error)

	CreateJournalEntry(ctx context.Context, entry domain.JournalEntry, lines []domain.JournalLine) (*domain.JournalEntry, error)
	GetJournalEntryByID(ctx context.Context, id uuid.UUID) (*domain.JournalEntry, error)
	GetJournalEntryByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.JournalEntry, error)
	ListJournalEntries(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.JournalEntry, int64, error)
	PostJournalEntry(ctx context.Context, id uuid.UUID, postedBy uuid.UUID) (*domain.JournalEntry, error)
	ReverseJournalEntry(ctx context.Context, originalID uuid.UUID, reversalEntry domain.JournalEntry, reversalLines []domain.JournalLine, reversedBy uuid.UUID) (*domain.JournalEntry, error)
	GetTrialBalance(ctx context.Context, legalEntityID uuid.UUID, asOfDate *time.Time) (*domain.TrialBalanceReport, error)
}
