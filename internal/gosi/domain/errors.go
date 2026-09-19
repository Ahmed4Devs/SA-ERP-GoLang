package domain

import "errors"

var (
	ErrInsuranceProfileNotFound = errors.New("GOSI insurance profile not found")
	ErrCalculationNotFound      = errors.New("GOSI contribution calculation not found")
	ErrEmployeeScopeMismatch    = errors.New("employee does not belong to the legal entity")
	ErrPeriodAlreadyExported    = errors.New("GOSI period contains calculations that were already exported")
	ErrInvalidPeriod            = errors.New("invalid GOSI calculation period")
)
