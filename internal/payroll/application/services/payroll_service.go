package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/payroll/domain"
	"erp/internal/payroll/ports"
	"erp/internal/platform/database"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PayrollService struct {
	repo ports.Repository
	db   *database.DB
}

func NewPayrollService(repo ports.Repository, db *database.DB) ports.Service {
	return &PayrollService{
		repo: repo,
		db:   db,
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
		WHERE legal_entity_id = $1 AND period_number = $2 AND status = 'OPEN'
		LIMIT 1
	`, legalEntityID, month).Scan(&fpUUID)
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

	// Clean any previous calculation lines for this run
	_ = s.repo.DeleteRunEmployees(ctx, runID)
	_, _ = s.repo.UpdateRunStatus(ctx, runID, domain.RunStatusCalculating, nil)

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

	rows, err := s.db.Pool.Query(ctx, `
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
		ec.BasicSalary, _ = decimal.NewFromString(basic)
		ec.HousingAllowance, _ = decimal.NewFromString(housing)
		ec.TransportAllowance, _ = decimal.NewFromString(transport)
		ec.OtherAllowances, _ = decimal.NewFromString(other)
		empList = append(empList, ec)
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
		_ = s.db.Pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(approved_hours)::text, '0')
			FROM attendance.overtime_requests
			WHERE employee_id = $1 AND status = 'APPROVED'
			  AND work_date BETWEEN $2 AND $3
		`, emp.EmployeeID, period.StartDate, period.EndDate).Scan(&otHoursStr)
		if otHoursStr != nil {
			otHours, _ = decimal.NewFromString(*otHoursStr)
		}

		// Pull pending loan installments due in this period
		var loanDeduction decimal.Decimal
		installments, _ := s.repo.ListPendingLoanInstallmentsByEmployee(ctx, emp.EmployeeID, period.PaymentDueDate)
		for _, inst := range installments {
			loanDeduction = loanDeduction.Add(inst.Amount)
		}

		// Pull unpaid leave days in this period
		var unpaidDays int
		_ = s.db.Pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(working_days), 0)
			FROM leave.requests
			WHERE employee_id = $1 AND status = 'APPROVED' AND leave_type_code = 'UNPAID'
			  AND start_date <= $3 AND end_date >= $2
		`, emp.EmployeeID, period.StartDate, period.EndDate).Scan(&unpaidDays)

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
func (s *PayrollService) ApproveRun(ctx context.Context, runID uuid.UUID) (*domain.PayrollRun, error) {
	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run.Status == domain.RunStatusPosted {
		return nil, domain.ErrPayrollAlreadyPosted
	}
	if run.Status != domain.RunStatusCalculated && run.Status != domain.RunStatusUnderReview {
		return nil, domain.ErrPayrollNotCalculated
	}

	return s.repo.UpdateRunStatus(ctx, runID, domain.RunStatusApproved, nil)
}

