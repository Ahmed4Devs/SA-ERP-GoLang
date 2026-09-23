package documents_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"erp/internal/documents"
	"erp/internal/documents/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/storage"
	"erp/internal/platform/transaction"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func setupDocumentsIntegrationTest(t *testing.T) (*documents.Module, *database.DB, storage.BlobStore, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("integration database setup failed: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)

	// 1. Seed Organization
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Admin User
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	} else {
		newAdmin, err := iamMod.AuthService.SeedAdmin(ctx, "doc_admin", "doc_admin@erp.sa", "مدير الوثائق", "Admin@ERP123!")
		if err == nil && newAdmin != nil {
			adminID = newAdmin.ID
		}
	}

	// 3. Setup temporary local blob store
	tempDir, err := os.MkdirTemp("", "erp_doc_test_*")
	if err != nil {
		t.Fatalf("failed to create temp storage dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	blobStore, err := storage.NewLocalFilesystemBlobStore(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize local blob store: %v", err)
	}

	docMod := documents.NewModule(cfg, db, blobStore)

	return docMod, db, blobStore, le.ID, adminID, ctx
}

func TestDocuments_UploadAndDownload(t *testing.T) {
	docMod, db, _, leID, userID, ctx := setupDocumentsIntegrationTest(t)
	defer db.Close()

	payload := []byte("%PDF-1.7 Test PDF Document for Saudi ERP 2026 - ZATCA & GOSI Compliance")
	params := domain.UploadFileParams{
		LegalEntityID:    leID,
		Title:            "شهادة التأهيل الفني 2026",
		Description:      "وثيقة التأهيل الفني والامتثال للأنظمة السعودية",
		Category:         domain.CategoryLegal,
		FolderPath:       "/compliance",
		OriginalFilename: "technical_qualification_2026.pdf",
		ContentType:      "application/pdf",
		Content:          bytes.NewReader(payload),
		UserID:           &userID,
	}

	// 1. Upload file
	uploaded, err := docMod.Service.UploadFile(ctx, params)
	if err != nil {
		t.Fatalf("failed to upload document: %v", err)
	}

	if uploaded.ID == uuid.Nil {
		t.Fatal("expected non-nil uploaded file ID")
	}
	if uploaded.Title != params.Title {
		t.Errorf("expected title %q, got %q", params.Title, uploaded.Title)
	}
	if uploaded.Category != domain.CategoryLegal {
		t.Errorf("expected category LEGAL, got %s", uploaded.Category)
	}
	if uploaded.CurrentVersion == nil {
		t.Fatal("expected non-nil CurrentVersion")
	}
	if uploaded.CurrentVersion.VersionNumber != 1 {
		t.Errorf("expected version number 1, got %d", uploaded.CurrentVersion.VersionNumber)
	}
	if uploaded.CurrentVersion.ByteSize != int64(len(payload)) {
		t.Errorf("expected byte size %d, got %d", len(payload), uploaded.CurrentVersion.ByteSize)
	}
	if uploaded.CurrentVersion.SHA256Hash == "" {
		t.Error("expected non-empty SHA-256 hash")
	}

	// 2. Download file and verify content
	stream, ver, err := docMod.Service.DownloadFile(ctx, leID, uploaded.ID, nil)
	if err != nil {
		t.Fatalf("failed to download document: %v", err)
	}

	downloadedBytes, err := io.ReadAll(stream)
	_ = stream.Close()
	if err != nil {
		t.Fatalf("failed to read downloaded stream: %v", err)
	}
	if !bytes.Equal(downloadedBytes, payload) {
		t.Errorf("downloaded content mismatch: expected %d bytes, got %d bytes", len(payload), len(downloadedBytes))
	}
	if ver.ID != uploaded.CurrentVersion.ID {
		t.Errorf("expected version ID %s, got %s", uploaded.CurrentVersion.ID, ver.ID)
	}

	// 3. Test HTTP Download Handler
	r := chi.NewRouter()
	docMod.RegisterRoutes(r, nil)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/documents/%s/download?legal_entity_id=%s", uploaded.ID, leID), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on download, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %s", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "technical_qualification_2026.pdf") {
		t.Errorf("expected Content-Disposition with filename, got: %s", rec.Header().Get("Content-Disposition"))
	}
	if !bytes.Equal(rec.Body.Bytes(), payload) {
		t.Errorf("HTTP response body does not match payload")
	}
}

