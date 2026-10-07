package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ledgerService struct {
	repo    ports.AccountingRepository
	posting ports.PostingEngine
}

func NewLedgerService(repo ports.AccountingRepository, posting ports.PostingEngine) ports.LedgerService {
	return &ledgerService{repo: repo, posting: posting}
}

func (s *ledgerService) CreateJournalEntry(ctx context.Context, entry domain.JournalEntry, lines []domain.JournalLine) (*domain.JournalEntry, error) {
	if entry.AccountingDate.IsZero() {
		entry.AccountingDate = time.Now().UTC()
	}

	// 1. Resolve Fiscal Period
	if entry.FiscalPeriodID == uuid.Nil {
		period, err := s.repo.GetFiscalPeriodByDate(ctx, entry.LegalEntityID, entry.AccountingDate)
		if err != nil {
			return nil, domain.ErrFiscalPeriodNotFound
		}
		entry.FiscalPeriodID = period.ID
	}

	period, err := s.repo.GetFiscalPeriodByID(ctx, entry.FiscalPeriodID)
	if err != nil {
		return nil, domain.ErrFiscalPeriodNotFound
	}
	if err := domain.ValidatePeriodCanPost(period.Status); err != nil {
		return nil, err
	}

	// 2. Validate rate and currency
	if entry.ExchangeRate.IsZero() {
		entry.ExchangeRate = decimal.NewFromInt(1)
	}
	if entry.CurrencyCode == "" {
		entry.CurrencyCode = "SAR"
	}

	// 3. Process and validate each line
	processedLines := make([]domain.JournalLine, 0, len(lines))
	for i, line := range lines {
		line.LineNumber = int16(i + 1)
		line.LegalEntityID = entry.LegalEntityID

		if line.ExchangeRate.IsZero() {
			line.ExchangeRate = entry.ExchangeRate
		}
		if line.CurrencyCode == "" {
			line.CurrencyCode = entry.CurrencyCode
		}

		// Calculate base amounts if not set
		if line.BaseDebit.IsZero() && line.TransactionDebit.GreaterThan(decimal.Zero) {
			line.BaseDebit = line.TransactionDebit.Mul(line.ExchangeRate).Round(4)
		}
		if line.BaseCredit.IsZero() && line.TransactionCredit.GreaterThan(decimal.Zero) {
			line.BaseCredit = line.TransactionCredit.Mul(line.ExchangeRate).Round(4)
		}

		// Verify account exists and can post
		acc, err := s.repo.GetAccountByID(ctx, line.AccountID)
		if err != nil {
			return nil, err
		}
		if err := domain.ValidateAccountCanPost(acc.AllowPosting, acc.IsActive); err != nil {
			return nil, fmt.Errorf("line %d: %w", line.LineNumber, err)
		}

		processedLines = append(processedLines, line)
	}

	// 4. Validate double-entry mathematical balance
	totalDebit, totalCredit, err := domain.ValidateJournalEntryBalance(processedLines)
	if err != nil {
		return nil, err
	}
	entry.TotalDebit = totalDebit
	entry.TotalCredit = totalCredit

	// 5. Default fields
	if entry.EntryNumber == "" {
		entry.EntryNumber = identifiers.NewDocumentNumber("JE", entry.AccountingDate)
	}
	if entry.Status == "" {
		entry.Status = domain.JournalEntryStatusDraft
	}
	if entry.SourceModule == "" {
		entry.SourceModule = "GL"
	}

	return s.repo.CreateJournalEntry(ctx, entry, processedLines)
}

func (s *ledgerService) GetJournalEntryByID(ctx context.Context, id uuid.UUID) (*domain.JournalEntry, error) {
	return s.repo.GetJournalEntryByID(ctx, id)
}

func (s *ledgerService) GetJournalEntryByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.JournalEntry, error) {
	return s.repo.GetJournalEntryByNumber(ctx, legalEntityID, number)
}

func (s *ledgerService) ListJournalEntries(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.JournalEntry, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListJournalEntries(ctx, legalEntityID, limit, offset)
}

func (s *ledgerService) PostJournalEntry(ctx context.Context, id uuid.UUID, postedBy uuid.UUID) (*domain.JournalEntry, error) {
	result, err := s.posting.PostDraft(ctx, id, postedBy)
	if err != nil {
		return nil, err
	}
	return s.repo.GetJournalEntryByID(ctx, result.JournalEntryID)
}

func (s *ledgerService) ReverseJournalEntry(ctx context.Context, id uuid.UUID, reason string, reversedBy uuid.UUID) (*domain.JournalEntry, error) {
	result, err := s.posting.Reverse(ctx, id, reason, reversedBy)
	if err != nil {
		return nil, err
	}
	return s.repo.GetJournalEntryByID(ctx, result.JournalEntryID)
}

func (s *ledgerService) GetTrialBalance(ctx context.Context, legalEntityID uuid.UUID, asOfDate *time.Time) (*domain.TrialBalanceReport, error) {
	return s.repo.GetTrialBalance(ctx, legalEntityID, asOfDate)
}
