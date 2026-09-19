package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"erp/internal/notifications"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/mail"
	"erp/internal/platform/observability"
	"erp/internal/platform/outbox"
	"erp/internal/zatca"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load worker config: %v\n", err)
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
	if err := db.Check(); err != nil {
		logger.Error("Worker database schema is not ready", "error", err)
		os.Exit(1)
	}

	var mailer mail.Mailer = mail.NewMockMailer()
	if cfg.Mail.Driver == "smtp" {
		mailer = mail.NewSMTPMailer(mail.SMTPConfig{
			Host: cfg.Mail.Host, Port: cfg.Mail.Port, Username: cfg.Mail.Username,
			Password: cfg.Mail.Password, From: cfg.Mail.From,
		})
	}
	notificationsModule := notifications.NewModule(cfg, db, mailer)
	dispatcher := outbox.NewDispatcher(notificationsModule.Outbox, logger, outbox.DefaultDispatcherConfig())
	if err := notificationsModule.RegisterOutboxHandlers(dispatcher); err != nil {
		logger.Error("Failed to register outbox handler", "error", err)
		os.Exit(1)
	}

	jobs := []string{"outbox", "notification_delivery"}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		dispatcher.Run(ctx)
	}()

	if cfg.ZATCA.Enabled {
		zatcaModule := zatca.NewModule(cfg, db)
		jobs = append(jobs, "zatca_submission_retries")
		wg.Add(1)
		go func() {
			defer wg.Done()
			zatcaModule.RetryWorker.Start(ctx)
		}()
	}

	logger.Info("Worker daemon initialized", "jobs", jobs)
	<-ctx.Done()
	wg.Wait()
	logger.Info("Worker daemon stopped gracefully")
}
