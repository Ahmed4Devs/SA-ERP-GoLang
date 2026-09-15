package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/tax/domain"
	"erp/internal/tax/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaxHandler struct {
	taxService ports.TaxService
}

func NewTaxHandler(taxService ports.TaxService) *TaxHandler {
	return &TaxHandler{
		taxService: taxService,
	}
}

func (h *TaxHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Categories & Reasons
		r.Get("/categories", h.ListCategories)
		r.Get("/reasons/exemption", h.ListExemptionReasons)
		r.Get("/reasons/zero-rate", h.ListZeroRateReasons)

		// Tax Codes & Rates
		r.Get("/codes", h.ListTaxCodes)
		r.Post("/codes", h.CreateTaxCode)
		r.Get("/codes/{id}", h.GetTaxCode)
		r.Post("/codes/{id}/rates", h.AddRateVersion)

		// Calculation Preview
		r.Post("/calculate", h.CalculateTax)

		// Tax Ledger Transactions
		r.Get("/transactions", h.ListTaxTransactions)
		r.Get("/transactions/{id}", h.GetTaxTransaction)
		r.Post("/transactions", h.PostTaxTransaction)

		// Reporting Periods
		r.Get("/periods", h.ListPeriods)
		r.Post("/periods", h.CreatePeriod)
		r.Post("/periods/{id}/close", h.ClosePeriod)

		// VAT Returns (ZATCA Form)
		r.Get("/returns", h.ListVATReturns)
		r.Post("/returns/draft", h.GenerateVATReturnDraft)
		r.Post("/returns/{id}/submit", h.SubmitVATReturn)
	})

	return r
}

func (h *TaxHandler) resolveLegalEntityID(r *http.Request) (uuid.UUID, error) {
	if header := r.Header.Get("X-Legal-Entity-ID"); header != "" {
		if u, err := uuid.Parse(header); err == nil && u != uuid.Nil {
			return u, nil
		}
	}

	leParam := r.URL.Query().Get("legal_entity_id")
	if leParam != "" {
		return uuid.Parse(leParam)
	}

	return uuid.Nil, errors.New("missing legal entity context (X-Legal-Entity-ID header or legal_entity_id query)")
}

// 1. Categories & Reasons
func (h *TaxHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.taxService.ListCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": cats})
}

