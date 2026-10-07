package accounting_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"erp/internal/accounting/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestJournalImmutabilityLocksPostingAndBothLineParents(t *testing.T) {
	mod, _, _, db, entityID, actorID, parentCtx := setupAccountingTest(t)
	defer db.Close()
	ctx, cancel := context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()
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
	draft := func() *domain.JournalEntry {
		t.Helper()
		entry, err := mod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
			LegalEntityID: entityID, JournalID: journals[0].ID, AccountingDate: time.Now().UTC(),
			CreatedBy: actorID, Description: "Concurrent immutability test", EntryNumber: "LOCK-" + uuid.NewString(),
		}, []domain.JournalLine{
			{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(10)},
			{AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(10)},
		})
		if err != nil {
			t.Fatal(err)
		}
		return entry
	}
	posted, other := draft(), draft()
	postingTx, err := db.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer postingTx.Rollback(context.Background())
	if _, err := postingTx.Exec(ctx, "UPDATE accounting.journal_entries SET status='POSTED', posted_at=now(), posted_by=$2 WHERE id=$1", posted.ID, actorID); err != nil {
		t.Fatal(err)
	}
	editor, err := db.Pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer editor.Release()
	var editorPID int
	if err := editor.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&editorPID); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() {
		_, err := editor.Exec(ctx, "UPDATE accounting.journal_lines SET description='forbidden concurrent edit' WHERE journal_entry_id=$1 AND line_number=1", posted.ID)
		result <- err
	}()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	waiting := false
	for !waiting {
		select {
		case err := <-result:
			t.Fatalf("line write did not wait for posting lock: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
			if err := db.Pool.QueryRow(ctx, "SELECT cardinality(pg_blocking_pids($1)) > 0", editorPID).Scan(&waiting); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := postingTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "Cannot UPDATE journal lines") {
			t.Fatalf("edit survived concurrent posting: %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	// A line cannot be moved from an editable parent into a posted parent.
	_, err = db.Pool.Exec(ctx, "UPDATE accounting.journal_lines SET journal_entry_id=$1, line_number=3 WHERE journal_entry_id=$2 AND line_number=1", posted.ID, other.ID)
	if err == nil || !strings.Contains(err.Error(), "Cannot UPDATE journal lines") {
		t.Fatalf("line moved into posted journal: %v", err)
	}
	// Changing status to REVERSED is not an escape hatch to change posted money.
	_, err = db.Pool.Exec(ctx, `UPDATE accounting.journal_entries SET status='REVERSED', reversed_at=now(), reversal_of_id=$2,
 version=version+1, total_debit=11, total_credit=11 WHERE id=$1`, posted.ID, other.ID)
	if err == nil || !strings.Contains(err.Error(), "Cannot UPDATE posted") {
		t.Fatalf("reversal rewrote original amounts: %v", err)
	}
	var description string
	if err := db.Pool.QueryRow(ctx, "SELECT coalesce(description,'') FROM accounting.journal_lines WHERE journal_entry_id=$1 AND line_number=1", posted.ID).Scan(&description); err != nil {
		t.Fatal(err)
	}
	if description == "forbidden concurrent edit" {
		t.Fatal("posted line changed")
	}
}
