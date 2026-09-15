package domain

import "errors"

var (
	ErrEventNotFound       = errors.New("audit event not found")
	ErrHashMismatch        = errors.New("audit hash chain integrity mismatch")
	ErrAppendOnlyViolation = errors.New("audit history is strictly append-only; mutations are forbidden")
)
