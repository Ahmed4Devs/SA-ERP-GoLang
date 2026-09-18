package observability_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/platform/config"
	"erp/internal/platform/observability"
)

func TestObservabilityLoggingAndTracing(t *testing.T) {
	logger := observability.NewLogger("test-service", "testing", "debug")
	ctx := observability.WithLogger(context.Background(), logger)

	extracted := observability.FromContext(ctx)
	if extracted == nil {
		t.Error("expected extracted logger to not be nil")
	}

	tp, err := observability.InitTracer(context.Background(), "test-service", "1.0.0", config.TelemetryConfig{})
	if err != nil {
		t.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	_, span := observability.StartSpan(ctx, "test-tracer", "test-span")
	if span == nil {
		t.Error("expected non-nil span")
	}
	span.End()
}

func TestObservabilityMetrics(t *testing.T) {
	metrics := observability.NewMetrics()

	testHandler := metrics.HTTPMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	rr := httptest.NewRecorder()

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	// Query metrics endpoint
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRR := httptest.NewRecorder()

	metrics.Handler().ServeHTTP(metricsRR, metricsReq)

	if metricsRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 for metrics, got %d", metricsRR.Code)
	}

	body, err := io.ReadAll(metricsRR.Body)
	if err != nil {
		t.Fatalf("failed to read metrics body: %v", err)
	}

	bodyStr := string(body)
	if len(bodyStr) == 0 {
		t.Error("expected non-empty metrics output")
	}
}
