package treasury_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	iamdomain "erp/internal/iam/domain"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/treasury/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestTreasuryHTTP_ActorApprovalAndLedger(t *testing.T) {
	mod, _, db, entity, maker, ctx := setupTreasuryTest(t)
	checker := treasuryChecker(t, db)
	svc := mod.TreasuryService
	cashGL, expenseGL := treasuryGL(t, db, entity, "111001"), treasuryGL(t, db, entity, "523000")
	account, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID: entity, AccountCode: "HTTP-" + uuid.NewString(), NameEn: "HTTP box", NameAr: "صندوق",
		AccountType: domain.CashAccountPettyCash, CurrencyCode: "SAR", GLAccountID: cashGL,
		CurrentBalance: decimal.NewFromInt(100),
	})
	if err != nil {
		t.Fatal(err)
	}
	router := mod.Handler.Routes(nil)
	call := func(path string, actor uuid.UUID, payload any) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		if actor != uuid.Nil {
			req = req.WithContext(iamhttp.ContextWithUser(ctx, &iamdomain.User{ID: actor}))
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	// Client-supplied creator/approver/journal fields cannot bypass maker-checker.
	forgedJournal := uuid.New()
	payload := domain.Payment{
		LegalEntityID: entity, CashAccountID: account.ID, PaymentDate: time.Now(), PaymentType: domain.PaymentMiscellaneous,
		PaidTo: "Payee", Amount: decimal.NewFromInt(30), CurrencyCode: "SAR", GLOffsetAccountID: expenseGL,
		CreatedByUserID: &checker, ApprovedByUserID: &checker, JournalEntryID: &forgedJournal,
	}
	if rec := call("/payments", uuid.Nil, payload); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated create: %d %s", rec.Code, rec.Body.String())
	}
	rec := call("/payments", maker, payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var payment domain.Payment
	if err := json.Unmarshal(rec.Body.Bytes(), &payment); err != nil {
		t.Fatal(err)
	}
	if payment.CreatedByUserID == nil || *payment.CreatedByUserID != maker || payment.ApprovedByUserID != nil || payment.JournalEntryID != nil {
		t.Fatalf("untrusted metadata accepted: %+v", payment)
	}
	path := fmt.Sprintf("/payments/%s/", payment.ID)
	query := "?legal_entity_id=" + entity.String()
	if rec := call(path+"pay"+query, checker, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("draft payment posted: %d", rec.Code)
	}
	if rec := call(path+"approve"+query, maker, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("maker approved own payment: %d", rec.Code)
	}
	if _, err := db.Pool.Exec(ctx, `INSERT INTO workflow.approval_limits(legal_entity_id,user_id,resource_type,maximum_base_amount,updated_by) VALUES($1,$2,'TREASURY_PAYMENT',29.999999,$3)`, entity, checker, maker); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(ctx, `DELETE FROM workflow.approval_limits WHERE legal_entity_id=$1 AND user_id=$2`, entity, checker)
	})
	if rec := call(path+"approve"+query, checker, nil); rec.Code != http.StatusBadRequest {
		t.Fatal("payment exceeded approver's configured limit")
	}
	if _, err := db.Pool.Exec(ctx, `UPDATE workflow.approval_limits SET maximum_base_amount=30 WHERE legal_entity_id=$1 AND user_id=$2`, entity, checker); err != nil {
		t.Fatal(err)
	}
	if rec := call(path+"approve"+query, checker, nil); rec.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call(path+"pay"+query, maker, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("maker posted own payment: %d", rec.Code)
	}
	if rec := call(path+"pay"+query, checker, nil); rec.Code != http.StatusOK {
		t.Fatalf("pay: %d %s", rec.Code, rec.Body.String())
	}
	if rec := call(path+"pay"+query, checker, nil); rec.Code != http.StatusOK {
		t.Fatalf("retry: %d %s", rec.Code, rec.Body.String())
	}
	paid, err := svc.GetPayment(ctx, entity, payment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if paid.JournalEntryID == nil {
		t.Fatal("paid payment has no journal")
	}
	var journalActor uuid.UUID
	var debits, credits decimal.Decimal
	if err := db.Pool.QueryRow(ctx, `SELECT j.posted_by, SUM(l.base_debit), SUM(l.base_credit) FROM accounting.journal_entries j JOIN accounting.journal_lines l ON l.journal_entry_id=j.id WHERE j.id=$1 GROUP BY j.posted_by`, *paid.JournalEntryID).Scan(&journalActor, &debits, &credits); err != nil {
		t.Fatal(err)
	}
	if journalActor != checker || !debits.Equal(decimal.NewFromInt(30)) || !credits.Equal(debits) {
		t.Fatalf("incorrect ledger actor/totals: %s %s %s", journalActor, debits, credits)
	}
	var count int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE resource_id=$1 AND action IN ('treasury.payment.approve','treasury.payment.pay') AND user_id=$2`, payment.ID.String(), checker).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected exactly two treasury audit transitions, got %d", count)
	}
	updated, err := svc.GetCashAccount(ctx, entity, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.CurrentBalance.Equal(decimal.NewFromInt(70)) {
		t.Fatalf("unexpected cash balance after retry: %s", updated.CurrentBalance)
	}

	// The receipt and transfer routes also derive both maker and posting actors
	// from authentication, even when the payload supplies a different user.
	receiptPayload := domain.Receipt{LegalEntityID: entity, CashAccountID: account.ID,
		ReceiptType: domain.ReceiptMiscellaneous, ReceivedFrom: "Payer", Amount: decimal.NewFromInt(10),
		CreatedByUserID: &checker, GLOffsetAccountID: treasuryGL(t, db, entity, "410000")}
	rec = call("/receipts", maker, receiptPayload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("receipt create: %d %s", rec.Code, rec.Body.String())
	}
	var receipt domain.Receipt
	if err := json.Unmarshal(rec.Body.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.CreatedByUserID == nil || *receipt.CreatedByUserID != maker {
		t.Fatal("receipt accepted forged creator")
	}
	rec = call(fmt.Sprintf("/receipts/%s/post%s", receipt.ID, query), checker, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("receipt post: %d %s", rec.Code, rec.Body.String())
	}
	destination, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID: entity, AccountCode: "DEST-" + uuid.NewString(), NameEn: "Destination", NameAr: "صندوق",
		AccountType: domain.CashAccountPettyCash, GLAccountID: cashGL,
	})
	if err != nil {
		t.Fatal(err)
	}
	transferPayload := domain.Transfer{LegalEntityID: entity, SourceType: domain.TransferSourceCash,
		SourceCashAccountID: &account.ID, DestinationType: domain.TransferDestinationCash,
		DestinationCashAccountID: &destination.ID, Amount: decimal.NewFromInt(10), InitiatedByUserID: &checker}
	rec = call("/transfers", maker, transferPayload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("transfer create: %d %s", rec.Code, rec.Body.String())
	}
	var transfer domain.Transfer
	if err := json.Unmarshal(rec.Body.Bytes(), &transfer); err != nil {
		t.Fatal(err)
	}
	if transfer.InitiatedByUserID == nil || *transfer.InitiatedByUserID != maker {
		t.Fatal("transfer accepted forged initiator")
	}
	completePath := fmt.Sprintf("/transfers/%s/complete%s", transfer.ID, query)
	if rec = call(completePath, maker, nil); rec.Code != http.StatusBadRequest {
		t.Fatal("initiator completed own transfer")
	}
	if rec = call(completePath, checker, nil); rec.Code != http.StatusOK {
		t.Fatalf("transfer complete: %d %s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &transfer); err != nil {
		t.Fatal(err)
	}
	if transfer.CompletedByUserID == nil || *transfer.CompletedByUserID != checker || transfer.JournalEntryID == nil {
		t.Fatal("transfer missing authenticated completer or journal")
	}
}

func TestTreasuryReceipt_AuditFailureRollsBackLedgerAndCash(t *testing.T) {
	mod, _, db, entity, maker, ctx := setupTreasuryTest(t)
	checker := treasuryChecker(t, db)
	svc := mod.TreasuryService
	account, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID: entity, AccountCode: "AUDIT-" + uuid.NewString(), NameEn: "Audit box", NameAr: "صندوق",
		AccountType: domain.CashAccountPettyCash, GLAccountID: treasuryGL(t, db, entity, "111001"), CurrentBalance: decimal.NewFromInt(100),
	})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := svc.CreateReceipt(ctx, &domain.Receipt{
		LegalEntityID: entity, CashAccountID: account.ID, ReceiptType: domain.ReceiptMiscellaneous, ReceivedFrom: "Payer",
		Amount: decimal.NewFromInt(25), CreatedByUserID: &maker, GLOffsetAccountID: treasuryGL(t, db, entity, "410000"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PostReceipt(ctx, entity, receipt.ID, maker); !errors.Is(err, domain.ErrMakerCannotCheck) {
		t.Fatalf("expected maker rejection, got %v", err)
	}
	// Fail after the GL posting and source state change, at the source audit append.
	_, err = db.Pool.Exec(ctx, `CREATE FUNCTION treasury.reject_test_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='treasury.receipt.post' THEN RAISE EXCEPTION 'injected treasury audit failure'; END IF; RETURN NEW; END $$; CREATE TRIGGER reject_test_audit BEFORE INSERT ON audit.events FOR EACH ROW EXECUTE FUNCTION treasury.reject_test_audit()`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(ctx, `DROP TRIGGER IF EXISTS reject_test_audit ON audit.events; DROP FUNCTION IF EXISTS treasury.reject_test_audit()`)
	})
	if _, err := svc.PostReceipt(ctx, entity, receipt.ID, checker); err == nil {
		t.Fatal("expected audit failure")
	}
	updated, err := svc.GetCashAccount(ctx, entity, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := svc.GetReceipt(ctx, entity, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.CurrentBalance.Equal(decimal.NewFromInt(100)) || unchanged.Status != domain.ReceiptDraft || unchanged.JournalEntryID != nil {
		t.Fatal("audit failure did not roll back cash/document")
	}
	var count int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM accounting.journal_entries WHERE source_module='TREASURY' AND source_id=$1`, receipt.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("audit failure left %d journal entries", count)
	}
	if _, err := db.Pool.Exec(ctx, `DROP TRIGGER reject_test_audit ON audit.events; DROP FUNCTION treasury.reject_test_audit()`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PostReceipt(ctx, entity, receipt.ID, checker); err != nil {
		t.Fatalf("valid retry failed: %v", err)
	}
}

