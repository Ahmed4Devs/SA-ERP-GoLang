package budgeting_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/accounting"
	"erp/internal/budgeting"
	"erp/internal/budgeting/domain"
	"erp/internal/budgeting/ports"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupBudgetingTest(t *testing.T) (*budgeting.Module, *accounting.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	bMod := budgeting.NewModule(cfg, db)

	// 1. Seed Organization
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Seed default Chart of Accounts and Fiscal Calendar
	if err := accMod.SeedDefaultCOA(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed default COA: %v", err)
	}

	// 3. Admin User
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

	// Retrieve active fiscal year
	var fiscalYearID uuid.UUID
	err = db.Pool.QueryRow(ctx, `SELECT id FROM accounting.fiscal_years WHERE legal_entity_id = $1 LIMIT 1`, le.ID).Scan(&fiscalYearID)
	if err != nil {
		t.Fatalf("failed to query fiscal year: %v", err)
	}

	return bMod, accMod, db, le.ID, fiscalYearID, adminID, ctx
}

func TestBudget_CreationAndLifecycle(t *testing.T) {
	bMod, _, _, leID, fyID, adminID, ctx := setupBudgetingTest(t)

	// 1. Create Draft Budget
	code := "BUD-TEST-" + uuid.New().String()[:6]
	created, err := bMod.Service.CreateBudget(ctx, ports.CreateBudgetRequest{
		LegalEntityID: leID,
		Code:          code,
		NameEn:        "Marketing & Growth Budget",
		NameAr:        "ميزانية التسويق والنمو",
		DescriptionEn: "FY2026 digital marketing allocation",
		DescriptionAr: "مخصصات التسويق الرقمي للعام المالي 2026",
		FiscalYearID:  fyID,
		CurrencyCode:  "SAR",
		BudgetType:    domain.BudgetTypeOperating,
		CreatedBy:     &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	if created.Status != domain.BudgetStatusDraft {
		t.Errorf("expected status DRAFT, got %s", created.Status)
	}

	// Verify Version 1 auto-created
	versions, err := bMod.Service.ListVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to list versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 initial version, got %d", len(versions))
	}
	if versions[0].VersionNumber != 1 {
		t.Errorf("expected version number 1, got %d", versions[0].VersionNumber)
	}

	// 2. Submit Budget
	submitted, err := bMod.Service.SubmitBudget(ctx, created.ID, leID)
	if err != nil {
		t.Fatalf("failed to submit budget: %v", err)
	}
	if submitted.Status != domain.BudgetStatusSubmitted {
		t.Errorf("expected status SUBMITTED, got %s", submitted.Status)
	}

	// 3. Approve Budget
	approved, err := bMod.Service.ApproveBudget(ctx, created.ID, leID, adminID)
	if err != nil {
		t.Fatalf("failed to approve budget: %v", err)
	}
	if approved.Status != domain.BudgetStatusActive {
		t.Errorf("expected status ACTIVE, got %s", approved.Status)
	}
}

func TestBudget_MultiDimensionalLines(t *testing.T) {
	bMod, _, db, leID, fyID, adminID, ctx := setupBudgetingTest(t)

	// Retrieve expense account
	var accID uuid.UUID
	err := db.Pool.QueryRow(ctx, `SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND account_type = 'EXPENSE' LIMIT 1`, leID).Scan(&accID)
	if err != nil {
		t.Fatalf("failed to query expense account: %v", err)
	}

	// Retrieve department or create if none exists
	var deptID *uuid.UUID
	var dID uuid.UUID
	err = db.Pool.QueryRow(ctx, `SELECT id FROM organization.departments WHERE legal_entity_id = $1 LIMIT 1`, leID).Scan(&dID)
	if err == nil {
		deptID = &dID
	} else {
		newDID := uuid.New()
		_, errIns := db.Pool.Exec(ctx, `INSERT INTO organization.departments (id, legal_entity_id, code, name_en, name_ar) VALUES ($1, $2, $3, $4, $5)`,
			newDID, leID, "DEP-OPS-"+uuid.New().String()[:4], "Operations", "إدارة العمليات")
		if errIns == nil {
			deptID = &newDID
		}
	}

	// Create Budget
	code := "BUD-LINES-" + uuid.New().String()[:6]
	b, err := bMod.Service.CreateBudget(ctx, ports.CreateBudgetRequest{
		LegalEntityID: leID,
		Code:          code,
		NameEn:        "Operations Budget",
		NameAr:        "الميزانية التشغيلية",
		FiscalYearID:  fyID,
		CurrencyCode:  "SAR",
		BudgetType:    domain.BudgetTypeOperating,
		CreatedBy:     &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	versions, err := bMod.Service.ListVersions(ctx, b.ID)
	if err != nil || len(versions) == 0 {
		t.Fatalf("failed to get versions: %v", err)
	}
	verID := versions[0].ID

	// Add Line 1: 50,000 SAR
	line1, err := bMod.Service.AddBudgetLine(ctx, ports.AddBudgetLineRequest{
		BudgetID:          b.ID,
		BudgetVersionID:   verID,
		LegalEntityID:     leID,
		AccountID:         accID,
		DepartmentID:      deptID,
		ProjectCode:       "PRJ-ALPHA",
		Amount:            decimal.NewFromInt(50000),
		LineDescriptionEn: "Software licenses and tools",
		LineDescriptionAr: "تراخيص البرمجيات والأدوات التقنية",
	})
	if err != nil {
		t.Fatalf("failed to add budget line 1: %v", err)
	}

	if !line1.Amount.Equal(decimal.NewFromInt(50000)) {
		t.Errorf("expected line amount 50000, got %s", line1.Amount)
	}

	// Add Line 2: 25,000 SAR
	_, err = bMod.Service.AddBudgetLine(ctx, ports.AddBudgetLineRequest{
		BudgetID:          b.ID,
		BudgetVersionID:   verID,
		LegalEntityID:     leID,
		AccountID:         accID,
		DepartmentID:      deptID,
		ProjectCode:       "PRJ-BETA",
		Amount:            decimal.NewFromInt(25000),
		LineDescriptionEn: "Cloud hosting infrastructure",
		LineDescriptionAr: "خدمات الاستضافة السحابية",
	})
	if err != nil {
		t.Fatalf("failed to add budget line 2: %v", err)
	}

	// Verify Lines List
	lines, err := bMod.Service.ListBudgetLines(ctx, verID)
	if err != nil {
		t.Fatalf("failed to list lines: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Verify Budget Total Recalculation
	updatedBudget, err := bMod.Service.GetBudget(ctx, b.ID, leID)
	if err != nil {
		t.Fatalf("failed to get updated budget: %v", err)
	}
	if !updatedBudget.TotalBudgetedAmount.Equal(decimal.NewFromInt(75000)) {
		t.Errorf("expected total budgeted amount 75000, got %s", updatedBudget.TotalBudgetedAmount)
	}
}

func TestBudget_FundTransfers(t *testing.T) {
	bMod, _, db, leID, fyID, adminID, ctx := setupBudgetingTest(t)

	var accID uuid.UUID
	_ = db.Pool.QueryRow(ctx, `SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND account_type = 'EXPENSE' LIMIT 1`, leID).Scan(&accID)

	b, err := bMod.Service.CreateBudget(ctx, ports.CreateBudgetRequest{
		LegalEntityID: leID,
		Code:          "BUD-TRF-" + uuid.New().String()[:6],
		NameEn:        "Transfer Test Budget",
		NameAr:        "ميزانية اختبار المناقلات",
		FiscalYearID:  fyID,
		CurrencyCode:  "SAR",
		BudgetType:    domain.BudgetTypeOperating,
		CreatedBy:     &adminID,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	versions, _ := bMod.Service.ListVersions(ctx, b.ID)
	verID := versions[0].ID

	// Line 1: Source (10,000 SAR)
	line1, err := bMod.Service.AddBudgetLine(ctx, ports.AddBudgetLineRequest{
		BudgetID:        b.ID,
		BudgetVersionID: verID,
		LegalEntityID:   leID,
		AccountID:       accID,
		Amount:          decimal.NewFromInt(10000),
	})
	if err != nil {
		t.Fatalf("failed to create line 1: %v", err)
	}

	// Line 2: Target (5,000 SAR)
	line2, err := bMod.Service.AddBudgetLine(ctx, ports.AddBudgetLineRequest{
		BudgetID:        b.ID,
		BudgetVersionID: verID,
		LegalEntityID:   leID,
		AccountID:       accID,
		Amount:          decimal.NewFromInt(5000),
	})
	if err != nil {
		t.Fatalf("failed to create line 2: %v", err)
	}

	// Transfer 3,000 SAR from Line 1 -> Line 2
	trf, err := bMod.Service.TransferBudget(ctx, ports.TransferBudgetRequest{
		LegalEntityID: leID,
		SourceLineID:  line1.ID,
		TargetLineID:  line2.ID,
		Amount:        decimal.NewFromInt(3000),
		Reason:        "Reallocating surplus IT budget to facility repairs",
		TransferredBy: &adminID,
	})
	if err != nil {
		t.Fatalf("failed to transfer budget: %v", err)
	}

	if trf.Status != domain.TransferStatusCompleted {
		t.Errorf("expected transfer status COMPLETED, got %s", trf.Status)
	}

	// Verify updated line amounts
	lines, _ := bMod.Service.ListBudgetLines(ctx, verID)
	for _, l := range lines {
		if l.ID == line1.ID {
			if !l.Amount.Equal(decimal.NewFromInt(7000)) {
				t.Errorf("expected source line amount 7000, got %s", l.Amount)
			}
		} else if l.ID == line2.ID {
			if !l.Amount.Equal(decimal.NewFromInt(8000)) {
				t.Errorf("expected target line amount 8000, got %s", l.Amount)
			}
		}
	}
}

func TestBudget_SeedDefaultData(t *testing.T) {
	bMod, _, db, leID, _, _, ctx := setupBudgetingTest(t)

	// Clean up any test budgets so seed procedure runs cleanly
	if _, err := db.Pool.Exec(ctx, "TRUNCATE budgeting.budget_transfers, budgeting.budget_lines, budgeting.budget_versions, budgeting.budgets CASCADE"); err != nil {
		t.Fatalf("failed to truncate budgets: %v", err)
	}

	// Seed procedure
	if err := bMod.SeedDefaultBudgetingData(ctx, leID); err != nil {
		t.Fatalf("failed to seed default budgeting data: %v", err)
	}

	// Verify seeded budgets exist
	budgets, err := bMod.Service.ListBudgets(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list budgets: %v", err)
	}

	if len(budgets) < 2 {
		t.Fatalf("expected at least 2 seeded budgets (OPEX & CAPEX), got %d", len(budgets))
	}

	// Verify Variance / Budget vs Actual
	var opexBudget *domain.Budget
	for _, b := range budgets {
		if b.Code == "BUD-2026-OPEX" {
			opexBudget = b
			break
		}
	}

	if opexBudget == nil {
		t.Fatalf("expected to find BUD-2026-OPEX")
	}

	items, err := bMod.Service.GetBudgetVsActual(ctx, opexBudget.ID, leID)
	if err != nil {
		t.Fatalf("failed to get budget vs actual: %v", err)
	}

	if len(items) == 0 {
		t.Fatalf("expected budget vs actual items for OPEX budget")
	}

	for _, it := range items {
		if it.BudgetedAmount.IsZero() {
			t.Errorf("expected non-zero budgeted amount")
		}
		if it.VarianceStatus == "" {
			t.Errorf("expected variance status to be set")
		}
	}

	// Summary test
	summary, err := bMod.Service.GetBudgetSummary(ctx, leID)
	if err != nil {
		t.Fatalf("failed to get budget summary: %v", err)
	}
	if summary.ActiveBudgetCount < 1 {
		t.Errorf("expected at least 1 active budget, got %d", summary.ActiveBudgetCount)
	}
	if summary.TotalBudgeted.IsZero() {
		t.Errorf("expected total budgeted amount to be non-zero")
	}
}

func TestBudget_HTTPTransport(t *testing.T) {
	bMod, _, _, leID, _, _, ctx := setupBudgetingTest(t)

	_ = bMod.SeedDefaultBudgetingData(ctx, leID)

	r := chi.NewRouter()
	bMod.RegisterRoutes(r, nil)

	// 1. GET /api/v1/budgeting/budgets
	req := httptest.NewRequest(http.MethodGet, "/api/v1/budgeting/budgets?legal_entity_id="+leID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. GET /api/v1/budgeting/summary
	reqSum := httptest.NewRequest(http.MethodGet, "/api/v1/budgeting/summary?legal_entity_id="+leID.String(), nil)
	wSum := httptest.NewRecorder()
	r.ServeHTTP(wSum, reqSum)

	if wSum.Code != http.StatusOK {
		t.Fatalf("expected summary status 200, got %d: %s", wSum.Code, wSum.Body.String())
	}
}
