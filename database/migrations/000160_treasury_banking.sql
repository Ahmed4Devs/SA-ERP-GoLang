-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS treasury;
CREATE SCHEMA IF NOT EXISTS banking;

-- ============================================================================
-- 1. TREASURY: CASH ACCOUNTS (الصناديق والعهد المالية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS treasury.cash_accounts (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    branch_id               UUID REFERENCES organization.branches(id) ON DELETE SET NULL,
    account_code            TEXT NOT NULL,
    name_en                 TEXT NOT NULL,
    name_ar                 TEXT NOT NULL,
    account_type            TEXT NOT NULL CHECK (account_type IN ('PETTY_CASH', 'CASH_REGISTER', 'MAIN_VAULT')),
    gl_account_id           UUID REFERENCES accounting.accounts(id),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    custodian_user_id       UUID REFERENCES iam.users(id),
    max_limit               NUMERIC(18, 4) NOT NULL DEFAULT 10000.0000 CHECK (max_limit >= 0),
    current_balance         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    status                  TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'CLOSED')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_cash_account_code
        UNIQUE (legal_entity_id, account_code)
);

CREATE INDEX idx_cash_accounts_entity ON treasury.cash_accounts(legal_entity_id);
CREATE INDEX idx_cash_accounts_type ON treasury.cash_accounts(account_type);
CREATE INDEX idx_cash_accounts_gl ON treasury.cash_accounts(gl_account_id);

-- ============================================================================
-- 2. BANKING: BANK ACCOUNTS (الحسابات المصرفية الرسمية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS banking.bank_accounts (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    account_code            TEXT NOT NULL,
    bank_name_en            TEXT NOT NULL,
    bank_name_ar            TEXT NOT NULL,
    branch_name             TEXT,
    account_number          TEXT NOT NULL,
    account_number_masked   TEXT NOT NULL,
    iban                    TEXT NOT NULL,
    swift_bic               TEXT NOT NULL,
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    gl_account_id           UUID REFERENCES accounting.accounts(id),
    opening_balance         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    current_balance         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    available_balance       NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    overdraft_limit         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (overdraft_limit >= 0),
    status                  TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FROZEN', 'CLOSED')),
    is_primary              BOOLEAN NOT NULL DEFAULT false,
    contact_person          TEXT,
    contact_phone           TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bank_account_code
        UNIQUE (legal_entity_id, account_code),
    CONSTRAINT uq_bank_account_iban
        UNIQUE (legal_entity_id, iban)
);

CREATE INDEX idx_bank_accounts_entity ON banking.bank_accounts(legal_entity_id);
CREATE INDEX idx_bank_accounts_gl ON banking.bank_accounts(gl_account_id);
CREATE INDEX idx_bank_accounts_iban ON banking.bank_accounts(iban);

-- ============================================================================
-- 3. TREASURY: RECEIPTS (سندات القبض النقدية والشيكات)
-- ============================================================================
CREATE TABLE IF NOT EXISTS treasury.receipts (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    receipt_number          TEXT NOT NULL,
    cash_account_id         UUID NOT NULL REFERENCES treasury.cash_accounts(id) ON DELETE RESTRICT,
    receipt_date            DATE NOT NULL,
    receipt_type            TEXT NOT NULL CHECK (receipt_type IN ('CUSTOMER_PAYMENT', 'MISCELLANEOUS', 'CASH_SALE', 'DEPOSIT', 'INTER_BRANCH')),
    received_from           TEXT NOT NULL,
    amount                  NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    payment_method          TEXT NOT NULL DEFAULT 'CASH' CHECK (payment_method IN ('CASH', 'CHEQUE')),
    cheque_number           TEXT,
    cheque_date             DATE,
    reference_number        TEXT,
    description             TEXT,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'POSTED', 'VOIDED')),
    gl_offset_account_id    UUID REFERENCES accounting.accounts(id),
    journal_entry_id        UUID REFERENCES accounting.journal_entries(id),
    created_by_user_id      UUID REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_treasury_receipt_num
        UNIQUE (legal_entity_id, receipt_number)
);

