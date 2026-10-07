package accounting_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/accounting/application/posting"
	"erp/internal/accounting/application/services"
	"erp/internal/accounting/domain"
	"erp/internal/accounting/infrastructure/postgres"
	"erp/internal/accounting/ports"
	auditservices "erp/internal/audit/application/services"
	auditpostgres "erp/internal/audit/infrastructure/postgres"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

func manualDraft(t *testing.T, ctx context.Context, mod *accounting.Module, entityID, creatorID uuid.UUID) *domain.JournalEntry {
	t.Helper()
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
	entry, err := mod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID: entityID, JournalID: journals[0].ID, AccountingDate: time.Now().UTC(),
		CreatedBy: creatorID, Description: "Manual posting boundary",
	}, []domain.JournalLine{
		{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(123)},
		{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(123)},
	})
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

func manualBalance(t *testing.T, ctx context.Context, db *database.DB, accountID uuid.UUID) decimal.Decimal {
	t.Helper()
	var balance decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT COALESCE(sum(closing_balance),0) FROM accounting.account_balances WHERE account_id=$1", accountID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	return balance
}

func manualAuditCount(t *testing.T, ctx context.Context, db *database.DB, entryID uuid.UUID) int {
	t.Helper()
	var count int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE resource_id=$1 OR payload->'original_journal_entry'->>'id'=$1`, entryID.String()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestManualLedger_ConcurrentPostAndReversePreserveIdentityAndProjectExactlyOnce(t *testing.T) {
	mod, _, iam, db, entityID, creatorID, ctx := setupAccountingTest(t)
	defer db.Close()
	actor, err := iam.AuthService.SeedAdmin(ctx, "manual_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Posting actor", "Password@123")
	if err != nil {
		t.Fatal(err)
	}
	draft := manualDraft(t, ctx, mod, entityID, creatorID)
	beforeCash := manualBalance(t, ctx, db, draft.Lines[0].AccountID)
	beforeRevenue := manualBalance(t, ctx, db, draft.Lines[1].AccountID)
	const attempts = 8
	for _, reverse := range []bool{false, true} {
		results := make(chan error, attempts)
		var workers sync.WaitGroup
		for range attempts {
			workers.Go(func() {
				if reverse {
					_, err := mod.LedgerService.ReverseJournalEntry(ctx, draft.ID, "Approved manual correction", actor.ID)
					results <- err
				} else {
					_, err := mod.LedgerService.PostJournalEntry(ctx, draft.ID, actor.ID)
					results <- err
				}
			})
		}
		workers.Wait()
		close(results)
		successes := 0
		expected := domain.ErrAlreadyPosted
		if reverse {
			expected = domain.ErrAlreadyReversed
		}
		for err := range results {
			if err == nil {
				successes++
			} else if !errors.Is(err, expected) {
				t.Errorf("unexpected retry failure: %v", err)
			}
		}
		if successes != 1 {
			t.Fatalf("reverse=%v: expected one success, got %d", reverse, successes)
		}
		saved, err := mod.LedgerService.GetJournalEntryByID(ctx, draft.ID)
		if err != nil {
			t.Fatal(err)
		}
		if saved.ID != draft.ID || saved.EntryNumber != draft.EntryNumber || saved.CreatedBy != creatorID || saved.PostedBy == nil || *saved.PostedBy != actor.ID || saved.Lines[0].ID != draft.Lines[0].ID || saved.Lines[1].ID != draft.Lines[1].ID {
			t.Fatal("manual posting replaced draft identity, lines, creator or posting actor")
		}
		delta := decimal.NewFromInt(123)
		auditCount := 1
		if reverse {
			delta = decimal.Zero
			auditCount = 2
			if saved.Status != domain.JournalEntryStatusReversed || saved.ReversalOfID == nil {
				t.Fatal("original not reversed")
			}
			reversal, err := mod.LedgerService.GetJournalEntryByID(ctx, *saved.ReversalOfID)
			if err != nil {
				t.Fatal(err)
			}
			if reversal.PostedBy == nil || *reversal.PostedBy != actor.ID {
				t.Fatal("reversal lost authenticated actor")
			}
		}
		if !manualBalance(t, ctx, db, draft.Lines[0].AccountID).Equal(beforeCash.Add(delta)) || !manualBalance(t, ctx, db, draft.Lines[1].AccountID).Equal(beforeRevenue.Sub(delta)) {
			t.Fatal("manual retries duplicated or omitted balance projections")
		}
		if manualAuditCount(t, ctx, db, draft.ID) != auditCount {
			t.Fatal("manual retries duplicated or omitted financial audit")
		}
	}
	audit := auditservices.NewAuditService(auditpostgres.NewAuditRepository(db.Pool))
	check, err := audit.VerifyChainIntegrity(ctx)
	if err != nil || !check.IsValid {
		t.Fatalf("audit chain: %+v %v", check, err)
	}
}

func TestManualLedger_PostAndReverseFailuresRollBackBalancesAndAudit(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
	for _, failure := range []string{"balance", "audit"} {
		t.Run(failure, func(t *testing.T) {
			draft := manualDraft(t, ctx, mod, entityID, actorID)
			repo := postgres.NewPostingRepository(db.Pool)
			var postingRepo ports.PostingRepository = repo
			var audit ports.PostingAuditRecorder = auditservices.NewAuditService(auditpostgres.NewAuditRepository(db.Pool))
			expected := errAuditWrite
			if failure == "balance" {
				postingRepo = &failingBalanceRepository{PostingRepository: repo, failAccount: draft.Lines[1].AccountID}
				expected = errBalanceWrite
			} else {
				audit = failingAuditRecorder{audit}
			}
			ledger := services.NewLedgerService(repo, posting.NewPostingEngine(postingRepo, transaction.NewManager(db.Pool), audit))
			beforeCash := manualBalance(t, ctx, db, draft.Lines[0].AccountID)
			beforeRevenue := manualBalance(t, ctx, db, draft.Lines[1].AccountID)
			if _, err := ledger.PostJournalEntry(ctx, draft.ID, actorID); !errors.Is(err, expected) {
				t.Fatalf("expected failure: %v", err)
			}
			saved, err := mod.LedgerService.GetJournalEntryByID(ctx, draft.ID)
			if err != nil {
				t.Fatal(err)
			}
			if saved.Status != domain.JournalEntryStatusDraft || saved.PostedBy != nil || saved.Version != draft.Version || manualAuditCount(t, ctx, db, draft.ID) != 0 ||
				!manualBalance(t, ctx, db, draft.Lines[0].AccountID).Equal(beforeCash) || !manualBalance(t, ctx, db, draft.Lines[1].AccountID).Equal(beforeRevenue) {
				t.Fatal("failed manual post left state, balance or audit writes")
			}
			posted, err := mod.LedgerService.PostJournalEntry(ctx, draft.ID, actorID)
			if err != nil {
				t.Fatal(err)
			}
			beforeCash = manualBalance(t, ctx, db, draft.Lines[0].AccountID)
			beforeRevenue = manualBalance(t, ctx, db, draft.Lines[1].AccountID)
			if _, err := ledger.ReverseJournalEntry(ctx, draft.ID, "Rollback correction", actorID); !errors.Is(err, expected) {
				t.Fatalf("expected reversal failure: %v", err)
			}
			saved, err = mod.LedgerService.GetJournalEntryByID(ctx, draft.ID)
			if err != nil {
				t.Fatal(err)
			}
			var reversals int
			if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM accounting.journal_entries WHERE legal_entity_id=$1 AND entry_number=$2", entityID, "REV-"+draft.EntryNumber).Scan(&reversals); err != nil {
				t.Fatal(err)
			}
			if saved.Status != domain.JournalEntryStatusPosted || saved.ReversalOfID != nil || saved.Version != posted.Version || reversals != 0 || manualAuditCount(t, ctx, db, draft.ID) != 1 ||
				!manualBalance(t, ctx, db, draft.Lines[0].AccountID).Equal(beforeCash) || !manualBalance(t, ctx, db, draft.Lines[1].AccountID).Equal(beforeRevenue) {
				t.Fatal("failed manual reversal left state, balance or audit writes")
			}
		})
	}
}

func TestManualLedger_RevalidatesSavedDraftBeforePosting(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
	for _, test := range []struct {
		name    string
		update  string
		restore string
		target  func(*domain.JournalEntry) uuid.UUID
		want    error
	}{
		{"inactive account", "UPDATE accounting.accounts SET is_active=false WHERE id=$1", "UPDATE accounting.accounts SET is_active=true WHERE id=$1", func(e *domain.JournalEntry) uuid.UUID { return e.Lines[0].AccountID }, domain.ErrAccountInactive},
		{"nonposting account", "UPDATE accounting.accounts SET allow_posting=false WHERE id=$1", "UPDATE accounting.accounts SET allow_posting=true WHERE id=$1", func(e *domain.JournalEntry) uuid.UUID { return e.Lines[0].AccountID }, domain.ErrAccountNotPostable},
		{"inactive journal", "UPDATE accounting.journals SET is_active=false WHERE id=$1", "UPDATE accounting.journals SET is_active=true WHERE id=$1", func(e *domain.JournalEntry) uuid.UUID { return e.JournalID }, domain.ErrJournalNotFound},
		{"closed period", "UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", "UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1", func(e *domain.JournalEntry) uuid.UUID { return e.FiscalPeriodID }, domain.ErrPeriodNotOpen},
		{"wrong date period", "UPDATE accounting.journal_entries SET accounting_date=accounting_date + interval '1 month' WHERE id=$1", "", func(e *domain.JournalEntry) uuid.UUID { return e.ID }, domain.ErrFiscalPeriodNotFound},
		{"cancelled draft", "UPDATE accounting.journal_entries SET status='CANCELLED' WHERE id=$1", "", func(e *domain.JournalEntry) uuid.UUID { return e.ID }, domain.ErrInvalidPostingRequest},
		{"header totals", "UPDATE accounting.journal_entries SET total_debit=124,total_credit=124 WHERE id=$1", "", func(e *domain.JournalEntry) uuid.UUID { return e.ID }, domain.ErrUnbalancedJournalEntry},
		{"invalid header rate", "UPDATE accounting.journal_entries SET exchange_rate=0 WHERE id=$1", "", func(e *domain.JournalEntry) uuid.UUID { return e.ID }, domain.ErrInvalidPostingRequest},
		{"invalid line rate", "UPDATE accounting.journal_lines SET exchange_rate=0 WHERE journal_entry_id=$1", "", func(e *domain.JournalEntry) uuid.UUID { return e.ID }, domain.ErrInvalidPostingRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			draft := manualDraft(t, ctx, mod, entityID, actorID)
			target := test.target(draft)
			if _, err := db.Pool.Exec(ctx, test.update, target); err != nil {
				t.Fatal(err)
			}
			if test.restore != "" {
				defer func() {
					if _, err := db.Pool.Exec(ctx, test.restore, target); err != nil {
						t.Error(err)
					}
				}()
			}
			before := manualBalance(t, ctx, db, draft.Lines[0].AccountID)
			if _, err := mod.LedgerService.PostJournalEntry(ctx, draft.ID, actorID); !errors.Is(err, test.want) {
				t.Fatalf("wanted %v; got %v", test.want, err)
			}
			if manualAuditCount(t, ctx, db, draft.ID) != 0 || !manualBalance(t, ctx, db, draft.Lines[0].AccountID).Equal(before) {
				t.Fatal("invalid draft left projections or audit")
			}
			var posted bool
			if err := db.Pool.QueryRow(ctx, "SELECT posted_at IS NOT NULL FROM accounting.journal_entries WHERE id=$1", draft.ID).Scan(&posted); err != nil {
				t.Fatal(err)
			}
			if posted {
				t.Fatal("invalid draft was posted")
			}
		})
	}
	draft := manualDraft(t, ctx, mod, entityID, actorID)
	if _, err := mod.LedgerService.PostJournalEntry(ctx, draft.ID, uuid.Nil); !errors.Is(err, domain.ErrInvalidPostingRequest) {
		t.Fatalf("missing actor accepted: %v", err)
	}
}

func TestManualLedger_RejectsCrossEntityJournalAndPeriod(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
	otherID := uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organization.legal_entities(id, organization_id, code, name_ar, name_en)
        SELECT $1, organization_id, $2, 'منشأة أخرى', 'Other entity' FROM organization.legal_entities WHERE id=$3`, otherID, "MANUAL-"+otherID.String(), entityID); err != nil {
		t.Fatal(err)
	}
	if err := mod.SeedDefaultCOA(ctx, otherID); err != nil {
		t.Fatal(err)
	}
	journals, err := mod.AccountingService.ListJournals(ctx, otherID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("foreign journals: %v", err)
	}
	period, err := mod.AccountingService.GetFiscalPeriodByDate(ctx, otherID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		column string
		id     uuid.UUID
		want   error
	}{
		{"journal_id", journals[0].ID, domain.ErrJournalNotFound},
		{"fiscal_period_id", period.ID, domain.ErrFiscalPeriodNotFound},
	} {
		t.Run(test.column, func(t *testing.T) {
			draft := manualDraft(t, ctx, mod, entityID, actorID)
			// Legacy header foreign keys reference IDs independently; the posting
			// boundary must reject an inconsistent entity even for a saved draft.
			if _, err := db.Pool.Exec(ctx, "UPDATE accounting.journal_entries SET "+test.column+"=$1 WHERE id=$2", test.id, draft.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := mod.LedgerService.PostJournalEntry(ctx, draft.ID, actorID); !errors.Is(err, test.want) {
				t.Fatalf("cross-entity %s: %v", test.column, err)
			}
			if manualAuditCount(t, ctx, db, draft.ID) != 0 {
				t.Fatal("foreign draft produced posting audit")
			}
		})
	}
}

