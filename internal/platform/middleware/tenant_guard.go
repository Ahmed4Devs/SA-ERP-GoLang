package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type tenantContextKey struct{}

// TenantBoundaryGuard validates and enforces legal entity isolation.
func TenantBoundaryGuard() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			leStr := r.URL.Query().Get("legal_entity_id")
			if leStr == "" {
				leStr = r.Header.Get("X-Legal-Entity-ID")
			}

			if leStr != "" {
				leID, err := uuid.Parse(leStr)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"error": "Invalid legal_entity_id: must be a valid RFC 4122 / UUIDv7",
					})
					return
				}

				ctx := context.WithValue(r.Context(), tenantContextKey{}, leID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantFromContext extracts the legal entity ID from context if present.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	id, ok := ctx.Value(tenantContextKey{}).(uuid.UUID)
	return id, ok
}
