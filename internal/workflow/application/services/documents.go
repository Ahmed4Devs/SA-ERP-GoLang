package services

import (
	"context"
	"encoding/json"
	auditports "erp/internal/audit/ports"
	"erp/internal/workflow/domain"
	"erp/internal/workflow/ports"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Known financial documents never use client-supplied workflow amounts.
func (s *workflowService) snapshotDocument(ctx context.Context, cmd *ports.StartWorkflowCommand) error {
	if cmd.ResourceType != "SALES_INVOICE" {
		return nil
	}
	if cmd.DefinitionCode != "SALES_INVOICE_APPROVAL" {
		return fmt.Errorf("sales invoices require the sales invoice approval definition")
	}
	id, err := uuid.Parse(cmd.ResourceID)
	if err != nil {
		return fmt.Errorf("invoice UUID is required")
	}
	var amount, rate decimal.Decimal
	var currency, status string
	var maker *uuid.UUID
	err = s.db.DBTX(ctx).QueryRow(ctx, `SELECT total_amount,exchange_rate,currency_code,status,created_by_user_id FROM sales.invoices WHERE id=$1 AND legal_entity_id=$2 FOR UPDATE`, id, cmd.LegalEntityID).Scan(&amount, &rate, &currency, &status, &maker)
	if err != nil {
		return err
	}
	if status != "DRAFT" || maker == nil || *maker != cmd.InitiatorID || !rate.IsPositive() {
		return fmt.Errorf("only the recorded creator can submit a draft invoice with a valid exchange rate")
	}
	cmd.ContextData = map[string]any{"amount": amount.Mul(rate).Round(6).String(), "transaction_amount": amount.String(), "exchange_rate": rate.String(), "currency_code": currency}
	return nil
}

func (s *workflowService) linkDocument(ctx context.Context, inst *domain.WorkflowInstance) error {
	if inst.ResourceType != "SALES_INVOICE" {
		return nil
	}
	_, err := s.db.DBTX(ctx).Exec(ctx, `UPDATE sales.invoices SET status='SUBMITTED',approval_instance_id=$2,approved_by=NULL,approved_at=NULL,updated_at=now() WHERE id=$1 AND legal_entity_id=$3 AND status='DRAFT'`, inst.ResourceID, inst.ID, inst.LegalEntityID)
	return err
}

func (s *workflowService) validateDocumentAuthority(ctx context.Context, inst *domain.WorkflowInstance, actor uuid.UUID) error {
	if inst.ResourceType != "SALES_INVOICE" {
		return nil
	}
	var amount, rate decimal.Decimal
	var currency string
	var maker *uuid.UUID
	err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT total_amount,exchange_rate,currency_code,created_by_user_id FROM sales.invoices WHERE id=$1 AND legal_entity_id=$2 AND approval_instance_id=$3 AND status='SUBMITTED' FOR UPDATE`, inst.ResourceID, inst.LegalEntityID, inst.ID).Scan(&amount, &rate, &currency, &maker)
	if err != nil {
		return fmt.Errorf("invoice approval snapshot is unavailable: %w", err)
	}
	if maker == nil || *maker == actor || *maker != inst.InitiatorID {
		return domain.ErrMakerCannotApprove
	}
	var snapshot map[string]string
	if err = json.Unmarshal(inst.ContextData, &snapshot); err != nil {
		return err
	}
	if snapshot["transaction_amount"] != amount.String() || snapshot["exchange_rate"] != rate.String() || snapshot["currency_code"] != currency {
		return fmt.Errorf("invoice changed since submission")
	}
	return CheckApprovalLimit(ctx, s.db, inst.LegalEntityID, actor, "SALES_INVOICE", amount.Mul(rate).Round(6))
}

func (s *workflowService) completeDocument(ctx context.Context, inst *domain.WorkflowInstance, actor uuid.UUID, approved bool) error {
	if inst.ResourceType != "SALES_INVOICE" {
		return nil
	}
	query := `UPDATE sales.invoices SET status='DRAFT',approval_instance_id=NULL,updated_at=now() WHERE id=$1 AND legal_entity_id=$2 AND approval_instance_id=$3 AND status='SUBMITTED'`
	args := []any{inst.ResourceID, inst.LegalEntityID, inst.ID}
	if approved {
		query = `UPDATE sales.invoices SET status='APPROVED',approved_by=$4,approved_at=now(),updated_at=now() WHERE id=$1 AND legal_entity_id=$2 AND approval_instance_id=$3 AND status='SUBMITTED'`
		args = append(args, actor)
	}
	tag, err := s.db.DBTX(ctx).Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("invoice approval state changed")
	}
	return nil
}

func (s *workflowService) recordWorkflow(ctx context.Context, inst *domain.WorkflowInstance, actor uuid.UUID, action string) error {
	payload, err := json.Marshal(inst)
	if err != nil {
		return err
	}
	_, err = s.audit.RecordEvent(ctx, auditports.RecordEventCommand{UserID: &actor, LegalEntityID: &inst.LegalEntityID, Action: action, ResourceType: "workflow.instance", ResourceID: inst.ID.String(), Payload: payload})
	return err
}
