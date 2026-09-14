-- name: CreateNumberSequence :one
INSERT INTO organization.number_sequences (
    legal_entity_id,
    sequence_code,
    prefix,
    suffix,
    next_value,
    step,
    padding_length,
    reset_interval,
    last_reset_year
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetNumberSequenceByEntityAndCode :one
SELECT * FROM organization.number_sequences
WHERE legal_entity_id = $1 AND sequence_code = $2;

-- name: GetNumberSequenceForUpdate :one
SELECT * FROM organization.number_sequences
WHERE legal_entity_id = $1 AND sequence_code = $2
FOR UPDATE;

-- name: IncrementNumberSequence :one
UPDATE organization.number_sequences
SET
    next_value = next_value + step,
    last_reset_year = $3,
    updated_at = now(),
    version = version + 1
WHERE legal_entity_id = $1 AND sequence_code = $2
RETURNING *;

-- name: ResetAndIncrementNumberSequence :one
UPDATE organization.number_sequences
SET
    next_value = 1 + step,
    last_reset_year = $3,
    updated_at = now(),
    version = version + 1
WHERE legal_entity_id = $1 AND sequence_code = $2
RETURNING *;

-- name: ListNumberSequencesByEntity :many
SELECT * FROM organization.number_sequences
WHERE legal_entity_id = $1
ORDER BY sequence_code ASC;
