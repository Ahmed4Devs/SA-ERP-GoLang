package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/workflow/domain"
	"erp/internal/workflow/ports"

	"github.com/google/uuid"
)

type workflowService struct {
	repo ports.WorkflowRepository
	tx   transaction.Manager
}

func NewWorkflowService(repo ports.WorkflowRepository, db *database.DB) ports.WorkflowService {
	return &workflowService{
		repo: repo,
		tx:   transaction.NewManager(db.Pool),
	}
}

func (s *workflowService) StartWorkflow(ctx context.Context, cmd ports.StartWorkflowCommand) (*domain.WorkflowInstance, error) {
	cmd.DefinitionCode = strings.TrimSpace(cmd.DefinitionCode)
	cmd.ResourceType = strings.TrimSpace(cmd.ResourceType)
	cmd.ResourceID = strings.TrimSpace(cmd.ResourceID)
	if cmd.DefinitionCode == "" || cmd.ResourceType == "" || cmd.ResourceID == "" || cmd.LegalEntityID == uuid.Nil || cmd.InitiatorID == uuid.Nil {
		return nil, domain.ErrInvalidWorkflowCommand
	}
	if cmd.ContextData == nil {
		cmd.ContextData = map[string]any{}
	}

	def, err := s.repo.GetDefinitionByCode(ctx, cmd.DefinitionCode)
	if err != nil {
		return nil, fmt.Errorf("workflow definition not found: %w", err)
	}

	ver, err := s.repo.GetPublishedVersion(ctx, def.ID)
	if err != nil {
		return nil, fmt.Errorf("no published version available for definition %s: %w", def.Code, err)
	}

	steps, err := s.repo.ListSteps(ctx, ver.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load steps for workflow version %s: %w", ver.ID, err)
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow version %s has no steps defined", ver.ID)
	}

	contextBytes, err := json.Marshal(cmd.ContextData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode workflow context: %w", err)
	}

	var firstStep *domain.WorkflowStep
	for i := range steps {
		eligible, err := domain.EvaluateConditions(steps[i].Conditions, cmd.ContextData)
		if err != nil {
			return nil, err
		}
		if eligible {
			firstStep = &steps[i]
			break
		}
	}
	if firstStep == nil {
		return nil, domain.ErrNoEligibleWorkflowStep
	}

	instance := domain.WorkflowInstance{
		DefinitionID:  def.ID,
		VersionID:     ver.ID,
		LegalEntityID: cmd.LegalEntityID,
		ResourceType:  cmd.ResourceType,
		ResourceID:    cmd.ResourceID,
		InitiatorID:   cmd.InitiatorID,
		Status:        "PENDING",
		ContextData:   contextBytes,
	}

	var createdInst *domain.WorkflowInstance
	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdInst, err = s.repo.CreateInstance(txCtx, &instance)
		if err != nil {
			return fmt.Errorf("failed to create workflow instance: %w", err)
		}

		instStep := &domain.InstanceStep{
			InstanceID: createdInst.ID,
			StepID:     firstStep.ID,
			StepOrder:  firstStep.StepOrder,
			Status:     "IN_PROGRESS",
		}
		createdStep, err := s.repo.CreateInstanceStep(txCtx, instStep)
		if err != nil {
			return fmt.Errorf("failed to create instance step: %w", err)
		}

		if err := s.assignStepApprovers(txCtx, createdStep.ID, *firstStep, cmd.LegalEntityID); err != nil {
			return fmt.Errorf("failed to assign workflow approvers: %w", err)
		}

		createdInst.Steps = append(createdInst.Steps, *createdStep)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return createdInst, nil
}

