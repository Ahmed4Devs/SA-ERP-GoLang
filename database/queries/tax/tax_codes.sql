-- name: ListTaxCodesByEntity :many
SELECT
    tc.id,
    tc.legal_entity_id,
    tc.code,
    tc.name_en,
    tc.name_ar,
    tc.tax_category_code,
    tc.tax_type,
    tc.sales_tax_account_id,
    tc.purchase_tax_account_id,
    tc.is_recoverable,
    tc.is_active,
    tc.created_at,
    tc.updated_at,
    tc.version,
    cat.name_en AS category_name_en,
    cat.name_ar AS category_name_ar,
    cur_rate.rate_percent AS current_rate_percent
FROM tax.tax_codes tc
JOIN tax.tax_categories cat ON cat.code = tc.tax_category_code
LEFT JOIN LATERAL (
    SELECT rate_percent
    FROM tax.tax_rate_versions trv
    WHERE trv.tax_code_id = tc.id
      AND trv.effective_from <= CURRENT_DATE
      AND (trv.effective_to IS NULL OR trv.effective_to >= CURRENT_DATE)
    ORDER BY trv.effective_from DESC
    LIMIT 1
) cur_rate ON true
WHERE tc.legal_entity_id = $1 AND tc.is_active = true
ORDER BY tc.code ASC;

-- name: GetTaxCodeByID :one
SELECT
    tc.id,
    tc.legal_entity_id,
    tc.code,
    tc.name_en,
    tc.name_ar,
    tc.tax_category_code,
    tc.tax_type,
    tc.sales_tax_account_id,
    tc.purchase_tax_account_id,
    tc.is_recoverable,
    tc.is_active,
    tc.created_at,
    tc.updated_at,
    tc.version
FROM tax.tax_codes tc
WHERE tc.id = $1;

-- name: GetTaxCodeByEntityAndCode :one
SELECT
    tc.id,
    tc.legal_entity_id,
    tc.code,
    tc.name_en,
    tc.name_ar,
    tc.tax_category_code,
    tc.tax_type,
    tc.sales_tax_account_id,
    tc.purchase_tax_account_id,
    tc.is_recoverable,
    tc.is_active,
    tc.created_at,
    tc.updated_at,
    tc.version
FROM tax.tax_codes tc
WHERE tc.legal_entity_id = $1 AND tc.code = $2;

-- name: CreateTaxCode :one
INSERT INTO tax.tax_codes (
    legal_entity_id,
    code,
    name_en,
    name_ar,
    tax_category_code,
    tax_type,
    sales_tax_account_id,
    purchase_tax_account_id,
    is_recoverable,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateTaxCode :one
UPDATE tax.tax_codes
SET
    name_en = $2,
    name_ar = $3,
    tax_type = $4,
    sales_tax_account_id = $5,
    purchase_tax_account_id = $6,
    is_recoverable = $7,
    is_active = $8,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;

-- name: CreateTaxRateVersion :one
INSERT INTO tax.tax_rate_versions (
    tax_code_id,
    rate_percent,
    effective_from,
    effective_to,
    description,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetActiveTaxRateVersion :one
SELECT
    id,
    tax_code_id,
    rate_percent,
    effective_from,
    effective_to,
    description,
    created_by,
    created_at
FROM tax.tax_rate_versions
WHERE tax_code_id = $1
  AND effective_from <= $2
  AND (effective_to IS NULL OR effective_to >= $2)
ORDER BY effective_from DESC
LIMIT 1;

-- name: ListTaxRateVersionsByCode :many
SELECT
    id,
    tax_code_id,
    rate_percent,
    effective_from,
    effective_to,
    description,
    created_by,
    created_at
FROM tax.tax_rate_versions
WHERE tax_code_id = $1
ORDER BY effective_from DESC;
