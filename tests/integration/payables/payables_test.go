package payables_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/payables"
	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupPayablesTest(t *testing.T) (*payables.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("integration database setup failed: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	payMod := payables.NewModule(cfg, db)

	// 1. Ensure organization & legal entity exist
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Seed accounting defaults (Chart of accounts)
	_ = accMod.SeedDefaultCOA(ctx, le.ID)

	// 3. Seed payables defaults
	_ = payMod.SeedDefaultPayablesData(ctx, le.ID)

	// 4. Get admin user
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	}

	// 5. Lookup accounts: AP (211000), Bank (111002), posting Expense (523000)
	var apAccountID, bankAccountID, expAccountID uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '211000' LIMIT 1", le.ID).Scan(&apAccountID)
	if err != nil {
		t.Fatalf("failed to find AP account: %v", err)
	}
	err = db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '111002' LIMIT 1", le.ID).Scan(&bankAccountID)
	if err != nil {
		t.Fatalf("failed to find Bank account: %v", err)
	}
	err = db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '523000' AND is_active AND allow_posting LIMIT 1", le.ID).Scan(&expAccountID)
	if err != nil {
		t.Fatalf("failed to find Expense account: %v", err)
	}

	return payMod, db, le.ID, adminID, apAccountID, bankAccountID, expAccountID, ctx
}

func TestSupplier_CreationAndSaudiValidation(t *testing.T) {
	payMod, db, leID, _, apAccountID, _, expAccountID, ctx := setupPayablesTest(t)
	defer db.Close()

	suppSvc := payMod.SupplierService

	validTaxID := "310123456700003"
	validCR := "1010897654"
	validIBAN := "SA0380000000123456789012"

	// 1. Positive: Create Valid Local Supplier
	cmd := ports.CreateSupplierCmd{
		LegalEntityID: leID,
		SupplierType:  domain.SupplierTypeLocal,
		NameEn:        "Rawabi Industrial Supplies Co.",
		NameAr:        "شركة روابي للتجهيزات الصناعية",
		TaxID:         &validTaxID,
		CRNumber:      &validCR,
		IBAN:          &validIBAN,
		BankName:      strPtr("Al-Rajhi Bank"),
		Address: domain.SaudiAddress{
			BuildingNo:   "4567",
			StreetName:   "King Abdulaziz Road",
			District:     "Al-Morouj",
			City:         "Riyadh",
			PostalCode:   "12284",
			AdditionalNo: "8912",
			CountryCode:  "SA",
		},
		Email:                     strPtr("sales@rawabi.example.com"),
		Phone:                     strPtr("+966112345678"),
		ContactPerson:             strPtr("Eng. Fahad Al-Rawabi"),
		DefaultExpenseAccountID:   &expAccountID,
		DefaultLiabilityAccountID: &apAccountID,
	}

	supp, err := suppSvc.CreateSupplier(ctx, cmd)
	if err != nil {
		t.Fatalf("expected valid supplier creation, got: %v", err)
	}

	if supp.ID == uuid.Nil {
		t.Errorf("expected generated UUID for supplier, got nil")
	}
	if supp.SupplierNumber == "" {
		t.Errorf("expected generated supplier number, got empty")
	}
	if supp.Status != domain.SupplierStatusActive {
		t.Errorf("expected ACTIVE status, got %s", supp.Status)
	}

	// 2. Negative: Invalid Saudi Tax ID (not 15 digits or not ending with 3)
	badTaxID := "123456789012345"
	badCmd := cmd
	badCmd.SupplierNumber = "BAD-SUP-1"
	badCmd.TaxID = &badTaxID
	_, err = suppSvc.CreateSupplier(ctx, badCmd)
	if !errors.Is(err, domain.ErrInvalidTaxID) {
		t.Errorf("expected ErrInvalidTaxID, got: %v", err)
	}

	// 3. Negative: Invalid CR Number (not 10 digits)
	badCR := "12345"
	badCmd2 := cmd
	badCmd2.SupplierNumber = "BAD-SUP-2"
	badCmd2.TaxID = &validTaxID
	badCmd2.CRNumber = &badCR
	_, err = suppSvc.CreateSupplier(ctx, badCmd2)
	if !errors.Is(err, domain.ErrInvalidCRNumber) {
		t.Errorf("expected ErrInvalidCRNumber, got: %v", err)
	}

	// 4. Negative: Invalid IBAN format
	badIBAN := "INVALIDIBAN123"
	badCmd3 := cmd
	badCmd3.SupplierNumber = "BAD-SUP-3"
	badCmd3.TaxID = &validTaxID
	badCmd3.CRNumber = &validCR
	badCmd3.IBAN = &badIBAN
	_, err = suppSvc.CreateSupplier(ctx, badCmd3)
	if !errors.Is(err, domain.ErrInvalidIBAN) {
		t.Errorf("expected ErrInvalidIBAN, got: %v", err)
	}

	// 5. Negative: Missing City
	badCmd4 := cmd
	badCmd4.SupplierNumber = "BAD-SUP-4"
	badCmd4.TaxID = &validTaxID
	badCmd4.CRNumber = &validCR
	badCmd4.IBAN = &validIBAN
	badCmd4.Address.City = ""
	_, err = suppSvc.CreateSupplier(ctx, badCmd4)
	if !errors.Is(err, domain.ErrCityRequired) {
		t.Errorf("expected ErrCityRequired, got: %v", err)
	}
}

