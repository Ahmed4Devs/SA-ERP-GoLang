-- Supplier Payments Queries
-- name: CreateSupplierPayment :one
INSERT INTO payables.payments (
    legal_entity_id, payment_number, supplier_id, payment_date,
    payment_method, disbursement_account_id, amount, currency_code,
    reference_number, status, notes, allocated_amount, unallocated_amount,
    journal_entry_id, created_by_user_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15
)
RETURNING *;

-- name: GetSupplierPaymentByID :one
SELECT 
    p.*,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar,
    s.supplier_number,
    a.code AS disbursement_account_number,
    a.name_ar AS disbursement_account_name_ar
FROM payables.payments p
JOIN payables.suppliers s ON p.supplier_id = s.id
JOIN accounting.accounts a ON p.disbursement_account_id = a.id
WHERE p.id = $1;

-- name: ListSupplierPayments :many
SELECT 
    p.*,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar,
    s.supplier_number
FROM payables.payments p
JOIN payables.suppliers s ON p.supplier_id = s.id
WHERE p.legal_entity_id = $1
  AND ($2::uuid IS NULL OR p.supplier_id = $2)
  AND (
      $3::text = '' 
      OR p.payment_number ILIKE '%' || $3 || '%'
      OR p.reference_number ILIKE '%' || $3 || '%'
      OR s.name_en ILIKE '%' || $3 || '%'
      OR s.name_ar ILIKE '%' || $3 || '%'
  )
ORDER BY p.payment_date DESC, p.created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountSupplierPayments :one
SELECT COUNT(*) FROM payables.payments p
JOIN payables.suppliers s ON p.supplier_id = s.id
WHERE p.legal_entity_id = $1
  AND ($2::uuid IS NULL OR p.supplier_id = $2)
  AND (
      $3::text = '' 
      OR p.payment_number ILIKE '%' || $3 || '%'
      OR p.reference_number ILIKE '%' || $3 || '%'
      OR s.name_en ILIKE '%' || $3 || '%'
      OR s.name_ar ILIKE '%' || $3 || '%'
  );

-- name: UpdateSupplierPaymentStatus :exec
UPDATE payables.payments
SET 
    status = $2,
    allocated_amount = $3,
    unallocated_amount = $4,
    journal_entry_id = COALESCE($5, journal_entry_id)
WHERE id = $1;

-- Payment Allocations Queries
-- name: CreateSupplierPaymentAllocation :one
INSERT INTO payables.payment_allocations (
    payment_id, open_item_id, allocated_amount
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: ListSupplierPaymentAllocations :many
SELECT 
    pa.*,
    oi.document_type,
    oi.document_number,
    oi.due_date,
    oi.original_amount,
    oi.remaining_amount
FROM payables.payment_allocations pa
JOIN payables.open_items oi ON pa.open_item_id = oi.id
WHERE pa.payment_id = $1
ORDER BY pa.allocated_at ASC;

-- Open Items Queries
-- name: CreatePayablesOpenItem :one
INSERT INTO payables.open_items (
    legal_entity_id, supplier_id, document_type, document_number, document_id,
    document_date, due_date, currency_code, original_amount, allocated_amount,
    remaining_amount, status
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12
)
RETURNING *;

-- name: GetPayablesOpenItemByID :one
SELECT * FROM payables.open_items
WHERE id = $1;

-- name: LockPayablesOpenItem :one
SELECT * FROM payables.open_items
WHERE id = $1
FOR UPDATE;

-- name: ListOpenItemsBySupplier :many
SELECT 
    oi.*,
    s.supplier_number,
    s.name_ar AS supplier_name_ar,
    (CURRENT_DATE - oi.due_date)::integer AS days_overdue
FROM payables.open_items oi
JOIN payables.suppliers s ON oi.supplier_id = s.id
WHERE oi.supplier_id = $1
  AND ($2::boolean = false OR oi.status IN ('OPEN', 'PARTIALLY_PAID'))
ORDER BY oi.due_date ASC;

-- name: ListOpenItemsByLegalEntityPayables :many
SELECT 
    oi.*,
    s.supplier_number,
    s.name_ar AS supplier_name_ar,
    (CURRENT_DATE - oi.due_date)::integer AS days_overdue
FROM payables.open_items oi
JOIN payables.suppliers s ON oi.supplier_id = s.id
WHERE oi.legal_entity_id = $1
  AND ($2::boolean = false OR oi.status IN ('OPEN', 'PARTIALLY_PAID'))
ORDER BY oi.due_date ASC
LIMIT $3 OFFSET $4;

-- name: UpdatePayablesOpenItemAllocation :one
UPDATE payables.open_items
SET 
    allocated_amount = $2,
    remaining_amount = $3,
    status = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetSupplierAgingSummary :one
SELECT 
    COALESCE(SUM(remaining_amount), 0.0000)::numeric AS total_outstanding,
    COALESCE(SUM(CASE WHEN due_date >= CURRENT_DATE THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS current_amount,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date BETWEEN 1 AND 30 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_1_30,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date BETWEEN 31 AND 60 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_31_60,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date BETWEEN 61 AND 90 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_61_90,
    COALESCE(SUM(CASE WHEN CURRENT_DATE - due_date > 90 THEN remaining_amount ELSE 0 END), 0.0000)::numeric AS overdue_over_90
FROM payables.open_items
WHERE supplier_id = $1
  AND status IN ('OPEN', 'PARTIALLY_PAID');
