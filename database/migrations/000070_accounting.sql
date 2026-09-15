-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS accounting;

-- 1. Chart of Accounts (COA Header per Legal Entity)
CREATE TABLE IF NOT EXISTS accounting.chart_of_accounts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    description         TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_chart_of_accounts_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_chart_of_accounts_entity
ON accounting.chart_of_accounts (legal_entity_id, is_active);

-- 2. General Ledger Accounts (Hierarchical Tree)
CREATE TABLE IF NOT EXISTS accounting.accounts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    chart_id            UUID NOT NULL REFERENCES accounting.chart_of_accounts(id) ON DELETE CASCADE,
    parent_id           UUID REFERENCES accounting.accounts(id) ON DELETE RESTRICT,
    code                TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    account_type        TEXT NOT NULL,
    normal_balance      TEXT NOT NULL,
    allow_posting       BOOLEAN NOT NULL DEFAULT true,
    is_control_account  BOOLEAN NOT NULL DEFAULT false,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_accounts_entity_code
        UNIQUE (legal_entity_id, code),
    CONSTRAINT ck_accounts_type
        CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
    CONSTRAINT ck_accounts_normal_balance
        CHECK (normal_balance IN ('DEBIT', 'CREDIT'))
);

CREATE INDEX IF NOT EXISTS idx_accounts_entity_chart
ON accounting.accounts (legal_entity_id, chart_id, code);
CREATE INDEX IF NOT EXISTS idx_accounts_parent
ON accounting.accounts (parent_id);
CREATE INDEX IF NOT EXISTS idx_accounts_type
ON accounting.accounts (legal_entity_id, account_type);

-- 3. Fiscal Years
CREATE TABLE IF NOT EXISTS accounting.fiscal_years (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'OPEN',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_fiscal_years_entity_code
        UNIQUE (legal_entity_id, code),
    CONSTRAINT ck_fiscal_years_dates
        CHECK (end_date > start_date),
    CONSTRAINT ck_fiscal_years_status
        CHECK (status IN ('OPEN', 'CLOSED'))
);

CREATE INDEX IF NOT EXISTS idx_fiscal_years_entity
ON accounting.fiscal_years (legal_entity_id, start_date, end_date);

-- 4. Fiscal Periods (Monthly breakdown)
CREATE TABLE IF NOT EXISTS accounting.fiscal_periods (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    fiscal_year_id      UUID NOT NULL REFERENCES accounting.fiscal_years(id) ON DELETE CASCADE,
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    period_number       SMALLINT NOT NULL,
    period_name_en      TEXT NOT NULL,
    period_name_ar      TEXT NOT NULL,
    quarter             SMALLINT NOT NULL,
    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'OPEN',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_fiscal_periods_year_num
        UNIQUE (fiscal_year_id, period_number),
    CONSTRAINT ck_fiscal_periods_quarter
        CHECK (quarter BETWEEN 1 AND 4),
    CONSTRAINT ck_fiscal_periods_dates
        CHECK (end_date >= start_date),
    CONSTRAINT ck_fiscal_periods_status
        CHECK (status IN ('OPEN', 'SOFT_CLOSED', 'CLOSED', 'LOCKED'))
);

CREATE INDEX IF NOT EXISTS idx_fiscal_periods_lookup
ON accounting.fiscal_periods (legal_entity_id, start_date, end_date, status);

-- 5. Journals (Subledgers)
CREATE TABLE IF NOT EXISTS accounting.journals (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    journal_type        TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_journals_entity_code
        UNIQUE (legal_entity_id, code),
    CONSTRAINT ck_journals_type
        CHECK (journal_type IN ('GENERAL', 'SALES', 'PURCHASE', 'CASH', 'BANK', 'PAYROLL', 'INVENTORY', 'ADJUSTMENT'))
);

CREATE INDEX IF NOT EXISTS idx_journals_entity
ON accounting.journals (legal_entity_id, is_active);

