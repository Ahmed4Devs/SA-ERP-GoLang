package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"erp/internal/purchasing/domain"
	"erp/internal/purchasing/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PurchasingHandler struct {
	purchasingService ports.PurchasingService
}

func NewPurchasingHandler(purchasingService ports.PurchasingService) *PurchasingHandler {
	return &PurchasingHandler{
		purchasingService: purchasingService,
	}
}

func (h *PurchasingHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Purchase Requests
		r.Post("/requests", h.CreatePurchaseRequest)
		r.Get("/requests", h.ListPurchaseRequests)
		r.Get("/requests/{id}", h.GetPurchaseRequest)
		r.Post("/requests/{id}/approve", h.ApprovePurchaseRequest)

		// RFQs
		r.Post("/rfqs", h.CreateRFQ)
		r.Get("/rfqs", h.ListRFQs)
		r.Get("/rfqs/{id}", h.GetRFQ)

		// Supplier Quotes
		r.Post("/quotes", h.RecordSupplierQuote)
		r.Get("/quotes", h.ListSupplierQuotes)
		r.Get("/quotes/{id}", h.GetSupplierQuote)
		r.Post("/quotes/{id}/award", h.AwardSupplierQuote)

		// Purchase Orders
		r.Post("/orders", h.CreatePurchaseOrder)
		r.Get("/orders", h.ListPurchaseOrders)
		r.Get("/orders/{id}", h.GetPurchaseOrder)
		r.Post("/orders/{id}/approve", h.ApprovePurchaseOrder)
		r.Post("/orders/{id}/issue", h.IssuePurchaseOrder)

		// Goods Receipts
		r.Post("/receipts", h.CreateGoodsReceipt)
		r.Get("/receipts", h.ListGoodsReceipts)
		r.Get("/receipts/{id}", h.GetGoodsReceipt)
		r.Post("/receipts/{id}/accept", h.InspectAndAcceptGoodsReceipt)

		// Three-Way Matching
		r.Post("/matching/verify", h.VerifyThreeWayMatch)
	})

	return r
}

// ----------------------------------------------------------------------------
// Purchase Request Handlers
// ----------------------------------------------------------------------------

func (h *PurchasingHandler) CreatePurchaseRequest(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	reqDate, err := time.Parse("2006-01-02", req.RequestDate)
	if err != nil {
		reqDate = time.Now()
	}
	var reqByDate *time.Time
	if req.RequiredByDate != nil && *req.RequiredByDate != "" {
		t, err := time.Parse("2006-01-02", *req.RequiredByDate)
		if err == nil {
			reqByDate = &t
		}
	}

	cmd := ports.CreatePRCmd{
		LegalEntityID:     req.LegalEntityID,
		RequestedByUserID: uuid.Nil, // In real flow, from context user
		DepartmentID:      req.DepartmentID,
		RequestDate:       reqDate,
		RequiredByDate:    reqByDate,
		Purpose:           req.Purpose,
		Notes:             req.Notes,
		CurrencyCode:      req.CurrencyCode,
	}

	for _, l := range req.Lines {
		cmd.Lines = append(cmd.Lines, ports.CreatePRLineCmd{
			ItemID:              l.ItemID,
			Description:         l.Description,
			UOM:                 l.UOM,
			Quantity:            l.Quantity,
			EstimatedUnitCost:   l.EstimatedUnitCost,
			SuggestedSupplierID: l.SuggestedSupplierID,
		})
	}

	pr, err := h.purchasingService.CreatePurchaseRequest(r.Context(), cmd)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, pr)
}

func (h *PurchasingHandler) GetPurchaseRequest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid purchase request id"})
		return
	}

	pr, err := h.purchasingService.GetPurchaseRequest(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, pr)
}

func (h *PurchasingHandler) ListPurchaseRequests(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id required"})
		return
	}

	var status *domain.PRStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.PRStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	prs, err := h.purchasingService.ListPurchaseRequests(r.Context(), leID, status, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, prs)
}

func (h *PurchasingHandler) ApprovePurchaseRequest(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid purchase request id"})
		return
	}

	pr, err := h.purchasingService.ApprovePurchaseRequest(r.Context(), id, uuid.Nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, pr)
}

// ----------------------------------------------------------------------------
// RFQ Handlers
// ----------------------------------------------------------------------------

