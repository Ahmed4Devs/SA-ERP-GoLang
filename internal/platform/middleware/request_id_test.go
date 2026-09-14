package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/platform/middleware"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("generates new request ID if missing", func(t *testing.T) {
		handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.RequestIDFromContext(r.Context())
			if reqID == "" {
				t.Error("expected request ID in context, got empty string")
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		headerID := rr.Header().Get(middleware.RequestIDHeader)
		if headerID == "" {
			t.Error("expected X-Request-ID response header to be set")
		}
	})

	t.Run("preserves existing request ID if supplied", func(t *testing.T) {
		incomingID := "custom-req-id-12345"

		handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.RequestIDFromContext(r.Context())
			if reqID != incomingID {
				t.Errorf("expected context request ID %s, got %s", incomingID, reqID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.RequestIDHeader, incomingID)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		headerID := rr.Header().Get(middleware.RequestIDHeader)
		if headerID != incomingID {
			t.Errorf("expected header %s, got %s", incomingID, headerID)
		}
	})
}
