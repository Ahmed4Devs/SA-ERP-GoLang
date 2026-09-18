package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
	"erp/internal/zatca/signing"
	ublxml "erp/internal/zatca/xml"

	"github.com/google/uuid"
)

type ZATCAService struct {
	repo                      ports.Repository
	xmlGen                    ports.XMLGenerator
	validator                 ports.InvoiceValidator
	signer                    ports.InvoiceSigner
	db                        *database.DB
	allowSimulatedCredentials bool
}

func NewZATCAService(
	repo ports.Repository,
	xmlGen ports.XMLGenerator,
	validator ports.InvoiceValidator,
	signer ports.InvoiceSigner,
	db *database.DB,
	allowSimulatedCredentials bool,
) *ZATCAService {
	return &ZATCAService{
		repo:                      repo,
		xmlGen:                    xmlGen,
		validator:                 validator,
		signer:                    signer,
		db:                        db,
		allowSimulatedCredentials: allowSimulatedCredentials,
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

	status := "PENDING_OTP"
	if s.allowSimulatedCredentials {
		status = "ACTIVE"
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
		Status:               status,
	}

	createdUnit, err := s.repo.CreateEGSUnit(ctx, unit)
	if err != nil {
		return nil, err
	}

	// Local credentials exist only in the explicit simulation environment. Real
	// environments must complete the OTP onboarding flow against FATOORA.
	if s.allowSimulatedCredentials {
		if _, err := s.GenerateCSID(ctx, createdUnit.ID, req.LegalEntityID, domain.CSIDTypeCompliance, "simulation"); err != nil {
			return nil, fmt.Errorf("failed to create simulation CSID: %w", err)
		}
	}

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
	if !s.allowSimulatedCredentials {
		return nil, fmt.Errorf("local CSID generation is disabled; use the ZATCA onboarding endpoint")
	}
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

	var created *domain.ZATCADocument
	txManager := transaction.NewManager(s.db.Pool)
	if err := txManager.WithinTx(ctx, func(txCtx context.Context) error {
		// Reserving the counter, calculating the final document hash, updating
		// the chain, and storing the document are one atomic operation. The row
		// update serializes concurrent invoices for the same EGS/type sequence.
		icv, pih, err := s.repo.ReserveNextSequence(txCtx, unit.ID, string(data.InvoiceType))
		if err != nil {
			return err
		}
		data.ICV = icv
		data.PIH = pih

		xmlBytes, err := s.xmlGen.GenerateInvoice(txCtx, data)
		if err != nil {
			return domain.ErrXMLGenerationFailed
		}
		invHash, err := s.signer.Hash(xmlBytes)
		if err != nil {
			return fmt.Errorf("failed to hash final UBL XML: %w", err)
		}
		if err := s.repo.SetSequenceHash(txCtx, unit.ID, string(data.InvoiceType), icv, invHash); err != nil {
			return err
		}

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
			return err
		}

		subMode := domain.SubmissionModeClearance
		if data.InvoiceType == domain.InvoiceTypeSimplified {
			subMode = domain.SubmissionModeReporting
		}

		created, err = s.repo.CreateDocument(txCtx, &domain.ZATCADocument{
			ID:             uuid.New(),
			LegalEntityID:  legalEntityID,
			EGSUnitID:      unit.ID,
			InvoiceUUID:    data.UUID,
			InvoiceNumber:  data.InvoiceNumber,
			InvoiceType:    data.InvoiceType,
			InvoiceSubtype: data.InvoiceSubtype,
			SubmissionMode: subMode,
			ICV:            icv,
			PIH:            pih,
			IssueDate:      data.IssueDate,
			IssueTime:      data.IssueTime,
			UBLXML:         string(xmlBytes),
			InvoiceHash:    invHash,
			QRCodeData:     qr,
			Status:         domain.DocumentStatusGenerated,
		})
		return err
	}); err != nil {
		return nil, fmt.Errorf("failed to create chained ZATCA document: %w", err)
	}
	return created, nil
}

func (s *ZATCAService) ValidateDocument(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ValidationResult, error) {
	doc, err := s.repo.GetDocumentByID(ctx, documentID, legalEntityID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}

	invData, err := ublxml.ParseInvoiceData([]byte(doc.UBLXML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse stored UBL document: %w", err)
	}
	if invData.InvoiceNumber != doc.InvoiceNumber || invData.UUID != doc.InvoiceUUID ||
		invData.ICV != doc.ICV || invData.PIH != doc.PIH {
		return nil, fmt.Errorf("stored UBL metadata does not match document record")
	}

	res, err := s.validator.Validate(ctx, []byte(doc.UBLXML), invData)
	if err != nil {
		return nil, err
	}
	res.DocumentID = doc.ID

	if err := s.repo.SaveValidationResult(ctx, res); err != nil {
		return nil, fmt.Errorf("failed to save ZATCA validation result: %w", err)
	}

	if res.OverallStatus != "FAILED" {
		if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusValidated); err != nil {
			return nil, fmt.Errorf("failed to mark ZATCA document validated: %w", err)
		}
	} else {
		if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusRejected); err != nil {
			return nil, fmt.Errorf("failed to mark ZATCA document rejected: %w", err)
		}
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

	invData, err := ublxml.ParseInvoiceData([]byte(doc.UBLXML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse stored UBL document for QR generation: %w", err)
	}
	computedHash, err := s.signer.Hash([]byte(doc.UBLXML))
	if err != nil {
		return nil, fmt.Errorf("failed to hash stored UBL document: %w", err)
	}
	if computedHash != doc.InvoiceHash {
		return nil, fmt.Errorf("stored UBL document hash does not match its persisted hash")
	}

	// Generate Phase 2 QR from the persisted invoice values.
	timestamp := fmt.Sprintf("%sT%sZ", doc.IssueDate, doc.IssueTime)
	fullQR, err := s.signer.GenerateQR(
		invData.Seller.Name,
		invData.Seller.VATNumber,
		timestamp,
		invData.TaxInclusiveAmount,
		invData.TaxTotal,
		doc.InvoiceHash,
		sig,
		pubKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed ZATCA QR data: %w", err)
	}

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
