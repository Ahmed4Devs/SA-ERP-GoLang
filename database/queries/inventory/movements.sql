-- name: InsertStockMovement :one
INSERT INTO inventory.stock_movements (
    legal_entity_id,
    movement_number,
    movement_type,
    item_id,
    sku_id,
    warehouse_id,
    bin_id,
    lot_number,
    serial_number,
    quantity,
    uom_id,
    unit_cost,
    total_cost,
    source_doc_type,
    source_doc_id,
    source_doc_line_id,
    occurred_at,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: GetMovementByID :one
SELECT
    m.id,
    m.legal_entity_id,
    m.movement_number,
    m.movement_type,
    m.item_id,
    m.sku_id,
    m.warehouse_id,
    m.bin_id,
    m.lot_number,
    m.serial_number,
    m.quantity,
    m.uom_id,
    m.unit_cost,
    m.total_cost,
    m.source_doc_type,
    m.source_doc_id,
    m.source_doc_line_id,
    m.occurred_at,
    m.created_by,
    m.created_at,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar,
    bin.code AS bin_code,
    uom.code AS uom_code
FROM inventory.stock_movements m
JOIN products.items i ON i.id = m.item_id
JOIN warehouse.warehouses wh ON wh.id = m.warehouse_id
JOIN warehouse.bin_locations bin ON bin.id = m.bin_id
JOIN products.units_of_measure uom ON uom.id = m.uom_id
WHERE m.id = $1;

-- name: GetMovementByNumber :one
SELECT
    m.id,
    m.legal_entity_id,
    m.movement_number,
    m.movement_type,
    m.item_id,
    m.sku_id,
    m.warehouse_id,
    m.bin_id,
    m.lot_number,
    m.serial_number,
    m.quantity,
    m.uom_id,
    m.unit_cost,
    m.total_cost,
    m.source_doc_type,
    m.source_doc_id,
    m.source_doc_line_id,
    m.occurred_at,
    m.created_by,
    m.created_at,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar,
    bin.code AS bin_code,
    uom.code AS uom_code
FROM inventory.stock_movements m
JOIN products.items i ON i.id = m.item_id
JOIN warehouse.warehouses wh ON wh.id = m.warehouse_id
JOIN warehouse.bin_locations bin ON bin.id = m.bin_id
JOIN products.units_of_measure uom ON uom.id = m.uom_id
WHERE m.legal_entity_id = $1 AND m.movement_number = $2;

-- name: ListMovementsByEntity :many
SELECT
    m.id,
    m.legal_entity_id,
    m.movement_number,
    m.movement_type,
    m.item_id,
    m.sku_id,
    m.warehouse_id,
    m.bin_id,
    m.lot_number,
    m.serial_number,
    m.quantity,
    m.uom_id,
    m.unit_cost,
    m.total_cost,
    m.source_doc_type,
    m.source_doc_id,
    m.occurred_at,
    m.created_at,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar,
    bin.code AS bin_code,
    uom.code AS uom_code
FROM inventory.stock_movements m
JOIN products.items i ON i.id = m.item_id
JOIN warehouse.warehouses wh ON wh.id = m.warehouse_id
JOIN warehouse.bin_locations bin ON bin.id = m.bin_id
JOIN products.units_of_measure uom ON uom.id = m.uom_id
WHERE m.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('item_id')::UUID IS NULL OR m.item_id = sqlc.narg('item_id'))
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR m.warehouse_id = sqlc.narg('warehouse_id'))
  AND (sqlc.narg('movement_type')::TEXT IS NULL OR m.movement_type = sqlc.narg('movement_type'))
ORDER BY m.occurred_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetItemBaseUOM :one
SELECT base_uom_id FROM products.items WHERE id = $1;
