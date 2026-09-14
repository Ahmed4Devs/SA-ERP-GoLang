package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/organization/domain"
	"erp/internal/organization/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	orgService ports.OrganizationService
	seqService ports.NumberSequenceService
}

func NewOrganizationHandler(orgService ports.OrganizationService, seqService ports.NumberSequenceService) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
		seqService: seqService,
	}
}

func (h *OrganizationHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Protected Organization Endpoints
	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Organizations
		r.Get("/organizations", h.ListOrganizations)
		r.Post("/organizations", h.CreateOrganization)

		// Legal Entities
		r.Get("/legal-entities", h.ListLegalEntities)
		r.Post("/legal-entities", h.CreateLegalEntity)

		// Branches
		r.Get("/branches", h.ListBranches)
		r.Post("/branches", h.CreateBranch)

		// Departments
		r.Get("/departments", h.ListDepartments)
		r.Post("/departments", h.CreateDepartment)

		// Cost Centers
		r.Get("/cost-centers", h.ListCostCenters)
		r.Post("/cost-centers", h.CreateCostCenter)

		// Number Sequences
		r.Get("/number-sequences", h.ListSequences)
		r.Post("/number-sequences", h.DefineSequence)
		r.Post("/number-sequences/next", h.GenerateNextSequence)
	})

	return r
}

func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.orgService.ListOrganizations(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": orgs})
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	org, err := h.orgService.CreateOrganization(r.Context(), req.Code, req.NameAr, req.NameEn)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": org})
}

func (h *OrganizationHandler) ListLegalEntities(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("organization_id")
	if orgIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "organization_id query parameter is required")
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid organization_id UUID")
		return
	}

	entities, err := h.orgService.ListLegalEntities(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": entities})
}

func (h *OrganizationHandler) CreateLegalEntity(w http.ResponseWriter, r *http.Request) {
	var req CreateLegalEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid organization_id UUID")
		return
	}

	entity, err := h.orgService.CreateLegalEntity(r.Context(), orgID, req.Code, req.NameAr, req.NameEn, req.CRNumber, req.BaseCurrency)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": entity})
}

func (h *OrganizationHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	entityIDStr := r.URL.Query().Get("legal_entity_id")
	if entityIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "legal_entity_id query parameter is required")
		return
	}
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	branches, err := h.orgService.ListBranches(r.Context(), entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": branches})
}

func (h *OrganizationHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	var req CreateBranchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	entityID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	branch, err := h.orgService.CreateBranch(r.Context(), entityID, req.Code, req.NameAr, req.NameEn)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": branch})
}

func (h *OrganizationHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	entityIDStr := r.URL.Query().Get("legal_entity_id")
	if entityIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "legal_entity_id query parameter is required")
		return
	}
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	depts, err := h.orgService.ListDepartments(r.Context(), entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": depts})
}

func (h *OrganizationHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	entityID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	dept, err := h.orgService.CreateDepartment(r.Context(), entityID, parentID, req.Code, req.NameAr, req.NameEn)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": dept})
}

func (h *OrganizationHandler) ListCostCenters(w http.ResponseWriter, r *http.Request) {
	entityIDStr := r.URL.Query().Get("legal_entity_id")
	if entityIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "legal_entity_id query parameter is required")
		return
	}
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	ccs, err := h.orgService.ListCostCenters(r.Context(), entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": ccs})
}

func (h *OrganizationHandler) CreateCostCenter(w http.ResponseWriter, r *http.Request) {
	var req CreateCostCenterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	entityID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err == nil {
			parentID = &pid
		}
	}

	cc, err := h.orgService.CreateCostCenter(r.Context(), entityID, parentID, req.Code, req.NameAr, req.NameEn)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": cc})
}

func (h *OrganizationHandler) ListSequences(w http.ResponseWriter, r *http.Request) {
	entityIDStr := r.URL.Query().Get("legal_entity_id")
	if entityIDStr == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAM", "legal_entity_id query parameter is required")
		return
	}
	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	seqs, err := h.seqService.ListSequences(r.Context(), entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": seqs})
}

func (h *OrganizationHandler) DefineSequence(w http.ResponseWriter, r *http.Request) {
	var req DefineSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	entityID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	seq, err := h.seqService.DefineSequence(
		r.Context(),
		entityID,
		req.Code,
		req.Prefix,
		req.Suffix,
		req.StartValue,
		req.Step,
		req.PaddingLength,
		domain.ResetInterval(req.ResetInterval),
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": seq})
}

func (h *OrganizationHandler) GenerateNextSequence(w http.ResponseWriter, r *http.Request) {
	var req GenerateNextSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	entityID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "invalid legal_entity_id UUID")
		return
	}

	num, err := h.seqService.GenerateNext(r.Context(), entityID, req.SequenceCode)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"number": num})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}
