-- name: ListCategoriesByEntity :many
SELECT
    c.id,
    c.legal_entity_id,
    c.parent_id,
    c.code,
    c.name_en,
    c.name_ar,
    c.description,
    c.is_active,
    c.created_at,
    c.updated_at,
    p.name_ar AS parent_name_ar,
    p.name_en AS parent_name_en
FROM products.categories c
LEFT JOIN products.categories p ON p.id = c.parent_id
WHERE c.legal_entity_id = $1 AND c.is_active = true
ORDER BY c.code ASC;

-- name: GetCategoryByID :one
SELECT
    id,
    legal_entity_id,
    parent_id,
    code,
    name_en,
    name_ar,
    description,
    is_active,
    created_at,
    updated_at
FROM products.categories
WHERE id = $1;

-- name: GetCategoryByCode :one
SELECT
    id,
    legal_entity_id,
    parent_id,
    code,
    name_en,
    name_ar,
    description,
    is_active,
    created_at,
    updated_at
FROM products.categories
WHERE legal_entity_id = $1 AND code = $2;

-- name: CreateCategory :one
INSERT INTO products.categories (
    legal_entity_id, parent_id, code, name_en, name_ar, description
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateCategory :one
UPDATE products.categories
SET
    parent_id = $2,
    name_en = $3,
    name_ar = $4,
    description = $5,
    is_active = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;
