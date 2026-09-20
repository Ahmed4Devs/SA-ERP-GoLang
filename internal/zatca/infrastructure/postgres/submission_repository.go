package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SubmissionRepository struct {
	db *database.DB
}

func NewSubmissionRepository(db *database.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) CreateSubmission(ctx context.Context, sub *domain.ZATCASubmission) (*domain.ZATCASubmission, error) {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	if sub.SubmittedAt.IsZero() {
		sub.SubmittedAt = time.Now()
	}

	query := `
		INSERT INTO zatca.submissions (
			id, document_id, submission_mode, environment, zatca_status,
			zatca_request_id, clearance_status, reporting_status, submitted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, document_id, submission_mode, environment, zatca_status, zatca_request_id, clearance_status, reporting_status, submitted_at
	`

	var clearanceStatus, reportingStatus *string
	if sub.SubmissionMode == domain.SubmissionModeClearance {
		clearanceStatus = &sub.Status
	} else {
		reportingStatus = &sub.Status
	}

	var savedID, docID uuid.UUID
	var mode, environment, status string
	var reqID *string
	var cStatus, rStatus *string
	var submittedAt time.Time

	err := r.db.Pool.QueryRow(ctx, query,
		sub.ID,
		sub.DocumentID,
		string(sub.SubmissionMode),
		string(sub.Environment),
		sub.Status,
		sub.ZATCARequestID,
		clearanceStatus,
		reportingStatus,
		sub.SubmittedAt,
	).Scan(&savedID, &docID, &mode, &environment, &status, &reqID, &cStatus, &rStatus, &submittedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	result := *sub
	result.ID = savedID
	result.SubmittedAt = submittedAt
	return &result, nil
}

func (r *SubmissionRepository) UpdateSubmission(ctx context.Context, sub *domain.ZATCASubmission) (*domain.ZATCASubmission, error) {
	query := `
		UPDATE zatca.submissions
		SET zatca_status = $2, zatca_request_id = $3
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, sub.ID, sub.Status, sub.ZATCARequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to update submission: %w", err)
	}
	return sub, nil
}

func (r *SubmissionRepository) GetSubmissionByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCASubmission, error) {
	query := `
		SELECT s.id, s.document_id, d.legal_entity_id, s.submission_mode, s.environment, s.zatca_status,
		       s.zatca_request_id, s.clearance_status, s.reporting_status, s.submitted_at
		FROM zatca.submissions s
		JOIN zatca.documents d ON d.id = s.document_id
		WHERE s.id = $1 AND d.legal_entity_id = $2
	`

	var sub domain.ZATCASubmission
	var mode, environment, status string
	var reqID, cStatus, rStatus *string

	err := r.db.Pool.QueryRow(ctx, query, id, legalEntityID).Scan(
		&sub.ID,
		&sub.DocumentID,
		&sub.LegalEntityID,
		&mode,
		&environment,
		&status,
		&reqID,
		&cStatus,
		&rStatus,
		&sub.SubmittedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("submission not found")
		}
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	sub.SubmissionMode = domain.SubmissionMode(mode)
	sub.Environment = domain.Environment(environment)
	sub.Status = status
	if reqID != nil {
		sub.ZATCARequestID = *reqID
	}

	// Fetch attempts
	attempts, err := r.ListAttempts(ctx, sub.ID)
	if err == nil {
		sub.Attempts = attempts
		sub.AttemptsCount = len(attempts)
	}

	return &sub, nil
}

func (r *SubmissionRepository) GetSubmissionByDocumentID(ctx context.Context, docID, legalEntityID uuid.UUID) (*domain.ZATCASubmission, error) {
	query := `
		SELECT s.id, s.document_id, d.legal_entity_id, s.submission_mode, s.environment, s.zatca_status,
		       s.zatca_request_id, s.clearance_status, s.reporting_status, s.submitted_at
		FROM zatca.submissions s
		JOIN zatca.documents d ON d.id = s.document_id
		WHERE s.document_id = $1 AND d.legal_entity_id = $2
		ORDER BY s.submitted_at DESC
		LIMIT 1
	`

	var sub domain.ZATCASubmission
	var mode, environment, status string
	var reqID, cStatus, rStatus *string

	err := r.db.Pool.QueryRow(ctx, query, docID, legalEntityID).Scan(
		&sub.ID,
		&sub.DocumentID,
		&sub.LegalEntityID,
		&mode,
		&environment,
		&status,
		&reqID,
		&cStatus,
		&rStatus,
		&sub.SubmittedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get submission by doc ID: %w", err)
	}

	sub.SubmissionMode = domain.SubmissionMode(mode)
	sub.Environment = domain.Environment(environment)
	sub.Status = status
	if reqID != nil {
		sub.ZATCARequestID = *reqID
	}

	attempts, err := r.ListAttempts(ctx, sub.ID)
	if err == nil {
		sub.Attempts = attempts
		sub.AttemptsCount = len(attempts)
	}

	return &sub, nil
}

func (r *SubmissionRepository) ListSubmissions(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.ZATCASubmission, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT s.id, s.document_id, d.legal_entity_id, s.submission_mode, s.environment, s.zatca_status,
		       s.zatca_request_id, s.clearance_status, s.reporting_status, s.submitted_at
		FROM zatca.submissions s
		JOIN zatca.documents d ON d.id = s.document_id
		WHERE d.legal_entity_id = $1
	`
	args := []any{filter.LegalEntityID}
	argIdx := 2

	if filter.DocumentID != nil {
		query += fmt.Sprintf(" AND s.document_id = $%d", argIdx)
		args = append(args, *filter.DocumentID)
		argIdx++
	}
	if filter.SubmissionMode != nil {
		query += fmt.Sprintf(" AND s.submission_mode = $%d", argIdx)
		args = append(args, string(*filter.SubmissionMode))
		argIdx++
	}
	if filter.Environment != nil {
		query += fmt.Sprintf(" AND s.environment = $%d", argIdx)
		args = append(args, string(*filter.Environment))
		argIdx++
	}
	if filter.Status != nil {
		query += fmt.Sprintf(" AND s.zatca_status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY s.submitted_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list submissions: %w", err)
	}
	defer rows.Close()

	var results []*domain.ZATCASubmission
	for rows.Next() {
		var sub domain.ZATCASubmission
		var mode, environment, status string
		var reqID, cStatus, rStatus *string

		if err := rows.Scan(
			&sub.ID,
			&sub.DocumentID,
			&sub.LegalEntityID,
			&mode,
			&environment,
			&status,
			&reqID,
			&cStatus,
			&rStatus,
			&sub.SubmittedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan submission row: %w", err)
		}

		sub.SubmissionMode = domain.SubmissionMode(mode)
		sub.Environment = domain.Environment(environment)
		sub.Status = status
		if reqID != nil {
			sub.ZATCARequestID = *reqID
		}
		results = append(results, &sub)
	}

	return results, nil
}

func (r *SubmissionRepository) RecordAttempt(ctx context.Context, attempt *domain.SubmissionAttempt) error {
	if attempt.ID == uuid.Nil {
		attempt.ID = uuid.New()
	}
	if attempt.AttemptedAt.IsZero() {
		attempt.AttemptedAt = time.Now()
	}

	query := `
		INSERT INTO zatca.submission_attempts (
			id, submission_id, attempt_number, http_status, response_payload, error_message, attempted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var payloadJSON []byte
	if attempt.ResponsePayload != "" {
		payloadJSON = []byte(attempt.ResponsePayload)
		if !json.Valid(payloadJSON) {
			payloadJSON, _ = json.Marshal(map[string]string{"raw": attempt.ResponsePayload})
		}
	} else {
		payloadJSON = []byte("{}")
	}

	_, err := r.db.Pool.Exec(ctx, query,
		attempt.ID,
		attempt.SubmissionID,
		attempt.AttemptNumber,
		attempt.HTTPStatusCode,
		payloadJSON,
		attempt.ErrorMessage,
		attempt.AttemptedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record submission attempt: %w", err)
	}
	return nil
}

func (r *SubmissionRepository) ListAttempts(ctx context.Context, submissionID uuid.UUID) ([]domain.SubmissionAttempt, error) {
	query := `
		SELECT id, submission_id, attempt_number, http_status, response_payload, error_message, attempted_at
		FROM zatca.submission_attempts
		WHERE submission_id = $1
		ORDER BY attempt_number ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list submission attempts: %w", err)
	}
	defer rows.Close()

	var attempts []domain.SubmissionAttempt
	for rows.Next() {
		var a domain.SubmissionAttempt
		var rawPayload []byte
		var errMsg *string

		if err := rows.Scan(
			&a.ID,
			&a.SubmissionID,
			&a.AttemptNumber,
			&a.HTTPStatusCode,
			&rawPayload,
			&errMsg,
			&a.AttemptedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attempt row: %w", err)
		}

		if rawPayload != nil {
			a.ResponsePayload = string(rawPayload)
		}
		if errMsg != nil {
			a.ErrorMessage = *errMsg
		}
		attempts = append(attempts, a)
	}

	return attempts, nil
}

