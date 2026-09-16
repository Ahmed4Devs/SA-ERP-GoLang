-- name: CreateGoodsReceipt :one
INSERT INTO purchasing.goods_receipts (
    id,
    legal_entity_id,
    receipt_number,
    purchase_order_id,
    supplier_id,
    warehouse_id,
    receipt_date,
    delivery_note_number,
    status,
    notes,
    received_by_user_id,
    inspected_by_user_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: CreateGoodsReceiptLine :one
INSERT INTO purchasing.goods_receipt_lines (
    id,
    goods_receipt_id,
    line_number,
    po_line_id,
    item_id,
    warehouse_id,
    bin_id,
    received_quantity,
    accepted_quantity,
    rejected_quantity,
    unit_cost,
    total_cost,
    inventory_movement_id,
    rejection_reason
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetGoodsReceipt :one
SELECT * FROM purchasing.goods_receipts
WHERE id = $1;

-- name: GetGoodsReceiptByNumber :one
SELECT * FROM purchasing.goods_receipts
WHERE legal_entity_id = $1 AND receipt_number = $2;

-- name: ListGoodsReceipts :many
SELECT * FROM purchasing.goods_receipts
WHERE legal_entity_id = $1
  AND (sqlc.narg('purchase_order_id')::UUID IS NULL OR purchase_order_id = sqlc.narg('purchase_order_id'))
  AND (sqlc.narg('supplier_id')::UUID IS NULL OR supplier_id = sqlc.narg('supplier_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListGoodsReceiptLines :many
SELECT * FROM purchasing.goods_receipt_lines
WHERE goods_receipt_id = $1
ORDER BY line_number ASC;

-- name: UpdateGoodsReceiptStatus :one
UPDATE purchasing.goods_receipts
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateGoodsReceiptLineMovement :one
UPDATE purchasing.goods_receipt_lines
SET inventory_movement_id = $2
WHERE id = $1
RETURNING *;
