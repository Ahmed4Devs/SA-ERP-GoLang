package submission

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
	"github.com/google/uuid"
)

type SubmissionService struct {
	repo        ports.Repository
	subRepo     ports.SubmissionRepository
	zatcaClient ports.ZATCAClient
}

func NewSubmissionService(
	repo ports.Repository,
	subRepo ports.SubmissionRepository,
	zatcaClient ports.ZATCAClient,
) *SubmissionService {
	return &SubmissionService{
		repo:        repo,
		subRepo:     subRepo,
		zatcaClient: zatcaClient,
	}
}

func (s *SubmissionService) SubmitClearance(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ClearanceResponse, error) {
	// 1. Retrieve Document
	doc, err := s.repo.GetDocumentByID(ctx, documentID, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if doc.SignedXML == "" {
		return nil, fmt.Errorf("document must be digitally signed before clearance")
	}

	// 2. Retrieve CSID
	csid, err := s.repo.GetActiveCSID(ctx, doc.EGSUnitID)
	if err != nil || csid == nil {
		return nil, fmt.Errorf("active CSID certificate not found for EGS unit: %w", err)
	}

	// 3. Mark Submitting & Create Submission Record
	_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusSubmitting)

	submission := &domain.ZATCASubmission{
		ID:             uuid.New(),
		DocumentID:     doc.ID,
		LegalEntityID:  legalEntityID,
		SubmissionMode: domain.SubmissionModeClearance,
		Environment:    domain.EnvironmentSandbox,
		Status:         "PENDING",
		SubmittedAt:    time.Now(),
	}
	savedSub, err := s.subRepo.CreateSubmission(ctx, submission)
	if err != nil {
		return nil, fmt.Errorf("failed to record submission audit: %w", err)
	}

	// 4. Dispatch to ZATCA Clearance API
	invoiceBase64 := base64.StdEncoding.EncodeToString([]byte(doc.SignedXML))
	req := domain.ClearanceRequest{
		InvoiceHash: doc.InvoiceHash,
		UUID:        doc.InvoiceUUID.String(),
		Invoice:     invoiceBase64,
	}

	resp, clientErr := s.zatcaClient.Clearance(ctx, req, csid)

	// 5. Handle Network / Gateway Errors (Enqueue for Retry)
	if clientErr != nil {
		_ = s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
			SubmissionID:   savedSub.ID,
			AttemptNumber:  1,
			HTTPStatusCode: 503,
			ErrorMessage:   clientErr.Error(),
		})

		savedSub.Status = "RETRY_PENDING"
		_, _ = s.subRepo.UpdateSubmission(ctx, savedSub)
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusRetryPending)

		return nil, fmt.Errorf("ZATCA clearance gateway error (queued for retry): %w", clientErr)
	}

	// 6. Handle ZATCA Application Response
	respJSON, _ := json.Marshal(resp)
	attemptStatus := 200
	if resp.ClearanceStatus != "CLEARED" {
		attemptStatus = 400
	}

	_ = s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
		SubmissionID:    savedSub.ID,
		AttemptNumber:   1,
		HTTPStatusCode:  attemptStatus,
		ResponsePayload: string(respJSON),
	})

	savedSub.ZATCARequestID = resp.ZATCARequestID
	savedSub.Status = resp.ClearanceStatus
	savedSub.ResponsePayload = string(respJSON)

	if resp.ClearanceStatus == "CLEARED" {
		savedSub.ClearedInvoiceXML = resp.ClearedInvoice
		_, _ = s.subRepo.UpdateSubmission(ctx, savedSub)
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusCleared)
	} else {
		_, _ = s.subRepo.UpdateSubmission(ctx, savedSub)
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusRejected)
	}

	return &resp, nil
}

