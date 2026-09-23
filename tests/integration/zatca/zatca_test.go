package zatca_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/zatca"
	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupZATCATest(t *testing.T) (*zatca.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
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
	_, _ = db.Pool.Exec(ctx, `TRUNCATE TABLE zatca.egs_units, zatca.csids, zatca.invoice_sequences, zatca.documents, zatca.document_hashes, zatca.submissions, zatca.submission_attempts, zatca.validation_results, zatca.validation_messages CASCADE`)

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	zMod := zatca.NewModule(cfg, db)

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
		createdAdmin, _ := iamMod.AuthService.SeedAdmin(ctx, "admin", "admin@erp.sa", "System Administrator", "Admin@ERP123!")
		if createdAdmin != nil {
			adminID = createdAdmin.ID
		}
	}

	return zMod, db, le.ID, adminID, ctx
}

func sampleInvoiceData() domain.InvoiceData {
	return domain.InvoiceData{
		InvoiceNumber:       "INV-2026-" + uuid.New().String()[:8],
		UUID:                uuid.New(),
		IssueDate:           time.Now().Format("2006-01-02"),
		IssueTime:           "10:30:00",
		InvoiceType:         domain.InvoiceTypeStandard,
		InvoiceSubtype:      domain.InvoiceSubtypeTaxInvoice,
		CurrencyCode:        "SAR",
		LineExtensionAmount: decimal.NewFromInt(1000),
		TaxExclusiveAmount:  decimal.NewFromInt(1000),
		TaxInclusiveAmount:  decimal.NewFromInt(1150),
		AllowanceTotal:      decimal.Zero,
		TaxTotal:            decimal.NewFromInt(150),
		PayableAmount:       decimal.NewFromInt(1150),
		ICV:                 1,
		PIH:                 "NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz",
		Seller: domain.PartyInfo{
			Name:           "Saudi Horizon Co. Ltd.",
			VATNumber:      "300000000000003",
			Street:         "King Fahd Road",
			BuildingNumber: "1234",
			District:       "Al-Olaya",
			City:           "Riyadh",
			PostalCode:     "12345",
			CountryCode:    "SA",
		},
		Buyer: domain.PartyInfo{
			Name:           "Advanced Solutions Corp",
			VATNumber:      "311111111111113",
			Street:         "Takhassusi Street",
			BuildingNumber: "4567",
			District:       "Al-Muhammadiyah",
			City:           "Riyadh",
			PostalCode:     "54321",
			CountryCode:    "SA",
		},
		Lines: []domain.InvoiceLineData{
			{
				ID:        "1",
				Name:      "Enterprise Cloud ERP Annual License",
				Quantity:  decimal.NewFromInt(1),
				UnitPrice: decimal.NewFromInt(1000),
				Discount:  decimal.Zero,
				VATRate:   decimal.NewFromInt(15),
				VATAmount: decimal.NewFromInt(150),
				LineTotal: decimal.NewFromInt(1150),
			},
		},
	}
}

func TestZATCA_EGSAndCSIDLifecycle(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	// 1. Register EGS Unit
	unit, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             "EGS-TEST-01",
		Model:                "POS-Terminal-2026",
		SerialNumber:         "SN-TEST-" + uuid.New().String()[:6],
		SolutionName:         "Saudi ERP Suite",
		VATNumber:            "300000000000003",
		OrganizationName:     "شركة الأفق المتقدمة",
		OrganizationUnitName: "Riyadh Branch",
		InvoiceTypeMask:      "1100",
	})
	if err != nil {
		t.Fatalf("failed to register EGS unit: %v", err)
	}

	if unit.Status != "ACTIVE" {
		t.Errorf("expected status ACTIVE, got %s", unit.Status)
	}

	// 2. Verify CSID generated automatically
	csids, err := zMod.Service.ListCSIDs(ctx, unit.ID, leID)
	if err != nil {
		t.Fatalf("failed to list CSIDs: %v", err)
	}
	if len(csids) == 0 {
		t.Fatalf("expected at least 1 CSID generated for EGS")
	}

	activeCSID := csids[0]
	if activeCSID.PrivateKeyPEM == "" || activeCSID.PublicKeyPEM == "" {
		t.Errorf("expected generated ECDSA key pair")
	}
}

func TestZATCA_UBL21GenerationAndHashing(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	unit, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             "EGS-UBL-01",
		SerialNumber:         "SN-UBL-" + uuid.New().String()[:6],
		VATNumber:            "300000000000003",
		OrganizationName:     "شركة الأفق",
		OrganizationUnitName: "HQ",
	})
	if err != nil {
		t.Fatalf("failed to register EGS unit: %v", err)
	}

	invData := sampleInvoiceData()
	doc, err := zMod.Service.GenerateDocumentFromInvoiceData(ctx, leID, unit.ID, invData)
	if err != nil {
		t.Fatalf("failed to generate document: %v", err)
	}

	if doc.UBLXML == "" {
		t.Errorf("expected non-empty UBL XML")
	}
	if doc.InvoiceHash == "" {
		t.Errorf("expected calculated invoice SHA-256 hash")
	}
	if doc.QRCodeData == "" {
		t.Errorf("expected generated QR code")
	}
}

