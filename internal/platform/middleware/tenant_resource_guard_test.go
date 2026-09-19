package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type ownershipResult struct {
	here  bool
	other bool
	err   error
}

type fakeOwnershipChecker map[uuid.UUID]ownershipResult

func (f fakeOwnershipChecker) Ownership(_ context.Context, resourceID, _ uuid.UUID) (bool, bool, error) {
	result := f[resourceID]
	return result.here, result.other, result.err
}

func TestTenantResourceGuardHidesCrossTenantPathResource(t *testing.T) {
	tenantID := uuid.New()
	foreignID := uuid.New()
	checker := fakeOwnershipChecker{foreignID: {other: true}}
	handler := TenantBoundaryGuard()(TenantResourceGuard(checker, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+foreignID.String()+"?legal_entity_id="+tenantID.String(), nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected cross-tenant resource to be hidden with 404, got %d", recorder.Code)
	}
}

func TestTenantResourceGuardChecksJSONForeignKeysAndPreservesBody(t *testing.T) {
	tenantID := uuid.New()
	ownedID := uuid.New()
	checker := fakeOwnershipChecker{ownedID: {here: true}}
	called := false
	handler := TenantBoundaryGuard()(TenantResourceGuard(checker, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), ownedID.String()) {
			t.Fatal("guard did not preserve JSON body")
		}
		w.WriteHeader(http.StatusCreated)
	})))

	body := `{"legal_entity_id":"` + tenantID.String() + `","lines":[{"item_id":"` + ownedID.String() + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated || !called {
		t.Fatalf("expected owned JSON reference to reach handler, got %d", recorder.Code)
	}
}

func TestTenantResourceGuardAllowsGlobalOrUnknownID(t *testing.T) {
	tenantID := uuid.New()
	globalID := uuid.New()
	handler := TenantBoundaryGuard()(TenantResourceGuard(fakeOwnershipChecker{}, nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+globalID.String()+"?legal_entity_id="+tenantID.String(), nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected unknown/global UUID to pass ownership guard, got %d", recorder.Code)
	}
}
