package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrSupplierNotFound     = errors.New("supplier not found")
	ErrDuplicateSupplierNum = errors.New("supplier number already exists")
	ErrInvalidSupplierType  = errors.New("invalid supplier type, must be LOCAL, INTERNATIONAL, or GOVERNMENT")
	ErrInvalidTaxID         = errors.New("invalid Saudi VAT number, must be 15 digits starting and ending with 3")
	ErrInvalidCRNumber      = errors.New("invalid Saudi commercial registration number, must be 10 digits")
	ErrInvalidIBAN          = errors.New("invalid IBAN format")
	ErrCityRequired         = errors.New("city is required in supplier address")
	ErrPaymentTermNotFound  = errors.New("payment term not found")
	ErrInvoiceNotFound      = errors.New("supplier invoice not found")
	ErrInvoiceAlreadyPosted = errors.New("supplier invoice is already posted")
	ErrInvoiceAlreadyPaid   = errors.New("supplier invoice is already fully paid")
	ErrInvalidInvoiceStatus = errors.New("invalid supplier invoice status transition")
	ErrPaymentNotFound      = errors.New("supplier payment not found")
	ErrPaymentAlreadyPosted = errors.New("supplier payment is already posted")
	ErrOpenItemNotFound     = errors.New("open item not found")
	ErrOpenItemAlreadyPaid  = errors.New("open item is already paid")
	ErrOverAllocation       = errors.New("allocated amount exceeds open item or payment balance")
	ErrCreditNoteNotFound   = errors.New("supplier credit note not found")
	ErrSupplierBlocked      = errors.New("supplier is suspended or blocked")
)

var (
	vatRegex  = regexp.MustCompile(`^3[0-9]{13}3$`)
	crRegex   = regexp.MustCompile(`^[0-9]{10}$`)
	ibanRegex = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{4,30}$`)
)

type SupplierType string

const (
	SupplierTypeLocal         SupplierType = "LOCAL"
	SupplierTypeInternational SupplierType = "INTERNATIONAL"
	SupplierTypeGovernment    SupplierType = "GOVERNMENT"
)

type SupplierStatus string

const (
	SupplierStatusActive    SupplierStatus = "ACTIVE"
	SupplierStatusSuspended SupplierStatus = "SUSPENDED"
	SupplierStatusBlocked   SupplierStatus = "BLOCKED"
)

type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodCash         PaymentMethod = "CASH"
	PaymentMethodCheque       PaymentMethod = "CHEQUE"
	PaymentMethodMada         PaymentMethod = "MADA"
	PaymentMethodSadad        PaymentMethod = "SADAD"
)

type SupplierInvoiceStatus string

const (
	InvoiceStatusDraft     SupplierInvoiceStatus = "DRAFT"
	InvoiceStatusApproved  SupplierInvoiceStatus = "APPROVED"
	InvoiceStatusPosted    SupplierInvoiceStatus = "POSTED"
	InvoiceStatusPaid      SupplierInvoiceStatus = "PAID"
	InvoiceStatusCancelled SupplierInvoiceStatus = "CANCELLED"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid        PaymentStatus = "UNPAID"
	PaymentStatusPartiallyPaid PaymentStatus = "PARTIALLY_PAID"
	PaymentStatusPaid          PaymentStatus = "PAID"
)

type OpenItemDocumentType string

const (
	DocumentTypeInvoice            OpenItemDocumentType = "INVOICE"
	DocumentTypeCreditNote         OpenItemDocumentType = "CREDIT_NOTE"
	DocumentTypeUnallocatedPayment OpenItemDocumentType = "UNALLOCATED_PAYMENT"
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

type Supplier struct {
	ID                          uuid.UUID      `json:"id"`
	LegalEntityID               uuid.UUID      `json:"legal_entity_id"`
	SupplierNumber              string         `json:"supplier_number"`
	SupplierType                SupplierType   `json:"supplier_type"`
	NameEn                      string         `json:"name_en"`
	NameAr                      string         `json:"name_ar"`
	TaxID                       *string        `json:"tax_id,omitempty"`
	CRNumber                    *string        `json:"cr_number,omitempty"`
	NationalID                  *string        `json:"national_id,omitempty"`
	IBAN                        *string        `json:"iban,omitempty"`
	BankName                    *string        `json:"bank_name,omitempty"`
	SwiftCode                   *string        `json:"swift_code,omitempty"`
	Address                     SaudiAddress   `json:"address"`
	Email                       *string        `json:"email,omitempty"`
	Phone                       *string        `json:"phone,omitempty"`
	ContactPerson               *string        `json:"contact_person,omitempty"`
	Website                     *string        `json:"website,omitempty"`
	PaymentTermsID              *uuid.UUID     `json:"payment_terms_id,omitempty"`
	PaymentTermsCode            string         `json:"payment_terms_code,omitempty"`
	PaymentTermsNameAr          string         `json:"payment_terms_name_ar,omitempty"`
	PaymentTermsDueDays         int32          `json:"payment_terms_due_days,omitempty"`
	CurrencyCode                string         `json:"currency_code"`
	DefaultExpenseAccountID     *uuid.UUID     `json:"default_expense_account_id,omitempty"`
	DefaultLiabilityAccountID   *uuid.UUID     `json:"default_liability_account_id,omitempty"`
	DefaultExpenseAccountCode   string         `json:"default_expense_account_code,omitempty"`
	DefaultLiabilityAccountCode string         `json:"default_liability_account_code,omitempty"`
	Status                      SupplierStatus `json:"status"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
}

