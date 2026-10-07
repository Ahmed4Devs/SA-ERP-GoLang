-- name: InsertCostLayer :one
INSERT INTO inventory.cost_layers (
    legal_entity_id,
    item_id,
    warehouse_id,
    source_movement_id,
    original_quantity,
    remaining_quantity,
    unit_cost,
    total_cost
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetActiveCostLayersForUpdate :many
SELECT
    id,
    legal_entity_id,
    item_id,
    warehouse_id,
    source_movement_id,
    original_quantity,
    remaining_quantity,
    unit_cost,
    total_cost,
    created_at,
    updated_at
FROM inventory.cost_layers
WHERE legal_entity_id = $1
  AND item_id = $2
  AND warehouse_id = $3
  AND remaining_quantity > 0
ORDER BY created_at ASC
FOR UPDATE;

-- name: UpdateCostLayerRemaining :one
UPDATE inventory.cost_layers
SET
    remaining_quantity = $2,
    total_cost = $2 * unit_cost,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: InsertCostLayerConsumption :one
INSERT INTO inventory.cost_layer_consumptions (
    legal_entity_id,
    cost_layer_id,
    issue_movement_id,
    quantity_consumed,
    unit_cost,
    total_cost
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListCostLayers :many
SELECT
    cl.id,
    cl.legal_entity_id,
    cl.item_id,
    cl.warehouse_id,
    cl.source_movement_id,
    cl.original_quantity,
    cl.remaining_quantity,
    cl.unit_cost,
    cl.total_cost,
    cl.created_at,
    cl.updated_at,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.cost_layers cl
JOIN products.items i ON i.id = cl.item_id
JOIN warehouse.warehouses wh ON wh.id = cl.warehouse_id
WHERE cl.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('item_id')::UUID IS NULL OR cl.item_id = sqlc.narg('item_id'))
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR cl.warehouse_id = sqlc.narg('warehouse_id'))
  AND (sqlc.narg('only_active')::BOOLEAN IS NULL OR (sqlc.narg('only_active')::BOOLEAN = true AND cl.remaining_quantity > 0))
ORDER BY cl.created_at ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateInventoryRevaluation :one
INSERT INTO inventory.inventory_revaluations (
    legal_entity_id,
    revaluation_number,
    warehouse_id,
    item_id,
    quantity_on_hand,
    old_unit_cost,
    new_unit_cost,
    old_total_value,
    new_total_value,
    variance_amount,
    reason,
    journal_entry_id,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: ListInventoryRevaluations :many
SELECT
    r.id,
    r.legal_entity_id,
    r.revaluation_number,
    r.warehouse_id,
    r.item_id,
    r.quantity_on_hand,
    r.old_unit_cost,
    r.new_unit_cost,
    r.old_total_value,
    r.new_total_value,
    r.variance_amount,
    r.reason,
    r.journal_entry_id,
    r.created_by,
    r.created_at,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.inventory_revaluations r
JOIN products.items i ON i.id = r.item_id
JOIN warehouse.warehouses wh ON wh.id = r.warehouse_id
WHERE r.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR r.warehouse_id = sqlc.narg('warehouse_id'))
  AND (sqlc.narg('item_id')::UUID IS NULL OR r.item_id = sqlc.narg('item_id'))
ORDER BY r.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetValuationSummary :many
SELECT
    b.legal_entity_id,
    b.item_id,
    b.warehouse_id,
    b.quantity_on_hand,
    b.quantity_reserved,
    b.quantity_available,
    b.average_cost,
    b.total_value,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    i.costing_method,
    i.standard_cost,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.stock_balances b
JOIN products.items i ON i.id = b.item_id
JOIN warehouse.warehouses wh ON wh.id = b.warehouse_id
WHERE b.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR b.warehouse_id = sqlc.narg('warehouse_id'))
  AND (sqlc.narg('item_id')::UUID IS NULL OR b.item_id = sqlc.narg('item_id'))
ORDER BY i.code ASC, wh.code ASC;

-- name: UpdateStockBalanceCostAndValue :exec
UPDATE inventory.stock_balances
SET
    average_cost = $4,
    total_value = round(quantity_on_hand * $4, 6),
    version = version + 1,
    updated_at = now()
WHERE legal_entity_id = $1
  AND item_id = $2
  AND warehouse_id = $3
;

-- name: GetItemCostingDetails :one
SELECT
    id,
    costing_method,
    cost_price,
    standard_cost
FROM products.items
WHERE id = $1;
