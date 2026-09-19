-- name: CreateTreasuryReceipt :one
INSERT INTO treasury.receipts (
    id, legal_entity_id, receipt_number, cash_account_id, receipt_date,
    receipt_type, received_from, amount, currency_code, payment_method,
    cheque_number, cheque_date, reference_number, description, status,
    gl_offset_account_id, journal_entry_id, created_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15,
    $16, $17, $18, now(), now()
)
RETURNING *;

-- name: GetTreasuryReceiptByID :one
SELECT * FROM treasury.receipts
WHERE id = $1 AND legal_entity_id = $2;

-- name: LockTreasuryReceiptByID :one
SELECT * FROM treasury.receipts
WHERE id = $1 AND legal_entity_id = $2
FOR UPDATE;

-- name: GetTreasuryReceiptByNumber :one
SELECT * FROM treasury.receipts
WHERE legal_entity_id = $1 AND receipt_number = $2;

-- name: ListTreasuryReceipts :many
SELECT * FROM treasury.receipts
WHERE legal_entity_id = $1
ORDER BY receipt_date DESC, receipt_number DESC;

-- name: UpdateTreasuryReceiptStatus :one
UPDATE treasury.receipts
SET status = $3,
    journal_entry_id = COALESCE($4, journal_entry_id),
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
