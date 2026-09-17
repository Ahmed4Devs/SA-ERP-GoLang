package zatca_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"erp/internal/zatca/infrastructure/client"
	"erp/internal/zatca/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupZATCAIntegrationTest(t *testing.T) (*zatca.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("PostgreSQL not accessible: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	_ = database.MigrateDown(ctx, cfg.Database.URL, "../../../database/migrations")
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

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

func createAndSignTestDoc(t *testing.T, zMod *zatca.Module, leID, egsID uuid.UUID, invType domain.InvoiceType, ctx context.Context) *domain.ZATCADocument {
	t.Helper()

	invData := domain.InvoiceData{
		InvoiceNumber:  fmt.Sprintf("INV-INT-%s", uuid.New().String()[:6]),
		UUID:           uuid.New(),
		IssueDate:      time.Now().Format("2006-01-02"),
		IssueTime:      time.Now().Format("15:04:05"),
		InvoiceType:    invType,
		InvoiceSubtype: domain.InvoiceSubtypeTaxInvoice,
		CurrencyCode:   "SAR",
		Seller: domain.PartyInfo{
			Name:           "شركة الأفق المتقدمة المحدودة",
			VATNumber:      "300012345600003",
			Street:         "شارع الملك فهد",
			BuildingNumber: "1234",
			District:       "العليا",
			City:           "الرياض",
			PostalCode:     "12214",
			CountryCode:    "SA",
		},
		Buyer: domain.PartyInfo{
			Name:           "مؤسسة الرائد للمقاولات",
			VATNumber:      "310987654300003",
			Street:         "طريق مكة المكرمة",
			BuildingNumber: "5678",
			District:       "الربوة",
			City:           "الرياض",
			PostalCode:     "12813",
			CountryCode:    "SA",
		},
		Lines: []domain.InvoiceLineData{
			{
				ID:        "1",
				Name:      "خوادم سحابية وتراخيص سنوية",
				Quantity:  decimal.NewFromInt(1),
				UnitPrice: decimal.NewFromInt(1000),
				Discount:  decimal.Zero,
				VATRate:   decimal.NewFromInt(15),
				VATAmount: decimal.NewFromInt(150),
				LineTotal: decimal.NewFromInt(1150),
			},
		},
		LineExtensionAmount: decimal.NewFromInt(1000),
		TaxExclusiveAmount:  decimal.NewFromInt(1000),
		TaxInclusiveAmount:  decimal.NewFromInt(1150),
		AllowanceTotal:      decimal.Zero,
		TaxTotal:            decimal.NewFromInt(150),
		PayableAmount:       decimal.NewFromInt(1150),
		ICV:                 1,
		PIH:                 "NWZlY2ViNjZmZmM4NmYzOGQ5NTI3ODZjNmQ2OTZjNzljMmRiYzIzOWRkNGU5MWI0NjQyYjc4NTAzMzYzZDE5MA==",
	}

	doc, err := zMod.Service.GenerateDocumentFromInvoiceData(ctx, leID, egsID, invData)
	if err != nil {
		t.Fatalf("failed to generate doc: %v", err)
	}

	signedDoc, err := zMod.Service.SignDocument(ctx, doc.ID, leID)
	if err != nil {
		t.Fatalf("failed to sign doc: %v", err)
	}

	return signedDoc
}

