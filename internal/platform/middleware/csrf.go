package middleware

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type CSRFConfig struct {
	AllowedOrigins []string
	ExemptPaths    []string
}

func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		ExemptPaths:    []string{"/health", "/ready", "/metrics"},
	}
}

// CSRFProtection returns a middleware that validates request Origin/Referer on mutating requests.
func CSRFProtection(cfg CSRFConfig) func(http.Handler) http.Handler {
	allowedMap := make(map[string]bool)
	for _, o := range cfg.AllowedOrigins {
		allowedMap[strings.ToLower(o)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Safe HTTP methods do not require CSRF validation
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
				next.ServeHTTP(w, r)
				return
			}

			// Check path exemptions
			for _, exempt := range cfg.ExemptPaths {
				if strings.HasPrefix(r.URL.Path, exempt) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Validate Origin header if present
			origin := r.Header.Get("Origin")
			if origin != "" {
				if !isOriginAllowed(origin, r.Host, allowedMap) {
					rejectCSRF(w, "Forbidden: Invalid or untrusted Origin")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Fallback: Validate Referer header if present
			referer := r.Header.Get("Referer")
			if referer != "" {
				refURL, err := url.Parse(referer)
				if err != nil || !isHostAllowed(refURL.Host, r.Host, allowedMap) {
					rejectCSRF(w, "Forbidden: Invalid or untrusted Referer")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// If neither Origin nor Referer is present, standard API client with custom header or same-host is accepted
			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, requestHost string, allowedMap map[string]bool) bool {
	originLower := strings.ToLower(origin)
	if allowedMap[originLower] {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	return strings.EqualFold(u.Host, requestHost)
}

func isHostAllowed(refererHost string, requestHost string, allowedMap map[string]bool) bool {
	if strings.EqualFold(refererHost, requestHost) {
		return true
	}
	for o := range allowedMap {
		if u, err := url.Parse(o); err == nil && strings.EqualFold(u.Host, refererHost) {
			return true
		}
	}
	return false
}

func rejectCSRF(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
