package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrPayrollAlreadyPosted = errors.New("posted payroll is financially immutable and cannot be modified")
	ErrInvalidPayrollState  = errors.New("invalid payroll run status transition")
	ErrDuplicatePayrollPost = errors.New("payroll run has already been posted to the general ledger")
	ErrPeriodClosed         = errors.New("payroll period is closed")
	ErrPayrollNotCalculated = errors.New("payroll run must be in CALCULATED state to be approved")
	ErrPayrollNotApproved   = errors.New("payroll run must be in APPROVED state to be posted")
	ErrEmployeeNotFound     = errors.New("employee not found or has no active contract")
	ErrNegativeNetSalary    = errors.New("total deductions cannot exceed gross salary")
)

type ComponentType string

const (
	ComponentTypeEarning   ComponentType = "EARNING"
	ComponentTypeDeduction ComponentType = "DEDUCTION"
)

type RunStatus string

const (
	RunStatusDraft       RunStatus = "DRAFT"
	RunStatusCalculating RunStatus = "CALCULATING"
	RunStatusCalculated  RunStatus = "CALCULATED"
	RunStatusUnderReview RunStatus = "UNDER_REVIEW"
	RunStatusApproved    RunStatus = "APPROVED"
	RunStatusPosted      RunStatus = "POSTED"
	RunStatusPaid        RunStatus = "PAID"
	RunStatusCancelled   RunStatus = "CANCELLED"
	RunStatusReversed    RunStatus = "REVERSED"
)

type TerminationReason string

const (
	TerminationResignation     TerminationReason = "RESIGNATION"
	TerminationByEmployer      TerminationReason = "TERMINATION_BY_EMPLOYER"
	TerminationContractExpiry  TerminationReason = "CONTRACT_EXPIRY"
	TerminationMutualAgreement TerminationReason = "MUTUAL_AGREEMENT"
	TerminationForceMajeure    TerminationReason = "FORCE_MAJEURE"
)

// SalaryComponent represents an individual earning or deduction line item.
type SalaryComponent struct {
	ID                   uuid.UUID
	LegalEntityID        uuid.UUID
	Code                 string
	NameEn               string
	NameAr               string
	ComponentType        ComponentType
	IsTaxableOrGosiBasis bool
	CalculationType      string // FIXED, FORMULA, PERCENTAGE_OF_BASIC
	FormulaExpression    *string
	DefaultAmount        decimal.Decimal
	IsActive             bool
	CreatedAt            time.Time
}

// SalaryStructure represents a group of salary rules and components.
type SalaryStructure struct {
	ID            uuid.UUID
	LegalEntityID uuid.UUID
	Code          string
	NameEn        string
	NameAr        string
	CurrencyCode  string
	IsActive      bool
	CreatedAt     time.Time
}

// SalaryStructureLine represents a line in a salary structure.
type SalaryStructureLine struct {
	ID                uuid.UUID
	SalaryStructureID uuid.UUID
	ComponentID       uuid.UUID
	LineNumber        int
	RatePercentage    decimal.Decimal
	FixedAmount       decimal.Decimal
	CreatedAt         time.Time
}

// EmployeeSalaryAssignment represents an employee's salary package.
type EmployeeSalaryAssignment struct {
	ID                 uuid.UUID
	LegalEntityID      uuid.UUID
	EmployeeID         uuid.UUID
	SalaryStructureID  uuid.UUID
	EffectiveFrom      time.Time
	EffectiveTo        *time.Time
	BasicSalary        decimal.Decimal
	HousingAllowance   decimal.Decimal
	TransportAllowance decimal.Decimal
	OtherAllowances    decimal.Decimal
	GrossSalary        decimal.Decimal
	CreatedAt          time.Time
}

// PayrollPeriod represents a monthly payroll cycle.
type PayrollPeriod struct {
	ID             uuid.UUID
	LegalEntityID  uuid.UUID
	FiscalPeriodID *uuid.UUID
	PeriodNameEn   string
	PeriodNameAr   string
	PeriodYear     int
	PeriodMonth    int
	StartDate      time.Time
	EndDate        time.Time
	PaymentDueDate time.Time
	IsClosed       bool
	CreatedAt      time.Time
}

