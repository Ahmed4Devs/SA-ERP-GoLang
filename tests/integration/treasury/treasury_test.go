package treasury_test

import (
	"context"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/banking"
	bankDomain "erp/internal/banking/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/treasury"
	"erp/internal/treasury/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupTreasuryTest(t *testing.T) (*treasury.Module, *banking.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
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
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	treasuryMod := treasury.NewModule(cfg, db)
	bankingMod := banking.NewModule(cfg, db)

	// 1. Organization & Legal Entity
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Foundation COA & Treasury Defaults
	_ = accMod.SeedDefaultCOA(ctx, le.ID)
	_ = treasuryMod.SeedDefaultTreasuryData(ctx, le.ID)

	// 3. Admin user
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	}

	return treasuryMod, bankingMod, db, le.ID, adminID, ctx
}

func TestCashAccount_CreationAndLimits(t *testing.T) {
	treasuryMod, _, _, leID, adminID, ctx := setupTreasuryTest(t)
	svc := treasuryMod.TreasuryService

	// Create custom branch cash register
	acc, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID:   leID,
		AccountCode:     "CASH-TEST-" + uuid.New().String()[:8],
		NameEn:          "Test Branch Register",
		NameAr:          "صندوق الفرع التجريبي",
		AccountType:     domain.CashAccountCashRegister,
		CurrencyCode:    "SAR",
		CustodianUserID: &adminID,
		MaxLimit:        decimal.NewFromInt(5000),
		CurrentBalance:  decimal.NewFromInt(1000),
		Status:          domain.CashAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}
	if acc.ID == uuid.Nil {
		t.Fatalf("expected non-nil cash account ID")
	}

	fetched, err := svc.GetCashAccount(ctx, leID, acc.ID)
	if err != nil {
		t.Fatalf("failed to get cash account: %v", err)
	}
	if !fetched.CurrentBalance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("expected balance 1000, got %s", fetched.CurrentBalance.String())
	}
	if !fetched.MaxLimit.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected max limit 5000, got %s", fetched.MaxLimit.String())
	}
}