// Pause after validation but before the status update so another connection can
// try changing the reference data used to accept this saved draft.
type pausedManualPostRepository struct {
	ports.PostingRepository
	validated chan struct{}
	resume    chan struct{}
}

func (r pausedManualPostRepository) PostJournalEntry(ctx context.Context, id, actorID uuid.UUID) (*domain.JournalEntry, error) {
	close(r.validated)
	select {
	case <-r.resume:
		return r.PostingRepository.PostJournalEntry(ctx, id, actorID)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestManualLedger_ValidationLocksLastThroughPosting(t *testing.T) {
	mod, _, _, db, entityID, actorID, setupCtx := setupAccountingTest(t)
	defer db.Close()
	ctx, cancel := context.WithTimeout(setupCtx, 10*time.Second)
	defer cancel()
	draft := manualDraft(t, ctx, mod, entityID, actorID)
	repo := pausedManualPostRepository{postgres.NewPostingRepository(db.Pool), make(chan struct{}), make(chan struct{})}
	audit := auditservices.NewAuditService(auditpostgres.NewAuditRepository(db.Pool))
	engine := posting.NewPostingEngine(repo, transaction.NewManager(db.Pool), audit)
	completed := make(chan error, 1)
	go func() { _, err := engine.PostDraft(ctx, draft.ID, actorID); completed <- err }()
	select {
	case <-repo.validated:
	case err := <-completed:
		t.Fatalf("posting ended before validation barrier: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	for _, update := range []struct {
		sql string
		id  uuid.UUID
	}{
		{"UPDATE accounting.accounts SET is_active=false WHERE id=$1", draft.Lines[0].AccountID},
		{"UPDATE accounting.journals SET is_active=false WHERE id=$1", draft.JournalID},
		{"UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", draft.FiscalPeriodID},
	} {
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, "SET LOCAL lock_timeout='100ms'"); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, update.sql, update.id)
		_ = tx.Rollback(ctx)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "55P03" {
			t.Errorf("posting reference was not locked: %s: %v", update.sql, err)
		}
	}
	close(repo.resume)
	select {
	case err := <-completed:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
