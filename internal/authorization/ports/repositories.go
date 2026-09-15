package ports

import (
	"context"

	"erp/internal/authorization/domain"
	"github.com/google/uuid"
)

type RoleRepository interface {
	GetByCode(ctx context.Context, code string) (*domain.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	List(ctx context.Context) ([]*domain.Role, error)
	ListPermissions(ctx context.Context) ([]*domain.Permission, error)
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error)
	AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error
}

type AssignmentRepository interface {
	Assign(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) (*domain.RoleAssignment, error)
	Revoke(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.RoleAssignment, error)
	GetPermissionsInScope(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID) ([]string, error)
}

type PolicyCache interface {
	Get(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID) ([]string, bool)
	Set(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID, permissions []string)
	InvalidateUser(ctx context.Context, userID uuid.UUID)
	InvalidateAll(ctx context.Context)
}
