package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents a canonical, immutable, hash-chained audit event per Section 40.28.
type AuditEvent struct {
	ID            uuid.UUID       `json:"id"`
	UserID        *uuid.UUID      `json:"user_id,omitempty"`
	LegalEntityID *uuid.UUID      `json:"legal_entity_id,omitempty"`
	Action        string          `json:"action"`
	ResourceType  string          `json:"resource_type"`
	ResourceID    string          `json:"resource_id"`
	Payload       json.RawMessage `json:"payload"`
	IPAddress     string          `json:"ip_address,omitempty"`
	UserAgent     string          `json:"user_agent,omitempty"`
	PreviousHash  string          `json:"previous_hash"`
	RecordHash    string          `json:"record_hash"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Changes       []EntityChange  `json:"changes,omitempty"`
}

// EntityChange represents the state diff (before and after) for a specific entity table record.
type EntityChange struct {
	ID          uuid.UUID       `json:"id"`
	EventID     uuid.UUID       `json:"event_id"`
	TableName   string          `json:"table_name"`
	RecordID    string          `json:"record_id"`
	BeforeState json.RawMessage `json:"before_state,omitempty"`
	AfterState  json.RawMessage `json:"after_state,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SecurityEvent represents a security-relevant event (authentication, permission rejection, abuse).
type SecurityEvent struct {
	ID         uuid.UUID       `json:"id"`
	UserID     *uuid.UUID      `json:"user_id,omitempty"`
	EventType  string          `json:"event_type"`
	Status     string          `json:"status"`
	Details    json.RawMessage `json:"details"`
	IPAddress  string          `json:"ip_address,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
}

// AccessEvent represents an HTTP request access log record.
type AccessEvent struct {
	ID           uuid.UUID  `json:"id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	ResourcePath string     `json:"resource_path"`
	Method       string     `json:"method"`
	StatusCode   int        `json:"status_code"`
	IPAddress    string     `json:"ip_address,omitempty"`
	LatencyMS    int        `json:"latency_ms"`
	OccurredAt   time.Time  `json:"occurred_at"`
}

// AuditFilter contains criteria for querying audit events.
type AuditFilter struct {
	LegalEntityID *uuid.UUID `json:"legal_entity_id,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	Action        string     `json:"action,omitempty"`
	ResourceType  string     `json:"resource_type,omitempty"`
	ResourceID    string     `json:"resource_id,omitempty"`
	FromDate      *time.Time `json:"from_date,omitempty"`
	ToDate        *time.Time `json:"to_date,omitempty"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// IntegrityVerificationResult summarizes the audit hash chain verification.
type IntegrityVerificationResult struct {
	TotalRecords    int64      `json:"total_records"`
	VerifiedRecords int64      `json:"verified_records"`
	IsValid         bool       `json:"is_valid"`
	FailedAtEventID *uuid.UUID `json:"failed_at_event_id,omitempty"`
	Details         string     `json:"details"`
	VerifiedAt      time.Time  `json:"verified_at"`
}
