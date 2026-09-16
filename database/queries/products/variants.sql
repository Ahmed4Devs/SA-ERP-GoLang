-- name: ListVariantsByItem :many
SELECT id, item_id, variant_code, name_en, name_ar, attributes, is_active, created_at, updated_at
FROM products.item_variants
WHERE item_id = $1 AND is_active = true
ORDER BY variant_code ASC;

-- name: GetVariantByID :one
SELECT id, item_id, variant_code, name_en, name_ar, attributes, is_active, created_at, updated_at
FROM products.item_variants
WHERE id = $1;

-- name: GetVariantByCode :one
SELECT id, item_id, variant_code, name_en, name_ar, attributes, is_active, created_at, updated_at
FROM products.item_variants
WHERE item_id = $1 AND variant_code = $2;

-- name: CreateVariant :one
INSERT INTO products.item_variants (
    item_id, variant_code, name_en, name_ar, attributes, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateVariant :one
UPDATE products.item_variants
SET
    name_en = $2,
    name_ar = $3,
    attributes = $4,
    is_active = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;
