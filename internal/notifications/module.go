package notifications

import (
	"context"
	"fmt"
	"net/http"

	"erp/internal/notifications/application"
	"erp/internal/notifications/infrastructure/postgres"
	"erp/internal/notifications/ports"
	nhttp "erp/internal/notifications/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/mail"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module coordinates the notification and messaging domain.
type Module struct {
	pool    *pgxpool.Pool
	Service ports.Service
	Repo    ports.Repository
	Mailer  mail.Mailer
	Handler *nhttp.NotificationHandler
}

// NewModule initializes a new Notifications module.
func NewModule(cfg *config.Config, db *database.DB, mailer mail.Mailer) *Module {
	repo := postgres.NewNotificationRepository(db)
	svc := application.NewNotificationService(repo, mailer)
	handler := nhttp.NewNotificationHandler(svc)

	return &Module{
		pool:    db.Pool,
		Service: svc,
		Repo:    repo,
		Mailer:  mailer,
		Handler: handler,
	}
}

// RegisterRoutes mounts all notification endpoints on the Chi router under /api/v1/notifications.
func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/notifications", m.Handler.Routes(authMiddleware))
}

// SeedDefaultNotificationsData invokes the stored procedure to populate baseline templates and welcome alert.
func (m *Module) SeedDefaultNotificationsData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := m.pool.Exec(ctx, "CALL notifications.seed_default_notifications_data($1)", legalEntityID)
	if err != nil {
		return fmt.Errorf("failed to seed default notifications: %w", err)
	}
	return nil
}
