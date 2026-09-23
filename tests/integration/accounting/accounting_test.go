package accounting_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/accounting/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupAccountingTest(t *testing.T) (*accounting.Module, *organization.Module, *iam.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
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
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)

	// Ensure legal entity exists
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization structure: %v", err)
	}

	// Seed default Chart of Accounts and Fiscal Calendar
	if err := accMod.SeedDefaultCOA(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed default COA: %v", err)
	}

	// Create test user
	adminUser, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var userID uuid.UUID
	if err != nil || adminUser == nil {
		email := uuid.NewString()[:8] + "@acc.test"
		user, err := iamMod.AuthService.SeedAdmin(ctx, "acc_user_"+uuid.NewString()[:6], email, "Test User", "Password@123")
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
		userID = user.ID
	} else {
		userID = adminUser.ID
	}

	// Ensure journals exist
	journals, _ := accMod.AccountingService.ListJournals(ctx, le.ID)
	if len(journals) == 0 {
		_, _ = accMod.AccountingService.CreateJournal(ctx, domain.Journal{
			LegalEntityID: le.ID,
			Code:          "GEN",
			NameEn:        "General Journal",
			NameAr:        "دفتر اليومية العامة",
			JournalType:   domain.JournalTypeGeneral,
			IsActive:      true,
		})
	}

	return accMod, orgMod, iamMod, db, le.ID, userID, ctx
}

func TestAccounting_ChartOfAccountsHierarchy(t *testing.T) {
	accMod, _, _, _, leID, _, ctx := setupAccountingTest(t)

	// 1. List Charts of Accounts
	coas, err := accMod.AccountingService.ListChartOfAccounts(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list chart of accounts: %v", err)
	}
	if len(coas) == 0 {
		t.Fatal("expected at least 1 chart of accounts seeded")
	}

	chartID := coas[0].ID

	// 2. List Accounts and verify 5 major categories exist
	accounts, err := accMod.AccountingService.ListAccounts(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list accounts: %v", err)
	}
	if len(accounts) == 0 {
		t.Fatal("expected seeded standard Saudi accounts")
	}

	categories := map[domain.AccountType]bool{}
	for _, acc := range accounts {
		categories[acc.AccountType] = true
	}

	for _, expected := range []domain.AccountType{
		domain.AccountTypeAsset,
		domain.AccountTypeLiability,
		domain.AccountTypeEquity,
		domain.AccountTypeRevenue,
		domain.AccountTypeExpense,
	} {
		if !categories[expected] {
			t.Errorf("expected account category %s to exist in COA", expected)
		}
	}

	// 3. Create a custom sub-account
	cashAccount, err := accMod.AccountingService.GetAccountByCode(ctx, leID, "111000")
	if err != nil {
		t.Fatalf("failed to get parent cash account: %v", err)
	}

	newAccCode := fmt.Sprintf("111%03d", (time.Now().UnixNano()/1000)%900+100)
	createdAcc, err := accMod.AccountingService.CreateAccount(ctx, domain.Account{
		LegalEntityID:    leID,
		ChartID:          chartID,
		ParentID:         &cashAccount.ID,
		Code:             newAccCode,
		NameEn:           "Riyad Bank Operational SAR",
		NameAr:           "بنك الرياض - الحساب التشغيلي",
		AccountType:      domain.AccountTypeAsset,
		NormalBalance:    domain.NormalBalanceDebit,
		AllowPosting:     true,
		IsControlAccount: false,
		IsActive:         true,
	})
	if err != nil {
		t.Fatalf("failed to create sub-account: %v", err)
	}

	if createdAcc.Code != newAccCode || *createdAcc.ParentID != cashAccount.ID {
		t.Fatalf("sub-account fields mismatched: %+v", createdAcc)
	}
}

