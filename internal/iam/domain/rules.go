package domain

import "time"

func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooWeak
	}
	return nil
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive
}
