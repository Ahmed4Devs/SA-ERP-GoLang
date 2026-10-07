package treasury

import (
	"context"
	"net/http"

	"erp/internal/accounting"
	"erp/internal/accounting/application/posting"
	"erp/internal/audit"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/treasury/application/services"
	"erp/internal/treasury/infrastructure/postgres"
	"erp/internal/treasury/ports"
	treasuryhttp "erp/internal/treasury/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool            *pgxpool.Pool
	TreasuryService ports.TreasuryService
	Handler         *treasuryhttp.TreasuryHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewTreasuryRepository(db)
	auditService := audit.NewModule(cfg, db).Service
	ledger := posting.NewSourceService(db, accounting.NewModule(cfg, db).PostingEngine, auditService)
	svc := services.NewTreasuryService(repo, ledger, auditService)
	handler := treasuryhttp.NewTreasuryHandler(svc)

	return &Module{
		pool:            db.Pool,
		TreasuryService: svc,
		Handler:         handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/treasury", m.Handler.Routes(authMiddleware))
}

// SeedDefaultTreasuryData initializes default cash accounts, bank accounts and sample vouchers for a legal entity.
func (m *Module) SeedDefaultTreasuryData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT treasury.seed_default_treasury_data($1)", legalEntityID)
	return err
}
