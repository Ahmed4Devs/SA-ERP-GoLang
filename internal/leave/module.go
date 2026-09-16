package leave

import (
	"context"
	"net/http"

	"erp/internal/leave/application/services"
	"erp/internal/leave/infrastructure/postgres"
	"erp/internal/leave/ports"
	leavehttp "erp/internal/leave/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool         *pgxpool.Pool
	LeaveService ports.LeaveService
	Repo         ports.LeaveRepository
	Handler      *leavehttp.LeaveHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewLeaveRepository(db)
	svc := services.NewLeaveService(repo, db)
	handler := leavehttp.NewLeaveHandler(svc)

	return &Module{
		pool:         db.Pool,
		LeaveService: svc,
		Repo:         repo,
		Handler:      handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/leave", m.Handler.Routes(authMiddleware))
}

// SeedDefaultLeaveData initializes Saudi statutory leave types, default policy, and balances.
func (m *Module) SeedDefaultLeaveData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT leave.seed_default_leave_data($1)", legalEntityID)
	return err
}
