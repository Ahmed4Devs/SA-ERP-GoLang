package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/notifications/domain"
	"erp/internal/notifications/ports"
	"erp/internal/platform/mail"
	"github.com/google/uuid"
)

// NotificationService coordinates multi-channel notification dispatch and user feeds.
type NotificationService struct {
	repo   ports.Repository
	mailer mail.Mailer
}

var _ ports.Service = (*NotificationService)(nil)

// NewNotificationService creates a new NotificationService instance.
func NewNotificationService(repo ports.Repository, mailer mail.Mailer) *NotificationService {
	return &NotificationService{
		repo:   repo,
		mailer: mailer,
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
	var errMsg *string

	// If channel includes EMAIL, dispatch through the Mailer
	if channel == domain.ChannelEmail {
		if strings.TrimSpace(params.RecipientEmail) == "" {
			return nil, fmt.Errorf("%w: recipient_email is required for email channel", domain.ErrRecipientRequired)
		}

		emailMsg := mail.EmailMessage{
			To:       []string{params.RecipientEmail},
			Subject:  titleEn,
			BodyText: fmt.Sprintf("%s\n\n%s", bodyAr, bodyEn),
			BodyHTML: fmt.Sprintf(`
				<div dir="rtl" style="font-family: sans-serif; padding: 20px; line-height: 1.6;">
					<h2 style="color: #1e3a8a;">%s</h2>
					<p style="font-size: 16px;">%s</p>
					<hr style="margin: 20px 0; border: 0; border-top: 1px solid #e5e7eb;" />
					<div dir="ltr" style="text-align: left;">
						<h3 style="color: #4b5563;">%s</h3>
						<p style="font-size: 14px; color: #4b5563;">%s</p>
					</div>
				</div>
			`, titleAr, bodyAr, titleEn, bodyEn),
		}

		if err := s.mailer.Send(ctx, emailMsg); err != nil {
			status = domain.StatusFailed
			errStr := err.Error()
			errMsg = &errStr
			sentAt = nil
		}
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
		ErrorMessage:    errMsg,
		SentAt:          sentAt,
		CreatedAt:       now,
	}

	created, err := s.repo.Create(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to record notification: %w", err)
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
