package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	inventoryService ports.InventoryService
}

func NewInventoryHandler(inventoryService ports.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

func (h *InventoryHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Warehouses & Bins
		r.Get("/warehouses", h.ListWarehouses)
		r.Post("/warehouses", h.CreateWarehouse)
		r.Get("/warehouses/{id}", h.GetWarehouse)
		r.Get("/warehouses/{id}/bins", h.ListBins)
		r.Post("/warehouses/{id}/bins", h.CreateBin)

		// Movements & Ledger
		r.Get("/movements", h.ListMovements)
		r.Post("/movements", h.PostMovement)
		r.Get("/movements/{id}", h.GetMovement)

		// Balances
		r.Get("/balances", h.ListBalances)

		// Reservations
		r.Post("/reservations", h.ReserveStock)
		r.Post("/reservations/{id}/release", h.ReleaseStock)

		// Transfers
		r.Get("/transfers", h.ListTransferOrders)
		r.Post("/transfers", h.CreateTransferOrder)
		r.Get("/transfers/{id}", h.GetTransferOrder)
		r.Post("/transfers/{id}/dispatch", h.DispatchTransfer)
		r.Post("/transfers/{id}/receive", h.ReceiveTransfer)

		// Adjustments
		r.Get("/adjustments", h.ListAdjustments)
		r.Post("/adjustments", h.CreateAdjustment)
		r.Post("/adjustments/{id}/post", h.PostAdjustment)

		// Counts
		r.Get("/counts", h.ListStockCounts)
		r.Post("/counts", h.CreateStockCount)
		r.Post("/counts/{id}/reconcile", h.ReconcileStockCount)

		// Costing & Valuation
		r.Get("/costing/valuation", h.GetValuationReport)
		r.Get("/costing/layers", h.ListCostLayers)
		r.Post("/costing/revaluations", h.ExecuteRevaluation)
		r.Get("/costing/revaluations", h.ListRevaluations)
	})

	return r
}

