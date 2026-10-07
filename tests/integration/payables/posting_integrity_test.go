package payables_test

import (
	"testing"
	"time"

	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSupplierPostingFXTaxAndRollback(t *testing.T) {
	mod, db, entity, actor, ap, _, expense, ctx := setupPayablesTest(t)
	defer db.Close()
	supplier, err := mod.SupplierService.CreateSupplier(ctx, ports.CreateSupplierCmd{LegalEntityID: entity, SupplierType: domain.SupplierTypeLocal, NameEn: uuid.NewString(), NameAr: "مورد اختبار", Address: domain.SaudiAddress{City: "Riyadh"}, DefaultLiabilityAccountID: &ap, DefaultExpenseAccountID: &expense})
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(time.Now().Year(), 2, 12, 0, 0, 0, 0, time.UTC)
	var code, fiscal, taxPeriod uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM tax.tax_codes WHERE legal_entity_id=$1 AND tax_type IN ('INPUT','BOTH') AND tax_category_code='S' ORDER BY code LIMIT 1", entity).Scan(&code); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id=$1 AND $2 BETWEEN start_date AND end_date", entity, date).Scan(&fiscal); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM tax.tax_periods WHERE legal_entity_id=$1 AND $2 BETWEEN start_date AND end_date", entity, date).Scan(&taxPeriod); err != nil {
		t.Fatal(err)
	}
	create := func() *domain.SupplierInvoice {
		t.Helper()
		invoice, err := mod.PayableService.CreateSupplierInvoice(ctx, ports.CreateSupplierInvoiceCmd{LegalEntityID: entity, SupplierID: supplier.ID, SupplierInvoiceNumber: uuid.NewString(), InvoiceDate: date.Format("2006-01-02"), CurrencyCode: "USD", ExchangeRate: decimal.RequireFromString("3.751234"), CreatedByUserID: &actor, Lines: []ports.CreateSupplierInvoiceLineCmd{
			{ExpenseAccountID: expense, Description: "first", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.RequireFromString("19.1234"), TaxCodeID: &code, TaxRate: decimal.NewFromInt(99)},
			{ExpenseAccountID: expense, Description: "second", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.RequireFromString("13.4321"), TaxCodeID: &code, TaxRate: decimal.NewFromInt(99)},
		}})
		if err != nil {
			t.Fatal(err)
		}
		if !invoice.Lines[0].TaxRate.Equal(decimal.NewFromInt(15)) {
			t.Fatal("client tax rate overrode configured rate")
		}
		return invoice
	}
	invoice := create()
	posted, err := mod.PayableService.PostSupplierInvoice(ctx, invoice.ID, &actor)
	if err != nil {
		t.Fatal(err)
	}
	var debit, credit, vat, tax decimal.Decimal
	var period, postedBy uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT total_debit,total_credit,fiscal_period_id,posted_by FROM accounting.journal_entries WHERE id=$1", posted.JournalEntryID).Scan(&debit, &credit, &period, &postedBy); err != nil {
		t.Fatal(err)
	}
	if period != fiscal || postedBy != actor || !debit.Equal(credit) || !debit.Equal(invoice.TotalAmount.Mul(invoice.ExchangeRate).Round(4)) {
		t.Fatalf("incorrect posting: %s/%s period=%s actor=%s", debit, credit, period, postedBy)
	}
	if err := db.Pool.QueryRow(ctx, "SELECT SUM(l.base_debit) FROM accounting.journal_lines l JOIN accounting.accounts a ON a.id=l.account_id WHERE l.journal_entry_id=$1 AND a.code='114000'", posted.JournalEntryID).Scan(&vat); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, "SELECT total_tax_amount FROM tax.tax_transactions WHERE journal_entry_id=$1 AND transaction_type='INPUT' AND created_by=$2", posted.JournalEntryID, actor).Scan(&tax); err != nil {
		t.Fatal(err)
	}
	if !vat.Equal(tax) {
		t.Fatalf("input VAT differs from GL: %s / %s", vat, tax)
	}
	var journalCount int
	if _, err := mod.PayableService.PostSupplierInvoice(ctx, invoice.ID, &actor); err == nil {
		t.Fatal("duplicate posting accepted")
	}
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM accounting.journal_entries WHERE source_id=$1", invoice.ID).Scan(&journalCount); err != nil || journalCount != 1 {
		t.Fatalf("duplicate journal: %d %v", journalCount, err)
	}

	for _, failure := range []string{"accounting-period", "tax-period", "audit"} {
		t.Run(failure, func(t *testing.T) {
			candidate := create()
			switch failure {
			case "accounting-period":
				if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", fiscal); err != nil {
					t.Fatal(err)
				}
				defer db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1", fiscal)
			case "tax-period":
				if _, err := db.Pool.Exec(ctx, "UPDATE tax.tax_periods SET status='CLOSED' WHERE id=$1", taxPeriod); err != nil {
					t.Fatal(err)
				}
				defer db.Pool.Exec(ctx, "UPDATE tax.tax_periods SET status='OPEN' WHERE id=$1", taxPeriod)
			case "audit":
				if _, err := db.Pool.Exec(ctx, `CREATE FUNCTION audit.reject_supplier_post() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected audit failure'; END $$; CREATE TRIGGER reject_supplier_post BEFORE INSERT ON audit.events FOR EACH ROW EXECUTE FUNCTION audit.reject_supplier_post()`); err != nil {
					t.Fatal(err)
				}
				defer db.Pool.Exec(ctx, `DROP TRIGGER reject_supplier_post ON audit.events; DROP FUNCTION audit.reject_supplier_post()`)
			}
			if _, err := mod.PayableService.PostSupplierInvoice(ctx, candidate.ID, &actor); err == nil {
				t.Fatal("expected mandatory posting failure")
			}
			current, err := mod.PayableService.GetSupplierInvoice(ctx, candidate.ID)
			if err != nil {
				t.Fatal(err)
			}
			if current.Status != domain.InvoiceStatusDraft || current.JournalEntryID != nil {
				t.Fatal("failed posting mutated invoice")
			}
			var count int
			if err := db.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM accounting.journal_entries WHERE source_id=$1)+(SELECT count(*) FROM payables.open_items WHERE document_id=$1)+(SELECT count(*) FROM tax.tax_transactions WHERE source_id=$1)`, candidate.ID).Scan(&count); err != nil || count != 0 {
				t.Fatalf("failed posting left financial rows: %d %v", count, err)
			}
		})
	}
}
