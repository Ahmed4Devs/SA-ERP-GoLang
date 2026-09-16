package ports

import (
	"context"
	"erp/internal/payables/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SupplierFilter struct {
	LegalEntityID uuid.UUID
	Search        string
	SupplierType  string
	Status        string
	Limit         int32
	Offset        int32
}

type SupplierInvoiceFilter struct {
	LegalEntityID uuid.UUID
	Search        string
	SupplierID    *uuid.UUID
	Status        string
	PaymentStatus string
	Limit         int32
	Offset        int32
}

type SupplierPaymentFilter struct {
	LegalEntityID uuid.UUID
	SupplierID    *uuid.UUID
	Search        string
	Limit         int32
	Offset        int32
}

type SupplierRepository interface {
	CreateSupplier(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error)
	GetSupplierByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	GetSupplierByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Supplier, error)
	ListSuppliers(ctx context.Context, filter SupplierFilter) ([]*domain.Supplier, int64, error)
	UpdateSupplier(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error)
	SetSupplierStatus(ctx context.Context, id uuid.UUID, status domain.SupplierStatus) error

	CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error)
	GetPaymentTermByID(ctx context.Context, id uuid.UUID) (*domain.PaymentTerm, error)
	ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID, includeInactive bool) ([]*domain.PaymentTerm, error)

	CreateSupplierAccount(ctx context.Context, supplierID, glAccountID uuid.UUID, currencyCode string, isPrimary bool) error
}

type PayableRepository interface {
	// Supplier Invoices
	CreateSupplierInvoice(ctx context.Context, inv *domain.SupplierInvoice) (*domain.SupplierInvoice, error)
	GetSupplierInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.SupplierInvoice, error)
	GetSupplierInvoiceByNumber(ctx context.Context, legalEntityID uuid.UUID, invoiceNumber string) (*domain.SupplierInvoice, error)
	ListSupplierInvoices(ctx context.Context, filter SupplierInvoiceFilter) ([]*domain.SupplierInvoice, int64, error)
	UpdateSupplierInvoiceStatus(ctx context.Context, id uuid.UUID, status domain.SupplierInvoiceStatus, journalEntryID *uuid.UUID) error
	UpdateSupplierInvoicePaymentStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error
	ListSupplierInvoiceLines(ctx context.Context, invoiceID uuid.UUID) ([]domain.SupplierInvoiceLine, error)

	// Credit Notes
	CreateSupplierCreditNote(ctx context.Context, cn *domain.SupplierCreditNote) (*domain.SupplierCreditNote, error)
	GetSupplierCreditNoteByID(ctx context.Context, id uuid.UUID) (*domain.SupplierCreditNote, error)
	ListSupplierCreditNotes(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID) ([]*domain.SupplierCreditNote, error)

	// Payments
	CreateSupplierPayment(ctx context.Context, p *domain.SupplierPayment) (*domain.SupplierPayment, error)
	GetSupplierPaymentByID(ctx context.Context, id uuid.UUID) (*domain.SupplierPayment, error)
	ListSupplierPayments(ctx context.Context, filter SupplierPaymentFilter) ([]*domain.SupplierPayment, int64, error)
	UpdateSupplierPaymentStatus(ctx context.Context, id uuid.UUID, status string, allocated, unallocated decimal.Decimal, journalEntryID *uuid.UUID) error

	// Allocations
	CreatePaymentAllocation(ctx context.Context, alloc *domain.PaymentAllocation) (*domain.PaymentAllocation, error)
	ListPaymentAllocations(ctx context.Context, paymentID uuid.UUID) ([]domain.PaymentAllocation, error)

	// Open Items
	CreateOpenItem(ctx context.Context, item *domain.PayablesOpenItem) (*domain.PayablesOpenItem, error)
	GetOpenItemByID(ctx context.Context, id uuid.UUID) (*domain.PayablesOpenItem, error)
	LockOpenItem(ctx context.Context, id uuid.UUID) (*domain.PayablesOpenItem, error)
	ListOpenItemsBySupplier(ctx context.Context, supplierID uuid.UUID, openOnly bool) ([]*domain.PayablesOpenItem, error)
	ListOpenItemsByLegalEntity(ctx context.Context, legalEntityID uuid.UUID, openOnly bool, limit, offset int32) ([]*domain.PayablesOpenItem, error)
	UpdateOpenItemAllocation(ctx context.Context, id uuid.UUID, allocated, remaining decimal.Decimal, status domain.OpenItemStatus) (*domain.PayablesOpenItem, error)

	// Aging
	GetSupplierAgingSummary(ctx context.Context, supplierID uuid.UUID) (*domain.SupplierAgingSummary, error)
}
