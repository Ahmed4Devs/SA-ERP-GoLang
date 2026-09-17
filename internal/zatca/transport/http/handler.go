package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ZATCAHandler struct {
	svc           ports.Service
	onboardingSvc ports.OnboardingService
	submissionSvc ports.SubmissionService
}

func NewZATCAHandler(svc ports.Service, onboardingSvc ports.OnboardingService, submissionSvc ports.SubmissionService) *ZATCAHandler {
	return &ZATCAHandler{
		svc:           svc,
		onboardingSvc: onboardingSvc,
		submissionSvc: submissionSvc,
	}
}

func (h *ZATCAHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// EGS Units
		protected.Get("/egs-units", h.ListEGSUnits)
		protected.Post("/egs-units", h.RegisterEGSUnit)
		protected.Get("/egs-units/{id}", h.GetEGSUnit)
		protected.Get("/egs-units/{id}/csids", h.ListCSIDs)
		protected.Post("/egs-units/{id}/csids", h.GenerateCSID)

		// Onboarding (Phase 26)
		protected.Post("/onboarding/start", h.StartOnboarding)
		protected.Get("/onboarding/status", h.GetOnboardingStatus)

		// Documents
		protected.Get("/documents", h.ListDocuments)
		protected.Post("/documents/generate", h.GenerateDocument)
		protected.Get("/documents/{id}", h.GetDocument)
		protected.Get("/documents/{id}/xml", h.GetDocumentXML)
		protected.Post("/documents/{id}/validate", h.ValidateDocument)
		protected.Post("/documents/{id}/sign", h.SignDocument)

		// Submissions (Phase 26 Clearance & Reporting)
		protected.Post("/documents/{id}/clearance", h.SubmitClearance)
		protected.Post("/documents/{id}/reporting", h.SubmitReporting)
		protected.Get("/submissions", h.ListSubmissions)
		protected.Get("/submissions/{id}", h.GetSubmission)
		protected.Post("/submissions/{id}/retry", h.RetrySubmission)

		// Summary
		protected.Get("/summary", h.GetSummary)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ----------------------------------------------------------------------
// EGS Units & CSIDs
// ----------------------------------------------------------------------

func (h *ZATCAHandler) ListEGSUnits(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	units, err := h.svc.ListEGSUnits(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, units)
}

type RegisterEGSUnitPayload struct {
	LegalEntityID        string  `json:"legal_entity_id"`
	BranchID             *string `json:"branch_id,omitempty"`
	CustomID             string  `json:"custom_id"`
	Model                string  `json:"model"`
	SerialNumber         string  `json:"serial_number"`
	SolutionName         string  `json:"solution_name"`
	VATNumber            string  `json:"vat_number"`
	OrganizationName     string  `json:"organization_name"`
	OrganizationUnitName string  `json:"organization_unit_name"`
	InvoiceTypeMask      string  `json:"invoice_type_mask"`
}

func (h *ZATCAHandler) RegisterEGSUnit(w http.ResponseWriter, r *http.Request) {
	var payload RegisterEGSUnitPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(payload.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	var branchID *uuid.UUID
	if payload.BranchID != nil && *payload.BranchID != "" {
		bID, parseErr := uuid.Parse(*payload.BranchID)
		if parseErr == nil {
			branchID = &bID
		}
	}

	unit, err := h.svc.RegisterEGSUnit(r.Context(), ports.RegisterEGSRequest{
		LegalEntityID:        leID,
		BranchID:             branchID,
		CustomID:             payload.CustomID,
		Model:                payload.Model,
		SerialNumber:         payload.SerialNumber,
		SolutionName:         payload.SolutionName,
		VATNumber:            payload.VATNumber,
		OrganizationName:     payload.OrganizationName,
		OrganizationUnitName: payload.OrganizationUnitName,
		InvoiceTypeMask:      payload.InvoiceTypeMask,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, unit)
}

func (h *ZATCAHandler) GetEGSUnit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid egs id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	unit, err := h.svc.GetEGSUnit(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, unit)
}

func (h *ZATCAHandler) ListCSIDs(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid egs id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	csids, err := h.svc.ListCSIDs(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, csids)
}

type GenerateCSIDPayload struct {
	CSIDType string `json:"csid_type"`
	OTP      string `json:"otp"`
}

func (h *ZATCAHandler) GenerateCSID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid egs id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	var payload GenerateCSIDPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)

	csType := domain.CSIDTypeCompliance
	if payload.CSIDType == "PRODUCTION" {
		csType = domain.CSIDTypeProduction
	}

	csid, err := h.svc.GenerateCSID(r.Context(), id, leID, csType, payload.OTP)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, csid)
}

