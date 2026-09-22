package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"erp/internal/audit/ports"
	iamhttp "erp/internal/iam/transport/http"
	platformmiddleware "erp/internal/platform/middleware"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	if w.statusCode != 0 {
		return
	}
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterInterceptor) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseWriterInterceptor) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// AuditRequestMiddleware synchronously records authorized API access and
// successful state-changing requests. Audit persistence failures are logged;
// they cannot safely change a response after a domain transaction committed.
func AuditRequestMiddleware(service ports.AuditService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriterInterceptor{ResponseWriter: w}

			next.ServeHTTP(wrapped, r)

			latency := time.Since(start).Milliseconds()
			statusCode := wrapped.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			var userID *uuid.UUID
			if user := iamhttp.UserFromContext(r.Context()); user != nil {
				id := user.ID
				userID = &id
			}

			ipAddress := clientIPAddress(r.RemoteAddr)

			if isSuccessfulMutation(r.Method, statusCode) && !strings.HasPrefix(r.URL.Path, "/api/v1/audit") {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				legalEntityID := legalEntityFromRequest(r)
				payload, _ := json.Marshal(map[string]any{
					"method":      r.Method,
					"path":        r.URL.Path,
					"status_code": statusCode,
					"request_id":  middleware.GetReqID(r.Context()),
				})
				_, err := service.RecordEvent(ctx, ports.RecordEventCommand{
					UserID:        userID,
					LegalEntityID: legalEntityID,
					Action:        mutationAction(r.Method),
					ResourceType:  resourceType(r.URL.Path),
					ResourceID:    resourceID(r.URL.Path),
					Payload:       payload,
					IPAddress:     ipAddress,
					UserAgent:     r.UserAgent(),
				})
				cancel()
				if err != nil {
					slog.Error("failed to record automatic mutation audit event", "error", err, "method", r.Method, "path", r.URL.Path)
				}
			}

			accessCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := service.RecordAccessEvent(accessCtx, ports.RecordAccessEventCommand{
				UserID:       userID,
				ResourcePath: r.URL.Path,
				Method:       r.Method,
				StatusCode:   statusCode,
				IPAddress:    ipAddress,
				LatencyMS:    int(latency),
			})
			cancel()
			if err != nil {
				slog.Error("failed to record API access audit event", "error", err, "method", r.Method, "path", r.URL.Path)
			}
		})
	}
}

func clientIPAddress(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func isSuccessfulMutation(method string, statusCode int) bool {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return false
	}
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func mutationAction(method string) string {
	switch method {
	case http.MethodPost:
		return "CREATE"
	case http.MethodDelete:
		return "DELETE"
	default:
		return "UPDATE"
	}
}

func legalEntityFromRequest(r *http.Request) *uuid.UUID {
	if tenantID, ok := platformmiddleware.TenantFromContext(r.Context()); ok {
		return &tenantID
	}
	for _, candidate := range []string{r.URL.Query().Get("legal_entity_id"), r.Header.Get("X-Legal-Entity-ID")} {
		if parsed, err := uuid.Parse(candidate); err == nil && parsed != uuid.Nil {
			return &parsed
		}
	}
	return nil
}

func resourceType(path string) string {
	segments := strings.Split(strings.Trim(strings.TrimPrefix(path, "/api/v1/"), "/"), "/")
	named := make([]string, 0, 2)
	for _, segment := range segments {
		if segment == "" || isIdentifierSegment(segment) {
			break
		}
		named = append(named, segment)
		if len(named) == 2 {
			break
		}
	}
	if len(named) == 0 {
		return "API_RESOURCE"
	}
	return strings.ToUpper(strings.ReplaceAll(strings.Join(named, "_"), "-", "_"))
}

// isIdentifierSegment reports whether a path segment carries a record
// identifier rather than a resource name, so identifiers stay out of the
// audit resource type and its cardinality stays bounded.
func isIdentifierSegment(segment string) bool {
	if _, err := uuid.Parse(segment); err == nil {
		return true
	}
	if _, err := strconv.ParseUint(segment, 10, 64); err == nil {
		return true
	}
	return false
}

func resourceID(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if _, err := uuid.Parse(segments[i]); err == nil {
			return segments[i]
		}
	}
	return path
}
