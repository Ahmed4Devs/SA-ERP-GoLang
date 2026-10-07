-- +goose Up
-- +goose StatementBegin
-- Widen stored decimals without updating, rounding, or recomputing historical
-- financial records. Existing values remain exactly equal. This migration is
-- forward-only because narrowing would destroy newly representable amounts.
-- Recreate known reporting views and preserve their owner and explicit grants.
-- DROP uses RESTRICT so unrecognized customer dependencies stop the migration
-- atomically for review instead of being silently removed.
DO $$ BEGIN
CREATE TEMP TABLE canonical_view_metadata ON COMMIT DROP AS
SELECT c.relname, pg_get_userbyid(c.relowner) AS owner_name, c.relacl
FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname='reporting' AND c.relkind='v'
    AND c.relname IN ('trial_balance','general_ledger','customer_aging','supplier_aging','inventory_valuation','sales_daily','payroll_summary');
END $$;


DROP VIEW reporting.trial_balance;

DROP VIEW reporting.general_ledger;

DROP VIEW reporting.customer_aging;

DROP VIEW reporting.supplier_aging;

DROP VIEW reporting.inventory_valuation;

DROP VIEW reporting.sales_daily;

DROP VIEW reporting.payroll_summary;

ALTER TABLE inventory.stock_balances DROP COLUMN quantity_available;

ALTER TABLE inventory.stock_count_lines DROP COLUMN variance;

ALTER TABLE accounting.account_balances
    ALTER COLUMN opening_balance TYPE NUMERIC(24,6),
    ALTER COLUMN period_debit TYPE NUMERIC(24,6),
    ALTER COLUMN period_credit TYPE NUMERIC(24,6),
    ALTER COLUMN closing_balance TYPE NUMERIC(24,6);

ALTER TABLE accounting.journal_entries
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN total_debit TYPE NUMERIC(24,6),
    ALTER COLUMN total_credit TYPE NUMERIC(24,6);

ALTER TABLE accounting.journal_lines
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN transaction_debit TYPE NUMERIC(24,6),
    ALTER COLUMN transaction_credit TYPE NUMERIC(24,6),
    ALTER COLUMN base_debit TYPE NUMERIC(24,6),
    ALTER COLUMN base_credit TYPE NUMERIC(24,6);

ALTER TABLE attendance.overtime_requests
    ALTER COLUMN overtime_hours TYPE NUMERIC(24,6),
    ALTER COLUMN rate_multiplier TYPE NUMERIC(24,6);

ALTER TABLE attendance.records
    ALTER COLUMN expected_work_hours TYPE NUMERIC(24,6),
    ALTER COLUMN actual_work_hours TYPE NUMERIC(24,6);

ALTER TABLE banking.bank_accounts
    ALTER COLUMN opening_balance TYPE NUMERIC(24,6),
    ALTER COLUMN current_balance TYPE NUMERIC(24,6),
    ALTER COLUMN available_balance TYPE NUMERIC(24,6),
    ALTER COLUMN overdraft_limit TYPE NUMERIC(24,6);

ALTER TABLE banking.reconciliation_matches
    ALTER COLUMN matched_amount TYPE NUMERIC(24,6),
    ALTER COLUMN difference_amount TYPE NUMERIC(24,6);

ALTER TABLE banking.reconciliations
    ALTER COLUMN statement_ending_balance TYPE NUMERIC(24,6),
    ALTER COLUMN gl_book_balance TYPE NUMERIC(24,6),
    ALTER COLUMN unreconciled_difference TYPE NUMERIC(24,6);

ALTER TABLE banking.statement_lines
    ALTER COLUMN amount TYPE NUMERIC(24,6),
    ALTER COLUMN balance_after TYPE NUMERIC(24,6);

ALTER TABLE banking.statements
    ALTER COLUMN opening_balance TYPE NUMERIC(24,6),
    ALTER COLUMN closing_balance TYPE NUMERIC(24,6),
    ALTER COLUMN total_credits TYPE NUMERIC(24,6),
    ALTER COLUMN total_debits TYPE NUMERIC(24,6);

ALTER TABLE budgeting.budget_lines
    ALTER COLUMN amount TYPE NUMERIC(24,6),
    ALTER COLUMN actual_amount TYPE NUMERIC(24,6),
    ALTER COLUMN committed_amount TYPE NUMERIC(24,6);

