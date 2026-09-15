package services

import (
	"context"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"

	"github.com/google/uuid"
)

type accountingService struct {
	repo ports.AccountingRepository
}

func NewAccountingService(repo ports.AccountingRepository) ports.AccountingService {
	return &accountingService{repo: repo}
}

func (s *accountingService) ListChartOfAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ChartOfAccounts, error) {
	return s.repo.ListChartOfAccounts(ctx, legalEntityID)
}

func (s *accountingService) GetChartOfAccountsByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error) {
	return s.repo.GetChartOfAccountsByID(ctx, id)
}

func (s *accountingService) CreateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error) {
	if acc.ParentID != nil && *acc.ParentID != uuid.Nil {
		parent, err := s.repo.GetAccountByID(ctx, *acc.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.LegalEntityID != acc.LegalEntityID {
			return nil, domain.ErrAccountNotFound
		}
	}
	return s.repo.CreateAccount(ctx, acc)
}

func (s *accountingService) UpdateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error) {
	return s.repo.UpdateAccount(ctx, acc)
}

func (s *accountingService) GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	return s.repo.GetAccountByID(ctx, id)
}

func (s *accountingService) GetAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Account, error) {
	return s.repo.GetAccountByCode(ctx, legalEntityID, code)
}

func (s *accountingService) ListAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Account, error) {
	return s.repo.ListAccounts(ctx, legalEntityID)
}

func (s *accountingService) ListFiscalYears(ctx context.Context, legalEntityID uuid.UUID) ([]domain.FiscalYear, error) {
	return s.repo.ListFiscalYears(ctx, legalEntityID)
}

func (s *accountingService) GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error) {
	return s.repo.GetFiscalYearByID(ctx, id)
}

func (s *accountingService) ListFiscalPeriods(ctx context.Context, fiscalYearID uuid.UUID) ([]domain.FiscalPeriod, error) {
	return s.repo.ListFiscalPeriods(ctx, fiscalYearID)
}

func (s *accountingService) GetFiscalPeriodByID(ctx context.Context, id uuid.UUID) (*domain.FiscalPeriod, error) {
	return s.repo.GetFiscalPeriodByID(ctx, id)
}

func (s *accountingService) GetFiscalPeriodByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.FiscalPeriod, error) {
	return s.repo.GetFiscalPeriodByDate(ctx, legalEntityID, date)
}

func (s *accountingService) UpdateFiscalPeriodStatus(ctx context.Context, id uuid.UUID, status domain.FiscalPeriodStatus) (*domain.FiscalPeriod, error) {
	return s.repo.UpdateFiscalPeriodStatus(ctx, id, status)
}

func (s *accountingService) ListJournals(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Journal, error) {
	return s.repo.ListJournals(ctx, legalEntityID)
}

func (s *accountingService) GetJournalByID(ctx context.Context, id uuid.UUID) (*domain.Journal, error) {
	return s.repo.GetJournalByID(ctx, id)
}

func (s *accountingService) CreateJournal(ctx context.Context, j domain.Journal) (*domain.Journal, error) {
	return s.repo.CreateJournal(ctx, j)
}
