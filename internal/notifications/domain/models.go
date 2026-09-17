package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationChannel defines transmission mechanisms.
type NotificationChannel string

const (
	ChannelInApp NotificationChannel = "IN_APP"
	ChannelEmail NotificationChannel = "EMAIL"
	ChannelSMS   NotificationChannel = "SMS"
)

// NotificationStatus defines delivery and consumption lifecycle states.
type NotificationStatus string

const (
	StatusPending NotificationStatus = "PENDING"
	StatusSent    NotificationStatus = "SENT"
	StatusFailed  NotificationStatus = "FAILED"
	StatusRead    NotificationStatus = "READ"
)

// NotificationPriority defines urgency level.
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "LOW"
	PriorityNormal NotificationPriority = "NORMAL"
	PriorityHigh   NotificationPriority = "HIGH"
	PriorityUrgent NotificationPriority = "URGENT"
)

// NotificationCategory defines domain grouping.
type NotificationCategory string

const (
	CategoryWorkflow  NotificationCategory = "WORKFLOW"
	CategoryZATCA     NotificationCategory = "ZATCA"
	CategoryFinance   NotificationCategory = "FINANCE"
	CategoryInventory NotificationCategory = "INVENTORY"
	CategoryPayroll   NotificationCategory = "PAYROLL"
	CategorySystem    NotificationCategory = "SYSTEM"
)

// Notification represents an in-app or outbound notification message.
type Notification struct {
	ID              uuid.UUID            `json:"id"`
	LegalEntityID   uuid.UUID            `json:"legal_entity_id"`
	RecipientUserID *uuid.UUID           `json:"recipient_user_id,omitempty"`
	RecipientEmail  string               `json:"recipient_email,omitempty"`
	Channel         NotificationChannel  `json:"channel"`
	TitleAr         string               `json:"title_ar"`
	TitleEn         string               `json:"title_en"`
	BodyAr          string               `json:"body_ar"`
	BodyEn          string               `json:"body_en"`
	Category        NotificationCategory `json:"category"`
	Priority        NotificationPriority `json:"priority"`
	ActionURL       string               `json:"action_url,omitempty"`
	EntityType      string               `json:"entity_type,omitempty"`
	EntityID        *uuid.UUID           `json:"entity_id,omitempty"`
	Status          NotificationStatus   `json:"status"`
	ErrorMessage    *string              `json:"error_message,omitempty"`
	SentAt          *time.Time           `json:"sent_at,omitempty"`
	ReadAt          *time.Time           `json:"read_at,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
}

// NotificationTemplate represents a pre-configured notification blueprint.
type NotificationTemplate struct {
	ID        uuid.UUID            `json:"id"`
	Code      string               `json:"code"`
	TitleAr   string               `json:"title_ar"`
	TitleEn   string               `json:"title_en"`
	BodyAr    string               `json:"body_ar"`
	BodyEn    string               `json:"body_en"`
	Category  NotificationCategory `json:"category"`
	Channel   NotificationChannel  `json:"channel"`
	IsActive  bool                 `json:"is_active"`
	CreatedAt time.Time            `json:"created_at"`
}

// SendNotificationParams contains parameters to dispatch a notification.
type SendNotificationParams struct {
	LegalEntityID   uuid.UUID
	RecipientUserID *uuid.UUID
	RecipientEmail  string
	Channel         NotificationChannel
	TitleAr         string
	TitleEn         string
	BodyAr          string
	BodyEn          string
	Category        NotificationCategory
	Priority        NotificationPriority
	ActionURL       string
	EntityType      string
	EntityID        *uuid.UUID
}

// NotificationFilter contains search and pagination criteria.
type NotificationFilter struct {
	LegalEntityID   uuid.UUID
	RecipientUserID *uuid.UUID
	Category        *NotificationCategory
	Status          *NotificationStatus
	Channel         *NotificationChannel
	Limit           int
	Offset          int
}

// NotificationSummary provides analytics for user and system notification feeds.
type NotificationSummary struct {
	TotalCount      int64            `json:"total_count"`
	UnreadCount     int64            `json:"unread_count"`
	SentCount       int64            `json:"sent_count"`
	FailedCount     int64            `json:"failed_count"`
	CategoriesCount map[string]int64 `json:"categories_count"`
}