func TestZATCA_BRKSARuleValidation(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	unit, _ := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             "EGS-VAL-01",
		SerialNumber:         "SN-VAL-" + uuid.New().String()[:6],
		VATNumber:            "300000000000003",
		OrganizationName:     "شركة الأفق",
		OrganizationUnitName: "HQ",
	})

	invData := sampleInvoiceData()
	doc, err := zMod.Service.GenerateDocumentFromInvoiceData(ctx, leID, unit.ID, invData)
	if err != nil {
		t.Fatalf("failed to generate document: %v", err)
	}

	// Validate against BR-KSA business rules
	valResult, err := zMod.Service.ValidateDocument(ctx, doc.ID, leID)
	if err != nil {
		t.Fatalf("failed to validate document: %v", err)
	}

	if valResult.OverallStatus != "PASSED" {
		t.Errorf("expected validation to pass, got %s: %+v", valResult.OverallStatus, valResult.Messages)
	}

	// Verify status updated to VALIDATED
	updatedDoc, _ := zMod.Service.GetDocument(ctx, doc.ID, leID)
	if updatedDoc.Status != domain.DocumentStatusValidated {
		t.Errorf("expected document status VALIDATED, got %s", updatedDoc.Status)
	}
}

func TestZATCA_SigningAndTLVQR(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	unit, _ := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             "EGS-SIGN-01",
		SerialNumber:         "SN-SIGN-" + uuid.New().String()[:6],
		VATNumber:            "300000000000003",
		OrganizationName:     "شركة الأفق",
		OrganizationUnitName: "HQ",
	})

	invData := sampleInvoiceData()
	doc, _ := zMod.Service.GenerateDocumentFromInvoiceData(ctx, leID, unit.ID, invData)
	_, _ = zMod.Service.ValidateDocument(ctx, doc.ID, leID)

	// Digitally Sign Document
	signedDoc, err := zMod.Service.SignDocument(ctx, doc.ID, leID)
	if err != nil {
		t.Fatalf("failed to digitally sign document: %v", err)
	}

	if signedDoc.Status != domain.DocumentStatusSigned {
		t.Errorf("expected status SIGNED, got %s", signedDoc.Status)
	}
	if signedDoc.SignedXML == "" {
		t.Errorf("expected signed XML with XAdES block")
	}
	if signedDoc.DigitalSignature == "" {
		t.Errorf("expected digital signature string")
	}
}

func TestZATCA_ChainingSequences(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	unit, _ := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             "EGS-CHAIN-01",
		SerialNumber:         "SN-CHAIN-" + uuid.New().String()[:6],
		VATNumber:            "300000000000003",
		OrganizationName:     "شركة الأفق",
		OrganizationUnitName: "HQ",
	})

	// Invoice 1
	data1 := sampleInvoiceData()
	data1.InvoiceNumber = "INV-CHAIN-001"
	doc1, err := zMod.Service.GenerateDocumentFromInvoiceData(ctx, leID, unit.ID, data1)
	if err != nil {
		t.Fatalf("failed to generate doc 1: %v", err)
	}

	if doc1.ICV != 1 {
		t.Errorf("expected ICV 1 for first invoice, got %d", doc1.ICV)
	}

	// Invoice 2
	data2 := sampleInvoiceData()
	data2.InvoiceNumber = "INV-CHAIN-002"
	doc2, err := zMod.Service.GenerateDocumentFromInvoiceData(ctx, leID, unit.ID, data2)
	if err != nil {
		t.Fatalf("failed to generate doc 2: %v", err)
	}

	if doc2.ICV != 2 {
		t.Errorf("expected ICV 2 for second invoice, got %d", doc2.ICV)
	}

	// PIH of doc2 must equal InvoiceHash of doc1!
	if doc2.PIH != doc1.InvoiceHash {
		t.Errorf("expected PIH chaining: doc2.PIH (%s) must equal doc1.InvoiceHash (%s)", doc2.PIH, doc1.InvoiceHash)
	}
}

func TestZATCA_SeedProcedure(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	if err := zMod.SeedDefaultZATCAData(ctx, leID); err != nil {
		t.Fatalf("failed to seed default zatca data: %v", err)
	}

	units, err := zMod.Service.ListEGSUnits(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list EGS units: %v", err)
	}
	if len(units) == 0 {
		t.Fatalf("expected seeded EGS units")
	}

	summary, err := zMod.Service.GetSummary(ctx, leID)
	if err != nil {
		t.Fatalf("failed to get summary: %v", err)
	}
	if summary.ActiveEGSCount == 0 {
		t.Errorf("expected active EGS count > 0")
	}
}

func TestZATCA_HTTPTransport(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCATest(t)

	_ = zMod.SeedDefaultZATCAData(ctx, leID)

	r := chi.NewRouter()
	zMod.RegisterRoutes(r, nil)

	// 1. GET /api/v1/zatca/egs-units
	req := httptest.NewRequest(http.MethodGet, "/api/v1/zatca/egs-units?legal_entity_id="+leID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. GET /api/v1/zatca/summary
	reqSum := httptest.NewRequest(http.MethodGet, "/api/v1/zatca/summary?legal_entity_id="+leID.String(), nil)
	wSum := httptest.NewRecorder()
	r.ServeHTTP(wSum, reqSum)

	if wSum.Code != http.StatusOK {
		t.Fatalf("expected summary status 200, got %d: %s", wSum.Code, wSum.Body.String())
	}
}