func TestAccounting_JournalEntryBalanceInvariant(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, err := accMod.AccountingService.ListJournals(ctx, leID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("failed to list journals: %v", err)
	}
	genJournal := journals[0]

	// Find cash (asset) and revenue accounts
	cashAcc, err := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	if err != nil {
		t.Fatalf("failed to find cash account: %v", err)
	}
	revAcc, err := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")
	if err != nil {
		t.Fatalf("failed to find revenue account: %v", err)
	}

	// 1. Negative Test: Unbalanced Entry (Debit 1000 vs Credit 500)
	unbalancedLines := []domain.JournalLine{
		{
			LineNumber:        1,
			AccountID:         cashAcc.ID,
			TransactionDebit:  decimal.NewFromInt(1000),
			TransactionCredit: decimal.Zero,
			BaseDebit:         decimal.NewFromInt(1000),
			BaseCredit:        decimal.Zero,
		},
		{
			LineNumber:        2,
			AccountID:         revAcc.ID,
			TransactionDebit:  decimal.Zero,
			TransactionCredit: decimal.NewFromInt(500), // UNBALANCED
			BaseDebit:         decimal.Zero,
			BaseCredit:        decimal.NewFromInt(500),
		},
	}

	_, err = accMod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      genJournal.ID,
		AccountingDate: time.Now().UTC(),
		Description:    "Unbalanced test entry",
		CreatedBy:      userID,
	}, unbalancedLines)

	if !errors.Is(err, domain.ErrUnbalancedJournalEntry) {
		t.Fatalf("expected ErrUnbalancedJournalEntry, got: %v", err)
	}

	// 2. Positive Test: Balanced Entry (Debit 5,000 SAR, Credit 5,000 SAR)
	balancedLines := []domain.JournalLine{
		{
			LineNumber:        1,
			AccountID:         cashAcc.ID,
			TransactionDebit:  decimal.NewFromInt(5000),
			TransactionCredit: decimal.Zero,
			BaseDebit:         decimal.NewFromInt(5000),
			BaseCredit:        decimal.Zero,
		},
		{
			LineNumber:        2,
			AccountID:         revAcc.ID,
			TransactionDebit:  decimal.Zero,
			TransactionCredit: decimal.NewFromInt(5000),
			BaseDebit:         decimal.Zero,
			BaseCredit:        decimal.NewFromInt(5000),
		},
	}

	entry, err := accMod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      genJournal.ID,
		AccountingDate: time.Now().UTC(),
		Description:    "Cash sales collection",
		CreatedBy:      userID,
	}, balancedLines)
	if err != nil {
		t.Fatalf("failed to create balanced journal entry: %v", err)
	}

	if entry.Status != domain.JournalEntryStatusDraft {
		t.Errorf("expected initial status DRAFT, got: %s", entry.Status)
	}
	if !entry.TotalDebit.Equal(decimal.NewFromInt(5000)) || !entry.TotalCredit.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("unexpected entry totals: debit=%s, credit=%s", entry.TotalDebit, entry.TotalCredit)
	}

	// 3. Post the balanced entry
	posted, err := accMod.LedgerService.PostJournalEntry(ctx, entry.ID, userID)
	if err != nil {
		t.Fatalf("failed to post balanced journal entry: %v", err)
	}

	if posted.Status != domain.JournalEntryStatusPosted {
		t.Errorf("expected posted status POSTED, got: %s", posted.Status)
	}
	if posted.PostedAt == nil {
		t.Error("expected posted_at timestamp to be set")
	}
}

func TestAccounting_ImmutabilityTriggerEnforcement(t *testing.T) {
	accMod, _, _, db, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	// Create and post an entry
	entry, err := accMod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      journals[0].ID,
		AccountingDate: time.Now().UTC(),
		Description:    "Immutability test entry",
		CreatedBy:      userID,
	}, []domain.JournalLine{
		{
			LineNumber:        1,
			AccountID:         cashAcc.ID,
			TransactionDebit:  decimal.NewFromInt(2500),
			TransactionCredit: decimal.Zero,
		},
		{
			LineNumber:        2,
			AccountID:         revAcc.ID,
			TransactionDebit:  decimal.Zero,
			TransactionCredit: decimal.NewFromInt(2500),
		},
	})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	posted, err := accMod.LedgerService.PostJournalEntry(ctx, entry.ID, userID)
	if err != nil {
		t.Fatalf("failed to post entry: %v", err)
	}

	// 1. Direct raw SQL attempt to UPDATE posted entry description
	_, err = db.Pool.Exec(ctx, "UPDATE accounting.journal_entries SET description = 'Tampered' WHERE id = $1", posted.ID)
	if err == nil {
		t.Fatal("expected database trigger to block UPDATE of posted journal entry, but succeeded!")
	}

	// 2. Direct raw SQL attempt to DELETE a posted journal line
	_, err = db.Pool.Exec(ctx, "DELETE FROM accounting.journal_lines WHERE journal_entry_id = $1", posted.ID)
	if err == nil {
		t.Fatal("expected database trigger to block DELETE of posted journal lines, but succeeded!")
	}
}

