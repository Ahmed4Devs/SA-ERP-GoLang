package ports

import (
	"context"
	"time"

	"erp/internal/payroll/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Repository interface {
	// Components & Structures
	CreateComponent(ctx context.Context, comp domain.SalaryComponent) (*domain.SalaryComponent, error)
	GetComponentByID(ctx context.Context, id uuid.UUID) (*domain.SalaryComponent, error)
	ListComponents(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryComponent, error)
	CreateStructure(ctx context.Context, structObj domain.SalaryStructure) (*domain.SalaryStructure, error)
	GetStructureByID(ctx context.Context, id uuid.UUID) (*domain.SalaryStructure, error)
	ListStructures(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryStructure, error)
	CreateStructureLine(ctx context.Context, line domain.SalaryStructureLine) (*domain.SalaryStructureLine, error)
	ListStructureLines(ctx context.Context, structureID uuid.UUID) ([]domain.SalaryStructureLine, error)
	CreateSalaryAssignment(ctx context.Context, assign domain.EmployeeSalaryAssignment) (*domain.EmployeeSalaryAssignment, error)
	GetActiveSalaryAssignment(ctx context.Context, employeeID uuid.UUID, asOfDate time.Time) (*domain.EmployeeSalaryAssignment, error)

	// Periods & Runs
	CreatePeriod(ctx context.Context, period domain.PayrollPeriod) (*domain.PayrollPeriod, error)
	GetPeriodByID(ctx context.Context, id uuid.UUID) (*domain.PayrollPeriod, error)
	ListPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollPeriod, error)
	CreateRun(ctx context.Context, run domain.PayrollRun) (*domain.PayrollRun, error)
	GetRunByID(ctx context.Context, id uuid.UUID) (*domain.PayrollRun, error)
	ListRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollRun, error)
	UpdateRunStatus(ctx context.Context, id uuid.UUID, status domain.RunStatus, postedBy *uuid.UUID) (*domain.PayrollRun, error)
	UpdateRunTotals(ctx context.Context, id uuid.UUID, basic, allow, gross, ded, net decimal.Decimal, count int) (*domain.PayrollRun, error)
	LinkRunJournalEntry(ctx context.Context, id uuid.UUID, journalEntryID uuid.UUID) (*domain.PayrollRun, error)
	UpdateRunWpsStatus(ctx context.Context, id uuid.UUID, fileName string) (*domain.PayrollRun, error)

	// Run Employees & Lines
	CreateRunEmployee(ctx context.Context, emp domain.RunEmployee) (*domain.RunEmployee, error)
	ListRunEmployees(ctx context.Context, runID uuid.UUID) ([]domain.RunEmployee, error)
	GetRunEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.RunEmployee, error)
	DeleteRunEmployees(ctx context.Context, runID uuid.UUID) error
	CreateCalculationLine(ctx context.Context, line domain.CalculationLine) (*domain.CalculationLine, error)
	ListCalculationLines(ctx context.Context, runEmployeeID uuid.UUID) ([]domain.CalculationLine, error)

	// Payslips
	CreatePayslip(ctx context.Context, slip domain.Payslip) (*domain.Payslip, error)
	GetPayslipByID(ctx context.Context, id uuid.UUID) (*domain.Payslip, error)
	ListPayslipsByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.Payslip, error)
	ListPayslipsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]domain.Payslip, error)

	// Loans & Installments
	CreateLoan(ctx context.Context, loan domain.EmployeeLoan) (*domain.EmployeeLoan, error)
	GetLoanByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeLoan, error)
	ListLoans(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EmployeeLoan, error)
	CreateLoanInstallment(ctx context.Context, inst domain.LoanInstallment) (*domain.LoanInstallment, error)
	ListPendingLoanInstallmentsByEmployee(ctx context.Context, employeeID uuid.UUID, dueDate time.Time) ([]domain.LoanInstallment, error)
	UpdateLoanInstallmentStatus(ctx context.Context, id uuid.UUID, status string, runID *uuid.UUID) (*domain.LoanInstallment, error)
	UpdateLoanRemainingBalance(ctx context.Context, id uuid.UUID, deductionAmount decimal.Decimal) (*domain.EmployeeLoan, error)

	// End of Service
	CreateEndOfService(ctx context.Context, eos domain.EndOfServiceCalculation) (*domain.EndOfServiceCalculation, error)
	GetEndOfServiceByID(ctx context.Context, id uuid.UUID) (*domain.EndOfServiceCalculation, error)
	ListEndOfService(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EndOfServiceCalculation, error)

	// Seed
	SeedDefaultPayrollData(ctx context.Context, legalEntityID uuid.UUID) error
}

type Service interface {
	// Components & Structures
	CreateComponent(ctx context.Context, legalEntityID uuid.UUID, code, nameEn, nameAr string, compType domain.ComponentType, calcType string, defaultAmt decimal.Decimal) (*domain.SalaryComponent, error)
	ListComponents(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryComponent, error)
	CreateStructure(ctx context.Context, legalEntityID uuid.UUID, code, nameEn, nameAr string, lines []domain.SalaryStructureLine) (*domain.SalaryStructure, error)
	ListStructures(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryStructure, error)
	AssignSalary(ctx context.Context, assign domain.EmployeeSalaryAssignment) (*domain.EmployeeSalaryAssignment, error)

	// Periods & Runs
	CreatePeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int, nameEn, nameAr string, start, end, due time.Time) (*domain.PayrollPeriod, error)
	ListPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollPeriod, error)
	CreateRun(ctx context.Context, legalEntityID, periodID uuid.UUID, runNumber string, runDate, cutoffDate time.Time, createdBy *uuid.UUID) (*domain.PayrollRun, error)
	GetRun(ctx context.Context, runID uuid.UUID) (*domain.PayrollRun, error)
	ListRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollRun, error)
	CalculateRun(ctx context.Context, runID uuid.UUID) (*domain.PayrollRun, error)
	ApproveRun(ctx context.Context, runID, actor uuid.UUID) (*domain.PayrollRun, error)
	PostRun(ctx context.Context, runID uuid.UUID, postedBy *uuid.UUID) (*domain.PayrollRun, error)
	GenerateWps(ctx context.Context, runID uuid.UUID, employerMolID, bankCode string) (string, string, error)

	// Payslips
	GetPayslipsByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.Payslip, error)
	GetPayslipByID(ctx context.Context, payslipID uuid.UUID) (*domain.Payslip, error)

	// Loans
	CreateLoan(ctx context.Context, legalEntityID, employeeID uuid.UUID, loanNumber string, principal, monthlyInstallment decimal.Decimal, disbursementDate time.Time, reason *string) (*domain.EmployeeLoan, error)
	ListLoans(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EmployeeLoan, error)

	// End of Service
	CalculateEndOfService(ctx context.Context, legalEntityID, employeeID uuid.UUID, terminationDate time.Time, reason domain.TerminationReason) (*domain.EndOfServiceCalculation, error)
	ListEndOfService(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EndOfServiceCalculation, error)

	// Seed
	SeedDefaultData(ctx context.Context, legalEntityID uuid.UUID) error
}
