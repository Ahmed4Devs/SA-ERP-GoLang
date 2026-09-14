package http

type CreateOrganizationRequest struct {
	Code   string `json:"code"`
	NameAr string `json:"name_ar"`
	NameEn string `json:"name_en"`
}

type CreateLegalEntityRequest struct {
	OrganizationID string `json:"organization_id"`
	Code           string `json:"code"`
	NameAr         string `json:"name_ar"`
	NameEn         string `json:"name_en"`
	CRNumber       string `json:"cr_number"`
	BaseCurrency   string `json:"base_currency"`
}

type CreateBranchRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	Code          string `json:"code"`
	NameAr        string `json:"name_ar"`
	NameEn        string `json:"name_en"`
}

type CreateDepartmentRequest struct {
	LegalEntityID string  `json:"legal_entity_id"`
	ParentID      *string `json:"parent_id,omitempty"`
	Code          string  `json:"code"`
	NameAr        string  `json:"name_ar"`
	NameEn        string  `json:"name_en"`
}

type CreateCostCenterRequest struct {
	LegalEntityID string  `json:"legal_entity_id"`
	ParentID      *string `json:"parent_id,omitempty"`
	Code          string  `json:"code"`
	NameAr        string  `json:"name_ar"`
	NameEn        string  `json:"name_en"`
}

type DefineSequenceRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	Code          string `json:"code"`
	Prefix        string `json:"prefix"`
	Suffix        string `json:"suffix"`
	StartValue    int64  `json:"start_value"`
	Step          int32  `json:"step"`
	PaddingLength int32  `json:"padding_length"`
	ResetInterval string `json:"reset_interval"`
}

type GenerateNextSequenceRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	SequenceCode  string `json:"sequence_code"`
}
