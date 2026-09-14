package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Recovery middleware recovers from any panics, logs the error and stack trace structured,
// and returns a safe JSON 500 Internal Server Error without leaking internal implementation details.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID := RequestIDFromContext(r.Context())
					stack := string(debug.Stack())

					if logger != nil {
						logger.Error("HTTP handler panicked",
							"error", fmt.Sprintf("%v", err),
							"request_id", reqID,
							"method", r.Method,
							"path", r.URL.Path,
							"stack", stack,
						)
					} else {
						slog.Error("HTTP handler panicked",
							"error", fmt.Sprintf("%v", err),
							"request_id", reqID,
							"method", r.Method,
							"path", r.URL.Path,
							"stack", stack,
						)
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					resp := ErrorResponse{
						Error: ErrorDetails{
							Code:      "INTERNAL_SERVER_ERROR",
							Message:   "An unexpected internal error occurred",
							RequestID: reqID,
						},
					}

					_ = json.NewEncoder(w).Encode(resp)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
