package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientRecord struct {
	tokens     float64
	lastRefill time.Time
}

type RateLimiterConfig struct {
	GeneralLimit   int           // Max requests allowed per window
	AuthLimit      int           // Max requests allowed for auth endpoints per window
	WindowDuration time.Duration // Time window (e.g. 1 minute)
	TrustedProxies []string
}

func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		GeneralLimit:   120,
		AuthLimit:      10,
		WindowDuration: 1 * time.Minute,
		TrustedProxies: []string{"127.0.0.1", "::1"},
	}
}

type RateLimiter struct {
	cfg     RateLimiterConfig
	mu      sync.Mutex
	clients map[string]*clientRecord
}

func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		cfg:     cfg,
		clients: make(map[string]*clientRecord),
	}

	// Periodic cleanup of stale clients
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, rec := range rl.clients {
				if now.Sub(rec.lastRefill) > 10*time.Minute {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client identifier (IP from Real-IP, X-Forwarded-For, or RemoteAddr)
			ip := clientIP(r.RemoteAddr)
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" && isTrustedProxy(rl.cfg.TrustedProxies, ip) {
				ip = strings.TrimSpace(strings.Split(forwarded, ",")[0])
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" && isTrustedProxy(rl.cfg.TrustedProxies, ip) {
				ip = realIP
			}

			// Differentiate between sensitive auth endpoints and general API endpoints
			isAuth := strings.HasPrefix(r.URL.Path, "/api/v1/auth/login") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/register")

			limit := rl.cfg.GeneralLimit
			if isAuth {
				limit = rl.cfg.AuthLimit
			}

			clientKey := fmt.Sprintf("%s:%t", ip, isAuth)

			rl.mu.Lock()
			now := time.Now()
			rec, exists := rl.clients[clientKey]
			if !exists {
				rec = &clientRecord{
					tokens:     float64(limit),
					lastRefill: now,
				}
				rl.clients[clientKey] = rec
			}

			// Refill tokens based on elapsed time
			elapsed := now.Sub(rec.lastRefill)
			refillRate := float64(limit) / rl.cfg.WindowDuration.Seconds()
			rec.tokens += elapsed.Seconds() * refillRate
			if rec.tokens > float64(limit) {
				rec.tokens = float64(limit)
			}
			rec.lastRefill = now

			if rec.tokens < 1.0 {
				rl.mu.Unlock()
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.cfg.WindowDuration.Seconds()))
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "Too Many Requests: rate limit exceeded, please try again later",
				})
				return
			}

			rec.tokens -= 1.0
			remaining := int(rec.tokens)
			rl.mu.Unlock()

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}

func isTrustedProxy(values []string, target string) bool {
	targetIP := net.ParseIP(strings.TrimSpace(target))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if strings.EqualFold(value, target) {
			return true
		}
		if _, network, err := net.ParseCIDR(value); err == nil && targetIP != nil && network.Contains(targetIP) {
			return true
		}
	}
	return false
}
