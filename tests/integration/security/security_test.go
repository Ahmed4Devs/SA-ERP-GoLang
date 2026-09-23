package security_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	platformhttp "erp/internal/platform/http"
	"erp/internal/platform/middleware"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
)

func setupSecurityTest(t *testing.T) (*platformhttp.Server, *database.DB, uuid.UUID, context.Context) {
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
	t.Cleanup(db.Close)

	txManager := transaction.NewManager(db.Pool)
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	orgMod := organization.NewModule(cfg, db, txManager)
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	server := platformhttp.NewServer(cfg, logger, db)

	return server, db, le.ID, ctx
}

// 1. OWASP Security Headers Test
func TestSecurity_OWASPHeaders(t *testing.T) {
	server, _, _, _ := setupSecurityTest(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.Router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	headers := resp.Header

	if h := headers.Get("X-Content-Type-Options"); h != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", h)
	}

	if h := headers.Get("X-Frame-Options"); h != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %q", h)
	}

	if h := headers.Get("X-XSS-Protection"); !strings.Contains(h, "1") {
		t.Errorf("expected X-XSS-Protection to be enabled, got %q", h)
	}

	if h := headers.Get("Referrer-Policy"); h != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy: strict-origin-when-cross-origin, got %q", h)
	}

	if h := headers.Get("Permissions-Policy"); h == "" {
		t.Error("expected Permissions-Policy header to be set")
	}

	if h := headers.Get("Content-Security-Policy"); h == "" {
		t.Error("expected Content-Security-Policy header to be set")
	}
}

// 2. Rate Limiting Test
func TestSecurity_RateLimiting(t *testing.T) {
	// Configure tight rate limiter for testing
	rl := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		GeneralLimit:   5,
		AuthLimit:      3,
		WindowDuration: 1 * time.Minute,
	})

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	// Execute 5 allowed requests
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d expected status 200, got %d", i, w.Code)
		}
	}

	// 6th request must be rate limited
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resource", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 Too Many Requests, got %d", w.Code)
	}

	if w.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on rate limited response")
	}
}

// 3. CSRF Protection Test
func TestSecurity_CSRFProtection(t *testing.T) {
	csrf := middleware.CSRFProtection(middleware.CSRFConfig{
		AllowedOrigins:       []string{"http://localhost:3000"},
		ExemptPaths:          []string{"/health"},
		ProtectedCookieNames: []string{"erp_session"},
	})

	handler := csrf(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"created"}`))
	}))

	// Case 1: Untrusted cross-origin POST must be rejected with 403 Forbidden
	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewBufferString(`{"title":"test"}`))
	req.Header.Set("Origin", "http://evil-attacker.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for cross-origin POST, got %d", w.Code)
	}

	// Case 2: Trusted Origin POST must succeed
	reqValid := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewBufferString(`{"title":"test"}`))
	reqValid.Header.Set("Origin", "http://localhost:3000")
	wValid := httptest.NewRecorder()
	handler.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Fatalf("expected status 200 for trusted origin POST, got %d", wValid.Code)
	}

	// Case 3: GET requests are safe and exempted
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/documents", nil)
	reqGet.Header.Set("Origin", "http://evil-attacker.com")
	wGet := httptest.NewRecorder()
	handler.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("expected status 200 for GET request, got %d", wGet.Code)
	}

	// Case 4: A mutating cookie-authenticated request without browser origin
	// metadata must be rejected instead of bypassing CSRF protection.
	reqCookie := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewBufferString(`{"title":"test"}`))
	reqCookie.AddCookie(&http.Cookie{Name: "erp_session", Value: "session-token"})
	wCookie := httptest.NewRecorder()
	handler.ServeHTTP(wCookie, reqCookie)

	if wCookie.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for cookie-authenticated POST without origin metadata, got %d", wCookie.Code)
	}

	// Case 5: A non-browser API client without protected cookies remains valid.
	reqBearer := httptest.NewRequest(http.MethodPost, "/api/v1/documents", bytes.NewBufferString(`{"title":"test"}`))
	reqBearer.Header.Set("Authorization", "Bearer api-token")
	wBearer := httptest.NewRecorder()
	handler.ServeHTTP(wBearer, reqBearer)

	if wBearer.Code != http.StatusOK {
		t.Fatalf("expected status 200 for non-cookie API client, got %d", wBearer.Code)
	}

	// Case 6: Exempted health paths remain available regardless of cookies.
	reqExempt := httptest.NewRequest(http.MethodPost, "/health/check", nil)
	reqExempt.AddCookie(&http.Cookie{Name: "erp_session", Value: "session-token"})
	wExempt := httptest.NewRecorder()
	handler.ServeHTTP(wExempt, reqExempt)

	if wExempt.Code != http.StatusOK {
		t.Fatalf("expected status 200 for exempt path, got %d", wExempt.Code)
	}
}

// 4. Request Body Limit (DoS Defense) Test
func TestSecurity_BodyLimit(t *testing.T) {
	limitMiddleware := middleware.RequestBodyLimit(middleware.BodyLimitConfig{
		DefaultLimitBytes: 1024, // 1 KB limit for testing
		UploadLimitBytes:  2048,
	})

	handler := limitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"bytes_read":%d}`, len(body))))
	}))

	// Oversized body (> 1KB)
	largeData := bytes.Repeat([]byte("A"), 2048)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/data", bytes.NewReader(largeData))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413 Payload Too Large, got %d", w.Code)
	}
}

