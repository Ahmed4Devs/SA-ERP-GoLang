package domain

import "errors"

var (
	ErrWorkflowDefinitionNotFound = errors.New("workflow definition not found or inactive")
	ErrPublishedVersionNotFound   = errors.New("published workflow definition version not found")
	ErrWorkflowInstanceNotFound   = errors.New("workflow instance not found")
	ErrApprovalRequestNotFound    = errors.New("approval request not found")
	ErrMakerCannotApprove         = errors.New("segregation of duties violation: initiator cannot approve their own document")
	ErrUnauthorizedApprover       = errors.New("actor is not authorized to approve this request")
	ErrRequestAlreadyDecided      = errors.New("approval request is already decided")
	ErrSelfDelegationForbidden    = errors.New("delegating approval authority to yourself is forbidden")
	ErrInvalidDelegationPeriod    = errors.New("delegation end date must be after start date")
)
