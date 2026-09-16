-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS receivables;

-- 1. Payment Terms
CREATE TABLE IF NOT EXISTS receivables.payment_terms (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    due_days            INT NOT NULL DEFAULT 0 CHECK (due_days >= 0),
    discount_days       INT NOT NULL DEFAULT 0 CHECK (discount_days >= 0),
    discount_percent    NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_payment_terms_code
        UNIQUE (legal_entity_id, code)
);

-- 2. Customers Master
CREATE TABLE IF NOT EXISTS receivables.customers (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    customer_number     TEXT NOT NULL,
    customer_type       TEXT NOT NULL CHECK (customer_type IN ('B2B', 'B2C', 'GOVERNMENT')),
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    tax_id              TEXT CHECK (tax_id IS NULL OR tax_id ~ '^[0-9]{15}$'),
    cr_number           TEXT CHECK (cr_number IS NULL OR cr_number ~ '^[0-9]{10}$'),
    national_id         TEXT CHECK (national_id IS NULL OR national_id ~ '^[0-9]{10}$'),
    
    -- Saudi National Address (ZATCA compliant)
    building_no         TEXT CHECK (building_no IS NULL OR building_no ~ '^[0-9]{4}$'),
    street_name         TEXT,
    district            TEXT,
    city                TEXT NOT NULL,
    postal_code         TEXT CHECK (postal_code IS NULL OR postal_code ~ '^[0-9]{5}$'),
    additional_no       TEXT CHECK (additional_no IS NULL OR additional_no ~ '^[0-9]{4}$'),
    country_code        TEXT NOT NULL DEFAULT 'SA',

    -- Contact details
    email               TEXT,
    phone               TEXT,
    contact_person      TEXT,

    -- Financial & Accounting Setup
    payment_terms_id    UUID REFERENCES receivables.payment_terms(id),
    currency_code       TEXT NOT NULL DEFAULT 'SAR',
    default_account_id  UUID REFERENCES accounting.accounts(id),
    status              TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'BLOCKED')),
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_customer_number
        UNIQUE (legal_entity_id, customer_number)
);

CREATE INDEX idx_customers_legal_entity ON receivables.customers(legal_entity_id);
CREATE INDEX idx_customers_tax_id ON receivables.customers(tax_id) WHERE tax_id IS NOT NULL;
CREATE INDEX idx_customers_cr_number ON receivables.customers(cr_number) WHERE cr_number IS NOT NULL;

-- 3. Customer GL Accounts (Subledger mapping per currency)
CREATE TABLE IF NOT EXISTS receivables.customer_accounts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    customer_id         UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE CASCADE,
    gl_account_id       UUID NOT NULL REFERENCES accounting.accounts(id),
    currency_code       TEXT NOT NULL DEFAULT 'SAR',
    is_primary          BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_customer_currency_account
        UNIQUE (customer_id, currency_code)
);

-- 4. Customer Credit Limits & Policy
CREATE TABLE IF NOT EXISTS receivables.customer_credit_limits (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    customer_id         UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE CASCADE,
    credit_limit        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (credit_limit >= 0),
    current_balance     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    max_overdue_days    INT NOT NULL DEFAULT 0 CHECK (max_overdue_days >= 0),
    block_action        TEXT NOT NULL DEFAULT 'WARN' CHECK (block_action IN ('WARN', 'BLOCK', 'REQUIRE_APPROVAL')),
    is_credit_hold      BOOLEAN NOT NULL DEFAULT false,
    hold_reason         TEXT,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_customer_credit_limit
        UNIQUE (customer_id)
);

-- 5. Accounts Receivable Open Items Ledger
CREATE TABLE IF NOT EXISTS receivables.open_items (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    customer_id         UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE CASCADE,
    document_type       TEXT NOT NULL CHECK (document_type IN ('INVOICE', 'DEBIT_NOTE', 'CREDIT_NOTE', 'UNALLOCATED_RECEIPT')),
    document_number     TEXT NOT NULL,
    document_id         UUID,
    document_date       DATE NOT NULL,
    due_date            DATE NOT NULL,
    currency_code       TEXT NOT NULL DEFAULT 'SAR',
    original_amount     NUMERIC(18, 4) NOT NULL CHECK (original_amount > 0),
    allocated_amount    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (allocated_amount >= 0),
    remaining_amount    NUMERIC(18, 4) NOT NULL CHECK (remaining_amount >= 0),
    status              TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'PARTIALLY_PAID', 'PAID', 'CANCELLED')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_customer_open_item_doc
        UNIQUE (legal_entity_id, document_type, document_number)
);

