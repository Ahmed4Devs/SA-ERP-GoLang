package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type tenantContextKey struct{}

// TenantBoundaryGuard validates and enforces legal entity isolation.
func TenantBoundaryGuard() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			candidates := []string{
				r.URL.Query().Get("legal_entity_id"),
				r.Header.Get("X-Legal-Entity-ID"),
			}
			contentType := strings.ToLower(r.Header.Get("Content-Type"))
			if strings.HasPrefix(contentType, "application/json") && r.Body != nil {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					writeTenantError(w, http.StatusBadRequest, "Unable to read request body")
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
				if len(body) > 0 {
					var envelope struct {
						LegalEntityID string `json:"legal_entity_id"`
					}
					if json.Unmarshal(body, &envelope) == nil {
						candidates = append(candidates, envelope.LegalEntityID)
					}
				}
			}
			if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				if err := r.ParseMultipartForm(32 << 20); err == nil {
					candidates = append(candidates, r.FormValue("legal_entity_id"))
				}
			}

			var tenantID uuid.UUID
			for _, candidate := range candidates {
				if strings.TrimSpace(candidate) == "" {
					continue
				}
				parsed, err := uuid.Parse(candidate)
				if err != nil {
					writeTenantError(w, http.StatusBadRequest, "Invalid legal_entity_id: must be a valid UUID")
					return
				}
				if tenantID != uuid.Nil && tenantID != parsed {
					writeTenantError(w, http.StatusBadRequest, "Conflicting legal_entity_id values in request")
					return
				}
				tenantID = parsed
			}

			if tenantID != uuid.Nil {
				ctx := context.WithValue(r.Context(), tenantContextKey{}, tenantID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeTenantError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// TenantFromContext extracts the legal entity ID from context if present.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	id, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
	return id, ok
}
