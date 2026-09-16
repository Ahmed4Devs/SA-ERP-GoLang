-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS hr;

-- ============================================================================
-- 1. JOBS (دليل الوظائف والمهام)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.jobs (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                TEXT NOT NULL,
    title_en            TEXT NOT NULL,
    title_ar            TEXT NOT NULL,
    job_family          TEXT NOT NULL CHECK (job_family IN ('ENGINEERING', 'FINANCE', 'SALES', 'OPERATIONS', 'HR', 'LEGAL', 'MANAGEMENT')),
    job_level           TEXT NOT NULL CHECK (job_level IN ('ENTRY', 'MID', 'SENIOR', 'LEAD', 'MANAGER', 'DIRECTOR', 'EXECUTIVE')),
    description         TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_jobs_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX idx_hr_jobs_entity ON hr.jobs(legal_entity_id);
CREATE INDEX idx_hr_jobs_family ON hr.jobs(job_family);

-- ============================================================================
-- 2. POSITIONS (المناصب الشاغرة والهيكل الوظيفي)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.positions (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                    TEXT NOT NULL,
    job_id                  UUID NOT NULL REFERENCES hr.jobs(id) ON DELETE RESTRICT,
    department_id           UUID NOT NULL REFERENCES organization.departments(id) ON DELETE RESTRICT,
    branch_id               UUID REFERENCES organization.branches(id) ON DELETE SET NULL,
    reports_to_position_id  UUID REFERENCES hr.positions(id) ON DELETE SET NULL,
    headcount               INT NOT NULL DEFAULT 1 CHECK (headcount >= 1),
    status                  TEXT NOT NULL DEFAULT 'VACANT' CHECK (status IN ('VACANT', 'FILLED', 'FROZEN')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_positions_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX idx_hr_positions_entity ON hr.positions(legal_entity_id);
CREATE INDEX idx_hr_positions_job ON hr.positions(job_id);
CREATE INDEX idx_hr_positions_dept ON hr.positions(department_id);
CREATE INDEX idx_hr_positions_status ON hr.positions(status);

-- ============================================================================
-- 3. EMPLOYEES (سجل الموظفين الرئيسي)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.employees (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    employee_number         TEXT NOT NULL,
    user_id                 UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    first_name_en           TEXT NOT NULL,
    last_name_en            TEXT NOT NULL,
    first_name_ar           TEXT NOT NULL,
    last_name_ar            TEXT NOT NULL,
    gender                  TEXT NOT NULL CHECK (gender IN ('MALE', 'FEMALE')),
    marital_status          TEXT NOT NULL DEFAULT 'SINGLE' CHECK (marital_status IN ('SINGLE', 'MARRIED', 'DIVORCED', 'WIDOWED')),
    date_of_birth           DATE NOT NULL,
    nationality_country_code TEXT NOT NULL DEFAULT 'SA',
    is_saudi_national       BOOLEAN NOT NULL DEFAULT true,
    hire_date               DATE NOT NULL,
    probation_end_date      DATE,
    termination_date        DATE,
    employment_status       TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (employment_status IN ('PROBATION', 'ACTIVE', 'ON_LEAVE', 'SUSPENDED', 'TERMINATED')),
    primary_position_id     UUID REFERENCES hr.positions(id) ON DELETE SET NULL,
    department_id           UUID REFERENCES organization.departments(id) ON DELETE SET NULL,
    branch_id               UUID REFERENCES organization.branches(id) ON DELETE SET NULL,
    manager_employee_id     UUID REFERENCES hr.employees(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_employee_number
        UNIQUE (legal_entity_id, employee_number)
);

CREATE INDEX idx_hr_employees_entity ON hr.employees(legal_entity_id);
CREATE INDEX idx_hr_employees_status ON hr.employees(employment_status);
CREATE INDEX idx_hr_employees_dept ON hr.employees(department_id);
CREATE INDEX idx_hr_employees_saudi ON hr.employees(is_saudi_national);

-- ============================================================================
-- 4. EMPLOYEE IDENTIFIERS (الهويات الرسمية: الهوية الوطنية، الإقامة، الجواز)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.employee_identifiers (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id             UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    identifier_type         TEXT NOT NULL CHECK (identifier_type IN ('NATIONAL_ID', 'IQAMA', 'PASSPORT', 'BORDER_NUMBER')),
    identifier_value        TEXT NOT NULL,
    identifier_value_masked TEXT NOT NULL,
    issue_date              DATE,
    expiry_date             DATE,
    issue_place             TEXT,
    is_primary              BOOLEAN NOT NULL DEFAULT true,
    is_verified             BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_employee_identifier
        UNIQUE (employee_id, identifier_type, identifier_value)
);

CREATE INDEX idx_hr_identifiers_emp ON hr.employee_identifiers(employee_id);
CREATE INDEX idx_hr_identifiers_type ON hr.employee_identifiers(identifier_type);

-- ============================================================================
-- 5. EMPLOYEE CONTACTS (العنوان الوطني ووسائل الاتصال)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.employee_contacts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id         UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    work_email          TEXT,
    personal_email      TEXT,
    mobile_phone        TEXT NOT NULL,
    work_phone          TEXT,
    extension           TEXT,
    building_no         TEXT,
    street_name         TEXT,
    district            TEXT,
    city                TEXT NOT NULL DEFAULT 'Riyadh',
    postal_code         TEXT,
    additional_no       TEXT,
    country_code        TEXT NOT NULL DEFAULT 'SA',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_employee_contact
        UNIQUE (employee_id)
);

CREATE INDEX idx_hr_contacts_emp ON hr.employee_contacts(employee_id);

-- ============================================================================
-- 6. EMERGENCY CONTACTS (بيانات الاتصال للطوارئ)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.emergency_contacts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id         UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    contact_name        TEXT NOT NULL,
    relationship        TEXT NOT NULL CHECK (relationship IN ('SPOUSE', 'PARENT', 'SIBLING', 'CHILD', 'OTHER')),
    primary_phone       TEXT NOT NULL,
    secondary_phone     TEXT,
    address             TEXT,
    is_primary          BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hr_emergency_emp ON hr.emergency_contacts(employee_id);

-- ============================================================================
-- 7. QUALIFICATIONS (المؤهلات العلمية والشهادات)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.qualifications (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id         UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    degree_type         TEXT NOT NULL CHECK (degree_type IN ('HIGH_SCHOOL', 'DIPLOMA', 'BACHELOR', 'MASTER', 'DOCTORATE', 'PROFESSIONAL_CERTIFICATE')),
    field_of_study      TEXT NOT NULL,
    institution_name    TEXT NOT NULL,
    graduation_year     INT NOT NULL CHECK (graduation_year >= 1950 AND graduation_year <= 2100),
    grade_gpa           TEXT,
    country_code        TEXT NOT NULL DEFAULT 'SA',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hr_qualifications_emp ON hr.qualifications(employee_id);

-- ============================================================================
-- 8. CONTRACTS (عقود العمل - متوافق مع منصة قوى)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.contracts (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    employee_id             UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    contract_number         TEXT NOT NULL,
    contract_type           TEXT NOT NULL CHECK (contract_type IN ('UNSPECIFIED_TERM', 'SPECIFIED_TERM', 'TEMPORARY', 'SEASONAL')),
    start_date              DATE NOT NULL,
    end_date                DATE,
    probation_days          INT NOT NULL DEFAULT 90 CHECK (probation_days >= 0 AND probation_days <= 180),
    notice_period_days      INT NOT NULL DEFAULT 60 CHECK (notice_period_days >= 0),
    working_hours_per_week  NUMERIC(5, 2) NOT NULL DEFAULT 40.00 CHECK (working_hours_per_week > 0),
    work_location           TEXT NOT NULL DEFAULT 'Riyadh HQ',
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'ACTIVE', 'EXPIRED', 'TERMINATED')),
    notes                   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_contract_number
        UNIQUE (legal_entity_id, contract_number)
);

CREATE INDEX idx_hr_contracts_entity ON hr.contracts(legal_entity_id);
CREATE INDEX idx_hr_contracts_emp ON hr.contracts(employee_id);
CREATE INDEX idx_hr_contracts_status ON hr.contracts(status);

-- ============================================================================
-- 9. CONTRACT VERSIONS (تاريخ إصدارات العقد والرواتب والبدلات)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.contract_versions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    contract_id         UUID NOT NULL REFERENCES hr.contracts(id) ON DELETE CASCADE,
    version_number      INT NOT NULL DEFAULT 1,
    effective_from      DATE NOT NULL,
    effective_to        DATE,
    basic_salary        NUMERIC(18, 4) NOT NULL CHECK (basic_salary >= 0),
    housing_allowance   NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (housing_allowance >= 0),
    transport_allowance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (transport_allowance >= 0),
    other_allowances    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (other_allowances >= 0),
    gross_salary        NUMERIC(18, 4) NOT NULL CHECK (gross_salary >= 0),
    change_reason       TEXT,
    approved_by_user_id UUID REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_contract_version
        UNIQUE (contract_id, version_number)
);

CREATE INDEX idx_hr_contract_versions_contract ON hr.contract_versions(contract_id);

-- ============================================================================
-- 10. BANK ACCOUNTS (حسابات الرواتب البنكية - نظام حماية الأجور WPS)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.bank_accounts (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id     UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    bank_name_en    TEXT NOT NULL,
    bank_name_ar    TEXT NOT NULL,
    iban            TEXT NOT NULL,
    iban_masked     TEXT NOT NULL,
    swift_bic       TEXT NOT NULL,
    is_primary      BOOLEAN NOT NULL DEFAULT true,
    account_status  TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (account_status IN ('ACTIVE', 'INACTIVE')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_hr_employee_iban
        UNIQUE (employee_id, iban)
);

CREATE INDEX idx_hr_bank_accounts_emp ON hr.bank_accounts(employee_id);

-- ============================================================================
-- 11. EMPLOYEE DOCUMENTS (المستندات والوثائق المرفقة)
-- ============================================================================
CREATE TABLE IF NOT EXISTS hr.employee_documents (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id         UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    document_type       TEXT NOT NULL CHECK (document_type IN ('NATIONAL_ID_COPY', 'PASSPORT_COPY', 'IQAMA_COPY', 'CONTRACT_SIGNED', 'CERTIFICATE', 'RESUME', 'OTHER')),
    title               TEXT NOT NULL,
    file_name           TEXT NOT NULL,
    file_size_bytes     BIGINT NOT NULL DEFAULT 0,
    mime_type           TEXT NOT NULL DEFAULT 'application/pdf',
    storage_path        TEXT NOT NULL,
    uploaded_by_user_id UUID REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hr_documents_emp ON hr.employee_documents(employee_id);

-- ============================================================================
-- 12. PERMISSIONS SEEDING
-- ============================================================================
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'hr.employee.create',    'Create Employee',             'إنشاء ملف موظف جديد',       'HR', now()),
    (uuidv7(), 'hr.employee.read',      'Read Employee Records',       'استعراض بيانات الموظفين',    'HR', now()),
    (uuidv7(), 'hr.employee.update',    'Update Employee Details',     'تعديل بيانات الموظف',        'HR', now()),
    (uuidv7(), 'hr.employee.terminate', 'Terminate Employee Contract', 'إنهاء خدمات الموظف',        'HR', now()),
    (uuidv7(), 'hr.contract.create',    'Create Contract',             'إنشاء عقد عمل جديد',         'HR', now()),
    (uuidv7(), 'hr.contract.read',      'Read Contracts',              'استعراض عقود العمل',        'HR', now()),
    (uuidv7(), 'hr.contract.amend',     'Amend Contract Version',      'تعديل إصدار وبيانات العقد',  'HR', now()),
    (uuidv7(), 'hr.contract.activate',  'Activate Contract',           'اعتماد وتفعيل عقد العمل',    'HR', now()),
    (uuidv7(), 'hr.job.manage',         'Manage Jobs and Positions',   'إدارة الوظائف والمناصب',    'HR', now()),
    (uuidv7(), 'hr.sensitive.view',     'View Sensitive Identity Data','عرض الهويات والبيانات الحساسة','HR', now())
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- 13. SEED FUNCTION
-- ============================================================================
CREATE OR REPLACE FUNCTION hr.seed_default_hr_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_admin_id UUID;
    v_dept_eng_id UUID;
    v_dept_fin_id UUID;
    v_dept_hr_id UUID;
    v_branch_id UUID;
    v_job_eng_id UUID;
    v_job_cfo_id UUID;
    v_job_hr_id UUID;
    v_pos_eng_id UUID;
    v_pos_cfo_id UUID;
    v_pos_hr_id UUID;
    v_emp_saudi_id UUID;
    v_emp_expat_id UUID;
    v_contract_1_id UUID;
    v_contract_2_id UUID;
BEGIN
    -- 1. Admin user
    SELECT id INTO v_admin_id FROM iam.users WHERE username = 'admin' LIMIT 1;
    IF v_admin_id IS NULL THEN
        SELECT id INTO v_admin_id FROM iam.users LIMIT 1;
    END IF;

    -- 2. Departments
    SELECT id INTO v_dept_eng_id FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'ENG' LIMIT 1;
    IF v_dept_eng_id IS NULL THEN
        SELECT id INTO v_dept_eng_id FROM organization.departments WHERE legal_entity_id = p_legal_entity_id LIMIT 1;
    END IF;
    SELECT id INTO v_dept_fin_id FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'FIN' LIMIT 1;
    IF v_dept_fin_id IS NULL THEN
        v_dept_fin_id := v_dept_eng_id;
    END IF;
    SELECT id INTO v_dept_hr_id FROM organization.departments WHERE legal_entity_id = p_legal_entity_id AND code = 'HR' LIMIT 1;
    IF v_dept_hr_id IS NULL THEN
        v_dept_hr_id := v_dept_eng_id;
    END IF;

    -- 3. Branch
    SELECT id INTO v_branch_id FROM organization.branches WHERE legal_entity_id = p_legal_entity_id LIMIT 1;

    -- 4. Seed Standard Jobs
    INSERT INTO hr.jobs (
        id, legal_entity_id, code, title_en, title_ar, job_family, job_level, description
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'JOB-ENG-01', 'Senior Full-Stack Engineer', 'كبير مهندسي برمجيات',
        'ENGINEERING', 'SENIOR', 'تصميم وتطوير الأنظمة السحابية والخدمات الدقيقة'
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET title_ar = EXCLUDED.title_ar
    RETURNING id INTO v_job_eng_id;

    INSERT INTO hr.jobs (
        id, legal_entity_id, code, title_en, title_ar, job_family, job_level, description
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'JOB-FIN-01', 'Chief Financial Officer', 'المدير المالي التنفيذي',
        'FINANCE', 'EXECUTIVE', 'التخطيط المالي والرقابة المحاسبية والميزانيات'
    ) ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_job_cfo_id;

    INSERT INTO hr.jobs (
        id, legal_entity_id, code, title_en, title_ar, job_family, job_level, description
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'JOB-HR-01', 'HR Specialist', 'أخصائي موارد بشرية وتوطين',
        'HR', 'MID', 'شؤون الموظفين، التوطين ونطاقات، ومنصة قوى والتأمينات'
    ) ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_job_hr_id;

    -- 5. Seed Positions
    IF v_job_eng_id IS NOT NULL AND v_dept_eng_id IS NOT NULL THEN
        INSERT INTO hr.positions (
            id, legal_entity_id, code, job_id, department_id, branch_id, headcount, status
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'POS-ENG-001', v_job_eng_id, v_dept_eng_id, v_branch_id, 1, 'FILLED'
        ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET status = 'FILLED'
        RETURNING id INTO v_pos_eng_id;
    END IF;

    IF v_job_cfo_id IS NOT NULL AND v_dept_fin_id IS NOT NULL THEN
        INSERT INTO hr.positions (
            id, legal_entity_id, code, job_id, department_id, branch_id, headcount, status
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'POS-CFO-001', v_job_cfo_id, v_dept_fin_id, v_branch_id, 1, 'FILLED'
        ) ON CONFLICT (legal_entity_id, code) DO NOTHING
        RETURNING id INTO v_pos_cfo_id;
    END IF;

    IF v_job_hr_id IS NOT NULL AND v_dept_hr_id IS NOT NULL THEN
        INSERT INTO hr.positions (
            id, legal_entity_id, code, job_id, department_id, branch_id, headcount, status
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'POS-HR-001', v_job_hr_id, v_dept_hr_id, v_branch_id, 1, 'VACANT'
        ) ON CONFLICT (legal_entity_id, code) DO NOTHING
        RETURNING id INTO v_pos_hr_id;
    END IF;

    -- 6. Seed Employees: Employee 1 (Saudi Citizen)
    INSERT INTO hr.employees (
        id, legal_entity_id, employee_number, user_id, first_name_en, last_name_en,
        first_name_ar, last_name_ar, gender, marital_status, date_of_birth,
        nationality_country_code, is_saudi_national, hire_date, probation_end_date,
        employment_status, primary_position_id, department_id, branch_id
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'EMP-1001', v_admin_id, 'Mohammed', 'Al-Ghamdi',
        'محمد', 'الغامدي', 'MALE', 'MARRIED', '1992-05-14',
        'SA', true, '2024-01-01', '2024-04-01',
        'ACTIVE', v_pos_eng_id, v_dept_eng_id, v_branch_id
    ) ON CONFLICT (legal_entity_id, employee_number) DO UPDATE
        SET first_name_ar = EXCLUDED.first_name_ar
    RETURNING id INTO v_emp_saudi_id;

    -- Employee 2 (Expatriate Professional)
    INSERT INTO hr.employees (
        id, legal_entity_id, employee_number, first_name_en, last_name_en,
        first_name_ar, last_name_ar, gender, marital_status, date_of_birth,
        nationality_country_code, is_saudi_national, hire_date, probation_end_date,
        employment_status, primary_position_id, department_id, branch_id
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'EMP-1002', 'Ahmed', 'Al-Masry',
        'أحمد', 'المصري', 'MALE', 'MARRIED', '1988-11-20',
        'EG', false, '2023-06-01', '2023-09-01',
        'ACTIVE', v_pos_cfo_id, v_dept_fin_id, v_branch_id
    ) ON CONFLICT (legal_entity_id, employee_number) DO NOTHING
    RETURNING id INTO v_emp_expat_id;

    -- 7. Seed Identifiers
    IF v_emp_saudi_id IS NOT NULL THEN
        INSERT INTO hr.employee_identifiers (
            id, employee_id, identifier_type, identifier_value, identifier_value_masked,
            issue_date, expiry_date, issue_place, is_primary, is_verified
        ) VALUES (
            uuidv7(), v_emp_saudi_id, 'NATIONAL_ID', '1088776655', '1088****55',
            '2020-01-10', '2030-01-10', 'الرياض', true, true
        ) ON CONFLICT (employee_id, identifier_type, identifier_value) DO NOTHING;

        INSERT INTO hr.employee_contacts (
            id, employee_id, work_email, personal_email, mobile_phone,
            building_no, street_name, district, city, postal_code, additional_no, country_code
        ) VALUES (
            uuidv7(), v_emp_saudi_id, 'm.ghamdi@erp.sa', 'ghamdi92@gmail.com', '+966501234567',
            '7821', 'طريق أنس بن مالك', 'الملقا', 'Riyadh', '13524', '3210', 'SA'
        ) ON CONFLICT (employee_id) DO NOTHING;

        INSERT INTO hr.bank_accounts (
            id, employee_id, bank_name_en, bank_name_ar, iban, iban_masked,
            swift_bic, is_primary, account_status
        ) VALUES (
            uuidv7(), v_emp_saudi_id, 'Al-Rajhi Bank', 'مصرف الراجحي',
            'SA4480000456123456789012', 'SA4480000456****9012', 'RJHISARI', true, 'ACTIVE'
        ) ON CONFLICT (employee_id, iban) DO NOTHING;

        -- Saudi Contract (Unspecified Term)
        INSERT INTO hr.contracts (
            id, legal_entity_id, employee_id, contract_number, contract_type,
            start_date, end_date, probation_days, notice_period_days, working_hours_per_week,
            work_location, status, notes
        ) VALUES (
            uuidv7(), p_legal_entity_id, v_emp_saudi_id, 'QIWA-2024-001', 'UNSPECIFIED_TERM',
            '2024-01-01', NULL, 90, 60, 40.00,
            'الرياض - المقر الرئيسي', 'ACTIVE', 'عقد عمل معتمد وموثق عبر منصة قوى'
        ) ON CONFLICT (legal_entity_id, contract_number) DO NOTHING
        RETURNING id INTO v_contract_1_id;

        IF v_contract_1_id IS NOT NULL THEN
            INSERT INTO hr.contract_versions (
                id, contract_id, version_number, effective_from, effective_to,
                basic_salary, housing_allowance, transport_allowance, other_allowances, gross_salary,
                change_reason, approved_by_user_id
            ) VALUES (
                uuidv7(), v_contract_1_id, 1, '2024-01-01', NULL,
                16000.0000, 4000.0000, 1500.0000, 1000.0000, 22500.0000,
                'الراتب الأساسي التعاقدي عند التعيين', v_admin_id
            ) ON CONFLICT (contract_id, version_number) DO NOTHING;
        END IF;
    END IF;

    IF v_emp_expat_id IS NOT NULL THEN
        INSERT INTO hr.employee_identifiers (
            id, employee_id, identifier_type, identifier_value, identifier_value_masked,
            issue_date, expiry_date, issue_place, is_primary, is_verified
        ) VALUES (
            uuidv7(), v_emp_expat_id, 'IQAMA', '2345678901', '2345****01',
            '2023-05-15', '2027-05-14', 'الرياض', true, true
        ) ON CONFLICT (employee_id, identifier_type, identifier_value) DO NOTHING;

        -- Expatriate Contract (Specified Term)
        INSERT INTO hr.contracts (
            id, legal_entity_id, employee_id, contract_number, contract_type,
            start_date, end_date, probation_days, notice_period_days, working_hours_per_week,
            work_location, status, notes
        ) VALUES (
            uuidv7(), p_legal_entity_id, v_emp_expat_id, 'QIWA-2023-042', 'SPECIFIED_TERM',
            '2023-06-01', '2027-05-31', 90, 60, 40.00,
            'الرياض - الإدارة المالية', 'ACTIVE', 'عقد عمل محدد المدة وموثق عبر منصة قوى'
        ) ON CONFLICT (legal_entity_id, contract_number) DO NOTHING
        RETURNING id INTO v_contract_2_id;

        IF v_contract_2_id IS NOT NULL THEN
            INSERT INTO hr.contract_versions (
                id, contract_id, version_number, effective_from, effective_to,
                basic_salary, housing_allowance, transport_allowance, other_allowances, gross_salary,
                change_reason, approved_by_user_id
            ) VALUES (
                uuidv7(), v_contract_2_id, 1, '2023-06-01', NULL,
                22000.0000, 5500.0000, 2000.0000, 0.0000, 29500.0000,
                'الراتب الأساسي المعتمد للإدارة المالية', v_admin_id
            ) ON CONFLICT (contract_id, version_number) DO NOTHING;
        END IF;
    END IF;

END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS hr.seed_default_hr_data(UUID);
DROP TABLE IF EXISTS hr.employee_documents CASCADE;
DROP TABLE IF EXISTS hr.bank_accounts CASCADE;
DROP TABLE IF EXISTS hr.contract_versions CASCADE;
DROP TABLE IF EXISTS hr.contracts CASCADE;
DROP TABLE IF EXISTS hr.qualifications CASCADE;
DROP TABLE IF EXISTS hr.emergency_contacts CASCADE;
DROP TABLE IF EXISTS hr.employee_contacts CASCADE;
DROP TABLE IF EXISTS hr.employee_identifiers CASCADE;
DROP TABLE IF EXISTS hr.employees CASCADE;
DROP TABLE IF EXISTS hr.positions CASCADE;
DROP TABLE IF EXISTS hr.jobs CASCADE;
DROP SCHEMA IF EXISTS hr CASCADE;
-- +goose StatementEnd