func TestPayables_PaymentTerms(t *testing.T) {
	payMod, db, leID, _, _, _, _, ctx := setupPayablesTest(t)
	defer db.Close()

	suppSvc := payMod.SupplierService

	// 1. Create custom payment term
	term, err := suppSvc.CreatePaymentTerm(ctx, &domain.PaymentTerm{
		LegalEntityID:   leID,
		Code:            fmt.Sprintf("NET45-%s", uuid.New().String()[:8]),
		NameEn:          "Net 45 Days with 2% early payment",
		NameAr:          "صافي 45 يوماً مع خصم تعجيل 2%",
		DueDays:         45,
		DiscountDays:    10,
		DiscountPercent: decimal.NewFromFloat(2.00),
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("failed to create payment term: %v", err)
	}

	if term.ID == uuid.Nil {
		t.Errorf("expected generated UUID for payment term")
	}

	// 2. List terms
	terms, err := suppSvc.ListPaymentTerms(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list payment terms: %v", err)
	}
	if len(terms) == 0 {
		t.Errorf("expected at least one payment term, got 0")
	}
}

func TestPayables_SupplierInvoiceCreationAndGLPosting(t *testing.T) {
	payMod, db, leID, adminID, apAccountID, _, expAccountID, ctx := setupPayablesTest(t)
	defer db.Close()

	suppSvc := payMod.SupplierService
	payableSvc := payMod.PayableService

	// 1. Create a Supplier
	taxID := "310987654300003"
	crNum := "1010765432"
	supp, err := suppSvc.CreateSupplier(ctx, ports.CreateSupplierCmd{
		LegalEntityID: leID,
		SupplierType:  domain.SupplierTypeLocal,
		NameEn:        "Saudi Modern Hardware Tech",
		NameAr:        "شركة التقنية الحديثة للمعدات",
		TaxID:         &taxID,
		CRNumber:      &crNum,
		Address: domain.SaudiAddress{
			City:        "Riyadh",
			CountryCode: "SA",
		},
		DefaultExpenseAccountID:   &expAccountID,
		DefaultLiabilityAccountID: &apAccountID,
	})
	if err != nil {
		t.Fatalf("failed to create supplier: %v", err)
	}

	// 2. Create a Supplier Invoice
	// Line 1: 10 units @ 100 SAR, 0% discount, 15% VAT = 1,000 subtotal, 150 VAT, 1,150 total
	// Line 2: 5 units @ 200 SAR, 10% discount, 15% VAT = 900 subtotal, 135 VAT, 1,035 total
	// Total: 1,900 subtotal, 285 VAT, 2,185 grand total
	invoiceCmd := ports.CreateSupplierInvoiceCmd{
		LegalEntityID:         leID,
		SupplierInvoiceNumber: "INV-SUP-8899",
		SupplierID:            supp.ID,
		InvoiceDate:           time.Now().Format("2006-01-02"),
		DueDate:               time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		CurrencyCode:          "SAR",
		Lines: []ports.CreateSupplierInvoiceLineCmd{
			{
				ExpenseAccountID: expAccountID,
				Description:      "Server Hardware Upgrades",
				Quantity:         decimal.NewFromInt(10),
				UnitPrice:        decimal.NewFromInt(100),
				DiscountPercent:  decimal.Zero,
				TaxRate:          decimal.NewFromInt(15),
			},
			{
				ExpenseAccountID: expAccountID,
				Description:      "Network Cabling Infrastructure",
				Quantity:         decimal.NewFromInt(5),
				UnitPrice:        decimal.NewFromInt(200),
				DiscountPercent:  decimal.NewFromInt(10),
				TaxRate:          decimal.NewFromInt(15),
			},
		},
	}

	inv, err := payableSvc.CreateSupplierInvoice(ctx, invoiceCmd)
	if err != nil {
		t.Fatalf("failed to create supplier invoice: %v", err)
	}

	expectedSubtotal := decimal.NewFromInt(1900)
	expectedTax := decimal.NewFromFloat(285)
	expectedTotal := decimal.NewFromFloat(2185)

	if !inv.Subtotal.Equal(expectedSubtotal) {
		t.Errorf("expected subtotal %s, got %s", expectedSubtotal, inv.Subtotal)
	}
	if !inv.TaxAmount.Equal(expectedTax) {
		t.Errorf("expected tax amount %s, got %s", expectedTax, inv.TaxAmount)
	}
	if !inv.TotalAmount.Equal(expectedTotal) {
		t.Errorf("expected total amount %s, got %s", expectedTotal, inv.TotalAmount)
	}
	if inv.Status != domain.InvoiceStatusDraft {
		t.Errorf("expected DRAFT status, got %s", inv.Status)
	}
	if inv.PaymentStatus != domain.PaymentStatusUnpaid {
		t.Errorf("expected UNPAID payment status, got %s", inv.PaymentStatus)
	}

	// 3. Post Supplier Invoice
	postedInv, err := payableSvc.PostSupplierInvoice(ctx, inv.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to post supplier invoice: %v", err)
	}

	if postedInv.Status != domain.InvoiceStatusPosted {
		t.Errorf("expected POSTED status, got %s", postedInv.Status)
	}
	if postedInv.JournalEntryID == nil {
		t.Errorf("expected linked journal entry ID on posted invoice")
	}

	// 4. Verify Open Item created in payables.open_items
	items, err := payableSvc.ListOpenItems(ctx, supp.ID, true)
	if err != nil {
		t.Fatalf("failed to list open items: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected at least 1 open item for supplier, got 0")
	}
	if !items[0].RemainingAmount.Equal(expectedTotal) {
		t.Errorf("expected open item remaining amount %s, got %s", expectedTotal, items[0].RemainingAmount)
	}

	// 5. Verify Double-Entry GL Journal Entry is Balanced
	var totalDebit, totalCredit decimal.Decimal
	var jeStatus string
	err = db.Pool.QueryRow(ctx, `
		SELECT total_debit, total_credit, status
		FROM accounting.journal_entries
		WHERE id = $1
	`, postedInv.JournalEntryID).Scan(&totalDebit, &totalCredit, &jeStatus)
	if err != nil {
		t.Fatalf("failed to query journal entry: %v", err)
	}

	if jeStatus != "POSTED" {
		t.Errorf("expected journal entry status POSTED, got %s", jeStatus)
	}
	if !totalDebit.Equal(expectedTotal) || !totalCredit.Equal(expectedTotal) {
		t.Errorf("expected balanced debit and credit of %s, got debit=%s credit=%s", expectedTotal, totalDebit, totalCredit)
	}
}

func TestPayables_PaymentAndAllocation(t *testing.T) {
	payMod, db, leID, adminID, apAccountID, bankAccountID, expAccountID, ctx := setupPayablesTest(t)
	defer db.Close()

	suppSvc := payMod.SupplierService
	payableSvc := payMod.PayableService

	// 1. Create Supplier and Invoice
	taxID := "310887766500003"
	supp, err := suppSvc.CreateSupplier(ctx, ports.CreateSupplierCmd{
		LegalEntityID: leID,
		SupplierType:  domain.SupplierTypeLocal,
		NameEn:        "Al-Jazeera Printing Presses",
		NameAr:        "مطابع الجزيرة الحديثة",
		TaxID:         &taxID,
		Address: domain.SaudiAddress{
			City: "Riyadh",
		},
		DefaultExpenseAccountID:   &expAccountID,
		DefaultLiabilityAccountID: &apAccountID,
	})
	if err != nil {
		t.Fatalf("failed to create supplier: %v", err)
	}

	inv, err := payableSvc.CreateSupplierInvoice(ctx, ports.CreateSupplierInvoiceCmd{
		LegalEntityID:         leID,
		SupplierInvoiceNumber: "INV-PAY-TEST",
		SupplierID:            supp.ID,
		InvoiceDate:           time.Now().Format("2006-01-02"),
		DueDate:               time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		CurrencyCode:          "SAR",
		Lines: []ports.CreateSupplierInvoiceLineCmd{
			{
				ExpenseAccountID: expAccountID,
				Description:      "Print Catalogues",
				Quantity:         decimal.NewFromInt(10),
				UnitPrice:        decimal.NewFromInt(100),
				TaxRate:          decimal.NewFromInt(15),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}

	postedInv, err := payableSvc.PostSupplierInvoice(ctx, inv.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to post invoice: %v", err)
	}

	openItems, err := payableSvc.ListOpenItems(ctx, supp.ID, true)
	if err != nil || len(openItems) == 0 {
		t.Fatalf("expected open item for supplier")
	}
	openItem := openItems[0]

	// 2. Partial Payment (Pay 500 out of 1150)
	partialAmount := decimal.NewFromInt(500)
	payCmd := ports.CreateSupplierPaymentCmd{
		LegalEntityID:         leID,
		SupplierID:            supp.ID,
		PaymentDate:           time.Now().Format("2006-01-02"),
		PaymentMethod:         domain.PaymentMethodBankTransfer,
		DisbursementAccountID: bankAccountID,
		Amount:                partialAmount,
		Allocations: []ports.PaymentAllocationCmd{
			{
				OpenItemID: openItem.ID,
				Amount:     partialAmount,
			},
		},
	}

	payment, err := payableSvc.CreateSupplierPayment(ctx, payCmd)
	if err != nil {
		t.Fatalf("failed to create supplier payment: %v", err)
	}

	postedPayment, err := payableSvc.PostSupplierPayment(ctx, payment.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to post supplier payment: %v", err)
	}

	if postedPayment.Status != "POSTED" {
		t.Errorf("expected payment status POSTED, got %s", postedPayment.Status)
	}

	// 3. Verify Open Item Remaining Balance is 1150 - 500 = 650
	itemsAfterPartial, err := payableSvc.ListOpenItems(ctx, supp.ID, false)
	if err != nil || len(itemsAfterPartial) == 0 {
		t.Fatalf("failed to fetch open items after partial payment")
	}
	expectedRemaining := postedInv.TotalAmount.Sub(partialAmount)
	if !itemsAfterPartial[0].RemainingAmount.Equal(expectedRemaining) {
		t.Errorf("expected remaining amount %s, got %s", expectedRemaining, itemsAfterPartial[0].RemainingAmount)
	}
	if itemsAfterPartial[0].Status != domain.OpenItemStatusPartiallyPaid {
		t.Errorf("expected status PARTIALLY_PAID, got %s", itemsAfterPartial[0].Status)
	}

	// 4. Verify Invoice payment_status is PARTIALLY_PAID
	refreshedInv, err := payableSvc.GetSupplierInvoice(ctx, postedInv.ID)
	if err != nil {
		t.Fatalf("failed to fetch refreshed invoice: %v", err)
	}
	if refreshedInv.PaymentStatus != domain.PaymentStatusPartiallyPaid {
		t.Errorf("expected invoice payment status PARTIALLY_PAID, got %s", refreshedInv.PaymentStatus)
	}

	// 5. Complete Final Payment (Pay remaining 650)
	finalPayCmd := ports.CreateSupplierPaymentCmd{
		LegalEntityID:         leID,
		SupplierID:            supp.ID,
		PaymentDate:           time.Now().Format("2006-01-02"),
		PaymentMethod:         domain.PaymentMethodBankTransfer,
		DisbursementAccountID: bankAccountID,
		Amount:                expectedRemaining,
		Allocations: []ports.PaymentAllocationCmd{
			{
				OpenItemID: openItem.ID,
				Amount:     expectedRemaining,
			},
		},
	}

	finalPayment, err := payableSvc.CreateSupplierPayment(ctx, finalPayCmd)
	if err != nil {
		t.Fatalf("failed to create final payment: %v", err)
	}
	_, err = payableSvc.PostSupplierPayment(ctx, finalPayment.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to post final payment: %v", err)
	}

	// 6. Verify Invoice and Open Item are PAID
	refreshedInv2, err := payableSvc.GetSupplierInvoice(ctx, postedInv.ID)
	if err != nil {
		t.Fatalf("failed to fetch final invoice: %v", err)
	}
	if refreshedInv2.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("expected invoice payment status PAID, got %s", refreshedInv2.PaymentStatus)
	}

	itemsAfterFull, err := payableSvc.ListOpenItems(ctx, supp.ID, false)
	if err != nil || len(itemsAfterFull) == 0 {
		t.Fatalf("failed to fetch open items after full payment")
	}
	if itemsAfterFull[0].Status != domain.OpenItemStatusPaid {
		t.Errorf("expected open item status PAID, got %s", itemsAfterFull[0].Status)
	}
	if !itemsAfterFull[0].RemainingAmount.IsZero() {
		t.Errorf("expected 0 remaining amount, got %s", itemsAfterFull[0].RemainingAmount)
	}
}

func TestPayables_CreditNoteLifecycle(t *testing.T) {
	payMod, db, leID, adminID, apAccountID, _, expAccountID, ctx := setupPayablesTest(t)
	defer db.Close()

	suppSvc := payMod.SupplierService
	payableSvc := payMod.PayableService

	// 1. Create Supplier
	taxID := "310776655400003"
	supp, err := suppSvc.CreateSupplier(ctx, ports.CreateSupplierCmd{
		LegalEntityID: leID,
		SupplierType:  domain.SupplierTypeLocal,
		NameEn:        "National Packing Materials Co.",
		NameAr:        "الشركة الوطنية لمواد التعبئة",
		TaxID:         &taxID,
		Address: domain.SaudiAddress{
			City: "Dammam",
		},
		DefaultExpenseAccountID:   &expAccountID,
		DefaultLiabilityAccountID: &apAccountID,
	})
	if err != nil {
		t.Fatalf("failed to create supplier: %v", err)
	}

	// 2. Create Credit Note
	// Line 1: 2 units @ 150 = 300 subtotal, 45 tax = 345 total
	cnCmd := ports.CreateCreditNoteCmd{
		LegalEntityID:            leID,
		SupplierCreditNoteNumber: strPtr("VCN-4556"),
		SupplierID:               supp.ID,
		CreditNoteDate:           time.Now().Format("2006-01-02"),
		CurrencyCode:             "SAR",
		Reason:                   strPtr("Damaged goods return"),
		CreatedByUserID:          &adminID,
		Lines: []ports.CreateCreditNoteLineCmd{
			{
				ExpenseAccountID: expAccountID,
				Description:      "Return of damaged carton packing",
				Quantity:         decimal.NewFromInt(2),
				UnitPrice:        decimal.NewFromInt(150),
				TaxRate:          decimal.NewFromInt(15),
			},
		},
	}

	cn, err := payableSvc.CreateSupplierCreditNote(ctx, cnCmd)
	if err != nil {
		t.Fatalf("failed to create credit note: %v", err)
	}

	expectedTotal := decimal.NewFromInt(345)
	if !cn.TotalAmount.Equal(expectedTotal) {
		t.Errorf("expected total amount %s, got %s", expectedTotal, cn.TotalAmount)
	}
	if cn.Status != "POSTED" {
		t.Errorf("expected status POSTED, got %s", cn.Status)
	}

	// 3. List Credit Notes
	notes, err := payableSvc.ListSupplierCreditNotes(ctx, leID, &supp.ID)
	if err != nil {
		t.Fatalf("failed to list credit notes: %v", err)
	}
	if len(notes) == 0 {
		t.Errorf("expected at least 1 credit note, got 0")
	}
}

func TestPayables_SupplierAgingSummary(t *testing.T) {
	payMod, db, leID, _, apAccountID, _, expAccountID, ctx := setupPayablesTest(t)
	defer db.Close()

	suppSvc := payMod.SupplierService
	payableSvc := payMod.PayableService

	// 1. Create Supplier
	taxID := "310665544300003"
	supp, err := suppSvc.CreateSupplier(ctx, ports.CreateSupplierCmd{
		LegalEntityID: leID,
		SupplierType:  domain.SupplierTypeLocal,
		NameEn:        "Saudi Logistics Transport Services",
		NameAr:        "خدمات النقل اللوجستي السعودي",
		TaxID:         &taxID,
		Address: domain.SaudiAddress{
			City: "Jeddah",
		},
		DefaultExpenseAccountID:   &expAccountID,
		DefaultLiabilityAccountID: &apAccountID,
	})
	if err != nil {
		t.Fatalf("failed to create supplier: %v", err)
	}

	// Helper to insert test open items with specific due dates directly into payables.open_items
	insertOpenItem := func(amount decimal.Decimal, daysOffset int) {
		dueDate := time.Now().AddDate(0, 0, daysOffset)
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO payables.open_items (
				id, legal_entity_id, supplier_id, document_type, document_number,
				document_date, due_date, currency_code, original_amount, allocated_amount,
				remaining_amount, status
			) VALUES (
				uuidv7(), $1, $2, 'INVOICE', $3, CURRENT_DATE, $4, 'SAR', $5, 0, $5, 'OPEN'
			)
		`, leID, supp.ID, fmt.Sprintf("DOC-%s", uuid.New().String()[:12]), dueDate, amount)
		if err != nil {
			t.Fatalf("failed to insert open item for aging test: %v", err)
		}
	}

	// A. Current (due in future, +10 days): 1000 SAR
	insertOpenItem(decimal.NewFromInt(1000), 10)
	// B. Overdue 1-30 days (-15 days): 2000 SAR
	insertOpenItem(decimal.NewFromInt(2000), -15)
	// C. Overdue 31-60 days (-45 days): 3000 SAR
	insertOpenItem(decimal.NewFromInt(3000), -45)
	// D. Overdue 61-90 days (-75 days): 4000 SAR
	insertOpenItem(decimal.NewFromInt(4000), -75)
	// E. Overdue > 90 days (-120 days): 5000 SAR
	insertOpenItem(decimal.NewFromInt(5000), -120)

	// 2. Query Aging Summary
	summary, err := payableSvc.GetSupplierAgingSummary(ctx, supp.ID)
	if err != nil {
		t.Fatalf("failed to get supplier aging summary: %v", err)
	}

	if !summary.CurrentAmount.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("expected CurrentAmount 1000, got %s", summary.CurrentAmount)
	}
	if !summary.Overdue1to30.Equal(decimal.NewFromInt(2000)) {
		t.Errorf("expected Overdue1to30 2000, got %s", summary.Overdue1to30)
	}
	if !summary.Overdue31to60.Equal(decimal.NewFromInt(3000)) {
		t.Errorf("expected Overdue31to60 3000, got %s", summary.Overdue31to60)
	}
	if !summary.Overdue61to90.Equal(decimal.NewFromInt(4000)) {
		t.Errorf("expected Overdue61to90 4000, got %s", summary.Overdue61to90)
	}
	if !summary.OverdueOver90.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected OverdueOver90 5000, got %s", summary.OverdueOver90)
	}
	if !summary.TotalOutstanding.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected TotalOutstanding 15000, got %s", summary.TotalOutstanding)
	}
}

func strPtr(s string) *string {
	return &s
}
