package hr

import (
	"context"
	"net/http"

	"erp/internal/hr/application/services"
	"erp/internal/hr/infrastructure/postgres"
	"erp/internal/hr/ports"
	hrhttp "erp/internal/hr/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool      *pgxpool.Pool
	HRService ports.HRService
	Repo      ports.HRRepository
	Handler   *hrhttp.HRHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewHRRepository(db)
	svc := services.NewHRService(repo, db)
	handler := hrhttp.NewHRHandler(svc)

	return &Module{
		pool:      db.Pool,
		HRService: svc,
		Repo:      repo,
		Handler:   handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/hr", m.Handler.Routes(authMiddleware))
}

// SeedDefaultHRData initializes standard jobs, positions, employees, identifiers, and contracts.
func (m *Module) SeedDefaultHRData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT hr.seed_default_hr_data($1)", legalEntityID)
	return err
}