// PayrollRun represents a batch payroll calculation for a period.
type PayrollRun struct {
	ApprovedBy       *uuid.UUID
	ApprovedAt       *time.Time
	ID               uuid.UUID
	LegalEntityID    uuid.UUID
	PeriodID         uuid.UUID
	RunNumber        string
	RunDate          time.Time
	CutoffDate       time.Time
	Status           RunStatus
	TotalBasic       decimal.Decimal
	TotalAllowances  decimal.Decimal
	TotalGross       decimal.Decimal
	TotalDeductions  decimal.Decimal
	TotalNet         decimal.Decimal
	EmployeeCount    int
	JournalEntryID   *uuid.UUID
	WpsFileGenerated bool
	WpsFileName      *string
	PostedAt         *time.Time
	PostedBy         *uuid.UUID
	CreatedBy        *uuid.UUID
	CreatedAt        time.Time
}

// RunEmployee represents a single employee's computed payroll for the run.
type RunEmployee struct {
	ID                    uuid.UUID
	RunID                 uuid.UUID
	EmployeeID            uuid.UUID
	ContractID            *uuid.UUID
	BasicSalary           decimal.Decimal
	HousingAllowance      decimal.Decimal
	TransportAllowance    decimal.Decimal
	OtherAllowances       decimal.Decimal
	OvertimeAmount        decimal.Decimal
	GrossSalary           decimal.Decimal
	GosiEmployeeDeduction decimal.Decimal
	LoanDeduction         decimal.Decimal
	AbsenceDeduction      decimal.Decimal
	OtherDeductions       decimal.Decimal
	TotalDeductions       decimal.Decimal
	NetSalary             decimal.Decimal
	PaymentStatus         string
	CreatedAt             time.Time
}

// CalculationLine provides full explainability for each earning or deduction item.
type CalculationLine struct {
	ID            uuid.UUID
	RunEmployeeID uuid.UUID
	ComponentID   *uuid.UUID
	LineType      ComponentType
	Code          string
	NameEn        string
	NameAr        string
	Rate          decimal.Decimal
	BasisAmount   decimal.Decimal
	Amount        decimal.Decimal
	FormulaNote   string
	CreatedAt     time.Time
}

// Payslip represents an employee's formal compensation document.
type Payslip struct {
	ID                  uuid.UUID
	LegalEntityID       uuid.UUID
	RunEmployeeID       uuid.UUID
	PayslipNumber       string
	EmployeeID          uuid.UUID
	PeriodID            uuid.UUID
	IssueDate           time.Time
	GrossSalary         decimal.Decimal
	TotalDeductions     decimal.Decimal
	NetSalary           decimal.Decimal
	CalculationSnapshot map[string]any
	IsViewedByEmployee  bool
	CreatedAt           time.Time
}

// EmployeeLoan tracks company loans and monthly installments.
type EmployeeLoan struct {
	ID                 uuid.UUID
	LegalEntityID      uuid.UUID
	EmployeeID         uuid.UUID
	LoanNumber         string
	PrincipalAmount    decimal.Decimal
	MonthlyInstallment decimal.Decimal
	RemainingBalance   decimal.Decimal
	DisbursementDate   time.Time
	Status             string // ACTIVE, PAID_OFF
	Reason             *string
	CreatedAt          time.Time
}

// LoanInstallment tracks each monthly deduction for a loan.
type LoanInstallment struct {
	ID                uuid.UUID
	LoanID            uuid.UUID
	PayrollRunID      *uuid.UUID
	InstallmentNumber int
	DueDate           time.Time
	Amount            decimal.Decimal
	Status            string // PENDING, DEDUCTED, WAIVED
	DeductedAt        *time.Time
	CreatedAt         time.Time
}

// EndOfServiceCalculation represents a statutory severance computation per Articles 84 & 85.
type EndOfServiceCalculation struct {
	ID                    uuid.UUID
	LegalEntityID         uuid.UUID
	EmployeeID            uuid.UUID
	CalculationNumber     string
	HireDate              time.Time
	TerminationDate       time.Time
	TerminationReason     TerminationReason
	ServiceYears          decimal.Decimal
	LastGrossSalary       decimal.Decimal
	AwardAmount           decimal.Decimal
	StatutoryRateFraction decimal.Decimal
	FinalSettlementAmount decimal.Decimal
	CalculationBreakdown  map[string]any
	Status                string
	CreatedAt             time.Time
}

// EmployeePayrollInput contains all parameters required for deterministic payroll calculation.
type EmployeePayrollInput struct {
	EmployeeID         uuid.UUID
	ContractID         *uuid.UUID
	IsSaudi            bool
	BasicSalary        decimal.Decimal
	HousingAllowance   decimal.Decimal
	TransportAllowance decimal.Decimal
	OtherAllowances    decimal.Decimal
	OvertimeHours      decimal.Decimal
	OvertimeMultiplier decimal.Decimal // Default 1.50 per Article 107
	LoanDeduction      decimal.Decimal
	UnpaidDays         int
}

