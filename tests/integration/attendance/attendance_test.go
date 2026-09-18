package attendance_test

import (
	"context"
	"testing"
	"time"

	"erp/internal/attendance"
	"erp/internal/attendance/domain"
	"erp/internal/hr"
	hrdomain "erp/internal/hr/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupAttendanceTest(t *testing.T) (*attendance.Module, *hr.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	attendanceMod := attendance.NewModule(cfg, db)

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
		EmployeeNumber:         "EMP-ATT-" + uuid.New().String()[:6],
		FirstNameEn:            "Rami",
		LastNameEn:             "Al-Saleh",
		FirstNameAr:            "رامي",
		LastNameAr:             "الصالح",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               time.Now(),
		EmploymentStatus:       hrdomain.EmploymentStatusActive,
	}, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to onboard employee for attendance: %v", err)
	}

	return attendanceMod, hrMod, db, le.ID, emp.ID, adminID, ctx
}

func TestShift_CreationAndAssignment(t *testing.T) {
	attMod, _, _, leID, empID, _, ctx := setupAttendanceTest(t)
	svc := attMod.AttendanceService

	// 1. Create Shift
	shiftCode := "SHIFT-TEST-" + uuid.New().String()[:6]
	shift, err := svc.CreateShift(ctx, &domain.Shift{
		LegalEntityID:           leID,
		Code:                    shiftCode,
		NameEn:                  "Engineering Shift",
		NameAr:                  "وردية الهندسة والتطوير",
		StartTime:               "08:00:00",
		EndTime:                 "17:00:00",
		GracePeriodMinutes:      15,
		BreakDurationMinutes:    60,
		HalfDayThresholdMinutes: 240,
		IsActive:                true,
	})
	if err != nil {
		t.Fatalf("failed to create shift: %v", err)
	}
	if shift.ID == uuid.Nil {
		t.Fatalf("expected valid shift id")
	}

	// 2. Assign Shift to Employee
	assign, err := svc.AssignShift(ctx, &domain.ShiftAssignment{
		EmployeeID: empID,
		ShiftID:    shift.ID,
		ValidFrom:  time.Now().AddDate(0, 0, -1),
	})
	if err != nil {
		t.Fatalf("failed to assign shift: %v", err)
	}
	if !assign.IsActive {
		t.Errorf("expected assignment to be active")
	}

	// 3. Verify Active Assignment
	activeAssign, err := attMod.Repo.GetActiveShiftAssignment(ctx, empID)
	if err != nil {
		t.Fatalf("failed to get active assignment: %v", err)
	}
	if activeAssign == nil || activeAssign.ShiftID != shift.ID {
		t.Fatalf("expected active assignment pointing to created shift")
	}
}

func TestAttendance_ClockInOutAndCalculation(t *testing.T) {
	attMod, _, _, leID, empID, _, ctx := setupAttendanceTest(t)
	svc := attMod.AttendanceService

	// 1. Create Shift with 08:00:00 start and 15 min grace
	shift, err := svc.CreateShift(ctx, &domain.Shift{
		LegalEntityID:           leID,
		Code:                    "SHIFT-PUNCH-" + uuid.New().String()[:6],
		NameEn:                  "Morning Shift",
		NameAr:                  "الوردية الصباحية",
		StartTime:               "08:00:00",
		EndTime:                 "17:00:00",
		GracePeriodMinutes:      15,
		BreakDurationMinutes:    60,
		HalfDayThresholdMinutes: 240,
		IsActive:                true,
	})
	if err != nil {
		t.Fatalf("failed to create shift: %v", err)
	}
	_, err = svc.AssignShift(ctx, &domain.ShiftAssignment{
		EmployeeID: empID,
		ShiftID:    shift.ID,
		ValidFrom:  time.Now().AddDate(0, 0, -1),
	})
	if err != nil {
		t.Fatalf("failed to assign shift: %v", err)
	}

	// 2. Clock In at 08:25 (25 min late from 08:00, beyond 15 min grace)
	now := time.Now()
	clockInTime := time.Date(now.Year(), now.Month(), now.Day(), 8, 25, 0, 0, time.UTC)
	record, err := svc.ClockIn(ctx, leID, empID, domain.MethodBiometric, clockInTime)
	if err != nil {
		t.Fatalf("failed to clock in: %v", err)
	}

	if record.LateMinutes != 25 {
		t.Errorf("expected 25 late minutes, got %d", record.LateMinutes)
	}
	if record.Status != domain.StatusPresent {
		t.Errorf("expected status PRESENT, got %s", record.Status)
	}

	// Test duplicate clock in rejection
	_, err = svc.ClockIn(ctx, leID, empID, domain.MethodBiometric, clockInTime)
	if err != domain.ErrAlreadyClockedIn {
		t.Errorf("expected ErrAlreadyClockedIn, got %v", err)
	}

	// 3. Clock Out at 17:30 (Total elapsed 9h 05m = 545m, minus 60m break = 485m = 8.08 hrs)
	clockOutTime := time.Date(now.Year(), now.Month(), now.Day(), 17, 30, 0, 0, time.UTC)
	outRecord, err := svc.ClockOut(ctx, leID, empID, domain.MethodBiometric, clockOutTime)
	if err != nil {
		t.Fatalf("failed to clock out: %v", err)
	}

	if outRecord.ClockOut == nil {
		t.Fatalf("expected valid clock out time")
	}
	if outRecord.ActualWorkHours.LessThan(decimal.NewFromFloat(8.0)) {
		t.Errorf("expected >= 8 actual work hours, got %s", outRecord.ActualWorkHours.String())
	}
	if outRecord.OvertimeMinutes < 5 {
		t.Errorf("expected positive overtime minutes beyond 8 hrs, got %d", outRecord.OvertimeMinutes)
	}
}