func (r *SubmissionRepository) GetPendingRetrySubmissions(ctx context.Context, limit int) ([]*domain.ZATCASubmission, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT s.id, s.document_id, d.legal_entity_id, s.submission_mode, s.environment, s.zatca_status,
		       s.zatca_request_id, s.clearance_status, s.reporting_status, s.submitted_at
		FROM zatca.submissions s
		JOIN zatca.documents d ON d.id = s.document_id
		WHERE s.zatca_status = 'RETRY_PENDING'
		ORDER BY s.submitted_at ASC
		LIMIT $1
	`
	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending retries: %w", err)
	}
	defer rows.Close()

	var results []*domain.ZATCASubmission
	for rows.Next() {
		var sub domain.ZATCASubmission
		var mode, environment, status string
		var reqID, cStatus, rStatus *string

		if err := rows.Scan(
			&sub.ID,
			&sub.DocumentID,
			&sub.LegalEntityID,
			&mode,
			&environment,
			&status,
			&reqID,
			&cStatus,
			&rStatus,
			&sub.SubmittedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan retry row: %w", err)
		}

		sub.SubmissionMode = domain.SubmissionMode(mode)
		sub.Environment = domain.Environment(environment)
		sub.Status = status
		if reqID != nil {
			sub.ZATCARequestID = *reqID
		}
		results = append(results, &sub)
	}

	return results, nil
}

var _ ports.SubmissionRepository = (*SubmissionRepository)(nil)
