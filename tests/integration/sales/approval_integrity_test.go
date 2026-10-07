package sales_test

import (
	"encoding/json"
	"erp/internal/platform/config"
	recdomain "erp/internal/receivables/domain"
	salesdomain "erp/internal/sales/domain"
	salesports "erp/internal/sales/ports"
	"erp/internal/workflow"
	workflowports "erp/internal/workflow/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestInvoiceApprovalBindsAmountMakerAndHistory(t *testing.T) {
	mod, rec, db, entity, maker, item, _, ctx := setupSalesTest(t)
	defer db.Close()
	customer := createTestCustomer(t, rec, entity, recdomain.CustomerTypeB2C, "عميل الاعتماد", "Approval customer", nil, decimal.NewFromInt(100000), ctx)
	invoice, err := mod.SalesService.CreateInvoice(ctx, salesports.CreateInvoiceCommand{LegalEntityID: entity, CustomerID: customer.ID, CustomerNameAr: customer.NameAr, CustomerNameEn: customer.NameEn, InvoiceType: salesdomain.InvoiceTypeSimplified, IssueDate: time.Now(), DueDate: time.Now(), CurrencyCode: "USD", ExchangeRate: decimal.RequireFromString("3.75"), CreatedByUserID: &maker, Lines: []salesports.CreateInvoiceLineCommand{{ItemID: item, Description: "Approval snapshot", UOM: "EA", Quantity: decimal.NewFromInt(1), UnitPrice: decimal.NewFromInt(16000), TaxCode: "VAT-EXEMPT-0"}}})
	if err != nil {
		t.Fatal(err)
	}
	checker, chief := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{checker, chief} {
		if _, err = db.Pool.Exec(ctx, `INSERT INTO iam.users(id,username,display_name,status) VALUES($1,$2,'Approver','ACTIVE')`, id, "checker-"+id.String()); err != nil {
			t.Fatal(err)
		}
	}
	for role, id := range map[string]uuid.UUID{"SALES_OFFICER": checker, "CHIEF_ACCOUNTANT": chief} {
		if _, err = db.Pool.Exec(ctx, `INSERT INTO authz.user_role_assignments(user_id,role_id,legal_entity_id) SELECT $1,id,$2 FROM authz.roles WHERE code=$3`, id, entity, role); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = mod.SalesService.PostInvoice(ctx, invoice.ID, checker); err == nil {
		t.Fatal("draft bypassed approval")
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	wf := workflow.NewModule(cfg, db)
	cmd := workflowports.StartWorkflowCommand{DefinitionCode: "SALES_INVOICE_APPROVAL", ResourceType: "SALES_INVOICE", ResourceID: invoice.ID.String(), LegalEntityID: entity, InitiatorID: checker, ContextData: map[string]any{"amount": "1"}}
	if _, err = wf.Service.StartWorkflow(ctx, cmd); err == nil {
		t.Fatal("non-creator submitted someone else's invoice")
	}
	cmd.InitiatorID = maker
	inst, err := wf.Service.StartWorkflow(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]string
	if err = json.Unmarshal(inst.ContextData, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot["amount"] != "60000" {
		t.Fatalf("client-controlled amount affected workflow: %s", inst.ContextData)
	}
	for _, q := range []string{`UPDATE sales.invoices SET total_amount=1 WHERE id=$1`, `UPDATE sales.invoice_lines SET quantity=2 WHERE invoice_id=$1`, `DELETE FROM sales.invoice_lines WHERE invoice_id=$1`} {
		if _, err = db.Pool.Exec(ctx, q, invoice.ID); err == nil {
			t.Fatal("submitted financial snapshot was mutable")
		}
	}
	inbox, err := wf.Service.ListPendingApprovals(ctx, checker)
	if err != nil {
		t.Fatal(err)
	}
	var request uuid.UUID
	for _, r := range inbox {
		if r.InstanceID == inst.ID {
			request = r.RequestID
		}
	}
	if request == uuid.Nil {
		t.Fatal("missing review request")
	}
	if _, err = wf.Service.SubmitDecision(ctx, workflowports.SubmitDecisionCommand{RequestID: request, ActorID: maker, Decision: "APPROVE"}); err == nil {
		t.Fatal("creator approved own invoice")
	}
	if _, err = db.Pool.Exec(ctx, `INSERT INTO workflow.approval_limits(legal_entity_id,user_id,resource_type,maximum_base_amount,updated_by) VALUES($1,$2,'SALES_INVOICE',59999.999999,$3)`, entity, checker, maker); err != nil {
		t.Fatal(err)
	}
	decision := workflowports.SubmitDecisionCommand{RequestID: request, ActorID: checker, Decision: "APPROVE"}
	if _, err = wf.Service.SubmitDecision(ctx, decision); err == nil {
		t.Fatal("approver exceeded configured base currency limit")
	}
	if _, err = db.Pool.Exec(ctx, `UPDATE workflow.approval_limits SET maximum_base_amount=60000 WHERE legal_entity_id=$1 AND user_id=$2`, entity, checker); err != nil {
		t.Fatal(err)
	}
	if _, err = wf.Service.SubmitDecision(ctx, decision); err != nil {
		t.Fatal(err)
	}
	if _, err = mod.SalesService.PostInvoice(ctx, invoice.ID, checker); err == nil {
		t.Fatal("first review bypassed required high-value approval")
	}
	inbox, err = wf.Service.ListPendingApprovals(ctx, chief)
	if err != nil {
		t.Fatal(err)
	}
	request = uuid.Nil
	for _, r := range inbox {
		if r.InstanceID == inst.ID {
			request = r.RequestID
		}
	}
	if request == uuid.Nil {
		t.Fatal("high-value invoice failed to route to finance")
	}
	if _, err = wf.Service.SubmitDecision(ctx, workflowports.SubmitDecisionCommand{RequestID: request, ActorID: chief, Decision: "APPROVE"}); err != nil {
		t.Fatal(err)
	}
	if _, err = mod.SalesService.PostInvoice(ctx, invoice.ID, maker); err == nil {
		t.Fatal("creator posted own invoice")
	}
	if _, err = mod.SalesService.PostInvoice(ctx, invoice.ID, checker); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Pool.Exec(ctx, `UPDATE sales.invoices SET status='DRAFT' WHERE id=$1`, invoice.ID); err == nil {
		t.Fatal("posted invoice reverted to mutable draft")
	}
}