ALTER TABLE budgeting.budget_transfers
    ALTER COLUMN amount TYPE NUMERIC(24,6);

ALTER TABLE budgeting.budgets
    ALTER COLUMN total_budgeted_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_actual_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_committed_amount TYPE NUMERIC(24,6);

ALTER TABLE fixedassets.categories
    ALTER COLUMN salvage_value_rate TYPE NUMERIC(12,8);

ALTER TABLE gosi.contribution_calculation_lines
    ALTER COLUMN basis_amount TYPE NUMERIC(24,6),
    ALTER COLUMN employee_rate TYPE NUMERIC(12,8),
    ALTER COLUMN employer_rate TYPE NUMERIC(12,8),
    ALTER COLUMN employee_amount TYPE NUMERIC(24,6),
    ALTER COLUMN employer_amount TYPE NUMERIC(24,6);

ALTER TABLE gosi.contribution_calculations
    ALTER COLUMN contribution_basis TYPE NUMERIC(24,6),
    ALTER COLUMN total_employee_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_employer_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE gosi.contribution_rule_versions
    ALTER COLUMN employee_rate TYPE NUMERIC(12,8),
    ALTER COLUMN employer_rate TYPE NUMERIC(12,8),
    ALTER COLUMN minimum_basis TYPE NUMERIC(24,6),
    ALTER COLUMN maximum_basis TYPE NUMERIC(24,6);

ALTER TABLE gosi.employee_insurance_profiles
    ALTER COLUMN contribution_basis_override TYPE NUMERIC(24,6);

ALTER TABLE gosi.export_batches
    ALTER COLUMN total_employee_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_employer_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE hr.contract_versions
    ALTER COLUMN basic_salary TYPE NUMERIC(24,6),
    ALTER COLUMN housing_allowance TYPE NUMERIC(24,6),
    ALTER COLUMN transport_allowance TYPE NUMERIC(24,6),
    ALTER COLUMN other_allowances TYPE NUMERIC(24,6),
    ALTER COLUMN gross_salary TYPE NUMERIC(24,6);

ALTER TABLE hr.contracts
    ALTER COLUMN working_hours_per_week TYPE NUMERIC(24,6);

ALTER TABLE inventory.cost_layer_consumptions
    ALTER COLUMN quantity_consumed TYPE NUMERIC(24,6),
    ALTER COLUMN unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_cost TYPE NUMERIC(24,6);

ALTER TABLE inventory.cost_layers
    ALTER COLUMN original_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN remaining_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_cost TYPE NUMERIC(24,6);

ALTER TABLE inventory.inventory_revaluations
    ALTER COLUMN quantity_on_hand TYPE NUMERIC(24,6),
    ALTER COLUMN old_unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN new_unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN old_total_value TYPE NUMERIC(24,6),
    ALTER COLUMN new_total_value TYPE NUMERIC(24,6),
    ALTER COLUMN variance_amount TYPE NUMERIC(24,6);

ALTER TABLE inventory.stock_adjustment_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_cost TYPE NUMERIC(24,6);

ALTER TABLE inventory.stock_balances
    ALTER COLUMN quantity_on_hand TYPE NUMERIC(24,6),
    ALTER COLUMN quantity_reserved TYPE NUMERIC(24,6),
    ALTER COLUMN average_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_value TYPE NUMERIC(24,6);

ALTER TABLE inventory.stock_count_lines
    ALTER COLUMN system_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN counted_quantity TYPE NUMERIC(24,6);

ALTER TABLE inventory.stock_movements
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_cost TYPE NUMERIC(24,6);

ALTER TABLE inventory.stock_reservations
    ALTER COLUMN quantity TYPE NUMERIC(24,6);

ALTER TABLE inventory.transfer_order_lines
    ALTER COLUMN requested_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN dispatched_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN received_quantity TYPE NUMERIC(24,6);

ALTER TABLE leave.employee_balances
    ALTER COLUMN allocated_days TYPE NUMERIC(24,6),
    ALTER COLUMN accrued_days TYPE NUMERIC(24,6),
    ALTER COLUMN taken_days TYPE NUMERIC(24,6),
    ALTER COLUMN pending_days TYPE NUMERIC(24,6),
    ALTER COLUMN carryover_days TYPE NUMERIC(24,6),
    ALTER COLUMN closing_balance TYPE NUMERIC(24,6);

