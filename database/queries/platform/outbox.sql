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
RETURNING id, event_type, aggregate_type, aggregate_id, payload, occurred_at, available_at,
          published_at, attempt_count, last_error, dead_lettered_at;

-- name: FetchPendingOutboxEvents :many
SELECT id, event_type, aggregate_type, aggregate_id, payload, occurred_at, available_at,
       published_at, attempt_count, last_error, dead_lettered_at
FROM platform.outbox_events
WHERE published_at IS NULL AND dead_lettered_at IS NULL AND available_at <= now()
ORDER BY available_at ASC, occurred_at ASC
LIMIT $1;

-- name: ClaimPendingOutboxEvents :many
WITH candidates AS (
    SELECT id
    FROM platform.outbox_events
    WHERE published_at IS NULL AND dead_lettered_at IS NULL AND available_at <= now()
    ORDER BY available_at ASC, occurred_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(batch_size)
)
UPDATE platform.outbox_events AS event
SET available_at = sqlc.arg(lease_until)
FROM candidates
WHERE event.id = candidates.id
RETURNING event.id, event.event_type, event.aggregate_type, event.aggregate_id,
          event.payload, event.occurred_at, event.available_at, event.published_at,
          event.attempt_count, event.last_error, event.dead_lettered_at;

-- name: MarkOutboxEventPublished :exec
UPDATE platform.outbox_events
SET published_at = now(),
    last_error = NULL
WHERE id = $1;

-- name: MarkOutboxEventDeadLettered :exec
UPDATE platform.outbox_events
SET attempt_count = attempt_count + 1,
    last_error = $2,
    dead_lettered_at = now()
WHERE id = $1;

-- name: RecordOutboxEventAttempt :exec
UPDATE platform.outbox_events
SET attempt_count = attempt_count + 1,
    last_error = $2,
    available_at = $3
WHERE id = $1;