func (h *TaxHandler) ListExemptionReasons(w http.ResponseWriter, r *http.Request) {
	reasons, err := h.taxService.ListExemptionReasons(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": reasons})
}

func (h *TaxHandler) ListZeroRateReasons(w http.ResponseWriter, r *http.Request) {
	reasons, err := h.taxService.ListZeroRateReasons(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": reasons})
}

// 2. Tax Codes & Rates
func (h *TaxHandler) ListTaxCodes(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	codes, err := h.taxService.ListTaxCodes(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": codes})
}

func (h *TaxHandler) GetTaxCode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tax code UUID")
		return
	}

	code, err := h.taxService.GetTaxCode(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": code})
}

func (h *TaxHandler) CreateTaxCode(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())

	var req CreateTaxCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	leID := req.LegalEntityID
	if leID == uuid.Nil {
		var err error
		leID, err = h.resolveLegalEntityID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
			return
		}
	}

	effectiveFrom := time.Now().UTC()
	if req.EffectiveFrom != nil {
		effectiveFrom = *req.EffectiveFrom
	}

	var actorID *uuid.UUID
	if user != nil {
		actorID = &user.ID
	}

	code := domain.TaxCode{
		LegalEntityID:        leID,
		Code:                 req.Code,
		NameEn:               req.NameEn,
		NameAr:               req.NameAr,
		TaxCategoryCode:      req.TaxCategoryCode,
		TaxType:              req.TaxType,
		SalesTaxAccountID:    req.SalesTaxAccountID,
		PurchaseTaxAccountID: req.PurchaseTaxAccountID,
		IsRecoverable:        req.IsRecoverable,
		IsActive:             true,
	}

	created, err := h.taxService.CreateTaxCode(r.Context(), code, req.InitialRate, effectiveFrom, actorID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *TaxHandler) AddRateVersion(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	taxCodeID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tax code UUID")
		return
	}

	var req AddRateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	var actorID *uuid.UUID
	if user != nil {
		actorID = &user.ID
	}

	ver := domain.TaxRateVersion{
		TaxCodeID:     taxCodeID,
		RatePercent:   req.RatePercent,
		EffectiveFrom: req.EffectiveFrom,
		EffectiveTo:   req.EffectiveTo,
		Description:   req.Description,
		CreatedBy:     actorID,
	}

	created, err := h.taxService.CreateRateVersion(r.Context(), ver)
	if err != nil {
		writeError(w, http.StatusBadRequest, "FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 3. Calculation Preview
func (h *TaxHandler) CalculateTax(w http.ResponseWriter, r *http.Request) {
	var req CalculateTaxRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	date := time.Now().UTC()
	if req.TransactionDate != nil {
		date = *req.TransactionDate
	}

	inputs := make([]domain.TaxLineCalculationInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		inputs = append(inputs, domain.TaxLineCalculationInput{
			LineNumber:          l.LineNumber,
			TaxCodeID:           l.TaxCodeID,
			Quantity:            l.Quantity,
			UnitPrice:           l.UnitPrice,
			DiscountAmount:      l.DiscountAmount,
			ExemptionReasonCode: l.ExemptionReasonCode,
			ZeroRateReasonCode:  l.ZeroRateReasonCode,
		})
	}

	result, err := h.taxService.CalculateTax(r.Context(), ports.CalculateTaxRequest{
		LegalEntityID:   req.LegalEntityID,
		TransactionDate: date,
		Lines:           inputs,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "CALCULATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// 4. Tax Transactions
func (h *TaxHandler) ListTaxTransactions(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	limit := int32(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	offset := int32(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	txs, err := h.taxService.ListTaxTransactions(r.Context(), leID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": txs})
}

func (h *TaxHandler) GetTaxTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transaction UUID")
		return
	}

	tx, err := h.taxService.GetTaxTransaction(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if tx == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Tax transaction not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": tx})
}

func (h *TaxHandler) PostTaxTransaction(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())

	var req PostTaxTransactionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	leID := req.LegalEntityID
	if leID == uuid.Nil {
		var err error
		leID, err = h.resolveLegalEntityID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
			return
		}
	}

	date := time.Now().UTC()
	if req.TransactionDate != nil {
		date = *req.TransactionDate
	}

	var actorID *uuid.UUID
	if user != nil {
		actorID = &user.ID
	}

	lines := make([]domain.TaxTransactionLine, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, domain.TaxTransactionLine{
			TaxCodeID:           l.TaxCodeID,
			TaxableAmount:       l.TaxableAmount,
			ExemptionReasonCode: l.ExemptionReasonCode,
			ZeroRateReasonCode:  l.ZeroRateReasonCode,
		})
	}

	created, err := h.taxService.PostTaxTransaction(r.Context(), ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      req.SourceModule,
		SourceType:        req.SourceType,
		SourceID:          req.SourceID,
		TransactionNumber: req.TransactionNumber,
		TransactionDate:   date,
		TransactionType:   req.TransactionType,
		PartnerID:         req.PartnerID,
		PartnerName:       req.PartnerName,
		PartnerTaxNumber:  req.PartnerTaxNumber,
		JournalEntryID:    req.JournalEntryID,
		CreatedBy:         actorID,
		Lines:             lines,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "POST_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 5. Reporting Periods
func (h *TaxHandler) ListPeriods(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	periods, err := h.taxService.ListPeriods(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": periods})
}

func (h *TaxHandler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
	var req CreateTaxPeriodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	leID := req.LegalEntityID
	if leID == uuid.Nil {
		var err error
		leID, err = h.resolveLegalEntityID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
			return
		}
	}

	created, err := h.taxService.CreatePeriod(r.Context(), domain.TaxPeriod{
		LegalEntityID: leID,
		PeriodNumber:  req.PeriodNumber,
		FiscalYear:    req.FiscalYear,
		PeriodType:    req.PeriodType,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		DueDate:       req.DueDate,
		Status:        domain.TaxPeriodStatusOpen,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *TaxHandler) ClosePeriod(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	periodID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid period UUID")
		return
	}

	closed, err := h.taxService.ClosePeriod(r.Context(), periodID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": closed})
}

// 6. VAT Returns
func (h *TaxHandler) ListVATReturns(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	returns, err := h.taxService.ListVATReturns(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": returns})
}

func (h *TaxHandler) GenerateVATReturnDraft(w http.ResponseWriter, r *http.Request) {
	var req GenerateVATReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	leID := req.LegalEntityID
	if leID == uuid.Nil {
		var err error
		leID, err = h.resolveLegalEntityID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
			return
		}
	}

	draft, err := h.taxService.GenerateVATReturnDraft(r.Context(), leID, req.TaxPeriodID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "DRAFT_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": draft})
}

func (h *TaxHandler) SubmitVATReturn(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	returnID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid return UUID")
		return
	}

	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	submitted, err := h.taxService.SubmitVATReturn(r.Context(), leID, returnID, user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "SUBMISSION_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": submitted})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
