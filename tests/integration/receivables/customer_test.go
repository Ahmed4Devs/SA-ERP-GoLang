package receivables_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/receivables"
	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupReceivablesTest(t *testing.T) (*receivables.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("PostgreSQL not accessible: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	recMod := receivables.NewModule(cfg, db)

	// Ensure organization & legal entity exist
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// Seed accounting defaults (Chart of accounts)
	_ = accMod.SeedDefaultCOA(ctx, le.ID)

	// Seed receivables defaults
	_ = recMod.SeedDefaultReceivablesData(ctx, le.ID)

	// Get admin user
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	}

	// Lookup AR Control Account (112000) and Bank Account (111002)
	var arAccountID, bankAccountID uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '112000' LIMIT 1", le.ID).Scan(&arAccountID)
	if err != nil {
		t.Fatalf("failed to find AR account: %v", err)
	}
	err = db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '111002' LIMIT 1", le.ID).Scan(&bankAccountID)
	if err != nil {
		t.Fatalf("failed to find Bank account: %v", err)
	}

	return recMod, db, le.ID, adminID, arAccountID, bankAccountID, ctx
}

func TestCustomer_CreationAndSaudiValidation(t *testing.T) {
	recMod, db, leID, _, arAccountID, _, ctx := setupReceivablesTest(t)
	defer db.Close()

	custSvc := recMod.CustomerService

	validTaxID := "310123456700003"
	validCR := "1010897654"

	// 1. Positive: Create Valid B2B Customer
	cmd := ports.CreateCustomerCommand{
		LegalEntityID: leID,
		CustomerType:  domain.CustomerTypeB2B,
		NameEn:        "Al-Yamama Engineering Consultancies",
		NameAr:        "شركة اليمامة للاستشارات الهندسية",
		TaxID:         &validTaxID,
		CRNumber:      &validCR,
		Address: domain.SaudiAddress{
			BuildingNo:   "4512",
			StreetName:   "Takhassusi Street",
			District:     "Al-Mohammadiyah",
			City:         "Riyadh",
			PostalCode:   "12362",
			AdditionalNo: "8891",
			CountryCode:  "SA",
		},
		Email:            strPtr("info@alyamama.sa"),
		Phone:            strPtr("+966112345678"),
		ContactPerson:    strPtr("م. فهد القحطاني"),
		CurrencyCode:     "SAR",
		DefaultAccountID: &arAccountID,
		CreditLimit:      decimal.NewFromInt(150000),
		MaxOverdueDays:   30,
		BlockAction:      domain.BlockActionBlock,
	}

	customer, err := custSvc.CreateCustomer(ctx, cmd)
	if err != nil {
		t.Fatalf("expected customer creation to succeed, got error: %v", err)
	}

	if customer.ID == uuid.Nil {
		t.Errorf("expected valid customer ID, got Nil")
	}
	if customer.CustomerNumber == "" {
		t.Errorf("expected generated customer number")
	}
	if customer.Status != domain.CustomerStatusActive {
		t.Errorf("expected ACTIVE status, got %s", customer.Status)
	}

	// 2. Negative: Invalid Saudi Tax ID (not 15 digits or not starting/ending with 3)
	badTaxID := "123456789012345"
	badCmd := cmd
	badCmd.TaxID = &badTaxID
	_, err = custSvc.CreateCustomer(ctx, badCmd)
	if !errors.Is(err, domain.ErrInvalidTaxID) {
		t.Errorf("expected ErrInvalidTaxID, got: %v", err)
	}

	// 3. Negative: Invalid CR number (must be 10 digits)
	badCR := "12345"
	badCmd2 := cmd
	badCmd2.TaxID = &validTaxID
	badCmd2.CRNumber = &badCR
	_, err = custSvc.CreateCustomer(ctx, badCmd2)
	if !errors.Is(err, domain.ErrInvalidCRNumber) {
		t.Errorf("expected ErrInvalidCRNumber, got: %v", err)
	}

	// 4. Negative: Missing City
	badCmd3 := cmd
	badCmd3.TaxID = &validTaxID
	badCmd3.CRNumber = &validCR
	badCmd3.Address.City = ""
	_, err = custSvc.CreateCustomer(ctx, badCmd3)
	if !errors.Is(err, domain.ErrCityRequired) {
		t.Errorf("expected ErrCityRequired, got: %v", err)
	}
}

