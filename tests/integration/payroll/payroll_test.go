package payroll_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/hr"
	hrdomain "erp/internal/hr/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/payroll"
	"erp/internal/payroll/domain"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupPayrollTest(t *testing.T) (*payroll.Module, *hr.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	accountingMod := accounting.NewModule(cfg, db)
	hrMod := hr.NewModule(cfg, db)
	payrollMod := payroll.NewModule(cfg, db)

	// 1. Seed Organization & Accounting
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}
	if err := accountingMod.SeedDefaultCOA(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed payroll accounting defaults: %v", err)
	}

	// 2. Admin User
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	} else {
		createdAdmin, _ := iamMod.AuthService.SeedAdmin(ctx, "admin", "admin@erp.sa", "System Administrator", "Admin@ERP123!")
		if createdAdmin != nil {
			adminID = createdAdmin.ID
		}
	}

	// 3. Onboard Employee
	hireDate := time.Now().AddDate(-3, 0, 0)
	emp, err := hrMod.HRService.OnboardEmployee(ctx, &hrdomain.Employee{
		LegalEntityID:          le.ID,
		EmployeeNumber:         "EMP-PAY-" + uuid.New().String()[:6],
		FirstNameEn:            "Fahad",
		LastNameEn:             "Al-Harbi",
		FirstNameAr:            "فهد",
		LastNameAr:             "الحربي",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               hireDate,
		EmploymentStatus:       hrdomain.EmploymentStatusActive,
	}, &hrdomain.EmployeeIdentifier{
		IdentifierType:  hrdomain.IdentifierNationalID,
		IdentifierValue: "1098765432",
		IsPrimary:       true,
		IsVerified:      true,
	}, nil, &hrdomain.EmployeeBankAccount{
		BankNameEn:    "National Commercial Bank",
		BankNameAr:    "البنك الأهلي السعودي",
		IBAN:          "SA4410000001234567890123",
		SwiftBIC:      "NCBK",
		IsPrimary:     true,
		AccountStatus: "ACTIVE",
	})
	if err != nil {
		t.Fatalf("failed to onboard employee for payroll: %v", err)
	}

	// 4. Create Contract with Salary Decomposition
	reason := "INITIAL_HIRE"
	_, _, err = hrMod.HRService.CreateContract(ctx, &hrdomain.Contract{
		LegalEntityID:       le.ID,
		EmployeeID:          emp.ID,
		ContractNumber:      "QIWA-" + uuid.New().String()[:8],
		ContractType:        hrdomain.ContractUnspecifiedTerm,
		StartDate:           hireDate,
		ProbationDays:       90,
		NoticePeriodDays:    60,
		WorkLocation:        "Riyadh HQ",
		WorkingHoursPerWeek: decimal.NewFromInt(40),
		Status:              hrdomain.ContractActive,
	}, &hrdomain.ContractVersion{
		EffectiveFrom:      hireDate,
		BasicSalary:        decimal.NewFromInt(12000),
		HousingAllowance:   decimal.NewFromInt(3000),
		TransportAllowance: decimal.NewFromInt(1000),
		OtherAllowances:    decimal.NewFromInt(500),
		ChangeReason:       &reason,
	})
	if err != nil {
		t.Fatalf("failed to create contract for employee: %v", err)
	}

	return payrollMod, hrMod, db, le.ID, emp.ID, adminID, ctx
}

func TestPayroll_DeterministicCalculation(t *testing.T) {
	// Proves invariant: Same inputs ALWAYS produce 100% identical outputs
	input := domain.EmployeePayrollInput{
		EmployeeID:         uuid.New(),
		IsSaudi:            true,
		BasicSalary:        decimal.NewFromInt(10000),
		HousingAllowance:   decimal.NewFromInt(2500),
		TransportAllowance: decimal.NewFromInt(1000),
		OtherAllowances:    decimal.NewFromInt(500),
		OvertimeHours:      decimal.NewFromFloat(10.0),
		OvertimeMultiplier: decimal.NewFromFloat(1.50),
		LoanDeduction:      decimal.NewFromInt(500),
		UnpaidDays:         2,
	}

	res1, err1 := domain.CalculateDeterministicPayroll(input)
	if err1 != nil {
		t.Fatalf("run 1 failed: %v", err1)
	}

	res2, err2 := domain.CalculateDeterministicPayroll(input)
	if err2 != nil {
		t.Fatalf("run 2 failed: %v", err2)
	}

	if !res1.GrossSalary.Equal(res2.GrossSalary) {
		t.Errorf("expected gross %s, got %s", res1.GrossSalary, res2.GrossSalary)
	}
	if !res1.NetSalary.Equal(res2.NetSalary) {
		t.Errorf("expected net %s, got %s", res1.NetSalary, res2.NetSalary)
	}
	if !res1.TotalDeductions.Equal(res2.TotalDeductions) {
		t.Errorf("expected total deductions %s, got %s", res1.TotalDeductions, res2.TotalDeductions)
	}
	if len(res1.Lines) != len(res2.Lines) {
		t.Errorf("expected %d calculation lines, got %d", len(res1.Lines), len(res2.Lines))
	}
}

