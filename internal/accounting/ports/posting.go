package ports

import (
	"context"

	"erp/internal/accounting/domain"

	"github.com/google/uuid"
)

// PostingEngine is the central interface through which ALL modules create
// accounting entries. It guarantees atomic, balanced, idempotent posting
// with period locking enforcement and audit trail recording.
//
// Spec reference: project-prompt.md Section 21 (Accounting Posting Engine)
//
// Invariants enforced:
//   - balanced debit/credit
//   - open accounting period
//   - active accounts with allow_posting = true
//   - idempotency via (source_module, source_type, source_id) triple
//   - atomic transaction (all-or-nothing within a single DB transaction)
//   - immutable posted journal (database triggers prevent mutation)
type PostingEngine interface {
	// Post creates a balanced journal entry, posts it atomically,
	// and updates account balances within a single database transaction.
	// Returns ErrDuplicatePosting if the source triple already exists (idempotent).
	Post(
		ctx context.Context,
		request domain.PostingRequest,
	) (domain.PostingResult, error)

	// Reverse creates a mirror journal entry that exactly offsets the original,
	// marks the original as REVERSED, and updates account balances accordingly.
	// Only POSTED entries can be reversed; reversing a REVERSED entry returns ErrAlreadyReversed.
	Reverse(
		ctx context.Context,
		journalEntryID uuid.UUID,
		reason string,
		actorID uuid.UUID,
	) (domain.PostingResult, error)
}

// PostingRepository extends AccountingRepository with the additional
// methods needed by the PostingEngine for idempotency and balance tracking.
type PostingRepository interface {
	AccountingRepository

	// GetJournalEntryBySource checks for existing entries by source triple (idempotency).
	GetJournalEntryBySource(ctx context.Context, legalEntityID uuid.UUID, sourceModule, sourceType string, sourceID uuid.UUID) (*domain.JournalEntry, error)

	// UpsertAccountBalance atomically inserts or updates account period balances.
	// Uses ON CONFLICT DO UPDATE for thread-safe accumulation.
	UpsertAccountBalance(ctx context.Context, balance domain.AccountBalance) (*domain.AccountBalance, error)
}
