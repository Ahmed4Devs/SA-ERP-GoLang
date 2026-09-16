-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS payables;

-- 1. Payment Terms
CREATE TABLE IF NOT EXISTS payables.payment_terms (
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

    CONSTRAINT uq_payables_payment_terms_code
        UNIQUE (legal_entity_id, code)
);

-- 2. Suppliers Master Data
CREATE TABLE IF NOT EXISTS payables.suppliers (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    supplier_number             TEXT NOT NULL,
    supplier_type               TEXT NOT NULL CHECK (supplier_type IN ('LOCAL', 'INTERNATIONAL', 'GOVERNMENT')),
    name_en                     TEXT NOT NULL,
    name_ar                     TEXT NOT NULL,
    tax_id                      TEXT CHECK (tax_id IS NULL OR tax_id ~ '^[0-9]{15}$'),
    cr_number                   TEXT CHECK (cr_number IS NULL OR cr_number ~ '^[0-9]{10}$'),
    national_id                 TEXT CHECK (national_id IS NULL OR national_id ~ '^[0-9]{10}$'),
    iban                        TEXT CHECK (iban IS NULL OR iban ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{4,30}$'),
    bank_name                   TEXT,
    swift_code                  TEXT,
    
    -- Saudi National Address (ZATCA & GACA compliant)
    building_no                 TEXT CHECK (building_no IS NULL OR building_no ~ '^[0-9]{4}$'),
    street_name                 TEXT,
    district                    TEXT,
    city                        TEXT NOT NULL,
    postal_code                 TEXT CHECK (postal_code IS NULL OR postal_code ~ '^[0-9]{5}$'),
    additional_no               TEXT CHECK (additional_no IS NULL OR additional_no ~ '^[0-9]{4}$'),
    country_code                TEXT NOT NULL DEFAULT 'SA',

    -- Contact details
    email                       TEXT,
    phone                       TEXT,
    contact_person              TEXT,
    website                     TEXT,

    -- Financial & Accounting Setup
    payment_terms_id            UUID REFERENCES payables.payment_terms(id),
    currency_code               TEXT NOT NULL DEFAULT 'SAR',
    default_expense_account_id  UUID REFERENCES accounting.accounts(id),
    default_liability_account_id UUID REFERENCES accounting.accounts(id),
    status                      TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'BLOCKED')),
    
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_number
        UNIQUE (legal_entity_id, supplier_number)
);

CREATE INDEX idx_suppliers_legal_entity ON payables.suppliers(legal_entity_id);
CREATE INDEX idx_suppliers_tax_id ON payables.suppliers(tax_id) WHERE tax_id IS NOT NULL;
CREATE INDEX idx_suppliers_cr_number ON payables.suppliers(cr_number) WHERE cr_number IS NOT NULL;

-- 3. Supplier Accounts (Subledger mapping per currency)
CREATE TABLE IF NOT EXISTS payables.supplier_accounts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    supplier_id         UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE CASCADE,
    gl_account_id       UUID NOT NULL REFERENCES accounting.accounts(id),
    currency_code       TEXT NOT NULL DEFAULT 'SAR',
    is_primary          BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_currency_account
        UNIQUE (supplier_id, currency_code)
);

-- 4. Supplier Invoices (AP Invoices / Bills)
CREATE TABLE IF NOT EXISTS payables.supplier_invoices (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    invoice_number              TEXT NOT NULL,
    supplier_invoice_number     TEXT NOT NULL,
    supplier_id                 UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE RESTRICT,
    invoice_date                DATE NOT NULL,
    due_date                    DATE NOT NULL,
    currency_code               TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate               NUMERIC(18, 6) NOT NULL DEFAULT 1.000000 CHECK (exchange_rate > 0),
    
    -- Financial breakdown
    subtotal                    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (subtotal >= 0),
    discount_amount             NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (discount_amount >= 0),
    taxable_amount              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (taxable_amount >= 0),
    tax_amount                  NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    total_amount                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_amount >= 0),
    
    status                      TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'POSTED', 'PAID', 'CANCELLED')),
    payment_status              TEXT NOT NULL DEFAULT 'UNPAID' CHECK (payment_status IN ('UNPAID', 'PARTIALLY_PAID', 'PAID')),
    journal_entry_id            UUID REFERENCES accounting.journal_entries(id),
    notes                       TEXT,
    created_by_user_id          UUID REFERENCES iam.users(id),
    
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_invoice_internal_number
        UNIQUE (legal_entity_id, invoice_number),
    CONSTRAINT uq_supplier_invoice_external_number
        UNIQUE (legal_entity_id, supplier_id, supplier_invoice_number)
);

