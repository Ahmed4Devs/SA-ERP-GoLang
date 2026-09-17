package application

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"erp/internal/reporting/domain"
	"erp/internal/reporting/ports"

	"github.com/google/uuid"
)

type reportingService struct {
	repo ports.ReportingRepository
}

func NewReportingService(repo ports.ReportingRepository) ports.ReportingService {
	return &reportingService{
		repo: repo,
	}
}

func (s *reportingService) GetCatalog(ctx context.Context) []domain.ReportCatalogItem {
	return []domain.ReportCatalogItem{
		{
			Code:        "TRIAL_BALANCE",
			Category:    "FINANCIAL",
			NameEn:      "Trial Balance",
			NameAr:      "ميزان المراجعة",
			Description: "Comprehensive list of all ledger accounts with debits, credits, and net ending balances.",
			Endpoint:    "/api/v1/reports/trial-balance",
		},
		{
			Code:        "GENERAL_LEDGER",
			Category:    "FINANCIAL",
			NameEn:      "General Ledger",
			NameAr:      "دفتر الأستاذ العام",
			Description: "Detailed chronological record of all posted journal transactions across accounts.",
			Endpoint:    "/api/v1/reports/general-ledger",
		},
		{
			Code:        "BALANCE_SHEET",
			Category:    "FINANCIAL",
			NameEn:      "Balance Sheet",
			NameAr:      "الميزانية العمومية",
			Description: "Statement of financial position showing total assets, liabilities, and owner's equity.",
			Endpoint:    "/api/v1/reports/balance-sheet",
		},
		{
			Code:        "INCOME_STATEMENT",
			Category:    "FINANCIAL",
			NameEn:      "Income Statement (P&L)",
			NameAr:      "قائمة الدخل والأرباح والخسائر",
			Description: "Operating revenues, expenses, and net profit/loss for the fiscal period.",
			Endpoint:    "/api/v1/reports/income-statement",
		},
		{
			Code:        "CASH_FLOW",
			Category:    "FINANCIAL",
			NameEn:      "Cash Flow Statement",
			NameAr:      "قائمة التدفقات النقدية",
			Description: "Cash inflows and outflows across operating, investing, and financing activities.",
			Endpoint:    "/api/v1/reports/cash-flow",
		},
		{
			Code:        "AR_AGING",
			Category:    "COMMERCIAL",
			NameEn:      "Accounts Receivable Aging",
			NameAr:      "أعمار ديون العملاء",
			Description: "Outstanding customer receivables categorized by age buckets (Current, 1-30, 31-60, 61-90, 90+ days).",
			Endpoint:    "/api/v1/reports/ar-aging",
		},
		{
			Code:        "AP_AGING",
			Category:    "COMMERCIAL",
			NameEn:      "Accounts Payable Aging",
			NameAr:      "أعمار ديون الموردين",
			Description: "Unpaid vendor obligations categorized by age buckets for liquidity and cash planning.",
			Endpoint:    "/api/v1/reports/ap-aging",
		},
		{
			Code:        "CUSTOMER_STATEMENT",
			Category:    "COMMERCIAL",
			NameEn:      "Customer Account Statement",
			NameAr:      "كشف حساب عميل",
			Description: "Full transaction ledger and running balance for a specific customer over a date range.",
			Endpoint:    "/api/v1/reports/customer-statement",
		},
		{
			Code:        "SUPPLIER_STATEMENT",
			Category:    "COMMERCIAL",
			NameEn:      "Supplier Account Statement",
			NameAr:      "كشف حساب مورد",
			Description: "Full transaction ledger and running balance for a specific supplier over a date range.",
			Endpoint:    "/api/v1/reports/supplier-statement",
		},
		{
			Code:        "INVENTORY_VALUATION",
			Category:    "SUPPLY_CHAIN",
			NameEn:      "Inventory Valuation",
			NameAr:      "تقييم المخزون والسلع",
			Description: "Physical stock balances on hand multiplied by weighted average cost across warehouses.",
			Endpoint:    "/api/v1/reports/inventory-valuation",
		},
		{
			Code:        "SALES_REPORT",
			Category:    "SUPPLY_CHAIN",
			NameEn:      "Sales Performance Report",
			NameAr:      "تقرير حركة المبيعات",
			Description: "Comprehensive analytics of invoiced sales, taxable totals, discounts, and customer breakdown.",
			Endpoint:    "/api/v1/reports/sales",
		},
		{
			Code:        "PURCHASING_REPORT",
			Category:    "SUPPLY_CHAIN",
			NameEn:      "Purchasing Summary Report",
			NameAr:      "تقرير حركة المشتريات",
			Description: "Analysis of purchase orders, vendor invoices, and procurement expenditure.",
			Endpoint:    "/api/v1/reports/purchasing",
		},
		{
			Code:        "VAT_RETURN",
			Category:    "COMPLIANCE",
			NameEn:      "VAT Return 15% (ZATCA)",
			NameAr:      "إقرار ضريبة القيمة المضافة 15%",
			Description: "Standard 15% Saudi VAT return boxes with output tax, input tax, and net payable/refundable.",
			Endpoint:    "/api/v1/reports/vat-return",
		},
		{
			Code:        "PAYROLL_SUMMARY",
			Category:    "HR_PAYROLL",
			NameEn:      "Payroll Summary Report",
			NameAr:      "مسير الرواتب المجمع",
			Description: "Monthly breakdown of basic salary, allowances, deductions, and net disbursed wages.",
			Endpoint:    "/api/v1/reports/payroll-summary",
		},
		{
			Code:        "GOSI_CONTRIBUTION",
			Category:    "HR_PAYROLL",
			NameEn:      "GOSI Social Insurance Report",
			NameAr:      "تقرير التأمينات الاجتماعية GOSI",
			Description: "Statutory GOSI contributory wages and monthly employer/employee remittance shares.",
			Endpoint:    "/api/v1/reports/gosi-summary",
		},
		{
			Code:        "BUDGET_VS_ACTUAL",
			Category:    "MANAGEMENT",
			NameEn:      "Budget vs Actual Variance",
			NameAr:      "الموازنة مقابل الفعلي والحيود",
			Description: "Comparative analysis of allocated budget limits versus posted ledger actuals.",
			Endpoint:    "/api/v1/reports/budget-vs-actual",
		},
	}
}