func (s *SubmissionService) SubmitReporting(ctx context.Context, documentID, legalEntityID uuid.UUID) (*domain.ReportingResponse, error) {
	// 1. Retrieve Document
	doc, err := s.repo.GetDocumentByID(ctx, documentID, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	if doc.SignedXML == "" {
		return nil, fmt.Errorf("document must be digitally signed before reporting")
	}

	// 2. Retrieve CSID
	csid, err := s.repo.GetActiveCSID(ctx, doc.EGSUnitID)
	if err != nil || csid == nil {
		return nil, fmt.Errorf("active CSID certificate not found for EGS unit: %w", err)
	}

	// 3. Mark Submitting & Create Submission Record
	_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusSubmitting)

	submission := &domain.ZATCASubmission{
		ID:             uuid.New(),
		DocumentID:     doc.ID,
		LegalEntityID:  legalEntityID,
		SubmissionMode: domain.SubmissionModeReporting,
		Environment:    domain.EnvironmentSandbox,
		Status:         "PENDING",
		SubmittedAt:    time.Now(),
	}
	savedSub, err := s.subRepo.CreateSubmission(ctx, submission)
	if err != nil {
		return nil, fmt.Errorf("failed to record submission audit: %w", err)
	}

	// 4. Dispatch to ZATCA Reporting API
	invoiceBase64 := base64.StdEncoding.EncodeToString([]byte(doc.SignedXML))
	req := domain.ReportingRequest{
		InvoiceHash: doc.InvoiceHash,
		UUID:        doc.InvoiceUUID.String(),
		Invoice:     invoiceBase64,
	}

	resp, clientErr := s.zatcaClient.Report(ctx, req, csid)

	// 5. Handle Network / Gateway Errors
	if clientErr != nil {
		_ = s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
			SubmissionID:   savedSub.ID,
			AttemptNumber:  1,
			HTTPStatusCode: 503,
			ErrorMessage:   clientErr.Error(),
		})

		savedSub.Status = "RETRY_PENDING"
		_, _ = s.subRepo.UpdateSubmission(ctx, savedSub)
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusRetryPending)

		return nil, fmt.Errorf("ZATCA reporting gateway error (queued for retry): %w", clientErr)
	}

	// 6. Handle ZATCA Response
	respJSON, _ := json.Marshal(resp)
	attemptStatus := 200
	if resp.ReportingStatus != "REPORTED" {
		attemptStatus = 400
	}

	_ = s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
		SubmissionID:    savedSub.ID,
		AttemptNumber:   1,
		HTTPStatusCode:  attemptStatus,
		ResponsePayload: string(respJSON),
	})

	savedSub.ZATCARequestID = resp.ZATCARequestID
	savedSub.Status = resp.ReportingStatus
	savedSub.ResponsePayload = string(respJSON)

	if resp.ReportingStatus == "REPORTED" {
		docStatus := domain.DocumentStatusReported
		if len(resp.WarningMessages) > 0 {
			docStatus = domain.DocumentStatusWarning
		}
		_, _ = s.subRepo.UpdateSubmission(ctx, savedSub)
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, docStatus)
	} else {
		_, _ = s.subRepo.UpdateSubmission(ctx, savedSub)
		_, _ = s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusRejected)
	}

	return &resp, nil
}

func (s *SubmissionService) RetrySubmission(ctx context.Context, submissionID, legalEntityID uuid.UUID) error {
	sub, err := s.subRepo.GetSubmissionByID(ctx, submissionID, legalEntityID)
	if err != nil {
		return fmt.Errorf("submission not found: %w", err)
	}

	if sub.SubmissionMode == domain.SubmissionModeClearance {
		_, err = s.SubmitClearance(ctx, sub.DocumentID, legalEntityID)
		return err
	}

	_, err = s.SubmitReporting(ctx, sub.DocumentID, legalEntityID)
	return err
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.ZATCASubmission, error) {
	return s.subRepo.ListSubmissions(ctx, filter)
}

func (s *SubmissionService) GetSubmission(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCASubmission, error) {
	return s.subRepo.GetSubmissionByID(ctx, id, legalEntityID)
}

var _ ports.SubmissionService = (*SubmissionService)(nil)
