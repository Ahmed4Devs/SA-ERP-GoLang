package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/attendance/domain"
	"erp/internal/attendance/ports"
	"erp/internal/platform/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type attendanceService struct {
	repo ports.AttendanceRepository
	db   *database.DB
}

func NewAttendanceService(repo ports.AttendanceRepository, db *database.DB) ports.AttendanceService {
	return &attendanceService{
		repo: repo,
		db:   db,
	}
}

// ============================================================================
// SHIFTS
// ============================================================================

func (s *attendanceService) CreateShift(ctx context.Context, shift *domain.Shift) (*domain.Shift, error) {
	if shift.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if shift.Code == "" {
		return nil, fmt.Errorf("shift code is required")
	}
	if shift.NameEn == "" || shift.NameAr == "" {
		return nil, fmt.Errorf("bilingual shift names are required")
	}
	if shift.StartTime == "" {
		shift.StartTime = "08:00:00"
	}
	if shift.EndTime == "" {
		shift.EndTime = "17:00:00"
	}
	if shift.GracePeriodMinutes <= 0 {
		shift.GracePeriodMinutes = 15
	}
	if shift.BreakDurationMinutes <= 0 {
		shift.BreakDurationMinutes = 60
	}
	if shift.HalfDayThresholdMinutes <= 0 {
		shift.HalfDayThresholdMinutes = 240
	}
	return s.repo.CreateShift(ctx, shift)
}

func (s *attendanceService) ListShifts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Shift, error) {
	return s.repo.ListShifts(ctx, legalEntityID)
}

func (s *attendanceService) AssignShift(ctx context.Context, assignment *domain.ShiftAssignment) (*domain.ShiftAssignment, error) {
	if assignment.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}
	if assignment.ShiftID == uuid.Nil {
		return nil, fmt.Errorf("shift_id is required")
	}
	if assignment.ValidFrom.IsZero() {
		assignment.ValidFrom = time.Now()
	}
	assignment.IsActive = true
	return s.repo.CreateShiftAssignment(ctx, assignment)
}

// ============================================================================
// CLOCK IN / CLOCK OUT
// ============================================================================

func (s *attendanceService) ClockIn(
	ctx context.Context,
	legalEntityID, employeeID uuid.UUID,
	method domain.CheckMethod,
	clockInTime time.Time,
) (*domain.AttendanceRecord, error) {
	if clockInTime.IsZero() {
		clockInTime = time.Now()
	}
	workDate := time.Date(clockInTime.Year(), clockInTime.Month(), clockInTime.Day(), 0, 0, 0, 0, time.UTC)

	// 1. Check if record already exists
	existing, err := s.repo.GetAttendanceRecordByEmployeeAndDate(ctx, employeeID, workDate)
	if err == nil && existing != nil && existing.ClockIn != nil {
		return nil, domain.ErrAlreadyClockedIn
	}

	// 2. Fetch active shift assignment
	assignment, _ := s.repo.GetActiveShiftAssignment(ctx, employeeID)
	var shiftID *uuid.UUID
	lateMinutes := int32(0)
	expectedHours := decimal.NewFromInt(8)

	if assignment != nil {
		shiftID = &assignment.ShiftID
		shift, err := s.repo.GetShiftByID(ctx, legalEntityID, assignment.ShiftID)
		if err == nil && shift != nil {
			lateMinutes = domain.CalculateLateArrival(clockInTime, shift.StartTime, shift.GracePeriodMinutes)
		}
	}

	rec := &domain.AttendanceRecord{
		LegalEntityID:     legalEntityID,
		EmployeeID:        employeeID,
		ShiftID:           shiftID,
		WorkDate:          workDate,
		ClockIn:           &clockInTime,
		CheckInMethod:     method,
		ExpectedWorkHours: expectedHours,
		LateMinutes:       lateMinutes,
		Status:            domain.StatusPresent,
	}

	return s.repo.CreateAttendanceRecord(ctx, rec)
}

func (s *attendanceService) ClockOut(
	ctx context.Context,
	legalEntityID, employeeID uuid.UUID,
	method domain.CheckMethod,
	clockOutTime time.Time,
) (*domain.AttendanceRecord, error) {
	if clockOutTime.IsZero() {
		clockOutTime = time.Now()
	}
	workDate := time.Date(clockOutTime.Year(), clockOutTime.Month(), clockOutTime.Day(), 0, 0, 0, 0, time.UTC)

	// 1. Get today's attendance record
	record, err := s.repo.GetAttendanceRecordByEmployeeAndDate(ctx, employeeID, workDate)
	if err != nil || record == nil || record.ClockIn == nil {
		return nil, domain.ErrNotClockedIn
	}
	if record.ClockOut != nil {
		return nil, domain.ErrAlreadyClockedOut
	}
	if clockOutTime.Before(*record.ClockIn) {
		return nil, domain.ErrInvalidTimeRange
	}

	// 2. Determine shift end time & break
	shiftEndStr := "17:00:00"
	breakMinutes := int32(60)
	if record.ShiftID != nil {
		shift, err := s.repo.GetShiftByID(ctx, legalEntityID, *record.ShiftID)
		if err == nil && shift != nil {
			shiftEndStr = shift.EndTime
			breakMinutes = shift.BreakDurationMinutes
		}
	}

	// 3. Compute actual hours, early departure, overtime
	actualHours, earlyMinutes, overtimeMinutes := domain.CalculateHoursAndEarlyDeparture(
		*record.ClockIn, clockOutTime, shiftEndStr, breakMinutes, record.ExpectedWorkHours,
	)

	record.ClockOut = &clockOutTime
	record.CheckOutMethod = &method
	record.ActualWorkHours = actualHours
	record.EarlyDepartureMinutes = earlyMinutes
	record.OvertimeMinutes = overtimeMinutes

	// Status adjustments (e.g. if actual hours < 4 -> HALF_DAY)
	if actualHours.LessThan(decimal.NewFromInt(4)) {
		record.Status = domain.StatusHalfDay
	} else {
		record.Status = domain.StatusPresent
	}

	return s.repo.UpdateAttendanceRecordClockOut(ctx, record)
}

