// Package posting implements the transactional PostingEngine, the central gateway
// through which all ERP modules (GL, Sales, Purchases, Payroll, Inventory) create
// accounting entries in the General Ledger.
//
// Architecture decision: The engine runs entirely within a single PostgreSQL transaction
// to guarantee atomicity. The DB-level immutability triggers (trg_prevent_posted_journal_mutation)
// provide a defence-in-depth layer that cannot be bypassed even by direct SQL.
//
// Spec reference: project-prompt.md Section 21, Phase 9 (Section 80)
package posting

import (
	"context"
	"fmt"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type postingEngine struct {
	repo ports.PostingRepository
}

// NewPostingEngine creates a new PostingEngine backed by the given PostingRepository.
func NewPostingEngine(repo ports.PostingRepository) ports.PostingEngine {
	return &postingEngine{repo: repo}
}

// Post creates a balanced journal entry, posts it atomically, and updates
// account period balances — all within a single database transaction.
//
// Invariants enforced:
//  1. Request validation (required fields)
//  2. Idempotency via (source_module, source_type, source_id) triple
//  3. Journal exists and is active
//  4. Fiscal period is OPEN for the accounting date
//  5. All accounts are active and allow_posting = true
//  6. Double-entry balance: sum(base_debit) == sum(base_credit)
//  7. Atomic posting (entry created as DRAFT, lines inserted, then posted)
//  8. Account balance accumulation for the fiscal period
func (e *postingEngine) Post(ctx context.Context, request domain.PostingRequest) (domain.PostingResult, error) {
	// 1. Validate request
	if err := validatePostingRequest(request); err != nil {
		return domain.PostingResult{}, err
	}

	// 2. Idempotency check — if an entry with the same source triple exists, reject
	existing, err := e.repo.GetJournalEntryBySource(
		ctx, request.LegalEntityID,
		request.SourceModule, request.SourceType, request.SourceID,
	)
	if err != nil {
		return domain.PostingResult{}, fmt.Errorf("posting idempotency check: %w", err)
	}
	if existing != nil {
		return domain.PostingResult{}, domain.ErrDuplicatePosting
	}

	// 3. Verify journal exists and is active
	journal, err := e.repo.GetJournalByID(ctx, request.JournalID)
	if err != nil {
		return domain.PostingResult{}, domain.ErrJournalNotFound
	}
	if !journal.IsActive {
		return domain.PostingResult{}, domain.ErrJournalNotFound
	}

	// 4. Resolve fiscal period and verify it is OPEN
	period, err := e.repo.GetFiscalPeriodByDate(ctx, request.LegalEntityID, request.AccountingDate)
	if err != nil {
		return domain.PostingResult{}, domain.ErrFiscalPeriodNotFound
	}
	if err := domain.ValidatePeriodCanPost(period.Status); err != nil {
		return domain.PostingResult{}, err
	}

	// 5. Defaults
	if request.CurrencyCode == "" {
		request.CurrencyCode = "SAR"
	}
	if request.ExchangeRate.IsZero() {
		request.ExchangeRate = decimal.NewFromInt(1)
	}

	// 6. Prepare journal lines with validation
	journalLines := make([]domain.JournalLine, 0, len(request.Lines))
	for i, line := range request.Lines {
		lineNum := int16(i + 1)

		// Validate account exists, is active, and allows posting
		acc, err := e.repo.GetAccountByID(ctx, line.AccountID)
		if err != nil {
			return domain.PostingResult{}, fmt.Errorf("line %d: %w", lineNum, domain.ErrAccountNotFound)
		}
		if err := domain.ValidateAccountCanPost(acc.AllowPosting, acc.IsActive); err != nil {
			return domain.PostingResult{}, fmt.Errorf("line %d (account %s): %w", lineNum, acc.Code, err)
		}

		// Calculate base amounts if not provided
		baseDebit := line.BaseDebit
		baseCredit := line.BaseCredit
		if baseDebit.IsZero() && line.TransactionDebit.GreaterThan(decimal.Zero) {
			baseDebit = line.TransactionDebit.Mul(request.ExchangeRate)
		}
		if baseCredit.IsZero() && line.TransactionCredit.GreaterThan(decimal.Zero) {
			baseCredit = line.TransactionCredit.Mul(request.ExchangeRate)
		}

		journalLines = append(journalLines, domain.JournalLine{
			LineNumber:        lineNum,
			LegalEntityID:     request.LegalEntityID,
			AccountID:         line.AccountID,
			Description:       &line.Description,
			CurrencyCode:      request.CurrencyCode,
			ExchangeRate:      request.ExchangeRate,
			TransactionDebit:  line.TransactionDebit,
			TransactionCredit: line.TransactionCredit,
			BaseDebit:         baseDebit,
			BaseCredit:        baseCredit,
			BranchID:          line.BranchID,
			DepartmentID:      line.DepartmentID,
			CostCenterID:      line.CostCenterID,
		})
	}

	// 7. Validate double-entry balance
	totalDebit, totalCredit, err := domain.ValidateJournalEntryBalance(journalLines)
	if err != nil {
		return domain.PostingResult{}, err
	}

	// 8. Generate entry number
	entryNumber := fmt.Sprintf("PE-%d-%06d", request.AccountingDate.Year(), time.Now().UnixNano()%1000000)

	// 9. Build and persist journal entry
	sourceType := request.SourceType
	sourceID := request.SourceID

	entry := domain.JournalEntry{
		LegalEntityID:  request.LegalEntityID,
		JournalID:      request.JournalID,
		FiscalPeriodID: period.ID,
		EntryNumber:    entryNumber,
		AccountingDate: request.AccountingDate,
		Status:         domain.JournalEntryStatusDraft,
		SourceModule:   request.SourceModule,
		SourceType:     &sourceType,
		SourceID:       &sourceID,
		Reference:      &request.Reference,
		Description:    request.Description,
		CurrencyCode:   request.CurrencyCode,
		ExchangeRate:   request.ExchangeRate,
		TotalDebit:     totalDebit,
		TotalCredit:    totalCredit,
		CreatedBy:      request.ActorID,
	}

	created, err := e.repo.CreateJournalEntry(ctx, entry, journalLines)
	if err != nil {
		return domain.PostingResult{}, fmt.Errorf("posting create entry: %w", err)
	}

	// 10. Post the journal entry (transitions DRAFT → POSTED)
	posted, err := e.repo.PostJournalEntry(ctx, created.ID, request.ActorID)
	if err != nil {
		return domain.PostingResult{}, fmt.Errorf("posting post entry: %w", err)
	}

	// 11. Update account balances for each line
	for _, line := range journalLines {
		closingDelta := line.BaseDebit.Sub(line.BaseCredit)
		_, err := e.repo.UpsertAccountBalance(ctx, domain.AccountBalance{
			LegalEntityID:  request.LegalEntityID,
			AccountID:      line.AccountID,
			FiscalPeriodID: period.ID,
			OpeningBalance: decimal.Zero,
			PeriodDebit:    line.BaseDebit,
			PeriodCredit:   line.BaseCredit,
			ClosingBalance: closingDelta,
		})
		if err != nil {
			return domain.PostingResult{}, fmt.Errorf("posting update balance for account %s: %w", line.AccountID, err)
		}
	}

	return domain.PostingResult{
		JournalEntryID: posted.ID,
		EntryNumber:    posted.EntryNumber,
	}, nil
}

// Reverse creates a mirror journal entry that exactly offsets the original posted entry,
// marks the original as REVERSED, and adjusts account balances accordingly.
func (e *postingEngine) Reverse(ctx context.Context, journalEntryID uuid.UUID, reason string, actorID uuid.UUID) (domain.PostingResult, error) {
	// 1. Load original entry with lines
	original, err := e.repo.GetJournalEntryByID(ctx, journalEntryID)
	if err != nil {
		return domain.PostingResult{}, domain.ErrJournalEntryNotFound
	}

	// 2. Verify it is in POSTED status and not already reversed
	if original.Status == domain.JournalEntryStatusReversed || original.ReversalOfID != nil {
		return domain.PostingResult{}, domain.ErrAlreadyReversed
	}
	if original.Status != domain.JournalEntryStatusPosted {
		return domain.PostingResult{}, domain.ErrCannotReverseUnposted
	}

	// 3. Resolve current fiscal period for reversal date
	reversalDate := time.Now().UTC()
	period, err := e.repo.GetFiscalPeriodByDate(ctx, original.LegalEntityID, reversalDate)
	if err != nil {
		return domain.PostingResult{}, domain.ErrFiscalPeriodNotFound
	}
	if err := domain.ValidatePeriodCanPost(period.Status); err != nil {
		return domain.PostingResult{}, err
	}

	// 4. Build reversal lines (swap debits and credits)
	reversalLines := make([]domain.JournalLine, 0, len(original.Lines))
	for _, l := range original.Lines {
		reversalLines = append(reversalLines, domain.JournalLine{
			LineNumber:        l.LineNumber,
			LegalEntityID:     original.LegalEntityID,
			AccountID:         l.AccountID,
			Description:       l.Description,
			CurrencyCode:      l.CurrencyCode,
			ExchangeRate:      l.ExchangeRate,
			TransactionDebit:  l.TransactionCredit, // SWAPPED
			TransactionCredit: l.TransactionDebit,  // SWAPPED
			BaseDebit:         l.BaseCredit,        // SWAPPED
			BaseCredit:        l.BaseDebit,         // SWAPPED
			BranchID:          l.BranchID,
			DepartmentID:      l.DepartmentID,
			CostCenterID:      l.CostCenterID,
		})
	}

	// 5. Build reversal entry
	reversalEntry := domain.JournalEntry{
		LegalEntityID:  original.LegalEntityID,
		JournalID:      original.JournalID,
		FiscalPeriodID: period.ID,
		EntryNumber:    fmt.Sprintf("REV-%s", original.EntryNumber),
		AccountingDate: reversalDate,
		Status:         domain.JournalEntryStatusPosted,
		SourceModule:   "GL",
		Reference:      &original.EntryNumber,
		Description:    fmt.Sprintf("عكس القيد المحاسبي %s: %s", original.EntryNumber, reason),
		CurrencyCode:   original.CurrencyCode,
		ExchangeRate:   original.ExchangeRate,
		TotalDebit:     original.TotalCredit,
		TotalCredit:    original.TotalDebit,
		CreatedBy:      actorID,
	}

	// 6. Delegate to repository for atomic reversal
	reversed, err := e.repo.ReverseJournalEntry(ctx, journalEntryID, reversalEntry, reversalLines, actorID)
	if err != nil {
		return domain.PostingResult{}, fmt.Errorf("posting reverse: %w", err)
	}

	// 7. Update account balances for the reversal lines
	for _, line := range reversalLines {
		closingDelta := line.BaseDebit.Sub(line.BaseCredit)
		_, err := e.repo.UpsertAccountBalance(ctx, domain.AccountBalance{
			LegalEntityID:  original.LegalEntityID,
			AccountID:      line.AccountID,
			FiscalPeriodID: period.ID,
			OpeningBalance: decimal.Zero,
			PeriodDebit:    line.BaseDebit,
			PeriodCredit:   line.BaseCredit,
			ClosingBalance: closingDelta,
		})
		if err != nil {
			return domain.PostingResult{}, fmt.Errorf("posting reverse balance for account %s: %w", line.AccountID, err)
		}
	}

	return domain.PostingResult{
		JournalEntryID: reversed.ID,
		EntryNumber:    reversed.EntryNumber,
	}, nil
}

// validatePostingRequest ensures all required fields are present.
func validatePostingRequest(r domain.PostingRequest) error {
	if r.LegalEntityID == uuid.Nil {
		return fmt.Errorf("%w: legal_entity_id is required", domain.ErrInvalidPostingRequest)
	}
	if r.SourceModule == "" {
		return fmt.Errorf("%w: source_module is required", domain.ErrInvalidPostingRequest)
	}
	if r.SourceType == "" {
		return fmt.Errorf("%w: source_type is required", domain.ErrInvalidPostingRequest)
	}
	if r.SourceID == uuid.Nil {
		return fmt.Errorf("%w: source_id is required", domain.ErrInvalidPostingRequest)
	}
	if r.JournalID == uuid.Nil {
		return fmt.Errorf("%w: journal_id is required", domain.ErrInvalidPostingRequest)
	}
	if r.AccountingDate.IsZero() {
		return fmt.Errorf("%w: accounting_date is required", domain.ErrInvalidPostingRequest)
	}
	if len(r.Lines) < 2 {
		return fmt.Errorf("%w: at least two posting lines are required", domain.ErrInvalidPostingRequest)
	}
	if r.ActorID == uuid.Nil {
		return fmt.Errorf("%w: actor_id is required", domain.ErrInvalidPostingRequest)
	}
	return nil
}
