package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/dashboard/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DashboardHandler struct {
	svc ports.DashboardService
}

func NewDashboardHandler(svc ports.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		protected.Get("/", h.GetDashboard)
		protected.Get("/summary", h.GetDashboard)
	})

	return r
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	leStr := q.Get("legal_entity_id")
	if leStr == "" {
		leStr = r.Header.Get("X-Legal-Entity-ID")
	}

	var leID uuid.UUID
	if leStr != "" {
		parsed, err := uuid.Parse(leStr)
		if err == nil {
			leID = parsed
		}
	}

	data, err := h.svc.GetDashboardData(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, data)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
