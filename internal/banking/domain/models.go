package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrBankAccountNotFound    = errors.New("bank account not found")
	ErrStatementNotFound      = errors.New("bank statement not found")
	ErrStatementLineNotFound  = errors.New("statement line not found")
	ErrReconciliationNotFound = errors.New("bank reconciliation not found")
	ErrReconciliationLocked   = errors.New("bank reconciliation is locked and cannot be modified")
	ErrInvalidIBAN            = errors.New("invalid IBAN format")
	ErrInvalidAmount          = errors.New("amount must be greater than zero")
)

type BankAccountStatus string

const (
	BankAccountActive BankAccountStatus = "ACTIVE"
	BankAccountFrozen BankAccountStatus = "FROZEN"
	BankAccountClosed BankAccountStatus = "CLOSED"
)

type StatementFormat string

const (
	StatementFormatMT940   StatementFormat = "MT940"
	StatementFormatCAMT053 StatementFormat = "CAMT053"
	StatementFormatCSV     StatementFormat = "CSV"
	StatementFormatManual  StatementFormat = "MANUAL"
)

type StatementStatus string

const (
	StatementDraft               StatementStatus = "DRAFT"
	StatementImported            StatementStatus = "IMPORTED"
	StatementPartiallyReconciled StatementStatus = "PARTIALLY_RECONCILED"
	StatementReconciled          StatementStatus = "RECONCILED"
)

type TransactionType string

const (
	TransactionCredit TransactionType = "CREDIT"
	TransactionDebit  TransactionType = "DEBIT"
)

type ReconciledStatus string

const (
	ReconciledUnmatched        ReconciledStatus = "UNMATCHED"
	ReconciledPartiallyMatched ReconciledStatus = "PARTIALLY_MATCHED"
	ReconciledMatched          ReconciledStatus = "MATCHED"
)

type ReconciliationStatus string

const (
	ReconciliationDraft      ReconciliationStatus = "DRAFT"
	ReconciliationInProgress ReconciliationStatus = "IN_PROGRESS"
	ReconciliationCompleted  ReconciliationStatus = "COMPLETED"
	ReconciliationLocked     ReconciliationStatus = "LOCKED"
)

type MatchType string

const (
	MatchExact      MatchType = "EXACT"
	MatchAutoRule   MatchType = "AUTO_RULE"
	MatchManual     MatchType = "MANUAL"
	MatchAdjustment MatchType = "ADJUSTMENT"
)

type MatchedEntityType string

const (
	MatchedEntityJournalEntry      MatchedEntityType = "JOURNAL_ENTRY"
	MatchedEntityReceipt           MatchedEntityType = "RECEIPT"
	MatchedEntityPayment           MatchedEntityType = "PAYMENT"
	MatchedEntityTransfer          MatchedEntityType = "TRANSFER"
	MatchedEntityBankFeeAdjustment MatchedEntityType = "BANK_FEE_ADJUSTMENT"
)

