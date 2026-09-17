package reporting_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/reporting"
	"erp/internal/reporting/domain"

	"github.com/google/uuid"
)

func setupReportingTest(t *testing.T) (*reporting.Module, uuid.UUID, context.Context) {
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

	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	reportingMod := reporting.NewModule(cfg, db)

	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	_ = accMod.SeedDefaultCOA(ctx, le.ID)
	_ = reportingMod.SeedDefaultReportingData(ctx, le.ID)

	return reportingMod, le.ID, ctx
}

func TestReporting_Catalog(t *testing.T) {
	mod, _, ctx := setupReportingTest(t)

	catalog := mod.Service.GetCatalog(ctx)
	if len(catalog) != 16 {
		t.Fatalf("expected 16 canonical reports in catalog, got %d", len(catalog))
	}

	codes := make(map[string]bool)
	for _, item := range catalog {
		if item.Code == "" || item.NameEn == "" || item.NameAr == "" || item.Endpoint == "" {
			t.Errorf("incomplete catalog item: %+v", item)
		}
		codes[item.Code] = true
	}

	expectedCodes := []string{
		"TRIAL_BALANCE", "GENERAL_LEDGER", "BALANCE_SHEET", "INCOME_STATEMENT",
		"CASH_FLOW", "AR_AGING", "AP_AGING", "CUSTOMER_STATEMENT", "SUPPLIER_STATEMENT",
		"INVENTORY_VALUATION", "SALES_REPORT", "PURCHASING_REPORT", "VAT_RETURN",
		"PAYROLL_SUMMARY", "GOSI_CONTRIBUTION", "BUDGET_VS_ACTUAL",
	}

	for _, code := range expectedCodes {
		if !codes[code] {
			t.Errorf("missing report code %s in catalog", code)
		}
	}
}

func TestReporting_FinancialStatements(t *testing.T) {
	mod, leID, ctx := setupReportingTest(t)

	filter := domain.ReportFilter{
		LegalEntityID: leID,
	}

	// 1. Trial Balance
	tb, err := mod.Service.GetTrialBalance(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get trial balance: %v", err)
	}
	if tb.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, tb.LegalEntityID)
	}
	if !tb.IsBalanced {
		t.Errorf("expected trial balance to be balanced, total debits=%s, credits=%s", tb.TotalDebits, tb.TotalCredits)
	}

	// 2. General Ledger
	gl, err := mod.Service.GetGeneralLedger(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get general ledger: %v", err)
	}
	if gl.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, gl.LegalEntityID)
	}

	// 3. Balance Sheet
	bs, err := mod.Service.GetBalanceSheet(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get balance sheet: %v", err)
	}
	if bs.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, bs.LegalEntityID)
	}
	if !bs.IsBalanced {
		t.Errorf("expected balance sheet to balance: assets=%s, liabilities+equity=%s", bs.TotalAssets, bs.TotalLiabEquity)
	}

	// 4. Income Statement
	now := time.Now()
	monthAgo := now.AddDate(0, -1, 0)
	filter.StartDate = &monthAgo
	filter.EndDate = &now
	is, err := mod.Service.GetIncomeStatement(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get income statement: %v", err)
	}
	expectedNet := is.TotalRevenue.Sub(is.TotalExpenses)
	if !is.NetIncome.Equal(expectedNet) {
		t.Errorf("expected net income %s, got %s", expectedNet, is.NetIncome)
	}

	// 5. Cash Flow
	cf, err := mod.Service.GetCashFlow(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get cash flow: %v", err)
	}
	if cf.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, cf.LegalEntityID)
	}
}

func TestReporting_CommercialAndAging(t *testing.T) {
	mod, leID, ctx := setupReportingTest(t)

	filter := domain.ReportFilter{
		LegalEntityID: leID,
	}

	// AR Aging
	ar, err := mod.Service.GetARAging(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get AR aging: %v", err)
	}
	if ar.ReportType != "AR" {
		t.Errorf("expected AR report type, got %s", ar.ReportType)
	}

	// AP Aging
	ap, err := mod.Service.GetAPAging(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get AP aging: %v", err)
	}
	if ap.ReportType != "AP" {
		t.Errorf("expected AP report type, got %s", ap.ReportType)
	}

	// Customer statement requires customer_id
	_, err = mod.Service.GetCustomerStatement(ctx, filter)
	if err == nil {
		t.Errorf("expected error when customer_id is missing, got nil")
	}

	// Supplier statement requires supplier_id
	_, err = mod.Service.GetSupplierStatement(ctx, filter)
	if err == nil {
		t.Errorf("expected error when supplier_id is missing, got nil")
	}
}

