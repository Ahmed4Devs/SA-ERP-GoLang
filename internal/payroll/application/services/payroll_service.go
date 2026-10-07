package services

import (
	"context"
	"encoding/json"
	postingapp "erp/internal/accounting/application/posting"
	accountingdomain "erp/internal/accounting/domain"
	auditports "erp/internal/audit/ports"
	"fmt"
	"time"

	"erp/internal/payroll/domain"
	"erp/internal/payroll/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PayrollService struct {
	ledger *postingapp.SourceService
	audit  auditports.AuditService
	repo   ports.Repository
	db     *database.DB
	tx     transaction.Manager
}

func NewPayrollService(repo ports.Repository, db *database.DB, ledger *postingapp.SourceService, audit auditports.AuditService) ports.Service {
	return &PayrollService{
		repo:   repo,
		ledger: ledger, audit: audit,
		db: db,
		tx: transaction.NewManager(db.Pool),
	}
}

// 1. Components & Structures
func (s *PayrollService) CreateComponent(
	ctx context.Context,
	legalEntityID uuid.UUID,
	code, nameEn, nameAr string,
	compType domain.ComponentType,
	calcType string,
	defaultAmt decimal.Decimal,
) (*domain.SalaryComponent, error) {
	comp := domain.SalaryComponent{
		ID:                   uuid.New(),
		LegalEntityID:        legalEntityID,
		Code:                 code,
		NameEn:               nameEn,
		NameAr:               nameAr,
		ComponentType:        compType,
		IsTaxableOrGosiBasis: true,
		CalculationType:      calcType,
		DefaultAmount:        defaultAmt,
		IsActive:             true,
	}
	return s.repo.CreateComponent(ctx, comp)
}

func (s *PayrollService) ListComponents(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryComponent, error) {
	return s.repo.ListComponents(ctx, legalEntityID)
}

func (s *PayrollService) CreateStructure(
	ctx context.Context,
	legalEntityID uuid.UUID,
	code, nameEn, nameAr string,
	lines []domain.SalaryStructureLine,
) (*domain.SalaryStructure, error) {
	structObj := domain.SalaryStructure{
		ID:            uuid.New(),
		LegalEntityID: legalEntityID,
		Code:          code,
		NameEn:        nameEn,
		NameAr:        nameAr,
		CurrencyCode:  "SAR",
		IsActive:      true,
	}
	created, err := s.repo.CreateStructure(ctx, structObj)
	if err != nil {
		return nil, err
	}

	for idx, line := range lines {
		line.ID = uuid.New()
		line.SalaryStructureID = created.ID
		line.LineNumber = idx + 1
		_, err := s.repo.CreateStructureLine(ctx, line)
		if err != nil {
			return nil, fmt.Errorf("failed to create structure line: %w", err)
		}
	}

	return created, nil
}

func (s *PayrollService) ListStructures(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryStructure, error) {
	return s.repo.ListStructures(ctx, legalEntityID)
}

func (s *PayrollService) AssignSalary(ctx context.Context, assign domain.EmployeeSalaryAssignment) (*domain.EmployeeSalaryAssignment, error) {
	if assign.ID == uuid.Nil {
		assign.ID = uuid.New()
	}
	assign.GrossSalary = assign.BasicSalary.Add(assign.HousingAllowance).Add(assign.TransportAllowance).Add(assign.OtherAllowances)
	return s.repo.CreateSalaryAssignment(ctx, assign)
}

