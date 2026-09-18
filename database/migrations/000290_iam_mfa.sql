-- +goose Up
-- +goose StatementBegin

ALTER TABLE iam.mfa_factors
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_used_counter BIGINT;

-- Earlier releases defined this table but did not implement a usable MFA
-- flow or encrypted secret format. Never treat those legacy rows as active.
UPDATE iam.mfa_factors
SET is_verified = FALSE,
    verified_at = NULL,
    last_used_counter = NULL;

DELETE FROM iam.mfa_factors factor
USING (
    SELECT ctid,
           row_number() OVER (
               PARTITION BY user_id, factor_type
               ORDER BY created_at DESC, id DESC
           ) AS row_number
    FROM iam.mfa_factors
) duplicate
WHERE factor.ctid = duplicate.ctid
  AND duplicate.row_number > 1;

CREATE UNIQUE INDEX IF NOT EXISTS uq_mfa_factors_user_type
ON iam.mfa_factors (user_id, factor_type);

CREATE TABLE IF NOT EXISTS iam.mfa_recovery_codes (
    id          UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id     UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    code_hash   TEXT NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, code_hash)
);

CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_unused
ON iam.mfa_recovery_codes (user_id)
WHERE used_at IS NULL;

CREATE TABLE IF NOT EXISTS iam.mfa_login_challenges (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id       UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL UNIQUE,
    ip_address    TEXT,
    user_agent    TEXT,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    expires_at    TIMESTAMPTZ NOT NULL,
    consumed_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_mfa_login_challenge_attempts CHECK (attempt_count BETWEEN 0 AND 5)
);

CREATE INDEX IF NOT EXISTS idx_mfa_login_challenges_active
ON iam.mfa_login_challenges (token_hash, expires_at)
WHERE consumed_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam.mfa_login_challenges;
DROP TABLE IF EXISTS iam.mfa_recovery_codes;
DROP INDEX IF EXISTS iam.uq_mfa_factors_user_type;
ALTER TABLE iam.mfa_factors
    DROP COLUMN IF EXISTS last_used_counter,
    DROP COLUMN IF EXISTS verified_at;
-- +goose StatementEnd
