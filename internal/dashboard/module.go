package dashboard

import (
	"net/http"

	"erp/internal/dashboard/application/services"
	"erp/internal/dashboard/infrastructure/postgres"
	"erp/internal/dashboard/ports"
	dhttp "erp/internal/dashboard/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool    *pgxpool.Pool
	Service ports.DashboardService
	Repo    ports.DashboardRepository
	Handler *dhttp.DashboardHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewDashboardRepository(db)
	svc := services.NewDashboardService(repo)
	handler := dhttp.NewDashboardHandler(svc)

	return &Module{
		pool:    db.Pool,
		Service: svc,
		Repo:    repo,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/dashboard", m.Handler.Routes(authMiddleware))
}
