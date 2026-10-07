package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"erp/internal/iam/domain"
	"erp/internal/iam/ports"
	"erp/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MFARepository struct {
	db *database.DB
}

func NewMFARepository(db *database.DB) ports.MFARepository {
	return &MFARepository{db: db}
}

func (r *MFARepository) UpsertTOTPFactor(ctx context.Context, userID uuid.UUID, encryptedSecret string) (*domain.MFAFactor, error) {
	row := r.db.DBTX(ctx).QueryRow(ctx, `
		INSERT INTO iam.mfa_factors (user_id, factor_type, secret, is_verified, verified_at, last_used_counter)
		VALUES ($1, 'TOTP', $2, FALSE, NULL, NULL)
		ON CONFLICT (user_id, factor_type) DO UPDATE SET
			secret = EXCLUDED.secret,
			is_verified = FALSE,
			verified_at = NULL,
			last_used_counter = NULL,
			created_at = now()
		WHERE iam.mfa_factors.is_verified = FALSE
		RETURNING id, user_id, factor_type, secret, is_verified, verified_at, last_used_counter`, userID, encryptedSecret)
	factor, err := scanFactor(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMFAAlreadyEnabled
	}
	return factor, err
}

func (r *MFARepository) GetTOTPFactor(ctx context.Context, userID uuid.UUID, verifiedOnly bool) (*domain.MFAFactor, error) {
	query := `SELECT id, user_id, factor_type, secret, is_verified, verified_at, last_used_counter
		FROM iam.mfa_factors WHERE user_id = $1 AND factor_type = 'TOTP'`
	if verifiedOnly {
		query += " AND is_verified = TRUE"
	}
	factor, err := scanFactor(r.db.DBTX(ctx).QueryRow(ctx, query, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMFAUnavailable
	}
	return factor, err
}

func (r *MFARepository) VerifyTOTPFactor(ctx context.Context, userID uuid.UUID, counter int64) error {
	result, err := r.db.DBTX(ctx).Exec(ctx, `
		UPDATE iam.mfa_factors
		SET is_verified = TRUE, verified_at = now(), last_used_counter = $2
		WHERE user_id = $1 AND factor_type = 'TOTP' AND is_verified = FALSE`, userID, counter)
	if err != nil {
		return fmt.Errorf("failed to verify MFA factor: %w", err)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrMFAUnavailable
	}
	return nil
}

func (r *MFARepository) UpdateLastUsedCounter(ctx context.Context, userID uuid.UUID, counter int64) (bool, error) {
	result, err := r.db.DBTX(ctx).Exec(ctx, `
		UPDATE iam.mfa_factors SET last_used_counter = $2
		WHERE user_id = $1 AND factor_type = 'TOTP' AND is_verified = TRUE
		  AND (last_used_counter IS NULL OR last_used_counter < $2)`, userID, counter)
	if err != nil {
		return false, fmt.Errorf("failed to update MFA counter: %w", err)
	}
	return result.RowsAffected() == 1, nil
}

func (r *MFARepository) ReplaceRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	if _, err := r.db.DBTX(ctx).Exec(ctx, "DELETE FROM iam.mfa_recovery_codes WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("failed to clear MFA recovery codes: %w", err)
	}
	for _, hash := range codeHashes {
		if _, err := r.db.DBTX(ctx).Exec(ctx, `INSERT INTO iam.mfa_recovery_codes (user_id, code_hash) VALUES ($1, $2)`, userID, hash); err != nil {
			return fmt.Errorf("failed to store MFA recovery code: %w", err)
		}
	}
	return nil
}

func (r *MFARepository) ConsumeRecoveryCode(ctx context.Context, userID uuid.UUID, codeHash string) (bool, error) {
	result, err := r.db.DBTX(ctx).Exec(ctx, `
		UPDATE iam.mfa_recovery_codes SET used_at = now()
		WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL`, userID, codeHash)
	if err != nil {
		return false, fmt.Errorf("failed to consume MFA recovery code: %w", err)
	}
	return result.RowsAffected() == 1, nil
}

