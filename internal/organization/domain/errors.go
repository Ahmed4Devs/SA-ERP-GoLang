package domain

import "errors"

var (
	ErrOrganizationNotFound   = errors.New("organization not found")
	ErrLegalEntityNotFound    = errors.New("legal entity not found")
	ErrBranchNotFound         = errors.New("branch not found")
	ErrDepartmentNotFound     = errors.New("department not found")
	ErrCostCenterNotFound     = errors.New("cost center not found")
	ErrNumberSequenceNotFound = errors.New("number sequence not found")
	ErrInvalidCode            = errors.New("code cannot be empty and must be alphanumeric with dashes or underscores")
	ErrInvalidCRNumber        = errors.New("commercial registration number (CR) must consist of exactly 10 digits")
	ErrInvalidVATNumber       = errors.New("KSA VAT registration number must consist of exactly 15 digits starting and ending with 3")
	ErrInvalidNationalAddress = errors.New("invalid KSA national address parameters")
	ErrDuplicateCode          = errors.New("an entity with this code already exists in the given scope")
)
