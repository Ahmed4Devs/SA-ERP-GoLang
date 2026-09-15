-- name: CreateTaxReturn :one
INSERT INTO tax.tax_returns (
    legal_entity_id,
    tax_period_id,
    return_number,
    filing_date,
    status,
    total_sales_taxable,
    total_sales_vat,
    total_purchases_taxable,
    total_purchases_vat,
    net_vat_due,
    declaration_boxes,
    submitted_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetTaxReturnByID :one
SELECT
    id,
    legal_entity_id,
    tax_period_id,
    return_number,
    filing_date,
    status,
    total_sales_taxable,
    total_sales_vat,
    total_purchases_taxable,
    total_purchases_vat,
    net_vat_due,
    declaration_boxes,
    submitted_at,
    submitted_by,
    created_at,
    updated_at,
    version
FROM tax.tax_returns
WHERE id = $1;

-- name: GetTaxReturnByPeriod :one
SELECT
    id,
    legal_entity_id,
    tax_period_id,
    return_number,
    filing_date,
    status,
    total_sales_taxable,
    total_sales_vat,
    total_purchases_taxable,
    total_purchases_vat,
    net_vat_due,
    declaration_boxes,
    submitted_at,
    submitted_by,
    created_at,
    updated_at,
    version
FROM tax.tax_returns
WHERE legal_entity_id = $1 AND tax_period_id = $2;

-- name: ListTaxReturnsByEntity :many
SELECT
    tr.id,
    tr.legal_entity_id,
    tr.tax_period_id,
    tr.return_number,
    tr.filing_date,
    tr.status,
    tr.total_sales_taxable,
    tr.total_sales_vat,
    tr.total_purchases_taxable,
    tr.total_purchases_vat,
    tr.net_vat_due,
    tr.submitted_at,
    tr.created_at,
    tp.period_number,
    tp.fiscal_year,
    tp.period_type
FROM tax.tax_returns tr
JOIN tax.tax_periods tp ON tp.id = tr.tax_period_id
WHERE tr.legal_entity_id = $1
ORDER BY tp.fiscal_year DESC, tp.period_number DESC;

-- name: UpdateTaxReturn :one
UPDATE tax.tax_returns
SET
    status = $2,
    filing_date = $3,
    total_sales_taxable = $4,
    total_sales_vat = $5,
    total_purchases_taxable = $6,
    total_purchases_vat = $7,
    net_vat_due = $8,
    declaration_boxes = $9,
    submitted_at = $10,
    submitted_by = $11,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;

-- name: AggregateTaxPeriodSummary :many
SELECT
    tt.transaction_type,
    ttl.tax_category_code,
    ttl.tax_rate_percent,
    ttl.exemption_reason_code,
    ttl.zero_rate_reason_code,
    COALESCE(SUM(ttl.taxable_amount), 0)::NUMERIC(18,4) AS total_taxable,
    COALESCE(SUM(ttl.tax_amount), 0)::NUMERIC(18,4) AS total_tax,
    COALESCE(SUM(ttl.gross_amount), 0)::NUMERIC(18,4) AS total_gross
FROM tax.tax_transactions tt
JOIN tax.tax_transaction_lines ttl ON ttl.tax_transaction_id = tt.id
WHERE tt.tax_period_id = $1
  AND tt.status = 'POSTED'
GROUP BY
    tt.transaction_type,
    ttl.tax_category_code,
    ttl.tax_rate_percent,
    ttl.exemption_reason_code,
    ttl.zero_rate_reason_code;
