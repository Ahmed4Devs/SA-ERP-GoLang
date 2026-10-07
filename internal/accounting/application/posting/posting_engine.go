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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"
	auditports "erp/internal/audit/ports"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type postingEngine struct {
	repo  ports.PostingRepository
	tx    transaction.Manager
	audit ports.PostingAuditRecorder
}

// NewPostingEngine creates a new PostingEngine backed by the given PostingRepository.
func NewPostingEngine(repo ports.PostingRepository, tx transaction.Manager, audit ports.PostingAuditRecorder) ports.PostingEngine {
	return &postingEngine{repo: repo, tx: tx, audit: audit}
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
	var result domain.PostingResult
	err := e.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := e.lockPosting(txCtx); err != nil {
			return err
		}
		var err error
		result, err = e.post(txCtx, request)
		return err
	})
	return result, err
}

func (e *postingEngine) post(ctx context.Context, request domain.PostingRequest) (domain.PostingResult, error) {
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
	if !journal.IsActive || journal.LegalEntityID != request.LegalEntityID {
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
		if acc.LegalEntityID != request.LegalEntityID {
			return domain.PostingResult{}, domain.ErrAccountNotFound
		}
		if err := domain.ValidateAccountCanPost(acc.AllowPosting, acc.IsActive); err != nil {
			return domain.PostingResult{}, fmt.Errorf("line %d (account %s): %w", lineNum, acc.Code, err)
		}

		// Calculate base amounts if not provided
		baseDebit := line.BaseDebit
		baseCredit := line.BaseCredit
		if baseDebit.IsZero() && line.TransactionDebit.GreaterThan(decimal.Zero) {
			baseDebit = line.TransactionDebit.Mul(request.ExchangeRate).Round(4)
		}
		if baseCredit.IsZero() && line.TransactionCredit.GreaterThan(decimal.Zero) {
			baseCredit = line.TransactionCredit.Mul(request.ExchangeRate).Round(4)
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
	entryNumber := identifiers.NewDocumentNumber("PE", request.AccountingDate)

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

	return e.postEntry(ctx, created, request.ActorID)
}

// PostDraft preserves the draft ID, number, creator and persisted lines. The
// journal lock serializes retries before any balances or audit are written.
func (e *postingEngine) PostDraft(ctx context.Context, journalEntryID, actorID uuid.UUID) (domain.PostingResult, error) {
	var result domain.PostingResult
	err := e.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if actorID == uuid.Nil || journalEntryID == uuid.Nil {
			return fmt.Errorf("%w: journal_entry_id and actor_id are required", domain.ErrInvalidPostingRequest)
		}
		if err := e.lockPosting(txCtx); err != nil {
			return err
		}
		entry, err := e.repo.GetJournalEntryByID(txCtx, journalEntryID)
		if err != nil {
			return err
		}
		switch entry.Status {
		case domain.JournalEntryStatusPosted:
			return domain.ErrAlreadyPosted
		case domain.JournalEntryStatusReversed:
			return domain.ErrAlreadyReversed
		case domain.JournalEntryStatusDraft, domain.JournalEntryStatusSubmitted, domain.JournalEntryStatusApproved:
		default:
			return fmt.Errorf("%w: journal status %s cannot be posted", domain.ErrInvalidPostingRequest, entry.Status)
		}
		result, err = e.postEntry(txCtx, entry, actorID)
		return err
	})
	return result, err
}

func (e *postingEngine) postEntry(ctx context.Context, entry *domain.JournalEntry, actorID uuid.UUID) (domain.PostingResult, error) {
	if err := e.validateEntryForPosting(ctx, entry); err != nil {
		return domain.PostingResult{}, err
	}
	posted, err := e.repo.PostJournalEntry(ctx, entry.ID, actorID)
	if err != nil {
		return domain.PostingResult{}, fmt.Errorf("posting post entry: %w", err)
	}
	if err := e.updateBalances(ctx, posted); err != nil {
		return domain.PostingResult{}, err
	}
	if err := e.recordPostingAudit(ctx, "accounting.journal.post", actorID, posted, nil, ""); err != nil {
		return domain.PostingResult{}, err
	}
	return domain.PostingResult{JournalEntryID: posted.ID, EntryNumber: posted.EntryNumber}, nil
}

