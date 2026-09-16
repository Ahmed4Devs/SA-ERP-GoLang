package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PayableHandler struct {
	supplierService ports.SupplierService
	payableService  ports.PayableService
}

func NewPayableHandler(
	supplierService ports.SupplierService,
	payableService ports.PayableService,
) *PayableHandler {
	return &PayableHandler{
		supplierService: supplierService,
		payableService:  payableService,
	}
}

func (h *PayableHandler) Routes(authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		if authMiddleware != nil {
			r.Use(authMiddleware)
		}

		// Suppliers
		r.Get("/suppliers", h.ListSuppliers)
		r.Post("/suppliers", h.CreateSupplier)
		r.Get("/suppliers/{id}", h.GetSupplier)
		r.Put("/suppliers/{id}", h.UpdateSupplier)
		r.Patch("/suppliers/{id}/status", h.SetSupplierStatus)
		r.Get("/suppliers/{id}/aging", h.GetSupplierAging)
		r.Get("/suppliers/{id}/open-items", h.ListSupplierOpenItems)

		// Payment Terms
		r.Get("/payment-terms", h.ListPaymentTerms)
		r.Post("/payment-terms", h.CreatePaymentTerm)

		// Supplier Invoices
		r.Get("/invoices", h.ListInvoices)
		r.Post("/invoices", h.CreateInvoice)
		r.Get("/invoices/{id}", h.GetInvoice)
		r.Post("/invoices/{id}/post", h.PostInvoice)

		// Supplier Payments
		r.Get("/payments", h.ListPayments)
		r.Post("/payments", h.CreatePayment)
		r.Get("/payments/{id}", h.GetPayment)
		r.Post("/payments/{id}/post", h.PostPayment)

		// Credit Notes
		r.Get("/credit-notes", h.ListCreditNotes)
		r.Post("/credit-notes", h.CreateCreditNote)
		r.Get("/credit-notes/{id}", h.GetCreditNote)
	})

	return r
}

// Supplier Handlers

func (h *PayableHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	cmd := ports.CreateSupplierCmd{
		LegalEntityID:  req.LegalEntityID,
		SupplierNumber: req.SupplierNumber,
		SupplierType:   req.SupplierType,
		NameEn:         req.NameEn,
		NameAr:         req.NameAr,
		TaxID:          req.TaxID,
		CRNumber:       req.CRNumber,
		NationalID:     req.NationalID,
		IBAN:           req.IBAN,
		BankName:       req.BankName,
		SwiftCode:      req.SwiftCode,
		Address: domain.SaudiAddress{
			BuildingNo:   req.BuildingNo,
			StreetName:   req.StreetName,
			District:     req.District,
			City:         req.City,
			PostalCode:   req.PostalCode,
			AdditionalNo: req.AdditionalNo,
			CountryCode:  req.CountryCode,
		},
		Email:                     req.Email,
		Phone:                     req.Phone,
		ContactPerson:             req.ContactPerson,
		Website:                   req.Website,
		PaymentTermsID:            req.PaymentTermsID,
		CurrencyCode:              req.CurrencyCode,
		DefaultExpenseAccountID:   req.DefaultExpenseAccountID,
		DefaultLiabilityAccountID: req.DefaultLiabilityAccountID,
	}

	created, err := h.supplierService.CreateSupplier(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *PayableHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid supplier ID"})
		return
	}

	supp, err := h.supplierService.GetSupplier(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "supplier not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, supp)
}

func (h *PayableHandler) ListSuppliers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	leIDStr := q.Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id is required"})
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ports.SupplierFilter{
		LegalEntityID: leID,
		Search:        q.Get("search"),
		SupplierType:  q.Get("type"),
		Status:        q.Get("status"),
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	suppliers, total, err := h.supplierService.ListSuppliers(r.Context(), filter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   suppliers,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *PayableHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid supplier ID"})
		return
	}

	var req UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	supp, err := h.supplierService.GetSupplier(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "supplier not found"})
		return
	}

	supp.NameEn = req.NameEn
	supp.NameAr = req.NameAr
	supp.TaxID = req.TaxID
	supp.CRNumber = req.CRNumber
	supp.NationalID = req.NationalID
	supp.IBAN = req.IBAN
	supp.BankName = req.BankName
	supp.SwiftCode = req.SwiftCode
	supp.Address.BuildingNo = req.BuildingNo
	supp.Address.StreetName = req.StreetName
	supp.Address.District = req.District
	supp.Address.City = req.City
	supp.Address.PostalCode = req.PostalCode
	supp.Address.AdditionalNo = req.AdditionalNo
	if req.CountryCode != "" {
		supp.Address.CountryCode = req.CountryCode
	}
	supp.Email = req.Email
	supp.Phone = req.Phone
	supp.ContactPerson = req.ContactPerson
	supp.Website = req.Website
	supp.PaymentTermsID = req.PaymentTermsID
	supp.DefaultExpenseAccountID = req.DefaultExpenseAccountID
	supp.DefaultLiabilityAccountID = req.DefaultLiabilityAccountID

	updated, err := h.supplierService.UpdateSupplier(r.Context(), supp)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (h *PayableHandler) SetSupplierStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid supplier ID"})
		return
	}

	var req SetSupplierStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.supplierService.SetSupplierStatus(r.Context(), id, req.Status); err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": string(req.Status)})
}

