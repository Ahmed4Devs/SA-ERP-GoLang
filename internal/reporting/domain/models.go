package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReportFilter holds common query filters across all ERP reports
type ReportFilter struct {
	LegalEntityID uuid.UUID
	StartDate     *time.Time
	EndDate       *time.Time
	AsOfDate      *time.Time
	AccountID     *uuid.UUID
	CustomerID    *uuid.UUID
	SupplierID    *uuid.UUID
	WarehouseID   *uuid.UUID
	BudgetID      *uuid.UUID
	Status        *string
	Year          *int
	Month         *int
}

// ============================================================================
// 1. TRIAL BALANCE
// ============================================================================

type TrialBalanceRow struct {
	AccountID     uuid.UUID       `json:"account_id"`
	AccountCode   string          `json:"account_code"`
	AccountNameEn string          `json:"account_name_en"`
	AccountNameAr string          `json:"account_name_ar"`
	AccountType   string          `json:"account_type"`
	NormalBalance string          `json:"normal_balance"`
	TotalDebit    decimal.Decimal `json:"total_debit"`
	TotalCredit   decimal.Decimal `json:"total_credit"`
	NetBalance    decimal.Decimal `json:"net_balance"`
}

type TrialBalanceReport struct {
	LegalEntityID uuid.UUID         `json:"legal_entity_id"`
	GeneratedAt   time.Time         `json:"generated_at"`
	StartDate     *time.Time        `json:"start_date,omitempty"`
	EndDate       *time.Time        `json:"end_date,omitempty"`
	Rows          []TrialBalanceRow `json:"rows"`
	TotalDebits   decimal.Decimal   `json:"total_debits"`
	TotalCredits  decimal.Decimal   `json:"total_credits"`
	IsBalanced    bool              `json:"is_balanced"`
}

// ============================================================================
// 2. GENERAL LEDGER
// ============================================================================

type GeneralLedgerEntry struct {
	EntryID        uuid.UUID       `json:"entry_id"`
	EntryNumber    string          `json:"entry_number"`
	AccountingDate time.Time       `json:"accounting_date"`
	SourceModule   string          `json:"source_module"`
	Reference      *string         `json:"reference,omitempty"`
	LineID         uuid.UUID       `json:"line_id"`
	LineNumber     int             `json:"line_number"`
	AccountID      uuid.UUID       `json:"account_id"`
	AccountCode    string          `json:"account_code"`
	AccountNameEn  string          `json:"account_name_en"`
	AccountNameAr  string          `json:"account_name_ar"`
	AccountType    string          `json:"account_type"`
	Debit          decimal.Decimal `json:"debit"`
	Credit         decimal.Decimal `json:"credit"`
	Description    *string         `json:"description,omitempty"`
}

type GeneralLedgerReport struct {
	LegalEntityID uuid.UUID            `json:"legal_entity_id"`
	GeneratedAt   time.Time            `json:"generated_at"`
	AccountID     *uuid.UUID           `json:"account_id,omitempty"`
	StartDate     *time.Time           `json:"start_date,omitempty"`
	EndDate       *time.Time           `json:"end_date,omitempty"`
	Entries       []GeneralLedgerEntry `json:"entries"`
	TotalDebit    decimal.Decimal      `json:"total_debit"`
	TotalCredit   decimal.Decimal      `json:"total_credit"`
}

// ============================================================================
// 3. BALANCE SHEET
// ============================================================================

type BalanceSheetItem struct {
	AccountID     uuid.UUID       `json:"account_id"`
	AccountCode   string          `json:"account_code"`
	AccountNameEn string          `json:"account_name_en"`
	AccountNameAr string          `json:"account_name_ar"`
	Balance       decimal.Decimal `json:"balance"`
}

type BalanceSheetCategory struct {
	CategoryNameEn string             `json:"category_name_en"`
	CategoryNameAr string             `json:"category_name_ar"`
	Items          []BalanceSheetItem `json:"items"`
	Subtotal       decimal.Decimal    `json:"subtotal"`
}

