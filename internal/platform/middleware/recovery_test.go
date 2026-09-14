package middleware_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/platform/middleware"
)

func TestRecoveryMiddleware(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	recoveryMw := middleware.Recovery(discardLogger)

	panickingHandler := recoveryMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("database connection suddenly severed")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	panickingHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var resp middleware.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if resp.Error.Code != "INTERNAL_SERVER_ERROR" {
		t.Errorf("expected code INTERNAL_SERVER_ERROR, got %s", resp.Error.Code)
	}

	if resp.Error.Message != "An unexpected internal error occurred" {
		t.Errorf("expected sanitized message, got %s", resp.Error.Message)
	}
}
