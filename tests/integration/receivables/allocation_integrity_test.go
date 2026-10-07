package receivables_test

import (
	"sync"
	"testing"
	"time"

	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestReceiptAllocationRejectsInvalidTotalsAndPartyCurrencyMismatch(t *testing.T) {
	mod, db, entity, actor, ar, bank, ctx := setupReceivablesTest(t)
	defer db.Close()
	createCustomer := func() uuid.UUID {
		t.Helper()
		c, err := mod.CustomerService.CreateCustomer(ctx, ports.CreateCustomerCommand{LegalEntityID: entity, CustomerType: domain.CustomerTypeB2B, NameEn: uuid.NewString(), NameAr: "عميل اختبار", Address: domain.SaudiAddress{City: "Riyadh", CountryCode: "SA"}, DefaultAccountID: &ar, CurrencyCode: "SAR"})
		if err != nil {
			t.Fatal(err)
		}
		return c.ID
	}
	customer, other := createCustomer(), createCustomer()
	createItem := func(party uuid.UUID, currency string) uuid.UUID {
		t.Helper()
		item, err := mod.ReceivablesService.CreateOpenItem(ctx, &domain.OpenItem{LegalEntityID: entity, CustomerID: party, DocumentType: domain.DocumentTypeInvoice, DocumentNumber: uuid.NewString(), DocumentDate: time.Now(), DueDate: time.Now(), CurrencyCode: currency, OriginalAmount: decimal.NewFromInt(100)})
		if err != nil {
			t.Fatal(err)
		}
		return item.ID
	}
	valid, wrongParty, wrongCurrency := createItem(customer, "SAR"), createItem(other, "SAR"), createItem(customer, "USD")
	command := ports.CreateReceiptCommand{LegalEntityID: entity, CustomerID: customer, PaymentMethod: domain.PaymentMethodBankTransfer, DepositAccountID: bank, Amount: decimal.NewFromInt(100), CurrencyCode: "SAR", CreatedByUserID: &actor}
	receipt, err := mod.ReceivablesService.CreateReceipt(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name        string
		allocations []ports.AllocationCommand
	}{
		{"negative-cancellation", []ports.AllocationCommand{{OpenItemID: valid, AllocatedAmount: decimal.NewFromInt(100)}, {OpenItemID: wrongParty, AllocatedAmount: decimal.NewFromInt(-100)}}},
		{"zero", []ports.AllocationCommand{{OpenItemID: valid, AllocatedAmount: decimal.Zero}}},
		{"different-customer", []ports.AllocationCommand{{OpenItemID: wrongParty, AllocatedAmount: decimal.NewFromInt(10)}}},
		{"different-currency", []ports.AllocationCommand{{OpenItemID: wrongCurrency, AllocatedAmount: decimal.NewFromInt(10)}}},
		{"duplicate", []ports.AllocationCommand{{OpenItemID: valid, AllocatedAmount: decimal.NewFromInt(10)}, {OpenItemID: valid, AllocatedAmount: decimal.NewFromInt(10)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := mod.ReceivablesService.AllocateReceipt(ctx, receipt.ID, tc.allocations, &actor); err == nil {
				t.Fatal("invalid allocation accepted")
			}
			command.Allocations = tc.allocations
			if _, err := mod.ReceivablesService.CreateReceipt(ctx, command); err == nil {
				t.Fatal("invalid initial allocation accepted")
			}
		})
	}
	var remaining decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT remaining_amount FROM receivables.open_items WHERE id=$1", valid).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if !remaining.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("rejection changed invoice: %s", remaining)
	}
	// Concurrent requests must not over-settle the receipt or its invoice.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			results <- mod.ReceivablesService.AllocateReceipt(ctx, receipt.ID, []ports.AllocationCommand{{OpenItemID: valid, AllocatedAmount: decimal.NewFromInt(80)}}, &actor)
		})
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful concurrent allocations=%d, want 1", successes)
	}
	got, err := mod.ReceivablesService.GetReceipt(ctx, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.AllocatedAmount.Equal(decimal.NewFromInt(80)) || !got.UnallocatedAmount.Equal(decimal.NewFromInt(20)) {
		t.Fatal("receipt totals do not reconcile")
	}
}