func (r *MFARepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.db.DBTX(ctx).Exec(ctx, "DELETE FROM iam.mfa_recovery_codes WHERE user_id = $1", userID); err != nil {
		return err
	}
	_, err := r.db.DBTX(ctx).Exec(ctx, "DELETE FROM iam.mfa_factors WHERE user_id = $1 AND factor_type = 'TOTP'", userID)
	return err
}

func (r *MFARepository) CreateChallenge(ctx context.Context, challenge *domain.MFAChallenge) (*domain.MFAChallenge, error) {
	row := r.db.DBTX(ctx).QueryRow(ctx, `
		INSERT INTO iam.mfa_login_challenges (user_id, token_hash, ip_address, user_agent, expires_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5)
		RETURNING id, user_id, token_hash, COALESCE(ip_address, ''), COALESCE(user_agent, ''), attempt_count, expires_at,
		(SELECT version FROM iam.users WHERE id = $1)`,
		challenge.UserID, challenge.TokenHash, challenge.IPAddress, challenge.UserAgent, challenge.ExpiresAt)
	return scanChallenge(row)
}

func (r *MFARepository) GetActiveChallenge(ctx context.Context, tokenHash string) (*domain.MFAChallenge, error) {
	challenge, err := scanChallenge(r.db.DBTX(ctx).QueryRow(ctx, `
		SELECT c.id, c.user_id, c.token_hash, COALESCE(c.ip_address, ''), COALESCE(c.user_agent, ''), c.attempt_count, c.expires_at, u.version
		FROM iam.mfa_login_challenges c JOIN iam.users u ON u.id = c.user_id
		WHERE c.token_hash = $1 AND c.consumed_at IS NULL AND c.expires_at > now() AND c.attempt_count < 5 AND u.status = 'ACTIVE'
		FOR UPDATE OF c`, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrMFAChallengeExpired
	}
	return challenge, err
}

func (r *MFARepository) IncrementChallengeAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.DBTX(ctx).Exec(ctx, `UPDATE iam.mfa_login_challenges SET attempt_count = attempt_count + 1 WHERE id = $1 AND consumed_at IS NULL AND attempt_count < 5`, id)
	return err
}

func (r *MFARepository) DeleteLoginChallenges(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.DBTX(ctx).Exec(ctx, "DELETE FROM iam.mfa_login_challenges WHERE user_id=$1", userID)
	return err
}

func (r *MFARepository) ConsumeChallenge(ctx context.Context, id uuid.UUID) (bool, error) {
	result, err := r.db.DBTX(ctx).Exec(ctx, `UPDATE iam.mfa_login_challenges SET consumed_at = now() WHERE id = $1 AND consumed_at IS NULL AND expires_at > now()`, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() == 1, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanFactor(row rowScanner) (*domain.MFAFactor, error) {
	var factor domain.MFAFactor
	var verifiedAt *time.Time
	var lastUsedCounter *int64
	if err := row.Scan(&factor.ID, &factor.UserID, &factor.FactorType, &factor.EncryptedSecret, &factor.IsVerified, &verifiedAt, &lastUsedCounter); err != nil {
		return nil, err
	}
	factor.VerifiedAt = verifiedAt
	factor.LastUsedCounter = lastUsedCounter
	return &factor, nil
}

func scanChallenge(row rowScanner) (*domain.MFAChallenge, error) {
	var challenge domain.MFAChallenge
	if err := row.Scan(&challenge.ID, &challenge.UserID, &challenge.TokenHash, &challenge.IPAddress, &challenge.UserAgent, &challenge.AttemptCount, &challenge.ExpiresAt, &challenge.UserVersion); err != nil {
		return nil, err
	}
	return &challenge, nil
}
