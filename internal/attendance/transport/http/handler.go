package http

import (
	"encoding/json"
	"net/http"
	"time"

	"erp/internal/attendance/domain"
	"erp/internal/attendance/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	svc ports.AttendanceService
}

func NewAttendanceHandler(svc ports.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{svc: svc}
}

func (h *AttendanceHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// Shifts
		protected.Get("/shifts", h.ListShifts)
		protected.Post("/shifts", h.CreateShift)
		protected.Post("/shifts/assign", h.AssignShift)

		// Punch In/Out
		protected.Post("/clock-in", h.ClockIn)
		protected.Post("/clock-out", h.ClockOut)
		protected.Get("/daily", h.GetDailyRecords)
		protected.Get("/employees/{employee_id}", h.GetEmployeeRecords)

		// Overtime
		protected.Get("/overtime", h.ListOvertime)
		protected.Post("/overtime", h.SubmitOvertime)
		protected.Post("/overtime/{id}/approve", h.ApproveOvertime)
		protected.Post("/overtime/{id}/reject", h.RejectOvertime)

		// Adjustments
		protected.Post("/adjustments", h.SubmitAdjustment)
		protected.Post("/adjustments/{id}/approve", h.ApproveAdjustment)
	})

	return r
}

func (h *AttendanceHandler) ListShifts(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	shifts, err := h.svc.ListShifts(r.Context(), legalEntityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, shifts)
}

func (h *AttendanceHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var shift domain.Shift
	if err := json.NewDecoder(r.Body).Decode(&shift); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.CreateShift(r.Context(), &shift)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AttendanceHandler) AssignShift(w http.ResponseWriter, r *http.Request) {
	var assign domain.ShiftAssignment
	if err := json.NewDecoder(r.Body).Decode(&assign); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.AssignShift(r.Context(), &assign)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

type ClockRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	EmployeeID    string `json:"employee_id"`
	Method        string `json:"method"`
	Timestamp     string `json:"timestamp,omitempty"`
}

func (h *AttendanceHandler) ClockIn(w http.ResponseWriter, r *http.Request) {
	var req ClockRequest
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

	method := domain.CheckMethod(req.Method)
	if method == "" {
		method = domain.MethodManual
	}

	clockTime := time.Now()
	if req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			clockTime = t
		}
	}

	record, err := h.svc.ClockIn(r.Context(), leID, empID, method, clockTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (h *AttendanceHandler) ClockOut(w http.ResponseWriter, r *http.Request) {
	var req ClockRequest
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

	method := domain.CheckMethod(req.Method)
	if method == "" {
		method = domain.MethodManual
	}

	clockTime := time.Now()
	if req.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.Timestamp); err == nil {
			clockTime = t
		}
	}

	record, err := h.svc.ClockOut(r.Context(), leID, empID, method, clockTime)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (h *AttendanceHandler) GetDailyRecords(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = parsed
		}
	}

	records, err := h.svc.GetDailyRecords(r.Context(), legalEntityID, targetDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (h *AttendanceHandler) GetEmployeeRecords(w http.ResponseWriter, r *http.Request) {
	empID, err := uuid.Parse(chi.URLParam(r, "employee_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from := time.Now().AddDate(0, 0, -30)
	to := time.Now()
	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	records, err := h.svc.GetEmployeeRecords(r.Context(), empID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (h *AttendanceHandler) ListOvertime(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	list, err := h.svc.ListOvertime(r.Context(), legalEntityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *AttendanceHandler) SubmitOvertime(w http.ResponseWriter, r *http.Request) {
	var req domain.OvertimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.SubmitOvertime(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *AttendanceHandler) ApproveOvertime(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid overtime id")
		return
	}

	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id is required")
		return
	}

	approverID := uuid.Nil
	approved, err := h.svc.ApproveOvertime(r.Context(), legalEntityID, id, approverID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, approved)
}

func (h *AttendanceHandler) RejectOvertime(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid overtime id")
		return
	}

	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id is required")
		return
	}

	approverID := uuid.Nil
	rejected, err := h.svc.RejectOvertime(r.Context(), legalEntityID, id, approverID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rejected)
}

func (h *AttendanceHandler) SubmitAdjustment(w http.ResponseWriter, r *http.Request) {
	var adj domain.AttendanceAdjustment
	if err := json.NewDecoder(r.Body).Decode(&adj); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.svc.SubmitAdjustment(r.Context(), &adj)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

type ApproveAdjustmentReq struct {
	Notes string `json:"notes"`
}

func (h *AttendanceHandler) ApproveAdjustment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid adjustment id")
		return
	}

	var req ApproveAdjustmentReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	approverID := uuid.Nil
	approved, err := h.svc.ApproveAdjustment(r.Context(), id, approverID, req.Notes)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, approved)
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
