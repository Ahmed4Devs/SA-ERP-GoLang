package domain

import "errors"

var (
	ErrWorkflowDefinitionNotFound = errors.New("workflow definition not found or inactive")
	ErrPublishedVersionNotFound   = errors.New("published workflow definition version not found")
	ErrWorkflowInstanceNotFound   = errors.New("workflow instance not found")
	ErrWorkflowStepNotFound       = errors.New("workflow step not found")
	ErrApprovalRequestNotFound    = errors.New("approval request not found")
	ErrMakerCannotApprove         = errors.New("segregation of duties violation: initiator cannot approve their own document")
	ErrUnauthorizedApprover       = errors.New("actor is not authorized to approve this request")
	ErrRequestAlreadyDecided      = errors.New("approval request is already decided")
	ErrInvalidDecision            = errors.New("decision must be APPROVE or REJECT")
	ErrInvalidWorkflowCommand     = errors.New("workflow command contains missing or invalid identifiers")
	ErrNoEligibleWorkflowStep     = errors.New("workflow has no eligible step for the supplied context")
	ErrNoApproversAvailable       = errors.New("workflow step has no available approvers")
	ErrSelfDelegationForbidden    = errors.New("delegating approval authority to yourself is forbidden")
	ErrInvalidDelegationPeriod    = errors.New("delegation end date must be after start date")
)
