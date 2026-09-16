package attendance

import (
	"context"
	"net/http"

	"erp/internal/attendance/application/services"
	"erp/internal/attendance/infrastructure/postgres"
	"erp/internal/attendance/ports"
	attendancehttp "erp/internal/attendance/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	pool              *pgxpool.Pool
	AttendanceService ports.AttendanceService
	Repo              ports.AttendanceRepository
	Handler           *attendancehttp.AttendanceHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	repo := postgres.NewAttendanceRepository(db)
	svc := services.NewAttendanceService(repo, db)
	handler := attendancehttp.NewAttendanceHandler(svc)

	return &Module{
		pool:              db.Pool,
		AttendanceService: svc,
		Repo:              repo,
		Handler:           handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/attendance", m.Handler.Routes(authMiddleware))
}

// SeedDefaultAttendanceData initializes standard shifts and employee assignments.
func (m *Module) SeedDefaultAttendanceData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "SELECT attendance.seed_default_attendance_data($1)", legalEntityID)
	return err
}
