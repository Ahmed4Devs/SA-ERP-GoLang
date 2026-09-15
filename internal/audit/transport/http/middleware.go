package http

import (
	"context"
	"net/http"
	"time"

	"erp/internal/audit/ports"
	iamhttp "erp/internal/iam/transport/http"

	"github.com/google/uuid"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// AuditAccessMiddleware logs HTTP access requests to the audit access log.
func AuditAccessMiddleware(service ports.AuditService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			latency := time.Since(start).Milliseconds()

			var userID *uuid.UUID
			if user := iamhttp.UserFromContext(r.Context()); user != nil {
				userID = &user.ID
			}

			// Run in background with a detached context
			go func(path, method, ip string, status, lat int, uID *uuid.UUID) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, _ = service.RecordAccessEvent(ctx, ports.RecordAccessEventCommand{
					UserID:       uID,
					ResourcePath: path,
					Method:       method,
					StatusCode:   status,
					IPAddress:    ip,
					LatencyMS:    lat,
				})
			}(r.URL.Path, r.Method, r.RemoteAddr, wrapped.statusCode, int(latency), userID)
		})
	}
}
