package inventory

import (
	"context"
	"net/http"

	"erp/internal/inventory/application/services"
	"erp/internal/inventory/infrastructure/postgres"
	"erp/internal/inventory/ports"
	inventoryhttp "erp/internal/inventory/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool             *pgxpool.Pool
	InventoryService ports.InventoryService
	Handler          *inventoryhttp.InventoryHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewInventoryRepository(db)
	invSvc := services.NewInventoryService(repo, db)
	handler := inventoryhttp.NewInventoryHandler(invSvc)

	return &Module{
		pool:             db.Pool,
		InventoryService: invSvc,
		Handler:          handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/inventory", m.Handler.Routes(authMiddleware))
}

// SeedDefaultInventoryData initializes the primary central warehouse, default bin locations,
// and initial stock balances for a legal entity.
func (m *Module) SeedDefaultInventoryData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT inventory.seed_default_inventory_data($1)", legalEntityID)
	return err
}
