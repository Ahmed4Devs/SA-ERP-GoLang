-- name: GetIdempotencyKey :one
SELECT id, scope, idempotency_key, request_hash, state, response_status, response_body, created_at, completed_at, expires_at
FROM platform.idempotency_keys
WHERE scope = $1 AND idempotency_key = $2;

-- name: InsertIdempotencyKey :one
INSERT INTO platform.idempotency_keys (
    scope,
    idempotency_key,
    request_hash,
    state,
    expires_at
) VALUES (
    $1, $2, $3, 'PROCESSING', $4
)
RETURNING id, scope, idempotency_key, request_hash, state, response_status, response_body, created_at, completed_at, expires_at;

-- name: CompleteIdempotencyKey :exec
UPDATE platform.idempotency_keys
SET state = 'COMPLETED',
    response_status = $3,
    response_body = $4,
    completed_at = now()
WHERE scope = $1 AND idempotency_key = $2;

-- name: FailIdempotencyKey :exec
UPDATE platform.idempotency_keys
SET state = 'FAILED',
    completed_at = now()
WHERE scope = $1 AND idempotency_key = $2;
