package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"erp/internal/products/domain"
	"erp/internal/products/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productService ports.ProductService
}

func NewProductHandler(productService ports.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Units of Measure
		r.Get("/uoms", h.ListUOMs)
		r.Post("/uoms", h.CreateUOM)
		r.Post("/uoms/convert", h.ConvertUOM)

		// Categories
		r.Get("/categories", h.ListCategories)
		r.Post("/categories", h.CreateCategory)

		// Items Master
		r.Get("/items", h.ListItems)
		r.Post("/items", h.CreateItem)
		r.Get("/items/search", h.SearchItems)
		r.Get("/items/{id}", h.GetItem)
		r.Put("/items/{id}", h.UpdateItem)

		// Variants & SKUs
		r.Get("/items/{id}/variants", h.ListVariants)
		r.Post("/items/{id}/variants", h.CreateVariant)
		r.Get("/items/{id}/skus", h.ListSKUs)

		// Barcodes
		r.Get("/barcodes/lookup", h.LookupBarcode)
		r.Get("/items/{id}/barcodes", h.ListBarcodes)
		r.Post("/items/{id}/barcodes", h.AddBarcode)

		// Packaging UOMs
		r.Get("/items/{id}/uoms", h.ListItemUOMs)
		r.Post("/items/{id}/uoms", h.AddItemUOM)

		// Tax Profiles
		r.Get("/items/{id}/tax-profile", h.GetItemTaxProfile)
		r.Post("/items/{id}/tax-profile", h.SetItemTaxProfile)
	})

	return r
}

func (h *ProductHandler) resolveLegalEntityID(r *http.Request) (uuid.UUID, error) {
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

// 1. Units of Measure
func (h *ProductHandler) ListUOMs(w http.ResponseWriter, r *http.Request) {
	uoms, err := h.productService.ListUOMs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": uoms})
}

func (h *ProductHandler) CreateUOM(w http.ResponseWriter, r *http.Request) {
	var req CreateUOMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.Code == "" || req.NameEn == "" || req.NameAr == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Code, NameEn, and NameAr are required")
		return
	}

	uom := domain.UnitOfMeasure{
		Code:     req.Code,
		NameEn:   req.NameEn,
		NameAr:   req.NameAr,
		SymbolEn: req.SymbolEn,
		SymbolAr: req.SymbolAr,
		UOMType:  req.UOMType,
	}

	created, err := h.productService.CreateUOM(r.Context(), uom)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *ProductHandler) ConvertUOM(w http.ResponseWriter, r *http.Request) {
	var req ConvertUOMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	converted, err := h.productService.ConvertQuantity(r.Context(), req.FromUOMID, req.ToUOMID, req.Quantity)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CONVERSION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"from_uom_id":        req.FromUOMID,
			"to_uom_id":          req.ToUOMID,
			"original_quantity":  req.Quantity,
			"converted_quantity": converted,
		},
	})
}

// 2. Categories
func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	cats, err := h.productService.ListCategories(r.Context(), leID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": cats})
}

