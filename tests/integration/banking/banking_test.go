package banking_test

import (
	"context"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/banking"
	"erp/internal/banking/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupBankingTest(t *testing.T) (*banking.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
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
	bankingMod := banking.NewModule(cfg, db)

	// 1. Seed org & COA
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}
	_ = accMod.SeedDefaultCOA(ctx, le.ID)

	// 2. Admin user
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	}

	return bankingMod, db, le.ID, adminID, ctx
}

func TestBankAccount_MaskingAndManagement(t *testing.T) {
	bankingMod, _, leID, _, ctx := setupBankingTest(t)
	svc := bankingMod.BankingService

	code := "BANK-TEST-" + uuid.New().String()[:8]
	iban := "SA12200000" + uuid.New().String()[:12]
	accNum := "10203040506070"

	acc, err := svc.CreateBankAccount(ctx, &domain.BankAccount{
		LegalEntityID:  leID,
		AccountCode:    code,
		BankNameEn:     "Banque Saudi Fransi",
		BankNameAr:     "البنك السعودي الفرنسي",
		AccountNumber:  accNum,
		IBAN:           iban,
		SwiftBIC:       "BSFRSARI",
		CurrencyCode:   "SAR",
		OpeningBalance: decimal.NewFromInt(50000),
		Status:         domain.BankAccountActive,
	})
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	if acc.AccountNumberMasked != "**** **** **** 6070" {
		t.Errorf("expected masked account number '**** **** **** 6070', got '%s'", acc.AccountNumberMasked)
	}
	if !acc.CurrentBalance.Equal(decimal.NewFromInt(50000)) {
		t.Errorf("expected balance 50000, got %s", acc.CurrentBalance.String())
	}

	// Fetch by ID
	fetched, err := svc.GetBankAccount(ctx, leID, acc.ID)
	if err != nil {
		t.Fatalf("failed to get bank account: %v", err)
	}
	if fetched.AccountCode != code {
		t.Errorf("expected code %s, got %s", code, fetched.AccountCode)
	}
}

func TestBankStatement_ImportAndLines(t *testing.T) {
	bankingMod, _, leID, adminID, ctx := setupBankingTest(t)
	svc := bankingMod.BankingService

	// 1. Create bank account
	bankAcc, err := svc.CreateBankAccount(ctx, &domain.BankAccount{
		LegalEntityID:  leID,
		AccountCode:    "BANK-STMT-" + uuid.New().String()[:8],
		BankNameEn:     "Arab National Bank",
		BankNameAr:     "البنك العربي الوطني",
		AccountNumber:  "554433221100",
		IBAN:           "SA99400000554433221100",
		SwiftBIC:       "ARNBSARI",
		CurrencyCode:   "SAR",
		OpeningBalance: decimal.NewFromInt(10000),
	})
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	// 2. Prepare Statement & 2 Lines
	now := time.Now()
	lines := []*domain.StatementLine{
		{
			TransactionDate: now,
			Description:     "Client Payment Wire",
			ReferenceNumber: func(s string) *string { return &s }("WIRE-8891"),
			TransactionType: domain.TransactionCredit,
			Amount:          decimal.NewFromInt(3000),
		},
		{
			TransactionDate: now,
			Description:     "Supplier Online Transfer",
			ReferenceNumber: func(s string) *string { return &s }("SUP-WIRE-112"),
			TransactionType: domain.TransactionDebit,
			Amount:          decimal.NewFromInt(1200),
		},
	}

	filename := "anb_jan_2026.csv"
	stmt, err := svc.ImportBankStatement(ctx, &domain.Statement{
		LegalEntityID:    leID,
		BankAccountID:    bankAcc.ID,
		StatementNumber:  "ANB-STMT-001",
		StatementDate:    now,
		StartDate:        now.AddDate(0, 0, -30),
		EndDate:          now,
		OpeningBalance:   decimal.NewFromInt(10000),
		FileFormat:       domain.StatementFormatCSV,
		SourceFilename:   &filename,
		ImportedByUserID: &adminID,
	}, lines)
	if err != nil {
		t.Fatalf("failed to import statement: %v", err)
	}

	if !stmt.TotalCredits.Equal(decimal.NewFromInt(3000)) {
		t.Errorf("expected credits 3000, got %s", stmt.TotalCredits.String())
	}
	if !stmt.TotalDebits.Equal(decimal.NewFromInt(1200)) {
		t.Errorf("expected debits 1200, got %s", stmt.TotalDebits.String())
	}
	// Expected Closing: 10000 + 3000 - 1200 = 11800
	if !stmt.ClosingBalance.Equal(decimal.NewFromInt(11800)) {
		t.Errorf("expected closing balance 11800, got %s", stmt.ClosingBalance.String())
	}

	// Check lines
	fetchedLines, err := svc.ListStatementLines(ctx, stmt.ID)
	if err != nil {
		t.Fatalf("failed to list statement lines: %v", err)
	}
	if len(fetchedLines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(fetchedLines))
	}
}