func (h *InventoryHandler) resolveLegalEntityID(r *http.Request) (uuid.UUID, error) {
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

// 1. Warehouses & Bins
func (h *InventoryHandler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	warehouses, err := h.inventoryService.ListWarehouses(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": warehouses})
}

func (h *InventoryHandler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid warehouse UUID")
		return
	}

	wh, err := h.inventoryService.GetWarehouse(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWarehouseNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Warehouse not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": wh})
}

func (h *InventoryHandler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var dto CreateWarehouseRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.Code == "" || dto.NameEn == "" || dto.NameAr == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Code, NameEn, and NameAr are required")
		return
	}

	wh := domain.Warehouse{
		LegalEntityID: leID,
		Code:          dto.Code,
		NameEn:        dto.NameEn,
		NameAr:        dto.NameAr,
		WarehouseType: domain.WarehouseType(dto.WarehouseType),
		AddressLine1:  dto.AddressLine1,
		City:          dto.City,
		PostalCode:    dto.PostalCode,
		CountryCode:   dto.CountryCode,
	}

	created, err := h.inventoryService.CreateWarehouse(r.Context(), wh)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *InventoryHandler) ListBins(w http.ResponseWriter, r *http.Request) {
	whIDStr := chi.URLParam(r, "id")
	whID, err := uuid.Parse(whIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid warehouse UUID")
		return
	}

	bins, err := h.inventoryService.ListBins(r.Context(), whID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": bins})
}

func (h *InventoryHandler) CreateBin(w http.ResponseWriter, r *http.Request) {
	whIDStr := chi.URLParam(r, "id")
	whID, err := uuid.Parse(whIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid warehouse UUID")
		return
	}

	var dto CreateBinRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.Code == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Bin Code is required")
		return
	}

	bin := domain.BinLocation{
		WarehouseID: whID,
		Code:        dto.Code,
		Aisle:       dto.Aisle,
		Rack:        dto.Rack,
		Shelf:       dto.Shelf,
		Bin:         dto.Bin,
		IsDefault:   dto.IsDefault,
	}

	created, err := h.inventoryService.CreateBin(r.Context(), bin)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 2. Movements & Ledger
func (h *InventoryHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var itemID *uuid.UUID
	if iStr := r.URL.Query().Get("item_id"); iStr != "" {
		if id, err := uuid.Parse(iStr); err == nil {
			itemID = &id
		}
	}

	var warehouseID *uuid.UUID
	if wStr := r.URL.Query().Get("warehouse_id"); wStr != "" {
		if id, err := uuid.Parse(wStr); err == nil {
			warehouseID = &id
		}
	}

	var movType *string
	if mt := r.URL.Query().Get("movement_type"); mt != "" {
		movType = &mt
	}

	limit := int32(50)
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	offset := int32(0)
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	movements, err := h.inventoryService.ListMovements(r.Context(), leID, itemID, warehouseID, movType, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": movements})
}

func (h *InventoryHandler) GetMovement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid movement UUID")
		return
	}

	mov, err := h.inventoryService.GetMovement(r.Context(), id)
	if err != nil || mov == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Stock movement not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": mov})
}

func (h *InventoryHandler) PostMovement(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	var dto PostMovementRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.ItemID == uuid.Nil || dto.WarehouseID == uuid.Nil || dto.UOMID == uuid.Nil || dto.Quantity.IsZero() {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ItemID, WarehouseID, UOMID, and non-zero Quantity are required")
		return
	}

	req := ports.PostMovementRequest{
		LegalEntityID:   leID,
		MovementType:    domain.MovementType(dto.MovementType),
		ItemID:          dto.ItemID,
		SKUID:           dto.SKUID,
		WarehouseID:     dto.WarehouseID,
		BinID:           dto.BinID,
		LotNumber:       dto.LotNumber,
		SerialNumber:    dto.SerialNumber,
		Quantity:        dto.Quantity,
		UOMID:           dto.UOMID,
		UnitCost:        dto.UnitCost,
		SourceDocType:   dto.SourceDocType,
		SourceDocID:     dto.SourceDocID,
		SourceDocLineID: dto.SourceDocLineID,
		CreatedBy:       userID,
	}

	created, err := h.inventoryService.PostStockMovement(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) {
			writeError(w, http.StatusConflict, "INSUFFICIENT_STOCK", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "POST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 3. Balances
func (h *InventoryHandler) ListBalances(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var itemID *uuid.UUID
	if iStr := r.URL.Query().Get("item_id"); iStr != "" {
		if id, err := uuid.Parse(iStr); err == nil {
			itemID = &id
		}
	}

	var warehouseID *uuid.UUID
	if wStr := r.URL.Query().Get("warehouse_id"); wStr != "" {
		if id, err := uuid.Parse(wStr); err == nil {
			warehouseID = &id
		}
	}

	limit := int32(50)
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	offset := int32(0)
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	balances, err := h.inventoryService.ListBalances(r.Context(), leID, itemID, warehouseID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": balances})
}

// 4. Reservations
func (h *InventoryHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var dto ReserveStockRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.ItemID == uuid.Nil || dto.WarehouseID == uuid.Nil || dto.Quantity.IsZero() || dto.Quantity.IsNegative() {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "ItemID, WarehouseID, and positive Quantity are required")
		return
	}

	req := ports.ReserveStockRequest{
		LegalEntityID:    leID,
		ItemID:           dto.ItemID,
		WarehouseID:      dto.WarehouseID,
		BinID:            dto.BinID,
		Quantity:         dto.Quantity,
		DemandSourceType: dto.DemandSourceType,
		DemandSourceID:   dto.DemandSourceID,
	}

	created, err := h.inventoryService.ReserveStock(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) {
			writeError(w, http.StatusConflict, "INSUFFICIENT_STOCK", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "RESERVE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *InventoryHandler) ReleaseStock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid reservation UUID")
		return
	}

	if err := h.inventoryService.ReleaseStock(r.Context(), id); err != nil {
		writeError(w, http.StatusBadRequest, "RELEASE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "RELEASED"})
}

// 5. Transfers
func (h *InventoryHandler) ListTransferOrders(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	limit := int32(50)
	offset := int32(0)

	orders, err := h.inventoryService.ListTransferOrders(r.Context(), leID, status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": orders})
}

func (h *InventoryHandler) CreateTransferOrder(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	var dto CreateTransferRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	lines := make([]ports.CreateTransferLineRequest, 0, len(dto.Lines))
	for _, l := range dto.Lines {
		lines = append(lines, ports.CreateTransferLineRequest{
			ItemID:            l.ItemID,
			SKUID:             l.SKUID,
			RequestedQuantity: l.RequestedQuantity,
			UOMID:             l.UOMID,
			Notes:             l.Notes,
		})
	}

	req := ports.CreateTransferRequest{
		LegalEntityID:   leID,
		FromWarehouseID: dto.FromWarehouseID,
		ToWarehouseID:   dto.ToWarehouseID,
		Notes:           dto.Notes,
		CreatedBy:       userID,
		Lines:           lines,
	}

	order, err := h.inventoryService.CreateTransferOrder(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": order})
}

func (h *InventoryHandler) GetTransferOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transfer order UUID")
		return
	}

	order, err := h.inventoryService.GetTransferOrder(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transfer order not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": order})
}

func (h *InventoryHandler) DispatchTransfer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transfer order UUID")
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	order, err := h.inventoryService.DispatchTransfer(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "DISPATCH_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": order})
}

func (h *InventoryHandler) ReceiveTransfer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transfer order UUID")
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	order, err := h.inventoryService.ReceiveTransfer(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "RECEIVE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": order})
}

// 6. Adjustments
func (h *InventoryHandler) ListAdjustments(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var warehouseID *uuid.UUID
	if wStr := r.URL.Query().Get("warehouse_id"); wStr != "" {
		if id, err := uuid.Parse(wStr); err == nil {
			warehouseID = &id
		}
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	adjustments, err := h.inventoryService.ListAdjustments(r.Context(), leID, warehouseID, status, 50, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": adjustments})
}

func (h *InventoryHandler) CreateAdjustment(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	var dto CreateAdjustmentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	lines := make([]ports.CreateAdjustmentLineRequest, 0, len(dto.Lines))
	for _, l := range dto.Lines {
		lines = append(lines, ports.CreateAdjustmentLineRequest{
			ItemID:       l.ItemID,
			BinID:        l.BinID,
			LotNumber:    l.LotNumber,
			SerialNumber: l.SerialNumber,
			Quantity:     l.Quantity,
			UnitCost:     l.UnitCost,
			Reason:       l.Reason,
		})
	}

	req := ports.CreateAdjustmentRequest{
		LegalEntityID:  leID,
		WarehouseID:    dto.WarehouseID,
		AdjustmentType: domain.AdjustmentType(dto.AdjustmentType),
		ReasonCode:     dto.ReasonCode,
		Notes:          dto.Notes,
		CreatedBy:      userID,
		Lines:          lines,
	}

	adj, err := h.inventoryService.CreateAdjustment(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": adj})
}

func (h *InventoryHandler) PostAdjustment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid adjustment UUID")
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	adj, err := h.inventoryService.PostAdjustment(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "POST_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": adj})
}

// 7. Counts
func (h *InventoryHandler) ListStockCounts(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var warehouseID *uuid.UUID
	if wStr := r.URL.Query().Get("warehouse_id"); wStr != "" {
		if id, err := uuid.Parse(wStr); err == nil {
			warehouseID = &id
		}
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	counts, err := h.inventoryService.ListStockCounts(r.Context(), leID, warehouseID, status, 50, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": counts})
}

func (h *InventoryHandler) CreateStockCount(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var dto CreateCountRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	lines := make([]ports.CreateCountLineRequest, 0, len(dto.Lines))
	for _, l := range dto.Lines {
		lines = append(lines, ports.CreateCountLineRequest{
			ItemID:          l.ItemID,
			BinID:           l.BinID,
			CountedQuantity: l.CountedQuantity,
			Notes:           l.Notes,
		})
	}

	req := ports.CreateCountRequest{
		LegalEntityID: leID,
		WarehouseID:   dto.WarehouseID,
		CountType:     domain.CountType(dto.CountType),
		Lines:         lines,
	}

	count, err := h.inventoryService.CreateStockCount(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": count})
}

func (h *InventoryHandler) ReconcileStockCount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid count UUID")
		return
	}

	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	count, err := h.inventoryService.ReconcileStockCount(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "RECONCILE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": count})
}

// Costing & Valuation Endpoints

func (h *InventoryHandler) GetValuationReport(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ENTITY", err.Error())
		return
	}

	var whIDPtr, itemIDPtr *uuid.UUID
	if whIDStr := r.URL.Query().Get("warehouse_id"); whIDStr != "" {
		if id, err := uuid.Parse(whIDStr); err == nil {
			whIDPtr = &id
		}
	}
	if itemIDStr := r.URL.Query().Get("item_id"); itemIDStr != "" {
		if id, err := uuid.Parse(itemIDStr); err == nil {
			itemIDPtr = &id
		}
	}

	report, err := h.inventoryService.GetValuationReport(r.Context(), leID, whIDPtr, itemIDPtr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "VALUATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": report})
}

func (h *InventoryHandler) ListCostLayers(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ENTITY", err.Error())
		return
	}

	var whIDPtr, itemIDPtr *uuid.UUID
	var onlyActivePtr *bool
	if whIDStr := r.URL.Query().Get("warehouse_id"); whIDStr != "" {
		if id, err := uuid.Parse(whIDStr); err == nil {
			whIDPtr = &id
		}
	}
	if itemIDStr := r.URL.Query().Get("item_id"); itemIDStr != "" {
		if id, err := uuid.Parse(itemIDStr); err == nil {
			itemIDPtr = &id
		}
	}
	if activeStr := r.URL.Query().Get("only_active"); activeStr != "" {
		b := activeStr == "true" || activeStr == "1"
		onlyActivePtr = &b
	}

	limit := int32(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = int32(val)
		}
	}
	offset := int32(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = int32(val)
		}
	}

	layers, err := h.inventoryService.ListCostLayers(r.Context(), leID, itemIDPtr, whIDPtr, onlyActivePtr, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "FETCH_LAYERS_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": layers})
}