// ----------------------------------------------------------------------
// Documents
// ----------------------------------------------------------------------

func (h *ZATCAHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	docs, err := h.svc.ListDocuments(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

type GenerateDocumentPayload struct {
	LegalEntityID string             `json:"legal_entity_id"`
	EGSUnitID     string             `json:"egs_unit_id"`
	Data          domain.InvoiceData `json:"data"`
}

func (h *ZATCAHandler) GenerateDocument(w http.ResponseWriter, r *http.Request) {
	var payload GenerateDocumentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(payload.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	egsID, err := uuid.Parse(payload.EGSUnitID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid egs_unit_id")
		return
	}

	doc, err := h.svc.GenerateDocumentFromInvoiceData(r.Context(), leID, egsID, payload.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, doc)
}

func (h *ZATCAHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	doc, err := h.svc.GetDocument(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (h *ZATCAHandler) GetDocumentXML(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	doc, err := h.svc.GetDocument(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	xmlContent := doc.UBLXML
	if doc.SignedXML != "" {
		xmlContent = doc.SignedXML
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xmlContent))
}

func (h *ZATCAHandler) ValidateDocument(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	res, err := h.svc.ValidateDocument(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *ZATCAHandler) SignDocument(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	doc, err := h.svc.SignDocument(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (h *ZATCAHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	summary, err := h.svc.GetSummary(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// ----------------------------------------------------------------------------
// Phase 26 Handlers (Onboarding, Clearance, Reporting, Retries, Audit)
// ----------------------------------------------------------------------------

type StartOnboardingPayload struct {
	LegalEntityID string `json:"legal_entity_id"`
	EGSUnitID     string `json:"egs_unit_id"`
	OTP           string `json:"otp"`
}

func (h *ZATCAHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	if h.onboardingSvc == nil {
		writeError(w, http.StatusNotImplemented, "onboarding service not enabled")
		return
	}

	var payload StartOnboardingPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, err := uuid.Parse(payload.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	egsID, err := uuid.Parse(payload.EGSUnitID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid egs_unit_id")
		return
	}

	res, err := h.onboardingSvc.StartOnboarding(r.Context(), leID, egsID, payload.OTP)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *ZATCAHandler) GetOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	if h.onboardingSvc == nil {
		writeError(w, http.StatusNotImplemented, "onboarding service not enabled")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	egsID, err := uuid.Parse(r.URL.Query().Get("egs_unit_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "egs_unit_id query parameter is required")
		return
	}

	status, err := h.onboardingSvc.GetOnboardingStatus(r.Context(), leID, egsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *ZATCAHandler) SubmitClearance(w http.ResponseWriter, r *http.Request) {
	if h.submissionSvc == nil {
		writeError(w, http.StatusNotImplemented, "submission service not enabled")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	resp, err := h.submissionSvc.SubmitClearance(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ZATCAHandler) SubmitReporting(w http.ResponseWriter, r *http.Request) {
	if h.submissionSvc == nil {
		writeError(w, http.StatusNotImplemented, "submission service not enabled")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	resp, err := h.submissionSvc.SubmitReporting(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ZATCAHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	if h.submissionSvc == nil {
		writeError(w, http.StatusNotImplemented, "submission service not enabled")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	filter := domain.SubmissionFilter{
		LegalEntityID: leID,
		Limit:         50,
	}

	if docIDStr := r.URL.Query().Get("document_id"); docIDStr != "" {
		if dID, parseErr := uuid.Parse(docIDStr); parseErr == nil {
			filter.DocumentID = &dID
		}
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		filter.Status = &statusStr
	}

	submissions, err := h.submissionSvc.ListSubmissions(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, submissions)
}

func (h *ZATCAHandler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	if h.submissionSvc == nil {
		writeError(w, http.StatusNotImplemented, "submission service not enabled")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid submission id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	sub, err := h.submissionSvc.GetSubmission(r.Context(), id, leID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (h *ZATCAHandler) RetrySubmission(w http.ResponseWriter, r *http.Request) {
	if h.submissionSvc == nil {
		writeError(w, http.StatusNotImplemented, "submission service not enabled")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid submission id")
		return
	}

	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}

	if err := h.submissionSvc.RetrySubmission(r.Context(), id, leID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "RETRY_TRIGGERED"})
}

