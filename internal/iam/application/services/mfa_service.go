package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"erp/internal/iam/domain"
	iamcrypto "erp/internal/iam/infrastructure/crypto"
	"erp/internal/iam/ports"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type MFAService struct {
	repo          ports.MFARepository
	tokens        ports.TokenGenerator
	secretBox     *iamcrypto.SecretBox
	txManager     transaction.Manager
	issuer        string
	challengeTTL  time.Duration
	recoveryCount int
}

type MFAEnrollment struct {
	Secret string   `json:"secret"`
	URI    string   `json:"otpauth_uri"`
	Codes  []string `json:"recovery_codes"`
}

func NewMFAService(repo ports.MFARepository, tokens ports.TokenGenerator, secretBox *iamcrypto.SecretBox, txManager transaction.Manager, issuer string) *MFAService {
	return &MFAService{repo: repo, tokens: tokens, secretBox: secretBox, txManager: txManager, issuer: issuer, challengeTTL: 5 * time.Minute, recoveryCount: 10}
}

func (s *MFAService) IsEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	if s == nil {
		return false, nil
	}
	_, err := s.repo.GetTOTPFactor(ctx, userID, true)
	if errors.Is(err, domain.ErrMFAUnavailable) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if s.secretBox == nil {
		return false, domain.ErrMFAUnavailable
	}
	return true, nil
}

func (s *MFAService) BeginEnrollment(ctx context.Context, user *domain.User) (*MFAEnrollment, error) {
	if s == nil || s.secretBox == nil {
		return nil, domain.ErrMFAUnavailable
	}
	accountName := user.ID.String()
	if user.Email != nil && *user.Email != "" {
		accountName = *user.Email
	} else if user.Username != nil && *user.Username != "" {
		accountName = *user.Username
	}
	key, err := totp.Generate(totp.GenerateOpts{Issuer: s.issuer, AccountName: accountName})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	encrypted, err := s.secretBox.Encrypt(key.Secret())
	if err != nil {
		return nil, err
	}
	codes, hashes, err := generateRecoveryCodes(s.recoveryCount)
	if err != nil {
		return nil, err
	}
	if err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.UpsertTOTPFactor(txCtx, user.ID, encrypted); err != nil {
			return err
		}
		return s.repo.ReplaceRecoveryCodes(txCtx, user.ID, hashes)
	}); err != nil {
		return nil, err
	}
	return &MFAEnrollment{Secret: key.Secret(), URI: key.URL(), Codes: codes}, nil
}

func (s *MFAService) ConfirmEnrollment(ctx context.Context, userID uuid.UUID, code string) error {
	factor, err := s.repo.GetTOTPFactor(ctx, userID, false)
	if err != nil || factor.IsVerified {
		return domain.ErrMFAUnavailable
	}
	secret, err := s.secretBox.Decrypt(factor.EncryptedSecret)
	if err != nil {
		return err
	}
	counter, valid := validateTOTP(code, secret, time.Now())
	if !valid {
		return domain.ErrMFAInvalidCode
	}
	return s.repo.VerifyTOTPFactor(ctx, userID, counter)
}

func (s *MFAService) Disable(ctx context.Context, userID uuid.UUID, code string) error {
	return s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.verifyCode(txCtx, userID, code); err != nil {
			return err
		}
		return s.repo.DisableMFA(txCtx, userID)
	})
}

func (s *MFAService) CreateLoginChallenge(ctx context.Context, userID uuid.UUID, ip, userAgent string) (string, time.Time, error) {
	raw, hash, err := s.tokens.GenerateSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(s.challengeTTL)
	_, err = s.repo.CreateChallenge(ctx, &domain.MFAChallenge{UserID: userID, TokenHash: hash, IPAddress: ip, UserAgent: userAgent, ExpiresAt: expiresAt})
	return raw, expiresAt, err
}

func (s *MFAService) CompleteLogin(ctx context.Context, challengeToken, code string) (*domain.MFAChallenge, error) {
	var challenge *domain.MFAChallenge
	err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		challenge, err = s.repo.GetActiveChallenge(txCtx, iamcrypto.HashToken(challengeToken))
		if err != nil {
			return domain.ErrMFAChallengeExpired
		}
		if err := s.verifyCode(txCtx, challenge.UserID, code); err != nil {
			return err
		}
		consumed, err := s.repo.ConsumeChallenge(txCtx, challenge.ID)
		if err != nil {
			return err
		}
		if !consumed {
			return domain.ErrMFAChallengeExpired
		}
		return nil
	})
	if err != nil {
		if challenge != nil && errors.Is(err, domain.ErrMFAInvalidCode) {
			if incrementErr := s.repo.IncrementChallengeAttempts(ctx, challenge.ID); incrementErr != nil {
				return nil, incrementErr
			}
		}
		return nil, err
	}
	return challenge, nil
}

func (s *MFAService) verifyCode(ctx context.Context, userID uuid.UUID, code string) error {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	if len(normalized) == 6 {
		factor, err := s.repo.GetTOTPFactor(ctx, userID, true)
		if err != nil {
			return domain.ErrMFAUnavailable
		}
		secret, err := s.secretBox.Decrypt(factor.EncryptedSecret)
		if err != nil {
			return err
		}
		counter, valid := validateTOTP(normalized, secret, time.Now())
		if valid {
			updated, err := s.repo.UpdateLastUsedCounter(ctx, userID, counter)
			if err != nil {
				return err
			}
			if updated {
				return nil
			}
		}
	}

	hash := sha256.Sum256([]byte(normalized))
	consumed, err := s.repo.ConsumeRecoveryCode(ctx, userID, base64.RawURLEncoding.EncodeToString(hash[:]))
	if err != nil {
		return err
	}
	if !consumed {
		return domain.ErrMFAInvalidCode
	}
	return nil
}

func validateTOTP(code, secret string, now time.Time) (int64, bool) {
	counter := now.Unix() / 30
	for offset := int64(-1); offset <= 1; offset++ {
		candidateTime := time.Unix((counter+offset)*30, 0)
		valid, err := totp.ValidateCustom(code, secret, candidateTime, totp.ValidateOpts{
			Period:    30,
			Skew:      0,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err == nil && valid {
			return counter + offset, true
		}
	}
	return 0, false
}

func generateRecoveryCodes(count int) ([]string, []string, error) {
	codes := make([]string, 0, count)
	hashes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		raw := make([]byte, 10)
		if _, err := rand.Read(raw); err != nil {
			return nil, nil, err
		}
		encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)
		code := encoded[:8] + "-" + encoded[8:]
		normalized := strings.ReplaceAll(code, "-", "")
		hash := sha256.Sum256([]byte(normalized))
		codes = append(codes, code)
		hashes = append(hashes, base64.RawURLEncoding.EncodeToString(hash[:]))
	}
	return codes, hashes, nil
}
