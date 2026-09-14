package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationStatus string

const (
	OrgStatusActive   OrganizationStatus = "ACTIVE"
	OrgStatusInactive OrganizationStatus = "INACTIVE"
)

type Organization struct {
	ID        uuid.UUID          `json:"id"`
	Code      string             `json:"code"`
	NameAr    string             `json:"name_ar"`
	NameEn    string             `json:"name_en"`
	Status    OrganizationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Version   int64              `json:"version"`
}

type LegalEntity struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Code           string    `json:"code"`
	NameAr         string    `json:"name_ar"`
	NameEn         string    `json:"name_en"`
	CRNumber       string    `json:"cr_number,omitempty"`
	BaseCurrency   string    `json:"base_currency"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Version        int64     `json:"version"`
}

type TaxRegistration struct {
	ID                 uuid.UUID  `json:"id"`
	LegalEntityID      uuid.UUID  `json:"legal_entity_id"`
	TaxType            string     `json:"tax_type"`
	RegistrationNumber string     `json:"registration_number"`
	EffectiveFrom      time.Time  `json:"effective_from"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type Branch struct {
	ID            uuid.UUID `json:"id"`
	LegalEntityID uuid.UUID `json:"legal_entity_id"`
	Code          string    `json:"code"`
	NameAr        string    `json:"name_ar"`
	NameEn        string    `json:"name_en"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int64     `json:"version"`
}

type Department struct {
	ID            uuid.UUID  `json:"id"`
	LegalEntityID uuid.UUID  `json:"legal_entity_id"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	Code          string     `json:"code"`
	NameAr        string     `json:"name_ar"`
	NameEn        string     `json:"name_en"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Version       int64      `json:"version"`
}

type CostCenter struct {
	ID            uuid.UUID  `json:"id"`
	LegalEntityID uuid.UUID  `json:"legal_entity_id"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	Code          string     `json:"code"`
	NameAr        string     `json:"name_ar"`
	NameEn        string     `json:"name_en"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Version       int64      `json:"version"`
}

type Address struct {
	ID               uuid.UUID  `json:"id"`
	LegalEntityID    *uuid.UUID `json:"legal_entity_id,omitempty"`
	BranchID         *uuid.UUID `json:"branch_id,omitempty"`
	AddressType      string     `json:"address_type"`
	BuildingNumber   string     `json:"building_number"`
	StreetNameAr     string     `json:"street_name_ar"`
	StreetNameEn     string     `json:"street_name_en"`
	DistrictAr       string     `json:"district_ar"`
	DistrictEn       string     `json:"district_en"`
	CityAr           string     `json:"city_ar"`
	CityEn           string     `json:"city_en"`
	PostalCode       string     `json:"postal_code"`
	AdditionalNumber string     `json:"additional_number,omitempty"`
	CountryCode      string     `json:"country_code"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ResetInterval string

const (
	ResetNever   ResetInterval = "NEVER"
	ResetYearly  ResetInterval = "YEARLY"
	ResetMonthly ResetInterval = "MONTHLY"
)

type NumberSequence struct {
	ID            uuid.UUID     `json:"id"`
	LegalEntityID uuid.UUID     `json:"legal_entity_id"`
	SequenceCode  string        `json:"sequence_code"`
	Prefix        string        `json:"prefix"`
	Suffix        string        `json:"suffix"`
	NextValue     int64         `json:"next_value"`
	Step          int32         `json:"step"`
	PaddingLength int32         `json:"padding_length"`
	ResetInterval ResetInterval `json:"reset_interval"`
	LastResetYear int32         `json:"last_reset_year"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Version       int64         `json:"version"`
}
