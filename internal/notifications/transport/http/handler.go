package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"erp/internal/notifications/domain"
	"erp/internal/notifications/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NotificationHandler handles HTTP routes for the notifications module.
type NotificationHandler struct {
	svc ports.Service
}

// NewNotificationHandler creates a new NotificationHandler instance.
func NewNotificationHandler(svc ports.Service) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// Routes configures all notification routes on the Chi router.
func (h *NotificationHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		protected.Get("/", h.List)
		protected.Get("/unread-count", h.UnreadCount)
		protected.Post("/send", h.Send)
		protected.Get("/summary", h.Summary)
		protected.Get("/{id}", h.Get)
		protected.Patch("/{id}/read", h.MarkAsRead)
		protected.Post("/mark-all-read", h.MarkAllAsRead)
		protected.Delete("/{id}", h.Delete)
	})

	return r
}

// List returns notifications matching filter criteria.
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	filter := domain.NotificationFilter{
		LegalEntityID: leID,
		Limit:         50,
		Offset:        0,
	}

	if userIDStr := r.URL.Query().Get("recipient_user_id"); userIDStr != "" {
		if parsed, err := uuid.Parse(userIDStr); err == nil {
			filter.RecipientUserID = &parsed
		}
	}

	if catStr := r.URL.Query().Get("category"); catStr != "" {
		cat := domain.NotificationCategory(strings.ToUpper(catStr))
		filter.Category = &cat
	}

	if stStr := r.URL.Query().Get("status"); stStr != "" {
		st := domain.NotificationStatus(strings.ToUpper(stStr))
		filter.Status = &st
	}

	if chStr := r.URL.Query().Get("channel"); chStr != "" {
		ch := domain.NotificationChannel(strings.ToUpper(chStr))
		filter.Channel = &ch
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			filter.Limit = val
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	notifications, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":   notifications,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// UnreadCount returns the count of unread notifications.
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	var recipientIDPtr *uuid.UUID
	if userIDStr := r.URL.Query().Get("recipient_user_id"); userIDStr != "" {
		if parsed, err := uuid.Parse(userIDStr); err == nil {
			recipientIDPtr = &parsed
		}
	}

	count, err := h.svc.GetUnreadCount(r.Context(), leID, recipientIDPtr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}

// Send dispatches a new notification.
func (h *NotificationHandler) Send(w http.ResponseWriter, r *http.Request) {
	var params domain.SendNotificationParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	notification, err := h.svc.Send(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, notification)
}

// Get returns a single notification.
func (h *NotificationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	notification, err := h.svc.Get(r.Context(), leID, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, notification)
}

// MarkAsRead marks a notification as read.
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), leID, id); err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "marked_as_read"})
}

// MarkAllAsRead marks all notifications for a recipient as read.
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	var recipientIDPtr *uuid.UUID
	if userIDStr := r.URL.Query().Get("recipient_user_id"); userIDStr != "" {
		if parsed, err := uuid.Parse(userIDStr); err == nil {
			recipientIDPtr = &parsed
		}
	}

	if err := h.svc.MarkAllAsRead(r.Context(), leID, recipientIDPtr); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "all_marked_as_read"})
}

// Delete removes a notification.
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	if err := h.svc.Delete(r.Context(), leID, id); err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Summary returns high-level notification analytics.
func (h *NotificationHandler) Summary(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "valid legal_entity_id is required")
		return
	}

	var recipientIDPtr *uuid.UUID
	if userIDStr := r.URL.Query().Get("recipient_user_id"); userIDStr != "" {
		if parsed, err := uuid.Parse(userIDStr); err == nil {
			recipientIDPtr = &parsed
		}
	}

	summary, err := h.svc.GetSummary(r.Context(), leID, recipientIDPtr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
