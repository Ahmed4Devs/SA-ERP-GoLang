package iam_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"erp/internal/iam"
	"erp/internal/iam/domain"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

func setupIAMTest(t *testing.T) (*iam.Module, *database.DB) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" {
		cfg.Database.URL = testDBURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	migrationsDir := "../../../database/migrations"
	if err := database.MigrateUp(ctx, cfg.Database.URL, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	iamMod := iam.NewModule(cfg, db, txManager)

	return iamMod, db
}

func TestIAMAuthenticationLifecycle(t *testing.T) {
	iamMod, db := setupIAMTest(t)
	defer db.Close()

	ctx := context.Background()
	testUsername := "admin_" + uuid.NewString()[:8]
	testEmail := testUsername + "@erp.sa"
	testPassword := "ComplexPass123!"

	// 1. Seed/Create User
	user, err := iamMod.AuthService.SeedAdmin(ctx, testUsername, testEmail, "مدير النظام", testPassword)
	if err != nil {
		t.Fatalf("failed to seed admin user: %v", err)
	}

	if user.DisplayName != "مدير النظام" {
		t.Errorf("unexpected display name: %s", user.DisplayName)
	}

	// 2. Login Failure with wrong password
	_, _, _, err = iamMod.AuthService.Login(ctx, testUsername, "WrongPassword999", "127.0.0.1", "TestAgent/1.0")
	if err == nil || err != domain.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials on wrong password, got: %v", err)
	}

	// 3. Login Success
	session, rawToken, loggedInUser, err := iamMod.AuthService.Login(ctx, testUsername, testPassword, "127.0.0.1", "TestAgent/1.0")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if session == nil || rawToken == "" || loggedInUser.ID != user.ID {
		t.Fatalf("unexpected login result: session=%v, token=%s, user=%+v", session, rawToken, loggedInUser)
	}

	// 4. Validate Session
	validUser, validSession, err := iamMod.AuthService.ValidateSession(ctx, rawToken)
	if err != nil {
		t.Fatalf("failed to validate session: %v", err)
	}
	if validUser.ID != user.ID || validSession.ID != session.ID {
		t.Errorf("session validation returned unexpected user or session")
	}

	// 5. Logout
	if err := iamMod.AuthService.Logout(ctx, rawToken); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// 6. Validate Session after Logout must fail
	_, _, err = iamMod.AuthService.ValidateSession(ctx, rawToken)
	if err == nil || err != domain.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound after logout, got: %v", err)
	}
}

func TestIAMHTTPTransportAndMiddleware(t *testing.T) {
	iamMod, db := setupIAMTest(t)
	defer db.Close()

	ctx := context.Background()
	testUsername := "user_" + uuid.NewString()[:8]
	testEmail := testUsername + "@erp.sa"
	testPassword := "SecuredP@ssw0rd!"

	// Seed user
	_, err := iamMod.AuthService.SeedAdmin(ctx, testUsername, testEmail, "مستخدم تجريبي", testPassword)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	handler := iamMod.Handler.Routes()

	// 1. HTTP Login Request
	loginPayload := map[string]string{
		"identifier": testUsername,
		"password":   testPassword,
	}
	body, _ := json.Marshal(loginPayload)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Check cookie
	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == iamhttp.SessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("expected erp_session cookie to be set")
	}
	if !sessionCookie.HttpOnly {
		t.Error("expected HttpOnly cookie")
	}

	// 2. Protected Route /me with valid cookie
	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.AddCookie(sessionCookie)
	meRR := httptest.NewRecorder()

	handler.ServeHTTP(meRR, meReq)

	if meRR.Code != http.StatusOK {
		t.Fatalf("expected 200 on /me with cookie, got %d", meRR.Code)
	}

	// 3. Protected Route /me without cookie must return 401
	unauthReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	unauthRR := httptest.NewRecorder()

	handler.ServeHTTP(unauthRR, unauthReq)

	if unauthRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on /me without cookie, got %d", unauthRR.Code)
	}
}
