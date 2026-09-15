-- name: GetWorkflowDefinitionByCode :one
SELECT * FROM workflow.definitions
WHERE code = $1 AND is_active = true;

-- name: GetWorkflowDefinitionByID :one
SELECT * FROM workflow.definitions
WHERE id = $1;

-- name: ListWorkflowDefinitions :many
SELECT * FROM workflow.definitions
ORDER BY code ASC;

-- name: GetPublishedVersionByDefinitionID :one
SELECT * FROM workflow.definition_versions
WHERE definition_id = $1 AND status = 'PUBLISHED'
ORDER BY version DESC
LIMIT 1;

-- name: GetDefinitionVersionByID :one
SELECT * FROM workflow.definition_versions
WHERE id = $1;

-- name: ListStepsByVersionID :many
SELECT * FROM workflow.steps
WHERE version_id = $1
ORDER BY step_order ASC;

-- name: ListStepConditionsByStepID :many
SELECT * FROM workflow.step_conditions
WHERE step_id = $1;

-- name: ListStepApproversByStepID :many
SELECT * FROM workflow.step_approvers
WHERE step_id = $1;

-- name: CreateWorkflowDefinition :one
INSERT INTO workflow.definitions (
    code,
    name_ar,
    name_en,
    module,
    is_active
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: CreateDefinitionVersion :one
INSERT INTO workflow.definition_versions (
    definition_id,
    version,
    status,
    published_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: CreateWorkflowStep :one
INSERT INTO workflow.steps (
    version_id,
    step_order,
    code,
    name_ar,
    name_en,
    approval_type,
    timeout_hours
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: CreateStepCondition :one
INSERT INTO workflow.step_conditions (
    step_id,
    field_name,
    operator,
    field_value
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: CreateStepApprover :one
INSERT INTO workflow.step_approvers (
    step_id,
    approver_type,
    role_id,
    user_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;
