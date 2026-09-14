package middleware

import (
	"net/http"
)

type SecurityHeadersConfig struct {
	IsProduction bool
	CSP          string
}

// DefaultSecurityHeadersConfig returns standard OWASP-recommended security settings.
func DefaultSecurityHeadersConfig(isProduction bool) SecurityHeadersConfig {
	return SecurityHeadersConfig{
		IsProduction: isProduction,
		CSP:          "default-src 'self'",
	}
}

// SecurityHeaders applies baseline HTTP security headers to protect against common web attacks.
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

			if cfg.CSP != "" {
				w.Header().Set("Content-Security-Policy", cfg.CSP)
			}

			if cfg.IsProduction {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			next.ServeHTTP(w, r)
		})
	}
}