func (h *PayableHandler) GetSupplierAging(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid supplier ID"})
		return
	}

	summary, err := h.payableService.GetSupplierAgingSummary(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (h *PayableHandler) ListSupplierOpenItems(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid supplier ID"})
		return
	}

	openOnly := r.URL.Query().Get("open_only") != "false"
	items, err := h.payableService.ListOpenItems(r.Context(), id, openOnly)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// Payment Terms

func (h *PayableHandler) ListPaymentTerms(w http.ResponseWriter, r *http.Request) {
	leIDStr := r.URL.Query().Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id is required"})
		return
	}

	terms, err := h.supplierService.ListPaymentTerms(r.Context(), leID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, terms)
}

func (h *PayableHandler) CreatePaymentTerm(w http.ResponseWriter, r *http.Request) {
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

	created, err := h.supplierService.CreatePaymentTerm(r.Context(), term)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

// Supplier Invoices

func (h *PayableHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateSupplierInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	var lines []ports.CreateSupplierInvoiceLineCmd
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateSupplierInvoiceLineCmd{
			ItemID:           l.ItemID,
			ExpenseAccountID: l.ExpenseAccountID,
			Description:      l.Description,
			Quantity:         l.Quantity,
			UnitPrice:        l.UnitPrice,
			DiscountPercent:  l.DiscountPercent,
			TaxRate:          l.TaxRate,
		})
	}

	cmd := ports.CreateSupplierInvoiceCmd{
		LegalEntityID:         req.LegalEntityID,
		InvoiceNumber:         req.InvoiceNumber,
		SupplierInvoiceNumber: req.SupplierInvoiceNumber,
		SupplierID:            req.SupplierID,
		InvoiceDate:           req.InvoiceDate,
		DueDate:               req.DueDate,
		CurrencyCode:          req.CurrencyCode,
		ExchangeRate:          req.ExchangeRate,
		Notes:                 req.Notes,
		Lines:                 lines,
	}

	created, err := h.payableService.CreateSupplierInvoice(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *PayableHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice ID"})
		return
	}

	inv, err := h.payableService.GetSupplierInvoice(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "supplier invoice not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, inv)
}

func (h *PayableHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	leIDStr := q.Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id is required"})
		return
	}

	var suppID *uuid.UUID
	if sStr := q.Get("supplier_id"); sStr != "" {
		if sUUID, err := uuid.Parse(sStr); err == nil {
			suppID = &sUUID
		}
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ports.SupplierInvoiceFilter{
		LegalEntityID: leID,
		Search:        q.Get("search"),
		SupplierID:    suppID,
		Status:        q.Get("status"),
		PaymentStatus: q.Get("payment_status"),
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	invoices, total, err := h.payableService.ListSupplierInvoices(r.Context(), filter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   invoices,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *PayableHandler) PostInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid invoice ID"})
		return
	}

	posted, err := h.payableService.PostSupplierInvoice(r.Context(), id, nil)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, posted)
}

