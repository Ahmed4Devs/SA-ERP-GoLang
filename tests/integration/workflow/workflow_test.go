package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"erp/internal/authorization"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/workflow"
	"erp/internal/workflow/domain"
	"erp/internal/workflow/ports"

	"github.com/google/uuid"
)

func setupWorkflowTest(t *testing.T) (*workflow.Module, *authorization.Module, *organization.Module, *iam.Module, *database.DB, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("integration database setup failed: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	iamMod := iam.NewModule(cfg, db, txManager)
	authzMod := authorization.NewModule(cfg, db)
	orgMod := organization.NewModule(cfg, db, txManager)
	wfMod := workflow.NewModule(cfg, db)
	t.Cleanup(db.Close)

	return wfMod, authzMod, orgMod, iamMod, db, ctx
}

func assignWorkflowRole(t *testing.T, ctx context.Context, authzMod *authorization.Module, userID, legalEntityID uuid.UUID, roleCode string) {
	t.Helper()
	roles, err := authzMod.Service.ListRoles(ctx)
	if err != nil {
		t.Fatalf("failed to list roles: %v", err)
	}
	for _, role := range roles {
		if role.Code == roleCode {
			if _, err := authzMod.Service.AssignRole(ctx, userID, role.ID, &legalEntityID, nil); err != nil {
				t.Fatalf("failed to assign %s role: %v", roleCode, err)
			}
			return
		}
	}
	t.Fatalf("role %s not found", roleCode)
}

func approvalRequestForInstance(t *testing.T, ctx context.Context, wfMod *workflow.Module, userID, instanceID uuid.UUID) uuid.UUID {
	t.Helper()
	items, err := wfMod.Service.ListPendingApprovals(ctx, userID)
	if err != nil {
		t.Fatalf("failed to list pending approvals: %v", err)
	}
	for _, item := range items {
		if item.InstanceID == instanceID {
			return item.RequestID
		}
	}
	t.Fatalf("pending approval for instance %s not found", instanceID)
	return uuid.Nil
}

func TestWorkflow_DefinitionsAndPublishedVersions(t *testing.T) {
	wfMod, _, _, _, _, ctx := setupWorkflowTest(t)

	defs, err := wfMod.Service.ListDefinitions(ctx)
	if err != nil {
		t.Fatalf("failed to list workflow definitions: %v", err)
	}

	if len(defs) == 0 {
		t.Fatal("expected seeded workflow definitions, got 0")
	}

	var foundSalesInvoice bool
	for _, def := range defs {
		if def.Code == "SALES_INVOICE_APPROVAL" {
			foundSalesInvoice = true
			if len(def.Versions) == 0 {
				t.Fatalf("expected definition %s to have at least 1 published version", def.Code)
			}
			steps := def.Versions[0].Steps
			if len(steps) < 2 {
				t.Fatalf("expected at least 2 steps for %s, got %d", def.Code, len(steps))
			}
		}
	}

	if !foundSalesInvoice {
		t.Fatal("expected SALES_INVOICE_APPROVAL definition to be seeded")
	}
}

func TestWorkflow_MakerCheckerEnforcement(t *testing.T) {
	wfMod, authzMod, orgMod, iamMod, _, ctx := setupWorkflowTest(t)

	// 1. Get or seed Legal Entity
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed legal entity: %v", err)
	}

	uid := uuid.NewString()[:8]
	// Create Maker (Initiator)
	makerUser, err := iamMod.AuthService.SeedAdmin(ctx, "maker_"+uid, "maker_"+uid+"@wf.test", "منشئ الوثيقة", "Password123!")
	if err != nil {
		t.Fatalf("failed to seed maker: %v", err)
	}

	// Create Checker (Approver)
	checkerUser, err := iamMod.AuthService.SeedAdmin(ctx, "checker_"+uid, "checker_"+uid+"@wf.test", "معتمد الوثيقة", "Password123!")
	if err != nil {
		t.Fatalf("failed to seed checker: %v", err)
	}
	assignWorkflowRole(t, ctx, authzMod, checkerUser.ID, le.ID, "SALES_OFFICER")

	// 2. Start Workflow for a document
	inst, err := wfMod.Service.StartWorkflow(ctx, ports.StartWorkflowCommand{
		DefinitionCode: "SALES_INVOICE_APPROVAL",
		LegalEntityID:  le.ID,
		ResourceType:   "TEST_SALES_DOCUMENT",
		ResourceID:     "INV-2026-TEST01",
		InitiatorID:    makerUser.ID,
		ContextData: map[string]any{
			"amount": "15000.00",
		},
	})
	if err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}

	if inst.Status != "PENDING" {
		t.Fatalf("expected instance status PENDING, got %s", inst.Status)
	}

	if len(inst.Steps) == 0 {
		t.Fatal("expected at least 1 instance step created")
	}

	requestID := approvalRequestForInstance(t, ctx, wfMod, checkerUser.ID, inst.ID)

	// Maker tries to approve the checker's real request.
	_, err = wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: requestID,
		ActorID:   makerUser.ID,
		Decision:  "APPROVE",
		Notes:     "Self-approval attempt",
	})
	if !errors.Is(err, domain.ErrMakerCannotApprove) {
		t.Fatalf("expected ErrMakerCannotApprove, got: %v", err)
	}

	decision, err := wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: requestID,
		ActorID:   checkerUser.ID,
		Decision:  "APPROVE",
		Notes:     "Approved by independent checker",
	})
	if err != nil {
		t.Fatalf("checker approval failed: %v", err)
	}
	if decision.InstanceID != inst.ID {
		t.Fatalf("decision references instance %s, want %s", decision.InstanceID, inst.ID)
	}

	completed, err := wfMod.Service.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow instance: %v", err)
	}
	if completed.Status != "APPROVED" {
		t.Fatalf("expected workflow APPROVED, got %s", completed.Status)
	}

	_, err = wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: requestID,
		ActorID:   checkerUser.ID,
		Decision:  "APPROVE",
	})
	if !errors.Is(err, domain.ErrRequestAlreadyDecided) {
		t.Fatalf("expected duplicate decision to fail with ErrRequestAlreadyDecided, got: %v", err)
	}
}

