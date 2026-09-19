package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrCashAccountNotFound      = errors.New("cash account not found")
	ErrBankAccountNotFound      = errors.New("bank account not found")
	ErrReceiptNotFound          = errors.New("treasury receipt not found")
	ErrPaymentNotFound          = errors.New("treasury payment not found")
	ErrTransferNotFound         = errors.New("treasury transfer not found")
	ErrCashAccountLimitExceeded = errors.New("cash balance exceeds account maximum limit")
	ErrInsufficientCashBalance  = errors.New("insufficient cash balance in account")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
	ErrInvalidAmount            = errors.New("amount must be greater than zero")
	ErrInvalidTransferAccounts  = errors.New("transfer source and destination accounts are invalid")
	ErrInactiveTreasuryAccount  = errors.New("treasury account is not active")
	ErrCurrencyMismatch         = errors.New("transaction currency does not match treasury account currency")
)

type CashAccountType string

const (
	CashAccountPettyCash    CashAccountType = "PETTY_CASH"
	CashAccountCashRegister CashAccountType = "CASH_REGISTER"
	CashAccountMainVault    CashAccountType = "MAIN_VAULT"
)

type CashAccountStatus string

const (
	CashAccountActive    CashAccountStatus = "ACTIVE"
	CashAccountSuspended CashAccountStatus = "SUSPENDED"
	CashAccountClosed    CashAccountStatus = "CLOSED"
)

type ReceiptType string

const (
	ReceiptCustomerPayment ReceiptType = "CUSTOMER_PAYMENT"
	ReceiptMiscellaneous   ReceiptType = "MISCELLANEOUS"
	ReceiptCashSale        ReceiptType = "CASH_SALE"
	ReceiptDeposit         ReceiptType = "DEPOSIT"
	ReceiptInterBranch     ReceiptType = "INTER_BRANCH"
)

type PaymentMethod string

const (
	PaymentMethodCash   PaymentMethod = "CASH"
	PaymentMethodCheque PaymentMethod = "CHEQUE"
)

type ReceiptStatus string

const (
	ReceiptDraft  ReceiptStatus = "DRAFT"
	ReceiptPosted ReceiptStatus = "POSTED"
	ReceiptVoided ReceiptStatus = "VOIDED"
)

type PaymentType string

const (
	PaymentPettyCashExpense PaymentType = "PETTY_CASH_EXPENSE"
	PaymentSupplierPayment  PaymentType = "SUPPLIER_PAYMENT"
	PaymentCashAdvance      PaymentType = "CASH_ADVANCE"
	PaymentMiscellaneous    PaymentType = "MISCELLANEOUS"
)

type PaymentStatus string

const (
	PaymentDraft    PaymentStatus = "DRAFT"
	PaymentApproved PaymentStatus = "APPROVED"
	PaymentPaid     PaymentStatus = "PAID"
	PaymentVoided   PaymentStatus = "VOIDED"
)

type TransferSourceType string

const (
	TransferSourceCash TransferSourceType = "CASH"
	TransferSourceBank TransferSourceType = "BANK"
)

type TransferDestinationType string

const (
	TransferDestinationCash TransferDestinationType = "CASH"
	TransferDestinationBank TransferDestinationType = "BANK"
)

type TransferStatus string

const (
	TransferDraft     TransferStatus = "DRAFT"
	TransferInTransit TransferStatus = "IN_TRANSIT"
	TransferCompleted TransferStatus = "COMPLETED"
	TransferCancelled TransferStatus = "CANCELLED"
)

