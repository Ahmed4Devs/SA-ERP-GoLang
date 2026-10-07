package accounting_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"erp/internal/accounting/application/posting"
	"erp/internal/accounting/application/services"
	"erp/internal/accounting/domain"
	"erp/internal/audit"
	"erp/internal/platform/config"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestCanonicalReferences_RatesMappingsAndEntityIsolation(t *testing.T) {
	mod, _, _, db, entity, actor, ctx := setupAccountingTest(t)
	defer db.Close()
	cash, _ := mod.AccountingService.GetAccountByCode(ctx, entity, "111002")
	revenue, _ := mod.AccountingService.GetAccountByCode(ctx, entity, "410000")
	cmd := services.ExchangeRateCommand{FromCurrency: "USD", ToCurrency: "SAR", EffectiveDate: "2026-01-01", Rate: decimal.RequireFromString("3.750000123456"), Source: "Treasury quotation"}
	rateID, err := mod.ReferenceService.CreateExchangeRate(ctx, entity, actor, cmd)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Rate = decimal.RequireFromString("3.760000123456")
	if _, err = mod.ReferenceService.CreateExchangeRate(ctx, entity, actor, cmd); err != nil {
		t.Fatal(err)
	}
	var value string
	if err = db.Pool.QueryRow(ctx, `SELECT rate::text FROM accounting.exchange_rates WHERE id=$1`, rateID).Scan(&value); err != nil || value != "3.750000123456" {
		t.Fatalf("historic rate changed: %s %v", value, err)
	}
	if _, err = db.Pool.Exec(ctx, `UPDATE accounting.exchange_rates SET rate=4 WHERE id=$1`, rateID); err == nil {
		t.Fatal("exchange-rate history was mutable")
	}
	data, total, err := mod.ReferenceService.List(ctx, entity, "exchange-rates", 1, 1)
	if err != nil || total != 2 {
		t.Fatalf("pagination count %d: %v", total, err)
	}
	var records []map[string]any
	if err = json.Unmarshal(data, &records); err != nil || len(records) != 1 {
		t.Fatalf("page length: %s %v", data, err)
	}
	_, total, err = mod.ReferenceService.List(ctx, uuid.New(), "exchange-rates", 50, 0)
	if err != nil || total != 0 {
		t.Fatalf("scope leaked: %d %v", total, err)
	}
	rule := services.PostingRuleCommand{SourceModule: "PAYABLES", Code: "INPUT_VAT", NameEn: "VAT account", EffectiveFrom: "2026-01-01", AccountID: cash.ID}
	if _, err = mod.ReferenceService.CreatePostingRule(ctx, entity, actor, rule); err != nil {
		t.Fatal(err)
	}
	rule.EffectiveFrom = "2026-07-01"
	rule.AccountID = revenue.ID
	if _, err = mod.ReferenceService.CreatePostingRule(ctx, entity, actor, rule); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	ledger := posting.NewSourceService(db, mod.PostingEngine, audit.NewModule(cfg, db).Service)
	for _, example := range []struct {
		date    string
		account uuid.UUID
	}{{"2026-06-30", cash.ID}, {"2026-07-01", revenue.ID}} {
		date, _ := time.Parse(time.DateOnly, example.date)
		if err = ledger.WithinTx(ctx, func(ctx context.Context) error {
			id, err := ledger.AccountForDate(ctx, entity, nil, "PAYABLES", "INPUT_VAT", "114000", date)
			if err == nil && id != example.account {
				t.Fatalf("effective account: %s", id)
			}
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	other := uuid.New()
	if _, err = db.Pool.Exec(ctx, `INSERT INTO organization.legal_entities(id,organization_id,code,name_ar,name_en) SELECT $1,organization_id,$2,'أخرى','Other' FROM organization.legal_entities WHERE id=$3`, other, other.String(), entity); err != nil {
		t.Fatal(err)
	}
	if _, err = mod.ReferenceService.CreatePostingRule(ctx, other, actor, rule); err == nil {
		t.Fatal("cross-entity posting mapping accepted")
	}
	dim, err := mod.ReferenceService.CreateDimension(ctx, entity, actor, services.DimensionCommand{Code: uuid.NewString(), NameEn: "Project"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = mod.ReferenceService.CreateDimension(ctx, other, actor, services.DimensionCommand{DimensionID: dim, Code: "X", NameEn: "Wrong company"}, true); err == nil {
		t.Fatal("cross-entity dimension value accepted")
	}
}

func TestCanonicalReferences_OpeningBalanceReconciliationAndPrecision(t *testing.T) {
	mod, _, iamMod, db, entity, actor, ctx := setupAccountingTest(t)
	defer db.Close()
	checker, err := iamMod.AuthService.SeedAdmin(ctx, "ref_"+uuid.NewString()[:8], uuid.NewString()+"@test.local", "Checker", "Password@123")
	if err != nil {
		t.Fatal(err)
	}
	cash, _ := mod.AccountingService.GetAccountByCode(ctx, entity, "111002")
	capital, _ := mod.AccountingService.GetAccountByCode(ctx, entity, "310000")
	period, err := mod.AccountingService.GetFiscalPeriodByDate(ctx, entity, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	amount := decimal.RequireFromString("987.123456")
	id, err := mod.ReferenceService.CreateOpeningBalance(ctx, entity, actor, services.OpeningBalanceCommand{FiscalPeriodID: period.ID, AccountID: cash.ID, CurrencyCode: "SAR", Debit: amount})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = mod.ReferenceService.PostOpeningBalance(ctx, entity, actor, id, capital.ID, decimal.NewFromInt(1)); err == nil {
		t.Fatal("creator posted own opening balance")
	}
	journal, err := mod.ReferenceService.PostOpeningBalance(ctx, entity, checker.ID, id, capital.ID, decimal.NewFromInt(1))
	if err != nil {
		t.Fatal(err)
	}
	var debit string
	if err = db.Pool.QueryRow(ctx, `SELECT total_debit::text FROM accounting.journal_entries WHERE id=$1`, journal).Scan(&debit); err != nil || debit != amount.String() {
		t.Fatalf("six decimals lost: %s %v", debit, err)
	}
	if _, err = db.Pool.Exec(ctx, `UPDATE accounting.opening_balances SET debit=1 WHERE id=$1`, id); err == nil {
		t.Fatal("posted opening balance mutable")
	}
	reversal, err := mod.PostingEngine.Reverse(ctx, journal, "Opening correction", checker.ID)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Pool.Query(ctx, `SELECT id FROM accounting.journal_lines WHERE journal_entry_id=ANY($1::uuid[]) AND account_id=$2`, []uuid.UUID{journal, reversal.JournalEntryID}, cash.ID)
	if err != nil {
		t.Fatal(err)
	}
	var lines []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		lines = append(lines, id)
	}
	rows.Close()
	reconciliation, err := mod.ReferenceService.CreateReconciliation(ctx, entity, actor, services.ReconciliationCommand{AccountID: cash.ID, Reference: uuid.NewString(), ReconciliationDate: "2026-12-31", CurrencyCode: "SAR", JournalLineIDs: lines})
	if err != nil {
		t.Fatal(err)
	}
	if err = mod.ReferenceService.CompleteReconciliation(ctx, entity, actor, reconciliation); err == nil {
		t.Fatal("creator completed own reconciliation")
	}
	if err = mod.ReferenceService.CompleteReconciliation(ctx, entity, checker.ID, reconciliation); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Pool.Exec(ctx, `DELETE FROM accounting.reconciliation_lines WHERE reconciliation_id=$1`, reconciliation); err == nil {
		t.Fatal("completed reconciliation lines mutable")
	}
	journals, _ := mod.AccountingService.ListJournals(ctx, entity)
	request := domain.PostingRequest{LegalEntityID: entity, SourceModule: "TEST", SourceType: "PRECISION", SourceID: uuid.New(), JournalID: journals[0].ID, AccountingDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), CurrencyCode: "SAR", ExchangeRate: decimal.RequireFromString("1.000000123456"), ActorID: actor, Lines: []domain.PostingLine{{AccountID: cash.ID, TransactionDebit: amount}, {AccountID: capital.ID, TransactionCredit: amount}}}
	posted, err := mod.PostingEngine.Post(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := mod.LedgerService.GetJournalEntryByID(ctx, posted.JournalEntryID)
	if err != nil {
		t.Fatal(err)
	}
	if !entry.TotalDebit.Equal(amount.Mul(request.ExchangeRate).Round(6)) {
		t.Fatalf("FX precision mismatch: %s", entry.TotalDebit)
	}
	request.SourceID = uuid.New()
	request.Lines[0].TransactionDebit = decimal.RequireFromString("0.0000001")
	request.Lines[1].TransactionCredit = request.Lines[0].TransactionDebit
	if _, err = mod.PostingEngine.Post(ctx, request); err == nil {
		t.Fatal("sub-micro amounts silently rounded")
	}
	var badTypes int
	if err = db.Pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema NOT IN ('pg_catalog','information_schema') AND data_type='numeric' AND NOT ((numeric_precision=24 AND numeric_scale IN (6,12)) OR(numeric_precision=12 AND numeric_scale=8) OR(table_schema='reporting' AND numeric_precision IS NULL))`).Scan(&badTypes); err != nil || badTypes != 0 {
		t.Fatalf("noncanonical stored numeric types: %d %v", badTypes, err)
	}
}

func TestCanonicalReferences_DimensionsBecomeImmutableWithJournal(t *testing.T) {
	mod, _, _, db, entity, actor, ctx := setupAccountingTest(t)
	defer db.Close()
	dimension, err := mod.ReferenceService.CreateDimension(ctx, entity, actor, services.DimensionCommand{Code: uuid.NewString(), NameEn: "Contract"}, false)
	if err != nil {
		t.Fatal(err)
	}
	value, err := mod.ReferenceService.CreateDimension(ctx, entity, actor, services.DimensionCommand{DimensionID: dimension, Code: uuid.NewString(), NameEn: "Contract value"}, true)
	if err != nil {
		t.Fatal(err)
	}
	cash, _ := mod.AccountingService.GetAccountByCode(ctx, entity, "111002")
	revenue, _ := mod.AccountingService.GetAccountByCode(ctx, entity, "410000")
	journals, _ := mod.AccountingService.ListJournals(ctx, entity)
	entry, err := mod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{LegalEntityID: entity, JournalID: journals[0].ID, AccountingDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), CreatedBy: actor, Description: "Custom dimensions"}, []domain.JournalLine{{AccountID: cash.ID, TransactionDebit: decimal.NewFromInt(10)}, {AccountID: revenue.ID, TransactionCredit: decimal.NewFromInt(10)}})
	if err != nil {
		t.Fatal(err)
	}
	var line uuid.UUID
	if err = db.Pool.QueryRow(ctx, `SELECT id FROM accounting.journal_lines WHERE journal_entry_id=$1 AND line_number=1`, entry.ID).Scan(&line); err != nil {
		t.Fatal(err)
	}
	if err = mod.ReferenceService.AssignDimension(ctx, entity, actor, line, dimension, value); err != nil {
		t.Fatal(err)
	}
	if _, err = mod.PostingEngine.PostDraft(ctx, entry.ID, actor); err != nil {
		t.Fatal(err)
	}
	if err = mod.ReferenceService.AssignDimension(ctx, entity, actor, line, dimension, value); err == nil {
		t.Fatal("posted dimension assignment changed")
	}
	if _, err = db.Pool.Exec(ctx, `DELETE FROM accounting.journal_line_dimensions WHERE journal_line_id=$1`, line); err == nil {
		t.Fatal("posted dimension assignment deleted")
	}
	reversal, err := mod.PostingEngine.Reverse(ctx, entry.ID, "Dimension correction", actor)
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := mod.LedgerService.GetJournalEntryByID(ctx, reversal.JournalEntryID)
	if err != nil {
		t.Fatal(err)
	}
	if reversed.Lines[0].Dimensions[dimension] != value {
		t.Fatal("reversal lost custom dimensions")
	}
}
