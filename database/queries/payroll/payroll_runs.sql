-- name: CreatePayrollPeriod :one
INSERT INTO payroll.periods (
    id, legal_entity_id, fiscal_period_id, period_name_en, period_name_ar,
    period_year, period_month, start_date, end_date, payment_due_date, is_closed
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (legal_entity_id, period_year, period_month)
DO UPDATE SET period_name_en = EXCLUDED.period_name_en
RETURNING *;

-- name: GetPayrollPeriodByID :one
SELECT * FROM payroll.periods
WHERE id = $1 LIMIT 1;

-- name: ListPayrollPeriods :many
SELECT * FROM payroll.periods
WHERE legal_entity_id = $1
ORDER BY period_year DESC, period_month DESC;

-- name: CreatePayrollRun :one
INSERT INTO payroll.runs (
    id, legal_entity_id, period_id, run_number, run_date, cutoff_date,
    status, total_basic, total_allowances, total_gross, total_deductions,
    total_net, employee_count, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetPayrollRunByID :one
SELECT * FROM payroll.runs
WHERE id = $1 LIMIT 1;

-- name: ListPayrollRuns :many
SELECT r.*, p.period_name_en, p.period_name_ar, p.period_year, p.period_month
FROM payroll.runs r
JOIN payroll.periods p ON p.id = r.period_id
WHERE r.legal_entity_id = $1
ORDER BY r.created_at DESC;

-- name: UpdatePayrollRunStatus :one
UPDATE payroll.runs
SET status = sqlc.arg('status')::VARCHAR,
    posted_at = CASE WHEN sqlc.arg('status')::VARCHAR = 'POSTED' THEN CURRENT_TIMESTAMP ELSE posted_at END,
    posted_by = CASE WHEN sqlc.arg('status')::VARCHAR = 'POSTED' THEN sqlc.arg('posted_by')::UUID ELSE posted_by END
WHERE id = $1
RETURNING *;

-- name: UpdatePayrollRunTotals :one
UPDATE payroll.runs
SET total_basic = $2,
    total_allowances = $3,
    total_gross = $4,
    total_deductions = $5,
    total_net = $6,
    employee_count = $7,
    status = 'CALCULATED'
WHERE id = $1
RETURNING *;

-- name: LinkPayrollRunJournalEntry :one
UPDATE payroll.runs
SET journal_entry_id = $2
WHERE id = $1
RETURNING *;

-- name: UpdatePayrollRunWpsStatus :one
UPDATE payroll.runs
SET wps_file_generated = true,
    wps_file_name = $2
WHERE id = $1
RETURNING *;

-- name: CreatePayrollRunEmployee :one
INSERT INTO payroll.run_employees (
    id, run_id, employee_id, contract_id, basic_salary, housing_allowance,
    transport_allowance, other_allowances, overtime_amount, gross_salary,
    gosi_employee_deduction, loan_deduction, absence_deduction, other_deductions,
    total_deductions, net_salary, payment_status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: ListPayrollRunEmployees :many
SELECT re.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM payroll.run_employees re
JOIN hr.employees e ON e.id = re.employee_id
WHERE re.run_id = $1
ORDER BY e.employee_number ASC;

-- name: GetPayrollRunEmployeeByID :one
SELECT * FROM payroll.run_employees
WHERE id = $1 LIMIT 1;

-- name: DeletePayrollRunEmployees :exec
DELETE FROM payroll.run_employees
WHERE run_id = $1;

-- name: CreateCalculationLine :one
INSERT INTO payroll.calculation_lines (
    id, run_employee_id, component_id, line_type, code, name_en, name_ar,
    rate, basis_amount, amount, formula_note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: ListCalculationLinesByRunEmployee :many
SELECT * FROM payroll.calculation_lines
WHERE run_employee_id = $1
ORDER BY line_type ASC, code ASC;