ALTER TABLE leave.policies
    ALTER COLUMN accrual_rate_per_period TYPE NUMERIC(24,6),
    ALTER COLUMN max_carryover_days TYPE NUMERIC(24,6);

ALTER TABLE leave.requests
    ALTER COLUMN total_days TYPE NUMERIC(24,6);

ALTER TABLE leave.transactions
    ALTER COLUMN days TYPE NUMERIC(24,6);

ALTER TABLE leave.types
    ALTER COLUMN max_days_per_year TYPE NUMERIC(24,6);

ALTER TABLE payables.open_items
    ALTER COLUMN original_amount TYPE NUMERIC(24,6),
    ALTER COLUMN allocated_amount TYPE NUMERIC(24,6),
    ALTER COLUMN remaining_amount TYPE NUMERIC(24,6);

ALTER TABLE payables.payment_allocations
    ALTER COLUMN allocated_amount TYPE NUMERIC(24,6);

ALTER TABLE payables.payment_terms
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8);

ALTER TABLE payables.payments
    ALTER COLUMN amount TYPE NUMERIC(24,6),
    ALTER COLUMN allocated_amount TYPE NUMERIC(24,6),
    ALTER COLUMN unallocated_amount TYPE NUMERIC(24,6);

ALTER TABLE payables.supplier_credit_note_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE payables.supplier_credit_notes
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE payables.supplier_invoice_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE payables.supplier_invoices
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE payroll.advances
    ALTER COLUMN amount TYPE NUMERIC(24,6);

ALTER TABLE payroll.calculation_lines
    ALTER COLUMN rate TYPE NUMERIC(24,6),
    ALTER COLUMN basis_amount TYPE NUMERIC(24,6),
    ALTER COLUMN amount TYPE NUMERIC(24,6);

ALTER TABLE payroll.employee_salary_assignments
    ALTER COLUMN basic_salary TYPE NUMERIC(24,6),
    ALTER COLUMN housing_allowance TYPE NUMERIC(24,6),
    ALTER COLUMN transport_allowance TYPE NUMERIC(24,6),
    ALTER COLUMN other_allowances TYPE NUMERIC(24,6),
    ALTER COLUMN gross_salary TYPE NUMERIC(24,6);

ALTER TABLE payroll.end_of_service_calculations
    ALTER COLUMN service_years TYPE NUMERIC(24,6),
    ALTER COLUMN last_gross_salary TYPE NUMERIC(24,6),
    ALTER COLUMN award_amount TYPE NUMERIC(24,6),
    ALTER COLUMN statutory_rate_fraction TYPE NUMERIC(12,8),
    ALTER COLUMN final_settlement_amount TYPE NUMERIC(24,6);

ALTER TABLE payroll.loan_installments
    ALTER COLUMN amount TYPE NUMERIC(24,6);

ALTER TABLE payroll.loans
    ALTER COLUMN principal_amount TYPE NUMERIC(24,6),
    ALTER COLUMN monthly_installment TYPE NUMERIC(24,6),
    ALTER COLUMN remaining_balance TYPE NUMERIC(24,6);

ALTER TABLE payroll.payslips
    ALTER COLUMN gross_salary TYPE NUMERIC(24,6),
    ALTER COLUMN total_deductions TYPE NUMERIC(24,6),
    ALTER COLUMN net_salary TYPE NUMERIC(24,6);

ALTER TABLE payroll.run_employees
    ALTER COLUMN basic_salary TYPE NUMERIC(24,6),
    ALTER COLUMN housing_allowance TYPE NUMERIC(24,6),
    ALTER COLUMN transport_allowance TYPE NUMERIC(24,6),
    ALTER COLUMN other_allowances TYPE NUMERIC(24,6),
    ALTER COLUMN overtime_amount TYPE NUMERIC(24,6),
    ALTER COLUMN gross_salary TYPE NUMERIC(24,6),
    ALTER COLUMN gosi_employee_deduction TYPE NUMERIC(24,6),
    ALTER COLUMN loan_deduction TYPE NUMERIC(24,6),
    ALTER COLUMN absence_deduction TYPE NUMERIC(24,6),
    ALTER COLUMN other_deductions TYPE NUMERIC(24,6),
    ALTER COLUMN total_deductions TYPE NUMERIC(24,6),
    ALTER COLUMN net_salary TYPE NUMERIC(24,6);

