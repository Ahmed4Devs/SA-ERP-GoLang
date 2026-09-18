package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
)

type RetryWorkerConfig struct {
	PollInterval time.Duration
	MaxRetries   int
	BatchSize    int
}

func DefaultWorkerConfig() RetryWorkerConfig {
	return RetryWorkerConfig{
		PollInterval: 1 * time.Minute,
		MaxRetries:   5,
		BatchSize:    20,
	}
}

type RetryWorker struct {
	cfg        RetryWorkerConfig
	subRepo    ports.SubmissionRepository
	docRepo    ports.Repository
	subService ports.SubmissionService
	stopChan   chan struct{}
}

func NewRetryWorker(
	cfg RetryWorkerConfig,
	subRepo ports.SubmissionRepository,
	docRepo ports.Repository,
	subService ports.SubmissionService,
) *RetryWorker {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 1 * time.Minute
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 5
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 20
	}

	return &RetryWorker{
		cfg:        cfg,
		subRepo:    subRepo,
		docRepo:    docRepo,
		subService: subService,
		stopChan:   make(chan struct{}),
	}
}

// ProcessPending runs a single pass over pending retry submissions.
func (w *RetryWorker) ProcessPending(ctx context.Context) (int, error) {
	pending, err := w.subRepo.GetPendingRetrySubmissions(ctx, w.cfg.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to query pending retry submissions: %w", err)
	}

	processed := 0
	for _, sub := range pending {
		select {
		case <-ctx.Done():
			return processed, ctx.Err()
		default:
		}

		attempts, err := w.subRepo.ListAttempts(ctx, sub.ID)
		if err != nil {
			return processed, fmt.Errorf("failed to list attempts for submission %s: %w", sub.ID, err)
		}
		if len(attempts) >= w.cfg.MaxRetries {
			// Mark permanently failed
			sub.Status = "FAILED"
			if _, err := w.subRepo.UpdateSubmission(ctx, sub); err != nil {
				return processed, fmt.Errorf("failed to mark submission %s failed: %w", sub.ID, err)
			}
			if _, err := w.docRepo.UpdateDocumentStatus(ctx, sub.DocumentID, sub.LegalEntityID, domain.DocumentStatusFailed); err != nil {
				return processed, fmt.Errorf("failed to mark document %s failed: %w", sub.DocumentID, err)
			}
			processed++
			continue
		}

		// Attempt retry
		err = w.subService.RetrySubmission(ctx, sub.ID, sub.LegalEntityID)
		if err == nil {
			processed++
		}
	}

	return processed, nil
}

// Start begins the background processing loop.
func (w *RetryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			processed, err := w.ProcessPending(ctx)
			if err != nil {
				slog.Error("ZATCA retry pass failed", "error", err)
			} else if processed > 0 {
				slog.Info("ZATCA retry pass completed", "processed", processed)
			}
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (w *RetryWorker) Stop() {
	close(w.stopChan)
}
