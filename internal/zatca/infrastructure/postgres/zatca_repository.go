package postgres

import (
	"context"
	"fmt"
	"strings"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	platformsecurity "erp/internal/platform/security"
	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct {
	db             *database.DB
	q              *sqlc.Queries
	credentialsBox *platformsecurity.SecretBox
}

func NewRepository(db *database.DB, credentialsBox ...*platformsecurity.SecretBox) *Repository {
	var box *platformsecurity.SecretBox
	if len(credentialsBox) > 0 {
		box = credentialsBox[0]
	}
	return &Repository{
		db:             db,
		q:              sqlc.New(db.Pool),
		credentialsBox: box,
	}
}

func (r *Repository) GetEGSUnitByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.EGSUnit, error) {
	row, err := r.q.GetEGSUnitByID(ctx, sqlc.GetEGSUnitByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		return nil, err
	}
	return mapEGSToDomain(row), nil
}

func (r *Repository) GetEGSUnitByCustomID(ctx context.Context, customID string, legalEntityID uuid.UUID) (*domain.EGSUnit, error) {
	row, err := r.q.GetEGSUnitByCustomID(ctx, sqlc.GetEGSUnitByCustomIDParams{
		CustomID:      customID,
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		return nil, err
	}
	return mapEGSToDomain(row), nil
}

func (r *Repository) ListEGSUnits(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.EGSUnit, error) {
	rows, err := r.q.ListEGSUnits(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.EGSUnit, len(rows))
	for i, row := range rows {
		res[i] = mapEGSToDomain(row)
	}
	return res, nil
}

func (r *Repository) CreateEGSUnit(ctx context.Context, unit *domain.EGSUnit) (*domain.EGSUnit, error) {
	row, err := r.q.CreateEGSUnit(ctx, sqlc.CreateEGSUnitParams{
		ID:                   toPgUUID(unit.ID),
		LegalEntityID:        toPgUUID(unit.LegalEntityID),
		BranchID:             toPgUUIDPtr(unit.BranchID),
		Uuid:                 toPgUUID(unit.UUID),
		CustomID:             unit.CustomID,
		Model:                unit.Model,
		SerialNumber:         unit.SerialNumber,
		SolutionName:         unit.SolutionName,
		VatNumber:            unit.VATNumber,
		OrganizationName:     unit.OrganizationName,
		OrganizationUnitName: unit.OrganizationUnitName,
		CountryCode:          unit.CountryCode,
		InvoiceTypeMask:      unit.InvoiceTypeMask,
		Status:               unit.Status,
	})
	if err != nil {
		return nil, err
	}
	return mapEGSToDomain(row), nil
}

func (r *Repository) UpdateEGSUnitStatus(ctx context.Context, id, legalEntityID uuid.UUID, status string) (*domain.EGSUnit, error) {
	row, err := r.q.UpdateEGSUnitStatus(ctx, sqlc.UpdateEGSUnitStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        status,
	})
	if err != nil {
		return nil, err
	}
	return mapEGSToDomain(row), nil
}

func (r *Repository) GetActiveCSID(ctx context.Context, egsUnitID uuid.UUID) (*domain.CSID, error) {
	row, err := r.q.GetActiveCSID(ctx, toPgUUID(egsUnitID))
	if err != nil {
		return nil, err
	}
	return r.mapCSIDToDomain(row)
}

