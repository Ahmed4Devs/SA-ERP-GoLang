package http

import (
	"encoding/json"
	iamhttp "erp/internal/iam/transport/http"
	"net/http"
	"time"

	"erp/internal/payroll/domain"
	"erp/internal/payroll/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PayrollHandler struct {
	svc ports.Service
}

func NewPayrollHandler(svc ports.Service) *PayrollHandler {
	return &PayrollHandler{svc: svc}
}

func (h *PayrollHandler) Routes(authMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		if authMiddleware != nil {
			protected.Use(authMiddleware)
		}

		// Components & Structures
		protected.Get("/components", h.ListComponents)
		protected.Post("/components", h.CreateComponent)
		protected.Get("/structures", h.ListStructures)
		protected.Post("/structures", h.CreateStructure)
		protected.Post("/assignments", h.AssignSalary)

		// Periods
		protected.Get("/periods", h.ListPeriods)
		protected.Post("/periods", h.CreatePeriod)

		// Runs
		protected.Get("/runs", h.ListRuns)
		protected.Post("/runs", h.CreateRun)
		protected.Get("/runs/{id}", h.GetRun)
		protected.Post("/runs/{id}/calculate", h.CalculateRun)
		protected.Post("/runs/{id}/approve", h.ApproveRun)
		protected.Post("/runs/{id}/post", h.PostRun)
		protected.Get("/runs/{id}/wps", h.GenerateWps)

		// Payslips
		protected.Get("/payslips", h.ListPayslips)
		protected.Get("/payslips/{id}", h.GetPayslip)

		// Loans
		protected.Get("/loans", h.ListLoans)
		protected.Post("/loans", h.CreateLoan)

		// End of Service
		protected.Post("/eos/calculate", h.CalculateEndOfService)
		protected.Get("/eos", h.ListEndOfService)
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

// 1. Components
func (h *PayrollHandler) ListComponents(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}
	comps, err := h.svc.ListComponents(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comps)
}

type CreateComponentRequest struct {
	LegalEntityID   string          `json:"legal_entity_id"`
	Code            string          `json:"code"`
	NameEn          string          `json:"name_en"`
	NameAr          string          `json:"name_ar"`
	ComponentType   string          `json:"component_type"`
	CalculationType string          `json:"calculation_type"`
	DefaultAmount   decimal.Decimal `json:"default_amount"`
}

func (h *PayrollHandler) CreateComponent(w http.ResponseWriter, r *http.Request) {
	var req CreateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}
	comp, err := h.svc.CreateComponent(r.Context(), leID, req.Code, req.NameEn, req.NameAr, domain.ComponentType(req.ComponentType), req.CalculationType, req.DefaultAmount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, comp)
}

// 2. Structures
func (h *PayrollHandler) ListStructures(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}
	structs, err := h.svc.ListStructures(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, structs)
}

type CreateStructureRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	Code          string `json:"code"`
	NameEn        string `json:"name_en"`
	NameAr        string `json:"name_ar"`
	Lines         []struct {
		ComponentID    string          `json:"component_id"`
		RatePercentage decimal.Decimal `json:"rate_percentage"`
		FixedAmount    decimal.Decimal `json:"fixed_amount"`
	} `json:"lines"`
}

