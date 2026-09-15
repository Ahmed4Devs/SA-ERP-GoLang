-- name: CreateTaxTransaction :one
INSERT INTO tax.tax_transactions (
    legal_entity_id,
    tax_period_id,
    source_module,
    source_type,
    source_id,
    transaction_number,
    transaction_date,
    transaction_type,
    partner_id,
    partner_name,
    partner_tax_number,
    total_taxable_amount,
    total_tax_amount,
    total_gross_amount,
    journal_entry_id,
    status,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: CreateTaxTransactionLine :one
INSERT INTO tax.tax_transaction_lines (
    tax_transaction_id,
    line_number,
    tax_code_id,
    tax_category_code,
    tax_rate_percent,
    taxable_amount,
    tax_amount,
    gross_amount,
    exemption_reason_code,
    zero_rate_reason_code,
    is_recoverable
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetTaxTransactionByID :one
SELECT
    id,
    legal_entity_id,
    tax_period_id,
    source_module,
    source_type,
    source_id,
    transaction_number,
    transaction_date,
    transaction_type,
    partner_id,
    partner_name,
    partner_tax_number,
    total_taxable_amount,
    total_tax_amount,
    total_gross_amount,
    journal_entry_id,
    status,
    created_by,
    created_at
FROM tax.tax_transactions
WHERE id = $1;

-- name: GetTaxTransactionBySource :one
SELECT
    id,
    legal_entity_id,
    tax_period_id,
    source_module,
    source_type,
    source_id,
    transaction_number,
    transaction_date,
    transaction_type,
    partner_id,
    partner_name,
    partner_tax_number,
    total_taxable_amount,
    total_tax_amount,
    total_gross_amount,
    journal_entry_id,
    status,
    created_by,
    created_at
FROM tax.tax_transactions
WHERE legal_entity_id = $1
  AND source_module = $2
  AND source_type = $3
  AND source_id = $4;

-- name: ListTaxTransactionsByEntity :many
SELECT
    id,
    legal_entity_id,
    tax_period_id,
    source_module,
    source_type,
    source_id,
    transaction_number,
    transaction_date,
    transaction_type,
    partner_id,
    partner_name,
    partner_tax_number,
    total_taxable_amount,
    total_tax_amount,
    total_gross_amount,
    journal_entry_id,
    status,
    created_by,
    created_at
FROM tax.tax_transactions
WHERE legal_entity_id = $1
ORDER BY transaction_date DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTaxTransactionsByPeriod :many
SELECT
    id,
    legal_entity_id,
    tax_period_id,
    source_module,
    source_type,
    source_id,
    transaction_number,
    transaction_date,
    transaction_type,
    partner_id,
    partner_name,
    partner_tax_number,
    total_taxable_amount,
    total_tax_amount,
    total_gross_amount,
    journal_entry_id,
    status,
    created_by,
    created_at
FROM tax.tax_transactions
WHERE tax_period_id = $1
ORDER BY transaction_date ASC;

-- name: ListTaxTransactionLines :many
SELECT
    ttl.id,
    ttl.tax_transaction_id,
    ttl.line_number,
    ttl.tax_code_id,
    ttl.tax_category_code,
    ttl.tax_rate_percent,
    ttl.taxable_amount,
    ttl.tax_amount,
    ttl.gross_amount,
    ttl.exemption_reason_code,
    ttl.zero_rate_reason_code,
    ttl.is_recoverable,
    ttl.created_at,
    tc.code AS tax_code,
    tc.name_ar AS tax_name_ar,
    tc.name_en AS tax_name_en
FROM tax.tax_transaction_lines ttl
JOIN tax.tax_codes tc ON tc.id = ttl.tax_code_id
WHERE ttl.tax_transaction_id = $1
ORDER BY ttl.line_number ASC;

-- name: UpdateTaxTransactionStatus :one
UPDATE tax.tax_transactions
SET
    status = $2
WHERE id = $1
RETURNING *;
