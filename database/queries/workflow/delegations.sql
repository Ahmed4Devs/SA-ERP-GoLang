-- name: CreateDelegation :one
INSERT INTO workflow.delegations (
    delegator_id,
    delegate_id,
    legal_entity_id,
    start_date,
    end_date,
    is_active,
    reason
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetActiveDelegation :one
SELECT * FROM workflow.delegations
WHERE delegator_id = $1
  AND delegate_id = $2
  AND is_active = true
  AND now() BETWEEN start_date AND end_date
  AND ($3::uuid IS NULL OR legal_entity_id IS NULL OR legal_entity_id = $3)
LIMIT 1;

-- name: ListActiveDelegationsForDelegate :many
SELECT d.*, u.display_name AS delegator_name, u.username AS delegator_username
FROM workflow.delegations d
JOIN iam.users u ON u.id = d.delegator_id
WHERE d.delegate_id = $1
  AND d.is_active = true
  AND now() BETWEEN d.start_date AND d.end_date
ORDER BY d.created_at DESC;

-- name: ListDelegationsByDelegator :many
SELECT d.*, u.display_name AS delegate_name, u.username AS delegate_username
FROM workflow.delegations d
JOIN iam.users u ON u.id = d.delegate_id
WHERE d.delegator_id = $1
ORDER BY d.created_at DESC;

-- name: RevokeDelegation :exec
UPDATE workflow.delegations
SET is_active = false
WHERE id = $1 AND delegator_id = $2;