CREATE INDEX idx_treasury_receipts_entity ON treasury.receipts(legal_entity_id);
CREATE INDEX idx_treasury_receipts_cash_acc ON treasury.receipts(cash_account_id);
CREATE INDEX idx_treasury_receipts_date ON treasury.receipts(receipt_date);
CREATE INDEX idx_treasury_receipts_status ON treasury.receipts(status);

-- ============================================================================
-- 4. TREASURY: PAYMENTS (سندات الصرف النقدية والمصروفات النثرية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS treasury.payments (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    payment_number          TEXT NOT NULL,
    cash_account_id         UUID NOT NULL REFERENCES treasury.cash_accounts(id) ON DELETE RESTRICT,
    payment_date            DATE NOT NULL,
    payment_type            TEXT NOT NULL CHECK (payment_type IN ('PETTY_CASH_EXPENSE', 'SUPPLIER_PAYMENT', 'CASH_ADVANCE', 'MISCELLANEOUS')),
    paid_to                 TEXT NOT NULL,
    amount                  NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    reference_number        TEXT,
    description             TEXT,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'PAID', 'VOIDED')),
    gl_offset_account_id    UUID REFERENCES accounting.accounts(id),
    journal_entry_id        UUID REFERENCES accounting.journal_entries(id),
    created_by_user_id      UUID REFERENCES iam.users(id),
    approved_by_user_id     UUID REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_treasury_payment_num
        UNIQUE (legal_entity_id, payment_number)
);

CREATE INDEX idx_treasury_payments_entity ON treasury.payments(legal_entity_id);
CREATE INDEX idx_treasury_payments_cash_acc ON treasury.payments(cash_account_id);
CREATE INDEX idx_treasury_payments_date ON treasury.payments(payment_date);
CREATE INDEX idx_treasury_payments_status ON treasury.payments(status);

-- ============================================================================
-- 5. TREASURY: TRANSFERS (التحويلات بين الخزائن والحسابات البنكية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS treasury.transfers (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    transfer_number             TEXT NOT NULL,
    transfer_date               DATE NOT NULL,
    source_type                 TEXT NOT NULL CHECK (source_type IN ('CASH', 'BANK')),
    source_cash_account_id      UUID REFERENCES treasury.cash_accounts(id) ON DELETE SET NULL,
    source_bank_account_id      UUID REFERENCES banking.bank_accounts(id) ON DELETE SET NULL,
    destination_type            TEXT NOT NULL CHECK (destination_type IN ('CASH', 'BANK')),
    destination_cash_account_id UUID REFERENCES treasury.cash_accounts(id) ON DELETE SET NULL,
    destination_bank_account_id UUID REFERENCES banking.bank_accounts(id) ON DELETE SET NULL,
    amount                      NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    currency_code               TEXT NOT NULL DEFAULT 'SAR',
    fee_amount                  NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (fee_amount >= 0),
    fee_gl_account_id           UUID REFERENCES accounting.accounts(id),
    status                      TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'IN_TRANSIT', 'COMPLETED', 'CANCELLED')),
    reference_number            TEXT,
    notes                       TEXT,
    journal_entry_id            UUID REFERENCES accounting.journal_entries(id),
    initiated_by_user_id        UUID REFERENCES iam.users(id),
    completed_by_user_id        UUID REFERENCES iam.users(id),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_treasury_transfer_num
        UNIQUE (legal_entity_id, transfer_number)
);

CREATE INDEX idx_treasury_transfers_entity ON treasury.transfers(legal_entity_id);
CREATE INDEX idx_treasury_transfers_status ON treasury.transfers(status);

