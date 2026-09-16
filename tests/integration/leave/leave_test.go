package leave_test

import (
	"context"
	"testing"
	"time"

	hrdomain "erp/internal/hr/domain"
	"erp/internal/hr"
	"erp/internal/iam"
	"erp/internal/leave"
	"erp/internal/leave/domain"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupLeaveTest(t *testing.T) (*leave.Module, *hr.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	hrMod := hr.NewModule(cfg, db)
	leaveMod := leave.NewModule(cfg, db)

	// 1. Seed org
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Admin user
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

	// 3. Create active test employee
	emp, err := hrMod.HRService.OnboardEmployee(ctx, &hrdomain.Employee{
		LegalEntityID:          le.ID,
		EmployeeNumber:         "EMP-LV-" + uuid.New().String()[:6],
		FirstNameEn:            "Mona",
		LastNameEn:             "Al-Zahrani",
		FirstNameAr:            "منى",
		LastNameAr:             "الزهراني",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               time.Now(),
		EmploymentStatus:       hrdomain.EmploymentStatusActive,
	}, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to onboard employee for leave: %v", err)
	}

	return leaveMod, hrMod, db, le.ID, emp.ID, adminID, ctx
}

func TestLeave_StatutoryTypesAndPolicy(t *testing.T) {
	leaveMod, _, _, leID, _, _, ctx := setupLeaveTest(t)

	// 1. Seed statutory Saudi leave types
	err := leaveMod.SeedDefaultLeaveData(ctx, leID)
	if err != nil {
		t.Fatalf("failed to seed default leave data: %v", err)
	}

	// 2. List leave types
	types, err := leaveMod.LeaveService.ListLeaveTypes(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list leave types: %v", err)
	}
	if len(types) < 9 {
		t.Errorf("expected at least 9 statutory leave types, got %d", len(types))
	}

	// Verify Annual Leave
	annualFound := false
	for _, lt := range types {
		if lt.Code == "ANNUAL" {
			annualFound = true
			if !lt.IsPaid {
				t.Errorf("expected annual leave to be paid")
			}
		}
	}
	if !annualFound {
		t.Errorf("expected ANNUAL leave type to be present")
	}

	// 3. Verify Policies
	policies, err := leaveMod.LeaveService.ListPolicies(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list leave policies: %v", err)
	}
	if len(policies) == 0 {
		t.Errorf("expected at least 1 leave policy")
	}
}

func TestLeave_AccrualAndBalances(t *testing.T) {
	leaveMod, _, _, leID, empID, _, ctx := setupLeaveTest(t)
	svc := leaveMod.LeaveService

	// 1. Create specific leave type
	lt, err := svc.CreateLeaveType(ctx, &domain.LeaveType{
		LegalEntityID:  leID,
		Code:           "ANNUAL-TEST-" + uuid.New().String()[:6],
		NameEn:         "Annual Vacation",
		NameAr:         "الإجازة السنوية",
		IsPaid:         true,
		IsActive:       true,
	})
	if err != nil {
		t.Fatalf("failed to create leave type: %v", err)
	}

	// 2. Accrue 10.5 days
	year := int32(2026)
	accrualDays := decimal.NewFromFloat(10.5)
	bal, err := svc.AccrueLeave(ctx, empID, lt.ID, year, accrualDays, "Monthly standard accrual")
	if err != nil {
		t.Fatalf("failed to accrue leave: %v", err)
	}

	if !bal.ClosingBalance.Equal(accrualDays) {
		t.Errorf("expected closing balance %s, got %s", accrualDays.String(), bal.ClosingBalance.String())
	}
	if !bal.AccruedDays.Equal(accrualDays) {
		t.Errorf("expected accrued days %s, got %s", accrualDays.String(), bal.AccruedDays.String())
	}

	// 3. Verify Transactions
	txs, err := leaveMod.Repo.ListLeaveTransactions(ctx, empID, lt.ID, year)
	if err != nil {
		t.Fatalf("failed to list leave transactions: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}
	if txs[0].TransactionType != domain.TxAccrual {
		t.Errorf("expected ACCRUAL transaction, got %s", txs[0].TransactionType)
	}
}

func TestLeave_RequestWorkflow(t *testing.T) {
	leaveMod, _, _, leID, empID, adminID, ctx := setupLeaveTest(t)
	svc := leaveMod.LeaveService

	// 1. Create Leave Type & Accrue Balance
	lt, err := svc.CreateLeaveType(ctx, &domain.LeaveType{
		LegalEntityID: leID,
		Code:          "ANNUAL", // Code "ANNUAL" triggers balance checking
		NameEn:        "Annual Leave",
		NameAr:        "إجازة اعتيادية",
		IsPaid:        true,
		IsActive:      true,
	})
	if err != nil {
		// If code already exists from seed, query it
		lt, err = leaveMod.Repo.GetLeaveTypeByCode(ctx, leID, "ANNUAL")
		if err != nil {
			t.Fatalf("failed to get annual leave type: %v", err)
		}
	}

	year := int32(2026)
	_, err = svc.AccrueLeave(ctx, empID, lt.ID, year, decimal.NewFromInt(15), "Initial Balance")
	if err != nil {
		t.Fatalf("failed to accrue initial balance: %v", err)
	}

	// 2. Submit Leave Request for 3 working days (Sunday to Tuesday: 2026-05-10 to 2026-05-12)
	startDate := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
	req, err := svc.SubmitLeaveRequest(ctx, &domain.LeaveRequest{
		LegalEntityID: leID,
		EmployeeID:    empID,
		LeaveTypeID:   lt.ID,
		StartDate:     startDate,
		EndDate:       endDate,
		Reason:        "Family vacation",
	})
	if err != nil {
		t.Fatalf("failed to submit leave request: %v", err)
	}
	if req.Status != domain.RequestPending {
		t.Errorf("expected PENDING status, got %s", req.Status)
	}
	if !req.TotalDays.Equal(decimal.NewFromInt(3)) {
		t.Errorf("expected 3 working days, got %s", req.TotalDays.String())
	}

	// Verify balance reservation: 15 - 3 = 12 closing balance, 3 pending
	balAfterSubmit, err := leaveMod.Repo.GetEmployeeLeaveBalance(ctx, empID, lt.ID, year)
	if err != nil {
		t.Fatalf("failed to get balance: %v", err)
	}
	if !balAfterSubmit.ClosingBalance.Equal(decimal.NewFromInt(12)) {
		t.Errorf("expected 12 closing balance after submit, got %s", balAfterSubmit.ClosingBalance.String())
	}
	if !balAfterSubmit.PendingDays.Equal(decimal.NewFromInt(3)) {
		t.Errorf("expected 3 pending days, got %s", balAfterSubmit.PendingDays.String())
	}

	// 3. Approve Leave Request
	approvedReq, err := svc.ApproveLeaveRequest(ctx, leID, req.ID, adminID, "Approved by HR Director")
	if err != nil {
		t.Fatalf("failed to approve leave request: %v", err)
	}
	if approvedReq.Status != domain.RequestApproved {
		t.Errorf("expected APPROVED status, got %s", approvedReq.Status)
	}

	// Verify balance committed: pending = 0, taken = 3, closing = 12
	balAfterApprove, err := leaveMod.Repo.GetEmployeeLeaveBalance(ctx, empID, lt.ID, year)
	if err != nil {
		t.Fatalf("failed to get balance: %v", err)
	}
	if !balAfterApprove.PendingDays.Equal(decimal.Zero) {
		t.Errorf("expected 0 pending days after approve, got %s", balAfterApprove.PendingDays.String())
	}
	if !balAfterApprove.TakenDays.Equal(decimal.NewFromInt(3)) {
		t.Errorf("expected 3 taken days, got %s", balAfterApprove.TakenDays.String())
	}

	// 4. Verify Immutable Transaction Ledger (TxUsage)
	txs, err := leaveMod.Repo.ListLeaveTransactions(ctx, empID, lt.ID, year)
	if err != nil {
		t.Fatalf("failed to list transactions: %v", err)
	}
	usageFound := false
	for _, tx := range txs {
		if tx.TransactionType == domain.TxUsage && tx.Days.Equal(decimal.NewFromInt(3)) {
			usageFound = true
			if tx.ReferenceID == nil || *tx.ReferenceID != req.ID {
				t.Errorf("expected transaction reference to match request ID")
			}
		}
	}
	if !usageFound {
		t.Errorf("expected TxUsage transaction in ledger")
	}
}

func TestLeave_InsufficientBalance(t *testing.T) {
	leaveMod, _, _, leID, empID, _, ctx := setupLeaveTest(t)
	svc := leaveMod.LeaveService

	lt, err := leaveMod.Repo.GetLeaveTypeByCode(ctx, leID, "ANNUAL")
	if err != nil {
		lt, err = svc.CreateLeaveType(ctx, &domain.LeaveType{
			LegalEntityID: leID,
			Code:          "ANNUAL",
			NameEn:        "Annual Leave",
			NameAr:        "إجازة اعتيادية",
			IsPaid:        true,
			IsActive:      true,
		})
		if err != nil {
			t.Fatalf("failed to prepare annual leave type: %v", err)
		}
	}

	// Current balance is 0 or less than 30 days
	startDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) // ~40+ working days

	_, err = svc.SubmitLeaveRequest(ctx, &domain.LeaveRequest{
		LegalEntityID: leID,
		EmployeeID:    empID,
		LeaveTypeID:   lt.ID,
		StartDate:     startDate,
		EndDate:       endDate,
		Reason:        "Long sabbatical",
	})
	if err != domain.ErrInsufficientBalance {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestLeave_RejectionAndBalanceReversion(t *testing.T) {
	leaveMod, _, _, leID, empID, adminID, ctx := setupLeaveTest(t)
	svc := leaveMod.LeaveService

	lt, err := leaveMod.Repo.GetLeaveTypeByCode(ctx, leID, "ANNUAL")
	if err != nil {
		lt, err = svc.CreateLeaveType(ctx, &domain.LeaveType{
			LegalEntityID: leID,
			Code:          "ANNUAL",
			NameEn:        "Annual Leave",
			NameAr:        "إجازة اعتيادية",
			IsPaid:        true,
			IsActive:      true,
		})
		if err != nil {
			t.Fatalf("failed to prepare annual leave: %v", err)
		}
	}

	year := int32(2026)
	_, err = svc.AccrueLeave(ctx, empID, lt.ID, year, decimal.NewFromInt(10), "Initial Accrual")
	if err != nil {
		t.Fatalf("failed to accrue: %v", err)
	}

	// Submit 2 working days (Sunday to Monday)
	startDate := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	req, err := svc.SubmitLeaveRequest(ctx, &domain.LeaveRequest{
		LegalEntityID: leID,
		EmployeeID:    empID,
		LeaveTypeID:   lt.ID,
		StartDate:     startDate,
		EndDate:       endDate,
		Reason:        "Personal trip",
	})
	if err != nil {
		t.Fatalf("failed to submit request: %v", err)
	}

	// Reject request
	rejectedReq, err := svc.RejectLeaveRequest(ctx, leID, req.ID, adminID, "Peak business period; request postponed")
	if err != nil {
		t.Fatalf("failed to reject request: %v", err)
	}
	if rejectedReq.Status != domain.RequestRejected {
		t.Errorf("expected REJECTED status, got %s", rejectedReq.Status)
	}

	// Verify balance reverted: pending = 0, closing balance restored to 10
	bal, err := leaveMod.Repo.GetEmployeeLeaveBalance(ctx, empID, lt.ID, year)
	if err != nil {
		t.Fatalf("failed to get balance: %v", err)
	}
	if !bal.PendingDays.Equal(decimal.Zero) {
		t.Errorf("expected pending days reverted to 0, got %s", bal.PendingDays.String())
	}
	if !bal.ClosingBalance.Equal(decimal.NewFromInt(10)) {
		t.Errorf("expected closing balance restored to 10, got %s", bal.ClosingBalance.String())
	}
}
