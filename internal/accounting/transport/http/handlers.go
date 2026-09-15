package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"
	iamhttp "erp/internal/iam/transport/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AccountingHandler struct {
	accountingService ports.AccountingService
	ledgerService     ports.LedgerService
}

func NewAccountingHandler(accountingService ports.AccountingService, ledgerService ports.LedgerService) *AccountingHandler {
	return &AccountingHandler{
		accountingService: accountingService,
		ledgerService:     ledgerService,
	}
}

func (h *AccountingHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Chart of Accounts & Accounts
		r.Get("/chart-of-accounts", h.ListChartOfAccounts)
		r.Get("/accounts", h.ListAccounts)
		r.Post("/accounts", h.CreateAccount)
		r.Get("/accounts/{id}", h.GetAccount)
		r.Put("/accounts/{id}", h.UpdateAccount)

		// Fiscal Years & Periods
		r.Get("/fiscal-years", h.ListFiscalYears)
		r.Get("/fiscal-years/{id}/periods", h.ListFiscalPeriods)
		r.Put("/fiscal-periods/{id}/status", h.UpdateFiscalPeriodStatus)

		// Journals
		r.Get("/journals", h.ListJournals)

		// Journal Entries
		r.Get("/journal-entries", h.ListJournalEntries)
		r.Post("/journal-entries", h.CreateJournalEntry)
		r.Get("/journal-entries/{id}", h.GetJournalEntry)
		r.Post("/journal-entries/{id}/post", h.PostJournalEntry)
		r.Post("/journal-entries/{id}/reverse", h.ReverseJournalEntry)

		// Reports
		r.Get("/reports/trial-balance", h.GetTrialBalance)
	})

	return r
}

func (h *AccountingHandler) resolveLegalEntityID(r *http.Request) (uuid.UUID, error) {
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

func (h *AccountingHandler) ListChartOfAccounts(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	coas, err := h.accountingService.ListChartOfAccounts(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": coas})
}

func (h *AccountingHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	accounts, err := h.accountingService.ListAccounts(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": accounts})
}

func (h *AccountingHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	acc, err := h.accountingService.CreateAccount(r.Context(), domain.Account{
		LegalEntityID:    leID,
		ChartID:          req.ChartID,
		ParentID:         req.ParentID,
		Code:             req.Code,
		NameEn:           req.NameEn,
		NameAr:           req.NameAr,
		AccountType:      req.AccountType,
		NormalBalance:    req.NormalBalance,
		AllowPosting:     req.AllowPosting,
		IsControlAccount: req.IsControlAccount,
		IsActive:         true,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": acc})
}

func (h *AccountingHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid account UUID")
		return
	}

	acc, err := h.accountingService.GetAccountByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": acc})
}

func (h *AccountingHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid account UUID")
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	acc, err := h.accountingService.UpdateAccount(r.Context(), domain.Account{
		ID:               id,
		NameEn:           req.NameEn,
		NameAr:           req.NameAr,
		AllowPosting:     req.AllowPosting,
		IsControlAccount: req.IsControlAccount,
		IsActive:         req.IsActive,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": acc})
}

func (h *AccountingHandler) ListFiscalYears(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	years, err := h.accountingService.ListFiscalYears(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": years})
}

func (h *AccountingHandler) ListFiscalPeriods(w http.ResponseWriter, r *http.Request) {
	yearID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid fiscal year UUID")
		return
	}

	periods, err := h.accountingService.ListFiscalPeriods(r.Context(), yearID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": periods})
}

func (h *AccountingHandler) UpdateFiscalPeriodStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid fiscal period UUID")
		return
	}

	var req UpdateFiscalPeriodStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	period, err := h.accountingService.UpdateFiscalPeriodStatus(r.Context(), id, req.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": period})
}

func (h *AccountingHandler) ListJournals(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	journals, err := h.accountingService.ListJournals(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": journals})
}

func (h *AccountingHandler) ListJournalEntries(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	limit := int32(50)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	entries, total, err := h.ledgerService.ListJournalEntries(r.Context(), leID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": entries, "total": total, "limit": limit, "offset": offset})
}

func (h *AccountingHandler) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	var req CreateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON payload")
		return
	}

	accountingDate := time.Now().UTC()
	if req.AccountingDate != nil {
		accountingDate = *req.AccountingDate
	}

	var periodID uuid.UUID
	if req.FiscalPeriodID != nil {
		periodID = *req.FiscalPeriodID
	}

	lines := make([]domain.JournalLine, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, domain.JournalLine{
			AccountID:         l.AccountID,
			Description:       l.Description,
			CurrencyCode:      l.CurrencyCode,
			ExchangeRate:      l.ExchangeRate,
			TransactionDebit:  l.TransactionDebit,
			TransactionCredit: l.TransactionCredit,
			BranchID:          l.BranchID,
			DepartmentID:      l.DepartmentID,
			CostCenterID:      l.CostCenterID,
		})
	}

	entry, err := h.ledgerService.CreateJournalEntry(r.Context(), domain.JournalEntry{
		LegalEntityID:  leID,
		JournalID:      req.JournalID,
		FiscalPeriodID: periodID,
		EntryNumber:    req.EntryNumber,
		AccountingDate: accountingDate,
		Reference:      req.Reference,
		Description:    req.Description,
		CurrencyCode:   req.CurrencyCode,
		ExchangeRate:   req.ExchangeRate,
		CreatedBy:      user.ID,
	}, lines)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": entry})
}

func (h *AccountingHandler) GetJournalEntry(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid journal entry UUID")
		return
	}

	entry, err := h.ledgerService.GetJournalEntryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrJournalEntryNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Journal entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": entry})
}

func (h *AccountingHandler) PostJournalEntry(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid journal entry UUID")
		return
	}

	entry, err := h.ledgerService.PostJournalEntry(r.Context(), id, user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "POST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": entry})
}

func (h *AccountingHandler) ReverseJournalEntry(w http.ResponseWriter, r *http.Request) {
	user := iamhttp.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid journal entry UUID")
		return
	}

	var req ReverseJournalEntryRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "طلب إلغاء وعكس القيد المحاسبي"
	}

	rev, err := h.ledgerService.ReverseJournalEntry(r.Context(), id, req.Reason, user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REVERSE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": rev})
}

func (h *AccountingHandler) GetTrialBalance(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", "Legal entity context required")
		return
	}

	var asOfDate *time.Time
	if d := r.URL.Query().Get("as_of_date"); d != "" {
		if parsed, err := time.Parse("2006-01-02", d); err == nil {
			asOfDate = &parsed
		}
	}

	tb, err := h.ledgerService.GetTrialBalance(r.Context(), leID, asOfDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": tb})
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
