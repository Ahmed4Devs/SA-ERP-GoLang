package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// InitTracer initializes OpenTelemetry tracing and configures W3C trace context propagation.
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// Tracer returns an OpenTelemetry tracer for the specified package or module.
func Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

// StartSpan creates a new active span in the context.
func StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, trace.Span) {
	return Tracer(tracerName).Start(ctx, spanName)
}
