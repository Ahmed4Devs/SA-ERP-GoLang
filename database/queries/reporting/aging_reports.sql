-- name: GetCustomerAgingReport :many
SELECT
    customer_id,
    customer_number,
    customer_name_en,
    customer_name_ar,
    COUNT(open_item_id)::INT AS invoice_count,
    SUM(remaining_amount)::NUMERIC(18, 4) AS total_outstanding,
    SUM(current_amount)::NUMERIC(18, 4) AS current_amount,
    SUM(aging_1_30)::NUMERIC(18, 4) AS aging_1_30,
    SUM(aging_31_60)::NUMERIC(18, 4) AS aging_31_60,
    SUM(aging_61_90)::NUMERIC(18, 4) AS aging_61_90,
    SUM(aging_91_120)::NUMERIC(18, 4) AS aging_91_120,
    SUM(aging_over_120)::NUMERIC(18, 4) AS aging_over_120
FROM reporting.customer_aging
WHERE legal_entity_id = $1
  AND ($2::UUID IS NULL OR customer_id = $2::UUID)
GROUP BY customer_id, customer_number, customer_name_en, customer_name_ar
ORDER BY total_outstanding DESC;

-- name: GetSupplierAgingReport :many
SELECT
    supplier_id,
    supplier_number,
    supplier_name_en,
    supplier_name_ar,
    COUNT(open_item_id)::INT AS bill_count,
    SUM(remaining_amount)::NUMERIC(18, 4) AS total_outstanding,
    SUM(current_amount)::NUMERIC(18, 4) AS current_amount,
    SUM(aging_1_30)::NUMERIC(18, 4) AS aging_1_30,
    SUM(aging_31_60)::NUMERIC(18, 4) AS aging_31_60,
    SUM(aging_61_90)::NUMERIC(18, 4) AS aging_61_90,
    SUM(aging_91_120)::NUMERIC(18, 4) AS aging_91_120,
    SUM(aging_over_120)::NUMERIC(18, 4) AS aging_over_120
FROM reporting.supplier_aging
WHERE legal_entity_id = $1
  AND ($2::UUID IS NULL OR supplier_id = $2::UUID)
GROUP BY supplier_id, supplier_number, supplier_name_en, supplier_name_ar
ORDER BY total_outstanding DESC;

-- name: GetCustomerStatementLines :many
SELECT
    oi.id AS item_id,
    oi.document_type,
    oi.document_number,
    oi.document_date,
    oi.due_date,
    oi.currency_code,
    oi.original_amount,
    oi.allocated_amount,
    oi.remaining_amount,
    oi.status
FROM receivables.open_items oi
WHERE oi.legal_entity_id = $1
  AND oi.customer_id = $2
  AND ($3::DATE IS NULL OR oi.document_date >= $3::DATE)
  AND ($4::DATE IS NULL OR oi.document_date <= $4::DATE)
ORDER BY oi.document_date ASC, oi.document_number ASC;

-- name: GetSupplierStatementLines :many
SELECT
    oi.id AS item_id,
    oi.document_type,
    oi.document_number,
    oi.document_date,
    oi.due_date,
    oi.currency_code,
    oi.original_amount,
    oi.allocated_amount,
    oi.remaining_amount,
    oi.status
FROM payables.open_items oi
WHERE oi.legal_entity_id = $1
  AND oi.supplier_id = $2
  AND ($3::DATE IS NULL OR oi.document_date >= $3::DATE)
  AND ($4::DATE IS NULL OR oi.document_date <= $4::DATE)
ORDER BY oi.document_date ASC, oi.document_number ASC;
