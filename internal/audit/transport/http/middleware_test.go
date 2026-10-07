package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"erp/internal/audit/domain"
	"erp/internal/audit/ports"
	iamdomain "erp/internal/iam/domain"
	iamhttp "erp/internal/iam/transport/http"

	"github.com/google/uuid"
)

type recordingAuditService struct {
	events   []ports.RecordEventCommand
	accesses []ports.RecordAccessEventCommand
}

func (s *recordingAuditService) LockChain(context.Context) error { return nil }

func (s *recordingAuditService) RecordEvent(_ context.Context, cmd ports.RecordEventCommand) (*domain.AuditEvent, error) {
	s.events = append(s.events, cmd)
	return &domain.AuditEvent{ID: uuid.New()}, nil
}

func (s *recordingAuditService) GetEvent(context.Context, uuid.UUID) (*domain.AuditEvent, error) {
	return nil, nil
}

func (s *recordingAuditService) SearchEvents(context.Context, domain.AuditFilter) ([]domain.AuditEvent, int64, error) {
	return nil, 0, nil
}

func (s *recordingAuditService) VerifyChainIntegrity(context.Context) (*domain.IntegrityVerificationResult, error) {
	return nil, nil
}

func (s *recordingAuditService) RecordSecurityEvent(context.Context, ports.RecordSecurityEventCommand) (*domain.SecurityEvent, error) {
	return nil, nil
}

func (s *recordingAuditService) ListSecurityEvents(context.Context, *uuid.UUID, string, string, int32, int32) ([]domain.SecurityEvent, error) {
	return nil, nil
}

func (s *recordingAuditService) RecordAccessEvent(_ context.Context, cmd ports.RecordAccessEventCommand) (*domain.AccessEvent, error) {
	s.accesses = append(s.accesses, cmd)
	return &domain.AccessEvent{ID: uuid.New()}, nil
}

func (s *recordingAuditService) ListAccessEvents(context.Context, *uuid.UUID, string, int32, int32) ([]domain.AccessEvent, error) {
	return nil, nil
}

func TestAuditRequestMiddlewareRecordsSuccessfulMutationAndAccess(t *testing.T) {
	service := &recordingAuditService{}
	userID := uuid.New()
	legalEntityID := uuid.New()
	handler := AuditRequestMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/products/categories?legal_entity_id="+legalEntityID.String(),
		nil,
	)
	req.RemoteAddr = "127.0.0.1:43120"
	req = req.WithContext(iamhttp.ContextWithUser(req.Context(), &iamdomain.User{ID: userID}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
	if len(service.events) != 1 {
		t.Fatalf("expected one mutation event, got %d", len(service.events))
	}
	event := service.events[0]
	if event.UserID == nil || *event.UserID != userID {
		t.Fatalf("expected audit user %s, got %v", userID, event.UserID)
	}
	if event.LegalEntityID == nil || *event.LegalEntityID != legalEntityID {
		t.Fatalf("expected legal entity %s, got %v", legalEntityID, event.LegalEntityID)
	}
	if event.Action != "CREATE" || event.ResourceType != "PRODUCTS_CATEGORIES" {
		t.Fatalf("unexpected mutation classification: %s %s", event.Action, event.ResourceType)
	}
	if len(service.accesses) != 1 || service.accesses[0].StatusCode != http.StatusCreated {
		t.Fatalf("expected one 201 access event, got %+v", service.accesses)
	}
	if service.accesses[0].IPAddress != "127.0.0.1" {
		t.Fatalf("expected normalized client IP, got %q", service.accesses[0].IPAddress)
	}
}

func TestAuditRequestMiddlewareDoesNotRecordReadOrFailedMutationAsBusinessEvent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		status int
	}{
		{name: "read", method: http.MethodGet, status: http.StatusOK},
		{name: "failed mutation", method: http.MethodPatch, status: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &recordingAuditService{}
			handler := AuditRequestMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(tc.method, "/api/v1/products/items", nil))

			if len(service.events) != 0 {
				t.Fatalf("expected no business event, got %d", len(service.events))
			}
			if len(service.accesses) != 1 {
				t.Fatalf("expected one access event, got %d", len(service.accesses))
			}
		})
	}
}

func TestResourceTypeExcludesRecordIdentifiers(t *testing.T) {
	recordID := uuid.New()
	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/api/v1/documents/" + recordID.String(), want: "DOCUMENTS"},
		{path: "/api/v1/notifications/" + recordID.String(), want: "NOTIFICATIONS"},
		{path: "/api/v1/products/categories", want: "PRODUCTS_CATEGORIES"},
		{path: "/api/v1/sales/invoices/" + recordID.String(), want: "SALES_INVOICES"},
		{path: "/api/v1/reports/42", want: "REPORTS"},
		{path: "/api/v1/", want: "API_RESOURCE"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			if got := resourceType(tc.path); got != tc.want {
				t.Fatalf("resourceType(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestAuditRequestMiddlewareRecordsDeleteWithoutIdentifierInResourceType(t *testing.T) {
	service := &recordingAuditService{}
	recordID := uuid.New()
	handler := AuditRequestMiddleware(service)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/"+recordID.String(), nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if len(service.events) != 1 {
		t.Fatalf("expected one mutation event, got %d", len(service.events))
	}
	event := service.events[0]
	if event.Action != "DELETE" || event.ResourceType != "DOCUMENTS" {
		t.Fatalf("unexpected classification: %s %s", event.Action, event.ResourceType)
	}
	if event.ResourceID != recordID.String() {
		t.Fatalf("expected resource id %s, got %s", recordID, event.ResourceID)
	}
}
