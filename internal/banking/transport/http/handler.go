package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/banking/domain"
	"erp/internal/banking/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BankingHandler struct {
	service ports.BankingService
}

func NewBankingHandler(svc ports.BankingService) *BankingHandler {
	return &BankingHandler{service: svc}
}

func (h *BankingHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	if authMiddleware != nil {
		r.Use(authMiddleware)
	}

	// Bank Accounts
	r.Get("/bank-accounts", h.listBankAccounts)
	r.Post("/bank-accounts", h.createBankAccount)
	r.Get("/bank-accounts/{id}", h.getBankAccount)

	// Statements
	r.Get("/statements", h.listStatements)
	r.Post("/statements/import", h.importStatement)
	r.Get("/statements/{id}", h.getStatement)
	r.Get("/statements/{id}/lines", h.listStatementLines)

	// Reconciliations
	r.Get("/reconciliations", h.listReconciliations)
	r.Post("/reconciliations", h.createReconciliation)
	r.Get("/reconciliations/{id}", h.getReconciliation)
	r.Post("/reconciliations/{id}/auto-match", h.runAutoMatch)
	r.Post("/reconciliations/{id}/matches", h.addManualMatch)
	r.Get("/reconciliations/{id}/matches", h.listMatches)
	r.Post("/reconciliations/{id}/lock", h.lockReconciliation)

	return r
}

func (h *BankingHandler) listBankAccounts(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	accounts, err := h.service.ListBankAccounts(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *BankingHandler) createBankAccount(w http.ResponseWriter, r *http.Request) {
	var req domain.BankAccount
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	acc, err := h.service.CreateBankAccount(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

func (h *BankingHandler) getBankAccount(w http.ResponseWriter, r *http.Request) {
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
	acc, err := h.service.GetBankAccount(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

type ImportStatementRequest struct {
	Statement domain.Statement        `json:"statement"`
	Lines     []*domain.StatementLine `json:"lines"`
}

func (h *BankingHandler) importStatement(w http.ResponseWriter, r *http.Request) {
	var req ImportStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	stmt, err := h.service.ImportBankStatement(r.Context(), &req.Statement, req.Lines)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, stmt)
}

func (h *BankingHandler) listStatements(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	accIDStr := r.URL.Query().Get("bank_account_id")
	accID, err := uuid.Parse(accIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing bank_account_id"}`, http.StatusBadRequest)
		return
	}
	stmts, err := h.service.ListBankStatements(r.Context(), leID, accID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, stmts)
}

func (h *BankingHandler) getStatement(w http.ResponseWriter, r *http.Request) {
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
	stmt, err := h.service.GetBankStatement(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, stmt)
}

func (h *BankingHandler) listStatementLines(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	lines, err := h.service.ListStatementLines(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, lines)
}

func (h *BankingHandler) listReconciliations(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	accIDStr := r.URL.Query().Get("bank_account_id")
	accID, err := uuid.Parse(accIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing bank_account_id"}`, http.StatusBadRequest)
		return
	}
	recons, err := h.service.ListReconciliations(r.Context(), leID, accID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, recons)
}

func (h *BankingHandler) createReconciliation(w http.ResponseWriter, r *http.Request) {
	var req domain.Reconciliation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	recon, err := h.service.CreateReconciliation(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, recon)
}

func (h *BankingHandler) getReconciliation(w http.ResponseWriter, r *http.Request) {
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
	recon, err := h.service.GetReconciliation(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, recon)
}

type AutoMatchRequest struct {
	Candidates []domain.CandidateLedgerEntry `json:"candidates"`
}

func (h *BankingHandler) runAutoMatch(w http.ResponseWriter, r *http.Request) {
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
	var req AutoMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	results, err := h.service.RunAutoMatch(r.Context(), leID, id, req.Candidates)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (h *BankingHandler) addManualMatch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	var req domain.ReconciliationMatch
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	req.ReconciliationID = id
	created, err := h.service.AddManualMatch(r.Context(), leID, &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *BankingHandler) listMatches(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	matches, err := h.service.ListMatches(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, matches)
}

func (h *BankingHandler) lockReconciliation(w http.ResponseWriter, r *http.Request) {
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
	recon, err := h.service.LockReconciliation(r.Context(), leID, id, uuid.Nil)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, recon)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
