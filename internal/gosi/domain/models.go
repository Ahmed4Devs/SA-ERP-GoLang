package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ContributionRule represents a type of GOSI contribution (e.g., Annuity, OHRSI, SANED).
type ContributionRule struct {
	ID            uuid.UUID
	LegalEntityID uuid.UUID
	Code          string
	NameEn        string
	NameAr        string
	DescriptionEn string
	DescriptionAr string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RuleVersion represents an effective-dated version of a contribution rule with specific rates.
// Regulatory rates MUST come from versioned database rules — never hard-coded.
type RuleVersion struct {
	ID                uuid.UUID
	RuleID            uuid.UUID
	LegalEntityID     uuid.UUID
	VersionNumber     int
	EffectiveFrom     time.Time
	EffectiveTo       *time.Time
	NationalityScope  string // SAUDI, NON_SAUDI, ALL
	EmployeeRate      decimal.Decimal
	EmployerRate      decimal.Decimal
	MinimumBasis      decimal.Decimal
	MaximumBasis      decimal.Decimal
	CalculationConfig map[string]any
	LegalReference    string
	Notes             string
	IsActive          bool
	CreatedAt         time.Time

	// Joined fields from contribution_rules
	RuleCode   string
	RuleNameEn string
	RuleNameAr string
}

// InsuranceProfile represents an employee's GOSI registration.
type InsuranceProfile struct {
	ID                        uuid.UUID
	LegalEntityID             uuid.UUID
	EmployeeID                uuid.UUID
	GosiNumber                string
	RegistrationDate          *time.Time
	DeregistrationDate        *time.Time
	NationalityScope          string // SAUDI or NON_SAUDI
	IsActive                  bool
	ContributionBasisOverride *decimal.Decimal
	Notes                     string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time

	// Joined fields from hr.employees
	EmployeeNumber  string
	FirstNameEn     string
	LastNameEn      string
	FirstNameAr     string
	LastNameAr      string
	NationalityCode string
}

// ContributionCalculation represents a monthly contribution calculation for an employee.
type ContributionCalculation struct {
	ID                  uuid.UUID
	LegalEntityID       uuid.UUID
	EmployeeID          uuid.UUID
	ProfileID           uuid.UUID
	CalculationYear     int
	CalculationMonth    int
	ContributionBasis   decimal.Decimal
	TotalEmployeeAmount decimal.Decimal
	TotalEmployerAmount decimal.Decimal
	TotalAmount         decimal.Decimal
	CalculationSnapshot map[string]any
	ExportBatchID       *uuid.UUID
	Status              string
	CreatedAt           time.Time

	// Joined fields
	EmployeeNumber string
	FirstNameEn    string
	LastNameEn     string
	FirstNameAr    string
	LastNameAr     string
}

// CalculationLine represents a single rule-application line within a contribution calculation.
type CalculationLine struct {
	ID             uuid.UUID
	CalculationID  uuid.UUID
	RuleVersionID  uuid.UUID
	RuleCode       string
	RuleNameEn     string
	RuleNameAr     string
	BasisAmount    decimal.Decimal
	EmployeeRate   decimal.Decimal
	EmployerRate   decimal.Decimal
	EmployeeAmount decimal.Decimal
	EmployerAmount decimal.Decimal
	FormulaNote    string
	CreatedAt      time.Time
}

// ExportBatch represents a monthly GOSI export batch.
type ExportBatch struct {
	ID                  uuid.UUID
	LegalEntityID       uuid.UUID
	BatchNumber         string
	ExportYear          int
	ExportMonth         int
	EmployeeCount       int
	TotalEmployeeAmount decimal.Decimal
	TotalEmployerAmount decimal.Decimal
	TotalAmount         decimal.Decimal
	FileContent         *string
	FileFormat          string
	Status              string
	SubmittedAt         *time.Time
	ConfirmedAt         *time.Time
	CreatedBy           *uuid.UUID
	CreatedAt           time.Time
}

// IntegrationAttempt tracks a submission attempt to the GOSI portal.
type IntegrationAttempt struct {
	ID              uuid.UUID
	LegalEntityID   uuid.UUID
	BatchID         uuid.UUID
	AttemptNumber   int
	AdapterType     string // MANUAL, FILE_EXPORT, API_INTEGRATION
	RequestPayload  map[string]any
	ResponsePayload map[string]any
	Status          string // PENDING, IN_PROGRESS, SUCCESS, FAILED, TIMEOUT
	ErrorMessage    string
	StartedAt       time.Time
	CompletedAt     *time.Time
}

// ContributionInput contains all parameters required for a single employee contribution calculation.
// Matches the GOSIService interface from Section 26 of the project specification.
type ContributionInput struct {
	LegalEntityID     uuid.UUID
	EmployeeID        uuid.UUID
	PayrollDate       time.Time
	NationalityCode   string
	ContributionBasis decimal.Decimal
}

// ContributionResult contains the calculated contribution amounts.
// Matches the GOSIService interface from Section 26 of the project specification.
type ContributionResult struct {
	RuleVersionID       uuid.UUID
	Basis               decimal.Decimal
	EmployeeRate        decimal.Decimal
	EmployerRate        decimal.Decimal
	EmployeeAmount      decimal.Decimal
	EmployerAmount      decimal.Decimal
	CalculationSnapshot map[string]any
}

// SingleRuleResult represents the calculation result for a single rule version.
type SingleRuleResult struct {
	RuleVersion    RuleVersion
	BasisApplied   decimal.Decimal
	EmployeeAmount decimal.Decimal
	EmployerAmount decimal.Decimal
	FormulaNote    string
}

// CalculateContribution applies a set of active rule versions to a contribution input.
// It determines which rules apply based on the employee's nationality scope.
// Rates are read from the database — never hard-coded.
func CalculateContribution(input ContributionInput, activeRules []RuleVersion) ([]SingleRuleResult, error) {
	isSaudi := input.NationalityCode == "SA"
	var results []SingleRuleResult

	for _, rv := range activeRules {
		// Check nationality scope applicability
		switch rv.NationalityScope {
		case "SAUDI":
			if !isSaudi {
				continue
			}
		case "NON_SAUDI":
			if isSaudi {
				continue
			}
		case "ALL":
			// Applies to everyone
		default:
			continue
		}

		// Apply basis capping
		basis := input.ContributionBasis
		if basis.LessThan(rv.MinimumBasis) {
			basis = rv.MinimumBasis
		}
		if basis.GreaterThan(rv.MaximumBasis) {
			basis = rv.MaximumBasis
		}

		// Calculate amounts using rates from DB
		employeeRate := rv.EmployeeRate.Div(decimal.NewFromInt(100))
		employerRate := rv.EmployerRate.Div(decimal.NewFromInt(100))

		employeeAmount := basis.Mul(employeeRate).Round(2)
		employerAmount := basis.Mul(employerRate).Round(2)

		note := rv.RuleCode + ": "
		if isSaudi {
			note += "Saudi national"
		} else {
			note += "Non-Saudi"
		}
		note += " | Basis=" + basis.String() +
			" | EmpRate=" + rv.EmployeeRate.String() + "%" +
			" | ErRate=" + rv.EmployerRate.String() + "%" +
			" | Ref: " + rv.LegalReference

		results = append(results, SingleRuleResult{
			RuleVersion:    rv,
			BasisApplied:   basis,
			EmployeeAmount: employeeAmount,
			EmployerAmount: employerAmount,
			FormulaNote:    note,
		})
	}

	return results, nil
}

// DashboardSummary provides the GOSI dashboard KPI values.
type DashboardSummary struct {
	TotalEmployeeContributions decimal.Decimal
	TotalEmployerContributions decimal.Decimal
	TotalContributions         decimal.Decimal
	SubscriberCount            int
	ComplianceRate             string
}
