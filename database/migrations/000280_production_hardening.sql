-- +goose Up
-- +goose StatementBegin

-- Repair reference rows on databases that already applied migration 000001.
INSERT INTO ref.currencies (code, name_en, name_ar, decimal_places, is_active) VALUES
    ('SAR', 'Saudi Riyal', 'الريال السعودي', 2, TRUE),
    ('USD', 'US Dollar', 'الدولار الأمريكي', 2, TRUE),
    ('EUR', 'Euro', 'اليورو', 2, TRUE),
    ('AED', 'UAE Dirham', 'الدرهم الإماراتي', 2, TRUE)
ON CONFLICT (code) DO UPDATE SET
    name_en = EXCLUDED.name_en,
    name_ar = EXCLUDED.name_ar,
    decimal_places = EXCLUDED.decimal_places,
    is_active = EXCLUDED.is_active;

INSERT INTO ref.countries (code, name_en, name_ar, phone_prefix, is_active) VALUES
    ('SA', 'Saudi Arabia', 'المملكة العربية السعودية', '+966', TRUE),
    ('AE', 'United Arab Emirates', 'الإمارات العربية المتحدة', '+971', TRUE),
    ('BH', 'Bahrain', 'البحرين', '+973', TRUE),
    ('KW', 'Kuwait', 'الكويت', '+965', TRUE),
    ('OM', 'Oman', 'عُمان', '+968', TRUE),
    ('QA', 'Qatar', 'قطر', '+974', TRUE)
ON CONFLICT (code) DO UPDATE SET
    name_en = EXCLUDED.name_en,
    name_ar = EXCLUDED.name_ar,
    phone_prefix = EXCLUDED.phone_prefix,
    is_active = EXCLUDED.is_active;

-- Repair the already-installed treasury seed function without replaying a
-- historical migration. pg_get_functiondef returns normalized source text.
DO $$
DECLARE
    function_definition TEXT;
BEGIN
    SELECT pg_get_functiondef('treasury.seed_default_treasury_data(uuid)'::regprocedure)
    INTO function_definition;

    function_definition := replace(
        function_definition,
        'SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_rajhi_bank_id',
        'SET bank_name_ar = EXCLUDED.bank_name_ar
    RETURNING id INTO v_rajhi_bank_id'
    );

    EXECUTE function_definition;
END;
$$;

-- Credentials shipped by an old demo seed must never remain active after an
-- upgrade. Real CSIDs are created exclusively by the FATOORA onboarding flow.
DELETE FROM zatca.csids
WHERE secret = 'mock-zatca-api-secret-2026'
   OR private_key_pem LIKE '%IBmJ3w0F8v1qg3k1...%';

-- Replace the old demo procedure so an upgraded database cannot recreate the
-- known mock production credential.
CREATE OR REPLACE PROCEDURE zatca.seed_default_zatca_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    v_egs_id UUID;
    v_vat_number VARCHAR(20);
    v_org_name VARCHAR(255);
    v_branch_id UUID;
