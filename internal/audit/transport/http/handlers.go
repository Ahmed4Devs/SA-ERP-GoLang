package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"erp/internal/audit/domain"
	"erp/internal/audit/ports"
	iamhttp "erp/internal/iam/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuditHandler struct {
	auditService ports.AuditService
}

func NewAuditHandler(auditService ports.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

func (h *AuditHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(AuditAccessMiddleware(h.auditService))

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		r.Get("/events", h.ListEvents)
		r.Get("/events/{id}", h.GetEvent)
		r.Get("/verify", h.VerifyChainIntegrity)
		r.Get("/security-events", h.ListSecurityEvents)
		r.Post("/events", h.RecordEvent)
	})

	return r
}

func (h *AuditHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := domain.AuditFilter{
		Action:       q.Get("action"),
		ResourceType: q.Get("resource_type"),
		ResourceID:   q.Get("resource_id"),
		Limit:        50,
		Offset:       0,
	}

	if leStr := q.Get("legal_entity_id"); leStr != "" {
		if parsed, err := uuid.Parse(leStr); err == nil {
			filter.LegalEntityID = &parsed
		}
	}

	if uStr := q.Get("user_id"); uStr != "" {
		if parsed, err := uuid.Parse(uStr); err == nil {
			filter.UserID = &parsed
		}
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	if fromStr := q.Get("from_date"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.FromDate = &t
		}
	}

	if toStr := q.Get("to_date"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ToDate = &t
		}
	}

	events, total, err := h.auditService.SearchEvents(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": events,
		"meta": map[string]any{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

func (h *AuditHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid audit event UUID")
		return
	}

	event, err := h.auditService.GetEvent(r.Context(), id)
	if err != nil {
		if err == domain.ErrEventNotFound {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "audit event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": event})
}

func (h *AuditHandler) VerifyChainIntegrity(w http.ResponseWriter, r *http.Request) {
	result, err := h.auditService.VerifyChainIntegrity(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTEGRITY_CHECK_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

func (h *AuditHandler) ListSecurityEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var userID *uuid.UUID
	if uStr := q.Get("user_id"); uStr != "" {
		if parsed, err := uuid.Parse(uStr); err == nil {
			userID = &parsed
		}
	}

	limit := int32(50)
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	offset := int32(0)
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	events, err := h.auditService.ListSecurityEvents(
		r.Context(),
		userID,
		q.Get("event_type"),
		q.Get("status"),
		limit,
		offset,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": events})
}

func (h *AuditHandler) RecordEvent(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req RecordEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
		return
	}

	var leID *uuid.UUID
	if req.LegalEntityID != nil && *req.LegalEntityID != "" {
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil {
			leID = &parsed
		}
	}

	payloadBytes, _ := json.Marshal(req.Payload)

	event, err := h.auditService.RecordEvent(r.Context(), ports.RecordEventCommand{
		UserID:        &user.ID,
		LegalEntityID: leID,
		Action:        req.Action,
		ResourceType:  req.ResourceType,
		ResourceID:    req.ResourceID,
		Payload:       payloadBytes,
		IPAddress:     r.RemoteAddr,
		UserAgent:     r.UserAgent(),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": event})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}
