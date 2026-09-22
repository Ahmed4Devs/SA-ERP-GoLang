package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/audit/domain"
	"erp/internal/audit/ports"

	"github.com/google/uuid"
)

type auditService struct {
	repo ports.AuditRepository
}

func NewAuditService(repo ports.AuditRepository) ports.AuditService {
	return &auditService{
		repo: repo,
	}
}

func (s *auditService) RecordEvent(ctx context.Context, cmd ports.RecordEventCommand) (*domain.AuditEvent, error) {
	canonicalPayload := domain.CanonicalizeJSON(cmd.Payload)

	eventToInsert := &domain.AuditEvent{
		UserID:        cmd.UserID,
		LegalEntityID: cmd.LegalEntityID,
		Action:        cmd.Action,
		ResourceType:  cmd.ResourceType,
		ResourceID:    cmd.ResourceID,
		Payload:       canonicalPayload,
		IPAddress:     cmd.IPAddress,
		UserAgent:     cmd.UserAgent,
	}

	changes := make([]domain.EntityChange, 0, len(cmd.Changes))
	for _, change := range cmd.Changes {
		beforeState := change.BeforeState
		if len(beforeState) == 0 {
			beforeState = []byte("null")
		}
		afterState := change.AfterState
		if len(afterState) == 0 {
			afterState = []byte("null")
		}

		changes = append(changes, domain.EntityChange{
			TableName:   change.TableName,
			RecordID:    change.RecordID,
			BeforeState: beforeState,
			AfterState:  afterState,
		})
	}

	insertedEvent, err := s.repo.AppendEvent(ctx, eventToInsert, changes)
	if err != nil {
		return nil, fmt.Errorf("failed to atomically persist audit event: %w", err)
	}
	return insertedEvent, nil
}

func (s *auditService) GetEvent(ctx context.Context, id uuid.UUID) (*domain.AuditEvent, error) {
	event, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}

	changes, err := s.repo.ListEntityChanges(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load entity changes: %w", err)
	}
	event.Changes = changes

	return event, nil
}

func (s *auditService) SearchEvents(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditEvent, int64, error) {
	return s.repo.ListEvents(ctx, filter)
}

func (s *auditService) VerifyChainIntegrity(ctx context.Context) (*domain.IntegrityVerificationResult, error) {
	const batchSize = 500
	var offset int32 = 0
	var totalRecords int64 = 0
	var verifiedRecords int64 = 0
	var prevExpectedHash = domain.GenesisHash

	for {
		batch, err := s.repo.ListEventsChronological(ctx, batchSize, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to load audit batch at offset %d: %w", offset, err)
		}

		if len(batch) == 0 {
			break
		}

		for _, event := range batch {
			totalRecords++

			// Check previous hash link
			if event.PreviousHash != prevExpectedHash {
				return &domain.IntegrityVerificationResult{
					TotalRecords:    totalRecords,
					VerifiedRecords: verifiedRecords,
					IsValid:         false,
					FailedAtEventID: &event.ID,
					Details: fmt.Sprintf("Hash link broken at event %s. Expected previous hash: %s, got: %s",
						event.ID, prevExpectedHash, event.PreviousHash),
					VerifiedAt: time.Now().UTC(),
				}, nil
			}

			// Validate record's own cryptographic hash
			if !domain.ValidateRecordHash(&event) {
				return &domain.IntegrityVerificationResult{
					TotalRecords:    totalRecords,
					VerifiedRecords: verifiedRecords,
					IsValid:         false,
					FailedAtEventID: &event.ID,
					Details:         fmt.Sprintf("Cryptographic record hash mismatch at event %s", event.ID),
					VerifiedAt:      time.Now().UTC(),
				}, nil
			}

			verifiedRecords++
			prevExpectedHash = event.RecordHash
		}

		offset += int32(len(batch))
		if len(batch) < batchSize {
			break
		}
	}

	return &domain.IntegrityVerificationResult{
		TotalRecords:    totalRecords,
		VerifiedRecords: verifiedRecords,
		IsValid:         true,
		Details:         "All audit events cryptographically verified. Hash chain is intact and untampered.",
		VerifiedAt:      time.Now().UTC(),
	}, nil
}

func (s *auditService) RecordSecurityEvent(ctx context.Context, cmd ports.RecordSecurityEventCommand) (*domain.SecurityEvent, error) {
	details := cmd.Details
	if len(details) == 0 {
		details = []byte("{}")
	}

	return s.repo.InsertSecurityEvent(ctx, &domain.SecurityEvent{
		UserID:    cmd.UserID,
		EventType: cmd.EventType,
		Status:    cmd.Status,
		Details:   details,
		IPAddress: cmd.IPAddress,
	})
}

func (s *auditService) ListSecurityEvents(ctx context.Context, userID *uuid.UUID, eventType, status string, limit, offset int32) ([]domain.SecurityEvent, error) {
	return s.repo.ListSecurityEvents(ctx, userID, eventType, status, limit, offset)
}

func (s *auditService) RecordAccessEvent(ctx context.Context, cmd ports.RecordAccessEventCommand) (*domain.AccessEvent, error) {
	return s.repo.InsertAccessEvent(ctx, &domain.AccessEvent{
		UserID:       cmd.UserID,
		ResourcePath: cmd.ResourcePath,
		Method:       cmd.Method,
		StatusCode:   cmd.StatusCode,
		IPAddress:    cmd.IPAddress,
		LatencyMS:    cmd.LatencyMS,
	})
}

func (s *auditService) ListAccessEvents(ctx context.Context, userID *uuid.UUID, resourcePath string, limit, offset int32) ([]domain.AccessEvent, error) {
	return s.repo.ListAccessEvents(ctx, userID, resourcePath, limit, offset)
}
