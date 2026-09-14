package ports

import (
	"context"
	"time"

	"erp/internal/iam/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByIdentifier(ctx context.Context, identifier string) (*domain.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type PasswordRepository interface {
	SetPassword(ctx context.Context, userID uuid.UUID, hash string) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordCredential, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) (*domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, *domain.User, error)
	UpdateLastActive(ctx context.Context, tokenHash string) error
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

type LoginAttemptRepository interface {
	Record(ctx context.Context, identifier, ip string, success bool) error
	CountRecentFailed(ctx context.Context, identifier string, since time.Time) (int64, error)
}
