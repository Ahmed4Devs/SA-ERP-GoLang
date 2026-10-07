package sales_test

import (
	"context"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	salesports "erp/internal/sales/ports"
	"erp/internal/workflow"
	workflowports "erp/internal/workflow/ports"
	"github.com/google/uuid"
	"testing"
)

func approveSalesInvoice(t *testing.T, ctx context.Context, db *database.DB, svc salesports.SalesService, id uuid.UUID) uuid.UUID {
	t.Helper()
	inv, err := svc.GetInvoice(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	checker := uuid.New()
	if _, err = db.Pool.Exec(ctx, `INSERT INTO iam.users(id,username,display_name,status) VALUES($1,$2,'Sales approver','ACTIVE')`, checker, "sales-checker-"+checker.String()); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Pool.Exec(ctx, `INSERT INTO authz.user_role_assignments(user_id,role_id,legal_entity_id) SELECT $1,id,$2 FROM authz.roles WHERE code IN ('SALES_OFFICER','CHIEF_ACCOUNTANT')`, checker, inv.LegalEntityID); err != nil {
		t.Fatal(err)
	}
	if inv.CreatedByUserID == nil {
		t.Fatal("approval fixture requires creator")
	}
	if _, err = svc.SubmitInvoice(ctx, id, *inv.CreatedByUserID); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	wf := workflow.NewModule(cfg, db)
	for i := 0; i < 5; i++ {
		inst, err := wf.Service.GetInstanceByResource(ctx, "SALES_INVOICE", id.String())
		if err != nil {
			t.Fatal(err)
		}
		if inst.Status == "APPROVED" {
			return checker
		}
		inbox, err := wf.Service.ListPendingApprovals(ctx, checker)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, req := range inbox {
			if req.InstanceID != inst.ID {
				continue
			}
			found = true
			if _, err = wf.Service.SubmitDecision(ctx, workflowports.SubmitDecisionCommand{RequestID: req.RequestID, ActorID: checker, Decision: "APPROVE"}); err != nil {
				t.Fatal(err)
			}
		}
		if !found {
			t.Fatal("invoice workflow has no checker request")
		}
	}
	t.Fatal("invoice approval did not complete")
	return uuid.Nil
}
