-- name: CreateTransferOrder :one
INSERT INTO inventory.transfer_orders (
    legal_entity_id,
    order_number,
    from_warehouse_id,
    to_warehouse_id,
    status,
    shipping_date,
    expected_delivery_date,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateTransferOrderStatus :one
UPDATE inventory.transfer_orders
SET
    status = $2,
    received_date = CASE WHEN $2 = 'RECEIVED' THEN now() ELSE received_date END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetTransferOrderByID :one
SELECT
    t.id,
    t.legal_entity_id,
    t.order_number,
    t.from_warehouse_id,
    t.to_warehouse_id,
    t.status,
    t.shipping_date,
    t.expected_delivery_date,
    t.received_date,
    t.notes,
    t.created_by,
    t.created_at,
    t.updated_at,
    fw.code AS from_warehouse_code,
    fw.name_ar AS from_warehouse_name_ar,
    tw.code AS to_warehouse_code,
    tw.name_ar AS to_warehouse_name_ar
FROM inventory.transfer_orders t
JOIN warehouse.warehouses fw ON fw.id = t.from_warehouse_id
JOIN warehouse.warehouses tw ON tw.id = t.to_warehouse_id
WHERE t.id = $1;

-- name: ListTransferOrdersByEntity :many
SELECT
    t.id,
    t.legal_entity_id,
    t.order_number,
    t.from_warehouse_id,
    t.to_warehouse_id,
    t.status,
    t.shipping_date,
    t.expected_delivery_date,
    t.received_date,
    t.notes,
    t.created_at,
    fw.code AS from_warehouse_code,
    fw.name_ar AS from_warehouse_name_ar,
    tw.code AS to_warehouse_code,
    tw.name_ar AS to_warehouse_name_ar
FROM inventory.transfer_orders t
JOIN warehouse.warehouses fw ON fw.id = t.from_warehouse_id
JOIN warehouse.warehouses tw ON tw.id = t.to_warehouse_id
WHERE t.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('status')::TEXT IS NULL OR t.status = sqlc.narg('status'))
ORDER BY t.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateTransferOrderLine :one
INSERT INTO inventory.transfer_order_lines (
    transfer_order_id,
    item_id,
    sku_id,
    requested_quantity,
    dispatched_quantity,
    received_quantity,
    uom_id,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListTransferOrderLines :many
SELECT
    l.id,
    l.transfer_order_id,
    l.item_id,
    l.sku_id,
    l.requested_quantity,
    l.dispatched_quantity,
    l.received_quantity,
    l.uom_id,
    l.notes,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    uom.code AS uom_code
FROM inventory.transfer_order_lines l
JOIN products.items i ON i.id = l.item_id
JOIN products.units_of_measure uom ON uom.id = l.uom_id
WHERE l.transfer_order_id = $1
ORDER BY l.id ASC;

-- name: UpdateTransferOrderLineQuantities :one
UPDATE inventory.transfer_order_lines
SET
    dispatched_quantity = $2,
    received_quantity = $3
WHERE id = $1
RETURNING *;
