package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"erp/internal/platform/middleware"
)

type loggerContextKey struct{}

// NewLogger instantiates a structured slog.Logger based on environment and log level.
func NewLogger(serviceName, env, levelStr string) *slog.Logger {
	level := parseLogLevel(levelStr)
	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		slog.String("service", serviceName),
		slog.String("env", env),
	)
}

// WithLogger injects logger into the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// FromContext extracts logger from context, falling back to default logger if missing.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger); ok && logger != nil {
			return logger
		}
	}
	return slog.Default()
}

// HTTPLoggingMiddleware logs structured request outcomes including timing, status, and request correlation.
func HTTPLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.RequestIDFromContext(r.Context())

			reqLogger := logger.With(slog.String("request_id", reqID))
			ctx := WithLogger(r.Context(), reqLogger)

			rw := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r.WithContext(ctx))

			latency := time.Since(start)

			logLevel := slog.LevelInfo
			if rw.statusCode >= 500 {
				logLevel = slog.LevelError
			} else if rw.statusCode >= 400 {
				logLevel = slog.LevelWarn
			}

			reqLogger.Log(ctx, logLevel, "HTTP request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("latency", latency),
				slog.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
				slog.Int("bytes", rw.bytesWritten),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriterInterceptor) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
