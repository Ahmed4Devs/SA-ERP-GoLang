package domain

import "errors"

var (
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrRecipientRequired     = errors.New("recipient user id or email address is required")
	ErrTitleRequired         = errors.New("notification title is required")
	ErrBodyRequired          = errors.New("notification body is required")
	ErrInvalidChannel        = errors.New("invalid notification channel")
	ErrTemplateNotFound      = errors.New("notification template not found")
	ErrNotificationDispatch  = errors.New("failed to dispatch notification")
)