func TestReporting_OperationalReports(t *testing.T) {
	mod, leID, ctx := setupReportingTest(t)

	filter := domain.ReportFilter{
		LegalEntityID: leID,
	}

	// 1. Inventory Valuation
	iv, err := mod.Service.GetInventoryValuation(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get inventory valuation: %v", err)
	}
	if iv.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, iv.LegalEntityID)
	}

	// 2. Sales Report
	sr, err := mod.Service.GetSalesReport(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get sales report: %v", err)
	}
	if sr.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, sr.LegalEntityID)
	}

	// 3. Purchasing Report
	pr, err := mod.Service.GetPurchasingReport(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get purchasing report: %v", err)
	}
	if pr.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, pr.LegalEntityID)
	}

	// 4. VAT Return 15%
	vr, err := mod.Service.GetVATReturnReport(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get VAT return report: %v", err)
	}
	if len(vr.SalesBoxes) != 4 || len(vr.PurchaseBoxes) != 4 {
		t.Errorf("expected 4 sales boxes and 4 purchase boxes, got %d and %d", len(vr.SalesBoxes), len(vr.PurchaseBoxes))
	}
	expectedNetVAT := vr.TotalOutputVAT.Sub(vr.TotalInputVAT)
	if !vr.NetVATDueOrRefundable.Equal(expectedNetVAT) {
		t.Errorf("expected net VAT %s, got %s", expectedNetVAT, vr.NetVATDueOrRefundable)
	}

	// 5. Payroll Summary
	ps, err := mod.Service.GetPayrollSummary(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get payroll summary: %v", err)
	}
	if ps.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, ps.LegalEntityID)
	}

	// 6. GOSI Contribution
	gosiRpt, err := mod.Service.GetGOSIContribution(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get GOSI contribution report: %v", err)
	}
	if gosiRpt.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, gosiRpt.LegalEntityID)
	}

	// 7. Budget vs Actual
	bva, err := mod.Service.GetBudgetVsActual(ctx, filter)
	if err != nil {
		t.Fatalf("failed to get budget vs actual report: %v", err)
	}
	if bva.LegalEntityID != leID {
		t.Errorf("expected legal entity %s, got %s", leID, bva.LegalEntityID)
	}
}

func TestReporting_CSVExport(t *testing.T) {
	mod, leID, ctx := setupReportingTest(t)

	filter := domain.ReportFilter{
		LegalEntityID: leID,
	}

	reportsToExport := []string{
		"TRIAL_BALANCE", "GENERAL_LEDGER", "INVENTORY_VALUATION",
		"SALES_REPORT", "AR_AGING", "AP_AGING", "BUDGET_VS_ACTUAL",
	}

	for _, code := range reportsToExport {
		data, filename, err := mod.Service.ExportReportToCSV(ctx, code, filter)
		if err != nil {
			t.Fatalf("failed to export %s to CSV: %v", code, err)
		}
		if len(data) == 0 {
			t.Errorf("expected non-empty CSV data for %s", code)
		}
		if filename == "" {
			t.Errorf("expected non-empty filename for %s", code)
		}
	}
}

func TestReporting_HTTPEndpoints(t *testing.T) {
	mod, leID, _ := setupReportingTest(t)

	handler := mod.Handler.Routes(nil)

	endpoints := []string{
		"/catalog",
		"/trial-balance?legal_entity_id=" + leID.String(),
		"/general-ledger?legal_entity_id=" + leID.String(),
		"/balance-sheet?legal_entity_id=" + leID.String(),
		"/income-statement?legal_entity_id=" + leID.String(),
		"/cash-flow?legal_entity_id=" + leID.String(),
		"/ar-aging?legal_entity_id=" + leID.String(),
		"/ap-aging?legal_entity_id=" + leID.String(),
		"/inventory-valuation?legal_entity_id=" + leID.String(),
		"/sales?legal_entity_id=" + leID.String(),
		"/purchasing?legal_entity_id=" + leID.String(),
		"/vat-return?legal_entity_id=" + leID.String(),
		"/payroll-summary?legal_entity_id=" + leID.String(),
		"/gosi-summary?legal_entity_id=" + leID.String(),
		"/budget-vs-actual?legal_entity_id=" + leID.String(),
		"/trial-balance?legal_entity_id=" + leID.String() + "&export=csv",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("endpoint %s returned status %d: %s", ep, rec.Code, rec.Body.String())
		}
	}
}