// Supplier Payments

func (h *PayableHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreateSupplierPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	var allocs []ports.PaymentAllocationCmd
	for _, a := range req.Allocations {
		allocs = append(allocs, ports.PaymentAllocationCmd{
			OpenItemID: a.OpenItemID,
			Amount:     a.Amount,
		})
	}

	cmd := ports.CreateSupplierPaymentCmd{
		LegalEntityID:         req.LegalEntityID,
		PaymentNumber:         req.PaymentNumber,
		SupplierID:            req.SupplierID,
		PaymentDate:           req.PaymentDate,
		PaymentMethod:         req.PaymentMethod,
		DisbursementAccountID: req.DisbursementAccountID,
		Amount:                req.Amount,
		CurrencyCode:          req.CurrencyCode,
		ReferenceNumber:       req.ReferenceNumber,
		Notes:                 req.Notes,
		Allocations:           allocs,
	}

	created, err := h.payableService.CreateSupplierPayment(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *PayableHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment ID"})
		return
	}

	p, err := h.payableService.GetSupplierPayment(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "supplier payment not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, p)
}

func (h *PayableHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	leIDStr := q.Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id is required"})
		return
	}

	var suppID *uuid.UUID
	if sStr := q.Get("supplier_id"); sStr != "" {
		if sUUID, err := uuid.Parse(sStr); err == nil {
			suppID = &sUUID
		}
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ports.SupplierPaymentFilter{
		LegalEntityID: leID,
		SupplierID:    suppID,
		Search:        q.Get("search"),
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	payments, total, err := h.payableService.ListSupplierPayments(r.Context(), filter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   payments,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

func (h *PayableHandler) PostPayment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment ID"})
		return
	}

	posted, err := h.payableService.PostSupplierPayment(r.Context(), id, nil)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, posted)
}

// Supplier Credit Notes

func (h *PayableHandler) CreateCreditNote(w http.ResponseWriter, r *http.Request) {
	var req CreateCreditNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: " + err.Error()})
		return
	}

	var lines []ports.CreateCreditNoteLineCmd
	for _, l := range req.Lines {
		lines = append(lines, ports.CreateCreditNoteLineCmd{
			ItemID:           l.ItemID,
			ExpenseAccountID: l.ExpenseAccountID,
			Description:      l.Description,
			Quantity:         l.Quantity,
			UnitPrice:        l.UnitPrice,
			TaxRate:          l.TaxRate,
		})
	}

	cmd := ports.CreateCreditNoteCmd{
		LegalEntityID:            req.LegalEntityID,
		CreditNoteNumber:         req.CreditNoteNumber,
		SupplierCreditNoteNumber: req.SupplierCreditNoteNumber,
		SupplierID:               req.SupplierID,
		OriginalInvoiceID:        req.OriginalInvoiceID,
		CreditNoteDate:           req.CreditNoteDate,
		CurrencyCode:             req.CurrencyCode,
		Reason:                   req.Reason,
		Lines:                    lines,
	}

	created, err := h.payableService.CreateSupplierCreditNote(r.Context(), cmd)
	if err != nil {
		respondJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, created)
}

func (h *PayableHandler) GetCreditNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid credit note ID"})
		return
	}

	cn, err := h.payableService.GetSupplierCreditNote(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCreditNoteNotFound) {
			respondJSON(w, http.StatusNotFound, map[string]string{"error": "supplier credit note not found"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, cn)
}

func (h *PayableHandler) ListCreditNotes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	leIDStr := q.Get("legal_entity_id")
	leID, err := uuid.Parse(leIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "valid legal_entity_id is required"})
		return
	}

	var suppID *uuid.UUID
	if sStr := q.Get("supplier_id"); sStr != "" {
		if sUUID, err := uuid.Parse(sStr); err == nil {
			suppID = &sUUID
		}
	}

	notes, err := h.payableService.ListSupplierCreditNotes(r.Context(), leID, suppID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, notes)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