func TestZATCA_OnboardingLifecycle(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCAIntegrationTest(t)

	// 1. Register EGS Unit
	egs, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             fmt.Sprintf("EGS-ONBOARD-%s", uuid.New().String()[:6]),
		Model:                "POS-Terminal-2026",
		SerialNumber:         fmt.Sprintf("SN-%s", uuid.New().String()[:8]),
		SolutionName:         "Saudi ERP FATOORA Onboarder",
		VATNumber:            "300012345600003",
		OrganizationName:     "شركة الأفق المتقدمة المحدودة",
		OrganizationUnitName: "فرع الرياض الرئيسي",
		InvoiceTypeMask:      "1100",
	})
	if err != nil {
		t.Fatalf("failed to register EGS unit: %v", err)
	}

	// 2. Initial Onboarding Status
	status, err := zMod.OnboardingService.GetOnboardingStatus(ctx, leID, egs.ID)
	if err != nil {
		t.Fatalf("failed to get onboarding status: %v", err)
	}
	if status.HasProductionCSID {
		t.Errorf("expected new EGS to not have production CSID")
	}

	// 3. Execute Complete Automated Onboarding with OTP
	res, err := zMod.OnboardingService.StartOnboarding(ctx, leID, egs.ID, "123456")
	if err != nil {
		t.Fatalf("failed to start onboarding: %v", err)
	}

	if res.Status != "COMPLETED" {
		t.Errorf("expected onboarding status COMPLETED, got %s", res.Status)
	}
	if res.ProductionCSIDID == uuid.Nil {
		t.Errorf("expected valid Production CSID UUID")
	}

	// 4. Verify Final Onboarding Status
	finalStatus, err := zMod.OnboardingService.GetOnboardingStatus(ctx, leID, egs.ID)
	if err != nil {
		t.Fatalf("failed to get final onboarding status: %v", err)
	}
	if !finalStatus.HasComplianceCSID {
		t.Errorf("expected EGS to have Compliance CSID")
	}
	if !finalStatus.HasProductionCSID {
		t.Errorf("expected EGS to have Production CSID")
	}
	if finalStatus.Status != "ONBOARDED" {
		t.Errorf("expected status ONBOARDED, got %s", finalStatus.Status)
	}
}

func TestZATCA_ClearanceWorkflow(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCAIntegrationTest(t)

	// 1. Setup EGS and Onboard
	egs, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             fmt.Sprintf("EGS-CLR-%s", uuid.New().String()[:6]),
		Model:                "Server-2026",
		SerialNumber:         fmt.Sprintf("SN-%s", uuid.New().String()[:8]),
		SolutionName:         "ERP Clearance Engine",
		VATNumber:            "300012345600003",
		OrganizationName:     "شركة الأفق المتقدمة المحدودة",
		OrganizationUnitName: "الإدارة المالية",
		InvoiceTypeMask:      "1100",
	})
	if err != nil {
		t.Fatalf("failed to register EGS: %v", err)
	}

	_, err = zMod.OnboardingService.StartOnboarding(ctx, leID, egs.ID, "123456")
	if err != nil {
		t.Fatalf("failed to onboard EGS: %v", err)
	}

	// 2. Generate and Sign Standard B2B Invoice
	signedDoc := createAndSignTestDoc(t, zMod, leID, egs.ID, domain.InvoiceTypeStandard, ctx)

	// 3. Submit Clearance
	resp, err := zMod.SubmissionService.SubmitClearance(ctx, signedDoc.ID, leID)
	if err != nil {
		t.Fatalf("failed to submit clearance: %v", err)
	}

	if resp.ClearanceStatus != "CLEARED" {
		t.Errorf("expected clearance status CLEARED, got %s", resp.ClearanceStatus)
	}
	if resp.ClearedInvoice == "" {
		t.Errorf("expected non-empty cleared invoice")
	}
	if resp.ZATCARequestID == "" {
		t.Errorf("expected non-empty ZATCA request ID")
	}

	// 4. Verify Document Status updated to CLEARED
	updatedDoc, err := zMod.Service.GetDocument(ctx, signedDoc.ID, leID)
	if err != nil {
		t.Fatalf("failed to get updated doc: %v", err)
	}
	if updatedDoc.Status != domain.DocumentStatusCleared {
		t.Errorf("expected document status CLEARED, got %s", updatedDoc.Status)
	}

	// 5. Verify Submissions Audit Trail
	sub, err := zMod.SubRepo.GetSubmissionByDocumentID(ctx, signedDoc.ID, leID)
	if err != nil || sub == nil {
		t.Fatalf("expected submission record for cleared invoice: %v", err)
	}
	if sub.Status != "CLEARED" {
		t.Errorf("expected submission status CLEARED, got %s", sub.Status)
	}
	if len(sub.Attempts) == 0 {
		t.Errorf("expected submission attempt log")
	}
	if sub.Attempts[0].HTTPStatusCode != 200 {
		t.Errorf("expected HTTP 200 in attempt log, got %d", sub.Attempts[0].HTTPStatusCode)
	}
}

