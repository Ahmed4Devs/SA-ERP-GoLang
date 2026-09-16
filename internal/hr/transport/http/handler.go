package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/hr/domain"
	"erp/internal/hr/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type HRHandler struct {
	service ports.HRService
}

func NewHRHandler(svc ports.HRService) *HRHandler {
	return &HRHandler{service: svc}
}

func (h *HRHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	if authMiddleware != nil {
		r.Use(authMiddleware)
	}

	// Jobs & Positions
	r.Get("/jobs", h.listJobs)
	r.Post("/jobs", h.createJob)
	r.Get("/positions", h.listPositions)
	r.Post("/positions", h.createPosition)

	// Employees
	r.Get("/employees", h.listEmployees)
	r.Post("/employees", h.onboardEmployee)
	r.Get("/employees/{id}", h.getEmployee)
	r.Post("/employees/{id}/terminate", h.terminateEmployee)

	// Contracts
	r.Get("/contracts", h.listContracts)
	r.Post("/contracts", h.createContract)
	r.Get("/contracts/{id}", h.getContract)
	r.Post("/contracts/{id}/amend", h.amendContract)
	r.Post("/contracts/{id}/activate", h.activateContract)

	// Analytics
	r.Get("/saudization", h.getSaudizationMetrics)

	return r
}

func (h *HRHandler) listJobs(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	jobs, err := h.service.ListJobs(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (h *HRHandler) createJob(w http.ResponseWriter, r *http.Request) {
	var req domain.Job
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	job, err := h.service.CreateJob(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (h *HRHandler) listPositions(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	positions, err := h.service.ListPositions(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, positions)
}

func (h *HRHandler) createPosition(w http.ResponseWriter, r *http.Request) {
	var req domain.Position
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	pos, err := h.service.CreatePosition(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, pos)
}

func (h *HRHandler) listEmployees(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	employees, err := h.service.ListEmployees(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, employees)
}

type OnboardEmployeeRequest struct {
	Employee    domain.Employee            `json:"employee"`
	Identifier  *domain.EmployeeIdentifier `json:"identifier,omitempty"`
	Contact     *domain.EmployeeContact    `json:"contact,omitempty"`
	BankAccount *domain.EmployeeBankAccount`json:"bank_account,omitempty"`
}

func (h *HRHandler) onboardEmployee(w http.ResponseWriter, r *http.Request) {
	var req OnboardEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	emp, err := h.service.OnboardEmployee(r.Context(), &req.Employee, req.Identifier, req.Contact, req.BankAccount)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, emp)
}

func (h *HRHandler) getEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	emp, err := h.service.GetEmployee(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

func (h *HRHandler) terminateEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	emp, err := h.service.TerminateEmployee(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

type CreateContractRequest struct {
	Contract       domain.Contract        `json:"contract"`
	InitialVersion *domain.ContractVersion `json:"initial_version,omitempty"`
}

func (h *HRHandler) createContract(w http.ResponseWriter, r *http.Request) {
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	c, ver, err := h.service.CreateContract(r.Context(), &req.Contract, req.InitialVersion)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"contract": c, "version": ver})
}

func (h *HRHandler) getContract(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	c, err := h.service.GetContract(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *HRHandler) listContracts(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	contracts, err := h.service.ListContracts(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, contracts)
}

func (h *HRHandler) amendContract(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	var req domain.ContractVersion
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	ver, err := h.service.AmendContractVersion(r.Context(), leID, id, &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, ver)
}

func (h *HRHandler) activateContract(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	c, err := h.service.ActivateContract(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *HRHandler) getSaudizationMetrics(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	metrics, err := h.service.GetSaudizationMetrics(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
