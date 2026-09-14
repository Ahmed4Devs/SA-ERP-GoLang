package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusDisabled  UserStatus = "DISABLED"
)

type User struct {
	ID                uuid.UUID  `json:"id"`
	Username          *string    `json:"username,omitempty"`
	Email             *string    `json:"email,omitempty"`
	PhoneNumber       *string    `json:"phone_number,omitempty"`
	DisplayName       string     `json:"display_name"`
	PreferredLocale   string     `json:"preferred_locale"`
	PreferredTimezone string     `json:"preferred_timezone"`
	Status            UserStatus `json:"status"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DisabledAt        *time.Time `json:"disabled_at,omitempty"`
	Version           int64      `json:"version"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	TokenHash    string    `json:"-"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

type PasswordCredential struct {
	UserID       uuid.UUID `json:"user_id"`
	PasswordHash string    `json:"-"`
	Algorithm    string    `json:"algorithm"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
