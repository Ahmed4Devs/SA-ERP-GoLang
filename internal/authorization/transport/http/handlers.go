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

	// Self-service authorization endpoints need authentication only. They are
	// used by the web client to discover the caller's scopes and permissions.
	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}
		r.Get("/assignments/me", h.ListMyAssignments)
		r.Get("/permissions/me", h.GetMyPermissions)
		r.Post("/check", h.CheckPermission)
	})

	// Role catalog and assignment mutations are administrative operations.
	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}
		// Role administration is deployment-wide. Never infer authority from
		// the target entity carried by the assignment request.
		r.Group(func(read chi.Router) {
			read.Use(RequirePermission(h.authzService, "authz.view", nil))
			read.Get("/roles", h.ListRoles)
			read.Get("/permissions", h.ListPermissions)
			read.Get("/assignments", h.ListAssignments)
		})
		r.Group(func(write chi.Router) {
			write.Use(RequirePermission(h.authzService, "authz.manage", nil))
			write.Post("/assignments", h.AssignRole)
			write.Post("/assignments/revoke", h.RevokeRole)
		})
	})

	return r
}

func (h *AuthorizationHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil || id == uuid.Nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Valid user_id is required")
		return
	}
	assignments, err := h.authzService.ListUserAssignments(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list assignments")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": assignments})
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
		} else {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid legal_entity_id")
			return
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
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil && parsed != uuid.Nil {
			leID = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid legal_entity_id")
			return
		}
	}

	var brID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		parsed, err := uuid.Parse(*req.BranchID)
		if err != nil || parsed == uuid.Nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid branch_id")
			return
		}
		brID = &parsed
	}
	decision, err := h.authzService.Authorize(r.Context(), domain.AuthorizationRequest{
		UserID:        user.ID,
		Permission:    req.Permission,
		LegalEntityID: leID,
		BranchID:      brID,
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
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil && parsed != uuid.Nil {
			leID = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid legal_entity_id")
			return
		}
	}

	var brID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		if parsed, err := uuid.Parse(*req.BranchID); err == nil && parsed != uuid.Nil {
			brID = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid branch_id")
			return
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
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil && parsed != uuid.Nil {
			leID = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid legal_entity_id")
			return
		}
	}

	var brID *uuid.UUID
	if req.BranchID != nil && *req.BranchID != "" {
		if parsed, err := uuid.Parse(*req.BranchID); err == nil && parsed != uuid.Nil {
			brID = &parsed
		} else {
			writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid branch_id")
			return
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
