package fixedassets_test

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"testing"
)

func TestDepreciationPostingRollbackAndHistory(t *testing.T) {
	mod, _, db, entity, maker, _, ctx := setupFixedAssetsTest(t)
	defer db.Close()
	books, err := mod.Service.ListDepreciationBooks(ctx, entity)
	if err != nil {
		t.Fatal(err)
	}
	var book uuid.UUID
	for _, b := range books {
		if b.IsPrimary {
			book = b.ID
		}
	}
	run, lines, err := mod.Service.RunMonthlyDepreciation(ctx, entity, book, 2026, 11, &maker)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.Service.PostDepreciationRun(ctx, run.ID, &maker); err == nil {
		t.Fatal("creator posted depreciation")
	}
	checker := assetChecker(t, db, ctx)
	var period uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id=$1 AND $2 BETWEEN start_date AND end_date", entity, run.RunDate).Scan(&period); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", period); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.Service.PostDepreciationRun(ctx, run.ID, checker); err == nil {
		t.Fatal("depreciation posted into closed period")
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1", period); err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		asset, err := mod.Service.GetAsset(ctx, line.AssetID)
		if err != nil {
			t.Fatal(err)
		}
		if !asset.AccumulatedDepreciation.Equal(line.AccumulatedBefore) {
			t.Fatal("failed posting changed carrying value")
		}
	}
	posted, err := mod.Service.PostDepreciationRun(ctx, run.ID, checker)
	if err != nil {
		t.Fatal(err)
	}
	if posted.JournalEntryID == nil {
		t.Fatal("depreciation journal missing")
	}
	var debit, credit decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT total_debit,total_credit FROM accounting.journal_entries WHERE id=$1", posted.JournalEntryID).Scan(&debit, &credit); err != nil {
		t.Fatal(err)
	}
	if !debit.Equal(run.TotalDepreciationAmount) || !credit.Equal(debit) {
		t.Fatal("depreciation GL differs from subledger")
	}
	if _, _, err := mod.Service.RunMonthlyDepreciation(ctx, entity, book, 2026, 11, &maker); err == nil {
		t.Fatal("recalculation overwrote a posted run")
	}
	current, _, err := mod.Service.GetDepreciationRun(ctx, run.ID)
	if err != nil || current.Status != "POSTED" || current.JournalEntryID == nil {
		t.Fatal("posted depreciation history changed")
	}
}
