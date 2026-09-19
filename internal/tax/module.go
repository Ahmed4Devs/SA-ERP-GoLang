package tax

import (
	"context"
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/tax/application/services"
	"erp/internal/tax/infrastructure/postgres"
	"erp/internal/tax/ports"
	taxhttp "erp/internal/tax/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool       *pgxpool.Pool
	TaxService ports.TaxService
	Handler    *taxhttp.TaxHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewTaxRepository(db.Pool)
	taxSvc := services.NewTaxService(repo, db)
	handler := taxhttp.NewTaxHandler(taxSvc)

	return &Module{
		pool:       db.Pool,
		TaxService: taxSvc,
		Handler:    handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/tax", m.Handler.Routes(authMiddleware))
}

// SeedDefaultTaxData initializes the Saudi standard VAT codes, rates (historical 5% & current 15%),
// and 2026 reporting periods for a legal entity.
func (m *Module) SeedDefaultTaxData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT tax.seed_default_tax_data($1)", legalEntityID)
	return err
}
