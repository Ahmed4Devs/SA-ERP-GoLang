package treasury_test

import (
	"context"
	"errors"
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
	t.Cleanup(db.Close)

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
	if _, err := svc.PostReceipt(ctx, leID, rc.ID, nil); err != nil {
		t.Fatalf("idempotent receipt post failed: %v", err)
	}
	updatedAcc, err = svc.GetCashAccount(ctx, leID, acc.ID)
	if err != nil {
		t.Fatalf("failed to fetch account after repeated post: %v", err)
	}
	if !updatedAcc.CurrentBalance.Equal(decimal.NewFromInt(1400)) {
		t.Fatalf("repeated receipt post changed balance to %s", updatedAcc.CurrentBalance)
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

func TestTreasuryPayment_ConcurrentPaymentsCannotOverdraw(t *testing.T) {
	treasuryMod, _, _, leID, adminID, ctx := setupTreasuryTest(t)
	svc := treasuryMod.TreasuryService

	account, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID:   leID,
		AccountCode:     "CASH-RACE-" + uuid.New().String()[:8],
		NameEn:          "Concurrent Payment Box",
		NameAr:          "صندوق اختبار التزامن",
		AccountType:     domain.CashAccountPettyCash,
		CurrencyCode:    "SAR",
		CustodianUserID: &adminID,
		MaxLimit:        decimal.NewFromInt(1000),
		CurrentBalance:  decimal.NewFromInt(100),
		Status:          domain.CashAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}

	payments := make([]*domain.Payment, 2)
	for i := range payments {
		payments[i], err = svc.CreatePayment(ctx, &domain.Payment{
			LegalEntityID:   leID,
			CashAccountID:   account.ID,
			PaymentDate:     time.Now(),
			PaymentType:     domain.PaymentMiscellaneous,
			PaidTo:          "Concurrent payee",
			Amount:          decimal.NewFromInt(80),
			CurrencyCode:    "SAR",
			CreatedByUserID: &adminID,
		})
		if err != nil {
			t.Fatalf("failed to create payment: %v", err)
		}
	}

	start := make(chan struct{})
	results := make(chan error, len(payments))
	for _, payment := range payments {
		payment := payment
		go func() {
			<-start
			_, err := svc.PayPayment(ctx, leID, payment.ID, nil)
			results <- err
		}()
	}
	close(start)

	successes := 0
	insufficient := 0
	for range payments {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrInsufficientCashBalance):
			insufficient++
		default:
			t.Fatalf("unexpected payment result: %v", err)
		}
	}
	if successes != 1 || insufficient != 1 {
		t.Fatalf("expected one payment and one insufficient-balance error, got successes=%d insufficient=%d", successes, insufficient)
	}

	updated, err := svc.GetCashAccount(ctx, leID, account.ID)
	if err != nil {
		t.Fatalf("failed to reload cash account: %v", err)
	}
	if !updated.CurrentBalance.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("expected final balance 20, got %s", updated.CurrentBalance)
	}
}

