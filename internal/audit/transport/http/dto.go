package http

import "time"

type RecordEventRequest struct {
	LegalEntityID *string `json:"legal_entity_id,omitempty"`
	Action        string  `json:"action"`
	ResourceType  string  `json:"resource_type"`
	ResourceID    string  `json:"resource_id"`
	Payload       any     `json:"payload,omitempty"`
}

type VerificationResponse struct {
	TotalRecords    int64      `json:"total_records"`
	VerifiedRecords int64      `json:"verified_records"`
	IsValid         bool       `json:"is_valid"`
	FailedAtEventID *string    `json:"failed_at_event_id,omitempty"`
	Details         string     `json:"details"`
	VerifiedAt      time.Time  `json:"verified_at"`
}
