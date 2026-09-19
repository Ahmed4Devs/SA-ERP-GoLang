package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/workflow/domain"
	"erp/internal/workflow/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type workflowRepository struct {
	db *database.DB
}

func NewWorkflowRepository(db *database.DB) ports.WorkflowRepository {
	return &workflowRepository{db: db}
}

func (r *workflowRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

func (r *workflowRepository) GetDefinitionByCode(ctx context.Context, code string) (*domain.WorkflowDefinition, error) {
	row, err := r.queries(ctx).GetWorkflowDefinitionByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowDefinitionNotFound
		}
		return nil, err
	}
	return toDomainDefinition(&row), nil
}

func (r *workflowRepository) GetDefinitionByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowDefinition, error) {
	row, err := r.queries(ctx).GetWorkflowDefinitionByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowDefinitionNotFound
		}
		return nil, err
	}
	return toDomainDefinition(&row), nil
}

func (r *workflowRepository) ListDefinitions(ctx context.Context) ([]domain.WorkflowDefinition, error) {
	rows, err := r.queries(ctx).ListWorkflowDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.WorkflowDefinition, 0, len(rows))
	for _, row := range rows {
		def := toDomainDefinition(&row)
		verRow, err := r.queries(ctx).GetPublishedVersionByDefinitionID(ctx, row.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result = append(result, *def)
				continue
			}
			return nil, err
		}
		ver := toDomainVersion(&verRow)
		steps, err := r.ListSteps(ctx, ver.ID)
		if err != nil {
			return nil, err
		}
		ver.Steps = steps
		def.Versions = []domain.DefinitionVersion{*ver}
		result = append(result, *def)
	}
	return result, nil
}

func (r *workflowRepository) GetPublishedVersion(ctx context.Context, defID uuid.UUID) (*domain.DefinitionVersion, error) {
	row, err := r.queries(ctx).GetPublishedVersionByDefinitionID(ctx, toPgUUID(defID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPublishedVersionNotFound
		}
		return nil, err
	}
	return toDomainVersion(&row), nil
}

func (r *workflowRepository) ListSteps(ctx context.Context, versionID uuid.UUID) ([]domain.WorkflowStep, error) {
	rows, err := r.queries(ctx).ListStepsByVersionID(ctx, toPgUUID(versionID))
	if err != nil {
		return nil, err
	}

	steps := make([]domain.WorkflowStep, 0, len(rows))
	for _, row := range rows {
		s := toDomainStep(&row)

		// Load conditions
		condRows, err := r.queries(ctx).ListStepConditionsByStepID(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		for _, c := range condRows {
			s.Conditions = append(s.Conditions, *toDomainCondition(&c))
		}

		// Load approvers
		appRows, err := r.queries(ctx).ListStepApproversByStepID(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		for _, a := range appRows {
			s.Approvers = append(s.Approvers, *toDomainApprover(&a))
		}

		steps = append(steps, *s)
	}
	return steps, nil
}

func (r *workflowRepository) CreateInstance(ctx context.Context, inst *domain.WorkflowInstance) (*domain.WorkflowInstance, error) {
	contextData := inst.ContextData
	if len(contextData) == 0 {
		contextData = []byte("{}")
	}

	row, err := r.queries(ctx).CreateWorkflowInstance(ctx, sqlc.CreateWorkflowInstanceParams{
		DefinitionID:  toPgUUID(inst.DefinitionID),
		VersionID:     toPgUUID(inst.VersionID),
		LegalEntityID: toPgUUID(inst.LegalEntityID),
		ResourceType:  inst.ResourceType,
		ResourceID:    inst.ResourceID,
		InitiatorID:   toPgUUID(inst.InitiatorID),
		Status:        inst.Status,
		ContextData:   contextData,
	})
	if err != nil {
		return nil, err
	}
	return toDomainInstance(&row), nil
}

func (r *workflowRepository) GetInstanceByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowInstance, error) {
	row, err := r.queries(ctx).GetWorkflowInstanceByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowInstanceNotFound
		}
		return nil, err
	}
	return toDomainInstance(&row), nil
}

