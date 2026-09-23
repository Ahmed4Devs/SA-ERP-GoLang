package audit_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"erp/internal/audit"
	"erp/internal/audit/domain"
	"erp/internal/audit/ports"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cfg, err := config.Load()
	if err == nil {
		if db, err := database.NewPool(ctx, cfg); err == nil {
			_, _ = db.Pool.Exec(ctx, "TRUNCATE audit.access_events, audit.security_events, audit.entity_changes, audit.events CASCADE;")
			db.Close()
		}
	}
	os.Exit(m.Run())
}

func setupAuditTest(t *testing.T) (*audit.Module, *organization.Module, *iam.Module, *database.DB, context.Context) {
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
	orgMod := organization.NewModule(cfg, db, txManager)
	auditMod := audit.NewModule(cfg, db)

	return auditMod, orgMod, iamMod, db, ctx
}

func TestAudit_AppendOnlyTriggerEnforcement(t *testing.T) {
	auditMod, _, _, db, ctx := setupAuditTest(t)

	// 1. Insert a test audit event
	event, err := auditMod.Service.RecordEvent(ctx, ports.RecordEventCommand{
		Action:       "CREATE",
		ResourceType: "TEST_RESOURCE",
		ResourceID:   uuid.NewString(),
		Payload:      []byte(`{"immutable": true}`),
		IPAddress:    "127.0.0.1",
		UserAgent:    "IntegrationTest/1.0",
	})
	if err != nil {
		t.Fatalf("failed to record audit event: %v", err)
	}

	// 2. Attempt UPDATE on audit.events -> MUST FAIL
	_, err = db.Pool.Exec(ctx, "UPDATE audit.events SET action = 'TAMPERED' WHERE id = $1", event.ID)
	if err == nil {
		t.Fatal("expected UPDATE on audit.events to fail due to append-only trigger, but it succeeded!")
	}
	if !strings.Contains(err.Error(), "strictly append-only") {
		t.Fatalf("expected trigger error message mentioning 'strictly append-only', got: %v", err)
	}

	// 3. Attempt DELETE on audit.events -> MUST FAIL
	_, err = db.Pool.Exec(ctx, "DELETE FROM audit.events WHERE id = $1", event.ID)
	if err == nil {
		t.Fatal("expected DELETE on audit.events to fail due to append-only trigger, but it succeeded!")
	}
	if !strings.Contains(err.Error(), "strictly append-only") {
		t.Fatalf("expected trigger error message mentioning 'strictly append-only', got: %v", err)
	}

	// 4. Test security_events immutability
	secEvent, err := auditMod.Service.RecordSecurityEvent(ctx, ports.RecordSecurityEventCommand{
		EventType: "LOGIN_SUCCESS",
		Status:    "SUCCESS",
		Details:   []byte(`{"test": true}`),
		IPAddress: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("failed to record security event: %v", err)
	}

	_, err = db.Pool.Exec(ctx, "UPDATE audit.security_events SET status = 'TAMPERED' WHERE id = $1", secEvent.ID)
	if err == nil {
		t.Fatal("expected UPDATE on audit.security_events to fail, but it succeeded!")
	}
	_, err = db.Pool.Exec(ctx, "DELETE FROM audit.security_events WHERE id = $1", secEvent.ID)
	if err == nil {
		t.Fatal("expected DELETE on audit.security_events to fail, but it succeeded!")
	}
}

func TestAudit_CryptographicHashChaining(t *testing.T) {
	auditMod, _, _, _, ctx := setupAuditTest(t)

	const totalEvents = 5
	events := make([]*domain.AuditEvent, totalEvents)

	for i := 0; i < totalEvents; i++ {
		ev, err := auditMod.Service.RecordEvent(ctx, ports.RecordEventCommand{
			Action:       "CHAIN_TEST",
			ResourceType: "FINANCIAL_ENTRY",
			ResourceID:   uuid.NewString(),
			Payload:      []byte(`{"step": ` + string(rune('0'+i)) + `}`),
			IPAddress:    "127.0.0.1",
		})
		if err != nil {
			t.Fatalf("failed to record event %d: %v", i, err)
		}
		events[i] = ev
	}

	// Verify sequential hash chaining
	for i := 1; i < totalEvents; i++ {
		if events[i].PreviousHash != events[i-1].RecordHash {
			t.Fatalf("chain broken at index %d: previous_hash %s does not equal previous event record_hash %s",
				i, events[i].PreviousHash, events[i-1].RecordHash)
		}
		if !domain.ValidateRecordHash(events[i]) {
			t.Fatalf("event %d record hash validation failed", i)
		}
	}
}

func TestAudit_ConcurrentWritersAcrossServiceInstancesKeepSingleChain(t *testing.T) {
	auditMod, _, _, db, ctx := setupAuditTest(t)
	secondInstance := audit.NewModule(&config.Config{}, db)

	const writerCount = 24
	errors := make(chan error, writerCount)
	var wg sync.WaitGroup
	for i := 0; i < writerCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			service := auditMod.Service
			if index%2 == 1 {
				service = secondInstance.Service
			}
			_, err := service.RecordEvent(ctx, ports.RecordEventCommand{
				Action:       "CONCURRENT_WRITE",
				ResourceType: "AUDIT_CHAIN_TEST",
				ResourceID:   uuid.NewString(),
				Payload:      []byte(`{"writer": "concurrent"}`),
			})
			errors <- err
		}(i)
	}
	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent audit write failed: %v", err)
		}
	}
	result, err := auditMod.Service.VerifyChainIntegrity(ctx)
	if err != nil {
		t.Fatalf("verify concurrent chain: %v", err)
	}
	if !result.IsValid {
		t.Fatalf("concurrent writers forked the audit chain: %s", result.Details)
	}
}

