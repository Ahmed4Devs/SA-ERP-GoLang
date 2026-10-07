package iam_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"erp/internal/authorization"
	"erp/internal/iam/application/services"
	"erp/internal/iam/domain"
	"erp/internal/platform/config"
	"github.com/google/uuid"
)

func TestUserAdministrationLifecycle(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx := context.Background()
	admin, err := mod.AuthService.SeedAdmin(ctx, "owner_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Owner", "Owner-password-123")
	if err != nil {
		t.Fatal(err)
	}
	authz := authorization.NewModule(&config.Config{}, db)
	if err := authz.SeedAdminRole(ctx, admin.ID); err != nil {
		t.Fatal(err)
	}
	mod.ConfigureUserManagement(authz.Service)
	_, token, _, err := mod.AuthService.CreateSessionForUser(ctx, admin.ID, admin.Version, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: "erp_session", Value: token}
	handler := mod.Handler.Routes()
	username := "managed_" + uuid.NewString()[:8]
	body := map[string]any{"username": username, "email": username + "@example.test", "display_name": "مستخدم اختبار", "password": "Managed-password-123", "preferred_locale": "ar-SA"}

	if response := performJSONRequest(t, handler, "GET", "/users", nil, nil); response.Code != 401 {
		t.Fatalf("unauthenticated list: %d", response.Code)
	}
	created := performJSONRequest(t, handler, "POST", "/users", body, cookie)
	if created.Code != 201 {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	if strings.Contains(created.Body.String(), "password") {
		t.Fatal("password leaked")
	}
	var user domain.User
	if err := json.Unmarshal(created.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	duplicate := performJSONRequest(t, handler, "POST", "/users", body, cookie)
	if duplicate.Code != 409 {
		t.Fatalf("duplicate user: %d %s", duplicate.Code, duplicate.Body.String())
	}
	body["role_id"] = uuid.NewString()
	if response := performJSONRequest(t, handler, "POST", "/users", body, cookie); response.Code != 400 {
		t.Fatalf("mass assignment: %d", response.Code)
	}
	list := performJSONRequest(t, handler, "GET", "/users?search="+username, nil, cookie)
	if list.Code != 200 || !strings.Contains(list.Body.String(), username) {
		t.Fatalf("list: %d %s", list.Code, list.Body.String())
	}
	if response := performJSONRequest(t, handler, "GET", "/users?limit=2147483648", nil, cookie); response.Code != 400 {
		t.Fatalf("overflow limit: %d", response.Code)
	}

	_, userToken, _, err := mod.AuthService.Login(ctx, username, "Managed-password-123", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	if response := performJSONRequest(t, handler, "GET", "/users", nil, &http.Cookie{Name: "erp_session", Value: userToken}); response.Code != 403 {
		t.Fatalf("unauthorized list: %d", response.Code)
	}
	stale := performJSONRequest(t, handler, "POST", "/users/"+user.ID.String()+"/status", map[string]any{"status": "DISABLED", "expected_version": 999}, cookie)
	if stale.Code != 409 {
		t.Fatalf("stale status: %d", stale.Code)
	}
	disabled := performJSONRequest(t, handler, "POST", "/users/"+user.ID.String()+"/status", map[string]any{"status": "DISABLED", "expected_version": user.Version}, cookie)
	if disabled.Code != 200 {
		t.Fatalf("disable: %d %s", disabled.Code, disabled.Body.String())
	}
	if _, _, err := mod.AuthService.ValidateSession(ctx, userToken); err == nil {
		t.Fatal("disabled user session still valid")
	}
	if _, _, _, err := mod.AuthService.Login(ctx, username, "Managed-password-123", "127.0.0.1", "test"); !errors.Is(err, domain.ErrUserDisabled) {
		t.Fatalf("disabled login: %v", err)
	}
	if err := json.Unmarshal(disabled.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	active, err := mod.UserService.ChangeStatus(ctx, admin.ID, user.ID, domain.UserStatusActive, user.Version)
	if err != nil {
		t.Fatal(err)
	}
	_, userToken, _, err = mod.AuthService.Login(ctx, username, "Managed-password-123", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	reset := performJSONRequest(t, handler, "POST", "/users/"+user.ID.String()+"/password", map[string]any{"password": "Replacement-password-123", "expected_version": active.Version}, cookie)
	if reset.Code != 200 {
		t.Fatalf("reset: %d %s", reset.Code, reset.Body.String())
	}
	if _, _, err := mod.AuthService.ValidateSession(ctx, userToken); err == nil {
		t.Fatal("old session survives password reset")
	}
	if _, _, _, err := mod.AuthService.Login(ctx, username, "Managed-password-123", "127.0.0.1", "test"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("old password: %v", err)
	}
	if _, _, _, err := mod.AuthService.Login(ctx, username, "Replacement-password-123", "127.0.0.1", "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.UserService.ChangeStatus(ctx, admin.ID, admin.ID, domain.UserStatusDisabled, admin.Version); !errors.Is(err, domain.ErrSelfDeactivation) {
		t.Fatalf("self deactivation: %v", err)
	}
	var audited int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM audit.security_events WHERE details->>'target_user_id' = $1", user.ID.String()).Scan(&audited); err != nil {
		t.Fatal(err)
	}
	if audited != 4 {
		t.Fatalf("expected four atomic administration audit records, got %d", audited)
	}
}

func TestUserCreationRollsBackWhenAuditFails(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	username := "rollback_" + uuid.NewString()[:8]
	_, err := mod.UserService.Create(context.Background(), uuid.New(), services.CreateUserCommand{Username: username, DisplayName: "Rollback", Password: "Rollback-password"})
	if err == nil {
		t.Fatal("nonexistent audit actor must fail")
	}
	if _, err := mod.AuthService.GetUserByIdentifier(context.Background(), username); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("user survived rollback: %v", err)
	}
}

func TestEntityScopedUserAdministratorCannotManageGlobalIdentities(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx := context.Background()
	user, err := mod.AuthService.SeedAdmin(ctx, "scoped_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Scoped", "Scoped-password")
	if err != nil {
		t.Fatal(err)
	}
	var orgID, entityID uuid.UUID
	if err := db.Pool.QueryRow(ctx, "INSERT INTO organization.organizations (code, name_ar, name_en) VALUES ($1, 'اختبار', 'Test') RETURNING id", uuid.NewString()).Scan(&orgID); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, "INSERT INTO organization.legal_entities (organization_id, code, name_ar, name_en, base_currency) VALUES ($1, $2, 'اختبار', 'Test', 'SAR') RETURNING id", orgID, uuid.NewString()).Scan(&entityID); err != nil {
		t.Fatal(err)
	}
	var roleID uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM authz.roles WHERE code = 'SUPER_ADMIN'").Scan(&roleID); err != nil {
		t.Fatal(err)
	}
	authz := authorization.NewModule(&config.Config{}, db)
	if _, err := authz.Service.AssignRole(ctx, user.ID, roleID, &entityID, nil); err != nil {
		t.Fatal(err)
	}
	mod.ConfigureUserManagement(authz.Service)
	_, token, _, err := mod.AuthService.CreateSessionForUser(ctx, user.ID, user.Version, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	response := performJSONRequest(t, mod.Handler.Routes(), "GET", "/users?legal_entity_id="+entityID.String(), nil, &http.Cookie{Name: "erp_session", Value: token})
	if response.Code != 403 {
		t.Fatalf("entity permission escalated: %d %s", response.Code, response.Body.String())
	}
}