func (s *attendanceService) GetDailyRecords(ctx context.Context, legalEntityID uuid.UUID, date time.Time) ([]*domain.AttendanceRecord, error) {
	if date.IsZero() {
		date = time.Now()
	}
	workDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	return s.repo.ListAttendanceRecordsByDate(ctx, legalEntityID, workDate)
}

func (s *attendanceService) GetEmployeeRecords(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]*domain.AttendanceRecord, error) {
	if from.IsZero() {
		from = time.Now().AddDate(0, 0, -30)
	}
	if to.IsZero() {
		to = time.Now()
	}
	return s.repo.ListEmployeeAttendanceRecords(ctx, employeeID, from, to)
}

// ============================================================================
// OVERTIME
// ============================================================================

func (s *attendanceService) SubmitOvertime(ctx context.Context, req *domain.OvertimeRequest) (*domain.OvertimeRequest, error) {
	if req.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if req.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}
	if req.OvertimeHours.IsZero() || req.OvertimeHours.IsNegative() {
		return nil, fmt.Errorf("overtime hours must be greater than zero")
	}
	if req.RateMultiplier.IsZero() {
		req.RateMultiplier = decimal.NewFromFloat(1.50) // Article 107
	}
	if req.RequestDate.IsZero() {
		req.RequestDate = time.Now()
	}
	if req.OvertimeDate.IsZero() {
		req.OvertimeDate = time.Now()
	}
	req.Status = domain.StatusPending
	return s.repo.CreateOvertimeRequest(ctx, req)
}

func (s *attendanceService) ApproveOvertime(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID) (*domain.OvertimeRequest, error) {
	return s.repo.UpdateOvertimeStatus(ctx, legalEntityID, id, domain.StatusApproved, approverUserID)
}

func (s *attendanceService) RejectOvertime(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID) (*domain.OvertimeRequest, error) {
	return s.repo.UpdateOvertimeStatus(ctx, legalEntityID, id, domain.StatusRejected, approverUserID)
}

func (s *attendanceService) ListOvertime(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.OvertimeRequest, error) {
	return s.repo.ListOvertimeRequests(ctx, legalEntityID)
}

// ============================================================================
// ADJUSTMENTS
// ============================================================================

func (s *attendanceService) SubmitAdjustment(ctx context.Context, adj *domain.AttendanceAdjustment) (*domain.AttendanceAdjustment, error) {
	if adj.AttendanceRecordID == uuid.Nil {
		return nil, fmt.Errorf("attendance_record_id is required")
	}
	if adj.RequestedByEmployeeID == uuid.Nil {
		return nil, fmt.Errorf("requested_by_employee_id is required")
	}
	if adj.AdjustedClockOut.Before(adj.AdjustedClockIn) {
		return nil, domain.ErrInvalidTimeRange
	}
	adj.Status = domain.StatusPending
	return s.repo.CreateAttendanceAdjustment(ctx, adj)
}

func (s *attendanceService) ApproveAdjustment(ctx context.Context, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.AttendanceAdjustment, error) {
	adj, err := s.repo.GetAttendanceAdjustmentByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updatedAdj, err := s.repo.UpdateAdjustmentStatus(ctx, id, domain.StatusApproved, approverUserID, notes)
	if err != nil {
		return nil, err
	}

	// Update attendance record with adjusted times
	record, err := s.repo.GetAttendanceRecordByID(ctx, adj.AttendanceRecordID)
	if err == nil && record != nil {
		record.ClockIn = &adj.AdjustedClockIn
		record.ClockOut = &adj.AdjustedClockOut
		actualHours, earlyMin, otMin := domain.CalculateHoursAndEarlyDeparture(
			adj.AdjustedClockIn, adj.AdjustedClockOut, "17:00:00", 60, record.ExpectedWorkHours,
		)
		record.ActualWorkHours = actualHours
		record.EarlyDepartureMinutes = earlyMin
		record.OvertimeMinutes = otMin
		record.Status = domain.StatusPresent
		_, _ = s.repo.UpdateAttendanceRecordClockOut(ctx, record)
	}

	return updatedAdj, nil
}
