-- name: GetItemTaxProfile :one
SELECT
    itp.id,
    itp.item_id,
    itp.tax_code_id,
    itp.exemption_reason_code,
    itp.zero_rate_reason_code,
    itp.is_active,
    itp.created_at,
    tc.code AS tax_code,
    tc.name_ar AS tax_name_ar,
    tc.tax_category_code
FROM products.item_tax_profiles itp
JOIN tax.tax_codes tc ON tc.id = itp.tax_code_id
WHERE itp.item_id = $1;

-- name: UpsertItemTaxProfile :one
INSERT INTO products.item_tax_profiles (
    item_id, tax_code_id, exemption_reason_code, zero_rate_reason_code, is_active
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (item_id) DO UPDATE SET
    tax_code_id = EXCLUDED.tax_code_id,
    exemption_reason_code = EXCLUDED.exemption_reason_code,
    zero_rate_reason_code = EXCLUDED.zero_rate_reason_code,
    is_active = EXCLUDED.is_active
RETURNING *;

-- name: DeleteItemTaxProfile :exec
DELETE FROM products.item_tax_profiles
WHERE item_id = $1;