func (s *Supplier) Validate() error {
	if s.SupplierNumber == "" {
		return errors.New("supplier number is required")
	}
	if s.NameEn == "" || s.NameAr == "" {
		return errors.New("both Arabic and English supplier names are required")
	}
	switch s.SupplierType {
	case SupplierTypeLocal, SupplierTypeInternational, SupplierTypeGovernment:
	default:
		return ErrInvalidSupplierType
	}
	if s.Address.City == "" {
		return ErrCityRequired
	}
	if s.TaxID != nil && *s.TaxID != "" {
		if !vatRegex.MatchString(*s.TaxID) {
			return ErrInvalidTaxID
		}
	}
	if s.CRNumber != nil && *s.CRNumber != "" {
		if !crRegex.MatchString(*s.CRNumber) {
			return ErrInvalidCRNumber
		}
	}
	if s.IBAN != nil && *s.IBAN != "" {
		clean := strings.ToUpper(strings.ReplaceAll(*s.IBAN, " ", ""))
		if !ibanRegex.MatchString(clean) {
			return ErrInvalidIBAN
		}
	}
	return nil
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

type SupplierInvoice struct {
	ID                     uuid.UUID             `json:"id"`
	LegalEntityID          uuid.UUID             `json:"legal_entity_id"`
	InvoiceNumber          string                `json:"invoice_number"`
	SupplierInvoiceNumber  string                `json:"supplier_invoice_number"`
	SupplierID             uuid.UUID             `json:"supplier_id"`
	SupplierNameEn         string                `json:"supplier_name_en,omitempty"`
	SupplierNameAr         string                `json:"supplier_name_ar,omitempty"`
	SupplierTaxID          *string               `json:"supplier_tax_id,omitempty"`
	SupplierCRNumber       *string               `json:"supplier_cr_number,omitempty"`
	SupplierIBAN           *string               `json:"supplier_iban,omitempty"`
	SupplierLiabilityAccID *uuid.UUID            `json:"supplier_liability_account_id,omitempty"`
	InvoiceDate            time.Time             `json:"invoice_date"`
	DueDate                time.Time             `json:"due_date"`
	CurrencyCode           string                `json:"currency_code"`
	ExchangeRate           decimal.Decimal       `json:"exchange_rate"`
	Subtotal               decimal.Decimal       `json:"subtotal"`
	DiscountAmount         decimal.Decimal       `json:"discount_amount"`
	TaxableAmount          decimal.Decimal       `json:"taxable_amount"`
	TaxAmount              decimal.Decimal       `json:"tax_amount"`
	TotalAmount            decimal.Decimal       `json:"total_amount"`
	Status                 SupplierInvoiceStatus `json:"status"`
	PaymentStatus          PaymentStatus         `json:"payment_status"`
	JournalEntryID         *uuid.UUID            `json:"journal_entry_id,omitempty"`
	Notes                  *string               `json:"notes,omitempty"`
	CreatedByUserID        *uuid.UUID            `json:"created_by_user_id,omitempty"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
	Lines                  []SupplierInvoiceLine `json:"lines,omitempty"`
}

type SupplierInvoiceLine struct {
	TaxCodeID            *uuid.UUID      `json:"tax_code_id,omitempty"`
	ID                   uuid.UUID       `json:"id"`
	InvoiceID            uuid.UUID       `json:"invoice_id"`
	LineNumber           int32           `json:"line_number"`
	ItemID               *uuid.UUID      `json:"item_id,omitempty"`
	ItemCode             string          `json:"item_code,omitempty"`
	ItemNameEn           string          `json:"item_name_en,omitempty"`
	ItemNameAr           string          `json:"item_name_ar,omitempty"`
	ExpenseAccountID     uuid.UUID       `json:"expense_account_id"`
	ExpenseAccountNumber string          `json:"expense_account_number,omitempty"`
	ExpenseAccountNameAr string          `json:"expense_account_name_ar,omitempty"`
	Description          string          `json:"description"`
	Quantity             decimal.Decimal `json:"quantity"`
	UnitPrice            decimal.Decimal `json:"unit_price"`
	DiscountPercent      decimal.Decimal `json:"discount_percent"`
	DiscountAmount       decimal.Decimal `json:"discount_amount"`
	TaxRate              decimal.Decimal `json:"tax_rate"`
	TaxAmount            decimal.Decimal `json:"tax_amount"`
	Subtotal             decimal.Decimal `json:"subtotal"`
	TotalAmount          decimal.Decimal `json:"total_amount"`
	CreatedAt            time.Time       `json:"created_at"`
}

type SupplierCreditNote struct {
	ID                       uuid.UUID                `json:"id"`
	LegalEntityID            uuid.UUID                `json:"legal_entity_id"`
	CreditNoteNumber         string                   `json:"credit_note_number"`
	SupplierCreditNoteNumber *string                  `json:"supplier_credit_note_number,omitempty"`
	SupplierID               uuid.UUID                `json:"supplier_id"`
	SupplierNameEn           string                   `json:"supplier_name_en,omitempty"`
	SupplierNameAr           string                   `json:"supplier_name_ar,omitempty"`
	OriginalInvoiceID        *uuid.UUID               `json:"original_invoice_id,omitempty"`
	CreditNoteDate           time.Time                `json:"credit_note_date"`
	CurrencyCode             string                   `json:"currency_code"`
	Subtotal                 decimal.Decimal          `json:"subtotal"`
	TaxAmount                decimal.Decimal          `json:"tax_amount"`
	TotalAmount              decimal.Decimal          `json:"total_amount"`
	Reason                   *string                  `json:"reason,omitempty"`
	Status                   string                   `json:"status"`
	JournalEntryID           *uuid.UUID               `json:"journal_entry_id,omitempty"`
	CreatedByUserID          *uuid.UUID               `json:"created_by_user_id,omitempty"`
	CreatedAt                time.Time                `json:"created_at"`
	Lines                    []SupplierCreditNoteLine `json:"lines,omitempty"`
}

type SupplierCreditNoteLine struct {
	ID               uuid.UUID       `json:"id"`
	CreditNoteID     uuid.UUID       `json:"credit_note_id"`
	LineNumber       int32           `json:"line_number"`
	ItemID           *uuid.UUID      `json:"item_id,omitempty"`
	ExpenseAccountID uuid.UUID       `json:"expense_account_id"`
	Description      string          `json:"description"`
	Quantity         decimal.Decimal `json:"quantity"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	TaxRate          decimal.Decimal `json:"tax_rate"`
	TaxAmount        decimal.Decimal `json:"tax_amount"`
	Subtotal         decimal.Decimal `json:"subtotal"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	CreatedAt        time.Time       `json:"created_at"`
}

type SupplierPayment struct {
	ID                        uuid.UUID           `json:"id"`
	LegalEntityID             uuid.UUID           `json:"legal_entity_id"`
	PaymentNumber             string              `json:"payment_number"`
	SupplierID                uuid.UUID           `json:"supplier_id"`
	SupplierNameEn            string              `json:"supplier_name_en,omitempty"`
	SupplierNameAr            string              `json:"supplier_name_ar,omitempty"`
	SupplierNumber            string              `json:"supplier_number,omitempty"`
	PaymentDate               time.Time           `json:"payment_date"`
	PaymentMethod             PaymentMethod       `json:"payment_method"`
	DisbursementAccountID     uuid.UUID           `json:"disbursement_account_id"`
	DisbursementAccountNumber string              `json:"disbursement_account_number,omitempty"`
	DisbursementAccountNameAr string              `json:"disbursement_account_name_ar,omitempty"`
	Amount                    decimal.Decimal     `json:"amount"`
	CurrencyCode              string              `json:"currency_code"`
	ReferenceNumber           *string             `json:"reference_number,omitempty"`
	Status                    string              `json:"status"`
	Notes                     *string             `json:"notes,omitempty"`
	AllocatedAmount           decimal.Decimal     `json:"allocated_amount"`
	UnallocatedAmount         decimal.Decimal     `json:"unallocated_amount"`
	JournalEntryID            *uuid.UUID          `json:"journal_entry_id,omitempty"`
	CreatedByUserID           *uuid.UUID          `json:"created_by_user_id,omitempty"`
	CreatedAt                 time.Time           `json:"created_at"`
	Allocations               []PaymentAllocation `json:"allocations,omitempty"`
}

type PaymentAllocation struct {
	ID              uuid.UUID            `json:"id"`
	PaymentID       uuid.UUID            `json:"payment_id"`
	OpenItemID      uuid.UUID            `json:"open_item_id"`
	DocumentType    OpenItemDocumentType `json:"document_type,omitempty"`
	DocumentNumber  string               `json:"document_number,omitempty"`
	DueDate         time.Time            `json:"due_date,omitempty"`
	OriginalAmount  decimal.Decimal      `json:"original_amount,omitempty"`
	RemainingAmount decimal.Decimal      `json:"remaining_amount,omitempty"`
	AllocatedAmount decimal.Decimal      `json:"allocated_amount"`
	AllocatedAt     time.Time            `json:"allocated_at"`
}

type PayablesOpenItem struct {
	ID              uuid.UUID            `json:"id"`
	LegalEntityID   uuid.UUID            `json:"legal_entity_id"`
	SupplierID      uuid.UUID            `json:"supplier_id"`
	SupplierNumber  string               `json:"supplier_number,omitempty"`
	SupplierNameAr  string               `json:"supplier_name_ar,omitempty"`
	DocumentType    OpenItemDocumentType `json:"document_type"`
	DocumentNumber  string               `json:"document_number"`
	DocumentID      *uuid.UUID           `json:"document_id,omitempty"`
	DocumentDate    time.Time            `json:"document_date"`
	DueDate         time.Time            `json:"due_date"`
	CurrencyCode    string               `json:"currency_code"`
	OriginalAmount  decimal.Decimal      `json:"original_amount"`
	AllocatedAmount decimal.Decimal      `json:"allocated_amount"`
	RemainingAmount decimal.Decimal      `json:"remaining_amount"`
	Status          OpenItemStatus       `json:"status"`
	DaysOverdue     int32                `json:"days_overdue"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type SupplierAgingSummary struct {
	TotalOutstanding decimal.Decimal `json:"total_outstanding"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	Overdue1to30     decimal.Decimal `json:"overdue_1_30"`
	Overdue31to60    decimal.Decimal `json:"overdue_31_60"`
	Overdue61to90    decimal.Decimal `json:"overdue_61_90"`
	OverdueOver90    decimal.Decimal `json:"overdue_over_90"`
}

func CalculateLineAmounts(qty, unitPrice, discountPct, taxRate decimal.Decimal) (subtotal, discountAmt, taxAmt, totalAmt decimal.Decimal) {
	gross := qty.Mul(unitPrice).Round(6)
	if discountPct.GreaterThan(decimal.Zero) {
		discountAmt = gross.Mul(discountPct).Div(decimal.NewFromInt(100)).Round(6)
	} else {
		discountAmt = decimal.Zero
	}
	subtotal = gross.Sub(discountAmt).Round(6)
	if taxRate.GreaterThan(decimal.Zero) {
		taxAmt = subtotal.Mul(taxRate).Div(decimal.NewFromInt(100)).Round(6)
	} else {
		taxAmt = decimal.Zero
	}
	totalAmt = subtotal.Add(taxAmt).Round(6)
	return
}
