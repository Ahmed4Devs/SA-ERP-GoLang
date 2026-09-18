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
	if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusSubmitting); err != nil {
		return nil, fmt.Errorf("failed to mark document as submitting: %w", err)
	}

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
		if err := s.recordRetryFailure(ctx, savedSub, doc, 1, clientErr); err != nil {
			return nil, fmt.Errorf("ZATCA clearance gateway error: %w", err)
		}

		return nil, fmt.Errorf("ZATCA clearance gateway error (queued for retry): %w", clientErr)
	}

	// 6. Handle ZATCA Application Response
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ZATCA clearance response: %w", err)
	}
	attemptStatus := 200
	if resp.ClearanceStatus != "CLEARED" {
		attemptStatus = 400
	}

	if err := s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
		SubmissionID:    savedSub.ID,
		AttemptNumber:   1,
		HTTPStatusCode:  attemptStatus,
		ResponsePayload: string(respJSON),
	}); err != nil {
		return nil, fmt.Errorf("failed to record ZATCA clearance attempt: %w", err)
	}

	savedSub.ZATCARequestID = resp.ZATCARequestID
	savedSub.Status = resp.ClearanceStatus
	savedSub.ResponsePayload = string(respJSON)

	documentStatus := domain.DocumentStatusRejected
	if resp.ClearanceStatus == "CLEARED" {
		savedSub.ClearedInvoiceXML = resp.ClearedInvoice
		documentStatus = domain.DocumentStatusCleared
	}
	if _, err := s.subRepo.UpdateSubmission(ctx, savedSub); err != nil {
		return nil, fmt.Errorf("failed to update ZATCA clearance submission: %w", err)
	}
	if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, documentStatus); err != nil {
		return nil, fmt.Errorf("failed to update document after ZATCA clearance: %w", err)
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
	if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, domain.DocumentStatusSubmitting); err != nil {
		return nil, fmt.Errorf("failed to mark document as submitting: %w", err)
	}

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
		if err := s.recordRetryFailure(ctx, savedSub, doc, 1, clientErr); err != nil {
			return nil, fmt.Errorf("ZATCA reporting gateway error: %w", err)
		}

		return nil, fmt.Errorf("ZATCA reporting gateway error (queued for retry): %w", clientErr)
	}

	// 6. Handle ZATCA Response
	respJSON, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to encode ZATCA reporting response: %w", err)
	}
	attemptStatus := 200
	if resp.ReportingStatus != "REPORTED" {
		attemptStatus = 400
	}

	if err := s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
		SubmissionID:    savedSub.ID,
		AttemptNumber:   1,
		HTTPStatusCode:  attemptStatus,
		ResponsePayload: string(respJSON),
	}); err != nil {
		return nil, fmt.Errorf("failed to record ZATCA reporting attempt: %w", err)
	}

	savedSub.ZATCARequestID = resp.ZATCARequestID
	savedSub.Status = resp.ReportingStatus
	savedSub.ResponsePayload = string(respJSON)
	docStatus := domain.DocumentStatusRejected
	if resp.ReportingStatus == "REPORTED" {
		docStatus = domain.DocumentStatusReported
		if len(resp.WarningMessages) > 0 {
			docStatus = domain.DocumentStatusWarning
		}
	}
	if _, err := s.subRepo.UpdateSubmission(ctx, savedSub); err != nil {
		return nil, fmt.Errorf("failed to update ZATCA reporting submission: %w", err)
	}
	if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, docStatus); err != nil {
		return nil, fmt.Errorf("failed to update document after ZATCA reporting: %w", err)
	}

	return &resp, nil
}

