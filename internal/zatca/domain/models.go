package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InvoiceType string

const (
	InvoiceTypeStandard   InvoiceType = "STANDARD"   // B2B Tax Invoice (requires Clearance)
	InvoiceTypeSimplified InvoiceType = "SIMPLIFIED" // B2C Simplified Tax Invoice (requires Reporting)
)

type InvoiceSubtype string

const (
	InvoiceSubtypeTaxInvoice InvoiceSubtype = "TAX_INVOICE" // Code 388
	InvoiceSubtypeCreditNote InvoiceSubtype = "CREDIT_NOTE" // Code 381
	InvoiceSubtypeDebitNote  InvoiceSubtype = "DEBIT_NOTE"  // Code 383
)

type SubmissionMode string

const (
	SubmissionModeClearance SubmissionMode = "CLEARANCE"
	SubmissionModeReporting SubmissionMode = "REPORTING"
)

type DocumentStatus string

const (
	DocumentStatusNotGenerated DocumentStatus = "NOT_GENERATED"
	DocumentStatusGenerated    DocumentStatus = "GENERATED"
	DocumentStatusValidated    DocumentStatus = "VALIDATED"
	DocumentStatusSigned       DocumentStatus = "SIGNED"
	DocumentStatusQueued       DocumentStatus = "QUEUED"
	DocumentStatusSubmitting   DocumentStatus = "SUBMITTING"
	DocumentStatusCleared      DocumentStatus = "CLEARED"
	DocumentStatusReported     DocumentStatus = "REPORTED"
	DocumentStatusWarning      DocumentStatus = "WARNING"
	DocumentStatusRejected     DocumentStatus = "REJECTED"
	DocumentStatusFailed       DocumentStatus = "FAILED"
)

type CSIDType string

const (
	CSIDTypeCompliance CSIDType = "COMPLIANCE"
	CSIDTypeProduction CSIDType = "PRODUCTION"
)

// EGSUnit represents an Electronic Generation Solution unit.
type EGSUnit struct {
	ID                   uuid.UUID  `json:"id"`
	LegalEntityID        uuid.UUID  `json:"legal_entity_id"`
	BranchID             *uuid.UUID `json:"branch_id,omitempty"`
	UUID                 uuid.UUID  `json:"uuid"`
	CustomID             string     `json:"custom_id"`
	Model                string     `json:"model"`
	SerialNumber         string     `json:"serial_number"`
	SolutionName         string     `json:"solution_name"`
	VATNumber            string     `json:"vat_number"`
	OrganizationName     string     `json:"organization_name"`
	OrganizationUnitName string     `json:"organization_unit_name"`
	CountryCode          string     `json:"country_code"`
	InvoiceTypeMask      string     `json:"invoice_type_mask"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CSID represents an X.509 Cryptographic Stamp Identifier certificate and ECDSA key pair.
type CSID struct {
	ID                   uuid.UUID  `json:"id"`
	EGSUnitID            uuid.UUID  `json:"egs_unit_id"`
	CSIDType             CSIDType   `json:"csid_type"`
	BinarySecurityToken  string     `json:"binary_security_token"`
	Secret               string     `json:"secret"`
	PrivateKeyPEM        string     `json:"private_key_pem"`
	PublicKeyPEM         string     `json:"public_key_pem"`
	CertificatePEM       string     `json:"certificate_pem"`
	RequestID            string     `json:"request_id,omitempty"`
	ComplianceRequestID  string     `json:"compliance_request_id,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
}

