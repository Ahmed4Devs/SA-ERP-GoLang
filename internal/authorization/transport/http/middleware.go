package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/authorization/domain"
	"erp/internal/authorization/ports"
	iamhttp "erp/internal/iam/transport/http"
	platformmiddleware "erp/internal/platform/middleware"
	"github.com/google/uuid"
)

// RequirePermission constructs an HTTP middleware that verifies user holds the required permission
// within the extracted legal entity scope.
func RequirePermission(
	authzService ports.AuthorizationService,
	permissionCode string,
	extractScope func(r *http.Request) *uuid.UUID,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := iamhttp.UserFromContext(r.Context())
			if user == nil {
				writeForbidden(w, "User authentication required for permission check")
				return
			}

			var scope *uuid.UUID
			if extractScope != nil {
				scope = extractScope(r)
			}

			decision, err := authzService.Authorize(r.Context(), domain.AuthorizationRequest{
				UserID:        user.ID,
				Permission:    permissionCode,
				LegalEntityID: scope,
			})
			if err != nil {
				writeForbidden(w, "Failed to evaluate authorization: "+err.Error())
				return
			}

			if !decision.Allowed {
				writeForbidden(w, decision.Reason)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireModuleAccess selects read or manage permission from the HTTP method
// and evaluates it in the legal-entity scope established by the tenant guard.
func RequireModuleAccess(authzService ports.AuthorizationService, viewPermission, managePermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := iamhttp.UserFromContext(r.Context())
			if user == nil {
				writeForbidden(w, "User authentication required for permission check")
				return
			}

			permission := managePermission
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				permission = viewPermission
			}

			var scope *uuid.UUID
			if tenantID, ok := platformmiddleware.TenantFromContext(r.Context()); ok {
				scope = &tenantID
			}

			decision, err := authzService.Authorize(r.Context(), domain.AuthorizationRequest{
				UserID: user.ID, Permission: permission, LegalEntityID: scope,
			})
			if err != nil {
				writeForbidden(w, "Failed to evaluate authorization")
				return
			}
			if !decision.Allowed {
				writeForbidden(w, decision.Reason)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "FORBIDDEN",
			"message": msg,
		},
	})
}
