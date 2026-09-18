package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrJobNotFound             = errors.New("job not found")
	ErrPositionNotFound        = errors.New("position not found")
	ErrEmployeeNotFound        = errors.New("employee not found")
	ErrContractNotFound        = errors.New("contract not found")
	ErrInvalidSalary           = errors.New("basic salary must be greater than or equal to zero")
	ErrInvalidProbationPeriod  = errors.New("probation period cannot exceed 180 days")
	ErrInvalidStatusTransition = errors.New("invalid employment or contract status transition")
)

type JobFamily string

const (
	JobFamilyEngineering JobFamily = "ENGINEERING"
	JobFamilyFinance     JobFamily = "FINANCE"
	JobFamilySales       JobFamily = "SALES"
	JobFamilyOperations  JobFamily = "OPERATIONS"
	JobFamilyHR          JobFamily = "HR"
	JobFamilyLegal       JobFamily = "LEGAL"
	JobFamilyManagement  JobFamily = "MANAGEMENT"
)

type JobLevel string

const (
	JobLevelEntry     JobLevel = "ENTRY"
	JobLevelMid       JobLevel = "MID"
	JobLevelSenior    JobLevel = "SENIOR"
	JobLevelLead      JobLevel = "LEAD"
	JobLevelManager   JobLevel = "MANAGER"
	JobLevelDirector  JobLevel = "DIRECTOR"
	JobLevelExecutive JobLevel = "EXECUTIVE"
)

type PositionStatus string

