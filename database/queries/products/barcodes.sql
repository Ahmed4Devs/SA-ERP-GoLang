-- name: ListBarcodesByItem :many
SELECT
    b.id,
    b.legal_entity_id,
    b.item_id,
    b.sku_id,
    b.uom_id,
    b.barcode,
    b.barcode_format,
    b.is_primary,
    b.created_at,
    uom.code AS uom_code,
    uom.name_ar AS uom_name_ar,
    s.sku_code
FROM products.barcodes b
LEFT JOIN products.units_of_measure uom ON uom.id = b.uom_id
LEFT JOIN products.skus s ON s.id = b.sku_id
WHERE b.item_id = $1
ORDER BY b.is_primary DESC, b.created_at ASC;

-- name: GetBarcodeByCode :one
SELECT
    b.id,
    b.legal_entity_id,
    b.item_id,
    b.sku_id,
    b.uom_id,
    b.barcode,
    b.barcode_format,
    b.is_primary,
    b.created_at,
    i.code AS item_code,
    i.name_ar AS item_name_ar,
    i.name_en AS item_name_en,
    i.item_type,
    i.list_price,
    i.cost_price,
    uom.code AS uom_code,
    uom.name_ar AS uom_name_ar,
    s.sku_code
FROM products.barcodes b
JOIN products.items i ON i.id = b.item_id
LEFT JOIN products.units_of_measure uom ON uom.id = b.uom_id
LEFT JOIN products.skus s ON s.id = b.sku_id
WHERE b.legal_entity_id = $1 AND b.barcode = $2;

-- name: CreateBarcode :one
INSERT INTO products.barcodes (
    legal_entity_id, item_id, sku_id, uom_id, barcode, barcode_format, is_primary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: DeleteBarcode :exec
DELETE FROM products.barcodes
WHERE id = $1;