func TestAudit_VerifyChainIntegrity(t *testing.T) {
	auditMod, _, _, _, ctx := setupAuditTest(t)

	// Run full chain integrity verification across all historical records
	res, err := auditMod.Service.VerifyChainIntegrity(ctx)
	if err != nil {
		t.Fatalf("VerifyChainIntegrity failed with unexpected error: %v", err)
	}

	if !res.IsValid {
		t.Fatalf("cryptographic audit chain integrity verification reported invalid: %s (failed at: %v)",
			res.Details, res.FailedAtEventID)
	}

	if res.VerifiedRecords == 0 {
		t.Fatal("expected at least 1 verified record, got 0")
	}

	if res.VerifiedRecords != res.TotalRecords {
		t.Fatalf("expected all %d records to be verified, verified: %d", res.TotalRecords, res.VerifiedRecords)
	}
}

func TestAudit_EntityStateChangesDiff(t *testing.T) {
	auditMod, _, _, _, ctx := setupAuditTest(t)

	recordID := uuid.NewString()
	beforeJSON := []byte(`{"status": "DRAFT", "amount": 1000.00}`)
	afterJSON := []byte(`{"status": "POSTED", "amount": 1000.00}`)

	event, err := auditMod.Service.RecordEvent(ctx, ports.RecordEventCommand{
		Action:       "POST",
		ResourceType: "JOURNAL_ENTRY",
		ResourceID:   recordID,
		Payload:      []byte(`{"note": "Monthly VAT settlement"}`),
		IPAddress:    "10.0.0.1",
		Changes: []ports.EntityChangeCommand{
			{
				TableName:   "accounting.journal_entries",
				RecordID:    recordID,
				BeforeState: beforeJSON,
				AfterState:  afterJSON,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to record event with changes: %v", err)
	}

	retrieved, err := auditMod.Service.GetEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to get event by ID: %v", err)
	}

	if len(retrieved.Changes) != 1 {
		t.Fatalf("expected 1 entity change attached, got %d", len(retrieved.Changes))
	}

	ch := retrieved.Changes[0]
	if ch.TableName != "accounting.journal_entries" {
		t.Fatalf("unexpected table name: %s", ch.TableName)
	}
	var beforeExpected, beforeActual map[string]any
	if err := json.Unmarshal(beforeJSON, &beforeExpected); err != nil {
		t.Fatalf("failed to unmarshal beforeJSON: %v", err)
	}
	if err := json.Unmarshal(ch.BeforeState, &beforeActual); err != nil {
		t.Fatalf("failed to unmarshal ch.BeforeState: %v", err)
	}
	if !reflect.DeepEqual(beforeExpected, beforeActual) {
		t.Fatalf("expected before state %v, got %v", beforeExpected, beforeActual)
	}

	var afterExpected, afterActual map[string]any
	if err := json.Unmarshal(afterJSON, &afterExpected); err != nil {
		t.Fatalf("failed to unmarshal afterJSON: %v", err)
	}
	if err := json.Unmarshal(ch.AfterState, &afterActual); err != nil {
		t.Fatalf("failed to unmarshal ch.AfterState: %v", err)
	}
	if !reflect.DeepEqual(afterExpected, afterActual) {
		t.Fatalf("expected after state %v, got %v", afterExpected, afterActual)
	}
}

func TestAudit_SearchAndFiltering(t *testing.T) {
	auditMod, _, _, _, ctx := setupAuditTest(t)

	uniqueAction := "UNIQUE_AUDIT_ACTION_" + uuid.NewString()[:8]
	_, err := auditMod.Service.RecordEvent(ctx, ports.RecordEventCommand{
		Action:       uniqueAction,
		ResourceType: "CUSTOM_RESOURCE",
		ResourceID:   "RES-001",
		Payload:      []byte(`{"test": true}`),
	})
	if err != nil {
		t.Fatalf("failed to record event: %v", err)
	}

	// Filter by unique action
	events, total, err := auditMod.Service.SearchEvents(ctx, domain.AuditFilter{
		Action: uniqueAction,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("search events failed: %v", err)
	}
	if total < 1 || len(events) < 1 {
		t.Fatalf("expected at least 1 match for action %s, got total=%d, len=%d", uniqueAction, total, len(events))
	}
	if events[0].Action != uniqueAction {
		t.Fatalf("expected action %s, got %s", uniqueAction, events[0].Action)
	}
}

func TestAudit_AccessEvents(t *testing.T) {
	auditMod, _, _, _, ctx := setupAuditTest(t)

	accessPath := "/api/v1/sales/invoices/" + uuid.NewString()[:6]
	accEvent, err := auditMod.Service.RecordAccessEvent(ctx, ports.RecordAccessEventCommand{
		ResourcePath: accessPath,
		Method:       "GET",
		StatusCode:   200,
		IPAddress:    "127.0.0.1",
		LatencyMS:    14,
	})
	if err != nil {
		t.Fatalf("failed to record access event: %v", err)
	}
	if accEvent.ID == uuid.Nil {
		t.Fatal("expected valid access event ID")
	}

	time.Sleep(10 * time.Millisecond)

	events, err := auditMod.Service.ListAccessEvents(ctx, nil, accessPath, 10, 0)
	if err != nil {
		t.Fatalf("failed to list access events: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected to find recorded access event for path %s", accessPath)
	}
	if events[0].ResourcePath != accessPath {
		t.Fatalf("expected path %s, got %s", accessPath, events[0].ResourcePath)
	}
}