func (h *PurchasingHandler) CreateRFQ(w http.ResponseWriter, r *http.Request) {
	var req CreateRFQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rfqDate, err := time.Parse("2006-01-02", req.RFQDate)
	if err != nil {
		rfqDate = time.Now()
	}
	closingDate, err := time.Parse("2006-01-02", req.ClosingDate)
	if err != nil {
		closingDate = rfqDate.AddDate(0, 0, 14)
	}

	cmd := ports.CreateRFQCmd{
		LegalEntityID:     req.LegalEntityID,
		PurchaseRequestID: req.PurchaseRequestID,
		Title:             req.Title,
		RFQDate:           rfqDate,
		ClosingDate:       closingDate,
		Notes:             req.Notes,
	}

	for _, l := range req.Lines {
		cmd.Lines = append(cmd.Lines, ports.CreateRFQLineCmd{
			PRLineID:        l.PRLineID,
			ItemID:          l.ItemID,
			Description:     l.Description,
			UOM:             l.UOM,
			Quantity:        l.Quantity,
			TargetUnitPrice: l.TargetUnitPrice,
		})
	}

	rfq, err := h.purchasingService.CreateRFQ(r.Context(), cmd)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, rfq)
}

func (h *PurchasingHandler) GetRFQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid rfq id"})
		return
	}
	rfq, err := h.purchasingService.GetRFQ(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rfq)
}

func (h *PurchasingHandler) ListRFQs(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id required"})
		return
	}

	var status *domain.RFQStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.RFQStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rfqs, err := h.purchasingService.ListRFQs(r.Context(), leID, status, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rfqs)
}

// ----------------------------------------------------------------------------
// Supplier Quote Handlers
// ----------------------------------------------------------------------------

func (h *PurchasingHandler) RecordSupplierQuote(w http.ResponseWriter, r *http.Request) {
	var req RecordQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	qDate, err := time.Parse("2006-01-02", req.QuoteDate)
	if err != nil {
		qDate = time.Now()
	}
	valUntil, err := time.Parse("2006-01-02", req.ValidUntil)
	if err != nil {
		valUntil = qDate.AddDate(0, 1, 0)
	}

	cmd := ports.RecordQuoteCmd{
		LegalEntityID:          req.LegalEntityID,
		RFQID:                  req.RFQID,
		SupplierID:             req.SupplierID,
		SupplierQuoteReference: req.SupplierQuoteReference,
		QuoteDate:              qDate,
		ValidUntil:             valUntil,
		CurrencyCode:           req.CurrencyCode,
		ExchangeRate:           req.ExchangeRate,
		PaymentTermsID:         req.PaymentTermsID,
		Notes:                  req.Notes,
	}

	for _, l := range req.Lines {
		cmd.Lines = append(cmd.Lines, ports.RecordQuoteLineCmd{
			RFQLineID:       l.RFQLineID,
			ItemID:          l.ItemID,
			Description:     l.Description,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			DiscountPercent: l.DiscountPercent,
			TaxRatePercent:  l.TaxRatePercent,
		})
	}

	quote, err := h.purchasingService.RecordSupplierQuote(r.Context(), cmd)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, quote)
}

func (h *PurchasingHandler) GetSupplierQuote(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid quote id"})
		return
	}
	quote, err := h.purchasingService.GetSupplierQuote(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

func (h *PurchasingHandler) ListSupplierQuotes(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id required"})
		return
	}

	var rfqID, suppID *uuid.UUID
	if idStr := r.URL.Query().Get("rfq_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			rfqID = &id
		}
	}
	if idStr := r.URL.Query().Get("supplier_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			suppID = &id
		}
	}

	var status *domain.QuoteStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.QuoteStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	quotes, err := h.purchasingService.ListSupplierQuotes(r.Context(), leID, rfqID, suppID, status, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, quotes)
}

func (h *PurchasingHandler) AwardSupplierQuote(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid quote id"})
		return
	}
	quote, err := h.purchasingService.AwardSupplierQuote(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

// ----------------------------------------------------------------------------
// Purchase Order Handlers
// ----------------------------------------------------------------------------

func (h *PurchasingHandler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	var req CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		orderDate = time.Now()
	}
	var expDelivery *time.Time
	if req.ExpectedDeliveryDate != nil && *req.ExpectedDeliveryDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpectedDeliveryDate)
		if err == nil {
			expDelivery = &t
		}
	}

	cmd := ports.CreatePOCmd{
		LegalEntityID:        req.LegalEntityID,
		PurchaseRequestID:    req.PurchaseRequestID,
		RFQID:                req.RFQID,
		SupplierQuoteID:      req.SupplierQuoteID,
		SupplierID:           req.SupplierID,
		OrderDate:            orderDate,
		ExpectedDeliveryDate: expDelivery,
		PaymentTermsID:       req.PaymentTermsID,
		DeliveryWarehouseID:  req.DeliveryWarehouseID,
		CurrencyCode:         req.CurrencyCode,
		ExchangeRate:         req.ExchangeRate,
		Notes:                req.Notes,
	}

	for _, l := range req.Lines {
		cmd.Lines = append(cmd.Lines, ports.CreatePOLineCmd{
			PRLineID:         l.PRLineID,
			ItemID:           l.ItemID,
			Description:      l.Description,
			UOM:              l.UOM,
			Quantity:         l.Quantity,
			UnitPrice:        l.UnitPrice,
			DiscountPercent:  l.DiscountPercent,
			TaxCode:          l.TaxCode,
			TaxRatePercent:   l.TaxRatePercent,
			ExpenseAccountID: l.ExpenseAccountID,
		})
	}

	po, err := h.purchasingService.CreatePurchaseOrder(r.Context(), cmd)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, po)
}

