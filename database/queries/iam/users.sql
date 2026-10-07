-- name: CreateUser :one
INSERT INTO iam.users (
    username,
    email,
    phone_number,
    display_name,
    preferred_locale,
    preferred_timezone,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, username, email, phone_number, display_name, preferred_locale, preferred_timezone, status, last_login_at, created_at, updated_at, disabled_at, version;

-- name: GetUserByID :one
SELECT id, username, email, phone_number, display_name, preferred_locale, preferred_timezone, status, last_login_at, created_at, updated_at, disabled_at, version
FROM iam.users
WHERE id = $1;

-- name: GetUserByUsernameOrEmail :one
SELECT id, username, email, phone_number, display_name, preferred_locale, preferred_timezone, status, last_login_at, created_at, updated_at, disabled_at, version
FROM iam.users
WHERE LOWER(username) = LOWER($1) OR LOWER(email) = LOWER($1)
LIMIT 1;

-- name: UpdateUserLastLogin :exec
UPDATE iam.users
SET last_login_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE iam.users
SET status = $2,
    disabled_at = CASE WHEN $2 = 'DISABLED' THEN now() ELSE NULL END,
    updated_at = now(),
    version = version + 1
WHERE id = $1;

-- name: InvalidateUserAuthentication :exec
UPDATE iam.users
SET version = version + 1, updated_at = now()
WHERE id = $1;