-- 6. Journal Entries (General Ledger Header)
CREATE TABLE IF NOT EXISTS accounting.journal_entries (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    journal_id          UUID NOT NULL REFERENCES accounting.journals(id) ON DELETE RESTRICT,
    fiscal_period_id    UUID NOT NULL REFERENCES accounting.fiscal_periods(id) ON DELETE RESTRICT,
    entry_number        TEXT NOT NULL,
    accounting_date     DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'DRAFT',
    source_module       TEXT NOT NULL DEFAULT 'GL',
    source_type         TEXT,
    source_id           UUID,
    reference           TEXT,
    description         TEXT NOT NULL,
    currency_code       CHAR(3) NOT NULL DEFAULT 'SAR' REFERENCES ref.currencies(code),
    exchange_rate       NUMERIC(18, 6) NOT NULL DEFAULT 1.000000,
    total_debit         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_credit        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    created_by          UUID NOT NULL REFERENCES iam.users(id) ON DELETE RESTRICT,
    approved_by         UUID REFERENCES iam.users(id) ON DELETE RESTRICT,
    posted_by           UUID REFERENCES iam.users(id) ON DELETE RESTRICT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at         TIMESTAMPTZ,
    posted_at           TIMESTAMPTZ,
    reversed_at         TIMESTAMPTZ,
    reversal_of_id      UUID REFERENCES accounting.journal_entries(id) ON DELETE RESTRICT,
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_journal_entries_entity_num
        UNIQUE (legal_entity_id, entry_number),
    CONSTRAINT ck_journal_entries_status
        CHECK (status IN ('DRAFT', 'SUBMITTED', 'APPROVED', 'POSTED', 'REVERSED', 'CANCELLED')),
    CONSTRAINT ck_journal_entries_balance
        CHECK (status NOT IN ('POSTED', 'REVERSED') OR total_debit = total_credit)
);

CREATE INDEX IF NOT EXISTS idx_journal_entries_entity_date
ON accounting.journal_entries (legal_entity_id, accounting_date DESC);
CREATE INDEX IF NOT EXISTS idx_journal_entries_period
ON accounting.journal_entries (fiscal_period_id, status);
CREATE INDEX IF NOT EXISTS idx_journal_entries_source
ON accounting.journal_entries (source_module, source_id);

-- 7. Journal Lines (Financial Ledger Lines)
CREATE TABLE IF NOT EXISTS accounting.journal_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    journal_entry_id    UUID NOT NULL REFERENCES accounting.journal_entries(id) ON DELETE CASCADE,
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    line_number         SMALLINT NOT NULL,
    account_id          UUID NOT NULL REFERENCES accounting.accounts(id) ON DELETE RESTRICT,
    description         TEXT,
    currency_code       CHAR(3) NOT NULL DEFAULT 'SAR' REFERENCES ref.currencies(code),
    exchange_rate       NUMERIC(18, 6) NOT NULL DEFAULT 1.000000,
    transaction_debit   NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    transaction_credit  NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    base_debit          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    base_credit         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    branch_id           UUID REFERENCES organization.branches(id) ON DELETE SET NULL,
    department_id       UUID REFERENCES organization.departments(id) ON DELETE SET NULL,
    cost_center_id      UUID REFERENCES organization.cost_centers(id) ON DELETE SET NULL,

    CONSTRAINT uq_journal_lines_entry_line
        UNIQUE (journal_entry_id, line_number),
    CONSTRAINT ck_journal_lines_xor_amount
        CHECK ((transaction_debit > 0 AND transaction_credit = 0) OR (transaction_credit > 0 AND transaction_debit = 0)),
    CONSTRAINT ck_journal_lines_positive
        CHECK (transaction_debit >= 0 AND transaction_credit >= 0 AND base_debit >= 0 AND base_credit >= 0)
);

