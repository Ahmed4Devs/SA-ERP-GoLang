-- name: GetEmployeeLeaveBalance :one
SELECT * FROM leave.employee_balances
WHERE employee_id = $1 AND leave_type_id = $2 AND fiscal_year = $3;

-- name: UpsertEmployeeLeaveBalance :one
INSERT INTO leave.employee_balances (
    id,
    employee_id,
    leave_type_id,
    fiscal_year,
    allocated_days,
    accrued_days,
    taken_days,
    pending_days,
    carryover_days,
    closing_balance
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (employee_id, leave_type_id, fiscal_year)
DO UPDATE SET
    allocated_days = EXCLUDED.allocated_days,
    accrued_days = EXCLUDED.accrued_days,
    taken_days = EXCLUDED.taken_days,
    pending_days = EXCLUDED.pending_days,
    carryover_days = EXCLUDED.carryover_days,
    closing_balance = EXCLUDED.closing_balance,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: ListEmployeeLeaveBalances :many
SELECT b.*, t.code AS leave_type_code, t.name_en AS leave_type_name_en, t.name_ar AS leave_type_name_ar
FROM leave.employee_balances b
JOIN leave.types t ON b.leave_type_id = t.id
WHERE b.employee_id = $1 AND b.fiscal_year = $2
ORDER BY t.code ASC;

-- name: CreateLeaveTransaction :one
INSERT INTO leave.transactions (
    id,
    employee_id,
    leave_type_id,
    fiscal_year,
    transaction_type,
    days,
    reference_id,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListLeaveTransactions :many
SELECT * FROM leave.transactions
WHERE employee_id = $1 AND leave_type_id = $2 AND fiscal_year = $3
ORDER BY created_at DESC;
