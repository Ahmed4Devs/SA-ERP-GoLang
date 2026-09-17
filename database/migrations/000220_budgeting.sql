-- +goose Up
-- Budgeting Module Schema (Phase 24)
-- Compliant with Master Specification Section 40.24 and 95

CREATE SCHEMA IF NOT EXISTS budgeting;

-- 1. Budgets (Master Header)
CREATE TABLE IF NOT EXISTS budgeting.budgets (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    description_en TEXT,
    description_ar TEXT,
    fiscal_year_id UUID NOT NULL REFERENCES accounting.fiscal_years(id) ON DELETE RESTRICT,
    currency_code VARCHAR(3) NOT NULL REFERENCES ref.currencies(code) DEFAULT 'SAR',
    budget_type VARCHAR(50) NOT NULL DEFAULT 'OPERATING', -- 'OPERATING', 'CAPITAL', 'PROJECT', 'CASH_FLOW'
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',          -- 'DRAFT', 'SUBMITTED', 'APPROVED', 'ACTIVE', 'REVISED', 'CLOSED'
    total_budgeted_amount NUMERIC(18,4) NOT NULL DEFAULT 0.0000,
    total_actual_amount NUMERIC(18,4) NOT NULL DEFAULT 0.0000,
    total_committed_amount NUMERIC(18,4) NOT NULL DEFAULT 0.0000,
    created_by_user_id UUID REFERENCES iam.users(id),
    approved_by_user_id UUID REFERENCES iam.users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_budgeting_budgets_code UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_budgeting_budgets_le ON budgeting.budgets(legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_budgets_fy ON budgeting.budgets(fiscal_year_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_budgets_status ON budgeting.budgets(status);

-- 2. Budget Versions (Revisions & Approvals)
CREATE TABLE IF NOT EXISTS budgeting.budget_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    budget_id UUID NOT NULL REFERENCES budgeting.budgets(id) ON DELETE CASCADE,
    version_number INT NOT NULL DEFAULT 1,
    version_name VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- 'DRAFT', 'ACTIVE', 'SUPERSEDED'
    notes TEXT,
    created_by_user_id UUID REFERENCES iam.users(id),
    approved_by_user_id UUID REFERENCES iam.users(id),
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_budgeting_versions UNIQUE (budget_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_budgeting_versions_budget ON budgeting.budget_versions(budget_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_versions_status ON budgeting.budget_versions(status);

-- 3. Budget Lines (Multi-Dimensional Allocations)
-- Dimensions per Section 40.24: account, department, cost_center, project, period
CREATE TABLE IF NOT EXISTS budgeting.budget_lines (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    budget_version_id UUID NOT NULL REFERENCES budgeting.budget_versions(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounting.accounts(id) ON DELETE RESTRICT,
    department_id UUID REFERENCES organization.departments(id) ON DELETE SET NULL,
    cost_center_id UUID REFERENCES organization.cost_centers(id) ON DELETE SET NULL,
    project_code VARCHAR(100),
    fiscal_period_id UUID REFERENCES accounting.fiscal_periods(id) ON DELETE SET NULL,
    amount NUMERIC(18,4) NOT NULL CHECK (amount >= 0),
    actual_amount NUMERIC(18,4) NOT NULL DEFAULT 0.0000,
    committed_amount NUMERIC(18,4) NOT NULL DEFAULT 0.0000,
    line_description_en TEXT,
    line_description_ar TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_budgeting_lines_version ON budgeting.budget_lines(budget_version_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_lines_account ON budgeting.budget_lines(account_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_lines_dept ON budgeting.budget_lines(department_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_lines_cc ON budgeting.budget_lines(cost_center_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_lines_period ON budgeting.budget_lines(fiscal_period_id);

-- 4. Budget Transfers & Reallocations (Inter-line movements with full audit)
CREATE TABLE IF NOT EXISTS budgeting.budget_transfers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    transfer_number VARCHAR(50) NOT NULL,
    source_line_id UUID NOT NULL REFERENCES budgeting.budget_lines(id) ON DELETE RESTRICT,
    target_line_id UUID NOT NULL REFERENCES budgeting.budget_lines(id) ON DELETE RESTRICT,
    amount NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    reason TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'COMPLETED', -- 'PENDING', 'COMPLETED', 'REJECTED'
    transferred_by_user_id UUID REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_budgeting_transfers_no UNIQUE (legal_entity_id, transfer_number)
);

CREATE INDEX IF NOT EXISTS idx_budgeting_transfers_le ON budgeting.budget_transfers(legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_transfers_src ON budgeting.budget_transfers(source_line_id);
CREATE INDEX IF NOT EXISTS idx_budgeting_transfers_tgt ON budgeting.budget_transfers(target_line_id);

-- 5. Seed Procedure for Default Budgeting Data
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE budgeting.seed_default_budgeting_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    v_fiscal_year_id UUID;
    v_budget_id UUID;
    v_version_id UUID;
    v_dept_fin UUID;
    v_dept_hr UUID;
    v_dept_it UUID;
    v_dept_sales UUID;
    v_cc_fin UUID;
    v_cc_sales UUID;
    v_acc_salaries UUID;
    v_acc_rent UUID;
    v_acc_marketing UUID;
    v_acc_depr UUID;
    v_period_q1 UUID;
    v_period_q2 UUID;
BEGIN
    -- Check if default seeded budget already exists for this legal entity
    IF EXISTS (SELECT 1 FROM budgeting.budgets WHERE legal_entity_id = p_legal_entity_id AND code = 'BUD-2026-OPEX') THEN
        RETURN;
    END IF;

    -- Retrieve active fiscal year
    SELECT id INTO v_fiscal_year_id
    FROM accounting.fiscal_years
    WHERE legal_entity_id = p_legal_entity_id
    ORDER BY start_date DESC
    LIMIT 1;

    IF v_fiscal_year_id IS NULL THEN
        RETURN;
    END IF;

    -- Retrieve sample departments
    SELECT id INTO v_dept_fin FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'DEP-FIN' LIMIT 1;
    SELECT id INTO v_dept_hr FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'DEP-HR' LIMIT 1;
    SELECT id INTO v_dept_it FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'DEP-IT' LIMIT 1;
    SELECT id INTO v_dept_sales FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'DEP-SALES' LIMIT 1;

    -- Retrieve sample cost centers
    SELECT id INTO v_cc_fin FROM organization.cost_centers WHERE legal_entity_id = p_legal_entity_id AND code = 'CC-ADMIN' LIMIT 1;
    SELECT id INTO v_cc_sales FROM organization.cost_centers WHERE legal_entity_id = p_legal_entity_id AND code = 'CC-SALES' LIMIT 1;

    -- Retrieve sample accounts
    SELECT id INTO v_acc_salaries FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND (code = '5101' OR code LIKE '51%') LIMIT 1;
    SELECT id INTO v_acc_rent FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND (code = '5201' OR code LIKE '52%') LIMIT 1;
    SELECT id INTO v_acc_marketing FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND (code = '5301' OR code LIKE '53%') LIMIT 1;
    SELECT id INTO v_acc_depr FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND (code = '5401' OR code LIKE '5%') LIMIT 1;

    -- Retrieve periods
    SELECT id INTO v_period_q1 FROM accounting.fiscal_periods WHERE fiscal_year_id = v_fiscal_year_id ORDER BY period_number ASC LIMIT 1;
    SELECT id INTO v_period_q2 FROM accounting.fiscal_periods WHERE fiscal_year_id = v_fiscal_year_id ORDER BY period_number ASC OFFSET 1 LIMIT 1;

    IF v_acc_salaries IS NULL THEN
        -- Fallback to any expense account
        SELECT id INTO v_acc_salaries FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND account_type = 'EXPENSE' LIMIT 1;
    END IF;

    IF v_acc_salaries IS NULL THEN
        RETURN;
    END IF;

    -- Create primary Annual Operating Budget
    INSERT INTO budgeting.budgets (
        id, legal_entity_id, code, name_en, name_ar, description_en, description_ar,
        fiscal_year_id, currency_code, budget_type, status,
        total_budgeted_amount, total_actual_amount, total_committed_amount,
        created_at, updated_at
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'BUD-2026-OPEX',
        'Annual Operating Budget 2026', 'الميزانية التشغيلية السنوية لعام 2026',
        'Approved operational expense and overhead budget for fiscal year 2026',
        'الميزانية التشغيلية المعتمدة للمصروفات العمومية والإدارية والتسويقية لعام 2026',
        v_fiscal_year_id, 'SAR', 'OPERATING', 'ACTIVE',
        1850000.0000, 320000.0000, 45000.0000,
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ) RETURNING id INTO v_budget_id;

    -- Create Version 1 (Approved / Active)
    INSERT INTO budgeting.budget_versions (
        id, budget_id, version_number, version_name, status, notes,
        approved_at, created_at, updated_at
    ) VALUES (
        uuidv7(), v_budget_id, 1, 'Initial Approved Budget v1.0', 'ACTIVE',
        'Board-approved baseline operational budget for FY 2026',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ) RETURNING id INTO v_version_id;

    -- Create Multi-Dimensional Budget Lines
    -- Line 1: Salaries & Wages (HR / Admin)
    INSERT INTO budgeting.budget_lines (
        id, budget_version_id, account_id, department_id, cost_center_id,
        project_code, fiscal_period_id, amount, actual_amount, committed_amount,
        line_description_en, line_description_ar, created_at, updated_at
    ) VALUES (
        uuidv7(), v_version_id, v_acc_salaries, v_dept_hr, v_cc_fin,
        'HQ-OPS', v_period_q1, 1000000.0000, 210000.0000, 0.0000,
        'Staff salaries, allowances, and social insurance contribution allocations',
        'مخصصات رواتب الموظفين والبدلات واشتراكات التأمينات الاجتماعية',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    );

    -- Line 2: Facilities & Rent
    IF v_acc_rent IS NOT NULL THEN
        INSERT INTO budgeting.budget_lines (
            id, budget_version_id, account_id, department_id, cost_center_id,
            project_code, fiscal_period_id, amount, actual_amount, committed_amount,
            line_description_en, line_description_ar, created_at, updated_at
        ) VALUES (
            uuidv7(), v_version_id, v_acc_rent, v_dept_fin, v_cc_fin,
            'HQ-FAC', v_period_q1, 350000.0000, 60000.0000, 15000.0000,
            'Office premises leases, maintenance, and utility bills',
            'عقود إيجار المقرات الرئيسية، الصيانة، وفواتير الخدمات العامة',
            CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        );
    END IF;

    -- Line 3: Marketing & Customer Acquisition
    IF v_acc_marketing IS NOT NULL THEN
        INSERT INTO budgeting.budget_lines (
            id, budget_version_id, account_id, department_id, cost_center_id,
            project_code, fiscal_period_id, amount, actual_amount, committed_amount,
            line_description_en, line_description_ar, created_at, updated_at
        ) VALUES (
            uuidv7(), v_version_id, v_acc_marketing, v_dept_sales, v_cc_sales,
            'MKT-Q1', v_period_q1, 300000.0000, 45000.0000, 25000.0000,
            'Digital advertising, brand awareness campaigns, and client exhibitions',
            'الحملات الإعلانية الرقمية، الترويج للعلامة التجارية، والمعارض',
            CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        );
    END IF;

    -- Line 4: Technology, Software & Depreciation
    IF v_acc_depr IS NOT NULL THEN
        INSERT INTO budgeting.budget_lines (
            id, budget_version_id, account_id, department_id, cost_center_id,
            project_code, fiscal_period_id, amount, actual_amount, committed_amount,
            line_description_en, line_description_ar, created_at, updated_at
        ) VALUES (
            uuidv7(), v_version_id, v_acc_depr, v_dept_it, v_cc_fin,
            'TECH-INFRA', v_period_q2, 200000.0000, 5000.0000, 5000.0000,
            'Cloud infrastructure, ERP subscriptions, and IT equipment amortization',
            'البنية التحتية السحابية، اشتراكات البرمجيات، وإهلاك معدات تقنية المعلومات',
            CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
        );
    END IF;

    -- Create a Capital Expenditure Budget (CAPEX)
    INSERT INTO budgeting.budgets (
        id, legal_entity_id, code, name_en, name_ar, description_en, description_ar,
        fiscal_year_id, currency_code, budget_type, status,
        total_budgeted_amount, total_actual_amount, total_committed_amount,
        created_at, updated_at
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'BUD-2026-CAPEX',
        'Capital Expenditure Budget 2026', 'ميزانية النفقات الرأسمالية لعام 2026',
        'Planned investments in machinery, fleet vehicles, and digital transformation',
        'الاستثمارات الرأسمالية المخططة في الآلات والمعدات، أسطول النقل والتحول الرقمي',
        v_fiscal_year_id, 'SAR', 'CAPITAL', 'APPROVED',
        750000.0000, 120000.0000, 80000.0000,
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    );

END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS budgeting.seed_default_budgeting_data(UUID);
DROP TABLE IF EXISTS budgeting.budget_transfers CASCADE;
DROP TABLE IF EXISTS budgeting.budget_lines CASCADE;
DROP TABLE IF EXISTS budgeting.budget_versions CASCADE;
DROP TABLE IF EXISTS budgeting.budgets CASCADE;
DROP SCHEMA IF EXISTS budgeting CASCADE;
