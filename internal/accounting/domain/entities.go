package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"
	AccountTypeLiability AccountType = "LIABILITY"
	AccountTypeEquity    AccountType = "EQUITY"
	AccountTypeRevenue   AccountType = "REVENUE"
	AccountTypeExpense   AccountType = "EXPENSE"
)

type NormalBalance string

const (
	NormalBalanceDebit  NormalBalance = "DEBIT"
	NormalBalanceCredit NormalBalance = "CREDIT"
)

type FiscalPeriodStatus string

const (
	FiscalPeriodStatusOpen       FiscalPeriodStatus = "OPEN"
	FiscalPeriodStatusSoftClosed FiscalPeriodStatus = "SOFT_CLOSED"
	FiscalPeriodStatusClosed     FiscalPeriodStatus = "CLOSED"
	FiscalPeriodStatusLocked     FiscalPeriodStatus = "LOCKED"
)

type JournalType string

const (
	JournalTypeGeneral    JournalType = "GENERAL"
	JournalTypeSales      JournalType = "SALES"
	JournalTypePurchase   JournalType = "PURCHASE"
	JournalTypeCash       JournalType = "CASH"
	JournalTypeBank       JournalType = "BANK"
	JournalTypePayroll    JournalType = "PAYROLL"
	JournalTypeInventory  JournalType = "INVENTORY"
	JournalTypeAdjustment JournalType = "ADJUSTMENT"
)

type JournalEntryStatus string

const (
	JournalEntryStatusDraft     JournalEntryStatus = "DRAFT"
	JournalEntryStatusSubmitted JournalEntryStatus = "SUBMITTED"
	JournalEntryStatusApproved  JournalEntryStatus = "APPROVED"
	JournalEntryStatusPosted    JournalEntryStatus = "POSTED"
	JournalEntryStatusReversed  JournalEntryStatus = "REVERSED"
	JournalEntryStatusCancelled JournalEntryStatus = "CANCELLED"
)

type ChartOfAccounts struct {
	ID            uuid.UUID `json:"id"`
	LegalEntityID uuid.UUID `json:"legal_entity_id"`
	Code          string    `json:"code"`
	NameEn        string    `json:"name_en"`
	NameAr        string    `json:"name_ar"`
	Description   string    `json:"description,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int64     `json:"version"`
}

type Account struct {
	ID               uuid.UUID     `json:"id"`
	LegalEntityID    uuid.UUID     `json:"legal_entity_id"`
	ChartID          uuid.UUID     `json:"chart_id"`
	ParentID         *uuid.UUID    `json:"parent_id,omitempty"`
	Code             string        `json:"code"`
	NameEn           string        `json:"name_en"`
	NameAr           string        `json:"name_ar"`
	AccountType      AccountType   `json:"account_type"`
	NormalBalance    NormalBalance `json:"normal_balance"`
	AllowPosting     bool          `json:"allow_posting"`
	IsControlAccount bool          `json:"is_control_account"`
	IsActive         bool          `json:"is_active"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	Version          int64         `json:"version"`
}

type FiscalYear struct {
	ID            uuid.UUID `json:"id"`
	LegalEntityID uuid.UUID `json:"legal_entity_id"`
	Code          string    `json:"code"`
	NameEn        string    `json:"name_en"`
	NameAr        string    `json:"name_ar"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int64     `json:"version"`
}

type FiscalPeriod struct {
	ID            uuid.UUID          `json:"id"`
	FiscalYearID  uuid.UUID          `json:"fiscal_year_id"`
	LegalEntityID uuid.UUID          `json:"legal_entity_id"`
	PeriodNumber  int16              `json:"period_number"`
	PeriodNameEn  string             `json:"period_name_en"`
	PeriodNameAr  string             `json:"period_name_ar"`
	Quarter       int16              `json:"quarter"`
	StartDate     time.Time          `json:"start_date"`
	EndDate       time.Time          `json:"end_date"`
	Status        FiscalPeriodStatus `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Version       int64              `json:"version"`
}

