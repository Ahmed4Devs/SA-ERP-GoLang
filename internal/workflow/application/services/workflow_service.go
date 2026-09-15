package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"erp/internal/workflow/domain"
	"erp/internal/workflow/ports"

	"github.com/google/uuid"
)

type workflowService struct {
	repo ports.WorkflowRepository
}

func NewWorkflowService(repo ports.WorkflowRepository) ports.WorkflowService {
	return &workflowService{
		repo: repo,
	}
}

func (s *workflowService) StartWorkflow(ctx context.Context, cmd ports.StartWorkflowCommand) (*domain.WorkflowInstance, error) {
	// 1. Fetch Definition
	def, err := s.repo.GetDefinitionByCode(ctx, cmd.DefinitionCode)
	if err != nil {
		return nil, fmt.Errorf("workflow definition not found: %w", err)
	}

	// 2. Fetch Published Version
	ver, err := s.repo.GetPublishedVersion(ctx, def.ID)
	if err != nil {
		return nil, fmt.Errorf("no published version available for definition %s: %w", def.Code, err)
	}

	// 3. Fetch Version Steps
	steps, err := s.repo.ListSteps(ctx, ver.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load steps for workflow version %s: %w", ver.ID, err)
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow version %s has no steps defined", ver.ID)
	}

	// 4. Create Instance
	contextBytes, err := json.Marshal(cmd.ContextData)
	if err != nil {
		contextBytes = []byte("{}")
	}

	instance := &domain.WorkflowInstance{
		DefinitionID:  def.ID,
		VersionID:     ver.ID,
		LegalEntityID: cmd.LegalEntityID,
		ResourceType:  cmd.ResourceType,
		ResourceID:    cmd.ResourceID,
		InitiatorID:   cmd.InitiatorID,
		Status:        "PENDING",
		ContextData:   contextBytes,
	}

	createdInst, err := s.repo.CreateInstance(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow instance: %w", err)
	}

	// 5. Activate First Eligible Step
	for _, step := range steps {
		eligible, err := domain.EvaluateConditions(step.Conditions, cmd.ContextData)
		if err != nil {
			return nil, err
		}
		if !eligible {
			// Skip step as conditions not met
			continue
		}

		// Activate this step
		instStep := &domain.InstanceStep{
			InstanceID: createdInst.ID,
			StepID:     step.ID,
			StepOrder:  step.StepOrder,
			Status:     "IN_PROGRESS",
		}
		createdStep, err := s.repo.CreateInstanceStep(ctx, instStep)
		if err != nil {
			return nil, fmt.Errorf("failed to create instance step: %w", err)
		}

		// Assign approvers
		if err := s.assignStepApprovers(ctx, createdStep.ID, step, cmd.InitiatorID, cmd.LegalEntityID); err != nil {
			return nil, err
		}

		createdInst.Steps = append(createdInst.Steps, *createdStep)
		break
	}

	return createdInst, nil
}

func (s *workflowService) assignStepApprovers(ctx context.Context, instanceStepID uuid.UUID, step domain.WorkflowStep, initiatorID, legalEntityID uuid.UUID) error {
	assignedCount := 0

	for _, app := range step.Approvers {
		if app.ApproverType == "USER" && app.UserID != nil {
			req := &domain.ApprovalRequest{
				InstanceStepID: instanceStepID,
				AssignedUserID: *app.UserID,
				Status:         "PENDING",
			}
			if _, err := s.repo.CreateApprovalRequest(ctx, req); err != nil {
				return err
			}
			assignedCount++
		} else if app.ApproverType == "ROLE" && app.RoleID != nil {
			userIDs, err := s.repo.ListUserIDsByRoleInScope(ctx, *app.RoleID, legalEntityID)
			if err == nil {
				for _, uID := range userIDs {
					req := &domain.ApprovalRequest{
						InstanceStepID: instanceStepID,
						AssignedUserID: uID,
						Status:         "PENDING",
					}
					if _, err := s.repo.CreateApprovalRequest(ctx, req); err == nil {
						assignedCount++
					}
				}
			}
		}
	}

	// Fallback to active system administrator if no role assignments yet exist
	if assignedCount == 0 {
		fallbackAdminID, err := s.repo.GetFallbackAdminUserID(ctx)
		if err == nil && fallbackAdminID != uuid.Nil {
			req := &domain.ApprovalRequest{
				InstanceStepID: instanceStepID,
				AssignedUserID: fallbackAdminID,
				Status:         "PENDING",
			}
			_, _ = s.repo.CreateApprovalRequest(ctx, req)
		}
	}

	return nil
}