func TestTreasuryReceipt_PostRollbackRestoresBalance(t *testing.T) {
	treasuryMod, _, _, leID, adminID, ctx := setupTreasuryTest(t)
	svc := treasuryMod.TreasuryService

	account, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID:   leID,
		AccountCode:     "CASH-ROLLBACK-" + uuid.New().String()[:8],
		NameEn:          "Receipt Rollback Box",
		NameAr:          "صندوق تراجع القبض",
		AccountType:     domain.CashAccountPettyCash,
		CurrencyCode:    "SAR",
		CustodianUserID: &adminID,
		MaxLimit:        decimal.NewFromInt(1000),
		CurrentBalance:  decimal.NewFromInt(100),
		Status:          domain.CashAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create cash account: %v", err)
	}
	receipt, err := svc.CreateReceipt(ctx, &domain.Receipt{
		LegalEntityID:   leID,
		CashAccountID:   account.ID,
		ReceiptDate:     time.Now(),
		ReceiptType:     domain.ReceiptMiscellaneous,
		ReceivedFrom:    "Rollback payer",
		Amount:          decimal.NewFromInt(25),
		CurrencyCode:    "SAR",
		PaymentMethod:   domain.PaymentMethodCash,
		CreatedByUserID: &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create receipt: %v", err)
	}

	invalidJournalID := uuid.New()
	if _, err := svc.PostReceipt(ctx, leID, receipt.ID, &invalidJournalID); err == nil {
		t.Fatal("expected invalid journal entry to abort receipt posting")
	}
	updatedAccount, err := svc.GetCashAccount(ctx, leID, account.ID)
	if err != nil {
		t.Fatalf("failed to reload cash account: %v", err)
	}
	if !updatedAccount.CurrentBalance.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected rolled-back balance 100, got %s", updatedAccount.CurrentBalance)
	}
	unchangedReceipt, err := svc.GetReceipt(ctx, leID, receipt.ID)
	if err != nil {
		t.Fatalf("failed to reload receipt: %v", err)
	}
	if unchangedReceipt.Status != domain.ReceiptDraft {
		t.Fatalf("expected receipt to remain DRAFT, got %s", unchangedReceipt.Status)
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
	updatedBank, err := bSvc.GetBankAccount(ctx, leID, bankAcc.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated bank account: %v", err)
	}
	if !updatedBank.CurrentBalance.Equal(decimal.NewFromInt(11000)) {
		t.Errorf("expected bank balance 11000, got %s", updatedBank.CurrentBalance.String())
	}
}

func TestTreasuryTransfer_RollbackRestoresBothCashAccounts(t *testing.T) {
	treasuryMod, _, _, leID, adminID, ctx := setupTreasuryTest(t)
	svc := treasuryMod.TreasuryService

	createAccount := func(code string, balance decimal.Decimal) *domain.CashAccount {
		t.Helper()
		account, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
			LegalEntityID:   leID,
			AccountCode:     code + uuid.New().String()[:8],
			NameEn:          "Transfer Rollback Box",
			NameAr:          "صندوق تراجع التحويل",
			AccountType:     domain.CashAccountMainVault,
			CurrencyCode:    "SAR",
			CustodianUserID: &adminID,
			MaxLimit:        decimal.NewFromInt(5000),
			CurrentBalance:  balance,
			Status:          domain.CashAccountActive,
		})
		if err != nil {
			t.Fatalf("failed to create cash account: %v", err)
		}
		return account
	}
	source := createAccount("CASH-TRF-RB-SRC-", decimal.NewFromInt(1000))
	destination := createAccount("CASH-TRF-RB-DST-", decimal.NewFromInt(100))
	transfer, err := svc.CreateTransfer(ctx, &domain.Transfer{
		LegalEntityID:            leID,
		TransferDate:             time.Now(),
		SourceType:               domain.TransferSourceCash,
		SourceCashAccountID:      &source.ID,
		DestinationType:          domain.TransferDestinationCash,
		DestinationCashAccountID: &destination.ID,
		Amount:                   decimal.NewFromInt(200),
		FeeAmount:                decimal.NewFromInt(10),
		CurrencyCode:             "SAR",
		InitiatedByUserID:        &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}

	invalidJournalID := uuid.New()
	if _, err := svc.CompleteTransfer(ctx, leID, transfer.ID, adminID, &invalidJournalID); err == nil {
		t.Fatal("expected invalid journal entry to abort transfer")
	}
	rolledBackSource, err := svc.GetCashAccount(ctx, leID, source.ID)
	if err != nil {
		t.Fatalf("failed to reload source account: %v", err)
	}
	rolledBackDestination, err := svc.GetCashAccount(ctx, leID, destination.ID)
	if err != nil {
		t.Fatalf("failed to reload destination account: %v", err)
	}
	if !rolledBackSource.CurrentBalance.Equal(decimal.NewFromInt(1000)) || !rolledBackDestination.CurrentBalance.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected balances to roll back to 1000/100, got %s/%s", rolledBackSource.CurrentBalance, rolledBackDestination.CurrentBalance)
	}
	unchangedTransfer, err := svc.GetTransfer(ctx, leID, transfer.ID)
	if err != nil {
		t.Fatalf("failed to reload transfer: %v", err)
	}
	if unchangedTransfer.Status != domain.TransferDraft {
		t.Fatalf("expected transfer to remain DRAFT, got %s", unchangedTransfer.Status)
	}
}
