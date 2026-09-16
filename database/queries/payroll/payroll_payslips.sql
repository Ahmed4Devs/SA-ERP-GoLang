-- name: CreatePayslip :one
INSERT INTO payroll.payslips (
    id, legal_entity_id, run_employee_id, payslip_number, employee_id,
    period_id, issue_date, gross_salary, total_deductions, net_salary, calculation_snapshot
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetPayslipByID :one
SELECT ps.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar,
       p.period_name_en, p.period_name_ar, p.start_date, p.end_date
FROM payroll.payslips ps
JOIN hr.employees e ON e.id = ps.employee_id
JOIN payroll.periods p ON p.id = ps.period_id
WHERE ps.id = $1 LIMIT 1;

-- name: ListPayslipsByPeriod :many
SELECT ps.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM payroll.payslips ps
JOIN hr.employees e ON e.id = ps.employee_id
WHERE ps.period_id = $1
ORDER BY e.employee_number ASC;

-- name: ListPayslipsByEmployee :many
SELECT ps.*, p.period_name_en, p.period_name_ar, p.period_year, p.period_month
FROM payroll.payslips ps
JOIN payroll.periods p ON p.id = ps.period_id
WHERE ps.employee_id = $1
ORDER BY ps.issue_date DESC;

-- name: CreateEmployeeLoan :one
INSERT INTO payroll.loans (
    id, legal_entity_id, employee_id, loan_number, principal_amount,
    monthly_installment, remaining_balance, disbursement_date, status, reason
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetEmployeeLoanByID :one
SELECT * FROM payroll.loans
WHERE id = $1 LIMIT 1;

-- name: ListEmployeeLoans :many
SELECT l.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM payroll.loans l
JOIN hr.employees e ON e.id = l.employee_id
WHERE l.legal_entity_id = $1
ORDER BY l.created_at DESC;

-- name: CreateLoanInstallment :one
INSERT INTO payroll.loan_installments (
    id, loan_id, payroll_run_id, installment_number, due_date, amount, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListPendingLoanInstallmentsByEmployee :many
SELECT li.*, l.employee_id
FROM payroll.loan_installments li
JOIN payroll.loans l ON l.id = li.loan_id
WHERE l.employee_id = $1 AND li.status = 'PENDING' AND li.due_date <= $2
ORDER BY li.installment_number ASC;

-- name: UpdateLoanInstallmentStatus :one
UPDATE payroll.loan_installments
SET status = $2,
    payroll_run_id = $3,
    deducted_at = CASE WHEN $2 = 'DEDUCTED' THEN CURRENT_TIMESTAMP ELSE deducted_at END
WHERE id = $1
RETURNING *;

-- name: UpdateLoanRemainingBalance :one
UPDATE payroll.loans
SET remaining_balance = remaining_balance - $2,
    status = CASE WHEN (remaining_balance - $2) <= 0 THEN 'PAID_OFF' ELSE status END
WHERE id = $1
RETURNING *;

-- name: CreateEndOfServiceCalculation :one
INSERT INTO payroll.end_of_service_calculations (
    id, legal_entity_id, employee_id, calculation_number, hire_date,
    termination_date, termination_reason, service_years, last_gross_salary,
    award_amount, statutory_rate_fraction, final_settlement_amount, calculation_breakdown, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetEndOfServiceCalculationByID :one
SELECT eos.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM payroll.end_of_service_calculations eos
JOIN hr.employees e ON e.id = eos.employee_id
WHERE eos.id = $1 LIMIT 1;

-- name: ListEndOfServiceCalculations :many
SELECT eos.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM payroll.end_of_service_calculations eos
JOIN hr.employees e ON e.id = eos.employee_id
WHERE eos.legal_entity_id = $1
ORDER BY eos.created_at DESC;