func (s *reportingService) validateFilter(filter domain.ReportFilter) error {
	if filter.LegalEntityID == uuid.Nil {
		return domain.ErrLegalEntityRequired
	}
	if filter.StartDate != nil && filter.EndDate != nil && filter.StartDate.After(*filter.EndDate) {
		return domain.ErrInvalidDateRange
	}
	return nil
}

func (s *reportingService) GetTrialBalance(ctx context.Context, filter domain.ReportFilter) (*domain.TrialBalanceReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetTrialBalance(ctx, filter)
}

func (s *reportingService) GetGeneralLedger(ctx context.Context, filter domain.ReportFilter) (*domain.GeneralLedgerReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetGeneralLedger(ctx, filter)
}

func (s *reportingService) GetBalanceSheet(ctx context.Context, filter domain.ReportFilter) (*domain.BalanceSheetReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetBalanceSheet(ctx, filter)
}

func (s *reportingService) GetIncomeStatement(ctx context.Context, filter domain.ReportFilter) (*domain.IncomeStatementReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetIncomeStatement(ctx, filter)
}

func (s *reportingService) GetCashFlow(ctx context.Context, filter domain.ReportFilter) (*domain.CashFlowReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetCashFlow(ctx, filter)
}

func (s *reportingService) GetARAging(ctx context.Context, filter domain.ReportFilter) (*domain.AgingReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetAgingReport(ctx, "AR", filter)
}

func (s *reportingService) GetAPAging(ctx context.Context, filter domain.ReportFilter) (*domain.AgingReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetAgingReport(ctx, "AP", filter)
}

func (s *reportingService) GetCustomerStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetCustomerStatement(ctx, filter)
}

func (s *reportingService) GetSupplierStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetSupplierStatement(ctx, filter)
}

func (s *reportingService) GetInventoryValuation(ctx context.Context, filter domain.ReportFilter) (*domain.InventoryValuationReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetInventoryValuation(ctx, filter)
}

func (s *reportingService) GetSalesReport(ctx context.Context, filter domain.ReportFilter) (*domain.SalesReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetSalesReport(ctx, filter)
}

func (s *reportingService) GetPurchasingReport(ctx context.Context, filter domain.ReportFilter) (*domain.PurchasingReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetPurchasingReport(ctx, filter)
}

func (s *reportingService) GetVATReturnReport(ctx context.Context, filter domain.ReportFilter) (*domain.VATReturnReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetVATReturnReport(ctx, filter)
}

func (s *reportingService) GetPayrollSummary(ctx context.Context, filter domain.ReportFilter) (*domain.PayrollSummaryReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetPayrollSummary(ctx, filter)
}

func (s *reportingService) GetGOSIContribution(ctx context.Context, filter domain.ReportFilter) (*domain.GOSIContributionReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetGOSIContribution(ctx, filter)
}

func (s *reportingService) GetBudgetVsActual(ctx context.Context, filter domain.ReportFilter) (*domain.BudgetVsActualReport, error) {
	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.GetBudgetVsActual(ctx, filter)
}

// ============================================================================
// CSV EXPORT ENGINE
// ============================================================================

