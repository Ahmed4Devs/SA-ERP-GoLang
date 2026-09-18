package products

import (
	"context"
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/products/application/services"
	"erp/internal/products/infrastructure/postgres"
	"erp/internal/products/ports"
	producthttp "erp/internal/products/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool           *pgxpool.Pool
	ProductService ports.ProductService
	Handler        *producthttp.ProductHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewProductRepository(db)
	productSvc := services.NewProductService(repo, db)
	handler := producthttp.NewProductHandler(productSvc)

	return &Module{
		pool:           db.Pool,
		ProductService: productSvc,
		Handler:        handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/products", m.Handler.Routes(authMiddleware))
}

// SeedDefaultProductsData initializes sample categories, units of measure,
// and default products for a legal entity.
func (m *Module) SeedDefaultProductsData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT products.seed_default_products_data($1)", legalEntityID)
	return err
}