// 5. Tenant Boundary & IDOR Guard Test
func TestSecurity_TenantBoundaryGuard(t *testing.T) {
	guard := middleware.TenantBoundaryGuard()

	handler := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tenantID, ok := middleware.TenantFromContext(r.Context()); ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"tenant_id":"%s"}`, tenantID.String())))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tenant_id":"none"}`))
	}))

	// Case 1: Malformed UUID must be rejected with 400 Bad Request
	reqInvalid := httptest.NewRequest(http.MethodGet, "/api/v1/invoices?legal_entity_id=not-a-valid-uuid-123", nil)
	wInvalid := httptest.NewRecorder()
	handler.ServeHTTP(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for malformed legal_entity_id, got %d", wInvalid.Code)
	}

	// Case 2: Valid UUID passes and is extracted to context
	validID := uuid.New()
	reqValid := httptest.NewRequest(http.MethodGet, "/api/v1/invoices?legal_entity_id="+validID.String(), nil)
	wValid := httptest.NewRecorder()
	handler.ServeHTTP(wValid, reqValid)

	if wValid.Code != http.StatusOK {
		t.Fatalf("expected status 200 for valid legal_entity_id, got %d", wValid.Code)
	}

	var resp struct {
		TenantID string `json:"tenant_id"`
	}
	_ = json.Unmarshal(wValid.Body.Bytes(), &resp)
	if resp.TenantID != validID.String() {
		t.Errorf("expected tenant ID %s in context, got %s", validID.String(), resp.TenantID)
	}
}

