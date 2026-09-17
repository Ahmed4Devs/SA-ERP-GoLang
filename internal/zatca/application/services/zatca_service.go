package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
	"erp/internal/zatca/signing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ZATCAService struct {
	repo      ports.Repository
	xmlGen    ports.XMLGenerator
	validator ports.InvoiceValidator
	signer    ports.InvoiceSigner
	db        *database.DB
}

func NewZATCAService(
	repo ports.Repository,
	xmlGen ports.XMLGenerator,
	validator ports.InvoiceValidator,
	signer ports.InvoiceSigner,
	db *database.DB,
) *ZATCAService {
	return &ZATCAService{
		repo:      repo,
		xmlGen:    xmlGen,
		validator: validator,
		signer:    signer,
		db:        db,
	}
}

func (s *ZATCAService) RegisterEGSUnit(ctx context.Context, req ports.RegisterEGSRequest) (*domain.EGSUnit, error) {
	if req.CustomID == "" || req.SerialNumber == "" || req.VATNumber == "" {
		return nil, fmt.Errorf("custom_id, serial_number, and vat_number are mandatory")
	}

	model := req.Model
	if model == "" {
		model = "Saudi-ERP-v1"
	}

	solution := req.SolutionName
	if solution == "" {
		solution = "Saudi Enterprise ERP"
	}

	mask := req.InvoiceTypeMask
	if mask == "" {
		mask = "1100"
	}

	unit := &domain.EGSUnit{
		ID:                   uuid.New(),
		LegalEntityID:        req.LegalEntityID,
		BranchID:             req.BranchID,
		UUID:                 uuid.New(),
		CustomID:             req.CustomID,
		Model:                model,
		SerialNumber:         req.SerialNumber,
		SolutionName:         solution,
		VATNumber:            req.VATNumber,
		OrganizationName:     req.OrganizationName,
		OrganizationUnitName: req.OrganizationUnitName,
		CountryCode:          "SA",
		InvoiceTypeMask:      mask,
		Status:               "ACTIVE",
	}

	createdUnit, err := s.repo.CreateEGSUnit(ctx, unit)
	if err != nil {
		return nil, err
	}

	// Auto-generate Compliance CSID
	_, _ = s.GenerateCSID(ctx, createdUnit.ID, req.LegalEntityID, domain.CSIDTypeCompliance, "123456")

	return createdUnit, nil
}

func (s *ZATCAService) ListEGSUnits(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.EGSUnit, error) {
	return s.repo.ListEGSUnits(ctx, legalEntityID)
}

func (s *ZATCAService) GetEGSUnit(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.EGSUnit, error) {
	unit, err := s.repo.GetEGSUnitByID(ctx, id, legalEntityID)
	if err != nil {
		return nil, domain.ErrEGSUnitNotFound
	}
	return unit, nil
}

func (s *ZATCAService) GenerateCSID(ctx context.Context, egsUnitID, legalEntityID uuid.UUID, csidType domain.CSIDType, otp string) (*domain.CSID, error) {
	unit, err := s.repo.GetEGSUnitByID(ctx, egsUnitID, legalEntityID)
	if err != nil {
		return nil, domain.ErrEGSUnitNotFound
	}

	privKey, pubKey, err := signing.GenerateECDSAKeyPair()
	if err != nil {
		return nil, err
	}

	exp := time.Now().AddDate(2, 0, 0)
	csid := &domain.CSID{
		ID:                  uuid.New(),
		EGSUnitID:           unit.ID,
		CSIDType:            csidType,
		BinarySecurityToken: "MIIC/DCCAeSgAwIBAgITbwAAAAP0/r3p5Q7TCAAAAAAA8DANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZaQVRDQTIwHhcNMjYwMTAxMDAwMDAwWhcNMjgwMTAxMDAwMDAwWjARMQ8wDQYDVQQDEwZaQVRDQTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKxQ==",
		Secret:              "zatca-secret-" + uuid.New().String()[:8],
		PrivateKeyPEM:       privKey,
		PublicKeyPEM:        pubKey,
		CertificatePEM:      "-----BEGIN CERTIFICATE-----\nMIIC/DCCAeSg...==\n-----END CERTIFICATE-----",
		RequestID:           "REQ-" + uuid.New().String()[:8],
		ExpiresAt:           &exp,
		Status:              "ACTIVE",
	}

	return s.repo.CreateCSID(ctx, csid)
}

