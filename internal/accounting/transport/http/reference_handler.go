package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"erp/internal/accounting/application/services"
	iamhttp "erp/internal/iam/transport/http"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (h *AccountingHandler) registerReferenceRoutes(r chi.Router) {
	if h.ReferenceService == nil {
		return
	}
	for _, resource := range []string{"dimensions", "dimension-values", "exchange-rates", "posting-rules", "opening-balances", "reconciliations"} {
		r.Get("/"+resource, h.listReference(resource))
		r.Post("/"+resource, h.mutateReference(resource))
	}
	r.Post("/journal-lines/{id}/dimensions", h.mutateReference("line-dimension"))
	r.Post("/opening-balances/{id}/post", h.mutateReference("post-opening"))
	r.Post("/reconciliations/{id}/complete", h.mutateReference("complete-reconciliation"))
}

func (h *AccountingHandler) listReference(resource string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entity, err := h.resolveLegalEntityID(r)
		if err != nil || entity == uuid.Nil {
			writeError(w, 400, "INVALID_SCOPE", "Legal entity context required")
			return
		}
		limit, offset := 50, 0
		if value := r.URL.Query().Get("limit"); value != "" {
			limit, err = strconv.Atoi(value)
			if err != nil || limit < 1 || limit > 200 {
				writeError(w, 400, "INVALID_PAGE", "Limit must be between 1 and 200")
				return
			}
		}
		if value := r.URL.Query().Get("offset"); value != "" {
			offset, err = strconv.Atoi(value)
			if err != nil || offset < 0 {
				writeError(w, 400, "INVALID_PAGE", "Offset must be non-negative")
				return
			}
		}
		data, total, err := h.ReferenceService.List(r.Context(), entity, resource, limit, offset)
		if err != nil {
			writeError(w, 500, "READ_FAILED", "Unable to read accounting records")
			return
		}
		writeJSON(w, 200, map[string]any{"data": data, "total": total, "limit": limit, "offset": offset})
	}
}

func (h *AccountingHandler) mutateReference(resource string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actor := iamhttp.UserFromContext(r.Context())
		if actor == nil {
			writeError(w, 401, "UNAUTHORIZED", "Authenticated actor required")
			return
		}
		entity, err := h.resolveLegalEntityID(r)
		if err != nil || entity == uuid.Nil {
			writeError(w, 400, "INVALID_SCOPE", "Legal entity context required")
			return
		}
		var target, result uuid.UUID
		if value := chi.URLParam(r, "id"); value != "" {
			target, err = uuid.Parse(value)
			if err != nil {
				writeError(w, 400, "INVALID_ID", "Invalid UUID")
				return
			}
		}
		decode := func(value any) bool {
			if err = json.NewDecoder(r.Body).Decode(value); err != nil {
				writeError(w, 400, "INVALID_BODY", "Invalid JSON payload")
				return false
			}
			return true
		}
		switch resource {
		case "dimensions", "dimension-values":
			var cmd services.DimensionCommand
			if !decode(&cmd) {
				return
			}
			result, err = h.ReferenceService.CreateDimension(r.Context(), entity, actor.ID, cmd, resource == "dimension-values")
		case "exchange-rates":
			var cmd services.ExchangeRateCommand
			if !decode(&cmd) {
				return
			}
			result, err = h.ReferenceService.CreateExchangeRate(r.Context(), entity, actor.ID, cmd)
		case "posting-rules":
			var cmd services.PostingRuleCommand
			if !decode(&cmd) {
				return
			}
			result, err = h.ReferenceService.CreatePostingRule(r.Context(), entity, actor.ID, cmd)
		case "opening-balances":
			var cmd services.OpeningBalanceCommand
			if !decode(&cmd) {
				return
			}
			result, err = h.ReferenceService.CreateOpeningBalance(r.Context(), entity, actor.ID, cmd)
		case "reconciliations":
			var cmd services.ReconciliationCommand
			if !decode(&cmd) {
				return
			}
			result, err = h.ReferenceService.CreateReconciliation(r.Context(), entity, actor.ID, cmd)
		case "line-dimension":
			var cmd struct {
				DimensionID uuid.UUID `json:"dimension_id"`
				ValueID     uuid.UUID `json:"dimension_value_id"`
			}
			if !decode(&cmd) {
				return
			}
			err = h.ReferenceService.AssignDimension(r.Context(), entity, actor.ID, target, cmd.DimensionID, cmd.ValueID)
			result = target
		case "post-opening":
			var cmd struct {
				OffsetAccountID uuid.UUID       `json:"offset_account_id"`
				ExchangeRate    decimal.Decimal `json:"exchange_rate"`
			}
			if !decode(&cmd) {
				return
			}
			result, err = h.ReferenceService.PostOpeningBalance(r.Context(), entity, actor.ID, target, cmd.OffsetAccountID, cmd.ExchangeRate)
		case "complete-reconciliation":
			err = h.ReferenceService.CompleteReconciliation(r.Context(), entity, actor.ID, target)
			result = target
		}
		if err != nil {
			writeError(w, 400, "OPERATION_FAILED", err.Error())
			return
		}
		status := http.StatusCreated
		if resource == "complete-reconciliation" || resource == "line-dimension" || resource == "post-opening" {
			status = http.StatusOK
		}
		writeJSON(w, status, map[string]any{"data": map[string]any{"id": result}})
	}
}
