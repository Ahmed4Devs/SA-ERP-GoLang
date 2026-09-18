-- +goose Up
-- ZATCA Electronic Invoicing Core Schema (Phase 25)
-- Compliant with Master Specification Section 40.26 and Section 96 (ZATCA Phase 2 FATOORA Compliance)

CREATE SCHEMA IF NOT EXISTS zatca;

-- 1. EGS Units (Electronic Generation Solution Devices/Systems)
CREATE TABLE IF NOT EXISTS zatca.egs_units (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES organization.branches(id) ON DELETE SET NULL,
    uuid UUID NOT NULL DEFAULT uuidv7(),
    custom_id VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL DEFAULT 'Saudi-ERP-v1',
    serial_number VARCHAR(100) NOT NULL,
    solution_name VARCHAR(100) NOT NULL DEFAULT 'Saudi Enterprise ERP',
    vat_number VARCHAR(20) NOT NULL,
    organization_name VARCHAR(255) NOT NULL,
    organization_unit_name VARCHAR(255) NOT NULL,
    country_code VARCHAR(2) NOT NULL DEFAULT 'SA',
    invoice_type_mask VARCHAR(10) NOT NULL DEFAULT '1100', -- 1100: Standard + Simplified, 0100: Simplified, 1000: Standard
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',          -- 'PENDING_OTP', 'COMPLIANCE_ACTIVE', 'ACTIVE', 'REVOKED', 'INACTIVE'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_zatca_egs_custom_id UNIQUE (legal_entity_id, custom_id),
    CONSTRAINT uq_zatca_egs_serial UNIQUE (legal_entity_id, serial_number)
);

