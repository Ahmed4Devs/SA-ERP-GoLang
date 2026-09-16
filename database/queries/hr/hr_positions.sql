-- name: CreatePosition :one
INSERT INTO hr.positions (
    id, legal_entity_id, code, job_id, department_id,
    branch_id, reports_to_position_id, headcount, status,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    now(), now()
)
RETURNING *;

-- name: GetPositionByID :one
SELECT * FROM hr.positions
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetPositionByCode :one
SELECT * FROM hr.positions
WHERE legal_entity_id = $1 AND code = $2;

-- name: ListPositions :many
SELECT * FROM hr.positions
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: UpdatePositionStatus :one
UPDATE hr.positions
SET status = $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
