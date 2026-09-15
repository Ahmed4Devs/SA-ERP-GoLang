package http

type AssignRoleRequest struct {
	UserID        string  `json:"user_id"`
	RoleID        string  `json:"role_id"`
	LegalEntityID *string `json:"legal_entity_id,omitempty"`
	BranchID      *string `json:"branch_id,omitempty"`
}

type RevokeRoleRequest struct {
	UserID        string  `json:"user_id"`
	RoleID        string  `json:"role_id"`
	LegalEntityID *string `json:"legal_entity_id,omitempty"`
	BranchID      *string `json:"branch_id,omitempty"`
}

type CheckPermissionRequest struct {
	Permission    string  `json:"permission"`
	LegalEntityID *string `json:"legal_entity_id,omitempty"`
}
