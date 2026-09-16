-- name: CreateQuotation :one
INSERT INTO sales.quotations (
    id,
    legal_entity_id,
    quotation_number,
    customer_id,
    customer_name_en,
    customer_name_ar,
    quote_date,
    valid_until,
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: CreateQuotationLine :one
INSERT INTO sales.quotation_lines (
    id,
    quotation_id,
    line_number,
    item_id,
    description_en,
    description_ar,
    uom,
    quantity,
    unit_price,
    discount_percent,
    discount_amount,
    taxable_amount,
    tax_code,
    tax_rate_percent,
    tax_amount,
    total_amount
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetQuotation :one
SELECT * FROM sales.quotations
WHERE id = $1;

-- name: GetQuotationByNumber :one
SELECT * FROM sales.quotations
WHERE legal_entity_id = $1 AND quotation_number = $2;

-- name: ListQuotations :many
SELECT * FROM sales.quotations
WHERE legal_entity_id = $1
  AND (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListQuotationLines :many
SELECT * FROM sales.quotation_lines
WHERE quotation_id = $1
ORDER BY line_number ASC;

-- name: UpdateQuotationStatus :one
UPDATE sales.quotations
SET status = $2,
    converted_order_id = COALESCE($3, converted_order_id),
    updated_at = now()
WHERE id = $1
RETURNING *;
