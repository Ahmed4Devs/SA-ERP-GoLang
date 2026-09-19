package ports

import (
	"context"
	"time"

	"erp/internal/notifications/domain"
	"github.com/google/uuid"
)

// Repository defines data access contracts for notifications and templates.
type Repository interface {
	Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error)
	GetByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Notification, error)
	List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int, error)
	CountUnread(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, legalEntityID, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) error
	UpdateStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.NotificationStatus, sentAt *time.Time, errMsg *string) error
	Delete(ctx context.Context, legalEntityID, id uuid.UUID) error
	GetSummary(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (*domain.NotificationSummary, error)
	GetTemplateByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error)
}

// Service defines orchestration workflows for notification dispatching and feed management.
type Service interface {
	Send(ctx context.Context, params domain.SendNotificationParams) (*domain.Notification, error)
	Get(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Notification, error)
	List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int, error)
	GetUnreadCount(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, legalEntityID, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) error
	Delete(ctx context.Context, legalEntityID, id uuid.UUID) error
	GetSummary(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (*domain.NotificationSummary, error)
}
