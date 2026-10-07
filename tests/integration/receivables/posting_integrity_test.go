package receivables_test

import (
	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestReceiptPostingAndClosedPeriodRollback(t *testing.T) {
	mod, db, entity, actor, ar, bank, ctx := setupReceivablesTest(t)
	defer db.Close()
	customer, err := mod.CustomerService.CreateCustomer(ctx, ports.CreateCustomerCommand{LegalEntityID: entity, CustomerType: domain.CustomerTypeB2C, NameEn: uuid.NewString(), NameAr: "عميل اختبار", Address: domain.SaudiAddress{City: "Riyadh"}, CurrencyCode: "SAR", DefaultAccountID: &ar})
	if err != nil {
		t.Fatal(err)
	}
	item, err := mod.ReceivablesService.CreateOpenItem(ctx, &domain.OpenItem{LegalEntityID: entity, CustomerID: customer.ID, DocumentType: domain.DocumentTypeInvoice, DocumentNumber: uuid.NewString(), DocumentDate: time.Now(), DueDate: time.Now(), CurrencyCode: "SAR", OriginalAmount: decimal.NewFromInt(100), RemainingAmount: decimal.NewFromInt(100), Status: domain.OpenItemStatusOpen})
	if err != nil {
		t.Fatal(err)
	}
	cmd := ports.CreateReceiptCommand{LegalEntityID: entity, CustomerID: customer.ID, PaymentMethod: domain.PaymentMethodBankTransfer, DepositAccountID: bank, Amount: decimal.NewFromInt(40), CurrencyCode: "SAR", CreatedByUserID: &actor, Allocations: []ports.AllocationCommand{{OpenItemID: item.ID, AllocatedAmount: decimal.NewFromInt(40)}}}
	receipt, err := mod.ReceivablesService.CreateReceipt(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.JournalEntryID == nil {
		t.Fatal("receipt has no journal")
	}
	var debit, credit decimal.Decimal
	var period, postedBy uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT total_debit,total_credit,fiscal_period_id,posted_by FROM accounting.journal_entries WHERE id=$1", receipt.JournalEntryID).Scan(&debit, &credit, &period, &postedBy); err != nil {
		t.Fatal(err)
	}
	if !debit.Equal(cmd.Amount) || !credit.Equal(cmd.Amount) || postedBy != actor {
		t.Fatal("receipt GL amounts or actor differ")
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", period); err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1", period)
	if _, err := mod.ReceivablesService.CreateReceipt(ctx, cmd); err == nil {
		t.Fatal("closed-period receipt accepted")
	}
	var count int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM receivables.receipts WHERE customer_id=$1", customer.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("failed receipt persisted: %d %v", count, err)
	}
	var remaining decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT remaining_amount FROM receivables.open_items WHERE id=$1", item.ID).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if !remaining.Equal(decimal.NewFromInt(60)) {
		t.Fatalf("failed receipt consumed allocation: %s", remaining)
	}
}
