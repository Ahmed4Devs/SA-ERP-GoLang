package workflow

import (
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/workflow/application/services"
	"erp/internal/workflow/infrastructure/postgres"
	"erp/internal/workflow/ports"
	workflowhttp "erp/internal/workflow/transport/http"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	Service ports.WorkflowService
	Handler *workflowhttp.WorkflowHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewWorkflowRepository(db)
	service := services.NewWorkflowService(repo, db)
	handler := workflowhttp.NewWorkflowHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/workflow", m.Handler.Routes(authMiddleware))
}
