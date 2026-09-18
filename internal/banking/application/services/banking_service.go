package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/banking/domain"
	"erp/internal/banking/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type bankingService struct {
	repo ports.BankingRepository
	db   *database.DB
}

func NewBankingService(repo ports.BankingRepository, db *database.DB) ports.BankingService {
	return &bankingService{
		repo: repo,
		db:   db,
	}
}

// ============================================================================
// BANK ACCOUNTS
// ============================================================================

func (s *bankingService) CreateBankAccount(ctx context.Context, a *domain.BankAccount) (*domain.BankAccount, error) {
	if a.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if a.AccountCode == "" {
		return nil, fmt.Errorf("account_code is required")
	}
	if a.BankNameEn == "" || a.BankNameAr == "" {
		return nil, fmt.Errorf("bank_name_en and bank_name_ar are required")
	}
	if a.AccountNumber == "" {
		return nil, fmt.Errorf("account_number is required")
	}
	if a.IBAN == "" {
		return nil, fmt.Errorf("iban is required")
	}
	// Basic IBAN check
	cleanIBAN := strings.ReplaceAll(a.IBAN, " ", "")
	if len(cleanIBAN) < 15 {
		return nil, domain.ErrInvalidIBAN
	}
	a.IBAN = cleanIBAN

	if a.AccountNumberMasked == "" {
		a.AccountNumberMasked = domain.MaskAccountNumber(a.AccountNumber)
	}
	if a.CurrencyCode == "" {
		a.CurrencyCode = "SAR"
	}
	if a.Status == "" {
		a.Status = domain.BankAccountActive
	}
	if a.CurrentBalance.IsZero() && !a.OpeningBalance.IsZero() {
		a.CurrentBalance = a.OpeningBalance
		a.AvailableBalance = a.OpeningBalance
	}

	return s.repo.CreateBankAccount(ctx, a)
}

func (s *bankingService) GetBankAccount(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.BankAccount, error) {
	return s.repo.GetBankAccountByID(ctx, legalEntityID, id)
}

func (s *bankingService) ListBankAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BankAccount, error) {
	return s.repo.ListBankAccounts(ctx, legalEntityID)
}

// ============================================================================
// STATEMENTS
// ============================================================================

func (s *bankingService) ImportBankStatement(ctx context.Context, stmt *domain.Statement, lines []*domain.StatementLine) (*domain.Statement, error) {
	if stmt.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if stmt.BankAccountID == uuid.Nil {
		return nil, fmt.Errorf("bank_account_id is required")
	}
	if stmt.StatementNumber == "" {
		stmt.StatementNumber = identifiers.NewDocumentNumber("STMT", time.Now())
	}
	if stmt.StatementDate.IsZero() {
		stmt.StatementDate = time.Now()
	}
	if stmt.CurrencyCode == "" {
		stmt.CurrencyCode = "SAR"
	}
	if stmt.FileFormat == "" {
		stmt.FileFormat = domain.StatementFormatCSV
	}
	if stmt.Status == "" {
		stmt.Status = domain.StatementImported
	}

	// Compute totals
	totalCredits := decimal.Zero
	totalDebits := decimal.Zero
	for _, l := range lines {
		if l.TransactionType == domain.TransactionCredit {
			totalCredits = totalCredits.Add(l.Amount)
		} else {
			totalDebits = totalDebits.Add(l.Amount)
		}
	}
	stmt.TotalCredits = totalCredits
	stmt.TotalDebits = totalDebits
	if stmt.ClosingBalance.IsZero() && !stmt.OpeningBalance.IsZero() {
		stmt.ClosingBalance = stmt.OpeningBalance.Add(totalCredits).Sub(totalDebits)
	}

	createdStmt, err := s.repo.CreateBankStatement(ctx, stmt)
	if err != nil {
		return nil, err
	}

	for i, l := range lines {
		l.StatementID = createdStmt.ID
		l.LineNumber = int32(i + 1)
		if l.CurrencyCode == "" {
			l.CurrencyCode = createdStmt.CurrencyCode
		}
		if l.ReconciledStatus == "" {
			l.ReconciledStatus = domain.ReconciledUnmatched
		}
		if _, err := s.repo.CreateStatementLine(ctx, l); err != nil {
			return nil, err
		}
	}

	return createdStmt, nil
}

func (s *bankingService) GetBankStatement(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Statement, error) {
	return s.repo.GetBankStatementByID(ctx, legalEntityID, id)
}

func (s *bankingService) ListBankStatements(ctx context.Context, bankAccountID uuid.UUID) ([]*domain.Statement, error) {
	return s.repo.ListBankStatements(ctx, bankAccountID)
}

func (s *bankingService) ListStatementLines(ctx context.Context, statementID uuid.UUID) ([]*domain.StatementLine, error) {
	return s.repo.ListStatementLines(ctx, statementID)
}

// ============================================================================
// RECONCILIATIONS
// ============================================================================

func (s *bankingService) CreateReconciliation(ctx context.Context, recon *domain.Reconciliation) (*domain.Reconciliation, error) {
	if recon.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if recon.BankAccountID == uuid.Nil {
		return nil, fmt.Errorf("bank_account_id is required")
	}
	if recon.ReconciliationNumber == "" {
		recon.ReconciliationNumber = identifiers.NewDocumentNumber("REC", time.Now())
	}
	if recon.AsOfDate.IsZero() {
		recon.AsOfDate = time.Now()
	}
	if recon.Status == "" {
		recon.Status = domain.ReconciliationInProgress
	}
	recon.UnreconciledDifference = recon.StatementEndingBalance.Sub(recon.GLBookBalance).Abs()

	return s.repo.CreateBankReconciliation(ctx, recon)
}

