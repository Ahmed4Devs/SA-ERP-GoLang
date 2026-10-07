-- name: CreateStockAdjustment :one
INSERT INTO inventory.stock_adjustments (
    legal_entity_id,
    adjustment_number,
    warehouse_id,
    adjustment_type,
    reason_code,
    status,
    adjustment_date,
    notes,
    created_by, inventory_account_id, offset_account_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: UpdateStockAdjustmentStatus :one
UPDATE inventory.stock_adjustments
SET
    status = $2,
    posted_at = CASE WHEN $2 = 'POSTED' THEN now() ELSE posted_at END,
    posted_by = CASE WHEN $2 = 'POSTED' THEN $3 ELSE posted_by END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetStockAdjustmentByID :one
SELECT
    a.id,
    a.legal_entity_id,
    a.adjustment_number, a.inventory_account_id, a.offset_account_id, a.journal_entry_id,
    a.warehouse_id,
    a.adjustment_type,
    a.reason_code,
    a.status,
    a.adjustment_date,
    a.posted_at,
    a.posted_by,
    a.notes,
    a.created_by,
    a.created_at,
    a.updated_at,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.stock_adjustments a
JOIN warehouse.warehouses wh ON wh.id = a.warehouse_id
WHERE a.id = $1;

-- name: ListStockAdjustmentsByEntity :many
SELECT
    a.id,
    a.legal_entity_id,
    a.adjustment_number, a.inventory_account_id, a.offset_account_id, a.journal_entry_id,
    a.warehouse_id,
    a.adjustment_type,
    a.reason_code,
    a.status,
    a.adjustment_date,
    a.posted_at,
    a.posted_by,
    a.notes,
    a.created_at,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.stock_adjustments a
JOIN warehouse.warehouses wh ON wh.id = a.warehouse_id
WHERE a.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR a.warehouse_id = sqlc.narg('warehouse_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR a.status = sqlc.narg('status'))
ORDER BY a.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateStockAdjustmentLine :one
INSERT INTO inventory.stock_adjustment_lines (
    adjustment_id,
    item_id,
    bin_id,
    lot_number,
    serial_number,
    quantity,
    unit_cost,
    total_cost,
    reason
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListStockAdjustmentLines :many
SELECT
    al.id,
    al.adjustment_id,
    al.item_id,
    al.bin_id,
    al.lot_number,
    al.serial_number,
    al.quantity,
    al.unit_cost,
    al.total_cost,
    al.reason,
    i.base_uom_id AS uom_id,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    bin.code AS bin_code
FROM inventory.stock_adjustment_lines al
JOIN products.items i ON i.id = al.item_id
LEFT JOIN warehouse.bin_locations bin ON bin.id = al.bin_id
WHERE al.adjustment_id = $1
ORDER BY al.id ASC;
