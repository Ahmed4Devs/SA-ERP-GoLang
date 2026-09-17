-- +goose Up
-- Migration: GOSI Social Insurance Engine
-- Phase 22 — التأمينات الاجتماعية السعودية (GOSI)

CREATE SCHEMA IF NOT EXISTS gosi;

-- 1. Contribution Rules (Types of contributions: Annuity, OHRSI, SANED)
CREATE TABLE gosi.contribution_rules (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    code VARCHAR(30) NOT NULL,
    name_en VARCHAR(150) NOT NULL,
    name_ar VARCHAR(150) NOT NULL,
    description_en TEXT,
    description_ar TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_gosi_rule_code UNIQUE (legal_entity_id, code)
);

-- 2. Contribution Rule Versions (Effective-dated rate configurations)
CREATE TABLE gosi.contribution_rule_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    rule_id UUID NOT NULL REFERENCES gosi.contribution_rules(id) ON DELETE CASCADE,
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    version_number INT NOT NULL DEFAULT 1,
    effective_from DATE NOT NULL,
    effective_to DATE,
    nationality_scope VARCHAR(20) NOT NULL DEFAULT 'ALL'
        CHECK (nationality_scope IN ('SAUDI', 'NON_SAUDI', 'ALL')),
    employee_rate NUMERIC(8, 4) NOT NULL DEFAULT 0.0000,
    employer_rate NUMERIC(8, 4) NOT NULL DEFAULT 0.0000,
    minimum_basis NUMERIC(15, 4) NOT NULL DEFAULT 1500.0000,
    maximum_basis NUMERIC(15, 4) NOT NULL DEFAULT 45000.0000,
    calculation_config JSONB NOT NULL DEFAULT '{}',
    legal_reference VARCHAR(200),
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_gosi_rule_version UNIQUE (rule_id, version_number),
    CONSTRAINT ck_gosi_effective_dates CHECK (effective_to IS NULL OR effective_to > effective_from)
);

CREATE INDEX idx_gosi_rule_versions_active ON gosi.contribution_rule_versions (legal_entity_id, is_active, effective_from);

-- 3. Employee Insurance Profiles
CREATE TABLE gosi.employee_insurance_profiles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    employee_id UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    gosi_number VARCHAR(30),
    registration_date DATE,
    deregistration_date DATE,
    nationality_scope VARCHAR(20) NOT NULL DEFAULT 'SAUDI'
        CHECK (nationality_scope IN ('SAUDI', 'NON_SAUDI')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    contribution_basis_override NUMERIC(15, 4),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_gosi_employee_profile UNIQUE (legal_entity_id, employee_id)
);

CREATE INDEX idx_gosi_profiles_employee ON gosi.employee_insurance_profiles (employee_id);

