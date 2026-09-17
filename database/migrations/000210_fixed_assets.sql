-- +goose Up
-- Fixed Assets Module Schema (Phase 23)
-- Compliant with Saudi Corporate Standards, IFRS, and General Ledger integration

CREATE SCHEMA IF NOT EXISTS fixedassets;

-- 1. Asset Categories (Classes)
CREATE TABLE IF NOT EXISTS fixedassets.categories (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    description_en TEXT,
    description_ar TEXT,
    depreciation_method VARCHAR(50) NOT NULL DEFAULT 'STRAIGHT_LINE',
    useful_life_months INT NOT NULL CHECK (useful_life_months > 0),
    salvage_value_rate NUMERIC(6,4) NOT NULL DEFAULT 0.0000,
    asset_account_id UUID REFERENCES accounting.accounts(id),
    accumulated_depreciation_account_id UUID REFERENCES accounting.accounts(id),
    depreciation_expense_account_id UUID REFERENCES accounting.accounts(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_categories_code UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_categories_le ON fixedassets.categories(legal_entity_id);

-- 2. Asset Locations
CREATE TABLE IF NOT EXISTS fixedassets.asset_locations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    building VARCHAR(100),
    floor VARCHAR(50),
    room VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_locations_code UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_locations_le ON fixedassets.asset_locations(legal_entity_id);

-- 3. Depreciation Books (Commercial Accounting vs Tax)
CREATE TABLE IF NOT EXISTS fixedassets.depreciation_books (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_books_code UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_books_le ON fixedassets.depreciation_books(legal_entity_id);

-- 4. Fixed Assets Master
CREATE TABLE IF NOT EXISTS fixedassets.assets (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES fixedassets.categories(id),
    location_id UUID REFERENCES fixedassets.asset_locations(id),
    asset_number VARCHAR(100) NOT NULL,
    serial_number VARCHAR(100),
    barcode VARCHAR(100),
    name_en VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255) NOT NULL,
    description_en TEXT,
    description_ar TEXT,
    acquisition_date DATE NOT NULL,
    in_service_date DATE NOT NULL,
    acquisition_cost NUMERIC(24,6) NOT NULL CHECK (acquisition_cost >= 0),
    salvage_value NUMERIC(24,6) NOT NULL DEFAULT 0.000000 CHECK (salvage_value >= 0),
    useful_life_months INT NOT NULL CHECK (useful_life_months > 0),
    depreciation_method VARCHAR(50) NOT NULL DEFAULT 'STRAIGHT_LINE',
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- DRAFT, ACTIVE, FULLY_DEPRECIATED, DISPOSED, WRITTEN_OFF
    current_net_book_value NUMERIC(24,6) NOT NULL,
    accumulated_depreciation NUMERIC(24,6) NOT NULL DEFAULT 0.000000,
    impairment_total NUMERIC(24,6) NOT NULL DEFAULT 0.000000,
    supplier_id UUID REFERENCES payables.suppliers(id),
    purchase_invoice_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_assets_number UNIQUE (legal_entity_id, asset_number)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_assets_le ON fixedassets.assets(legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_fixedassets_assets_cat ON fixedassets.assets(category_id);
CREATE INDEX IF NOT EXISTS idx_fixedassets_assets_loc ON fixedassets.assets(location_id);
CREATE INDEX IF NOT EXISTS idx_fixedassets_assets_status ON fixedassets.assets(status);

-- 5. Asset Assignments (Custody / Responsibility)
CREATE TABLE IF NOT EXISTS fixedassets.asset_assignments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    asset_id UUID NOT NULL REFERENCES fixedassets.assets(id) ON DELETE CASCADE,
    assigned_to_type VARCHAR(50) NOT NULL DEFAULT 'EMPLOYEE', -- EMPLOYEE, DEPARTMENT, PROJECT
    employee_id UUID REFERENCES hr.employees(id),
    department_id UUID,
    assigned_date DATE NOT NULL,
    return_date DATE,
    notes TEXT,
    is_current BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_assign_asset ON fixedassets.asset_assignments(asset_id);
CREATE INDEX IF NOT EXISTS idx_fixedassets_assign_emp ON fixedassets.asset_assignments(employee_id);

-- 6. Depreciation Schedules (Projection by Period)
CREATE TABLE IF NOT EXISTS fixedassets.depreciation_schedules (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    asset_id UUID NOT NULL REFERENCES fixedassets.assets(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES fixedassets.depreciation_books(id),
    period_year INT NOT NULL,
    period_month INT NOT NULL,
    opening_net_book_value NUMERIC(24,6) NOT NULL,
    depreciation_amount NUMERIC(24,6) NOT NULL,
    closing_net_book_value NUMERIC(24,6) NOT NULL,
    accumulated_depreciation NUMERIC(24,6) NOT NULL,
    is_posted BOOLEAN NOT NULL DEFAULT false,
    depreciation_run_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_sched_period UNIQUE (asset_id, book_id, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_sched_asset ON fixedassets.depreciation_schedules(asset_id);
CREATE INDEX IF NOT EXISTS idx_fixedassets_sched_run ON fixedassets.depreciation_schedules(depreciation_run_id);

-- 7. Depreciation Runs (Monthly Batch Processing)
CREATE TABLE IF NOT EXISTS fixedassets.depreciation_runs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES fixedassets.depreciation_books(id),
    run_number VARCHAR(100) NOT NULL,
    period_year INT NOT NULL,
    period_month INT NOT NULL,
    run_date DATE NOT NULL,
    total_assets_processed INT NOT NULL DEFAULT 0,
    total_depreciation_amount NUMERIC(24,6) NOT NULL DEFAULT 0.000000,
    status VARCHAR(50) NOT NULL DEFAULT 'CALCULATED', -- DRAFT, CALCULATED, APPROVED, POSTED
    journal_entry_id UUID REFERENCES accounting.journal_entries(id),
    created_by UUID REFERENCES iam.users(id),
    approved_by UUID REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_runs_period UNIQUE (legal_entity_id, book_id, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_runs_le ON fixedassets.depreciation_runs(legal_entity_id);

-- 8. Depreciation Lines (Line item per asset in a run)
CREATE TABLE IF NOT EXISTS fixedassets.depreciation_lines (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    run_id UUID NOT NULL REFERENCES fixedassets.depreciation_runs(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES fixedassets.assets(id),
    asset_number VARCHAR(100) NOT NULL,
    cost_basis NUMERIC(24,6) NOT NULL,
    depreciation_amount NUMERIC(24,6) NOT NULL,
    accumulated_before NUMERIC(24,6) NOT NULL,
    accumulated_after NUMERIC(24,6) NOT NULL,
    net_book_value_after NUMERIC(24,6) NOT NULL,
    expense_account_id UUID REFERENCES accounting.accounts(id),
    accumulated_account_id UUID REFERENCES accounting.accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fixedassets_lines_run_asset UNIQUE (run_id, asset_id)
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_lines_run ON fixedassets.depreciation_lines(run_id);
CREATE INDEX IF NOT EXISTS idx_fixedassets_lines_asset ON fixedassets.depreciation_lines(asset_id);

-- 9. Asset Transfers (Movement Audit Log)
CREATE TABLE IF NOT EXISTS fixedassets.transfers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES fixedassets.assets(id),
    transfer_number VARCHAR(100) NOT NULL,
    transfer_date DATE NOT NULL,
    from_location_id UUID REFERENCES fixedassets.asset_locations(id),
    to_location_id UUID NOT NULL REFERENCES fixedassets.asset_locations(id),
    from_employee_id UUID REFERENCES hr.employees(id),
    to_employee_id UUID REFERENCES hr.employees(id),
    reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'COMPLETED',
    transferred_by UUID REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_transfers_asset ON fixedassets.transfers(asset_id);

-- 10. Asset Impairments
CREATE TABLE IF NOT EXISTS fixedassets.impairments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES fixedassets.assets(id),
    impairment_number VARCHAR(100) NOT NULL,
    impairment_date DATE NOT NULL,
    carrying_amount_before NUMERIC(24,6) NOT NULL,
    recoverable_amount NUMERIC(24,6) NOT NULL,
    impairment_loss NUMERIC(24,6) NOT NULL,
    reason TEXT NOT NULL,
    journal_entry_id UUID REFERENCES accounting.journal_entries(id),
    created_by UUID REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_impairments_asset ON fixedassets.impairments(asset_id);

-- 11. Asset Disposals
CREATE TABLE IF NOT EXISTS fixedassets.disposals (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES fixedassets.assets(id),
    disposal_number VARCHAR(100) NOT NULL,
    disposal_date DATE NOT NULL,
    disposal_type VARCHAR(50) NOT NULL DEFAULT 'SALE', -- SALE, SCRAP, WRITE_OFF, DONATION
    proceeds_amount NUMERIC(24,6) NOT NULL DEFAULT 0.000000,
    cost_at_disposal NUMERIC(24,6) NOT NULL,
    accumulated_depreciation_at_disposal NUMERIC(24,6) NOT NULL,
    net_book_value_at_disposal NUMERIC(24,6) NOT NULL,
    gain_or_loss_amount NUMERIC(24,6) NOT NULL,
    journal_entry_id UUID REFERENCES accounting.journal_entries(id),
    customer_id UUID REFERENCES receivables.customers(id),
    reason TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'POSTED',
    created_by UUID REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fixedassets_disposals_asset ON fixedassets.disposals(asset_id);

-- Permissions
INSERT INTO iam.permissions (id, code, name_en, name_ar, module, created_at) VALUES
    (uuidv7(), 'fixedassets:assets:read', 'View Fixed Assets', 'عرض الأصول الثابتة', 'FIXEDASSETS', CURRENT_TIMESTAMP),
    (uuidv7(), 'fixedassets:assets:write', 'Manage Fixed Assets', 'إدارة وتعديل الأصول الثابتة', 'FIXEDASSETS', CURRENT_TIMESTAMP),
    (uuidv7(), 'fixedassets:depreciation:run', 'Run Monthly Depreciation', 'تشغيل مسير الاستهلاك الشهري', 'FIXEDASSETS', CURRENT_TIMESTAMP),
    (uuidv7(), 'fixedassets:depreciation:post', 'Post Depreciation to General Ledger', 'اعتماد وترحيل الاستهلاك لدفتر الأستاذ', 'FIXEDASSETS', CURRENT_TIMESTAMP),
    (uuidv7(), 'fixedassets:transfers:manage', 'Manage Asset Transfers', 'إدارة نقل ومناقلات الأصول والعهد', 'FIXEDASSETS', CURRENT_TIMESTAMP),
    (uuidv7(), 'fixedassets:disposals:manage', 'Manage Asset Disposals & Write-offs', 'إدارة استبعاد وتخريد الأصول', 'FIXEDASSETS', CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- Seed Stored Procedure
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE fixedassets.seed_default_fixed_assets_data(p_legal_entity_id UUID)
LANGUAGE plpgsql AS $$
DECLARE
    v_chart_id UUID;
    v_nca_parent_id UUID;
    v_acc_depr_id UUID;
    v_exp_depr_id UUID;
    v_acc_veh_id UUID;
    v_acc_it_id UUID;
    v_acc_mch_id UUID;
    v_acc_fur_id UUID;

    v_cat_veh_id UUID;
    v_cat_it_id UUID;
    v_cat_mch_id UUID;
    v_cat_fur_id UUID;

    v_loc_hq_id UUID;
    v_loc_jed_id UUID;
    v_loc_dmm_id UUID;

    v_book_comm_id UUID;
    v_book_tax_id UUID;
BEGIN
    -- 1. Look up Chart of Accounts
    SELECT id INTO v_chart_id FROM accounting.chart_of_accounts WHERE legal_entity_id = p_legal_entity_id LIMIT 1;
    IF v_chart_id IS NOT NULL THEN
        -- Find Non-Current Assets parent (120000)
        SELECT id INTO v_nca_parent_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '120000';
        
        -- Find or create Accumulated Depreciation Contra-Asset account (129000)
        SELECT id INTO v_acc_depr_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '129000';
        IF v_acc_depr_id IS NULL THEN
            INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
            VALUES (p_legal_entity_id, v_chart_id, v_nca_parent_id, '129000', 'Accumulated Depreciation', 'مجمع الاستهلاك المتراكم', 'ASSET', 'CREDIT', true)
            ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
            RETURNING id INTO v_acc_depr_id;
        END IF;

        -- Find Depreciation Expense account (525000)
        SELECT id INTO v_exp_depr_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '525000';

        -- Asset Accounts for classes
        SELECT id INTO v_acc_veh_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '123000';
        IF v_acc_veh_id IS NULL THEN
            INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
            VALUES (p_legal_entity_id, v_chart_id, v_nca_parent_id, '123000', 'Vehicles & Transportation', 'السيارات ووسائل النقل', 'ASSET', 'DEBIT', true)
            ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
            RETURNING id INTO v_acc_veh_id;
        END IF;

        SELECT id INTO v_acc_it_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '124000';
        IF v_acc_it_id IS NULL THEN
            INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
            VALUES (p_legal_entity_id, v_chart_id, v_nca_parent_id, '124000', 'IT & Computer Equipment', 'أجهزة ومعدات تقنية المعلومات', 'ASSET', 'DEBIT', true)
            ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
            RETURNING id INTO v_acc_it_id;
        END IF;

        SELECT id INTO v_acc_mch_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '121000';
        IF v_acc_mch_id IS NULL THEN
            INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
            VALUES (p_legal_entity_id, v_chart_id, v_nca_parent_id, '121000', 'Machinery & Equipment', 'الآلات والمعدات والمصانع', 'ASSET', 'DEBIT', true)
            ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
            RETURNING id INTO v_acc_mch_id;
        END IF;

        SELECT id INTO v_acc_fur_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '125000';
        IF v_acc_fur_id IS NULL THEN
            INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
            VALUES (p_legal_entity_id, v_chart_id, v_nca_parent_id, '125000', 'Furniture & Fixtures', 'الأثاث والتجهيزات المكتبية', 'ASSET', 'DEBIT', true)
            ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
            RETURNING id INTO v_acc_fur_id;
        END IF;
    END IF;

    -- 2. Seed Depreciation Books (Commercial & Tax)
    INSERT INTO fixedassets.depreciation_books (legal_entity_id, code, name_en, name_ar, is_primary)
    VALUES (p_legal_entity_id, 'COMMERCIAL', 'Commercial Financial Book', 'الدفتر المالي والتجاري', true)
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_book_comm_id;
    IF v_book_comm_id IS NULL THEN
        SELECT id INTO v_book_comm_id FROM fixedassets.depreciation_books WHERE legal_entity_id = p_legal_entity_id AND code = 'COMMERCIAL';
    END IF;

    INSERT INTO fixedassets.depreciation_books (legal_entity_id, code, name_en, name_ar, is_primary)
    VALUES (p_legal_entity_id, 'TAX', 'ZATCA Tax Depreciation Book', 'دفتر الإهلاك الضريبي والزكوي', false)
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_book_tax_id;

    -- 3. Seed Asset Locations
    INSERT INTO fixedassets.asset_locations (legal_entity_id, code, name_en, name_ar, building, floor, room)
    VALUES (p_legal_entity_id, 'HQ-RYD', 'Riyadh Head Office', 'المقر الرئيسي - الرياض', 'Tower A', 'Floor 4', 'Office 402')
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_loc_hq_id;
    IF v_loc_hq_id IS NULL THEN
        SELECT id INTO v_loc_hq_id FROM fixedassets.asset_locations WHERE legal_entity_id = p_legal_entity_id AND code = 'HQ-RYD';
    END IF;

    INSERT INTO fixedassets.asset_locations (legal_entity_id, code, name_en, name_ar, building, floor, room)
    VALUES (p_legal_entity_id, 'BR-JED', 'Jeddah Regional Office', 'فرع جدة الإقليمي', 'Al-Madinah Center', 'Floor 2', 'Suite 210')
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_loc_jed_id;

    INSERT INTO fixedassets.asset_locations (legal_entity_id, code, name_en, name_ar, building, floor, room)
    VALUES (p_legal_entity_id, 'WH-DMM', 'Dammam Central Warehouse', 'مستودع الدمام المركزي', 'Warehouse Complex 3', 'Ground', 'Dock 5')
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_loc_dmm_id;

    -- 4. Seed Categories
    INSERT INTO fixedassets.categories (
        legal_entity_id, code, name_en, name_ar, description_en, description_ar,
        depreciation_method, useful_life_months, salvage_value_rate,
        asset_account_id, accumulated_depreciation_account_id, depreciation_expense_account_id
    ) VALUES
        (p_legal_entity_id, 'VEH', 'Vehicles & Fleet', 'السيارات وأسطول النقل',
         'Corporate transport and delivery fleet', 'سيارات الإدارة والنقل والتوزيع',
         'STRAIGHT_LINE', 48, 0.1000, v_acc_veh_id, v_acc_depr_id, v_exp_depr_id)
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_cat_veh_id;
    IF v_cat_veh_id IS NULL THEN
        SELECT id INTO v_cat_veh_id FROM fixedassets.categories WHERE legal_entity_id = p_legal_entity_id AND code = 'VEH';
    END IF;

    INSERT INTO fixedassets.categories (
        legal_entity_id, code, name_en, name_ar, description_en, description_ar,
        depreciation_method, useful_life_months, salvage_value_rate,
        asset_account_id, accumulated_depreciation_account_id, depreciation_expense_account_id
    ) VALUES
        (p_legal_entity_id, 'IT', 'IT & Computer Equipment', 'أجهزة ومعدات تقنية المعلومات',
         'Laptops, servers, switches and IT infrastructure', 'أجهزة الحاسب الآلي والخوادم والبنية التحتية التقنية',
         'STRAIGHT_LINE', 36, 0.0000, v_acc_it_id, v_acc_depr_id, v_exp_depr_id)
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_cat_it_id;
    IF v_cat_it_id IS NULL THEN
        SELECT id INTO v_cat_it_id FROM fixedassets.categories WHERE legal_entity_id = p_legal_entity_id AND code = 'IT';
    END IF;

    INSERT INTO fixedassets.categories (
        legal_entity_id, code, name_en, name_ar, description_en, description_ar,
        depreciation_method, useful_life_months, salvage_value_rate,
        asset_account_id, accumulated_depreciation_account_id, depreciation_expense_account_id
    ) VALUES
        (p_legal_entity_id, 'FUR', 'Furniture & Fixtures', 'الأثاث والتجهيزات المكتبية',
         'Office desks, chairs, meeting room fittings', 'المكاتب والمقاعد وتجهيزات قاعات الاجتماعات',
         'STRAIGHT_LINE', 60, 0.0000, v_acc_fur_id, v_acc_depr_id, v_exp_depr_id)
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_cat_fur_id;
    IF v_cat_fur_id IS NULL THEN
        SELECT id INTO v_cat_fur_id FROM fixedassets.categories WHERE legal_entity_id = p_legal_entity_id AND code = 'FUR';
    END IF;

    -- 5. Seed Sample Assets
    INSERT INTO fixedassets.assets (
        legal_entity_id, category_id, location_id, asset_number, serial_number, barcode,
        name_en, name_ar, description_en, description_ar,
        acquisition_date, in_service_date, acquisition_cost, salvage_value, useful_life_months,
        depreciation_method, status, current_net_book_value, accumulated_depreciation
    ) VALUES
        (p_legal_entity_id, v_cat_veh_id, v_loc_hq_id, 'AST-VEH-001', 'VIN-8933214567', 'BC-88901',
         'Toyota Hilux Commercial Truck 2024', 'شاحنة تويوتا هايلوكس تجارية 2024م',
         'Logistics and technical support truck', 'شاحنة الدعم الفني والخدمات اللوجستية',
         '2024-01-01', '2024-01-01', 120000.000000, 12000.000000, 48,
         'STRAIGHT_LINE', 'ACTIVE', 120000.000000, 0.000000),
        (p_legal_entity_id, v_cat_it_id, v_loc_hq_id, 'AST-IT-001', 'SRV-DELL-99128', 'BC-77201',
         'Dell PowerEdge Enterprise Server R750', 'خادم شبكة رئيسي ديل باور إيدج R750',
         'Main database and application server', 'خادم قواعد البيانات والتطبيقات الرئيسي',
         '2024-01-01', '2024-01-01', 72000.000000, 0.000000, 36,
         'STRAIGHT_LINE', 'ACTIVE', 72000.000000, 0.000000),
        (p_legal_entity_id, v_cat_fur_id, v_loc_hq_id, 'AST-FUR-001', 'FUR-CONF-0044', 'BC-66301',
         'Executive Conference Room Table & 16 Chairs', 'طاولة اجتماعات تنفيذية مع 16 مقعداً',
         'Boardroom conference setup', 'تجهيزات قاعة اجتماعات مجلس الإدارة',
         '2024-01-01', '2024-01-01', 30000.000000, 0.000000, 60,
         'STRAIGHT_LINE', 'ACTIVE', 30000.000000, 0.000000)
    ON CONFLICT (legal_entity_id, asset_number) DO NOTHING;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS fixedassets.seed_default_fixed_assets_data(UUID);
DROP TABLE IF EXISTS fixedassets.disposals CASCADE;
DROP TABLE IF EXISTS fixedassets.impairments CASCADE;
DROP TABLE IF EXISTS fixedassets.transfers CASCADE;
DROP TABLE IF EXISTS fixedassets.depreciation_lines CASCADE;
DROP TABLE IF EXISTS fixedassets.depreciation_runs CASCADE;
DROP TABLE IF EXISTS fixedassets.depreciation_schedules CASCADE;
DROP TABLE IF EXISTS fixedassets.depreciation_books CASCADE;
DROP TABLE IF EXISTS fixedassets.asset_assignments CASCADE;
DROP TABLE IF EXISTS fixedassets.assets CASCADE;
DROP TABLE IF EXISTS fixedassets.asset_locations CASCADE;
DROP TABLE IF EXISTS fixedassets.categories CASCADE;
DROP SCHEMA IF EXISTS fixedassets CASCADE;
