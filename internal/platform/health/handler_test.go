package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/platform/config"
)

type mockChecker struct {
	err error
}

func (m *mockChecker) Check() error {
	return m.err
}

func TestHealthLiveness(t *testing.T) {
	cfg := &config.Config{
		AppName:     "Saudi-ERP-Test",
		AppVersion:  "1.0.0",
		Environment: config.EnvTesting,
	}

	handler := NewHandler(cfg)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.Liveness(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != StatusUP {
		t.Errorf("expected status UP, got %s", resp.Status)
	}

	if resp.Service != "Saudi-ERP-Test" {
		t.Errorf("expected service 'Saudi-ERP-Test', got %s", resp.Service)
	}
}

func TestHealthReadiness(t *testing.T) {
	cfg := &config.Config{
		AppName:     "Saudi-ERP-Test",
		AppVersion:  "1.0.0",
		Environment: config.EnvTesting,
	}

	t.Run("readiness healthy", func(t *testing.T) {
		checker := &mockChecker{err: nil}
		handler := NewHandler(cfg, checker)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rr := httptest.NewRecorder()

		handler.Readiness(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}

		var resp HealthResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != StatusUP {
			t.Errorf("expected status UP, got %s", resp.Status)
		}
	})

	t.Run("readiness unhealthy", func(t *testing.T) {
		checker := &mockChecker{err: errors.New("db ping failed")}
		handler := NewHandler(cfg, checker)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rr := httptest.NewRecorder()

		handler.Readiness(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", rr.Code)
		}

		var resp HealthResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Status != StatusDOWN {
			t.Errorf("expected status DOWN, got %s", resp.Status)
		}
	})
}