func (h *ProductHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.Code == "" || req.NameEn == "" || req.NameAr == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Code, NameEn, and NameAr are required")
		return
	}

	cat := domain.Category{
		LegalEntityID: leID,
		ParentID:      req.ParentID,
		Code:          req.Code,
		NameEn:        req.NameEn,
		NameAr:        req.NameAr,
		Description:   req.Description,
	}

	created, err := h.productService.CreateCategory(r.Context(), cat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 3. Items Master
func (h *ProductHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var itemType *string
	if t := r.URL.Query().Get("item_type"); t != "" {
		itemType = &t
	}

	var categoryID *uuid.UUID
	if cidStr := r.URL.Query().Get("category_id"); cidStr != "" {
		if cid, err := uuid.Parse(cidStr); err == nil {
			categoryID = &cid
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

	items, err := h.productService.ListItems(r.Context(), leID, itemType, categoryID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *ProductHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	item, err := h.productService.GetItem(r.Context(), itemID)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h *ProductHandler) SearchItems(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	query := r.URL.Query().Get("q")
	items, err := h.productService.SearchItems(r.Context(), leID, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *ProductHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	var dto CreateItemRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.Code == "" || dto.NameEn == "" || dto.NameAr == "" || dto.BaseUOMID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Code, NameEn, NameAr, and BaseUOMID are required")
		return
	}

	req := ports.CreateItemRequest{
		LegalEntityID:     leID,
		CategoryID:        dto.CategoryID,
		Code:              dto.Code,
		NameEn:            dto.NameEn,
		NameAr:            dto.NameAr,
		DescriptionEn:     dto.DescriptionEn,
		DescriptionAr:     dto.DescriptionAr,
		ItemType:          domain.ItemType(dto.ItemType),
		TrackingType:      domain.TrackingType(dto.TrackingType),
		BaseUOMID:         dto.BaseUOMID,
		PurchaseUOMID:     dto.PurchaseUOMID,
		SalesUOMID:        dto.SalesUOMID,
		DefaultTaxCodeID:  dto.DefaultTaxCodeID,
		CostPrice:         dto.CostPrice,
		ListPrice:         dto.ListPrice,
		IsSalesEnabled:    dto.IsSalesEnabled,
		IsPurchaseEnabled: dto.IsPurchaseEnabled,
		DefaultBarcode:    dto.DefaultBarcode,
	}

	created, err := h.productService.CreateItem(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateItemCode) {
			writeError(w, http.StatusConflict, "DUPLICATE_CODE", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (h *ProductHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	existing, err := h.productService.GetItem(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Item not found")
		return
	}

	var dto UpdateItemRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	existing.CategoryID = dto.CategoryID
	existing.NameEn = dto.NameEn
	existing.NameAr = dto.NameAr
	existing.DescriptionEn = dto.DescriptionEn
	existing.DescriptionAr = dto.DescriptionAr
	existing.ItemType = domain.ItemType(dto.ItemType)
	existing.TrackingType = domain.TrackingType(dto.TrackingType)
	existing.BaseUOMID = dto.BaseUOMID
	existing.PurchaseUOMID = dto.PurchaseUOMID
	existing.SalesUOMID = dto.SalesUOMID
	existing.DefaultTaxCodeID = dto.DefaultTaxCodeID
	existing.CostPrice = dto.CostPrice
	existing.ListPrice = dto.ListPrice
	existing.IsSalesEnabled = dto.IsSalesEnabled
	existing.IsPurchaseEnabled = dto.IsPurchaseEnabled
	existing.IsActive = dto.IsActive

	updated, err := h.productService.UpdateItem(r.Context(), *existing)
	if err != nil {
		writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": updated})
}

// 4. Variants & SKUs
func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	variants, err := h.productService.ListVariants(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": variants})
}

func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	var dto CreateVariantRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.VariantCode == "" || dto.NameEn == "" || dto.NameAr == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "VariantCode, NameEn, and NameAr are required")
		return
	}

	v, sku, err := h.productService.CreateVariantWithSKU(r.Context(), leID, itemID, dto.VariantCode, dto.NameEn, dto.NameAr, dto.Attributes, dto.CostPrice, dto.SellingPrice)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"variant": v,
			"sku":     sku,
		},
	})
}

func (h *ProductHandler) ListSKUs(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	skus, err := h.productService.ListSKUs(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": skus})
}

// 5. Barcodes
func (h *ProductHandler) LookupBarcode(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	code := r.URL.Query().Get("barcode")
	if code == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "barcode parameter is required")
		return
	}

	bc, err := h.productService.LookupBarcode(r.Context(), leID, code)
	if err != nil {
		if errors.Is(err, domain.ErrBarcodeNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Barcode not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": bc})
}

func (h *ProductHandler) ListBarcodes(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	barcodes, err := h.productService.ListBarcodes(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": barcodes})
}

func (h *ProductHandler) AddBarcode(w http.ResponseWriter, r *http.Request) {
	leID, err := h.resolveLegalEntityID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SCOPE", err.Error())
		return
	}

	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	var dto AddBarcodeRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.Barcode == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Barcode is required")
		return
	}

	bc := domain.Barcode{
		LegalEntityID: leID,
		ItemID:        itemID,
		SKUID:         dto.SKUID,
		UOMID:         dto.UOMID,
		Barcode:       dto.Barcode,
		BarcodeFormat: dto.BarcodeFormat,
		IsPrimary:     dto.IsPrimary,
	}

	created, err := h.productService.AddBarcode(r.Context(), bc)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateBarcode) {
			writeError(w, http.StatusConflict, "DUPLICATE_BARCODE", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 6. Packaging UOMs
func (h *ProductHandler) ListItemUOMs(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	uoms, err := h.productService.ListItemUOMs(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": uoms})
}

func (h *ProductHandler) AddItemUOM(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	var dto AddItemUOMRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.UOMID == uuid.Nil || dto.ConversionRate.IsZero() || dto.ConversionRate.IsNegative() {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "UOMID and positive conversion_rate are required")
		return
	}

	iu := domain.ItemUOM{
		ItemID:         itemID,
		UOMID:          dto.UOMID,
		ConversionRate: dto.ConversionRate,
		IsBaseUOM:      false,
		IsSalesUOM:     dto.IsSalesUOM,
		IsPurchaseUOM:  dto.IsPurchaseUOM,
	}

	created, err := h.productService.AddItemUOM(r.Context(), iu)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

// 7. Item Tax Profiles
func (h *ProductHandler) GetItemTaxProfile(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	itp, err := h.productService.GetItemTaxProfile(r.Context(), itemID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Tax profile not found for item")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": itp})
}

func (h *ProductHandler) SetItemTaxProfile(w http.ResponseWriter, r *http.Request) {
	itemIDStr := chi.URLParam(r, "id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid item UUID")
		return
	}

	var dto SetTaxProfileRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if dto.TaxCodeID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "TaxCodeID is required")
		return
	}

	itp := domain.ItemTaxProfile{
		ItemID:              itemID,
		TaxCodeID:           dto.TaxCodeID,
		ExemptionReasonCode: dto.ExemptionReasonCode,
		ZeroRateReasonCode:  dto.ZeroRateReasonCode,
		IsActive:            true,
	}

	created, err := h.productService.SetItemTaxProfile(r.Context(), itp)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": created})
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
