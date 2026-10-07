package http

import (
	"net/http"
	"strconv"

	iamhttp "erp/internal/iam/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *InventoryHandler) registerValuationRoutes(r chi.Router) {
	if h.ValuationService == nil {
		return
	}
	r.Get("/costing/item-costs", h.listValuation(false))
	r.Get("/costing/snapshots", h.listValuation(true))
	r.Post("/costing/snapshots", h.createValuationSnapshot)
}

func (h *InventoryHandler) listValuation(snapshots bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entity, err := h.resolveLegalEntityID(r)
		if err != nil {
			writeError(w, 400, "INVALID_SCOPE", "Legal entity required")
			return
		}
		limit, offset := 50, 0
		if value := r.URL.Query().Get("limit"); value != "" {
			limit, err = strconv.Atoi(value)
			if err != nil || limit < 1 || limit > 200 {
				writeError(w, 400, "INVALID_PAGE", "Invalid limit")
				return
			}
		}
		if value := r.URL.Query().Get("offset"); value != "" {
			offset, err = strconv.Atoi(value)
			if err != nil || offset < 0 {
				writeError(w, 400, "INVALID_PAGE", "Invalid offset")
				return
			}
		}
		var batch *uuid.UUID
		if value := r.URL.Query().Get("snapshot_batch_id"); value != "" {
			id, err := uuid.Parse(value)
			if err != nil {
				writeError(w, 400, "INVALID_ID", "Invalid snapshot batch")
				return
			}
			batch = &id
		}
		data, total, err := h.ValuationService.List(r.Context(), entity, snapshots, batch, limit, offset)
		if err != nil {
			writeError(w, 500, "READ_FAILED", "Unable to read valuation records")
			return
		}
		writeJSON(w, 200, map[string]any{"data": data, "total": total, "limit": limit, "offset": offset})
	}
}

func (h *InventoryHandler) createValuationSnapshot(w http.ResponseWriter, r *http.Request) {
	actor := iamhttp.UserFromContext(r.Context())
	if actor == nil {
		writeError(w, 401, "UNAUTHORIZED", "Authenticated actor required")
		return
	}
	entity, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, 400, "INVALID_SCOPE", "Legal entity required")
		return
	}
	batch, count, err := h.ValuationService.CreateSnapshot(r.Context(), entity, actor.ID)
	if err != nil {
		writeError(w, 400, "SNAPSHOT_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"data": map[string]any{"snapshot_batch_id": batch, "records": count}})
}
