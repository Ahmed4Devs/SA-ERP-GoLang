-- name: ListTaxCategories :many
SELECT code, name_en, name_ar, description_en, description_ar, is_active, created_at
FROM tax.tax_categories
WHERE is_active = true
ORDER BY code ASC;

-- name: GetTaxCategoryByCode :one
SELECT code, name_en, name_ar, description_en, description_ar, is_active, created_at
FROM tax.tax_categories
WHERE code = $1;

-- name: ListExemptionReasons :many
SELECT code, category_code, name_en, name_ar, description_en, description_ar, is_active, created_at
FROM tax.exemption_reasons
WHERE is_active = true
ORDER BY code ASC;

-- name: GetExemptionReasonByCode :one
SELECT code, category_code, name_en, name_ar, description_en, description_ar, is_active, created_at
FROM tax.exemption_reasons
WHERE code = $1;

-- name: ListZeroRateReasons :many
SELECT code, category_code, name_en, name_ar, description_en, description_ar, is_active, created_at
FROM tax.zero_rate_reasons
WHERE is_active = true
ORDER BY code ASC;

-- name: GetZeroRateReasonByCode :one
SELECT code, category_code, name_en, name_ar, description_en, description_ar, is_active, created_at
FROM tax.zero_rate_reasons
WHERE code = $1;
