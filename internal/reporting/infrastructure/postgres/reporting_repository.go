package postgres

import (
	"context"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/reporting/domain"
	"erp/internal/reporting/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type reportingRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

func NewReportingRepository(db *database.DB) ports.ReportingRepository {
	return &reportingRepository{
		db:      db,
		queries: sqlc.New(db.Pool),
	}
}

// Helpers
func toPgUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

func toPgDate(t *time.Time) pgtype.Date {
	if t == nil || t.IsZero() {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromPgNumeric(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	v, err := n.Value()
	if err != nil || v == nil {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(fmt.Sprintf("%v", v))
	if err != nil {
		return decimal.Zero
	}
	return d
}

// ============================================================================
// 1. TRIAL BALANCE
// ============================================================================

func (r *reportingRepository) GetTrialBalance(ctx context.Context, filter domain.ReportFilter) (*domain.TrialBalanceReport, error) {
	rows, err := r.queries.GetTrialBalanceReport(ctx, sqlc.GetTrialBalanceReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(filter.StartDate),
		Column3:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.TrialBalanceReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Rows:          make([]domain.TrialBalanceRow, 0, len(rows)),
		TotalDebits:   decimal.Zero,
		TotalCredits:  decimal.Zero,
	}

	for _, row := range rows {
		debit := fromPgNumeric(row.TotalDebit)
		credit := fromPgNumeric(row.TotalCredit)
		net := fromPgNumeric(row.NetBalance)

		report.TotalDebits = report.TotalDebits.Add(debit)
		report.TotalCredits = report.TotalCredits.Add(credit)

		report.Rows = append(report.Rows, domain.TrialBalanceRow{
			AccountID:     toUUID(row.AccountID),
			AccountCode:   row.AccountCode,
			AccountNameEn: row.AccountNameEn,
			AccountNameAr: row.AccountNameAr,
			AccountType:   row.AccountType,
			NormalBalance: row.NormalBalance,
			TotalDebit:    debit,
			TotalCredit:   credit,
			NetBalance:    net,
		})
	}

	report.IsBalanced = report.TotalDebits.Equal(report.TotalCredits)
	return report, nil
}

// ============================================================================
// 2. GENERAL LEDGER
// ============================================================================

func (r *reportingRepository) GetGeneralLedger(ctx context.Context, filter domain.ReportFilter) (*domain.GeneralLedgerReport, error) {
	rows, err := r.queries.GetGeneralLedgerReport(ctx, sqlc.GetGeneralLedgerReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgUUIDPtr(filter.AccountID),
		Column3:       toPgDate(filter.StartDate),
		Column4:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.GeneralLedgerReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		AccountID:     filter.AccountID,
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Entries:       make([]domain.GeneralLedgerEntry, 0, len(rows)),
		TotalDebit:    decimal.Zero,
		TotalCredit:   decimal.Zero,
	}

	for _, row := range rows {
		debit := fromPgNumeric(row.BaseDebit)
		credit := fromPgNumeric(row.BaseCredit)
		report.TotalDebit = report.TotalDebit.Add(debit)
		report.TotalCredit = report.TotalCredit.Add(credit)

		var ref *string
		if row.Reference.Valid {
			ref = &row.Reference.String
		}
		var desc *string
		if row.Description.Valid {
			desc = &row.Description.String
		}

		report.Entries = append(report.Entries, domain.GeneralLedgerEntry{
			EntryID:        toUUID(row.EntryID),
			EntryNumber:    row.EntryNumber,
			AccountingDate: row.AccountingDate.Time,
			SourceModule:   row.SourceModule,
			Reference:      ref,
			LineID:         toUUID(row.LineID),
			LineNumber:     int(row.LineNumber),
			AccountID:      toUUID(row.AccountID),
			AccountCode:    row.AccountCode,
			AccountNameEn:  row.AccountNameEn,
			AccountNameAr:  row.AccountNameAr,
			AccountType:    row.AccountType,
			Debit:          debit,
			Credit:         credit,
			Description:    desc,
		})
	}

	return report, nil
}

// ============================================================================
// 3. BALANCE SHEET
// ============================================================================

func (r *reportingRepository) GetBalanceSheet(ctx context.Context, filter domain.ReportFilter) (*domain.BalanceSheetReport, error) {
	asOf := filter.AsOfDate
	if asOf == nil {
		now := time.Now()
		asOf = &now
	}

	rows, err := r.queries.GetBalanceSheetAccounts(ctx, sqlc.GetBalanceSheetAccountsParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(asOf),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.BalanceSheetReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		AsOfDate:      *asOf,
		Assets:        []domain.BalanceSheetCategory{},
		Liabilities:   []domain.BalanceSheetCategory{},
		Equity:        []domain.BalanceSheetCategory{},
	}

	assetItems := make([]domain.BalanceSheetItem, 0)
	liabItems := make([]domain.BalanceSheetItem, 0)
	equityItems := make([]domain.BalanceSheetItem, 0)

	for _, row := range rows {
		bal := fromPgNumeric(row.Balance)
		item := domain.BalanceSheetItem{
			AccountID:     toUUID(row.AccountID),
			AccountCode:   row.AccountCode,
			AccountNameEn: row.AccountNameEn,
			AccountNameAr: row.AccountNameAr,
			Balance:       bal,
		}

		switch row.AccountType {
		case "ASSET":
			assetItems = append(assetItems, item)
			report.TotalAssets = report.TotalAssets.Add(bal)
		case "LIABILITY":
			liabItems = append(liabItems, item)
			report.TotalLiabilities = report.TotalLiabilities.Add(bal)
		case "EQUITY":
			equityItems = append(equityItems, item)
			report.TotalEquity = report.TotalEquity.Add(bal)
		}
	}

	// In standard financial accounting, Current Period Net Income (Revenues - Expenses)
	// forms part of Owner's Equity before closing entries into Retained Earnings.
	var netIncome decimal.Decimal
	netIncomeQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN a.account_type = 'REVENUE' THEN (jl.base_credit - jl.base_debit) ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN a.account_type = 'EXPENSE' THEN (jl.base_debit - jl.base_credit) ELSE 0 END), 0) AS net_income
		FROM accounting.accounts a
		JOIN accounting.journal_lines jl ON jl.account_id = a.id
		JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
		WHERE je.legal_entity_id = $1
		  AND je.status = 'POSTED'
		  AND ($2::DATE IS NULL OR je.accounting_date <= $2::DATE)
	`
	_ = r.db.Pool.QueryRow(ctx, netIncomeQuery, filter.LegalEntityID, asOf).Scan(&netIncome)

	if !netIncome.IsZero() {
		equityItems = append(equityItems, domain.BalanceSheetItem{
			AccountCode:   "CURRENT-P&L",
			AccountNameEn: "Current Period Profit / (Loss)",
			AccountNameAr: "أرباح / (خسائر) الفترة الحالية",
			Balance:       netIncome,
		})
		report.TotalEquity = report.TotalEquity.Add(netIncome)
	}

	report.Assets = []domain.BalanceSheetCategory{{
		CategoryNameEn: "Assets",
		CategoryNameAr: "الأصول",
		Items:          assetItems,
		Subtotal:       report.TotalAssets,
	}}
	report.Liabilities = []domain.BalanceSheetCategory{{
		CategoryNameEn: "Liabilities",
		CategoryNameAr: "الالتزامات",
		Items:          liabItems,
		Subtotal:       report.TotalLiabilities,
	}}
	report.Equity = []domain.BalanceSheetCategory{{
		CategoryNameEn: "Equity",
		CategoryNameAr: "حقوق الملكية",
		Items:          equityItems,
		Subtotal:       report.TotalEquity,
	}}

	report.TotalLiabEquity = report.TotalLiabilities.Add(report.TotalEquity)
	report.IsBalanced = report.TotalAssets.Equal(report.TotalLiabEquity)

	return report, nil
}

// ============================================================================
// 4. INCOME STATEMENT (P&L)
// ============================================================================

func (r *reportingRepository) GetIncomeStatement(ctx context.Context, filter domain.ReportFilter) (*domain.IncomeStatementReport, error) {
	rows, err := r.queries.GetIncomeStatementAccounts(ctx, sqlc.GetIncomeStatementAccountsParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(filter.StartDate),
		Column3:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.IncomeStatementReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		Revenues:      []domain.IncomeStatementCategory{},
		Expenses:      []domain.IncomeStatementCategory{},
	}
	if filter.StartDate != nil {
		report.StartDate = *filter.StartDate
	}
	if filter.EndDate != nil {
		report.EndDate = *filter.EndDate
	}

	revItems := make([]domain.IncomeStatementItem, 0)
	expItems := make([]domain.IncomeStatementItem, 0)

	for _, row := range rows {
		amt := fromPgNumeric(row.PeriodBalance)
		item := domain.IncomeStatementItem{
			AccountID:     toUUID(row.AccountID),
			AccountCode:   row.AccountCode,
			AccountNameEn: row.AccountNameEn,
			AccountNameAr: row.AccountNameAr,
			Amount:        amt,
		}

		if row.AccountType == "REVENUE" {
			revItems = append(revItems, item)
			report.TotalRevenue = report.TotalRevenue.Add(amt)
		} else if row.AccountType == "EXPENSE" {
			expItems = append(expItems, item)
			report.TotalExpenses = report.TotalExpenses.Add(amt)
		}
	}

	report.Revenues = []domain.IncomeStatementCategory{{
		CategoryNameEn: "Operating Revenues",
		CategoryNameAr: "الإيرادات التشغيلية",
		Items:          revItems,
		TotalAmount:    report.TotalRevenue,
	}}
	report.Expenses = []domain.IncomeStatementCategory{{
		CategoryNameEn: "Operating Expenses",
		CategoryNameAr: "المصروفات التشغيلية",
		Items:          expItems,
		TotalAmount:    report.TotalExpenses,
	}}

	report.NetIncome = report.TotalRevenue.Sub(report.TotalExpenses)
	if report.TotalRevenue.IsPositive() {
		report.ProfitMarginRate = report.NetIncome.Div(report.TotalRevenue).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return report, nil
}

// ============================================================================
// 5. CASH FLOW
// ============================================================================

func (r *reportingRepository) GetCashFlow(ctx context.Context, filter domain.ReportFilter) (*domain.CashFlowReport, error) {
	// Query cash movements from receipts and payments
	query := `
		SELECT
			'RECEIPT' as movement_type,
			r.receipt_type,
			r.received_from,
			r.amount
		FROM treasury.receipts r
		WHERE r.legal_entity_id = $1
		  AND r.status = 'POSTED'
		  AND ($2::DATE IS NULL OR r.receipt_date >= $2::DATE)
		  AND ($3::DATE IS NULL OR r.receipt_date <= $3::DATE)
		UNION ALL
		SELECT
			'PAYMENT' as movement_type,
			p.payment_type,
			p.paid_to,
			p.amount
		FROM treasury.payments p
		WHERE p.legal_entity_id = $1
		  AND p.status = 'PAID'
		  AND ($2::DATE IS NULL OR p.payment_date >= $2::DATE)
		  AND ($3::DATE IS NULL OR p.payment_date <= $3::DATE)
	`
	rows, err := r.db.Pool.Query(ctx, query, filter.LegalEntityID, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &domain.CashFlowReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		OperatingActivities: domain.CashFlowSection{
			SectionEn: "Cash Flows from Operating Activities",
			SectionAr: "التدفقات النقدية من الأنشطة التشغيلية",
			Items:     []domain.CashFlowItem{},
		},
		InvestingActivities: domain.CashFlowSection{
			SectionEn: "Cash Flows from Investing Activities",
			SectionAr: "التدفقات النقدية من الأنشطة الاستثمارية",
			Items:     []domain.CashFlowItem{},
		},
		FinancingActivities: domain.CashFlowSection{
			SectionEn: "Cash Flows from Financing Activities",
			SectionAr: "التدفقات النقدية من الأنشطة التمويلية",
			Items:     []domain.CashFlowItem{},
		},
	}
	if filter.StartDate != nil {
		report.StartDate = *filter.StartDate
	}
	if filter.EndDate != nil {
		report.EndDate = *filter.EndDate
	}

	for rows.Next() {
		var movType, txType, counterparty string
		var amt decimal.Decimal
		if err := rows.Scan(&movType, &txType, &counterparty, &amt); err != nil {
			return nil, err
		}

		if movType == "RECEIPT" {
			report.OperatingActivities.Items = append(report.OperatingActivities.Items, domain.CashFlowItem{
				TitleEn: fmt.Sprintf("Cash Receipt: %s (%s)", txType, counterparty),
				TitleAr: fmt.Sprintf("سند قبض نقدي: %s (%s)", txType, counterparty),
				Amount:  amt,
			})
			report.OperatingActivities.NetTotal = report.OperatingActivities.NetTotal.Add(amt)
		} else {
			report.OperatingActivities.Items = append(report.OperatingActivities.Items, domain.CashFlowItem{
				TitleEn: fmt.Sprintf("Cash Payment: %s (%s)", txType, counterparty),
				TitleAr: fmt.Sprintf("سند صرف نقدي: %s (%s)", txType, counterparty),
				Amount:  amt.Neg(),
			})
			report.OperatingActivities.NetTotal = report.OperatingActivities.NetTotal.Sub(amt)
		}
	}

	report.NetCashChange = report.OperatingActivities.NetTotal.
		Add(report.InvestingActivities.NetTotal).
		Add(report.FinancingActivities.NetTotal)

	return report, nil
}

// ============================================================================
// 6 & 7. AGING REPORTS (AR & AP)
// ============================================================================

func (r *reportingRepository) GetAgingReport(ctx context.Context, reportType string, filter domain.ReportFilter) (*domain.AgingReport, error) {
	report := &domain.AgingReport{
		ReportType:    reportType,
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		Rows:          []domain.AgingBucketRow{},
	}

	if reportType == "AR" {
		rows, err := r.queries.GetCustomerAgingReport(ctx, sqlc.GetCustomerAgingReportParams{
			LegalEntityID: toPgUUID(filter.LegalEntityID),
			Column2:       toPgUUIDPtr(filter.CustomerID),
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			tot := fromPgNumeric(row.TotalOutstanding)
			curr := fromPgNumeric(row.CurrentAmount)
			a1 := fromPgNumeric(row.Aging130)
			a2 := fromPgNumeric(row.Aging3160)
			a3 := fromPgNumeric(row.Aging6190)
			a4 := fromPgNumeric(row.Aging91120)
			a5 := fromPgNumeric(row.AgingOver120)

			report.TotalOutstanding = report.TotalOutstanding.Add(tot)
			report.TotalCurrent = report.TotalCurrent.Add(curr)
			report.Total1To30 = report.Total1To30.Add(a1)
			report.Total31To60 = report.Total31To60.Add(a2)
			report.Total61To90 = report.Total61To90.Add(a3)
			report.Total91To120 = report.Total91To120.Add(a4)
			report.TotalOver120 = report.TotalOver120.Add(a5)

			report.Rows = append(report.Rows, domain.AgingBucketRow{
				EntityID:         toUUID(row.CustomerID),
				EntityNumber:     row.CustomerNumber,
				EntityNameEn:     row.CustomerNameEn,
				EntityNameAr:     row.CustomerNameAr,
				DocumentCount:    int(row.InvoiceCount),
				TotalOutstanding: tot,
				CurrentAmount:    curr,
				Aging1To30:       a1,
				Aging31To60:      a2,
				Aging61To90:      a3,
				Aging91To120:     a4,
				AgingOver120:     a5,
			})
		}
	} else {
		rows, err := r.queries.GetSupplierAgingReport(ctx, sqlc.GetSupplierAgingReportParams{
			LegalEntityID: toPgUUID(filter.LegalEntityID),
			Column2:       toPgUUIDPtr(filter.SupplierID),
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			tot := fromPgNumeric(row.TotalOutstanding)
			curr := fromPgNumeric(row.CurrentAmount)
			a1 := fromPgNumeric(row.Aging130)
			a2 := fromPgNumeric(row.Aging3160)
			a3 := fromPgNumeric(row.Aging6190)
			a4 := fromPgNumeric(row.Aging91120)
			a5 := fromPgNumeric(row.AgingOver120)

			report.TotalOutstanding = report.TotalOutstanding.Add(tot)
			report.TotalCurrent = report.TotalCurrent.Add(curr)
			report.Total1To30 = report.Total1To30.Add(a1)
			report.Total31To60 = report.Total31To60.Add(a2)
			report.Total61To90 = report.Total61To90.Add(a3)
			report.Total91To120 = report.Total91To120.Add(a4)
			report.TotalOver120 = report.TotalOver120.Add(a5)

			report.Rows = append(report.Rows, domain.AgingBucketRow{
				EntityID:         toUUID(row.SupplierID),
				EntityNumber:     row.SupplierNumber,
				EntityNameEn:     row.SupplierNameEn,
				EntityNameAr:     row.SupplierNameAr,
				DocumentCount:    int(row.BillCount),
				TotalOutstanding: tot,
				CurrentAmount:    curr,
				Aging1To30:       a1,
				Aging31To60:      a2,
				Aging61To90:      a3,
				Aging91To120:     a4,
				AgingOver120:     a5,
			})
		}
	}

	return report, nil
}

// ============================================================================
// 8. CUSTOMER STATEMENT
// ============================================================================

func (r *reportingRepository) GetCustomerStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error) {
	if filter.CustomerID == nil {
		return nil, domain.ErrEntityIDRequired
	}

	var nameEn, nameAr, num string
	err := r.db.Pool.QueryRow(ctx, "SELECT customer_number, name_en, name_ar FROM receivables.customers WHERE id = $1", *filter.CustomerID).
		Scan(&num, &nameEn, &nameAr)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetCustomerStatementLines(ctx, sqlc.GetCustomerStatementLinesParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		CustomerID:    toPgUUID(*filter.CustomerID),
		Column3:       toPgDate(filter.StartDate),
		Column4:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.EntityStatementReport{
		ReportType:    "CUSTOMER",
		LegalEntityID: filter.LegalEntityID,
		EntityID:      *filter.CustomerID,
		EntityNumber:  num,
		EntityNameEn:  nameEn,
		EntityNameAr:  nameAr,
		GeneratedAt:   time.Now(),
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Lines:         make([]domain.StatementLineItem, 0, len(rows)),
	}

	for _, row := range rows {
		orig := fromPgNumeric(row.OriginalAmount)
		alloc := fromPgNumeric(row.AllocatedAmount)
		rem := fromPgNumeric(row.RemainingAmount)

		report.TotalDebited = report.TotalDebited.Add(orig)
		report.TotalCredited = report.TotalCredited.Add(alloc)

		report.Lines = append(report.Lines, domain.StatementLineItem{
			ItemID:          toUUID(row.ItemID),
			DocumentType:    row.DocumentType,
			DocumentNumber:  row.DocumentNumber,
			DocumentDate:    row.DocumentDate.Time,
			DueDate:         row.DueDate.Time,
			CurrencyCode:    row.CurrencyCode,
			OriginalAmount:  orig,
			AllocatedAmount: alloc,
			RemainingAmount: rem,
			Status:          row.Status,
		})
	}

	report.EndingBalance = report.TotalDebited.Sub(report.TotalCredited)
	return report, nil
}

// ============================================================================
// 9. SUPPLIER STATEMENT
// ============================================================================

func (r *reportingRepository) GetSupplierStatement(ctx context.Context, filter domain.ReportFilter) (*domain.EntityStatementReport, error) {
	if filter.SupplierID == nil {
		return nil, domain.ErrEntityIDRequired
	}

	var nameEn, nameAr, num string
	err := r.db.Pool.QueryRow(ctx, "SELECT supplier_number, name_en, name_ar FROM payables.suppliers WHERE id = $1", *filter.SupplierID).
		Scan(&num, &nameEn, &nameAr)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetSupplierStatementLines(ctx, sqlc.GetSupplierStatementLinesParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		SupplierID:    toPgUUID(*filter.SupplierID),
		Column3:       toPgDate(filter.StartDate),
		Column4:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.EntityStatementReport{
		ReportType:    "SUPPLIER",
		LegalEntityID: filter.LegalEntityID,
		EntityID:      *filter.SupplierID,
		EntityNumber:  num,
		EntityNameEn:  nameEn,
		EntityNameAr:  nameAr,
		GeneratedAt:   time.Now(),
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Lines:         make([]domain.StatementLineItem, 0, len(rows)),
	}

	for _, row := range rows {
		orig := fromPgNumeric(row.OriginalAmount)
		alloc := fromPgNumeric(row.AllocatedAmount)
		rem := fromPgNumeric(row.RemainingAmount)

		report.TotalDebited = report.TotalDebited.Add(orig)
		report.TotalCredited = report.TotalCredited.Add(alloc)

		report.Lines = append(report.Lines, domain.StatementLineItem{
			ItemID:          toUUID(row.ItemID),
			DocumentType:    row.DocumentType,
			DocumentNumber:  row.DocumentNumber,
			DocumentDate:    row.DocumentDate.Time,
			DueDate:         row.DueDate.Time,
			CurrencyCode:    row.CurrencyCode,
			OriginalAmount:  orig,
			AllocatedAmount: alloc,
			RemainingAmount: rem,
			Status:          row.Status,
		})
	}

	report.EndingBalance = report.TotalDebited.Sub(report.TotalCredited)
	return report, nil
}

// ============================================================================
// 10. INVENTORY VALUATION
// ============================================================================

func (r *reportingRepository) GetInventoryValuation(ctx context.Context, filter domain.ReportFilter) (*domain.InventoryValuationReport, error) {
	rows, err := r.queries.GetInventoryValuationReport(ctx, sqlc.GetInventoryValuationReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgUUIDPtr(filter.WarehouseID),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.InventoryValuationReport{
		LegalEntityID: filter.LegalEntityID,
		WarehouseID:   filter.WarehouseID,
		GeneratedAt:   time.Now(),
		Items:         make([]domain.InventoryValuationRow, 0, len(rows)),
		TotalItems:    len(rows),
	}

	for _, row := range rows {
		onHand := fromPgNumeric(row.QuantityOnHand)
		res := fromPgNumeric(row.QuantityReserved)
		avail := fromPgNumeric(row.QuantityAvailable)
		cost := fromPgNumeric(row.AverageCost)
		val := fromPgNumeric(row.TotalValue)

		report.TotalQuantity = report.TotalQuantity.Add(onHand)
		report.TotalValuation = report.TotalValuation.Add(val)

		report.Items = append(report.Items, domain.InventoryValuationRow{
			ItemID:            toUUID(row.ItemID),
			ItemCode:          row.ItemCode,
			ItemNameEn:        row.ItemNameEn,
			ItemNameAr:        row.ItemNameAr,
			WarehouseID:       toUUID(row.WarehouseID),
			WarehouseCode:     row.WarehouseCode,
			WarehouseNameEn:   row.WarehouseNameEn,
			WarehouseNameAr:   row.WarehouseNameAr,
			QuantityOnHand:    onHand,
			QuantityReserved:  res,
			QuantityAvailable: avail,
			AverageCost:       cost,
			TotalValue:        val,
		})
	}

	return report, nil
}

// ============================================================================
// 11. SALES REPORT
// ============================================================================

func (r *reportingRepository) GetSalesReport(ctx context.Context, filter domain.ReportFilter) (*domain.SalesReport, error) {
	statusStr := ""
	if filter.Status != nil {
		statusStr = *filter.Status
	}

	rows, err := r.queries.GetSalesPeriodReport(ctx, sqlc.GetSalesPeriodReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(filter.StartDate),
		Column3:       toPgDate(filter.EndDate),
		Column4:       statusStr,
	})
	if err != nil {
		return nil, err
	}

	report := &domain.SalesReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Invoices:      make([]domain.SalesInvoiceRow, 0, len(rows)),
		TotalCount:    len(rows),
	}

	for _, row := range rows {
		sub := fromPgNumeric(row.Subtotal)
		tax := fromPgNumeric(row.TaxAmount)
		tot := fromPgNumeric(row.TotalAmount)

		report.TotalSubtotal = report.TotalSubtotal.Add(sub)
		report.TotalTax = report.TotalTax.Add(tax)
		report.TotalSales = report.TotalSales.Add(tot)

		report.Invoices = append(report.Invoices, domain.SalesInvoiceRow{
			InvoiceID:      toUUID(row.InvoiceID),
			InvoiceNumber:  row.InvoiceNumber,
			IssueDate:      row.IssueDate.Time,
			DueDate:        row.DueDate.Time,
			CustomerNameEn: row.CustomerNameEn,
			CustomerNameAr: row.CustomerNameAr,
			Subtotal:       sub,
			TaxAmount:      tax,
			TotalAmount:    tot,
			Status:         row.Status,
		})
	}

	return report, nil
}

// ============================================================================
// 12. PURCHASING REPORT
// ============================================================================

func (r *reportingRepository) GetPurchasingReport(ctx context.Context, filter domain.ReportFilter) (*domain.PurchasingReport, error) {
	statusStr := ""
	if filter.Status != nil {
		statusStr = *filter.Status
	}

	rows, err := r.queries.GetPurchasingPeriodReport(ctx, sqlc.GetPurchasingPeriodReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(filter.StartDate),
		Column3:       toPgDate(filter.EndDate),
		Column4:       statusStr,
	})
	if err != nil {
		return nil, err
	}

	report := &domain.PurchasingReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Orders:        make([]domain.PurchasingOrderRow, 0, len(rows)),
		TotalCount:    len(rows),
	}

	for _, row := range rows {
		sub := fromPgNumeric(row.Subtotal)
		tax := fromPgNumeric(row.TaxAmount)
		tot := fromPgNumeric(row.TotalAmount)

		report.TotalSubtotal = report.TotalSubtotal.Add(sub)
		report.TotalTax = report.TotalTax.Add(tax)
		report.TotalPurchases = report.TotalPurchases.Add(tot)

		report.Orders = append(report.Orders, domain.PurchasingOrderRow{
			POID:           toUUID(row.PoID),
			PONumber:       row.PoNumber,
			OrderDate:      row.OrderDate.Time,
			SupplierNumber: row.SupplierNumber,
			SupplierNameEn: row.SupplierNameEn,
			SupplierNameAr: row.SupplierNameAr,
			Subtotal:       sub,
			TaxAmount:      tax,
			TotalAmount:    tot,
			Status:         row.Status,
		})
	}

	return report, nil
}

// ============================================================================
// 13. VAT RETURN (ZATCA 15%)
// ============================================================================

func (r *reportingRepository) GetVATReturnReport(ctx context.Context, filter domain.ReportFilter) (*domain.VATReturnReport, error) {
	summary, err := r.queries.GetVATSummaryReport(ctx, sqlc.GetVATSummaryReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(filter.StartDate),
		Column3:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	salesAmt := fromPgNumeric(summary.StandardRatedSales)
	outputVAT := fromPgNumeric(summary.OutputVatAmount)
	purchasesAmt := fromPgNumeric(summary.StandardRatedPurchases)
	inputVAT := fromPgNumeric(summary.InputVatAmount)
	netVAT := fromPgNumeric(summary.NetVatPayable)

	salesBoxes := []domain.VATReturnBox{
		{
			BoxNumber:     1,
			DescriptionEn: "Standard rated sales (15%)",
			DescriptionAr: "المبيعات الخاضعة للنسبة الأساسية (15%)",
			TaxableAmount: salesAmt,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     outputVAT,
		},
		{
			BoxNumber:     2,
			DescriptionEn: "Sales to registered customers in GCC",
			DescriptionAr: "المبيعات للمواطنين المسجلين بدول الخليج",
			TaxableAmount: decimal.Zero,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     decimal.Zero,
		},
		{
			BoxNumber:     3,
			DescriptionEn: "Zero rated sales / Exports",
			DescriptionAr: "المبيعات الخاضعة لنسبة الصفر / الصادرات",
			TaxableAmount: decimal.Zero,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     decimal.Zero,
		},
		{
			BoxNumber:     4,
			DescriptionEn: "Exempt sales",
			DescriptionAr: "المبيعات المعفاة من الضريبة",
			TaxableAmount: decimal.Zero,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     decimal.Zero,
		},
	}

	purchaseBoxes := []domain.VATReturnBox{
		{
			BoxNumber:     7,
			DescriptionEn: "Standard rated domestic purchases (15%)",
			DescriptionAr: "المشتريات المحلية الخاضعة للنسبة الأساسية (15%)",
			TaxableAmount: purchasesAmt,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     inputVAT,
		},
		{
			BoxNumber:     8,
			DescriptionEn: "Imports subject to VAT paid at customs",
			DescriptionAr: "الواردات الخاضعة لضريبة القيمة المضافة المدفوعة بالجمارك",
			TaxableAmount: decimal.Zero,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     decimal.Zero,
		},
		{
			BoxNumber:     9,
			DescriptionEn: "Zero rated purchases",
			DescriptionAr: "المشتريات الخاضعة لنسبة الصفر",
			TaxableAmount: decimal.Zero,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     decimal.Zero,
		},
		{
			BoxNumber:     10,
			DescriptionEn: "Exempt purchases",
			DescriptionAr: "المشتريات المعفاة من الضريبة",
			TaxableAmount: decimal.Zero,
			AdjustmentAmt: decimal.Zero,
			VATAmount:     decimal.Zero,
		},
	}

	return &domain.VATReturnReport{
		LegalEntityID:         filter.LegalEntityID,
		GeneratedAt:           time.Now(),
		StartDate:             filter.StartDate,
		EndDate:               filter.EndDate,
		SalesBoxes:            salesBoxes,
		PurchaseBoxes:         purchaseBoxes,
		TotalOutputVAT:        outputVAT,
		TotalInputVAT:         inputVAT,
		NetVATDueOrRefundable: netVAT,
	}, nil
}

// ============================================================================
// 14. PAYROLL SUMMARY
// ============================================================================

func (r *reportingRepository) GetPayrollSummary(ctx context.Context, filter domain.ReportFilter) (*domain.PayrollSummaryReport, error) {
	rows, err := r.queries.GetPayrollRunsReport(ctx, sqlc.GetPayrollRunsReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgDate(filter.StartDate),
		Column3:       toPgDate(filter.EndDate),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.PayrollSummaryReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		Runs:          make([]domain.PayrollRunRow, 0, len(rows)),
	}

	for _, row := range rows {
		basic := fromPgNumeric(row.TotalBasic)
		allow := fromPgNumeric(row.TotalAllowances)
		gross := fromPgNumeric(row.TotalGross)
		ded := fromPgNumeric(row.TotalDeductions)
		net := fromPgNumeric(row.TotalNet)

		report.TotalGrossPaid = report.TotalGrossPaid.Add(gross)
		report.TotalDeductions = report.TotalDeductions.Add(ded)
		report.TotalNetPaid = report.TotalNetPaid.Add(net)

		report.Runs = append(report.Runs, domain.PayrollRunRow{
			RunID:           toUUID(row.RunID),
			RunNumber:       row.RunNumber,
			PeriodID:        toUUID(row.PeriodID),
			PeriodNameEn:    row.PeriodNameEn,
			PeriodNameAr:    row.PeriodNameAr,
			RunDate:         row.RunDate.Time,
			Status:          row.Status,
			EmployeeCount:   int(row.EmployeeCount),
			TotalBasic:      basic,
			TotalAllowances: allow,
			TotalGross:      gross,
			TotalDeductions: ded,
			TotalNet:        net,
		})
	}

	return report, nil
}

// ============================================================================
// 15. GOSI CONTRIBUTIONS
// ============================================================================

func (r *reportingRepository) GetGOSIContribution(ctx context.Context, filter domain.ReportFilter) (*domain.GOSIContributionReport, error) {
	var y, m int32
	if filter.Year != nil {
		y = int32(*filter.Year)
	}
	if filter.Month != nil {
		m = int32(*filter.Month)
	}

	rows, err := r.queries.GetGOSIContributionsReport(ctx, sqlc.GetGOSIContributionsReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       y,
		Column3:       m,
	})
	if err != nil {
		return nil, err
	}

	report := &domain.GOSIContributionReport{
		LegalEntityID: filter.LegalEntityID,
		GeneratedAt:   time.Now(),
		Year:          filter.Year,
		Month:         filter.Month,
		Calculations:  make([]domain.GOSICalculationRow, 0, len(rows)),
	}

	for _, row := range rows {
		basis := fromPgNumeric(row.ContributionBasis)
		empShare := fromPgNumeric(row.TotalEmployeeAmount)
		emprShare := fromPgNumeric(row.TotalEmployerAmount)
		totContrib := fromPgNumeric(row.TotalAmount)

		report.TotalEmployees++
		if row.IsSaudiNational {
			report.TotalSaudiCount++
		} else {
			report.TotalNonSaudiCount++
		}

		report.TotalContributoryWages = report.TotalContributoryWages.Add(basis)
		report.TotalEmployeeDeduction = report.TotalEmployeeDeduction.Add(empShare)
		report.TotalEmployerPayment = report.TotalEmployerPayment.Add(emprShare)
		report.TotalRemittanceDue = report.TotalRemittanceDue.Add(totContrib)

		nameEn := fmt.Sprintf("%v", row.EmployeeNameEn)
		nameAr := fmt.Sprintf("%v", row.EmployeeNameAr)

		report.Calculations = append(report.Calculations, domain.GOSICalculationRow{
			CalculationID:      toUUID(row.CalculationID),
			CalculationYear:    int(row.CalculationYear),
			CalculationMonth:   int(row.CalculationMonth),
			Status:             row.Status,
			EmployeeNumber:     row.EmployeeNumber,
			EmployeeNameEn:     nameEn,
			EmployeeNameAr:     nameAr,
			IsSaudiNational:    row.IsSaudiNational,
			ContributionBasis:  basis,
			TotalEmployeeShare: empShare,
			TotalEmployerShare: emprShare,
			TotalContribution:  totContrib,
			CreatedAt:          row.CreatedAt.Time,
		})
	}

	return report, nil
}

// ============================================================================
// 16. BUDGET VS ACTUAL
// ============================================================================

func (r *reportingRepository) GetBudgetVsActual(ctx context.Context, filter domain.ReportFilter) (*domain.BudgetVsActualReport, error) {
	rows, err := r.queries.GetBudgetVsActualReport(ctx, sqlc.GetBudgetVsActualReportParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       toPgUUIDPtr(filter.BudgetID),
	})
	if err != nil {
		return nil, err
	}

	report := &domain.BudgetVsActualReport{
		LegalEntityID: filter.LegalEntityID,
		BudgetID:      filter.BudgetID,
		GeneratedAt:   time.Now(),
		Variances:     make([]domain.BudgetVarianceRow, 0, len(rows)),
	}

	for _, row := range rows {
		bAmt := fromPgNumeric(row.BudgetedAmount)
		aAmt := fromPgNumeric(row.ActualAmount)
		vAmt := fromPgNumeric(row.VarianceAmount)
		vPct := decimal.NewFromFloat(row.VariancePercentage)

		report.TotalBudgeted = report.TotalBudgeted.Add(bAmt)
		report.TotalActual = report.TotalActual.Add(aAmt)
		report.TotalNetVariance = report.TotalNetVariance.Add(vAmt)

		report.Variances = append(report.Variances, domain.BudgetVarianceRow{
			BudgetID:           toUUID(row.BudgetID),
			BudgetNameEn:       row.BudgetNameEn,
			BudgetNameAr:       row.BudgetNameAr,
			FiscalYearName:     row.FiscalYearName,
			AccountID:          toUUID(row.AccountID),
			AccountCode:        row.AccountCode,
			AccountNameEn:      row.AccountNameEn,
			AccountNameAr:      row.AccountNameAr,
			AccountType:        row.AccountType,
			BudgetedAmount:     bAmt,
			ActualAmount:       aAmt,
			VarianceAmount:     vAmt,
			VariancePercentage: vPct,
		})
	}

	return report, nil
}