func TestPayroll_StateTransitionAndPosting(t *testing.T) {
	payMod, _, db, leID, empID, adminID, ctx := setupPayrollTest(t)
	svc := payMod.PayrollService

	// 1. Create Period
	now := time.Now()
	period, err := svc.CreatePeriod(ctx, leID, 2027, 1, "Period 2027-01", "فترة يناير 2027", now.AddDate(0, 0, -15), now.AddDate(0, 0, 15), now.AddDate(0, 0, 10))
	if err != nil {
		t.Fatalf("failed to create period: %v", err)
	}

	// 2. Create Payroll Run (DRAFT)
	runNum := "PR-" + uuid.New().String()[:8]
	run, err := svc.CreateRun(ctx, leID, period.ID, runNum, now, now, &adminID)
	if err != nil {
		t.Fatalf("failed to create payroll run: %v", err)
	}
	if run.Status != domain.RunStatusDraft {
		t.Fatalf("expected status DRAFT, got %s", run.Status)
	}

	// 3. Calculate Payroll Run (CALCULATED)
	calculated, err := svc.CalculateRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to calculate payroll: %v", err)
	}
	if calculated.Status != domain.RunStatusCalculated {
		t.Fatalf("expected status CALCULATED, got %s", calculated.Status)
	}
	if calculated.EmployeeCount < 1 {
		t.Fatalf("expected at least 1 employee, got %d", calculated.EmployeeCount)
	}
	if !calculated.TotalGross.GreaterThan(decimal.Zero) {
		t.Fatalf("expected positive total gross, got %s", calculated.TotalGross)
	}
	if !calculated.TotalNet.GreaterThan(decimal.Zero) {
		t.Fatalf("expected positive total net, got %s", calculated.TotalNet)
	}

	// 4. Approve Payroll Run (APPROVED)
	approved, err := svc.ApproveRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to approve run: %v", err)
	}
	if approved.Status != domain.RunStatusApproved {
		t.Fatalf("expected status APPROVED, got %s", approved.Status)
	}

	// 5. Post Payroll Run to General Ledger (POSTED)
	posted, err := svc.PostRun(ctx, run.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to post run: %v", err)
	}
	if posted.Status != domain.RunStatusPosted {
		t.Fatalf("expected status POSTED, got %s", posted.Status)
	}
	if posted.JournalEntryID == nil {
		t.Fatalf("expected linked journal entry id, got nil")
	}

	// 6. Verify Balanced General Ledger Journal Entry
	var totalDebit, totalCredit decimal.Decimal
	var debitStr, creditStr string
	err = db.Pool.QueryRow(ctx, `
		SELECT SUM(transaction_debit)::text, SUM(transaction_credit)::text
		FROM accounting.journal_lines
		WHERE journal_entry_id = $1
	`, *posted.JournalEntryID).Scan(&debitStr, &creditStr)
	if err != nil {
		t.Fatalf("failed to query journal lines: %v", err)
	}
	totalDebit, _ = decimal.NewFromString(debitStr)
	totalCredit, _ = decimal.NewFromString(creditStr)

	if !totalDebit.Equal(totalCredit) {
		t.Errorf("journal entry is unbalanced! Debit: %s, Credit: %s", totalDebit, totalCredit)
	}

	// 7. Verify Payslip Generation
	slips, err := svc.GetPayslipsByPeriod(ctx, period.ID)
	if err != nil {
		t.Fatalf("failed to get payslips: %v", err)
	}
	if len(slips) == 0 {
		t.Fatalf("expected at least 1 payslip, got 0")
	}

	foundEmp := false
	expectedEmpGross := decimal.NewFromInt(16500)
	for _, s := range slips {
		if s.EmployeeID == empID {
			foundEmp = true
			if !s.GrossSalary.Equal(expectedEmpGross) {
				t.Errorf("payslip gross mismatch: expected %s, got %s", expectedEmpGross, s.GrossSalary)
			}
		}
	}
	if !foundEmp {
		t.Fatalf("expected payslip for employee %s", empID)
	}
}

func TestPayroll_PostedImmutabilityAndDuplicatePostPrevention(t *testing.T) {
	payMod, _, _, leID, _, adminID, ctx := setupPayrollTest(t)
	svc := payMod.PayrollService

	now := time.Now()
	period, err := svc.CreatePeriod(ctx, leID, 2027, 2, "Period 2027-02", "فترة فبراير 2027", now.AddDate(0, 0, -10), now.AddDate(0, 0, 10), now.AddDate(0, 0, 5))
	if err != nil {
		t.Fatalf("failed to create period: %v", err)
	}

	runNum := "PR-IMM-" + uuid.New().String()[:8]
	run, _ := svc.CreateRun(ctx, leID, period.ID, runNum, now, now, &adminID)
	_, _ = svc.CalculateRun(ctx, run.ID)
	_, _ = svc.ApproveRun(ctx, run.ID)
	posted, err := svc.PostRun(ctx, run.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to post run: %v", err)
	}

	// Invariant 1: Attempting to recalculate a POSTED run must fail
	_, err = svc.CalculateRun(ctx, posted.ID)
	if err != domain.ErrPayrollAlreadyPosted {
		t.Errorf("expected ErrPayrollAlreadyPosted, got %v", err)
	}

	// Invariant 2: Attempting to post a POSTED run again must fail (Duplicate post prevention)
	_, err = svc.PostRun(ctx, posted.ID, &adminID)
	if err != domain.ErrDuplicatePayrollPost {
		t.Errorf("expected ErrDuplicatePayrollPost, got %v", err)
	}
}

