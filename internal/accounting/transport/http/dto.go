package http

import (
	"time"

	"erp/internal/accounting/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateAccountRequest struct {
	ChartID          uuid.UUID            `json:"chart_id"`
	ParentID         *uuid.UUID           `json:"parent_id,omitempty"`
	Code             string               `json:"code"`
	NameEn           string               `json:"name_en"`
	NameAr           string               `json:"name_ar"`
	AccountType      domain.AccountType   `json:"account_type"`
	NormalBalance    domain.NormalBalance `json:"normal_balance"`
	AllowPosting     bool                 `json:"allow_posting"`
	IsControlAccount bool                 `json:"is_control_account"`
}

type UpdateAccountRequest struct {
	NameEn           string `json:"name_en"`
	NameAr           string `json:"name_ar"`
	AllowPosting     bool   `json:"allow_posting"`
	IsControlAccount bool   `json:"is_control_account"`
	IsActive         bool   `json:"is_active"`
}

type CreateJournalLineRequest struct {
	AccountID         uuid.UUID       `json:"account_id"`
	Description       *string         `json:"description,omitempty"`
	CurrencyCode      string          `json:"currency_code,omitempty"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate,omitempty"`
	TransactionDebit  decimal.Decimal `json:"transaction_debit"`
	TransactionCredit decimal.Decimal `json:"transaction_credit"`
	BranchID          *uuid.UUID      `json:"branch_id,omitempty"`
	DepartmentID      *uuid.UUID      `json:"department_id,omitempty"`
	CostCenterID      *uuid.UUID      `json:"cost_center_id,omitempty"`
}

type CreateJournalEntryRequest struct {
	JournalID      uuid.UUID                  `json:"journal_id"`
	FiscalPeriodID *uuid.UUID                 `json:"fiscal_period_id,omitempty"`
	EntryNumber    string                     `json:"entry_number,omitempty"`
	AccountingDate *time.Time                 `json:"accounting_date,omitempty"`
	Reference      *string                    `json:"reference,omitempty"`
	Description    string                     `json:"description"`
	CurrencyCode   string                     `json:"currency_code,omitempty"`
	ExchangeRate   decimal.Decimal            `json:"exchange_rate,omitempty"`
	Lines          []CreateJournalLineRequest `json:"lines"`
}

type ReverseJournalEntryRequest struct {
	Reason string `json:"reason"`
}

type UpdateFiscalPeriodStatusRequest struct {
	Status domain.FiscalPeriodStatus `json:"status"`
}
