package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	iamdomain "erp/internal/iam/domain"
	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/sales/domain"
	"erp/internal/sales/ports"

	"github.com/google/uuid"
)

type actorRecordingSalesService struct {
	ports.SalesService
	actorID uuid.UUID
}

func (s *actorRecordingSalesService) PostInvoice(_ context.Context, id, actorID uuid.UUID) (*domain.SalesInvoice, error) {
	s.actorID = actorID
	return &domain.SalesInvoice{ID: id}, nil
}

func (s *actorRecordingSalesService) PostCreditNote(_ context.Context, id, actorID uuid.UUID) (*domain.CreditNote, error) {
	s.actorID = actorID
	return &domain.CreditNote{ID: id}, nil
}

func TestSalesPostingUsesAuthenticatedActor(t *testing.T) {
	for _, resource := range []string{"invoices", "credit-notes"} {
		t.Run(resource, func(t *testing.T) {
			service := &actorRecordingSalesService{}
			handler := NewSalesHandler(service).Routes(nil)
			path := "/" + resource + "/" + uuid.NewString() + "/post"
			request := httptest.NewRequest(http.MethodPost, path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized || service.actorID != uuid.Nil {
				t.Fatal("missing authentication reached posting service")
			}
			actorID := uuid.New()
			request = request.WithContext(iamhttp.ContextWithUser(request.Context(), &iamdomain.User{ID: actorID}))
			response = httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusOK || service.actorID != actorID {
				t.Fatalf("posting actor was not propagated: status=%d actor=%s", response.Code, service.actorID)
			}
		})
	}
}
