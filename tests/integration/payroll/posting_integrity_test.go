package payroll_test

import (
	"erp/internal/payroll/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestPayrollApprovalAndAtomicPosting(t *testing.T) {
	mod, _, db, entity, _, maker, ctx := setupPayrollTest(t)
	defer db.Close()
	svc := mod.PayrollService
	checker := payrollChecker(t, db, ctx)
	now := time.Now()
	period, err := svc.CreatePeriod(ctx, entity, now.Year(), int(now.Month()), "Posting test", "اختبار الترحيل", now.AddDate(0, 0, -1), now, now)
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.CreateRun(ctx, entity, period.ID, uuid.NewString(), now, now, &maker)
	if err != nil {
		t.Fatal(err)
	}
	run, err = svc.CalculateRun(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ApproveRun(ctx, run.ID, maker); err == nil {
		t.Fatal("maker approved own payroll")
	}
	if _, err := svc.PostRun(ctx, run.ID, &checker); err == nil {
		t.Fatal("unapproved payroll posted")
	}
	if _, err := db.Pool.Exec(ctx, `INSERT INTO workflow.approval_limits(legal_entity_id,user_id,resource_type,maximum_base_amount,updated_by) VALUES($1,$2,'PAYROLL_RUN',$3,$4)`, entity, checker, run.TotalGross.Sub(decimal.RequireFromString("0.000001")), maker); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ApproveRun(ctx, run.ID, checker); err == nil {
		t.Fatal("payroll exceeded approver authority")
	}
	if _, err := db.Pool.Exec(ctx, `DELETE FROM workflow.approval_limits WHERE legal_entity_id=$1 AND user_id=$2`, entity, checker); err != nil {
		t.Fatal(err)
	}
	run, err = svc.ApproveRun(ctx, run.ID, checker)
	if err != nil {
		t.Fatal(err)
	}
	if run.ApprovedBy == nil || *run.ApprovedBy != checker || run.ApprovedAt == nil {
		t.Fatal("approval actor not persisted")
	}
	if _, err := svc.CalculateRun(ctx, run.ID); err == nil {
		t.Fatal("approved calculations changed")
	}
	if _, err := svc.PostRun(ctx, run.ID, &maker); err == nil {
		t.Fatal("maker posted own payroll")
	}
	var fiscal uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id=$1 AND $2 BETWEEN start_date AND end_date", entity, now).Scan(&fiscal); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", fiscal); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.PostRun(ctx, run.ID, &checker); err == nil {
		t.Fatal("closed period accepted")
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1", fiscal); err != nil {
		t.Fatal(err)
	}
	current, err := svc.GetRun(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if current.Status != domain.RunStatusApproved || current.JournalEntryID != nil {
		t.Fatal("failed posting mutated payroll")
	}
	var slips int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM payroll.payslips p JOIN payroll.run_employees e ON e.id=p.run_employee_id WHERE e.run_id=$1", run.ID).Scan(&slips); err != nil || slips != 0 {
		t.Fatalf("failed posting left payslips: %d %v", slips, err)
	}
	posted, err := svc.PostRun(ctx, run.ID, &checker)
	if err != nil {
		t.Fatal(err)
	}
	if posted.JournalEntryID == nil {
		t.Fatal("payroll journal missing")
	}
	var debit, credit decimal.Decimal
	var actor, entryPeriod uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT total_debit,total_credit,posted_by,fiscal_period_id FROM accounting.journal_entries WHERE id=$1", posted.JournalEntryID).Scan(&debit, &credit, &actor, &entryPeriod); err != nil {
		t.Fatal(err)
	}
	if !debit.Equal(credit) || !debit.Equal(run.TotalGross) || actor != checker || entryPeriod != fiscal {
		t.Fatal("incorrect payroll journal")
	}
	var expectedGosi, postedGosi, postedVAT decimal.Decimal
	if err := db.Pool.QueryRow(ctx, `SELECT COALESCE(sum(gosi_employee_deduction),0) FROM payroll.run_employees WHERE run_id=$1`, run.ID).Scan(&expectedGosi); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, `SELECT COALESCE(sum(l.base_credit) FILTER(WHERE a.code='215000'),0),COALESCE(sum(l.base_credit) FILTER(WHERE a.code='213000'),0) FROM accounting.journal_lines l JOIN accounting.accounts a ON a.id=l.account_id WHERE l.journal_entry_id=$1`, posted.JournalEntryID).Scan(&postedGosi, &postedVAT); err != nil {
		t.Fatal(err)
	}
	if !expectedGosi.Equal(postedGosi) || !postedVAT.IsZero() {
		t.Fatalf("employee insurance posted to wrong liability: expected=%s insurance=%s VAT=%s", expectedGosi, postedGosi, postedVAT)
	}
	var audits int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit.events WHERE resource_id=$1 AND action='accounting.journal.post'", posted.JournalEntryID.String()).Scan(&audits); err != nil || audits != 1 {
		t.Fatalf("payroll audit missing: %d %v", audits, err)
	}
}