func TestZATCA_ReportingWorkflow(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCAIntegrationTest(t)

	// 1. Setup EGS and Onboard
	egs, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             fmt.Sprintf("EGS-REP-%s", uuid.New().String()[:6]),
		Model:                "POS-Register-B2C",
		SerialNumber:         fmt.Sprintf("SN-%s", uuid.New().String()[:8]),
		SolutionName:         "ERP Reporting Engine",
		VATNumber:            "300012345600003",
		OrganizationName:     "شركة الأفق المتقدمة المحدودة",
		OrganizationUnitName: "نقاط البيع",
		InvoiceTypeMask:      "0100",
	})
	if err != nil {
		t.Fatalf("failed to register EGS: %v", err)
	}

	_, err = zMod.OnboardingService.StartOnboarding(ctx, leID, egs.ID, "123456")
	if err != nil {
		t.Fatalf("failed to onboard EGS: %v", err)
	}

	// 2. Generate and Sign Simplified B2C Invoice
	signedDoc := createAndSignTestDoc(t, zMod, leID, egs.ID, domain.InvoiceTypeSimplified, ctx)

	// 3. Submit Reporting
	resp, err := zMod.SubmissionService.SubmitReporting(ctx, signedDoc.ID, leID)
	if err != nil {
		t.Fatalf("failed to submit reporting: %v", err)
	}

	if resp.ReportingStatus != "REPORTED" {
		t.Errorf("expected reporting status REPORTED, got %s", resp.ReportingStatus)
	}
	if resp.ZATCARequestID == "" {
		t.Errorf("expected non-empty ZATCA request ID")
	}

	// 4. Verify Document Status updated to REPORTED
	updatedDoc, err := zMod.Service.GetDocument(ctx, signedDoc.ID, leID)
	if err != nil {
		t.Fatalf("failed to get updated doc: %v", err)
	}
	if updatedDoc.Status != domain.DocumentStatusReported {
		t.Errorf("expected document status REPORTED, got %s", updatedDoc.Status)
	}
}

func TestZATCA_SubmissionRetriesAndWorker(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCAIntegrationTest(t)

	// 1. Setup EGS and Onboard
	egs, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             fmt.Sprintf("EGS-RETRY-%s", uuid.New().String()[:6]),
		Model:                "POS-Resilience",
		SerialNumber:         fmt.Sprintf("SN-%s", uuid.New().String()[:8]),
		SolutionName:         "ERP Resilience Engine",
		VATNumber:            "300012345600003",
		OrganizationName:     "شركة الأفق المتقدمة المحدودة",
		OrganizationUnitName: "إدارة العمليات",
		InvoiceTypeMask:      "1100",
	})
	if err != nil {
		t.Fatalf("failed to register EGS: %v", err)
	}

	_, _ = zMod.OnboardingService.StartOnboarding(ctx, leID, egs.ID, "123456")

	// 2. Sign document
	signedDoc := createAndSignTestDoc(t, zMod, leID, egs.ID, domain.InvoiceTypeStandard, ctx)

	// 3. Set simulator into network failure mode
	simClient, ok := zMod.ZATCAClient.(*client.SimulatorClient)
	if !ok {
		t.Fatalf("expected SimulatorClient")
	}
	simClient.SetFailureMode(true, false, false)

	// 4. Attempt clearance under network failure
	_, err = zMod.SubmissionService.SubmitClearance(ctx, signedDoc.ID, leID)
	if err == nil {
		t.Fatalf("expected error under network failure")
	}

	// 5. Verify status queued as RETRY_PENDING
	doc, _ := zMod.Service.GetDocument(ctx, signedDoc.ID, leID)
	if doc.Status != domain.DocumentStatusRetryPending {
		t.Errorf("expected document status RETRY_PENDING, got %s", doc.Status)
	}

	pending, err := zMod.SubRepo.GetPendingRetrySubmissions(ctx, 10)
	if err != nil {
		t.Fatalf("failed to get pending retries: %v", err)
	}
	if len(pending) == 0 {
		t.Fatalf("expected at least 1 pending retry submission")
	}

	// 6. Restore network connection and process retries with RetryWorker
	simClient.SetFailureMode(false, false, false)

	processed, err := zMod.RetryWorker.ProcessPending(ctx)
	if err != nil {
		t.Fatalf("failed to process pending retries: %v", err)
	}
	if processed == 0 {
		t.Errorf("expected retry worker to process at least 1 pending submission")
	}

	// 7. Verify document recovered to CLEARED
	finalDoc, _ := zMod.Service.GetDocument(ctx, signedDoc.ID, leID)
	if finalDoc.Status != domain.DocumentStatusCleared {
		t.Errorf("expected document status to recover to CLEARED, got %s", finalDoc.Status)
	}
}

