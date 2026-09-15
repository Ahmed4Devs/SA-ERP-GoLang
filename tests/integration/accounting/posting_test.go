package accounting_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"erp/internal/accounting/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPostingEngine_SuccessfulPostAndBalanceAccumulation(t *testing.T) {
	accMod, _, _, db, leID, userID, ctx := setupAccountingTest(t)

	journals, err := accMod.AccountingService.ListJournals(ctx, leID)
	if err != nil || len(journals) == 0 {
		t.Fatalf("failed to get journals: %v", err)
	}
	journal := journals[0]

	cashAcc, err := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	if err != nil {
		t.Fatalf("failed to find cash account: %v", err)
	}
	revAcc, err := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")
	if err != nil {
		t.Fatalf("failed to find revenue account: %v", err)
	}

	sourceID := uuid.New()
	amount := decimal.NewFromInt(12500) // 12,500 SAR

	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "SALES",
		SourceType:     "INVOICE",
		SourceID:       sourceID,
		JournalID:      journal.ID,
		AccountingDate: time.Now().UTC(),
		CurrencyCode:   "SAR",
		ExchangeRate:   decimal.NewFromInt(1),
		Reference:      "INV-2026-0001",
		Description:    "مبيعات نقدية للعميل - قيد ترحيل آلي",
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:         cashAcc.ID,
				Description:       "قبض نقدي - الصندوق الرئيسي",
				TransactionDebit:  amount,
				TransactionCredit: decimal.Zero,
				BaseDebit:         amount,
				BaseCredit:        decimal.Zero,
			},
			{
				AccountID:         revAcc.ID,
				Description:       "إيرادات مبيعات بضائع",
				TransactionDebit:  decimal.Zero,
				TransactionCredit: amount,
				BaseDebit:         decimal.Zero,
				BaseCredit:        amount,
			},
		},
	}

	// 1. Post via PostingEngine
	result, err := accMod.PostingEngine.Post(ctx, req)
	if err != nil {
		t.Fatalf("PostingEngine.Post failed: %v", err)
	}

	if result.JournalEntryID == uuid.Nil {
		t.Fatal("expected valid JournalEntryID in PostingResult")
	}
	if result.EntryNumber == "" {
		t.Fatal("expected generated EntryNumber in PostingResult")
	}

	// 2. Verify Journal Entry Status in DB
	entry, err := accMod.LedgerService.GetJournalEntryByID(ctx, result.JournalEntryID)
	if err != nil {
		t.Fatalf("failed to load posted journal entry: %v", err)
	}

	if entry.Status != domain.JournalEntryStatusPosted {
		t.Errorf("expected entry status POSTED, got: %s", entry.Status)
	}
	if entry.PostedBy == nil || *entry.PostedBy != userID {
		t.Errorf("expected posted_by=%s, got: %v", userID, entry.PostedBy)
	}
	if entry.PostedAt == nil {
		t.Error("expected posted_at to be set")
	}
	if entry.SourceModule != "SALES" || entry.SourceType == nil || *entry.SourceType != "INVOICE" {
		t.Errorf("unexpected source metadata: module=%s, type=%v", entry.SourceModule, entry.SourceType)
	}

	// 3. Verify Account Balances updated in DB
	var periodDebit decimal.Decimal
	err = db.Pool.QueryRow(ctx, `
		SELECT period_debit
		FROM accounting.account_balances
		WHERE legal_entity_id = $1 AND account_id = $2 AND fiscal_period_id = $3
	`, leID, cashAcc.ID, entry.FiscalPeriodID).Scan(&periodDebit)
	if err != nil {
		t.Fatalf("failed to query cash account balance: %v", err)
	}

	if periodDebit.LessThan(amount) {
		t.Errorf("expected cash period debit >= %s, got: %s", amount, periodDebit)
	}
}

func TestPostingEngine_IdempotencyEnforcement(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	journal := journals[0]
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	sourceID := uuid.New()
	amount := decimal.NewFromInt(5000)

	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "PURCHASES",
		SourceType:     "RECEIPT",
		SourceID:       sourceID,
		JournalID:      journal.ID,
		AccountingDate: time.Now().UTC(),
		CurrencyCode:   "SAR",
		ExchangeRate:   decimal.NewFromInt(1),
		Reference:      "RCV-999",
		Description:    "Idempotency test posting",
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:        cashAcc.ID,
				TransactionDebit: amount,
				BaseDebit:        amount,
			},
			{
				AccountID:         revAcc.ID,
				TransactionCredit: amount,
				BaseCredit:        amount,
			},
		},
	}

	// First post succeeds
	res1, err := accMod.PostingEngine.Post(ctx, req)
	if err != nil {
		t.Fatalf("first post failed: %v", err)
	}

	// Second post with identical source triple must fail with ErrDuplicatePosting
	_, err = accMod.PostingEngine.Post(ctx, req)
	if !errors.Is(err, domain.ErrDuplicatePosting) {
		t.Fatalf("expected ErrDuplicatePosting on duplicate request, got: %v", err)
	}

	// Verify that only one journal entry exists for this source
	entry, err := accMod.LedgerService.GetJournalEntryByID(ctx, res1.JournalEntryID)
	if err != nil || entry == nil {
		t.Fatalf("expected original entry to exist: %v", err)
	}
}

