-- name: CreateJob :one
INSERT INTO hr.jobs (
    id, legal_entity_id, code, title_en, title_ar,
    job_family, job_level, description, is_active,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    now(), now()
)
RETURNING *;

-- name: GetJobByID :one
SELECT * FROM hr.jobs
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetJobByCode :one
SELECT * FROM hr.jobs
WHERE legal_entity_id = $1 AND code = $2;

-- name: ListJobs :many
SELECT * FROM hr.jobs
WHERE legal_entity_id = $1
ORDER BY code ASC;
