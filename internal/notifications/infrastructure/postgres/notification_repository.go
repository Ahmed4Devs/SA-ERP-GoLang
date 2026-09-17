package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"erp/internal/notifications/domain"
	"erp/internal/notifications/ports"
	"erp/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// NotificationRepository implements ports.Repository using PostgreSQL.
type NotificationRepository struct {
	db *database.DB
}

var _ ports.Repository = (*NotificationRepository)(nil)

// NewNotificationRepository creates a new NotificationRepository instance.
func NewNotificationRepository(db *database.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create inserts a new notification record.
func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	if n.Status == "" {
		n.Status = domain.StatusPending
	}
	if n.Channel == "" {
		n.Channel = domain.ChannelInApp
	}
	if n.Priority == "" {
		n.Priority = domain.PriorityNormal
	}
	if n.Category == "" {
		n.Category = domain.CategorySystem
	}

	query := `
		INSERT INTO notifications.notifications (
			id, legal_entity_id, recipient_user_id, recipient_email, channel,
			title_ar, title_en, body_ar, body_en, category, priority,
			action_url, entity_type, entity_id, status, error_message,
			sent_at, read_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		RETURNING id, created_at
	`

	err := r.db.Pool.QueryRow(ctx, query,
		n.ID,
		n.LegalEntityID,
		n.RecipientUserID,
		n.RecipientEmail,
		string(n.Channel),
		n.TitleAr,
		n.TitleEn,
		n.BodyAr,
		n.BodyEn,
		string(n.Category),
		string(n.Priority),
		n.ActionURL,
		n.EntityType,
		n.EntityID,
		string(n.Status),
		n.ErrorMessage,
		n.SentAt,
		n.ReadAt,
		n.CreatedAt,
	).Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert notification: %w", err)
	}

	return n, nil
}

// GetByID retrieves a single notification by ID.
func (r *NotificationRepository) GetByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT 
			id, legal_entity_id, recipient_user_id, recipient_email, channel,
			title_ar, title_en, body_ar, body_en, category, priority,
			action_url, entity_type, entity_id, status, error_message,
			sent_at, read_at, created_at
		FROM notifications.notifications
		WHERE legal_entity_id = $1 AND id = $2
	`

	var n domain.Notification
	var ch, cat, prio, st string
	var actURL, entType, errMsg *string
	var recEmail *string

	err := r.db.Pool.QueryRow(ctx, query, legalEntityID, id).Scan(
		&n.ID,
		&n.LegalEntityID,
		&n.RecipientUserID,
		&recEmail,
		&ch,
		&n.TitleAr,
		&n.TitleEn,
		&n.BodyAr,
		&n.BodyEn,
		&cat,
		&prio,
		&actURL,
		&entType,
		&n.EntityID,
		&st,
		&errMsg,
		&n.SentAt,
		&n.ReadAt,
		&n.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if recEmail != nil {
		n.RecipientEmail = *recEmail
	}
	n.Channel = domain.NotificationChannel(ch)
	n.Category = domain.NotificationCategory(cat)
	n.Priority = domain.NotificationPriority(prio)
	n.Status = domain.NotificationStatus(st)
	if actURL != nil {
		n.ActionURL = *actURL
	}
	if entType != nil {
		n.EntityType = *entType
	}
	n.ErrorMessage = errMsg

	return &n, nil
}

// List returns filtered notifications.
func (r *NotificationRepository) List(ctx context.Context, filter domain.NotificationFilter) ([]domain.Notification, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("legal_entity_id = $%d", argIdx))
	args = append(args, filter.LegalEntityID)
	argIdx++

	if filter.RecipientUserID != nil && *filter.RecipientUserID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("(recipient_user_id = $%d OR recipient_user_id IS NULL)", argIdx))
		args = append(args, *filter.RecipientUserID)
		argIdx++
	}

	if filter.Category != nil && *filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, string(*filter.Category))
		argIdx++
	}

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.Channel != nil && *filter.Channel != "" {
		conditions = append(conditions, fmt.Sprintf("channel = $%d", argIdx))
		args = append(args, string(*filter.Channel))
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(id) FROM notifications.notifications WHERE %s", whereClause)
	var totalCount int
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			id, legal_entity_id, recipient_user_id, recipient_email, channel,
			title_ar, title_en, body_ar, body_en, category, priority,
			action_url, entity_type, entity_id, status, error_message,
			sent_at, read_at, created_at
		FROM notifications.notifications
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var ch, cat, prio, st string
		var actURL, entType, errMsg *string
		var recEmail *string

		err := rows.Scan(
			&n.ID,
			&n.LegalEntityID,
			&n.RecipientUserID,
			&recEmail,
			&ch,
			&n.TitleAr,
			&n.TitleEn,
			&n.BodyAr,
			&n.BodyEn,
			&cat,
			&prio,
			&actURL,
			&entType,
			&n.EntityID,
			&st,
			&errMsg,
			&n.SentAt,
			&n.ReadAt,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification row: %w", err)
		}

		if recEmail != nil {
			n.RecipientEmail = *recEmail
		}
		n.Channel = domain.NotificationChannel(ch)
		n.Category = domain.NotificationCategory(cat)
		n.Priority = domain.NotificationPriority(prio)
		n.Status = domain.NotificationStatus(st)
		if actURL != nil {
			n.ActionURL = *actURL
		}
		if entType != nil {
			n.EntityType = *entType
		}
		n.ErrorMessage = errMsg

		notifications = append(notifications, n)
	}

	return notifications, totalCount, nil
}

// CountUnread returns count of unread notifications for a recipient.
func (r *NotificationRepository) CountUnread(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(id)
		FROM notifications.notifications
		WHERE legal_entity_id = $1 AND status != 'READ'
		  AND ($2::uuid IS NULL OR recipient_user_id = $2 OR recipient_user_id IS NULL)
	`
	var count int64
	err := r.db.Pool.QueryRow(ctx, query, legalEntityID, recipientUserID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}