-- ============================================================================
-- 6. BANKING: STATEMENTS (كشوف الحسابات البنكية المستوردة)
-- ============================================================================
CREATE TABLE IF NOT EXISTS banking.statements (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    bank_account_id         UUID NOT NULL REFERENCES banking.bank_accounts(id) ON DELETE CASCADE,
    statement_number        TEXT NOT NULL,
    statement_date          DATE NOT NULL,
    start_date              DATE NOT NULL,
    end_date                DATE NOT NULL,
    opening_balance         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    closing_balance         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_credits           NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_credits >= 0),
    total_debits            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_debits >= 0),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    file_format             TEXT NOT NULL DEFAULT 'CSV' CHECK (file_format IN ('MT940', 'CAMT053', 'CSV', 'MANUAL')),
    status                  TEXT NOT NULL DEFAULT 'IMPORTED' CHECK (status IN ('DRAFT', 'IMPORTED', 'PARTIALLY_RECONCILED', 'RECONCILED')),
    source_filename         TEXT,
    imported_by_user_id     UUID REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bank_statement_num
        UNIQUE (bank_account_id, statement_number)
);

CREATE INDEX idx_bank_statements_acc ON banking.statements(bank_account_id);
CREATE INDEX idx_bank_statements_date ON banking.statements(statement_date);

-- ============================================================================
-- 7. BANKING: STATEMENT LINES (بنود وحركات كشف الحساب البنكي)
-- ============================================================================
CREATE TABLE IF NOT EXISTS banking.statement_lines (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    statement_id            UUID NOT NULL REFERENCES banking.statements(id) ON DELETE CASCADE,
    line_number             INT NOT NULL,
    transaction_date        DATE NOT NULL,
    value_date              DATE,
    description             TEXT NOT NULL,
    reference_number        TEXT,
    transaction_type        TEXT NOT NULL CHECK (transaction_type IN ('CREDIT', 'DEBIT')),
    amount                  NUMERIC(18, 4) NOT NULL CHECK (amount > 0),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    balance_after           NUMERIC(18, 4),
    reconciled_status       TEXT NOT NULL DEFAULT 'UNMATCHED' CHECK (reconciled_status IN ('UNMATCHED', 'PARTIALLY_MATCHED', 'MATCHED')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bank_statement_line
        UNIQUE (statement_id, line_number)
);

CREATE INDEX idx_statement_lines_stmt ON banking.statement_lines(statement_id);
CREATE INDEX idx_statement_lines_status ON banking.statement_lines(reconciled_status);
CREATE INDEX idx_statement_lines_date ON banking.statement_lines(transaction_date);

-- ============================================================================
-- 8. BANKING: RECONCILIATIONS (جلسات المطابقة والتسوية المصرفية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS banking.reconciliations (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    bank_account_id             UUID NOT NULL REFERENCES banking.bank_accounts(id) ON DELETE CASCADE,
    statement_id                UUID REFERENCES banking.statements(id) ON DELETE SET NULL,
    reconciliation_number       TEXT NOT NULL,
    as_of_date                  DATE NOT NULL,
    statement_ending_balance    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    gl_book_balance             NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    unreconciled_difference     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    status                      TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'IN_PROGRESS', 'COMPLETED', 'LOCKED')),
    notes                       TEXT,
    reconciled_by_user_id       UUID REFERENCES iam.users(id),
    reconciled_at               TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bank_recon_num
        UNIQUE (legal_entity_id, reconciliation_number)
);

CREATE INDEX idx_bank_reconciliations_entity ON banking.reconciliations(legal_entity_id);
CREATE INDEX idx_bank_reconciliations_acc ON banking.reconciliations(bank_account_id);
CREATE INDEX idx_bank_reconciliations_status ON banking.reconciliations(status);