func TestAccounting_ReversalEntry(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	entry, err := accMod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      journals[0].ID,
		AccountingDate: time.Now().UTC(),
		Description:    "Entry to be reversed",
		CreatedBy:      userID,
	}, []domain.JournalLine{
		{
			LineNumber:        1,
			AccountID:         cashAcc.ID,
			TransactionDebit:  decimal.NewFromInt(1200),
			TransactionCredit: decimal.Zero,
		},
		{
			LineNumber:        2,
			AccountID:         revAcc.ID,
			TransactionDebit:  decimal.Zero,
			TransactionCredit: decimal.NewFromInt(1200),
		},
	})
	if err != nil {
		t.Fatalf("failed to create entry: %v", err)
	}

	posted, err := accMod.LedgerService.PostJournalEntry(ctx, entry.ID, userID)
	if err != nil {
		t.Fatalf("failed to post entry: %v", err)
	}

	// Reverse the entry
	reversal, err := accMod.LedgerService.ReverseJournalEntry(ctx, posted.ID, "Incorrect invoice mapping", userID)
	if err != nil {
		t.Fatalf("failed to reverse journal entry: %v", err)
	}

	if reversal.Status != domain.JournalEntryStatusPosted {
		t.Errorf("expected reversal entry to be POSTED, got: %s", reversal.Status)
	}
	if len(reversal.Lines) != 2 {
		t.Fatalf("expected 2 lines in reversal entry, got %d", len(reversal.Lines))
	}

	// Verify original entry is now REVERSED
	origAfter, err := accMod.LedgerService.GetJournalEntryByID(ctx, posted.ID)
	if err != nil {
		t.Fatalf("failed to fetch original entry after reversal: %v", err)
	}
	if origAfter.Status != domain.JournalEntryStatusReversed {
		t.Errorf("expected original entry status to be REVERSED, got: %s", origAfter.Status)
	}
	if origAfter.ReversalOfID == nil || *origAfter.ReversalOfID != reversal.ID {
		t.Errorf("expected original entry reversal_of_id to point to reversal entry %v", reversal.ID)
	}
}

func TestAccounting_TrialBalanceIntegrity(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")
	opexAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "521000")

	// 1. Transaction 1: Revenue 8,000 SAR into Cash
	entry1, _ := accMod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      journals[0].ID,
		AccountingDate: time.Now().UTC(),
		Description:    "Revenue receipt",
		CreatedBy:      userID,
	}, []domain.JournalLine{
		{LineNumber: 1, AccountID: cashAcc.ID, TransactionDebit: decimal.NewFromInt(8000), TransactionCredit: decimal.Zero},
		{LineNumber: 2, AccountID: revAcc.ID, TransactionDebit: decimal.Zero, TransactionCredit: decimal.NewFromInt(8000)},
	})
	_, _ = accMod.LedgerService.PostJournalEntry(ctx, entry1.ID, userID)

	// 2. Transaction 2: Salary Payment 3,000 SAR from Cash
	entry2, _ := accMod.LedgerService.CreateJournalEntry(ctx, domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      journals[0].ID,
		AccountingDate: time.Now().UTC(),
		Description:    "Salary payout",
		CreatedBy:      userID,
	}, []domain.JournalLine{
		{LineNumber: 1, AccountID: opexAcc.ID, TransactionDebit: decimal.NewFromInt(3000), TransactionCredit: decimal.Zero},
		{LineNumber: 2, AccountID: cashAcc.ID, TransactionDebit: decimal.Zero, TransactionCredit: decimal.NewFromInt(3000)},
	})
	_, _ = accMod.LedgerService.PostJournalEntry(ctx, entry2.ID, userID)

	// 3. Extract Trial Balance
	report, err := accMod.LedgerService.GetTrialBalance(ctx, leID, nil)
	if err != nil {
		t.Fatalf("failed to compute trial balance: %v", err)
	}

	if !report.IsBalanced {
		t.Fatalf("expected Trial Balance to be balanced, got sum_debit=%s, sum_credit=%s", report.TotalDebitSum, report.TotalCreditSum)
	}

	if !report.TotalDebitSum.Equal(report.TotalCreditSum) {
		t.Fatalf("mathematical mismatch: debit sum %s != credit sum %s", report.TotalDebitSum, report.TotalCreditSum)
	}

	if len(report.Rows) == 0 {
		t.Fatal("expected trial balance rows to be populated")
	}
}