func (r *Repository) ListCSIDs(ctx context.Context, egsUnitID uuid.UUID) ([]*domain.CSID, error) {
	rows, err := r.q.ListCSIDs(ctx, toPgUUID(egsUnitID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.CSID, len(rows))
	for i, row := range rows {
		mapped, err := r.mapCSIDToDomain(row)
		if err != nil {
			return nil, err
		}
		res[i] = mapped
	}
	return res, nil
}

func (r *Repository) CreateCSID(ctx context.Context, csid *domain.CSID) (*domain.CSID, error) {
	var exp pgtype.Timestamptz
	if csid.ExpiresAt != nil {
		exp = toPgTimestamptz(*csid.ExpiresAt)
	}
	secret, privateKey, err := r.encryptCredentials(csid.Secret, csid.PrivateKeyPEM)
	if err != nil {
		return nil, err
	}
	row, err := r.q.CreateCSID(ctx, sqlc.CreateCSIDParams{
		ID:                  toPgUUID(csid.ID),
		EgsUnitID:           toPgUUID(csid.EGSUnitID),
		CsidType:            string(csid.CSIDType),
		BinarySecurityToken: csid.BinarySecurityToken,
		Secret:              secret,
		PrivateKeyPem:       privateKey,
		PublicKeyPem:        csid.PublicKeyPEM,
		CertificatePem:      toPgText(csid.CertificatePEM),
		RequestID:           toPgText(csid.RequestID),
		ComplianceRequestID: toPgText(csid.ComplianceRequestID),
		ExpiresAt:           exp,
		Status:              csid.Status,
	})
	if err != nil {
		return nil, err
	}
	return r.mapCSIDToDomain(row)
}

func (r *Repository) ReserveNextSequence(ctx context.Context, egsUnitID uuid.UUID, invoiceType string) (int64, string, error) {
	const initialPIH = "NWZlY2ViNjAxNWM4OTQyYzhkNTJhN2RlZWZmODM4NDFlZWY3YjFmMzBkYzYxOWQ5MTUzZGFhYTEzYTI3MWEz"
	var icv int64
	var pih string
	err := r.db.DBTX(ctx).QueryRow(ctx, `
		INSERT INTO zatca.invoice_sequences (
			id, egs_unit_id, invoice_type, last_icv, last_invoice_hash, updated_at
		) VALUES ($1, $2, $3, 1, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (egs_unit_id, invoice_type) DO UPDATE
		SET last_icv = zatca.invoice_sequences.last_icv + 1,
			updated_at = CURRENT_TIMESTAMP
		RETURNING last_icv, last_invoice_hash
	`, uuid.New(), egsUnitID, invoiceType, initialPIH).Scan(&icv, &pih)
	if err != nil {
		return 0, "", fmt.Errorf("failed to reserve ZATCA invoice sequence: %w", err)
	}
	return icv, pih, nil
}

func (r *Repository) SetSequenceHash(ctx context.Context, egsUnitID uuid.UUID, invoiceType string, icv int64, currentHash string) error {
	command, err := r.db.DBTX(ctx).Exec(ctx, `
		UPDATE zatca.invoice_sequences
		SET last_invoice_hash = $4, updated_at = CURRENT_TIMESTAMP
		WHERE egs_unit_id = $1 AND invoice_type = $2 AND last_icv = $3
	`, egsUnitID, invoiceType, icv, currentHash)
	if err != nil {
		return fmt.Errorf("failed to update ZATCA invoice hash chain: %w", err)
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("ZATCA invoice sequence changed while generating ICV %d", icv)
	}
	return nil
}

func (r *Repository) GetDocumentByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCADocument, error) {
	row, err := r.q.GetDocumentByID(ctx, sqlc.GetDocumentByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		return nil, err
	}
	return mapDocToDomain(row), nil
}

func (r *Repository) GetDocumentByUUID(ctx context.Context, u uuid.UUID) (*domain.ZATCADocument, error) {
	row, err := r.q.GetDocumentByUUID(ctx, toPgUUID(u))
	if err != nil {
		return nil, err
	}
	return mapDocToDomain(row), nil
}

func (r *Repository) ListDocuments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.ZATCADocument, error) {
	rows, err := r.q.ListDocuments(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.ZATCADocument, len(rows))
	for i, row := range rows {
		res[i] = mapDocToDomain(row)
	}
	return res, nil
}

func (r *Repository) CreateDocument(ctx context.Context, doc *domain.ZATCADocument) (*domain.ZATCADocument, error) {
	row, err := sqlc.New(r.db.DBTX(ctx)).CreateDocument(ctx, sqlc.CreateDocumentParams{
		ID:               toPgUUID(doc.ID),
		LegalEntityID:    toPgUUID(doc.LegalEntityID),
		EgsUnitID:        toPgUUID(doc.EGSUnitID),
		SalesInvoiceID:   toPgUUIDPtr(doc.SalesInvoiceID),
		InvoiceUuid:      toPgUUID(doc.InvoiceUUID),
		InvoiceNumber:    doc.InvoiceNumber,
		InvoiceType:      string(doc.InvoiceType),
		InvoiceSubtype:   string(doc.InvoiceSubtype),
		SubmissionMode:   string(doc.SubmissionMode),
		Icv:              doc.ICV,
		Pih:              doc.PIH,
		IssueDate:        toPgDate(doc.IssueDate),
		IssueTime:        toPgTime(doc.IssueTime),
		UblXml:           doc.UBLXML,
		SignedXml:        toPgText(doc.SignedXML),
		InvoiceHash:      doc.InvoiceHash,
		DigitalSignature: toPgText(doc.DigitalSignature),
		PublicKeyHashing: toPgText(doc.PublicKeyHashing),
		QrCodeData:       doc.QRCodeData,
		Status:           string(doc.Status),
	})
	if err != nil {
		return nil, err
	}
	return mapDocToDomain(row), nil
}