func (s *reportingService) ExportReportToCSV(ctx context.Context, reportCode string, filter domain.ReportFilter) ([]byte, string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s_%s.csv", reportCode, timestamp)

	switch reportCode {
	case "TRIAL_BALANCE":
		rpt, err := s.GetTrialBalance(ctx, filter)
		if err != nil {
			return nil, "", err
		}
		_ = w.Write([]string{"Account Code", "Account Name (EN)", "Account Name (AR)", "Account Type", "Normal Balance", "Total Debit", "Total Credit", "Net Balance"})
		for _, r := range rpt.Rows {
			_ = w.Write([]string{r.AccountCode, r.AccountNameEn, r.AccountNameAr, r.AccountType, r.NormalBalance, r.TotalDebit.String(), r.TotalCredit.String(), r.NetBalance.String()})
		}
		_ = w.Write([]string{"TOTAL", "", "", "", "", rpt.TotalDebits.String(), rpt.TotalCredits.String(), ""})

	case "GENERAL_LEDGER":
		rpt, err := s.GetGeneralLedger(ctx, filter)
		if err != nil {
			return nil, "", err
		}
		_ = w.Write([]string{"Entry Number", "Date", "Source Module", "Account Code", "Account Name", "Debit", "Credit", "Description"})
		for _, e := range rpt.Entries {
			desc := ""
			if e.Description != nil {
				desc = *e.Description
			}
			_ = w.Write([]string{e.EntryNumber, e.AccountingDate.Format("2006-01-02"), e.SourceModule, e.AccountCode, e.AccountNameEn, e.Debit.String(), e.Credit.String(), desc})
		}

	case "INVENTORY_VALUATION":
		rpt, err := s.GetInventoryValuation(ctx, filter)
		if err != nil {
			return nil, "", err
		}
		_ = w.Write([]string{"Item Code", "Item Name", "Warehouse Code", "Quantity On Hand", "Quantity Reserved", "Quantity Available", "Average Cost", "Total Valuation"})
		for _, i := range rpt.Items {
			_ = w.Write([]string{i.ItemCode, i.ItemNameEn, i.WarehouseCode, i.QuantityOnHand.String(), i.QuantityReserved.String(), i.QuantityAvailable.String(), i.AverageCost.String(), i.TotalValue.String()})
		}

	case "SALES_REPORT":
		rpt, err := s.GetSalesReport(ctx, filter)
		if err != nil {
			return nil, "", err
		}
		_ = w.Write([]string{"Invoice Number", "Issue Date", "Due Date", "Customer Name", "Subtotal", "VAT Amount", "Total Amount", "Status"})
		for _, inv := range rpt.Invoices {
			_ = w.Write([]string{inv.InvoiceNumber, inv.IssueDate.Format("2006-01-02"), inv.DueDate.Format("2006-01-02"), inv.CustomerNameEn, inv.Subtotal.String(), inv.TaxAmount.String(), inv.TotalAmount.String(), inv.Status})
		}

	case "AR_AGING", "AP_AGING":
		rpt, err := s.repo.GetAgingReport(ctx, reportCode, filter)
		if err != nil {
			return nil, "", err
		}
		_ = w.Write([]string{"Entity Number", "Entity Name (EN)", "Entity Name (AR)", "Invoices/Bills", "Total Outstanding", "Current", "1-30 Days", "31-60 Days", "61-90 Days", "91-120 Days", ">120 Days"})
		for _, r := range rpt.Rows {
			_ = w.Write([]string{r.EntityNumber, r.EntityNameEn, r.EntityNameAr, fmt.Sprintf("%d", r.DocumentCount), r.TotalOutstanding.String(), r.CurrentAmount.String(), r.Aging1To30.String(), r.Aging31To60.String(), r.Aging61To90.String(), r.Aging91To120.String(), r.AgingOver120.String()})
		}

	case "BUDGET_VS_ACTUAL":
		rpt, err := s.GetBudgetVsActual(ctx, filter)
		if err != nil {
			return nil, "", err
		}
		_ = w.Write([]string{"Budget Name", "Fiscal Year", "Account Code", "Account Name", "Budgeted Amount", "Actual Amount", "Variance Amount", "Variance %"})
		for _, v := range rpt.Variances {
			_ = w.Write([]string{v.BudgetNameEn, v.FiscalYearName, v.AccountCode, v.AccountNameEn, v.BudgetedAmount.String(), v.ActualAmount.String(), v.VarianceAmount.String(), v.VariancePercentage.String() + "%"})
		}

	default:
		// Default generic catalog or empty
		_ = w.Write([]string{"Report Code", "Generated At", "Status"})
		_ = w.Write([]string{reportCode, timestamp, "EXPORTED"})
	}

	w.Flush()
	return buf.Bytes(), filename, nil
}