func TestZATCA_HTTPIntegrationTransport(t *testing.T) {
	zMod, _, leID, _, ctx := setupZATCAIntegrationTest(t)

	handler := zMod.Handler.Routes(nil)

	// 1. Setup EGS Unit
	egs, err := zMod.Service.RegisterEGSUnit(ctx, ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		CustomID:             fmt.Sprintf("EGS-HTTP-%s", uuid.New().String()[:6]),
		Model:                "Cloud-Gateway",
		SerialNumber:         fmt.Sprintf("SN-%s", uuid.New().String()[:8]),
		SolutionName:         "ERP HTTP Gateway",
		VATNumber:            "300012345600003",
		OrganizationName:     "شركة الأفق المتقدمة المحدودة",
		OrganizationUnitName: "البوابة السحابية",
		InvoiceTypeMask:      "1100",
	})
	if err != nil {
		t.Fatalf("failed to register EGS: %v", err)
	}

	// 2. HTTP Onboarding POST
	onboardPayload := map[string]string{
		"legal_entity_id": leID.String(),
		"egs_unit_id":     egs.ID.String(),
		"otp":             "654321",
	}
	body, _ := json.Marshal(onboardPayload)
	req := httptest.NewRequest(http.MethodPost, "/onboarding/start", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from onboarding start, got %d: %s", w.Code, w.Body.String())
	}

	// 3. HTTP Onboarding Status GET
	statusReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/onboarding/status?legal_entity_id=%s&egs_unit_id=%s", leID, egs.ID), nil)
	statusW := httptest.NewRecorder()
	handler.ServeHTTP(statusW, statusReq)

	if statusW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from onboarding status, got %d", statusW.Code)
	}

	// 4. Create & Sign Doc
	signedDoc := createAndSignTestDoc(t, zMod, leID, egs.ID, domain.InvoiceTypeStandard, ctx)

	// 5. HTTP Clearance POST
	clrReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/documents/%s/clearance?legal_entity_id=%s", signedDoc.ID, leID), nil)
	clrW := httptest.NewRecorder()
	handler.ServeHTTP(clrW, clrReq)

	if clrW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from HTTP clearance, got %d: %s", clrW.Code, clrW.Body.String())
	}

	// 6. HTTP Submissions List GET
	listReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/submissions?legal_entity_id=%s", leID), nil)
	listW := httptest.NewRecorder()
	handler.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from list submissions, got %d", listW.Code)
	}

	var subs []*domain.ZATCASubmission
	if err := json.Unmarshal(listW.Body.Bytes(), &subs); err != nil {
		t.Fatalf("failed to unmarshal submissions: %v", err)
	}
	if len(subs) == 0 {
		t.Fatalf("expected at least 1 submission returned from HTTP API")
	}
}
