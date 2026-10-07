package postgres

import (
	"context"
	"errors"
	"fmt"

	"erp/internal/iam/domain"
	"erp/internal/iam/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepository struct {
	db *database.DB
}

func NewSessionRepository(db *database.DB) ports.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) (*domain.Session, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:    pgtype.UUID{Bytes: s.UserID, Valid: true},
		TokenHash: s.TokenHash,
		IpAddress: pgtype.Text{String: s.IPAddress, Valid: s.IPAddress != ""},
		UserAgent: pgtype.Text{String: s.UserAgent, Valid: s.UserAgent != ""},
		ExpiresAt: pgtype.Timestamptz{Time: s.ExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.Session{
		ID:           row.ID.Bytes,
		UserID:       row.UserID.Bytes,
		TokenHash:    row.TokenHash,
		IPAddress:    row.IpAddress.String,
		UserAgent:    row.UserAgent.String,
		ExpiresAt:    row.ExpiresAt.Time,
		CreatedAt:    row.CreatedAt.Time,
		LastActiveAt: row.LastActiveAt.Time,
	}, nil
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, *domain.User, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrSessionNotFound
		}
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}

	sess := &domain.Session{
		ID:           row.ID.Bytes,
		UserID:       row.UserID.Bytes,
		TokenHash:    row.TokenHash,
		IPAddress:    row.IpAddress.String,
		UserAgent:    row.UserAgent.String,
		ExpiresAt:    row.ExpiresAt.Time,
		CreatedAt:    row.CreatedAt.Time,
		LastActiveAt: row.LastActiveAt.Time,
	}

	u := &domain.User{
		ID:                row.UserID.Bytes,
		DisplayName:       row.DisplayName,
		Status:            domain.UserStatus(row.Status),
		PreferredLocale:   row.PreferredLocale,
		PreferredTimezone: row.PreferredTimezone,
	}
	if row.Username.Valid {
		u.Username = &row.Username.String
	}
	if row.Email.Valid {
		u.Email = &row.Email.String
	}

	return sess, u, nil
}

func (r *SessionRepository) UpdateLastActive(ctx context.Context, tokenHash string) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	return queries.UpdateSessionLastActive(ctx, tokenHash)
}

func (r *SessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	return queries.DeleteSessionByTokenHash(ctx, tokenHash)
}

func (r *SessionRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	if _, err := r.db.DBTX(ctx).Exec(ctx, "DELETE FROM iam.mfa_login_challenges WHERE user_id = $1", userID); err != nil {
		return err
	}
	return queries.DeleteAllSessionsForUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
}
