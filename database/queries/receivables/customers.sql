-- Payment Terms Queries
-- name: CreatePaymentTerm :one
INSERT INTO receivables.payment_terms (
    legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetPaymentTermByID :one
SELECT * FROM receivables.payment_terms
WHERE id = $1;

-- name: ListPaymentTerms :many
SELECT * FROM receivables.payment_terms
WHERE legal_entity_id = $1 AND (is_active = true OR $2::boolean = true)
ORDER BY due_days ASC;

-- Customer Queries
-- name: CreateCustomer :one
INSERT INTO receivables.customers (
    legal_entity_id, customer_number, customer_type, name_en, name_ar,
    tax_id, cr_number, national_id, building_no, street_name, district, city, postal_code, additional_no, country_code,
    email, phone, contact_person, payment_terms_id, currency_code, default_account_id, status
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22
)
RETURNING *;

-- name: GetCustomerByID :one
SELECT 
    c.*,
    pt.code AS payment_terms_code,
    pt.name_ar AS payment_terms_name_ar,
    pt.due_days AS payment_terms_due_days,
    cl.credit_limit,
    cl.current_balance,
    cl.max_overdue_days,
    cl.block_action,
    cl.is_credit_hold,
    cl.hold_reason
FROM receivables.customers c
LEFT JOIN receivables.payment_terms pt ON c.payment_terms_id = pt.id
LEFT JOIN receivables.customer_credit_limits cl ON c.id = cl.customer_id
WHERE c.id = $1;

-- name: GetCustomerByNumber :one
SELECT * FROM receivables.customers
WHERE legal_entity_id = $1 AND customer_number = $2;

-- name: ListCustomers :many
SELECT 
    c.*,
    pt.code AS payment_terms_code,
    pt.name_ar AS payment_terms_name_ar,
    COALESCE(cl.credit_limit, 0.0000)::numeric AS credit_limit,
    COALESCE(cl.current_balance, 0.0000)::numeric AS current_balance,
    COALESCE(cl.is_credit_hold, false) AS is_credit_hold
FROM receivables.customers c
LEFT JOIN receivables.payment_terms pt ON c.payment_terms_id = pt.id
LEFT JOIN receivables.customer_credit_limits cl ON c.id = cl.customer_id
WHERE c.legal_entity_id = $1
  AND (
      $2::text = '' 
      OR c.customer_number ILIKE '%' || $2 || '%'
      OR c.name_en ILIKE '%' || $2 || '%'
      OR c.name_ar ILIKE '%' || $2 || '%'
      OR c.tax_id ILIKE '%' || $2 || '%'
      OR c.phone ILIKE '%' || $2 || '%'
  )
  AND ($3::text = 'ALL' OR c.customer_type = $3)
  AND ($4::text = 'ALL' OR c.status = $4)
ORDER BY c.created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountCustomers :one
SELECT COUNT(*) FROM receivables.customers
WHERE legal_entity_id = $1
  AND (
      $2::text = '' 
      OR customer_number ILIKE '%' || $2 || '%'
      OR name_en ILIKE '%' || $2 || '%'
      OR name_ar ILIKE '%' || $2 || '%'
      OR tax_id ILIKE '%' || $2 || '%'
      OR phone ILIKE '%' || $2 || '%'
  )
  AND ($3::text = 'ALL' OR customer_type = $3)
  AND ($4::text = 'ALL' OR status = $4);

-- name: UpdateCustomer :one
UPDATE receivables.customers
SET
    name_en = $2,
    name_ar = $3,
    tax_id = $4,
    cr_number = $5,
    national_id = $6,
    building_no = $7,
    street_name = $8,
    district = $9,
    city = $10,
    postal_code = $11,
    additional_no = $12,
    country_code = $13,
    email = $14,
    phone = $15,
    contact_person = $16,
    payment_terms_id = $17,
    currency_code = $18,
    default_account_id = $19,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateCustomerStatus :one
UPDATE receivables.customers
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- Customer Accounts
-- name: UpsertCustomerAccount :one
INSERT INTO receivables.customer_accounts (
    customer_id, gl_account_id, currency_code, is_primary
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (customer_id, currency_code) DO UPDATE
SET gl_account_id = EXCLUDED.gl_account_id, is_primary = EXCLUDED.is_primary
RETURNING *;

-- name: GetCustomerAccounts :many
SELECT ca.*, a.code AS account_code, a.name_ar AS account_name_ar, a.name_en AS account_name_en
FROM receivables.customer_accounts ca
JOIN accounting.accounts a ON ca.gl_account_id = a.id
WHERE ca.customer_id = $1
ORDER BY ca.is_primary DESC;

-- Credit Limit Queries
-- name: UpsertCreditLimit :one
INSERT INTO receivables.customer_credit_limits (
    customer_id, credit_limit, current_balance, max_overdue_days, block_action, is_credit_hold, hold_reason
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (customer_id) DO UPDATE
SET
    credit_limit = EXCLUDED.credit_limit,
    max_overdue_days = EXCLUDED.max_overdue_days,
    block_action = EXCLUDED.block_action,
    is_credit_hold = EXCLUDED.is_credit_hold,
    hold_reason = EXCLUDED.hold_reason,
    updated_at = now()
RETURNING *;

-- name: GetCreditLimitByCustomerID :one
SELECT * FROM receivables.customer_credit_limits
WHERE customer_id = $1;

-- name: UpdateCustomerBalance :one
UPDATE receivables.customer_credit_limits
SET
    current_balance = current_balance + $2,
    updated_at = now()
WHERE customer_id = $1
RETURNING *;

-- name: SetCreditHold :one
UPDATE receivables.customer_credit_limits
SET
    is_credit_hold = $2,
    hold_reason = $3,
    updated_at = now()
WHERE customer_id = $1
RETURNING *;