// EmployeePayrollResult contains the deterministic outputs and explainable lines.
type EmployeePayrollResult struct {
	BasicSalary           decimal.Decimal
	HousingAllowance      decimal.Decimal
	TransportAllowance    decimal.Decimal
	OtherAllowances       decimal.Decimal
	OvertimeAmount        decimal.Decimal
	GrossSalary           decimal.Decimal
	GosiEmployeeDeduction decimal.Decimal
	LoanDeduction         decimal.Decimal
	AbsenceDeduction      decimal.Decimal
	OtherDeductions       decimal.Decimal
	TotalDeductions       decimal.Decimal
	NetSalary             decimal.Decimal
	Lines                 []CalculationLine
}

// CalculateDeterministicPayroll computes the exact, reproducible salary breakdown for an employee.
// Invariant: Given the exact same EmployeePayrollInput, this function will ALWAYS return identical results.
func CalculateDeterministicPayroll(input EmployeePayrollInput) (EmployeePayrollResult, error) {
	// 1. Basic & Allowances
	basic := input.BasicSalary.Round(4)
	housing := input.HousingAllowance.Round(4)
	transport := input.TransportAllowance.Round(4)
	other := input.OtherAllowances.Round(4)

	// 2. Overtime Calculation (Saudi Labor Law Art. 107 - 1.50x standard multiplier)
	multiplier := input.OvertimeMultiplier
	if multiplier.IsZero() {
		multiplier = decimal.NewFromFloat(1.50)
	}

	otAmount := decimal.Zero
	var otHourlyRate decimal.Decimal
	if input.OvertimeHours.GreaterThan(decimal.Zero) {
		// Standard monthly hours base in Saudi Labor Law: 30 days * 8 hours = 240 hours
		standardHours := decimal.NewFromInt(240)
		otHourlyRate = basic.Add(housing).DivRound(standardHours, 4)
		otAmount = input.OvertimeHours.Mul(otHourlyRate).Mul(multiplier).Round(4)
	}

	gross := basic.Add(housing).Add(transport).Add(other).Add(otAmount).Round(4)

	// 3. GOSI Employee Deduction
	// For Saudi employees: 9.75% of (Basic + Housing), subject to maximum contribution ceiling (45,000 SAR)
	gosiDeduction := decimal.Zero
	if input.IsSaudi {
		gosiBasis := basic.Add(housing)
		gosiCap := decimal.NewFromInt(45000)
		if gosiBasis.GreaterThan(gosiCap) {
			gosiBasis = gosiCap
		}
		// 9.75% = 9% Annuities (معاشات) + 0.75% Unemployment Insurance (ساند)
		gosiRate := decimal.NewFromFloat(0.0975)
		gosiDeduction = gosiBasis.Mul(gosiRate).Round(4)
	}

	// 4. Absence / Unpaid Leave Deduction: (Gross / 30) * unpaid_days
	absenceDeduction := decimal.Zero
	if input.UnpaidDays > 0 {
		dailyRate := gross.DivRound(decimal.NewFromInt(30), 4)
		absenceDeduction = dailyRate.Mul(decimal.NewFromInt(int64(input.UnpaidDays))).Round(4)
	}

	// 5. Loan Deduction
	loanDeduction := input.LoanDeduction.Round(4)

	// 6. Total Deductions & Net Salary
	totalDeductions := gosiDeduction.Add(absenceDeduction).Add(loanDeduction).Round(4)
	netSalary := gross.Sub(totalDeductions).Round(4)

	if netSalary.IsNegative() {
		return EmployeePayrollResult{}, ErrNegativeNetSalary
	}

	// 7. Calculation Breakdown Lines (Explainability)
	lines := []CalculationLine{
		{
			LineType:    ComponentTypeEarning,
			Code:        "BASIC",
			NameEn:      "Basic Salary",
			NameAr:      "الراتب الأساسي",
			Rate:        decimal.NewFromInt(1),
			BasisAmount: basic,
			Amount:      basic,
			FormulaNote: "Contractual basic salary",
		},
		{
			LineType:    ComponentTypeEarning,
			Code:        "HOUSING",
			NameEn:      "Housing Allowance",
			NameAr:      "بدل سكن",
			Rate:        decimal.NewFromInt(1),
			BasisAmount: housing,
			Amount:      housing,
			FormulaNote: "Standard housing allowance",
		},
		{
			LineType:    ComponentTypeEarning,
			Code:        "TRANSPORT",
			NameEn:      "Transportation Allowance",
			NameAr:      "بدل نقل",
			Rate:        decimal.NewFromInt(1),
			BasisAmount: transport,
			Amount:      transport,
			FormulaNote: "Standard transportation allowance",
		},
	}

	if other.GreaterThan(decimal.Zero) {
		lines = append(lines, CalculationLine{
			LineType:    ComponentTypeEarning,
			Code:        "OTHER_ALLOWANCE",
			NameEn:      "Other Allowances",
			NameAr:      "بدلات أخرى",
			Rate:        decimal.NewFromInt(1),
			BasisAmount: other,
			Amount:      other,
			FormulaNote: "Contractual special allowances",
		})
	}

	if otAmount.GreaterThan(decimal.Zero) {
		lines = append(lines, CalculationLine{
			LineType:    ComponentTypeEarning,
			Code:        "OVERTIME",
			NameEn:      "Overtime Wages",
			NameAr:      "أجر العمل الإضافي (المادة 107)",
			Rate:        multiplier,
			BasisAmount: otHourlyRate.Mul(input.OvertimeHours).Round(4),
			Amount:      otAmount,
			FormulaNote: fmt.Sprintf("%s hrs @ %s SAR/hr * %s multiplier", input.OvertimeHours.String(), otHourlyRate.String(), multiplier.String()),
		})
	}

	if gosiDeduction.GreaterThan(decimal.Zero) {
		lines = append(lines, CalculationLine{
			LineType:    ComponentTypeDeduction,
			Code:        "GOSI_EMPLOYEE",
			NameEn:      "GOSI Employee Share (9.75%)",
			NameAr:      "حصة الموظف في التأمينات الاجتماعية (9.75%)",
			Rate:        decimal.NewFromFloat(0.0975),
			BasisAmount: basic.Add(housing),
			Amount:      gosiDeduction,
			FormulaNote: "9.75% of (Basic + Housing) capped at 45,000 SAR per GOSI rules",
		})
	}

	if absenceDeduction.GreaterThan(decimal.Zero) {
		lines = append(lines, CalculationLine{
			LineType:    ComponentTypeDeduction,
			Code:        "ABSENCE_DEDUCTION",
			NameEn:      "Unpaid Leave / Absence Deduction",
			NameAr:      "استقطاع الغياب / الإجازة غير مدفوعة الأجر",
			Rate:        decimal.NewFromInt(int64(input.UnpaidDays)),
			BasisAmount: gross.DivRound(decimal.NewFromInt(30), 4),
			Amount:      absenceDeduction,
			FormulaNote: fmt.Sprintf("%d days * (Gross / 30 days)", input.UnpaidDays),
		})
	}

	if loanDeduction.GreaterThan(decimal.Zero) {
		lines = append(lines, CalculationLine{
			LineType:    ComponentTypeDeduction,
			Code:        "LOAN_DEDUCTION",
			NameEn:      "Company Loan Installment",
			NameAr:      "استقطاع قسط سلفة الشركة",
			Rate:        decimal.NewFromInt(1),
			BasisAmount: loanDeduction,
			Amount:      loanDeduction,
			FormulaNote: "Active monthly loan deduction",
		})
	}

	return EmployeePayrollResult{
		BasicSalary:           basic,
		HousingAllowance:      housing,
		TransportAllowance:    transport,
		OtherAllowances:       other,
		OvertimeAmount:        otAmount,
		GrossSalary:           gross,
		GosiEmployeeDeduction: gosiDeduction,
		LoanDeduction:         loanDeduction,
		AbsenceDeduction:      absenceDeduction,
		OtherDeductions:       decimal.Zero,
		TotalDeductions:       totalDeductions,
		NetSalary:             netSalary,
		Lines:                 lines,
	}, nil
}

