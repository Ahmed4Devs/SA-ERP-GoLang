-- name: RecordLoginAttempt :exec
INSERT INTO iam.login_attempts (
    identifier,
    ip_address,
    success,
    attempted_at
) VALUES (
    $1, $2, $3, now()
);

-- name: CountRecentFailedAttempts :one
SELECT COUNT(*)
FROM iam.login_attempts
WHERE identifier = $1 AND success = FALSE AND attempted_at > $2;