func (s *workflowService) assignStepApprovers(ctx context.Context, instanceStepID uuid.UUID, step domain.WorkflowStep, legalEntityID uuid.UUID) error {
	assigned := make(map[uuid.UUID]struct{})
	assign := func(userID uuid.UUID) error {
		if userID == uuid.Nil {
			return nil
		}
		if _, exists := assigned[userID]; exists {
			return nil
		}
		req := &domain.ApprovalRequest{
			InstanceStepID: instanceStepID,
			AssignedUserID: userID,
			Status:         "PENDING",
		}
		if _, err := s.repo.CreateApprovalRequest(ctx, req); err != nil {
			return err
		}
		assigned[userID] = struct{}{}
		return nil
	}

	for _, app := range step.Approvers {
		if app.ApproverType == "USER" && app.UserID != nil {
			if err := assign(*app.UserID); err != nil {
				return err
			}
		} else if app.ApproverType == "ROLE" && app.RoleID != nil {
			userIDs, err := s.repo.ListUserIDsByRoleInScope(ctx, *app.RoleID, legalEntityID)
			if err != nil {
				return err
			}
			for _, userID := range userIDs {
				if err := assign(userID); err != nil {
					return err
				}
			}
		}
	}

	if len(assigned) == 0 {
		fallbackAdminID, err := s.repo.GetFallbackAdminUserID(ctx)
		if err != nil {
			return fmt.Errorf("no configured approver and fallback administrator lookup failed: %w", err)
		}
		if err := assign(fallbackAdminID); err != nil {
			return err
		}
	}
	if len(assigned) == 0 {
		return domain.ErrNoApproversAvailable
	}

	return nil
}

