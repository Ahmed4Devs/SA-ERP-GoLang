-- name: GetBudgetByID :one
SELECT * FROM budgeting.budgets
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetBudgetByCode :one
SELECT * FROM budgeting.budgets
WHERE code = $1 AND legal_entity_id = $2;

-- name: ListBudgets :many
SELECT * FROM budgeting.budgets
WHERE legal_entity_id = $1
ORDER BY created_at DESC;

-- name: CreateBudget :one
INSERT INTO budgeting.budgets (
    id, legal_entity_id, code, name_en, name_ar, description_en, description_ar,
    fiscal_year_id, currency_code, budget_type, status,
    total_budgeted_amount, total_actual_amount, total_committed_amount,
    created_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14,
    $15, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateBudgetStatus :one
UPDATE budgeting.budgets
SET status = $3,
    approved_by_user_id = $4,
    approved_at = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: UpdateBudgetTotals :one
UPDATE budgeting.budgets
SET total_budgeted_amount = $3,
    total_actual_amount = $4,
    total_committed_amount = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: DeleteBudget :exec
DELETE FROM budgeting.budgets
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetBudgetVersions :many
SELECT * FROM budgeting.budget_versions
WHERE budget_id = $1
ORDER BY version_number ASC;

-- name: GetBudgetVersionByID :one
SELECT * FROM budgeting.budget_versions
WHERE id = $1;

-- name: GetActiveBudgetVersion :one
SELECT * FROM budgeting.budget_versions
WHERE budget_id = $1 AND status = 'ACTIVE'
LIMIT 1;

-- name: CreateBudgetVersion :one
INSERT INTO budgeting.budget_versions (
    id, budget_id, version_number, version_name, status, notes,
    created_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateBudgetVersionStatus :one
UPDATE budgeting.budget_versions
SET status = $2,
    approved_by_user_id = $3,
    approved_at = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListBudgetLines :many
SELECT 
    l.id,
    l.budget_version_id,
    l.account_id,
    l.department_id,
    l.cost_center_id,
    l.project_code,
    l.fiscal_period_id,
    l.amount,
    l.actual_amount,
    l.committed_amount,
    l.line_description_en,
    l.line_description_ar,
    l.created_at,
    l.updated_at,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    d.name_en AS department_name_en,
    d.name_ar AS department_name_ar,
    c.name_en AS cost_center_name_en,
    c.name_ar AS cost_center_name_ar,
    p.period_number,
    p.period_name_en,
    p.period_name_ar
FROM budgeting.budget_lines l
JOIN accounting.accounts a ON a.id = l.account_id
LEFT JOIN organization.departments d ON d.id = l.department_id
LEFT JOIN organization.cost_centers c ON c.id = l.cost_center_id
LEFT JOIN accounting.fiscal_periods p ON p.id = l.fiscal_period_id
WHERE l.budget_version_id = $1
ORDER BY a.code ASC, l.created_at ASC;

-- name: GetBudgetLineByID :one
SELECT * FROM budgeting.budget_lines
WHERE id = $1;

-- name: CreateBudgetLine :one
INSERT INTO budgeting.budget_lines (
    id, budget_version_id, account_id, department_id, cost_center_id,
    project_code, fiscal_period_id, amount, actual_amount, committed_amount,
    line_description_en, line_description_ar, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateBudgetLine :one
UPDATE budgeting.budget_lines
SET amount = $2,
    actual_amount = $3,
    committed_amount = $4,
    line_description_en = $5,
    line_description_ar = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteBudgetLine :exec
DELETE FROM budgeting.budget_lines
WHERE id = $1;

-- name: SumBudgetLinesByVersion :one
SELECT 
    COALESCE(SUM(amount), 0)::NUMERIC(18,4) AS total_budgeted,
    COALESCE(SUM(actual_amount), 0)::NUMERIC(18,4) AS total_actual,
    COALESCE(SUM(committed_amount), 0)::NUMERIC(18,4) AS total_committed
FROM budgeting.budget_lines
WHERE budget_version_id = $1;

-- name: CreateBudgetTransfer :one
INSERT INTO budgeting.budget_transfers (
    id, legal_entity_id, transfer_number, source_line_id, target_line_id,
    amount, reason, status, transferred_by_user_id, created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, CURRENT_TIMESTAMP
) RETURNING *;

-- name: ListBudgetTransfers :many
SELECT 
    t.id,
    t.legal_entity_id,
    t.transfer_number,
    t.source_line_id,
    t.target_line_id,
    t.amount,
    t.reason,
    t.status,
    t.transferred_by_user_id,
    t.created_at,
    sa.code AS source_account_code,
    sa.name_en AS source_account_name_en,
    sa.name_ar AS source_account_name_ar,
    ta.code AS target_account_code,
    ta.name_en AS target_account_name_en,
    ta.name_ar AS target_account_name_ar
FROM budgeting.budget_transfers t
JOIN budgeting.budget_lines sl ON sl.id = t.source_line_id
JOIN accounting.accounts sa ON sa.id = sl.account_id
JOIN budgeting.budget_lines tl ON tl.id = t.target_line_id
JOIN accounting.accounts ta ON ta.id = tl.account_id
WHERE t.legal_entity_id = $1
ORDER BY t.created_at DESC;
