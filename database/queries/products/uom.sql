-- name: ListUnitsOfMeasure :many
SELECT id, code, name_en, name_ar, symbol_en, symbol_ar, uom_type, is_active, created_at
FROM products.units_of_measure
WHERE is_active = true
ORDER BY uom_type, code;

-- name: GetUnitOfMeasureByID :one
SELECT id, code, name_en, name_ar, symbol_en, symbol_ar, uom_type, is_active, created_at
FROM products.units_of_measure
WHERE id = $1;

-- name: GetUnitOfMeasureByCode :one
SELECT id, code, name_en, name_ar, symbol_en, symbol_ar, uom_type, is_active, created_at
FROM products.units_of_measure
WHERE code = $1;

-- name: CreateUnitOfMeasure :one
INSERT INTO products.units_of_measure (
    code, name_en, name_ar, symbol_en, symbol_ar, uom_type
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListUOMConversions :many
SELECT
    uc.id,
    uc.from_uom_id,
    uc.to_uom_id,
    uc.conversion_rate,
    uc.created_at,
    f.code AS from_uom_code,
    f.name_ar AS from_uom_name_ar,
    t.code AS to_uom_code,
    t.name_ar AS to_uom_name_ar
FROM products.uom_conversions uc
JOIN products.units_of_measure f ON f.id = uc.from_uom_id
JOIN products.units_of_measure t ON t.id = uc.to_uom_id
ORDER BY f.code, t.code;

-- name: GetUOMConversionRate :one
SELECT conversion_rate
FROM products.uom_conversions
WHERE from_uom_id = $1 AND to_uom_id = $2;

-- name: CreateUOMConversion :one
INSERT INTO products.uom_conversions (
    from_uom_id, to_uom_id, conversion_rate
) VALUES (
    $1, $2, $3
)
ON CONFLICT (from_uom_id, to_uom_id) DO UPDATE SET conversion_rate = EXCLUDED.conversion_rate
RETURNING *;
