package postgres

import (
	"context"
	"encoding/json"
	"erp/internal/iam/domain"
	"erp/internal/iam/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ManagedUserRepository struct{ *UserRepository }

func NewManagedUserRepository(db *database.DB) ports.ManagedUserRepository {
	return &ManagedUserRepository{&UserRepository{db: db}}
}

func managedUser(row sqlc.IamUsers) *domain.User {
	return toDomainUser(row.ID, row.Username, row.Email, row.PhoneNumber, row.DisplayName, row.PreferredLocale, row.PreferredTimezone, row.Status, row.LastLoginAt, row.CreatedAt, row.UpdatedAt, row.DisabledAt, row.Version)
}

func (r *ManagedUserRepository) List(ctx context.Context, search string, limit, offset int32) ([]*domain.User, error) {
	rows, err := sqlc.New(r.db.DBTX(ctx)).ListManagedUsers(ctx, sqlc.ListManagedUsersParams{Search: search, PageLimit: limit, PageOffset: offset})
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, managedUser(row))
	}
	return users, nil
}

func (r *UserRepository) Lock(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := sqlc.New(r.db.DBTX(ctx)).LockManagedUser(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return managedUser(row), nil
}

func (r *ManagedUserRepository) UpdateState(ctx context.Context, id uuid.UUID, status domain.UserStatus, version int64) (*domain.User, error) {
	row, err := sqlc.New(r.db.DBTX(ctx)).UpdateManagedUserState(ctx, sqlc.UpdateManagedUserStateParams{ID: pgtype.UUID{Bytes: id, Valid: true}, Status: string(status), ExpectedVersion: version})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserVersionConflict
	}
	if err != nil {
		return nil, err
	}
	return managedUser(row), nil
}

func (r *ManagedUserRepository) LockAdministration(ctx context.Context) error {
	return sqlc.New(r.db.DBTX(ctx)).LockUserAdministration(ctx)
}

func (r *ManagedUserRepository) CanDeactivate(ctx context.Context, id uuid.UUID) (bool, error) {
	q := sqlc.New(r.db.DBTX(ctx))
	key := pgtype.UUID{Bytes: id, Valid: true}
	admin, err := q.IsGlobalAdministrator(ctx, key)
	if err != nil || !admin {
		return !admin, err
	}
	count, err := q.CountOtherActiveAdministrators(ctx, key)
	return count > 0, err
}

func (r *ManagedUserRepository) RecordChange(ctx context.Context, actorID, targetID uuid.UUID, action string) error {
	details, err := json.Marshal(map[string]string{"target_user_id": targetID.String()})
	if err != nil {
		return err
	}
	_, err = sqlc.New(r.db.DBTX(ctx)).InsertSecurityEvent(ctx, sqlc.InsertSecurityEventParams{UserID: pgtype.UUID{Bytes: actorID, Valid: true}, EventType: action, Status: "SUCCESS", Details: details})
	return err
}