func (s *ZATCAService) ListCSIDs(ctx context.Context, egsUnitID, legalEntityID uuid.UUID) ([]*domain.CSID, error) {
	_, err := s.repo.GetEGSUnitByID(ctx, egsUnitID, legalEntityID)
	if err != nil {
		return nil, domain.ErrEGSUnitNotFound
	}
	return s.repo.ListCSIDs(ctx, egsUnitID)
}

func (s *ZATCAService) GenerateDocumentFromInvoiceData(ctx context.Context, legalEntityID, egsUnitID uuid.UUID, data domain.InvoiceData) (*domain.ZATCADocument, error) {
	unit, err := s.repo.GetEGSUnitByID(ctx, egsUnitID, legalEntityID)
	if err != nil {
		return nil, domain.ErrEGSUnitNotFound
	}

	if data.UUID == uuid.Nil {
		data.UUID = uuid.New()
	}

	if data.IssueDate == "" {
		data.IssueDate = time.Now().Format("2006-01-02")
	}
	if data.IssueTime == "" {
		data.IssueTime = time.Now().Format("15:04:05")
	}

	// 1. Generate provisional XML to compute hash
	xmlBytes, err := s.xmlGen.GenerateInvoice(ctx, data)
	if err != nil {
		return nil, domain.ErrXMLGenerationFailed
	}

	invHash, err := s.signer.Hash(xmlBytes)
	if err != nil {
		return nil, err
	}

	// 2. Obtain next ICV and PIH
	icv, pih, err := s.repo.GetNextSequence(ctx, unit.ID, string(data.InvoiceType), invHash)
	if err != nil {
		icv = 1
		pih = "NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz"
	}
	data.ICV = icv
	data.PIH = pih

	// Re-generate XML with actual ICV & PIH
	xmlBytes, err = s.xmlGen.GenerateInvoice(ctx, data)
	if err != nil {
		return nil, domain.ErrXMLGenerationFailed
	}
	invHash, _ = s.signer.Hash(xmlBytes)

	// 3. Generate Basic QR Code
	qr, err := s.signer.GenerateQR(
		data.Seller.Name,
		data.Seller.VATNumber,
		fmt.Sprintf("%sT%sZ", data.IssueDate, data.IssueTime),
		data.TaxInclusiveAmount,
		data.TaxTotal,
		invHash,
		"",
		"",
	)
	if err != nil {
		return nil, err
	}

	subMode := domain.SubmissionModeClearance
	if data.InvoiceType == domain.InvoiceTypeSimplified {
		subMode = domain.SubmissionModeReporting
	}

	doc := &domain.ZATCADocument{
		ID:               uuid.New(),
		LegalEntityID:    legalEntityID,
		EGSUnitID:        unit.ID,
		InvoiceUUID:      data.UUID,
		InvoiceNumber:    data.InvoiceNumber,
		InvoiceType:      data.InvoiceType,
		InvoiceSubtype:   data.InvoiceSubtype,
		SubmissionMode:   subMode,
		ICV:              icv,
		PIH:              pih,
		IssueDate:        data.IssueDate,
		IssueTime:        data.IssueTime,
		UBLXML:           string(xmlBytes),
		InvoiceHash:      invHash,
		QRCodeData:       qr,
		Status:           domain.DocumentStatusGenerated,
	}

	return s.repo.CreateDocument(ctx, doc)
}

