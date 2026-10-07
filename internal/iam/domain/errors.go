package domain

import "errors"

var (
	ErrInvalidUserInput     = errors.New("invalid user details")
	ErrUserVersionConflict  = errors.New("user was changed by another operation")
	ErrSelfDeactivation     = errors.New("you cannot deactivate your own account")
	ErrLastAdministrator    = errors.New("the last active administrator cannot be deactivated")
	ErrInvalidCredentials   = errors.New("invalid username, email or password")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserDisabled         = errors.New("user account is disabled or suspended")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionNotFound      = errors.New("session not found")
	ErrTooManyLoginAttempts = errors.New("too many failed login attempts, please try again later")
	ErrPasswordTooWeak      = errors.New("password must be at least 8 characters long")
	ErrUsernameOrEmailTaken = errors.New("username or email is already registered")
	ErrMFARequired          = errors.New("multi-factor authentication is required")
	ErrMFAUnavailable       = errors.New("multi-factor authentication is not configured")
	ErrMFAAlreadyEnabled    = errors.New("multi-factor authentication is already enabled")
	ErrMFAInvalidCode       = errors.New("invalid or already used multi-factor code")
	ErrMFAChallengeExpired  = errors.New("multi-factor challenge is invalid or expired")
)