ALTER TABLE payroll.runs
    ALTER COLUMN total_basic TYPE NUMERIC(24,6),
    ALTER COLUMN total_allowances TYPE NUMERIC(24,6),
    ALTER COLUMN total_gross TYPE NUMERIC(24,6),
    ALTER COLUMN total_deductions TYPE NUMERIC(24,6),
    ALTER COLUMN total_net TYPE NUMERIC(24,6);

ALTER TABLE payroll.salary_components
    ALTER COLUMN default_amount TYPE NUMERIC(24,6);

ALTER TABLE payroll.salary_structure_lines
    ALTER COLUMN rate_percentage TYPE NUMERIC(12,8),
    ALTER COLUMN fixed_amount TYPE NUMERIC(24,6);

ALTER TABLE products.item_uoms
    ALTER COLUMN conversion_rate TYPE NUMERIC(24,12);

ALTER TABLE products.items
    ALTER COLUMN cost_price TYPE NUMERIC(24,6),
    ALTER COLUMN list_price TYPE NUMERIC(24,6),
    ALTER COLUMN standard_cost TYPE NUMERIC(24,6);

ALTER TABLE products.skus
    ALTER COLUMN cost_price TYPE NUMERIC(24,6),
    ALTER COLUMN selling_price TYPE NUMERIC(24,6);

ALTER TABLE products.uom_conversions
    ALTER COLUMN conversion_rate TYPE NUMERIC(24,12);

ALTER TABLE purchasing.goods_receipt_lines
    ALTER COLUMN received_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN accepted_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN rejected_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_cost TYPE NUMERIC(24,6);

ALTER TABLE purchasing.purchase_order_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN received_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN billed_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE purchasing.purchase_orders
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE purchasing.purchase_request_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN ordered_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN estimated_unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN estimated_total_cost TYPE NUMERIC(24,6);

ALTER TABLE purchasing.purchase_requests
    ALTER COLUMN estimated_total TYPE NUMERIC(24,6);

ALTER TABLE purchasing.rfq_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN target_unit_price TYPE NUMERIC(24,6);

ALTER TABLE purchasing.supplier_quote_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE purchasing.supplier_quotes
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE receivables.customer_credit_limits
    ALTER COLUMN credit_limit TYPE NUMERIC(24,6),
    ALTER COLUMN current_balance TYPE NUMERIC(24,6);

ALTER TABLE receivables.open_items
    ALTER COLUMN original_amount TYPE NUMERIC(24,6),
    ALTER COLUMN allocated_amount TYPE NUMERIC(24,6),
    ALTER COLUMN remaining_amount TYPE NUMERIC(24,6);

ALTER TABLE receivables.payment_terms
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8);

ALTER TABLE receivables.receipt_allocations
    ALTER COLUMN allocated_amount TYPE NUMERIC(24,6);

ALTER TABLE receivables.receipts
    ALTER COLUMN amount TYPE NUMERIC(24,6),
    ALTER COLUMN allocated_amount TYPE NUMERIC(24,6),
    ALTER COLUMN unallocated_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.credit_note_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.credit_notes
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.delivery_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_cost TYPE NUMERIC(24,6),
    ALTER COLUMN total_cost TYPE NUMERIC(24,6);

ALTER TABLE sales.invoice_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN gross_amount TYPE NUMERIC(24,6),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN net_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.invoices
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.order_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN delivered_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN invoiced_quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.orders
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.quotation_lines
    ALTER COLUMN quantity TYPE NUMERIC(24,6),
    ALTER COLUMN unit_price TYPE NUMERIC(24,6),
    ALTER COLUMN discount_percent TYPE NUMERIC(12,8),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE sales.quotations
    ALTER COLUMN exchange_rate TYPE NUMERIC(24,12),
    ALTER COLUMN subtotal TYPE NUMERIC(24,6),
    ALTER COLUMN discount_amount TYPE NUMERIC(24,6),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_amount TYPE NUMERIC(24,6);

