package banking

import (
	"net/http"

	"erp/internal/banking/application/services"
	"erp/internal/banking/infrastructure/postgres"
	"erp/internal/banking/ports"
	bankinghttp "erp/internal/banking/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool           *pgxpool.Pool
	BankingService ports.BankingService
	Handler        *bankinghttp.BankingHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewBankingRepository(db)
	svc := services.NewBankingService(repo, db)
	handler := bankinghttp.NewBankingHandler(svc)

	return &Module{
		pool:           db.Pool,
		BankingService: svc,
		Handler:        handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/banking", m.Handler.Routes(authMiddleware))
}