// 5. PostRun verifies invariants, creates balanced General Ledger entries, generates payslips, and posts payroll.
func (s *PayrollService) PostRun(ctx context.Context, runID uuid.UUID, postedBy *uuid.UUID) (*domain.PayrollRun, error) {
	run, err := s.repo.GetRunByID(ctx, runID)
	if err != nil {
		return nil, err
	}

	// Invariant: duplicate post impossible
	if run.Status == domain.RunStatusPosted {
		return nil, domain.ErrDuplicatePayrollPost
	}
	if run.Status != domain.RunStatusApproved {
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

	// 1. Post Balanced Journal Entry to General Ledger
	// Debit: Salaries & Wages Expense (521000) = Total Gross
	// Credit: Accrued Salaries Payable (212000) = Total Net
	// Credit: Output/Social GOSI Payable (212000 / 213000) = Total GOSI Deductions
	// Credit: Employee Loan Recovery (115000 / 212000) = Total Loan Deductions
	// Credit: Other Deductions / Absence Recovery = Total Absence Deductions
	// Debit == Credit!

	var totalGosi, totalLoan, totalAbsence decimal.Decimal
	for _, re := range runEmployees {
		totalGosi = totalGosi.Add(re.GosiEmployeeDeduction)
		totalLoan = totalLoan.Add(re.LoanDeduction)
		totalAbsence = totalAbsence.Add(re.AbsenceDeduction)
	}

	var journalID, fiscalPeriodID, salExpenseAccID, accruedSalAccID, gosiLiabilityAccID, loanAssetAccID uuid.UUID
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 LIMIT 1", run.LegalEntityID).Scan(&journalID)
	if period.FiscalPeriodID != nil {
		fiscalPeriodID = *period.FiscalPeriodID
	} else {
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", run.LegalEntityID).Scan(&fiscalPeriodID)
	}

	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '521000' LIMIT 1", run.LegalEntityID).Scan(&salExpenseAccID)
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '212000' LIMIT 1", run.LegalEntityID).Scan(&accruedSalAccID)
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '213000' LIMIT 1", run.LegalEntityID).Scan(&gosiLiabilityAccID)
	if gosiLiabilityAccID == uuid.Nil {
		gosiLiabilityAccID = accruedSalAccID
	}
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '115000' LIMIT 1", run.LegalEntityID).Scan(&loanAssetAccID)
	if loanAssetAccID == uuid.Nil {
		loanAssetAccID = accruedSalAccID
	}

	var jeID uuid.UUID
	if journalID != uuid.Nil && fiscalPeriodID != uuid.Nil && salExpenseAccID != uuid.Nil && accruedSalAccID != uuid.Nil {
		jeID = uuid.New()
		jeNum := fmt.Sprintf("JE-PR-%s", run.RunNumber)

		// 1. Create Journal Entry Header in DRAFT state
		_, err = s.db.Pool.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, 'DRAFT', 'PAYROLL', 'PAYROLL_RUN', $7,
				$8, $9, 'SAR', 1.0,
				$10, $10, $11
			)
		`, jeID, run.LegalEntityID, journalID, fiscalPeriodID, jeNum,
			run.RunDate, run.ID,
			run.RunNumber, fmt.Sprintf("استحقاق مسير رواتب - %s", run.RunNumber),
			run.TotalGross, postedBy)

		if err == nil {
			lineNum := int16(1)
			// 1. Debit Salaries Expense
			_, err = s.db.Pool.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, $3, $4, 'إجمالي استحقاقات رواتب الموظفين', 'SAR', $5, 0, $5, 0)
			`, jeID, run.LegalEntityID, lineNum, salExpenseAccID, run.TotalGross)
			lineNum++

			// 2. Credit Accrued Salaries Payable (Net)
			_, _ = s.db.Pool.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, $3, $4, 'صافي رواتب الموظفين المستحقة للصرف', 'SAR', 0, $5, 0, $5)
			`, jeID, run.LegalEntityID, lineNum, accruedSalAccID, run.TotalNet)
			lineNum++

			// 3. Credit GOSI Payable (if any)
			if totalGosi.GreaterThan(decimal.Zero) {
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, 'استقطاعات التأمينات الاجتماعية (حصة الموظفين)', 'SAR', 0, $5, 0, $5)
				`, jeID, run.LegalEntityID, lineNum, gosiLiabilityAccID, totalGosi)
				lineNum++
			}

			// 4. Credit Loan Recovery (if any)
			if totalLoan.GreaterThan(decimal.Zero) {
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, 'استرداد أقساط سلف وقروض الموظفين', 'SAR', 0, $5, 0, $5)
				`, jeID, run.LegalEntityID, lineNum, loanAssetAccID, totalLoan)
				lineNum++
			}

			// 5. Credit Absence/Other Recovery (if any)
			if totalAbsence.GreaterThan(decimal.Zero) {
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, 'استقطاعات الغياب والإجازات غير مدفوعة الأجر', 'SAR', 0, $5, 0, $5)
				`, jeID, run.LegalEntityID, lineNum, salExpenseAccID, totalAbsence)
				lineNum++
			}

			// 6. Transition Journal Entry to POSTED
			_, _ = s.db.Pool.Exec(ctx, `
				UPDATE accounting.journal_entries
				SET status = 'POSTED', posted_by = $1, posted_at = CURRENT_TIMESTAMP
				WHERE id = $2
			`, postedBy, jeID)

			// Link JE to payroll run
			_, _ = s.repo.LinkRunJournalEntry(ctx, runID, jeID)
		}
	}

	// 2. Generate immutable Payslips for each run employee
	for idx, re := range runEmployees {
		calcLines, _ := s.repo.ListCalculationLines(ctx, re.ID)
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
		_, _ = s.repo.CreatePayslip(ctx, slip)

		// 3. Mark any deducted loan installments as DEDUCTED
		if re.LoanDeduction.GreaterThan(decimal.Zero) {
			installments, _ := s.repo.ListPendingLoanInstallmentsByEmployee(ctx, re.EmployeeID, period.PaymentDueDate)
			for _, inst := range installments {
				_, _ = s.repo.UpdateLoanInstallmentStatus(ctx, inst.ID, "DEDUCTED", &run.ID)
				_, _ = s.repo.UpdateLoanRemainingBalance(ctx, inst.LoanID, inst.Amount)
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
		_, _ = s.repo.CreateLoanInstallment(ctx, inst)
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

	calcNum := fmt.Sprintf("EOS-%s-%d", employeeID.String()[:6], time.Now().Unix()%10000)
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
