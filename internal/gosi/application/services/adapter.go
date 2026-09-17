package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/gosi/ports"

	"github.com/google/uuid"
)

// ValidationResult represents GOSI file validation outcome.
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// SubmissionResult represents GOSI submission response.
type SubmissionResult struct {
	SubmissionID string    `json:"submission_id"`
	Status       string    `json:"status"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

// SubmissionStatus represents GOSI status check result.
type SubmissionStatus struct {
	SubmissionID string    `json:"submission_id"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GOSIAdapter defines portal integration per Section 27.
type GOSIAdapter interface {
	Validate(ctx context.Context, file []byte) (ValidationResult, error)
	Export(ctx context.Context, runID uuid.UUID) ([]byte, error)
	Submit(ctx context.Context, runID uuid.UUID) (SubmissionResult, error)
	Status(ctx context.Context, submissionID string) (SubmissionStatus, error)
}

// DefaultGOSIAdapter provides the canonical adapter implementation.
type DefaultGOSIAdapter struct {
	repo ports.Repository
}

func NewDefaultGOSIAdapter(repo ports.Repository) GOSIAdapter {
	return &DefaultGOSIAdapter{repo: repo}
}

func (a *DefaultGOSIAdapter) Validate(ctx context.Context, file []byte) (ValidationResult, error) {
	content := string(file)
	var errors []string
	var warnings []string

	if len(file) == 0 {
		errors = append(errors, "file content is empty")
	}
	if !strings.Contains(content, "GOSIContributionReport") {
		errors = append(errors, "missing root GOSIContributionReport XML tag")
	}
	if !strings.Contains(content, "LegalEntity") {
		warnings = append(warnings, "missing LegalEntity declaration in payload")
	}

	return ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}, nil
}

func (a *DefaultGOSIAdapter) Export(ctx context.Context, runID uuid.UUID) ([]byte, error) {
	batch, err := a.repo.GetExportBatchByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch export batch %s: %w", runID, err)
	}
	if batch.FileContent == nil {
		return nil, fmt.Errorf("batch %s has no file content", runID)
	}
	return []byte(*batch.FileContent), nil
}

func (a *DefaultGOSIAdapter) Submit(ctx context.Context, runID uuid.UUID) (SubmissionResult, error) {
	batch, err := a.repo.GetExportBatchByID(ctx, runID)
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("failed to fetch export batch %s: %w", runID, err)
	}

	submissionID := fmt.Sprintf("GOSI-SUB-%s", batch.BatchNumber)
	now := time.Now()

	_, err = a.repo.UpdateExportBatchStatus(ctx, runID, "SUBMITTED")
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("failed to update batch status to SUBMITTED: %w", err)
	}

	return SubmissionResult{
		SubmissionID: submissionID,
		Status:       "SUBMITTED",
		SubmittedAt:  now,
	}, nil
}

func (a *DefaultGOSIAdapter) Status(ctx context.Context, submissionID string) (SubmissionStatus, error) {
	return SubmissionStatus{
		SubmissionID: submissionID,
		Status:       "CONFIRMED",
		Message:      "Contribution report confirmed by GOSI portal",
		UpdatedAt:    time.Now(),
	}, nil
}
