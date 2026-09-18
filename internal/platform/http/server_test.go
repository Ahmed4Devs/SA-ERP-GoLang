package http_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/platform/config"
	platformhttp "erp/internal/platform/http"
	"erp/internal/platform/middleware"
)

func TestServerPipeline(t *testing.T) {
	cfg := &config.Config{
		AppName:     "Saudi-ERP-Test",
		AppVersion:  "0.1.0",
		Environment: config.EnvTesting,
		HTTP: config.HTTPConfig{
			Port: "9999",
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := platformhttp.NewServer(cfg, logger, nil)

	t.Run("serves root welcome endpoint with security headers and request id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		server.Router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		if rr.Header().Get(middleware.RequestIDHeader) == "" {
			t.Error("expected X-Request-ID header to be present")
		}

		if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Error("expected nosniff header")
		}

		if rr.Header().Get("X-Frame-Options") != "DENY" {
			t.Error("expected DENY header")
		}
	})

	t.Run("serves healthz endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()

		server.Router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("serves metrics endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rr := httptest.NewRecorder()

		server.Router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("protects metrics when token is configured", func(t *testing.T) {
		cfg.Security.MetricsToken = "test-metrics-token-with-32-characters"
		protectedServer := platformhttp.NewServer(cfg, logger, nil)
		unauthorized := httptest.NewRecorder()
		protectedServer.Router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		if unauthorized.Code != http.StatusUnauthorized {
			t.Fatalf("expected metrics status 401, got %d", unauthorized.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.Header.Set("Authorization", "Bearer "+cfg.Security.MetricsToken)
		authorized := httptest.NewRecorder()
		protectedServer.Router.ServeHTTP(authorized, req)
		if authorized.Code != http.StatusOK {
			t.Fatalf("expected authorized metrics status 200, got %d", authorized.Code)
		}
	})

	t.Run("recovers from panic gracefully", func(t *testing.T) {
		server.Router.Get("/test-panic", func(w http.ResponseWriter, r *http.Request) {
			panic("unexpected handler explosion")
		})

		req := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
		rr := httptest.NewRecorder()

		server.Router.ServeHTTP(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500 on panic, got %d", rr.Code)
		}
	})
}