func (r *workflowRepository) GetInstanceByResource(ctx context.Context, resourceType, resourceID string) (*domain.WorkflowInstance, error) {
	row, err := r.queries(ctx).GetWorkflowInstanceByResource(ctx, sqlc.GetWorkflowInstanceByResourceParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowInstanceNotFound
		}
		return nil, err
	}
	return toDomainInstance(&row), nil
}

func (r *workflowRepository) UpdateInstanceStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) error {
	_, err := r.queries(ctx).UpdateWorkflowInstanceStatus(ctx, sqlc.UpdateWorkflowInstanceStatusParams{
		ID:          toPgUUID(id),
		Status:      status,
		CompletedAt: toPgTimestamptz(completedAt),
	})
	return err
}

func (r *workflowRepository) CreateInstanceStep(ctx context.Context, step *domain.InstanceStep) (*domain.InstanceStep, error) {
	row, err := r.queries(ctx).CreateInstanceStep(ctx, sqlc.CreateInstanceStepParams{
		InstanceID: toPgUUID(step.InstanceID),
		StepID:     toPgUUID(step.StepID),
		StepOrder:  int32(step.StepOrder),
		Status:     step.Status,
	})
	if err != nil {
		return nil, err
	}
	return toDomainInstanceStep(&row), nil
}

func (r *workflowRepository) GetInstanceStepByID(ctx context.Context, id uuid.UUID) (*domain.InstanceStep, error) {
	row, err := r.queries(ctx).GetInstanceStepByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowStepNotFound
		}
		return nil, err
	}
	return toDomainInstanceStep(&row), nil
}

func (r *workflowRepository) LockInstanceStepByID(ctx context.Context, id uuid.UUID) (*domain.InstanceStep, error) {
	row, err := r.queries(ctx).LockInstanceStepByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWorkflowStepNotFound
		}
		return nil, err
	}
	return toDomainInstanceStep(&row), nil
}

func (r *workflowRepository) ListInstanceSteps(ctx context.Context, instanceID uuid.UUID) ([]domain.InstanceStep, error) {
	rows, err := r.queries(ctx).ListInstanceStepsByInstanceID(ctx, toPgUUID(instanceID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.InstanceStep, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainInstanceStep(&row))
	}
	return result, nil
}

func (r *workflowRepository) UpdateInstanceStepStatus(ctx context.Context, id uuid.UUID, status string, completedAt *time.Time) error {
	_, err := r.queries(ctx).UpdateInstanceStepStatus(ctx, sqlc.UpdateInstanceStepStatusParams{
		ID:          toPgUUID(id),
		Status:      status,
		CompletedAt: toPgTimestamptz(completedAt),
	})
	return err
}

func (r *workflowRepository) CreateApprovalRequest(ctx context.Context, req *domain.ApprovalRequest) (*domain.ApprovalRequest, error) {
	row, err := r.queries(ctx).CreateApprovalRequest(ctx, sqlc.CreateApprovalRequestParams{
		InstanceStepID: toPgUUID(req.InstanceStepID),
		AssignedUserID: toPgUUID(req.AssignedUserID),
		Status:         req.Status,
	})
	if err != nil {
		return nil, err
	}
	return toDomainApprovalRequest(&row), nil
}

func (r *workflowRepository) GetApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error) {
	row, err := r.queries(ctx).GetApprovalRequestByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrApprovalRequestNotFound
		}
		return nil, err
	}
	return toDomainApprovalRequest(&row), nil
}

func (r *workflowRepository) LockApprovalRequestByID(ctx context.Context, id uuid.UUID) (*domain.ApprovalRequest, error) {
	row, err := r.queries(ctx).LockApprovalRequestByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrApprovalRequestNotFound
		}
		return nil, err
	}
	return toDomainApprovalRequest(&row), nil
}

