package fixedassets

import (
	"context"
	"net/http"

	"erp/internal/fixedassets/application/services"
	"erp/internal/fixedassets/infrastructure/postgres"
	"erp/internal/fixedassets/ports"
	fahttp "erp/internal/fixedassets/transport/http"
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
	Handler *fahttp.FixedAssetsHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewRepository(db)
	svc := services.NewFixedAssetsService(repo, db)
	handler := fahttp.NewFixedAssetsHandler(svc)

	return &Module{
		pool:    db.Pool,
		Service: svc,
		Repo:    repo,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/fixedassets", m.Handler.Routes(authMiddleware))
}

// SeedDefaultFixedAssetsData initializes standard asset classes, locations, books, and sample assets.
func (m *Module) SeedDefaultFixedAssetsData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL fixedassets.seed_default_fixed_assets_data($1)", legalEntityID)
	return err
}
