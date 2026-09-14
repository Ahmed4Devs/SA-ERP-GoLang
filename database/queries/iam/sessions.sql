-- name: CreateSession :one
INSERT INTO iam.sessions (
    user_id,
    token_hash,
    ip_address,
    user_agent,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, user_id, token_hash, ip_address, user_agent, expires_at, created_at, last_active_at;

-- name: GetSessionByTokenHash :one
SELECT s.id, s.user_id, s.token_hash, s.ip_address, s.user_agent, s.expires_at, s.created_at, s.last_active_at,
       u.username, u.email, u.display_name, u.status, u.preferred_locale, u.preferred_timezone
FROM iam.sessions s
JOIN iam.users u ON u.id = s.user_id
WHERE s.token_hash = $1 AND s.expires_at > now() AND u.status = 'ACTIVE';

-- name: UpdateSessionLastActive :exec
UPDATE iam.sessions
SET last_active_at = now()
WHERE token_hash = $1;

-- name: DeleteSessionByTokenHash :exec
DELETE FROM iam.sessions
WHERE token_hash = $1;

-- name: DeleteAllSessionsForUser :exec
DELETE FROM iam.sessions
WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM iam.sessions
WHERE expires_at <= now();
