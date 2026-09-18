package middleware

import (
	"net/http"
	"strings"
)

type BodyLimitConfig struct {
	DefaultLimitBytes int64 // Standard API payloads (e.g. 10MB)
	UploadLimitBytes  int64 // Multipart document uploads (e.g. 32MB)
}

func DefaultBodyLimitConfig() BodyLimitConfig {
	return BodyLimitConfig{
		DefaultLimitBytes: 10 * 1024 * 1024, // 10 MB
		UploadLimitBytes:  32 * 1024 * 1024, // 32 MB
	}
}

// RequestBodyLimit wraps the request Body with http.MaxBytesReader to prevent memory exhaustion DoS.
func RequestBodyLimit(cfg BodyLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			limit := cfg.DefaultLimitBytes
			contentType := r.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "multipart/form-data") {
				limit = cfg.UploadLimitBytes
			}

			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
