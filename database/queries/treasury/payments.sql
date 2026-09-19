-- name: CreateTreasuryPayment :one
INSERT INTO treasury.payments (
    id, legal_entity_id, payment_number, cash_account_id, payment_date,
    payment_type, paid_to, amount, currency_code, reference_number,
    description, status, gl_offset_account_id, journal_entry_id,
    created_by_user_id, approved_by_user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14,
    $15, $16, now(), now()
)
RETURNING *;

-- name: GetTreasuryPaymentByID :one
SELECT * FROM treasury.payments
WHERE id = $1 AND legal_entity_id = $2;

-- name: LockTreasuryPaymentByID :one
SELECT * FROM treasury.payments
WHERE id = $1 AND legal_entity_id = $2
FOR UPDATE;

-- name: GetTreasuryPaymentByNumber :one
SELECT * FROM treasury.payments
WHERE legal_entity_id = $1 AND payment_number = $2;

-- name: ListTreasuryPayments :many
SELECT * FROM treasury.payments
WHERE legal_entity_id = $1
ORDER BY payment_date DESC, payment_number DESC;

-- name: UpdateTreasuryPaymentStatus :one
UPDATE treasury.payments
SET status = $3,
    journal_entry_id = COALESCE($4, journal_entry_id),
    approved_by_user_id = COALESCE($5, approved_by_user_id),
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
