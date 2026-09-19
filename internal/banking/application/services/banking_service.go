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
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type bankingService struct {
	repo ports.BankingRepository
	tx   transaction.Manager
}

func NewBankingService(repo ports.BankingRepository, db *database.DB) ports.BankingService {
	return &bankingService{
		repo: repo,
		tx:   transaction.NewManager(db.Pool),
	}
}

// ============================================================================
// BANK ACCOUNTS
// ============================================================================

func (s *bankingService) CreateBankAccount(ctx context.Context, a *domain.BankAccount) (*domain.BankAccount, error) {
	if a.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	a.AccountCode = strings.TrimSpace(a.AccountCode)
	a.BankNameEn = strings.TrimSpace(a.BankNameEn)
	a.BankNameAr = strings.TrimSpace(a.BankNameAr)
	a.AccountNumber = strings.TrimSpace(a.AccountNumber)
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
	if a.OverdraftLimit.IsNegative() {
		return nil, fmt.Errorf("overdraft_limit cannot be negative")
	}
	// Basic IBAN check
	cleanIBAN := strings.ReplaceAll(a.IBAN, " ", "")
	if len(cleanIBAN) < 15 {
		return nil, domain.ErrInvalidIBAN
	}
	a.IBAN = strings.ToUpper(cleanIBAN)

	if a.AccountNumberMasked == "" {
		a.AccountNumberMasked = domain.MaskAccountNumber(a.AccountNumber)
	}
	if a.CurrencyCode == "" {
		a.CurrencyCode = "SAR"
	}
	a.CurrencyCode = strings.ToUpper(strings.TrimSpace(a.CurrencyCode))
	if a.Status == "" {
		a.Status = domain.BankAccountActive
	} else if a.Status != domain.BankAccountActive {
		return nil, fmt.Errorf("new bank accounts must be ACTIVE")
	}
	a.CurrentBalance = a.OpeningBalance
	a.AvailableBalance = a.OpeningBalance
	if a.CurrentBalance.LessThan(a.OverdraftLimit.Neg()) || a.AvailableBalance.LessThan(a.OverdraftLimit.Neg()) {
		return nil, fmt.Errorf("opening balance exceeds the overdraft limit")
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
	if stmt.StartDate.IsZero() || stmt.EndDate.IsZero() || stmt.EndDate.Before(stmt.StartDate) {
		return nil, fmt.Errorf("a valid statement date range is required")
	}
	if stmt.CurrencyCode == "" {
		stmt.CurrencyCode = "SAR"
	}
	stmt.CurrencyCode = strings.ToUpper(strings.TrimSpace(stmt.CurrencyCode))
	if stmt.FileFormat == "" {
		stmt.FileFormat = domain.StatementFormatCSV
	}
	if stmt.Status == "" {
		stmt.Status = domain.StatementImported
	} else if stmt.Status != domain.StatementImported {
		return nil, fmt.Errorf("imported bank statements must have IMPORTED status")
	}

	totalCredits := decimal.Zero
	totalDebits := decimal.Zero
	for i, l := range lines {
		if l == nil || !l.Amount.IsPositive() || l.TransactionDate.IsZero() || strings.TrimSpace(l.Description) == "" {
			return nil, fmt.Errorf("%w at line %d", domain.ErrInvalidStatementLine, i+1)
		}
		if l.TransactionType != domain.TransactionCredit && l.TransactionType != domain.TransactionDebit {
			return nil, fmt.Errorf("%w at line %d", domain.ErrInvalidStatementLine, i+1)
		}
		if l.CurrencyCode == "" {
			l.CurrencyCode = stmt.CurrencyCode
		}
		l.CurrencyCode = strings.ToUpper(strings.TrimSpace(l.CurrencyCode))
		if l.CurrencyCode != stmt.CurrencyCode {
			return nil, fmt.Errorf("%w at line %d", domain.ErrCurrencyMismatch, i+1)
		}
		l.LineNumber = int32(i + 1)
		l.ReconciledStatus = domain.ReconciledUnmatched

		if l.TransactionType == domain.TransactionCredit {
			totalCredits = totalCredits.Add(l.Amount)
		} else if l.TransactionType == domain.TransactionDebit {
			totalDebits = totalDebits.Add(l.Amount)
		}
	}
	stmt.TotalCredits = totalCredits
	stmt.TotalDebits = totalDebits
	if stmt.ClosingBalance.IsZero() {
		stmt.ClosingBalance = stmt.OpeningBalance.Add(totalCredits).Sub(totalDebits)
	}

	var createdStmt *domain.Statement
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		account, err := s.repo.GetBankAccountByID(txCtx, stmt.LegalEntityID, stmt.BankAccountID)
		if err != nil {
			return err
		}
		if account.Status != domain.BankAccountActive {
			return domain.ErrBankAccountUnavailable
		}
		if !strings.EqualFold(account.CurrencyCode, stmt.CurrencyCode) {
			return domain.ErrCurrencyMismatch
		}

		createdStmt, err = s.repo.CreateBankStatement(txCtx, stmt)
		if err != nil {
			return err
		}
		for _, line := range lines {
			line.StatementID = createdStmt.ID
			if _, err := s.repo.CreateStatementLine(txCtx, line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdStmt, nil
}

func (s *bankingService) GetBankStatement(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Statement, error) {
	return s.repo.GetBankStatementByID(ctx, legalEntityID, id)
}

func (s *bankingService) ListBankStatements(ctx context.Context, legalEntityID, bankAccountID uuid.UUID) ([]*domain.Statement, error) {
	if _, err := s.repo.GetBankAccountByID(ctx, legalEntityID, bankAccountID); err != nil {
		return nil, err
	}
	return s.repo.ListBankStatements(ctx, bankAccountID)
}

func (s *bankingService) ListStatementLines(ctx context.Context, legalEntityID, statementID uuid.UUID) ([]*domain.StatementLine, error) {
	if _, err := s.repo.GetBankStatementByID(ctx, legalEntityID, statementID); err != nil {
		return nil, err
	}
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
	} else if recon.Status != domain.ReconciliationDraft && recon.Status != domain.ReconciliationInProgress {
		return nil, fmt.Errorf("new reconciliation must be DRAFT or IN_PROGRESS")
	}
	recon.UnreconciledDifference = recon.StatementEndingBalance.Sub(recon.GLBookBalance).Abs()
	account, err := s.repo.GetBankAccountByID(ctx, recon.LegalEntityID, recon.BankAccountID)
	if err != nil {
		return nil, err
	}
	if account.Status != domain.BankAccountActive {
		return nil, domain.ErrBankAccountUnavailable
	}
	if recon.StatementID != nil {
		statement, err := s.repo.GetBankStatementByID(ctx, recon.LegalEntityID, *recon.StatementID)
		if err != nil {
			return nil, err
		}
		if statement.BankAccountID != recon.BankAccountID {
			return nil, domain.ErrStatementAccountMismatch
		}
	}

	return s.repo.CreateBankReconciliation(ctx, recon)
}

func (s *bankingService) GetReconciliation(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error) {
	return s.repo.GetBankReconciliationByID(ctx, legalEntityID, id)
}

func (s *bankingService) ListReconciliations(ctx context.Context, legalEntityID, bankAccountID uuid.UUID) ([]*domain.Reconciliation, error) {
	if _, err := s.repo.GetBankAccountByID(ctx, legalEntityID, bankAccountID); err != nil {
		return nil, err
	}
	return s.repo.ListBankReconciliations(ctx, bankAccountID)
}

func (s *bankingService) RunAutoMatch(ctx context.Context, legalEntityID, reconciliationID uuid.UUID, candidateEntries []domain.CandidateLedgerEntry) ([]domain.AutoMatchResult, error) {
	var results []domain.AutoMatchResult
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		recon, err := s.repo.LockBankReconciliationByID(txCtx, legalEntityID, reconciliationID)
		if err != nil {
			return err
		}
		if recon.Status == domain.ReconciliationLocked {
			return domain.ErrReconciliationLocked
		}
		if recon.StatementID == nil {
			return fmt.Errorf("no statement attached to reconciliation session")
		}

		lines, err := s.repo.ListStatementLines(txCtx, *recon.StatementID)
		if err != nil {
			return err
		}
		matchedCandidateMap := make(map[uuid.UUID]bool)

		for _, lineSnapshot := range lines {
			line, err := s.repo.LockStatementLineByID(txCtx, lineSnapshot.ID)
			if err != nil {
				return err
			}
			if line.ReconciledStatus != domain.ReconciledUnmatched {
				continue
			}

			var bestCandidate *domain.CandidateLedgerEntry
			var bestMatchType domain.MatchType
			var confidence float64
			for i, candidate := range candidateEntries {
				if candidate.EntityID == uuid.Nil || matchedCandidateMap[candidate.EntityID] {
					continue
				}
				if line.Amount.Equal(candidate.Amount) {
					if line.ReferenceNumber != nil && *line.ReferenceNumber != "" && candidate.ReferenceNumber != "" &&
						strings.EqualFold(*line.ReferenceNumber, candidate.ReferenceNumber) {
						bestCandidate = &candidateEntries[i]
						bestMatchType = domain.MatchExact
						confidence = 1.0
						break
					}
					dateDiff := line.TransactionDate.Sub(candidate.Date)
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
				if !isValidMatchedEntityType(bestCandidate.EntityType) {
					return domain.ErrInvalidMatchedEntityType
				}
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
				if _, err := s.repo.CreateReconciliationMatch(txCtx, matchRecord); err != nil {
					return err
				}
				if _, err := s.repo.UpdateStatementLineReconciledStatus(txCtx, line.ID, domain.ReconciledMatched); err != nil {
					return err
				}
				matchedCandidateMap[bestCandidate.EntityID] = true
				results = append(results, domain.AutoMatchResult{
					StatementLineID: line.ID,
					Candidate:       *bestCandidate,
					MatchType:       bestMatchType,
					Confidence:      confidence,
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *bankingService) AddManualMatch(ctx context.Context, legalEntityID uuid.UUID, match *domain.ReconciliationMatch) (*domain.ReconciliationMatch, error) {
	if legalEntityID == uuid.Nil || match.ReconciliationID == uuid.Nil {
		return nil, fmt.Errorf("reconciliation_id is required")
	}
	if match.StatementLineID == uuid.Nil {
		return nil, fmt.Errorf("statement_line_id is required")
	}
	if !match.MatchedAmount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}
	if !isValidMatchedEntityType(match.MatchedEntityType) {
		return nil, domain.ErrInvalidMatchedEntityType
	}
	match.MatchType = domain.MatchManual

	var created *domain.ReconciliationMatch
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		reconciliation, err := s.repo.LockBankReconciliationByID(txCtx, legalEntityID, match.ReconciliationID)
		if err != nil {
			return err
		}
		if reconciliation.Status == domain.ReconciliationLocked {
			return domain.ErrReconciliationLocked
		}
		if reconciliation.StatementID == nil {
			return fmt.Errorf("no statement attached to reconciliation session")
		}
		line, err := s.repo.LockStatementLineByID(txCtx, match.StatementLineID)
		if err != nil {
			return err
		}
		if line.StatementID != *reconciliation.StatementID {
			return domain.ErrStatementLineMismatch
		}
		if line.ReconciledStatus == domain.ReconciledMatched {
			return domain.ErrStatementLineMatched
		}

		existingMatches, err := s.repo.ListReconciliationMatches(txCtx, reconciliation.ID)
		if err != nil {
			return err
		}
		alreadyMatched := decimal.Zero
		for _, existingMatch := range existingMatches {
			if existingMatch.StatementLineID == line.ID {
				alreadyMatched = alreadyMatched.Add(existingMatch.MatchedAmount)
			}
		}
		remainingAmount := line.Amount.Sub(alreadyMatched)
		if !remainingAmount.IsPositive() {
			return domain.ErrStatementLineMatched
		}
		if match.MatchedAmount.GreaterThan(remainingAmount) {
			return domain.ErrInvalidAmount
		}

		match.DifferenceAmount = remainingAmount.Sub(match.MatchedAmount)
		created, err = s.repo.CreateReconciliationMatch(txCtx, match)
		if err != nil {
			return err
		}
		status := domain.ReconciledMatched
		if match.DifferenceAmount.IsPositive() {
			status = domain.ReconciledPartiallyMatched
		}
		_, err = s.repo.UpdateStatementLineReconciledStatus(txCtx, match.StatementLineID, status)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func isValidMatchedEntityType(entityType domain.MatchedEntityType) bool {
	switch entityType {
	case domain.MatchedEntityJournalEntry,
		domain.MatchedEntityReceipt,
		domain.MatchedEntityPayment,
		domain.MatchedEntityTransfer,
		domain.MatchedEntityBankFeeAdjustment:
		return true
	default:
		return false
	}
}

func (s *bankingService) LockReconciliation(ctx context.Context, legalEntityID, id uuid.UUID, lockedBy uuid.UUID) (*domain.Reconciliation, error) {
	var locked *domain.Reconciliation
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		reconciliation, err := s.repo.LockBankReconciliationByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if reconciliation.Status == domain.ReconciliationLocked {
			locked = reconciliation
			return nil
		}

		now := time.Now()
		reconciliation.Status = domain.ReconciliationLocked
		reconciliation.ReconciledAt = &now
		if lockedBy != uuid.Nil {
			reconciliation.ReconciledByUserID = &lockedBy
		}
		locked, err = s.repo.UpdateBankReconciliation(txCtx, reconciliation)
		return err
	})
	if err != nil {
		return nil, err
	}
	return locked, nil
}

func (s *bankingService) ListMatches(ctx context.Context, legalEntityID, reconciliationID uuid.UUID) ([]*domain.ReconciliationMatch, error) {
	if _, err := s.repo.GetBankReconciliationByID(ctx, legalEntityID, reconciliationID); err != nil {
		return nil, err
	}
	return s.repo.ListReconciliationMatches(ctx, reconciliationID)
}
