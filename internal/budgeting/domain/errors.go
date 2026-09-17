package domain

import "errors"

var (
	ErrBudgetNotFound          = errors.New("budget not found")
	ErrBudgetCodeExists        = errors.New("budget code already exists for this legal entity")
	ErrBudgetVersionNotFound   = errors.New("budget version not found")
	ErrBudgetLineNotFound      = errors.New("budget line not found")
	ErrInvalidBudgetStatus     = errors.New("invalid budget status transition")
	ErrCannotModifyClosed      = errors.New("cannot modify a closed or superseded budget version")
	ErrInsufficientBudgetFunds = errors.New("insufficient budget funds on source line for reallocation")
	ErrInvalidTransferLines    = errors.New("source and target lines must belong to the same budget version")
	ErrInvalidDimensions       = errors.New("invalid budget line dimensions provided")
	ErrNegativeAmount          = errors.New("budget amount cannot be negative")
)