// 2. Periods & Runs
func (s *PayrollService) CreatePeriod(
	ctx context.Context,
	legalEntityID uuid.UUID,
	year, month int,
	nameEn, nameAr string,
	start, end, due time.Time,
) (*domain.PayrollPeriod, error) {
	var fpID *uuid.UUID
	var fpUUID uuid.UUID
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id FROM accounting.fiscal_periods
		WHERE legal_entity_id = $1 AND $2 BETWEEN start_date AND end_date AND status = 'OPEN'
		LIMIT 1
	`, legalEntityID, end).Scan(&fpUUID)
	if err == nil && fpUUID != uuid.Nil {
		fpID = &fpUUID
	}

	period := domain.PayrollPeriod{
		ID:             uuid.New(),
		LegalEntityID:  legalEntityID,
		FiscalPeriodID: fpID,
		PeriodNameEn:   nameEn,
		PeriodNameAr:   nameAr,
		PeriodYear:     year,
		PeriodMonth:    month,
		StartDate:      start,
		EndDate:        end,
		PaymentDueDate: due,
		IsClosed:       false,
	}
	return s.repo.CreatePeriod(ctx, period)
}

func (s *PayrollService) ListPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollPeriod, error) {
	return s.repo.ListPeriods(ctx, legalEntityID)
}

func (s *PayrollService) CreateRun(
	ctx context.Context,
	legalEntityID, periodID uuid.UUID,
	runNumber string,
	runDate, cutoffDate time.Time,
	createdBy *uuid.UUID,
) (*domain.PayrollRun, error) {
	period, err := s.repo.GetPeriodByID(ctx, periodID)
	if err != nil {
		return nil, fmt.Errorf("period not found: %w", err)
	}
	if period.LegalEntityID != legalEntityID {
		return nil, fmt.Errorf("payroll period legal entity mismatch")
	}
	if createdBy == nil || *createdBy == uuid.Nil {
		return nil, fmt.Errorf("payroll creator is required")
	}
	if period.IsClosed {
		return nil, domain.ErrPeriodClosed
	}

	run := domain.PayrollRun{
		ID:            uuid.New(),
		LegalEntityID: legalEntityID,
		PeriodID:      periodID,
		RunNumber:     runNumber,
		RunDate:       runDate,
		CutoffDate:    cutoffDate,
		Status:        domain.RunStatusDraft,
		CreatedBy:     createdBy,
	}
	return s.repo.CreateRun(ctx, run)
}

func (s *PayrollService) GetRun(ctx context.Context, runID uuid.UUID) (*domain.PayrollRun, error) {
	return s.repo.GetRunByID(ctx, runID)
}

func (s *PayrollService) ListRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollRun, error) {
	return s.repo.ListRuns(ctx, legalEntityID)
}

// 3. CalculateRun executes deterministic payroll calculation for all active employees.
func (s *PayrollService) CalculateRun(ctx context.Context, runID uuid.UUID) (*domain.PayrollRun, error) {
	var run *domain.PayrollRun
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		run, err = s.calculateRun(txCtx, runID)
		return err
	})
	return run, err
}

func (s *PayrollService) calculateRun(ctx context.Context, runID uuid.UUID) (*domain.PayrollRun, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM payroll.runs WHERE id = $1 FOR UPDATE", runID).Scan(&runID); err != nil {
		return nil, fmt.Errorf("failed to lock payroll run: %w", err)
	}
	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Financial Immutability Invariant: Posted payroll cannot be modified!
	if run.Status == domain.RunStatusPosted {
		return nil, domain.ErrPayrollAlreadyPosted
	}

	period, err := s.repo.GetPeriodByID(ctx, run.PeriodID)
	if err != nil {
		return nil, err
	}

	if run.Status == domain.RunStatusApproved {
		return nil, fmt.Errorf("approved payroll cannot be recalculated")
	}
	if period.IsClosed {
		return nil, domain.ErrPeriodClosed
	}

	// Clean any previous calculation lines for this run
	if err := s.repo.DeleteRunEmployees(ctx, runID); err != nil {
		return nil, fmt.Errorf("failed to clear previous payroll calculation: %w", err)
	}
	if _, err := s.repo.UpdateRunStatus(ctx, runID, domain.RunStatusCalculating, nil); err != nil {
		return nil, fmt.Errorf("failed to mark payroll run as calculating: %w", err)
	}

	// Fetch active employees with active contracts or salary assignments
	type EmpContractRow struct {
		EmployeeID         uuid.UUID
		ContractID         *uuid.UUID
		NationalityCode    string
		BasicSalary        decimal.Decimal
		HousingAllowance   decimal.Decimal
		TransportAllowance decimal.Decimal
		OtherAllowances    decimal.Decimal
	}

	rows, err := dbtx.Query(ctx, `
		SELECT 
			e.id as employee_id,
			c.id as contract_id,
			e.nationality_country_code,
			COALESCE(cv.basic_salary, 0),
			COALESCE(cv.housing_allowance, 0),
			COALESCE(cv.transport_allowance, 0),
			COALESCE(cv.other_allowances, 0)
		FROM hr.employees e
		LEFT JOIN hr.contracts c ON c.employee_id = e.id AND c.status = 'ACTIVE'
		LEFT JOIN LATERAL (
			SELECT basic_salary, housing_allowance, transport_allowance, other_allowances
			FROM hr.contract_versions
			WHERE contract_id = c.id
			ORDER BY version_number DESC
			LIMIT 1
		) cv ON true
		WHERE e.legal_entity_id = $1 AND e.employment_status IN ('ACTIVE', 'PROBATION')
	`, run.LegalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active employees: %w", err)
	}
	defer rows.Close()

	var empList []EmpContractRow
	for rows.Next() {
		var ec EmpContractRow
		var basic, housing, transport, other string
		if err := rows.Scan(&ec.EmployeeID, &ec.ContractID, &ec.NationalityCode, &basic, &housing, &transport, &other); err != nil {
			return nil, err
		}
		if ec.BasicSalary, err = decimal.NewFromString(basic); err != nil {
			return nil, fmt.Errorf("invalid basic salary for employee %s: %w", ec.EmployeeID, err)
		}
		if ec.HousingAllowance, err = decimal.NewFromString(housing); err != nil {
			return nil, fmt.Errorf("invalid housing allowance for employee %s: %w", ec.EmployeeID, err)
		}
		if ec.TransportAllowance, err = decimal.NewFromString(transport); err != nil {
			return nil, fmt.Errorf("invalid transport allowance for employee %s: %w", ec.EmployeeID, err)
		}
		if ec.OtherAllowances, err = decimal.NewFromString(other); err != nil {
			return nil, fmt.Errorf("invalid other allowance for employee %s: %w", ec.EmployeeID, err)
		}
		empList = append(empList, ec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading active employees: %w", err)
	}

	var totalBasic, totalAllowances, totalGross, totalDeductions, totalNet decimal.Decimal
	empCount := 0

	for _, emp := range empList {
		// Fallback to salary assignment if contract version was 0
		if emp.BasicSalary.IsZero() {
			assign, err := s.repo.GetActiveSalaryAssignment(ctx, emp.EmployeeID, period.EndDate)
			if err == nil && assign != nil {
				emp.BasicSalary = assign.BasicSalary
				emp.HousingAllowance = assign.HousingAllowance
				emp.TransportAllowance = assign.TransportAllowance
				emp.OtherAllowances = assign.OtherAllowances
			}
		}

		// Pull overtime hours approved in this period
		var otHours decimal.Decimal
		var otHoursStr *string
		if err := dbtx.QueryRow(ctx, `
			SELECT COALESCE(SUM(overtime_hours)::text, '0')
			FROM attendance.overtime_requests
			WHERE employee_id = $1 AND status = 'APPROVED'
			  AND overtime_date BETWEEN $2 AND $3
		`, emp.EmployeeID, period.StartDate, period.EndDate).Scan(&otHoursStr); err != nil {
			return nil, fmt.Errorf("failed to load overtime for employee %s: %w", emp.EmployeeID, err)
		}
		if otHoursStr != nil {
			if otHours, err = decimal.NewFromString(*otHoursStr); err != nil {
				return nil, fmt.Errorf("invalid overtime value for employee %s: %w", emp.EmployeeID, err)
			}
		}

		// Pull pending loan installments due in this period
		var loanDeduction decimal.Decimal
		installments, err := s.repo.ListPendingLoanInstallmentsByEmployee(ctx, emp.EmployeeID, period.PaymentDueDate)
		if err != nil {
			return nil, fmt.Errorf("failed to load loan installments for employee %s: %w", emp.EmployeeID, err)
		}
		for _, inst := range installments {
			loanDeduction = loanDeduction.Add(inst.Amount)
		}

		// Pull unpaid leave days in this period
		var unpaidDays int
		if err := dbtx.QueryRow(ctx, `
			SELECT COALESCE(SUM(r.total_days), 0)
			FROM leave.requests r
			JOIN leave.types t ON t.id = r.leave_type_id
			WHERE r.employee_id = $1 AND r.status = 'APPROVED' AND t.code = 'UNPAID'
			  AND r.start_date <= $3 AND r.end_date >= $2
		`, emp.EmployeeID, period.StartDate, period.EndDate).Scan(&unpaidDays); err != nil {
			return nil, fmt.Errorf("failed to load unpaid leave for employee %s: %w", emp.EmployeeID, err)
		}

		calcInput := domain.EmployeePayrollInput{
			EmployeeID:         emp.EmployeeID,
			ContractID:         emp.ContractID,
			IsSaudi:            emp.NationalityCode == "SA",
			BasicSalary:        emp.BasicSalary,
			HousingAllowance:   emp.HousingAllowance,
			TransportAllowance: emp.TransportAllowance,
			OtherAllowances:    emp.OtherAllowances,
			OvertimeHours:      otHours,
			OvertimeMultiplier: decimal.NewFromFloat(1.50),
			LoanDeduction:      loanDeduction,
			UnpaidDays:         unpaidDays,
		}

		result, err := domain.CalculateDeterministicPayroll(calcInput)
		if err != nil {
			return nil, fmt.Errorf("payroll calculation failed for employee %s: %w", emp.EmployeeID, err)
		}

		// Persist run employee
		runEmp := domain.RunEmployee{
			ID:                    uuid.New(),
			RunID:                 run.ID,
			EmployeeID:            emp.EmployeeID,
			ContractID:            emp.ContractID,
			BasicSalary:           result.BasicSalary,
			HousingAllowance:      result.HousingAllowance,
			TransportAllowance:    result.TransportAllowance,
			OtherAllowances:       result.OtherAllowances,
			OvertimeAmount:        result.OvertimeAmount,
			GrossSalary:           result.GrossSalary,
			GosiEmployeeDeduction: result.GosiEmployeeDeduction,
			LoanDeduction:         result.LoanDeduction,
			AbsenceDeduction:      result.AbsenceDeduction,
			OtherDeductions:       result.OtherDeductions,
			TotalDeductions:       result.TotalDeductions,
			NetSalary:             result.NetSalary,
			PaymentStatus:         "UNPAID",
		}

		createdRunEmp, err := s.repo.CreateRunEmployee(ctx, runEmp)
		if err != nil {
			return nil, fmt.Errorf("failed to save run employee: %w", err)
		}

		// Persist calculation lines (breakdown)
		for _, line := range result.Lines {
			line.ID = uuid.New()
			line.RunEmployeeID = createdRunEmp.ID
			if _, err := s.repo.CreateCalculationLine(ctx, line); err != nil {
				return nil, fmt.Errorf("failed to save calculation line: %w", err)
			}
		}

		totalBasic = totalBasic.Add(result.BasicSalary)
		totalAllowances = totalAllowances.Add(result.HousingAllowance).Add(result.TransportAllowance).Add(result.OtherAllowances).Add(result.OvertimeAmount)
		totalGross = totalGross.Add(result.GrossSalary)
		totalDeductions = totalDeductions.Add(result.TotalDeductions)
		totalNet = totalNet.Add(result.NetSalary)
		empCount++
	}

	return s.repo.UpdateRunTotals(ctx, runID, totalBasic, totalAllowances, totalGross, totalDeductions, totalNet, empCount)
}

// 4. ApproveRun
func (s *PayrollService) ApproveRun(ctx context.Context, runID, actor uuid.UUID) (*domain.PayrollRun, error) {
	if actor == uuid.Nil {
		return nil, fmt.Errorf("approval actor is required")
	}
	var approved *domain.PayrollRun
	err := s.ledger.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.db.DBTX(ctx).QueryRow(ctx, "SELECT id FROM payroll.runs WHERE id=$1 FOR UPDATE", runID).Scan(&runID); err != nil {
			return err
		}
		run, err := s.repo.GetRunByID(ctx, runID)
		if err != nil {
			return err
		}
		if run.CreatedBy == nil || *run.CreatedBy == actor {
			return fmt.Errorf("payroll requires an independent approver")
		}
		if run.Status != domain.RunStatusCalculated && run.Status != domain.RunStatusUnderReview {
			return domain.ErrPayrollNotCalculated
		}
		if err := s.ledger.CheckApprovalLimit(ctx, run.LegalEntityID, actor, "PAYROLL_RUN", "SAR", decimal.Zero, run.TotalGross); err != nil {
			return err
		}
		if _, err = s.db.DBTX(ctx).Exec(ctx, "UPDATE payroll.runs SET approved_by=$2,approved_at=now() WHERE id=$1", runID, actor); err != nil {
			return err
		}
		approved, err = s.repo.UpdateRunStatus(ctx, runID, domain.RunStatusApproved, nil)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(approved)
		if err != nil {
			return err
		}
		_, err = s.audit.RecordEvent(ctx, auditports.RecordEventCommand{UserID: &actor, LegalEntityID: &run.LegalEntityID, Action: "payroll.run.approve", ResourceType: "payroll.run", ResourceID: run.ID.String(), Payload: payload})
		return err
	})
	return approved, err
}

// 5. PostRun verifies invariants, creates balanced General Ledger entries, generates payslips, and posts payroll.
func (s *PayrollService) PostRun(ctx context.Context, runID uuid.UUID, postedBy *uuid.UUID) (*domain.PayrollRun, error) {
	if postedBy == nil || *postedBy == uuid.Nil {
		return nil, fmt.Errorf("posting actor is required")
	}
	var run *domain.PayrollRun
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		run, err = s.postRun(txCtx, runID, postedBy)
		return err
	})
	return run, err
}

func (s *PayrollService) postRun(ctx context.Context, runID uuid.UUID, postedBy *uuid.UUID) (*domain.PayrollRun, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM payroll.runs WHERE id = $1 FOR UPDATE", runID).Scan(&runID); err != nil {
		return nil, fmt.Errorf("failed to lock payroll run: %w", err)
	}
	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Invariant: duplicate post impossible
	if run.Status == domain.RunStatusPosted {
		return nil, domain.ErrDuplicatePayrollPost
	}
	if run.CreatedBy == nil || *run.CreatedBy == *postedBy {
		return nil, fmt.Errorf("payroll requires an independent posting actor")
	}
	if run.Status != domain.RunStatusApproved || run.ApprovedBy == nil || run.ApprovedAt == nil {
		return nil, domain.ErrPayrollNotApproved
	}

	period, err := s.repo.GetPeriodByID(ctx, run.PeriodID)
	if err != nil {
		return nil, err
	}

	runEmployees, err := s.repo.ListRunEmployees(ctx, runID)
	if err != nil {
		return nil, err
	}

	if period.IsClosed {
		return nil, domain.ErrPeriodClosed
	}
	var totalGosi, totalLoan, totalRecovery, totalNet, totalGross decimal.Decimal
	for _, re := range runEmployees {
		totalGosi = totalGosi.Add(re.GosiEmployeeDeduction)
		totalLoan = totalLoan.Add(re.LoanDeduction)
		totalRecovery = totalRecovery.Add(re.AbsenceDeduction).Add(re.OtherDeductions)
		totalNet = totalNet.Add(re.NetSalary)
		totalGross = totalGross.Add(re.GrossSalary)
	}
	if !totalGross.Equal(run.TotalGross) || !totalNet.Equal(run.TotalNet) || !totalGross.Equal(totalNet.Add(totalGosi).Add(totalLoan).Add(totalRecovery)) {
		return nil, fmt.Errorf("payroll employee totals differ from run totals")
	}
	lines := []accountingdomain.PostingLine{}
	for _, part := range []struct {
		code          string
		debit, credit decimal.Decimal
	}{
		{"521000", totalGross, decimal.Zero}, {"212000", decimal.Zero, totalNet},
		{"215000", decimal.Zero, totalGosi}, {"115000", decimal.Zero, totalLoan}, {"521000", decimal.Zero, totalRecovery},
	} {
		if part.debit.IsZero() && part.credit.IsZero() {
			continue
		}
		account, err := s.ledger.AccountForDate(ctx, run.LegalEntityID, nil, "PAYROLL", part.code, part.code, run.RunDate)
		if err != nil {
			return nil, err
		}
		lines = append(lines, accountingdomain.PostingLine{AccountID: account, TransactionDebit: part.debit, TransactionCredit: part.credit})
	}
	journal, err := s.ledger.Post(ctx, "GENERAL", accountingdomain.PostingRequest{LegalEntityID: run.LegalEntityID, SourceModule: "PAYROLL", SourceType: "PAYROLL_RUN", SourceID: run.ID, AccountingDate: run.RunDate, CurrencyCode: "SAR", Reference: run.RunNumber, ActorID: *postedBy, Lines: lines})
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.LinkRunJournalEntry(ctx, runID, journal); err != nil {
		return nil, err
	}

	// 2. Generate immutable Payslips for each run employee
	for idx, re := range runEmployees {
		calcLines, err := s.repo.ListCalculationLines(ctx, re.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load payroll calculation lines: %w", err)
		}
		snapshot := map[string]any{
			"basic":       re.BasicSalary.String(),
			"housing":     re.HousingAllowance.String(),
			"transport":   re.TransportAllowance.String(),
			"other":       re.OtherAllowances.String(),
			"overtime":    re.OvertimeAmount.String(),
			"gross":       re.GrossSalary.String(),
			"gosi":        re.GosiEmployeeDeduction.String(),
			"loan":        re.LoanDeduction.String(),
			"absence":     re.AbsenceDeduction.String(),
			"deductions":  re.TotalDeductions.String(),
			"net":         re.NetSalary.String(),
			"lines_count": len(calcLines),
		}

		payslipNum := fmt.Sprintf("PS-%s-%04d", run.RunNumber, idx+1)
		slip := domain.Payslip{
			ID:                  uuid.New(),
			LegalEntityID:       run.LegalEntityID,
			RunEmployeeID:       re.ID,
			PayslipNumber:       payslipNum,
			EmployeeID:          re.EmployeeID,
			PeriodID:            run.PeriodID,
			IssueDate:           run.RunDate,
			GrossSalary:         re.GrossSalary,
			TotalDeductions:     re.TotalDeductions,
			NetSalary:           re.NetSalary,
			CalculationSnapshot: snapshot,
		}
		if _, err := s.repo.CreatePayslip(ctx, slip); err != nil {
			return nil, fmt.Errorf("failed to create payslip: %w", err)
		}

		// 3. Mark any deducted loan installments as DEDUCTED
		if re.LoanDeduction.GreaterThan(decimal.Zero) {
			installments, err := s.repo.ListPendingLoanInstallmentsByEmployee(ctx, re.EmployeeID, period.PaymentDueDate)
			if err != nil {
				return nil, fmt.Errorf("failed to load loan installments: %w", err)
			}
			for _, inst := range installments {
				if _, err := s.repo.UpdateLoanInstallmentStatus(ctx, inst.ID, "DEDUCTED", &run.ID); err != nil {
					return nil, fmt.Errorf("failed to mark loan installment deducted: %w", err)
				}
				if _, err := s.repo.UpdateLoanRemainingBalance(ctx, inst.LoanID, inst.Amount); err != nil {
					return nil, fmt.Errorf("failed to update employee loan balance: %w", err)
				}
			}
		}
	}

	// Update run state to POSTED
	return s.repo.UpdateRunStatus(ctx, runID, domain.RunStatusPosted, postedBy)
}

// 6. GenerateWps produces official Saudi SIF format text.
func (s *PayrollService) GenerateWps(ctx context.Context, runID uuid.UUID, employerMolID, bankCode string) (string, string, error) {
	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return "", "", err
	}

	runEmployees, err := s.repo.ListRunEmployees(ctx, runID)
	if err != nil {
		return "", "", err
	}

	var wpsList []domain.WpsEmployeeDetail
	for _, re := range runEmployees {
		var idNum, nameEn, nameAr, iban, bCode string
		_ = s.db.Pool.QueryRow(ctx, `
			SELECT 
				COALESCE(ei.identifier_value, '1000000000'),
				e.first_name_en || ' ' || e.last_name_en,
				e.first_name_ar || ' ' || e.last_name_ar,
				COALESCE(ba.iban, 'SA0000000000000000000000'),
				COALESCE(ba.swift_bic, $2)
			FROM hr.employees e
			LEFT JOIN hr.employee_identifiers ei ON ei.employee_id = e.id AND ei.is_primary = true
			LEFT JOIN hr.bank_accounts ba ON ba.employee_id = e.id AND ba.is_primary = true
			WHERE e.id = $1
		`, re.EmployeeID, bankCode).Scan(&idNum, &nameEn, &nameAr, &iban, &bCode)

		otherEarnings := re.TransportAllowance.Add(re.OtherAllowances).Add(re.OvertimeAmount)

		wpsList = append(wpsList, domain.WpsEmployeeDetail{
			NationalIDOrIqama: idNum,
			EmployeeName:      nameEn,
			EmployeeNumber:    re.EmployeeID.String()[:8],
			BankCode:          bCode,
			IBAN:              iban,
			BasicSalary:       re.BasicSalary,
			HousingAllowance:  re.HousingAllowance,
			OtherEarnings:     otherEarnings,
			Deductions:        re.TotalDeductions,
			NetSalary:         re.NetSalary,
		})
	}

	sifContent := domain.GenerateWpsSifContent(employerMolID, bankCode, run.RunDate, wpsList)
	fileName := fmt.Sprintf("SIF_%s_%s_%s.txt", employerMolID, run.RunDate.Format("20060102"), run.RunNumber)

	_, _ = s.repo.UpdateRunWpsStatus(ctx, runID, fileName)

	return sifContent, fileName, nil
}

// 7. Payslips
func (s *PayrollService) GetPayslipsByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.Payslip, error) {
	return s.repo.ListPayslipsByPeriod(ctx, periodID)
}

func (s *PayrollService) GetPayslipByID(ctx context.Context, payslipID uuid.UUID) (*domain.Payslip, error) {
	return s.repo.GetPayslipByID(ctx, payslipID)
}

// 8. Loans
func (s *PayrollService) CreateLoan(
	ctx context.Context,
	legalEntityID, employeeID uuid.UUID,
	loanNumber string,
	principal, monthlyInstallment decimal.Decimal,
	disbursementDate time.Time,
	reason *string,
) (*domain.EmployeeLoan, error) {
	var loan *domain.EmployeeLoan
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		loan, err = s.createLoan(txCtx, legalEntityID, employeeID, loanNumber, principal, monthlyInstallment, disbursementDate, reason)
		return err
	})
	return loan, err
}

func (s *PayrollService) createLoan(
	ctx context.Context,
	legalEntityID, employeeID uuid.UUID,
	loanNumber string,
	principal, monthlyInstallment decimal.Decimal,
	disbursementDate time.Time,
	reason *string,
) (*domain.EmployeeLoan, error) {
	if principal.LessThanOrEqual(decimal.Zero) || monthlyInstallment.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("principal and monthly installment must be greater than zero")
	}
	loan := domain.EmployeeLoan{
		ID:                 uuid.New(),
		LegalEntityID:      legalEntityID,
		EmployeeID:         employeeID,
		LoanNumber:         loanNumber,
		PrincipalAmount:    principal,
		MonthlyInstallment: monthlyInstallment,
		RemainingBalance:   principal,
		DisbursementDate:   disbursementDate,
		Status:             "ACTIVE",
		Reason:             reason,
	}
	created, err := s.repo.CreateLoan(ctx, loan)
	if err != nil {
		return nil, err
	}

	// Schedule installments (e.g. principal / monthlyInstallment)
	numInstallments := int(principal.DivRound(monthlyInstallment, 0).IntPart())
	if numInstallments <= 0 {
		numInstallments = 1
	}

	for i := 1; i <= numInstallments; i++ {
		dueDate := disbursementDate.AddDate(0, i, 0)
		inst := domain.LoanInstallment{
			ID:                uuid.New(),
			LoanID:            created.ID,
			InstallmentNumber: i,
			DueDate:           dueDate,
			Amount:            monthlyInstallment,
			Status:            "PENDING",
		}
		if _, err := s.repo.CreateLoanInstallment(ctx, inst); err != nil {
			return nil, fmt.Errorf("failed to create loan installment %d: %w", i, err)
		}
	}

	return created, nil
}

func (s *PayrollService) ListLoans(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EmployeeLoan, error) {
	return s.repo.ListLoans(ctx, legalEntityID)
}

// 9. End of Service Calculator
func (s *PayrollService) CalculateEndOfService(
	ctx context.Context,
	legalEntityID, employeeID uuid.UUID,
	terminationDate time.Time,
	reason domain.TerminationReason,
) (*domain.EndOfServiceCalculation, error) {
	var hireDate time.Time
	var lastGross decimal.Decimal
	var grossStr string

	err := s.db.Pool.QueryRow(ctx, `
		SELECT 
			e.hire_date,
			COALESCE(cv.gross_salary::text, '0')
		FROM hr.employees e
		LEFT JOIN hr.contracts c ON c.employee_id = e.id AND c.status = 'ACTIVE'
		LEFT JOIN LATERAL (
			SELECT gross_salary
			FROM hr.contract_versions
			WHERE contract_id = c.id
			ORDER BY version_number DESC
			LIMIT 1
		) cv ON true
		WHERE e.id = $1
	`, employeeID).Scan(&hireDate, &grossStr)
	if err != nil {
		return nil, fmt.Errorf("employee or contract not found: %w", err)
	}

	lastGross, _ = decimal.NewFromString(grossStr)

	result, err := domain.CalculateEndOfServiceAward(hireDate, terminationDate, lastGross, reason)
	if err != nil {
		return nil, err
	}

	calcNum := identifiers.NewDocumentNumber("EOS", time.Now())
	eos := domain.EndOfServiceCalculation{
		ID:                    uuid.New(),
		LegalEntityID:         legalEntityID,
		EmployeeID:            employeeID,
		CalculationNumber:     calcNum,
		HireDate:              hireDate,
		TerminationDate:       terminationDate,
		TerminationReason:     reason,
		ServiceYears:          result.ServiceYears,
		LastGrossSalary:       lastGross,
		AwardAmount:           result.BaseAwardAmount,
		StatutoryRateFraction: result.StatutoryFraction,
		FinalSettlementAmount: result.FinalSettlementAmount,
		CalculationBreakdown: map[string]any{
			"tenure_breakdown": result.TenureBreakdown,
			"statutory_law":    "Saudi Labor Law Articles 84 and 85",
		},
		Status: "CALCULATED",
	}

	return s.repo.CreateEndOfService(ctx, eos)
}

func (s *PayrollService) ListEndOfService(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EndOfServiceCalculation, error) {
	return s.repo.ListEndOfService(ctx, legalEntityID)
}

// 10. Seed
func (s *PayrollService) SeedDefaultData(ctx context.Context, legalEntityID uuid.UUID) error {
	return s.repo.SeedDefaultPayrollData(ctx, legalEntityID)
}
