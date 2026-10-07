package sales

import (
	"context"
	"net/http"

	"erp/internal/accounting"
	"erp/internal/audit"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/sales/application/services"
	"erp/internal/sales/infrastructure/postgres"
	"erp/internal/sales/ports"
	saleshttp "erp/internal/sales/transport/http"
	taxpostgres "erp/internal/tax/infrastructure/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool         *pgxpool.Pool
	SalesService ports.SalesService
	Handler      *saleshttp.SalesHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewSalesRepository(db)
	svc := services.NewSalesService(repo, db, accounting.NewModule(cfg, db).PostingEngine, taxpostgres.NewTaxRepository(db.Pool), audit.NewModule(cfg, db).Service)
	handler := saleshttp.NewSalesHandler(svc)

	return &Module{
		pool:         db.Pool,
		SalesService: svc,
		Handler:      handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/sales", m.Handler.Routes(authMiddleware))
}

func (m *Module) SeedDefaultSalesData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT sales.seed_default_sales_data($1)", legalEntityID)
	return err
}
