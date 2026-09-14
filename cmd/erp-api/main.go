package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"erp/internal/iam"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	platformhttp "erp/internal/platform/http"
	"erp/internal/platform/observability"
	"erp/internal/platform/transaction"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup Structured Logger (slog)
	logger := observability.NewLogger(cfg.AppName, string(cfg.Environment), cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("Starting Saudi ERP Platform API",
		"service", cfg.AppName,
		"version", cfg.AppVersion,
		"env", string(cfg.Environment),
		"port", cfg.HTTP.Port,
	)

	// 3. Initialize OpenTelemetry Tracing
	tracerProvider, err := observability.InitTracer(cfg.AppName)
	if err != nil {
		logger.Warn("Failed to initialize OpenTelemetry tracer", "error", err)
	} else {
		defer func() {
			_ = tracerProvider.Shutdown(context.Background())
		}()
	}

	// 4. Setup Root Context with Signal Cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 5. Connect to PostgreSQL Pool (Optional/soft-fail on bootstrap if offline)
	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Warn("Database pool initialization deferred or unavailable", "error", err)
	} else {
		defer db.Close()
		logger.Info("Database pool connected successfully")
	}

	// 6. Setup Production HTTP Server
	server := platformhttp.NewServer(cfg, logger, db)

	// 7. Mount Domain Modules
	if db != nil {
		txManager := transaction.NewManager(db.Pool)
		iamModule := iam.NewModule(cfg, db, txManager)
		iamModule.RegisterRoutes(server.Router)

		// Seed initial system administrator if missing
		_, err := iamModule.AuthService.SeedAdmin(ctx, "admin", "admin@erp.sa", "مدير النظام الشامل", "Admin@ERP123!")
		if err != nil {
			logger.Debug("Initial admin user already exists or seeding skipped", "reason", err.Error())
		} else {
			logger.Info("Default admin user seeded successfully", "username", "admin", "email", "admin@erp.sa")
		}
	}

	// 7. Start HTTP Server with Graceful Shutdown
	if err := server.Start(ctx); err != nil {
		logger.Error("HTTP server exited with error", "error", err)
		os.Exit(1)
	}
}
