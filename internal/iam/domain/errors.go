package domain

import "errors"

var (
	ErrInvalidCredentials   = errors.New("invalid username, email or password")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserDisabled         = errors.New("user account is disabled or suspended")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionNotFound      = errors.New("session not found")
	ErrTooManyLoginAttempts = errors.New("too many failed login attempts, please try again later")
	ErrPasswordTooWeak      = errors.New("password must be at least 8 characters long")
	ErrUsernameOrEmailTaken = errors.New("username or email is already registered")
)
