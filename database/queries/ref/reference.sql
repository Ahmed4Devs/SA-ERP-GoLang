-- name: GetCurrency :one
SELECT code, name_en, name_ar, decimal_places, is_active
FROM ref.currencies
WHERE code = $1;

-- name: ListActiveCurrencies :many
SELECT code, name_en, name_ar, decimal_places, is_active
FROM ref.currencies
WHERE is_active = TRUE
ORDER BY code ASC;

-- name: UpsertCurrency :exec
INSERT INTO ref.currencies (code, name_en, name_ar, decimal_places, is_active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (code) DO UPDATE
SET name_en = EXCLUDED.name_en,
    name_ar = EXCLUDED.name_ar,
    decimal_places = EXCLUDED.decimal_places,
    is_active = EXCLUDED.is_active;

-- name: GetCountry :one
SELECT code, name_en, name_ar, phone_prefix, is_active
FROM ref.countries
WHERE code = $1;

-- name: ListActiveCountries :many
SELECT code, name_en, name_ar, phone_prefix, is_active
FROM ref.countries
WHERE is_active = TRUE
ORDER BY code ASC;

-- name: UpsertCountry :exec
INSERT INTO ref.countries (code, name_en, name_ar, phone_prefix, is_active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (code) DO UPDATE
SET name_en = EXCLUDED.name_en,
    name_ar = EXCLUDED.name_ar,
    phone_prefix = EXCLUDED.phone_prefix,
    is_active = EXCLUDED.is_active;