func TestAttendance_OvertimeWorkflow(t *testing.T) {
	attMod, _, _, leID, empID, adminID, ctx := setupAttendanceTest(t)
	svc := attMod.AttendanceService

	// 1. Submit Overtime Request (3 hours at 1.50 multiplier per Article 107)
	req, err := svc.SubmitOvertime(ctx, &domain.OvertimeRequest{
		LegalEntityID:  leID,
		EmployeeID:     empID,
		RequestDate:    time.Now(),
		OvertimeDate:   time.Now(),
		OvertimeHours:  decimal.NewFromFloat(3.0),
		RateMultiplier: decimal.NewFromFloat(1.50),
		Reason:         "Urgent financial month-end closing support",
	})
	if err != nil {
		t.Fatalf("failed to submit overtime: %v", err)
	}
	if req.Status != domain.StatusPending {
		t.Errorf("expected status PENDING, got %s", req.Status)
	}

	// 2. Approve Overtime
	approved, err := svc.ApproveOvertime(ctx, leID, req.ID, adminID)
	if err != nil {
		t.Fatalf("failed to approve overtime: %v", err)
	}
	if approved.Status != domain.StatusApproved {
		t.Errorf("expected status APPROVED, got %s", approved.Status)
	}
}

func TestAttendance_AdjustmentWorkflow(t *testing.T) {
	attMod, _, _, leID, empID, adminID, ctx := setupAttendanceTest(t)
	svc := attMod.AttendanceService

	// 1. Create a dummy attendance record with missed clock-out
	now := time.Now().AddDate(0, 0, -2)
	recTime := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC)
	record, err := attMod.Repo.CreateAttendanceRecord(ctx, &domain.AttendanceRecord{
		LegalEntityID:     leID,
		EmployeeID:        empID,
		WorkDate:          recTime,
		ClockIn:           &recTime,
		CheckInMethod:     domain.MethodManual,
		ExpectedWorkHours: decimal.NewFromInt(8),
		Status:            domain.StatusPresent,
	})
	if err != nil {
		t.Fatalf("failed to create attendance record: %v", err)
	}

	// 2. Submit adjustment for missed punch
	adjClockOut := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, time.UTC)
	adj, err := svc.SubmitAdjustment(ctx, &domain.AttendanceAdjustment{
		AttendanceRecordID:    record.ID,
		RequestedByEmployeeID: empID,
		OriginalClockIn:       &recTime,
		OriginalClockOut:      nil,
		AdjustedClockIn:       recTime,
		AdjustedClockOut:      adjClockOut,
		Reason:                "Biometric scanner network offline at end of shift",
	})
	if err != nil {
		t.Fatalf("failed to submit adjustment: %v", err)
	}
	if adj.Status != domain.StatusPending {
		t.Errorf("expected adjustment PENDING, got %s", adj.Status)
	}

	// 3. Approve adjustment
	approvedAdj, err := svc.ApproveAdjustment(ctx, adj.ID, adminID, "Approved after verifying CCTV logs")
	if err != nil {
		t.Fatalf("failed to approve adjustment: %v", err)
	}
	if approvedAdj.Status != domain.StatusApproved {
		t.Errorf("expected status APPROVED, got %s", approvedAdj.Status)
	}

	// 4. Verify attendance record is updated with 8 work hours
	updatedRec, err := attMod.Repo.GetAttendanceRecordByID(ctx, record.ID)
	if err != nil {
		t.Fatalf("failed to get updated record: %v", err)
	}
	if updatedRec.ClockOut == nil {
		t.Fatalf("expected updated record to have clock out")
	}
	if updatedRec.ActualWorkHours.LessThan(decimal.NewFromFloat(7.9)) {
		t.Errorf("expected 8 actual work hours after adjustment, got %s", updatedRec.ActualWorkHours.String())
	}
}

