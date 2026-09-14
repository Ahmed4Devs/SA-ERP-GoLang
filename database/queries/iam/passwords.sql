-- name: SetPassword :exec
INSERT INTO iam.password_credentials (
    user_id,
    password_hash,
    algorithm
) VALUES (
    $1, $2, 'argon2id'
)
ON CONFLICT (user_id) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    algorithm = EXCLUDED.algorithm,
    updated_at = now();

-- name: GetPasswordByUserID :one
SELECT user_id, password_hash, algorithm, created_at, updated_at
FROM iam.password_credentials
WHERE user_id = $1;
