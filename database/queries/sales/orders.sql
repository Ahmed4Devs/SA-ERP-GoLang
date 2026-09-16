-- name: CreateOrder :one
INSERT INTO sales.orders (
    id,
    legal_entity_id,
    order_number,
    customer_id,
    quotation_id,
    customer_po_number,
    order_date,
    expected_delivery_date,
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
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
) RETURNING *;

-- name: CreateOrderLine :one
INSERT INTO sales.order_lines (
    id,
    order_id,
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

-- name: GetOrder :one
SELECT * FROM sales.orders
WHERE id = $1;

-- name: GetOrderByNumber :one
SELECT * FROM sales.orders
WHERE legal_entity_id = $1 AND order_number = $2;

-- name: ListOrders :many
SELECT * FROM sales.orders
WHERE legal_entity_id = $1
  AND (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListOrderLines :many
SELECT * FROM sales.order_lines
WHERE order_id = $1
ORDER BY line_number ASC;

-- name: UpdateOrderStatus :one
UPDATE sales.orders
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateOrderLineDeliveredQty :one
UPDATE sales.order_lines
SET delivered_quantity = delivered_quantity + $2
WHERE id = $1
RETURNING *;

-- name: UpdateOrderLineInvoicedQty :one
UPDATE sales.order_lines
SET invoiced_quantity = invoiced_quantity + $2
WHERE id = $1
RETURNING *;