-- ============================================================================
-- 9. BANKING: RECONCILIATION MATCHES (سجلات الربط والمطابقة الفعلية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS banking.reconciliation_matches (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    reconciliation_id       UUID NOT NULL REFERENCES banking.reconciliations(id) ON DELETE CASCADE,
    statement_line_id       UUID NOT NULL REFERENCES banking.statement_lines(id) ON DELETE CASCADE,
    match_type              TEXT NOT NULL CHECK (match_type IN ('EXACT', 'AUTO_RULE', 'MANUAL', 'ADJUSTMENT')),
    matched_entity_type     TEXT NOT NULL CHECK (matched_entity_type IN ('JOURNAL_ENTRY', 'RECEIPT', 'PAYMENT', 'TRANSFER', 'BANK_FEE_ADJUSTMENT')),
    matched_entity_id       UUID,
    matched_amount          NUMERIC(18, 4) NOT NULL CHECK (matched_amount > 0),
    difference_amount       NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    notes                   TEXT,
    matched_by_user_id      UUID REFERENCES iam.users(id),
    matched_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recon_matches_session ON banking.reconciliation_matches(reconciliation_id);
CREATE INDEX idx_recon_matches_line ON banking.reconciliation_matches(statement_line_id);

-- ============================================================================
-- 10. PERMISSIONS SEEDING
-- ============================================================================
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'treasury.cash.create',       'Create Cash Account',          'إنشاء حساب صندوق / عهدة',        'TREASURY', now()),
    (uuidv7(), 'treasury.cash.read',         'Read Cash Accounts',           'عرض حسابات الصناديق والعهد',      'TREASURY', now()),
    (uuidv7(), 'treasury.receipt.create',    'Create Treasury Receipt',      'إنشاء سند قبض خزينة',            'TREASURY', now()),
    (uuidv7(), 'treasury.receipt.read',      'Read Treasury Receipts',       'عرض سندات قبض الخزينة',           'TREASURY', now()),
    (uuidv7(), 'treasury.receipt.post',      'Post Treasury Receipt',        'ترحيل سند قبض الخزينة',           'TREASURY', now()),
    (uuidv7(), 'treasury.payment.create',    'Create Treasury Payment',      'إنشاء سند صرف خزينة',            'TREASURY', now()),
    (uuidv7(), 'treasury.payment.read',      'Read Treasury Payments',       'عرض سندات صرف الخزينة',           'TREASURY', now()),
    (uuidv7(), 'treasury.payment.approve',   'Approve Treasury Payment',     'اعتماد سند صرف الخزينة',          'TREASURY', now()),
    (uuidv7(), 'treasury.transfer.create',   'Create Treasury Transfer',     'إنشاء تحويل نقدية / بنك',         'TREASURY', now()),
    (uuidv7(), 'treasury.transfer.read',     'Read Treasury Transfers',      'عرض تحويلات الخزينة والبنوك',     'TREASURY', now()),
    (uuidv7(), 'treasury.transfer.complete', 'Complete Treasury Transfer',  'إتمام وتأكيد التحويل المالي',     'TREASURY', now()),
    (uuidv7(), 'banking.account.create',     'Create Bank Account',          'إنشاء حساب بنكي',                'BANKING',  now()),
    (uuidv7(), 'banking.account.read',       'Read Bank Accounts',           'عرض الحسابات البنكية',            'BANKING',  now()),
    (uuidv7(), 'banking.statement.import',   'Import Bank Statement',        'استيراد كشف حساب بنكي',          'BANKING',  now()),
    (uuidv7(), 'banking.statement.read',     'Read Bank Statements',         'عرض كشوف الحسابات البنكية',       'BANKING',  now()),
    (uuidv7(), 'banking.reconcile.execute',  'Execute Bank Reconciliation',  'تنفيذ المطابقة البنكية',          'BANKING',  now()),
    (uuidv7(), 'banking.reconcile.lock',     'Lock Bank Reconciliation',     'إقفال واعتماد المطابقة البنكية',   'BANKING',  now())
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- 11. SEED FUNCTION
-- ============================================================================
CREATE OR REPLACE FUNCTION treasury.seed_default_treasury_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_admin_id UUID;
    v_branch_id UUID;
    v_petty_gl_id UUID;
    v_rajhi_gl_id UUID;
    v_snb_gl_id UUID;
    v_fees_gl_id UUID;
    v_petty_cash_id UUID;
    v_register_cash_id UUID;
    v_rajhi_bank_id UUID;
    v_snb_bank_id UUID;
    v_receipt_id UUID;
    v_payment_id UUID;
    v_transfer_id UUID;
    v_stmt_id UUID;
    v_line_1_id UUID;
    v_line_2_id UUID;
    v_recon_id UUID;
BEGIN
    -- 1. Admin user
    SELECT id INTO v_admin_id FROM iam.users WHERE username = 'admin' LIMIT 1;
    IF v_admin_id IS NULL THEN
        SELECT id INTO v_admin_id FROM iam.users LIMIT 1;
    END IF;

    -- 2. Default branch
    SELECT id INTO v_branch_id FROM organization.branches WHERE legal_entity_id = p_legal_entity_id LIMIT 1;

    -- 3. GL Accounts
    SELECT id INTO v_petty_gl_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '111001' LIMIT 1;
    SELECT id INTO v_rajhi_gl_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '111002' LIMIT 1;
    SELECT id INTO v_snb_gl_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '111003' LIMIT 1;
    SELECT id INTO v_fees_gl_id FROM accounting.accounts WHERE legal_entity_id = p_legal_entity_id AND code = '526000' LIMIT 1;

    -- 4. Seed Cash Accounts
    INSERT INTO treasury.cash_accounts (
        id, legal_entity_id, branch_id, account_code, name_en, name_ar,
        account_type, gl_account_id, currency_code, custodian_user_id,
        max_limit, current_balance, status
    ) VALUES (
        uuidv7(), p_legal_entity_id, v_branch_id, 'CASH-HO-01',
        'Head Office Petty Cash', 'صندوق المصروفات النثرية - الإدارة العامة',
        'PETTY_CASH', v_petty_gl_id, 'SAR', v_admin_id,
        5000.0000, 3500.0000, 'ACTIVE'
    ) ON CONFLICT (legal_entity_id, account_code) DO UPDATE
        SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_petty_cash_id;

    INSERT INTO treasury.cash_accounts (
        id, legal_entity_id, branch_id, account_code, name_en, name_ar,
        account_type, gl_account_id, currency_code, custodian_user_id,
        max_limit, current_balance, status
    ) VALUES (
        uuidv7(), p_legal_entity_id, v_branch_id, 'CASH-REG-01',
        'Main Branch POS Register', 'صندوق نقطة بيع الفرع الرئيسي',
        'CASH_REGISTER', v_petty_gl_id, 'SAR', v_admin_id,
        20000.0000, 12450.0000, 'ACTIVE'
    ) ON CONFLICT (legal_entity_id, account_code) DO NOTHING
    RETURNING id INTO v_register_cash_id;

    -- 5. Seed Bank Accounts
    INSERT INTO banking.bank_accounts (
        id, legal_entity_id, account_code, bank_name_en, bank_name_ar,
        branch_name, account_number, account_number_masked, iban, swift_bic,
        currency_code, gl_account_id, opening_balance, current_balance, available_balance,
        overdraft_limit, status, is_primary, contact_person, contact_phone
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'BANK-RAJHI-01',
        'Al-Rajhi Bank', 'مصرف الراجحي',
        'Riyadh Olaya Branch', '204608010001234', '**** **** **** 1234',
        'SA4480000204608010001234', 'RJHISARI',
        'SAR', v_rajhi_gl_id, 100000.0000, 185200.0000, 185200.0000,
        0.0000, 'ACTIVE', true, 'مدير علاقات الشركات', '+966112003000'
    ) ON CONFLICT (legal_entity_id, account_code) DO UPDATE
        SET bank_name_ar = EXCLUDED.bank_name_ar
    RETURNING id INTO v_rajhi_bank_id;

    INSERT INTO banking.bank_accounts (
        id, legal_entity_id, account_code, bank_name_en, bank_name_ar,
        branch_name, account_number, account_number_masked, iban, swift_bic,
        currency_code, gl_account_id, opening_balance, current_balance, available_balance,
        overdraft_limit, status, is_primary, contact_person, contact_phone
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'BANK-SNB-01',
        'Saudi National Bank (SNB)', 'البنك الأهلي السعودي',
        'King Fahd Road Branch', '101000987654321', '**** **** **** 4321',
        'SA8810000101000987654321', 'NCBKSARI',
        'SAR', v_snb_gl_id, 50000.0000, 94000.0000, 94000.0000,
        0.0000, 'ACTIVE', false, 'قسم الحسابات الخاصة', '+966114005000'
    ) ON CONFLICT (legal_entity_id, account_code) DO NOTHING
    RETURNING id INTO v_snb_bank_id;

    -- 6. Sample Receipt Voucher (Posted)
    IF v_register_cash_id IS NOT NULL THEN
        INSERT INTO treasury.receipts (
            id, legal_entity_id, receipt_number, cash_account_id, receipt_date,
            receipt_type, received_from, amount, currency_code, payment_method,
            reference_number, description, status, created_by_user_id
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'REC-2026-0001', v_register_cash_id, CURRENT_DATE,
            'CASH_SALE', 'مبيعات نقدية مباشرة للفرع', 2500.0000, 'SAR', 'CASH',
            'POS-REF-9901', 'توريد نقدي فوري لنقطة البيع رقم 1', 'POSTED', v_admin_id
        ) ON CONFLICT (legal_entity_id, receipt_number) DO NOTHING
        RETURNING id INTO v_receipt_id;
    END IF;

    -- 7. Sample Payment Voucher (Approved / Paid)
    IF v_petty_cash_id IS NOT NULL THEN
        INSERT INTO treasury.payments (
            id, legal_entity_id, payment_number, cash_account_id, payment_date,
            payment_type, paid_to, amount, currency_code,
            reference_number, description, status, created_by_user_id, approved_by_user_id
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'PMT-2026-0001', v_petty_cash_id, CURRENT_DATE,
            'PETTY_CASH_EXPENSE', 'شركة جرير للمستلزمات المكتبية', 350.0000, 'SAR',
            'INV-JARIR-4421', 'شراء قرطاسية وأحبار طباعة للمكتب الرئيسي', 'PAID', v_admin_id, v_admin_id
        ) ON CONFLICT (legal_entity_id, payment_number) DO NOTHING
        RETURNING id INTO v_payment_id;
    END IF;

    -- 8. Sample Transfer (Cash Register -> Bank)
    IF v_register_cash_id IS NOT NULL AND v_rajhi_bank_id IS NOT NULL THEN
        INSERT INTO treasury.transfers (
            id, legal_entity_id, transfer_number, transfer_date, source_type,
            source_cash_account_id, destination_type, destination_bank_account_id,
            amount, currency_code, fee_amount, status, reference_number, notes,
            initiated_by_user_id, completed_by_user_id
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'TRF-2026-0001', CURRENT_DATE, 'CASH',
            v_register_cash_id, 'BANK', v_rajhi_bank_id,
            5000.0000, 'SAR', 0.0000, 'COMPLETED', 'DEP-RAJHI-8891',
            'إيداع نقدي يومي بحساب الراجحي الجاري من إيرادات الصندوق',
            v_admin_id, v_admin_id
        ) ON CONFLICT (legal_entity_id, transfer_number) DO NOTHING
        RETURNING id INTO v_transfer_id;
    END IF;

    -- 9. Sample Bank Statement & Lines
    IF v_rajhi_bank_id IS NOT NULL THEN
        INSERT INTO banking.statements (
            id, legal_entity_id, bank_account_id, statement_number, statement_date,
            start_date, end_date, opening_balance, closing_balance, total_credits,
            total_debits, currency_code, file_format, status, source_filename, imported_by_user_id
        ) VALUES (
            uuidv7(), p_legal_entity_id, v_rajhi_bank_id, 'STMT-2026-01', CURRENT_DATE,
            CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE, 100000.0000, 104985.0000,
            5000.0000, 15.0000, 'SAR', 'CSV', 'PARTIALLY_RECONCILED', 'rajhi_statement_jan2026.csv', v_admin_id
        ) ON CONFLICT (bank_account_id, statement_number) DO NOTHING
        RETURNING id INTO v_stmt_id;

        IF v_stmt_id IS NOT NULL THEN
            INSERT INTO banking.statement_lines (
                id, statement_id, line_number, transaction_date, value_date,
                description, reference_number, transaction_type, amount, currency_code,
                balance_after, reconciled_status
            ) VALUES (
                uuidv7(), v_stmt_id, 1, CURRENT_DATE, CURRENT_DATE,
                'Cash Deposit Branch 01', 'DEP-RAJHI-8891', 'CREDIT', 5000.0000, 'SAR',
                105000.0000, 'MATCHED'
            ) ON CONFLICT (statement_id, line_number) DO NOTHING
            RETURNING id INTO v_line_1_id;

            INSERT INTO banking.statement_lines (
                id, statement_id, line_number, transaction_date, value_date,
                description, reference_number, transaction_type, amount, currency_code,
                balance_after, reconciled_status
            ) VALUES (
                uuidv7(), v_stmt_id, 2, CURRENT_DATE, CURRENT_DATE,
                'Monthly Account Maintenance Fee', 'FEE-2026-01', 'DEBIT', 15.0000, 'SAR',
                104985.0000, 'UNMATCHED'
            ) ON CONFLICT (statement_id, line_number) DO NOTHING
            RETURNING id INTO v_line_2_id;

            -- 10. Sample Reconciliation Session
            INSERT INTO banking.reconciliations (
                id, legal_entity_id, bank_account_id, statement_id,
                reconciliation_number, as_of_date, statement_ending_balance,
                gl_book_balance, unreconciled_difference, status, notes,
                reconciled_by_user_id
            ) VALUES (
                uuidv7(), p_legal_entity_id, v_rajhi_bank_id, v_stmt_id,
                'REC-2026-0001', CURRENT_DATE, 104985.0000,
                105000.0000, 15.0000, 'IN_PROGRESS', 'مطابقة شهرية - جاري تسوية عمولة البنك الشهرية 15 ريال',
                v_admin_id
            ) ON CONFLICT (legal_entity_id, reconciliation_number) DO NOTHING
            RETURNING id INTO v_recon_id;

            IF v_recon_id IS NOT NULL AND v_line_1_id IS NOT NULL AND v_transfer_id IS NOT NULL THEN
                INSERT INTO banking.reconciliation_matches (
                    id, reconciliation_id, statement_line_id, match_type,
                    matched_entity_type, matched_entity_id, matched_amount, difference_amount,
                    notes, matched_by_user_id
                ) VALUES (
                    uuidv7(), v_recon_id, v_line_1_id, 'EXACT',
                    'TRANSFER', v_transfer_id, 5000.0000, 0.0000,
                    'مطابقة تامة مع الحوالة الداخلية النقدية', v_admin_id
                );
            END IF;
        END IF;
    END IF;

END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS treasury.seed_default_treasury_data(UUID);
DROP TABLE IF EXISTS banking.reconciliation_matches CASCADE;
DROP TABLE IF EXISTS banking.reconciliations CASCADE;
DROP TABLE IF EXISTS banking.statement_lines CASCADE;
DROP TABLE IF EXISTS banking.statements CASCADE;
DROP TABLE IF EXISTS treasury.transfers CASCADE;
DROP TABLE IF EXISTS treasury.payments CASCADE;
DROP TABLE IF EXISTS treasury.receipts CASCADE;
DROP TABLE IF EXISTS banking.bank_accounts CASCADE;
DROP TABLE IF EXISTS treasury.cash_accounts CASCADE;
DROP SCHEMA IF EXISTS banking CASCADE;
DROP SCHEMA IF EXISTS treasury CASCADE;
-- +goose StatementEnd
