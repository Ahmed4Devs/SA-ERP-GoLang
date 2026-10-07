package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	iamhttp "erp/internal/iam/transport/http"
	"erp/internal/sales/domain"
	"erp/internal/sales/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SalesHandler struct {
	salesService ports.SalesService
}

func NewSalesHandler(salesService ports.SalesService) *SalesHandler {
	return &SalesHandler{
		salesService: salesService,
	}
}

func (h *SalesHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Quotations
		r.Post("/quotations", h.CreateQuotation)
		r.Get("/quotations", h.ListQuotations)
		r.Get("/quotations/{id}", h.GetQuotation)
		r.Post("/quotations/{id}/convert", h.ConvertQuotationToOrder)

		// Orders
		r.Post("/orders", h.CreateOrder)
		r.Get("/orders", h.ListOrders)
		r.Get("/orders/{id}", h.GetOrder)
		r.Post("/orders/{id}/confirm", h.ConfirmOrder)
		r.Post("/orders/{id}/invoice", h.CreateInvoiceFromOrder)

		// Deliveries
		r.Post("/deliveries", h.CreateDelivery)
		r.Get("/deliveries", h.ListDeliveries)
		r.Get("/deliveries/{id}", h.GetDelivery)
		r.Post("/deliveries/{id}/dispatch", h.DispatchDelivery)

		// Invoices
		r.Post("/invoices", h.CreateInvoice)
		r.Get("/invoices", h.ListInvoices)
		r.Get("/invoices/{id}", h.GetInvoice)
		r.Post("/invoices/{id}/post", h.PostInvoice)
		r.Post("/invoices/{id}/submit", h.SubmitInvoice)

		// Credit Notes
		r.Post("/credit-notes", h.CreateCreditNote)
		r.Get("/credit-notes", h.ListCreditNotes)
		r.Get("/credit-notes/{id}", h.GetCreditNote)
		r.Post("/credit-notes/{id}/post", h.PostCreditNote)
	})

	return r
}

// ----------------------------------------------------------------------------
// Quotation Handlers
// ----------------------------------------------------------------------------