func TestCustomer_CreditLimitAndHold(t *testing.T) {
	recMod, db, leID, _, arAccountID, _, ctx := setupReceivablesTest(t)
	defer db.Close()

	custSvc := recMod.CustomerService

	// Create customer with 50,000 SAR credit limit
	customer, err := custSvc.CreateCustomer(ctx, ports.CreateCustomerCommand{
		LegalEntityID:    leID,
		CustomerType:     domain.CustomerTypeB2B,
		NameEn:           "Najd Logistics Solutions",
		NameAr:           "شركة نجد للحلول اللوجستية",
		Address:          domain.SaudiAddress{City: "Riyadh", CountryCode: "SA"},
		DefaultAccountID: &arAccountID,
		CreditLimit:      decimal.NewFromInt(50000),
		BlockAction:      domain.BlockActionBlock,
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	// Fetch customer and check credit limit
	fetched, err := custSvc.GetCustomer(ctx, customer.ID)
	if err != nil {
		t.Fatalf("failed to get customer: %v", err)
	}
	if !fetched.CreditLimit.Equal(decimal.NewFromInt(50000)) {
		t.Errorf("expected credit limit 50000, got %s", fetched.CreditLimit)
	}
	if fetched.IsCreditHold {
		t.Errorf("expected is_credit_hold to be false")
	}

	// Set Credit Hold
	holdReason := "تعليق حساب العميل لمراجعة الضمانات البنكية"
	err = custSvc.SetCreditHold(ctx, customer.ID, true, &holdReason)
	if err != nil {
		t.Fatalf("failed to set credit hold: %v", err)
	}

	fetched2, err := custSvc.GetCustomer(ctx, customer.ID)
	if err != nil {
		t.Fatalf("failed to get customer: %v", err)
	}
	if !fetched2.IsCreditHold {
		t.Errorf("expected customer to be on credit hold")
	}

	// Release Credit Hold
	err = custSvc.SetCreditHold(ctx, customer.ID, false, nil)
	if err != nil {
		t.Fatalf("failed to release credit hold: %v", err)
	}
}

func TestReceivables_PaymentTerms(t *testing.T) {
	recMod, db, leID, _, _, _, ctx := setupReceivablesTest(t)
	defer db.Close()

	custSvc := recMod.CustomerService

	// Create custom payment term NET45
	termCode := fmt.Sprintf("NET45-%d", time.Now().UnixNano()%1000000)
	term, err := custSvc.CreatePaymentTerm(ctx, &domain.PaymentTerm{
		LegalEntityID:   leID,
		Code:            termCode,
		NameEn:          "Net 45 Days",
		NameAr:          "أجل 45 يوماً",
		DueDays:         45,
		DiscountDays:    0,
		DiscountPercent: decimal.Zero,
		IsActive:        true,
	})
	if err != nil {
		t.Fatalf("failed to create payment term: %v", err)
	}
	if term.ID == uuid.Nil {
		t.Errorf("expected valid term ID")
	}

	// List payment terms
	terms, err := custSvc.ListPaymentTerms(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list payment terms: %v", err)
	}
	if len(terms) < 4 {
		t.Errorf("expected at least 4 payment terms, got %d", len(terms))
	}
}

func TestReceivables_OpenItemLifecycle(t *testing.T) {
	recMod, db, leID, _, arAccountID, _, ctx := setupReceivablesTest(t)
	defer db.Close()

	custSvc := recMod.CustomerService
	recSvc := recMod.ReceivablesService

	// Create customer
	customer, err := custSvc.CreateCustomer(ctx, ports.CreateCustomerCommand{
		LegalEntityID:    leID,
		CustomerType:     domain.CustomerTypeB2B,
		NameEn:           "Horizon Distribution Co.",
		NameAr:           "شركة آفاق التوزيع المحدودة",
		Address:          domain.SaudiAddress{City: "Jeddah", CountryCode: "SA"},
		DefaultAccountID: &arAccountID,
		CreditLimit:      decimal.NewFromInt(100000),
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	// 1. Create Invoice Open Item (15,000 SAR)
	invoiceDate := time.Now().AddDate(0, 0, -5)
	dueDate := time.Now().AddDate(0, 0, 25)
	docNum := fmt.Sprintf("INV-LC-%d", time.Now().UnixNano()%1000000)
	openItem, err := recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: docNum,
		DocumentDate:   invoiceDate,
		DueDate:        dueDate,
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(15000),
	})
	if err != nil {
		t.Fatalf("failed to create open item: %v", err)
	}

	if openItem.Status != domain.OpenItemStatusOpen {
		t.Errorf("expected status OPEN, got %s", openItem.Status)
	}
	if !openItem.RemainingAmount.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected remaining amount 15000, got %s", openItem.RemainingAmount)
	}

	// Verify customer balance increased
	custFetched, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !custFetched.CurrentBalance.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected customer current balance 15000, got %s", custFetched.CurrentBalance)
	}
}