ALTER TABLE tax.tax_rate_versions
    ALTER COLUMN rate_percent TYPE NUMERIC(12,8);

ALTER TABLE tax.tax_returns
    ALTER COLUMN total_sales_taxable TYPE NUMERIC(24,6),
    ALTER COLUMN total_sales_vat TYPE NUMERIC(24,6),
    ALTER COLUMN total_purchases_taxable TYPE NUMERIC(24,6),
    ALTER COLUMN total_purchases_vat TYPE NUMERIC(24,6),
    ALTER COLUMN net_vat_due TYPE NUMERIC(24,6);

ALTER TABLE tax.tax_transaction_lines
    ALTER COLUMN tax_rate_percent TYPE NUMERIC(12,8),
    ALTER COLUMN taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN gross_amount TYPE NUMERIC(24,6);

ALTER TABLE tax.tax_transactions
    ALTER COLUMN total_taxable_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_tax_amount TYPE NUMERIC(24,6),
    ALTER COLUMN total_gross_amount TYPE NUMERIC(24,6);

ALTER TABLE treasury.cash_accounts
    ALTER COLUMN max_limit TYPE NUMERIC(24,6),
    ALTER COLUMN current_balance TYPE NUMERIC(24,6);

ALTER TABLE treasury.payments
    ALTER COLUMN amount TYPE NUMERIC(24,6);

ALTER TABLE treasury.receipts
    ALTER COLUMN amount TYPE NUMERIC(24,6);

ALTER TABLE treasury.transfers
    ALTER COLUMN amount TYPE NUMERIC(24,6),
    ALTER COLUMN fee_amount TYPE NUMERIC(24,6);

ALTER TABLE inventory.stock_balances ADD COLUMN quantity_available NUMERIC(24,6) GENERATED ALWAYS AS ((quantity_on_hand - quantity_reserved)) STORED;

ALTER TABLE inventory.stock_count_lines ADD COLUMN variance NUMERIC(24,6) GENERATED ALWAYS AS ((counted_quantity - system_quantity)) STORED;

CREATE VIEW reporting.trial_balance AS
SELECT
    a.legal_entity_id,
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(jl.base_debit), 0)::NUMERIC(24,6) AS total_debit,
    COALESCE(SUM(jl.base_credit), 0)::NUMERIC(24,6) AS total_credit,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(jl.base_debit), 0) - COALESCE(SUM(jl.base_credit), 0))
        ELSE (COALESCE(SUM(jl.base_credit), 0) - COALESCE(SUM(jl.base_debit), 0))
    END::NUMERIC(24,6) AS net_balance
FROM accounting.accounts a
LEFT JOIN (
    SELECT jl.* FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.status IN ('POSTED', 'REVERSED')
) jl ON jl.account_id = a.id AND jl.legal_entity_id = a.legal_entity_id
GROUP BY a.legal_entity_id, a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance;

CREATE VIEW reporting.general_ledger AS
SELECT
    je.legal_entity_id,
    je.id AS entry_id,
    je.entry_number,
    je.accounting_date,
    je.source_module,
    je.reference,
    jl.id AS line_id,
    jl.line_number,
    jl.account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    jl.base_debit,
    jl.base_credit,
    jl.description
FROM accounting.journal_entries je
JOIN accounting.journal_lines jl ON jl.journal_entry_id = je.id
JOIN accounting.accounts a ON a.id = jl.account_id
WHERE je.status IN ('POSTED', 'REVERSED');

CREATE VIEW reporting.customer_aging AS
SELECT
    oi.legal_entity_id,
    oi.customer_id,
    c.customer_number,
    c.name_en AS customer_name_en,
    c.name_ar AS customer_name_ar,
    oi.id AS open_item_id,
    oi.document_number,
    oi.document_date,
    oi.due_date,
    oi.original_amount,
    oi.remaining_amount,
    (CURRENT_DATE - oi.due_date) AS days_overdue,
    CASE WHEN oi.due_date >= CURRENT_DATE THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS current_amount,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 1 AND 30 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_1_30,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 31 AND 60 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_31_60,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 61 AND 90 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_61_90,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 91 AND 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_91_120,
    CASE WHEN CURRENT_DATE - oi.due_date > 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_over_120