func TestAttendance_SeedDefaultData(t *testing.T) {
	attMod, _, _, leID, _, _, ctx := setupAttendanceTest(t)

	err := attMod.SeedDefaultAttendanceData(ctx, leID)
	if err != nil {
		t.Fatalf("failed to seed attendance data: %v", err)
	}

	shifts, err := attMod.AttendanceService.ListShifts(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list shifts: %v", err)
	}
	if len(shifts) < 2 {
		t.Errorf("expected at least 2 seeded shifts, got %d", len(shifts))
	}
}

func TestAttendance_AdjustmentApprovalRollsBackWhenRecordUpdateFails(t *testing.T) {
	attMod, _, db, leID, empID, adminID, ctx := setupAttendanceTest(t)
	defer db.Close()

	now := time.Now().AddDate(0, 0, -3)
	clockIn := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC)
	note := "ROLLBACK-ADJUSTMENT"
	record, err := attMod.Repo.CreateAttendanceRecord(ctx, &domain.AttendanceRecord{
		LegalEntityID: leID, EmployeeID: empID, WorkDate: clockIn,
		ClockIn: &clockIn, CheckInMethod: domain.MethodManual,
		ExpectedWorkHours: decimal.NewFromInt(8), Status: domain.StatusPresent, Notes: &note,
	})
	if err != nil {
		t.Fatalf("failed to create attendance record: %v", err)
	}
	clockOut := clockIn.Add(9 * time.Hour)
	adjustment, err := attMod.AttendanceService.SubmitAdjustment(ctx, &domain.AttendanceAdjustment{
		AttendanceRecordID: record.ID, RequestedByEmployeeID: empID,
		OriginalClockIn: &clockIn, AdjustedClockIn: clockIn, AdjustedClockOut: clockOut,
		Reason: "forced rollback",
	})
	if err != nil {
		t.Fatalf("failed to create attendance adjustment: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION attendance.fail_test_record_update()
		RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
			IF OLD.notes = 'ROLLBACK-ADJUSTMENT' THEN
				RAISE EXCEPTION 'forced attendance record failure';
			END IF;
			RETURN NEW;
		END;
		$$;
		DROP TRIGGER IF EXISTS trg_fail_test_record_update ON attendance.records;
		CREATE TRIGGER trg_fail_test_record_update
		BEFORE UPDATE ON attendance.records
		FOR EACH ROW EXECUTE FUNCTION attendance.fail_test_record_update();
	`)
	if err != nil {
		t.Fatalf("failed to install rollback test trigger: %v", err)
	}
	defer func() {
		_, _ = db.Pool.Exec(ctx, `
			DROP TRIGGER IF EXISTS trg_fail_test_record_update ON attendance.records;
			DROP FUNCTION IF EXISTS attendance.fail_test_record_update();
		`)
	}()

	if _, err := attMod.AttendanceService.ApproveAdjustment(ctx, adjustment.ID, adminID, "forced rollback"); err == nil {
		t.Fatal("expected forced attendance record update failure")
	}
	reloaded, err := attMod.Repo.GetAttendanceAdjustmentByID(ctx, adjustment.ID)
	if err != nil {
		t.Fatalf("failed to reload attendance adjustment: %v", err)
	}
	if reloaded.Status != domain.StatusPending {
		t.Fatalf("expected adjustment status rollback to PENDING, got %s", reloaded.Status)
	}
}
