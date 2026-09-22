package audit

import (
	"net/http"

	"erp/internal/audit/application/services"
	"erp/internal/audit/infrastructure/postgres"
	"erp/internal/audit/ports"
	audithttp "erp/internal/audit/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service ports.AuditService
	Handler *audithttp.AuditHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewAuditRepository(db.Pool)
	service := services.NewAuditService(repo)
	handler := audithttp.NewAuditHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/audit", m.Handler.Routes(authMiddleware))
}

func (m *Module) AccessMiddleware() func(http.Handler) http.Handler {
	return audithttp.AuditRequestMiddleware(m.Service)
}
