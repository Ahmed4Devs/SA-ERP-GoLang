package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	customerService    ports.CustomerService
	receivablesService ports.ReceivablesService
}

func NewCustomerHandler(
	customerService ports.CustomerService,
	receivablesService ports.ReceivablesService,
) *CustomerHandler {
	return &CustomerHandler{
		customerService:    customerService,
		receivablesService: receivablesService,
	}
}

func (h *CustomerHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Customers
		r.Get("/customers", h.ListCustomers)
		r.Post("/customers", h.CreateCustomer)
		r.Get("/customers/{id}", h.GetCustomer)
		r.Put("/customers/{id}", h.UpdateCustomer)
		r.Patch("/customers/{id}/status", h.SetCustomerStatus)
		r.Patch("/customers/{id}/credit-hold", h.SetCreditHold)
		r.Get("/customers/{id}/aging", h.GetCustomerAging)
		r.Get("/customers/{id}/open-items", h.ListCustomerOpenItems)

		// Payment Terms
		r.Get("/payment-terms", h.ListPaymentTerms)
		r.Post("/payment-terms", h.CreatePaymentTerm)

		// Receipts (سندات القبض)
		r.Get("/receipts", h.ListReceipts)
		r.Post("/receipts", h.CreateReceipt)
		r.Get("/receipts/{id}", h.GetReceipt)
		r.Post("/receipts/{id}/allocate", h.AllocateReceipt)

		// Open Items (الفواتير المفتوحة)
		r.Post("/open-items", h.CreateOpenItem)
	})

	return r
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	cmd := ports.CreateCustomerCommand{
		LegalEntityID:  req.LegalEntityID,
		CustomerType:   req.CustomerType,
		NameEn:         req.NameEn,
		NameAr:         req.NameAr,
		TaxID:          req.TaxID,
		CRNumber:       req.CRNumber,
		NationalID:     req.NationalID,
		Address: domain.SaudiAddress{
			BuildingNo:   req.BuildingNo,
			StreetName:   req.StreetName,
			District:     req.District,
			City:         req.City,
			PostalCode:   req.PostalCode,
			AdditionalNo: req.AdditionalNo,
			CountryCode:  req.CountryCode,
		},
		Email:            req.Email,
		Phone:            req.Phone,
		ContactPerson:    req.ContactPerson,
		PaymentTermsID:   req.PaymentTermsID,
		CurrencyCode:     req.CurrencyCode,
		DefaultAccountID: req.DefaultAccountID,
		CreditLimit:      req.CreditLimit,
		MaxOverdueDays:   req.MaxOverdueDays,
		BlockAction:      req.BlockAction,
	}

	customer, err := h.customerService.CreateCustomer(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, customer)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid customer id"})
		return
	}

	customer, err := h.customerService.GetCustomer(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "customer not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	legalEntityID, _ := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	search := r.URL.Query().Get("search")
	custType := r.URL.Query().Get("customer_type")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := ports.CustomerFilter{
		LegalEntityID: legalEntityID,
		Search:        search,
		CustomerType:  custType,
		Status:        status,
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	customers, total, err := h.customerService.ListCustomers(r.Context(), filter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   customers,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid customer id"})
		return
	}

	var req UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	cmd := ports.UpdateCustomerCommand{
		ID:         id,
		NameEn:     req.NameEn,
		NameAr:     req.NameAr,
		TaxID:      req.TaxID,
		CRNumber:   req.CRNumber,
		NationalID: req.NationalID,
		Address: domain.SaudiAddress{
			BuildingNo:   req.BuildingNo,
			StreetName:   req.StreetName,
			District:     req.District,
			City:         req.City,
			PostalCode:   req.PostalCode,
			AdditionalNo: req.AdditionalNo,
			CountryCode:  req.CountryCode,
		},
		Email:            req.Email,
		Phone:            req.Phone,
		ContactPerson:    req.ContactPerson,
		PaymentTermsID:   req.PaymentTermsID,
		CurrencyCode:     req.CurrencyCode,
		DefaultAccountID: req.DefaultAccountID,
	}

	customer, err := h.customerService.UpdateCustomer(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) SetCustomerStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid customer id"})
		return
	}

	var req SetCustomerStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.customerService.SetCustomerStatus(r.Context(), id, req.Status); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *CustomerHandler) SetCreditHold(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid customer id"})
		return
	}

	var req SetCreditHoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.customerService.SetCreditHold(r.Context(), id, req.IsHold, req.Reason); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "credit hold updated"})
}

func (h *CustomerHandler) GetCustomerAging(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid customer id"})
		return
	}

	aging, err := h.receivablesService.GetCustomerAging(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, aging)
}