func (s *bankingService) GetReconciliation(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error) {
	return s.repo.GetBankReconciliationByID(ctx, legalEntityID, id)
}

func (s *bankingService) ListReconciliations(ctx context.Context, bankAccountID uuid.UUID) ([]*domain.Reconciliation, error) {
	return s.repo.ListBankReconciliations(ctx, bankAccountID)
}

func (s *bankingService) RunAutoMatch(ctx context.Context, legalEntityID, reconciliationID uuid.UUID, candidateEntries []domain.CandidateLedgerEntry) ([]domain.AutoMatchResult, error) {
	recon, err := s.repo.GetBankReconciliationByID(ctx, legalEntityID, reconciliationID)
	if err != nil {
		return nil, err
	}
	if recon.Status == domain.ReconciliationLocked {
		return nil, domain.ErrReconciliationLocked
	}
	if recon.StatementID == nil {
		return nil, fmt.Errorf("no statement attached to reconciliation session")
	}

	lines, err := s.repo.ListStatementLines(ctx, *recon.StatementID)
	if err != nil {
		return nil, err
	}

	var results []domain.AutoMatchResult
	matchedCandidateMap := make(map[uuid.UUID]bool)

	for _, line := range lines {
		if line.ReconciledStatus == domain.ReconciledMatched {
			continue
		}

		// 1. Try Exact Match (Amount + Reference)
		var bestCandidate *domain.CandidateLedgerEntry
		var bestMatchType domain.MatchType
		var confidence float64

		for i, cand := range candidateEntries {
			if matchedCandidateMap[cand.EntityID] {
				continue
			}

			// Check amount equality
			if line.Amount.Equal(cand.Amount) {
				// Exact reference match
				if line.ReferenceNumber != nil && *line.ReferenceNumber != "" && cand.ReferenceNumber != "" &&
					strings.EqualFold(*line.ReferenceNumber, cand.ReferenceNumber) {
					bestCandidate = &candidateEntries[i]
					bestMatchType = domain.MatchExact
					confidence = 1.0
					break
				}

				// If no reference matched yet, check date proximity (within 7 days)
				dateDiff := line.TransactionDate.Sub(cand.Date)
				if dateDiff < 0 {
					dateDiff = -dateDiff
				}
				if dateDiff <= 7*24*time.Hour && bestCandidate == nil {
					bestCandidate = &candidateEntries[i]
					bestMatchType = domain.MatchAutoRule
					confidence = 0.85
				}
			}
		}

		if bestCandidate != nil {
			matchedCandidateMap[bestCandidate.EntityID] = true
			matchRecord := &domain.ReconciliationMatch{
				ReconciliationID:  reconciliationID,
				StatementLineID:   line.ID,
				MatchType:         bestMatchType,
				MatchedEntityType: bestCandidate.EntityType,
				MatchedEntityID:   &bestCandidate.EntityID,
				MatchedAmount:     line.Amount,
				DifferenceAmount:  decimal.Zero,
				Notes:             &bestCandidate.Description,
				MatchedAt:         time.Now(),
			}
			if _, err := s.repo.CreateReconciliationMatch(ctx, matchRecord); err != nil {
				return nil, err
			}
			if _, err := s.repo.UpdateStatementLineReconciledStatus(ctx, line.ID, domain.ReconciledMatched); err != nil {
				return nil, err
			}

			results = append(results, domain.AutoMatchResult{
				StatementLineID: line.ID,
				Candidate:       *bestCandidate,
				MatchType:       bestMatchType,
				Confidence:      confidence,
			})
		}
	}

	return results, nil
}

func (s *bankingService) AddManualMatch(ctx context.Context, match *domain.ReconciliationMatch) (*domain.ReconciliationMatch, error) {
	if match.ReconciliationID == uuid.Nil {
		return nil, fmt.Errorf("reconciliation_id is required")
	}
	if match.StatementLineID == uuid.Nil {
		return nil, fmt.Errorf("statement_line_id is required")
	}
	if !match.MatchedAmount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}
	if match.MatchType == "" {
		match.MatchType = domain.MatchManual
	}

	created, err := s.repo.CreateReconciliationMatch(ctx, match)
	if err != nil {
		return nil, err
	}

	status := domain.ReconciledMatched
	if match.DifferenceAmount.IsPositive() {
		status = domain.ReconciledPartiallyMatched
	}
	if _, err := s.repo.UpdateStatementLineReconciledStatus(ctx, match.StatementLineID, status); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *bankingService) LockReconciliation(ctx context.Context, legalEntityID, id uuid.UUID, lockedBy uuid.UUID) (*domain.Reconciliation, error) {
	recon, err := s.repo.GetBankReconciliationByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if recon.Status == domain.ReconciliationLocked {
		return recon, nil
	}

	now := time.Now()
	recon.Status = domain.ReconciliationLocked
	recon.ReconciledAt = &now
	if lockedBy != uuid.Nil {
		recon.ReconciledByUserID = &lockedBy
	}

	return s.repo.UpdateBankReconciliation(ctx, recon)
}

func (s *bankingService) ListMatches(ctx context.Context, reconciliationID uuid.UUID) ([]*domain.ReconciliationMatch, error) {
	return s.repo.ListReconciliationMatches(ctx, reconciliationID)
}
