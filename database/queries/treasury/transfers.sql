-- name: CreateTreasuryTransfer :one
INSERT INTO treasury.transfers (
    id, legal_entity_id, transfer_number, transfer_date, source_type,
    source_cash_account_id, source_bank_account_id, destination_type,
    destination_cash_account_id, destination_bank_account_id, amount,
    currency_code, fee_amount, fee_gl_account_id, status,
    reference_number, notes, journal_entry_id, initiated_by_user_id,
    completed_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12, $13, $14, $15,
    $16, $17, $18, $19,
    $20, now(), now()
)
RETURNING *;

-- name: GetTreasuryTransferByID :one
SELECT * FROM treasury.transfers
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetTreasuryTransferByNumber :one
SELECT * FROM treasury.transfers
WHERE legal_entity_id = $1 AND transfer_number = $2;

-- name: ListTreasuryTransfers :many
SELECT * FROM treasury.transfers
WHERE legal_entity_id = $1
ORDER BY transfer_date DESC, transfer_number DESC;

-- name: UpdateTreasuryTransferStatus :one
UPDATE treasury.transfers
SET status = $3,
    journal_entry_id = COALESCE($4, journal_entry_id),
    completed_by_user_id = COALESCE($5, completed_by_user_id),
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
