-- Supplier Invoices Queries
-- name: CreateSupplierInvoice :one
INSERT INTO payables.supplier_invoices (
    legal_entity_id, invoice_number, supplier_invoice_number, supplier_id,
    invoice_date, due_date, currency_code, exchange_rate,
    subtotal, discount_amount, taxable_amount, tax_amount, total_amount,
    status, payment_status, notes, created_by_user_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15, $16, $17
)
RETURNING *;

-- name: GetSupplierInvoiceByID :one
SELECT 
    si.*,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar,
    s.tax_id AS supplier_tax_id,
    s.cr_number AS supplier_cr_number,
    s.iban AS supplier_iban,
    s.default_liability_account_id AS supplier_liability_account_id
FROM payables.supplier_invoices si
JOIN payables.suppliers s ON si.supplier_id = s.id
WHERE si.id = $1;

-- name: GetSupplierInvoiceByNumber :one
SELECT * FROM payables.supplier_invoices
WHERE legal_entity_id = $1 AND invoice_number = $2;

-- name: ListSupplierInvoices :many
SELECT 
    si.*,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar,
    s.tax_id AS supplier_tax_id
FROM payables.supplier_invoices si
JOIN payables.suppliers s ON si.supplier_id = s.id
WHERE si.legal_entity_id = $1
  AND (
      $2::text = '' 
      OR si.invoice_number ILIKE '%' || $2 || '%'
      OR si.supplier_invoice_number ILIKE '%' || $2 || '%'
      OR s.name_en ILIKE '%' || $2 || '%'
      OR s.name_ar ILIKE '%' || $2 || '%'
  )
  AND ($3::uuid IS NULL OR si.supplier_id = $3)
  AND ($4::text = 'ALL' OR si.status = $4)
  AND ($5::text = 'ALL' OR si.payment_status = $5)
ORDER BY si.invoice_date DESC, si.created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountSupplierInvoices :one
SELECT COUNT(*) FROM payables.supplier_invoices si
JOIN payables.suppliers s ON si.supplier_id = s.id
WHERE si.legal_entity_id = $1
  AND (
      $2::text = '' 
      OR si.invoice_number ILIKE '%' || $2 || '%'
      OR si.supplier_invoice_number ILIKE '%' || $2 || '%'
      OR s.name_en ILIKE '%' || $2 || '%'
      OR s.name_ar ILIKE '%' || $2 || '%'
  )
  AND ($3::uuid IS NULL OR si.supplier_id = $3)
  AND ($4::text = 'ALL' OR si.status = $4)
  AND ($5::text = 'ALL' OR si.payment_status = $5);

-- name: UpdateSupplierInvoiceStatus :exec
UPDATE payables.supplier_invoices
SET 
    status = $2,
    journal_entry_id = COALESCE($3, journal_entry_id),
    updated_at = now()
WHERE id = $1;

-- name: UpdateSupplierInvoicePaymentStatus :exec
UPDATE payables.supplier_invoices
SET 
    payment_status = $2,
    updated_at = now()
WHERE id = $1;

-- name: CreateSupplierInvoiceLine :one
INSERT INTO payables.supplier_invoice_lines (
    invoice_id, line_number, item_id, expense_account_id,
    description, quantity, unit_price,
    discount_percent, discount_amount,
    tax_rate, tax_amount, subtotal, total_amount
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9,
    $10, $11, $12, $13
)
RETURNING *;

-- name: ListSupplierInvoiceLines :many
SELECT 
    sil.*,
    p.code AS item_code,
    p.name_en AS item_name_en,
    p.name_ar AS item_name_ar,
    a.code AS expense_account_number,
    a.name_ar AS expense_account_name_ar
FROM payables.supplier_invoice_lines sil
LEFT JOIN products.items p ON sil.item_id = p.id
JOIN accounting.accounts a ON sil.expense_account_id = a.id
WHERE sil.invoice_id = $1
ORDER BY sil.line_number ASC;

-- Supplier Credit Notes
-- name: CreateSupplierCreditNote :one
INSERT INTO payables.supplier_credit_notes (
    legal_entity_id, credit_note_number, supplier_credit_note_number,
    supplier_id, original_invoice_id, credit_note_date, currency_code,
    subtotal, tax_amount, total_amount, reason, status, journal_entry_id, created_by_user_id
) VALUES (
    $1, $2, $3,
    $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: GetSupplierCreditNoteByID :one
SELECT 
    scn.*,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar
FROM payables.supplier_credit_notes scn
JOIN payables.suppliers s ON scn.supplier_id = s.id
WHERE scn.id = $1;

-- name: ListSupplierCreditNotes :many
SELECT 
    scn.*,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar
FROM payables.supplier_credit_notes scn
JOIN payables.suppliers s ON scn.supplier_id = s.id
WHERE scn.legal_entity_id = $1
  AND ($2::uuid IS NULL OR scn.supplier_id = $2)
ORDER BY scn.credit_note_date DESC, scn.created_at DESC;

-- name: CreateSupplierCreditNoteLine :one
INSERT INTO payables.supplier_credit_note_lines (
    credit_note_id, line_number, item_id, expense_account_id,
    description, quantity, unit_price, tax_rate, tax_amount, subtotal, total_amount
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: ListSupplierCreditNoteLines :many
SELECT * FROM payables.supplier_credit_note_lines
WHERE credit_note_id = $1
ORDER BY line_number ASC;
