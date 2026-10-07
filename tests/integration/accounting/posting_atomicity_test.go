package accounting_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"erp/internal/accounting/application/posting"
	"erp/internal/accounting/domain"
	"erp/internal/accounting/infrastructure/postgres"
	"erp/internal/accounting/ports"
	auditservices "erp/internal/audit/application/services"
	auditpostgres "erp/internal/audit/infrastructure/postgres"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var errBalanceWrite = errors.New("injected balance projection failure")

type failingBalanceRepository struct {
	ports.PostingRepository
	failAccount uuid.UUID
}

func (r *failingBalanceRepository) UpsertAccountBalance(ctx context.Context, balance domain.AccountBalance) (*domain.AccountBalance, error) {
	if balance.AccountID == r.failAccount {
		return nil, errBalanceWrite
	}
	return r.PostingRepository.UpsertAccountBalance(ctx, balance)
}

func TestPostingEngine_BalanceFailureRollsBackPostAndReversal(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
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
		LegalEntityID: entityID, SourceModule: "TEST", SourceType: "ATOMIC", SourceID: uuid.New(),
		JournalID: journals[0].ID, AccountingDate: time.Now().UTC(), ActorID: actorID,
		Lines: []domain.PostingLine{
			{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(71)},
			{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(71)},
		},
	}
	engine := posting.NewPostingEngine(&failingBalanceRepository{
		PostingRepository: postgres.NewPostingRepository(db.Pool), failAccount: revenue.ID,
	}, transaction.NewManager(db.Pool), auditservices.NewAuditService(auditpostgres.NewAuditRepository(db.Pool)))
	readCashBalance := func() decimal.Decimal {
		t.Helper()
		var result decimal.Decimal
		if err := db.Pool.QueryRow(ctx, "SELECT COALESCE(sum(closing_balance),0) FROM accounting.account_balances WHERE account_id=$1", cash.ID).Scan(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	before := readCashBalance()
	if _, err := engine.Post(ctx, request); !errors.Is(err, errBalanceWrite) {
		t.Fatalf("expected injected failure, got %v", err)
	}
	var count int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM accounting.journal_entries WHERE source_id=$1", request.SourceID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 || !readCashBalance().Equal(before) {
		t.Fatal("failed posting left a journal or partially updated balances")
	}
	posted, err := mod.PostingEngine.Post(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	before = readCashBalance()
	if _, err := engine.Reverse(ctx, posted.JournalEntryID, "Rollback test", actorID); !errors.Is(err, errBalanceWrite) {
		t.Fatalf("expected reversal failure, got %v", err)
	}
	original, err := mod.LedgerService.GetJournalEntryByID(ctx, posted.JournalEntryID)
	if err != nil {
		t.Fatal(err)
	}
	if original.Status != domain.JournalEntryStatusPosted || original.ReversalOfID != nil || !readCashBalance().Equal(before) {
		t.Fatal("failed reversal changed the original journal or balances")
	}
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM accounting.journal_entries WHERE entry_number=$1", "REV-"+posted.EntryNumber).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("failed reversal left a reversal journal")
	}
}

func TestPostingEngine_ConcurrentSourcePostsExactlyOnce(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
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
		LegalEntityID: entityID, SourceModule: "TEST", SourceType: "CONCURRENT", SourceID: uuid.New(),
		JournalID: journals[0].ID, AccountingDate: time.Now().UTC(), ActorID: actorID,
		Lines: []domain.PostingLine{
			{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(73)},
			{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(73)},
		},
	}
	const attempts = 8
	results := make(chan error, attempts)
	var workers sync.WaitGroup
	for range attempts {
		workers.Go(func() {
			_, err := mod.PostingEngine.Post(ctx, request)
			results <- err
		})
	}
	workers.Wait()
	close(results)
	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
		} else if !errors.Is(err, domain.ErrDuplicatePosting) {
			t.Errorf("unexpected posting failure: %v", err)
		}
	}
	if succeeded != 1 {
		t.Fatalf("expected one successful posting, got %d", succeeded)
	}
}
