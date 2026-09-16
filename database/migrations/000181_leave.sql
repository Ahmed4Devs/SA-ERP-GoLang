-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS leave;

-- 1. LEAVE TYPES (SAUDI LABOR LAW STATUTORY TYPES)
CREATE TABLE IF NOT EXISTS leave.types (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                    TEXT NOT NULL,
    name_en                 TEXT NOT NULL,
    name_ar                 TEXT NOT NULL,
    is_paid                 BOOLEAN NOT NULL DEFAULT TRUE,
    requires_attachment     BOOLEAN NOT NULL DEFAULT FALSE,
    deducts_from_annual     BOOLEAN NOT NULL DEFAULT FALSE,
    max_days_per_year       NUMERIC(5,2),
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_leave_types_code UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_leave_types_entity ON leave.types(legal_entity_id);

-- 2. LEAVE POLICIES
CREATE TABLE IF NOT EXISTS leave.policies (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    leave_type_id               UUID NOT NULL REFERENCES leave.types(id) ON DELETE CASCADE,
    accrual_frequency           TEXT NOT NULL DEFAULT 'MONTHLY' CHECK (accrual_frequency IN ('MONTHLY', 'ANNUAL', 'NONE')),
    accrual_rate_per_period     NUMERIC(5,2) NOT NULL DEFAULT 1.75,
    max_carryover_days          NUMERIC(5,2) NOT NULL DEFAULT 10.00,
    probation_restricted        BOOLEAN NOT NULL DEFAULT TRUE,
    encashment_allowed          BOOLEAN NOT NULL DEFAULT TRUE,
    min_service_days_required   INT NOT NULL DEFAULT 0,
    is_active                   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_leave_policy_type UNIQUE (legal_entity_id, leave_type_id)
);

-- 3. EMPLOYEE LEAVE BALANCES
CREATE TABLE IF NOT EXISTS leave.employee_balances (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id         UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    leave_type_id       UUID NOT NULL REFERENCES leave.types(id) ON DELETE RESTRICT,
    fiscal_year         INT NOT NULL,
    allocated_days      NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    accrued_days        NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    taken_days          NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    pending_days        NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    carryover_days      NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    closing_balance     NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_leave_emp_balance UNIQUE (employee_id, leave_type_id, fiscal_year)
);

CREATE INDEX IF NOT EXISTS idx_leave_balances_emp_year ON leave.employee_balances(employee_id, fiscal_year);

-- 4. LEAVE REQUESTS
CREATE TABLE IF NOT EXISTS leave.requests (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    employee_id                 UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    leave_type_id               UUID NOT NULL REFERENCES leave.types(id) ON DELETE RESTRICT,
    request_number              TEXT NOT NULL,
    start_date                  DATE NOT NULL,
    end_date                    DATE NOT NULL,
    total_days                  NUMERIC(5,2) NOT NULL,
    return_date                 DATE NOT NULL,
    replacement_employee_id     UUID REFERENCES hr.employees(id) ON DELETE SET NULL,
    reason                      TEXT NOT NULL,
    attachment_url              TEXT,
    status                      TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED')),
    reviewed_by_user_id         UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    reviewed_at                 TIMESTAMPTZ,
    review_notes                TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_leave_request_num UNIQUE (legal_entity_id, request_number)
);

CREATE INDEX IF NOT EXISTS idx_leave_requests_entity ON leave.requests(legal_entity_id, status);
CREATE INDEX IF NOT EXISTS idx_leave_requests_emp ON leave.requests(employee_id, start_date);

-- 5. LEAVE TRANSACTIONS (APPEND-ONLY LEDGER FOR 100% RECONCILABILITY)
CREATE TABLE IF NOT EXISTS leave.transactions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id         UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    leave_type_id       UUID NOT NULL REFERENCES leave.types(id) ON DELETE RESTRICT,
    fiscal_year         INT NOT NULL,
    transaction_type    TEXT NOT NULL CHECK (transaction_type IN ('ACCRUAL', 'USAGE', 'CANCELLATION', 'CARRYOVER', 'ADJUSTMENT', 'ENCASHMENT')),
    days                NUMERIC(5,2) NOT NULL,
    reference_id        UUID,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_leave_transactions_emp ON leave.transactions(employee_id, leave_type_id, fiscal_year);

-- 6. PERMISSIONS
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'leave.type.read',         'View Leave Types and Policies', 'استعراض أنواع الإجازات والسياسات', 'LEAVE', now()),
    (uuidv7(), 'leave.type.manage',       'Manage Leave Policies',         'إدارة سياسات الإجازات السنوية',    'LEAVE', now()),
    (uuidv7(), 'leave.request.read',      'View Leave Requests',           'استعراض طلبات الإجازات',           'LEAVE', now()),
    (uuidv7(), 'leave.request.create',    'Submit Leave Request',          'تقديم طلب إجازة',                  'LEAVE', now()),
    (uuidv7(), 'leave.request.approve',   'Approve Leave Request',         'اعتماد أو رفض الإجازات',           'LEAVE', now()),
    (uuidv7(), 'leave.balance.manage',    'Manage Leave Balances',         'إدارة وتسوية أرصدة الإجازات',      'LEAVE', now())
