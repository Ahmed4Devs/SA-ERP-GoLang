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
	mfa          *MFAService
	sessionTTL   time.Duration
}

func (s *AuthService) SetMFAService(mfa *MFAService) {
	s.mfa = mfa
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

	if s.mfa != nil {
		mfaEnabled, err := s.mfa.IsEnabled(ctx, user.ID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to determine MFA status: %w", err)
		}
		if mfaEnabled {
			return nil, "", user, domain.ErrMFARequired
		}
	}

	session, rawToken, currentUser, err := s.CreateSessionForUser(ctx, user.ID, user.Version, ip, userAgent)
	if err == nil {
		_ = s.attemptRepo.Record(ctx, identifier, ip, true)
	}
	return session, rawToken, currentUser, err
}

// CreateMFAChallenge rejects a password verification superseded by a password
// reset or status change, including changes while password hashing was running.
func (s *AuthService) CreateMFAChallenge(ctx context.Context, verifiedUser *domain.User, ip, userAgent string) (string, time.Time, error) {
	if verifiedUser == nil || verifiedUser.Version < 1 {
		return "", time.Time{}, domain.ErrInvalidCredentials
	}
	var token string
	var expires time.Time
	err := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		user, err := s.userRepo.Lock(ctx, verifiedUser.ID)
		if err != nil {
			return err
		}
		if !user.CanLogin() {
			return domain.ErrUserDisabled
		}
		if user.Version != verifiedUser.Version {
			return domain.ErrInvalidCredentials
		}
		if s.mfa == nil {
			return domain.ErrMFAUnavailable
		}
		token, expires, err = s.mfa.createLoginChallenge(ctx, user.ID, ip, userAgent)
		return err
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

// CreateSessionForUser requires the user version observed during authentication.
// The user lock orders issuance against password reset and account suspension.
func (s *AuthService) CreateSessionForUser(ctx context.Context, userID uuid.UUID, expectedVersion int64, ip, userAgent string) (*domain.Session, string, *domain.User, error) {
	var user *domain.User
	var createdSession *domain.Session
	var rawToken string
	err := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.userRepo.Lock(ctx, userID)
		if err != nil {
			return err
		}
		if !user.CanLogin() {
			return domain.ErrUserDisabled
		}
		if expectedVersion < 1 || user.Version != expectedVersion {
			return domain.ErrInvalidCredentials
		}
		if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
			return err
		}
		var tokenHash string
		rawToken, tokenHash, err = s.tokens.GenerateSessionToken()
		if err != nil {
			return fmt.Errorf("failed to generate session token: %w", err)
		}
		createdSession, err = s.sessionRepo.Create(ctx, &domain.Session{
			UserID: user.ID, TokenHash: tokenHash, IPAddress: ip, UserAgent: userAgent, ExpiresAt: time.Now().Add(s.sessionTTL),
		})
		if err != nil {
			return fmt.Errorf("failed to persist session: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, "", nil, err
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
	return s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := s.userRepo.Lock(ctx, userID); err != nil {
			return err
		}
		if err := s.userRepo.InvalidateAuthentication(ctx, userID); err != nil {
			return err
		}
		return s.sessionRepo.DeleteAllForUser(ctx, userID)
	})
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

func (s *AuthService) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	return s.userRepo.GetByIdentifier(ctx, identifier)
}
