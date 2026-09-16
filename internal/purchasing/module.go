package purchasing

import (
	"context"
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/purchasing/application/services"
	"erp/internal/purchasing/infrastructure/postgres"
	"erp/internal/purchasing/ports"
	purchasinghttp "erp/internal/purchasing/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool              *pgxpool.Pool
	PurchasingService ports.PurchasingService
	Handler           *purchasinghttp.PurchasingHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewPurchasingRepository(db)
	svc := services.NewPurchasingService(repo, db)
	handler := purchasinghttp.NewPurchasingHandler(svc)

	return &Module{
		pool:              db.Pool,
		PurchasingService: svc,
		Handler:           handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/purchasing", m.Handler.Routes(authMiddleware))
}

// SeedDefaultPurchasingData initializes default PR, PO and GRN for a legal entity.
func (m *Module) SeedDefaultPurchasingData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT purchasing.seed_default_purchasing_data($1)", legalEntityID)
	return err
}