CREATE INDEX idx_supplier_invoices_legal_entity ON payables.supplier_invoices(legal_entity_id);
CREATE INDEX idx_supplier_invoices_supplier ON payables.supplier_invoices(supplier_id);
CREATE INDEX idx_supplier_invoices_date ON payables.supplier_invoices(invoice_date);
CREATE INDEX idx_supplier_invoices_status ON payables.supplier_invoices(status, payment_status);

-- 5. Supplier Invoice Lines
CREATE TABLE IF NOT EXISTS payables.supplier_invoice_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    invoice_id          UUID NOT NULL REFERENCES payables.supplier_invoices(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    item_id             UUID REFERENCES products.items(id),
    expense_account_id  UUID NOT NULL REFERENCES accounting.accounts(id),
    description         TEXT NOT NULL,
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    discount_percent    NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    discount_amount     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (discount_amount >= 0),
    tax_rate            NUMERIC(5, 2) NOT NULL DEFAULT 15.00 CHECK (tax_rate >= 0),
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    subtotal            NUMERIC(18, 4) NOT NULL CHECK (subtotal >= 0),
    total_amount        NUMERIC(18, 4) NOT NULL CHECK (total_amount >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_invoice_line_number
        UNIQUE (invoice_id, line_number)
);

CREATE INDEX idx_supplier_invoice_lines_invoice ON payables.supplier_invoice_lines(invoice_id);

-- 6. Supplier Credit Notes
CREATE TABLE IF NOT EXISTS payables.supplier_credit_notes (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    credit_note_number          TEXT NOT NULL,
    supplier_credit_note_number TEXT,
    supplier_id                 UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE RESTRICT,
    original_invoice_id         UUID REFERENCES payables.supplier_invoices(id),
    credit_note_date            DATE NOT NULL,
    currency_code               TEXT NOT NULL DEFAULT 'SAR',
    subtotal                    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (subtotal >= 0),
    tax_amount                  NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    total_amount                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_amount >= 0),
    reason                      TEXT,
    status                      TEXT NOT NULL DEFAULT 'POSTED' CHECK (status IN ('DRAFT', 'POSTED', 'CANCELLED')),
    journal_entry_id            UUID REFERENCES accounting.journal_entries(id),
    created_by_user_id          UUID REFERENCES iam.users(id),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_credit_note_number
        UNIQUE (legal_entity_id, credit_note_number)
);