func (s *SubmissionService) RetrySubmission(ctx context.Context, submissionID, legalEntityID uuid.UUID) error {
	sub, err := s.subRepo.GetSubmissionByID(ctx, submissionID, legalEntityID)
	if err != nil {
		return fmt.Errorf("submission not found: %w", err)
	}

	doc, err := s.repo.GetDocumentByID(ctx, sub.DocumentID, legalEntityID)
	if err != nil {
		return fmt.Errorf("document not found for retry: %w", err)
	}
	csid, err := s.repo.GetActiveCSID(ctx, doc.EGSUnitID)
	if err != nil || csid == nil {
		return fmt.Errorf("active CSID certificate not found for retry: %w", err)
	}

	attempts, err := s.subRepo.ListAttempts(ctx, sub.ID)
	if err != nil {
		return fmt.Errorf("failed to list previous ZATCA submission attempts: %w", err)
	}
	attemptNumber := len(attempts) + 1
	invoiceBase64 := base64.StdEncoding.EncodeToString([]byte(doc.SignedXML))

	if sub.SubmissionMode == domain.SubmissionModeClearance {
		resp, clientErr := s.zatcaClient.Clearance(ctx, domain.ClearanceRequest{
			InvoiceHash: doc.InvoiceHash,
			UUID:        doc.InvoiceUUID.String(),
			Invoice:     invoiceBase64,
		}, csid)
		if clientErr != nil {
			if err := s.recordRetryFailure(ctx, sub, doc, attemptNumber, clientErr); err != nil {
				return fmt.Errorf("failed to persist ZATCA clearance retry failure: %w", err)
			}
			return clientErr
		}

		payload, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("failed to encode ZATCA clearance retry response: %w", err)
		}
		statusCode := 200
		documentStatus := domain.DocumentStatusCleared
		if resp.ClearanceStatus != "CLEARED" {
			statusCode = 400
			documentStatus = domain.DocumentStatusRejected
		}
		if err := s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
			SubmissionID: sub.ID, AttemptNumber: attemptNumber,
			HTTPStatusCode: statusCode, ResponsePayload: string(payload),
		}); err != nil {
			return fmt.Errorf("failed to record ZATCA clearance retry attempt: %w", err)
		}
		sub.Status = resp.ClearanceStatus
		sub.ZATCARequestID = resp.ZATCARequestID
		if _, err := s.subRepo.UpdateSubmission(ctx, sub); err != nil {
			return fmt.Errorf("failed to update ZATCA clearance retry: %w", err)
		}
		if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, documentStatus); err != nil {
			return fmt.Errorf("failed to update document after ZATCA clearance retry: %w", err)
		}
		return nil
	}

	resp, clientErr := s.zatcaClient.Report(ctx, domain.ReportingRequest{
		InvoiceHash: doc.InvoiceHash,
		UUID:        doc.InvoiceUUID.String(),
		Invoice:     invoiceBase64,
	}, csid)
	if clientErr != nil {
		if err := s.recordRetryFailure(ctx, sub, doc, attemptNumber, clientErr); err != nil {
			return fmt.Errorf("failed to persist ZATCA reporting retry failure: %w", err)
		}
		return clientErr
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to encode ZATCA reporting retry response: %w", err)
	}
	statusCode := 200
	documentStatus := domain.DocumentStatusReported
	if len(resp.WarningMessages) > 0 {
		documentStatus = domain.DocumentStatusWarning
	}
	if resp.ReportingStatus != "REPORTED" {
		statusCode = 400
		documentStatus = domain.DocumentStatusRejected
	}
	if err := s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
		SubmissionID: sub.ID, AttemptNumber: attemptNumber,
		HTTPStatusCode: statusCode, ResponsePayload: string(payload),
	}); err != nil {
		return fmt.Errorf("failed to record ZATCA reporting retry attempt: %w", err)
	}
	sub.Status = resp.ReportingStatus
	sub.ZATCARequestID = resp.ZATCARequestID
	if _, err := s.subRepo.UpdateSubmission(ctx, sub); err != nil {
		return fmt.Errorf("failed to update ZATCA reporting retry: %w", err)
	}
	if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, legalEntityID, documentStatus); err != nil {
		return fmt.Errorf("failed to update document after ZATCA reporting retry: %w", err)
	}
	return nil
}

func (s *SubmissionService) recordRetryFailure(ctx context.Context, sub *domain.ZATCASubmission, doc *domain.ZATCADocument, attemptNumber int, retryErr error) error {
	if err := s.subRepo.RecordAttempt(ctx, &domain.SubmissionAttempt{
		SubmissionID: sub.ID, AttemptNumber: attemptNumber,
		HTTPStatusCode: 503, ErrorMessage: retryErr.Error(),
	}); err != nil {
		return err
	}
	sub.Status = "RETRY_PENDING"
	if _, err := s.subRepo.UpdateSubmission(ctx, sub); err != nil {
		return fmt.Errorf("failed to mark ZATCA submission for retry: %w", err)
	}
	if _, err := s.repo.UpdateDocumentStatus(ctx, doc.ID, sub.LegalEntityID, domain.DocumentStatusRetryPending); err != nil {
		return fmt.Errorf("failed to mark ZATCA document for retry: %w", err)
	}
	return nil
}

func (s *SubmissionService) ListSubmissions(ctx context.Context, filter domain.SubmissionFilter) ([]*domain.ZATCASubmission, error) {
	return s.subRepo.ListSubmissions(ctx, filter)
}

func (s *SubmissionService) GetSubmission(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.ZATCASubmission, error) {
	return s.subRepo.GetSubmissionByID(ctx, id, legalEntityID)
}

var _ ports.SubmissionService = (*SubmissionService)(nil)
