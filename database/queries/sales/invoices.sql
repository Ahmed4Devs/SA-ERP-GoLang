-- name: CreateInvoice :one
INSERT INTO sales.invoices (
    id,
    legal_entity_id,
    invoice_number,
    invoice_type,
    order_id,
    delivery_id,
    customer_id,
    customer_name_ar,
    customer_name_en,
    customer_tax_id,
    customer_cr_number,
    customer_address_json,
    issue_date,
    due_date,
    payment_terms_id,
    currency_code,
    exchange_rate,
    subtotal,
    discount_amount,
    taxable_amount,
    tax_amount,
    total_amount,
    status,
    notes,
    created_by_user_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
) RETURNING *;

-- name: CreateInvoiceLine :one
INSERT INTO sales.invoice_lines (
    id,
    invoice_id,
    line_number,
    item_id,
    description,
    uom,
    quantity,
    unit_price,
    gross_amount,
    discount_amount,
    taxable_amount,
    tax_code,
    tax_rate_percent,
    tax_amount,
    net_amount,
    exemption_reason
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetInvoice :one
SELECT * FROM sales.invoices
WHERE id = $1;

-- name: GetInvoiceByNumber :one
SELECT * FROM sales.invoices
WHERE legal_entity_id = $1 AND invoice_number = $2;

-- name: ListInvoices :many
SELECT * FROM sales.invoices
WHERE legal_entity_id = $1
  AND (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('invoice_type')::TEXT IS NULL OR invoice_type = sqlc.narg('invoice_type'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY issue_date DESC, created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListInvoiceLines :many
SELECT * FROM sales.invoice_lines
WHERE invoice_id = $1
ORDER BY line_number ASC;

-- name: PostInvoice :one
UPDATE sales.invoices
SET status = 'POSTED',
    posted_at = now(),
    journal_entry_id = $2,
    open_item_id = $3,
    qr_code_data = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateInvoiceStatus :one
UPDATE sales.invoices
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;
