package accounting_test

import (
	"testing"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/platform/config"
	"erp/internal/reporting"
	reportdomain "erp/internal/reporting/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestReversalReportsPreserveOriginalAndExcludeDrafts(t *testing.T) {
	mod, _, _, db, entityID, actorID, ctx := setupAccountingTest(t)
	defer db.Close()
	reports := reporting.NewModule(&config.Config{}, db)
	cash, err := mod.AccountingService.GetAccountByCode(ctx, entityID, "111002")
	if err != nil {
		t.Fatal(err)
	}
	revenue, err := mod.AccountingService.GetAccountByCode(ctx, entityID, "410000")
	if err != nil {
		t.Fatal(err)
	}
	journals, err := mod.AccountingService.ListJournals(ctx, entityID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("journals: %v", err)
	}
	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)
	filter := reportdomain.ReportFilter{LegalEntityID: entityID}
	read := func(f reportdomain.ReportFilter) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
		t.Helper()
		tb, err := reports.Service.GetTrialBalance(ctx, f)
		if err != nil {
			t.Fatal(err)
		}
		var cashBalance decimal.Decimal
		for _, row := range tb.Rows {
			if row.AccountID == cash.ID {
				cashBalance = row.NetBalance
			}
		}
		bs, err := reports.Service.GetBalanceSheet(ctx, f)
		if err != nil {
			t.Fatal(err)
		}
		income, err := reports.Service.GetIncomeStatement(ctx, f)
		if err != nil {
			t.Fatal(err)
		}
		return cashBalance, bs.TotalAssets, income.TotalRevenue
	}
	beforeCash, beforeAssets, beforeRevenue := read(filter)
	historical := filter
	historical.EndDate, historical.AsOfDate = &yesterday, &yesterday
	oldCash, oldAssets, oldRevenue := read(historical)
	var viewBefore decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT net_balance FROM reporting.trial_balance WHERE account_id=$1", cash.ID).Scan(&viewBefore); err != nil {
		t.Fatal(err)
	}
	// A saved draft must never leak into the SQL reporting views.
	manualDraft(t, ctx, mod, entityID, actorID)
	amount := decimal.NewFromInt(100)
	posted, err := mod.PostingEngine.Post(ctx, domain.PostingRequest{
		LegalEntityID: entityID, SourceModule: "TEST", SourceType: "REPORT_REVERSAL", SourceID: uuid.New(),
		JournalID: journals[0].ID, AccountingDate: yesterday, ActorID: actorID,
		Lines: []domain.PostingLine{{AccountID: cash.ID, TransactionDebit: amount}, {AccountID: revenue.ID, TransactionCredit: amount}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertTotals := func(f reportdomain.ReportFilter, cashWant, assetsWant, revenueWant decimal.Decimal) {
		t.Helper()
		cashGot, assetsGot, revenueGot := read(f)
		if !cashGot.Equal(cashWant) || !assetsGot.Equal(assetsWant) || !revenueGot.Equal(revenueWant) {
			t.Fatalf("cash/assets/revenue=%s/%s/%s, want %s/%s/%s", cashGot, assetsGot, revenueGot, cashWant, assetsWant, revenueWant)
		}
	}
	assertTotals(filter, beforeCash.Add(amount), beforeAssets.Add(amount), beforeRevenue.Add(amount))
	reversed, err := mod.PostingEngine.Reverse(ctx, posted.JournalEntryID, "Reporting regression", actorID)
	if err != nil {
		t.Fatal(err)
	}
	assertTotals(filter, beforeCash, beforeAssets, beforeRevenue)
	assertTotals(historical, oldCash.Add(amount), oldAssets.Add(amount), oldRevenue.Add(amount))
	filter.AccountID = &cash.ID
	gl, err := reports.Service.GetGeneralLedger(ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	var seenOriginal, seenReversal bool
	for _, entry := range gl.Entries {
		seenOriginal = seenOriginal || entry.EntryID == posted.JournalEntryID
		seenReversal = seenReversal || entry.EntryID == reversed.JournalEntryID
	}
	if !seenOriginal || !seenReversal {
		t.Fatal("general ledger must retain original and compensating entries")
	}
	var viewAfter decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT net_balance FROM reporting.trial_balance WHERE account_id=$1", cash.ID).Scan(&viewAfter); err != nil {
		t.Fatal(err)
	}
	if !viewAfter.Equal(viewBefore) {
		t.Fatalf("view included draft or lost reversed original: %s != %s", viewAfter, viewBefore)
	}
	var count int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM reporting.general_ledger WHERE entry_id IN ($1,$2)", posted.JournalEntryID, reversed.JournalEntryID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("view has %d lines, want original and reversal's four lines", count)
	}
}
