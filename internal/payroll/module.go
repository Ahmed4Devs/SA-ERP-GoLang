package payroll

import (
	"context"
	"net/http"

	"erp/internal/payroll/application/services"
	"erp/internal/payroll/infrastructure/postgres"
	"erp/internal/payroll/ports"
	payrollhttp "erp/internal/payroll/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool           *pgxpool.Pool
	PayrollService ports.Service
	Repo           ports.Repository
	Handler        *payrollhttp.PayrollHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewRepository(db)
	svc := services.NewPayrollService(repo, db)
	handler := payrollhttp.NewPayrollHandler(svc)

	return &Module{
		pool:           db.Pool,
		PayrollService: svc,
		Repo:           repo,
		Handler:        handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/payroll", m.Handler.Routes(authMiddleware))
}

// SeedDefaultPayrollData initializes standard salary components, structures, and sample period.
func (m *Module) SeedDefaultPayrollData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL payroll.seed_default_payroll_data($1)", legalEntityID)
	return err
}
