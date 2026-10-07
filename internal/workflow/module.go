package workflow

import (
	"erp/internal/audit"
	"erp/internal/authorization"
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
	Service         ports.WorkflowService
	Handler         *workflowhttp.WorkflowHandler
	LimitHandler    *workflowhttp.ApprovalLimitHandler
	limitPermission func(http.Handler) http.Handler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewWorkflowRepository(db)
	service := services.NewWorkflowService(repo, db, audit.NewModule(cfg, db).Service)
	handler := workflowhttp.NewWorkflowHandler(service)

	return &Module{
		Service:         service,
		Handler:         handler,
		LimitHandler:    workflowhttp.NewApprovalLimitHandler(db, audit.NewModule(cfg, db).Service),
		limitPermission: authorization.NewModule(cfg, db).RequirePermission("workflow.definitions.manage", nil),
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	routes := m.Handler.Routes(authMiddleware)
	routes.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}
		r.Use(m.limitPermission)
		r.Get("/approval-limits", m.LimitHandler.List)
		r.Put("/approval-limits", m.LimitHandler.Save)
	})
	r.Mount("/api/v1/workflow", routes)
}