// MarkAsRead marks a specific notification as READ.
func (r *NotificationRepository) MarkAsRead(ctx context.Context, legalEntityID, id uuid.UUID) error {
	query := `
		UPDATE notifications.notifications
		SET status = 'READ', read_at = CURRENT_TIMESTAMP
		WHERE legal_entity_id = $1 AND id = $2
	`
	tag, err := r.db.Pool.Exec(ctx, query, legalEntityID, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotificationNotFound
	}
	return nil
}

// MarkAllAsRead marks all notifications for a recipient as READ.
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) error {
	query := `
		UPDATE notifications.notifications
		SET status = 'READ', read_at = CURRENT_TIMESTAMP
		WHERE legal_entity_id = $1 AND status != 'READ'
		  AND ($2::uuid IS NULL OR recipient_user_id = $2 OR recipient_user_id IS NULL)
	`
	_, err := r.db.Pool.Exec(ctx, query, legalEntityID, recipientUserID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// UpdateStatus updates the delivery status of a notification.
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus, sentAt *time.Time, errMsg *string) error {
	query := `
		UPDATE notifications.notifications
		SET status = $1, sent_at = COALESCE($2, sent_at), error_message = $3
		WHERE id = $4
	`
	_, err := r.db.Pool.Exec(ctx, query, string(status), sentAt, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}
	return nil
}

// Delete removes a notification record.
func (r *NotificationRepository) Delete(ctx context.Context, legalEntityID, id uuid.UUID) error {
	query := `DELETE FROM notifications.notifications WHERE legal_entity_id = $1 AND id = $2`
	tag, err := r.db.Pool.Exec(ctx, query, legalEntityID, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotificationNotFound
	}
	return nil
}

// GetSummary returns notification counts and category breakdown.
func (r *NotificationRepository) GetSummary(ctx context.Context, legalEntityID uuid.UUID, recipientUserID *uuid.UUID) (*domain.NotificationSummary, error) {
	query := `
		SELECT 
			COUNT(id) as total,
			COUNT(CASE WHEN status != 'READ' THEN 1 END) as unread,
			COUNT(CASE WHEN status = 'SENT' THEN 1 END) as sent,
			COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed
		FROM notifications.notifications
		WHERE legal_entity_id = $1
		  AND ($2::uuid IS NULL OR recipient_user_id = $2 OR recipient_user_id IS NULL)
	`

	summary := &domain.NotificationSummary{
		CategoriesCount: make(map[string]int64),
	}

	err := r.db.Pool.QueryRow(ctx, query, legalEntityID, recipientUserID).Scan(
		&summary.TotalCount,
		&summary.UnreadCount,
		&summary.SentCount,
		&summary.FailedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification summary: %w", err)
	}

	catQuery := `
		SELECT category, COUNT(id)
		FROM notifications.notifications
		WHERE legal_entity_id = $1
		  AND ($2::uuid IS NULL OR recipient_user_id = $2 OR recipient_user_id IS NULL)
		GROUP BY category
	`
	rows, err := r.db.Pool.Query(ctx, catQuery, legalEntityID, recipientUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category summary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cat string
		var count int64
		if err := rows.Scan(&cat, &count); err == nil {
			summary.CategoriesCount[cat] = count
		}
	}

	return summary, nil
}

// GetTemplateByCode retrieves an active notification template by unique code.
func (r *NotificationRepository) GetTemplateByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error) {
	query := `
		SELECT id, code, title_ar, title_en, body_ar, body_en, category, channel, is_active, created_at
		FROM notifications.templates
		WHERE code = $1 AND is_active = true
	`
	var t domain.NotificationTemplate
	var cat, ch string
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&t.ID,
		&t.Code,
		&t.TitleAr,
		&t.TitleEn,
		&t.BodyAr,
		&t.BodyEn,
		&cat,
		&ch,
		&t.IsActive,
		&t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get notification template: %w", err)
	}
	t.Category = domain.NotificationCategory(cat)
	t.Channel = domain.NotificationChannel(ch)
	return &t, nil
}
