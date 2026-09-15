package authorization

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"erp/internal/authorization/application/services"
	"erp/internal/authorization/infrastructure/cache"
	"erp/internal/authorization/infrastructure/postgres"
	"erp/internal/authorization/ports"
	authzhttp "erp/internal/authorization/transport/http"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Module struct {
	Service ports.AuthorizationService
	Handler *authzhttp.AuthorizationHandler
}

func NewModule(cfg *config.Config, db *database.DB) *Module {
	roleRepo := postgres.NewRoleRepository(db)
	assignmentRepo := postgres.NewAssignmentRepository(db)
	policyCache := cache.NewMemoryPolicyCache(5 * time.Minute)

	service := services.NewAuthorizationService(roleRepo, assignmentRepo, policyCache)
	handler := authzhttp.NewAuthorizationHandler(service)

	return &Module{
		Service: service,
		Handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Mount("/api/v1/authz", m.Handler.Routes(authMiddleware))
}

func (m *Module) RequirePermission(permission string, extractScope func(*http.Request) *uuid.UUID) func(http.Handler) http.Handler {
	return authzhttp.RequirePermission(m.Service, permission, extractScope)
}

// SeedAdminRole ensures the system administrator has global SUPER_ADMIN role.
func (m *Module) SeedAdminRole(ctx context.Context, adminUserID uuid.UUID) error {
	roles, err := m.Service.ListRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list roles: %w", err)
	}

	var superAdminRoleID uuid.UUID
	for _, r := range roles {
		if r.Code == "SUPER_ADMIN" {
			superAdminRoleID = r.ID
			break
		}
	}

	if superAdminRoleID == uuid.Nil {
		return fmt.Errorf("SUPER_ADMIN role not found")
	}

	_, err = m.Service.AssignRole(ctx, adminUserID, superAdminRoleID, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to assign SUPER_ADMIN to admin user: %w", err)
	}

	return nil
}
