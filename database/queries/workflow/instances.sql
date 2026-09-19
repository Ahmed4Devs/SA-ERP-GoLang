-- name: CreateWorkflowInstance :one
INSERT INTO workflow.instances (
    definition_id,
    version_id,
    legal_entity_id,
    resource_type,
    resource_id,
    initiator_id,
    status,
    context_data
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetWorkflowInstanceByID :one
SELECT * FROM workflow.instances
WHERE id = $1;

-- name: GetWorkflowInstanceByResource :one
SELECT * FROM workflow.instances
WHERE resource_type = $1 AND resource_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateWorkflowInstanceStatus :one
UPDATE workflow.instances
SET status = $2,
    completed_at = $3
WHERE id = $1
RETURNING *;

-- name: CreateInstanceStep :one
INSERT INTO workflow.instance_steps (
    instance_id,
    step_id,
    step_order,
    status
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetInstanceStepByID :one
SELECT * FROM workflow.instance_steps
WHERE id = $1;

-- name: LockInstanceStepByID :one
SELECT * FROM workflow.instance_steps
WHERE id = $1
FOR UPDATE;

-- name: ListInstanceStepsByInstanceID :many
SELECT * FROM workflow.instance_steps
WHERE instance_id = $1
ORDER BY step_order ASC;

-- name: UpdateInstanceStepStatus :one
UPDATE workflow.instance_steps
SET status = $2,
    completed_at = $3
WHERE id = $1
RETURNING *;

-- name: CreateApprovalRequest :one
INSERT INTO workflow.approval_requests (
    instance_step_id,
    assigned_user_id,
    status
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetApprovalRequestByID :one
SELECT * FROM workflow.approval_requests
WHERE id = $1;

-- name: LockApprovalRequestByID :one
SELECT * FROM workflow.approval_requests
WHERE id = $1
FOR UPDATE;

-- name: ListApprovalRequestsByInstanceStep :many
SELECT * FROM workflow.approval_requests
WHERE instance_step_id = $1
ORDER BY created_at ASC;

-- name: CancelPendingApprovalRequestsForStep :exec
UPDATE workflow.approval_requests
SET status = 'CANCELLED'
WHERE instance_step_id = $1 AND status = 'PENDING';

-- name: ListPendingApprovalRequestsForUser :many
SELECT ar.id,
       ar.instance_step_id,
       ar.assigned_user_id,
       ar.status,
       ar.created_at,
       inst.id AS instance_id,
       inst.resource_type,
       inst.resource_id,
       inst.legal_entity_id,
       inst.initiator_id,
       inst.context_data,
       st.name_ar AS step_name_ar,
       st.name_en AS step_name_en,
       u.display_name AS initiator_name
FROM workflow.approval_requests ar
JOIN workflow.instance_steps inst_step ON inst_step.id = ar.instance_step_id
JOIN workflow.instances inst ON inst.id = inst_step.instance_id
JOIN workflow.steps st ON st.id = inst_step.step_id
JOIN iam.users u ON u.id = inst.initiator_id
WHERE ar.assigned_user_id = $1 AND ar.status = 'PENDING'
ORDER BY ar.created_at DESC;

-- name: UpdateApprovalRequestStatus :one
UPDATE workflow.approval_requests
SET status = $2
WHERE id = $1
RETURNING *;

-- name: CreateApprovalDecision :one
INSERT INTO workflow.approval_decisions (
    request_id,
    instance_id,
    actor_id,
    decision,
    notes
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: ListApprovalDecisionsByInstanceID :many
SELECT ad.id,
       ad.request_id,
       ad.instance_id,
       ad.actor_id,
       ad.decision,
       ad.notes,
       ad.decided_at,
       u.display_name AS actor_name,
       u.username AS actor_username
FROM workflow.approval_decisions ad
JOIN iam.users u ON u.id = ad.actor_id
WHERE ad.instance_id = $1
ORDER BY ad.decided_at ASC;

-- name: ListUserIDsByRoleInScope :many
SELECT DISTINCT ura.user_id
FROM authz.user_role_assignments ura
JOIN iam.users u ON u.id = ura.user_id
WHERE ura.role_id = $1
  AND u.status = 'ACTIVE'
  AND (ura.legal_entity_id IS NULL OR ura.legal_entity_id = $2);

-- name: GetFallbackAdminUserID :one
SELECT u.id
FROM iam.users u
JOIN authz.user_role_assignments ura ON ura.user_id = u.id
JOIN authz.roles r ON r.id = ura.role_id
WHERE u.status = 'ACTIVE'
  AND r.code = 'SUPER_ADMIN'
  AND ura.legal_entity_id IS NULL
  AND ura.branch_id IS NULL
ORDER BY u.created_at ASC
LIMIT 1;