func TestReceivables_ReceiptCreationAndAllocation(t *testing.T) {
	recMod, db, leID, adminID, arAccountID, bankAccountID, ctx := setupReceivablesTest(t)
	defer db.Close()

	custSvc := recMod.CustomerService
	recSvc := recMod.ReceivablesService

	// Create customer
	customer, err := custSvc.CreateCustomer(ctx, ports.CreateCustomerCommand{
		LegalEntityID:    leID,
		CustomerType:     domain.CustomerTypeB2B,
		NameEn:           "Gulf Electro-Mechanical Co.",
		NameAr:           "شركة الخليج الكهروميكانيكية",
		Address:          domain.SaudiAddress{City: "Khobar", CountryCode: "SA"},
		DefaultAccountID: &arAccountID,
		CreditLimit:      decimal.NewFromInt(200000),
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	// Create 2 open items:
	// Item 1: 10,000 SAR
	docNum1 := fmt.Sprintf("INV-REC1-%d", time.Now().UnixNano()%1000000)
	item1, err := recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: docNum1,
		DocumentDate:   time.Now().AddDate(0, 0, -10),
		DueDate:        time.Now().AddDate(0, 0, 20),
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(10000),
	})
	if err != nil {
		t.Fatalf("failed to create open item 1: %v", err)
	}

	// Item 2: 20,000 SAR
	docNum2 := fmt.Sprintf("INV-REC2-%d", time.Now().UnixNano()%1000000)
	item2, err := recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: docNum2,
		DocumentDate:   time.Now().AddDate(0, 0, -5),
		DueDate:        time.Now().AddDate(0, 0, 25),
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(20000),
	})
	if err != nil {
		t.Fatalf("failed to create open item 2: %v", err)
	}

	// Check customer total balance = 30,000 SAR
	cCheck, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !cCheck.CurrentBalance.Equal(decimal.NewFromInt(30000)) {
		t.Errorf("expected customer balance 30000, got %s", cCheck.CurrentBalance)
	}

	// Create Receipt Voucher of 15,000 SAR:
	// - 10,000 fully pays Item 1
	// - 5,000 partially pays Item 2
	receipt, err := recSvc.CreateReceipt(ctx, ports.CreateReceiptCommand{
		LegalEntityID:    leID,
		CustomerID:       customer.ID,
		ReceiptDate:      time.Now().Format("2006-01-02"),
		PaymentMethod:    domain.PaymentMethodBankTransfer,
		DepositAccountID: bankAccountID,
		Amount:           decimal.NewFromInt(15000),
		CurrencyCode:     "SAR",
		ReferenceNumber:  strPtr("TRF-SNB-112233"),
		Notes:            strPtr("سداد الفاتورة 8001 وجزء من الفاتورة 8002"),
		CreatedByUserID:  &adminID,
		Allocations: []ports.AllocationCommand{
			{OpenItemID: item1.ID, AllocatedAmount: decimal.NewFromInt(10000)},
			{OpenItemID: item2.ID, AllocatedAmount: decimal.NewFromInt(5000)},
		},
	})
	if err != nil {
		t.Fatalf("failed to create receipt with allocations: %v", err)
	}

	if receipt.Status != domain.ReceiptStatusPosted {
		t.Errorf("expected POSTED receipt status, got %s", receipt.Status)
	}
	if !receipt.AllocatedAmount.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected allocated amount 15000, got %s", receipt.AllocatedAmount)
	}
	if !receipt.UnallocatedAmount.IsZero() {
		t.Errorf("expected unallocated amount 0, got %s", receipt.UnallocatedAmount)
	}

	// Verify Item 1 is fully PAID
	openItems, err := recSvc.ListOpenItems(ctx, customer.ID, false)
	if err != nil {
		t.Fatalf("failed to list open items: %v", err)
	}

	var foundItem1, foundItem2 *domain.OpenItem
	for _, it := range openItems {
		if it.ID == item1.ID {
			foundItem1 = it
		}
		if it.ID == item2.ID {
			foundItem2 = it
		}
	}

	if foundItem1.Status != domain.OpenItemStatusPaid {
		t.Errorf("expected item 1 status PAID, got %s", foundItem1.Status)
	}
	if !foundItem1.RemainingAmount.IsZero() {
		t.Errorf("expected item 1 remaining amount 0, got %s", foundItem1.RemainingAmount)
	}

	// Verify Item 2 is PARTIALLY_PAID with 15,000 SAR remaining
	if foundItem2.Status != domain.OpenItemStatusPartiallyPaid {
		t.Errorf("expected item 2 status PARTIALLY_PAID, got %s", foundItem2.Status)
	}
	if !foundItem2.RemainingAmount.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected item 2 remaining amount 15000, got %s", foundItem2.RemainingAmount)
	}

	// Verify customer balance decreased by 15,000 SAR (now 15,000)
	cCheck2, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !cCheck2.CurrentBalance.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected customer balance 15000, got %s", cCheck2.CurrentBalance)
	}
}

