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
	DocumentStatusRetryPending DocumentStatus = "RETRY_PENDING"
	DocumentStatusFailed       DocumentStatus = "FAILED"
)

type Environment string

const (
	EnvironmentSandbox    Environment = "SANDBOX"
	EnvironmentSimulation Environment = "SIMULATION"
	EnvironmentProduction Environment = "PRODUCTION"
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

// ----------------------------------------------------------------------------
// ZATCA FATOORA Phase 2 API Models
// ----------------------------------------------------------------------------

// ClearanceRequest is sent to the Clearance API (B2B Standard Invoices).
type ClearanceRequest struct {
	InvoiceHash string `json:"invoiceHash"`
	UUID        string `json:"uuid"`
	Invoice     string `json:"invoice"` // Base64-encoded signed UBL XML
}

// ClearanceResponse is received from the Clearance API.
type ClearanceResponse struct {
	ClearanceStatus       string              `json:"clearanceStatus"` // 'CLEARED' or 'NOT_CLEARED'
	ClearedInvoice        string              `json:"clearedInvoice,omitempty"` // Base64-encoded UBL with ZATCA clearance stamp
	ValidationResults     *ValidationResult   `json:"validationResults,omitempty"`
	WarningMessages       []ValidationMessage `json:"warningMessages,omitempty"`
	ErrorMessages         []ValidationMessage `json:"errorMessages,omitempty"`
	ZATCARequestID        string              `json:"zatcaRequestId,omitempty"`
}

// ReportingRequest is sent to the Reporting API (B2C Simplified Invoices).
type ReportingRequest struct {
	InvoiceHash string `json:"invoiceHash"`
	UUID        string `json:"uuid"`
	Invoice     string `json:"invoice"` // Base64-encoded signed UBL XML
}

// ReportingResponse is received from the Reporting API.
type ReportingResponse struct {
	ReportingStatus       string              `json:"reportingStatus"` // 'REPORTED' or 'NOT_REPORTED'
	ValidationResults     *ValidationResult   `json:"validationResults,omitempty"`
	WarningMessages       []ValidationMessage `json:"warningMessages,omitempty"`
	ErrorMessages         []ValidationMessage `json:"errorMessages,omitempty"`
	ZATCARequestID        string              `json:"zatcaRequestId,omitempty"`
}

// ComplianceCSIDRequest is sent to issue a Compliance CSID using an OTP.
type ComplianceCSIDRequest struct {
	CSR string `json:"csr"` // Base64-encoded PKCS#10 Certificate Signing Request
	OTP string `json:"otp"` // One-Time-Password from FATOORA portal
}

// ComplianceCSIDResponse contains the issued Compliance CSID certificate and secret.
type ComplianceCSIDResponse struct {
	RequestID           string `json:"requestID"`
	DispositionMessage  string `json:"dispositionMessage"`
	BinarySecurityToken string `json:"binarySecurityToken"` // Base64 X.509 cert
	Secret              string `json:"secret"`
	Errors              []string `json:"errors,omitempty"`
}

// ComplianceCheckRequest is used to validate sample invoices against ZATCA compliance checks.
type ComplianceCheckRequest struct {
	InvoiceHash string `json:"invoiceHash"`
	UUID        string `json:"uuid"`
	Invoice     string `json:"invoice"` // Base64 signed XML
	InvoiceType string `json:"invoiceType"` // 'Standard' or 'Simplified'
}

// ComplianceCheckResponse contains validation results from compliance testing.
type ComplianceCheckResponse struct {
	Status            string              `json:"status"` // 'PASS' or 'FAIL'
	ValidationResults *ValidationResult   `json:"validationResults,omitempty"`
	WarningMessages   []ValidationMessage `json:"warningMessages,omitempty"`
	ErrorMessages     []ValidationMessage `json:"errorMessages,omitempty"`
}

// ProductionCSIDRequest requests the permanent Production CSID after compliance checks pass.
type ProductionCSIDRequest struct {
	ComplianceRequestID string `json:"compliance_request_id"`
}

// ProductionCSIDResponse contains the issued Production CSID and token.
type ProductionCSIDResponse struct {
	RequestID           string `json:"requestID"`
	DispositionMessage  string `json:"dispositionMessage"`
	BinarySecurityToken string `json:"binarySecurityToken"` // Base64 Production X.509 cert
	Secret              string `json:"secret"`
	Errors              []string `json:"errors,omitempty"`
}

// RenewalRequest is used to renew an existing Production CSID.
type RenewalRequest struct {
	CSR string `json:"csr"`
	OTP string `json:"otp"`
}

// RenewalResponse is returned upon certificate renewal.
type RenewalResponse struct {
	RequestID           string `json:"requestID"`
	DispositionMessage  string `json:"dispositionMessage"`
	BinarySecurityToken string `json:"binarySecurityToken"`
	Secret              string `json:"secret"`
}

// ----------------------------------------------------------------------------
// Submissions & Audit
// ----------------------------------------------------------------------------

type ZATCASubmission struct {
	ID                uuid.UUID          `json:"id"`
	DocumentID        uuid.UUID          `json:"document_id"`
	LegalEntityID     uuid.UUID          `json:"legal_entity_id"`
	SubmissionMode    SubmissionMode     `json:"submission_mode"`
	Environment       Environment        `json:"environment"`
	Status            string             `json:"status"` // 'PENDING', 'CLEARED', 'REPORTED', 'REJECTED', 'RETRY_PENDING', 'FAILED'
	RequestPayload    string             `json:"request_payload,omitempty"`
	ResponsePayload   string             `json:"response_payload,omitempty"`
	ZATCARequestID    string             `json:"zatca_request_id,omitempty"`
	ClearedInvoiceXML string             `json:"cleared_invoice_xml,omitempty"`
	WarningMessages   []string           `json:"warning_messages,omitempty"`
	ErrorMessages     []string           `json:"error_messages,omitempty"`
	AttemptsCount     int                `json:"attempts_count"`
	SubmittedAt       time.Time          `json:"submitted_at"`
	CompletedAt       *time.Time         `json:"completed_at,omitempty"`
	Attempts          []SubmissionAttempt `json:"attempts,omitempty"`
}

type SubmissionAttempt struct {
	ID              uuid.UUID `json:"id"`
	SubmissionID    uuid.UUID `json:"submission_id"`
	AttemptNumber   int       `json:"attempt_number"`
	HTTPStatusCode  int       `json:"http_status_code"`
	RequestPayload  string    `json:"request_payload,omitempty"`
	ResponsePayload string    `json:"response_payload,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	AttemptedAt     time.Time `json:"attempted_at"`
}

type SubmissionFilter struct {
	LegalEntityID  uuid.UUID       `json:"legal_entity_id"`
	DocumentID     *uuid.UUID      `json:"document_id,omitempty"`
	SubmissionMode *SubmissionMode `json:"submission_mode,omitempty"`
	Environment    *Environment    `json:"environment,omitempty"`
	Status         *string         `json:"status,omitempty"`
	Limit          int             `json:"limit"`
	Offset         int             `json:"offset"`
}

type OnboardingStatus struct {
	EGSUnitID              uuid.UUID  `json:"egs_unit_id"`
	HasComplianceCSID      bool       `json:"has_compliance_csid"`
	ComplianceChecksPassed bool       `json:"compliance_checks_passed"`
	HasProductionCSID      bool       `json:"has_production_csid"`
	ProductionCSIDExpiresAt *time.Time `json:"production_csid_expires_at,omitempty"`
	CurrentEnvironment     Environment `json:"current_environment"`
	Status                 string     `json:"status"` // 'NOT_ONBOARDED', 'COMPLIANCE_IN_PROGRESS', 'ONBOARDED'
}

type OnboardingResult struct {
	EGSUnitID           uuid.UUID `json:"egs_unit_id"`
	ComplianceRequestID string    `json:"compliance_request_id"`
	ProductionCSIDID    uuid.UUID `json:"production_csid_id"`
	Status              string    `json:"status"`
	Message             string    `json:"message"`
}