CREATE INDEX IF NOT EXISTS idx_journal_lines_account_date
ON accounting.journal_lines (legal_entity_id, account_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_entry
ON accounting.journal_lines (journal_entry_id);

-- 8. Account Balances (Aggregated Periodic Ledger Balances)
CREATE TABLE IF NOT EXISTS accounting.account_balances (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    account_id          UUID NOT NULL REFERENCES accounting.accounts(id) ON DELETE CASCADE,
    fiscal_period_id    UUID NOT NULL REFERENCES accounting.fiscal_periods(id) ON DELETE CASCADE,
    opening_balance     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    period_debit        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    period_credit       NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    closing_balance     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_account_balances_entity_acc_period
        UNIQUE (legal_entity_id, account_id, fiscal_period_id)
);

CREATE INDEX IF NOT EXISTS idx_account_balances_lookup
ON accounting.account_balances (legal_entity_id, fiscal_period_id, account_id);

-- 9. Accounting Immutability Triggers (Section 40.8)
CREATE OR REPLACE FUNCTION accounting.prevent_posted_journal_mutation()
RETURNS TRIGGER AS $$
DECLARE
    v_status TEXT;
BEGIN
    IF TG_TABLE_NAME = 'journal_entries' THEN
        IF TG_OP = 'DELETE' THEN
            IF OLD.status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot DELETE posted or reversed journal entry (ID: %)', OLD.id;
            END IF;
            RETURN OLD;
        ELSIF TG_OP = 'UPDATE' THEN
            IF OLD.status IN ('POSTED', 'REVERSED') THEN
                -- Allow ONLY transitioning from POSTED to REVERSED when reversal metadata is recorded
                IF NOT (OLD.status = 'POSTED' AND NEW.status = 'REVERSED' AND NEW.reversal_of_id IS NOT NULL) THEN
                    RAISE EXCEPTION 'Cannot UPDATE posted or reversed journal entry (ID: %)', OLD.id;
                END IF;
            END IF;
            RETURN NEW;
        END IF;
    ELSIF TG_TABLE_NAME = 'journal_lines' THEN
        IF TG_OP IN ('DELETE', 'UPDATE') THEN
            SELECT status INTO v_status FROM accounting.journal_entries WHERE id = OLD.journal_entry_id;
            IF v_status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot % journal lines for posted or reversed journal entry (ID: %)', TG_OP, OLD.journal_entry_id;
            END IF;
            IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
        ELSIF TG_OP = 'INSERT' THEN
            SELECT status INTO v_status FROM accounting.journal_entries WHERE id = NEW.journal_entry_id;
            IF v_status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot INSERT journal lines into posted or reversed journal entry (ID: %)', NEW.journal_entry_id;
            END IF;
            RETURN NEW;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_posted_journal_entry_mutation ON accounting.journal_entries;
CREATE TRIGGER trg_prevent_posted_journal_entry_mutation
BEFORE UPDATE OR DELETE ON accounting.journal_entries
FOR EACH ROW
EXECUTE FUNCTION accounting.prevent_posted_journal_mutation();

DROP TRIGGER IF EXISTS trg_prevent_posted_journal_lines_mutation ON accounting.journal_lines;
CREATE TRIGGER trg_prevent_posted_journal_lines_mutation
BEFORE INSERT OR UPDATE OR DELETE ON accounting.journal_lines
FOR EACH ROW
EXECUTE FUNCTION accounting.prevent_posted_journal_mutation();

-- 10. Register Accounting Permissions
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('gl.journals.view', 'عرض القيود اليومية', 'View Journal Entries', 'ACCOUNTING'),
    ('gl.journals.post', 'ترحيل القيود اليومية', 'Post Journal Entries', 'ACCOUNTING'),
    ('gl.accounts.view', 'عرض دليل الحسابات', 'View Chart of Accounts', 'ACCOUNTING'),
    ('accounting.accounts.view', 'عرض دليل الحسابات وتفاصيلها', 'View Chart of Accounts', 'ACCOUNTING'),
    ('accounting.accounts.create', 'إنشاء حسابات جديدة في الدليل', 'Create Accounts', 'ACCOUNTING'),
    ('accounting.accounts.update', 'تعديل وتحديث بيانات الحسابات', 'Update Accounts', 'ACCOUNTING'),
    ('accounting.fiscal.view', 'عرض السنوات والفترات المالية', 'View Fiscal Years and Periods', 'ACCOUNTING'),
    ('accounting.fiscal.manage', 'إدارة وإقفال الفترات والسنوات المالية', 'Manage Fiscal Periods', 'ACCOUNTING'),
    ('accounting.journals.view', 'عرض دفاتر اليومية المتخصصة', 'View Journals', 'ACCOUNTING'),
    ('accounting.journal_entries.view', 'عرض قيود اليومية وسطورها', 'View Journal Entries', 'ACCOUNTING'),
    ('accounting.journal_entries.create', 'إنشاء مسودات قيود اليومية', 'Create Journal Entries', 'ACCOUNTING'),
    ('accounting.journal_entries.post', 'ترحيل القيود المتوازنة للأستاذ العام', 'Post Journal Entries', 'ACCOUNTING'),
    ('accounting.journal_entries.reverse', 'عكس القيود المحاسبية المرحّلة', 'Reverse Journal Entries', 'ACCOUNTING'),
    ('accounting.reports.trial_balance', 'استخراج وعرض ميزان المراجعة', 'View Trial Balance Report', 'ACCOUNTING'),
    ('accounting.reports.general_ledger', 'عرض كشف حساب الأستاذ العام', 'View General Ledger Report', 'ACCOUNTING')
ON CONFLICT (code) DO NOTHING;

-- Grant permissions to SUPER_ADMIN and CHIEF_ACCOUNTANT roles
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code IN ('SUPER_ADMIN', 'CHIEF_ACCOUNTANT')
  AND p.module = 'ACCOUNTING'
ON CONFLICT DO NOTHING;

-- 11. Function to Seed Canonical Saudi Chart of Accounts, Fiscal Year 2026, and Journals for a Legal Entity
CREATE OR REPLACE FUNCTION accounting.seed_default_coa(p_legal_entity_id UUID)
RETURNS void AS $$
DECLARE
    v_chart_id UUID;
    v_fy_id UUID;
    v_asset_id UUID;
    v_ca_id UUID;
    v_cash_id UUID;
    v_liab_id UUID;
    v_cl_id UUID;
    v_eq_id UUID;
    v_rev_id UUID;
    v_exp_id UUID;
    v_opex_id UUID;
    m INT;
    v_start DATE;
    v_end DATE;
    v_month_names_en TEXT[] := ARRAY['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
    v_month_names_ar TEXT[] := ARRAY['يناير', 'فبراير', 'مارس', 'أبريل', 'مايو', 'يونيو', 'يوليو', 'أغسطس', 'سبتمبر', 'أكتوبر', 'نوفمبر', 'ديسمبر'];
BEGIN
    -- A. Create Chart of Accounts
    INSERT INTO accounting.chart_of_accounts (
        legal_entity_id, code, name_en, name_ar, description
    ) VALUES (
        p_legal_entity_id, 'COA-KSA-STD', 'Saudi Standard Commercial Chart of Accounts', 'دليل الحسابات الموحد للشركات السعودية (SOCPA/IFRS)', 'دليل الحسابات المعتمد المتوافق مع متطلبات المعايير المحاسبية وهيئة الزكاة والضريبة والجمارك'
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET is_active = true
    RETURNING id INTO v_chart_id;

    -- B. Seed 5 Major Level 1 Account Categories
    -- 1. ASSETS
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, '100000', 'Assets', 'الأصول', 'ASSET', 'DEBIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_asset_id;

    -- 110000 Current Assets
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, v_asset_id, '110000', 'Current Assets', 'الأصول المتداولة', 'ASSET', 'DEBIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_ca_id;

    -- 111000 Cash and Cash Equivalents
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, v_ca_id, '111000', 'Cash and Cash Equivalents', 'النقد وما في حكمه', 'ASSET', 'DEBIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_cash_id;

    -- Posting Cash Accounts
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting) VALUES
        (p_legal_entity_id, v_chart_id, v_cash_id, '111001', 'Petty Cash - Head Office', 'صندوق المصروفات النثرية - الإدارة العامة', 'ASSET', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_cash_id, '111002', 'Al-Rajhi Bank - Operating SAR', 'مصرف الراجحي - الحساب الجاري بالريال', 'ASSET', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_cash_id, '111003', 'SNB Bank - Corporate SAR', 'البنك الأهلي السعودي - حساب الشركات', 'ASSET', 'DEBIT', true)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- Other Current Assets
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account) VALUES
        (p_legal_entity_id, v_chart_id, v_ca_id, '112000', 'Trade Receivables (AR)', 'العملاء والمدينون التجاريون', 'ASSET', 'DEBIT', true, true),
        (p_legal_entity_id, v_chart_id, v_ca_id, '113000', 'Inventory (Merchandise)', 'المخزون السلعي والبضاعة', 'ASSET', 'DEBIT', true, true),
        (p_legal_entity_id, v_chart_id, v_ca_id, '114000', 'Input VAT (15%) - Recoverable', 'ضريبة القيمة المضافة المدخلات (المشتريات)', 'ASSET', 'DEBIT', true, false),
        (p_legal_entity_id, v_chart_id, v_ca_id, '115000', 'Prepaid Expenses', 'المصروفات المدفوعة مقدماً', 'ASSET', 'DEBIT', true, false)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 120000 Non-Current Assets
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, v_asset_id, '120000', 'Non-Current Assets', 'الأصول غير المتداولة', 'ASSET', 'DEBIT', false)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 2. LIABILITIES
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, '200000', 'Liabilities', 'الخصوم والالتزامات', 'LIABILITY', 'CREDIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_liab_id;

    -- 210000 Current Liabilities
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, v_liab_id, '210000', 'Current Liabilities', 'الخصوم المتداولة', 'LIABILITY', 'CREDIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_cl_id;

    -- Posting Liabilities Accounts
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting, is_control_account) VALUES
        (p_legal_entity_id, v_chart_id, v_cl_id, '211000', 'Trade Payables (AP)', 'الموردون والدائنون التجاريون', 'LIABILITY', 'CREDIT', true, true),
        (p_legal_entity_id, v_chart_id, v_cl_id, '212000', 'Accrued Expenses and Other Payables', 'المصروفات المستحقة والمستحقات الأخرى', 'LIABILITY', 'CREDIT', true, false),
        (p_legal_entity_id, v_chart_id, v_cl_id, '213000', 'Output VAT (15%) - Payable', 'ضريبة القيمة المضافة المخرجات (المبيعات)', 'LIABILITY', 'CREDIT', true, false),
        (p_legal_entity_id, v_chart_id, v_cl_id, '214000', 'Zakat Provision Payable', 'مخصص الزكاة الشرعية المستحقة', 'LIABILITY', 'CREDIT', true, false),
        (p_legal_entity_id, v_chart_id, v_cl_id, '215000', 'GOSI Contributions Payable', 'مستحقات التأمينات الاجتماعية', 'LIABILITY', 'CREDIT', true, false)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 3. EQUITY
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, '300000', 'Equity', 'حقوق الملكية', 'EQUITY', 'CREDIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_eq_id;

    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting) VALUES
        (p_legal_entity_id, v_chart_id, v_eq_id, '310000', 'Share Capital', 'رأس المال المدفوع', 'EQUITY', 'CREDIT', true),
        (p_legal_entity_id, v_chart_id, v_eq_id, '320000', 'Statutory Reserve', 'الاحتياطي النظامي', 'EQUITY', 'CREDIT', true),
        (p_legal_entity_id, v_chart_id, v_eq_id, '330000', 'Retained Earnings', 'الأرباح (الخسائر) المبقاة', 'EQUITY', 'CREDIT', true)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 4. REVENUE
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, '400000', 'Revenue', 'الإيرادات', 'REVENUE', 'CREDIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_rev_id;

    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting) VALUES
        (p_legal_entity_id, v_chart_id, v_rev_id, '410000', 'Sales Revenue', 'إيرادات مبيعات البضائع والخدمات', 'REVENUE', 'CREDIT', true),
        (p_legal_entity_id, v_chart_id, v_rev_id, '420000', 'Consulting & Service Revenue', 'إيرادات الخدمات والاستشارات', 'REVENUE', 'CREDIT', true),
        (p_legal_entity_id, v_chart_id, v_rev_id, '490000', 'Other Income', 'إيرادات وتوزيعات أخرى', 'REVENUE', 'CREDIT', true)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 5. EXPENSES
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, '500000', 'Expenses', 'المصروفات', 'EXPENSE', 'DEBIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_exp_id;

    -- 510000 Cost of Goods Sold
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, v_exp_id, '510000', 'Cost of Goods Sold (COGS)', 'تكلفة البضاعة المباعة', 'EXPENSE', 'DEBIT', true)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 520000 Operating Expenses (OPEX)
    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting)
    VALUES (p_legal_entity_id, v_chart_id, v_exp_id, '520000', 'Operating Expenses', 'المصروفات التشغيلية والعمومية', 'EXPENSE', 'DEBIT', false)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar RETURNING id INTO v_opex_id;

    INSERT INTO accounting.accounts (legal_entity_id, chart_id, parent_id, code, name_en, name_ar, account_type, normal_balance, allow_posting) VALUES
        (p_legal_entity_id, v_chart_id, v_opex_id, '521000', 'Salaries and Employee Wages', 'الرواتب والأجور الأساسية', 'EXPENSE', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_opex_id, '522000', 'Employer GOSI Contribution', 'حصة صاحب العمل في التأمينات الاجتماعية', 'EXPENSE', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_opex_id, '523000', 'Office Rent Expense', 'مصروف إيجار المكاتب والفروع', 'EXPENSE', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_opex_id, '524000', 'Utilities & Telecommunications', 'المنافع والكهرباء والاتصالات', 'EXPENSE', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_opex_id, '525000', 'Depreciation Expense', 'مصروف استهلاك الأصول والمعدات', 'EXPENSE', 'DEBIT', true),
        (p_legal_entity_id, v_chart_id, v_opex_id, '526000', 'Bank & Electronic Payment Fees', 'رسوم الحوالات البنكية ونقاط البيع (مدى)', 'EXPENSE', 'DEBIT', true)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- C. Fiscal Year 2026 and 12 Monthly Periods
    INSERT INTO accounting.fiscal_years (
        legal_entity_id, code, name_en, name_ar, start_date, end_date, status
    ) VALUES (
        p_legal_entity_id, 'FY-2026', 'Fiscal Year 2026', 'السنة المالية 2026م', '2026-01-01', '2026-12-31', 'OPEN'
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET status = 'OPEN'
    RETURNING id INTO v_fy_id;

    FOR m IN 1..12 LOOP
        v_start := ('2026-' || LPAD(m::TEXT, 2, '0') || '-01')::DATE;
        v_end := (v_start + INTERVAL '1 month' - INTERVAL '1 day')::DATE;

        INSERT INTO accounting.fiscal_periods (
            fiscal_year_id, legal_entity_id, period_number, period_name_en, period_name_ar, quarter, start_date, end_date, status
        ) VALUES (
            v_fy_id, p_legal_entity_id, m, v_month_names_en[m] || ' 2026', v_month_names_ar[m] || ' 2026',
            ((m - 1) / 3 + 1)::SMALLINT, v_start, v_end, 'OPEN'
        ) ON CONFLICT (fiscal_year_id, period_number) DO NOTHING;
    END LOOP;

    -- D. Standard Journals
    INSERT INTO accounting.journals (legal_entity_id, code, name_en, name_ar, journal_type) VALUES
        (p_legal_entity_id, 'GEN', 'General Journal', 'دفتر اليومية العامة', 'GENERAL'),
        (p_legal_entity_id, 'SALES', 'Sales Journal', 'دفتر يومية المبيعات', 'SALES'),
        (p_legal_entity_id, 'PURCH', 'Purchases Journal', 'دفتر يومية المشتريات', 'PURCHASE'),
        (p_legal_entity_id, 'CASH', 'Cash Receipts & Payments', 'دفتر المقبوضات والمدفوعات النقدية', 'CASH'),
        (p_legal_entity_id, 'BANK', 'Bank Transactions', 'دفتر الحركات البنكية', 'BANK'),
        (p_legal_entity_id, 'PAYROLL', 'Payroll Journal', 'دفتر مخصصات ومسير الرواتب', 'PAYROLL')
    ON CONFLICT (legal_entity_id, code) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    le RECORD;
BEGIN
    FOR le IN SELECT id FROM organization.legal_entities WHERE is_active = true LOOP
        PERFORM accounting.seed_default_coa(le.id);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_prevent_posted_journal_lines_mutation ON accounting.journal_lines;
DROP TRIGGER IF EXISTS trg_prevent_posted_journal_entry_mutation ON accounting.journal_entries;
DROP FUNCTION IF EXISTS accounting.prevent_posted_journal_mutation();

DROP TABLE IF EXISTS accounting.account_balances;
DROP TABLE IF EXISTS accounting.journal_lines;
DROP TABLE IF EXISTS accounting.journal_entries;
DROP TABLE IF EXISTS accounting.journals;
DROP TABLE IF EXISTS accounting.fiscal_periods;
DROP TABLE IF EXISTS accounting.fiscal_years;
DROP TABLE IF EXISTS accounting.accounts;
DROP TABLE IF EXISTS accounting.chart_of_accounts;

DELETE FROM authz.role_permissions WHERE permission_id IN (
    SELECT id FROM authz.permissions WHERE code LIKE 'accounting.%'
);
DELETE FROM authz.permissions WHERE code LIKE 'accounting.%';

DROP SCHEMA IF EXISTS accounting CASCADE;
-- +goose StatementEnd
