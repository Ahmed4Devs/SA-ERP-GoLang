-- +goose Up
-- Documents & Blob Storage Management Schema (Phase 27)
-- Compliant with Master Specification Section 40.25, Section 33, and Section 98

CREATE SCHEMA IF NOT EXISTS documents;

-- 1. Documents Files Master (ملفات ووثائق المنشأة)
CREATE TABLE IF NOT EXISTS documents.files (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL DEFAULT 'GENERAL' CHECK (category IN ('GENERAL', 'CONTRACT', 'INVOICE', 'RECEIPT', 'LEGAL', 'IDENTITY', 'TAX', 'ATTACHMENT', 'OTHER')),
    folder_path VARCHAR(255) NOT NULL DEFAULT '/',
    current_version_id UUID,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_by_user_id UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_documents_files_le ON documents.files(legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_documents_files_cat ON documents.files(category);
CREATE INDEX IF NOT EXISTS idx_documents_files_archived ON documents.files(is_archived);
CREATE INDEX IF NOT EXISTS idx_documents_files_folder ON documents.files(folder_path);

-- 2. File Versions (إصدارات الملف والبيانات الثنائية)
CREATE TABLE IF NOT EXISTS documents.file_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    file_id UUID NOT NULL REFERENCES documents.files(id) ON DELETE CASCADE,
    version_number INT NOT NULL DEFAULT 1,
    storage_provider VARCHAR(50) NOT NULL DEFAULT 'local',
    storage_key TEXT NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL DEFAULT 'application/octet-stream',
    byte_size BIGINT NOT NULL DEFAULT 0,
    sha256_hash VARCHAR(64) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    uploaded_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_documents_file_version UNIQUE (file_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_documents_versions_file ON documents.file_versions(file_id);
CREATE INDEX IF NOT EXISTS idx_documents_versions_hash ON documents.file_versions(sha256_hash);

-- Foreign Key: current_version_id back to file_versions
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_documents_files_current_version'
    ) THEN
        ALTER TABLE documents.files
        ADD CONSTRAINT fk_documents_files_current_version
        FOREIGN KEY (current_version_id)
        REFERENCES documents.file_versions(id)
        ON DELETE SET NULL;
    END IF;
END $$;
-- +goose StatementEnd

-- 3. Entity Attachments (ربط المستندات بكيانات الـ ERP)
CREATE TABLE IF NOT EXISTS documents.entity_attachments (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    file_id UUID NOT NULL REFERENCES documents.files(id) ON DELETE CASCADE,
    entity_type VARCHAR(100) NOT NULL, -- 'sales_invoice', 'purchase_order', 'employee', 'customer', 'supplier', 'fixed_asset', 'journal_entry', 'custom'
    entity_id UUID NOT NULL,
    attached_by UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_documents_entity_attachment UNIQUE (file_id, entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_documents_attachments_entity ON documents.entity_attachments(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_documents_attachments_file ON documents.entity_attachments(file_id);

-- 4. Permissions Seeding
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'documents.file.read',    'Read Documents',           'استعراض الوثائق والمستندات',    'DOCUMENTS', CURRENT_TIMESTAMP),
    (uuidv7(), 'documents.file.upload',  'Upload Documents',         'رفع المستندات وإصداراتها',       'DOCUMENTS', CURRENT_TIMESTAMP),
    (uuidv7(), 'documents.file.delete',  'Delete Documents',         'حذف المستندات نهائياً',          'DOCUMENTS', CURRENT_TIMESTAMP),
    (uuidv7(), 'documents.file.archive', 'Archive Documents',        'أرشفة المستندات',               'DOCUMENTS', CURRENT_TIMESTAMP),
    (uuidv7(), 'documents.file.attach',  'Attach to Entities',       'ربط المستندات بالعمليات والكيانات', 'DOCUMENTS', CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- 5. Seed Procedure for Default Organization Documents
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE documents.seed_default_documents_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    v_cr_file_id UUID;
    v_cr_ver_id UUID;
    v_vat_file_id UUID;
    v_vat_ver_id UUID;
    v_bylaws_file_id UUID;
    v_bylaws_ver_id UUID;
BEGIN
    -- Check if documents already exist for this legal entity
    IF EXISTS (SELECT 1 FROM documents.files WHERE legal_entity_id = p_legal_entity_id) THEN
        RETURN;
    END IF;

    -- A. Commercial Registration (السجل التجاري)
    v_cr_file_id := uuidv7();
    v_cr_ver_id := uuidv7();

    INSERT INTO documents.files (id, legal_entity_id, title, description, category, folder_path, current_version_id, is_archived, created_at, updated_at)
    VALUES (v_cr_file_id, p_legal_entity_id, 'السجل التجاري الموحد (Commercial Registration)', 'نسخة السجل التجاري المعتمدة من وزارة التجارة', 'LEGAL', '/legal', NULL, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

    INSERT INTO documents.file_versions (id, file_id, version_number, storage_provider, storage_key, original_filename, mime_type, byte_size, sha256_hash, metadata, created_at)
    VALUES (v_cr_ver_id, v_cr_file_id, 1, 'local', 'legal/cr_certificate_2026.pdf', 'cr_certificate_2026.pdf', 'application/pdf', 1048576, 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', '{"source": "seed", "authority": "Ministry of Commerce"}'::jsonb, CURRENT_TIMESTAMP);

    UPDATE documents.files SET current_version_id = v_cr_ver_id WHERE id = v_cr_file_id;

    -- B. VAT Registration Certificate (شهادة تسجيل ضريبة القيمة المضافة)
    v_vat_file_id := uuidv7();
    v_vat_ver_id := uuidv7();

    INSERT INTO documents.files (id, legal_entity_id, title, description, category, folder_path, current_version_id, is_archived, created_at, updated_at)
    VALUES (v_vat_file_id, p_legal_entity_id, 'شهادة تسجيل ضريبة القيمة المضافة (VAT Certificate)', 'شهادة التسجيل في ضريبة القيمة المضافة من هيئة الزكاة والضريبة والجمارك', 'TAX', '/tax', NULL, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

    INSERT INTO documents.file_versions (id, file_id, version_number, storage_provider, storage_key, original_filename, mime_type, byte_size, sha256_hash, metadata, created_at)
    VALUES (v_vat_ver_id, v_vat_file_id, 1, 'local', 'tax/zatca_vat_cert_2026.pdf', 'zatca_vat_cert_2026.pdf', 'application/pdf', 524288, '1b4f0e9851971998e732078544c96b36c3d01cedf7caa332359d6f1ddb12879b', '{"source": "seed", "vat_number": "310123456700003"}'::jsonb, CURRENT_TIMESTAMP);

    UPDATE documents.files SET current_version_id = v_vat_ver_id WHERE id = v_vat_file_id;

    -- C. Articles of Association (عقد التأسيس ولوائح الشركة)
    v_bylaws_file_id := uuidv7();
    v_bylaws_ver_id := uuidv7();

    INSERT INTO documents.files (id, legal_entity_id, title, description, category, folder_path, current_version_id, is_archived, created_at, updated_at)
    VALUES (v_bylaws_file_id, p_legal_entity_id, 'عقد تأسيس الشركة (Articles of Association)', 'عقد تأسيس الشركة والنظام الأساسي المعتمد', 'CONTRACT', '/corporate', NULL, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

    INSERT INTO documents.file_versions (id, file_id, version_number, storage_provider, storage_key, original_filename, mime_type, byte_size, sha256_hash, metadata, created_at)
    VALUES (v_bylaws_ver_id, v_bylaws_file_id, 1, 'local', 'corporate/articles_of_association.pdf', 'articles_of_association.pdf', 'application/pdf', 2097152, 'ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb', '{"source": "seed", "document_type": "official_bylaws"}'::jsonb, CURRENT_TIMESTAMP);

    UPDATE documents.files SET current_version_id = v_bylaws_ver_id WHERE id = v_bylaws_file_id;

END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP PROCEDURE IF EXISTS documents.seed_default_documents_data(UUID);
DROP TABLE IF EXISTS documents.entity_attachments CASCADE;
ALTER TABLE documents.files DROP CONSTRAINT IF EXISTS fk_documents_files_current_version;
DROP TABLE IF EXISTS documents.file_versions CASCADE;
DROP TABLE IF EXISTS documents.files CASCADE;
DROP SCHEMA IF EXISTS documents CASCADE;
-- +goose StatementEnd
