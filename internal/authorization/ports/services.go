package ports

import (
	"context"

	"erp/internal/authorization/domain"
	"github.com/google/uuid"
)

// AuthorizationService represents the canonical authorization evaluation and management service per Section 34.
type AuthorizationService interface {
	Authorize(ctx context.Context, request domain.AuthorizationRequest) (domain.AuthorizationDecision, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) (*domain.RoleAssignment, error)
	RevokeRole(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID) ([]string, error)
	ListUserAssignments(ctx context.Context, userID uuid.UUID) ([]*domain.RoleAssignment, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	ListPermissions(ctx context.Context) ([]*domain.Permission, error)
}
