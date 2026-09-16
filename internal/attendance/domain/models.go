package domain

import (
	"errors"
	// "fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrShiftNotFound      = errors.New("shift not found")
	ErrRecordNotFound     = errors.New("attendance record not found")
	ErrAdjustmentNotFound = errors.New("attendance adjustment not found")
	ErrOvertimeNotFound   = errors.New("overtime request not found")
	ErrAlreadyClockedIn   = errors.New("employee already clocked in for this work date")
	ErrNotClockedIn       = errors.New("employee has not clocked in yet")
	ErrAlreadyClockedOut  = errors.New("employee already clocked out for this work date")
	ErrInvalidTimeRange   = errors.New("clock out time must be after clock in time")
)

type AttendanceStatus string

const (
	StatusPresent AttendanceStatus = "PRESENT"
	StatusAbsent  AttendanceStatus = "ABSENT"
	StatusHalfDay AttendanceStatus = "HALF_DAY"
	StatusOnLeave AttendanceStatus = "ON_LEAVE"
	StatusRestDay AttendanceStatus = "REST_DAY"
	StatusHoliday AttendanceStatus = "HOLIDAY"
)

type CheckMethod string

const (
	MethodManual    CheckMethod = "MANUAL"
	MethodBiometric CheckMethod = "BIOMETRIC"
	MethodMobileGPS CheckMethod = "MOBILE_GPS"
	MethodSystem    CheckMethod = "SYSTEM"
)

type RequestStatus string

const (
	StatusPending  RequestStatus = "PENDING"
	StatusApproved RequestStatus = "APPROVED"
	StatusRejected RequestStatus = "REJECTED"
)

type Shift struct {
	ID                      uuid.UUID `json:"id"`
	LegalEntityID           uuid.UUID `json:"legal_entity_id"`
	Code                    string    `json:"code"`
	NameEn                  string    `json:"name_en"`
	NameAr                  string    `json:"name_ar"`
	StartTime               string    `json:"start_time"` // "08:00:00"
	EndTime                 string    `json:"end_time"`   // "17:00:00"
	GracePeriodMinutes      int32     `json:"grace_period_minutes"`
	BreakDurationMinutes    int32     `json:"break_duration_minutes"`
	HalfDayThresholdMinutes int32     `json:"half_day_threshold_minutes"`
	IsNightShift            bool      `json:"is_night_shift"`
	WorkDays                []string  `json:"work_days"`
	IsActive                bool      `json:"is_active"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type ShiftAssignment struct {
	ID          uuid.UUID  `json:"id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	ShiftID     uuid.UUID  `json:"shift_id"`
	ShiftCode   *string    `json:"shift_code,omitempty"`
	ShiftNameEn *string    `json:"shift_name_en,omitempty"`
	ShiftNameAr *string    `json:"shift_name_ar,omitempty"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidTo     *time.Time `json:"valid_to,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
}

type AttendanceRecord struct {
	ID                    uuid.UUID        `json:"id"`
	LegalEntityID         uuid.UUID        `json:"legal_entity_id"`
	EmployeeID            uuid.UUID        `json:"employee_id"`
	EmployeeNumber        *string          `json:"employee_number,omitempty"`
	EmployeeNameEn        *string          `json:"employee_name_en,omitempty"`
	EmployeeNameAr        *string          `json:"employee_name_ar,omitempty"`
	ShiftID               *uuid.UUID       `json:"shift_id,omitempty"`
	WorkDate              time.Time        `json:"work_date"`
	ClockIn               *time.Time       `json:"clock_in,omitempty"`
	ClockOut              *time.Time       `json:"clock_out,omitempty"`
	CheckInMethod         CheckMethod      `json:"check_in_method"`
	CheckOutMethod        *CheckMethod     `json:"check_out_method,omitempty"`
	ExpectedWorkHours     decimal.Decimal  `json:"expected_work_hours"`
	ActualWorkHours       decimal.Decimal  `json:"actual_work_hours"`
	LateMinutes           int32            `json:"late_minutes"`
	EarlyDepartureMinutes int32            `json:"early_departure_minutes"`
	OvertimeMinutes       int32            `json:"overtime_minutes"`
	Status                AttendanceStatus `json:"status"`
	Notes                 *string          `json:"notes,omitempty"`
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
}

