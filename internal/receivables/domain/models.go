package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrCustomerNotFound     = errors.New("customer not found")
	ErrDuplicateCustomerNum = errors.New("customer number already exists")
	ErrInvalidCustomerType  = errors.New("invalid customer type, must be B2B, B2C, or GOVERNMENT")
	ErrInvalidTaxID         = errors.New("invalid Saudi VAT number, must be 15 digits starting and ending with 3")
	ErrInvalidCRNumber      = errors.New("invalid Saudi commercial registration number, must be 10 digits")
	ErrInvalidNationalID    = errors.New("invalid Saudi national ID, must be 10 digits")
	ErrCityRequired         = errors.New("city is required in customer address")
	ErrCreditLimitExceeded  = errors.New("customer credit limit exceeded")
	ErrCustomerCreditHold   = errors.New("customer account is on credit hold")
	ErrPaymentTermNotFound  = errors.New("payment term not found")
	ErrReceiptNotFound      = errors.New("receipt not found")
	ErrReceiptAlreadyPosted = errors.New("receipt is already posted")
	ErrReceiptAlreadyCancel = errors.New("receipt is already cancelled")
	ErrOpenItemNotFound     = errors.New("open item not found")
	ErrOpenItemAlreadyPaid  = errors.New("open item is already paid")
	ErrOverAllocation       = errors.New("allocated amount exceeds open item or receipt balance")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
)

var (
	vatRegex = regexp.MustCompile(`^3[0-9]{13}3$`)
	crRegex  = regexp.MustCompile(`^[0-9]{10}$`)
	nidRegex = regexp.MustCompile(`^[0-9]{10}$`)
)

type CustomerType string

const (
	CustomerTypeB2B        CustomerType = "B2B"
	CustomerTypeB2C        CustomerType = "B2C"
	CustomerTypeGovernment CustomerType = "GOVERNMENT"
)

type CustomerStatus string

const (
	CustomerStatusActive    CustomerStatus = "ACTIVE"
	CustomerStatusSuspended CustomerStatus = "SUSPENDED"
	CustomerStatusBlocked   CustomerStatus = "BLOCKED"
)

type BlockAction string

const (
	BlockActionWarn            BlockAction = "WARN"
	BlockActionBlock           BlockAction = "BLOCK"
	BlockActionRequireApproval BlockAction = "REQUIRE_APPROVAL"
)

type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodCash         PaymentMethod = "CASH"
	PaymentMethodCheque       PaymentMethod = "CHEQUE"
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodMada         PaymentMethod = "MADA"
	PaymentMethodSadad        PaymentMethod = "SADAD"
)

type ReceiptStatus string

const (
	ReceiptStatusDraft     ReceiptStatus = "DRAFT"
	ReceiptStatusPosted    ReceiptStatus = "POSTED"
	ReceiptStatusCancelled ReceiptStatus = "CANCELLED"
)

type DocumentType string

const (
	DocumentTypeInvoice            DocumentType = "INVOICE"
	DocumentTypeDebitNote          DocumentType = "DEBIT_NOTE"
	DocumentTypeCreditNote         DocumentType = "CREDIT_NOTE"
	DocumentTypeUnallocatedReceipt DocumentType = "UNALLOCATED_RECEIPT"
)

type OpenItemStatus string

const (
	OpenItemStatusOpen          OpenItemStatus = "OPEN"
	OpenItemStatusPartiallyPaid OpenItemStatus = "PARTIALLY_PAID"
	OpenItemStatusPaid          OpenItemStatus = "PAID"
	OpenItemStatusCancelled     OpenItemStatus = "CANCELLED"
)

type SaudiAddress struct {
	BuildingNo   string `json:"building_no,omitempty"`
	StreetName   string `json:"street_name,omitempty"`
	District     string `json:"district,omitempty"`
	City         string `json:"city"`
	PostalCode   string `json:"postal_code,omitempty"`
	AdditionalNo string `json:"additional_no,omitempty"`
	CountryCode  string `json:"country_code"`
}