func TestWorkflow_AdvancesConditionalStepAtomically(t *testing.T) {
	wfMod, authzMod, orgMod, iamMod, _, ctx := setupWorkflowTest(t)
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed legal entity: %v", err)
	}

	uid := uuid.NewString()[:8]
	maker, err := iamMod.AuthService.SeedAdmin(ctx, "wf_maker_"+uid, "wf_maker_"+uid+"@wf.test", "Workflow Maker", "Password123!")
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}
	salesApprover, err := iamMod.AuthService.SeedAdmin(ctx, "wf_sales_"+uid, "wf_sales_"+uid+"@wf.test", "Sales Approver", "Password123!")
	if err != nil {
		t.Fatalf("failed to create sales approver: %v", err)
	}
	financeApprover, err := iamMod.AuthService.SeedAdmin(ctx, "wf_finance_"+uid, "wf_finance_"+uid+"@wf.test", "Finance Approver", "Password123!")
	if err != nil {
		t.Fatalf("failed to create finance approver: %v", err)
	}
	assignWorkflowRole(t, ctx, authzMod, salesApprover.ID, le.ID, "SALES_OFFICER")
	assignWorkflowRole(t, ctx, authzMod, financeApprover.ID, le.ID, "CHIEF_ACCOUNTANT")

	inst, err := wfMod.Service.StartWorkflow(ctx, ports.StartWorkflowCommand{
		DefinitionCode: "SALES_INVOICE_APPROVAL",
		LegalEntityID:  le.ID,
		ResourceType:   "TEST_SALES_DOCUMENT",
		ResourceID:     "INV-HIGH-" + uid,
		InitiatorID:    maker.ID,
		ContextData:    map[string]any{"amount": "75000.00"},
	})
	if err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}

	salesRequestID := approvalRequestForInstance(t, ctx, wfMod, salesApprover.ID, inst.ID)
	firstDecision, err := wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: salesRequestID,
		ActorID:   salesApprover.ID,
		Decision:  "APPROVE",
	})
	if err != nil {
		t.Fatalf("failed to approve first step: %v", err)
	}
	if firstDecision.InstanceID != inst.ID {
		t.Fatalf("first decision references instance %s, want %s", firstDecision.InstanceID, inst.ID)
	}
	pending, err := wfMod.Service.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow: %v", err)
	}
	if pending.Status != "PENDING" {
		t.Fatalf("expected workflow to remain PENDING for finance approval, got %s", pending.Status)
	}

	financeRequestID := approvalRequestForInstance(t, ctx, wfMod, financeApprover.ID, inst.ID)
	if _, err := wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: financeRequestID,
		ActorID:   financeApprover.ID,
		Decision:  "APPROVE",
	}); err != nil {
		t.Fatalf("failed to approve finance step: %v", err)
	}
	approved, err := wfMod.Service.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("failed to reload completed workflow: %v", err)
	}
	if approved.Status != "APPROVED" {
		t.Fatalf("expected workflow APPROVED, got %s", approved.Status)
	}
}