CREATE INDEX idx_open_items_customer_status ON receivables.open_items(customer_id, status);
CREATE INDEX idx_open_items_due_date ON receivables.open_items(due_date);

-- 6. Customer Receipts (سندات القبض)
CREATE TABLE IF NOT EXISTS receivables.receipts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    receipt_number      TEXT NOT NULL,
    customer_id         UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE RESTRICT,
    receipt_date        DATE NOT NULL,
    payment_method      TEXT NOT NULL CHECK (payment_method IN ('BANK_TRANSFER', 'CASH', 'CHEQUE', 'CREDIT_CARD', 'MADA', 'SADAD')),
    deposit_account_id  UUID NOT NULL REFERENCES accounting.accounts(id),
    amount              NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    currency_code       TEXT NOT NULL DEFAULT 'SAR',
    reference_number    TEXT,
    status              TEXT NOT NULL DEFAULT 'POSTED' CHECK (status IN ('DRAFT', 'POSTED', 'CANCELLED')),
    notes               TEXT,
    allocated_amount    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (allocated_amount >= 0),
    unallocated_amount  NUMERIC(18, 4) NOT NULL CHECK (unallocated_amount >= 0),
    journal_entry_id    UUID REFERENCES accounting.journal_entries(id),
    created_by_user_id  UUID REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_receipt_number
        UNIQUE (legal_entity_id, receipt_number)
);

CREATE INDEX idx_receipts_customer ON receivables.receipts(customer_id);
CREATE INDEX idx_receipts_date ON receivables.receipts(receipt_date);

-- 7. Receipt Allocations (Matching Payments to Open Invoices)
CREATE TABLE IF NOT EXISTS receivables.receipt_allocations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    receipt_id          UUID NOT NULL REFERENCES receivables.receipts(id) ON DELETE CASCADE,
    open_item_id        UUID NOT NULL REFERENCES receivables.open_items(id) ON DELETE RESTRICT,
    allocated_amount    NUMERIC(18, 4) NOT NULL CHECK (allocated_amount > 0),
    allocated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by_user_id  UUID REFERENCES iam.users(id)
);

CREATE INDEX idx_allocations_receipt ON receivables.receipt_allocations(receipt_id);
CREATE INDEX idx_allocations_open_item ON receivables.receipt_allocations(open_item_id);

-- 8. Seed Authorization Permissions
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('receivables.customers.read', 'عرض العملاء', 'View Customers', 'RECEIVABLES'),
    ('receivables.customers.create', 'إنشاء عميل', 'Create Customer', 'RECEIVABLES'),
    ('receivables.customers.update', 'تعديل عميل', 'Update Customer', 'RECEIVABLES'),
    ('receivables.customers.delete', 'تعطيل عميل', 'Disable Customer', 'RECEIVABLES'),
    ('receivables.receipts.read', 'عرض سندات القبض', 'View Receipts', 'RECEIVABLES'),
    ('receivables.receipts.create', 'إنشاء سند قبض', 'Create Receipt', 'RECEIVABLES'),
    ('receivables.receipts.post', 'ترحيل سند قبض', 'Post Receipt', 'RECEIVABLES'),
    ('receivables.receipts.cancel', 'إلغاء سند قبض', 'Cancel Receipt', 'RECEIVABLES'),
    ('receivables.credit_limits.manage', 'إدارة حدود الائتمان', 'Manage Credit Limits', 'RECEIVABLES'),
    ('receivables.reports.read', 'تقارير الذمم المدينة', 'View AR Reports', 'RECEIVABLES')
ON CONFLICT (code) DO NOTHING;

-- Assign permissions to SUPER_ADMIN
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN' AND p.module = 'RECEIVABLES'
ON CONFLICT DO NOTHING;

-- 9. Seed Helper Function
CREATE OR REPLACE FUNCTION receivables.seed_default_receivables_data(p_legal_entity_id UUID)
RETURNS void AS $$
DECLARE
    v_term_imm_id       UUID;
    v_term_net30_id     UUID;
    v_term_net60_id     UUID;
    v_ar_account_id     UUID;
    v_bank_account_id   UUID;
    v_cust_1_id         UUID;
    v_cust_2_id         UUID;
    v_cust_3_id         UUID;
    v_open_item_1_id    UUID;
    v_open_item_2_id    UUID;
    v_receipt_1_id      UUID;
