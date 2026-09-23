package dashboard_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/accounting"
	"erp/internal/dashboard"
	"erp/internal/dashboard/domain"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
)

func setupDashboardTest(t *testing.T) (*dashboard.Module, uuid.UUID, context.Context) {
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

	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	dashMod := dashboard.NewModule(cfg, db)

	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	_ = accMod.SeedDefaultCOA(ctx, le.ID)

	return dashMod, le.ID, ctx
}

func TestDashboard_GetDashboardData(t *testing.T) {
	mod, leID, ctx := setupDashboardTest(t)

	data, err := mod.Service.GetDashboardData(ctx, leID)
	if err != nil {
		t.Fatalf("failed to get dashboard data: %v", err)
	}

	if data == nil {
		t.Fatal("expected dashboard data, got nil")
	}

	if data.LegalEntityID != leID {
		t.Errorf("expected legal entity ID %v, got %v", leID, data.LegalEntityID)
	}

	if data.Currency == "" {
		t.Error("expected currency to be populated")
	}

	// Verify all widgets are initialized
	if data.Financials.Currency == "" {
		t.Error("expected financials currency to be populated")
	}

	if data.CashPosition.Currency == "" {
		t.Error("expected cash position currency to be populated")
	}

	if data.ARAP.Currency == "" {
		t.Error("expected AR/AP currency to be populated")
	}

	if data.StockWarnings == nil {
		t.Error("expected stock warnings slice, got nil")
	}

	if data.Approvals.RecentRequests == nil {
		t.Error("expected approvals recent requests slice, got nil")
	}

	if data.RecentDocs == nil {
		t.Error("expected recent docs slice, got nil")
	}

	if data.ActivityFeed == nil {
		t.Error("expected activity feed slice, got nil")
	}
}

func TestDashboard_HTTPHandler(t *testing.T) {
	mod, leID, _ := setupDashboardTest(t)

	req := httptest.NewRequest(http.MethodGet, "/?legal_entity_id="+leID.String(), nil)
	w := httptest.NewRecorder()

	handler := mod.Handler.Routes(nil)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp domain.DashboardData
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.LegalEntityID != leID {
		t.Errorf("expected legal entity ID %v, got %v", leID, resp.LegalEntityID)
	}
}
