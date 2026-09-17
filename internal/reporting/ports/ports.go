package ports

import (
	"context"

	"erp/internal/reporting/domain"
)

type ReportingRepository interface {
	GetTrialBalance(ctx context.Context, filter domain.ReportFilter) (*domain.TrialBalanceReport, error)
	GetGeneralLedger(ctx context.Context, filter domain.ReportFilter) (*domain.GeneralLedgerReport, error)
	GetBalanceSheet(ctx context.Context, filter domain.ReportFilter) (*domain.BalanceSheetReport, error)
	GetIncomeStatement(ctx context.Context, filter domain.ReportFilter) (*domain.IncomeStatementReport, error)
	GetCashFlow(ctx context.Context, filter domain.ReportFilter) (*domain.CashFlowReport, error)
	GetAgingReport(ctx context.Context, reportType string, filter domain.ReportFilter) (*domain.AgingReport, error)
	GetCustomerStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error)
	GetSupplierStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error)
	GetInventoryValuation(ctx context.Context, filter domain.ReportFilter) (*domain.InventoryValuationReport, error)
	GetSalesReport(ctx context.Context, filter domain.ReportFilter) (*domain.SalesReport, error)
	GetPurchasingReport(ctx context.Context, filter domain.ReportFilter) (*domain.PurchasingReport, error)
	GetVATReturnReport(ctx context.Context, filter domain.ReportFilter) (*domain.VATReturnReport, error)
	GetPayrollSummary(ctx context.Context, filter domain.ReportFilter) (*domain.PayrollSummaryReport, error)
	GetGOSIContribution(ctx context.Context, filter domain.ReportFilter) (*domain.GOSIContributionReport, error)
	GetBudgetVsActual(ctx context.Context, filter domain.ReportFilter) (*domain.BudgetVsActualReport, error)
}

type ReportingService interface {
	GetCatalog(ctx context.Context) []domain.ReportCatalogItem
	GetTrialBalance(ctx context.Context, filter domain.ReportFilter) (*domain.TrialBalanceReport, error)
	GetGeneralLedger(ctx context.Context, filter domain.ReportFilter) (*domain.GeneralLedgerReport, error)
	GetBalanceSheet(ctx context.Context, filter domain.ReportFilter) (*domain.BalanceSheetReport, error)
	GetIncomeStatement(ctx context.Context, filter domain.ReportFilter) (*domain.IncomeStatementReport, error)
	GetCashFlow(ctx context.Context, filter domain.ReportFilter) (*domain.CashFlowReport, error)
	GetARAging(ctx context.Context, filter domain.ReportFilter) (*domain.AgingReport, error)
	GetAPAging(ctx context.Context, filter domain.ReportFilter) (*domain.AgingReport, error)
	GetCustomerStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error)
	GetSupplierStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error)
	GetInventoryValuation(ctx context.Context, filter domain.ReportFilter) (*domain.InventoryValuationReport, error)
	GetSalesReport(ctx context.Context, filter domain.ReportFilter) (*domain.SalesReport, error)
	GetPurchasingReport(ctx context.Context, filter domain.ReportFilter) (*domain.PurchasingReport, error)
	GetVATReturnReport(ctx context.Context, filter domain.ReportFilter) (*domain.VATReturnReport, error)
	GetPayrollSummary(ctx context.Context, filter domain.ReportFilter) (*domain.PayrollSummaryReport, error)
	GetGOSIContribution(ctx context.Context, filter domain.ReportFilter) (*domain.GOSIContributionReport, error)
	GetBudgetVsActual(ctx context.Context, filter domain.ReportFilter) (*domain.BudgetVsActualReport, error)

	ExportReportToCSV(ctx context.Context, reportCode string, filter domain.ReportFilter) ([]byte, string, error)
}
