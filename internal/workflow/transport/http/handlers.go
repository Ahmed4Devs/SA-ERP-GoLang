package http

import (
	"encoding/json"
	"net/http"

	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/workflow/domain"
	"erp/internal/workflow/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WorkflowHandler struct {
	workflowService ports.WorkflowService
}

func NewWorkflowHandler(workflowService ports.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: workflowService}
}

func (h *WorkflowHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		r.Get("/definitions", h.ListDefinitions)
		r.Get("/inbox", h.ListInbox)
		r.Post("/instances", h.StartWorkflow)
		r.Get("/instances/{id}", h.GetInstance)
		r.Post("/requests/{id}/decide", h.SubmitDecision)
		r.Get("/delegations", h.ListDelegations)
		r.Post("/delegations", h.CreateDelegation)
		r.Delete("/delegations/{id}", h.RevokeDelegation)
	})

	return r
}

func (h *WorkflowHandler) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	defs, err := h.workflowService.ListDefinitions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": defs})
}

func (h *WorkflowHandler) ListInbox(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	items, err := h.workflowService.ListPendingApprovals(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *WorkflowHandler) StartWorkflow(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req StartWorkflowRequest
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid legal_entity_id UUID")
		return
	}

	inst, err := h.workflowService.StartWorkflow(r.Context(), ports.StartWorkflowCommand{
		DefinitionCode: req.DefinitionCode,
		LegalEntityID:  leID,
		ResourceType:   req.ResourceType,
		ResourceID:     req.ResourceID,
		InitiatorID:    user.ID,
		ContextData:    req.ContextData,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "START_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": inst})
}

func (h *WorkflowHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid workflow instance UUID")
		return
	}

	inst, err := h.workflowService.GetInstance(r.Context(), id)
	if err != nil {
		if err == domain.ErrWorkflowInstanceNotFound {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Workflow instance not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": inst})
}

func (h *WorkflowHandler) SubmitDecision(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	reqIDStr := chi.URLParam(r, "id")
	reqID, err := uuid.Parse(reqIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid request UUID")
		return
	}

	var req SubmitDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Decision != "APPROVE" && req.Decision != "REJECT" {
		writeError(w, http.StatusBadRequest, "INVALID_DECISION", "Decision must be APPROVE or REJECT")
		return
	}

	decision, err := h.workflowService.SubmitDecision(r.Context(), ports.SubmitDecisionCommand{
		RequestID: reqID,
		ActorID:   user.ID,
		Decision:  req.Decision,
		Notes:     req.Notes,
	})
	if err != nil {
		if err == domain.ErrMakerCannotApprove {
			writeError(w, http.StatusForbidden, "MAKER_CHECKER_VIOLATION", err.Error())
			return
		}
		if err == domain.ErrUnauthorizedApprover {
			writeError(w, http.StatusForbidden, "UNAUTHORIZED_APPROVER", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "DECISION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": decision})
}

func (h *WorkflowHandler) ListDelegations(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	delegations, err := h.workflowService.ListMyDelegations(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": delegations})
}

func (h *WorkflowHandler) CreateDelegation(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CreateDelegationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	delegateID, err := uuid.Parse(req.DelegateID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "Invalid delegate_id UUID")
		return
	}

	var leID *uuid.UUID
	if req.LegalEntityID != nil && *req.LegalEntityID != "" {
		if parsed, err := uuid.Parse(*req.LegalEntityID); err == nil {
			leID = &parsed
		}
	}

	del, err := h.workflowService.CreateDelegation(r.Context(), ports.CreateDelegationCommand{
		DelegatorID:   user.ID,
		DelegateID:    delegateID,
		LegalEntityID: leID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Reason:        req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "DELEGATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": del})
}

func (h *WorkflowHandler) RevokeDelegation(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid delegation UUID")
		return
	}

	if err := h.workflowService.RevokeDelegation(r.Context(), id, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "REVOKE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "revoked"})
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
