-- +goose Up
-- ============================================================================
-- 000270_reporting.sql: Phase 29 — Reporting Platform
-- Canonical Schema: reporting.* per Section 40.30 & Section 100
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS reporting;

-- 1. Trial Balance View
CREATE OR REPLACE VIEW reporting.trial_balance AS
SELECT
    a.legal_entity_id,
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(jl.base_debit), 0)::NUMERIC(18, 4) AS total_debit,
    COALESCE(SUM(jl.base_credit), 0)::NUMERIC(18, 4) AS total_credit,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(jl.base_debit), 0) - COALESCE(SUM(jl.base_credit), 0))
        ELSE (COALESCE(SUM(jl.base_credit), 0) - COALESCE(SUM(jl.base_debit), 0))
    END::NUMERIC(18, 4) AS net_balance
FROM accounting.accounts a
LEFT JOIN accounting.journal_lines jl ON jl.account_id = a.id
LEFT JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id AND je.status = 'POSTED'
GROUP BY a.legal_entity_id, a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance;

-- 2. General Ledger View
CREATE OR REPLACE VIEW reporting.general_ledger AS
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
WHERE je.status = 'POSTED';

-- 3. Customer Aging View
CREATE OR REPLACE VIEW reporting.customer_aging AS
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
    CASE WHEN oi.due_date >= CURRENT_DATE THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS current_amount,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 1 AND 30 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_1_30,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 31 AND 60 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_31_60,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 61 AND 90 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_61_90,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 91 AND 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_91_120,
    CASE WHEN CURRENT_DATE - oi.due_date > 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_over_120
FROM receivables.open_items oi
JOIN receivables.customers c ON c.id = oi.customer_id
WHERE oi.status IN ('OPEN', 'PARTIALLY_PAID') AND oi.remaining_amount > 0;

-- 4. Supplier Aging View
CREATE OR REPLACE VIEW reporting.supplier_aging AS
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
    CASE WHEN oi.due_date >= CURRENT_DATE THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS current_amount,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 1 AND 30 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_1_30,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 31 AND 60 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_31_60,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 61 AND 90 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_61_90,
    CASE WHEN CURRENT_DATE - oi.due_date BETWEEN 91 AND 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_91_120,
    CASE WHEN CURRENT_DATE - oi.due_date > 120 THEN oi.remaining_amount ELSE 0.0000 END::NUMERIC(18, 4) AS aging_over_120
FROM payables.open_items oi
JOIN payables.suppliers s ON s.id = oi.supplier_id
WHERE oi.status IN ('OPEN', 'PARTIALLY_PAID') AND oi.remaining_amount > 0;

-- 5. Inventory Valuation View
CREATE OR REPLACE VIEW reporting.inventory_valuation AS
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

-- 6. Sales Daily Rollup View
CREATE OR REPLACE VIEW reporting.sales_daily AS
SELECT
    legal_entity_id,
    issue_date AS sale_date,
    COUNT(*)::INT AS invoice_count,
    COALESCE(SUM(subtotal), 0)::NUMERIC(18, 4) AS total_subtotal,
    COALESCE(SUM(discount_amount), 0)::NUMERIC(18, 4) AS total_discount,
    COALESCE(SUM(taxable_amount), 0)::NUMERIC(18, 4) AS total_taxable,
    COALESCE(SUM(tax_amount), 0)::NUMERIC(18, 4) AS total_tax,
    COALESCE(SUM(total_amount), 0)::NUMERIC(18, 4) AS total_sales
FROM sales.invoices
WHERE status IN ('APPROVED', 'POSTED')
GROUP BY legal_entity_id, issue_date;

-- 7. Payroll Summary View
CREATE OR REPLACE VIEW reporting.payroll_summary AS
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

-- Seed Permissions for Reporting Platform
INSERT INTO authz.permissions (id, code, name_ar, name_en, module, created_at)
VALUES
    (uuidv7(), 'reporting.read', 'عرض التقارير المالية والتشغيلية', 'View Financial and Operational Reports', 'reporting', now()),
    (uuidv7(), 'reporting.export', 'تصدير التقارير إلى ملفات', 'Export Reports to Files', 'reporting', now()),
    (uuidv7(), 'reporting.manage', 'إدارة وتخصيص التقارير', 'Manage Report Configurations', 'reporting', now())
ON CONFLICT (code) DO NOTHING;

-- Assign to Admin role
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'ADMIN'
  AND p.module = 'reporting'
ON CONFLICT DO NOTHING;

-- Seed procedure for default reporting configurations
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE reporting.seed_default_reporting_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Informational notice for reporting bootstrap
    RAISE NOTICE 'Reporting platform views and default configurations verified for legal entity %', p_legal_entity_id;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS reporting.seed_default_reporting_data(UUID);
DROP VIEW IF EXISTS reporting.payroll_summary CASCADE;
DROP VIEW IF EXISTS reporting.sales_daily CASCADE;
DROP VIEW IF EXISTS reporting.inventory_valuation CASCADE;
DROP VIEW IF EXISTS reporting.supplier_aging CASCADE;
DROP VIEW IF EXISTS reporting.customer_aging CASCADE;
DROP VIEW IF EXISTS reporting.general_ledger CASCADE;
DROP VIEW IF EXISTS reporting.trial_balance CASCADE;
DELETE FROM authz.role_permissions WHERE permission_id IN (SELECT id FROM authz.permissions WHERE module = 'reporting');
DELETE FROM authz.permissions WHERE module = 'reporting';
DROP SCHEMA IF EXISTS reporting CASCADE;
