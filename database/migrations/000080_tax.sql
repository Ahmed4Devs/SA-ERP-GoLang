-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS tax;

-- 1. Standard Tax Categories (ZATCA / UN-ECE 5305 Codes: S, Z, E, O)
CREATE TABLE IF NOT EXISTS tax.tax_categories (
    code            TEXT PRIMARY KEY,
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    description_en  TEXT,
    description_ar  TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO tax.tax_categories (code, name_en, name_ar, description_en, description_ar) VALUES
    ('S', 'Standard Rate', 'النسبة الأساسية (15%)', 'Goods and services subject to standard VAT rate', 'السلع والخدمات الخاضعة لنسبة الضريبة الأساسية 15%'),
    ('Z', 'Zero Rated', 'نسبة الصفر (0%)', 'Goods and services taxed at 0% with input tax deduction rights', 'السلع والخدمات الخاضعة لنسبة الصفر مع حق استرداد ضريبة المدخلات'),
    ('E', 'Exempt', 'معفى من الضريبة', 'Financial services and residential rent exempt from VAT', 'الأنشطة المعفاة نظاماً كالخدمات المالية وتأجير العقار السكني'),
    ('O', 'Out of Scope', 'خارج نطاق الضريبة', 'Supplies outside the scope of VAT regulation', 'التوريدات والأنشطة الواقعة خارج نطاق نظام ضريبة القيمة المضافة')
ON CONFLICT (code) DO NOTHING;

-- 2. ZATCA Exemption Reasons (Mandatory when category is 'E')
CREATE TABLE IF NOT EXISTS tax.exemption_reasons (
    code            TEXT PRIMARY KEY,
    category_code   TEXT NOT NULL REFERENCES tax.tax_categories(code),
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    description_en  TEXT,
    description_ar  TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO tax.exemption_reasons (code, category_code, name_en, name_ar, description_en, description_ar) VALUES
    ('VATEX-SA-29', 'E', 'Financial services', 'الخدمات المالية المعفاة', 'Financial services mentioned in Article 29 of the VAT Regulations', 'الخدمات المالية المنصوص عليها في المادة 29 من اللائحة التنفيذية'),
    ('VATEX-SA-30', 'E', 'Life insurance services', 'عقود التأمين على الحياة', 'Life insurance contracts mentioned in Article 30 of the VAT Regulations', 'عقود التأمين على الحياة وفق المادة 30 من اللائحة التنفيذية'),
    ('VATEX-SA-31', 'E', 'Real estate lease', 'تأجير العقارات السكنية', 'Lease of residential real estate exempt from VAT', 'تأجير العقارات السكنية المعفى من ضريبة القيمة المضافة')
ON CONFLICT (code) DO NOTHING;

-- 3. ZATCA Zero Rate Reasons (Mandatory when category is 'Z')
CREATE TABLE IF NOT EXISTS tax.zero_rate_reasons (
    code            TEXT PRIMARY KEY,
    category_code   TEXT NOT NULL REFERENCES tax.tax_categories(code),
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    description_en  TEXT,
    description_ar  TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO tax.zero_rate_reasons (code, category_code, name_en, name_ar, description_en, description_ar) VALUES
    ('VATEX-SA-32', 'Z', 'Export of goods', 'صادرات السلع خارج دول المجلس', 'Export of goods outside GCC territory', 'تصدير السلع إلى خارج إقليم دول مجلس التعاون الخليجي'),
    ('VATEX-SA-33', 'Z', 'Export of services', 'صادرات الخدمات خارج دول المجلس', 'Export of services to non-GCC residents', 'تصدير الخدمات إلى أشخاص غير مقيمين في دول المجلس'),
    ('VATEX-SA-34', 'Z', 'International transport', 'خدمات النقل الدولي وركابها', 'International transport services of goods or passengers', 'خدمات النقل الدولي للركاب والبضائع والخدمات المرتبطة بها'),
    ('VATEX-SA-35', 'Z', 'Qualified medicines & equipment', 'الأدوية والمعدات الطبية المؤهلة', 'Medicines and medical equipment approved by SFDA', 'الأدوية والمعدات الطبية المؤهلة والمعتمدة من هيئة الغذاء والدواء'),
    ('VATEX-SA-36', 'Z', 'Qualified investment metals', 'المعادن الاستثمارية المؤهلة', 'Gold, silver and platinum of 99% purity for investment', 'الذهب والفضة والبلاتين الاستثماري بنقاء لا يقل عن 99%'),
    ('VATEX-SA-EDU', 'Z', 'Private education for citizens', 'التعليم الأهلي للمواطنين', 'Private education services provided to Saudi citizens', 'خدمات التعليم الأهلي المقدمة للمواطنين وتتحملها الدولة'),
    ('VATEX-SA-HEA', 'Z', 'Private healthcare for citizens', 'الرعاية الصحية الأهلية للمواطنين', 'Private healthcare services provided to Saudi citizens', 'خدمات الرعاية الصحية الأهلية المقدمة للمواطنين وتتحملها الدولة')
ON CONFLICT (code) DO NOTHING;

-- 4. Tax Codes (Entity-Scoped Configuration)
CREATE TABLE IF NOT EXISTS tax.tax_codes (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                    TEXT NOT NULL,
    name_en                 TEXT NOT NULL,
    name_ar                 TEXT NOT NULL,
    tax_category_code       TEXT NOT NULL REFERENCES tax.tax_categories(code),
    tax_type                TEXT NOT NULL CHECK (tax_type IN ('OUTPUT', 'INPUT', 'BOTH')),
    sales_tax_account_id    UUID REFERENCES accounting.accounts(id) ON DELETE RESTRICT,
    purchase_tax_account_id UUID REFERENCES accounting.accounts(id) ON DELETE RESTRICT,
    is_recoverable          BOOLEAN NOT NULL DEFAULT true,
    is_active               BOOLEAN NOT NULL DEFAULT true,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    version                 BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_tax_codes_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_tax_codes_entity
ON tax.tax_codes (legal_entity_id, is_active);

-- 5. Tax Rate Versions (Effective-Dated Rates for Historical Reproducibility)
CREATE TABLE IF NOT EXISTS tax.tax_rate_versions (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    tax_code_id     UUID NOT NULL REFERENCES tax.tax_codes(id) ON DELETE CASCADE,
    rate_percent    NUMERIC(7, 4) NOT NULL CHECK (rate_percent >= 0),
    effective_from  DATE NOT NULL,
    effective_to    DATE,
    description     TEXT,
    created_by      UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ck_tax_rate_dates
        CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE INDEX IF NOT EXISTS idx_tax_rates_lookup
ON tax.tax_rate_versions (tax_code_id, effective_from, effective_to);

-- 6. Tax Reporting Periods (Monthly / Quarterly for ZATCA)
CREATE TABLE IF NOT EXISTS tax.tax_periods (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    period_number   INT NOT NULL,
    fiscal_year     INT NOT NULL,
    period_type     TEXT NOT NULL CHECK (period_type IN ('MONTHLY', 'QUARTERLY')),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    due_date        DATE NOT NULL,
    status          TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'FILED', 'CLOSED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tax_periods_entity_range
        UNIQUE (legal_entity_id, fiscal_year, period_type, period_number),
    CONSTRAINT ck_tax_period_dates
        CHECK (end_date >= start_date AND due_date >= end_date)
);

CREATE INDEX IF NOT EXISTS idx_tax_periods_entity_status
ON tax.tax_periods (legal_entity_id, status);

-- 7. Tax Transactions (General Tax Ledger Header)
CREATE TABLE IF NOT EXISTS tax.tax_transactions (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    tax_period_id           UUID REFERENCES tax.tax_periods(id) ON DELETE RESTRICT,
    source_module           TEXT NOT NULL,
    source_type             TEXT NOT NULL,
    source_id               UUID NOT NULL,
    transaction_number      TEXT NOT NULL,
    transaction_date        DATE NOT NULL,
    transaction_type        TEXT NOT NULL CHECK (transaction_type IN ('OUTPUT', 'INPUT')),
    partner_id              UUID,
    partner_name            TEXT,
    partner_tax_number      TEXT,
    total_taxable_amount    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_tax_amount        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_gross_amount      NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    journal_entry_id        UUID REFERENCES accounting.journal_entries(id) ON DELETE SET NULL,
    status                  TEXT NOT NULL DEFAULT 'POSTED' CHECK (status IN ('DRAFT', 'POSTED', 'CANCELLED', 'REVERSED')),
    created_by              UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tax_transactions_source
        UNIQUE (legal_entity_id, source_module, source_type, source_id)
);

CREATE INDEX IF NOT EXISTS idx_tax_transactions_entity_date
ON tax.tax_transactions (legal_entity_id, transaction_date, transaction_type);

CREATE INDEX IF NOT EXISTS idx_tax_transactions_period
ON tax.tax_transactions (tax_period_id);

-- 8. Tax Transaction Lines (Item-Level Breakdown)
CREATE TABLE IF NOT EXISTS tax.tax_transaction_lines (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    tax_transaction_id      UUID NOT NULL REFERENCES tax.tax_transactions(id) ON DELETE CASCADE,
    line_number             INT NOT NULL,
    tax_code_id             UUID NOT NULL REFERENCES tax.tax_codes(id) ON DELETE RESTRICT,
    tax_category_code       TEXT NOT NULL REFERENCES tax.tax_categories(code),
    tax_rate_percent        NUMERIC(7, 4) NOT NULL CHECK (tax_rate_percent >= 0),
    taxable_amount          NUMERIC(18, 4) NOT NULL,
    tax_amount              NUMERIC(18, 4) NOT NULL,
    gross_amount            NUMERIC(18, 4) NOT NULL,
    exemption_reason_code   TEXT REFERENCES tax.exemption_reasons(code),
    zero_rate_reason_code   TEXT REFERENCES tax.zero_rate_reasons(code),
    is_recoverable          BOOLEAN NOT NULL DEFAULT true,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tax_trans_lines_seq
        UNIQUE (tax_transaction_id, line_number)
);

-- 9. Tax Returns (ZATCA Official 16-Box VAT Declaration Form)
CREATE TABLE IF NOT EXISTS tax.tax_returns (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    tax_period_id           UUID NOT NULL REFERENCES tax.tax_periods(id) ON DELETE RESTRICT,
    return_number           TEXT NOT NULL,
    filing_date             DATE,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'SUBMITTED', 'PAID', 'AMENDED')),
    total_sales_taxable     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_sales_vat         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_purchases_taxable NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_purchases_vat     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    net_vat_due             NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    declaration_boxes       JSONB NOT NULL DEFAULT '{}'::jsonb,
    submitted_at            TIMESTAMPTZ,
    submitted_by            UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    version                 BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_tax_returns_period
        UNIQUE (legal_entity_id, tax_period_id)
);

-- 10. Seed Permissions
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('tax.view', 'عرض الضرائب والإقرارات', 'View Tax & Returns', 'TAX'),
    ('tax.manage', 'إدارة الرموز والنسب الضريبية', 'Manage Tax Codes & Rates', 'TAX'),
    ('tax.calculate', 'حساب الضريبة والمعاينة', 'Calculate Tax & Preview', 'TAX'),
    ('tax.post_transaction', 'ترحيل المعاملات الضريبية', 'Post Tax Transactions', 'TAX'),
    ('tax.file_return', 'إعداد واعتماد إقرار الضريبة', 'File VAT Returns', 'TAX')
ON CONFLICT (code) DO NOTHING;

-- Grant permissions to SUPER_ADMIN role
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN' AND p.code LIKE 'tax.%'
ON CONFLICT DO NOTHING;

-- 11. Helper Function: Seed Default Saudi VAT Structure for a Legal Entity
CREATE OR REPLACE FUNCTION tax.seed_default_tax_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_output_tax_acc_id UUID;
    v_input_tax_acc_id  UUID;
    v_std_tax_code_id   UUID;
    v_zero_tax_code_id  UUID;
    v_exempt_tax_code_id UUID;
    v_oos_tax_code_id   UUID;
    m INT;
    v_start DATE;
    v_end DATE;
    v_due DATE;
BEGIN
    -- 1. Identify standard VAT accounts from accounting (seeded in 000070)
    SELECT id INTO v_output_tax_acc_id
    FROM accounting.accounts
    WHERE legal_entity_id = p_legal_entity_id AND code = '213000'
    LIMIT 1;

    SELECT id INTO v_input_tax_acc_id
    FROM accounting.accounts
    WHERE legal_entity_id = p_legal_entity_id AND code = '114000'
    LIMIT 1;

    -- 2. Standard 15% VAT Code (VAT-STD-15)
    INSERT INTO tax.tax_codes (
        legal_entity_id, code, name_en, name_ar, tax_category_code, tax_type,
        sales_tax_account_id, purchase_tax_account_id, is_recoverable
    ) VALUES (
        p_legal_entity_id, 'VAT-STD-15', 'Standard Rate 15% VAT', 'ضريبة القيمة المضافة بالنسبة الأساسية 15%',
        'S', 'BOTH', v_output_tax_acc_id, v_input_tax_acc_id, true
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET
        sales_tax_account_id = EXCLUDED.sales_tax_account_id,
        purchase_tax_account_id = EXCLUDED.purchase_tax_account_id
    RETURNING id INTO v_std_tax_code_id;

    -- Seed Historical 5% Rate (2018-01-01 to 2020-06-30)
    INSERT INTO tax.tax_rate_versions (
        tax_code_id, rate_percent, effective_from, effective_to, description
    ) VALUES (
        v_std_tax_code_id, 5.0000, '2018-01-01', '2020-06-30', 'Historical 5% VAT rate in KSA'
    ) ON CONFLICT DO NOTHING;

    -- Seed Current 15% Rate (2020-07-01 onwards)
    INSERT INTO tax.tax_rate_versions (
        tax_code_id, rate_percent, effective_from, effective_to, description
    ) VALUES (
        v_std_tax_code_id, 15.0000, '2020-07-01', NULL, 'Current 15% standard VAT rate in KSA'
    ) ON CONFLICT DO NOTHING;

    -- 3. Zero-Rated Code (VAT-ZERO-0)
    INSERT INTO tax.tax_codes (
        legal_entity_id, code, name_en, name_ar, tax_category_code, tax_type,
        sales_tax_account_id, purchase_tax_account_id, is_recoverable
    ) VALUES (
        p_legal_entity_id, 'VAT-ZERO-0', 'Zero Rated 0% VAT', 'ضريبة القيمة المضافة بنسبة الصفر 0%',
        'Z', 'BOTH', v_output_tax_acc_id, v_input_tax_acc_id, true
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET
        sales_tax_account_id = EXCLUDED.sales_tax_account_id
    RETURNING id INTO v_zero_tax_code_id;

    INSERT INTO tax.tax_rate_versions (
        tax_code_id, rate_percent, effective_from, effective_to, description
    ) VALUES (
        v_zero_tax_code_id, 0.0000, '2018-01-01', NULL, 'Zero-rated VAT rate'
    ) ON CONFLICT DO NOTHING;

    -- 4. Exempt Code (VAT-EXEMPT-0)
    INSERT INTO tax.tax_codes (
        legal_entity_id, code, name_en, name_ar, tax_category_code, tax_type,
        sales_tax_account_id, purchase_tax_account_id, is_recoverable
    ) VALUES (
        p_legal_entity_id, 'VAT-EXEMPT-0', 'Exempt from VAT', 'معفى من ضريبة القيمة المضافة',
        'E', 'BOTH', v_output_tax_acc_id, v_input_tax_acc_id, false
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET
        sales_tax_account_id = EXCLUDED.sales_tax_account_id
    RETURNING id INTO v_exempt_tax_code_id;

    INSERT INTO tax.tax_rate_versions (
        tax_code_id, rate_percent, effective_from, effective_to, description
    ) VALUES (
        v_exempt_tax_code_id, 0.0000, '2018-01-01', NULL, 'Exempt supplies'
    ) ON CONFLICT DO NOTHING;

    -- 5. Out of Scope Code (VAT-OOS-0)
    INSERT INTO tax.tax_codes (
        legal_entity_id, code, name_en, name_ar, tax_category_code, tax_type,
        sales_tax_account_id, purchase_tax_account_id, is_recoverable
    ) VALUES (
        p_legal_entity_id, 'VAT-OOS-0', 'Out of Scope', 'خارج نطاق ضريبة القيمة المضافة',
        'O', 'BOTH', NULL, NULL, false
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET
        is_active = EXCLUDED.is_active
    RETURNING id INTO v_oos_tax_code_id;

    INSERT INTO tax.tax_rate_versions (
        tax_code_id, rate_percent, effective_from, effective_to, description
    ) VALUES (
        v_oos_tax_code_id, 0.0000, '2018-01-01', NULL, 'Out of scope transactions'
    ) ON CONFLICT DO NOTHING;

    -- 6. Generate 12 Monthly Tax Periods for 2026
    FOR m IN 1..12 LOOP
        v_start := ('2026-' || LPAD(m::TEXT, 2, '0') || '-01')::DATE;
        v_end := (v_start + INTERVAL '1 month' - INTERVAL '1 day')::DATE;
        v_due := (v_start + INTERVAL '2 months' - INTERVAL '1 day')::DATE; -- Due last day of subsequent month

        INSERT INTO tax.tax_periods (
            legal_entity_id, period_number, fiscal_year, period_type, start_date, end_date, due_date, status
        ) VALUES (
            p_legal_entity_id, m, 2026, 'MONTHLY', v_start, v_end, v_due, 'OPEN'
        ) ON CONFLICT (legal_entity_id, fiscal_year, period_type, period_number) DO NOTHING;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Automatically seed default tax data for existing legal entities
DO $$
DECLARE
    le RECORD;
BEGIN
    FOR le IN SELECT id FROM organization.legal_entities WHERE is_active = true LOOP
        PERFORM tax.seed_default_tax_data(le.id);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tax.tax_returns;
DROP TABLE IF EXISTS tax.tax_transaction_lines;
DROP TABLE IF EXISTS tax.tax_transactions;
DROP TABLE IF EXISTS tax.tax_periods;
DROP TABLE IF EXISTS tax.tax_rate_versions;
DROP TABLE IF EXISTS tax.tax_codes;
DROP TABLE IF EXISTS tax.zero_rate_reasons;
DROP TABLE IF EXISTS tax.exemption_reasons;
DROP TABLE IF EXISTS tax.tax_categories;

DELETE FROM authz.role_permissions WHERE permission_id IN (
    SELECT id FROM authz.permissions WHERE code LIKE 'tax.%'
);
DELETE FROM authz.permissions WHERE code LIKE 'tax.%';

DROP SCHEMA IF EXISTS tax CASCADE;
-- +goose StatementEnd