// InvoiceSequence tracks ICV and PIH hash chains per EGS unit and invoice type.
type InvoiceSequence struct {
	ID              uuid.UUID `json:"id"`
	EGSUnitID       uuid.UUID `json:"egs_unit_id"`
	InvoiceType     string    `json:"invoice_type"`
	LastICV         int64     `json:"last_icv"`
	LastInvoiceHash string    `json:"last_invoice_hash"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ZATCADocument represents an electronic tax invoice compliant with Phase 2.
type ZATCADocument struct {
	ID                uuid.UUID      `json:"id"`
	LegalEntityID     uuid.UUID      `json:"legal_entity_id"`
	EGSUnitID         uuid.UUID      `json:"egs_unit_id"`
	SalesInvoiceID    *uuid.UUID     `json:"sales_invoice_id,omitempty"`
	InvoiceUUID       uuid.UUID      `json:"invoice_uuid"`
	InvoiceNumber     string         `json:"invoice_number"`
	InvoiceType       InvoiceType    `json:"invoice_type"`
	InvoiceSubtype    InvoiceSubtype `json:"invoice_subtype"`
	SubmissionMode    SubmissionMode `json:"submission_mode"`
	ICV               int64          `json:"icv"`
	PIH               string         `json:"pih"`
	IssueDate         string         `json:"issue_date"` // YYYY-MM-DD
	IssueTime         string         `json:"issue_time"` // HH:MM:SS
	UBLXML            string         `json:"ubl_xml"`
	SignedXML         string         `json:"signed_xml,omitempty"`
	InvoiceHash       string         `json:"invoice_hash"`
	DigitalSignature  string         `json:"digital_signature,omitempty"`
	PublicKeyHashing  string         `json:"public_key_hashing,omitempty"`
	QRCodeData        string         `json:"qr_code_data"`
	Status            DocumentStatus `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// PartyInfo contains tax details for Seller or Buyer.
type PartyInfo struct {
	Name           string `json:"name"`
	VATNumber      string `json:"vat_number"`
	CRNumber       string `json:"cr_number,omitempty"`
	Street         string `json:"street"`
	BuildingNumber string `json:"building_number"`
	District       string `json:"district"`
	City           string `json:"city"`
	PostalCode     string `json:"postal_code"`
	CountryCode    string `json:"country_code"`
}

// InvoiceLineData represents an invoice line item.
type InvoiceLineData struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Quantity    decimal.Decimal `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Discount    decimal.Decimal `json:"discount"`
	VATRate     decimal.Decimal `json:"vat_rate"`   // e.g. 15.00
	VATAmount   decimal.Decimal `json:"vat_amount"`
	LineTotal   decimal.Decimal `json:"line_total"` // tax inclusive
}

// InvoiceData is the comprehensive data structure passed into the XML Generator.
type InvoiceData struct {
	InvoiceNumber       string            `json:"invoice_number"`
	UUID                uuid.UUID         `json:"uuid"`
	IssueDate           string            `json:"issue_date"` // YYYY-MM-DD
	IssueTime           string            `json:"issue_time"` // HH:MM:SS
	InvoiceType         InvoiceType       `json:"invoice_type"`
	InvoiceSubtype      InvoiceSubtype    `json:"invoice_subtype"`
	CurrencyCode        string            `json:"currency_code"`
	Seller              PartyInfo         `json:"seller"`
	Buyer               PartyInfo         `json:"buyer"`
	Lines               []InvoiceLineData `json:"lines"`
	LineExtensionAmount decimal.Decimal   `json:"line_extension_amount"`
	TaxExclusiveAmount  decimal.Decimal   `json:"tax_exclusive_amount"`
	TaxInclusiveAmount  decimal.Decimal   `json:"tax_inclusive_amount"`
	AllowanceTotal      decimal.Decimal   `json:"allowance_total"`
	TaxTotal            decimal.Decimal   `json:"tax_total"`
	PayableAmount       decimal.Decimal   `json:"payable_amount"`
	ICV                 int64             `json:"icv"`
	PIH                 string            `json:"pih"`
}

// ValidationResult captures BR-KSA rule evaluation outcomes.
type ValidationResult struct {
	ID            uuid.UUID           `json:"id"`
	DocumentID    uuid.UUID           `json:"document_id"`
	OverallStatus string              `json:"overall_status"` // 'PASSED', 'WARNING', 'FAILED'
	RuleVersion   string              `json:"rule_version"`
	ValidatedAt   time.Time           `json:"validated_at"`
	Messages      []ValidationMessage `json:"messages"`
}

type ValidationMessage struct {
	ID          uuid.UUID `json:"id"`
	MessageType string    `json:"message_type"` // 'ERROR', 'WARNING', 'INFO'
	Code        string    `json:"code"`
	Category    string    `json:"category"`
	MessageText string    `json:"message_text"`
	CreatedAt   time.Time `json:"created_at"`
}

type ZATCASummary struct {
	TotalDocuments int64 `json:"total_documents"`
	ClearedCount   int64 `json:"cleared_count"`
	ReportedCount  int64 `json:"reported_count"`
	WarningCount   int64 `json:"warning_count"`
	FailedCount    int64 `json:"failed_count"`
	ActiveEGSCount int   `json:"active_egs_count"`
}