func TestTreasuryReceipt_ClosedPeriodRollsBackCash(t *testing.T) {
	mod, _, db, entity, maker, ctx := setupTreasuryTest(t)
	svc := mod.TreasuryService
	account, err := svc.CreateCashAccount(ctx, &domain.CashAccount{
		LegalEntityID: entity, AccountCode: "PERIOD-" + uuid.NewString(), NameEn: "Period box", NameAr: "صندوق",
		AccountType: domain.CashAccountPettyCash, GLAccountID: treasuryGL(t, db, entity, "111001"),
	})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := svc.CreateReceipt(ctx, &domain.Receipt{LegalEntityID: entity, CashAccountID: account.ID, ReceiptType: domain.ReceiptMiscellaneous, ReceivedFrom: "Payer", Amount: decimal.NewFromInt(20), CreatedByUserID: &maker, GLOffsetAccountID: treasuryGL(t, db, entity, "410000")})
	if err != nil {
		t.Fatal(err)
	}
	var period uuid.UUID
	if err := db.Pool.QueryRow(ctx, `UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE legal_entity_id=$1 AND $2::date BETWEEN start_date AND end_date RETURNING id`, entity, receipt.ReceiptDate).Scan(&period); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(ctx, `UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1`, period)
	})
	if _, err := svc.PostReceipt(ctx, entity, receipt.ID, treasuryChecker(t, db)); err == nil {
		t.Fatal("closed period accepted posting")
	}
	updated, err := svc.GetCashAccount(ctx, entity, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := svc.GetReceipt(ctx, entity, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.CurrentBalance.IsZero() || unchanged.Status != domain.ReceiptDraft || unchanged.JournalEntryID != nil {
		t.Fatal("closed period failure changed financial state")
	}
}
