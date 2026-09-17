package budgeting

import (
	"context"
	"net/http"

	"erp/internal/budgeting/application/services"
	"erp/internal/budgeting/infrastructure/postgres"
	"erp/internal/budgeting/ports"
	bhttp "erp/internal/budgeting/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool    *pgxpool.Pool
	Service ports.Service
	Repo    ports.Repository
	Handler *bhttp.BudgetingHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewRepository(db)
	svc := services.NewBudgetingService(repo, db)
	handler := bhttp.NewBudgetingHandler(svc)

	return &Module{
		pool:    db.Pool,
		Service: svc,
		Repo:    repo,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/budgeting", m.Handler.Routes(authMiddleware))
}

// SeedDefaultBudgetingData initializes standard operational and capex budgets with multi-dimensional line allocations.
func (m *Module) SeedDefaultBudgetingData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL budgeting.seed_default_budgeting_data($1)", legalEntityID)
	return err
}
