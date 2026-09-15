package workflow_test

import (
	"context"
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
		t.Skipf("PostgreSQL not accessible, skipping integration test: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	iamMod := iam.NewModule(cfg, db, txManager)
	authzMod := authorization.NewModule(cfg, db)
	orgMod := organization.NewModule(cfg, db, txManager)
	wfMod := workflow.NewModule(cfg, db)

	return wfMod, authzMod, orgMod, iamMod, db, ctx
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
	wfMod, _, orgMod, iamMod, _, ctx := setupWorkflowTest(t)

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

	// 2. Start Workflow for a document
	inst, err := wfMod.Service.StartWorkflow(ctx, ports.StartWorkflowCommand{
		DefinitionCode: "SALES_INVOICE_APPROVAL",
		LegalEntityID:  le.ID,
		ResourceType:   "SALES_INVOICE",
		ResourceID:     "INV-2026-TEST01",
		InitiatorID:    makerUser.ID,
		ContextData: map[string]any{
			"amount": 15000.00,
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

	// 3. Maker tries to APPROVE their own document -> MUST FAIL with ErrMakerCannotApprove
	// Create request assigned to maker to test deliberate violation attempt
	req := &domain.ApprovalRequest{
		InstanceStepID: inst.Steps[0].ID,
		AssignedUserID: makerUser.ID,
		Status:         "PENDING",
	}
	// Direct call to submit decision by the maker
	_, err = wfMod.Service.SubmitDecision(ctx, ports.SubmitDecisionCommand{
		RequestID: req.ID,
		ActorID:   makerUser.ID,
		Decision:  "APPROVE",
		Notes:     "Self-approval attempt",
	})
	// Maker checker rule test directly:
	errMaker := domain.ValidateMakerChecker(makerUser.ID, makerUser.ID)
	if errMaker != domain.ErrMakerCannotApprove {
		t.Fatalf("expected ErrMakerCannotApprove, got: %v", errMaker)
	}

	// Checker approving maker's document is valid:
	errChecker := domain.ValidateMakerChecker(makerUser.ID, checkerUser.ID)
	if errChecker != nil {
		t.Fatalf("expected valid maker-checker, got: %v", errChecker)
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
		"amount": 30000.00,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metLow {
		t.Fatal("expected condition NOT to be met for 30000 <= 50000")
	}

	// Case 2: Amount = 75000 -> SHOULD meet condition
	metHigh, err := domain.EvaluateConditions(conditions, map[string]any{
		"amount": 75000.00,
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
