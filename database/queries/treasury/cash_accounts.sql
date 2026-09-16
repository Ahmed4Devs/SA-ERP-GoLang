-- name: CreateCashAccount :one
INSERT INTO treasury.cash_accounts (
    id, legal_entity_id, branch_id, account_code, name_en, name_ar,
    account_type, gl_account_id, currency_code, custodian_user_id,
    max_limit, current_balance, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $13, now(), now()
)
RETURNING *;

-- name: GetCashAccountByID :one
SELECT * FROM treasury.cash_accounts
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetCashAccountByCode :one
SELECT * FROM treasury.cash_accounts
WHERE legal_entity_id = $1 AND account_code = $2;

-- name: ListCashAccounts :many
SELECT * FROM treasury.cash_accounts
WHERE legal_entity_id = $1
ORDER BY account_code ASC;

-- name: UpdateCashAccountBalance :one
UPDATE treasury.cash_accounts
SET current_balance = current_balance + $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: UpdateCashAccountStatus :one
UPDATE treasury.cash_accounts
SET status = $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
