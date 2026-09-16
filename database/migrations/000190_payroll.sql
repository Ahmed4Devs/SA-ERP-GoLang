-- +goose Up
CREATE SCHEMA IF NOT EXISTS payroll;

-- 1. Salary Components Master
CREATE TABLE payroll.salary_components (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    name_ar VARCHAR(100) NOT NULL,
    component_type VARCHAR(20) NOT NULL CHECK (component_type IN ('EARNING', 'DEDUCTION')),
    is_taxable_or_gosi_basis BOOLEAN NOT NULL DEFAULT true,
    calculation_type VARCHAR(30) NOT NULL DEFAULT 'FIXED' CHECK (calculation_type IN ('FIXED', 'FORMULA', 'PERCENTAGE_OF_BASIC')),
    formula_expression VARCHAR(255),
    default_amount NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payroll_components_code UNIQUE (legal_entity_id, code)
);

-- 2. Salary Structures
CREATE TABLE payroll.salary_structures (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    name_ar VARCHAR(100) NOT NULL,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'SAR',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payroll_structures_code UNIQUE (legal_entity_id, code)
);

-- 3. Salary Structure Lines
CREATE TABLE payroll.salary_structure_lines (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    salary_structure_id UUID NOT NULL REFERENCES payroll.salary_structures(id) ON DELETE CASCADE,
    component_id UUID NOT NULL REFERENCES payroll.salary_components(id),
    line_number INT NOT NULL,
    rate_percentage NUMERIC(5,2) DEFAULT 0.00,
    fixed_amount NUMERIC(15,4) DEFAULT 0.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payroll_struct_line UNIQUE (salary_structure_id, component_id)
);

-- 4. Employee Salary Assignments
CREATE TABLE payroll.employee_salary_assignments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    salary_structure_id UUID NOT NULL REFERENCES payroll.salary_structures(id),
    effective_from DATE NOT NULL,
    effective_to DATE,
    basic_salary NUMERIC(15,4) NOT NULL,
    housing_allowance NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    transport_allowance NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    other_allowances NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    gross_salary NUMERIC(15,4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Payroll Periods
CREATE TABLE payroll.periods (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    fiscal_period_id UUID REFERENCES accounting.fiscal_periods(id),
    period_name_en VARCHAR(100) NOT NULL,
    period_name_ar VARCHAR(100) NOT NULL,
    period_year INT NOT NULL,
    period_month INT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    payment_due_date DATE NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payroll_period_month UNIQUE (legal_entity_id, period_year, period_month)
);

-- 6. Payroll Runs
CREATE TABLE payroll.runs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    period_id UUID NOT NULL REFERENCES payroll.periods(id),
    run_number VARCHAR(50) NOT NULL,
    run_date DATE NOT NULL,
    cutoff_date DATE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'DRAFT' CHECK (
        status IN ('DRAFT', 'CALCULATING', 'CALCULATED', 'UNDER_REVIEW', 'APPROVED', 'POSTED', 'PAID', 'CANCELLED', 'REVERSED')
    ),
    total_basic NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    total_allowances NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    total_gross NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    total_deductions NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    total_net NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    employee_count INT NOT NULL DEFAULT 0,
    journal_entry_id UUID REFERENCES accounting.journal_entries(id),
    wps_file_generated BOOLEAN NOT NULL DEFAULT false,
    wps_file_name VARCHAR(255),
    posted_at TIMESTAMPTZ,
    posted_by UUID REFERENCES iam.users(id),
    created_by UUID REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payroll_run_num UNIQUE (legal_entity_id, run_number)
);

