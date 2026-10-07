package services

import (
	"context"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"erp/internal/iam/domain"
	"erp/internal/iam/ports"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

type UserManagementService struct {
	users     ports.ManagedUserRepository
	passwords ports.PasswordRepository
	sessions  ports.SessionRepository
	hasher    ports.PasswordHasher
	tx        transaction.Manager
}

type CreateUserCommand struct {
	Username          string `json:"username"`
	Email             string `json:"email"`
	DisplayName       string `json:"display_name"`
	Password          string `json:"password"`
	PreferredLocale   string `json:"preferred_locale"`
	PreferredTimezone string `json:"preferred_timezone"`
}

func NewUserManagementService(users ports.ManagedUserRepository, passwords ports.PasswordRepository, sessions ports.SessionRepository, hasher ports.PasswordHasher, tx transaction.Manager) *UserManagementService {
	return &UserManagementService{users, passwords, sessions, hasher, tx}
}

func (s *UserManagementService) List(ctx context.Context, search string, limit, offset int32) ([]*domain.User, error) {
	if limit < 1 || limit > 100 || offset < 0 || len(search) > 200 {
		return nil, domain.ErrInvalidUserInput
	}
	return s.users.List(ctx, strings.TrimSpace(search), limit, offset)
}

func (s *UserManagementService) Create(ctx context.Context, actorID uuid.UUID, cmd CreateUserCommand) (*domain.User, error) {
	cmd.Username = strings.TrimSpace(cmd.Username)
	cmd.Email = strings.ToLower(strings.TrimSpace(cmd.Email))
	cmd.DisplayName = strings.TrimSpace(cmd.DisplayName)
	if actorID == uuid.Nil || len(cmd.Username) < 3 || len(cmd.Username) > 100 || cmd.DisplayName == "" || len(cmd.DisplayName) > 200 || len(cmd.Email) > 254 {
		return nil, domain.ErrInvalidUserInput
	}
	for _, ch := range cmd.Username {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !strings.ContainsRune("._-", ch) {
			return nil, domain.ErrInvalidUserInput
		}
	}
	if cmd.Email != "" {
		addr, err := mail.ParseAddress(cmd.Email)
		if err != nil || addr.Address != cmd.Email {
			return nil, domain.ErrInvalidUserInput
		}
	}
	if cmd.PreferredLocale == "" {
		cmd.PreferredLocale = "ar-SA"
	}
	if cmd.PreferredLocale != "ar-SA" && cmd.PreferredLocale != "en" {
		return nil, domain.ErrInvalidUserInput
	}
	if cmd.PreferredTimezone == "" {
		cmd.PreferredTimezone = "Asia/Riyadh"
	}
	if _, err := time.LoadLocation(cmd.PreferredTimezone); err != nil {
		return nil, domain.ErrInvalidUserInput
	}
	if err := validateManagedPassword(cmd.Password); err != nil {
		return nil, err
	}
	hash, err := s.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}
	var user *domain.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var email *string
		if cmd.Email != "" {
			email = &cmd.Email
		}
		user, err = s.users.Create(ctx, &domain.User{Username: &cmd.Username, Email: email, DisplayName: cmd.DisplayName, PreferredLocale: cmd.PreferredLocale, PreferredTimezone: cmd.PreferredTimezone, Status: domain.UserStatusActive})
		if err != nil {
			return err
		}
		if err := s.passwords.SetPassword(ctx, user.ID, hash); err != nil {
			return err
		}
		return s.users.RecordChange(ctx, actorID, user.ID, "IAM_USER_CREATED")
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func validateManagedPassword(password string) error {
	if len(password) > 1024 {
		return domain.ErrInvalidUserInput
	}
	return domain.ValidatePasswordStrength(password)
}

func (s *UserManagementService) ChangeStatus(ctx context.Context, actorID, targetID uuid.UUID, status domain.UserStatus, version int64) (*domain.User, error) {
	if actorID == uuid.Nil || targetID == uuid.Nil || version < 1 {
		return nil, domain.ErrInvalidUserInput
	}
	if status != domain.UserStatusActive && status != domain.UserStatusDisabled && status != domain.UserStatusSuspended {
		return nil, domain.ErrInvalidUserInput
	}
	if targetID == actorID && status != domain.UserStatusActive {
		return nil, domain.ErrSelfDeactivation
	}
	var user *domain.User
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.users.LockAdministration(ctx); err != nil {
			return err
		}
		current, err := s.users.Lock(ctx, targetID)
		if err != nil {
			return err
		}
		if current.Version != version {
			return domain.ErrUserVersionConflict
		}
		if status != domain.UserStatusActive {
			allowed, err := s.users.CanDeactivate(ctx, targetID)
			if err != nil {
				return err
			}
			if !allowed {
				return domain.ErrLastAdministrator
			}
		}
		user, err = s.users.UpdateState(ctx, targetID, status, version)
		if err != nil {
			return err
		}
		if status != domain.UserStatusActive {
			if err := s.sessions.DeleteAllForUser(ctx, targetID); err != nil {
				return err
			}
		}
		return s.users.RecordChange(ctx, actorID, targetID, "IAM_USER_"+string(status))
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserManagementService) ResetPassword(ctx context.Context, actorID, targetID uuid.UUID, password string, version int64) error {
	if actorID == uuid.Nil || targetID == uuid.Nil || version < 1 {
		return domain.ErrInvalidUserInput
	}
	if err := validateManagedPassword(password); err != nil {
		return err
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	return s.tx.WithinTx(ctx, func(ctx context.Context) error {
		user, err := s.users.Lock(ctx, targetID)
		if err != nil {
			return err
		}
		if user.Version != version {
			return domain.ErrUserVersionConflict
		}
		if _, err := s.users.UpdateState(ctx, targetID, user.Status, version); err != nil {
			return err
		}
		if err := s.passwords.SetPassword(ctx, targetID, hash); err != nil {
			return err
		}
		if err := s.sessions.DeleteAllForUser(ctx, targetID); err != nil {
			return err
		}
		return s.users.RecordChange(ctx, actorID, targetID, "IAM_PASSWORD_RESET")
	})
}
