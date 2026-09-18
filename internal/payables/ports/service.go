package ports

import (
	"context"
	"erp/internal/payables/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateSupplierCmd struct {
	LegalEntityID             uuid.UUID
	SupplierNumber            string
	SupplierType              domain.SupplierType
	NameEn                    string
	NameAr                    string
	TaxID                     *string
	CRNumber                  *string
	NationalID                *string
	IBAN                      *string
	BankName                  *string
	SwiftCode                 *string
	Address                   domain.SaudiAddress
	Email                     *string
	Phone                     *string
	ContactPerson             *string
	Website                   *string
	PaymentTermsID            *uuid.UUID
	CurrencyCode              string
	DefaultExpenseAccountID   *uuid.UUID
	DefaultLiabilityAccountID *uuid.UUID
}

type CreateSupplierInvoiceCmd struct {
	LegalEntityID         uuid.UUID
	InvoiceNumber         string
	SupplierInvoiceNumber string
	SupplierID            uuid.UUID
	InvoiceDate           string // YYYY-MM-DD
	DueDate               string // YYYY-MM-DD
	CurrencyCode          string
	ExchangeRate          decimal.Decimal
	Notes                 *string
	CreatedByUserID       *uuid.UUID
	Lines                 []CreateSupplierInvoiceLineCmd
}

type CreateSupplierInvoiceLineCmd struct {
	ItemID           *uuid.UUID
	ExpenseAccountID uuid.UUID
	Description      string
	Quantity         decimal.Decimal
	UnitPrice        decimal.Decimal
	DiscountPercent  decimal.Decimal
	TaxRate          decimal.Decimal
}

type CreateSupplierPaymentCmd struct {
	LegalEntityID         uuid.UUID
	PaymentNumber         string
	SupplierID            uuid.UUID
	PaymentDate           string // YYYY-MM-DD
	PaymentMethod         domain.PaymentMethod
	DisbursementAccountID uuid.UUID
	Amount                decimal.Decimal
	CurrencyCode          string
	ReferenceNumber       *string
	Notes                 *string
	CreatedByUserID       *uuid.UUID
	Allocations           []PaymentAllocationCmd
}

type PaymentAllocationCmd struct {
	OpenItemID uuid.UUID
	Amount     decimal.Decimal
}

type CreateCreditNoteCmd struct {
	LegalEntityID            uuid.UUID
	CreditNoteNumber         string
	SupplierCreditNoteNumber *string
	SupplierID               uuid.UUID
	OriginalInvoiceID        *uuid.UUID
	CreditNoteDate           string
	CurrencyCode             string
	Reason                   *string
	CreatedByUserID          *uuid.UUID
	Lines                    []CreateCreditNoteLineCmd
}

type CreateCreditNoteLineCmd struct {
	ItemID           *uuid.UUID
	ExpenseAccountID uuid.UUID
	Description      string
	Quantity         decimal.Decimal
	UnitPrice        decimal.Decimal
	TaxRate          decimal.Decimal
}

type SupplierService interface {
	CreateSupplier(ctx context.Context, cmd CreateSupplierCmd) (*domain.Supplier, error)
	GetSupplier(ctx context.Context, id uuid.UUID) (*domain.Supplier, error)
	ListSuppliers(ctx context.Context, filter SupplierFilter) ([]*domain.Supplier, int64, error)
	UpdateSupplier(ctx context.Context, s *domain.Supplier) (*domain.Supplier, error)
	SetSupplierStatus(ctx context.Context, id uuid.UUID, status domain.SupplierStatus) error

	CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error)
	ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.PaymentTerm, error)
}

type PayableService interface {
	// Supplier Invoices
	CreateSupplierInvoice(ctx context.Context, cmd CreateSupplierInvoiceCmd) (*domain.SupplierInvoice, error)
	GetSupplierInvoice(ctx context.Context, id uuid.UUID) (*domain.SupplierInvoice, error)
	ListSupplierInvoices(ctx context.Context, filter SupplierInvoiceFilter) ([]*domain.SupplierInvoice, int64, error)
	PostSupplierInvoice(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.SupplierInvoice, error)

	// Supplier Payments
	CreateSupplierPayment(ctx context.Context, cmd CreateSupplierPaymentCmd) (*domain.SupplierPayment, error)
	GetSupplierPayment(ctx context.Context, id uuid.UUID) (*domain.SupplierPayment, error)
	ListSupplierPayments(ctx context.Context, filter SupplierPaymentFilter) ([]*domain.SupplierPayment, int64, error)
	PostSupplierPayment(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.SupplierPayment, error)

	// Credit Notes
	CreateSupplierCreditNote(ctx context.Context, cmd CreateCreditNoteCmd) (*domain.SupplierCreditNote, error)
	GetSupplierCreditNote(ctx context.Context, id uuid.UUID) (*domain.SupplierCreditNote, error)
	ListSupplierCreditNotes(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID) ([]*domain.SupplierCreditNote, error)

	// Open Items & Aging
	ListOpenItems(ctx context.Context, supplierID uuid.UUID, openOnly bool) ([]*domain.PayablesOpenItem, error)
	GetSupplierAgingSummary(ctx context.Context, supplierID uuid.UUID) (*domain.SupplierAgingSummary, error)
}
