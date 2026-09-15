package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PostingRequest represents a request from any module (GL, Sales, Purchases, Payroll, etc.)
// to create and atomically post a balanced journal entry in the General Ledger.
type PostingRequest struct {
	LegalEntityID uuid.UUID

	// Source traceability — links the journal entry back to the originating document.
	SourceModule string    // e.g., "GL", "SALES", "PURCHASES", "PAYROLL", "INVENTORY"
	SourceType   string    // e.g., "INVOICE", "RECEIPT", "ADJUSTMENT"
	SourceID     uuid.UUID // UUID of the originating document

	JournalID uuid.UUID // Target sub-ledger journal

	AccountingDate time.Time // Date of the financial transaction

	CurrencyCode string          // ISO 4217 currency code
	ExchangeRate decimal.Decimal // Exchange rate to base currency (SAR)

	Reference   string // Human-readable reference
	Description string // Arabic/English description of the transaction

	Lines []PostingLine

	// Actor performing the posting (for audit trail)
	ActorID uuid.UUID
}

// PostingLine represents a single debit or credit line within a posting request.
type PostingLine struct {
	AccountID uuid.UUID

	Description string

	TransactionDebit  decimal.Decimal
	TransactionCredit decimal.Decimal

	// Base currency amounts (calculated from transaction * exchange rate if not provided)
	BaseDebit  decimal.Decimal
	BaseCredit decimal.Decimal

	// Optional dimensional analysis
	BranchID     *uuid.UUID
	DepartmentID *uuid.UUID
	CostCenterID *uuid.UUID
}

// PostingResult is returned after a successful post or reversal.
type PostingResult struct {
	JournalEntryID uuid.UUID
	EntryNumber    string
}

// AccountBalance represents a period-level balance accumulation for an account.
type AccountBalance struct {
	ID             uuid.UUID       `json:"id"`
	LegalEntityID  uuid.UUID       `json:"legal_entity_id"`
	AccountID      uuid.UUID       `json:"account_id"`
	FiscalPeriodID uuid.UUID       `json:"fiscal_period_id"`
	OpeningBalance decimal.Decimal `json:"opening_balance"`
	PeriodDebit    decimal.Decimal `json:"period_debit"`
	PeriodCredit   decimal.Decimal `json:"period_credit"`
	ClosingBalance decimal.Decimal `json:"closing_balance"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
