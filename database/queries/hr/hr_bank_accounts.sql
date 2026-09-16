-- name: CreateEmployeeBankAccount :one
INSERT INTO hr.bank_accounts (
    id, employee_id, bank_name_en, bank_name_ar,
    iban, iban_masked, swift_bic, is_primary,
    account_status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, now(), now()
)
RETURNING *;

-- name: ListEmployeeBankAccounts :many
SELECT * FROM hr.bank_accounts
WHERE employee_id = $1
ORDER BY is_primary DESC, created_at ASC;