BEGIN
    -- 1. Standard Payment Terms
    INSERT INTO receivables.payment_terms (legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent)
    VALUES (p_legal_entity_id, 'IMMEDIATE', 'Cash / Immediate Payment', 'الدفع الفوري نقداً', 0, 0, 0.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_term_imm_id;

    INSERT INTO receivables.payment_terms (legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent)
    VALUES (p_legal_entity_id, 'NET15', 'Net 15 Days', 'أجل 15 يوماً', 15, 0, 0.00)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    INSERT INTO receivables.payment_terms (legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent)
    VALUES (p_legal_entity_id, 'NET30', 'Net 30 Days (2% 10 Days)', 'أجل 30 يوماً (خصم 2% خلال 10 أيام)', 30, 10, 2.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_term_net30_id;

    INSERT INTO receivables.payment_terms (legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent)
    VALUES (p_legal_entity_id, 'NET60', 'Net 60 Days', 'أجل 60 يوماً', 60, 0, 0.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_term_net60_id;

    -- Lookup AR Control Account (112000) and Bank Account (111002)
    SELECT id INTO v_ar_account_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '112000' LIMIT 1;
    SELECT id INTO v_bank_account_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '111002' LIMIT 1;

    -- 2. Seed Sample Customers
    -- Customer 1: B2B Saudi Enterprise
    INSERT INTO receivables.customers (
        legal_entity_id, customer_number, customer_type, name_en, name_ar,
        tax_id, cr_number, building_no, street_name, district, city, postal_code, additional_no,
        email, phone, contact_person, payment_terms_id, default_account_id, status
    ) VALUES (
        p_legal_entity_id, 'CUST-2026-0001', 'B2B',
        'Al-Arabiya Modern Contracting Co.', 'شركة العربية الحديثة للمقاولات العامة',
        '310123456700003', '1010897654', '2415', 'King Fahd Road', 'Al-Olaya', 'Riyadh', '12211', '8921',
        'accounts@alarabiya-contracting.sa', '+966114567890', 'م. خالد الدوسري',
        v_term_net30_id, v_ar_account_id, 'ACTIVE'
    )
    ON CONFLICT (legal_entity_id, customer_number) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_cust_1_id;

    -- Customer 2: B2B Retail Distributor
    INSERT INTO receivables.customers (
        legal_entity_id, customer_number, customer_type, name_en, name_ar,
        tax_id, cr_number, building_no, street_name, district, city, postal_code, additional_no,
        email, phone, contact_person, payment_terms_id, default_account_id, status
    ) VALUES (
        p_legal_entity_id, 'CUST-2026-0002', 'B2B',
        'Al-Madina Commercial Trading Est.', 'مؤسسة المدينة للتجارة والتوريدات',
        '310987654300003', '2050123456', '7812', 'Prince Nayef Street', 'Al-Tubishi', 'Dammam', '32415', '4312',
        'info@almadina-trading.sa', '+966138901234', 'أ. طارق الشمري',
        v_term_net60_id, v_ar_account_id, 'ACTIVE'
    )
    ON CONFLICT (legal_entity_id, customer_number) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_cust_2_id;

    -- Customer 3: B2C Individual VIP Client
    INSERT INTO receivables.customers (
        legal_entity_id, customer_number, customer_type, name_en, name_ar,
        national_id, building_no, street_name, district, city, postal_code, additional_no,
        email, phone, contact_person, payment_terms_id, default_account_id, status
    ) VALUES (
        p_legal_entity_id, 'CUST-2026-0003', 'B2C',
        'Abdulaziz Mohammed Al-Otaibi', 'عبدالعزيز محمد العتيبي',
        '1087654321', '3310', 'Al-Uruba Street', 'Al-Rahmaniyah', 'Riyadh', '12341', '6543',
        'a.otaibi@gmail.com', '+966501234567', 'عبدالعزيز العتيبي',
        v_term_imm_id, v_ar_account_id, 'ACTIVE'
    )
    ON CONFLICT (legal_entity_id, customer_number) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_cust_3_id;

    -- 3. Customer Accounts & Credit Limits
    IF v_cust_1_id IS NOT NULL AND v_ar_account_id IS NOT NULL THEN
        INSERT INTO receivables.customer_accounts (customer_id, gl_account_id, currency_code, is_primary)
        VALUES (v_cust_1_id, v_ar_account_id, 'SAR', true)
        ON CONFLICT (customer_id, currency_code) DO NOTHING;

        INSERT INTO receivables.customer_credit_limits (customer_id, credit_limit, current_balance, max_overdue_days, block_action)
        VALUES (v_cust_1_id, 250000.0000, 34500.0000, 45, 'BLOCK')
        ON CONFLICT (customer_id) DO NOTHING;

        -- Open item for Customer 1
        INSERT INTO receivables.open_items (
            legal_entity_id, customer_id, document_type, document_number,
            document_date, due_date, currency_code, original_amount, allocated_amount, remaining_amount, status
        ) VALUES (
            p_legal_entity_id, v_cust_1_id, 'INVOICE', 'INV-2026-0045',
            '2026-08-15', '2026-09-14', 'SAR', 57500.0000, 23000.0000, 34500.0000, 'PARTIALLY_PAID'
        )
        ON CONFLICT (legal_entity_id, document_type, document_number) DO UPDATE SET remaining_amount = EXCLUDED.remaining_amount
        RETURNING id INTO v_open_item_1_id;
    END IF;

    IF v_cust_2_id IS NOT NULL AND v_ar_account_id IS NOT NULL THEN
        INSERT INTO receivables.customer_accounts (customer_id, gl_account_id, currency_code, is_primary)
        VALUES (v_cust_2_id, v_ar_account_id, 'SAR', true)
        ON CONFLICT (customer_id, currency_code) DO NOTHING;

        INSERT INTO receivables.customer_credit_limits (customer_id, credit_limit, current_balance, max_overdue_days, block_action)
        VALUES (v_cust_2_id, 100000.0000, 18400.0000, 60, 'WARN')
        ON CONFLICT (customer_id) DO NOTHING;

        -- Open item for Customer 2
        INSERT INTO receivables.open_items (
            legal_entity_id, customer_id, document_type, document_number,
            document_date, due_date, currency_code, original_amount, allocated_amount, remaining_amount, status
        ) VALUES (
            p_legal_entity_id, v_cust_2_id, 'INVOICE', 'INV-2026-0062',
            '2026-09-01', '2026-10-31', 'SAR', 18400.0000, 0.0000, 18400.0000, 'OPEN'
        )
        ON CONFLICT (legal_entity_id, document_type, document_number) DO NOTHING
        RETURNING id INTO v_open_item_2_id;
    END IF;

    -- 4. Sample Receipt Voucher for Customer 1
    IF v_cust_1_id IS NOT NULL AND v_bank_account_id IS NOT NULL AND v_open_item_1_id IS NOT NULL THEN
        INSERT INTO receivables.receipts (
            legal_entity_id, receipt_number, customer_id, receipt_date, payment_method,
            deposit_account_id, amount, currency_code, reference_number, status,
            notes, allocated_amount, unallocated_amount
        ) VALUES (
            p_legal_entity_id, 'REC-2026-0001', v_cust_1_id, '2026-09-10', 'BANK_TRANSFER',
            v_bank_account_id, 23000.0000, 'SAR', 'TRF-ALRAJHI-99882', 'POSTED',
            'دفعة مسددة لحساب الفاتورة INV-2026-0045 عبر تحويل مصرفي الراجحي', 23000.0000, 0.0000
        )
        ON CONFLICT (legal_entity_id, receipt_number) DO UPDATE SET notes = EXCLUDED.notes
        RETURNING id INTO v_receipt_1_id;

        IF v_receipt_1_id IS NOT NULL THEN
            INSERT INTO receivables.receipt_allocations (receipt_id, open_item_id, allocated_amount)
            VALUES (v_receipt_1_id, v_open_item_1_id, 23000.0000);
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Automatically seed default receivables data for existing legal entities
DO $$
DECLARE
    le RECORD;
BEGIN
    FOR le IN SELECT id FROM organization.legal_entities WHERE is_active = true LOOP
        PERFORM receivables.seed_default_receivables_data(le.id);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS receivables.seed_default_receivables_data(UUID);
DROP TABLE IF EXISTS receivables.receipt_allocations;
DROP TABLE IF EXISTS receivables.receipts;
DROP TABLE IF EXISTS receivables.open_items;
DROP TABLE IF EXISTS receivables.customer_credit_limits;
DROP TABLE IF EXISTS receivables.customer_accounts;
DROP TABLE IF EXISTS receivables.customers;
DROP TABLE IF EXISTS receivables.payment_terms;

DELETE FROM authz.role_permissions WHERE permission_id IN (
    SELECT id FROM authz.permissions WHERE module = 'RECEIVABLES'
);
DELETE FROM authz.permissions WHERE module = 'RECEIVABLES';

DROP SCHEMA IF EXISTS receivables CASCADE;
-- +goose StatementEnd
