package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/budgeting/domain"
	"erp/internal/budgeting/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetingHandler struct {
	svc ports.Service
}

func NewBudgetingHandler(svc ports.Service) *BudgetingHandler {
	return &BudgetingHandler{svc: svc}
}

func (h *BudgetingHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// Budgets Master
		protected.Get("/budgets", h.ListBudgets)
		protected.Post("/budgets", h.CreateBudget)
		protected.Get("/budgets/{id}", h.GetBudget)
		protected.Post("/budgets/{id}/submit", h.SubmitBudget)
		protected.Post("/budgets/{id}/approve", h.ApproveBudget)

		// Versions
		protected.Get("/budgets/{id}/versions", h.ListVersions)
		protected.Post("/budgets/{id}/versions", h.CreateVersion)

		// Lines
		protected.Get("/versions/{versionId}/lines", h.ListBudgetLines)
		protected.Post("/lines", h.AddBudgetLine)

		// Transfers & Reallocations
		protected.Get("/transfers", h.ListTransfers)
		protected.Post("/transfers", h.TransferBudget)

		// Analytics & Variance
		protected.Get("/budgets/{id}/vs-actual", h.GetBudgetVsActual)
		protected.Get("/summary", h.GetBudgetSummary)
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
// Budget Master Endpoints
// ----------------------------------------------------------------------

func (h *BudgetingHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	budgets, err := h.svc.ListBudgets(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, budgets)
}

type CreateBudgetPayload struct {
	LegalEntityID string `json:"legal_entity_id"`
	Code          string `json:"code"`
	NameEn        string `json:"name_en"`
	NameAr        string `json:"name_ar"`
	DescriptionEn string `json:"description_en"`
	DescriptionAr string `json:"description_ar"`
	FiscalYearID  string `json:"fiscal_year_id"`
	CurrencyCode  string `json:"currency_code"`
	BudgetType    string `json:"budget_type"`
}

func (h *BudgetingHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var payload CreateBudgetPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(payload.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	fyID, err := uuid.Parse(payload.FiscalYearID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid fiscal_year_id")
		return
	}

	req := ports.CreateBudgetRequest{
		LegalEntityID: leID,
		Code:          payload.Code,
		NameEn:        payload.NameEn,
		NameAr:        payload.NameAr,
		DescriptionEn: payload.DescriptionEn,
		DescriptionAr: payload.DescriptionAr,
		FiscalYearID:  fyID,
		CurrencyCode:  payload.CurrencyCode,
		BudgetType:    domain.BudgetType(payload.BudgetType),
	}

	created, err := h.svc.CreateBudget(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *BudgetingHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	budget, err := h.svc.GetBudget(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, budget)
}

func (h *BudgetingHandler) SubmitBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	updated, err := h.svc.SubmitBudget(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

type ApproveBudgetPayload struct {
	ApproverID string `json:"approver_id"`
}

func (h *BudgetingHandler) ApproveBudget(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	var payload ApproveBudgetPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)

	var appID uuid.UUID
	if payload.ApproverID != "" {
		appID, _ = uuid.Parse(payload.ApproverID)
	}
	if appID == uuid.Nil {
		appID = uuid.New()
	}

	updated, err := h.svc.ApproveBudget(r.Context(), id, leID, appID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// ----------------------------------------------------------------------
// Versions & Lines
// ----------------------------------------------------------------------

func (h *BudgetingHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	versions, err := h.svc.ListVersions(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, versions)
}

type CreateVersionPayload struct {
	VersionName string `json:"version_name"`
	Notes       string `json:"notes"`
}

func (h *BudgetingHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	var payload CreateVersionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ver, err := h.svc.CreateVersion(r.Context(), ports.CreateVersionRequest{
		BudgetID:    id,
		VersionName: payload.VersionName,
		Notes:       payload.Notes,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ver)
}

func (h *BudgetingHandler) ListBudgetLines(w http.ResponseWriter, r *http.Request) {
	versionID, err := uuid.Parse(chi.URLParam(r, "versionId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid version id")
		return
	}

	lines, err := h.svc.ListBudgetLines(r.Context(), versionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, lines)
}

type AddBudgetLinePayload struct {
	BudgetID          string          `json:"budget_id"`
	BudgetVersionID   string          `json:"budget_version_id"`
	LegalEntityID     string          `json:"legal_entity_id"`
	AccountID         string          `json:"account_id"`
	DepartmentID      *string         `json:"department_id,omitempty"`
	CostCenterID      *string         `json:"cost_center_id,omitempty"`
	ProjectCode       string          `json:"project_code,omitempty"`
	FiscalPeriodID    *string         `json:"fiscal_period_id,omitempty"`
	Amount            decimal.Decimal `json:"amount"`
	LineDescriptionEn string          `json:"line_description_en"`
	LineDescriptionAr string          `json:"line_description_ar"`
}

func (h *BudgetingHandler) AddBudgetLine(w http.ResponseWriter, r *http.Request) {
	var payload AddBudgetLinePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bID, err := uuid.Parse(payload.BudgetID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget_id")
		return
	}
	verID, err := uuid.Parse(payload.BudgetVersionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget_version_id")
		return
	}
	leID, err := uuid.Parse(payload.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	accID, err := uuid.Parse(payload.AccountID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account_id")
		return
	}

	var deptID *uuid.UUID
	if payload.DepartmentID != nil && *payload.DepartmentID != "" {
		parsed, err := uuid.Parse(*payload.DepartmentID)
		if err == nil {
			deptID = &parsed
		}
	}

	var ccID *uuid.UUID
	if payload.CostCenterID != nil && *payload.CostCenterID != "" {
		parsed, err := uuid.Parse(*payload.CostCenterID)
		if err == nil {
			ccID = &parsed
		}
	}

	var periodID *uuid.UUID
	if payload.FiscalPeriodID != nil && *payload.FiscalPeriodID != "" {
		parsed, err := uuid.Parse(*payload.FiscalPeriodID)
		if err == nil {
			periodID = &parsed
		}
	}

	line, err := h.svc.AddBudgetLine(r.Context(), ports.AddBudgetLineRequest{
		BudgetID:          bID,
		BudgetVersionID:   verID,
		LegalEntityID:     leID,
		AccountID:         accID,
		DepartmentID:      deptID,
		CostCenterID:      ccID,
		ProjectCode:       payload.ProjectCode,
		FiscalPeriodID:    periodID,
		Amount:            payload.Amount,
		LineDescriptionEn: payload.LineDescriptionEn,
		LineDescriptionAr: payload.LineDescriptionAr,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, line)
}

// ----------------------------------------------------------------------
// Transfers & Reallocations
// ----------------------------------------------------------------------

type TransferPayload struct {
	LegalEntityID string          `json:"legal_entity_id"`
	SourceLineID  string          `json:"source_line_id"`
	TargetLineID  string          `json:"target_line_id"`
	Amount        decimal.Decimal `json:"amount"`
	Reason        string          `json:"reason"`
}

func (h *BudgetingHandler) TransferBudget(w http.ResponseWriter, r *http.Request) {
	var payload TransferPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(payload.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	srcID, err := uuid.Parse(payload.SourceLineID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid source_line_id")
		return
	}
	tgtID, err := uuid.Parse(payload.TargetLineID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid target_line_id")
		return
	}

	transfer, err := h.svc.TransferBudget(r.Context(), ports.TransferBudgetRequest{
		LegalEntityID: leID,
		SourceLineID:  srcID,
		TargetLineID:  tgtID,
		Amount:        payload.Amount,
		Reason:        payload.Reason,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, transfer)
}

func (h *BudgetingHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	transfers, err := h.svc.ListTransfers(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, transfers)
}

// ----------------------------------------------------------------------
// Analytics & Summaries
// ----------------------------------------------------------------------

func (h *BudgetingHandler) GetBudgetVsActual(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid budget id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	items, err := h.svc.GetBudgetVsActual(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *BudgetingHandler) GetBudgetSummary(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	summary, err := h.svc.GetBudgetSummary(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
