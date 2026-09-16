-- name: ListItemsByEntity :many
SELECT
    i.id,
    i.legal_entity_id,
    i.category_id,
    i.code,
    i.name_en,
    i.name_ar,
    i.description_en,
    i.description_ar,
    i.item_type,
    i.tracking_type,
    i.base_uom_id,
    i.purchase_uom_id,
    i.sales_uom_id,
    i.default_tax_code_id,
    i.cost_price,
    i.list_price,
    i.is_sales_enabled,
    i.is_purchase_enabled,
    i.is_active,
    i.created_at,
    i.updated_at,
    i.version,
    cat.name_ar AS category_name_ar,
    cat.name_en AS category_name_en,
    uom.code AS base_uom_code,
    uom.name_ar AS base_uom_name_ar,
    tc.code AS tax_code,
    tc.name_ar AS tax_name_ar
FROM products.items i
LEFT JOIN products.categories cat ON cat.id = i.category_id
JOIN products.units_of_measure uom ON uom.id = i.base_uom_id
LEFT JOIN tax.tax_codes tc ON tc.id = i.default_tax_code_id
WHERE i.legal_entity_id = sqlc.arg('legal_entity_id')
  AND (sqlc.narg('item_type')::TEXT IS NULL OR i.item_type = sqlc.narg('item_type'))
  AND (sqlc.narg('category_id')::UUID IS NULL OR i.category_id = sqlc.narg('category_id'))
  AND i.is_active = true
ORDER BY i.code ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetItemByID :one
SELECT
    i.id,
    i.legal_entity_id,
    i.category_id,
    i.code,
    i.name_en,
    i.name_ar,
    i.description_en,
    i.description_ar,
    i.item_type,
    i.tracking_type,
    i.base_uom_id,
    i.purchase_uom_id,
    i.sales_uom_id,
    i.default_tax_code_id,
    i.cost_price,
    i.list_price,
    i.is_sales_enabled,
    i.is_purchase_enabled,
    i.is_active,
    i.created_at,
    i.updated_at,
    i.version,
    cat.name_ar AS category_name_ar,
    cat.name_en AS category_name_en,
    uom.code AS base_uom_code,
    uom.name_ar AS base_uom_name_ar,
    tc.code AS tax_code,
    tc.name_ar AS tax_name_ar
FROM products.items i
LEFT JOIN products.categories cat ON cat.id = i.category_id
JOIN products.units_of_measure uom ON uom.id = i.base_uom_id
LEFT JOIN tax.tax_codes tc ON tc.id = i.default_tax_code_id
WHERE i.id = $1;

-- name: GetItemByCode :one
SELECT
    i.id,
    i.legal_entity_id,
    i.category_id,
    i.code,
    i.name_en,
    i.name_ar,
    i.description_en,
    i.description_ar,
    i.item_type,
    i.tracking_type,
    i.base_uom_id,
    i.purchase_uom_id,
    i.sales_uom_id,
    i.default_tax_code_id,
    i.cost_price,
    i.list_price,
    i.is_sales_enabled,
    i.is_purchase_enabled,
    i.is_active,
    i.created_at,
    i.updated_at,
    i.version,
    cat.name_ar AS category_name_ar,
    cat.name_en AS category_name_en,
    uom.code AS base_uom_code,
    uom.name_ar AS base_uom_name_ar,
    tc.code AS tax_code,
    tc.name_ar AS tax_name_ar
FROM products.items i
LEFT JOIN products.categories cat ON cat.id = i.category_id
JOIN products.units_of_measure uom ON uom.id = i.base_uom_id
LEFT JOIN tax.tax_codes tc ON tc.id = i.default_tax_code_id
WHERE i.legal_entity_id = $1 AND i.code = $2;

-- name: CreateItem :one
INSERT INTO products.items (
    legal_entity_id,
    category_id,
    code,
    name_en,
    name_ar,
    description_en,
    description_ar,
    item_type,
    tracking_type,
    base_uom_id,
    purchase_uom_id,
    sales_uom_id,
    default_tax_code_id,
    cost_price,
    list_price,
    is_sales_enabled,
    is_purchase_enabled,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
RETURNING *;

-- name: UpdateItem :one
UPDATE products.items
SET
    category_id = $2,
    name_en = $3,
    name_ar = $4,
    description_en = $5,
    description_ar = $6,
    item_type = $7,
    tracking_type = $8,
    purchase_uom_id = $9,
    sales_uom_id = $10,
    default_tax_code_id = $11,
    cost_price = $12,
    list_price = $13,
    is_sales_enabled = $14,
    is_purchase_enabled = $15,
    is_active = $16,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;

-- name: SearchItems :many
SELECT
    i.id,
    i.legal_entity_id,
    i.code,
    i.name_en,
    i.name_ar,
    i.item_type,
    i.list_price,
    i.cost_price,
    uom.code AS base_uom_code
FROM products.items i
JOIN products.units_of_measure uom ON uom.id = i.base_uom_id
WHERE i.legal_entity_id = sqlc.arg('legal_entity_id')
  AND i.is_active = true
  AND (
      i.code ILIKE '%' || sqlc.arg('query') || '%'
      OR i.name_en ILIKE '%' || sqlc.arg('query') || '%'
      OR i.name_ar ILIKE '%' || sqlc.arg('query') || '%'
  )
ORDER BY i.code ASC
LIMIT 20;
