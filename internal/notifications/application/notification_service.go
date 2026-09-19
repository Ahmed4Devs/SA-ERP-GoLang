package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/notifications/domain"
	"erp/internal/notifications/ports"
	"erp/internal/platform/outbox"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

const EmailRequestedEventType = "notifications.email.requested"

type EmailRequestedPayload struct {
	LegalEntityID  uuid.UUID `json:"legal_entity_id"`
	NotificationID uuid.UUID `json:"notification_id"`
}

// NotificationService coordinates multi-channel notification dispatch and user feeds.
type NotificationService struct {
	repo   ports.Repository
	outbox outbox.Enqueuer
	tx     transaction.Manager
}

var _ ports.Service = (*NotificationService)(nil)

// NewNotificationService creates a new NotificationService instance.
func NewNotificationService(repo ports.Repository, eventOutbox outbox.Enqueuer, tx transaction.Manager) *NotificationService {
	return &NotificationService{
		repo:   repo,
		outbox: eventOutbox,
		tx:     tx,
	}
}

// Send dispatches a notification across configured delivery channels (In-App, Email).
func (s *NotificationService) Send(ctx context.Context, params domain.SendNotificationParams) (*domain.Notification, error) {
	if params.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}

	if (params.RecipientUserID == nil || *params.RecipientUserID == uuid.Nil) && strings.TrimSpace(params.RecipientEmail) == "" {
		return nil, domain.ErrRecipientRequired
	}

	titleAr := strings.TrimSpace(params.TitleAr)
	titleEn := strings.TrimSpace(params.TitleEn)
	if titleAr == "" && titleEn == "" {
		return nil, domain.ErrTitleRequired
	}
	if titleAr == "" {
		titleAr = titleEn
	}
	if titleEn == "" {
		titleEn = titleAr
	}

	bodyAr := strings.TrimSpace(params.BodyAr)
	bodyEn := strings.TrimSpace(params.BodyEn)
	if bodyAr == "" && bodyEn == "" {
		return nil, domain.ErrBodyRequired
	}
	if bodyAr == "" {
		bodyAr = bodyEn
	}
	if bodyEn == "" {
		bodyEn = bodyAr
	}

	channel := params.Channel
	if channel == "" {
		channel = domain.ChannelInApp
	}
	if channel != domain.ChannelInApp && channel != domain.ChannelEmail {
		return nil, domain.ErrInvalidChannel
	}

	category := params.Category
	if category == "" {
		category = domain.CategorySystem
	}

	priority := params.Priority
	if priority == "" {
		priority = domain.PriorityNormal
	}

	now := time.Now().UTC()
	status := domain.StatusSent
	sentAt := &now
	if channel == domain.ChannelEmail {
		if strings.TrimSpace(params.RecipientEmail) == "" {
			return nil, fmt.Errorf("%w: recipient_email is required for email channel", domain.ErrRecipientRequired)
		}
		status = domain.StatusPending
		sentAt = nil
	}

	notification := &domain.Notification{
		ID:              uuid.New(),
		LegalEntityID:   params.LegalEntityID,
		RecipientUserID: params.RecipientUserID,
		RecipientEmail:  params.RecipientEmail,
		Channel:         channel,
		TitleAr:         titleAr,
		TitleEn:         titleEn,
		BodyAr:          bodyAr,
		BodyEn:          bodyEn,
		Category:        category,
		Priority:        priority,
		ActionURL:       params.ActionURL,
		EntityType:      params.EntityType,
		EntityID:        params.EntityID,
		Status:          status,
		SentAt:          sentAt,
		CreatedAt:       now,
	}

	var created *domain.Notification
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.Create(txCtx, notification)
		if err != nil {
			return fmt.Errorf("failed to record notification: %w", err)
		}
		if channel == domain.ChannelEmail {
			aggregateID := created.ID
			if _, err := s.outbox.Enqueue(txCtx, EmailRequestedEventType, "Notification", &aggregateID, EmailRequestedPayload{
				LegalEntityID:  created.LegalEntityID,
				NotificationID: created.ID,
			}); err != nil {
				return fmt.Errorf("failed to enqueue email delivery: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

// Get retrieves a single notification.
func (s *NotificationService) Get(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Notification, error) {
	return s.repo.GetByID(ctx, legalEntityID, id)
}

// List queries notifications matching filter criteria.
func (s *NotificationService) List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int, error) {
	return s.repo.List(ctx, filter)
}

// GetUnreadCount returns the number of unread notifications for a recipient.
func (s *NotificationService) GetUnreadCount(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (int64, error) {
	return s.repo.CountUnread(ctx, legalEntityID, recipientUserID)
}

// MarkAsRead marks a specific notification as read.
func (s *NotificationService) MarkAsRead(ctx context.Context, legalEntityID, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, legalEntityID, id)
}

// MarkAllAsRead marks all notifications for a recipient as read.
func (s *NotificationService) MarkAllAsRead(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, legalEntityID, recipientUserID)
}

// Delete removes a notification.
func (s *NotificationService) Delete(ctx context.Context, legalEntityID, id uuid.UUID) error {
	return s.repo.Delete(ctx, legalEntityID, id)
}

// GetSummary returns notification counts and category breakdown.
func (s *NotificationService) GetSummary(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (*domain.NotificationSummary, error) {
	return s.repo.GetSummary(ctx, legalEntityID, recipientUserID)
}
