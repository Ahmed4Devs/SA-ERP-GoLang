-- name: ListChartOfAccountsByLegalEntity :many
SELECT id, legal_entity_id, code, name_en, name_ar, description, is_active, created_at, updated_at, version
FROM accounting.chart_of_accounts
WHERE legal_entity_id = $1
ORDER BY code;

-- name: GetChartOfAccountsByID :one
SELECT id, legal_entity_id, code, name_en, name_ar, description, is_active, created_at, updated_at, version
FROM accounting.chart_of_accounts
WHERE id = $1;

-- name: CreateAccount :one
INSERT INTO accounting.accounts (
    legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account, is_active, created_at, updated_at, version;

-- name: UpdateAccount :one
UPDATE accounting.accounts
SET name_en = $2,
    name_ar = $3,
    allow_posting = $4,
    is_control_account = $5,
    is_active = $6,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING id, legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account, is_active, created_at, updated_at, version;

-- name: GetAccountByID :one
SELECT id, legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account, is_active, created_at, updated_at, version
FROM accounting.accounts
WHERE id = $1;

-- name: GetAccountByCode :one
SELECT id, legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account, is_active, created_at, updated_at, version
FROM accounting.accounts
WHERE legal_entity_id = $1 AND code = $2;

-- name: ListAccountsByLegalEntity :many
SELECT id, legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account, is_active, created_at, updated_at, version
FROM accounting.accounts
WHERE legal_entity_id = $1
ORDER BY code ASC;
