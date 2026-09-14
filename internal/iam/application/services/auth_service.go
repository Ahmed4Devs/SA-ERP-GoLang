package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"erp/internal/iam/domain"
	"erp/internal/iam/infrastructure/crypto"
	"erp/internal/iam/ports"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo     ports.UserRepository
	passwordRepo ports.PasswordRepository
	sessionRepo  ports.SessionRepository
	attemptRepo  ports.LoginAttemptRepository
	hasher       ports.PasswordHasher
	tokens       ports.TokenGenerator
	txManager    transaction.Manager
	sessionTTL   time.Duration
}

func NewAuthService(
	userRepo ports.UserRepository,
	passwordRepo ports.PasswordRepository,
	sessionRepo ports.SessionRepository,
	attemptRepo ports.LoginAttemptRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenGenerator,
	txManager transaction.Manager,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		passwordRepo: passwordRepo,
		sessionRepo:  sessionRepo,
		attemptRepo:  attemptRepo,
		hasher:       hasher,
		tokens:       tokens,
		txManager:    txManager,
		sessionTTL:   7 * 24 * time.Hour,
	}
}

func (s *AuthService) Login(ctx context.Context, identifier, password, ip, userAgent string) (*domain.Session, string, *domain.User, error) {
	// 1. Rate Limiting Check: Max 5 failed attempts in the last 15 minutes
	recentWindow := time.Now().Add(-15 * time.Minute)
	failedCount, err := s.attemptRepo.CountRecentFailed(ctx, identifier, recentWindow)
	if err == nil && failedCount >= 5 {
		return nil, "", nil, domain.ErrTooManyLoginAttempts
	}

	// 2. Fetch User by username or email
	user, err := s.userRepo.GetByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			_ = s.attemptRepo.Record(ctx, identifier, ip, false)
			return nil, "", nil, domain.ErrInvalidCredentials
		}
		return nil, "", nil, err
	}

	// 3. Check Account Status
	if !user.CanLogin() {
		return nil, "", nil, domain.ErrUserDisabled
	}

	// 4. Verify Password Credential
	cred, err := s.passwordRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		_ = s.attemptRepo.Record(ctx, identifier, ip, false)
		return nil, "", nil, domain.ErrInvalidCredentials
	}

	valid, err := s.hasher.Verify(password, cred.PasswordHash)
	if err != nil || !valid {
		_ = s.attemptRepo.Record(ctx, identifier, ip, false)
		return nil, "", nil, domain.ErrInvalidCredentials
	}

	// 5. Record Success & Update Last Login
	_ = s.attemptRepo.Record(ctx, identifier, ip, true)
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// 6. Generate Session Token and Hash
	rawToken, tokenHash, err := s.tokens.GenerateSessionToken()
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// 7. Create Session
	session := &domain.Session{
		UserID:    user.ID,
		TokenHash: tokenHash,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(s.sessionTTL),
	}

	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to persist session: %w", err)
	}

	return createdSession, rawToken, user, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, rawToken string) (*domain.User, *domain.Session, error) {
	if rawToken == "" {
		return nil, nil, domain.ErrSessionNotFound
	}

	tokenHash := crypto.HashToken(rawToken)
	session, user, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, nil, err
	}

	if session.IsExpired() {
		_ = s.sessionRepo.DeleteByTokenHash(ctx, tokenHash)
		return nil, nil, domain.ErrSessionExpired
	}

	if !user.CanLogin() {
		return nil, nil, domain.ErrUserDisabled
	}

	// Debounced last active update
	if time.Since(session.LastActiveAt) > 5*time.Minute {
		_ = s.sessionRepo.UpdateLastActive(ctx, tokenHash)
	}

	return user, session, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	tokenHash := crypto.HashToken(rawToken)
	return s.sessionRepo.DeleteByTokenHash(ctx, tokenHash)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteAllForUser(ctx, userID)
}

func (s *AuthService) SeedAdmin(ctx context.Context, username, email, displayName, password string) (*domain.User, error) {
	if err := domain.ValidatePasswordStrength(password); err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash admin password: %w", err)
	}

	var createdUser *domain.User

	err = s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		user := &domain.User{
			Username:          &username,
			Email:             &email,
			DisplayName:       displayName,
			PreferredLocale:   "ar-SA",
			PreferredTimezone: "Asia/Riyadh",
			Status:            domain.UserStatusActive,
		}

		u, err := s.userRepo.Create(txCtx, user)
		if err != nil {
			return err
		}

		if err := s.passwordRepo.SetPassword(txCtx, u.ID, hash); err != nil {
			return err
		}

		createdUser = u
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
