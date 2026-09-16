-- name: CreateStockCount :one
INSERT INTO inventory.stock_counts (
    legal_entity_id,
    count_number,
    warehouse_id,
    count_type,
    status,
    count_date
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateStockCountStatus :one
UPDATE inventory.stock_counts
SET
    status = $2,
    reconciled_at = CASE WHEN $2 = 'POSTED' THEN now() ELSE reconciled_at END,
    reconciled_by = CASE WHEN $2 = 'POSTED' THEN $3 ELSE reconciled_by END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetStockCountByID :one
SELECT
    c.id,
    c.legal_entity_id,
    c.count_number,
    c.warehouse_id,
    c.count_type,
    c.status,
    c.count_date,
    c.reconciled_at,
    c.reconciled_by,
    c.created_at,
    c.updated_at,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.stock_counts c
JOIN warehouse.warehouses wh ON wh.id = c.warehouse_id
WHERE c.id = $1;

-- name: ListStockCountsByEntity :many
SELECT
    c.id,
    c.legal_entity_id,
    c.count_number,
    c.warehouse_id,
    c.count_type,
    c.status,
    c.count_date,
    c.reconciled_at,
    c.reconciled_by,
    c.created_at,
    wh.code AS warehouse_code,
    wh.name_ar AS warehouse_name_ar
FROM inventory.stock_counts c
JOIN warehouse.warehouses wh ON wh.id = c.warehouse_id
WHERE c.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('warehouse_id')::UUID IS NULL OR c.warehouse_id = sqlc.narg('warehouse_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR c.status = sqlc.narg('status'))
ORDER BY c.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateStockCountLine :one
INSERT INTO inventory.stock_count_lines (
    count_id,
    item_id,
    bin_id,
    system_quantity,
    counted_quantity,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListStockCountLines :many
SELECT
    cl.id,
    cl.count_id,
    cl.item_id,
    cl.bin_id,
    cl.system_quantity,
    cl.counted_quantity,
    cl.variance,
    cl.notes,
    i.base_uom_id AS uom_id,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    bin.code AS bin_code
FROM inventory.stock_count_lines cl
JOIN products.items i ON i.id = cl.item_id
LEFT JOIN warehouse.bin_locations bin ON bin.id = cl.bin_id
WHERE cl.count_id = $1
ORDER BY cl.id ASC;

-- name: UpdateStockCountLine :one
UPDATE inventory.stock_count_lines
SET
    counted_quantity = $2,
    notes = $3
WHERE id = $1
RETURNING *;
