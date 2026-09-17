package jobs

import (
	"context"
	"fmt"
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
	cfg         RetryWorkerConfig
	subRepo     ports.SubmissionRepository
	docRepo     ports.Repository
	subService  ports.SubmissionService
	stopChan    chan struct{}
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

		attempts, _ := w.subRepo.ListAttempts(ctx, sub.ID)
		if len(attempts) >= w.cfg.MaxRetries {
			// Mark permanently failed
			sub.Status = "FAILED"
			_, _ = w.subRepo.UpdateSubmission(ctx, sub)
			_, _ = w.docRepo.UpdateDocumentStatus(ctx, sub.DocumentID, sub.LegalEntityID, domain.DocumentStatusFailed)
			processed++
			continue
		}

		// Attempt retry
		err := w.subService.RetrySubmission(ctx, sub.ID, sub.LegalEntityID)
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
			_, _ = w.ProcessPending(ctx)
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