// Revalidate stored drafts at the posting boundary: the selected journal,
// period or accounts may have changed since the draft was saved. Repository
// reads retain shared row locks until the surrounding transaction finishes.
func (e *postingEngine) validateEntryForPosting(ctx context.Context, entry *domain.JournalEntry) error {
	if entry.ExchangeRate.LessThanOrEqual(decimal.Zero) || strings.TrimSpace(entry.CurrencyCode) == "" {
		return fmt.Errorf("%w: currency and a positive exchange rate are required", domain.ErrInvalidPostingRequest)
	}
	journal, err := e.repo.GetJournalByID(ctx, entry.JournalID)
	if err != nil {
		return err
	}
	if !journal.IsActive || journal.LegalEntityID != entry.LegalEntityID {
		return domain.ErrJournalNotFound
	}
	period, err := e.repo.GetFiscalPeriodByDate(ctx, entry.LegalEntityID, entry.AccountingDate)
	if err != nil {
		return err
	}
	if period.ID != entry.FiscalPeriodID {
		return domain.ErrFiscalPeriodNotFound
	}
	if err := domain.ValidatePeriodCanPost(period.Status); err != nil {
		return err
	}
	for _, line := range entry.Lines {
		if line.LegalEntityID != entry.LegalEntityID {
			return domain.ErrAccountNotFound
		}
		account, err := e.repo.GetAccountByID(ctx, line.AccountID)
		if err != nil {
			return err
		}
		if account.LegalEntityID != entry.LegalEntityID {
			return domain.ErrAccountNotFound
		}
		if err := domain.ValidateAccountCanPost(account.AllowPosting, account.IsActive); err != nil {
			return fmt.Errorf("line %d: %w", line.LineNumber, err)
		}
		if err := e.repo.ValidateLineDimensions(ctx, entry.LegalEntityID, line); err != nil {
			return fmt.Errorf("line %d: %w", line.LineNumber, err)
		}
		if line.ExchangeRate.LessThanOrEqual(decimal.Zero) || strings.TrimSpace(line.CurrencyCode) == "" {
			return fmt.Errorf("%w: line %d requires currency and a positive exchange rate", domain.ErrInvalidPostingRequest, line.LineNumber)
		}
		if line.TransactionDebit.IsNegative() || line.TransactionCredit.IsNegative() ||
			(line.TransactionDebit.IsPositive() && !line.BaseDebit.IsPositive()) ||
			(line.TransactionCredit.IsPositive() && !line.BaseCredit.IsPositive()) {
			return domain.ErrInvalidLineAmounts
		}
	}
	debit, credit, err := domain.ValidateJournalEntryBalance(entry.Lines)
	if err != nil {
		return err
	}
	if !debit.Equal(entry.TotalDebit) || !credit.Equal(entry.TotalCredit) {
		return fmt.Errorf("%w: journal totals differ from persisted lines", domain.ErrUnbalancedJournalEntry)
	}
	return nil
}

func (e *postingEngine) updateBalances(ctx context.Context, entry *domain.JournalEntry) error {
	for _, line := range entry.Lines {
		_, err := e.repo.UpsertAccountBalance(ctx, domain.AccountBalance{
			LegalEntityID: entry.LegalEntityID, AccountID: line.AccountID,
			FiscalPeriodID: entry.FiscalPeriodID, OpeningBalance: decimal.Zero,
			PeriodDebit: line.BaseDebit, PeriodCredit: line.BaseCredit,
			ClosingBalance: line.BaseDebit.Sub(line.BaseCredit),
		})
		if err != nil {
			return fmt.Errorf("posting update balance for account %s: %w", line.AccountID, err)
		}
	}
	return nil
}

// Reverse creates a mirror journal entry that exactly offsets the original posted entry,
// marks the original as REVERSED, and adjusts account balances accordingly.
func (e *postingEngine) Reverse(ctx context.Context, journalEntryID uuid.UUID, reason string, actorID uuid.UUID) (domain.PostingResult, error) {
	var result domain.PostingResult
	err := e.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := e.lockPosting(txCtx); err != nil {
			return err
		}
		var err error
		result, err = e.reverse(txCtx, journalEntryID, reason, actorID)
		return err
	})
	return result, err
}

func (e *postingEngine) lockPosting(ctx context.Context) error {
	if e.audit == nil {
		return fmt.Errorf("posting audit recorder is not configured")
	}
	// Acquire the audit-chain lock before financial row locks, including when
	// several posts are composed in one outer transaction. All posting callers
	// that lock their own business rows must acquire this same lock first.
	return e.audit.LockChain(ctx)
}

func (e *postingEngine) reverse(ctx context.Context, journalEntryID uuid.UUID, reason string, actorID uuid.UUID) (domain.PostingResult, error) {
	if actorID == uuid.Nil || strings.TrimSpace(reason) == "" {
		return domain.PostingResult{}, fmt.Errorf("%w: reversal actor and reason are required", domain.ErrInvalidPostingRequest)
	}
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

	reversalEntry.Lines = reversalLines
	if err := e.validateEntryForPosting(ctx, &reversalEntry); err != nil {
		return domain.PostingResult{}, err
	}

	// 6. Delegate to repository for atomic reversal
	reversed, err := e.repo.ReverseJournalEntry(ctx, journalEntryID, reversalEntry, reversalLines, actorID)
	if err != nil {
		return domain.PostingResult{}, fmt.Errorf("posting reverse: %w", err)
	}

	if err := e.updateBalances(ctx, reversed); err != nil {
		return domain.PostingResult{}, err
	}

	if err := e.recordPostingAudit(ctx, "accounting.journal.reverse", actorID, reversed, original, reason); err != nil {
		return domain.PostingResult{}, err
	}

	return domain.PostingResult{
		JournalEntryID: reversed.ID,
		EntryNumber:    reversed.EntryNumber,
	}, nil
}

// validatePostingRequest ensures all required fields are present.
func validatePostingRequest(r domain.PostingRequest) error {
	if r.ExchangeRate.IsNegative() {
		return fmt.Errorf("%w: exchange_rate must be positive", domain.ErrInvalidPostingRequest)
	}
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

func (e *postingEngine) recordPostingAudit(ctx context.Context, action string, actorID uuid.UUID, entry, original *domain.JournalEntry, reason string) error {
	if e.audit == nil {
		return fmt.Errorf("posting audit recorder is not configured")
	}
	payload, err := json.Marshal(struct {
		JournalEntry         *domain.JournalEntry `json:"journal_entry"`
		OriginalJournalEntry *domain.JournalEntry `json:"original_journal_entry,omitempty"`
		Reason               string               `json:"reason,omitempty"`
	}{entry, original, reason})
	if err != nil {
		return fmt.Errorf("encode posting audit: %w", err)
	}
	_, err = e.audit.RecordEvent(ctx, auditports.RecordEventCommand{
		UserID: &actorID, LegalEntityID: &entry.LegalEntityID, Action: action,
		ResourceType: "accounting.journal", ResourceID: entry.ID.String(), Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("record posting audit: %w", err)
	}
	return nil
}