func TestPayroll_WPS_SIF_Generation(t *testing.T) {
	payMod, _, _, leID, _, adminID, ctx := setupPayrollTest(t)
	svc := payMod.PayrollService

	now := time.Now()
	period, err := svc.CreatePeriod(ctx, leID, 2027, 3, "Period 2027-03", "فترة مارس 2027", now.AddDate(0, 0, -5), now.AddDate(0, 0, 20), now.AddDate(0, 0, 15))
	if err != nil {
		t.Fatalf("failed to create period: %v", err)
	}
	run, _ := svc.CreateRun(ctx, leID, period.ID, "PR-WPS-"+uuid.New().String()[:6], now, now, &adminID)
	calcRun, err := svc.CalculateRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("failed to calculate run for WPS: %v", err)
	}
	if calcRun.EmployeeCount == 0 {
		t.Fatalf("expected employees in run for WPS")
	}

	content, fileName, err := svc.GenerateWps(ctx, run.ID, "7009876543", "NCB")
	if err != nil {
		t.Fatalf("failed to generate WPS file: %v", err)
	}

	if !strings.HasPrefix(content, "SCR,NCB,7009876543") {
		t.Errorf("expected SCR header starting with SCR,NCB,7009876543, got: %s", content)
	}
	if !strings.Contains(content, "EDR,") {
		t.Errorf("expected EDR detail record, got: %s", content)
	}
	if !strings.HasPrefix(fileName, "SIF_7009876543_") {
		t.Errorf("expected filename prefix SIF_7009876543_, got %s", fileName)
	}
}

func TestPayroll_EndOfServiceCalculation(t *testing.T) {
	grossSalary := decimal.NewFromInt(12000)

	// Case 1: Resignation after 3.5 years (< 5 years => 1/3 factor)
	hireDate1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	termDate1 := time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC) // 3.5 years

	res1, err := domain.CalculateEndOfServiceAward(hireDate1, termDate1, grossSalary, domain.TerminationResignation)
	if err != nil {
		t.Fatalf("failed EOS calc 1: %v", err)
	}

	// Base award = 3.5 years * (0.5 * 12000) = 21,000 SAR
	// Resignation 2-5 years => 1/3 factor => 7,000 SAR
	expectedAward := decimal.NewFromInt(21000)
	if !res1.BaseAwardAmount.Equal(expectedAward) {
		t.Errorf("expected base award %s, got %s", expectedAward, res1.BaseAwardAmount)
	}

	expectedFinal := decimal.NewFromInt(7000)
	diff := res1.FinalSettlementAmount.Sub(expectedFinal).Abs()
	if diff.GreaterThan(decimal.NewFromInt(2)) {
		t.Errorf("expected final settlement ~%s, got %s", expectedFinal, res1.FinalSettlementAmount)
	}

	// Case 2: Employer termination after 7 years => 100% full award
	hireDate2 := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	termDate2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // 7.0 years

	res2, err := domain.CalculateEndOfServiceAward(hireDate2, termDate2, grossSalary, domain.TerminationByEmployer)
	if err != nil {
		t.Fatalf("failed EOS calc 2: %v", err)
	}

	// First 5 yrs: 5 * 0.5 * 12,000 = 30,000 SAR
	// Next 2 yrs: 2 * 1.0 * 12,000 = 24,000 SAR
	// Total Base = 54,000 SAR. Employer termination => 100% => 54,000 SAR
	expectedAward2 := decimal.NewFromInt(54000)
	if !res2.BaseAwardAmount.Equal(expectedAward2) {
		t.Errorf("expected base award %s, got %s", expectedAward2, res2.BaseAwardAmount)
	}
	if !res2.FinalSettlementAmount.Equal(expectedAward2) {
		t.Errorf("expected final settlement %s, got %s", expectedAward2, res2.FinalSettlementAmount)
	}
}

func TestPayroll_SeedDefaultData(t *testing.T) {
	payMod, _, _, leID, _, _, ctx := setupPayrollTest(t)

	err := payMod.SeedDefaultPayrollData(ctx, leID)
	if err != nil {
		t.Fatalf("failed to seed payroll data: %v", err)
	}

	comps, err := payMod.PayrollService.ListComponents(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list components: %v", err)
	}
	if len(comps) == 0 {
		t.Fatalf("expected seeded components, got 0")
	}

	structs, err := payMod.PayrollService.ListStructures(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list structures: %v", err)
	}
	if len(structs) == 0 {
		t.Fatalf("expected seeded salary structure, got 0")
	}
}
