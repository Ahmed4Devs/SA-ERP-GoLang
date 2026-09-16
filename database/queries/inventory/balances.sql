-- name: LockBalance :one
SELECT
    id,
    legal_entity_id,
    item_id,
    warehouse_id,
    bin_id,
    quantity_on_hand,
    quantity_reserved,
    quantity_available,
    average_cost,
    total_value,
    updated_at,
    version
FROM inventory.stock_balances
WHERE legal_entity_id = $1 AND item_id = $2 AND warehouse_id = $3 AND bin_id = $4
FOR UPDATE;

-- name: GetBalance :one
SELECT
    b.id,
    b.legal_entity_id,
    b.item_id,
    b.warehouse_id,
    b.bin_id,
    b.quantity_on_hand,
    b.quantity_reserved,
    b.quantity_available,
    b.average_cost,
    b.total_value,
    b.updated_at,
    b.version,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar,
    bin.code AS bin_code
FROM inventory.stock_balances b
JOIN products.items i ON i.id = b.item_id
JOIN warehouse.warehouses wh ON wh.id = b.warehouse_id
JOIN warehouse.bin_locations bin ON bin.id = b.bin_id
WHERE b.legal_entity_id = $1 AND b.item_id = $2 AND b.warehouse_id = $3 AND b.bin_id = $4;

-- name: ListBalancesByEntity :many
SELECT
    b.id,
    b.legal_entity_id,
    b.item_id,
    b.warehouse_id,
    b.bin_id,
    b.quantity_on_hand,
    b.quantity_reserved,
    b.quantity_available,
    b.average_cost,
    b.total_value,
    b.updated_at,
    b.version,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    uom.code AS base_uom_code,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar,
    bin.code AS bin_code
FROM inventory.stock_balances b
JOIN products.items i ON i.id = b.item_id
JOIN products.units_of_measure uom ON uom.id = i.base_uom_id
JOIN warehouse.warehouses wh ON wh.id = b.warehouse_id
JOIN warehouse.bin_locations bin ON bin.id = b.bin_id
WHERE b.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('item_id')::UUID IS NULL OR b.item_id = sqlc.narg('item_id'))
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR b.warehouse_id = sqlc.narg('warehouse_id'))
ORDER BY i.code ASC, wh.code ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: UpsertBalance :one
INSERT INTO inventory.stock_balances (
    legal_entity_id,
    item_id,
    warehouse_id,
    bin_id,
    quantity_on_hand,
    quantity_reserved,
    average_cost,
    total_value
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (legal_entity_id, item_id, warehouse_id, bin_id)
DO UPDATE SET
    quantity_on_hand = EXCLUDED.quantity_on_hand,
    quantity_reserved = EXCLUDED.quantity_reserved,
    average_cost = EXCLUDED.average_cost,
    total_value = EXCLUDED.total_value,
    updated_at = now(),
    version = inventory.stock_balances.version + 1
RETURNING *;
