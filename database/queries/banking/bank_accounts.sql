-- name: CreateBankAccount :one
INSERT INTO banking.bank_accounts (
    id, legal_entity_id, account_code, bank_name_en, bank_name_ar,
    branch_name, account_number, account_number_masked, iban, swift_bic,
    currency_code, gl_account_id, opening_balance, current_balance, available_balance,
    overdraft_limit, status, is_primary, contact_person, contact_phone,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20,
    now(), now()
)
RETURNING *;

-- name: GetBankAccountByID :one
SELECT * FROM banking.bank_accounts
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetBankAccountByCode :one
SELECT * FROM banking.bank_accounts
WHERE legal_entity_id = $1 AND account_code = $2;

-- name: GetBankAccountByIBAN :one
SELECT * FROM banking.bank_accounts
WHERE legal_entity_id = $1 AND iban = $2;

-- name: ListBankAccounts :many
SELECT * FROM banking.bank_accounts
WHERE legal_entity_id = $1
ORDER BY account_code ASC;

-- name: UpdateBankAccountBalance :one
UPDATE banking.bank_accounts
SET current_balance = current_balance + $3,
    available_balance = available_balance + $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: UpdateBankAccountStatus :one
UPDATE banking.bank_accounts
SET status = $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
