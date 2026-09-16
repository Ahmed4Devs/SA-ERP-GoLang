-- name: CreateBankStatement :one
INSERT INTO banking.statements (
    id, legal_entity_id, bank_account_id, statement_number, statement_date,
    start_date, end_date, opening_balance, closing_balance, total_credits,
    total_debits, currency_code, file_format, status, source_filename,
    imported_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15,
    $16, now(), now()
)
RETURNING *;

-- name: GetBankStatementByID :one
SELECT * FROM banking.statements
WHERE id = $1 AND legal_entity_id = $2;

-- name: ListBankStatements :many
SELECT * FROM banking.statements
WHERE bank_account_id = $1
ORDER BY statement_date DESC;

-- name: UpdateBankStatementStatus :one
UPDATE banking.statements
SET status = $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: CreateStatementLine :one
INSERT INTO banking.statement_lines (
    id, statement_id, line_number, transaction_date, value_date,
    description, reference_number, transaction_type, amount,
    currency_code, balance_after, reconciled_status, created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, now()
)
RETURNING *;

-- name: GetStatementLineByID :one
SELECT * FROM banking.statement_lines
WHERE id = $1;

-- name: ListStatementLines :many
SELECT * FROM banking.statement_lines
WHERE statement_id = $1
ORDER BY line_number ASC;

-- name: UpdateStatementLineReconciledStatus :one
UPDATE banking.statement_lines
SET reconciled_status = $2
WHERE id = $1
RETURNING *;
