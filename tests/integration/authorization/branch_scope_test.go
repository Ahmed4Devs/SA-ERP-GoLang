package authorization_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/authorization"

	"erp/internal/authorization/domain"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/middleware"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestBranchGrantDoesNotBecomeEntityOrGlobalGrant(t *testing.T) {
	authz, org, iam, ctx := setupAuthzTest(t)
	organization, err := org.OrgService.CreateOrganization(ctx, "ORG-"+uuid.NewString()[:8], "اختبار", "Test")
	if err != nil {
		t.Fatal(err)
	}
	entity, err := org.OrgService.CreateLegalEntity(ctx, organization.ID, "LE-A", "اختبار", "Test", "", "SAR")
	if err != nil {
		t.Fatal(err)
	}
	otherEntity, err := org.OrgService.CreateLegalEntity(ctx, organization.ID, "LE-B", "اختبار", "Test", "", "SAR")
	if err != nil {
		t.Fatal(err)
	}
	branch, err := org.OrgService.CreateBranch(ctx, entity.ID, "BR-A", "فرع", "Branch")
	if err != nil {
		t.Fatal(err)
	}
	otherBranch, err := org.OrgService.CreateBranch(ctx, entity.ID, "BR-B", "فرع", "Branch")
	if err != nil {
		t.Fatal(err)
	}
	user, err := iam.AuthService.SeedAdmin(ctx, "branch_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Branch", "Branch-password")
	if err != nil {
		t.Fatal(err)
	}
	roles, err := authz.Service.ListRoles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var roleID uuid.UUID
	for _, role := range roles {
		if role.Code == "SUPER_ADMIN" {
			roleID = role.ID
		}
	}
	if _, err := authz.Service.AssignRole(ctx, user.ID, roleID, &entity.ID, &branch.ID); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name           string
		entity, branch *uuid.UUID
		allow          bool
	}{
		{"own branch", &entity.ID, &branch.ID, true},
		{"another branch", &entity.ID, &otherBranch.ID, false},
		{"whole entity", &entity.ID, nil, false},
		{"global", nil, nil, false},
		{"branch without entity", nil, &branch.ID, false},
		{"mismatched entity", &otherEntity.ID, &branch.ID, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decision, err := authz.Service.Authorize(ctx, domain.AuthorizationRequest{UserID: user.ID, Permission: "iam.users.manage", LegalEntityID: tc.entity, BranchID: tc.branch})
			if err != nil || decision.Allowed != tc.allow {
				t.Fatalf("decision=%+v err=%v", decision, err)
			}
		})
	}
	if _, err := authz.Service.AssignRole(ctx, user.ID, roleID, nil, &branch.ID); err == nil {
		t.Fatal("global branch assignment accepted")
	}
	if _, err := authz.Service.AssignRole(ctx, user.ID, roleID, &otherEntity.ID, &branch.ID); err == nil {
		t.Fatal("cross-entity branch assignment accepted")
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Pool.Exec(ctx, "INSERT INTO authz.user_role_assignments (user_id, role_id, legal_entity_id, branch_id) VALUES ($1,$2,$3,$4)", user.ID, roleID, otherEntity.ID, branch.ID); err == nil {
		t.Fatal("database accepted a cross-entity branch assignment")
	}
	if _, err := db.Pool.Exec(ctx, "INSERT INTO authz.user_role_assignments (user_id, role_id, branch_id) VALUES ($1,$2,$3)", user.ID, roleID, branch.ID); err == nil {
		t.Fatal("database accepted a global branch assignment")
	}

	// Even an entity-scoped wildcard must not administer global roles using
	// an attacker-selected target entity in the body.
	if _, err := authz.Service.AssignRole(ctx, user.ID, roleID, &entity.ID, nil); err != nil {
		t.Fatal(err)
	}
	handler := middleware.TenantBoundaryGuard()(authz.Handler.Routes(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(iamhttp.ContextWithUser(r.Context(), user)))
		})
	}))
	body, _ := json.Marshal(map[string]any{"user_id": user.ID, "role_id": roleID, "legal_entity_id": entity.ID})
	request := httptest.NewRequest("POST", "/assignments", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 403 {
		t.Fatalf("scoped administrator changed global roles: %d %s", response.Code, response.Body.String())
	}
}

func TestRevocationAcrossReplicasAndLastAdministrator(t *testing.T) {
	first, _, iam, ctx := setupAuthzTest(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	second := authorization.NewModule(cfg, db)
	target, err := iam.AuthService.SeedAdmin(ctx, "revoked_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Revoked", "Revoked-password")
	if err != nil {
		t.Fatal(err)
	}
	other, err := iam.AuthService.SeedAdmin(ctx, "remaining_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Remaining", "Remaining-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := first.SeedAdminRole(ctx, target.ID); err != nil {
		t.Fatal(err)
	}
	if err := first.SeedAdminRole(ctx, other.ID); err != nil {
		t.Fatal(err)
	}
	request := domain.AuthorizationRequest{UserID: target.ID, Permission: "iam.users.manage"}
	if decision, err := first.Service.Authorize(ctx, request); err != nil || !decision.Allowed {
		t.Fatalf("initial grant: %+v %v", decision, err)
	}
	assignments, err := first.Service.ListUserAssignments(ctx, target.ID)
	if err != nil || len(assignments) != 1 {
		t.Fatalf("assignments: %v", err)
	}
	roleID := assignments[0].RoleID
	if err := second.Service.RevokeRole(ctx, target.ID, roleID, nil, nil); err != nil {
		t.Fatal(err)
	}
	if decision, err := first.Service.Authorize(ctx, request); err != nil || decision.Allowed {
		t.Fatalf("stale replica grant: %+v %v", decision, err)
	}
	if permissions, err := first.Service.GetUserPermissions(ctx, target.ID, nil); err != nil || len(permissions) != 0 {
		t.Fatalf("stale UI permissions: %v %v", permissions, err)
	}

	// Isolate the last-admin scenario in a transaction that is always rolled
	// back, preserving all other test fixtures in this package.
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	txCtx := transaction.ContextWithTx(ctx, tx)
	if _, err := tx.Exec(ctx, "DELETE FROM authz.user_role_assignments WHERE user_id <> $1", other.ID); err != nil {
		t.Fatal(err)
	}
	if err := second.Service.RevokeRole(txCtx, other.ID, roleID, nil, nil); err == nil {
		t.Fatal("last administrator revoked")
	}
	var count int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM authz.user_role_assignments WHERE user_id=$1", other.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("last grant changed: %d %v", count, err)
	}
}