func (h *CustomerHandler) ListCustomerOpenItems(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid customer id"})
		return
	}
	openOnly := r.URL.Query().Get("open_only") != "false"

	items, err := h.receivablesService.ListOpenItems(r.Context(), id, openOnly)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (h *CustomerHandler) CreatePaymentTerm(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	term := &domain.PaymentTerm{
		LegalEntityID:   req.LegalEntityID,
		Code:            req.Code,
		NameEn:          req.NameEn,
		NameAr:          req.NameAr,
		DueDays:         req.DueDays,
		DiscountDays:    req.DiscountDays,
		DiscountPercent: req.DiscountPercent,
		IsActive:        req.IsActive,
	}

	created, err := h.customerService.CreatePaymentTerm(r.Context(), term)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *CustomerHandler) ListPaymentTerms(w http.ResponseWriter, r *http.Request) {
	legalEntityID, err := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id is required"})
		return
	}

	terms, err := h.customerService.ListPaymentTerms(r.Context(), legalEntityID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, terms)
}

func (h *CustomerHandler) CreateReceipt(w http.ResponseWriter, r *http.Request) {
	var req CreateReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	allocations := make([]ports.AllocationCommand, len(req.Allocations))
	for i, a := range req.Allocations {
		allocations[i] = ports.AllocationCommand{
			OpenItemID:      a.OpenItemID,
			AllocatedAmount: a.AllocatedAmount,
		}
	}

	cmd := ports.CreateReceiptCommand{
		LegalEntityID:    req.LegalEntityID,
		CustomerID:       req.CustomerID,
		ReceiptDate:      req.ReceiptDate,
		PaymentMethod:    req.PaymentMethod,
		DepositAccountID: req.DepositAccountID,
		Amount:           req.Amount,
		CurrencyCode:     req.CurrencyCode,
		ReferenceNumber:  req.ReferenceNumber,
		Notes:            req.Notes,
		Allocations:      allocations,
	}

	receipt, err := h.receivablesService.CreateReceipt(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, receipt)
}

func (h *CustomerHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid receipt id"})
		return
	}

	receipt, err := h.receivablesService.GetReceipt(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrReceiptNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "receipt not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, receipt)
}

func (h *CustomerHandler) ListReceipts(w http.ResponseWriter, r *http.Request) {
	legalEntityID, _ := uuid.Parse(r.URL.Query().Get("legal_entity_id"))
	custIDStr := r.URL.Query().Get("customer_id")
	var custID *uuid.UUID
	if custIDStr != "" {
		if parsed, err := uuid.Parse(custIDStr); err == nil {
			custID = &parsed
		}
	}
	status := r.URL.Query().Get("status")
	method := r.URL.Query().Get("payment_method")
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := ports.ReceiptFilter{
		LegalEntityID: legalEntityID,
		CustomerID:    custID,
		Status:        status,
		PaymentMethod: method,
		Search:        search,
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	receipts, total, err := h.receivablesService.ListReceipts(r.Context(), filter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   receipts,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *CustomerHandler) AllocateReceipt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid receipt id"})
		return
	}

	var req AllocateReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	allocations := make([]ports.AllocationCommand, len(req.Allocations))
	for i, a := range req.Allocations {
		allocations[i] = ports.AllocationCommand{
			OpenItemID:      a.OpenItemID,
			AllocatedAmount: a.AllocatedAmount,
		}
	}

	if err := h.receivablesService.AllocateReceipt(r.Context(), id, allocations, nil); err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "allocated"})
}

func (h *CustomerHandler) CreateOpenItem(w http.ResponseWriter, r *http.Request) {
	var req CreateOpenItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	var docDate, dueDate time.Time
	if req.DocumentDate != "" {
		docDate, _ = time.Parse("2006-01-02", req.DocumentDate)
	}
	if docDate.IsZero() {
		docDate = time.Now()
	}
	if req.DueDate != "" {
		dueDate, _ = time.Parse("2006-01-02", req.DueDate)
	}
	if dueDate.IsZero() {
		dueDate = docDate.AddDate(0, 0, 30)
	}

	currencyCode := req.CurrencyCode
	if currencyCode == "" {
		currencyCode = "SAR"
	}

	item := &domain.OpenItem{
		LegalEntityID:  req.LegalEntityID,
		CustomerID:     req.CustomerID,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		DocumentID:     req.DocumentID,
		DocumentDate:   docDate,
		DueDate:        dueDate,
		CurrencyCode:   currencyCode,
		OriginalAmount: req.OriginalAmount,
	}

	created, err := h.receivablesService.CreateOpenItem(r.Context(), item)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
