package ports

import (
	"context"
	"time"

	"erp/internal/workflow/domain"

	"github.com/google/uuid"
)

// WorkflowRepository defines persistence contracts for workflow definitions, instances, and delegations.
type WorkflowRepository interface {
	GetDefinitionByCode(ctx context.Context, code string) (*domain.WorkflowDefinition, error)
	GetDefinitionByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowDefinition, error)
	ListDefinitions(ctx context.Context) ([]domain.WorkflowDefinition, error)
	GetPublishedVersion(ctx context.Context, defID uuid.UUID) (*domain.DefinitionVersion, error)
	ListSteps(ctx context.Context, versionID uuid.UUID) ([]domain.WorkflowStep, error)

	CreateInstance(ctx context.Context, inst *domain.WorkflowInstance) (*domain.WorkflowInstance, error)
	GetInstanceByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowInstance, error)
	GetInstanceByResource(ctx context.Context, resourceType, resourceID string) (*domain.WorkflowInstance, error)
	UpdateInstanceStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) error

	CreateInstanceStep(ctx context.Context, step *domain.InstanceStep) (*domain.InstanceStep, error)
	GetInstanceStepByID(ctx context.Context, id uuid.UUID) (*domain.InstanceStep, error)
	LockInstanceStepByID(ctx context.Context, id uuid.UUID) (*domain.InstanceStep, error)
	ListInstanceSteps(ctx context.Context, instanceID uuid.UUID) ([]domain.InstanceStep, error)
	UpdateInstanceStepStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) error

	CreateApprovalRequest(ctx context.Context, req *domain.ApprovalRequest) (*domain.ApprovalRequest, error)
	GetApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error)
	LockApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error)
	ListApprovalRequestsByInstanceStep(ctx context.Context, instanceStepID uuid.UUID) ([]domain.ApprovalRequest, error)
	CancelPendingApprovalRequestsForStep(ctx context.Context, instanceStepID uuid.UUID) error
	ListPendingRequestsForUser(ctx context.Context, userID uuid.UUID) ([]domain.PendingApprovalItem, error)
	UpdateApprovalRequestStatus(ctx context.Context, id uuid.UUID, status string) error

	CreateApprovalDecision(ctx context.Context, dec *domain.ApprovalDecision) (*domain.ApprovalDecision, error)
	ListApprovalDecisions(ctx context.Context, instanceID uuid.UUID) ([]domain.ApprovalDecision, error)

	ListUserIDsByRoleInScope(ctx context.Context, roleID, leID uuid.UUID) ([]uuid.UUID, error)
	GetFallbackAdminUserID(ctx context.Context) (uuid.UUID, error)

	CreateDelegation(ctx context.Context, del *domain.Delegation) (*domain.Delegation, error)
	GetActiveDelegation(ctx context.Context, delegatorID, delegateID uuid.UUID, leID *uuid.UUID) (*domain.Delegation, error)
	ListActiveDelegationsForDelegate(ctx context.Context, delegateID uuid.UUID) ([]domain.Delegation, error)
	ListDelegationsByDelegator(ctx context.Context, delegatorID uuid.UUID) ([]domain.Delegation, error)
	RevokeDelegation(ctx context.Context, id, delegatorID uuid.UUID) error
}

// WorkflowService defines high-level business operations for starting, transitioning, and deciding workflows.
type WorkflowService interface {
	StartWorkflow(ctx context.Context, cmd StartWorkflowCommand) (*domain.WorkflowInstance, error)
	SubmitDecision(ctx context.Context, cmd SubmitDecisionCommand) (*domain.ApprovalDecision, error)
	GetInstance(ctx context.Context, id uuid.UUID) (*domain.WorkflowInstance, error)
	GetInstanceByResource(ctx context.Context, resourceType, resourceID string) (*domain.WorkflowInstance, error)
	ListPendingApprovals(ctx context.Context, userID uuid.UUID) ([]domain.PendingApprovalItem, error)
	ListDefinitions(ctx context.Context) ([]domain.WorkflowDefinition, error)
	CreateDelegation(ctx context.Context, cmd CreateDelegationCommand) (*domain.Delegation, error)
	ListMyDelegations(ctx context.Context, userID uuid.UUID) ([]domain.Delegation, error)
	RevokeDelegation(ctx context.Context, id, delegatorID uuid.UUID) error
}

// Command Objects
type StartWorkflowCommand struct {
	DefinitionCode string
	LegalEntityID  uuid.UUID
	ResourceType   string
	ResourceID     string
	InitiatorID    uuid.UUID
	ContextData    map[string]any
}

type SubmitDecisionCommand struct {
	RequestID uuid.UUID
	ActorID   uuid.UUID
	Decision  string // APPROVE, REJECT
	Notes     string
}

type CreateDelegationCommand struct {
	DelegatorID   uuid.UUID
	DelegateID    uuid.UUID
	LegalEntityID *uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	Reason        string
}
