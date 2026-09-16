-- name: CreateRFQ :one
INSERT INTO purchasing.rfqs (
    id,
    legal_entity_id,
    rfq_number,
    purchase_request_id,
    title,
    rfq_date,
    closing_date,
    status,
    notes,
    created_by_user_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: CreateRFQLine :one
INSERT INTO purchasing.rfq_lines (
    id,
    rfq_id,
    line_number,
    pr_line_id,
    item_id,
    description,
    uom,
    quantity,
    target_unit_price
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetRFQ :one
SELECT * FROM purchasing.rfqs
WHERE id = $1;

-- name: ListRFQs :many
SELECT * FROM purchasing.rfqs
WHERE legal_entity_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListRFQLines :many
SELECT * FROM purchasing.rfq_lines
WHERE rfq_id = $1
ORDER BY line_number ASC;

-- name: UpdateRFQStatus :one
UPDATE purchasing.rfqs
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateSupplierQuote :one
INSERT INTO purchasing.supplier_quotes (
    id,
    legal_entity_id,
    rfq_id,
    supplier_id,
    supplier_quote_reference,
    quote_date,
    valid_until,
    currency_code,
    exchange_rate,
    payment_terms_id,
    subtotal,
    tax_amount,
    total_amount,
    status,
    notes
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: CreateSupplierQuoteLine :one
INSERT INTO purchasing.supplier_quote_lines (
    id,
    quote_id,
    line_number,
    rfq_line_id,
    item_id,
    description,
    quantity,
    unit_price,
    discount_percent,
    tax_rate_percent,
    tax_amount,
    total_amount,
    is_awarded
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetSupplierQuote :one
SELECT * FROM purchasing.supplier_quotes
WHERE id = $1;

-- name: ListSupplierQuotes :many
SELECT * FROM purchasing.supplier_quotes
WHERE legal_entity_id = $1
  AND (sqlc.narg('rfq_id')::UUID IS NULL OR rfq_id = sqlc.narg('rfq_id'))
  AND (sqlc.narg('supplier_id')::UUID IS NULL OR supplier_id = sqlc.narg('supplier_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListSupplierQuoteLines :many
SELECT * FROM purchasing.supplier_quote_lines
WHERE quote_id = $1
ORDER BY line_number ASC;

-- name: UpdateSupplierQuoteStatus :one
UPDATE purchasing.supplier_quotes
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: AwardSupplierQuoteLine :one
UPDATE purchasing.supplier_quote_lines
SET is_awarded = $2
WHERE id = $1
RETURNING *;