CREATE TABLE IF NOT EXISTS payables.supplier_credit_note_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    credit_note_id      UUID NOT NULL REFERENCES payables.supplier_credit_notes(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    item_id             UUID REFERENCES products.items(id),
    expense_account_id  UUID NOT NULL REFERENCES accounting.accounts(id),
    description         TEXT NOT NULL,
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    tax_rate            NUMERIC(5, 2) NOT NULL DEFAULT 15.00 CHECK (tax_rate >= 0),
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    subtotal            NUMERIC(18, 4) NOT NULL CHECK (subtotal >= 0),
    total_amount        NUMERIC(18, 4) NOT NULL CHECK (total_amount >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_credit_note_line_number
        UNIQUE (credit_note_id, line_number)
);

-- 7. Accounts Payable Open Items Ledger
CREATE TABLE IF NOT EXISTS payables.open_items (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    supplier_id         UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE CASCADE,
    document_type       TEXT NOT NULL CHECK (document_type IN ('INVOICE', 'CREDIT_NOTE', 'UNALLOCATED_PAYMENT')),
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

    CONSTRAINT uq_supplier_open_item_doc
        UNIQUE (legal_entity_id, document_type, document_number)
);

CREATE INDEX idx_payables_open_items_supplier_status ON payables.open_items(supplier_id, status);
CREATE INDEX idx_payables_open_items_due_date ON payables.open_items(due_date);

-- 8. Supplier Payments (سندات الصرف)
CREATE TABLE IF NOT EXISTS payables.payments (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    payment_number              TEXT NOT NULL,
    supplier_id                 UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE RESTRICT,
    payment_date                DATE NOT NULL,
    payment_method              TEXT NOT NULL CHECK (payment_method IN ('BANK_TRANSFER', 'CASH', 'CHEQUE', 'MADA', 'SADAD')),
    disbursement_account_id     UUID NOT NULL REFERENCES accounting.accounts(id),
    amount                      NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    currency_code               TEXT NOT NULL DEFAULT 'SAR',
    reference_number            TEXT,
    status                      TEXT NOT NULL DEFAULT 'POSTED' CHECK (status IN ('DRAFT', 'POSTED', 'CANCELLED')),
    notes                       TEXT,
    allocated_amount            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (allocated_amount >= 0),
    unallocated_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (unallocated_amount >= 0),
    journal_entry_id            UUID REFERENCES accounting.journal_entries(id),
    created_by_user_id          UUID REFERENCES iam.users(id),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_payables_payment_number
        UNIQUE (legal_entity_id, payment_number)
);

CREATE INDEX idx_payables_payments_supplier ON payables.payments(supplier_id);
CREATE INDEX idx_payables_payments_date ON payables.payments(payment_date);

-- 9. Payment Allocations (Matching Payments to Open Supplier Invoices)
CREATE TABLE IF NOT EXISTS payables.payment_allocations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    payment_id          UUID NOT NULL REFERENCES payables.payments(id) ON DELETE CASCADE,
    open_item_id        UUID NOT NULL REFERENCES payables.open_items(id) ON DELETE RESTRICT,
    allocated_amount    NUMERIC(18, 4) NOT NULL CHECK (allocated_amount > 0),
    allocated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payables_allocations_payment ON payables.payment_allocations(payment_id);
CREATE INDEX idx_payables_allocations_open_item ON payables.payment_allocations(open_item_id);

-- 10. Payables Permissions in Authz
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'payables.supplier.view', 'View Suppliers', 'عرض سجل الموردين وبياناتهم المالية', 'PAYABLES', now()),
    (uuidv7(), 'payables.supplier.create', 'Create Supplier', 'إنشاء مورد جديد والتحقق من الرقم الضريبي والعنوان الوطني', 'PAYABLES', now()),
    (uuidv7(), 'payables.supplier.edit', 'Edit Supplier', 'تعديل بيانات المورد والشروط الائتمانية والبنكية', 'PAYABLES', now()),
    (uuidv7(), 'payables.supplier.delete', 'Delete Supplier', 'حذف أو تعطيل المورد', 'PAYABLES', now()),
    (uuidv7(), 'payables.invoice.view', 'View Invoices', 'عرض فواتير الموردين ومطابقتها', 'PAYABLES', now()),
    (uuidv7(), 'payables.invoice.create', 'Create Invoice', 'إنشاء وإدخال فاتورة مورد جديدة', 'PAYABLES', now()),
    (uuidv7(), 'payables.invoice.approve', 'Approve Invoices', 'اعتماد فواتير الموردين', 'PAYABLES', now()),
    (uuidv7(), 'payables.invoice.post', 'Post Invoices', 'ترحيل فواتير الموردين وتوليد قيود الأستاذ العام', 'PAYABLES', now()),
    (uuidv7(), 'payables.payment.view', 'View Payments', 'عرض سندات الصرف للموردين وتوزيعاتها', 'PAYABLES', now()),
    (uuidv7(), 'payables.payment.create', 'Create Payment', 'إنشاء سند صرف للمورد', 'PAYABLES', now()),
    (uuidv7(), 'payables.payment.post', 'Post Payment', 'ترحيل سند الصرف وتوزيعه على الفواتير المفتوحة', 'PAYABLES', now()),
    (uuidv7(), 'payables.aging.view', 'View AP Aging', 'عرض تقرير ومحرك أعمار ديون الموردين', 'PAYABLES', now())
ON CONFLICT (code) DO NOTHING;