type BankAccount struct {
	ID                  uuid.UUID         `json:"id"`
	LegalEntityID       uuid.UUID         `json:"legal_entity_id"`
	AccountCode         string            `json:"account_code"`
	BankNameEn          string            `json:"bank_name_en"`
	BankNameAr          string            `json:"bank_name_ar"`
	BranchName          *string           `json:"branch_name,omitempty"`
	AccountNumber       string            `json:"account_number"`
	AccountNumberMasked string            `json:"account_number_masked"`
	IBAN                string            `json:"iban"`
	SwiftBIC            string            `json:"swift_bic"`
	CurrencyCode        string            `json:"currency_code"`
	GLAccountID         *uuid.UUID        `json:"gl_account_id,omitempty"`
	OpeningBalance      decimal.Decimal   `json:"opening_balance"`
	CurrentBalance      decimal.Decimal   `json:"current_balance"`
	AvailableBalance    decimal.Decimal   `json:"available_balance"`
	OverdraftLimit      decimal.Decimal   `json:"overdraft_limit"`
	Status              BankAccountStatus `json:"status"`
	IsPrimary           bool              `json:"is_primary"`
	ContactPerson       *string           `json:"contact_person,omitempty"`
	ContactPhone        *string           `json:"contact_phone,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type Statement struct {
	ID               uuid.UUID       `json:"id"`
	LegalEntityID    uuid.UUID       `json:"legal_entity_id"`
	BankAccountID    uuid.UUID       `json:"bank_account_id"`
	StatementNumber  string          `json:"statement_number"`
	StatementDate    time.Time       `json:"statement_date"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	OpeningBalance   decimal.Decimal `json:"opening_balance"`
	ClosingBalance   decimal.Decimal `json:"closing_balance"`
	TotalCredits     decimal.Decimal `json:"total_credits"`
	TotalDebits      decimal.Decimal `json:"total_debits"`
	CurrencyCode     string          `json:"currency_code"`
	FileFormat       StatementFormat `json:"file_format"`
	Status           StatementStatus `json:"status"`
	SourceFilename   *string         `json:"source_filename,omitempty"`
	ImportedByUserID *uuid.UUID      `json:"imported_by_user_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type StatementLine struct {
	ID               uuid.UUID        `json:"id"`
	StatementID      uuid.UUID        `json:"statement_id"`
	LineNumber       int32            `json:"line_number"`
	TransactionDate  time.Time        `json:"transaction_date"`
	ValueDate        *time.Time       `json:"value_date,omitempty"`
	Description      string           `json:"description"`
	ReferenceNumber  *string          `json:"reference_number,omitempty"`
	TransactionType  TransactionType  `json:"transaction_type"`
	Amount           decimal.Decimal  `json:"amount"`
	CurrencyCode     string           `json:"currency_code"`
	BalanceAfter     *decimal.Decimal `json:"balance_after,omitempty"`
	ReconciledStatus ReconciledStatus `json:"reconciled_status"`
	CreatedAt        time.Time        `json:"created_at"`
}

type Reconciliation struct {
	ID                     uuid.UUID            `json:"id"`
	LegalEntityID          uuid.UUID            `json:"legal_entity_id"`
	BankAccountID          uuid.UUID            `json:"bank_account_id"`
	StatementID            *uuid.UUID           `json:"statement_id,omitempty"`
	ReconciliationNumber   string               `json:"reconciliation_number"`
	AsOfDate               time.Time            `json:"as_of_date"`
	StatementEndingBalance decimal.Decimal      `json:"statement_ending_balance"`
	GLBookBalance          decimal.Decimal      `json:"gl_book_balance"`
	UnreconciledDifference decimal.Decimal      `json:"unreconciled_difference"`
	Status                 ReconciliationStatus `json:"status"`
	Notes                  *string              `json:"notes,omitempty"`
	ReconciledByUserID     *uuid.UUID           `json:"reconciled_by_user_id,omitempty"`
	ReconciledAt           *time.Time           `json:"reconciled_at,omitempty"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
}

type ReconciliationMatch struct {
	ID                uuid.UUID         `json:"id"`
	ReconciliationID  uuid.UUID         `json:"reconciliation_id"`
	StatementLineID   uuid.UUID         `json:"statement_line_id"`
	MatchType         MatchType         `json:"match_type"`
	MatchedEntityType MatchedEntityType `json:"matched_entity_type"`
	MatchedEntityID   *uuid.UUID        `json:"matched_entity_id,omitempty"`
	MatchedAmount     decimal.Decimal   `json:"matched_amount"`
	DifferenceAmount  decimal.Decimal   `json:"difference_amount"`
	Notes             *string           `json:"notes,omitempty"`
	MatchedByUserID   *uuid.UUID        `json:"matched_by_user_id,omitempty"`
	MatchedAt         time.Time         `json:"matched_at"`
}

// CandidateLedgerEntry represents an internal ERP transaction to be matched against bank statement lines
type CandidateLedgerEntry struct {
	EntityType      MatchedEntityType `json:"entity_type"`
	EntityID        uuid.UUID         `json:"entity_id"`
	ReferenceNumber string            `json:"reference_number"`
	Date            time.Time         `json:"date"`
	Amount          decimal.Decimal   `json:"amount"`
	Description     string            `json:"description"`
}

// AutoMatchResult holds suggested or executed matches
type AutoMatchResult struct {
	StatementLineID uuid.UUID            `json:"statement_line_id"`
	Candidate       CandidateLedgerEntry `json:"candidate"`
	MatchType       MatchType            `json:"match_type"`
	Confidence      float64              `json:"confidence"` // 1.0 = exact match
}

// MaskAccountNumber returns a masked representation e.g. **** **** **** 1234
func MaskAccountNumber(accNum string) string {
	cleaned := strings.ReplaceAll(accNum, " ", "")
	if len(cleaned) <= 4 {
		return "**** " + cleaned
	}
	last4 := cleaned[len(cleaned)-4:]
	return "**** **** **** " + last4
}