func TestWorkflow_ConcurrentAnyApprovalsCreateOneDecision(t *testing.T) {
	wfMod, authzMod, orgMod, iamMod, _, ctx := setupWorkflowTest(t)
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed legal entity: %v", err)
	}

	uid := uuid.NewString()[:8]
	maker, err := iamMod.AuthService.SeedAdmin(ctx, "wf_concurrent_maker_"+uid, "wf_concurrent_maker_"+uid+"@wf.test", "Concurrent Maker", "Password123!")
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}
	approvers := make([]uuid.UUID, 2)
	for i := range approvers {
		user, err := iamMod.AuthService.SeedAdmin(ctx, "wf_concurrent_"+uid+string(rune('a'+i)), "wf_concurrent_"+uid+string(rune('a'+i))+"@wf.test", "Concurrent Approver", "Password123!")
		if err != nil {
			t.Fatalf("failed to create approver: %v", err)
		}
		approvers[i] = user.ID
		assignWorkflowRole(t, ctx, authzMod, user.ID, le.ID, "SALES_OFFICER")
	}

	inst, err := wfMod.Service.StartWorkflow(ctx, ports.StartWorkflowCommand{
		DefinitionCode: "SALES_INVOICE_APPROVAL",
		LegalEntityID:  le.ID,
		ResourceType:   "TEST_SALES_DOCUMENT",
		ResourceID:     "INV-CONCURRENT-" + uid,
		InitiatorID:    maker.ID,
		ContextData:    map[string]any{"amount": "100.00"},
	})
	if err != nil {
		t.Fatalf("failed to start workflow: %v", err)
	}
	requestIDs := []uuid.UUID{
		approvalRequestForInstance(t, ctx, wfMod, approvers[0], inst.ID),
		approvalRequestForInstance(t, ctx, wfMod, approvers[1], inst.ID),
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	for i := range approvers {
		go func(i int) {
			<-start
			_, err := wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
				RequestID: requestIDs[i],
				ActorID:   approvers[i],
				Decision:  "APPROVE",
			})
			results <- err
		}(i)
	}
	close(start)

	successes := 0
	alreadyDecided := 0
	for range approvers {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrRequestAlreadyDecided):
			alreadyDecided++
		default:
			t.Fatalf("unexpected concurrent approval result: %v", err)
		}
	}
	if successes != 1 || alreadyDecided != 1 {
		t.Fatalf("expected one success and one already-decided result, got successes=%d already_decided=%d", successes, alreadyDecided)
	}

	completed, err := wfMod.Service.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("failed to reload workflow: %v", err)
	}
	if completed.Status != "APPROVED" {
		t.Fatalf("expected workflow APPROVED, got %s", completed.Status)
	}
}

