package authorization_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"erp/internal/authorization"
	"erp/internal/authorization/domain"
	"erp/internal/iam"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

func setupAuthzTest(t *testing.T) (*authorization.Module, *organization.Module, *iam.Module, context.Context) {
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

	return authzMod, orgMod, iamMod, ctx
}

func TestAuthorization_RolesAndPermissionsListing(t *testing.T) {
	authzMod, _, _, ctx := setupAuthzTest(t)

	roles, err := authzMod.Service.ListRoles(ctx)
	if err != nil {
		t.Fatalf("failed to list roles: %v", err)
	}
	if len(roles) == 0 {
		t.Fatalf("expected seeded system roles, got 0")
	}

	var foundSuperAdmin bool
	for _, r := range roles {
		if r.Code == "SUPER_ADMIN" {
			foundSuperAdmin = true
			break
		}
	}
	if !foundSuperAdmin {
		t.Errorf("SUPER_ADMIN role not found in seeded roles")
	}

	perms, err := authzMod.Service.ListPermissions(ctx)
	if err != nil {
		t.Fatalf("failed to list permissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatalf("expected seeded permissions, got 0")
	}
}

func TestAuthorization_LegalEntityScopeEnforcement(t *testing.T) {
	authzMod, orgMod, iamMod, ctx := setupAuthzTest(t)

	// 1. Setup two separate legal entities to test isolation boundary
	org, err := orgMod.OrgService.CreateOrganization(ctx, fmt.Sprintf("ORG-AUTHZ-%d", time.Now().UnixNano()%100000), "منظمة الصلاحيات", "Authz Org")
	if err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	entityA, err := orgMod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-A", "الشركة الأولى", "Entity A", "1010111111", "SAR")
	if err != nil {
		t.Fatalf("failed to create entity A: %v", err)
	}

	entityB, err := orgMod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-B", "الشركة الثانية", "Entity B", "1010222222", "SAR")
	if err != nil {
		t.Fatalf("failed to create entity B: %v", err)
	}

	// 2. Create actual user in IAM
	testUsername := fmt.Sprintf("acc_%s", uuid.New().String()[:8])
	createdUser, err := iamMod.AuthService.SeedAdmin(ctx, testUsername, testUsername+"@erp.sa", "محاسب تجريبي", "AccountantPass123!")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	testUserID := createdUser.ID

	// 3. Find CHIEF_ACCOUNTANT role
	roles, err := authzMod.Service.ListRoles(ctx)
	if err != nil {
		t.Fatalf("failed to list roles: %v", err)
	}
	var accountantRole *domain.Role
	for _, r := range roles {
		if r.Code == "CHIEF_ACCOUNTANT" {
			accountantRole = r
			break
		}
	}
	if accountantRole == nil {
		t.Fatalf("CHIEF_ACCOUNTANT role not found")
	}

	// 4. Assign CHIEF_ACCOUNTANT to testUserID scoped ONLY to Entity A
	_, err = authzMod.Service.AssignRole(ctx, testUserID, accountantRole.ID, &entityA.ID, nil)
	if err != nil {
		t.Fatalf("failed to assign scoped role: %v", err)
	}

	// 5. Test Authorization in Entity A scope: Should be ALLOWED
	decisionA, err := authzMod.Service.Authorize(ctx, domain.AuthorizationRequest{
		UserID:        testUserID,
		Permission:    "gl.journals.post",
		LegalEntityID: &entityA.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error during authorize A: %v", err)
	}
	if !decisionA.Allowed {
		t.Errorf("expected permission to be granted in Entity A scope, got denied: %s", decisionA.Reason)
	}

	// 6. Test Authorization in Entity B scope: MUST BE DENIED (Strict multi-entity isolation!)
	decisionB, err := authzMod.Service.Authorize(ctx, domain.AuthorizationRequest{
		UserID:        testUserID,
		Permission:    "gl.journals.post",
		LegalEntityID: &entityB.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error during authorize B: %v", err)
	}
	if decisionB.Allowed {
		t.Errorf("SECURITY BREACH: user granted permission in unauthorized Entity B scope!")
	}

	// 7. Test Unheld Permission in Entity A scope: Should be DENIED
	decisionUnheld, err := authzMod.Service.Authorize(ctx, domain.AuthorizationRequest{
		UserID:        testUserID,
		Permission:    "iam.users.manage",
		LegalEntityID: &entityA.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error during authorize unheld: %v", err)
	}
	if decisionUnheld.Allowed {
		t.Errorf("expected unauthorized permission to be denied, but was allowed")
	}

	// 8. Revoke role and verify cache invalidation
	err = authzMod.Service.RevokeRole(ctx, testUserID, accountantRole.ID, &entityA.ID, nil)
	if err != nil {
		t.Fatalf("failed to revoke role: %v", err)
	}

	decisionAfterRevoke, err := authzMod.Service.Authorize(ctx, domain.AuthorizationRequest{
		UserID:        testUserID,
		Permission:    "gl.journals.post",
		LegalEntityID: &entityA.ID,
	})
	if err != nil {
		t.Fatalf("unexpected error after revoke: %v", err)
	}
	if decisionAfterRevoke.Allowed {
		t.Errorf("expected permission to be denied after role revocation, but was allowed")
	}
}

func TestAuthorization_MiddlewareEnforcement(t *testing.T) {
	authzMod, orgMod, iamMod, ctx := setupAuthzTest(t)

	org, _ := orgMod.OrgService.CreateOrganization(ctx, fmt.Sprintf("ORG-MW-%d", time.Now().UnixNano()%100000), "منظمة البرمجية", "Middleware Org")
	entity, _ := orgMod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-MW", "الكيان البرمجي", "MW Entity", "1010333333", "SAR")

	testUsername := fmt.Sprintf("mw_%d", time.Now().UnixNano()%1000000)
	testUser, err := iamMod.AuthService.SeedAdmin(ctx, testUsername, testUsername+"@erp.sa", "مدير المالية", "ComplexPass123!")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Construct protected handler
	permMiddleware := authzMod.RequirePermission("gl.journals.post", func(r *http.Request) *uuid.UUID {
		return &entity.ID
	})

	protectedHandler := permMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ACCESS_GRANTED"))
	}))

	// Case 1: Unauthorized user -> 403 Forbidden
	req := httptest.NewRequest(http.MethodPost, "/api/v1/journals", nil)
	req = req.WithContext(iamhttp.ContextWithUser(req.Context(), testUser))
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for unauthorized user, got %d", rec.Code)
	}

	// Case 2: Grant SUPER_ADMIN to user -> 200 OK
	roles, _ := authzMod.Service.ListRoles(ctx)
	var superAdminID uuid.UUID
	for _, r := range roles {
		if r.Code == "SUPER_ADMIN" {
			superAdminID = r.ID
			break
		}
	}
	_, err = authzMod.Service.AssignRole(ctx, testUser.ID, superAdminID, nil, nil)
	if err != nil {
		t.Fatalf("failed to assign super admin: %v", err)
	}

	rec2 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for authorized super admin, got %d", rec2.Code)
	}
}
