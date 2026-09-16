package ports

import (
	"context"
	"time"

	"erp/internal/attendance/domain"
	"github.com/google/uuid"
)

type AttendanceRepository interface {
	// Shifts
	CreateShift(ctx context.Context, shift *domain.Shift) (*domain.Shift, error)
	GetShiftByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Shift, error)
	GetShiftByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Shift, error)
	ListShifts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Shift, error)

	// Shift Assignments
	CreateShiftAssignment(ctx context.Context, assignment *domain.ShiftAssignment) (*domain.ShiftAssignment, error)
	GetActiveShiftAssignment(ctx context.Context, employeeID uuid.UUID) (*domain.ShiftAssignment, error)
	ListShiftAssignments(ctx context.Context, employeeID uuid.UUID) ([]*domain.ShiftAssignment, error)

	// Attendance Records
	CreateAttendanceRecord(ctx context.Context, record *domain.AttendanceRecord) (*domain.AttendanceRecord, error)
	GetAttendanceRecordByID(ctx context.Context, id uuid.UUID) (*domain.AttendanceRecord, error)
	GetAttendanceRecordByEmployeeAndDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (*domain.AttendanceRecord, error)
	UpdateAttendanceRecordClockOut(ctx context.Context, record *domain.AttendanceRecord) (*domain.AttendanceRecord, error)
	ListAttendanceRecordsByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) ([]*domain.AttendanceRecord, error)
	ListEmployeeAttendanceRecords(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]*domain.AttendanceRecord, error)

	// Overtime Requests
	CreateOvertimeRequest(ctx context.Context, req *domain.OvertimeRequest) (*domain.OvertimeRequest, error)
	GetOvertimeRequestByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.OvertimeRequest, error)
	ListOvertimeRequests(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.OvertimeRequest, error)
	UpdateOvertimeStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.RequestStatus, approvedBy uuid.UUID) (*domain.OvertimeRequest, error)

	// Adjustments
	CreateAttendanceAdjustment(ctx context.Context, adj *domain.AttendanceAdjustment) (*domain.AttendanceAdjustment, error)
	GetAttendanceAdjustmentByID(ctx context.Context, id uuid.UUID) (*domain.AttendanceAdjustment, error)
	UpdateAdjustmentStatus(ctx context.Context, id uuid.UUID, status domain.RequestStatus, reviewedBy uuid.UUID, reviewNotes string) (*domain.AttendanceAdjustment, error)
}

type AttendanceService interface {
	// Shifts
	CreateShift(ctx context.Context, shift *domain.Shift) (*domain.Shift, error)
	ListShifts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Shift, error)
	AssignShift(ctx context.Context, assignment *domain.ShiftAssignment) (*domain.ShiftAssignment, error)

	// Clock In / Clock Out
	ClockIn(ctx context.Context, legalEntityID, employeeID uuid.UUID, method domain.CheckMethod, clockInTime time.Time) (*domain.AttendanceRecord, error)
	ClockOut(ctx context.Context, legalEntityID, employeeID uuid.UUID, method domain.CheckMethod, clockOutTime time.Time) (*domain.AttendanceRecord, error)
	GetDailyRecords(ctx context.Context, legalEntityID uuid.UUID, date time.Time) ([]*domain.AttendanceRecord, error)
	GetEmployeeRecords(ctx context.Context, employeeID uuid.UUID, from, to time.Time) ([]*domain.AttendanceRecord, error)

	// Overtime
	SubmitOvertime(ctx context.Context, req *domain.OvertimeRequest) (*domain.OvertimeRequest, error)
	ApproveOvertime(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID) (*domain.OvertimeRequest, error)
	RejectOvertime(ctx context.Context, legalEntityID, id uuid.UUID, approverUserID uuid.UUID) (*domain.OvertimeRequest, error)
	ListOvertime(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.OvertimeRequest, error)

	// Adjustments
	SubmitAdjustment(ctx context.Context, adj *domain.AttendanceAdjustment) (*domain.AttendanceAdjustment, error)
	ApproveAdjustment(ctx context.Context, id uuid.UUID, approverUserID uuid.UUID, notes string) (*domain.AttendanceAdjustment, error)
}
