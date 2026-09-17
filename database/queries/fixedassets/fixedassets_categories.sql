-- name: CreateAssetCategory :one
INSERT INTO fixedassets.categories (
    id, legal_entity_id, code, name_en, name_ar, description_en, description_ar,
    depreciation_method, useful_life_months, salvage_value_rate,
    asset_account_id, accumulated_depreciation_account_id, depreciation_expense_account_id, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: ListAssetCategories :many
SELECT * FROM fixedassets.categories
WHERE legal_entity_id = $1
ORDER BY code;

-- name: GetAssetCategoryByID :one
SELECT * FROM fixedassets.categories
WHERE id = $1;

-- name: CreateAssetLocation :one
INSERT INTO fixedassets.asset_locations (
    id, legal_entity_id, code, name_en, name_ar, building, floor, room, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListAssetLocations :many
SELECT * FROM fixedassets.asset_locations
WHERE legal_entity_id = $1
ORDER BY code;

-- name: GetAssetLocationByID :one
SELECT * FROM fixedassets.asset_locations
WHERE id = $1;

-- name: CreateDepreciationBook :one
INSERT INTO fixedassets.depreciation_books (
    id, legal_entity_id, code, name_en, name_ar, is_primary
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListDepreciationBooks :many
SELECT * FROM fixedassets.depreciation_books
WHERE legal_entity_id = $1
ORDER BY code;

-- name: GetDepreciationBookByCode :one
SELECT * FROM fixedassets.depreciation_books
WHERE legal_entity_id = $1 AND code = $2;