func (h *PayrollHandler) CreateStructure(w http.ResponseWriter, r *http.Request) {
	var req CreateStructureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leID, err := uuid.Parse(req.LegalEntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid legal_entity_id")
		return
	}

	var lines []domain.SalaryStructureLine
	for _, l := range req.Lines {
		cID, err := uuid.Parse(l.ComponentID)
		if err == nil {
			lines = append(lines, domain.SalaryStructureLine{
				ComponentID:    cID,
				RatePercentage: l.RatePercentage,
				FixedAmount:    l.FixedAmount,
			})
		}
	}

	created, err := h.svc.CreateStructure(r.Context(), leID, req.Code, req.NameEn, req.NameAr, lines)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *PayrollHandler) AssignSalary(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LegalEntityID      string          `json:"legal_entity_id"`
		EmployeeID         string          `json:"employee_id"`
		SalaryStructureID  string          `json:"salary_structure_id"`
		EffectiveFrom      string          `json:"effective_from"`
		BasicSalary        decimal.Decimal `json:"basic_salary"`
		HousingAllowance   decimal.Decimal `json:"housing_allowance"`
		TransportAllowance decimal.Decimal `json:"transport_allowance"`
		OtherAllowances    decimal.Decimal `json:"other_allowances"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	leID, _ := uuid.Parse(req.LegalEntityID)
	empID, _ := uuid.Parse(req.EmployeeID)
	stID, _ := uuid.Parse(req.SalaryStructureID)
	effFrom, _ := time.Parse("2006-01-02", req.EffectiveFrom)

	created, err := h.svc.AssignSalary(r.Context(), domain.EmployeeSalaryAssignment{
		LegalEntityID:      leID,
		EmployeeID:         empID,
		SalaryStructureID:  stID,
		EffectiveFrom:      effFrom,
		BasicSalary:        req.BasicSalary,
		HousingAllowance:   req.HousingAllowance,
		TransportAllowance: req.TransportAllowance,
		OtherAllowances:    req.OtherAllowances,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// 3. Periods
func (h *PayrollHandler) ListPeriods(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}
	periods, err := h.svc.ListPeriods(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, periods)
}

func (h *PayrollHandler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LegalEntityID  string `json:"legal_entity_id"`
		Year           int    `json:"year"`
		Month          int    `json:"month"`
		NameEn         string `json:"name_en"`
		NameAr         string `json:"name_ar"`
		StartDate      string `json:"start_date"`
		EndDate        string `json:"end_date"`
		PaymentDueDate string `json:"payment_due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leID, _ := uuid.Parse(req.LegalEntityID)
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	due, _ := time.Parse("2006-01-02", req.PaymentDueDate)

	created, err := h.svc.CreatePeriod(r.Context(), leID, req.Year, req.Month, req.NameEn, req.NameAr, start, end, due)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// 4. Runs
func (h *PayrollHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id query parameter is required")
		return
	}
	runs, err := h.svc.ListRuns(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (h *PayrollHandler) CreateRun(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil || actor.ID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req struct {
		LegalEntityID string `json:"legal_entity_id"`
		PeriodID      string `json:"period_id"`
		RunNumber     string `json:"run_number"`
		RunDate       string `json:"run_date"`
		CutoffDate    string `json:"cutoff_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leID, _ := uuid.Parse(req.LegalEntityID)
	pID, _ := uuid.Parse(req.PeriodID)
	rDate, _ := time.Parse("2006-01-02", req.RunDate)
	cDate, _ := time.Parse("2006-01-02", req.CutoffDate)

	created, err := h.svc.CreateRun(r.Context(), leID, pID, req.RunNumber, rDate, cDate, &actor.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *PayrollHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	run, err := h.svc.GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (h *PayrollHandler) CalculateRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	calculated, err := h.svc.CalculateRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calculated)
}

func (h *PayrollHandler) ApproveRun(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil || actor.ID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	approved, err := h.svc.ApproveRun(r.Context(), runID, actor.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, approved)
}

func (h *PayrollHandler) PostRun(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil || actor.ID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	posted, err := h.svc.PostRun(r.Context(), runID, &actor.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, posted)
}

func (h *PayrollHandler) GenerateWps(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	molID := r.URL.Query().Get("employer_mol_id")
	if molID == "" {
		molID = "7001234567"
	}
	bankCode := r.URL.Query().Get("bank_code")
	if bankCode == "" {
		bankCode = "NCB"
	}

	content, fileName, err := h.svc.GenerateWps(r.Context(), runID, molID, bankCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"file_name":    fileName,
		"file_content": content,
	})
}

// 5. Payslips
func (h *PayrollHandler) ListPayslips(w http.ResponseWriter, r *http.Request) {
	periodParam := r.URL.Query().Get("period_id")
	if periodParam != "" {
		pID, err := uuid.Parse(periodParam)
		if err == nil {
			slips, err := h.svc.GetPayslipsByPeriod(r.Context(), pID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, slips)
			return
		}
	}
	writeError(w, http.StatusBadRequest, "period_id query parameter is required")
}

func (h *PayrollHandler) GetPayslip(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid payslip id")
		return
	}
	slip, err := h.svc.GetPayslipByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "payslip not found")
		return
	}
	writeJSON(w, http.StatusOK, slip)
}

// 6. Loans
func (h *PayrollHandler) ListLoans(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id is required")
		return
	}
	loans, err := h.svc.ListLoans(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, loans)
}

func (h *PayrollHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LegalEntityID      string          `json:"legal_entity_id"`
		EmployeeID         string          `json:"employee_id"`
		LoanNumber         string          `json:"loan_number"`
		PrincipalAmount    decimal.Decimal `json:"principal_amount"`
		MonthlyInstallment decimal.Decimal `json:"monthly_installment"`
		DisbursementDate   string          `json:"disbursement_date"`
		Reason             *string         `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leID, _ := uuid.Parse(req.LegalEntityID)
	empID, _ := uuid.Parse(req.EmployeeID)
	disbDate, _ := time.Parse("2006-01-02", req.DisbursementDate)

	loan, err := h.svc.CreateLoan(r.Context(), leID, empID, req.LoanNumber, req.PrincipalAmount, req.MonthlyInstallment, disbDate, req.Reason)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, loan)
}

// 7. End of Service
func (h *PayrollHandler) CalculateEndOfService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LegalEntityID   string `json:"legal_entity_id"`
		EmployeeID      string `json:"employee_id"`
		TerminationDate string `json:"termination_date"`
		Reason          string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	leID, _ := uuid.Parse(req.LegalEntityID)
	empID, _ := uuid.Parse(req.EmployeeID)
	termDate, _ := time.Parse("2006-01-02", req.TerminationDate)

	eos, err := h.svc.CalculateEndOfService(r.Context(), leID, empID, termDate, domain.TerminationReason(req.Reason))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, eos)
}

func (h *PayrollHandler) ListEndOfService(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "legal_entity_id is required")
		return
	}
	eoses, err := h.svc.ListEndOfService(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, eoses)
}
