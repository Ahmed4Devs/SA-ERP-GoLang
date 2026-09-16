package ports

import (
	"context"
	"erp/internal/receivables/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateCustomerCommand struct {
	LegalEntityID    uuid.UUID
	CustomerType     domain.CustomerType
	NameEn           string
	NameAr           string
	TaxID            *string
	CRNumber         *string
	NationalID       *string
	Address          domain.SaudiAddress
	Email            *string
	Phone            *string
	ContactPerson    *string
	PaymentTermsID   *uuid.UUID
	CurrencyCode     string
	DefaultAccountID *uuid.UUID
	CreditLimit      decimal.Decimal
	MaxOverdueDays   int32
	BlockAction      domain.BlockAction
}

type UpdateCustomerCommand struct {
	ID               uuid.UUID
	NameEn           string
	NameAr           string
	TaxID            *string
	CRNumber         *string
	NationalID       *string
	Address          domain.SaudiAddress
	Email            *string
	Phone            *string
	ContactPerson    *string
	PaymentTermsID   *uuid.UUID
	CurrencyCode     string
	DefaultAccountID *uuid.UUID
}

type CreateReceiptCommand struct {
	LegalEntityID    uuid.UUID
	CustomerID       uuid.UUID
	ReceiptDate      string
	PaymentMethod    domain.PaymentMethod
	DepositAccountID uuid.UUID
	Amount           decimal.Decimal
	CurrencyCode     string
	ReferenceNumber  *string
	Notes            *string
	CreatedByUserID  *uuid.UUID
	Allocations      []AllocationCommand
}

type AllocationCommand struct {
	OpenItemID      uuid.UUID
	AllocatedAmount decimal.Decimal
}

type CustomerService interface {
	CreateCustomer(ctx context.Context, cmd CreateCustomerCommand) (*domain.Customer, error)
	GetCustomer(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	ListCustomers(ctx context.Context, filter CustomerFilter) ([]*domain.Customer, int64, error)
	UpdateCustomer(ctx context.Context, cmd UpdateCustomerCommand) (*domain.Customer, error)
	SetCustomerStatus(ctx context.Context, id uuid.UUID, status domain.CustomerStatus) error
	SetCreditHold(ctx context.Context, customerID uuid.UUID, isHold bool, reason *string) error

	CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error)
	ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.PaymentTerm, error)
}

type ReceivablesService interface {
	CreateReceipt(ctx context.Context, cmd CreateReceiptCommand) (*domain.Receipt, error)
	GetReceipt(ctx context.Context, id uuid.UUID) (*domain.Receipt, error)
	ListReceipts(ctx context.Context, filter ReceiptFilter) ([]*domain.Receipt, int64, error)
	AllocateReceipt(ctx context.Context, receiptID uuid.UUID, allocations []AllocationCommand, userID *uuid.UUID) error

	CreateOpenItem(ctx context.Context, item *domain.OpenItem) (*domain.OpenItem, error)
	ListOpenItems(ctx context.Context, customerID uuid.UUID, openOnly bool) ([]*domain.OpenItem, error)
	GetCustomerAging(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAgingSummary, error)
}
