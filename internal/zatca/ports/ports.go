package ports

import (
	"context"

	"erp/internal/zatca/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type RegisterEGSRequest struct {
	LegalEntityID        uuid.UUID  `json:"legal_entity_id"`
	BranchID             *uuid.UUID `json:"branch_id,omitempty"`
	CustomID             string     `json:"custom_id"`
	Model                string     `json:"model"`
	SerialNumber         string     `json:"serial_number"`
	SolutionName         string     `json:"solution_name"`
	VATNumber            string     `json:"vat_number"`
	OrganizationName     string     `json:"organization_name"`
	OrganizationUnitName string     `json:"organization_unit_name"`
	InvoiceTypeMask      string     `json:"invoice_type_mask"`
}

type XMLGenerator interface {
	GenerateInvoice(ctx context.Context, input domain.InvoiceData) ([]byte, error)
}

type InvoiceValidator interface {
	Validate(ctx context.Context, xmlBytes []byte, data domain.InvoiceData) (*domain.ValidationResult, error)
}

type InvoiceSigner interface {
	Hash(xmlBytes []byte) (string, error)
	Sign(ctx context.Context, xmlBytes []byte, csid *domain.CSID, invoiceHash string) (signedXML string, digitalSig string, pubKeyHash string, err error)
	GenerateQR(sellerName, vatNumber, timestamp string, totalAmount, vatAmount decimal.Decimal, invoiceHash, signature, publicKey string) (string, error)
}

type Repository interface {
	GetEGSUnitByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.EGSUnit, error)
	GetEGSUnitByCustomID(ctx context.Context, customID string, legalEntityID uuid.UUID) (*domain.EGSUnit, error)
	ListEGSUnits(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.EGSUnit, error)
	CreateEGSUnit(ctx context.Context, unit *domain.EGSUnit) (*domain.EGSUnit, error)
	UpdateEGSUnitStatus(ctx context.Context, id, legalEntityID uuid.UUID, status string) (*domain.EGSUnit, error)

	GetActiveCSID(ctx context.Context, egsUnitID uuid.UUID) (*domain.CSID, error)
	ListCSIDs(ctx context.Context, egsUnitID uuid.UUID) ([]*domain.CSID, error)
	CreateCSID(ctx context.Context, csid *domain.CSID) (*domain.CSID, error)

	GetNextSequence(ctx context.Context, egsUnitID uuid.UUID, invoiceType string, currentHash string) (nextICV int64, pih string, err error)

	GetDocumentByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCADocument, error)
	GetDocumentByUUID(ctx context.Context, u uuid.UUID) (*domain.ZATCADocument, error)
	ListDocuments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.ZATCADocument, error)
	CreateDocument(ctx context.Context, doc *domain.ZATCADocument) (*domain.ZATCADocument, error)
	UpdateDocumentSigned(ctx context.Context, id, legalEntityID uuid.UUID, signedXML, hash, signature, pubKey, qr string, status domain.DocumentStatus) (*domain.ZATCADocument, error)
	UpdateDocumentStatus(ctx context.Context, id, legalEntityID uuid.UUID, status domain.DocumentStatus) (*domain.ZATCADocument, error)

	SaveValidationResult(ctx context.Context, res *domain.ValidationResult) error
	GetValidationResults(ctx context.Context, documentID uuid.UUID) ([]*domain.ValidationResult, error)

	GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ZATCASummary, error)
}

type Service interface {
	RegisterEGSUnit(ctx context.Context, req RegisterEGSRequest) (*domain.EGSUnit, error)
	ListEGSUnits(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.EGSUnit, error)
	GetEGSUnit(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.EGSUnit, error)

	GenerateCSID(ctx context.Context, egsUnitID, legalEntityID uuid.UUID, csidType domain.CSIDType, otp string) (*domain.CSID, error)
	ListCSIDs(ctx context.Context, egsUnitID, legalEntityID uuid.UUID) ([]*domain.CSID, error)

	GenerateDocumentFromInvoiceData(ctx context.Context, legalEntityID, egsUnitID uuid.UUID, data domain.InvoiceData) (*domain.ZATCADocument, error)
	ValidateDocument(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ValidationResult, error)
	SignDocument(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ZATCADocument, error)

	GetDocument(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCADocument, error)
	ListDocuments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.ZATCADocument, error)
	GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ZATCASummary, error)
}
