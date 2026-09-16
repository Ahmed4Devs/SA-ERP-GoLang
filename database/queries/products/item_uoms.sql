-- name: ListItemUOMsByItem :many
SELECT
    iu.id,
    iu.item_id,
    iu.uom_id,
    iu.conversion_rate,
    iu.is_base_uom,
    iu.is_sales_uom,
    iu.is_purchase_uom,
    iu.created_at,
    uom.code AS uom_code,
    uom.name_ar AS uom_name_ar,
    uom.name_en AS uom_name_en,
    uom.symbol_ar AS uom_symbol_ar
FROM products.item_uoms iu
JOIN products.units_of_measure uom ON uom.id = iu.uom_id
WHERE iu.item_id = $1
ORDER BY iu.is_base_uom DESC, iu.conversion_rate ASC;

-- name: GetItemUOM :one
SELECT
    iu.id,
    iu.item_id,
    iu.uom_id,
    iu.conversion_rate,
    iu.is_base_uom,
    iu.is_sales_uom,
    iu.is_purchase_uom,
    iu.created_at,
    uom.code AS uom_code,
    uom.name_ar AS uom_name_ar
FROM products.item_uoms iu
JOIN products.units_of_measure uom ON uom.id = iu.uom_id
WHERE iu.item_id = $1 AND iu.uom_id = $2;

-- name: UpsertItemUOM :one
INSERT INTO products.item_uoms (
    item_id, uom_id, conversion_rate, is_base_uom, is_sales_uom, is_purchase_uom
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (item_id, uom_id) DO UPDATE SET
    conversion_rate = EXCLUDED.conversion_rate,
    is_base_uom = EXCLUDED.is_base_uom,
    is_sales_uom = EXCLUDED.is_sales_uom,
    is_purchase_uom = EXCLUDED.is_purchase_uom
RETURNING *;

-- name: DeleteItemUOM :exec
DELETE FROM products.item_uoms
WHERE id = $1;