func (s *workflowService) SubmitDecision(ctx context.Context, cmd ports.SubmitDecisionCommand) (*domain.ApprovalDecision, error) {
	cmd.Decision = strings.ToUpper(strings.TrimSpace(cmd.Decision))
	cmd.Notes = strings.TrimSpace(cmd.Notes)
	if cmd.RequestID == uuid.Nil || cmd.ActorID == uuid.Nil {
		return nil, domain.ErrInvalidWorkflowCommand
	}
	if cmd.Decision != "APPROVE" && cmd.Decision != "REJECT" {
		return nil, domain.ErrInvalidDecision
	}

	var persistedDecision *domain.ApprovalDecision
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		// Discover the step first, then serialize all decisions for that step. The
		// lock order (step before request) also avoids a deadlock when one decision
		// cancels the other pending requests.
		requestSnapshot, err := s.repo.GetApprovalRequestByID(txCtx, cmd.RequestID)
		if err != nil {
			return err
		}
		instanceStep, err := s.repo.LockInstanceStepByID(txCtx, requestSnapshot.InstanceStepID)
		if err != nil {
			return err
		}
		req, err := s.repo.LockApprovalRequestByID(txCtx, cmd.RequestID)
		if err != nil {
			return err
		}
		if req.Status != "PENDING" || instanceStep.Status != "IN_PROGRESS" {
			return domain.ErrRequestAlreadyDecided
		}

		inst, err := s.repo.GetInstanceByID(txCtx, instanceStep.InstanceID)
		if err != nil {
			return err
		}
		if inst.Status != "PENDING" {
			return domain.ErrRequestAlreadyDecided
		}
		if err := domain.ValidateMakerChecker(inst.InitiatorID, cmd.ActorID); err != nil {
			return err
		}

		if req.AssignedUserID != cmd.ActorID {
			delegation, err := s.repo.GetActiveDelegation(txCtx, req.AssignedUserID, cmd.ActorID, &inst.LegalEntityID)
			if err != nil {
				return err
			}
			if delegation == nil {
				return domain.ErrUnauthorizedApprover
			}
		}

		steps, err := s.repo.ListSteps(txCtx, inst.VersionID)
		if err != nil {
			return fmt.Errorf("failed to load workflow steps: %w", err)
		}
		var currentStep *domain.WorkflowStep
		for i := range steps {
			if steps[i].ID == instanceStep.StepID {
				currentStep = &steps[i]
				break
			}
		}
		if currentStep == nil {
			return domain.ErrWorkflowStepNotFound
		}

		persistedDecision, err = s.repo.CreateApprovalDecision(txCtx, &domain.ApprovalDecision{
			RequestID:  req.ID,
			InstanceID: inst.ID,
			ActorID:    cmd.ActorID,
			Decision:   cmd.Decision,
			Notes:      cmd.Notes,
			DecidedAt:  time.Now().UTC(),
		})
		if err != nil {
			return fmt.Errorf("failed to record approval decision: %w", err)
		}
		if err := s.repo.UpdateApprovalRequestStatus(txCtx, req.ID, "DECIDED"); err != nil {
			return fmt.Errorf("failed to close approval request: %w", err)
		}

		now := time.Now().UTC()
		if cmd.Decision == "REJECT" {
			if err := s.repo.CancelPendingApprovalRequestsForStep(txCtx, instanceStep.ID); err != nil {
				return fmt.Errorf("failed to cancel sibling approval requests: %w", err)
			}
			if err := s.repo.UpdateInstanceStepStatus(txCtx, instanceStep.ID, "REJECTED", &now); err != nil {
				return fmt.Errorf("failed to reject workflow step: %w", err)
			}
			if err := s.repo.UpdateInstanceStatus(txCtx, inst.ID, "REJECTED", &now); err != nil {
				return fmt.Errorf("failed to reject workflow instance: %w", err)
			}
			return nil
		}

		if currentStep.ApprovalType == "ALL" {
			requests, err := s.repo.ListApprovalRequestsByInstanceStep(txCtx, instanceStep.ID)
			if err != nil {
				return fmt.Errorf("failed to inspect step approvals: %w", err)
			}
			for _, request := range requests {
				if request.Status == "PENDING" {
					return nil
				}
			}
		} else if err := s.repo.CancelPendingApprovalRequestsForStep(txCtx, instanceStep.ID); err != nil {
			return fmt.Errorf("failed to cancel sibling approval requests: %w", err)
		}

		if err := s.repo.UpdateInstanceStepStatus(txCtx, instanceStep.ID, "APPROVED", &now); err != nil {
			return fmt.Errorf("failed to approve workflow step: %w", err)
		}

		contextData := make(map[string]any)
		decoder := json.NewDecoder(bytes.NewReader(inst.ContextData))
		decoder.UseNumber()
		if err := decoder.Decode(&contextData); err != nil {
			return fmt.Errorf("failed to decode workflow context: %w", err)
		}
		var nextStep *domain.WorkflowStep
		for i := range steps {
			candidate := &steps[i]
			if candidate.StepOrder <= instanceStep.StepOrder {
				continue
			}
			eligible, err := domain.EvaluateConditions(candidate.Conditions, contextData)
			if err != nil {
				return err
			}
			if eligible {
				nextStep = candidate
				break
			}
		}

		if nextStep == nil {
			if err := s.repo.UpdateInstanceStatus(txCtx, inst.ID, "APPROVED", &now); err != nil {
				return fmt.Errorf("failed to approve workflow instance: %w", err)
			}
			return nil
		}

		createdNextStep, err := s.repo.CreateInstanceStep(txCtx, &domain.InstanceStep{
			InstanceID: inst.ID,
			StepID:     nextStep.ID,
			StepOrder:  nextStep.StepOrder,
			Status:     "IN_PROGRESS",
		})
		if err != nil {
			return fmt.Errorf("failed to activate next workflow step: %w", err)
		}
		if err := s.assignStepApprovers(txCtx, createdNextStep.ID, *nextStep, inst.LegalEntityID); err != nil {
			return fmt.Errorf("failed to assign next-step approvers: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return persistedDecision, nil
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
	if err != nil {
		return nil, err
	}
	if len(delegations) > 0 {
		seen := make(map[uuid.UUID]bool)
		for _, d := range direct {
			seen[d.RequestID] = true
		}

		for _, del := range delegations {
			delegatedItems, err := s.repo.ListPendingRequestsForUser(ctx, del.DelegatorID)
			if err != nil {
				return nil, err
			}
			for _, item := range delegatedItems {
				if del.LegalEntityID != nil && item.LegalEntityID != *del.LegalEntityID {
					continue
				}
				if !seen[item.RequestID] {
					seen[item.RequestID] = true
					direct = append(direct, item)
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
