package postgres

import (
	"context"
	"errors"
	"fmt"

	"erp/internal/authorization/domain"
	"erp/internal/authorization/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RoleRepository struct {
	db *database.DB
}

func NewRoleRepository(db *database.DB) ports.RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetByCode(ctx context.Context, code string) (*domain.Role, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetRoleByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by code: %w", err)
	}

	return toDomainRole(row), nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetRoleByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	return toDomainRole(row), nil
}

func (r *RoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	res := make([]*domain.Role, len(rows))
	for i, row := range rows {
		res[i] = toDomainRole(row)
	}
	return res, nil
}

func (r *RoleRepository) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	res := make([]*domain.Permission, len(rows))
	for i, row := range rows {
		res[i] = &domain.Permission{
			ID:        toUUID(row.ID),
			Code:      row.Code,
			NameAr:    row.NameAr,
			NameEn:    row.NameEn,
			Module:    row.Module,
			CreatedAt: row.CreatedAt.Time,
		}
	}
	return res, nil
}

func (r *RoleRepository) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListPermissionsForRole(ctx, toPgUUID(roleID))
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions for role: %w", err)
	}

	res := make([]*domain.Permission, len(rows))
	for i, row := range rows {
		res[i] = &domain.Permission{
			ID:        toUUID(row.ID),
			Code:      row.Code,
			NameAr:    row.NameAr,
			NameEn:    row.NameEn,
			Module:    row.Module,
			CreatedAt: row.CreatedAt.Time,
		}
	}
	return res, nil
}

func (r *RoleRepository) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	queries := sqlc.New(r.db.DBTX(ctx))

	return queries.AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
		RoleID:       toPgUUID(roleID),
		PermissionID: toPgUUID(permissionID),
	})
}

func toDomainRole(row sqlc.AuthzRoles) *domain.Role {
	return &domain.Role{
		ID:          toUUID(row.ID),
		Code:        row.Code,
		NameAr:      row.NameAr,
		NameEn:      row.NameEn,
		Description: fromPgText(row.Description),
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