const (
	PositionVacant PositionStatus = "VACANT"
	PositionFilled PositionStatus = "FILLED"
	PositionFrozen PositionStatus = "FROZEN"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

type MaritalStatus string

const (
	MaritalStatusSingle   MaritalStatus = "SINGLE"
	MaritalStatusMarried  MaritalStatus = "MARRIED"
	MaritalStatusDivorced MaritalStatus = "DIVORCED"
	MaritalStatusWidowed  MaritalStatus = "WIDOWED"
)

type EmploymentStatus string

const (
	EmploymentStatusProbation  EmploymentStatus = "PROBATION"
	EmploymentStatusActive     EmploymentStatus = "ACTIVE"
	EmploymentStatusOnLeave    EmploymentStatus = "ON_LEAVE"
	EmploymentStatusSuspended  EmploymentStatus = "SUSPENDED"
	EmploymentStatusTerminated EmploymentStatus = "TERMINATED"
)

type IdentifierType string

const (
	IdentifierNationalID   IdentifierType = "NATIONAL_ID"
	IdentifierIqama        IdentifierType = "IQAMA"
	IdentifierPassport     IdentifierType = "PASSPORT"
	IdentifierBorderNumber IdentifierType = "BORDER_NUMBER"
)

type ContractType string

const (
	ContractUnspecifiedTerm ContractType = "UNSPECIFIED_TERM"
	ContractSpecifiedTerm   ContractType = "SPECIFIED_TERM"
	ContractTemporary       ContractType = "TEMPORARY"
	ContractSeasonal        ContractType = "SEASONAL"
)

type ContractStatus string

const (
	ContractDraft           ContractStatus = "DRAFT"
	ContractPendingApproval ContractStatus = "PENDING_APPROVAL"
	ContractActive          ContractStatus = "ACTIVE"
	ContractExpired         ContractStatus = "EXPIRED"
	ContractTerminated      ContractStatus = "TERMINATED"
)

type Job struct {
	ID            uuid.UUID `json:"id"`
	LegalEntityID uuid.UUID `json:"legal_entity_id"`
	Code          string    `json:"code"`
	TitleEn       string    `json:"title_en"`
	TitleAr       string    `json:"title_ar"`
	JobFamily     JobFamily `json:"job_family"`
	JobLevel      JobLevel  `json:"job_level"`
	Description   *string   `json:"description,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Position struct {
	ID                  uuid.UUID      `json:"id"`
	LegalEntityID       uuid.UUID      `json:"legal_entity_id"`
	Code                string         `json:"code"`
	JobID               uuid.UUID      `json:"job_id"`
	DepartmentID        uuid.UUID      `json:"department_id"`
	BranchID            *uuid.UUID     `json:"branch_id,omitempty"`
	ReportsToPositionID *uuid.UUID     `json:"reports_to_position_id,omitempty"`
	Headcount           int32          `json:"headcount"`
	Status              PositionStatus `json:"status"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type Employee struct {
	ID                     uuid.UUID        `json:"id"`
	LegalEntityID          uuid.UUID        `json:"legal_entity_id"`
	EmployeeNumber         string           `json:"employee_number"`
	UserID                 *uuid.UUID       `json:"user_id,omitempty"`
	FirstNameEn            string           `json:"first_name_en"`
	LastNameEn             string           `json:"last_name_en"`
	FirstNameAr            string           `json:"first_name_ar"`
	LastNameAr             string           `json:"last_name_ar"`
	Gender                 Gender           `json:"gender"`
	MaritalStatus          MaritalStatus    `json:"marital_status"`
	DateOfBirth            time.Time        `json:"date_of_birth"`
	NationalityCountryCode string           `json:"nationality_country_code"`
	IsSaudiNational        bool             `json:"is_saudi_national"`
	HireDate               time.Time        `json:"hire_date"`
	ProbationEndDate       *time.Time       `json:"probation_end_date,omitempty"`
	TerminationDate        *time.Time       `json:"termination_date,omitempty"`
	EmploymentStatus       EmploymentStatus `json:"employment_status"`
	PrimaryPositionID      *uuid.UUID       `json:"primary_position_id,omitempty"`
	DepartmentID           *uuid.UUID       `json:"department_id,omitempty"`
	BranchID               *uuid.UUID       `json:"branch_id,omitempty"`
	ManagerEmployeeID      *uuid.UUID       `json:"manager_employee_id,omitempty"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

type EmployeeIdentifier struct {
	ID                    uuid.UUID      `json:"id"`
	EmployeeID            uuid.UUID      `json:"employee_id"`
	IdentifierType        IdentifierType `json:"identifier_type"`
	IdentifierValue       string         `json:"identifier_value"`
	IdentifierValueMasked string         `json:"identifier_value_masked"`
	IssueDate             *time.Time     `json:"issue_date,omitempty"`
	ExpiryDate            *time.Time     `json:"expiry_date,omitempty"`
	IssuePlace            *string        `json:"issue_place,omitempty"`
	IsPrimary             bool           `json:"is_primary"`
	IsVerified            bool           `json:"is_verified"`
	CreatedAt             time.Time      `json:"created_at"`
}

type EmployeeContact struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	WorkEmail     *string   `json:"work_email,omitempty"`
	PersonalEmail *string   `json:"personal_email,omitempty"`
	MobilePhone   string    `json:"mobile_phone"`
	WorkPhone     *string   `json:"work_phone,omitempty"`
	Extension     *string   `json:"extension,omitempty"`
	BuildingNo    *string   `json:"building_no,omitempty"`
	StreetName    *string   `json:"street_name,omitempty"`
	District      *string   `json:"district,omitempty"`
	City          string    `json:"city"`
	PostalCode    *string   `json:"postal_code,omitempty"`
	AdditionalNo  *string   `json:"additional_no,omitempty"`
	CountryCode   string    `json:"country_code"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Contract struct {
	ID                  uuid.UUID       `json:"id"`
	LegalEntityID       uuid.UUID       `json:"legal_entity_id"`
	EmployeeID          uuid.UUID       `json:"employee_id"`
	ContractNumber      string          `json:"contract_number"`
	ContractType        ContractType    `json:"contract_type"`
	StartDate           time.Time       `json:"start_date"`
	EndDate             *time.Time      `json:"end_date,omitempty"`
	ProbationDays       int32           `json:"probation_days"`
	NoticePeriodDays    int32           `json:"notice_period_days"`
	WorkingHoursPerWeek decimal.Decimal `json:"working_hours_per_week"`
	WorkLocation        string          `json:"work_location"`
	Status              ContractStatus  `json:"status"`
	Notes               *string         `json:"notes,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type ContractVersion struct {
	ID                 uuid.UUID       `json:"id"`
	ContractID         uuid.UUID       `json:"contract_id"`
	VersionNumber      int32           `json:"version_number"`
	EffectiveFrom      time.Time       `json:"effective_from"`
	EffectiveTo        *time.Time      `json:"effective_to,omitempty"`
	BasicSalary        decimal.Decimal `json:"basic_salary"`
	HousingAllowance   decimal.Decimal `json:"housing_allowance"`
	TransportAllowance decimal.Decimal `json:"transport_allowance"`
	OtherAllowances    decimal.Decimal `json:"other_allowances"`
	GrossSalary        decimal.Decimal `json:"gross_salary"`
	ChangeReason       *string         `json:"change_reason,omitempty"`
	ApprovedByUserID   *uuid.UUID      `json:"approved_by_user_id,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

type EmployeeBankAccount struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	BankNameEn    string    `json:"bank_name_en"`
	BankNameAr    string    `json:"bank_name_ar"`
	IBAN          string    `json:"iban"`
	IBANMasked    string    `json:"iban_masked"`
	SwiftBIC      string    `json:"swift_bic"`
	IsPrimary     bool      `json:"is_primary"`
	AccountStatus string    `json:"account_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SaudizationMetric struct {
	SaudiCount              int64           `json:"saudi_count"`
	ExpatCount              int64           `json:"expat_count"`
	TotalCount              int64           `json:"total_count"`
	SaudizationRatioPercent decimal.Decimal `json:"saudization_ratio_percent"`
	NitaqatTier             string          `json:"nitaqat_tier"` // PLATINUM, HIGH_GREEN, MID_GREEN, LOW_GREEN, RED
}

// MaskIdentifier returns masked identification strings (e.g. 1088****55)
func MaskIdentifier(val string) string {
	cleaned := strings.TrimSpace(val)
	if len(cleaned) <= 4 {
		return "****"
	}
	prefix := cleaned[:4]
	suffix := cleaned[len(cleaned)-2:]
	return prefix + "****" + suffix
}

// MaskIBAN masks middle parts of IBAN
func MaskIBAN(iban string) string {
	cleaned := strings.ReplaceAll(iban, " ", "")
	if len(cleaned) <= 8 {
		return "****"
	}
	prefix := cleaned[:12]
	suffix := cleaned[len(cleaned)-4:]
	return prefix + "****" + suffix
}

// CalculateGrossSalary computes basic + housing + transport + other
func CalculateGrossSalary(basic, housing, transport, other decimal.Decimal) decimal.Decimal {
	return basic.Add(housing).Add(transport).Add(other)
}
