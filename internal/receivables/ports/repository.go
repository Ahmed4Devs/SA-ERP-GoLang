package ports

import (
	"context"
	"erp/internal/receivables/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CustomerFilter struct {
	LegalEntityID uuid.UUID
	Search        string
	CustomerType  string
	Status        string
	Limit         int32
	Offset        int32
}

type ReceiptFilter struct {
	LegalEntityID uuid.UUID
	CustomerID    *uuid.UUID
	Status        string
	PaymentMethod string
	Search        string
	Limit         int32
	Offset        int32
}

type CustomerRepository interface {
	CreateCustomer(ctx context.Context, customer *domain.Customer) (*domain.Customer, error)
	GetCustomerByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error)
	GetCustomerByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Customer, error)
	ListCustomers(ctx context.Context, filter CustomerFilter) ([]*domain.Customer, int64, error)
	UpdateCustomer(ctx context.Context, customer *domain.Customer) (*domain.Customer, error)
	UpdateCustomerStatus(ctx context.Context, id uuid.UUID, status domain.CustomerStatus) error

	CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error)
	GetPaymentTermByID(ctx context.Context, id uuid.UUID) (*domain.PaymentTerm, error)
	ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID, includeInactive bool) ([]*domain.PaymentTerm, error)

	UpsertCustomerAccount(ctx context.Context, account *domain.CustomerAccount) (*domain.CustomerAccount, error)
	GetCustomerAccounts(ctx context.Context, customerID uuid.UUID) ([]*domain.CustomerAccount, error)

	UpsertCreditLimit(ctx context.Context, limit *domain.CreditLimit) (*domain.CreditLimit, error)
	GetCreditLimit(ctx context.Context, customerID uuid.UUID) (*domain.CreditLimit, error)
	UpdateCustomerBalance(ctx context.Context, customerID uuid.UUID, delta decimal.Decimal) (*domain.CreditLimit, error)
	SetCreditHold(ctx context.Context, customerID uuid.UUID, isHold bool, reason *string) error
}

type ReceivablesRepository interface {
	CreateReceipt(ctx context.Context, receipt *domain.Receipt) (*domain.Receipt, error)
	GetReceiptByID(ctx context.Context, id uuid.UUID) (*domain.Receipt, error)
	ListReceipts(ctx context.Context, filter ReceiptFilter) ([]*domain.Receipt, int64, error)
	UpdateReceiptStatus(ctx context.Context, id uuid.UUID, status domain.ReceiptStatus) error
	UpdateReceiptAllocationAmounts(ctx context.Context, id uuid.UUID, allocated, unallocated decimal.Decimal) error

	CreateReceiptAllocation(ctx context.Context, allocation *domain.ReceiptAllocation) (*domain.ReceiptAllocation, error)
	ListReceiptAllocations(ctx context.Context, receiptID uuid.UUID) ([]*domain.ReceiptAllocation, error)

	CreateOpenItem(ctx context.Context, item *domain.OpenItem) (*domain.OpenItem, error)
	GetOpenItemByID(ctx context.Context, id uuid.UUID) (*domain.OpenItem, error)
	LockOpenItem(ctx context.Context, id uuid.UUID) (*domain.OpenItem, error)
	ListOpenItemsByCustomer(ctx context.Context, customerID uuid.UUID, openOnly bool) ([]*domain.OpenItem, error)
	ListOpenItemsByLegalEntity(ctx context.Context, legalEntityID uuid.UUID, openOnly bool, limit, offset int32) ([]*domain.OpenItem, error)
	UpdateOpenItemAllocation(ctx context.Context, id uuid.UUID, allocated, remaining decimal.Decimal, status domain.OpenItemStatus) (*domain.OpenItem, error)
	GetCustomerAgingSummary(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAgingSummary, error)
}
