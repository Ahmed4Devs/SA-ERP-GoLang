package accounting

import (
	"context"
	"net/http"

	"erp/internal/accounting/application/services"
	"erp/internal/accounting/infrastructure/postgres"
	"erp/internal/accounting/ports"
	accountinghttp "erp/internal/accounting/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool              *pgxpool.Pool
	AccountingService ports.AccountingService
	LedgerService     ports.LedgerService
	Handler           *accountinghttp.AccountingHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewAccountingRepository(db.Pool)
	accountingSvc := services.NewAccountingService(repo)
	ledgerSvc := services.NewLedgerService(repo)
	handler := accountinghttp.NewAccountingHandler(accountingSvc, ledgerSvc)

	return &Module{
		pool:              db.Pool,
		AccountingService: accountingSvc,
		LedgerService:     ledgerSvc,
		Handler:           handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/accounting", m.Handler.Routes(authMiddleware))
}

// SeedDefaultCOA initializes the Saudi standard chart of accounts, fiscal calendar, and journals for a legal entity.
func (m *Module) SeedDefaultCOA(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT accounting.seed_default_coa($1)", legalEntityID)
	return err
}
