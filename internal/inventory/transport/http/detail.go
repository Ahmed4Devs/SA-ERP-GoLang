package http

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

func (h *InventoryHandler) GetAdjustment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid adjustment", 400)
		return
	}
	entity, err := h.resolveLegalEntityID(r)
	if err != nil {
		http.Error(w, "Legal entity is required", 400)
		return
	}
	item, err := h.inventoryService.GetAdjustment(r.Context(), id)
	if err != nil || item.LegalEntityID != entity {
		http.Error(w, "Adjustment not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}
func (h *InventoryHandler) GetStockCount(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid stock count", 400)
		return
	}
	entity, err := h.resolveLegalEntityID(r)
	if err != nil {
		http.Error(w, "Legal entity is required", 400)
		return
	}
	item, err := h.inventoryService.GetStockCount(r.Context(), id)
	if err != nil || item.LegalEntityID != entity {
		http.Error(w, "Stock count not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}