func TestWorkflow_AllApprovalWaitsForEveryApprover(t *testing.T) {
	wfMod, _, orgMod, iamMod, db, ctx := setupWorkflowTest(t)
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed legal entity: %v", err)
	}

	uid := uuid.NewString()[:8]
	maker, err := iamMod.AuthService.SeedAdmin(ctx, "wf_all_maker_"+uid, "wf_all_maker_"+uid+"@wf.test", "ALL Maker", "Password123!")
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}
	approvers := make([]uuid.UUID, 2)
	for i := range approvers {
		suffix := string(rune('a' + i))
		user, err := iamMod.AuthService.SeedAdmin(ctx, "wf_all_"+uid+suffix, "wf_all_"+uid+suffix+"@wf.test", "ALL Approver", "Password123!")
		if err != nil {
			t.Fatalf("failed to create approver: %v", err)
		}
		approvers[i] = user.ID
	}

	definitionID := uuid.New()
	versionID := uuid.New()
	stepID := uuid.New()
	definitionCode := "ALL_APPROVAL_" + uid
	if _, err := db.Pool.Exec(ctx, `
		INSERT INTO workflow.definitions (id, code, name_ar, name_en, module)
		VALUES ($1, $2, 'اعتماد جماعي', 'All Approval Test', 'TEST')`, definitionID, definitionCode); err != nil {
		t.Fatalf("failed to create ALL definition: %v", err)
	}
	if _, err := db.Pool.Exec(ctx, `
		INSERT INTO workflow.definition_versions (id, definition_id, version, status, published_at)
		VALUES ($1, $2, 1, 'PUBLISHED', now())`, versionID, definitionID); err != nil {
		t.Fatalf("failed to create ALL definition version: %v", err)
	}
	if _, err := db.Pool.Exec(ctx, `
		INSERT INTO workflow.steps (id, version_id, step_order, code, name_ar, name_en, approval_type)
		VALUES ($1, $2, 1, 'ALL_REVIEW', 'مراجعة جماعية', 'All Review', 'ALL')`, stepID, versionID); err != nil {
		t.Fatalf("failed to create ALL step: %v", err)
	}
	for _, approverID := range approvers {
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO workflow.step_approvers (step_id, approver_type, user_id)
			VALUES ($1, 'USER', $2)`, stepID, approverID); err != nil {
			t.Fatalf("failed to create ALL step approver: %v", err)
		}
	}

	inst, err := wfMod.Service.StartWorkflow(ctx, ports.StartWorkflowCommand{
		DefinitionCode: definitionCode,
		LegalEntityID:  le.ID,
		ResourceType:   "ALL_APPROVAL_TEST",
		ResourceID:     "ALL-" + uid,
		InitiatorID:    maker.ID,
		ContextData:    map[string]any{},
	})
	if err != nil {
		t.Fatalf("failed to start ALL workflow: %v", err)
	}
	requestIDs := []uuid.UUID{
		approvalRequestForInstance(t, ctx, wfMod, approvers[0], inst.ID),
		approvalRequestForInstance(t, ctx, wfMod, approvers[1], inst.ID),
	}

	if _, err := wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: requestIDs[0],
		ActorID:   approvers[0],
		Decision:  "APPROVE",
	}); err != nil {
		t.Fatalf("first ALL approval failed: %v", err)
	}
	pending, err := wfMod.Service.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("failed to reload ALL workflow: %v", err)
	}
	if pending.Status != "PENDING" {
		t.Fatalf("expected ALL workflow to remain PENDING, got %s", pending.Status)
	}

	if _, err := wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: requestIDs[1],
		ActorID:   approvers[1],
		Decision:  "APPROVE",
	}); err != nil {
		t.Fatalf("second ALL approval failed: %v", err)
	}
	approved, err := wfMod.Service.GetInstance(ctx, inst.ID)
	if err != nil {
		t.Fatalf("failed to reload completed ALL workflow: %v", err)
	}
	if approved.Status != "APPROVED" {
		t.Fatalf("expected ALL workflow APPROVED, got %s", approved.Status)
	}
}

func TestWorkflow_ConditionsEvaluation(t *testing.T) {
	conditions := []domain.StepCondition{
		{
			FieldName:  "amount",
			Operator:   "GT",
			FieldValue: "50000.00",
		},
	}

	// Case 1: Amount = 30000 -> Should NOT meet condition
	metLow, err := domain.EvaluateConditions(conditions, map[string]any{
		"amount": "30000.00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metLow {
		t.Fatal("expected condition NOT to be met for 30000 <= 50000")
	}

	// Case 2: Amount = 75000 -> SHOULD meet condition
	metHigh, err := domain.EvaluateConditions(conditions, map[string]any{
		"amount": "75000.00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !metHigh {
		t.Fatal("expected condition to be met for 75000 > 50000")
	}
}

func TestWorkflow_DelegationRules(t *testing.T) {
	wfMod, _, _, iamMod, _, ctx := setupWorkflowTest(t)

	uid := uuid.NewString()[:8]
	userA, err := iamMod.AuthService.SeedAdmin(ctx, "delegator_"+uid, "delegator_"+uid+"@wf.test", "المفوض", "Password123!")
	if err != nil {
		t.Fatalf("failed to seed delegator: %v", err)
	}

	userB, err := iamMod.AuthService.SeedAdmin(ctx, "delegate_"+uid, "delegate_"+uid+"@wf.test", "المفوض إليه", "Password123!")
	if err != nil {
		t.Fatalf("failed to seed delegate: %v", err)
	}

	// 1. Self delegation must fail
	_, err = wfMod.Service.CreateDelegation(ctx, ports.CreateDelegationCommand{
		DelegatorID: userA.ID,
		DelegateID:  userA.ID,
		StartDate:   time.Now().UTC(),
		EndDate:     time.Now().UTC().Add(24 * time.Hour),
		Reason:      "Self delegation test",
	})
	if err != domain.ErrSelfDelegationForbidden {
		t.Fatalf("expected ErrSelfDelegationForbidden, got: %v", err)
	}

	// 2. Invalid date range must fail
	_, err = wfMod.Service.CreateDelegation(ctx, ports.CreateDelegationCommand{
		DelegatorID: userA.ID,
		DelegateID:  userB.ID,
		StartDate:   time.Now().UTC().Add(48 * time.Hour),
		EndDate:     time.Now().UTC(), // End before start
		Reason:      "Invalid date test",
	})
	if err != domain.ErrInvalidDelegationPeriod {
		t.Fatalf("expected ErrInvalidDelegationPeriod, got: %v", err)
	}

	// 3. Valid delegation creation
	del, err := wfMod.Service.CreateDelegation(ctx, ports.CreateDelegationCommand{
		DelegatorID: userA.ID,
		DelegateID:  userB.ID,
		StartDate:   time.Now().UTC().Add(-1 * time.Hour),
		EndDate:     time.Now().UTC().Add(24 * time.Hour),
		Reason:      "Annual leave delegation",
	})
	if err != nil {
		t.Fatalf("failed to create valid delegation: %v", err)
	}

	if !del.IsActive {
		t.Fatal("expected delegation to be active")
	}

	// 4. Verify Active status evaluation
	isActive := domain.IsDelegationActive(del, uuid.New(), time.Now().UTC())
	if !isActive {
		t.Fatal("expected delegation to evaluate as active")
	}

	// 5. Revoke delegation
	if err := wfMod.Service.RevokeDelegation(ctx, del.ID, userA.ID); err != nil {
		t.Fatalf("failed to revoke delegation: %v", err)
	}
}
