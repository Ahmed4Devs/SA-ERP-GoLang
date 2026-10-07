package workflow_test

import (
	"bytes"
	"encoding/json"
	iamdomain "erp/internal/iam/domain"
	iamhttp "erp/internal/iam/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApprovalLimitAdministrationRequiresGlobalAuthorityAndVersion(t *testing.T) {
	wf, authz, org, _, db, ctx := setupWorkflowTest(t)
	entity, err := org.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatal(err)
	}
	admin, clerk := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{admin, clerk} {
		if _, err = db.Pool.Exec(ctx, `INSERT INTO iam.users(id,username,display_name,status) VALUES($1,$2,'Limits test','ACTIVE')`, id, "limit-"+id.String()); err != nil {
			t.Fatal(err)
		}
	}
	if err = authz.SeedAdminRole(ctx, admin); err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	wf.RegisterRoutes(router, nil)
	call := func(actor uuid.UUID, version int, max string) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(map[string]any{"legal_entity_id": entity.ID, "user_id": clerk, "resource_type": "TREASURY_PAYMENT", "maximum_base_amount": max, "version": version})
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPut, "/api/v1/workflow/approval-limits", bytes.NewReader(body))
		if actor != uuid.Nil {
			req = req.WithContext(iamhttp.ContextWithUser(ctx, &iamdomain.User{ID: actor}))
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	for _, actor := range []uuid.UUID{uuid.Nil, clerk} {
		if r := call(actor, 0, "100"); r.Code != 403 {
			t.Fatalf("unauthorized limit change: %d", r.Code)
		}
	}
	if r := call(admin, 0, "100.000001"); r.Code != 200 {
		t.Fatalf("limit create failed: %d %s", r.Code, r.Body.String())
	}
	if r := call(admin, 0, "999999"); r.Code != 409 {
		t.Fatalf("stale create overwrote limit: %d", r.Code)
	}
	if r := call(admin, 1, "200"); r.Code != 200 {
		t.Fatalf("versioned update failed: %d %s", r.Code, r.Body.String())
	}
	if r := call(admin, 1, "999999"); r.Code != 409 {
		t.Fatalf("stale update overwrote limit: %d", r.Code)
	}
	if _, err = db.Pool.Exec(ctx, `CREATE FUNCTION audit.fail_approval_limit_test() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='workflow.approval_limit.save' THEN RAISE EXCEPTION 'injected limit audit failure'; END IF; RETURN NEW; END $$; CREATE TRIGGER fail_approval_limit_test BEFORE INSERT ON audit.events FOR EACH ROW EXECUTE FUNCTION audit.fail_approval_limit_test();`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(ctx, `DROP TRIGGER fail_approval_limit_test ON audit.events; DROP FUNCTION audit.fail_approval_limit_test();`)
	})
	if r := call(admin, 2, "900"); r.Code == 200 {
		t.Fatal("limit change ignored mandatory audit failure")
	}
	var version int
	var max string
	if err = db.Pool.QueryRow(ctx, `SELECT version,maximum_base_amount::text FROM workflow.approval_limits WHERE legal_entity_id=$1 AND user_id=$2`, entity.ID, clerk).Scan(&version, &max); err != nil {
		t.Fatal(err)
	}
	if version != 2 || max != "200.000000" {
		t.Fatalf("failed audit changed authority: v%d %s", version, max)
	}
}
