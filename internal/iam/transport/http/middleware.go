package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"erp/internal/iam/application/services"
	"erp/internal/iam/domain"
)

type userContextKey struct{}
type sessionContextKey struct{}

const SessionCookieName = "erp_session"

func RequireAuth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeUnauthorized(w, "Authentication required")
				return
			}

			user, session, err := authService.ValidateSession(r.Context(), token)
			if err != nil {
				writeUnauthorized(w, "Invalid or expired session")
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey{}, user)
			ctx = context.WithValue(ctx, sessionContextKey{}, session)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ContextWithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) *domain.User {
	if ctx == nil {
		return nil
	}
	if u, ok := ctx.Value(userContextKey{}).(*domain.User); ok {
		return u
	}
	return nil
}

func SessionFromContext(ctx context.Context) *domain.Session {
	if ctx == nil {
		return nil
	}
	if s, ok := ctx.Value(sessionContextKey{}).(*domain.Session); ok {
		return s
	}
	return nil
}

func extractToken(r *http.Request) string {
	// 1. Try Cookie first
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Try Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    "UNAUTHORIZED",
			"message": msg,
		},
	})
}
