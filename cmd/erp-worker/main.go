package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"erp/internal/platform/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load worker config: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting Saudi ERP Worker Daemon",
		"service", cfg.AppName+"-worker",
		"version", cfg.AppVersion,
		"env", string(cfg.Environment),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Worker run loop placeholder (will run River job queue in Phase 1+)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	logger.Info("Worker daemon initialized and ready for jobs")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Worker daemon shutting down gracefully")
			return
		case <-ticker.C:
			logger.Debug("Worker heartbeat check")
		}
	}
}