func (h *SalesHandler) CreateQuotation(w http.ResponseWriter, r *http.Request) {
	var req CreateQuotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	qDate, _ := time.Parse("2006-01-02", req.QuoteDate)
	if qDate.IsZero() {
		qDate = time.Now()
	}
	vUntil, _ := time.Parse("2006-01-02", req.ValidUntil)
	if vUntil.IsZero() {
		vUntil = qDate.AddDate(0, 0, 30)
	}

	var lines []ports.CreateQuotationLineCommand
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateQuotationLineCommand{
			ItemID:          l.ItemID,
			DescriptionEn:   l.DescriptionEn,
			DescriptionAr:   l.DescriptionAr,
			UOM:             l.UOM,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			DiscountPercent: l.DiscountPercent,
			TaxCode:         l.TaxCode,
			TaxRatePercent:  l.TaxRatePercent,
		})
	}

	quotation, err := h.salesService.CreateQuotation(r.Context(), ports.CreateQuotationCommand{
		LegalEntityID:   req.LegalEntityID,
		CustomerID:      req.CustomerID,
		CustomerNameEn:  req.CustomerNameEn,
		CustomerNameAr:  req.CustomerNameAr,
		QuoteDate:       qDate,
		ValidUntil:      vUntil,
		CurrencyCode:    req.CurrencyCode,
		ExchangeRate:    req.ExchangeRate,
		Notes:           req.Notes,
		Lines:           lines,
		CreatedByUserID: salesActorID(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(quotation)
}

func (h *SalesHandler) GetQuotation(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid quotation id", http.StatusBadRequest)
		return
	}

	q, err := h.salesService.GetQuotation(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(q)
}

func (h *SalesHandler) ListQuotations(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, "invalid or missing legal_entity_id", http.StatusBadRequest)
		return
	}

	var customerID *uuid.UUID
	if cStr := r.URL.Query().Get("customer_id"); cStr != "" {
		if c, err := uuid.Parse(cStr); err == nil {
			customerID = &c
		}
	}

	var status *domain.QuotationStatus
	if sStr := r.URL.Query().Get("status"); sStr != "" {
		s := domain.QuotationStatus(sStr)
		status = &s
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	quotes, err := h.salesService.ListQuotations(r.Context(), leID, customerID, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quotes)
}

func (h *SalesHandler) ConvertQuotationToOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid quotation id", http.StatusBadRequest)
		return
	}

	order, err := h.salesService.ConvertQuotationToOrder(r.Context(), id, salesActorID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

// ----------------------------------------------------------------------------
// Order Handlers
// ----------------------------------------------------------------------------

func (h *SalesHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oDate, _ := time.Parse("2006-01-02", req.OrderDate)
	if oDate.IsZero() {
		oDate = time.Now()
	}

	var expDelivery *time.Time
	if req.ExpectedDeliveryDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedDeliveryDate); err == nil {
			expDelivery = &t
		}
	}

	var lines []ports.CreateOrderLineCommand
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateOrderLineCommand{
			ItemID:          l.ItemID,
			DescriptionEn:   l.DescriptionEn,
			DescriptionAr:   l.DescriptionAr,
			UOM:             l.UOM,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			DiscountPercent: l.DiscountPercent,
			TaxCode:         l.TaxCode,
			TaxRatePercent:  l.TaxRatePercent,
		})
	}

	order, err := h.salesService.CreateOrder(r.Context(), ports.CreateOrderCommand{
		LegalEntityID:        req.LegalEntityID,
		CustomerID:           req.CustomerID,
		QuotationID:          req.QuotationID,
		CustomerPONumber:     req.CustomerPONumber,
		OrderDate:            oDate,
		ExpectedDeliveryDate: expDelivery,
		PaymentTermsID:       req.PaymentTermsID,
		CurrencyCode:         req.CurrencyCode,
		ExchangeRate:         req.ExchangeRate,
		Notes:                req.Notes,
		Lines:                lines,
		CreatedByUserID:      salesActorID(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)
}

func (h *SalesHandler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.salesService.ConfirmOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (h *SalesHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.salesService.GetOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (h *SalesHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, "invalid or missing legal_entity_id", http.StatusBadRequest)
		return
	}

	var customerID *uuid.UUID
	if cStr := r.URL.Query().Get("customer_id"); cStr != "" {
		if c, err := uuid.Parse(cStr); err == nil {
			customerID = &c
		}
	}

	var status *domain.OrderStatus
	if sStr := r.URL.Query().Get("status"); sStr != "" {
		s := domain.OrderStatus(sStr)
		status = &s
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.salesService.ListOrders(r.Context(), leID, customerID, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(orders)
}

func (h *SalesHandler) CreateInvoiceFromOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}

	invType := domain.InvoiceTypeStandard
	if r.URL.Query().Get("type") == "SIMPLIFIED" {
		invType = domain.InvoiceTypeSimplified
	}

	invoice, err := h.salesService.CreateInvoiceFromOrder(r.Context(), id, invType, salesActorID(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(invoice)
}

// ----------------------------------------------------------------------------
// Delivery Handlers
// ----------------------------------------------------------------------------

func (h *SalesHandler) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var req CreateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dDate, _ := time.Parse("2006-01-02", req.DeliveryDate)
	if dDate.IsZero() {
		dDate = time.Now()
	}

	var lines []ports.CreateDeliveryLineCommand
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateDeliveryLineCommand{
			OrderLineID: l.OrderLineID,
			ItemID:      l.ItemID,
			UOM:         l.UOM,
			Quantity:    l.Quantity,
			WarehouseID: l.WarehouseID,
			BinID:       l.BinID,
		})
	}

	delivery, err := h.salesService.CreateDelivery(r.Context(), ports.CreateDeliveryCommand{
		LegalEntityID:     req.LegalEntityID,
		OrderID:           req.OrderID,
		CustomerID:        req.CustomerID,
		WarehouseID:       req.WarehouseID,
		DeliveryDate:      dDate,
		ShippingReference: req.ShippingReference,
		Notes:             req.Notes,
		Lines:             lines,
		CreatedByUserID:   salesActorID(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(delivery)
}

func (h *SalesHandler) DispatchDelivery(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid delivery id", http.StatusBadRequest)
		return
	}

	delivery, err := h.salesService.DispatchDelivery(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(delivery)
}

func (h *SalesHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid delivery id", http.StatusBadRequest)
		return
	}

	delivery, err := h.salesService.GetDelivery(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(delivery)
}

func (h *SalesHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, "invalid or missing legal_entity_id", http.StatusBadRequest)
		return
	}

	var customerID, orderID *uuid.UUID
	if cStr := r.URL.Query().Get("customer_id"); cStr != "" {
		if c, err := uuid.Parse(cStr); err == nil {
			customerID = &c
		}
	}
	if oStr := r.URL.Query().Get("order_id"); oStr != "" {
		if o, err := uuid.Parse(oStr); err == nil {
			orderID = &o
		}
	}

	var status *domain.DeliveryStatus
	if sStr := r.URL.Query().Get("status"); sStr != "" {
		s := domain.DeliveryStatus(sStr)
		status = &s
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	deliveries, err := h.salesService.ListDeliveries(r.Context(), leID, customerID, orderID, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(deliveries)
}

// ----------------------------------------------------------------------------
// Invoice Handlers
// ----------------------------------------------------------------------------

func (h *SalesHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iDate, _ := time.Parse("2006-01-02", req.IssueDate)
	if iDate.IsZero() {
		iDate = time.Now()
	}
	dDate, _ := time.Parse("2006-01-02", req.DueDate)
	if dDate.IsZero() {
		dDate = iDate.AddDate(0, 0, 30)
	}

	var lines []ports.CreateInvoiceLineCommand
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateInvoiceLineCommand{
			ItemID:          l.ItemID,
			Description:     l.Description,
			UOM:             l.UOM,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			DiscountPercent: l.DiscountPercent,
			TaxCode:         l.TaxCode,
			TaxRatePercent:  l.TaxRatePercent,
			ExemptionReason: l.ExemptionReason,
		})
	}

	invoice, err := h.salesService.CreateInvoice(r.Context(), ports.CreateInvoiceCommand{
		LegalEntityID:       req.LegalEntityID,
		InvoiceType:         req.InvoiceType,
		OrderID:             req.OrderID,
		DeliveryID:          req.DeliveryID,
		CustomerID:          req.CustomerID,
		CustomerNameAr:      req.CustomerNameAr,
		CustomerNameEn:      req.CustomerNameEn,
		CustomerTaxID:       req.CustomerTaxID,
		CustomerCRNumber:    req.CustomerCRNumber,
		CustomerAddressJSON: req.CustomerAddressJSON,
		IssueDate:           iDate,
		DueDate:             dDate,
		PaymentTermsID:      req.PaymentTermsID,
		CurrencyCode:        req.CurrencyCode,
		ExchangeRate:        req.ExchangeRate,
		Notes:               req.Notes,
		Lines:               lines,
		CreatedByUserID:     salesActorID(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(invoice)
}

func (h *SalesHandler) PostInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}

	actorID := salesActorID(r)
	if actorID == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	invoice, err := h.salesService.PostInvoice(r.Context(), id, *actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

func (h *SalesHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid invoice id", http.StatusBadRequest)
		return
	}

	invoice, err := h.salesService.GetInvoice(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoice)
}

func (h *SalesHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, "invalid or missing legal_entity_id", http.StatusBadRequest)
		return
	}

	var customerID *uuid.UUID
	if cStr := r.URL.Query().Get("customer_id"); cStr != "" {
		if c, err := uuid.Parse(cStr); err == nil {
			customerID = &c
		}
	}

	var invType *domain.InvoiceType
	if tStr := r.URL.Query().Get("invoice_type"); tStr != "" {
		t := domain.InvoiceType(tStr)
		invType = &t
	}

	var status *domain.InvoiceStatus
	if sStr := r.URL.Query().Get("status"); sStr != "" {
		s := domain.InvoiceStatus(sStr)
		status = &s
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	invoices, err := h.salesService.ListInvoices(r.Context(), leID, customerID, invType, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoices)
}

// ----------------------------------------------------------------------------
// Credit Note Handlers
// ----------------------------------------------------------------------------

func (h *SalesHandler) CreateCreditNote(w http.ResponseWriter, r *http.Request) {
	var req CreateCreditNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	iDate, _ := time.Parse("2006-01-02", req.IssueDate)
	if iDate.IsZero() {
		iDate = time.Now()
	}

	var lines []ports.CreateCreditNoteLineCommand
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateCreditNoteLineCommand{
			ItemID:         l.ItemID,
			Description:    l.Description,
			Quantity:       l.Quantity,
			UnitPrice:      l.UnitPrice,
			TaxCode:        l.TaxCode,
			TaxRatePercent: l.TaxRatePercent,
		})
	}

	cn, err := h.salesService.CreateCreditNote(r.Context(), ports.CreateCreditNoteCommand{
		LegalEntityID:     req.LegalEntityID,
		OriginalInvoiceID: req.OriginalInvoiceID,
		CustomerID:        req.CustomerID,
		IssueDate:         iDate,
		ReasonCode:        req.ReasonCode,
		ReasonDetails:     req.ReasonDetails,
		CurrencyCode:      req.CurrencyCode,
		Lines:             lines,
		CreatedByUserID:   salesActorID(r),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cn)
}

func (h *SalesHandler) PostCreditNote(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid credit note id", http.StatusBadRequest)
		return
	}

	actorID := salesActorID(r)
	if actorID == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	cn, err := h.salesService.PostCreditNote(r.Context(), id, *actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cn)
}

func (h *SalesHandler) GetCreditNote(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid credit note id", http.StatusBadRequest)
		return
	}

	cn, err := h.salesService.GetCreditNote(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cn)
}

func (h *SalesHandler) ListCreditNotes(w http.ResponseWriter, r *http.Request) {
	leID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		http.Error(w, "invalid or missing legal_entity_id", http.StatusBadRequest)
		return
	}

	var customerID, origInvID *uuid.UUID
	if cStr := r.URL.Query().Get("customer_id"); cStr != "" {
		if c, err := uuid.Parse(cStr); err == nil {
			customerID = &c
		}
	}
	if iStr := r.URL.Query().Get("original_invoice_id"); iStr != "" {
		if i, err := uuid.Parse(iStr); err == nil {
			origInvID = &i
		}
	}

	var status *domain.CreditNoteStatus
	if sStr := r.URL.Query().Get("status"); sStr != "" {
		s := domain.CreditNoteStatus(sStr)
		status = &s
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	cns, err := h.salesService.ListCreditNotes(r.Context(), leID, customerID, origInvID, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cns)
}

func salesActorID(r *http.Request) *uuid.UUID {
	if user := iamhttp.UserFromContext(r.Context()); user != nil {
		return &user.ID
	}
	return nil
}