func TestTreasuryReceipt_Lifecycle(t *testing.T) {
	treasuryMod, _, _, leID, adminID, ctx := setupTreasuryTest(t)
	svc := treasuryMod.TreasuryService

	// 1. Create a cash account with 1000 SAR balance and 2000 SAR limit
	acc, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID:   leID,
		AccountCode:     "CASH-REC-" + uuid.New().String()[:8],
		NameEn:          "Receipt Test Box",
		NameAr:          "صندوق اختبار القبض",
		AccountType:     domain.CashAccountPettyCash,
		CurrencyCode:    "SAR",
		CustodianUserID: &adminID,
		MaxLimit:        decimal.NewFromInt(2000),
		CurrentBalance:  decimal.NewFromInt(1000),
		Status:          domain.CashAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}

	// 2. Create Receipt for 400 SAR (Within limit: 1000 + 400 = 1400 <= 2000)
	rc, err := svc.CreateReceipt(ctx, &domain.Receipt{
		LegalEntityID:   leID,
		CashAccountID:   acc.ID,
		ReceiptDate:     time.Now(),
		ReceiptType:     domain.ReceiptMiscellaneous,
		ReceivedFrom:    "عميل مبيعات نقدية",
		Amount:          decimal.NewFromInt(400),
		CurrencyCode:    "SAR",
		PaymentMethod:   domain.PaymentMethodCash,
		CreatedByUserID: &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create receipt: %v", err)
	}
	if rc.Status != domain.ReceiptDraft {
		t.Errorf("expected status DRAFT, got %s", rc.Status)
	}

	// 3. Post Receipt
	postedRc, err := svc.PostReceipt(ctx, leID, rc.ID, nil)
	if err != nil {
		t.Fatalf("failed to post receipt: %v", err)
	}
	if postedRc.Status != domain.ReceiptPosted {
		t.Errorf("expected status POSTED, got %s", postedRc.Status)
	}

	// 4. Verify account balance increased to 1400
	updatedAcc, err := svc.GetCashAccount(ctx, leID, acc.ID)
	if err != nil {
		t.Fatalf("failed to fetch account: %v", err)
	}
	if !updatedAcc.CurrentBalance.Equal(decimal.NewFromInt(1400)) {
		t.Errorf("expected balance 1400, got %s", updatedAcc.CurrentBalance.String())
	}

	// 5. Attempt Receipt that exceeds limit (1400 + 800 = 2200 > 2000)
	excessRc, err := svc.CreateReceipt(ctx, &domain.Receipt{
		LegalEntityID:   leID,
		CashAccountID:   acc.ID,
		ReceiptDate:     time.Now(),
		ReceiptType:     domain.ReceiptDeposit,
		ReceivedFrom:    "تغذية صندوق زائدة",
		Amount:          decimal.NewFromInt(800),
		CurrencyCode:    "SAR",
		PaymentMethod:   domain.PaymentMethodCash,
		CreatedByUserID: &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create excess receipt: %v", err)
	}
	_, err = svc.PostReceipt(ctx, leID, excessRc.ID, nil)
	if err == nil {
		t.Fatalf("expected error exceeding account limit, got nil")
	}
}

func TestTreasuryPayment_Lifecycle(t *testing.T) {
	treasuryMod, _, _, leID, adminID, ctx := setupTreasuryTest(t)
	svc := treasuryMod.TreasuryService

	// 1. Create a cash account with 500 SAR balance
	acc, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID:   leID,
		AccountCode:     "CASH-PMT-" + uuid.New().String()[:8],
		NameEn:          "Payment Test Box",
		NameAr:          "صندوق اختبار الصرف",
		AccountType:     domain.CashAccountPettyCash,
		CurrencyCode:    "SAR",
		CustodianUserID: &adminID,
		MaxLimit:        decimal.NewFromInt(5000),
		CurrentBalance:  decimal.NewFromInt(500),
		Status:          domain.CashAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}

	// 2. Create Payment for 200 SAR
	pmt, err := svc.CreatePayment(ctx, &domain.Payment{
		LegalEntityID:   leID,
		CashAccountID:   acc.ID,
		PaymentDate:     time.Now(),
		PaymentType:     domain.PaymentPettyCashExpense,
		PaidTo:          "مكتبة الهدى",
		Amount:          decimal.NewFromInt(200),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// 3. Approve Payment
	approved, err := svc.ApprovePayment(ctx, leID, pmt.ID, adminID)
	if err != nil {
		t.Fatalf("failed to approve payment: %v", err)
	}
	if approved.Status != domain.PaymentApproved {
		t.Errorf("expected status APPROVED, got %s", approved.Status)
	}

	// 4. Pay Payment
	paid, err := svc.PayPayment(ctx, leID, pmt.ID, nil)
	if err != nil {
		t.Fatalf("failed to pay payment: %v", err)
	}
	if paid.Status != domain.PaymentPaid {
		t.Errorf("expected status PAID, got %s", paid.Status)
	}

	// 5. Verify Cash Account Balance (500 - 200 = 300)
	updatedAcc, err := svc.GetCashAccount(ctx, leID, acc.ID)
	if err != nil {
		t.Fatalf("failed to get cash account: %v", err)
	}
	if !updatedAcc.CurrentBalance.Equal(decimal.NewFromInt(300)) {
		t.Errorf("expected balance 300, got %s", updatedAcc.CurrentBalance.String())
	}

	// 6. Attempt payment greater than balance (300 < 500)
	overPmt, err := svc.CreatePayment(ctx, &domain.Payment{
		LegalEntityID:   leID,
		CashAccountID:   acc.ID,
		PaymentDate:     time.Now(),
		PaymentType:     domain.PaymentPettyCashExpense,
		PaidTo:          "مورد صيانة",
		Amount:          decimal.NewFromInt(500),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create over payment: %v", err)
	}
	_, err = svc.PayPayment(ctx, leID, overPmt.ID, nil)
	if err == nil {
		t.Fatalf("expected error on insufficient cash balance, got nil")
	}
}

func TestTreasuryTransfer_CashToBank(t *testing.T) {
	treasuryMod, bankingMod, _, leID, adminID, ctx := setupTreasuryTest(t)
	tSvc := treasuryMod.TreasuryService
	bSvc := bankingMod.BankingService

	// 1. Create Cash Account with 3000 SAR
	cashAcc, err := tSvc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID:   leID,
		AccountCode:     "CASH-TRF-" + uuid.New().String()[:8],
		NameEn:          "Source Cash Register",
		NameAr:          "صندوق تحويل تجريبي",
		AccountType:     domain.CashAccountCashRegister,
		CurrencyCode:    "SAR",
		CustodianUserID: &adminID,
		MaxLimit:        decimal.NewFromInt(10000),
		CurrentBalance:  decimal.NewFromInt(3000),
		Status:          domain.CashAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}

	// 2. Create Destination Bank Account
	randSuffix := uuid.New().String()[:8]
	bankAcc, err := bSvc.CreateBankAccount(ctx, &bankDomain.BankAccount{
		LegalEntityID:  leID,
		AccountCode:    "BANK-TRF-" + randSuffix,
		BankNameEn:     "Riyad Bank",
		BankNameAr:     "بنك الرياض",
		AccountNumber:  "9900" + randSuffix,
		IBAN:           "SA55200009900" + randSuffix,
		SwiftBIC:       "RIBLSARI",
		CurrencyCode:   "SAR",
		OpeningBalance: decimal.NewFromInt(10000),
		CurrentBalance: decimal.NewFromInt(10000),
		Status:         bankDomain.BankAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	// 3. Initiate Transfer: 1000 SAR + 15 SAR fee
	trf, err := tSvc.CreateTransfer(ctx, &domain.Transfer{
		LegalEntityID:            leID,
		TransferDate:             time.Now(),
		SourceType:               domain.TransferSourceCash,
		SourceCashAccountID:      &cashAcc.ID,
		DestinationType:          domain.TransferDestinationBank,
		DestinationBankAccountID: &bankAcc.ID,
		Amount:                   decimal.NewFromInt(1000),
		FeeAmount:                decimal.NewFromInt(15),
		CurrencyCode:             "SAR",
		InitiatedByUserID:        &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}

	// 4. Complete Transfer
	completed, err := tSvc.CompleteTransfer(ctx, leID, trf.ID, adminID, nil)
	if err != nil {
		t.Fatalf("failed to complete transfer: %v", err)
	}
	if completed.Status != domain.TransferCompleted {
		t.Errorf("expected status COMPLETED, got %s", completed.Status)
	}

	// 5. Verify Cash Account Balance (3000 - 1000 - 15 = 1985)
	updatedCash, err := tSvc.GetCashAccount(ctx, leID, cashAcc.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated cash account: %v", err)
	}
	if !updatedCash.CurrentBalance.Equal(decimal.NewFromInt(1985)) {
		t.Errorf("expected cash balance 1985, got %s", updatedCash.CurrentBalance.String())
	}
}
