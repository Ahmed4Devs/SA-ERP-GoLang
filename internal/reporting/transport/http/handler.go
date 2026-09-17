package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"erp/internal/reporting/domain"
	"erp/internal/reporting/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReportingHandler struct {
	svc ports.ReportingService
}

func NewReportingHandler(svc ports.ReportingService) *ReportingHandler {
	return &ReportingHandler{svc: svc}
}

func (h *ReportingHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		protected.Get("/catalog", h.GetCatalog)
		protected.Get("/trial-balance", h.GetTrialBalance)
		protected.Get("/general-ledger", h.GetGeneralLedger)
		protected.Get("/balance-sheet", h.GetBalanceSheet)
		protected.Get("/income-statement", h.GetIncomeStatement)
		protected.Get("/cash-flow", h.GetCashFlow)
		protected.Get("/ar-aging", h.GetARAging)
		protected.Get("/ap-aging", h.GetAPAging)
		protected.Get("/customer-statement", h.GetCustomerStatement)
		protected.Get("/supplier-statement", h.GetSupplierStatement)
		protected.Get("/inventory-valuation", h.GetInventoryValuation)
		protected.Get("/sales", h.GetSalesReport)
		protected.Get("/purchasing", h.GetPurchasingReport)
		protected.Get("/vat-return", h.GetVATReturnReport)
		protected.Get("/payroll-summary", h.GetPayrollSummary)
		protected.Get("/gosi-summary", h.GetGOSIContribution)
		protected.Get("/budget-vs-actual", h.GetBudgetVsActual)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseFilter(r *http.Request) domain.ReportFilter {
	q := r.URL.Query()
	f := domain.ReportFilter{}

	// Legal Entity ID from query or header
	leStr := q.Get("legal_entity_id")
	if leStr == "" {
		leStr = r.Header.Get("X-Legal-Entity-ID")
	}
	if leID, err := uuid.Parse(leStr); err == nil {
		f.LegalEntityID = leID
	}

	// Dates
	if sd := q.Get("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			f.StartDate = &t
		}
	}
	if ed := q.Get("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			f.EndDate = &t
		}
	}
	if aod := q.Get("as_of_date"); aod != "" {
		if t, err := time.Parse("2006-01-02", aod); err == nil {
			f.AsOfDate = &t
		}
	}

	// UUID references
	if acc := q.Get("account_id"); acc != "" {
		if id, err := uuid.Parse(acc); err == nil {
			f.AccountID = &id
		}
	}
	if cust := q.Get("customer_id"); cust != "" {
		if id, err := uuid.Parse(cust); err == nil {
			f.CustomerID = &id
		}
	}
	if sup := q.Get("supplier_id"); sup != "" {
		if id, err := uuid.Parse(sup); err == nil {
			f.SupplierID = &id
		}
	}
	if wh := q.Get("warehouse_id"); wh != "" {
		if id, err := uuid.Parse(wh); err == nil {
			f.WarehouseID = &id
		}
	}
	if b := q.Get("budget_id"); b != "" {
		if id, err := uuid.Parse(b); err == nil {
			f.BudgetID = &id
		}
	}

	// Status, Year, Month
	if st := q.Get("status"); st != "" {
		f.Status = &st
	}
	if yStr := q.Get("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil {
			f.Year = &y
		}
	}
	if mStr := q.Get("month"); mStr != "" {
		if m, err := strconv.Atoi(mStr); err == nil {
			f.Month = &m
		}
	}

	return f
}

func (h *ReportingHandler) maybeExportCSV(w http.ResponseWriter, r *http.Request, code string, f domain.ReportFilter) bool {
	if r.URL.Query().Get("export") == "csv" {
		data, filename, err := h.svc.ExportReportToCSV(r.Context(), code, f)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return true
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		return true
	}
	return false
}

func (h *ReportingHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	catalog := h.svc.GetCatalog(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{"catalog": catalog})
}

func (h *ReportingHandler) GetTrialBalance(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "TRIAL_BALANCE", f) {
		return
	}
	report, err := h.svc.GetTrialBalance(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetGeneralLedger(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "GENERAL_LEDGER", f) {
		return
	}
	report, err := h.svc.GetGeneralLedger(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetBalanceSheet(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "BALANCE_SHEET", f) {
		return
	}
	report, err := h.svc.GetBalanceSheet(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetIncomeStatement(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "INCOME_STATEMENT", f) {
		return
	}
	report, err := h.svc.GetIncomeStatement(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetCashFlow(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "CASH_FLOW", f) {
		return
	}
	report, err := h.svc.GetCashFlow(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetARAging(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "AR_AGING", f) {
		return
	}
	report, err := h.svc.GetARAging(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetAPAging(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "AP_AGING", f) {
		return
	}
	report, err := h.svc.GetAPAging(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetCustomerStatement(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "CUSTOMER_STATEMENT", f) {
		return
	}
	report, err := h.svc.GetCustomerStatement(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetSupplierStatement(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "SUPPLIER_STATEMENT", f) {
		return
	}
	report, err := h.svc.GetSupplierStatement(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetInventoryValuation(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "INVENTORY_VALUATION", f) {
		return
	}
	report, err := h.svc.GetInventoryValuation(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "SALES_REPORT", f) {
		return
	}
	report, err := h.svc.GetSalesReport(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetPurchasingReport(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "PURCHASING_REPORT", f) {
		return
	}
	report, err := h.svc.GetPurchasingReport(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetVATReturnReport(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "VAT_RETURN", f) {
		return
	}
	report, err := h.svc.GetVATReturnReport(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetPayrollSummary(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "PAYROLL_SUMMARY", f) {
		return
	}
	report, err := h.svc.GetPayrollSummary(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetGOSIContribution(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "GOSI_CONTRIBUTION", f) {
		return
	}
	report, err := h.svc.GetGOSIContribution(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (h *ReportingHandler) GetBudgetVsActual(w http.ResponseWriter, r *http.Request) {
	f := parseFilter(r)
	if h.maybeExportCSV(w, r, "BUDGET_VS_ACTUAL", f) {
		return
	}
	report, err := h.svc.GetBudgetVsActual(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}
