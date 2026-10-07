package payables_test

import (
	"testing"
	"time"

	"erp/internal/payables/domain"
	"erp/internal/payables/infrastructure/postgres"
	"erp/internal/payables/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestSupplierPaymentAllocationRequiresSamePartyAndCurrency(t *testing.T) {
	mod, db, entity, actor, ap, bank, expense, ctx := setupPayablesTest(t)
	defer db.Close()
	createSupplier := func() uuid.UUID {
		t.Helper()
		s, err := mod.SupplierService.CreateSupplier(ctx, ports.CreateSupplierCmd{LegalEntityID: entity, SupplierType: domain.SupplierTypeLocal, NameEn: uuid.NewString(), NameAr: "مورد اختبار", Address: domain.SaudiAddress{City: "Riyadh"}, DefaultLiabilityAccountID: &ap, DefaultExpenseAccountID: &expense})
		if err != nil {
			t.Fatal(err)
		}
		return s.ID
	}
	supplier, other := createSupplier(), createSupplier()
	repo := postgres.NewPayableRepository(db)
	createItem := func(party uuid.UUID, currency string) uuid.UUID {
		t.Helper()
		item, err := repo.CreateOpenItem(ctx, &domain.PayablesOpenItem{LegalEntityID: entity, SupplierID: party, DocumentType: domain.DocumentTypeInvoice, DocumentNumber: uuid.NewString(), DocumentDate: time.Now(), DueDate: time.Now(), CurrencyCode: currency, OriginalAmount: decimal.NewFromInt(100), RemainingAmount: decimal.NewFromInt(100), Status: domain.OpenItemStatusOpen})
		if err != nil {
			t.Fatal(err)
		}
		return item.ID
	}
	valid, wrongParty, wrongCurrency := createItem(supplier, "SAR"), createItem(other, "SAR"), createItem(supplier, "USD")
	for _, tc := range []struct {
		name        string
		allocations []ports.PaymentAllocationCmd
	}{
		{"negative", []ports.PaymentAllocationCmd{{OpenItemID: valid, Amount: decimal.NewFromInt(-1)}}},
		{"zero", []ports.PaymentAllocationCmd{{OpenItemID: valid, Amount: decimal.Zero}}},
		{"different-supplier", []ports.PaymentAllocationCmd{{OpenItemID: wrongParty, Amount: decimal.NewFromInt(1)}}},
		{"different-currency", []ports.PaymentAllocationCmd{{OpenItemID: wrongCurrency, Amount: decimal.NewFromInt(1)}}},
		{"duplicate", []ports.PaymentAllocationCmd{{OpenItemID: valid, Amount: decimal.NewFromInt(1)}, {OpenItemID: valid, Amount: decimal.NewFromInt(1)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mod.PayableService.CreateSupplierPayment(ctx, ports.CreateSupplierPaymentCmd{LegalEntityID: entity, SupplierID: supplier, PaymentMethod: domain.PaymentMethodBankTransfer, DisbursementAccountID: bank, Amount: decimal.NewFromInt(100), CurrencyCode: "SAR", CreatedByUserID: &actor, Allocations: tc.allocations})
			if err == nil {
				t.Fatal("invalid allocation accepted")
			}
		})
	}
	item, err := repo.GetOpenItemByID(ctx, valid)
	if err != nil {
		t.Fatal(err)
	}
	if !item.RemainingAmount.Equal(decimal.NewFromInt(100)) {
		t.Fatal("rejected allocation changed invoice")
	}
}