type Journal struct {
	ID            uuid.UUID   `json:"id"`
	LegalEntityID uuid.UUID   `json:"legal_entity_id"`
	Code          string      `json:"code"`
	NameEn        string      `json:"name_en"`
	NameAr        string      `json:"name_ar"`
	JournalType   JournalType `json:"journal_type"`
	IsActive      bool        `json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Version       int64       `json:"version"`
}

type JournalEntry struct {
	ID             uuid.UUID          `json:"id"`
	LegalEntityID  uuid.UUID          `json:"legal_entity_id"`
	JournalID      uuid.UUID          `json:"journal_id"`
	FiscalPeriodID uuid.UUID          `json:"fiscal_period_id"`
	EntryNumber    string             `json:"entry_number"`
	AccountingDate time.Time          `json:"accounting_date"`
	Status         JournalEntryStatus `json:"status"`
	SourceModule   string             `json:"source_module"`
	SourceType     *string            `json:"source_type,omitempty"`
	SourceID       *uuid.UUID         `json:"source_id,omitempty"`
	Reference      *string            `json:"reference,omitempty"`
	Description    string             `json:"description"`
	CurrencyCode   string             `json:"currency_code"`
	ExchangeRate   decimal.Decimal    `json:"exchange_rate"`
	TotalDebit     decimal.Decimal    `json:"total_debit"`
	TotalCredit    decimal.Decimal    `json:"total_credit"`
	CreatedBy      uuid.UUID          `json:"created_by"`
	ApprovedBy     *uuid.UUID         `json:"approved_by,omitempty"`
	PostedBy       *uuid.UUID         `json:"posted_by,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	ApprovedAt     *time.Time         `json:"approved_at,omitempty"`
	PostedAt       *time.Time         `json:"posted_at,omitempty"`
	ReversedAt     *time.Time         `json:"reversed_at,omitempty"`
	ReversalOfID   *uuid.UUID         `json:"reversal_of_id,omitempty"`
	Version        int64              `json:"version"`
	Lines          []JournalLine      `json:"lines,omitempty"`
}

type JournalLine struct {
	ID                uuid.UUID       `json:"id"`
	JournalEntryID    uuid.UUID       `json:"journal_entry_id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	LineNumber        int16           `json:"line_number"`
	AccountID         uuid.UUID       `json:"account_id"`
	AccountCode       string          `json:"account_code,omitempty"`
	AccountNameEn     string          `json:"account_name_en,omitempty"`
	AccountNameAr     string          `json:"account_name_ar,omitempty"`
	Description       *string         `json:"description,omitempty"`
	CurrencyCode      string          `json:"currency_code"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	TransactionDebit  decimal.Decimal `json:"transaction_debit"`
	TransactionCredit decimal.Decimal `json:"transaction_credit"`
	BaseDebit         decimal.Decimal `json:"base_debit"`
	BaseCredit        decimal.Decimal `json:"base_credit"`
	BranchID          *uuid.UUID      `json:"branch_id,omitempty"`
	DepartmentID      *uuid.UUID      `json:"department_id,omitempty"`
	CostCenterID      *uuid.UUID      `json:"cost_center_id,omitempty"`
}

type TrialBalanceRow struct {
	AccountID     uuid.UUID       `json:"account_id"`
	AccountCode   string          `json:"account_code"`
	AccountNameEn string          `json:"account_name_en"`
	AccountNameAr string          `json:"account_name_ar"`
	AccountType   AccountType     `json:"account_type"`
	NormalBalance NormalBalance   `json:"normal_balance"`
	TotalDebit    decimal.Decimal `json:"total_debit"`
	TotalCredit   decimal.Decimal `json:"total_credit"`
	NetBalance    decimal.Decimal `json:"net_balance"`
}

type TrialBalanceReport struct {
	LegalEntityID  uuid.UUID         `json:"legal_entity_id"`
	AsOfDate       time.Time         `json:"as_of_date"`
	Rows           []TrialBalanceRow `json:"rows"`
	TotalDebitSum  decimal.Decimal   `json:"total_debit_sum"`
	TotalCreditSum decimal.Decimal   `json:"total_credit_sum"`
	IsBalanced     bool              `json:"is_balanced"`
}
