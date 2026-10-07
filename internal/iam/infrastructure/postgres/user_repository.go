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

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) ports.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username:          toPgText(u.Username),
		Email:             toPgText(u.Email),
		PhoneNumber:       toPgText(u.PhoneNumber),
		DisplayName:       u.DisplayName,
		PreferredLocale:   u.PreferredLocale,
		PreferredTimezone: u.PreferredTimezone,
		Status:            string(u.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return toDomainUser(row.ID, row.Username, row.Email, row.PhoneNumber, row.DisplayName, row.PreferredLocale, row.PreferredTimezone, row.Status, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.DisabledAt, row.Version), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetUserByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return toDomainUser(row.ID, row.Username, row.Email, row.PhoneNumber, row.DisplayName, row.PreferredLocale, row.PreferredTimezone, row.Status, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.DisabledAt, row.Version), nil
}

func (r *UserRepository) GetByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetUserByUsernameOrEmail(ctx, identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by identifier: %w", err)
	}

	return toDomainUser(row.ID, row.Username, row.Email, row.PhoneNumber, row.DisplayName, row.PreferredLocale, row.PreferredTimezone, row.Status, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.DisabledAt, row.Version), nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	return queries.UpdateUserLastLogin(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

func (r *UserRepository) InvalidateAuthentication(ctx context.Context, id uuid.UUID) error {
	return sqlc.New(r.db.DBTX(ctx)).InvalidateUserAuthentication(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toDomainUser(id pgtype.UUID, username, email, phone pgtype.Text, displayName, locale, tz, status string, lastLogin, created, updated, disabled pgtype.Timestamptz, version int64) *domain.User {
	u := &domain.User{
		ID:                id.Bytes,
		DisplayName:       displayName,
		PreferredLocale:   locale,
		PreferredTimezone: tz,
		Status:            domain.UserStatus(status),
		CreatedAt:         created.Time,
		UpdatedAt:         updated.Time,
		Version:           version,
	}
	if username.Valid {
		u.Username = &username.String
	}
	if email.Valid {
		u.Email = &email.String
	}
	if phone.Valid {
		u.PhoneNumber = &phone.String
	}
	if lastLogin.Valid {
		u.LastLoginAt = &lastLogin.Time
	}
	if disabled.Valid {
		u.DisabledAt = &disabled.Time
	}
	return u
}
