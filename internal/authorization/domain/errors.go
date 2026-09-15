package domain

import "errors"

var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrUnauthorizedAccess  = errors.New("unauthorized: user lacks required permission in specified legal entity scope")
	ErrInvalidScope        = errors.New("invalid authorization scope: non-super admin requires a legal entity scope")
)
