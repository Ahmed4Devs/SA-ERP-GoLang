package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrLeaveTypeNotFound      = errors.New("leave type not found")
	ErrPolicyNotFound         = errors.New("leave policy not found")
	ErrInsufficientBalance    = errors.New("insufficient leave balance")
	ErrLeaveRequestNotFound   = errors.New("leave request not found")
	ErrProbationRestricted    = errors.New("employee is still in probation period; leave request not permitted by policy")
	ErrInvalidLeaveDates      = errors.New("end date must be after or equal to start date")
	ErrAlreadyProcessed       = errors.New("leave request has already been processed")
)

type LeaveStatusCode string

const (
	RequestPending   LeaveStatusCode = "PENDING"
	RequestApproved  LeaveStatusCode = "APPROVED"
	RequestRejected  LeaveStatusCode = "REJECTED"
	RequestCancelled LeaveStatusCode = "CANCELLED"
)

type AccrualFrequency string

const (
	AccrualMonthly AccrualFrequency = "MONTHLY"
	AccrualAnnual  AccrualFrequency = "ANNUAL"
	AccrualNone    AccrualFrequency = "NONE"
)

type TransactionType string

const (
	TxAccrual      TransactionType = "ACCRUAL"
	TxUsage        TransactionType = "USAGE"
	TxCancellation TransactionType = "CANCELLATION"
	TxCarryover    TransactionType = "CARRYOVER"
	TxAdjustment   TransactionType = "ADJUSTMENT"
	TxEncashment   TransactionType = "ENCASHMENT"
)

type LeaveType struct {
	ID                 uuid.UUID        `json:"id"`
	LegalEntityID      uuid.UUID        `json:"legal_entity_id"`
	Code               string           `json:"code"`
	NameEn             string           `json:"name_en"`
	NameAr             string           `json:"name_ar"`
	IsPaid             bool             `json:"is_paid"`
	RequiresAttachment bool             `json:"requires_attachment"`
	DeductsFromAnnual  bool             `json:"deducts_from_annual"`
	MaxDaysPerYear     *decimal.Decimal `json:"max_days_per_year,omitempty"`
	IsActive           bool             `json:"is_active"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type LeavePolicy struct {
	ID                     uuid.UUID        `json:"id"`
	LegalEntityID          uuid.UUID        `json:"legal_entity_id"`
	LeaveTypeID            uuid.UUID        `json:"leave_type_id"`
	LeaveTypeCode          *string          `json:"leave_type_code,omitempty"`
	LeaveTypeNameEn        *string          `json:"leave_type_name_en,omitempty"`
	LeaveTypeNameAr        *string          `json:"leave_type_name_ar,omitempty"`
	AccrualFrequency       AccrualFrequency `json:"accrual_frequency"`
	AccrualRatePerPeriod   decimal.Decimal  `json:"accrual_rate_per_period"`
	MaxCarryoverDays      decimal.Decimal  `json:"max_carryover_days"`
	ProbationRestricted    bool             `json:"probation_restricted"`
	EncashmentAllowed      bool             `json:"encashment_allowed"`
	MinServiceDaysRequired int32            `json:"min_service_days_required"`
	IsActive               bool             `json:"is_active"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

type EmployeeLeaveBalance struct {
	ID              uuid.UUID       `json:"id"`
	EmployeeID      uuid.UUID       `json:"employee_id"`
	LeaveTypeID     uuid.UUID       `json:"leave_type_id"`
	LeaveTypeCode   *string         `json:"leave_type_code,omitempty"`
	LeaveTypeNameEn *string         `json:"leave_type_name_en,omitempty"`
	LeaveTypeNameAr *string         `json:"leave_type_name_ar,omitempty"`
	FiscalYear      int32           `json:"fiscal_year"`
	AllocatedDays   decimal.Decimal `json:"allocated_days"`
	AccruedDays     decimal.Decimal `json:"accrued_days"`
	TakenDays       decimal.Decimal `json:"taken_days"`
	PendingDays     decimal.Decimal `json:"pending_days"`
	CarryoverDays   decimal.Decimal `json:"carryover_days"`
	ClosingBalance  decimal.Decimal `json:"closing_balance"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type LeaveRequest struct {
	ID                    uuid.UUID       `json:"id"`
	LegalEntityID         uuid.UUID       `json:"legal_entity_id"`
	EmployeeID            uuid.UUID       `json:"employee_id"`
	EmployeeNumber        *string         `json:"employee_number,omitempty"`
	EmployeeNameEn        *string         `json:"employee_name_en,omitempty"`
	EmployeeNameAr        *string         `json:"employee_name_ar,omitempty"`
	LeaveTypeID           uuid.UUID       `json:"leave_type_id"`
	LeaveTypeCode         *string         `json:"leave_type_code,omitempty"`
	LeaveTypeNameEn       *string         `json:"leave_type_name_en,omitempty"`
	LeaveTypeNameAr       *string         `json:"leave_type_name_ar,omitempty"`
	RequestNumber         string          `json:"request_number"`
	StartDate             time.Time       `json:"start_date"`
	EndDate               time.Time       `json:"end_date"`
	TotalDays             decimal.Decimal `json:"total_days"`
	ReturnDate            time.Time       `json:"return_date"`
	ReplacementEmployeeID *uuid.UUID      `json:"replacement_employee_id,omitempty"`
	Reason                string          `json:"reason"`
	AttachmentURL         *string         `json:"attachment_url,omitempty"`
	Status                LeaveStatusCode `json:"status"`
	ReviewedByUserID      *uuid.UUID      `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt            *time.Time      `json:"reviewed_at,omitempty"`
	ReviewNotes           *string         `json:"review_notes,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type LeaveTransaction struct {
	ID              uuid.UUID       `json:"id"`
	EmployeeID      uuid.UUID       `json:"employee_id"`
	LeaveTypeID     uuid.UUID       `json:"leave_type_id"`
	FiscalYear      int32           `json:"fiscal_year"`
	TransactionType TransactionType `json:"transaction_type"`
	Days            decimal.Decimal `json:"days"`
	ReferenceID     *uuid.UUID      `json:"reference_id,omitempty"`
	Notes           *string         `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// CalculateWorkingDays counts weekdays excluding Friday and Saturday (Saudi standard weekend)
func CalculateWorkingDays(start, end time.Time) decimal.Decimal {
	if end.Before(start) {
		return decimal.Zero
	}

	days := 0
	curr := start
	for !curr.After(end) {
		weekday := curr.Weekday()
		if weekday != time.Friday && weekday != time.Saturday {
			days++
		}
		curr = curr.AddDate(0, 0, 1)
	}
	return decimal.NewFromInt(int64(days))
}
