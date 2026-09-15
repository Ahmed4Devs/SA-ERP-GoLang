-- name: InsertEntityChange :one
INSERT INTO audit.entity_changes (
    event_id,
    table_name,
    record_id,
    before_state,
    after_state
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListEntityChangesByEventID :many
SELECT * FROM audit.entity_changes
WHERE event_id = $1
ORDER BY created_at ASC;