type AttendanceAdjustment struct {
	ID                    uuid.UUID     `json:"id"`
	AttendanceRecordID    uuid.UUID     `json:"attendance_record_id"`
	RequestedByEmployeeID uuid.UUID     `json:"requested_by_employee_id"`
	OriginalClockIn       *time.Time    `json:"original_clock_in,omitempty"`
	OriginalClockOut      *time.Time    `json:"original_clock_out,omitempty"`
	AdjustedClockIn       time.Time     `json:"adjusted_clock_in"`
	AdjustedClockOut      time.Time     `json:"adjusted_clock_out"`
	Reason                string        `json:"reason"`
	Status                RequestStatus `json:"status"`
	ReviewedByUserID      *uuid.UUID    `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt            *time.Time    `json:"reviewed_at,omitempty"`
	ReviewNotes           *string       `json:"review_notes,omitempty"`
	CreatedAt             time.Time     `json:"created_at"`
}

type OvertimeRequest struct {
	ID               uuid.UUID       `json:"id"`
	LegalEntityID    uuid.UUID       `json:"legal_entity_id"`
	EmployeeID       uuid.UUID       `json:"employee_id"`
	EmployeeNumber   *string         `json:"employee_number,omitempty"`
	EmployeeNameEn   *string         `json:"employee_name_en,omitempty"`
	EmployeeNameAr   *string         `json:"employee_name_ar,omitempty"`
	RequestDate      time.Time       `json:"request_date"`
	OvertimeDate     time.Time       `json:"overtime_date"`
	OvertimeHours    decimal.Decimal `json:"overtime_hours"`
	RateMultiplier   decimal.Decimal `json:"rate_multiplier"` // Article 107 default 1.50
	Reason           string          `json:"reason"`
	Status           RequestStatus   `json:"status"`
	ApprovedByUserID *uuid.UUID      `json:"approved_by_user_id,omitempty"`
	ApprovedAt       *time.Time      `json:"approved_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// CalculateLateArrival evaluates late minutes beyond the shift grace period
func CalculateLateArrival(clockIn time.Time, shiftStartStr string, graceMinutes int32) int32 {
	parsedTime, err := time.Parse("15:04:05", shiftStartStr)
	if err != nil {
		return 0
	}

	shiftStartToday := time.Date(
		clockIn.Year(), clockIn.Month(), clockIn.Day(),
		parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0, clockIn.Location(),
	)

	allowedTime := shiftStartToday.Add(time.Duration(graceMinutes) * time.Minute)
	if clockIn.After(allowedTime) {
		diff := clockIn.Sub(shiftStartToday)
		return int32(math.Round(diff.Minutes()))
	}
	return 0
}

// CalculateHoursAndEarlyDeparture determines actual work duration and early departure
func CalculateHoursAndEarlyDeparture(
	clockIn, clockOut time.Time,
	shiftEndStr string,
	breakMinutes int32,
	expectedHours decimal.Decimal,
) (actualHours decimal.Decimal, earlyMinutes int32, overtimeMinutes int32) {
	duration := clockOut.Sub(clockIn)
	netMinutes := duration.Minutes() - float64(breakMinutes)
	if netMinutes < 0 {
		netMinutes = 0
	}

	actualHours = decimal.NewFromFloat(netMinutes / 60.0).Round(2)

	// Early departure check
	if shiftEndStr != "" {
		if parsedEnd, err := time.Parse("15:04:05", shiftEndStr); err == nil {
			shiftEndToday := time.Date(
				clockOut.Year(), clockOut.Month(), clockOut.Day(),
				parsedEnd.Hour(), parsedEnd.Minute(), parsedEnd.Second(), 0, clockOut.Location(),
			)
			if clockOut.Before(shiftEndToday) {
				diff := shiftEndToday.Sub(clockOut)
				earlyMinutes = int32(math.Round(diff.Minutes()))
			}
		}
	}

	// Calculate overtime if worked beyond expected hours
	expMinutes, _ := expectedHours.Mul(decimal.NewFromInt(60)).Float64()
	if netMinutes > expMinutes {
		overtimeMinutes = int32(math.Round(netMinutes - expMinutes))
	}

	return actualHours, earlyMinutes, overtimeMinutes
}

// DailyAttendanceSummary represents KPI aggregated statistics
type DailyAttendanceSummary struct {
	TotalAssigned  int64           `json:"total_assigned"`
	PresentCount   int64           `json:"present_count"`
	AbsentCount    int64           `json:"absent_count"`
	LateCount      int64           `json:"late_count"`
	OnLeaveCount   int64           `json:"on_leave_count"`
	AttendanceRate decimal.Decimal `json:"attendance_rate_percent"`
}
