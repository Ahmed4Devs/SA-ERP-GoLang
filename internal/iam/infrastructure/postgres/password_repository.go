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

type PasswordRepository struct {
	db *database.DB
}

func NewPasswordRepository(db *database.DB) ports.PasswordRepository {
	return &PasswordRepository{db: db}
}

func (r *PasswordRepository) SetPassword(ctx context.Context, userID uuid.UUID, hash string) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	return queries.SetPassword(ctx, sqlc.SetPasswordParams{
		UserID:       pgtype.UUID{Bytes: userID, Valid: true},
		PasswordHash: hash,
	})
}

func (r *PasswordRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordCredential, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetPasswordByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get password credentials: %w", err)
	}

	return &domain.PasswordCredential{
		UserID:       row.UserID.Bytes,
		PasswordHash: row.PasswordHash,
		Algorithm:    row.Algorithm,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}