type Customer struct {
	ID                 uuid.UUID       `json:"id"`
	LegalEntityID      uuid.UUID       `json:"legal_entity_id"`
	CustomerNumber     string          `json:"customer_number"`
	CustomerType       CustomerType    `json:"customer_type"`
	NameEn             string          `json:"name_en"`
	NameAr             string          `json:"name_ar"`
	TaxID              *string         `json:"tax_id,omitempty"`
	CRNumber           *string         `json:"cr_number,omitempty"`
	NationalID         *string         `json:"national_id,omitempty"`
	Address            SaudiAddress    `json:"address"`
	Email              *string         `json:"email,omitempty"`
	Phone              *string         `json:"phone,omitempty"`
	ContactPerson      *string         `json:"contact_person,omitempty"`
	PaymentTermsID     *uuid.UUID      `json:"payment_terms_id,omitempty"`
	CurrencyCode       string          `json:"currency_code"`
	DefaultAccountID   *uuid.UUID      `json:"default_account_id,omitempty"`
	Status             CustomerStatus  `json:"status"`
	PaymentTermsCode   *string         `json:"payment_terms_code,omitempty"`
	PaymentTermsNameAr *string         `json:"payment_terms_name_ar,omitempty"`
	CreditLimit        decimal.Decimal `json:"credit_limit"`
	CurrentBalance     decimal.Decimal `json:"current_balance"`
	IsCreditHold       bool            `json:"is_credit_hold"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

func (c *Customer) Validate() error {
	if c.CustomerType != CustomerTypeB2B && c.CustomerType != CustomerTypeB2C && c.CustomerType != CustomerTypeGovernment {
		return ErrInvalidCustomerType
	}
	if c.Address.City == "" {
		return ErrCityRequired
	}
	if c.TaxID != nil && *c.TaxID != "" {
		if !vatRegex.MatchString(*c.TaxID) {
			return ErrInvalidTaxID
		}
	}
	if c.CRNumber != nil && *c.CRNumber != "" {
		if !crRegex.MatchString(*c.CRNumber) {
			return ErrInvalidCRNumber
		}
	}
	if c.NationalID != nil && *c.NationalID != "" {
		if !nidRegex.MatchString(*c.NationalID) {
			return ErrInvalidNationalID
		}
	}
	return nil
}

type CustomerAccount struct {
	ID            uuid.UUID `json:"id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	GLAccountID   uuid.UUID `json:"gl_account_id"`
	CurrencyCode  string    `json:"currency_code"`
	IsPrimary     bool      `json:"is_primary"`
	AccountCode   string    `json:"account_code,omitempty"`
	AccountNameAr string    `json:"account_name_ar,omitempty"`
	AccountNameEn string    `json:"account_name_en,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreditLimit struct {
	ID             uuid.UUID       `json:"id"`
	CustomerID     uuid.UUID       `json:"customer_id"`
	CreditLimit    decimal.Decimal `json:"credit_limit"`
	CurrentBalance decimal.Decimal `json:"current_balance"`
	MaxOverdueDays int32           `json:"max_overdue_days"`
	BlockAction    BlockAction     `json:"block_action"`
	IsCreditHold   bool            `json:"is_credit_hold"`
	HoldReason     *string         `json:"hold_reason,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type PaymentTerm struct {
	ID              uuid.UUID       `json:"id"`
	LegalEntityID   uuid.UUID       `json:"legal_entity_id"`
	Code            string          `json:"code"`
	NameEn          string          `json:"name_en"`
	NameAr          string          `json:"name_ar"`
	DueDays         int32           `json:"due_days"`
	DiscountDays    int32           `json:"discount_days"`
	DiscountPercent decimal.Decimal `json:"discount_percent"`
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
}

type Receipt struct {
	ID                   uuid.UUID       `json:"id"`
	LegalEntityID        uuid.UUID       `json:"legal_entity_id"`
	ReceiptNumber        string          `json:"receipt_number"`
	CustomerID           uuid.UUID       `json:"customer_id"`
	ReceiptDate          time.Time       `json:"receipt_date"`
	PaymentMethod        PaymentMethod   `json:"payment_method"`
	DepositAccountID     uuid.UUID       `json:"deposit_account_id"`
	Amount               decimal.Decimal `json:"amount"`
	CurrencyCode         string          `json:"currency_code"`
	ReferenceNumber      *string         `json:"reference_number,omitempty"`
	Status               ReceiptStatus   `json:"status"`
	Notes                *string         `json:"notes,omitempty"`
	AllocatedAmount      decimal.Decimal `json:"allocated_amount"`
	UnallocatedAmount    decimal.Decimal `json:"unallocated_amount"`
	JournalEntryID       *uuid.UUID      `json:"journal_entry_id,omitempty"`
	CreatedByUserID      *uuid.UUID      `json:"created_by_user_id,omitempty"`
	CustomerNumber       string          `json:"customer_number,omitempty"`
	CustomerNameAr       string          `json:"customer_name_ar,omitempty"`
	CustomerNameEn       string          `json:"customer_name_en,omitempty"`
	DepositAccountCode   string          `json:"deposit_account_code,omitempty"`
	DepositAccountNameAr string          `json:"deposit_account_name_ar,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
}

type ReceiptAllocation struct {
	ID              uuid.UUID       `json:"id"`
	ReceiptID       uuid.UUID       `json:"receipt_id"`
	OpenItemID      uuid.UUID       `json:"open_item_id"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
	AllocatedAt     time.Time       `json:"allocated_at"`
	CreatedByUserID *uuid.UUID      `json:"created_by_user_id,omitempty"`
	DocumentType    DocumentType    `json:"document_type,omitempty"`
	DocumentNumber  string          `json:"document_number,omitempty"`
	DueDate         time.Time       `json:"due_date,omitempty"`
	OriginalAmount  decimal.Decimal `json:"original_amount,omitempty"`
	RemainingAmount decimal.Decimal `json:"remaining_amount,omitempty"`
}

type OpenItem struct {
	ID              uuid.UUID       `json:"id"`
	LegalEntityID   uuid.UUID       `json:"legal_entity_id"`
	CustomerID      uuid.UUID       `json:"customer_id"`
	DocumentType    DocumentType    `json:"document_type"`
	DocumentNumber  string          `json:"document_number"`
	DocumentID      *uuid.UUID      `json:"document_id,omitempty"`
	DocumentDate    time.Time       `json:"document_date"`
	DueDate         time.Time       `json:"due_date"`
	CurrencyCode    string          `json:"currency_code"`
	OriginalAmount  decimal.Decimal `json:"original_amount"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
	RemainingAmount decimal.Decimal `json:"remaining_amount"`
	Status          OpenItemStatus  `json:"status"`
	CustomerNumber  string          `json:"customer_number,omitempty"`
	CustomerNameAr  string          `json:"customer_name_ar,omitempty"`
	DaysOverdue     int32           `json:"days_overdue,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CustomerAgingSummary struct {
	TotalOutstanding decimal.Decimal `json:"total_outstanding"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	Overdue1to30     decimal.Decimal `json:"overdue_1_30"`
	Overdue31to60    decimal.Decimal `json:"overdue_31_60"`
	Overdue61to90    decimal.Decimal `json:"overdue_61_90"`
	OverdueOver90    decimal.Decimal `json:"overdue_over_90"`
}
