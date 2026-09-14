-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS iam;

CREATE TABLE IF NOT EXISTS iam.users (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    username            TEXT,
    email               TEXT,
    phone_number        TEXT,
    display_name        TEXT NOT NULL,
    preferred_locale    TEXT NOT NULL DEFAULT 'ar-SA',
    preferred_timezone  TEXT NOT NULL DEFAULT 'Asia/Riyadh',
    status              TEXT NOT NULL DEFAULT 'ACTIVE',
    last_login_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    disabled_at         TIMESTAMPTZ,
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT ck_users_status
        CHECK (status IN ('ACTIVE', 'SUSPENDED', 'DISABLED'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower
ON iam.users (LOWER(username))
WHERE username IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower
ON iam.users (LOWER(email))
WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_number
ON iam.users (phone_number)
WHERE phone_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS iam.password_credentials (
    user_id             UUID PRIMARY KEY REFERENCES iam.users(id) ON DELETE CASCADE,
    password_hash       TEXT NOT NULL,
    algorithm           TEXT NOT NULL DEFAULT 'argon2id',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam.sessions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    token_hash          TEXT NOT NULL UNIQUE,
    ip_address          TEXT,
    user_agent          TEXT,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_active_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id
ON iam.sessions (user_id);

CREATE TABLE IF NOT EXISTS iam.login_attempts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    identifier          TEXT NOT NULL,
    ip_address          TEXT NOT NULL,
    success             BOOLEAN NOT NULL,
    attempted_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_lookup
ON iam.login_attempts (identifier, attempted_at);

CREATE TABLE IF NOT EXISTS iam.mfa_factors (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    factor_type         TEXT NOT NULL,
    secret              TEXT NOT NULL,
    is_verified         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam.roles (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    code                TEXT NOT NULL UNIQUE,
    name_en             TEXT NOT NULL,
    name_ar             TEXT,
    description         TEXT,
    is_system           BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam.permissions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    code                TEXT NOT NULL UNIQUE,
    name_en             TEXT NOT NULL,
    name_ar             TEXT,
    module              TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam.role_permissions (
    role_id             UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    permission_id       UUID NOT NULL REFERENCES iam.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS iam.user_role_assignments (
    user_id             UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    role_id             UUID NOT NULL REFERENCES iam.roles(id) ON DELETE CASCADE,
    assigned_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS iam.service_accounts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    name                TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS iam.api_keys (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID REFERENCES iam.users(id) ON DELETE CASCADE,
    key_hash            TEXT NOT NULL UNIQUE,
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam.api_keys;
DROP TABLE IF EXISTS iam.service_accounts;
DROP TABLE IF EXISTS iam.user_role_assignments;
DROP TABLE IF EXISTS iam.role_permissions;
DROP TABLE IF EXISTS iam.permissions;
DROP TABLE IF EXISTS iam.roles;
DROP TABLE IF EXISTS iam.mfa_factors;
DROP TABLE IF EXISTS iam.login_attempts;
DROP TABLE IF EXISTS iam.sessions;
DROP TABLE IF EXISTS iam.password_credentials;
DROP TABLE IF EXISTS iam.users;
DROP SCHEMA IF EXISTS iam CASCADE;
-- +goose StatementEnd