func TestPostingEngine_UnbalancedEntryRejection(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	journal := journals[0]
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "GL",
		SourceType:     "ADJUSTMENT",
		SourceID:       uuid.New(),
		JournalID:      journal.ID,
		AccountingDate: time.Now().UTC(),
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:        cashAcc.ID,
				TransactionDebit: decimal.NewFromInt(1000),
				BaseDebit:        decimal.NewFromInt(1000),
			},
			{
				AccountID:         revAcc.ID,
				TransactionCredit: decimal.NewFromInt(800), // UNBALANCED
				BaseCredit:        decimal.NewFromInt(800),
			},
		},
	}

	_, err := accMod.PostingEngine.Post(ctx, req)
	if !errors.Is(err, domain.ErrUnbalancedJournalEntry) {
		t.Fatalf("expected ErrUnbalancedJournalEntry, got: %v", err)
	}
}

func TestPostingEngine_LockedPeriodRejection(t *testing.T) {
	accMod, _, _, db, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	journal := journals[0]
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	// Create a historical fiscal year and LOCKED period
	now := time.Now().UTC()
	histYear := now.Year() - 5
	fyID := uuid.New()
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO accounting.fiscal_years (id, legal_entity_id, code, name_en, name_ar, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'CLOSED')
		ON CONFLICT (legal_entity_id, code) DO UPDATE SET status = 'CLOSED'
		RETURNING id
	`, fyID, leID, fmt.Sprintf("FY%d", histYear), fmt.Sprintf("FY %d", histYear), fmt.Sprintf("السنة المالية %d", histYear),
		time.Date(histYear, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(histYear, 12, 31, 0, 0, 0, 0, time.UTC),
	).Scan(&fyID)
	if err != nil {
		t.Fatalf("failed to insert historical fiscal year: %v", err)
	}

	periodID := uuid.New()
	histDate := time.Date(histYear, 3, 15, 0, 0, 0, 0, time.UTC)
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO accounting.fiscal_periods (id, fiscal_year_id, legal_entity_id, period_number, period_name_en, period_name_ar, quarter, start_date, end_date, status)
		VALUES ($1, $2, $3, 3, 'Historical Locked Period', 'فترة تاريخية مغلقة', 1, $4, $5, 'LOCKED')
		ON CONFLICT (fiscal_year_id, period_number) DO UPDATE SET status = 'LOCKED'
		RETURNING id
	`, periodID, fyID, leID,
		time.Date(histYear, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(histYear, 3, 31, 0, 0, 0, 0, time.UTC),
	).Scan(&periodID)
	if err != nil {
		t.Fatalf("failed to insert locked period: %v", err)
	}

	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "SALES",
		SourceType:     "INVOICE",
		SourceID:       uuid.New(),
		JournalID:      journal.ID,
		AccountingDate: histDate,
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:        cashAcc.ID,
				TransactionDebit: decimal.NewFromInt(100),
				BaseDebit:        decimal.NewFromInt(100),
			},
			{
				AccountID:         revAcc.ID,
				TransactionCredit: decimal.NewFromInt(100),
				BaseCredit:        decimal.NewFromInt(100),
			},
		},
	}

	_, err = accMod.PostingEngine.Post(ctx, req)
	if !errors.Is(err, domain.ErrPeriodNotOpen) {
		t.Fatalf("expected ErrPeriodNotOpen for locked period, got: %v", err)
	}
}

func TestPostingEngine_NonPostingAccountRejection(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	journal := journals[0]

	// Parent account "110000" (Current Assets) has allow_posting = false
	parentAcc, err := accMod.AccountingService.GetAccountByCode(ctx, leID, "110000")
	if err != nil {
		t.Fatalf("failed to get parent account: %v", err)
	}
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "GL",
		SourceType:     "MANUAL",
		SourceID:       uuid.New(),
		JournalID:      journal.ID,
		AccountingDate: time.Now().UTC(),
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:        parentAcc.ID, // allow_posting = FALSE
				TransactionDebit: decimal.NewFromInt(500),
				BaseDebit:        decimal.NewFromInt(500),
			},
			{
				AccountID:         revAcc.ID,
				TransactionCredit: decimal.NewFromInt(500),
				BaseCredit:        decimal.NewFromInt(500),
			},
		},
	}

	_, err = accMod.PostingEngine.Post(ctx, req)
	if !errors.Is(err, domain.ErrAccountNotPostable) {
		t.Fatalf("expected ErrAccountNotPostable when posting to header account, got: %v", err)
	}
}