-- 4. Contribution Calculations (Monthly per-employee calculations)
CREATE TABLE gosi.contribution_calculations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    employee_id UUID NOT NULL REFERENCES hr.employees(id),
    profile_id UUID NOT NULL REFERENCES gosi.employee_insurance_profiles(id),
    calculation_year INT NOT NULL,
    calculation_month INT NOT NULL,
    contribution_basis NUMERIC(15, 4) NOT NULL,
    total_employee_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    total_employer_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    total_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    calculation_snapshot JSONB NOT NULL DEFAULT '{}',
    export_batch_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'CALCULATED'
        CHECK (status IN ('CALCULATED', 'EXPORTED', 'SUBMITTED', 'CONFIRMED', 'REJECTED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_gosi_calc_period UNIQUE (legal_entity_id, employee_id, calculation_year, calculation_month)
);

CREATE INDEX idx_gosi_calc_period ON gosi.contribution_calculations (legal_entity_id, calculation_year, calculation_month);

-- 5. Contribution Calculation Lines (Breakdown per rule applied)
CREATE TABLE gosi.contribution_calculation_lines (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    calculation_id UUID NOT NULL REFERENCES gosi.contribution_calculations(id) ON DELETE CASCADE,
    rule_version_id UUID NOT NULL REFERENCES gosi.contribution_rule_versions(id),
    rule_code VARCHAR(30) NOT NULL,
    rule_name_en VARCHAR(150),
    rule_name_ar VARCHAR(150),
    basis_amount NUMERIC(15, 4) NOT NULL,
    employee_rate NUMERIC(8, 4) NOT NULL,
    employer_rate NUMERIC(8, 4) NOT NULL,
    employee_amount NUMERIC(15, 4) NOT NULL,
    employer_amount NUMERIC(15, 4) NOT NULL,
    formula_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_gosi_calc_lines ON gosi.contribution_calculation_lines (calculation_id);

-- 6. Export Batches
CREATE TABLE gosi.export_batches (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    batch_number VARCHAR(50) NOT NULL,
    export_year INT NOT NULL,
    export_month INT NOT NULL,
    employee_count INT NOT NULL DEFAULT 0,
    total_employee_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    total_employer_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    total_amount NUMERIC(15, 4) NOT NULL DEFAULT 0,
    file_content TEXT,
    file_format VARCHAR(20) NOT NULL DEFAULT 'GOSI_XML',
    status VARCHAR(20) NOT NULL DEFAULT 'GENERATED'
        CHECK (status IN ('GENERATED', 'VALIDATED', 'SUBMITTED', 'CONFIRMED', 'REJECTED', 'FAILED')),
    submitted_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_gosi_batch_num UNIQUE (legal_entity_id, batch_number)
);

-- 7. Integration Attempts (Tracking submissions to GOSI portal)
CREATE TABLE gosi.integration_attempts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    batch_id UUID NOT NULL REFERENCES gosi.export_batches(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL DEFAULT 1,
    adapter_type VARCHAR(30) NOT NULL DEFAULT 'FILE_EXPORT'
        CHECK (adapter_type IN ('MANUAL', 'FILE_EXPORT', 'API_INTEGRATION')),
    request_payload JSONB,
    response_payload JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'IN_PROGRESS', 'SUCCESS', 'FAILED', 'TIMEOUT')),
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    CONSTRAINT uq_gosi_attempt UNIQUE (batch_id, attempt_number)
);

-- Indexes
CREATE INDEX idx_gosi_batches_period ON gosi.export_batches (legal_entity_id, export_year, export_month);
CREATE INDEX idx_gosi_integration_batch ON gosi.integration_attempts (batch_id);

-- 8. Permissions
INSERT INTO authz.permissions (code, name_en, name_ar, module) VALUES
    ('gosi:rules:read', 'View GOSI Rules', 'عرض قواعد التأمينات', 'GOSI'),
    ('gosi:rules:write', 'Manage GOSI Rules', 'إدارة قواعد التأمينات', 'GOSI'),
    ('gosi:profiles:read', 'View Insurance Profiles', 'عرض ملفات التأمين', 'GOSI'),
    ('gosi:profiles:write', 'Manage Insurance Profiles', 'إدارة ملفات التأمين', 'GOSI'),
    ('gosi:calculations:read', 'View Contribution Calculations', 'عرض احتسابات الاشتراك', 'GOSI'),
    ('gosi:calculations:run', 'Run Monthly Calculations', 'تشغيل الاحتساب الشهري', 'GOSI'),
    ('gosi:export:run', 'Export GOSI Batches', 'تصدير دفعات التأمينات', 'GOSI'),
    ('gosi:export:submit', 'Submit to GOSI Portal', 'إرسال لبوابة التأمينات', 'GOSI')
ON CONFLICT (code) DO NOTHING;

-- 9. Seed Default GOSI Data Procedure
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE gosi.seed_default_gosi_data(p_legal_entity_id UUID)
LANGUAGE plpgsql AS $$
DECLARE
    v_annuity_rule_id UUID;
    v_ohrsi_rule_id UUID;
    v_saned_rule_id UUID;