type BalanceSheetReport struct {
	LegalEntityID    uuid.UUID              `json:"legal_entity_id"`
	GeneratedAt      time.Time              `json:"generated_at"`
	AsOfDate         time.Time              `json:"as_of_date"`
	Assets           []BalanceSheetCategory `json:"assets"`
	Liabilities      []BalanceSheetCategory `json:"liabilities"`
	Equity           []BalanceSheetCategory `json:"equity"`
	TotalAssets      decimal.Decimal        `json:"total_assets"`
	TotalLiabilities decimal.Decimal        `json:"total_liabilities"`
	TotalEquity      decimal.Decimal        `json:"total_equity"`
	TotalLiabEquity  decimal.Decimal        `json:"total_liabilities_and_equity"`
	IsBalanced       bool                   `json:"is_balanced"`
}

// ============================================================================
// 4. INCOME STATEMENT (PROFIT & LOSS)
// ============================================================================

type IncomeStatementItem struct {
	AccountID     uuid.UUID       `json:"account_id"`
	AccountCode   string          `json:"account_code"`
	AccountNameEn string          `json:"account_name_en"`
	AccountNameAr string          `json:"account_name_ar"`
	Amount        decimal.Decimal `json:"amount"`
}

type IncomeStatementCategory struct {
	CategoryNameEn string                `json:"category_name_en"`
	CategoryNameAr string                `json:"category_name_ar"`
	Items          []IncomeStatementItem `json:"items"`
	TotalAmount    decimal.Decimal       `json:"total_amount"`
}

type IncomeStatementReport struct {
	LegalEntityID    uuid.UUID                 `json:"legal_entity_id"`
	GeneratedAt      time.Time                 `json:"generated_at"`
	StartDate        time.Time                 `json:"start_date"`
	EndDate          time.Time                 `json:"end_date"`
	Revenues         []IncomeStatementCategory `json:"revenues"`
	Expenses         []IncomeStatementCategory `json:"expenses"`
	TotalRevenue     decimal.Decimal           `json:"total_revenue"`
	TotalExpenses    decimal.Decimal           `json:"total_expenses"`
	NetIncome        decimal.Decimal           `json:"net_income"`
	ProfitMarginRate decimal.Decimal           `json:"profit_margin_rate"`
}

// ============================================================================
// 5. CASH FLOW STATEMENT
// ============================================================================

type CashFlowItem struct {
	TitleEn string          `json:"title_en"`
	TitleAr string          `json:"title_ar"`
	Amount  decimal.Decimal `json:"amount"`
}

type CashFlowSection struct {
	SectionEn string          `json:"section_en"`
	SectionAr string          `json:"section_ar"`
	Items     []CashFlowItem  `json:"items"`
	NetTotal  decimal.Decimal `json:"net_total"`
}

type CashFlowReport struct {
	LegalEntityID       uuid.UUID       `json:"legal_entity_id"`
	GeneratedAt         time.Time       `json:"generated_at"`
	StartDate           time.Time       `json:"start_date"`
	EndDate             time.Time       `json:"end_date"`
	OperatingActivities CashFlowSection `json:"operating_activities"`
	InvestingActivities CashFlowSection `json:"investing_activities"`
	FinancingActivities CashFlowSection `json:"financing_activities"`
	NetCashChange       decimal.Decimal `json:"net_cash_change"`
}

// ============================================================================
// 6 & 7. AGING REPORTS (AR & AP)
// ============================================================================

