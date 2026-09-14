-- name: InsertOutboxEvent :one
INSERT INTO platform.outbox_events (
    event_type,
    aggregate_type,
    aggregate_id,
    payload,
    occurred_at,
    available_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, event_type, aggregate_type, aggregate_id, payload, occurred_at, available_at, published_at, attempt_count, last_error;

-- name: FetchPendingOutboxEvents :many
SELECT id, event_type, aggregate_type, aggregate_id, payload, occurred_at, available_at, published_at, attempt_count, last_error
FROM platform.outbox_events
WHERE published_at IS NULL AND available_at <= now()
ORDER BY available_at ASC, occurred_at ASC
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE platform.outbox_events
SET published_at = now()
WHERE id = $1;

-- name: RecordOutboxEventAttempt :exec
UPDATE platform.outbox_events
SET attempt_count = attempt_count + 1,
    last_error = $2,
    available_at = $3
WHERE id = $1;