func (s *workflowService) SubmitDecision(ctx context.Context, cmd ports.SubmitDecisionCommand) (*domain.ApprovalDecision, error) {
	// 1. Load Request
	req, err := s.repo.GetApprovalRequestByID(ctx, cmd.RequestID)
	if err != nil {
		return nil, err
	}
	if req.Status != "PENDING" {
		return nil, domain.ErrRequestAlreadyDecided
	}

	// 2. Load Instance Step & Instance
	instSteps, err := s.repo.ListInstanceSteps(ctx, req.InstanceStepID) // Load step via list or get
	_ = instSteps

	// Load instance
	decisions, _ := s.repo.ListApprovalDecisions(ctx, req.InstanceStepID)
	_ = decisions

	// To get the instance ID, fetch all pending requests or trace instance
	// Or we query approval decision directly
	// Let's verify authorization and Maker/Checker:
	// Find the instance
	instList, err := s.repo.ListPendingRequestsForUser(ctx, req.AssignedUserID)
	var initiatorID uuid.UUID
	var instanceID uuid.UUID
	var legalEntityID uuid.UUID
	var stepOrder int
	var contextDataMap map[string]any

	for _, item := range instList {
		if item.RequestID == req.ID {
			initiatorID = item.InitiatorID
			instanceID = item.InstanceID
			legalEntityID = item.LegalEntityID
			_ = json.Unmarshal(item.ContextData, &contextDataMap)
			break
		}
	}

	// If not found in pending list, fetch directly
	if instanceID == uuid.Nil {
		// Fallback lookup
		instanceID = req.InstanceStepID // fallback identifier
	}

	// 3. Maker-Checker Enforcement (cannot_approve_own_document)
	if initiatorID != uuid.Nil && initiatorID == cmd.ActorID {
		return nil, domain.ErrMakerCannotApprove
	}

	// 4. Verify Authorization (Direct or Delegated)
	if req.AssignedUserID != cmd.ActorID {
		// Check delegation
		del, err := s.repo.GetActiveDelegation(ctx, req.AssignedUserID, cmd.ActorID, &legalEntityID)
		if err != nil || del == nil {
			return nil, domain.ErrUnauthorizedApprover
		}
	}

	// 5. Record Decision
	dec := &domain.ApprovalDecision{
		RequestID:  req.ID,
		InstanceID: instanceID,
		ActorID:    cmd.ActorID,
		Decision:   cmd.Decision,
		Notes:      cmd.Notes,
		DecidedAt:  time.Now().UTC(),
	}

	persistedDec, err := s.repo.CreateApprovalDecision(ctx, dec)
	if err != nil {
		return nil, fmt.Errorf("failed to record approval decision: %w", err)
	}

	// 6. Update Request Status
	_ = s.repo.UpdateApprovalRequestStatus(ctx, req.ID, "DECIDED")

	// 7. Progress or Complete Workflow
	now := time.Now().UTC()
	if cmd.Decision == "REJECT" {
		_ = s.repo.UpdateInstanceStepStatus(ctx, req.InstanceStepID, "REJECTED", &now)
		_ = s.repo.UpdateInstanceStatus(ctx, instanceID, "REJECTED", &now)
	} else if cmd.Decision == "APPROVE" {
		_ = s.repo.UpdateInstanceStepStatus(ctx, req.InstanceStepID, "APPROVED", &now)

		// Check if there are further steps in this workflow definition
		inst, err := s.repo.GetInstanceByID(ctx, instanceID)
		if err == nil && inst != nil {
			steps, err := s.repo.ListSteps(ctx, inst.VersionID)
			if err == nil {
				var nextStep *domain.WorkflowStep
				for _, st := range steps {
					if st.StepOrder > stepOrder {
						eligible, _ := domain.EvaluateConditions(st.Conditions, contextDataMap)
						if eligible {
							nextStep = &st
							break
						}
					}
				}

				if nextStep != nil {
					// Activate next step
					newStep := &domain.InstanceStep{
						InstanceID: inst.ID,
						StepID:     nextStep.ID,
						StepOrder:  nextStep.StepOrder,
						Status:     "IN_PROGRESS",
					}
					createdNextStep, _ := s.repo.CreateInstanceStep(ctx, newStep)
					if createdNextStep != nil {
						_ = s.assignStepApprovers(ctx, createdNextStep.ID, *nextStep, inst.InitiatorID, inst.LegalEntityID)
					}
				} else {
					// All steps completed -> mark workflow APPROVED
					_ = s.repo.UpdateInstanceStatus(ctx, inst.ID, "APPROVED", &now)
				}
			}
		}
	}

	return persistedDec, nil
}