func (h *InventoryHandler) ExecuteRevaluation(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ENTITY", err.Error())
		return
	}
	user := iamhttp.UserFromContext(r.Context())
	userID := uuid.Nil
	if user != nil {
		userID = user.ID
	}

	var req ExecuteRevaluationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	rev, err := h.inventoryService.ExecuteRevaluation(r.Context(), ports.ExecuteRevaluationRequest{
		LegalEntityID:  leID,
		WarehouseID:    req.WarehouseID,
		ItemID:         req.ItemID,
		NewUnitCost:    req.NewUnitCost,
		Reason:         req.Reason,
		JournalEntryID: req.JournalEntryID,
		CreatedBy:      userID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrZeroQuantityRevaluation) {
			writeError(w, http.StatusConflict, "ZERO_QUANTITY", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "REVALUATION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": rev})
}

func (h *InventoryHandler) ListRevaluations(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ENTITY", err.Error())
		return
	}

	var whIDPtr, itemIDPtr *uuid.UUID
	if whIDStr := r.URL.Query().Get("warehouse_id"); whIDStr != "" {
		if id, err := uuid.Parse(whIDStr); err == nil {
			whIDPtr = &id
		}
	}
	if itemIDStr := r.URL.Query().Get("item_id"); itemIDStr != "" {
		if id, err := uuid.Parse(itemIDStr); err == nil {
			itemIDPtr = &id
		}
	}

	limit := int32(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = int32(val)
		}
	}
	offset := int32(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = int32(val)
		}
	}

	revs, err := h.inventoryService.ListRevaluations(r.Context(), leID, whIDPtr, itemIDPtr, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "FETCH_REVALUATIONS_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": revs})
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
