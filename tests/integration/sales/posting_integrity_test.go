package sales_test

import (
	"errors"
	"testing"
	"time"

	accountingdomain "erp/internal/accounting/domain"
	recdomain "erp/internal/receivables/domain"
	"erp/internal/sales/domain"
	"erp/internal/sales/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSales_ConfiguredTaxAndForeignCurrencyPosting(t *testing.T) {
	mod, rec, db, entityID, actorID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()
	customer := createTestCustomer(t, rec, entityID, recdomain.CustomerTypeB2C, "عميل", "Customer", nil, decimal.NewFromInt(10000), ctx)
	date := time.Date(2026, time.February, 15, 0, 0, 0, 0, time.UTC)
	invoice, err := mod.SalesService.CreateInvoice(ctx, ports.CreateInvoiceCommand{
		LegalEntityID: entityID, CustomerID: customer.ID, CustomerNameAr: customer.NameAr,
		CustomerNameEn: customer.NameEn, InvoiceType: domain.InvoiceTypeSimplified,
		IssueDate: date, DueDate: date.AddDate(0, 0, 30), CurrencyCode: "USD",
		ExchangeRate: decimal.RequireFromString("3.75"), CreatedByUserID: &actorID,
		Lines: []ports.CreateInvoiceLineCommand{{ItemID: itemID, Description: "Configured tax",
			UOM: "EA", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.NewFromInt(100),
			TaxCode: "VAT-STD-15", TaxRatePercent: decimal.NewFromInt(99)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !invoice.TaxAmount.Equal(decimal.NewFromInt(15)) || !invoice.Lines[0].TaxRatePercent.Equal(decimal.NewFromInt(15)) {
		t.Fatalf("client tax rate overrode configured rate: %+v", invoice)
	}
	posted, err := mod.SalesService.PostInvoice(ctx, invoice.ID, actorID)
	if err != nil {
		t.Fatal(err)
	}
	var rate, debit, credit decimal.Decimal
	var periodDate time.Time
	if err := db.Pool.QueryRow(ctx, `SELECT e.exchange_rate, e.total_debit, e.total_credit, p.start_date
		FROM accounting.journal_entries e JOIN accounting.fiscal_periods p ON p.id=e.fiscal_period_id
		WHERE e.id=$1`, posted.JournalEntryID).Scan(&rate, &debit, &credit, &periodDate); err != nil {
		t.Fatal(err)
	}
	if !rate.Equal(decimal.RequireFromString("3.75")) || !debit.Equal(decimal.RequireFromString("431.25")) || !debit.Equal(credit) || periodDate.Month() != time.February {
		t.Fatalf("wrong foreign currency posting: rate=%s debit=%s credit=%s period=%s", rate, debit, credit, periodDate)
	}
	var vat decimal.Decimal
	if err := db.Pool.QueryRow(ctx, `SELECT total_tax_amount FROM tax.tax_transactions WHERE source_id=$1`, invoice.ID).Scan(&vat); err != nil {
		t.Fatal(err)
	}
	if !vat.Equal(decimal.RequireFromString("56.25")) {
		t.Fatalf("VAT ledger must use base currency: %s", vat)
	}
	var balance decimal.Decimal
	if err := db.Pool.QueryRow(ctx, `SELECT b.period_debit FROM accounting.account_balances b
		JOIN accounting.journal_lines l ON l.account_id=b.account_id
		JOIN accounting.journal_entries e ON e.id=l.journal_entry_id AND e.fiscal_period_id=b.fiscal_period_id
		WHERE e.id=$1 AND l.line_number=1`, posted.JournalEntryID).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if balance.LessThan(debit) {
		t.Fatalf("AR projection was not updated: %s", balance)
	}
	creditNote, err := mod.SalesService.CreateCreditNote(ctx, ports.CreateCreditNoteCommand{
		LegalEntityID: entityID, OriginalInvoiceID: invoice.ID, CustomerID: customer.ID,
		IssueDate: date, ReasonCode: domain.CreditNoteReason("RETURN"), CreatedByUserID: &actorID,
		Lines: []ports.CreateCreditNoteLineCommand{{ItemID: itemID, Description: "Return", Quantity: decimal.NewFromInt(1),
			UnitPrice: decimal.NewFromInt(100), TaxCode: "VAT-STD-15", TaxRatePercent: decimal.NewFromInt(99)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if creditNote.CurrencyCode != "USD" || !creditNote.ExchangeRate.Equal(rate) || !creditNote.TaxAmount.Equal(invoice.TaxAmount) {
		t.Fatalf("credit note did not preserve original invoice calculation: %+v", creditNote)
	}
	if _, err := mod.SalesService.PostCreditNote(ctx, creditNote.ID, actorID); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, `SELECT total_tax_amount FROM tax.tax_transactions WHERE source_id=$1`, creditNote.ID).Scan(&vat); err != nil {
		t.Fatal(err)
	}
	if !vat.Equal(decimal.RequireFromString("-56.25")) {
		t.Fatalf("credit note must reduce VAT liability: %s", vat)
	}
	duplicate, err := mod.SalesService.CreateCreditNote(ctx, ports.CreateCreditNoteCommand{
		LegalEntityID: entityID, OriginalInvoiceID: invoice.ID, CustomerID: customer.ID,
		IssueDate: date, ReasonCode: domain.CreditNoteReasonReturn, CreatedByUserID: &actorID,
		Lines: []ports.CreateCreditNoteLineCommand{{ItemID: itemID, Description: "Excess return", Quantity: decimal.NewFromInt(1),
			UnitPrice: decimal.NewFromInt(1), TaxCode: "VAT-STD-15"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.SalesService.PostCreditNote(ctx, duplicate.ID, actorID); err == nil {
		t.Fatal("credit notes exceeding the original invoice amount must be rejected")
	}
}

func TestSales_ClosedPeriodsRollbackEntirePosting(t *testing.T) {
	mod, rec, db, entityID, actorID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()
	customer := createTestCustomer(t, rec, entityID, recdomain.CustomerTypeB2C, "عميل", "Customer", nil, decimal.NewFromInt(10000), ctx)
	date := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)
	for _, schema := range []string{"accounting", "tax"} {
		t.Run(schema, func(t *testing.T) {
			invoice, err := mod.SalesService.CreateInvoice(ctx, ports.CreateInvoiceCommand{
				LegalEntityID: entityID, CustomerID: customer.ID, CustomerNameAr: customer.NameAr,
				CustomerNameEn: customer.NameEn, InvoiceType: domain.InvoiceTypeSimplified,
				IssueDate: date, DueDate: date, CurrencyCode: "SAR", CreatedByUserID: &actorID,
				Lines: []ports.CreateInvoiceLineCommand{{ItemID: itemID, Description: "Blocked", UOM: "EA",
					Quantity: decimal.NewFromInt(1), UnitPrice: decimal.NewFromInt(100), TaxCode: "VAT-STD-15"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			table := "accounting.fiscal_periods"
			if schema == "tax" {
				table = "tax.tax_periods"
			}
			var periodID uuid.UUID
			if err := db.Pool.QueryRow(ctx, "UPDATE "+table+" SET status='CLOSED' WHERE legal_entity_id=$1 AND $2::date BETWEEN start_date AND end_date RETURNING id", entityID, date).Scan(&periodID); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if _, err := db.Pool.Exec(ctx, "UPDATE "+table+" SET status='OPEN' WHERE id=$1", periodID); err != nil {
					t.Error(err)
				}
			}()
			_, err = mod.SalesService.PostInvoice(ctx, invoice.ID, actorID)
			if err == nil {
				t.Fatal("closed period accepted posting")
			}
			if schema == "accounting" && !errors.Is(err, accountingdomain.ErrPeriodNotOpen) {
				t.Fatalf("expected closed accounting period, got %v", err)
			}
			for _, query := range []string{
				"SELECT count(*) FROM accounting.journal_entries WHERE source_id=$1",
				"SELECT count(*) FROM tax.tax_transactions WHERE source_id=$1",
				"SELECT count(*) FROM receivables.open_items WHERE document_id=$1",
			} {
				var count int
				if err := db.Pool.QueryRow(ctx, query, invoice.ID).Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 0 {
					t.Fatalf("failed posting leaked %d records: %s", count, query)
				}
			}
			stored, err := mod.SalesService.GetInvoice(ctx, invoice.ID)
			if err != nil || stored.Status != domain.InvoiceStatusDraft {
				t.Fatalf("failed posting altered invoice: %v %+v", err, stored)
			}
		})
	}
}

func TestSales_DraftLineFailureRollsBackHeader(t *testing.T) {
	mod, rec, db, entityID, actorID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()
	customer := createTestCustomer(t, rec, entityID, recdomain.CustomerTypeB2C, "عميل", "Customer", nil, decimal.NewFromInt(10000), ctx)
	date := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)
	_, err := mod.SalesService.CreateInvoice(ctx, ports.CreateInvoiceCommand{
		LegalEntityID: entityID, CustomerID: customer.ID, CustomerNameAr: customer.NameAr,
		CustomerNameEn: customer.NameEn, InvoiceType: domain.InvoiceTypeSimplified,
		IssueDate: date, DueDate: date, CurrencyCode: "SAR", CreatedByUserID: &actorID,
		Lines: []ports.CreateInvoiceLineCommand{
			{ItemID: itemID, Description: "Valid line", UOM: "EA", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.NewFromInt(100), TaxCode: "VAT-STD-15"},
			{ItemID: uuid.New(), Description: "Nonexistent item", UOM: "EA", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.NewFromInt(100), TaxCode: "VAT-STD-15"},
		},
	})
	if err == nil {
		t.Fatal("expected nonexistent item to fail")
	}
	var count int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM sales.invoices WHERE customer_id=$1", customer.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("failed draft creation left its header and first line persisted")
	}
}

func TestSales_PostingActorAndFractionalVATReconcile(t *testing.T) {
	mod, rec, db, entityID, creatorID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()
	posterID := uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO iam.users(id, username, display_name, status)
		VALUES($1,$2,'Posting user','ACTIVE')`, posterID, "poster-"+posterID.String()); err != nil {
		t.Fatal(err)
	}
	customer := createTestCustomer(t, rec, entityID, recdomain.CustomerTypeB2C, "عميل", "Customer", nil, decimal.NewFromInt(10000), ctx)
	date := time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC)
	line := ports.CreateInvoiceLineCommand{ItemID: itemID, Description: "Fractional line", UOM: "EA",
		Quantity: decimal.NewFromInt(1), UnitPrice: decimal.RequireFromString("0.01"), TaxCode: "VAT-STD-15"}
	invoice, err := mod.SalesService.CreateInvoice(ctx, ports.CreateInvoiceCommand{
		LegalEntityID: entityID, CustomerID: customer.ID, CustomerNameAr: customer.NameAr,
		CustomerNameEn: customer.NameEn, InvoiceType: domain.InvoiceTypeSimplified,
		IssueDate: date, DueDate: date, CurrencyCode: "USD", ExchangeRate: decimal.RequireFromString("3.75"),
		CreatedByUserID: &creatorID, Lines: []ports.CreateInvoiceLineCommand{line, line},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.SalesService.PostInvoice(ctx, invoice.ID, uuid.Nil); err == nil {
		t.Fatal("a posting command must require its authenticated actor")
	}
	posted, err := mod.SalesService.PostInvoice(ctx, invoice.ID, posterID)
	if err != nil {
		t.Fatal(err)
	}
	if posted.CreatedByUserID == nil || *posted.CreatedByUserID != creatorID {
		t.Fatal("posting must preserve the original document creator")
	}
	checkPosting := func(sourceID uuid.UUID, expectedTax string, reversal bool) {
		t.Helper()
		var journalActor, auditActor, taxActor uuid.UUID
		var glTotal, vatTotal, taxGross, ledgerTax, linesTax decimal.Decimal
		if err := db.Pool.QueryRow(ctx, `SELECT e.posted_by,a.user_id,t.created_by,e.total_debit,t.total_tax_amount,t.total_gross_amount,
			(SELECT sum(l.base_credit-l.base_debit) FROM accounting.journal_lines l JOIN accounting.accounts acc ON acc.id=l.account_id WHERE l.journal_entry_id=e.id AND acc.code='213000'),
			(SELECT sum(l.tax_amount) FROM tax.tax_transaction_lines l WHERE l.tax_transaction_id=t.id)
			FROM accounting.journal_entries e
			JOIN audit.events a ON a.resource_id=e.id::text AND a.action='accounting.journal.post'
			JOIN tax.tax_transactions t ON t.source_id=e.source_id AND t.source_module='SALES'
			WHERE e.source_id=$1`, sourceID).Scan(&journalActor, &auditActor, &taxActor, &glTotal, &vatTotal, &taxGross, &ledgerTax, &linesTax); err != nil {
			t.Fatal(err)
		}
		if journalActor != posterID || auditActor != posterID || taxActor != posterID {
			t.Fatalf("posting was attributed to the creator instead of the posting user: GL=%s audit=%s VAT=%s", journalActor, auditActor, taxActor)
		}
		if !vatTotal.Equal(decimal.RequireFromString(expectedTax)) || !vatTotal.Equal(ledgerTax) || !vatTotal.Equal(linesTax) {
			t.Fatalf("VAT conversion does not reconcile: tax=%s GL=%s lines=%s", vatTotal, ledgerTax, linesTax)
		}
		if !glTotal.Equal(decimal.RequireFromString("0.0863")) || !glTotal.Equal(taxGross.Abs()) || (reversal != taxGross.IsNegative()) {
			t.Fatalf("GL and tax document totals disagree: GL=%s tax=%s", glTotal, taxGross)
		}
	}
	checkPosting(invoice.ID, "0.0113", false)
	creditLine := ports.CreateCreditNoteLineCommand{ItemID: itemID, Description: "Fractional return",
		Quantity: line.Quantity, UnitPrice: line.UnitPrice, TaxCode: line.TaxCode}
	creditNote, err := mod.SalesService.CreateCreditNote(ctx, ports.CreateCreditNoteCommand{
		LegalEntityID: entityID, OriginalInvoiceID: invoice.ID, CustomerID: customer.ID,
		IssueDate: date, ReasonCode: domain.CreditNoteReasonReturn, CreatedByUserID: &creatorID,
		Lines: []ports.CreateCreditNoteLineCommand{creditLine, creditLine},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.SalesService.PostCreditNote(ctx, creditNote.ID, posterID); err != nil {
		t.Fatal(err)
	}
	checkPosting(creditNote.ID, "-0.0113", true)
}