func TestDocuments_Versioning(t *testing.T) {
	docMod, db, _, leID, userID, ctx := setupDocumentsIntegrationTest(t)
	defer db.Close()

	v1Payload := []byte("Initial Document Version 1 Content - 2026-01-01")
	v2Payload := []byte("Revised Document Version 2 Content - 2026-02-15 With Updated Signatures")

	// 1. Create Version 1
	file, err := docMod.Service.UploadFile(ctx, domain.UploadFileParams{
		LegalEntityID:    leID,
		Title:            "عقد توريد رئيسي (Master Supply Agreement)",
		Category:         domain.CategoryContract,
		FolderPath:       "/contracts/suppliers",
		OriginalFilename: "supply_agreement_v1.pdf",
		ContentType:      "application/pdf",
		Content:          bytes.NewReader(v1Payload),
		UserID:           &userID,
	})
	if err != nil {
		t.Fatalf("failed to create file v1: %v", err)
	}
	v1ID := file.CurrentVersion.ID

	// 2. Upload Version 2
	v2, err := docMod.Service.UploadVersion(ctx, domain.UploadVersionParams{
		FileID:           file.ID,
		LegalEntityID:    leID,
		OriginalFilename: "supply_agreement_v2_amended.pdf",
		ContentType:      "application/pdf",
		Content:          bytes.NewReader(v2Payload),
		UserID:           &userID,
	})
	if err != nil {
		t.Fatalf("failed to upload version 2: %v", err)
	}

	if v2.VersionNumber != 2 {
		t.Fatalf("expected version number 2, got %d", v2.VersionNumber)
	}

	// 3. Inspect updated File record
	refreshed, err := docMod.Service.GetFile(ctx, leID, file.ID)
	if err != nil {
		t.Fatalf("failed to get refreshed file: %v", err)
	}
	if refreshed.CurrentVersion.ID != v2.ID {
		t.Errorf("expected current version ID to be v2 (%s), got %s", v2.ID, refreshed.CurrentVersion.ID)
	}
	if refreshed.CurrentVersion.VersionNumber != 2 {
		t.Errorf("expected current version number 2, got %d", refreshed.CurrentVersion.VersionNumber)
	}

	// 4. Download specific v1
	stream1, ver1, err := docMod.Service.DownloadFile(ctx, leID, file.ID, &v1ID)
	if err != nil {
		t.Fatalf("failed to download v1: %v", err)
	}
	b1, _ := io.ReadAll(stream1)
	_ = stream1.Close()
	if !bytes.Equal(b1, v1Payload) {
		t.Errorf("v1 payload mismatch")
	}
	if ver1.VersionNumber != 1 {
		t.Errorf("expected ver1 number 1, got %d", ver1.VersionNumber)
	}

	// 5. Download latest (v2 default)
	stream2, ver2, err := docMod.Service.DownloadFile(ctx, leID, file.ID, nil)
	if err != nil {
		t.Fatalf("failed to download latest: %v", err)
	}
	b2, _ := io.ReadAll(stream2)
	_ = stream2.Close()
	if !bytes.Equal(b2, v2Payload) {
		t.Errorf("v2 payload mismatch")
	}
	if ver2.VersionNumber != 2 {
		t.Errorf("expected ver2 number 2, got %d", ver2.VersionNumber)
	}
}