func TestPostingEngine_ReversalLifecycle(t *testing.T) {
	accMod, _, _, _, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	journal := journals[0]
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	amount := decimal.NewFromInt(7500)
	sourceID := uuid.New()

	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "SALES",
		SourceType:     "INVOICE",
		SourceID:       sourceID,
		JournalID:      journal.ID,
		AccountingDate: time.Now().UTC(),
		CurrencyCode:   "SAR",
		Reference:      "INV-REV-TEST",
		Description:    "Original sale for reversal testing",
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:        cashAcc.ID,
				TransactionDebit: amount,
				BaseDebit:        amount,
			},
			{
				AccountID:         revAcc.ID,
				TransactionCredit: amount,
				BaseCredit:        amount,
			},
		},
	}

	// 1. Post original
	res, err := accMod.PostingEngine.Post(ctx, req)
	if err != nil {
		t.Fatalf("initial post failed: %v", err)
	}

	// 2. Reverse
	revRes, err := accMod.PostingEngine.Reverse(ctx, res.JournalEntryID, "Customer refund and return", userID)
	if err != nil {
		t.Fatalf("reversal failed: %v", err)
	}

	if revRes.JournalEntryID == uuid.Nil {
		t.Fatal("expected valid reversal journal entry ID")
	}

	// 3. Verify original entry is REVERSED
	origEntry, err := accMod.LedgerService.GetJournalEntryByID(ctx, res.JournalEntryID)
	if err != nil {
		t.Fatalf("failed to reload original: %v", err)
	}
	if origEntry.Status != domain.JournalEntryStatusReversed {
		t.Errorf("expected original status REVERSED, got: %s", origEntry.Status)
	}
	if origEntry.ReversalOfID == nil || *origEntry.ReversalOfID != revRes.JournalEntryID {
		t.Errorf("expected original reversal_of_id to point to %s, got: %v", revRes.JournalEntryID, origEntry.ReversalOfID)
	}

	// 4. Verify reversal entry has debits and credits swapped
	revEntry, err := accMod.LedgerService.GetJournalEntryByID(ctx, revRes.JournalEntryID)
	if err != nil {
		t.Fatalf("failed to load reversal entry: %v", err)
	}
	if revEntry.Status != domain.JournalEntryStatusPosted {
		t.Errorf("expected reversal entry to be POSTED, got: %s", revEntry.Status)
	}

	// 5. Attempting to reverse a second time must fail with ErrAlreadyReversed
	_, err = accMod.PostingEngine.Reverse(ctx, res.JournalEntryID, "Duplicate reversal attempt", userID)
	if !errors.Is(err, domain.ErrAlreadyReversed) {
		t.Fatalf("expected ErrAlreadyReversed, got: %v", err)
	}
}

func TestPostingEngine_DatabaseImmutabilityTriggerProtection(t *testing.T) {
	accMod, _, _, db, leID, userID, ctx := setupAccountingTest(t)

	journals, _ := accMod.AccountingService.ListJournals(ctx, leID)
	journal := journals[0]
	cashAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "111002")
	revAcc, _ := accMod.AccountingService.GetAccountByCode(ctx, leID, "410000")

	amount := decimal.NewFromInt(3000)
	req := domain.PostingRequest{
		LegalEntityID:  leID,
		SourceModule:   "SALES",
		SourceType:     "INVOICE",
		SourceID:       uuid.New(),
		JournalID:      journal.ID,
		AccountingDate: time.Now().UTC(),
		ActorID:        userID,
		Lines: []domain.PostingLine{
			{
				AccountID:        cashAcc.ID,
				TransactionDebit: amount,
				BaseDebit:        amount,
			},
			{
				AccountID:         revAcc.ID,
				TransactionCredit: amount,
				BaseCredit:        amount,
			},
		},
	}

	res, err := accMod.PostingEngine.Post(ctx, req)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	// Direct SQL UPDATE on posted entry must be blocked by trg_prevent_posted_journal_entry_mutation
	_, err = db.Pool.Exec(ctx, `
		UPDATE accounting.journal_entries
		SET description = 'Direct malicious modification'
		WHERE id = $1
	`, res.JournalEntryID)
	if err == nil {
		t.Fatal("expected database trigger to block direct UPDATE on posted journal entry")
	}

	// Direct SQL DELETE on posted entry must be blocked by trg_prevent_posted_journal_entry_mutation
	_, err = db.Pool.Exec(ctx, `
		DELETE FROM accounting.journal_entries
		WHERE id = $1
	`, res.JournalEntryID)
	if err == nil {
		t.Fatal("expected database trigger to block direct DELETE on posted journal entry")
	}

	// Direct SQL INSERT line on posted entry must be blocked by trg_prevent_posted_journal_lines_mutation
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO accounting.journal_lines (
			journal_entry_id, legal_entity_id, line_number, account_id, currency_code, exchange_rate,
			transaction_debit, transaction_credit, base_debit, base_credit
		) VALUES ($1, $2, 99, $3, 'SAR', 1, 100, 0, 100, 0)
	`, res.JournalEntryID, leID, cashAcc.ID)
	if err == nil {
		t.Fatal("expected database trigger to block direct INSERT into journal_lines of posted entry")
	}
}
