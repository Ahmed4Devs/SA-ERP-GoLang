-- name: ListWarehousesByEntity :many
SELECT
    id,
    legal_entity_id,
    code,
    name_en,
    name_ar,
    warehouse_type,
    is_active,
    address_line1,
    city,
    postal_code,
    country_code,
    created_at,
    updated_at
FROM warehouse.warehouses
WHERE legal_entity_id = $1 AND is_active = true
ORDER BY code ASC;

-- name: GetWarehouseByID :one
SELECT
    id,
    legal_entity_id,
    code,
    name_en,
    name_ar,
    warehouse_type,
    is_active,
    address_line1,
    city,
    postal_code,
    country_code,
    created_at,
    updated_at
FROM warehouse.warehouses
WHERE id = $1;

-- name: GetWarehouseByCode :one
SELECT
    id,
    legal_entity_id,
    code,
    name_en,
    name_ar,
    warehouse_type,
    is_active,
    address_line1,
    city,
    postal_code,
    country_code,
    created_at,
    updated_at
FROM warehouse.warehouses
WHERE legal_entity_id = $1 AND code = $2;

-- name: CreateWarehouse :one
INSERT INTO warehouse.warehouses (
    legal_entity_id,
    code,
    name_en,
    name_ar,
    warehouse_type,
    is_active,
    address_line1,
    city,
    postal_code,
    country_code
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateWarehouse :one
UPDATE warehouse.warehouses
SET
    name_en = $2,
    name_ar = $3,
    warehouse_type = $4,
    is_active = $5,
    address_line1 = $6,
    city = $7,
    postal_code = $8,
    country_code = $9,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListBinsByWarehouse :many
SELECT
    id,
    warehouse_id,
    code,
    aisle,
    rack,
    shelf,
    bin,
    is_default,
    is_active,
    created_at,
    updated_at
FROM warehouse.bin_locations
WHERE warehouse_id = $1 AND is_active = true
ORDER BY code ASC;

-- name: GetBinByID :one
SELECT
    id,
    warehouse_id,
    code,
    aisle,
    rack,
    shelf,
    bin,
    is_default,
    is_active,
    created_at,
    updated_at
FROM warehouse.bin_locations
WHERE id = $1;

-- name: GetDefaultBinByWarehouse :one
SELECT
    id,
    warehouse_id,
    code,
    aisle,
    rack,
    shelf,
    bin,
    is_default,
    is_active,
    created_at,
    updated_at
FROM warehouse.bin_locations
WHERE warehouse_id = $1 AND is_default = true
LIMIT 1;

-- name: CreateBin :one
INSERT INTO warehouse.bin_locations (
    warehouse_id,
    code,
    aisle,
    rack,
    shelf,
    bin,
    is_default,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;
