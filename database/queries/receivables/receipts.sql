-- Receipts Queries
-- name: CreateReceipt :one
INSERT INTO receivables.receipts (
    legal_entity_id, receipt_number, customer_id, receipt_date, payment_method,
    deposit_account_id, amount, currency_code, reference_number, status,
    notes, allocated_amount, unallocated_amount, journal_entry_id, created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15
)
RETURNING *;

-- name: GetReceiptByID :one
SELECT 
    r.*,
    c.customer_number,
    c.name_ar AS customer_name_ar,
    c.name_en AS customer_name_en,
    a.code AS deposit_account_code,
    a.name_ar AS deposit_account_name_ar,
    a.name_en AS deposit_account_name_en
FROM receivables.receipts r
JOIN receivables.customers c ON r.customer_id = c.id
JOIN accounting.accounts a ON r.deposit_account_id = a.id
WHERE r.id = $1;

-- name: GetReceiptByNumber :one
SELECT * FROM receivables.receipts
WHERE legal_entity_id = $1 AND receipt_number = $2;

-- name: ListReceipts :many
SELECT 
    r.*,
    c.customer_number,
    c.name_ar AS customer_name_ar,
    c.name_en AS customer_name_en,
    a.code AS deposit_account_code,
    a.name_ar AS deposit_account_name_ar
FROM receivables.receipts r
JOIN receivables.customers c ON r.customer_id = c.id
JOIN accounting.accounts a ON r.deposit_account_id = a.id
WHERE r.legal_entity_id = $1
  AND ($2::uuid IS NULL OR r.customer_id = $2)
  AND ($3::text = 'ALL' OR r.status = $3)
  AND ($4::text = 'ALL' OR r.payment_method = $4)
  AND (
      $5::text = ''
      OR r.receipt_number ILIKE '%' || $5 || '%'
      OR r.reference_number ILIKE '%' || $5 || '%'
      OR c.name_ar ILIKE '%' || $5 || '%'
      OR c.customer_number ILIKE '%' || $5 || '%'
  )
ORDER BY r.receipt_date DESC, r.created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountReceipts :one
SELECT COUNT(*) FROM receivables.receipts r
JOIN receivables.customers c ON r.customer_id = c.id
WHERE r.legal_entity_id = $1
  AND ($2::uuid IS NULL OR r.customer_id = $2)
  AND ($3::text = 'ALL' OR r.status = $3)
  AND ($4::text = 'ALL' OR r.payment_method = $4)
  AND (
      $5::text = ''
      OR r.receipt_number ILIKE '%' || $5 || '%'
      OR r.reference_number ILIKE '%' || $5 || '%'
      OR c.name_ar ILIKE '%' || $5 || '%'
      OR c.customer_number ILIKE '%' || $5 || '%'
  );

-- name: UpdateReceiptStatus :one
UPDATE receivables.receipts
SET status = $2
WHERE id = $1
RETURNING *;

-- name: UpdateReceiptAllocationAmounts :one
UPDATE receivables.receipts
SET 
    allocated_amount = $2,
    unallocated_amount = $3
WHERE id = $1
RETURNING *;

-- Receipt Allocations Queries
-- name: CreateReceiptAllocation :one
INSERT INTO receivables.receipt_allocations (
    receipt_id, open_item_id, allocated_amount, created_by_user_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListReceiptAllocations :many
SELECT 
    ra.*,
    oi.document_type,
    oi.document_number,
    oi.due_date,
    oi.original_amount,
    oi.remaining_amount
FROM receivables.receipt_allocations ra
JOIN receivables.open_items oi ON ra.open_item_id = oi.id
WHERE ra.receipt_id = $1
ORDER BY ra.allocated_at ASC;

-- Open Items Queries
-- name: CreateOpenItem :one
INSERT INTO receivables.open_items (
    legal_entity_id, customer_id, document_type, document_number, document_id,
    document_date, due_date, currency_code, original_amount, allocated_amount,
    remaining_amount, status
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12
)
RETURNING *;

-- name: GetOpenItemByID :one
SELECT * FROM receivables.open_items
WHERE id = $1;

-- name: LockOpenItem :one
SELECT * FROM receivables.open_items
WHERE id = $1
FOR UPDATE;

-- name: ListOpenItemsByCustomer :many
SELECT 
    oi.*,
    c.customer_number,
    c.name_ar AS customer_name_ar,
    (CURRENT_DATE - oi.due_date)::integer AS days_overdue
FROM receivables.open_items oi
JOIN receivables.customers c ON oi.customer_id = c.id
WHERE oi.customer_id = $1
  AND ($2::boolean = false OR oi.status IN ('OPEN', 'PARTIALLY_PAID'))
ORDER BY oi.due_date ASC;

-- name: ListOpenItemsByLegalEntity :many
SELECT 
    oi.*,
    c.customer_number,
    c.name_ar AS customer_name_ar,
    (CURRENT_DATE - oi.due_date)::integer AS days_overdue
FROM receivables.open_items oi
JOIN receivables.customers c ON oi.customer_id = c.id
WHERE oi.legal_entity_id = $1
  AND ($2::boolean = false OR oi.status IN ('OPEN', 'PARTIALLY_PAID'))
ORDER BY oi.due_date ASC
LIMIT $3 OFFSET $4;

-- name: UpdateOpenItemAllocation :one
UPDATE receivables.open_items
SET 
    allocated_amount = $2,
    remaining_amount = $3,
    status = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetCustomerAgingSummary :one
SELECT 
    COALESCE(SUM(remaining_amount), 0.0000)::numeric AS total_outstanding,
    COALESCE(SUM(CASE WHEN due_date >= CURRENT_DATE THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS current_amount,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date BETWEEN 1 AND 30 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date BETWEEN 31 AND 60 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date BETWEEN 61 AND 90 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date > 90 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_over_90
FROM receivables.open_items
WHERE customer_id = $1
  AND status IN ('OPEN', 'PARTIALLY_PAID');
