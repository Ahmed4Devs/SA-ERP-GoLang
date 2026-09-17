package reporting

import (
	"context"
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/reporting/application"
	"erp/internal/reporting/infrastructure/postgres"
	"erp/internal/reporting/ports"
	rhttp "erp/internal/reporting/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool    *pgxpool.Pool
	Service ports.ReportingService
	Repo    ports.ReportingRepository
	Handler *rhttp.ReportingHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewReportingRepository(db)
	svc := application.NewReportingService(repo)
	handler := rhttp.NewReportingHandler(svc)

	return &Module{
		pool:    db.Pool,
		Service: svc,
		Repo:    repo,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/reports", m.Handler.Routes(authMiddleware))
}

// SeedDefaultReportingData executes the database procedure for reporting configuration verification.
func (m *Module) SeedDefaultReportingData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL reporting.seed_default_reporting_data($1)", legalEntityID)
	return err
}