ON CONFLICT (code) DO NOTHING;

-- 7. SEED DEFAULT LEAVE DATA FUNCTION
CREATE OR REPLACE FUNCTION leave.seed_default_leave_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_annual_id UUID;
    v_sick_id UUID;
    v_mat_id UUID;
    v_pat_id UUID;
    v_mar_id UUID;
    v_ber_id UUID;
    v_hajj_id UUID;
    v_exam_id UUID;
    v_unpaid_id UUID;
    v_emp_id UUID;
    v_current_year INT := EXTRACT(YEAR FROM CURRENT_DATE);
BEGIN
    -- 1. Statutory Leave Types
    -- Annual Leave (Article 109 - 21 to 30 days)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'ANNUAL', 'Annual Leave', 'الإجازة السنوية الاعتيادية', TRUE, FALSE, 30.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_annual_id;

    -- Sick Leave (Article 117 - 120 days total)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'SICK', 'Sick Leave', 'الإجازة المرضية (تقرير طبي معتمد)', TRUE, TRUE, 120.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_sick_id;

    -- Maternity Leave (Article 151 - 10-12 weeks)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'MATERNITY', 'Maternity Leave', 'إجازة الوضع والأمومة (10 أسابيع)', TRUE, TRUE, 70.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_mat_id;

    -- Paternity Leave (Article 113 - 3 days)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'PATERNITY', 'Paternity Leave', 'إجازة مولود (3 أيام براتب كامل)', TRUE, TRUE, 3.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_pat_id;

    -- Marriage Leave (Article 113 - 5 days)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'MARRIAGE', 'Marriage Leave', 'إجازة زواج (5 أيام براتب كامل)', TRUE, TRUE, 5.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_mar_id;

    -- Bereavement Leave (Article 113 - 5 days)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'BEREAVEMENT', 'Bereavement Leave', 'إجازة وفاة قريب (5 أيام براتب كامل)', TRUE, TRUE, 5.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_ber_id;

    -- Hajj Leave (Article 114 - 10-15 days once)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'HAJJ', 'Hajj Pilgrimage Leave', 'إجازة أداء فريضة الحج (لمرة واحدة)', TRUE, TRUE, 15.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_hajj_id;

    -- Exam Leave (Article 115)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'EXAM', 'Study and Exam Leave', 'إجازة أداء الامتحانات الدراسية', TRUE, TRUE, 15.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_exam_id;

    -- Unpaid Leave (Article 116)
    INSERT INTO leave.types (id, legal_entity_id, code, name_en, name_ar, is_paid, requires_attachment, max_days_per_year)
    VALUES (uuidv7(), p_legal_entity_id, 'UNPAID', 'Unpaid Leave', 'إجازة بدون راتب (باتفاق الطرفين)', FALSE, FALSE, 90.00)
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en RETURNING id INTO v_unpaid_id;

    -- 2. Standard Annual Leave Policy (1.75 days / month = 21 days / year)
    INSERT INTO leave.policies (
        id, legal_entity_id, leave_type_id, accrual_frequency, accrual_rate_per_period,
        max_carryover_days, probation_restricted, encashment_allowed
    ) VALUES (
        uuidv7(), p_legal_entity_id, v_annual_id, 'MONTHLY', 1.75, 10.00, TRUE, TRUE
    )
    ON CONFLICT (legal_entity_id, leave_type_id) DO NOTHING;

    -- 3. Initialize Balances for active employees for Annual Leave
    FOR v_emp_id IN (SELECT id FROM hr.employees WHERE legal_entity_id = p_legal_entity_id AND employment_status = 'ACTIVE') LOOP
        IF NOT EXISTS (SELECT 1 FROM leave.employee_balances WHERE employee_id = v_emp_id AND leave_type_id = v_annual_id AND fiscal_year = v_current_year) THEN
            INSERT INTO leave.employee_balances (
                id, employee_id, leave_type_id, fiscal_year, allocated_days, accrued_days, closing_balance
            ) VALUES (
                uuidv7(), v_emp_id, v_annual_id, v_current_year, 21.00, 10.50, 10.50
            );

            INSERT INTO leave.transactions (
                id, employee_id, leave_type_id, fiscal_year, transaction_type, days, notes
            ) VALUES (
                uuidv7(), v_emp_id, v_annual_id, v_current_year, 'ACCRUAL', 10.50, 'Initial opening accrual balance for fiscal year'
            );
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS leave.seed_default_leave_data(UUID);
DROP TABLE IF EXISTS leave.transactions CASCADE;
DROP TABLE IF EXISTS leave.requests CASCADE;
DROP TABLE IF EXISTS leave.employee_balances CASCADE;
DROP TABLE IF EXISTS leave.policies CASCADE;
DROP TABLE IF EXISTS leave.types CASCADE;
DROP SCHEMA IF EXISTS leave CASCADE;
-- +goose StatementEnd
