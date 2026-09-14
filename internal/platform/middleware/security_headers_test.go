package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/platform/middleware"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	t.Run("applies base security headers", func(t *testing.T) {
		cfg := middleware.DefaultSecurityHeadersConfig(false)
		mw := middleware.SecurityHeaders(cfg)

		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Error("expected nosniff for X-Content-Type-Options")
		}
		if rr.Header().Get("X-Frame-Options") != "DENY" {
			t.Error("expected DENY for X-Frame-Options")
		}
		if rr.Header().Get("X-XSS-Protection") != "1; mode=block" {
			t.Error("expected 1; mode=block for X-XSS-Protection")
		}
		if rr.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
			t.Error("expected strict-origin-when-cross-origin for Referrer-Policy")
		}
		if rr.Header().Get("Content-Security-Policy") != "default-src 'self'" {
			t.Error("expected default-src 'self' for Content-Security-Policy")
		}
		if rr.Header().Get("Strict-Transport-Security") != "" {
			t.Error("did not expect HSTS in non-production")
		}
	})

	t.Run("applies HSTS in production", func(t *testing.T) {
		cfg := middleware.DefaultSecurityHeadersConfig(true)
		mw := middleware.SecurityHeaders(cfg)

		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		hsts := rr.Header().Get("Strict-Transport-Security")
		if hsts == "" {
			t.Error("expected Strict-Transport-Security in production")
		}
	})
}
