package services

import (
	"context"
	"fmt"
	"strings"

	"erp/internal/authorization/domain"
	"erp/internal/authorization/ports"
	"github.com/google/uuid"
)

type AuthorizationService struct {
	roleRepo       ports.RoleRepository
	assignmentRepo ports.AssignmentRepository
	cache          ports.PolicyCache
}

func NewAuthorizationService(
	roleRepo ports.RoleRepository,
	assignmentRepo ports.AssignmentRepository,
	cache ports.PolicyCache,
) ports.AuthorizationService {
	return &AuthorizationService{
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
		cache:          cache,
	}
}

// Authorize implements the canonical Section 34 contract with fast caching and legal entity scope enforcement.
func (s *AuthorizationService) Authorize(ctx context.Context, request domain.AuthorizationRequest) (domain.AuthorizationDecision, error) {
	if request.UserID == uuid.Nil {
		return domain.AuthorizationDecision{
			Allowed: false,
			Reason:  "missing user id in authorization request",
		}, nil
	}

	reqPerm := strings.ToLower(strings.TrimSpace(request.Permission))
	if reqPerm == "" {
		return domain.AuthorizationDecision{
			Allowed: false,
			Reason:  "empty permission requested",
		}, nil
	}

	// 1. Try cache lookup
	permissions, found := s.cache.Get(ctx, request.UserID, request.LegalEntityID)
	if !found {
		// 2. Query database for permissions in this user + legal entity scope
		var err error
		permissions, err = s.assignmentRepo.GetPermissionsInScope(ctx, request.UserID, request.LegalEntityID)
		if err != nil {
			return domain.AuthorizationDecision{
				Allowed: false,
				Reason:  fmt.Sprintf("failed to evaluate permissions: %v", err),
			}, err
		}

		// 3. Populate cache
		s.cache.Set(ctx, request.UserID, request.LegalEntityID, permissions)
	}

	// 4. Evaluate permissions
	for _, p := range permissions {
		pLower := strings.ToLower(p)
		if pLower == "*" || pLower == reqPerm {
			return domain.AuthorizationDecision{
				Allowed: true,
				Reason:  "permission granted",
			}, nil
		}
		// Wildcard module matching, e.g. "org.*" matches "org.branches.view"
		if strings.HasSuffix(pLower, ".*") {
			prefix := strings.TrimSuffix(pLower, ".*")
			if strings.HasPrefix(reqPerm, prefix+".") {
				return domain.AuthorizationDecision{
					Allowed: true,
					Reason:  "wildcard permission granted",
				}, nil
			}
		}
	}

	return domain.AuthorizationDecision{
		Allowed: false,
		Reason:  fmt.Sprintf("user lacks permission %q in legal entity scope %v", request.Permission, request.LegalEntityID),
	}, nil
}

func (s *AuthorizationService) AssignRole(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) (*domain.RoleAssignment, error) {
	assignment, err := s.assignmentRepo.Assign(ctx, userID, roleID, legalEntityID, branchID)
	if err != nil {
		return nil, err
	}

	// Invalidate user permissions in policy cache
	s.cache.InvalidateUser(ctx, userID)

	return assignment, nil
}

func (s *AuthorizationService) RevokeRole(ctx context.Context, userID, roleID uuid.UUID, legalEntityID, branchID *uuid.UUID) error {
	err := s.assignmentRepo.Revoke(ctx, userID, roleID, legalEntityID, branchID)
	if err != nil {
		return err
	}

	// Invalidate user permissions in policy cache
	s.cache.InvalidateUser(ctx, userID)

	return nil
}

func (s *AuthorizationService) GetUserPermissions(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID) ([]string, error) {
	// Try cache first
	if perms, found := s.cache.Get(ctx, userID, legalEntityID); found {
		return perms, nil
	}

	perms, err := s.assignmentRepo.GetPermissionsInScope(ctx, userID, legalEntityID)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, userID, legalEntityID, perms)
	return perms, nil
}

func (s *AuthorizationService) ListUserAssignments(ctx context.Context, userID uuid.UUID) ([]*domain.RoleAssignment, error) {
	return s.assignmentRepo.ListByUser(ctx, userID)
}

func (s *AuthorizationService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roleRepo.List(ctx)
}

func (s *AuthorizationService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.roleRepo.ListPermissions(ctx)
}
