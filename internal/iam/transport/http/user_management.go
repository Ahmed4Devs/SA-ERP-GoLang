package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	authzdomain "erp/internal/authorization/domain"
	authzports "erp/internal/authorization/ports"
	"erp/internal/iam/application/services"
	"erp/internal/iam/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *AuthHandler) ConfigureUserManagement(users *services.UserManagementService, authorizer authzports.AuthorizationService) {
	h.users, h.authorizer = users, authorizer
}

// User identities belong to the deployment. Entity-scoped administrators must
// never be allowed to modify a global identity through a selected entity header.
func (h *AuthHandler) allowUserManagement(w http.ResponseWriter, r *http.Request, permission string) bool {
	user := UserFromContext(r.Context())
	if user == nil {
		writeUnauthorized(w, "Authentication required")
		return false
	}
	if h.users == nil || h.authorizer == nil {
		writeError(w, http.StatusServiceUnavailable, "UNAVAILABLE", "User management is unavailable")
		return false
	}
	decision, err := h.authorizer.Authorize(r.Context(), authzdomain.AuthorizationRequest{UserID: user.ID, Permission: permission})
	if err != nil || !decision.Allowed {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Global user administration permission is required")
		return false
	}
	w.Header().Set("Cache-Control", "no-store")
	return true
}

type CreateUserRequest struct {
	Username          string `json:"username"`
	Email             string `json:"email"`
	DisplayName       string `json:"display_name"`
	Password          string `json:"password"`
	PreferredLocale   string `json:"preferred_locale"`
	PreferredTimezone string `json:"preferred_timezone"`
}

type ChangeUserStatusRequest struct {
	Status          string `json:"status"`
	ExpectedVersion int64  `json:"expected_version"`
}

type ResetUserPasswordRequest struct {
	Password        string `json:"password"`
	ExpectedVersion int64  `json:"expected_version"`
}

func strictUserDecoder(r *http.Request) *json.Decoder {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder
}

func endOfUserRequest(decoder *json.Decoder) bool {
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.allowUserManagement(w, r, "iam.users.view") {
		return
	}
	limit, offset := int64(50), int64(0)
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.ParseInt(value, 10, 32)
	}
	if err != nil {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.ParseInt(value, 10, 32)
	}
	if err != nil {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	users, err := h.users.List(r.Context(), r.URL.Query().Get("search"), int32(limit), int32(offset))
	if err != nil {
		userManagementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": users})
}

func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !h.allowUserManagement(w, r, "iam.users.manage") {
		return
	}
	var req CreateUserRequest
	decoder := strictUserDecoder(r)
	if err := decoder.Decode(&req); err != nil || !endOfUserRequest(decoder) {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	user, err := h.users.Create(r.Context(), UserFromContext(r.Context()).ID, services.CreateUserCommand{Username: req.Username, Email: req.Email, DisplayName: req.DisplayName, Password: req.Password, PreferredLocale: req.PreferredLocale, PreferredTimezone: req.PreferredTimezone})
	if err != nil {
		userManagementError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) ChangeUserStatus(w http.ResponseWriter, r *http.Request) {
	if !h.allowUserManagement(w, r, "iam.users.manage") {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	var req ChangeUserStatusRequest
	decoder := strictUserDecoder(r)
	if err := decoder.Decode(&req); err != nil || !endOfUserRequest(decoder) {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	user, err := h.users.ChangeStatus(r.Context(), UserFromContext(r.Context()).ID, id, domain.UserStatus(req.Status), req.ExpectedVersion)
	if err != nil {
		userManagementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	if !h.allowUserManagement(w, r, "iam.users.manage") {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	var req ResetUserPasswordRequest
	decoder := strictUserDecoder(r)
	if err := decoder.Decode(&req); err != nil || !endOfUserRequest(decoder) {
		userManagementError(w, domain.ErrInvalidUserInput)
		return
	}
	if err := h.users.ResetPassword(r.Context(), UserFromContext(r.Context()).ID, id, req.Password, req.ExpectedVersion); err != nil {
		userManagementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"updated": true})
}

func userManagementError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, domain.ErrInvalidUserInput):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Check the user details and expected version")
	case errors.Is(err, domain.ErrPasswordTooWeak):
		writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "Password must contain at least eight characters")
	case errors.Is(err, domain.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
	case errors.Is(err, domain.ErrUserVersionConflict):
		writeError(w, http.StatusConflict, "CONCURRENCY_CONFLICT", "User changed; reload and try again")
	case errors.Is(err, domain.ErrSelfDeactivation):
		writeError(w, http.StatusConflict, "SELF_DEACTIVATION", "You cannot deactivate your own account")
	case errors.Is(err, domain.ErrLastAdministrator):
		writeError(w, http.StatusConflict, "LAST_ADMINISTRATOR", "The last active administrator cannot be deactivated")
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		writeError(w, http.StatusConflict, "USER_EXISTS", "Username or email already exists")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to complete user administration")
	}
}
