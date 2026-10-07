package accounting_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"erp/internal/accounting/application/posting"
	"erp/internal/accounting/domain"
	"erp/internal/accounting/infrastructure/postgres"
	"erp/internal/accounting/ports"
	auditservices "erp/internal/audit/application/services"
	auditdomain "erp/internal/audit/domain"
	auditpostgres "erp/internal/audit/infrastructure/postgres"
	auditports "erp/internal/audit/ports"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var errAuditWrite = errors.New("injected audit acknowledgement failure")

// Simulate a mandatory failure after the actual audit insert, so both the
// audit row and financial writes must roll back rather than leave an orphan.
type failingAuditRecorder struct{ ports.PostingAuditRecorder }

func (r failingAuditRecorder) RecordEvent(ctx context.Context, command auditports.RecordEventCommand) (*auditdomain.AuditEvent, error) {
	if _, err := r.PostingAuditRecorder.RecordEvent(ctx, command); err != nil {
		return nil, err
	}
	return nil, errAuditWrite
}

type notifyingAuditLocker struct {
	ports.PostingAuditRecorder
	started chan struct{}
}

func (r notifyingAuditLocker) LockChain(ctx context.Context) error {
	close(r.started)
	return r.PostingAuditRecorder.LockChain(ctx)
}

func TestPostingEngine_ComposedPostsKeepAuditBeforeBalanceLockOrder(t *testing.T) {
	mod, _, _, db, entityID, actorID, setupCtx := setupAccountingTest(t)
	defer db.Close()
	ctx, cancel := context.WithTimeout(setupCtx, 10*time.Second)
	defer cancel()
	audit := auditservices.NewAuditService(auditpostgres.NewAuditRepository(db.Pool))
	journals, err := mod.AccountingService.ListJournals(ctx, entityID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("journals: %v", err)
	}
	accountID := func(code string) uuid.UUID {
		t.Helper()
		account, err := mod.AccountingService.GetAccountByCode(ctx, entityID, code)
		if err != nil {
			t.Fatal(err)
		}
		return account.ID
	}
	first := domain.PostingRequest{
		LegalEntityID: entityID, SourceModule: "TEST", SourceType: "BATCH_LOCK", SourceID: uuid.New(),
		JournalID: journals[0].ID, AccountingDate: time.Now().UTC(), ActorID: actorID,
		Lines: []domain.PostingLine{
			{AccountID: accountID("111001"), TransactionDebit: decimal.NewFromInt(87)},
			{AccountID: accountID("510000"), TransactionCredit: decimal.NewFromInt(87)},
		},
	}
	second := first
	second.SourceID = uuid.New()
	second.Lines = []domain.PostingLine{
		{AccountID: accountID("111002"), TransactionDebit: decimal.NewFromInt(89)},
		{AccountID: accountID("410000"), TransactionCredit: decimal.NewFromInt(89)},
	}
	competing := second
	competing.SourceID = uuid.New()
	started := make(chan struct{})
	completed := make(chan error, 1)
	otherEngine := posting.NewPostingEngine(postgres.NewPostingRepository(db.Pool), transaction.NewManager(db.Pool), notifyingAuditLocker{audit, started})
	err = transaction.NewManager(db.Pool).WithinTx(ctx, func(txCtx context.Context) error {
		if _, err := mod.PostingEngine.Post(txCtx, first); err != nil {
			return err
		}
		go func() {
			_, err := otherEngine.Post(ctx, competing)
			completed <- err
		}()
		select {
		case <-started:
		case <-ctx.Done():
			return ctx.Err()
		}
		// The competing transaction must wait for the audit lock before it can
		// hold the account balances needed by this transaction's second post.
		_, err := mod.PostingEngine.Post(txCtx, second)
		return err
	})
	if err != nil {
		t.Fatalf("composed posting failed: %v", err)
	}
	select {
	case err := <-completed:
		if err != nil {
			t.Fatalf("competing posting failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("competing posting did not finish after the outer transaction committed")
	}
	verification, err := audit.VerifyChainIntegrity(ctx)
	if err != nil || !verification.IsValid {
		t.Fatalf("composed posting damaged audit chain: %+v %v", verification, err)
	}
}

func TestPostingEngine_AuditIsMandatoryAtomicAndIdempotent(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
	audit := auditservices.NewAuditService(auditpostgres.NewAuditRepository(db.Pool))
	journals, err := mod.AccountingService.ListJournals(ctx, entityID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("journals: %v", err)
	}
	cash, err := mod.AccountingService.GetAccountByCode(ctx, entityID, "111002")
	if err != nil {
		t.Fatal(err)
	}
	revenue, err := mod.AccountingService.GetAccountByCode(ctx, entityID, "410000")
	if err != nil {
		t.Fatal(err)
	}
	request := domain.PostingRequest{
		LegalEntityID: entityID, SourceModule: "TEST", SourceType: "AUDIT", SourceID: uuid.New(),
		JournalID: journals[0].ID, AccountingDate: time.Now().UTC(), ActorID: actorID,
		Lines: []domain.PostingLine{
			{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(81)},
			{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(81)},
		},
	}
	engine := posting.NewPostingEngine(postgres.NewPostingRepository(db.Pool), transaction.NewManager(db.Pool), failingAuditRecorder{audit})
	if _, err := engine.Post(ctx, request); !errors.Is(err, errAuditWrite) {
		t.Fatalf("expected audit failure, got %v", err)
	}
	countRows := func(query string, args ...any) int {
		t.Helper()
		var count int
		if err := db.Pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
			t.Fatal(err)
		}
		return count
	}
	if countRows("SELECT count(*) FROM accounting.journal_entries WHERE source_id=$1", request.SourceID) != 0 ||
		countRows("SELECT count(*) FROM audit.events WHERE payload->'journal_entry'->>'source_id'=$1", request.SourceID.String()) != 0 {
		t.Fatal("audit failure left financial or audit rows committed")
	}
	posted, err := mod.PostingEngine.Post(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.PostingEngine.Post(ctx, request); !errors.Is(err, domain.ErrDuplicatePosting) {
		t.Fatalf("duplicate posting must be rejected: %v", err)
	}
	if countRows("SELECT count(*) FROM audit.events WHERE resource_id=$1 AND action='accounting.journal.post'", posted.JournalEntryID.String()) != 1 {
		t.Fatal("posting and duplicate retry must leave exactly one posting audit event")
	}
	if _, err := engine.Reverse(ctx, posted.JournalEntryID, "Mandatory audit failure", actorID); !errors.Is(err, errAuditWrite) {
		t.Fatalf("expected reversal audit failure, got %v", err)
	}
	original, err := mod.LedgerService.GetJournalEntryByID(ctx, posted.JournalEntryID)
	if err != nil {
		t.Fatal(err)
	}
	if original.Status != domain.JournalEntryStatusPosted || original.ReversalOfID != nil ||
		countRows("SELECT count(*) FROM audit.events WHERE payload->'original_journal_entry'->>'id'=$1", original.ID.String()) != 0 {
		t.Fatal("failed reversal audit did not roll back the original status and audit event")
	}
	reversed, err := mod.PostingEngine.Reverse(ctx, original.ID, "Approved correction", actorID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.PostingEngine.Reverse(ctx, original.ID, "Duplicate retry", actorID); !errors.Is(err, domain.ErrAlreadyReversed) {
		t.Fatalf("duplicate reversal must be rejected: %v", err)
	}
	if countRows("SELECT count(*) FROM audit.events WHERE resource_id=$1 AND action='accounting.journal.reverse'", reversed.JournalEntryID.String()) != 1 {
		t.Fatal("reversal and retry must leave exactly one reversal audit event")
	}
	var reason, auditActor, originalID string
	if err := db.Pool.QueryRow(ctx, `SELECT payload->>'reason', user_id::text, payload->'original_journal_entry'->>'id'
		FROM audit.events WHERE resource_id=$1 AND action='accounting.journal.reverse'`, reversed.JournalEntryID.String()).Scan(&reason, &auditActor, &originalID); err != nil {
		t.Fatal(err)
	}
	if reason != "Approved correction" || auditActor != actorID.String() || originalID != original.ID.String() {
		t.Fatal("reversal audit lost its actor, original document or reason")
	}
	verification, err := audit.VerifyChainIntegrity(ctx)
	if err != nil || !verification.IsValid {
		t.Fatalf("hash chain failed after successful and rolled-back postings: %+v %v", verification, err)
	}
}