type AgingBucketRow struct {
	EntityID         uuid.UUID       `json:"entity_id"`
	EntityNumber     string          `json:"entity_number"`
	EntityNameEn     string          `json:"entity_name_en"`
	EntityNameAr     string          `json:"entity_name_ar"`
	DocumentCount    int             `json:"document_count"`
	TotalOutstanding decimal.Decimal `json:"total_outstanding"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	Aging1To30       decimal.Decimal `json:"aging_1_30"`
	Aging31To60      decimal.Decimal `json:"aging_31_60"`
	Aging61To90      decimal.Decimal `json:"aging_61_90"`
	Aging91To120     decimal.Decimal `json:"aging_91_120"`
	AgingOver120     decimal.Decimal `json:"aging_over_120"`
}

type AgingReport struct {
	ReportType       string           `json:"report_type"` // 'AR' or 'AP'
	LegalEntityID    uuid.UUID        `json:"legal_entity_id"`
	GeneratedAt      time.Time        `json:"generated_at"`
	Rows             []AgingBucketRow `json:"rows"`
	TotalOutstanding decimal.Decimal  `json:"total_outstanding"`
	TotalCurrent     decimal.Decimal  `json:"total_current"`
	Total1To30       decimal.Decimal  `json:"total_1_30"`
	Total31To60      decimal.Decimal  `json:"total_31_60"`
	Total61To90      decimal.Decimal  `json:"total_61_90"`
	Total91To120     decimal.Decimal  `json:"total_91_120"`
	TotalOver120     decimal.Decimal  `json:"total_over_120"`
}

// ============================================================================
// 8 & 9. STATEMENTS (CUSTOMER & SUPPLIER)
// ============================================================================

type StatementLineItem struct {
	ItemID          uuid.UUID       `json:"item_id"`
	DocumentType    string          `json:"document_type"`
	DocumentNumber  string          `json:"document_number"`
	DocumentDate    time.Time       `json:"document_date"`
	DueDate         time.Time       `json:"due_date"`
	CurrencyCode    string          `json:"currency_code"`
	OriginalAmount  decimal.Decimal `json:"original_amount"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
	RemainingAmount decimal.Decimal `json:"remaining_amount"`
	Status          string          `json:"status"`
}

type EntityStatementReport struct {
	ReportType    string              `json:"report_type"` // 'CUSTOMER' or 'SUPPLIER'
	LegalEntityID uuid.UUID           `json:"legal_entity_id"`
	EntityID      uuid.UUID           `json:"entity_id"`
	EntityNumber  string              `json:"entity_number"`
	EntityNameEn  string              `json:"entity_name_en"`
	EntityNameAr  string              `json:"entity_name_ar"`
	GeneratedAt   time.Time           `json:"generated_at"`
	StartDate     *time.Time          `json:"start_date,omitempty"`
	EndDate       *time.Time          `json:"end_date,omitempty"`
	Lines         []StatementLineItem `json:"lines"`
	TotalDebited  decimal.Decimal     `json:"total_debited"`
	TotalCredited decimal.Decimal     `json:"total_credited"`
	EndingBalance decimal.Decimal     `json:"ending_balance"`
}

// ============================================================================
// 10. INVENTORY VALUATION
// ============================================================================

type InventoryValuationRow struct {
	ItemID            uuid.UUID       `json:"item_id"`
	ItemCode          string          `json:"item_code"`
	ItemNameEn        string          `json:"item_name_en"`
	ItemNameAr        string          `json:"item_name_ar"`
	WarehouseID       uuid.UUID       `json:"warehouse_id"`
	WarehouseCode     string          `json:"warehouse_code"`
	WarehouseNameEn   string          `json:"warehouse_name_en"`
	WarehouseNameAr   string          `json:"warehouse_name_ar"`
	QuantityOnHand    decimal.Decimal `json:"quantity_on_hand"`
	QuantityReserved  decimal.Decimal `json:"quantity_reserved"`
	QuantityAvailable decimal.Decimal `json:"quantity_available"`
	AverageCost       decimal.Decimal `json:"average_cost"`
	TotalValue        decimal.Decimal `json:"total_value"`
}

type InventoryValuationReport struct {
	LegalEntityID  uuid.UUID               `json:"legal_entity_id"`
	WarehouseID    *uuid.UUID              `json:"warehouse_id,omitempty"`
	GeneratedAt    time.Time               `json:"generated_at"`
	Items          []InventoryValuationRow `json:"items"`
	TotalItems     int                     `json:"total_items"`
	TotalQuantity  decimal.Decimal         `json:"total_quantity"`
	TotalValuation decimal.Decimal         `json:"total_valuation"`
}

// ============================================================================
// 11. SALES REPORT
// ============================================================================

type SalesInvoiceRow struct {
	InvoiceID      uuid.UUID       `json:"invoice_id"`
	InvoiceNumber  string          `json:"invoice_number"`
	IssueDate      time.Time       `json:"issue_date"`
	DueDate        time.Time       `json:"due_date"`
	CustomerNameEn string          `json:"customer_name_en"`
	CustomerNameAr string          `json:"customer_name_ar"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	TaxAmount      decimal.Decimal `json:"tax_amount"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	Status         string          `json:"status"`
}

