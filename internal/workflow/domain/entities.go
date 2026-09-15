package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WorkflowDefinition represents a configurable workflow definition.
type WorkflowDefinition struct {
	ID        uuid.UUID           `json:"id"`
	Code      string              `json:"code"`
	NameAr    string              `json:"name_ar"`
	NameEn    string              `json:"name_en"`
	Module    string              `json:"module"`
	IsActive  bool                `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Versions  []DefinitionVersion `json:"versions,omitempty"`
}

// DefinitionVersion represents an immutable published or draft version of a definition.
type DefinitionVersion struct {
	ID           uuid.UUID      `json:"id"`
	DefinitionID uuid.UUID      `json:"definition_id"`
	Version      int            `json:"version"`
	Status       string         `json:"status"` // DRAFT, PUBLISHED, ARCHIVED
	CreatedAt    time.Time      `json:"created_at"`
	PublishedAt  *time.Time     `json:"published_at,omitempty"`
	Steps        []WorkflowStep `json:"steps,omitempty"`
}

// WorkflowStep represents an approval stage within an immutable version.
type WorkflowStep struct {
	ID           uuid.UUID       `json:"id"`
	VersionID    uuid.UUID       `json:"version_id"`
	StepOrder    int             `json:"step_order"`
	Code         string          `json:"code"`
	NameAr       string          `json:"name_ar"`
	NameEn       string          `json:"name_en"`
	ApprovalType string          `json:"approval_type"` // ANY, ALL
	TimeoutHours int             `json:"timeout_hours"`
	Conditions   []StepCondition `json:"conditions,omitempty"`
	Approvers    []StepApprover  `json:"approvers,omitempty"`
}

// StepCondition represents conditional triggers for a step (e.g. amount > 50000).
type StepCondition struct {
	ID         uuid.UUID `json:"id"`
	StepID     uuid.UUID `json:"step_id"`
	FieldName  string    `json:"field_name"`
	Operator   string    `json:"operator"` // GT, GTE, LT, LTE, EQ, NEQ
	FieldValue string    `json:"field_value"`
}

// StepApprover defines a role or specific user authorized to approve a step.
type StepApprover struct {
	ID           uuid.UUID  `json:"id"`
	StepID       uuid.UUID  `json:"step_id"`
	ApproverType string     `json:"approver_type"` // ROLE, USER
	RoleID       *uuid.UUID `json:"role_id,omitempty"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
}

// WorkflowInstance represents a running approval workflow for a document.
type WorkflowInstance struct {
	ID            uuid.UUID       `json:"id"`
	DefinitionID  uuid.UUID       `json:"definition_id"`
	VersionID     uuid.UUID       `json:"version_id"`
	LegalEntityID uuid.UUID       `json:"legal_entity_id"`
	ResourceType  string          `json:"resource_type"`
	ResourceID    string          `json:"resource_id"`
	InitiatorID   uuid.UUID       `json:"initiator_id"`
	Status        string          `json:"status"` // PENDING, APPROVED, REJECTED, CANCELLED
	ContextData   json.RawMessage `json:"context_data"`
	CreatedAt     time.Time       `json:"created_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	Steps         []InstanceStep  `json:"steps,omitempty"`
}

// InstanceStep represents execution progress of an individual step.
type InstanceStep struct {
	ID          uuid.UUID         `json:"id"`
	InstanceID  uuid.UUID         `json:"instance_id"`
	StepID      uuid.UUID         `json:"step_id"`
	StepOrder   int               `json:"step_order"`
	Status      string            `json:"status"` // PENDING, IN_PROGRESS, APPROVED, REJECTED, SKIPPED
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Requests    []ApprovalRequest `json:"requests,omitempty"`
}

// ApprovalRequest represents an assigned task to an approver.
type ApprovalRequest struct {
	ID             uuid.UUID `json:"id"`
	InstanceStepID uuid.UUID `json:"instance_step_id"`
	AssignedUserID uuid.UUID `json:"assigned_user_id"`
	Status         string    `json:"status"` // PENDING, DECIDED, CANCELLED, DELEGATED
	CreatedAt      time.Time `json:"created_at"`
}

// ApprovalDecision represents a recorded approval or rejection decision.
type ApprovalDecision struct {
	ID         uuid.UUID `json:"id"`
	RequestID  uuid.UUID `json:"request_id"`
	InstanceID uuid.UUID `json:"instance_id"`
	ActorID    uuid.UUID `json:"actor_id"`
	Decision   string    `json:"decision"` // APPROVE, REJECT
	Notes      string    `json:"notes,omitempty"`
	DecidedAt  time.Time `json:"decided_at"`
}

// Delegation represents a temporary delegation of approval authority.
type Delegation struct {
	ID            uuid.UUID  `json:"id"`
	DelegatorID   uuid.UUID  `json:"delegator_id"`
	DelegateID    uuid.UUID  `json:"delegate_id"`
	LegalEntityID *uuid.UUID `json:"legal_entity_id,omitempty"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	IsActive      bool       `json:"is_active"`
	Reason        string     `json:"reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// PendingApprovalItem represents an inbox item with full document context.
type PendingApprovalItem struct {
	RequestID      uuid.UUID       `json:"request_id"`
	InstanceStepID uuid.UUID       `json:"instance_step_id"`
	InstanceID     uuid.UUID       `json:"instance_id"`
	ResourceType   string          `json:"resource_type"`
	ResourceID     string          `json:"resource_id"`
	LegalEntityID  uuid.UUID       `json:"legal_entity_id"`
	InitiatorID    uuid.UUID       `json:"initiator_id"`
	InitiatorName  string          `json:"initiator_name"`
	StepNameAr     string          `json:"step_name_ar"`
	StepNameEn     string          `json:"step_name_en"`
	ContextData    json.RawMessage `json:"context_data"`
	CreatedAt      time.Time       `json:"created_at"`
}
