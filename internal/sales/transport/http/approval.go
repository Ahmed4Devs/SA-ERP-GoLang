package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

func (h *SalesHandler) SubmitInvoice(w http.ResponseWriter, r *http.Request) {
	actor := salesActorID(r)
	if actor == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid invoice", http.StatusBadRequest)
		return
	}
	invoice, err := h.salesService.SubmitInvoice(r.Context(), id, *actor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}
