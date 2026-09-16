package receivables

import (
	"context"
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/receivables/application/services"
	"erp/internal/receivables/infrastructure/postgres"
	"erp/internal/receivables/ports"
	receivableshttp "erp/internal/receivables/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool               *pgxpool.Pool
	CustomerService    ports.CustomerService
	ReceivablesService ports.ReceivablesService
	Handler            *receivableshttp.CustomerHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	customerRepo := postgres.NewCustomerRepository(db.Pool)
	receivablesRepo := postgres.NewReceivablesRepository(db.Pool)

	customerSvc := services.NewCustomerService(customerRepo)
	receivablesSvc := services.NewReceivablesService(customerRepo, receivablesRepo)

	handler := receivableshttp.NewCustomerHandler(customerSvc, receivablesSvc)

	return &Module{
		pool:               db.Pool,
		CustomerService:    customerSvc,
		ReceivablesService: receivablesSvc,
		Handler:            handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/receivables", m.Handler.Routes(authMiddleware))
}

// SeedDefaultReceivablesData initializes default payment terms and sample customers for a legal entity.
func (m *Module) SeedDefaultReceivablesData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT receivables.seed_default_receivables_data($1)", legalEntityID)
	return err
}