// EndOfServiceResult represents the outcome of Saudi Labor Law Articles 84 and 85 calculations.
type EndOfServiceResult struct {
	ServiceYears          decimal.Decimal
	BaseAwardAmount       decimal.Decimal
	StatutoryFraction     decimal.Decimal
	FinalSettlementAmount decimal.Decimal
	TenureBreakdown       string
}

// CalculateEndOfServiceAward calculates severance pay under Saudi Labor Law:
// Article 84:
// - Half a month's wage for each of the first five years.
// - A full month's wage for each of the following years.
// Article 85 (Resignation):
// - Less than 2 continuous years: 0% (nothing).
// - Between 2 and 5 continuous years: 1/3 (one-third).
// - Between 5 and 10 continuous years: 2/3 (two-thirds).
// - 10 years or more (or termination by employer): Full award (100%).
func CalculateEndOfServiceAward(hireDate time.Time, terminationDate time.Time, lastGrossSalary decimal.Decimal, reason TerminationReason) (EndOfServiceResult, error) {
	if terminationDate.Before(hireDate) {
		return EndOfServiceResult{}, errors.New("termination date cannot be before hire date")
	}

	days := terminationDate.Sub(hireDate).Hours() / 24.0
	yearsFloat := days / 365.25
	years := decimal.NewFromFloat(yearsFloat).Round(2)

	// Article 84 base computation:
	// First 5 years: 0.5 month per year
	// Beyond 5 years: 1.0 month per year
	yearsVal, _ := years.Float64()
	first5Years := math.Min(yearsVal, 5.0)
	remainingYears := math.Max(0.0, yearsVal-5.0)

	halfMonthSalary := lastGrossSalary.Mul(decimal.NewFromFloat(0.5))
	fullMonthSalary := lastGrossSalary

	awardPart1 := halfMonthSalary.Mul(decimal.NewFromFloat(first5Years))
	awardPart2 := fullMonthSalary.Mul(decimal.NewFromFloat(remainingYears))
	baseAward := awardPart1.Add(awardPart2).Round(4)

	// Article 85 Resignation discount factor
	fraction := decimal.NewFromInt(1) // 100% default for employer termination, expiry, force majeure

	if reason == TerminationResignation {
		if yearsVal < 2.0 {
			fraction = decimal.Zero
		} else if yearsVal < 5.0 {
			fraction = decimal.NewFromFloat(1.0 / 3.0).Round(4)
		} else if yearsVal < 10.0 {
			fraction = decimal.NewFromFloat(2.0 / 3.0).Round(4)
		} else {
			fraction = decimal.NewFromInt(1)
		}
	}

	finalAmount := baseAward.Mul(fraction).Round(4)

	breakdown := fmt.Sprintf("Service: %.2f yrs. First 5 yrs: %.2f yrs @ 50%% wage, Remaining: %.2f yrs @ 100%% wage. Resignation factor: %s.",
		yearsVal, first5Years, remainingYears, fraction.String())

	return EndOfServiceResult{
		ServiceYears:          years,
		BaseAwardAmount:       baseAward,
		StatutoryFraction:     fraction,
		FinalSettlementAmount: finalAmount,
		TenureBreakdown:       breakdown,
	}, nil
}

