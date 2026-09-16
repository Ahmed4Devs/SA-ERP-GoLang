package payables

import (
	"context"
	"net/http"

	"erp/internal/payables/application/services"
	"erp/internal/payables/infrastructure/postgres"
	"erp/internal/payables/ports"
	payableshttp "erp/internal/payables/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool            *pgxpool.Pool
	SupplierService ports.SupplierService
	PayableService  ports.PayableService
	Handler         *payableshttp.PayableHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	supplierRepo := postgres.NewSupplierRepository(db)
	payableRepo := postgres.NewPayableRepository(db)

	supplierSvc := services.NewSupplierService(supplierRepo)
	payableSvc := services.NewPayableService(supplierRepo, payableRepo, db)

	handler := payableshttp.NewPayableHandler(supplierSvc, payableSvc)

	return &Module{
		pool:            db.Pool,
		SupplierService: supplierSvc,
		PayableService:  payableSvc,
		Handler:         handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/payables", m.Handler.Routes(authMiddleware))
}

// SeedDefaultPayablesData initializes default payment terms and sample suppliers for a legal entity.
func (m *Module) SeedDefaultPayablesData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT payables.seed_default_payables_data($1)", legalEntityID)
	return err
}
