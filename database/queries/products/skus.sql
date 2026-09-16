-- name: ListSKUsByItem :many
SELECT
    s.id,
    s.legal_entity_id,
    s.item_id,
    s.variant_id,
    s.sku_code,
    s.cost_price,
    s.selling_price,
    s.is_active,
    s.created_at,
    s.updated_at,
    v.variant_code,
    v.name_ar AS variant_name_ar
FROM products.skus s
LEFT JOIN products.item_variants v ON v.id = s.variant_id
WHERE s.item_id = $1 AND s.is_active = true
ORDER BY s.sku_code ASC;

-- name: GetSKUByID :one
SELECT
    id,
    legal_entity_id,
    item_id,
    variant_id,
    sku_code,
    cost_price,
    selling_price,
    is_active,
    created_at,
    updated_at
FROM products.skus
WHERE id = $1;

-- name: GetSKUByCode :one
SELECT
    id,
    legal_entity_id,
    item_id,
    variant_id,
    sku_code,
    cost_price,
    selling_price,
    is_active,
    created_at,
    updated_at
FROM products.skus
WHERE legal_entity_id = $1 AND sku_code = $2;

-- name: CreateSKU :one
INSERT INTO products.skus (
    legal_entity_id, item_id, variant_id, sku_code, cost_price, selling_price, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateSKUPrices :one
UPDATE products.skus
SET
    cost_price = $2,
    selling_price = $3,
    is_active = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;