type CashAccount struct {
	ID              uuid.UUID         `json:"id"`
	LegalEntityID   uuid.UUID         `json:"legal_entity_id"`
	BranchID        *uuid.UUID        `json:"branch_id,omitempty"`
	AccountCode     string            `json:"account_code"`
	NameEn          string            `json:"name_en"`
	NameAr          string            `json:"name_ar"`
	AccountType     CashAccountType   `json:"account_type"`
	GLAccountID     *uuid.UUID        `json:"gl_account_id,omitempty"`
	CurrencyCode    string            `json:"currency_code"`
	CustodianUserID *uuid.UUID        `json:"custodian_user_id,omitempty"`
	MaxLimit        decimal.Decimal   `json:"max_limit"`
	CurrentBalance  decimal.Decimal   `json:"current_balance"`
	Status          CashAccountStatus `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// BankAccountBalance contains the bank-account fields needed to settle a
// treasury transfer without coupling the treasury service to the banking module.
type BankAccountBalance struct {
	ID               uuid.UUID
	LegalEntityID    uuid.UUID
	CurrencyCode     string
	CurrentBalance   decimal.Decimal
	AvailableBalance decimal.Decimal
	OverdraftLimit   decimal.Decimal
	Status           string
}

type Receipt struct {
	ID                uuid.UUID       `json:"id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	ReceiptNumber     string          `json:"receipt_number"`
	CashAccountID     uuid.UUID       `json:"cash_account_id"`
	ReceiptDate       time.Time       `json:"receipt_date"`
	ReceiptType       ReceiptType     `json:"receipt_type"`
	ReceivedFrom      string          `json:"received_from"`
	Amount            decimal.Decimal `json:"amount"`
	CurrencyCode      string          `json:"currency_code"`
	PaymentMethod     PaymentMethod   `json:"payment_method"`
	ChequeNumber      *string         `json:"cheque_number,omitempty"`
	ChequeDate        *time.Time      `json:"cheque_date,omitempty"`
	ReferenceNumber   *string         `json:"reference_number,omitempty"`
	Description       *string         `json:"description,omitempty"`
	Status            ReceiptStatus   `json:"status"`
	GLOffsetAccountID *uuid.UUID      `json:"gl_offset_account_id,omitempty"`
	JournalEntryID    *uuid.UUID      `json:"journal_entry_id,omitempty"`
	CreatedByUserID   *uuid.UUID      `json:"created_by_user_id,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type Payment struct {
	ID                uuid.UUID       `json:"id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	PaymentNumber     string          `json:"payment_number"`
	CashAccountID     uuid.UUID       `json:"cash_account_id"`
	PaymentDate       time.Time       `json:"payment_date"`
	PaymentType       PaymentType     `json:"payment_type"`
	PaidTo            string          `json:"paid_to"`
	Amount            decimal.Decimal `json:"amount"`
	CurrencyCode      string          `json:"currency_code"`
	ReferenceNumber   *string         `json:"reference_number,omitempty"`
	Description       *string         `json:"description,omitempty"`
	Status            PaymentStatus   `json:"status"`
	GLOffsetAccountID *uuid.UUID      `json:"gl_offset_account_id,omitempty"`
	JournalEntryID    *uuid.UUID      `json:"journal_entry_id,omitempty"`
	CreatedByUserID   *uuid.UUID      `json:"created_by_user_id,omitempty"`
	ApprovedByUserID  *uuid.UUID      `json:"approved_by_user_id,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type Transfer struct {
	ID                       uuid.UUID               `json:"id"`
	LegalEntityID            uuid.UUID               `json:"legal_entity_id"`
	TransferNumber           string                  `json:"transfer_number"`
	TransferDate             time.Time               `json:"transfer_date"`
	SourceType               TransferSourceType      `json:"source_type"`
	SourceCashAccountID      *uuid.UUID              `json:"source_cash_account_id,omitempty"`
	SourceBankAccountID      *uuid.UUID              `json:"source_bank_account_id,omitempty"`
	DestinationType          TransferDestinationType `json:"destination_type"`
	DestinationCashAccountID *uuid.UUID              `json:"destination_cash_account_id,omitempty"`
	DestinationBankAccountID *uuid.UUID              `json:"destination_bank_account_id,omitempty"`
	Amount                   decimal.Decimal         `json:"amount"`
	CurrencyCode             string                  `json:"currency_code"`
	FeeAmount                decimal.Decimal         `json:"fee_amount"`
	FeeGLAccountID           *uuid.UUID              `json:"fee_gl_account_id,omitempty"`
	Status                   TransferStatus          `json:"status"`
	ReferenceNumber          *string                 `json:"reference_number,omitempty"`
	Notes                    *string                 `json:"notes,omitempty"`
	JournalEntryID           *uuid.UUID              `json:"journal_entry_id,omitempty"`
	InitiatedByUserID        *uuid.UUID              `json:"initiated_by_user_id,omitempty"`
	CompletedByUserID        *uuid.UUID              `json:"completed_by_user_id,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`
}
