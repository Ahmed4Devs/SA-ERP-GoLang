package services

import (
	"context"
	"erp/internal/sales/domain"
	workflowports "erp/internal/workflow/ports"
	"github.com/google/uuid"
)

func (s *salesService) SubmitInvoice(ctx context.Context, invoiceID, actorID uuid.UUID) (*domain.SalesInvoice, error) {
	invoice, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	_, err = s.workflow.StartWorkflow(ctx, workflowports.StartWorkflowCommand{DefinitionCode: "SALES_INVOICE_APPROVAL", ResourceType: "SALES_INVOICE", ResourceID: invoiceID.String(), LegalEntityID: invoice.LegalEntityID, InitiatorID: actorID})
	if err != nil {
		return nil, err
	}
	return s.repo.GetInvoice(ctx, invoiceID)
}
