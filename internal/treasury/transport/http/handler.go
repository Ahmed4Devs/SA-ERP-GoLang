package http

import (
	"encoding/json"
	"net/http"

	"erp/internal/treasury/domain"
	"erp/internal/treasury/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TreasuryHandler struct {
	service ports.TreasuryService
}

func NewTreasuryHandler(svc ports.TreasuryService) *TreasuryHandler {
	return &TreasuryHandler{service: svc}
}

func (h *TreasuryHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	if authMiddleware != nil {
		r.Use(authMiddleware)
	}

	// Cash accounts
	r.Get("/cash-accounts", h.listCashAccounts)
	r.Post("/cash-accounts", h.createCashAccount)
	r.Get("/cash-accounts/{id}", h.getCashAccount)

	// Receipts
	r.Get("/receipts", h.listReceipts)
	r.Post("/receipts", h.createReceipt)
	r.Post("/receipts/{id}/post", h.postReceipt)

	// Payments
	r.Get("/payments", h.listPayments)
	r.Post("/payments", h.createPayment)
	r.Post("/payments/{id}/approve", h.approvePayment)
	r.Post("/payments/{id}/pay", h.payPayment)

	// Transfers
	r.Get("/transfers", h.listTransfers)
	r.Post("/transfers", h.createTransfer)
	r.Post("/transfers/{id}/complete", h.completeTransfer)

	return r
}

func (h *TreasuryHandler) listCashAccounts(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	accounts, err := h.service.ListCashAccounts(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *TreasuryHandler) createCashAccount(w http.ResponseWriter, r *http.Request) {
	var req domain.CashAccount
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	account, err := h.service.CreateCashAccount(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, account)
}

func (h *TreasuryHandler) getCashAccount(w http.ResponseWriter, r *http.Request) {
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
	acc, err := h.service.GetCashAccount(r.Context(), leID, id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (h *TreasuryHandler) listReceipts(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	receipts, err := h.service.ListReceipts(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, receipts)
}

func (h *TreasuryHandler) createReceipt(w http.ResponseWriter, r *http.Request) {
	var req domain.Receipt
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	rc, err := h.service.CreateReceipt(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, rc)
}

func (h *TreasuryHandler) postReceipt(w http.ResponseWriter, r *http.Request) {
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
	rc, err := h.service.PostReceipt(r.Context(), leID, id, nil)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, rc)
}

func (h *TreasuryHandler) listPayments(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	pmts, err := h.service.ListPayments(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, pmts)
}

func (h *TreasuryHandler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req domain.Payment
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	pmt, err := h.service.CreatePayment(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, pmt)
}

func (h *TreasuryHandler) approvePayment(w http.ResponseWriter, r *http.Request) {
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
	pmt, err := h.service.ApprovePayment(r.Context(), leID, id, uuid.Nil)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, pmt)
}

func (h *TreasuryHandler) payPayment(w http.ResponseWriter, r *http.Request) {
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
	pmt, err := h.service.PayPayment(r.Context(), leID, id, nil)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, pmt)
}

func (h *TreasuryHandler) listTransfers(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		http.Error(w, `{"error":"invalid or missing legal_entity_id"}`, http.StatusBadRequest)
		return
	}
	trfs, err := h.service.ListTransfers(r.Context(), leID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, trfs)
}

func (h *TreasuryHandler) createTransfer(w http.ResponseWriter, r *http.Request) {
	var req domain.Transfer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}
	trf, err := h.service.CreateTransfer(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, trf)
}

func (h *TreasuryHandler) completeTransfer(w http.ResponseWriter, r *http.Request) {
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
	trf, err := h.service.CompleteTransfer(r.Context(), leID, id, uuid.Nil, nil)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, trf)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
