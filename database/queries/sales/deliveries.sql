-- name: CreateDelivery :one
INSERT INTO sales.deliveries (
    id,
    legal_entity_id,
    delivery_number,
    order_id,
    customer_id,
    warehouse_id,
    delivery_date,
    status,
    shipping_reference,
    notes,
    created_by_user_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: CreateDeliveryLine :one
INSERT INTO sales.delivery_lines (
    id,
    delivery_id,
    line_number,
    order_line_id,
    item_id,
    uom,
    quantity,
    warehouse_id,
    bin_id,
    inventory_movement_id,
    unit_cost,
    total_cost
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetDelivery :one
SELECT * FROM sales.deliveries
WHERE id = $1;

-- name: GetDeliveryByNumber :one
SELECT * FROM sales.deliveries
WHERE legal_entity_id = $1 AND delivery_number = $2;

-- name: ListDeliveries :many
SELECT * FROM sales.deliveries
WHERE legal_entity_id = $1
  AND (sqlc.narg('customer_id')::UUID IS NULL OR customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('order_id')::UUID IS NULL OR order_id = sqlc.narg('order_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListDeliveryLines :many
SELECT * FROM sales.delivery_lines
WHERE delivery_id = $1
ORDER BY line_number ASC;

-- name: UpdateDeliveryStatus :one
UPDATE sales.deliveries
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;