-- 7. Payroll Run Employees
CREATE TABLE payroll.run_employees (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    run_id UUID NOT NULL REFERENCES payroll.runs(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    contract_id UUID REFERENCES hr.contracts(id),
    basic_salary NUMERIC(15,4) NOT NULL,
    housing_allowance NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    transport_allowance NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    other_allowances NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    overtime_amount NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    gross_salary NUMERIC(15,4) NOT NULL,
    gosi_employee_deduction NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    loan_deduction NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    absence_deduction NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    other_deductions NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    total_deductions NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    net_salary NUMERIC(15,4) NOT NULL,
    payment_status VARCHAR(30) NOT NULL DEFAULT 'UNPAID' CHECK (payment_status IN ('UNPAID', 'PAID', 'ON_HOLD')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_run_employee UNIQUE (run_id, employee_id)
);

-- 8. Calculation Lines (Explainability Breakdown)
CREATE TABLE payroll.calculation_lines (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    run_employee_id UUID NOT NULL REFERENCES payroll.run_employees(id) ON DELETE CASCADE,
    component_id UUID REFERENCES payroll.salary_components(id),
    line_type VARCHAR(20) NOT NULL CHECK (line_type IN ('EARNING', 'DEDUCTION')),
    code VARCHAR(50) NOT NULL,
    name_en VARCHAR(100) NOT NULL,
    name_ar VARCHAR(100) NOT NULL,
    rate NUMERIC(10,4) NOT NULL DEFAULT 1.0000,
    basis_amount NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    amount NUMERIC(15,4) NOT NULL DEFAULT 0.0000,
    formula_note VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 9. Payslips
CREATE TABLE payroll.payslips (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    run_employee_id UUID NOT NULL REFERENCES payroll.run_employees(id) ON DELETE CASCADE,
    payslip_number VARCHAR(50) NOT NULL,
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    period_id UUID NOT NULL REFERENCES payroll.periods(id),
    issue_date DATE NOT NULL,
    gross_salary NUMERIC(15,4) NOT NULL,
    total_deductions NUMERIC(15,4) NOT NULL,
    net_salary NUMERIC(15,4) NOT NULL,
    calculation_snapshot JSONB NOT NULL DEFAULT '{}',
    is_viewed_by_employee BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_payslip_num UNIQUE (legal_entity_id, payslip_number)
);

-- 10. Employee Loans
CREATE TABLE payroll.loans (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    loan_number VARCHAR(50) NOT NULL,
    principal_amount NUMERIC(15,4) NOT NULL,
    monthly_installment NUMERIC(15,4) NOT NULL,
    remaining_balance NUMERIC(15,4) NOT NULL,
    disbursement_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('PENDING', 'ACTIVE', 'PAID_OFF', 'CANCELLED')),
    reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_loan_num UNIQUE (legal_entity_id, loan_number)
);

-- 11. Loan Installments
CREATE TABLE payroll.loan_installments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    loan_id UUID NOT NULL REFERENCES payroll.loans(id) ON DELETE CASCADE,
    payroll_run_id UUID REFERENCES payroll.runs(id),
    installment_number INT NOT NULL,
    due_date DATE NOT NULL,
    amount NUMERIC(15,4) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'DEDUCTED', 'WAIVED')),
    deducted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 12. Advances
CREATE TABLE payroll.advances (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    amount NUMERIC(15,4) NOT NULL,
    request_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'DEDUCTED', 'REJECTED')),
    deducted_in_run_id UUID REFERENCES payroll.runs(id),
    reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 13. End of Service Calculations (Saudi Labor Law Art. 84 & 85)
CREATE TABLE payroll.end_of_service_calculations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    calculation_number VARCHAR(50) NOT NULL,
    hire_date DATE NOT NULL,
    termination_date DATE NOT NULL,
    termination_reason VARCHAR(50) NOT NULL CHECK (
        termination_reason IN ('RESIGNATION', 'TERMINATION_BY_EMPLOYER', 'CONTRACT_EXPIRY', 'MUTUAL_AGREEMENT', 'FORCE_MAJEURE')
    ),
    service_years NUMERIC(5,2) NOT NULL,
    last_gross_salary NUMERIC(15,4) NOT NULL,
    award_amount NUMERIC(15,4) NOT NULL,
    statutory_rate_fraction NUMERIC(5,2) NOT NULL DEFAULT 1.00,
    final_settlement_amount NUMERIC(15,4) NOT NULL,
    calculation_breakdown JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'CALCULATED' CHECK (status IN ('CALCULATED', 'APPROVED', 'PAID')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_eos_calc_num UNIQUE (legal_entity_id, calculation_number)
);

