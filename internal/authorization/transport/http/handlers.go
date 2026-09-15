package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/authorization/domain"
	"erp/internal/authorization/ports"
	iamhttp "erp/internal/iam/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthorizationHandler struct {
	authzService ports.AuthorizationService
}

func NewAuthorizationHandler(authzService ports.AuthorizationService) *AuthorizationHandler {
	return &AuthorizationHandler{authzService: authzService}
}

func (h *AuthorizationHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		r.Get("/roles", h.ListRoles)
		r.Get("/permissions", h.ListPermissions)
		r.Get("/assignments/me", h.ListMyAssignments)
		r.Get("/permissions/me", h.GetMyPermissions)
		r.Post("/check", h.CheckPermission)
		r.Post("/assignments", h.AssignRole)
		r.Post("/assignments/revoke", h.RevokeRole)
	})

	return r
}

func (h *AuthorizationHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.authzService.ListRoles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": roles})
}

func (h *AuthorizationHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.authzService.ListPermissions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": perms})
}

func (h *AuthorizationHandler) ListMyAssignments(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	assignments, err := h.authzService.ListUserAssignments(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": assignments})
}

func (h *AuthorizationHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var leID *uuid.UUID
	if entityStr := r.URL.Query().Get("legal_entity_id"); entityStr != "" {
		if parsed, err := uuid.Parse(entityStr); err == nil {
			leID = &parsed
		}
	}

	perms, err := h.authzService.GetUserPermissions(r.Context(), user.ID, leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": perms})
}

func (h *AuthorizationHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	var leID *uuid.UUID
	if req.LegalEntityID != nil && *req.LegalEntityID != "" {
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil {
			leID = &parsed
		}
	}

	decision, err := h.authzService.Authorize(r.Context(), domain.AuthorizationRequest{
		UserID:        user.ID,
		Permission:    req.Permission,
		LegalEntityID: leID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, decision)
}

func (h *AuthorizationHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid user_id UUID")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid role_id UUID")
		return
	}

	var leID *uuid.UUID
	if req.LegalEntityID != nil && *req.LegalEntityID != "" {
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil {
			leID = &parsed
		}
	}

	var brID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		if parsed, err := uuid.Parse(*req.BranchID); err == nil {
			brID = &parsed
		}
	}

	assignment, err := h.authzService.AssignRole(r.Context(), userID, roleID, leID, brID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": assignment})
}

func (h *AuthorizationHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	var req RevokeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid user_id UUID")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid role_id UUID")
		return
	}

	var leID *uuid.UUID
	if req.LegalEntityID != nil && *req.LegalEntityID != "" {
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil {
			leID = &parsed
		}
	}

	var brID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		if parsed, err := uuid.Parse(*req.BranchID); err == nil {
			brID = &parsed
		}
	}

	err = h.authzService.RevokeRole(r.Context(), userID, roleID, leID, brID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "revoked"})
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
