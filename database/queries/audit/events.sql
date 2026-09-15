-- name: InsertAuditEvent :one
INSERT INTO audit.events (
    user_id,
    legal_entity_id,
    action,
    resource_type,
    resource_id,
    payload,
    ip_address,
    user_agent,
    previous_hash,
    record_hash,
    occurred_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetLatestAuditEvent :one
SELECT * FROM audit.events
ORDER BY occurred_at DESC, id DESC
LIMIT 1;

-- name: GetAuditEventByID :one
SELECT * FROM audit.events
WHERE id = $1;

-- name: ListAuditEvents :many
SELECT * FROM audit.events
WHERE
    (sqlc.narg('legal_entity_id')::uuid IS NULL OR legal_entity_id = sqlc.narg('legal_entity_id'))
    AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
    AND (sqlc.narg('action')::text IS NULL OR action = sqlc.narg('action'))
    AND (sqlc.narg('resource_type')::text IS NULL OR resource_type = sqlc.narg('resource_type'))
    AND (sqlc.narg('resource_id')::text IS NULL OR resource_id = sqlc.narg('resource_id'))
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR occurred_at >= sqlc.narg('from_date'))
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR occurred_at <= sqlc.narg('to_date'))
ORDER BY occurred_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditEvents :one
SELECT count(*) FROM audit.events
WHERE
    (sqlc.narg('legal_entity_id')::uuid IS NULL OR legal_entity_id = sqlc.narg('legal_entity_id'))
    AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
    AND (sqlc.narg('action')::text IS NULL OR action = sqlc.narg('action'))
    AND (sqlc.narg('resource_type')::text IS NULL OR resource_type = sqlc.narg('resource_type'))
    AND (sqlc.narg('resource_id')::text IS NULL OR resource_id = sqlc.narg('resource_id'))
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR occurred_at >= sqlc.narg('from_date'))
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR occurred_at <= sqlc.narg('to_date'));

-- name: ListAuditEventsChronological :many
SELECT * FROM audit.events
ORDER BY occurred_at ASC, id ASC
LIMIT $1 OFFSET $2;