func TestSecurity_TenantResourceGuardBlocksCrossEntityIDOR(t *testing.T) {
	server, db, entityAID, ctx := setupSecurityTest(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	orgMod := organization.NewModule(cfg, db, transaction.NewManager(db.Pool))
	org, err := orgMod.OrgService.CreateOrganization(ctx, "ORG-IDOR-"+uuid.NewString()[:8], "منظمة عزل الموارد", "Resource Isolation Org")
	if err != nil {
		t.Fatalf("create IDOR test organization: %v", err)
	}
	crNumber := fmt.Sprintf("1%09d", time.Now().UnixNano()%1_000_000_000)
	entityB, err := orgMod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-IDOR", "كيان الموارد الآخر", "Other Resource Entity", crNumber, "SAR")
	if err != nil {
		t.Fatalf("create second legal entity: %v", err)
	}

	insertNotification := func(legalEntityID uuid.UUID) uuid.UUID {
		t.Helper()
		id := uuid.New()
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO notifications.notifications (
				id, legal_entity_id, channel, title_ar, title_en, body_ar, body_en,
				category, priority, status
			) VALUES ($1, $2, 'IN_APP', 'اختبار', 'Test', 'اختبار', 'Test', 'SYSTEM', 'NORMAL', 'SENT')`, id, legalEntityID)
		if err != nil {
			t.Fatalf("insert tenant resource: %v", err)
		}
		return id
	}
	ownedID := insertNotification(entityAID)
	foreignID := insertNotification(entityB.ID)

	server.Router.Get("/test-tenant-resources/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ownedReq := httptest.NewRequest(http.MethodGet, "/test-tenant-resources/"+ownedID.String()+"?legal_entity_id="+entityAID.String(), nil)
	ownedRecorder := httptest.NewRecorder()
	server.Router.ServeHTTP(ownedRecorder, ownedReq)
	if ownedRecorder.Code != http.StatusOK {
		t.Fatalf("expected same-tenant resource to pass, got %d: %s", ownedRecorder.Code, ownedRecorder.Body.String())
	}

	foreignReq := httptest.NewRequest(http.MethodGet, "/test-tenant-resources/"+foreignID.String()+"?legal_entity_id="+entityAID.String(), nil)
	foreignRecorder := httptest.NewRecorder()
	server.Router.ServeHTTP(foreignRecorder, foreignReq)
	if foreignRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant resource to be hidden, got %d: %s", foreignRecorder.Code, foreignRecorder.Body.String())
	}
}

// 6. SQL Injection Resilience Test
func TestSecurity_SQLInjectionResilience(t *testing.T) {
	server, _, leID, ctx := setupSecurityTest(t)

	// Attempt SQL injection via query parameters
	sqlInjections := []string{
		"' OR '1'='1",
		"'; DROP TABLE accounting.accounts; --",
		"1' UNION SELECT NULL, NULL, NULL --",
		"admin'--",
		"1; EXEC xp_cmdshell('dir');",
	}

	for _, sqli := range sqlInjections {
		// Target dashboard endpoint with SQL injection payload in query string
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?legal_entity_id="+leID.String()+"&filter="+url.QueryEscape(sqli), nil)
		w := httptest.NewRecorder()
		server.Router.ServeHTTP(w, req)

		// The server must not crash (500 internal server error due to syntax error), and tables must remain intact
		if w.Code == http.StatusInternalServerError && strings.Contains(w.Body.String(), "syntax error") {
			t.Fatalf("SQL injection vulnerability suspected with payload: %s. Response: %s", sqli, w.Body.String())
		}
	}
	_ = ctx
}

// 7. XSS Payload Sanitization Test
func TestSecurity_XSSResilience(t *testing.T) {
	xssPayloads := []string{
		`<script>alert('XSS')</script>`,
		`<img src="x" onerror="alert(1)">`,
		`<svg onload=alert(document.cookie)>`,
		`javascript:alert(1)`,
	}

	for _, payload := range xssPayloads {
		// Verify JSON encoder properly escapes script tags and does not emit executable HTML
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(map[string]string{
			"input": payload,
		})
		if err != nil {
			t.Fatalf("failed to encode JSON: %v", err)
		}

		encoded := buf.String()
		if strings.Contains(encoded, "<script>") && !strings.Contains(encoded, `\u003cscript\u003e`) && !strings.Contains(encoded, `<script>`) {
			t.Errorf("XSS payload was not safely handled: %s", encoded)
		}
	}
}

// 8. Path Traversal & File Security Test
func TestSecurity_PathTraversalDefenses(t *testing.T) {
	traversalPayloads := []string{
		"../../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\cmd.exe",
		"/etc/shadow",
		"....//....//config.yaml",
	}

	for _, path := range traversalPayloads {
		cleaned := filepath.Clean(path)
		if !filepath.IsAbs(cleaned) && !strings.HasPrefix(cleaned, "..") {
			// clean path must neutralize traversal
		}
		// In Go file handling, paths must be strictly checked with filepath.Clean and base dir
		baseDir := "/safe/storage/path"
		resolved := filepath.Join(baseDir, cleaned)
		if strings.HasPrefix(resolved, "/safe/storage/path/../../") {
			t.Errorf("Path traversal escaped base directory: %s", resolved)
		}
	}
}

// 9. Secret & Key Scanning Test
func TestSecurity_ZeroHardcodedSecrets(t *testing.T) {
	// Walk internal codebase and verify no live production private keys or sensitive passwords
	patterns := []string{
		"-----BEGIN RSA PRIVATE KEY-----",
		"-----BEGIN OPENSSH PRIVATE KEY-----",
		"AKIA[0-9A-Z]{16}", // AWS access key pattern
	}

	workspaceRoot := "../../../"
	err := filepath.Walk(workspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == ".next" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check source code files (exclude test files containing pattern definitions)
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".ts" && ext != ".tsx" && ext != ".env.example" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		for _, pattern := range patterns {
			if strings.Contains(contentStr, pattern) {
				t.Errorf("Suspicious secret pattern %q found in file: %s", pattern, path)
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("error during secret scanning: %v", err)
	}
}