func TestReceivables_CustomerAgingSummary(t *testing.T) {
	recMod, db, leID, _, arAccountID, _, ctx := setupReceivablesTest(t)
	defer db.Close()

	custSvc := recMod.CustomerService
	recSvc := recMod.ReceivablesService

	// Create customer
	customer, err := custSvc.CreateCustomer(ctx, ports.CreateCustomerCommand{
		LegalEntityID:    leID,
		CustomerType:     domain.CustomerTypeB2B,
		NameEn:           "Rawabi Commercial Services",
		NameAr:           "خدمات الروابي التجارية",
		Address:          domain.SaudiAddress{City: "Dammam", CountryCode: "SA"},
		DefaultAccountID: &arAccountID,
		CreditLimit:      decimal.NewFromInt(500000),
	})
	if err != nil {
		t.Fatalf("failed to create customer: %v", err)
	}

	now := time.Now()

	// 1. Current invoice (not overdue, due in 10 days) -> 5,000 SAR
	doc1 := fmt.Sprintf("INV-AG1-%d", time.Now().UnixNano()%1000000)
	_, err = recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: doc1,
		DocumentDate:   now.AddDate(0, 0, -5),
		DueDate:        now.AddDate(0, 0, 10),
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(5000),
	})
	if err != nil {
		t.Fatalf("failed to create open item 1: %v", err)
	}

	// 2. Overdue 1-30 days (due 15 days ago) -> 12,000 SAR
	doc2 := fmt.Sprintf("INV-AG2-%d", time.Now().UnixNano()%1000000)
	_, err = recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: doc2,
		DocumentDate:   now.AddDate(0, 0, -45),
		DueDate:        now.AddDate(0, 0, -15),
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(12000),
	})
	if err != nil {
		t.Fatalf("failed to create open item 2: %v", err)
	}

	// 3. Overdue 31-60 days (due 45 days ago) -> 8,000 SAR
	doc3 := fmt.Sprintf("INV-AG3-%d", time.Now().UnixNano()%1000000)
	_, err = recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: doc3,
		DocumentDate:   now.AddDate(0, 0, -75),
		DueDate:        now.AddDate(0, 0, -45),
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(8000),
	})
	if err != nil {
		t.Fatalf("failed to create open item 3: %v", err)
	}

	// 4. Overdue > 90 days (due 110 days ago) -> 20,000 SAR
	doc4 := fmt.Sprintf("INV-AG4-%d", time.Now().UnixNano()%1000000)
	_, err = recSvc.CreateOpenItem(ctx, &domain.OpenItem{
		LegalEntityID:  leID,
		CustomerID:     customer.ID,
		DocumentType:   domain.DocumentTypeInvoice,
		DocumentNumber: doc4,
		DocumentDate:   now.AddDate(0, 0, -140),
		DueDate:        now.AddDate(0, 0, -110),
		CurrencyCode:   "SAR",
		OriginalAmount: decimal.NewFromInt(20000),
	})
	if err != nil {
		t.Fatalf("failed to create open item 4: %v", err)
	}

	// Fetch Aging Summary
	aging, err := recSvc.GetCustomerAging(ctx, customer.ID)
	if err != nil {
		t.Fatalf("failed to get customer aging: %v", err)
	}

	totalExpected := decimal.NewFromInt(45000)
	if !aging.TotalOutstanding.Equal(totalExpected) {
		t.Errorf("expected total outstanding 45000, got %s", aging.TotalOutstanding)
	}
	if !aging.CurrentAmount.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected current amount 5000, got %s", aging.CurrentAmount)
	}
	if !aging.Overdue1to30.Equal(decimal.NewFromInt(12000)) {
		t.Errorf("expected overdue 1-30 12000, got %s", aging.Overdue1to30)
	}
	if !aging.Overdue31to60.Equal(decimal.NewFromInt(8000)) {
		t.Errorf("expected overdue 31-60 8000, got %s", aging.Overdue31to60)
	}
	if !aging.OverdueOver90.Equal(decimal.NewFromInt(20000)) {
		t.Errorf("expected overdue >90 20000, got %s", aging.OverdueOver90)
	}
}

func strPtr(s string) *string {
	return &s
}
