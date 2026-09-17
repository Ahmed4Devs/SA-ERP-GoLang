package documents

import (
	"context"
	"fmt"
	"net/http"

	"erp/internal/documents/application"
	"erp/internal/documents/infrastructure/postgres"
	"erp/internal/documents/ports"
	dhttp "erp/internal/documents/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module coordinates the documents and blob storage domain.
type Module struct {
	pool      *pgxpool.Pool
	Service   ports.Service
	Repo      ports.Repository
	BlobStore storage.BlobStore
	Handler   *dhttp.DocumentHandler
}

// NewModule creates and wires a new documents module.
func NewModule(cfg *config.Config, db *database.DB, blobStore storage.BlobStore) *Module {
	repo := postgres.NewDocumentRepository(db)
	svc := application.NewDocumentService(repo, blobStore)
	handler := dhttp.NewDocumentHandler(svc)

	return &Module{
		pool:      db.Pool,
		Service:   svc,
		Repo:      repo,
		BlobStore: blobStore,
		Handler:   handler,
	}
}

// RegisterRoutes mounts all document routes onto the main Chi router under /api/v1/documents.
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/documents", m.Handler.Routes(authMiddleware))
}

const ensureSeedProcedureSQL = `
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
    IF EXISTS (SELECT 1 FROM documents.files WHERE legal_entity_id = p_legal_entity_id) THEN
        RETURN;
    END IF;

    -- A. Commercial Registration
    v_cr_file_id := uuidv7();
    v_cr_ver_id := uuidv7();

    INSERT INTO documents.files (id, legal_entity_id, title, description, category, folder_path, current_version_id, is_archived, created_at, updated_at)
    VALUES (v_cr_file_id, p_legal_entity_id, 'السجل التجاري الموحد (Commercial Registration)', 'نسخة السجل التجاري المعتمدة من وزارة التجارة', 'LEGAL', '/legal', NULL, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

    INSERT INTO documents.file_versions (id, file_id, version_number, storage_provider, storage_key, original_filename, mime_type, byte_size, sha256_hash, metadata, created_at)
    VALUES (v_cr_ver_id, v_cr_file_id, 1, 'local', 'legal/cr_certificate_2026.pdf', 'cr_certificate_2026.pdf', 'application/pdf', 1048576, 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855', '{"source": "seed", "authority": "Ministry of Commerce"}'::jsonb, CURRENT_TIMESTAMP);

    UPDATE documents.files SET current_version_id = v_cr_ver_id WHERE id = v_cr_file_id;

    -- B. VAT Certificate
    v_vat_file_id := uuidv7();
    v_vat_ver_id := uuidv7();

    INSERT INTO documents.files (id, legal_entity_id, title, description, category, folder_path, current_version_id, is_archived, created_at, updated_at)
    VALUES (v_vat_file_id, p_legal_entity_id, 'شهادة تسجيل ضريبة القيمة المضافة (VAT Certificate)', 'شهادة التسجيل في ضريبة القيمة المضافة من هيئة الزكاة والضريبة والجمارك', 'TAX', '/tax', NULL, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

    INSERT INTO documents.file_versions (id, file_id, version_number, storage_provider, storage_key, original_filename, mime_type, byte_size, sha256_hash, metadata, created_at)
    VALUES (v_vat_ver_id, v_vat_file_id, 1, 'local', 'tax/zatca_vat_cert_2026.pdf', 'zatca_vat_cert_2026.pdf', 'application/pdf', 524288, '1b4f0e9851971998e732078544c96b36c3d01cedf7caa332359d6f1ddb12879b', '{"source": "seed", "vat_number": "310123456700003"}'::jsonb, CURRENT_TIMESTAMP);

    UPDATE documents.files SET current_version_id = v_vat_ver_id WHERE id = v_vat_file_id;

    -- C. Articles of Association
    v_bylaws_file_id := uuidv7();
    v_bylaws_ver_id := uuidv7();

    INSERT INTO documents.files (id, legal_entity_id, title, description, category, folder_path, current_version_id, is_archived, created_at, updated_at)
    VALUES (v_bylaws_file_id, p_legal_entity_id, 'عقد تأسيس الشركة (Articles of Association)', 'عقد تأسيس الشركة والنظام الأساسي المعتمد', 'CONTRACT', '/corporate', NULL, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

    INSERT INTO documents.file_versions (id, file_id, version_number, storage_provider, storage_key, original_filename, mime_type, byte_size, sha256_hash, metadata, created_at)
    VALUES (v_bylaws_ver_id, v_bylaws_file_id, 1, 'local', 'corporate/articles_of_association.pdf', 'articles_of_association.pdf', 'application/pdf', 2097152, 'ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb', '{"source": "seed", "document_type": "official_bylaws"}'::jsonb, CURRENT_TIMESTAMP);

    UPDATE documents.files SET current_version_id = v_bylaws_ver_id WHERE id = v_bylaws_file_id;
END;
$$;
`

// SeedDefaultDocumentsData invokes the stored procedure to populate baseline company documents.
func (m *Module) SeedDefaultDocumentsData(ctx context.Context, legalEntityID uuid.UUID) error {
	if _, err := m.pool.Exec(ctx, ensureSeedProcedureSQL); err != nil {
		return fmt.Errorf("failed to ensure seed procedure: %w", err)
	}

	_, err := m.pool.Exec(ctx, "CALL documents.seed_default_documents_data($1)", legalEntityID)
	if err != nil {
		return fmt.Errorf("failed to seed default documents: %w", err)
	}
	return nil
}
