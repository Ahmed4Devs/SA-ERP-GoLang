-- name: InsertSecurityEvent :one
INSERT INTO audit.security_events (
    user_id,
    event_type,
    status,
    details,
    ip_address
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListSecurityEvents :many
SELECT * FROM audit.security_events
WHERE
    (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
    AND (sqlc.narg('event_type')::text IS NULL OR event_type = sqlc.narg('event_type'))
    AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
ORDER BY occurred_at DESC, id DESC
LIMIT $1 OFFSET $2;