BEGIN
    -- 1. Create Contribution Rules (Annuity, OHRSI, SANED)
    INSERT INTO gosi.contribution_rules (id, legal_entity_id, code, name_en, name_ar, description_en, description_ar)
    VALUES (uuidv7(), p_legal_entity_id, 'ANNUITY', 'Annuity / Pension', 'فرع المعاشات',
            'Retirement pension contributions under Saudi Social Insurance Law', 'اشتراكات المعاشات التقاعدية وفقاً لنظام التأمينات الاجتماعية السعودي')
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_annuity_rule_id;
    IF v_annuity_rule_id IS NULL THEN
        SELECT id INTO v_annuity_rule_id FROM gosi.contribution_rules WHERE legal_entity_id = p_legal_entity_id AND code = 'ANNUITY';
    END IF;

    INSERT INTO gosi.contribution_rules (id, legal_entity_id, code, name_en, name_ar, description_en, description_ar)
    VALUES (uuidv7(), p_legal_entity_id, 'OHRSI', 'Occupational Hazards', 'فرع الأخطار المهنية',
            'Occupational hazards and work injuries insurance', 'تأمين إصابات العمل والأخطار المهنية')
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_ohrsi_rule_id;
    IF v_ohrsi_rule_id IS NULL THEN
        SELECT id INTO v_ohrsi_rule_id FROM gosi.contribution_rules WHERE legal_entity_id = p_legal_entity_id AND code = 'OHRSI';
    END IF;

    INSERT INTO gosi.contribution_rules (id, legal_entity_id, code, name_en, name_ar, description_en, description_ar)
    VALUES (uuidv7(), p_legal_entity_id, 'SANED', 'Unemployment Insurance (SANED)', 'نظام ساند (التعطل عن العمل)',
            'Unemployment insurance under SANED system for Saudi nationals', 'تأمين التعطل عن العمل وفقاً لنظام ساند للمواطنين السعوديين')
    ON CONFLICT (legal_entity_id, code) DO NOTHING
    RETURNING id INTO v_saned_rule_id;
    IF v_saned_rule_id IS NULL THEN
        SELECT id INTO v_saned_rule_id FROM gosi.contribution_rules WHERE legal_entity_id = p_legal_entity_id AND code = 'SANED';
    END IF;

    -- 2. Seed Rule Versions (2024+ rates per GOSI regulations)
    -- Annuity: Saudi only - 9% employee + 9% employer
    INSERT INTO gosi.contribution_rule_versions (rule_id, legal_entity_id, version_number, effective_from, nationality_scope,
        employee_rate, employer_rate, minimum_basis, maximum_basis, legal_reference)
    VALUES (v_annuity_rule_id, p_legal_entity_id, 1, '2024-01-01', 'SAUDI',
        9.0000, 9.0000, 1500.0000, 45000.0000,
        'نظام التأمينات الاجتماعية - المادة 18 (فرع المعاشات)')
    ON CONFLICT (rule_id, version_number) DO NOTHING;

    -- OHRSI: All nationalities - 0% employee + 2% employer
    INSERT INTO gosi.contribution_rule_versions (rule_id, legal_entity_id, version_number, effective_from, nationality_scope,
        employee_rate, employer_rate, minimum_basis, maximum_basis, legal_reference)
    VALUES (v_ohrsi_rule_id, p_legal_entity_id, 1, '2024-01-01', 'ALL',
        0.0000, 2.0000, 1500.0000, 45000.0000,
        'نظام التأمينات الاجتماعية - المادة 18 (فرع الأخطار المهنية)')
    ON CONFLICT (rule_id, version_number) DO NOTHING;

    -- SANED: Saudi only - 0.75% employee + 0.75% employer
    INSERT INTO gosi.contribution_rule_versions (rule_id, legal_entity_id, version_number, effective_from, nationality_scope,
        employee_rate, employer_rate, minimum_basis, maximum_basis, legal_reference)
    VALUES (v_saned_rule_id, p_legal_entity_id, 1, '2024-01-01', 'SAUDI',
        0.7500, 0.7500, 1500.0000, 45000.0000,
        'نظام ساند - المادة 4 (التعطل عن العمل)')
    ON CONFLICT (rule_id, version_number) DO NOTHING;
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS gosi.seed_default_gosi_data;
DROP TABLE IF EXISTS gosi.integration_attempts CASCADE;
DROP TABLE IF EXISTS gosi.export_batches CASCADE;
DROP TABLE IF EXISTS gosi.contribution_calculation_lines CASCADE;
DROP TABLE IF EXISTS gosi.contribution_calculations CASCADE;
DROP TABLE IF EXISTS gosi.employee_insurance_profiles CASCADE;
DROP TABLE IF EXISTS gosi.contribution_rule_versions CASCADE;
DROP TABLE IF EXISTS gosi.contribution_rules CASCADE;
DROP SCHEMA IF EXISTS gosi CASCADE;