func (h *PurchasingHandler) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}
	po, err := h.purchasingService.GetPurchaseOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, po)
}

func (h *PurchasingHandler) ListPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id required"})
		return
	}

	var suppID *uuid.UUID
	if idStr := r.URL.Query().Get("supplier_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			suppID = &id
		}
	}

	var status *domain.POStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.POStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.purchasingService.ListPurchaseOrders(r.Context(), leID, suppID, status, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *PurchasingHandler) ApprovePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}
	po, err := h.purchasingService.ApprovePurchaseOrder(r.Context(), id, uuid.Nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, po)
}

func (h *PurchasingHandler) IssuePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}
	po, err := h.purchasingService.IssuePurchaseOrder(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, po)
}

// ----------------------------------------------------------------------------
// Goods Receipt Handlers
// ----------------------------------------------------------------------------

func (h *PurchasingHandler) CreateGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	var req CreateReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	rcptDate, err := time.Parse("2006-01-02", req.ReceiptDate)
	if err != nil {
		rcptDate = time.Now()
	}

	cmd := ports.CreateReceiptCmd{
		LegalEntityID:      req.LegalEntityID,
		PurchaseOrderID:    req.PurchaseOrderID,
		SupplierID:         req.SupplierID,
		WarehouseID:        req.WarehouseID,
		ReceiptDate:        rcptDate,
		DeliveryNoteNumber: req.DeliveryNoteNumber,
		Notes:              req.Notes,
	}

	for _, l := range req.Lines {
		cmd.Lines = append(cmd.Lines, ports.CreateReceiptLineCmd{
			POLineID:         l.POLineID,
			ItemID:           l.ItemID,
			WarehouseID:      l.WarehouseID,
			BinID:            l.BinID,
			ReceivedQuantity: l.ReceivedQuantity,
			AcceptedQuantity: l.AcceptedQuantity,
			RejectedQuantity: l.RejectedQuantity,
			UnitCost:         l.UnitCost,
			RejectionReason:  l.RejectionReason,
		})
	}

	gr, err := h.purchasingService.CreateGoodsReceipt(r.Context(), cmd)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, gr)
}

func (h *PurchasingHandler) GetGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid receipt id"})
		return
	}
	gr, err := h.purchasingService.GetGoodsReceipt(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, gr)
}

func (h *PurchasingHandler) ListGoodsReceipts(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id required"})
		return
	}

	var poID, suppID *uuid.UUID
	if idStr := r.URL.Query().Get("purchase_order_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			poID = &id
		}
	}
	if idStr := r.URL.Query().Get("supplier_id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			suppID = &id
		}
	}

	var status *domain.GRNStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.GRNStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	receipts, err := h.purchasingService.ListGoodsReceipts(r.Context(), leID, poID, suppID, status, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, receipts)
}

func (h *PurchasingHandler) InspectAndAcceptGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid receipt id"})
		return
	}
	gr, err := h.purchasingService.InspectAndAcceptGoodsReceipt(r.Context(), id, uuid.Nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, gr)
}

// ----------------------------------------------------------------------------
// Three-Way Match Handler
// ----------------------------------------------------------------------------

func (h *PurchasingHandler) VerifyThreeWayMatch(w http.ResponseWriter, r *http.Request) {
	var req ThreeWayMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	result, err := h.purchasingService.VerifyThreeWayMatch(r.Context(), ports.ThreeWayMatchReq{
		POLineID:              req.POLineID,
		GRNLineID:             req.GRNLineID,
		InvoiceLineID:         req.InvoiceLineID,
		InvoicedQuantity:      req.InvoicedQuantity,
		InvoicedUnitPrice:     req.InvoicedUnitPrice,
		PriceTolerancePercent: req.PriceTolerancePercent,
		QtyTolerancePercent:   req.QtyTolerancePercent,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ----------------------------------------------------------------------------
// Helper
// ----------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