func TestBankReconciliation_AutoMatch(t *testing.T) {
	bankingMod, _, leID, adminID, ctx := setupBankingTest(t)
	svc := bankingMod.BankingService

	// 1. Bank Account
	bankAcc, err := svc.CreateBankAccount(ctx, &domain.BankAccount{
		LegalEntityID:  leID,
		AccountCode:    "BANK-RECON-" + uuid.New().String()[:8],
		BankNameEn:     "Alinma Bank",
		BankNameAr:     "مصرف الإنماء",
		AccountNumber:  "887766554433",
		IBAN:           "SA77050000887766554433",
		SwiftBIC:       "INMASARI",
		CurrencyCode:   "SAR",
		OpeningBalance: decimal.NewFromInt(20000),
	})
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	// 2. Statement with 2 lines
	now := time.Now()
	ref1 := "REF-INV-9900"
	ref2 := "REF-MISC-7700"
	lines := []*domain.StatementLine{
		{
			TransactionDate: now,
			Description:     "Payment for Invoice 9900",
			ReferenceNumber: &ref1,
			TransactionType: domain.TransactionCredit,
			Amount:          decimal.NewFromInt(5000),
		},
		{
			TransactionDate: now,
			Description:     "Bank Service Fee",
			ReferenceNumber: &ref2,
			TransactionType: domain.TransactionDebit,
			Amount:          decimal.NewFromInt(25),
		},
	}

	stmt, err := svc.ImportBankStatement(ctx, &domain.Statement{
		LegalEntityID:   leID,
		BankAccountID:   bankAcc.ID,
		StatementNumber: "INMA-001",
		StatementDate:   now,
		StartDate:       now.AddDate(0, 0, -30),
		EndDate:         now,
		OpeningBalance:  decimal.NewFromInt(20000),
	}, lines)
	if err != nil {
		t.Fatalf("failed to import statement: %v", err)
	}

	// 3. Create Reconciliation Session
	recon, err := svc.CreateReconciliation(ctx, &domain.Reconciliation{
		LegalEntityID:          leID,
		BankAccountID:          bankAcc.ID,
		StatementID:            &stmt.ID,
		AsOfDate:               now,
		StatementEndingBalance: decimal.NewFromInt(24975),
		GLBookBalance:          decimal.NewFromInt(25000),
		ReconciledByUserID:     &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create reconciliation: %v", err)
	}

	// 4. Candidate Ledger Entries (1 exact match with ref, 1 match by amount + proximity)
	candidateInvoiceID := uuid.New()
	candidates := []domain.CandidateLedgerEntry{
		{
			EntityType:      domain.MatchedEntityReceipt,
			EntityID:        candidateInvoiceID,
			ReferenceNumber: "REF-INV-9900",
			Date:            now,
			Amount:          decimal.NewFromInt(5000),
			Description:     "AR Receipt for Customer Invoice 9900",
		},
	}

	// 5. Run Auto Match
	results, err := svc.RunAutoMatch(ctx, leID, recon.ID, candidates)
	if err != nil {
		t.Fatalf("failed to run auto-match: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 match result, got %d", len(results))
	}
	if results[0].MatchType != domain.MatchExact {
		t.Errorf("expected EXACT match type, got %s", results[0].MatchType)
	}
	if results[0].Confidence != 1.0 {
		t.Errorf("expected confidence 1.0, got %f", results[0].Confidence)
	}

	// Verify matches recorded in DB
	matches, err := svc.ListMatches(ctx, recon.ID)
	if err != nil {
		t.Fatalf("failed to list matches: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match record, got %d", len(matches))
	}
	if *matches[0].MatchedEntityID != candidateInvoiceID {
		t.Errorf("expected matched entity ID %s, got %s", candidateInvoiceID, matches[0].MatchedEntityID.String())
	}
}

func TestBankReconciliation_Completion(t *testing.T) {
	bankingMod, _, leID, adminID, ctx := setupBankingTest(t)
	svc := bankingMod.BankingService

	// 1. Bank Account
	bankAcc, err := svc.CreateBankAccount(ctx, &domain.BankAccount{
		LegalEntityID:  leID,
		AccountCode:    "BANK-LOCK-" + uuid.New().String()[:8],
		BankNameEn:     "Bank AlJazira",
		BankNameAr:     "بنك الجزيرة",
		AccountNumber:  "123123123123",
		IBAN:           "SA60600000123123123123",
		SwiftBIC:       "BJAZSARI",
		CurrencyCode:   "SAR",
		OpeningBalance: decimal.NewFromInt(15000),
	})
	if err != nil {
		t.Fatalf("failed to create bank account: %v", err)
	}

	// 2. Reconciliation session
	recon, err := svc.CreateReconciliation(ctx, &domain.Reconciliation{
		LegalEntityID:          leID,
		BankAccountID:          bankAcc.ID,
		AsOfDate:               time.Now(),
		StatementEndingBalance: decimal.NewFromInt(15000),
		GLBookBalance:          decimal.NewFromInt(15000),
		Status:                 domain.ReconciliationInProgress,
	})
	if err != nil {
		t.Fatalf("failed to create reconciliation: %v", err)
	}

	// 3. Lock Session
	locked, err := svc.LockReconciliation(ctx, leID, recon.ID, adminID)
	if err != nil {
		t.Fatalf("failed to lock reconciliation: %v", err)
	}
	if locked.Status != domain.ReconciliationLocked {
		t.Errorf("expected status LOCKED, got %s", locked.Status)
	}
	if locked.ReconciledAt == nil {
		t.Errorf("expected non-nil ReconciledAt")
	}

	// 4. Verify that running auto-match on locked session returns error
	_, err = svc.RunAutoMatch(ctx, leID, recon.ID, nil)
	if err != domain.ErrReconciliationLocked {
		t.Fatalf("expected ErrReconciliationLocked, got %v", err)
	}
}
