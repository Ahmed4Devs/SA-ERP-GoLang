-- name: GetInventoryValuationReport :many
SELECT
    item_id,
    item_code,
    item_name_en,
    item_name_ar,
    warehouse_id,
    warehouse_code,
    warehouse_name_en,
    warehouse_name_ar,
    quantity_on_hand,
    quantity_reserved,
    quantity_available,
    average_cost,
    total_value
FROM reporting.inventory_valuation
WHERE legal_entity_id = $1
  AND ($2::UUID IS NULL OR warehouse_id = $2::UUID)
ORDER BY item_code ASC;

-- name: GetSalesPeriodReport :many
SELECT
    i.id AS invoice_id,
    i.invoice_number,
    i.issue_date,
    i.due_date,
    i.customer_name_en,
    i.customer_name_ar,
    i.subtotal,
    i.tax_amount,
    i.total_amount,
    i.status
FROM sales.invoices i
WHERE i.legal_entity_id = $1
  AND ($2::DATE IS NULL OR i.issue_date >= $2::DATE)
  AND ($3::DATE IS NULL OR i.issue_date <= $3::DATE)
  AND ($4::TEXT IS NULL OR i.status = $4::TEXT)
ORDER BY i.issue_date DESC, i.invoice_number DESC;

-- name: GetPurchasingPeriodReport :many
SELECT
    po.id AS po_id,
    po.po_number,
    po.order_date,
    s.supplier_number,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar,
    po.subtotal,
    po.tax_amount,
    po.total_amount,
    po.status
FROM purchasing.purchase_orders po
JOIN payables.suppliers s ON s.id = po.supplier_id
WHERE po.legal_entity_id = $1
  AND ($2::DATE IS NULL OR po.order_date >= $2::DATE)
  AND ($3::DATE IS NULL OR po.order_date <= $3::DATE)
  AND ($4::TEXT IS NULL OR po.status = $4::TEXT)
ORDER BY po.order_date DESC, po.po_number DESC;

-- name: GetVATSummaryReport :one
SELECT
    -- Output VAT (Sales)
    COALESCE(SUM(CASE WHEN tt.transaction_type = 'OUTPUT' THEN tt.total_taxable_amount ELSE 0 END), 0)::NUMERIC(18, 4) AS standard_rated_sales,
    COALESCE(SUM(CASE WHEN tt.transaction_type = 'OUTPUT' THEN tt.total_tax_amount ELSE 0 END), 0)::NUMERIC(18, 4) AS output_vat_amount,
    -- Input VAT (Purchases)
    COALESCE(SUM(CASE WHEN tt.transaction_type = 'INPUT' THEN tt.total_taxable_amount ELSE 0 END), 0)::NUMERIC(18, 4) AS standard_rated_purchases,
    COALESCE(SUM(CASE WHEN tt.transaction_type = 'INPUT' THEN tt.total_tax_amount ELSE 0 END), 0)::NUMERIC(18, 4) AS input_vat_amount,
    -- Net VAT Payable / (Refundable)
    (COALESCE(SUM(CASE WHEN tt.transaction_type = 'OUTPUT' THEN tt.total_tax_amount ELSE 0 END), 0) -
     COALESCE(SUM(CASE WHEN tt.transaction_type = 'INPUT' THEN tt.total_tax_amount ELSE 0 END), 0))::NUMERIC(18, 4) AS net_vat_payable
FROM tax.tax_transactions tt
WHERE tt.legal_entity_id = $1
  AND ($2::DATE IS NULL OR tt.transaction_date >= $2::DATE)
  AND ($3::DATE IS NULL OR tt.transaction_date <= $3::DATE);

-- name: GetPayrollRunsReport :many
SELECT
    run_id,
    run_number,
    period_id,
    period_name_en,
    period_name_ar,
    run_date,
    status,
    employee_count,
    total_basic,
    total_allowances,
    total_gross,
    total_deductions,
    total_net
FROM reporting.payroll_summary
WHERE legal_entity_id = $1
  AND ($2::DATE IS NULL OR run_date >= $2::DATE)
  AND ($3::DATE IS NULL OR run_date <= $3::DATE)
ORDER BY run_date DESC;

-- name: GetGOSIContributionsReport :many
SELECT
    c.id AS calculation_id,
    c.calculation_year,
    c.calculation_month,
    c.status,
    e.employee_number,
    e.first_name_en || ' ' || e.last_name_en AS employee_name_en,
    e.first_name_ar || ' ' || e.last_name_ar AS employee_name_ar,
    e.is_saudi_national,
    c.contribution_basis,
    c.total_employee_amount,
    c.total_employer_amount,
    c.total_amount,
    c.created_at
FROM gosi.contribution_calculations c
JOIN hr.employees e ON e.id = c.employee_id
WHERE c.legal_entity_id = $1
  AND ($2::INT IS NULL OR c.calculation_year = $2::INT)
  AND ($3::INT IS NULL OR c.calculation_month = $3::INT)
ORDER BY c.calculation_year DESC, c.calculation_month DESC, c.created_at DESC;

-- name: GetBudgetVsActualReport :many
SELECT
    b.id AS budget_id,
    b.name_en AS budget_name_en,
    b.name_ar AS budget_name_ar,
    fy.name_en AS fiscal_year_name,
    bl.account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    bl.amount AS budgeted_amount,
    COALESCE(act.actual_amount, 0)::NUMERIC(18, 4) AS actual_amount,
    (bl.amount - COALESCE(act.actual_amount, 0))::NUMERIC(18, 4) AS variance_amount,
    CASE
        WHEN bl.amount != 0 THEN
            ROUND(((bl.amount - COALESCE(act.actual_amount, 0)) / bl.amount) * 100, 2)::NUMERIC(18, 2)
        ELSE 0.00
    END AS variance_percentage
FROM budgeting.budget_lines bl
JOIN budgeting.budget_versions bv ON bv.id = bl.budget_version_id
JOIN budgeting.budgets b ON b.id = bv.budget_id
JOIN accounting.fiscal_years fy ON fy.id = b.fiscal_year_id
JOIN accounting.accounts a ON a.id = bl.account_id
LEFT JOIN (
    SELECT
        jl.account_id,
        CASE
            WHEN acc.normal_balance = 'DEBIT' THEN (SUM(jl.base_debit) - SUM(jl.base_credit))
            ELSE (SUM(jl.base_credit) - SUM(jl.base_debit))
        END AS actual_amount
    FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    JOIN accounting.accounts acc ON acc.id = jl.account_id
    WHERE je.legal_entity_id = $1
      AND je.status = 'POSTED'
    GROUP BY jl.account_id, acc.normal_balance
) act ON act.account_id = bl.account_id
WHERE b.legal_entity_id = $1
  AND ($2::UUID IS NULL OR b.id = $2::UUID)
ORDER BY a.code ASC;
