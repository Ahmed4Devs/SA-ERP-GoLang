package http

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/health"
	"erp/internal/platform/middleware"
	"erp/internal/platform/observability"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg        *config.Config
	logger     *slog.Logger
	Router     *chi.Mux
	httpServer *http.Server
	Metrics    *observability.Metrics
	db         *database.DB
}

func NewServer(cfg *config.Config, logger *slog.Logger, db *database.DB, dependencyCheckers ...health.Checker) *Server {
	router := chi.NewRouter()
	metrics := observability.NewMetrics()

	isProd := cfg.Environment == config.EnvProduction

	// 1. Canonical Middleware Chain in strict execution order
	router.Use(middleware.RequestID)
	rateLimitConfig := middleware.DefaultRateLimiterConfig()
	rateLimitConfig.TrustedProxies = cfg.Security.TrustedProxies
	router.Use(middleware.NewRateLimiter(rateLimitConfig).Middleware())
	router.Use(middleware.RequestBodyLimit(middleware.DefaultBodyLimitConfig()))
	corsConfig := middleware.DefaultCORSConfig()
	corsConfig.AllowedOrigins = cfg.Security.AllowedOrigins
	router.Use(middleware.CORS(corsConfig))
	csrfConfig := middleware.DefaultCSRFConfig()
	csrfConfig.AllowedOrigins = cfg.Security.AllowedOrigins
	router.Use(middleware.CSRFProtection(csrfConfig))
	router.Use(observability.HTTPLoggingMiddleware(logger))
	router.Use(metrics.HTTPMetricsMiddleware)
	router.Use(middleware.SecurityHeaders(middleware.DefaultSecurityHeadersConfig(isProd)))
	router.Use(middleware.TenantBoundaryGuard())
	router.Use(middleware.Recovery(logger))
	router.Use(chimiddleware.Timeout(60 * time.Second))

	s := &Server{
		cfg:     cfg,
		logger:  logger,
		Router:  router,
		Metrics: metrics,
		db:      db,
	}

	s.registerCoreRoutes(dependencyCheckers...)

	s.httpServer = &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return s
}

func (s *Server) registerCoreRoutes(dependencyCheckers ...health.Checker) {
	// Health and Readiness
	var checkers []health.Checker
	if s.db != nil {
		checkers = append(checkers, s.db)
	}
	checkers = append(checkers, dependencyCheckers...)
	healthHandler := health.NewHandler(s.cfg, checkers...)
	s.Router.Mount("/", healthHandler.Routes())

	// Prometheus Metrics
	s.Router.Handle("/metrics", protectMetrics(s.cfg.Security.MetricsToken, s.Metrics.Handler()))

	// Welcome Root
	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"message":"Welcome to Saudi ERP API Platform","version":"%s","status":"operational"}`, s.cfg.AppVersion)))
	})
}

func protectMetrics(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("Authorization")
		expected := "Bearer " + token
		if len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="metrics"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Start launches the HTTP server and handles graceful shutdown when ctx is canceled.
func (s *Server) Start(ctx context.Context) error {
	serverErr := make(chan error, 1)

	go func() {
		s.logger.Info("HTTP server starting",
			"addr", s.httpServer.Addr,
			"env", string(s.cfg.Environment),
			"version", s.cfg.AppVersion,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server failed: %w", err)
	case <-ctx.Done():
		s.logger.Info("Shutting down HTTP server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server forced shutdown: %w", err)
		}
		s.logger.Info("HTTP server stopped gracefully")
		return nil
	}
}
