package domain

import "errors"

var (
	ErrEGSUnitNotFound       = errors.New("egs unit not found")
	ErrCSIDNotFound          = errors.New("active CSID certificate not found for EGS unit")
	ErrDocumentNotFound      = errors.New("zatca document not found")
	ErrSequenceNotFound      = errors.New("invoice sequence not found for EGS unit")
	ErrInvalidVATNumber      = errors.New("invalid VAT registration number: must be 15 digits starting and ending with 3")
	ErrInvalidUUID           = errors.New("invalid invoice UUID")
	ErrInvalidPIH            = errors.New("invalid previous invoice hash: must be valid Base64 SHA-256")
	ErrValidationFailed      = errors.New("zatca document validation failed against BR-KSA business rules")
	ErrDocumentAlreadySigned = errors.New("zatca document is already signed")
	ErrCannotSignUnvalidated = errors.New("cannot sign document that has not passed validation")
	ErrXMLGenerationFailed   = errors.New("failed to generate canonical UBL 2.1 XML")
	ErrCryptographicSigning  = errors.New("failed to digitally sign document using ECDSA certificate")
)