CREATE INDEX IF NOT EXISTS idx_zatca_egs_le ON zatca.egs_units(legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_zatca_egs_status ON zatca.egs_units(status);

-- 2. CSIDs (Cryptographic Stamp Identifiers - Compliance & Production X.509 Certificates)
CREATE TABLE IF NOT EXISTS zatca.csids (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    egs_unit_id UUID NOT NULL REFERENCES zatca.egs_units(id) ON DELETE CASCADE,
    csid_type VARCHAR(50) NOT NULL DEFAULT 'COMPLIANCE', -- 'COMPLIANCE', 'PRODUCTION'
    binary_security_token TEXT NOT NULL,                -- Base64 encoded X.509 certificate
    secret TEXT NOT NULL,                               -- API basic auth secret
    private_key_pem TEXT NOT NULL,                      -- ECDSA secp256k1 private key in PEM format
    public_key_pem TEXT NOT NULL,                       -- ECDSA public key in PEM format
    certificate_pem TEXT,                               -- Formatted PEM certificate
    request_id VARCHAR(100),
    compliance_request_id VARCHAR(100),
    expires_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',       -- 'ACTIVE', 'EXPIRED', 'REVOKED'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_csids_egs ON zatca.csids(egs_unit_id);
CREATE INDEX IF NOT EXISTS idx_zatca_csids_type ON zatca.csids(csid_type);

-- 3. Invoice Sequences (Invoice Counter Value & Previous Invoice Hash chaining)
CREATE TABLE IF NOT EXISTS zatca.invoice_sequences (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    egs_unit_id UUID NOT NULL REFERENCES zatca.egs_units(id) ON DELETE CASCADE,
    invoice_type VARCHAR(50) NOT NULL, -- 'STANDARD', 'SIMPLIFIED'
    last_icv BIGINT NOT NULL DEFAULT 0,
    last_invoice_hash VARCHAR(255) NOT NULL DEFAULT 'NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_zatca_inv_seq UNIQUE (egs_unit_id, invoice_type)
);

CREATE INDEX IF NOT EXISTS idx_zatca_seq_egs ON zatca.invoice_sequences(egs_unit_id);

-- 4. Documents (Phase 2 Electronic Invoices & Credit/Debit Notes)
CREATE TABLE IF NOT EXISTS zatca.documents (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    egs_unit_id UUID NOT NULL REFERENCES zatca.egs_units(id) ON DELETE RESTRICT,
    sales_invoice_id UUID REFERENCES sales.invoices(id) ON DELETE SET NULL,
    invoice_uuid UUID NOT NULL,
    invoice_number VARCHAR(100) NOT NULL,
    invoice_type VARCHAR(50) NOT NULL,                                       -- 'STANDARD', 'SIMPLIFIED'
    invoice_subtype VARCHAR(50) NOT NULL DEFAULT 'TAX_INVOICE',              -- 'TAX_INVOICE', 'CREDIT_NOTE', 'DEBIT_NOTE'
    submission_mode VARCHAR(50) NOT NULL DEFAULT 'CLEARANCE',                -- 'CLEARANCE', 'REPORTING'
    icv BIGINT NOT NULL,
    pih VARCHAR(255) NOT NULL,
    issue_date DATE NOT NULL,
    issue_time TIME NOT NULL,
    ubl_xml TEXT NOT NULL,
    signed_xml TEXT,
    invoice_hash VARCHAR(255) NOT NULL,
    digital_signature TEXT,
    public_key_hashing TEXT,
    qr_code_data TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'GENERATED', -- 'NOT_GENERATED', 'GENERATED', 'VALIDATED', 'SIGNED', 'QUEUED', 'SUBMITTING', 'CLEARED', 'REPORTED', 'WARNING', 'REJECTED', 'FAILED'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_zatca_documents_uuid UNIQUE (invoice_uuid),
    CONSTRAINT uq_zatca_documents_no UNIQUE (legal_entity_id, invoice_number)
);

CREATE INDEX IF NOT EXISTS idx_zatca_docs_le ON zatca.documents(legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_zatca_docs_egs ON zatca.documents(egs_unit_id);
CREATE INDEX IF NOT EXISTS idx_zatca_docs_status ON zatca.documents(status);
CREATE INDEX IF NOT EXISTS idx_zatca_docs_date ON zatca.documents(issue_date);

-- 5. Document Hashes
CREATE TABLE IF NOT EXISTS zatca.document_hashes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    document_id UUID NOT NULL REFERENCES zatca.documents(id) ON DELETE CASCADE,
    hash_type VARCHAR(50) NOT NULL DEFAULT 'SHA256',
    hash_value VARCHAR(255) NOT NULL,
    canonical_xml_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_hashes_doc ON zatca.document_hashes(document_id);

-- 6. Submissions & Submission Attempts
CREATE TABLE IF NOT EXISTS zatca.submissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    document_id UUID NOT NULL REFERENCES zatca.documents(id) ON DELETE CASCADE,
    submission_mode VARCHAR(50) NOT NULL, -- 'CLEARANCE', 'REPORTING'
    zatca_status VARCHAR(50) NOT NULL,    -- 'CLEARED', 'REPORTED', 'WARNING', 'REJECTED'
    zatca_request_id VARCHAR(100),
    clearance_status VARCHAR(50),
    reporting_status VARCHAR(50),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_sub_doc ON zatca.submissions(document_id);

CREATE TABLE IF NOT EXISTS zatca.submission_attempts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    submission_id UUID NOT NULL REFERENCES zatca.submissions(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL DEFAULT 1,
    http_status INT NOT NULL,
    response_payload JSONB,
    error_message TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_sub_attempts ON zatca.submission_attempts(submission_id);

-- 7. Validation Results & Messages (BR-KSA rules)
CREATE TABLE IF NOT EXISTS zatca.validation_results (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    document_id UUID NOT NULL REFERENCES zatca.documents(id) ON DELETE CASCADE,
    overall_status VARCHAR(50) NOT NULL DEFAULT 'PASSED', -- 'PASSED', 'WARNING', 'FAILED'
    rule_version VARCHAR(50) NOT NULL DEFAULT 'ZATCA-2024-v2',
    validated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_val_res_doc ON zatca.validation_results(document_id);

CREATE TABLE IF NOT EXISTS zatca.validation_messages (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    validation_result_id UUID NOT NULL REFERENCES zatca.validation_results(id) ON DELETE CASCADE,
    message_type VARCHAR(20) NOT NULL DEFAULT 'INFO', -- 'ERROR', 'WARNING', 'INFO'
    code VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'BUSINESS_RULES', -- 'XML_STRUCTURE', 'SCHEMA', 'BUSINESS_RULES', 'CRYPTOGRAPHIC_STAMP'
    message_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_val_msgs ON zatca.validation_messages(validation_result_id);

-- 8. Configuration Versions (Regulatory & API endpoints)
CREATE TABLE IF NOT EXISTS zatca.configuration_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    version_tag VARCHAR(50) NOT NULL DEFAULT 'v2.1-2026',
    environment VARCHAR(50) NOT NULL DEFAULT 'SIMULATION', -- 'SANDBOX', 'SIMULATION', 'PRODUCTION'
    clearance_api_url TEXT NOT NULL DEFAULT 'https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/invoices/clearance/single',
    reporting_api_url TEXT NOT NULL DEFAULT 'https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/invoices/reporting/single',
    compliance_api_url TEXT NOT NULL DEFAULT 'https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation/compliance',
    is_active BOOLEAN NOT NULL DEFAULT true,
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zatca_config_le ON zatca.configuration_versions(legal_entity_id);

-- 9. Seed Procedure for Default ZATCA Configuration and EGS Unit
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE zatca.seed_default_zatca_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    v_egs_id UUID;
    v_vat_number VARCHAR(20);
    v_org_name VARCHAR(255);
    v_branch_id UUID;
BEGIN
    -- Check if EGS already exists for this legal entity
    IF EXISTS (SELECT 1 FROM zatca.egs_units WHERE legal_entity_id = p_legal_entity_id) THEN
        RETURN;
    END IF;

    -- Query VAT number & org name from legal entity
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

    -- Query default branch
    SELECT id INTO v_branch_id
    FROM organization.branches
    WHERE legal_entity_id = p_legal_entity_id
    ORDER BY created_at ASC
    LIMIT 1;

    -- 1. Create Default EGS Unit
    INSERT INTO zatca.egs_units (
        id, legal_entity_id, branch_id, uuid, custom_id, model, serial_number,
        solution_name, vat_number, organization_name, organization_unit_name,
        country_code, invoice_type_mask, status, created_at, updated_at
    ) VALUES (
        uuidv7(), p_legal_entity_id, v_branch_id, uuidv7(),
        'EGS-RYD-HQ01', 'Saudi-ERP-v1', 'SN-ERP-2026-HQ01',
        'Saudi Enterprise ERP Platform', v_vat_number, v_org_name, 'Headquarters',
        'SA', '1100', 'ACTIVE', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ) RETURNING id INTO v_egs_id;

    -- 2. Create Initial Invoice Sequences for Standard & Simplified
    INSERT INTO zatca.invoice_sequences (
        id, egs_unit_id, invoice_type, last_icv, last_invoice_hash, updated_at
    ) VALUES 
    (uuidv7(), v_egs_id, 'STANDARD', 0, 'NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz', CURRENT_TIMESTAMP),
    (uuidv7(), v_egs_id, 'SIMPLIFIED', 0, 'NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz', CURRENT_TIMESTAMP);

    -- 3. Create Default Configuration Version
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
-- +goose StatementEnd

-- +goose Down
DROP PROCEDURE IF EXISTS zatca.seed_default_zatca_data(UUID);
DROP TABLE IF EXISTS zatca.validation_messages CASCADE;
DROP TABLE IF EXISTS zatca.validation_results CASCADE;
DROP TABLE IF EXISTS zatca.submission_attempts CASCADE;
DROP TABLE IF EXISTS zatca.submissions CASCADE;
DROP TABLE IF EXISTS zatca.document_hashes CASCADE;
DROP TABLE IF EXISTS zatca.documents CASCADE;
DROP TABLE IF EXISTS zatca.invoice_sequences CASCADE;
DROP TABLE IF EXISTS zatca.csids CASCADE;
DROP TABLE IF EXISTS zatca.egs_units CASCADE;
DROP TABLE IF EXISTS zatca.configuration_versions CASCADE;
DROP SCHEMA IF EXISTS zatca CASCADE;
