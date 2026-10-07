package ports

import (
	"context"
	"erp/internal/treasury/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TreasuryRepository interface {
	// Cash Accounts
	CreateCashAccount(ctx context.Context, account *domain.CashAccount) (*domain.CashAccount, error)
	GetCashAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.CashAccount, error)
	LockCashAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.CashAccount, error)
	GetCashAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.CashAccount, error)
	ListCashAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.CashAccount, error)
	UpdateCashAccountBalance(ctx context.Context, legalEntityID, id uuid.UUID, delta decimal.Decimal) (*domain.CashAccount, error)
	UpdateCashAccountStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.CashAccountStatus) (*domain.CashAccount, error)

	// Receipts
	CreateReceipt(ctx context.Context, receipt *domain.Receipt) (*domain.Receipt, error)
	GetReceiptByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Receipt, error)
	LockReceiptByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Receipt, error)
	GetReceiptByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Receipt, error)
	ListReceipts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Receipt, error)
	UpdateReceiptStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.ReceiptStatus, journalEntryID *uuid.UUID) (*domain.Receipt, error)

	// Payments
	CreatePayment(ctx context.Context, payment *domain.Payment) (*domain.Payment, error)
	GetPaymentByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Payment, error)
	LockPaymentByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Payment, error)
	GetPaymentByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Payment, error)
	ListPayments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Payment, error)
	UpdatePaymentStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.PaymentStatus, journalEntryID, approvedBy *uuid.UUID) (*domain.Payment, error)

	// Transfers
	CreateTransfer(ctx context.Context, transfer *domain.Transfer) (*domain.Transfer, error)
	GetTransferByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error)
	LockTransferByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error)
	GetTransferByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Transfer, error)
	ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Transfer, error)
	UpdateTransferStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.TransferStatus, journalEntryID, completedBy *uuid.UUID) (*domain.Transfer, error)
	LockBankAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.BankAccountBalance, error)
	UpdateBankAccountBalance(ctx context.Context, legalEntityID, id uuid.UUID, delta decimal.Decimal) (*domain.BankAccountBalance, error)
}

type TreasuryService interface {
	CreateCashAccount(ctx context.Context, account *domain.CashAccount) (*domain.CashAccount, error)
	GetCashAccount(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.CashAccount, error)
	ListCashAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.CashAccount, error)

	CreateReceipt(ctx context.Context, receipt *domain.Receipt) (*domain.Receipt, error)
	PostReceipt(ctx context.Context, legalEntityID, id uuid.UUID, actorID uuid.UUID) (*domain.Receipt, error)
	GetReceipt(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Receipt, error)
	ListReceipts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Receipt, error)

	CreatePayment(ctx context.Context, payment *domain.Payment) (*domain.Payment, error)
	ApprovePayment(ctx context.Context, legalEntityID, id uuid.UUID, approverID uuid.UUID) (*domain.Payment, error)
	PayPayment(ctx context.Context, legalEntityID, id uuid.UUID, actorID uuid.UUID) (*domain.Payment, error)
	GetPayment(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Payment, error)
	ListPayments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Payment, error)

	CreateTransfer(ctx context.Context, transfer *domain.Transfer) (*domain.Transfer, error)
	CompleteTransfer(ctx context.Context, legalEntityID, id uuid.UUID, completedBy uuid.UUID) (*domain.Transfer, error)
	GetTransfer(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error)
	ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Transfer, error)
}