type SalesReport struct {
	LegalEntityID uuid.UUID         `json:"legal_entity_id"`
	GeneratedAt   time.Time         `json:"generated_at"`
	StartDate     *time.Time        `json:"start_date,omitempty"`
	EndDate       *time.Time        `json:"end_date,omitempty"`
	Invoices      []SalesInvoiceRow `json:"invoices"`
	TotalCount    int               `json:"total_count"`
	TotalSubtotal decimal.Decimal   `json:"total_subtotal"`
	TotalTax      decimal.Decimal   `json:"total_tax"`
	TotalSales    decimal.Decimal   `json:"total_sales"`
}

// ============================================================================
// 12. PURCHASING REPORT
// ============================================================================

type PurchasingOrderRow struct {
	POID           uuid.UUID       `json:"po_id"`
	PONumber       string          `json:"po_number"`
	OrderDate      time.Time       `json:"order_date"`
	SupplierNumber string          `json:"supplier_number"`
	SupplierNameEn string          `json:"supplier_name_en"`
	SupplierNameAr string          `json:"supplier_name_ar"`
	Subtotal       decimal.Decimal `json:"subtotal"`
	TaxAmount      decimal.Decimal `json:"tax_amount"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	Status         string          `json:"status"`
}

type PurchasingReport struct {
	LegalEntityID  uuid.UUID            `json:"legal_entity_id"`
	GeneratedAt    time.Time            `json:"generated_at"`
	StartDate      *time.Time           `json:"start_date,omitempty"`
	EndDate        *time.Time           `json:"end_date,omitempty"`
	Orders         []PurchasingOrderRow `json:"orders"`
	TotalCount     int                  `json:"total_count"`
	TotalSubtotal  decimal.Decimal      `json:"total_subtotal"`
	TotalTax       decimal.Decimal      `json:"total_tax"`
	TotalPurchases decimal.Decimal      `json:"total_purchases"`
}

// ============================================================================
// 13. VAT RETURN REPORT (SAUDI ZATCA 15%)
// ============================================================================

type VATReturnBox struct {
	BoxNumber     int             `json:"box_number"`
	DescriptionEn string          `json:"description_en"`
	DescriptionAr string          `json:"description_ar"`
	TaxableAmount decimal.Decimal `json:"taxable_amount"`
	AdjustmentAmt decimal.Decimal `json:"adjustment_amt"`
	VATAmount     decimal.Decimal `json:"vat_amount"`
}

type VATReturnReport struct {
	LegalEntityID         uuid.UUID       `json:"legal_entity_id"`
	GeneratedAt           time.Time       `json:"generated_at"`
	StartDate             *time.Time      `json:"start_date,omitempty"`
	EndDate               *time.Time      `json:"end_date,omitempty"`
	SalesBoxes            []VATReturnBox  `json:"sales_boxes"`
	PurchaseBoxes         []VATReturnBox  `json:"purchase_boxes"`
	TotalOutputVAT        decimal.Decimal `json:"total_output_vat"`
	TotalInputVAT         decimal.Decimal `json:"total_input_vat"`
	NetVATDueOrRefundable decimal.Decimal `json:"net_vat_due_or_refundable"`
}

// ============================================================================
// 14. PAYROLL SUMMARY REPORT
// ============================================================================

type PayrollRunRow struct {
	RunID           uuid.UUID       `json:"run_id"`
	RunNumber       string          `json:"run_number"`
	PeriodID        uuid.UUID       `json:"period_id"`
	PeriodNameEn    string          `json:"period_name_en"`
	PeriodNameAr    string          `json:"period_name_ar"`
	RunDate         time.Time       `json:"run_date"`
	Status          string          `json:"status"`
	EmployeeCount   int             `json:"employee_count"`
	TotalBasic      decimal.Decimal `json:"total_basic"`
	TotalAllowances decimal.Decimal `json:"total_allowances"`
	TotalGross      decimal.Decimal `json:"total_gross"`
	TotalDeductions decimal.Decimal `json:"total_deductions"`
	TotalNet        decimal.Decimal `json:"total_net"`
}

type PayrollSummaryReport struct {
	LegalEntityID   uuid.UUID       `json:"legal_entity_id"`
	GeneratedAt     time.Time       `json:"generated_at"`
	StartDate       *time.Time      `json:"start_date,omitempty"`
	EndDate         *time.Time      `json:"end_date,omitempty"`
	Runs            []PayrollRunRow `json:"runs"`
	TotalGrossPaid  decimal.Decimal `json:"total_gross_paid"`
	TotalDeductions decimal.Decimal `json:"total_deductions"`
	TotalNetPaid    decimal.Decimal `json:"total_net_paid"`
}

// ============================================================================
// 15. GOSI CONTRIBUTION REPORT
// ============================================================================

type GOSICalculationRow struct {
	CalculationID      uuid.UUID       `json:"calculation_id"`
	CalculationYear    int             `json:"calculation_year"`
	CalculationMonth   int             `json:"calculation_month"`
	Status             string          `json:"status"`
	EmployeeNumber     string          `json:"employee_number"`
	EmployeeNameEn     string          `json:"employee_name_en"`
	EmployeeNameAr     string          `json:"employee_name_ar"`
	IsSaudiNational    bool            `json:"is_saudi_national"`
	ContributionBasis  decimal.Decimal `json:"contribution_basis"`
	TotalEmployeeShare decimal.Decimal `json:"total_employee_share"`
	TotalEmployerShare decimal.Decimal `json:"total_employer_share"`
	TotalContribution  decimal.Decimal `json:"total_contribution"`
	CreatedAt          time.Time       `json:"created_at"`
}

type GOSIContributionReport struct {
	LegalEntityID          uuid.UUID            `json:"legal_entity_id"`
	GeneratedAt            time.Time            `json:"generated_at"`
	Year                   *int                 `json:"year,omitempty"`
	Month                  *int                 `json:"month,omitempty"`
	Calculations           []GOSICalculationRow `json:"calculations"`
	TotalEmployees         int                  `json:"total_employees"`
	TotalSaudiCount        int                  `json:"total_saudi_count"`
	TotalNonSaudiCount     int                  `json:"total_non_saudi_count"`
	TotalContributoryWages decimal.Decimal      `json:"total_contributory_wages"`
	TotalEmployeeDeduction decimal.Decimal      `json:"total_employee_deduction"`
	TotalEmployerPayment   decimal.Decimal      `json:"total_employer_payment"`
	TotalRemittanceDue     decimal.Decimal      `json:"total_remittance_due"`
}

// ============================================================================
// 16. BUDGET VS ACTUAL REPORT
// ============================================================================

type BudgetVarianceRow struct {
	BudgetID           uuid.UUID       `json:"budget_id"`
	BudgetNameEn       string          `json:"budget_name_en"`
	BudgetNameAr       string          `json:"budget_name_ar"`
	FiscalYearName     string          `json:"fiscal_year_name"`
	AccountID          uuid.UUID       `json:"account_id"`
	AccountCode        string          `json:"account_code"`
	AccountNameEn      string          `json:"account_name_en"`
	AccountNameAr      string          `json:"account_name_ar"`
	AccountType        string          `json:"account_type"`
	BudgetedAmount     decimal.Decimal `json:"budgeted_amount"`
	ActualAmount       decimal.Decimal `json:"actual_amount"`
	VarianceAmount     decimal.Decimal `json:"variance_amount"`
	VariancePercentage decimal.Decimal `json:"variance_percentage"`
}

type BudgetVsActualReport struct {
	LegalEntityID    uuid.UUID           `json:"legal_entity_id"`
	BudgetID         *uuid.UUID          `json:"budget_id,omitempty"`
	GeneratedAt      time.Time           `json:"generated_at"`
	Variances        []BudgetVarianceRow `json:"variances"`
	TotalBudgeted    decimal.Decimal     `json:"total_budgeted"`
	TotalActual      decimal.Decimal     `json:"total_actual"`
	TotalNetVariance decimal.Decimal     `json:"total_net_variance"`
}

// ============================================================================
// REPORT CATALOG (DIRECTORY OF AVAILABLE REPORTS)
// ============================================================================

type ReportCatalogItem struct {
	Code        string `json:"code"`
	Category    string `json:"category"`
	NameEn      string `json:"name_en"`
	NameAr      string `json:"name_ar"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
}