FROM receivables.open_items oi
JOIN receivables.customers c ON c.id = oi.customer_id
WHERE oi.status IN ('OPEN', 'PARTIALLY_PAID') AND oi.remaining_amount > 0;

CREATE VIEW reporting.supplier_aging AS
SELECT
    oi.legal_entity_id,
    oi.supplier_id,
    s.supplier_number,
    s.name_en AS supplier_name_en,
    s.name_ar AS supplier_name_ar,
    oi.id AS open_item_id,
    oi.document_number,
    oi.document_date,
    oi.due_date,
    oi.original_amount,
    oi.remaining_amount,
    (CURRENT_DATE - oi.due_date) AS days_overdue,
    CASE WHEN oi.due_date >= CURRENT_DATE THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS current_amount,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 1 AND 30 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_1_30,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 31 AND 60 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_31_60,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 61 AND 90 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_61_90,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 91 AND 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_91_120,
    CASE WHEN CURRENT_DATE - oi.due_date > 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(24,6) AS aging_over_120
FROM payables.open_items oi
JOIN payables.suppliers s ON s.id = oi.supplier_id
WHERE oi.status IN ('OPEN', 'PARTIALLY_PAID') AND oi.remaining_amount > 0;

CREATE VIEW reporting.inventory_valuation AS
SELECT
    sb.legal_entity_id,
    sb.item_id,
    i.code AS item_code,
    i.name_en AS item_name_en,
    i.name_ar AS item_name_ar,
    sb.warehouse_id,
    w.code AS warehouse_code,
    w.name_en AS warehouse_name_en,
    w.name_ar AS warehouse_name_ar,
    sb.quantity_on_hand,
    sb.quantity_reserved,
    sb.quantity_available,
    sb.average_cost,
    sb.total_value
FROM inventory.stock_balances sb
JOIN products.items i ON i.id = sb.item_id
JOIN warehouse.warehouses w ON w.id = sb.warehouse_id;

CREATE VIEW reporting.sales_daily AS
SELECT
    legal_entity_id,
    issue_date AS sale_date,
    COUNT(*)::INT AS invoice_count,
    COALESCE(SUM(subtotal), 0)::NUMERIC(24,6) AS total_subtotal,
    COALESCE(SUM(discount_amount), 0)::NUMERIC(24,6) AS total_discount,
    COALESCE(SUM(taxable_amount), 0)::NUMERIC(24,6) AS total_taxable,
    COALESCE(SUM(tax_amount), 0)::NUMERIC(24,6) AS total_tax,
    COALESCE(SUM(total_amount), 0)::NUMERIC(24,6) AS total_sales
FROM sales.invoices
WHERE status IN ('APPROVED', 'POSTED')
GROUP BY legal_entity_id, issue_date;

CREATE VIEW reporting.payroll_summary AS
SELECT
    r.legal_entity_id,
    r.id AS run_id,
    r.run_number,
    r.period_id,
    p.period_name_en,
    p.period_name_ar,
    r.run_date,
    r.status,
    r.employee_count,
    r.total_basic,
    r.total_allowances,
    r.total_gross,
    r.total_deductions,
    r.total_net
FROM payroll.runs r
JOIN payroll.periods p ON p.id = r.period_id;

DO $$
DECLARE v_view RECORD; v_grant RECORD;
BEGIN
    FOR v_view IN SELECT * FROM canonical_view_metadata LOOP
        EXECUTE format('ALTER VIEW reporting.%I OWNER TO %I',v_view.relname,v_view.owner_name);
        FOR v_grant IN SELECT * FROM aclexplode(v_view.relacl) LOOP
            EXECUTE format('GRANT %s ON reporting.%I TO %s%s',v_grant.privilege_type,v_view.relname,
                CASE WHEN v_grant.grantee=0 THEN 'PUBLIC' ELSE quote_ident(pg_get_userbyid(v_grant.grantee)) END,
                CASE WHEN v_grant.is_grantable THEN ' WITH GRANT OPTION' ELSE '' END);
        END LOOP;
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Precision migrations are forward-only; restore a verified backup to roll back'; END $$;
-- +goose StatementEnd
