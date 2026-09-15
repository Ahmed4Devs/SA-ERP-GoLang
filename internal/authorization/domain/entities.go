package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	NameAr      string    `json:"name_ar"`
	NameEn      string    `json:"name_en"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	NameAr    string    `json:"name_ar"`
	NameEn    string    `json:"name_en"`
	Module    string    `json:"module"`
	CreatedAt time.Time `json:"created_at"`
}

type RoleAssignment struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              uuid.UUID  `json:"user_id"`
	RoleID              uuid.UUID  `json:"role_id"`
	RoleCode            string     `json:"role_code"`
	RoleNameAr          string     `json:"role_name_ar"`
	RoleNameEn          string     `json:"role_name_en"`
	LegalEntityID       *uuid.UUID `json:"legal_entity_id,omitempty"`
	LegalEntityNameAr   string     `json:"legal_entity_name_ar,omitempty"`
	LegalEntityNameEn   string     `json:"legal_entity_name_en,omitempty"`
	BranchID            *uuid.UUID `json:"branch_id,omitempty"`
	BranchNameAr        string     `json:"branch_name_ar,omitempty"`
	BranchNameEn        string     `json:"branch_name_en,omitempty"`
	AssignedAt          time.Time  `json:"assigned_at"`
}

// Canonical Authorization Contracts per Section 34
type AuthorizationRequest struct {
	UserID        uuid.UUID
	Permission    string
	LegalEntityID *uuid.UUID
	BranchID      *uuid.UUID
	ResourceType  string
	ResourceID    *uuid.UUID
}

type AuthorizationDecision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}