-- Indexes for performance
CREATE INDEX idx_payroll_runs_legal_entity ON payroll.runs(legal_entity_id, period_id);
CREATE INDEX idx_payroll_run_employees_run ON payroll.run_employees(run_id);
CREATE INDEX idx_payroll_calc_lines_emp ON payroll.calculation_lines(run_employee_id);
CREATE INDEX idx_payroll_payslips_emp ON payroll.payslips(employee_id, period_id);
CREATE INDEX idx_payroll_loans_emp ON payroll.loans(employee_id, status);

-- RBAC Permissions
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'payroll:runs:read', 'View Payroll Runs', 'عرض مسيرات الرواتب', 'PAYROLL', CURRENT_TIMESTAMP),
    (uuidv7(), 'payroll:runs:calculate', 'Calculate Payroll', 'احتساب مسيرات الرواتب', 'PAYROLL', CURRENT_TIMESTAMP),
    (uuidv7(), 'payroll:runs:manage', 'Manage Payroll Runs', 'إدارة مسيرات الرواتب', 'PAYROLL', CURRENT_TIMESTAMP),
    (uuidv7(), 'payroll:runs:post', 'Post Payroll to Ledger', 'اعتماد وترحيل مسيرات الرواتب', 'PAYROLL', CURRENT_TIMESTAMP),
    (uuidv7(), 'payroll:payslips:read', 'View Payslips', 'عرض قسائم الرواتب', 'PAYROLL', CURRENT_TIMESTAMP),
    (uuidv7(), 'payroll:loans:manage', 'Manage Employee Loans & Advances', 'إدارة السلف وقروض الموظفين', 'PAYROLL', CURRENT_TIMESTAMP),
    (uuidv7(), 'payroll:eos:manage', 'Manage End of Service Calculations', 'إدارة وتصفية مكافأة نهاية الخدمة', 'PAYROLL', CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- Seed Procedure
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE payroll.seed_default_payroll_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    v_basic_id UUID := uuidv7();
    v_housing_id UUID := uuidv7();
    v_transport_id UUID := uuidv7();
    v_other_id UUID := uuidv7();
    v_ot_id UUID := uuidv7();
    v_gosi_id UUID := uuidv7();
    v_loan_id UUID := uuidv7();
    v_absence_id UUID := uuidv7();
    v_struct_id UUID := uuidv7();
    v_fp_id UUID;
    v_period_id UUID := uuidv7();
BEGIN
    -- 1. Standard Components
    INSERT INTO payroll.salary_components (
        id, legal_entity_id, code, name_en, name_ar, component_type,
        is_taxable_or_gosi_basis, calculation_type, formula_expression, default_amount
    ) VALUES
        (v_basic_id, p_legal_entity_id, 'BASIC', 'Basic Salary', 'الراتب الأساسي', 'EARNING', true, 'FIXED', NULL, 0.0000),
        (v_housing_id, p_legal_entity_id, 'HOUSING', 'Housing Allowance', 'بدل سكن', 'EARNING', true, 'PERCENTAGE_OF_BASIC', 'basic * 0.25', 0.0000),
        (v_transport_id, p_legal_entity_id, 'TRANSPORT', 'Transportation Allowance', 'بدل نقل', 'EARNING', false, 'FIXED', NULL, 1000.0000),
        (v_other_id, p_legal_entity_id, 'OTHER_ALLOWANCE', 'Other Allowances', 'بدلات أخرى', 'EARNING', false, 'FIXED', NULL, 0.0000),
        (v_ot_id, p_legal_entity_id, 'OVERTIME', 'Overtime Wage (Art. 107)', 'أجر العمل الإضافي (المادة 107)', 'EARNING', false, 'FORMULA', 'hours * hourly_rate * 1.50', 0.0000),
        (v_gosi_id, p_legal_entity_id, 'GOSI_EMPLOYEE', 'GOSI Employee Share (Social Insurance)', 'حصة الموظف في التأمينات (معاشات/ساند)', 'DEDUCTION', false, 'PERCENTAGE_OF_BASIC', '(basic + housing) * 0.0975', 0.0000),
        (v_loan_id, p_legal_entity_id, 'LOAN_DEDUCTION', 'Company Loan Installment', 'استقطاع قسط سلفة الشركة', 'DEDUCTION', false, 'FIXED', NULL, 0.0000),
        (v_absence_id, p_legal_entity_id, 'ABSENCE_DEDUCTION', 'Unexcused Absence Deduction', 'استقطاع أيام الغياب غير المبرر', 'DEDUCTION', false, 'FORMULA', '(gross / 30) * absent_days', 0.0000)
    ON CONFLICT (legal_entity_id, code) DO NOTHING;

    -- 2. Standard Salary Structure
    INSERT INTO payroll.salary_structures (
        id, legal_entity_id, code, name_en, name_ar, currency_code
    ) VALUES (
        v_struct_id, p_legal_entity_id, 'STANDARD_SAUDI_STRUCT', 'Standard Saudi Employee Structure', 'الهيكل المعياري للرواتب للمنشآت السعودية', 'SAR'
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_struct_id;

    -- 3. Structure Lines
    INSERT INTO payroll.salary_structure_lines (
        salary_structure_id, component_id, line_number, rate_percentage, fixed_amount
    )
    SELECT v_struct_id, id, 1, 100.00, 0.0000 FROM payroll.salary_components WHERE legal_entity_id = p_legal_entity_id AND code = 'BASIC'
    ON CONFLICT (salary_structure_id, component_id) DO NOTHING;

    INSERT INTO payroll.salary_structure_lines (
        salary_structure_id, component_id, line_number, rate_percentage, fixed_amount
    )
    SELECT v_struct_id, id, 2, 25.00, 0.0000 FROM payroll.salary_components WHERE legal_entity_id = p_legal_entity_id AND code = 'HOUSING'
    ON CONFLICT (salary_structure_id, component_id) DO NOTHING;

    INSERT INTO payroll.salary_structure_lines (
        salary_structure_id, component_id, line_number, rate_percentage, fixed_amount
    )
    SELECT v_struct_id, id, 3, 0.00, 1000.0000 FROM payroll.salary_components WHERE legal_entity_id = p_legal_entity_id AND code = 'TRANSPORT'
    ON CONFLICT (salary_structure_id, component_id) DO NOTHING;

    -- 4. Initial Period (e.g. May 2025 / Current Fiscal Period)
    SELECT id INTO v_fp_id FROM accounting.fiscal_periods WHERE legal_entity_id = p_legal_entity_id AND period_number = 5 LIMIT 1;

    INSERT INTO payroll.periods (
        id, legal_entity_id, fiscal_period_id, period_name_en, period_name_ar,
        period_year, period_month, start_date, end_date, payment_due_date
    ) VALUES (
        v_period_id, p_legal_entity_id, v_fp_id, 'Payroll May 2025', 'مسير رواتب شهر مايو 2025م',
        2025, 5, '2025-05-01', '2025-05-31', '2025-05-27'
    ) ON CONFLICT (legal_entity_id, period_year, period_month) DO NOTHING;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS payroll.seed_default_payroll_data(UUID);
DROP TABLE IF EXISTS payroll.end_of_service_calculations CASCADE;
DROP TABLE IF EXISTS payroll.advances CASCADE;
DROP TABLE IF EXISTS payroll.loan_installments CASCADE;
DROP TABLE IF EXISTS payroll.loans CASCADE;
DROP TABLE IF EXISTS payroll.payslips CASCADE;
DROP TABLE IF EXISTS payroll.calculation_lines CASCADE;
DROP TABLE IF EXISTS payroll.run_employees CASCADE;
DROP TABLE IF EXISTS payroll.runs CASCADE;
DROP TABLE IF EXISTS payroll.periods CASCADE;
DROP TABLE IF EXISTS payroll.employee_salary_assignments CASCADE;
DROP TABLE IF EXISTS payroll.salary_structure_lines CASCADE;
DROP TABLE IF EXISTS payroll.salary_structures CASCADE;
DROP TABLE IF EXISTS payroll.salary_components CASCADE;
DROP SCHEMA IF EXISTS payroll CASCADE;