BEGIN
    IF EXISTS (SELECT 1 FROM zatca.egs_units WHERE legal_entity_id = p_legal_entity_id) THEN
        RETURN;
    END IF;

    SELECT name_ar INTO v_org_name
    FROM organization.legal_entities
    WHERE id = p_legal_entity_id;

    SELECT registration_number INTO v_vat_number
    FROM organization.tax_registrations
    WHERE legal_entity_id = p_legal_entity_id AND tax_type = 'VAT'
    LIMIT 1;

    IF v_vat_number IS NULL OR length(v_vat_number) < 15 THEN
        v_vat_number := '300000000000003';
    END IF;
    IF v_org_name IS NULL THEN
        v_org_name := 'شركة الأفق المتقدمة المحدودة';
    END IF;

    SELECT id INTO v_branch_id
    FROM organization.branches
    WHERE legal_entity_id = p_legal_entity_id
    ORDER BY created_at ASC
    LIMIT 1;

    INSERT INTO zatca.egs_units (
        id, legal_entity_id, branch_id, uuid, custom_id, model, serial_number,
        solution_name, vat_number, organization_name, organization_unit_name,
        country_code, invoice_type_mask, status, created_at, updated_at
    ) VALUES (
        uuidv7(), p_legal_entity_id, v_branch_id, uuidv7(),
        'EGS-RYD-HQ01', 'Saudi-ERP-v1', 'SN-ERP-2026-HQ01',
        'Saudi Enterprise ERP Platform', v_vat_number, v_org_name, 'Headquarters',
        'SA', '1100', 'PENDING_OTP', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    )
    ON CONFLICT (legal_entity_id, custom_id) DO UPDATE SET
        organization_name = EXCLUDED.organization_name,
        vat_number = EXCLUDED.vat_number,
        updated_at = CURRENT_TIMESTAMP
    RETURNING id INTO v_egs_id;

    INSERT INTO zatca.invoice_sequences (
        id, egs_unit_id, invoice_type, last_icv, last_invoice_hash, updated_at
    ) VALUES
        (uuidv7(), v_egs_id, 'STANDARD', 0, 'NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz', CURRENT_TIMESTAMP),
        (uuidv7(), v_egs_id, 'SIMPLIFIED', 0, 'NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz', CURRENT_TIMESTAMP)
    ON CONFLICT (egs_unit_id, invoice_type) DO NOTHING;

    INSERT INTO zatca.configuration_versions (
        id, legal_entity_id, version_tag, environment,
        clearance_api_url, reporting_api_url, compliance_api_url,
        is_active, effective_from, created_at
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'v2.1-2026', 'SIMULATION',
        'https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/invoices/clearance/single',
        'https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/invoices/reporting/single',
        'https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/compliance',
        true, CURRENT_DATE, CURRENT_TIMESTAMP
    );
END;
$$;

-- Complete the permission catalogue used by the route-level authorization
-- guard. Existing specialized permissions remain available.
INSERT INTO authz.permissions (code, name_ar, name_en, module)
SELECT module_code || suffix,
       CASE suffix WHEN '.view' THEN 'عرض ' || module_name_ar ELSE 'إدارة ' || module_name_ar END,
       CASE suffix WHEN '.view' THEN 'View ' || module_name_en ELSE 'Manage ' || module_name_en END,
       upper(module_code)
FROM (VALUES
    ('org', 'المنظمة', 'Organization'),
    ('authz', 'الصلاحيات', 'Authorization'),
    ('audit', 'سجل التدقيق', 'Audit'),
    ('workflow', 'سير العمل', 'Workflow'),
    ('accounting', 'المحاسبة', 'Accounting'),
    ('tax', 'الضرائب', 'Tax'),
    ('products', 'المنتجات', 'Products'),
    ('inventory', 'المخزون', 'Inventory'),
    ('receivables', 'المدينون', 'Receivables'),
    ('sales', 'المبيعات', 'Sales'),
    ('payables', 'الدائنون', 'Payables'),
    ('purchasing', 'المشتريات', 'Purchasing'),
    ('treasury', 'الخزينة', 'Treasury'),
    ('banking', 'البنوك', 'Banking'),
    ('hr', 'الموارد البشرية', 'Human Resources'),
    ('attendance', 'الحضور', 'Attendance'),
    ('leave', 'الإجازات', 'Leave'),
    ('payroll', 'الرواتب', 'Payroll'),
    ('gosi', 'التأمينات', 'GOSI'),
    ('fixedassets', 'الأصول الثابتة', 'Fixed Assets'),
    ('budgeting', 'الميزانيات', 'Budgeting'),
    ('zatca', 'الفوترة الإلكترونية', 'ZATCA'),
    ('documents', 'المستندات', 'Documents'),
    ('notifications', 'الإشعارات', 'Notifications'),
    ('reporting', 'التقارير', 'Reporting'),
    ('dashboard', 'لوحة المعلومات', 'Dashboard')
) AS modules(module_code, module_name_ar, module_name_en)
CROSS JOIN (VALUES ('.view'), ('.manage')) AS actions(suffix)
ON CONFLICT (code) DO UPDATE SET
    name_ar = EXCLUDED.name_ar,
    name_en = EXCLUDED.name_en,
    module = EXCLUDED.module;

INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- This hardening migration deliberately has no destructive down migration:
-- removing reference data or reintroducing leaked credentials is unsafe.
SELECT 1;
