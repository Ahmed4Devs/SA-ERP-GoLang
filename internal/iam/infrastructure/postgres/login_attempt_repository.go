package postgres

import (
	"context"
	"time"

	"erp/internal/iam/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type LoginAttemptRepository struct {
	db *database.DB
}

func NewLoginAttemptRepository(db *database.DB) ports.LoginAttemptRepository {
	return &LoginAttemptRepository{db: db}
}

func (r *LoginAttemptRepository) Record(ctx context.Context, identifier, ip string, success bool) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	return queries.RecordLoginAttempt(ctx, sqlc.RecordLoginAttemptParams{
		Identifier: identifier,
		IpAddress:  ip,
		Success:    success,
	})
}

func (r *LoginAttemptRepository) CountRecentFailed(ctx context.Context, identifier string, since time.Time) (int64, error) {
	queries := sqlc.New(r.db.DBTX(ctx))
	return queries.CountRecentFailedAttempts(ctx, sqlc.CountRecentFailedAttemptsParams{
		Identifier:  identifier,
		AttemptedAt: pgtype.Timestamptz{Time: since, Valid: true},
	})
}
