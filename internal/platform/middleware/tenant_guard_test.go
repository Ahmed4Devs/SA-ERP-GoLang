package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestTenantBoundaryGuardRejectsConflictingScopes(t *testing.T) {
	queryID := uuid.New()
	bodyID := uuid.New()
	handler := TenantBoundaryGuard()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	body := []byte(`{"legal_entity_id":"` + bodyID.String() + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/resource?legal_entity_id="+queryID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected conflicting tenant scope to return 400, got %d", recorder.Code)
	}
}

func TestTenantBoundaryGuardPreservesJSONBodyAndContext(t *testing.T) {
	tenantID := uuid.New()
	handler := TenantBoundaryGuard()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID, ok := TenantFromContext(r.Context())
		if !ok || contextID != tenantID {
			t.Fatalf("expected tenant context %s, got %s", tenantID, contextID)
		}
		var payload struct {
			LegalEntityID string `json:"legal_entity_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("expected preserved JSON body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	body := []byte(`{"legal_entity_id":"` + tenantID.String() + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/resource", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
}
