package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetStatus string

const (
	BudgetStatusDraft     BudgetStatus = "DRAFT"
	BudgetStatusSubmitted BudgetStatus = "SUBMITTED"
	BudgetStatusApproved  BudgetStatus = "APPROVED"
	BudgetStatusActive    BudgetStatus = "ACTIVE"
	BudgetStatusRevised   BudgetStatus = "REVISED"
	BudgetStatusClosed    BudgetStatus = "CLOSED"
)

type BudgetType string

const (
	BudgetTypeOperating BudgetType = "OPERATING"
	BudgetTypeCapital   BudgetType = "CAPITAL"
	BudgetTypeProject   BudgetType = "PROJECT"
	BudgetTypeCashFlow  BudgetType = "CASH_FLOW"
)

type VersionStatus string

const (
	VersionStatusDraft      VersionStatus = "DRAFT"
	VersionStatusActive     VersionStatus = "ACTIVE"
	VersionStatusSuperseded VersionStatus = "SUPERSEDED"
)

type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "PENDING"
	TransferStatusCompleted TransferStatus = "COMPLETED"
	TransferStatusRejected  TransferStatus = "REJECTED"
)

// Budget represents the master budget header for an entity and fiscal year.
type Budget struct {
	ID                   uuid.UUID       `json:"id"`
	LegalEntityID        uuid.UUID       `json:"legal_entity_id"`
	Code                 string          `json:"code"`
	NameEn               string          `json:"name_en"`
	NameAr               string          `json:"name_ar"`
	DescriptionEn        string          `json:"description_en"`
	DescriptionAr        string          `json:"description_ar"`
	FiscalYearID         uuid.UUID       `json:"fiscal_year_id"`
	CurrencyCode         string          `json:"currency_code"`
	BudgetType           BudgetType      `json:"budget_type"`
	Status               BudgetStatus    `json:"status"`
	TotalBudgetedAmount  decimal.Decimal `json:"total_budgeted_amount"`
	TotalActualAmount    decimal.Decimal `json:"total_actual_amount"`
	TotalCommittedAmount decimal.Decimal `json:"total_committed_amount"`
	CreatedByUserID      *uuid.UUID      `json:"created_by_user_id,omitempty"`
	ApprovedByUserID     *uuid.UUID      `json:"approved_by_user_id,omitempty"`
	ApprovedAt           *time.Time      `json:"approved_at,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// BudgetVersion tracks revisions and historical baselines.
type BudgetVersion struct {
	ID               uuid.UUID     `json:"id"`
	BudgetID         uuid.UUID     `json:"budget_id"`
	VersionNumber    int32         `json:"version_number"`
	VersionName      string        `json:"version_name"`
	Status           VersionStatus `json:"status"`
	Notes            string        `json:"notes"`
	CreatedByUserID  *uuid.UUID    `json:"created_by_user_id,omitempty"`
	ApprovedByUserID *uuid.UUID    `json:"approved_by_user_id,omitempty"`
	ApprovedAt       *time.Time    `json:"approved_at,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// BudgetLine represents a multi-dimensional line allocation.
// Canonical dimensions per Section 40.24: account, department, cost_center, project, period.
type BudgetLine struct {
	ID                 uuid.UUID       `json:"id"`
	BudgetVersionID    uuid.UUID       `json:"budget_version_id"`
	AccountID          uuid.UUID       `json:"account_id"`
	DepartmentID       *uuid.UUID      `json:"department_id,omitempty"`
	CostCenterID       *uuid.UUID      `json:"cost_center_id,omitempty"`
	ProjectCode        string          `json:"project_code,omitempty"`
	FiscalPeriodID     *uuid.UUID      `json:"fiscal_period_id,omitempty"`
	Amount             decimal.Decimal `json:"amount"`
	ActualAmount       decimal.Decimal `json:"actual_amount"`
	CommittedAmount    decimal.Decimal `json:"committed_amount"`
	LineDescriptionEn  string          `json:"line_description_en"`
	LineDescriptionAr  string          `json:"line_description_ar"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	AccountCode        string          `json:"account_code,omitempty"`
	AccountNameEn      string          `json:"account_name_en,omitempty"`
	AccountNameAr      string          `json:"account_name_ar,omitempty"`
	DepartmentNameEn   string          `json:"department_name_en,omitempty"`
	DepartmentNameAr   string          `json:"department_name_ar,omitempty"`
	CostCenterNameEn   string          `json:"cost_center_name_en,omitempty"`
	CostCenterNameAr   string          `json:"cost_center_name_ar,omitempty"`
	PeriodNumber       *int16          `json:"period_number,omitempty"`
	PeriodNameEn       string          `json:"period_name_en,omitempty"`
	PeriodNameAr       string          `json:"period_name_ar,omitempty"`
}

// BudgetTransfer models inter-line fund transfers/reallocations.
type BudgetTransfer struct {
	ID                  uuid.UUID       `json:"id"`
	LegalEntityID       uuid.UUID       `json:"legal_entity_id"`
	TransferNumber      string          `json:"transfer_number"`
	SourceLineID        uuid.UUID       `json:"source_line_id"`
	TargetLineID        uuid.UUID       `json:"target_line_id"`
	Amount              decimal.Decimal `json:"amount"`
	Reason              string          `json:"reason"`
	Status              TransferStatus  `json:"status"`
	TransferredByUserID *uuid.UUID      `json:"transferred_by_user_id,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	SourceAccountCode   string          `json:"source_account_code,omitempty"`
	SourceAccountNameEn string          `json:"source_account_name_en,omitempty"`
	SourceAccountNameAr string          `json:"source_account_name_ar,omitempty"`
	TargetAccountCode   string          `json:"target_account_code,omitempty"`
	TargetAccountNameEn string          `json:"target_account_name_en,omitempty"`
	TargetAccountNameAr string          `json:"target_account_name_ar,omitempty"`
}

// BudgetVsActualItem represents variance analysis per line/account.
type BudgetVsActualItem struct {
	LineID           uuid.UUID       `json:"line_id"`
	AccountID        uuid.UUID       `json:"account_id"`
	AccountCode      string          `json:"account_code"`
	AccountNameEn    string          `json:"account_name_en"`
	AccountNameAr    string          `json:"account_name_ar"`
	DepartmentNameEn string          `json:"department_name_en,omitempty"`
	DepartmentNameAr string          `json:"department_name_ar,omitempty"`
	CostCenterNameEn string          `json:"cost_center_name_en,omitempty"`
	CostCenterNameAr string          `json:"cost_center_name_ar,omitempty"`
	PeriodNameEn     string          `json:"period_name_en,omitempty"`
	PeriodNameAr     string          `json:"period_name_ar,omitempty"`
	ProjectCode      string          `json:"project_code,omitempty"`
	BudgetedAmount   decimal.Decimal `json:"budgeted_amount"`
	ActualAmount     decimal.Decimal `json:"actual_amount"`
	CommittedAmount  decimal.Decimal `json:"committed_amount"`
	RemainingAmount  decimal.Decimal `json:"remaining_amount"`
	VarianceAmount   decimal.Decimal `json:"variance_amount"`
	ConsumptionRate  decimal.Decimal `json:"consumption_rate"` // percentage
	VarianceStatus   string          `json:"variance_status"`  // 'ON_TRACK', 'WARNING', 'OVER_BUDGET'
}

// BudgetSummary gives KPI aggregate metrics.
type BudgetSummary struct {
	TotalBudgeted     decimal.Decimal `json:"total_budgeted"`
	TotalActual       decimal.Decimal `json:"total_actual"`
	TotalCommitted    decimal.Decimal `json:"total_committed"`
	TotalRemaining    decimal.Decimal `json:"total_remaining"`
	AverageUsageRate  decimal.Decimal `json:"average_usage_rate"`
	ActiveBudgetCount int             `json:"active_budget_count"`
}
