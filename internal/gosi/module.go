package gosi

import (
	"context"
	"net/http"

	"erp/internal/gosi/application/services"
	"erp/internal/gosi/infrastructure/postgres"
	"erp/internal/gosi/ports"
	gosihttp "erp/internal/gosi/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool        *pgxpool.Pool
	GosiService ports.Service
	Adapter     services.GOSIAdapter
	Repo        ports.Repository
	Handler     *gosihttp.GosiHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewRepository(db)
	svc := services.NewGosiService(repo, db)
	adapter := services.NewDefaultGOSIAdapter(repo)
	handler := gosihttp.NewGosiHandler(svc)

	return &Module{
		pool:        db.Pool,
		GosiService: svc,
		Adapter:     adapter,
		Repo:        repo,
		Handler:     handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/gosi", m.Handler.Routes(authMiddleware))
}

// SeedDefaultGOSIData initializes standard contribution rules and versioned rates for Annuity, OHRSI, and SANED.
func (m *Module) SeedDefaultGOSIData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL gosi.seed_default_gosi_data($1)", legalEntityID)
	return err
}
