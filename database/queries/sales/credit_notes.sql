-- name: CreateCreditNote :one
INSERT INTO sales.credit_notes (
    id,
    legal_entity_id,
    credit_note_number,
    original_invoice_id,
    customer_id,
    issue_date,
    reason_code,
    reason_details,
    currency_code,
    exchange_rate,
    subtotal,
    tax_amount,
    total_amount,
    status,
    created_by_user_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: CreateCreditNoteLine :one
INSERT INTO sales.credit_note_lines (
    id,
    credit_note_id,
    line_number,
    item_id,
    description,
    quantity,
    unit_price,
    taxable_amount,
    tax_code,
    tax_rate_percent,
    tax_amount,
    total_amount
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetCreditNote :one
SELECT * FROM sales.credit_notes
WHERE id = $1;

-- name: GetCreditNoteByNumber :one
SELECT * FROM sales.credit_notes
WHERE legal_entity_id = $1 AND credit_note_number = $2;

-- name: ListCreditNotes :many
SELECT * FROM sales.credit_notes
WHERE legal_entity_id = $1
  AND (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('original_invoice_id')::UUID IS NULL OR original_invoice_id = sqlc.narg('original_invoice_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY issue_date DESC, created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListCreditNoteLines :many
SELECT * FROM sales.credit_note_lines
WHERE credit_note_id = $1
ORDER BY line_number ASC;

-- name: PostCreditNote :one
UPDATE sales.credit_notes
SET status = 'POSTED',
    posted_at = now(),
    journal_entry_id = $2,
    open_item_id = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;
