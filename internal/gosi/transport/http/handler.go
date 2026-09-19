package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"erp/internal/gosi/domain"
	"erp/internal/gosi/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GosiHandler struct {
	svc ports.Service
}

func NewGosiHandler(svc ports.Service) *GosiHandler {
	return &GosiHandler{svc: svc}
}

func (h *GosiHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// Rules
		protected.Get("/rules", h.ListRules)
		protected.Get("/rules/active", h.GetActiveRules)

		// Profiles
		protected.Get("/profiles", h.ListProfiles)
		protected.Post("/profiles", h.RegisterEmployee)

		// Calculations
		protected.Post("/calculate", h.Calculate)
		protected.Post("/calculate/monthly", h.RunMonthlyCalculation)
		protected.Get("/calculations", h.ListCalculations)
		protected.Get("/calculations/{id}", h.GetCalculationDetails)

		// Export Batches
		protected.Post("/export", h.CreateExportBatch)
		protected.Get("/export", h.ListExportBatches)

		// Dashboard
		protected.Get("/dashboard", h.GetDashboardSummary)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ----------------------------------------------------------------------
// 1. Rules
// ----------------------------------------------------------------------

func (h *GosiHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	rules, err := h.svc.ListContributionRules(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (h *GosiHandler) GetActiveRules(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	asOf := time.Now()
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			asOf = t
		}
	}

	versions, err := h.svc.GetActiveRuleVersions(r.Context(), leID, asOf)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

// ----------------------------------------------------------------------
// 2. Profiles
// ----------------------------------------------------------------------

func (h *GosiHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	profiles, err := h.svc.ListInsuranceProfiles(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, profiles)
}

type RegisterEmployeeRequest struct {
	LegalEntityID             string           `json:"legal_entity_id"`
	EmployeeID                string           `json:"employee_id"`
	GosiNumber                string           `json:"gosi_number"`
	NationalityScope          string           `json:"nationality_scope"` // SAUDI or NON_SAUDI
	ContributionBasisOverride *decimal.Decimal `json:"contribution_basis_override"`
	Notes                     string           `json:"notes"`
}

func (h *GosiHandler) RegisterEmployee(w http.ResponseWriter, r *http.Request) {
	var req RegisterEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	scope := req.NationalityScope
	if scope == "" {
		scope = "SAUDI"
	}

	profile := domain.InsuranceProfile{
		ID:                        uuid.New(),
		LegalEntityID:             leID,
		EmployeeID:                empID,
		GosiNumber:                req.GosiNumber,
		NationalityScope:          scope,
		ContributionBasisOverride: req.ContributionBasisOverride,
		Notes:                     req.Notes,
		IsActive:                  true,
	}

	saved, err := h.svc.RegisterEmployee(r.Context(), profile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, saved)
}

// ----------------------------------------------------------------------
// 3. Calculations
// ----------------------------------------------------------------------

type CalculateRequest struct {
	LegalEntityID     string          `json:"legal_entity_id"`
	EmployeeID        string          `json:"employee_id"`
	PayrollDate       string          `json:"payroll_date"`
	NationalityCode   string          `json:"nationality_code"`
	ContributionBasis decimal.Decimal `json:"contribution_basis"`
}

func (h *GosiHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	empID, _ := uuid.Parse(req.EmployeeID)

	payrollDate := time.Now()
	if req.PayrollDate != "" {
		if t, err := time.Parse("2006-01-02", req.PayrollDate); err == nil {
			payrollDate = t
		}
	}

	results, err := h.svc.Calculate(r.Context(), domain.ContributionInput{
		LegalEntityID:     leID,
		EmployeeID:        empID,
		PayrollDate:       payrollDate,
		NationalityCode:   req.NationalityCode,
		ContributionBasis: req.ContributionBasis,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

type RunMonthlyCalculationRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	Year          int    `json:"year"`
	Month         int    `json:"month"`
}

func (h *GosiHandler) RunMonthlyCalculation(w http.ResponseWriter, r *http.Request) {
	var req RunMonthlyCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	if req.Year <= 2000 || req.Month < 1 || req.Month > 12 {
		writeError(w, http.StatusBadRequest, "invalid year or month")
		return
	}

	calcs, err := h.svc.RunMonthlyCalculation(r.Context(), leID, req.Year, req.Month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calcs)
}

func (h *GosiHandler) ListCalculations(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	calcs, err := h.svc.ListCalculationsByPeriod(r.Context(), leID, year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calcs)
}

func (h *GosiHandler) GetCalculationDetails(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id query parameter is required")
		return
	}
	idStr := chi.URLParam(r, "id")
	calcID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid calculation id")
		return
	}

	calc, lines, err := h.svc.GetCalculationDetails(r.Context(), leID, calcID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"calculation": calc,
		"lines":       lines,
	})
}

// ----------------------------------------------------------------------
// 4. Export Batches
// ----------------------------------------------------------------------

type CreateExportBatchRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	Year          int    `json:"year"`
	Month         int    `json:"month"`
}

func (h *GosiHandler) CreateExportBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateExportBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	batch, err := h.svc.CreateExportBatch(r.Context(), leID, req.Year, req.Month, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, batch)
}

func (h *GosiHandler) ListExportBatches(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	batches, err := h.svc.ListExportBatches(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, batches)
}

// ----------------------------------------------------------------------
// 5. Dashboard Summary
// ----------------------------------------------------------------------

func (h *GosiHandler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}

	summary, err := h.svc.GetDashboardSummary(r.Context(), leID, year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
