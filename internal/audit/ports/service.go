package ports

import (
	"context"

	"erp/internal/audit/domain"

	"github.com/google/uuid"
)

// AuditRepository defines the persistence operations for audit trail data.
type AuditRepository interface {
	AppendEvent(ctx context.Context, event *domain.AuditEvent, changes []domain.EntityChange) (*domain.AuditEvent, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (*domain.AuditEvent, error)
	ListEvents(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditEvent, int64, error)
	ListEventsChronological(ctx context.Context, limit, offset int32) ([]domain.AuditEvent, error)

	ListEntityChanges(ctx context.Context, eventID uuid.UUID) ([]domain.EntityChange, error)

	InsertSecurityEvent(ctx context.Context, event *domain.SecurityEvent) (*domain.SecurityEvent, error)
	ListSecurityEvents(ctx context.Context, userID *uuid.UUID, eventType, status string, limit, offset int32) ([]domain.SecurityEvent, error)

	InsertAccessEvent(ctx context.Context, event *domain.AccessEvent) (*domain.AccessEvent, error)
	ListAccessEvents(ctx context.Context, userID *uuid.UUID, resourcePath string, limit, offset int32) ([]domain.AccessEvent, error)
}

// AuditService defines the application operations for managing and verifying the audit trail.
type AuditService interface {
	RecordEvent(ctx context.Context, cmd RecordEventCommand) (*domain.AuditEvent, error)
	GetEvent(ctx context.Context, id uuid.UUID) (*domain.AuditEvent, error)
	SearchEvents(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditEvent, int64, error)
	VerifyChainIntegrity(ctx context.Context) (*domain.IntegrityVerificationResult, error)

	RecordSecurityEvent(ctx context.Context, cmd RecordSecurityEventCommand) (*domain.SecurityEvent, error)
	ListSecurityEvents(ctx context.Context, userID *uuid.UUID, eventType, status string, limit, offset int32) ([]domain.SecurityEvent, error)

	RecordAccessEvent(ctx context.Context, cmd RecordAccessEventCommand) (*domain.AccessEvent, error)
	ListAccessEvents(ctx context.Context, userID *uuid.UUID, resourcePath string, limit, offset int32) ([]domain.AccessEvent, error)
}

// Commands
type RecordEventCommand struct {
	UserID        *uuid.UUID
	LegalEntityID *uuid.UUID
	Action        string
	ResourceType  string
	ResourceID    string
	Payload       []byte
	IPAddress     string
	UserAgent     string
	Changes       []EntityChangeCommand
}

type EntityChangeCommand struct {
	TableName   string
	RecordID    string
	BeforeState []byte
	AfterState  []byte
}

type RecordSecurityEventCommand struct {
	UserID    *uuid.UUID
	EventType string
	Status    string
	Details   []byte
	IPAddress string
}

type RecordAccessEventCommand struct {
	UserID       *uuid.UUID
	ResourcePath string
	Method       string
	StatusCode   int
	IPAddress    string
	LatencyMS    int
}
