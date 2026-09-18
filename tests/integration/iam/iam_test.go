package iam_test

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"github.com/pquerna/otp/totp"
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
	cfg.Security.MFAKey = base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))

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

func TestIAMMFAAuthenticationLifecycle(t *testing.T) {
	iamMod, db := setupIAMTest(t)
	defer db.Close()

	ctx := context.Background()
	username := "mfa_" + uuid.NewString()[:8]
	password := "MFA-ComplexPass123!"
	_, err := iamMod.AuthService.SeedAdmin(ctx, username, username+"@erp.sa", "MFA User", password)
	if err != nil {
		t.Fatalf("failed to create MFA test user: %v", err)
	}

	handler := iamMod.Handler.Routes()
	initialLogin := performJSONRequest(t, handler, http.MethodPost, "/login", map[string]string{
		"identifier": username,
		"password":   password,
	}, nil)
	if initialLogin.Code != http.StatusOK {
		t.Fatalf("initial login status = %d: %s", initialLogin.Code, initialLogin.Body.String())
	}
	sessionCookie := findSessionCookie(t, initialLogin)

	enroll := performJSONRequest(t, handler, http.MethodPost, "/mfa/enroll", map[string]string{}, sessionCookie)
	if enroll.Code != http.StatusCreated {
		t.Fatalf("MFA enrollment status = %d: %s", enroll.Code, enroll.Body.String())
	}
	if enroll.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("MFA enrollment response must not be cached")
	}
	var enrollment struct {
		Secret string   `json:"secret"`
		Codes  []string `json:"recovery_codes"`
	}
	if err := json.NewDecoder(enroll.Body).Decode(&enrollment); err != nil {
		t.Fatalf("failed to decode enrollment: %v", err)
	}
	if enrollment.Secret == "" || len(enrollment.Codes) != 10 {
		t.Fatalf("invalid enrollment response: %+v", enrollment)
	}

	confirmationCode, err := totp.GenerateCode(enrollment.Secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate enrollment TOTP: %v", err)
	}
	confirm := performJSONRequest(t, handler, http.MethodPost, "/mfa/confirm", map[string]string{"code": confirmationCode}, sessionCookie)
	if confirm.Code != http.StatusOK {
		t.Fatalf("MFA confirmation status = %d: %s", confirm.Code, confirm.Body.String())
	}

	reEnroll := performJSONRequest(t, handler, http.MethodPost, "/mfa/enroll", map[string]string{}, sessionCookie)
	if reEnroll.Code != http.StatusConflict {
		t.Fatalf("re-enrollment status = %d, want 409: %s", reEnroll.Code, reEnroll.Body.String())
	}

	challengeResponse := performJSONRequest(t, handler, http.MethodPost, "/login", map[string]string{
		"identifier": username,
		"password":   password,
	}, nil)
	if challengeResponse.Code != http.StatusAccepted {
		t.Fatalf("MFA password login status = %d: %s", challengeResponse.Code, challengeResponse.Body.String())
	}
	if len(challengeResponse.Result().Cookies()) != 0 {
		t.Fatal("password-only MFA login unexpectedly set a session cookie")
	}
	challengeToken := decodeChallengeToken(t, challengeResponse)

	complete := performJSONRequest(t, handler, http.MethodPost, "/login/mfa", map[string]string{
		"challenge_token": challengeToken,
		"code":            enrollment.Codes[0],
	}, nil)
	if complete.Code != http.StatusOK {
		t.Fatalf("MFA completion status = %d: %s", complete.Code, complete.Body.String())
	}
	findSessionCookie(t, complete)

	replayedChallenge := performJSONRequest(t, handler, http.MethodPost, "/login/mfa", map[string]string{
		"challenge_token": challengeToken,
		"code":            enrollment.Codes[0],
	}, nil)
	if replayedChallenge.Code != http.StatusUnauthorized {
		t.Fatalf("replayed MFA challenge status = %d, want 401", replayedChallenge.Code)
	}

	secondChallengeResponse := performJSONRequest(t, handler, http.MethodPost, "/login", map[string]string{
		"identifier": username,
		"password":   password,
	}, nil)
	secondChallenge := decodeChallengeToken(t, secondChallengeResponse)
	replayedRecovery := performJSONRequest(t, handler, http.MethodPost, "/login/mfa", map[string]string{
		"challenge_token": secondChallenge,
		"code":            enrollment.Codes[0],
	}, nil)
	if replayedRecovery.Code != http.StatusUnauthorized {
		t.Fatalf("replayed recovery code status = %d, want 401", replayedRecovery.Code)
	}
	validRecovery := performJSONRequest(t, handler, http.MethodPost, "/login/mfa", map[string]string{
		"challenge_token": secondChallenge,
		"code":            enrollment.Codes[1],
	}, nil)
	if validRecovery.Code != http.StatusOK {
		t.Fatalf("second recovery code status = %d: %s", validRecovery.Code, validRecovery.Body.String())
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, payload any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func findSessionCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == iamhttp.SessionCookieName && cookie.Value != "" {
			return cookie
		}
	}
	t.Fatal("response does not contain a session cookie")
	return nil
}

func decodeChallengeToken(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("challenge status = %d: %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		MFARequired    bool   `json:"mfa_required"`
		ChallengeToken string `json:"challenge_token"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode challenge response: %v", err)
	}
	if !response.MFARequired || response.ChallengeToken == "" {
		t.Fatalf("invalid challenge response: %+v", response)
	}
	return response.ChallengeToken
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