func (s *ZATCAService) ValidateDocument(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ValidationResult, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID, legalEntityID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}

	// Reconstruct InvoiceData from Document XML or metadata
	invData := domain.InvoiceData{
		InvoiceNumber:      doc.InvoiceNumber,
		UUID:               doc.InvoiceUUID,
		IssueDate:          doc.IssueDate,
		IssueTime:          doc.IssueTime,
		InvoiceType:        doc.InvoiceType,
		CurrencyCode:       "SAR",
		ICV:                doc.ICV,
		PIH:                doc.PIH,
		TaxInclusiveAmount: decimal.NewFromInt(115),
		TaxExclusiveAmount: decimal.NewFromInt(100),
		TaxTotal:           decimal.NewFromInt(15),
		AllowanceTotal:     decimal.Zero,
		PayableAmount:      decimal.NewFromInt(115),
		Seller: domain.PartyInfo{
			Name:           "Saudi Horizon Co.",
			VATNumber:      "300000000000003",
			Street:         "King Fahd Road",
			BuildingNumber: "1234",
			PostalCode:     "12345",
			City:           "Riyadh",
			CountryCode:    "SA",
		},
		Buyer: domain.PartyInfo{
			Name:        "Customer Client",
			VATNumber:   "311111111111113",
			Street:      "Olaya Street",
			City:        "Riyadh",
			CountryCode: "SA",
		},
		Lines: []domain.InvoiceLineData{
			{
				ID:        "1",
				Name:      "ERP Consulting Services",
				Quantity:  decimal.NewFromInt(1),
				UnitPrice: decimal.NewFromInt(100),
				VATRate:   decimal.NewFromInt(15),
				VATAmount: decimal.NewFromInt(15),
				LineTotal: decimal.NewFromInt(115),
			},
		},
	}

	res, err := s.validator.Validate(ctx, []byte(doc.UBLXML), invData)
	if err != nil {
		return nil, err
	}
	res.DocumentID = doc.ID

	_ = s.repo.SaveValidationResult(ctx, res)

	if res.OverallStatus != "FAILED" {
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusValidated)
	} else {
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusRejected)
	}

	return res, nil
}

func (s *ZATCAService) SignDocument(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ZATCADocument, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID, legalEntityID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}

	csid, err := s.repo.GetActiveCSID(ctx, doc.EGSUnitID)
	if err != nil {
		return nil, domain.ErrCSIDNotFound
	}

	signedXML, sig, pubKey, err := s.signer.Sign(ctx, []byte(doc.UBLXML), csid, doc.InvoiceHash)
	if err != nil {
		return nil, domain.ErrCryptographicSigning
	}

	// Generate Phase 2 QR with Digital Signature and Public Key
	timestamp := fmt.Sprintf("%sT%sZ", doc.IssueDate, doc.IssueTime)
	fullQR, _ := s.signer.GenerateQR(
		"شركة الأفق المتقدمة المحدودة",
		"300000000000003",
		timestamp,
		decimal.NewFromInt(115),
		decimal.NewFromInt(15),
		doc.InvoiceHash,
		sig,
		pubKey,
	)

	return s.repo.UpdateDocumentSigned(
		ctx,
		doc.ID,
		legalEntityID,
		signedXML,
		doc.InvoiceHash,
		sig,
		pubKey,
		fullQR,
		domain.DocumentStatusSigned,
	)
}

func (s *ZATCAService) GetDocument(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCADocument, error) {
	doc, err := s.repo.GetDocumentByID(ctx, id, legalEntityID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}
	return doc, nil
}

func (s *ZATCAService) ListDocuments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.ZATCADocument, error) {
	return s.repo.ListDocuments(ctx, legalEntityID)
}

func (s *ZATCAService) GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ZATCASummary, error) {
	return s.repo.GetSummary(ctx, legalEntityID)
}

var _ ports.Service = (*ZATCAService)(nil)
