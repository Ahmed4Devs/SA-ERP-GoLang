package ports

import (
	"context"
	"erp/internal/iam/domain"
	"github.com/google/uuid"
)

type ManagedUserRepository interface {
	UserRepository
	List(ctx context.Context, search string, limit, offset int32) ([]*domain.User, error)
	Lock(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateState(ctx context.Context, id uuid.UUID, status domain.UserStatus, version int64) (*domain.User, error)
	LockAdministration(ctx context.Context) error
	CanDeactivate(ctx context.Context, id uuid.UUID) (bool, error)
	RecordChange(ctx context.Context, actorID, targetID uuid.UUID, action string) error
}
