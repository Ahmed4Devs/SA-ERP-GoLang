package http

import "time"

type StartWorkflowRequest struct {
	DefinitionCode string         `json:"definition_code"`
	LegalEntityID  string         `json:"legal_entity_id"`
	ResourceType   string         `json:"resource_type"`
	ResourceID     string         `json:"resource_id"`
	ContextData    map[string]any `json:"context_data"`
}

type SubmitDecisionRequest struct {
	Decision string `json:"decision"` // APPROVE, REJECT
	Notes    string `json:"notes,omitempty"`
}

type CreateDelegationRequest struct {
	DelegateID    string    `json:"delegate_id"`
	LegalEntityID *string   `json:"legal_entity_id,omitempty"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Reason        string    `json:"reason,omitempty"`
}
