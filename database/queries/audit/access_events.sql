-- name: InsertAccessEvent :one
INSERT INTO audit.access_events (
    user_id,
    resource_path,
    method,
    status_code,
    ip_address,
    latency_ms
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListAccessEvents :many
SELECT * FROM audit.access_events
WHERE
    (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
    AND (sqlc.narg('resource_path')::text IS NULL OR resource_path = sqlc.narg('resource_path'))
ORDER BY occurred_at DESC, id DESC
LIMIT $1 OFFSET $2;