func (r *workflowRepository) ListApprovalRequestsByInstanceStep(ctx context.Context, instanceStepID uuid.UUID) ([]domain.ApprovalRequest, error) {
	rows, err := r.queries(ctx).ListApprovalRequestsByInstanceStep(ctx, toPgUUID(instanceStepID))
	if err != nil {
		return nil, err
	}
	result := make([]domain.ApprovalRequest, 0, len(rows))
	for i := range rows {
		result = append(result, *toDomainApprovalRequest(&rows[i]))
	}
	return result, nil
}

func (r *workflowRepository) CancelPendingApprovalRequestsForStep(ctx context.Context, instanceStepID uuid.UUID) error {
	return r.queries(ctx).CancelPendingApprovalRequestsForStep(ctx, toPgUUID(instanceStepID))
}

func (r *workflowRepository) ListPendingRequestsForUser(ctx context.Context, userID uuid.UUID) ([]domain.PendingApprovalItem, error) {
	rows, err := r.queries(ctx).ListPendingApprovalRequestsForUser(ctx, toPgUUID(userID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.PendingApprovalItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.PendingApprovalItem{
			RequestID:      toUUID(row.ID),
			InstanceStepID: toUUID(row.InstanceStepID),
			InstanceID:     toUUID(row.InstanceID),
			ResourceType:   row.ResourceType,
			ResourceID:     row.ResourceID,
			LegalEntityID:  toUUID(row.LegalEntityID),
			InitiatorID:    toUUID(row.InitiatorID),
			InitiatorName:  row.InitiatorName,
			StepNameAr:     row.StepNameAr,
			StepNameEn:     row.StepNameEn,
			ContextData:    row.ContextData,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *workflowRepository) UpdateApprovalRequestStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.queries(ctx).UpdateApprovalRequestStatus(ctx, sqlc.UpdateApprovalRequestStatusParams{
		ID:     toPgUUID(id),
		Status: status,
	})
	return err
}

func (r *workflowRepository) CreateApprovalDecision(ctx context.Context, dec *domain.ApprovalDecision) (*domain.ApprovalDecision, error) {
	row, err := r.queries(ctx).CreateApprovalDecision(ctx, sqlc.CreateApprovalDecisionParams{
		RequestID:  toPgUUID(dec.RequestID),
		InstanceID: toPgUUID(dec.InstanceID),
		ActorID:    toPgUUID(dec.ActorID),
		Decision:   dec.Decision,
		Notes:      toPgText(dec.Notes),
	})
	if err != nil {
		return nil, err
	}
	return toDomainApprovalDecision(&row), nil
}

func (r *workflowRepository) ListApprovalDecisions(ctx context.Context, instanceID uuid.UUID) ([]domain.ApprovalDecision, error) {
	rows, err := r.queries(ctx).ListApprovalDecisionsByInstanceID(ctx, toPgUUID(instanceID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.ApprovalDecision, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ApprovalDecision{
			ID:         toUUID(row.ID),
			RequestID:  toUUID(row.RequestID),
			InstanceID: toUUID(row.InstanceID),
			ActorID:    toUUID(row.ActorID),
			Decision:   row.Decision,
			Notes:      fromPgText(row.Notes),
			DecidedAt:  row.DecidedAt.Time,
		})
	}
	return result, nil
}

func (r *workflowRepository) CreateDelegation(ctx context.Context, del *domain.Delegation) (*domain.Delegation, error) {
	row, err := r.queries(ctx).CreateDelegation(ctx, sqlc.CreateDelegationParams{
		DelegatorID:   toPgUUID(del.DelegatorID),
		DelegateID:    toPgUUID(del.DelegateID),
		LegalEntityID: toPgUUIDPtr(del.LegalEntityID),
		StartDate:     pgtype.Timestamptz{Time: del.StartDate, Valid: true},
		EndDate:       pgtype.Timestamptz{Time: del.EndDate, Valid: true},
		IsActive:      del.IsActive,
		Reason:        toPgText(del.Reason),
	})
	if err != nil {
		return nil, err
	}
	return toDomainDelegation(&row), nil
}

func (r *workflowRepository) GetActiveDelegation(ctx context.Context, delegatorID, delegateID uuid.UUID, leID *uuid.UUID) (*domain.Delegation, error) {
	row, err := r.queries(ctx).GetActiveDelegation(ctx, sqlc.GetActiveDelegationParams{
		DelegatorID: toPgUUID(delegatorID),
		DelegateID:  toPgUUID(delegateID),
		Column3:     toPgUUIDPtr(leID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainDelegation(&row), nil
}

func (r *workflowRepository) ListActiveDelegationsForDelegate(ctx context.Context, delegateID uuid.UUID) ([]domain.Delegation, error) {
	rows, err := r.queries(ctx).ListActiveDelegationsForDelegate(ctx, toPgUUID(delegateID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.Delegation, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Delegation{
			ID:            toUUID(row.ID),
			DelegatorID:   toUUID(row.DelegatorID),
			DelegateID:    toUUID(row.DelegateID),
			LegalEntityID: toUUIDPtr(row.LegalEntityID),
			StartDate:     row.StartDate.Time,
			EndDate:       row.EndDate.Time,
			IsActive:      row.IsActive,
			Reason:        fromPgText(row.Reason),
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *workflowRepository) ListDelegationsByDelegator(ctx context.Context, delegatorID uuid.UUID) ([]domain.Delegation, error) {
	rows, err := r.queries(ctx).ListDelegationsByDelegator(ctx, toPgUUID(delegatorID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.Delegation, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Delegation{
			ID:            toUUID(row.ID),
			DelegatorID:   toUUID(row.DelegatorID),
			DelegateID:    toUUID(row.DelegateID),
			LegalEntityID: toUUIDPtr(row.LegalEntityID),
			StartDate:     row.StartDate.Time,
			EndDate:       row.EndDate.Time,
			IsActive:      row.IsActive,
			Reason:        fromPgText(row.Reason),
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *workflowRepository) RevokeDelegation(ctx context.Context, id, delegatorID uuid.UUID) error {
	return r.queries(ctx).RevokeDelegation(ctx, sqlc.RevokeDelegationParams{
		ID:          toPgUUID(id),
		DelegatorID: toPgUUID(delegatorID),
	})
}

func (r *workflowRepository) ListUserIDsByRoleInScope(ctx context.Context, roleID, leID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.queries(ctx).ListUserIDsByRoleInScope(ctx, sqlc.ListUserIDsByRoleInScopeParams{
		RoleID:        toPgUUID(roleID),
		LegalEntityID: toPgUUID(leID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		result = append(result, toUUID(row))
	}
	return result, nil
}

func (r *workflowRepository) GetFallbackAdminUserID(ctx context.Context) (uuid.UUID, error) {
	row, err := r.queries(ctx).GetFallbackAdminUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return toUUID(row), nil
}

// Helpers
func toDomainDefinition(r *sqlc.WorkflowDefinitions) *domain.WorkflowDefinition {
	return &domain.WorkflowDefinition{
		ID:        toUUID(r.ID),
		Code:      r.Code,
		NameAr:    r.NameAr,
		NameEn:    r.NameEn,
		Module:    r.Module,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

func toDomainVersion(r *sqlc.WorkflowDefinitionVersions) *domain.DefinitionVersion {
	var pub *time.Time
	if r.PublishedAt.Valid {
		pub = &r.PublishedAt.Time
	}
	return &domain.DefinitionVersion{
		ID:           toUUID(r.ID),
		DefinitionID: toUUID(r.DefinitionID),
		Version:      int(r.Version),
		Status:       r.Status,
		CreatedAt:    r.CreatedAt.Time,
		PublishedAt:  pub,
	}
}

func toDomainStep(r *sqlc.WorkflowSteps) *domain.WorkflowStep {
	timeout := 24
	if r.TimeoutHours.Valid {
		timeout = int(r.TimeoutHours.Int32)
	}
	return &domain.WorkflowStep{
		ID:           toUUID(r.ID),
		VersionID:    toUUID(r.VersionID),
		StepOrder:    int(r.StepOrder),
		Code:         r.Code,
		NameAr:       r.NameAr,
		NameEn:       r.NameEn,
		ApprovalType: r.ApprovalType,
		TimeoutHours: timeout,
	}
}

func toDomainCondition(r *sqlc.WorkflowStepConditions) *domain.StepCondition {
	return &domain.StepCondition{
		ID:         toUUID(r.ID),
		StepID:     toUUID(r.StepID),
		FieldName:  r.FieldName,
		Operator:   r.Operator,
		FieldValue: r.FieldValue,
	}
}

func toDomainApprover(r *sqlc.WorkflowStepApprovers) *domain.StepApprover {
	return &domain.StepApprover{
		ID:           toUUID(r.ID),
		StepID:       toUUID(r.StepID),
		ApproverType: r.ApproverType,
		RoleID:       toUUIDPtr(r.RoleID),
		UserID:       toUUIDPtr(r.UserID),
	}
}

func toDomainInstance(r *sqlc.WorkflowInstances) *domain.WorkflowInstance {
	var comp *time.Time
	if r.CompletedAt.Valid {
		comp = &r.CompletedAt.Time
	}
	return &domain.WorkflowInstance{
		ID:            toUUID(r.ID),
		DefinitionID:  toUUID(r.DefinitionID),
		VersionID:     toUUID(r.VersionID),
		LegalEntityID: toUUID(r.LegalEntityID),
		ResourceType:  r.ResourceType,
		ResourceID:    r.ResourceID,
		InitiatorID:   toUUID(r.InitiatorID),
		Status:        r.Status,
		ContextData:   r.ContextData,
		CreatedAt:     r.CreatedAt.Time,
		CompletedAt:   comp,
	}
}

func toDomainInstanceStep(r *sqlc.WorkflowInstanceSteps) *domain.InstanceStep {
	var comp *time.Time
	if r.CompletedAt.Valid {
		comp = &r.CompletedAt.Time
	}
	return &domain.InstanceStep{
		ID:          toUUID(r.ID),
		InstanceID:  toUUID(r.InstanceID),
		StepID:      toUUID(r.StepID),
		StepOrder:   int(r.StepOrder),
		Status:      r.Status,
		StartedAt:   r.StartedAt.Time,
		CompletedAt: comp,
	}
}

func toDomainApprovalRequest(r *sqlc.WorkflowApprovalRequests) *domain.ApprovalRequest {
	return &domain.ApprovalRequest{
		ID:             toUUID(r.ID),
		InstanceStepID: toUUID(r.InstanceStepID),
		AssignedUserID: toUUID(r.AssignedUserID),
		Status:         r.Status,
		CreatedAt:      r.CreatedAt.Time,
	}
}

func toDomainApprovalDecision(r *sqlc.WorkflowApprovalDecisions) *domain.ApprovalDecision {
	return &domain.ApprovalDecision{
		ID:         toUUID(r.ID),
		RequestID:  toUUID(r.RequestID),
		InstanceID: toUUID(r.InstanceID),
		ActorID:    toUUID(r.ActorID),
		Decision:   r.Decision,
		Notes:      fromPgText(r.Notes),
		DecidedAt:  r.DecidedAt.Time,
	}
}

func toDomainDelegation(r *sqlc.WorkflowDelegations) *domain.Delegation {
	return &domain.Delegation{
		ID:            toUUID(r.ID),
		DelegatorID:   toUUID(r.DelegatorID),
		DelegateID:    toUUID(r.DelegateID),
		LegalEntityID: toUUIDPtr(r.LegalEntityID),
		StartDate:     r.StartDate.Time,
		EndDate:       r.EndDate.Time,
		IsActive:      r.IsActive,
		Reason:        fromPgText(r.Reason),
		CreatedAt:     r.CreatedAt.Time,
	}
}