func (r *Repository) UpdateDocumentSigned(ctx context.Context, id, legalEntityID uuid.UUID, signedXML, hash, signature, pubKey, qr string, status domain.DocumentStatus) (*domain.ZATCADocument, error) {
	row, err := r.q.UpdateDocumentSigned(ctx, sqlc.UpdateDocumentSignedParams{
		ID:               toPgUUID(id),
		LegalEntityID:    toPgUUID(legalEntityID),
		SignedXml:        toPgText(signedXML),
		InvoiceHash:      hash,
		DigitalSignature: toPgText(signature),
		PublicKeyHashing: toPgText(pubKey),
		QrCodeData:       qr,
		Status:           string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapDocToDomain(row), nil
}

func (r *Repository) UpdateDocumentStatus(ctx context.Context, id, legalEntityID uuid.UUID, status domain.DocumentStatus) (*domain.ZATCADocument, error) {
	row, err := r.q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapDocToDomain(row), nil
}

func (r *Repository) SaveValidationResult(ctx context.Context, res *domain.ValidationResult) error {
	vRow, err := r.q.CreateValidationResult(ctx, sqlc.CreateValidationResultParams{
		ID:            toPgUUID(res.ID),
		DocumentID:    toPgUUID(res.DocumentID),
		OverallStatus: res.OverallStatus,
		RuleVersion:   res.RuleVersion,
	})
	if err != nil {
		return err
	}

	for _, msg := range res.Messages {
		_, errMsg := r.q.CreateValidationMessage(ctx, sqlc.CreateValidationMessageParams{
			ID:                 toPgUUID(msg.ID),
			ValidationResultID: vRow.ID,
			MessageType:        msg.MessageType,
			Code:               msg.Code,
			Category:           msg.Category,
			MessageText:        msg.MessageText,
		})
		if errMsg != nil {
			return errMsg
		}
	}

	return nil
}

func (r *Repository) GetValidationResults(ctx context.Context, documentID uuid.UUID) ([]*domain.ValidationResult, error) {
	rows, err := r.q.GetValidationResultsByDocument(ctx, toPgUUID(documentID))
	if err != nil {
		return nil, err
	}

	res := make([]*domain.ValidationResult, len(rows))
	for i, row := range rows {
		msgs, errMsgs := r.q.GetValidationMessagesByResult(ctx, row.ID)
		if errMsgs != nil {
			return nil, fmt.Errorf("failed to get validation messages: %w", errMsgs)
		}
		msgList := []domain.ValidationMessage{}
		for _, m := range msgs {
			msgList = append(msgList, domain.ValidationMessage{
				ID:          toUUID(m.ID),
				MessageType: m.MessageType,
				Code:        m.Code,
				Category:    m.Category,
				MessageText: m.MessageText,
				CreatedAt:   fromPgTimestamptz(m.CreatedAt),
			})
		}

		res[i] = &domain.ValidationResult{
			ID:            toUUID(row.ID),
			DocumentID:    toUUID(row.DocumentID),
			OverallStatus: row.OverallStatus,
			RuleVersion:   row.RuleVersion,
			ValidatedAt:   fromPgTimestamptz(row.ValidatedAt),
			Messages:      msgList,
		}
	}

	return res, nil
}

func (r *Repository) GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ZATCASummary, error) {
	counts, err := r.q.CountDocumentsByStatus(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}

	egsList, err := r.q.ListEGSUnits(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}

	return &domain.ZATCASummary{
		TotalDocuments: counts.TotalDocuments,
		ClearedCount:   counts.ClearedCount,
		ReportedCount:  counts.ReportedCount,
		WarningCount:   counts.WarningCount,
		FailedCount:    counts.FailedCount,
		ActiveEGSCount: len(egsList),
	}, nil
}

// Helpers
func mapEGSToDomain(row sqlc.ZatcaEgsUnits) *domain.EGSUnit {
	return &domain.EGSUnit{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		BranchID:             toUUIDPtr(row.BranchID),
		UUID:                 toUUID(row.Uuid),
		CustomID:             row.CustomID,
		Model:                row.Model,
		SerialNumber:         row.SerialNumber,
		SolutionName:         row.SolutionName,
		VATNumber:            row.VatNumber,
		OrganizationName:     row.OrganizationName,
		OrganizationUnitName: row.OrganizationUnitName,
		CountryCode:          row.CountryCode,
		InvoiceTypeMask:      row.InvoiceTypeMask,
		Status:               row.Status,
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:            fromPgTimestamptz(row.UpdatedAt),
	}
}

func (r *Repository) mapCSIDToDomain(row sqlc.ZatcaCsids) (*domain.CSID, error) {
	secret, privateKey, err := r.decryptCredentials(row.Secret, row.PrivateKeyPem)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CSID %s credentials: %w", toUUID(row.ID), err)
	}
	return &domain.CSID{
		ID:                  toUUID(row.ID),
		EGSUnitID:           toUUID(row.EgsUnitID),
		CSIDType:            domain.CSIDType(row.CsidType),
		BinarySecurityToken: row.BinarySecurityToken,
		Secret:              secret,
		PrivateKeyPEM:       privateKey,
		PublicKeyPEM:        row.PublicKeyPem,
		CertificatePEM:      fromPgText(row.CertificatePem),
		RequestID:           fromPgText(row.RequestID),
		ComplianceRequestID: fromPgText(row.ComplianceRequestID),
		ExpiresAt:           fromPgTimestamptzPtr(row.ExpiresAt),
		Status:              row.Status,
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

const encryptedCredentialPrefix = "enc:v1:"

func (r *Repository) encryptCredentials(secret, privateKey string) (string, string, error) {
	if r.credentialsBox == nil {
		return secret, privateKey, nil
	}
	encryptedSecret, err := r.credentialsBox.Encrypt(secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt CSID secret: %w", err)
	}
	encryptedPrivateKey, err := r.credentialsBox.Encrypt(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt CSID private key: %w", err)
	}
	return encryptedCredentialPrefix + encryptedSecret, encryptedCredentialPrefix + encryptedPrivateKey, nil
}

func (r *Repository) decryptCredentials(secret, privateKey string) (string, string, error) {
	if r.credentialsBox == nil {
		return secret, privateKey, nil
	}
	if !strings.HasPrefix(secret, encryptedCredentialPrefix) || !strings.HasPrefix(privateKey, encryptedCredentialPrefix) {
		return "", "", fmt.Errorf("unencrypted CSID credentials are not accepted when credential encryption is enabled")
	}
	decryptedSecret, err := r.credentialsBox.Decrypt(strings.TrimPrefix(secret, encryptedCredentialPrefix))
	if err != nil {
		return "", "", err
	}
	decryptedPrivateKey, err := r.credentialsBox.Decrypt(strings.TrimPrefix(privateKey, encryptedCredentialPrefix))
	if err != nil {
		return "", "", err
	}
	return decryptedSecret, decryptedPrivateKey, nil
}

func mapDocToDomain(row sqlc.ZatcaDocuments) *domain.ZATCADocument {
	return &domain.ZATCADocument{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		EGSUnitID:        toUUID(row.EgsUnitID),
		SalesInvoiceID:   toUUIDPtr(row.SalesInvoiceID),
		InvoiceUUID:      toUUID(row.InvoiceUuid),
		InvoiceNumber:    row.InvoiceNumber,
		InvoiceType:      domain.InvoiceType(row.InvoiceType),
		InvoiceSubtype:   domain.InvoiceSubtype(row.InvoiceSubtype),
		SubmissionMode:   domain.SubmissionMode(row.SubmissionMode),
		ICV:              row.Icv,
		PIH:              row.Pih,
		IssueDate:        fromPgDate(row.IssueDate),
		IssueTime:        fromPgTime(row.IssueTime),
		UBLXML:           row.UblXml,
		SignedXML:        fromPgText(row.SignedXml),
		InvoiceHash:      row.InvoiceHash,
		DigitalSignature: fromPgText(row.DigitalSignature),
		PublicKeyHashing: fromPgText(row.PublicKeyHashing),
		QRCodeData:       row.QrCodeData,
		Status:           domain.DocumentStatus(row.Status),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:        fromPgTimestamptz(row.UpdatedAt),
	}
}

var _ ports.Repository = (*Repository)(nil)