// WpsEmployeeDetail contains employee information formatted for WPS/SIF files.
type WpsEmployeeDetail struct {
	NationalIDOrIqama string
	EmployeeName      string
	EmployeeNumber    string
	BankCode          string
	IBAN              string
	BasicSalary       decimal.Decimal
	HousingAllowance  decimal.Decimal
	OtherEarnings     decimal.Decimal
	Deductions        decimal.Decimal
	NetSalary         decimal.Decimal
}

// GenerateWpsSifContent creates the official Saudi Standard Information File (SIF) text.
// SIF format standard:
// SCR Header: SCR,<BankCode>,<EmployerMolID>,<ValueDate YYYY-MM-DD>,<TotalAmount>,<RecordCount>,SAR
// EDR Detail: EDR,<ID>,<Name>,<IBAN>,<BankCode>,<Basic>,<Housing>,<Other>,<Deductions>,<Net>,<Ref>
func GenerateWpsSifContent(employerMolID string, bankCode string, valueDate time.Time, employees []WpsEmployeeDetail) string {
	var totalNet decimal.Decimal
	for _, emp := range employees {
		totalNet = totalNet.Add(emp.NetSalary)
	}

	var sb strings.Builder
	dateStr := valueDate.Format("2006-01-02")

	// 1. Header Record (SCR)
	sb.WriteString(fmt.Sprintf("SCR,%s,%s,%s,%s,%d,SAR\n",
		bankCode,
		employerMolID,
		dateStr,
		totalNet.StringFixed(2),
		len(employees),
	))

	// 2. Detail Records (EDR)
	for _, emp := range employees {
		sb.WriteString(fmt.Sprintf("EDR,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			emp.NationalIDOrIqama,
			emp.EmployeeName,
			emp.IBAN,
			emp.BankCode,
			emp.BasicSalary.StringFixed(2),
			emp.HousingAllowance.StringFixed(2),
			emp.OtherEarnings.StringFixed(2),
			emp.Deductions.StringFixed(2),
			emp.NetSalary.StringFixed(2),
			emp.EmployeeNumber,
		))
	}

	return sb.String()
}
