-- Payment Terms Queries
-- name: CreatePayablesPaymentTerm :one
INSERT INTO payables.payment_terms (
    legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetPayablesPaymentTermByID :one
SELECT * FROM payables.payment_terms
WHERE id = $1;

-- name: ListPayablesPaymentTerms :many
SELECT * FROM payables.payment_terms
WHERE legal_entity_id = $1 AND (is_active = true OR $2::boolean = true)
ORDER BY due_days ASC;

-- Supplier Queries
-- name: CreateSupplier :one
INSERT INTO payables.suppliers (
    legal_entity_id, supplier_number, supplier_type, name_en, name_ar,
    tax_id, cr_number, national_id, iban, bank_name, swift_code,
    building_no, street_name, district, city, postal_code, additional_no, country_code,
    email, phone, contact_person, website,
    payment_terms_id, currency_code, default_expense_account_id, default_liability_account_id, status
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10, $11,
    $12, $13, $14, $15, $16, $17, $18,
    $19, $20, $21, $22,
    $23, $24, $25, $26, $27
)
RETURNING *;

-- name: GetSupplierByID :one
SELECT 
    s.*,
    pt.code AS payment_terms_code,
    pt.name_ar AS payment_terms_name_ar,
    pt.due_days AS payment_terms_due_days,
    exp_acc.code AS default_expense_account_number,
    liab_acc.code AS default_liability_account_number
FROM payables.suppliers s
LEFT JOIN payables.payment_terms pt ON s.payment_terms_id = pt.id
LEFT JOIN accounting.accounts exp_acc ON s.default_expense_account_id = exp_acc.id
LEFT JOIN accounting.accounts liab_acc ON s.default_liability_account_id = liab_acc.id
WHERE s.id = $1;

-- name: GetSupplierByNumber :one
SELECT * FROM payables.suppliers
WHERE legal_entity_id = $1 AND supplier_number = $2;

-- name: ListSuppliers :many
SELECT 
    s.*,
    pt.code AS payment_terms_code,
    pt.name_ar AS payment_terms_name_ar,
    pt.due_days AS payment_terms_due_days,
    exp_acc.code AS default_expense_account_number,
    liab_acc.code AS default_liability_account_number
FROM payables.suppliers s
LEFT JOIN payables.payment_terms pt ON s.payment_terms_id = pt.id
LEFT JOIN accounting.accounts exp_acc ON s.default_expense_account_id = exp_acc.id
LEFT JOIN accounting.accounts liab_acc ON s.default_liability_account_id = liab_acc.id
WHERE s.legal_entity_id = $1
  AND (
      $2::text = '' 
      OR s.supplier_number ILIKE '%' || $2 || '%'
      OR s.name_en ILIKE '%' || $2 || '%'
      OR s.name_ar ILIKE '%' || $2 || '%'
      OR s.tax_id ILIKE '%' || $2 || '%'
      OR s.phone ILIKE '%' || $2 || '%'
  )
  AND ($3::text = 'ALL' OR s.supplier_type = $3)
  AND ($4::text = 'ALL' OR s.status = $4)
ORDER BY s.created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountSuppliers :one
SELECT COUNT(*) FROM payables.suppliers
WHERE legal_entity_id = $1
  AND (
      $2::text = '' 
      OR supplier_number ILIKE '%' || $2 || '%'
      OR name_en ILIKE '%' || $2 || '%'
      OR name_ar ILIKE '%' || $2 || '%'
      OR tax_id ILIKE '%' || $2 || '%'
      OR phone ILIKE '%' || $2 || '%'
  )
  AND ($3::text = 'ALL' OR supplier_type = $3)
  AND ($4::text = 'ALL' OR status = $4);

-- name: UpdateSupplier :one
UPDATE payables.suppliers
SET
    name_en = $3,
    name_ar = $4,
    tax_id = $5,
    cr_number = $6,
    national_id = $7,
    iban = $8,
    bank_name = $9,
    swift_code = $10,
    building_no = $11,
    street_name = $12,
    district = $13,
    city = $14,
    postal_code = $15,
    additional_no = $16,
    country_code = $17,
    email = $18,
    phone = $19,
    contact_person = $20,
    website = $21,
    payment_terms_id = $22,
    default_expense_account_id = $23,
    default_liability_account_id = $24,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: SetSupplierStatus :exec
UPDATE payables.suppliers
SET status = $2, updated_at = now()
WHERE id = $1;

-- name: CreateSupplierAccount :one
INSERT INTO payables.supplier_accounts (
    supplier_id, gl_account_id, currency_code, is_primary
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListSupplierAccounts :many
SELECT sa.*, a.code AS account_number, a.name_en AS account_name_en, a.name_ar AS account_name_ar
FROM payables.supplier_accounts sa
JOIN accounting.accounts a ON sa.gl_account_id = a.id
WHERE sa.supplier_id = $1
ORDER BY sa.is_primary DESC, sa.created_at ASC;
