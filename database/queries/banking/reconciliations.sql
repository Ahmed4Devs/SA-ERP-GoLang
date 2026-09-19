-- name: CreateBankReconciliation :one
INSERT INTO banking.reconciliations (
    id, legal_entity_id, bank_account_id, statement_id,
    reconciliation_number, as_of_date, statement_ending_balance,
    gl_book_balance, unreconciled_difference, status, notes,
    reconciled_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, $11,
    $12, now(), now()
)
RETURNING *;

-- name: GetBankReconciliationByID :one
SELECT * FROM banking.reconciliations
WHERE id = $1 AND legal_entity_id = $2;

-- name: LockBankReconciliationByID :one
SELECT * FROM banking.reconciliations
WHERE id = $1 AND legal_entity_id = $2
FOR UPDATE;

-- name: ListBankReconciliations :many
SELECT * FROM banking.reconciliations
WHERE bank_account_id = $1
ORDER BY as_of_date DESC, created_at DESC;

-- name: UpdateBankReconciliation :one
UPDATE banking.reconciliations
SET status = $3,
    statement_ending_balance = $4,
    gl_book_balance = $5,
    unreconciled_difference = $6,
    notes = $7,
    reconciled_by_user_id = COALESCE($8, reconciled_by_user_id),
    reconciled_at = COALESCE($9, reconciled_at),
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: CreateReconciliationMatch :one
INSERT INTO banking.reconciliation_matches (
    id, reconciliation_id, statement_line_id, match_type,
    matched_entity_type, matched_entity_id, matched_amount,
    difference_amount, notes, matched_by_user_id, matched_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, now()
)
RETURNING *;

-- name: ListReconciliationMatches :many
SELECT * FROM banking.reconciliation_matches
WHERE reconciliation_id = $1
ORDER BY matched_at ASC;

-- name: DeleteReconciliationMatch :exec
DELETE FROM banking.reconciliation_matches
WHERE id = $1;
