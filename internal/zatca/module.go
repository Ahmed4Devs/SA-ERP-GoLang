package zatca

import (
	"context"
	"net/http"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/zatca/application/services"
	"erp/internal/zatca/infrastructure/postgres"
	"erp/internal/zatca/ports"
	"erp/internal/zatca/signing"
	zhttp "erp/internal/zatca/transport/http"
	"erp/internal/zatca/validation"
	"erp/internal/zatca/xml"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool    *pgxpool.Pool
	Service ports.Service
	Repo    ports.Repository
	Handler *zhttp.ZATCAHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewRepository(db)
	xmlGen := xml.NewUBLGenerator()
	validator := validation.NewInvoiceValidator()
	signer := signing.NewInvoiceSigner()
	svc := services.NewZATCAService(repo, xmlGen, validator, signer, db)
	handler := zhttp.NewZATCAHandler(svc)

	return &Module{
		pool:    db.Pool,
		Service: svc,
		Repo:    repo,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/zatca", m.Handler.Routes(authMiddleware))
}

// SeedDefaultZATCAData initializes baseline EGS units, configuration versions, and compliance CSID credentials.
func (m *Module) SeedDefaultZATCAData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL zatca.seed_default_zatca_data($1)", legalEntityID)
	return err
}
