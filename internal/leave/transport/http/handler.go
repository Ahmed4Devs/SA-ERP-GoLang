package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"erp/internal/leave/domain"
	"erp/internal/leave/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LeaveHandler struct {
	svc ports.LeaveService
}

func NewLeaveHandler(svc ports.LeaveService) *LeaveHandler {
	return &LeaveHandler{svc: svc}
}

func (h *LeaveHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// Types & Policies
		protected.Get("/types", h.ListTypes)
		protected.Post("/types", h.CreateType)
		protected.Get("/policies", h.ListPolicies)
		protected.Post("/policies", h.DefinePolicy)

		// Balances
		protected.Get("/balances/employees/{employee_id}", h.GetEmployeeBalances)
		protected.Post("/balances/accrue", h.AccrueLeave)

		// Requests
		protected.Get("/requests", h.ListRequests)
		protected.Post("/requests", h.SubmitRequest)
		protected.Get("/requests/employees/{employee_id}", h.ListEmployeeRequests)
		protected.Post("/requests/{id}/approve", h.ApproveRequest)
		protected.Post("/requests/{id}/reject", h.RejectRequest)
	})

	return r
}

func (h *LeaveHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	types, err := h.svc.ListLeaveTypes(r.Context(), legalEntityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, types)
}

func (h *LeaveHandler) CreateType(w http.ResponseWriter, r *http.Request) {
	var lt domain.LeaveType
	if err := json.NewDecoder(r.Body).Decode(&lt); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.CreateLeaveType(r.Context(), &lt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *LeaveHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	policies, err := h.svc.ListPolicies(r.Context(), legalEntityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (h *LeaveHandler) DefinePolicy(w http.ResponseWriter, r *http.Request) {
	var p domain.LeavePolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.DefinePolicy(r.Context(), &p)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *LeaveHandler) GetEmployeeBalances(w http.ResponseWriter, r *http.Request) {
	empID, err := uuid.Parse(chi.URLParam(r, "employee_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	year := int32(time.Now().Year())
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = int32(y)
		}
	}

	balances, err := h.svc.GetEmployeeBalances(r.Context(), empID, year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, balances)
}

type AccrueRequest struct {
	EmployeeID  string          `json:"employee_id"`
	LeaveTypeID string          `json:"leave_type_id"`
	FiscalYear  int32           `json:"fiscal_year"`
	Days        decimal.Decimal `json:"days"`
	Reason      string          `json:"reason"`
}

func (h *LeaveHandler) AccrueLeave(w http.ResponseWriter, r *http.Request) {
	var req AccrueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	empID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	ltID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid leave_type_id")
		return
	}

	updated, err := h.svc.AccrueLeave(r.Context(), empID, ltID, req.FiscalYear, req.Days, req.Reason)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *LeaveHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	requests, err := h.svc.ListRequests(r.Context(), legalEntityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, requests)
}

func (h *LeaveHandler) ListEmployeeRequests(w http.ResponseWriter, r *http.Request) {
	empID, err := uuid.Parse(chi.URLParam(r, "employee_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	requests, err := h.svc.ListEmployeeRequests(r.Context(), empID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, requests)
}

func (h *LeaveHandler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	var req domain.LeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.SubmitLeaveRequest(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

type ReviewLeaveRequest struct {
	Notes string `json:"notes"`
}

func (h *LeaveHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request id")
		return
	}
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	var rev ReviewLeaveRequest
	_ = json.NewDecoder(r.Body).Decode(&rev)

	approverID := uuid.Nil
	approved, err := h.svc.ApproveLeaveRequest(r.Context(), legalEntityID, id, approverID, rev.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, approved)
}

func (h *LeaveHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request id")
		return
	}
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	var rev ReviewLeaveRequest
	_ = json.NewDecoder(r.Body).Decode(&rev)

	approverID := uuid.Nil
	rejected, err := h.svc.RejectLeaveRequest(r.Context(), legalEntityID, id, approverID, rev.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rejected)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