func (s *workflowService) GetInstance(ctx context.Context, id uuid.UUID) (*domain.WorkflowInstance, error) {
	return s.repo.GetInstanceByID(ctx, id)
}

func (s *workflowService) GetInstanceByResource(ctx context.Context, resourceType, resourceID string) (*domain.WorkflowInstance, error) {
	return s.repo.GetInstanceByResource(ctx, resourceType, resourceID)
}

func (s *workflowService) ListPendingApprovals(ctx context.Context, userID uuid.UUID) ([]domain.PendingApprovalItem, error) {
	// 1. Direct assignments
	direct, err := s.repo.ListPendingRequestsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Delegated assignments (tasks assigned to users who delegated to this user)
	delegations, err := s.repo.ListActiveDelegationsForDelegate(ctx, userID)
	if err == nil && len(delegations) > 0 {
		seen := make(map[uuid.UUID]bool)
		for _, d := range direct {
			seen[d.RequestID] = true
		}

		for _, del := range delegations {
			delegatedItems, err := s.repo.ListPendingRequestsForUser(ctx, del.DelegatorID)
			if err == nil {
				for _, item := range delegatedItems {
					if !seen[item.RequestID] {
						seen[item.RequestID] = true
						direct = append(direct, item)
					}
				}
			}
		}
	}

	return direct, nil
}

func (s *workflowService) ListDefinitions(ctx context.Context) ([]domain.WorkflowDefinition, error) {
	return s.repo.ListDefinitions(ctx)
}

func (s *workflowService) CreateDelegation(ctx context.Context, cmd ports.CreateDelegationCommand) (*domain.Delegation, error) {
	if cmd.DelegatorID == cmd.DelegateID {
		return nil, domain.ErrSelfDelegationForbidden
	}
	if !cmd.EndDate.After(cmd.StartDate) {
		return nil, domain.ErrInvalidDelegationPeriod
	}

	del := &domain.Delegation{
		DelegatorID:   cmd.DelegatorID,
		DelegateID:    cmd.DelegateID,
		LegalEntityID: cmd.LegalEntityID,
		StartDate:     cmd.StartDate,
		EndDate:       cmd.EndDate,
		IsActive:      true,
		Reason:        cmd.Reason,
	}

	return s.repo.CreateDelegation(ctx, del)
}

func (s *workflowService) ListMyDelegations(ctx context.Context, userID uuid.UUID) ([]domain.Delegation, error) {
	return s.repo.ListDelegationsByDelegator(ctx, userID)
}

func (s *workflowService) RevokeDelegation(ctx context.Context, id, delegatorID uuid.UUID) error {
	return s.repo.RevokeDelegation(ctx, id, delegatorID)
}
