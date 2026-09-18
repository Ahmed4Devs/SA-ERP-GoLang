package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/observability"
	"erp/internal/zatca"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load worker config: %v\n", err)
		os.Exit(1)
	}
	if !cfg.ZATCA.Enabled {
		fmt.Fprintln(os.Stderr, "ZATCA worker is disabled; set ZATCA_ENABLED=true only for an approved deployment")
		os.Exit(1)
	}

	logger := observability.NewLogger(cfg.AppName+"-worker", string(cfg.Environment), cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("Starting Saudi ERP Worker Daemon",
		"service", cfg.AppName+"-worker",
		"version", cfg.AppVersion,
		"env", string(cfg.Environment),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		logger.Error("Worker database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	zatcaModule := zatca.NewModule(cfg, db)
	logger.Info("Worker daemon initialized", "jobs", []string{"zatca_submission_retries"})
	zatcaModule.RetryWorker.Start(ctx)
	logger.Info("Worker daemon stopped gracefully")
}