-- 11. Seed Default Data Function
CREATE OR REPLACE FUNCTION payables.seed_default_payables_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_net30_id UUID;
    v_net15_id UUID;
    v_immediate_id UUID;
    v_ap_account_id UUID;
    v_exp_account_id UUID;
BEGIN
    -- 1. Payment Terms
    INSERT INTO payables.payment_terms (legal_entity_id, code, name_en, name_ar, due_days, discount_days, discount_percent)
    VALUES 
        (p_legal_entity_id, 'NET30', 'Net 30 Days', 'صافي 30 يوماً', 30, 0, 0.00),
        (p_legal_entity_id, 'NET15', 'Net 15 Days', 'صافي 15 يوماً', 15, 0, 0.00),
        (p_legal_entity_id, 'IMMEDIATE', 'Cash on Delivery', 'سداد فوري عند الاستلام', 0, 0, 0.00)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    SELECT id INTO v_net30_id FROM payables.payment_terms WHERE legal_entity_id = p_legal_entity_id AND code = 'NET30';
    SELECT id INTO v_immediate_id FROM payables.payment_terms WHERE legal_entity_id = p_legal_entity_id AND code = 'IMMEDIATE';

    -- Find AP GL account (211000) and Operating Expense (520000)
    SELECT id INTO v_ap_account_id FROM accounting.accounts 
    WHERE legal_entity_id = p_legal_entity_id AND code = '211000' LIMIT 1;

    SELECT id INTO v_exp_account_id FROM accounting.accounts 
    WHERE legal_entity_id = p_legal_entity_id AND code = '520000' LIMIT 1;

    -- 2. Default Saudi Suppliers
    IF v_ap_account_id IS NOT NULL THEN
        INSERT INTO payables.suppliers (
            legal_entity_id, supplier_number, supplier_type, name_en, name_ar,
            tax_id, cr_number, iban, bank_name,
            building_no, street_name, district, city, postal_code, additional_no, country_code,
            email, phone, contact_person, payment_terms_id, currency_code,
            default_liability_account_id, default_expense_account_id, status
        )
        VALUES 
            (
                p_legal_entity_id, 'SUP-0001', 'LOCAL',
                'Al-Fozan Building Materials Co.', 'شركة الفوزان لمواد البناء',
                '310123456789013', '1010345678', 'SA0380000000123456789012', 'Al Rajhi Bank',
                '7890', 'King Fahd Road', 'Al Olaya', 'Riyadh', '12214', '3456', 'SA',
                'procurement@alfozan.example.com', '+966114567890', 'Eng. Mohammed Al-Fozan',
                v_net30_id, 'SAR', v_ap_account_id, v_exp_account_id, 'ACTIVE'
            ),
            (
                p_legal_entity_id, 'SUP-0002', 'LOCAL',
                'Saudi Electricity Company (SEC)', 'الشركة السعودية للكهرباء',
                '300012345678903', '1010000012', 'SA4410000000987654321098', 'National Commercial Bank (SNB)',
                '1234', 'Olaya Main Street', 'Al Malaz', 'Riyadh', '11411', '5678', 'SA',
                'billing@se.com.sa', '+966118001234', 'Billing Services',
                v_immediate_id, 'SAR', v_ap_account_id, v_exp_account_id, 'ACTIVE'
            )
        ON CONFLICT (legal_entity_id, supplier_number) DO NOTHING;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS payables.seed_default_payables_data(UUID);
DROP TABLE IF EXISTS payables.payment_allocations CASCADE;
DROP TABLE IF EXISTS payables.payments CASCADE;
DROP TABLE IF EXISTS payables.open_items CASCADE;
DROP TABLE IF EXISTS payables.supplier_credit_note_lines CASCADE;
DROP TABLE IF EXISTS payables.supplier_credit_notes CASCADE;
DROP TABLE IF EXISTS payables.supplier_invoice_lines CASCADE;
DROP TABLE IF EXISTS payables.supplier_invoices CASCADE;
DROP TABLE IF EXISTS payables.supplier_accounts CASCADE;
DROP TABLE IF EXISTS payables.suppliers CASCADE;
DROP TABLE IF EXISTS payables.payment_terms CASCADE;
DROP SCHEMA IF EXISTS payables CASCADE;
-- +goose StatementEnd