func TestDocuments_EntityAttachments(t *testing.T) {
	docMod, db, _, leID, userID, ctx := setupDocumentsIntegrationTest(t)
	defer db.Close()

	salesInvoiceID := uuid.New()
	employeeID := uuid.New()

	// 1. Upload document and link directly to sales invoice
	file, err := docMod.Service.UploadFile(ctx, domain.UploadFileParams{
		LegalEntityID:    leID,
		Title:            "فاتورة ضريبية موقعة ومختومة",
		Category:         domain.CategoryInvoice,
		OriginalFilename: "signed_tax_invoice.pdf",
		ContentType:      "application/pdf",
		Content:          bytes.NewReader([]byte("Signed invoice content")),
		UserID:           &userID,
		EntityType:       "sales_invoice",
		EntityID:         &salesInvoiceID,
	})
	if err != nil {
		t.Fatalf("failed to upload invoice document: %v", err)
	}

	// 2. Query attachments by entity
	attachments, err := docMod.Service.GetEntityAttachments(ctx, "sales_invoice", salesInvoiceID)
	if err != nil {
		t.Fatalf("failed to get attachments for sales_invoice: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].FileID != file.ID {
		t.Errorf("expected attachment file ID %s, got %s", file.ID, attachments[0].FileID)
	}
	if attachments[0].File == nil || attachments[0].File.Title != file.Title {
		t.Errorf("expected embedded file with title %q", file.Title)
	}

	// 3. Link the same document to an employee as well
	err = docMod.Service.AttachToEntity(ctx, file.ID, "employee", employeeID, &userID)
	if err != nil {
		t.Fatalf("failed to attach document to employee: %v", err)
	}

	empAttachments, err := docMod.Service.GetEntityAttachments(ctx, "employee", employeeID)
	if err != nil {
		t.Fatalf("failed to get employee attachments: %v", err)
	}
	if len(empAttachments) != 1 {
		t.Fatalf("expected 1 employee attachment, got %d", len(empAttachments))
	}

	// 4. Detach from sales invoice
	err = docMod.Service.DetachFromEntity(ctx, file.ID, "sales_invoice", salesInvoiceID)
	if err != nil {
		t.Fatalf("failed to detach document from invoice: %v", err)
	}

	afterDetach, err := docMod.Service.GetEntityAttachments(ctx, "sales_invoice", salesInvoiceID)
	if err != nil {
		t.Fatalf("failed to query after detach: %v", err)
	}
	if len(afterDetach) != 0 {
		t.Errorf("expected 0 attachments after detach, got %d", len(afterDetach))
	}
}

func TestDocuments_ArchivingAndFiltering(t *testing.T) {
	docMod, db, _, leID, userID, ctx := setupDocumentsIntegrationTest(t)
	defer db.Close()

	// 1. Upload 2 files
	f1, err := docMod.Service.UploadFile(ctx, domain.UploadFileParams{
		LegalEntityID:    leID,
		Title:            "سند قبض إيجار",
		Category:         domain.CategoryReceipt,
		OriginalFilename: "receipt_001.pdf",
		ContentType:      "application/pdf",
		Content:          bytes.NewReader([]byte("Receipt 1")),
		UserID:           &userID,
	})
	if err != nil {
		t.Fatalf("failed to upload f1: %v", err)
	}

	f2, err := docMod.Service.UploadFile(ctx, domain.UploadFileParams{
		LegalEntityID:    leID,
		Title:            "بطاقة هوية وطنية",
		Category:         domain.CategoryIdentity,
		OriginalFilename: "national_id.pdf",
		ContentType:      "application/pdf",
		Content:          bytes.NewReader([]byte("Identity 1")),
		UserID:           &userID,
	})
	if err != nil {
		t.Fatalf("failed to upload f2: %v", err)
	}

	// 2. Filter by Category
	catReceipt := domain.CategoryReceipt
	receipts, total, err := docMod.Service.ListFiles(ctx, domain.FileFilter{
		LegalEntityID: leID,
		Category:      &catReceipt,
	})
	if err != nil {
		t.Fatalf("failed to list by category: %v", err)
	}
	if total < 1 || len(receipts) < 1 {
		t.Fatalf("expected at least 1 receipt, got %d", total)
	}

	// 3. Archive f1
	if err := docMod.Service.ArchiveFile(ctx, leID, f1.ID); err != nil {
		t.Fatalf("failed to archive f1: %v", err)
	}

	// 4. List without IncludeArchived (should exclude f1)
	activeList, _, err := docMod.Service.ListFiles(ctx, domain.FileFilter{
		LegalEntityID:   leID,
		IncludeArchived: false,
	})
	if err != nil {
		t.Fatalf("failed to list active files: %v", err)
	}
	for _, f := range activeList {
		if f.ID == f1.ID {
			t.Errorf("archived file f1 should not appear in active list")
		}
	}

	// 5. Attempting to add version to archived file must fail
	_, err = docMod.Service.UploadVersion(ctx, domain.UploadVersionParams{
		FileID:           f1.ID,
		LegalEntityID:    leID,
		OriginalFilename: "new_version.pdf",
		Content:          bytes.NewReader([]byte("new content")),
	})
	if err != domain.ErrFileArchived {
		t.Errorf("expected ErrFileArchived, got: %v", err)
	}

	// 6. Summary metrics
	summary, err := docMod.Service.GetSummary(ctx, leID)
	if err != nil {
		t.Fatalf("failed to get summary: %v", err)
	}
	if summary.TotalFiles < 1 {
		t.Errorf("expected total files >= 1, got %d", summary.TotalFiles)
	}

	_ = f2
}

func TestDocuments_SeedDefaultProcedure(t *testing.T) {
	docMod, db, _, _, _, ctx := setupDocumentsIntegrationTest(t)
	defer db.Close()

	// Create a fresh legal entity to test seed procedure without interference from previous tests
	var freshLEID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO organization.legal_entities (
			id, organization_id, code, name_en, name_ar, cr_number, 
			base_currency, is_active
		) VALUES (
			uuidv7(), (SELECT organization_id FROM organization.legal_entities LIMIT 1),
			'DOC-SEED-' || substr(uuidv7()::text, 1, 8), 
			'Seed Doc Corp', 'شركة الوثائق التجريبية', 
			'1010' || substr((random()*1000000)::int::text, 1, 6),
			'SAR', true
		) RETURNING id
	`).Scan(&freshLEID)
	if err != nil {
		t.Fatalf("failed to create fresh legal entity for seed test: %v", err)
	}

	// Seed procedure
	if err := docMod.SeedDefaultDocumentsData(ctx, freshLEID); err != nil {
		t.Fatalf("failed to seed default documents data: %v", err)
	}

	// Verify files created
	files, total, err := docMod.Service.ListFiles(ctx, domain.FileFilter{
		LegalEntityID: freshLEID,
	})
	if err != nil {
		t.Fatalf("failed to list seeded documents: %v", err)
	}
	if total < 3 {
		t.Fatalf("expected at least 3 default documents seeded, got %d", total)
	}

	foundCR := false
	foundVAT := false
	for _, f := range files {
		if strings.Contains(f.Title, "السجل التجاري") {
			foundCR = true
		}
		if strings.Contains(f.Title, "ضريبة القيمة المضافة") {
			foundVAT = true
		}
	}

	if !foundCR {
		t.Error("expected Commercial Registration document in seeded list")
	}
	if !foundVAT {
		t.Error("expected VAT Certificate document in seeded list")
	}
}

func TestDocuments_HTTPMultipartUpload(t *testing.T) {
	docMod, db, _, leID, _, _ := setupDocumentsIntegrationTest(t)
	defer db.Close()

	r := chi.NewRouter()
	docMod.RegisterRoutes(r, nil)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	_ = writer.WriteField("legal_entity_id", leID.String())
	_ = writer.WriteField("title", "عقد عمل استشاري")
	_ = writer.WriteField("category", "CONTRACT")
	_ = writer.WriteField("description", "عقد استشارات إدارية معتمد")

	fileWriter, err := writer.CreateFormFile("file", "consulting_contract.pdf")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, _ = fileWriter.Write([]byte("Sample consulting contract content for Saudi ERP"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var created domain.File
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode created file response: %v", err)
	}
	if created.Title != "عقد عمل استشاري" {
		t.Errorf("expected title 'عقد عمل استشاري', got %q", created.Title)
	}
	if created.CurrentVersion == nil || created.CurrentVersion.VersionNumber != 1 {
		t.Errorf("expected version 1 in response")
	}
}
