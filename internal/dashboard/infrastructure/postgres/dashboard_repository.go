package postgres

import (
	"context"
	"math"

	"erp/internal/dashboard/domain"
	"erp/internal/dashboard/ports"
	"erp/internal/platform/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type dashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(db *database.DB) ports.DashboardRepository {
	return &dashboardRepository{
		pool: db.Pool,
	}
}

func (r *dashboardRepository) GetLegalEntityInfo(ctx context.Context, legalEntityID uuid.UUID) (uuid.UUID, string, string, error) {
	var id uuid.UUID
	var name string
	var currency string

	if legalEntityID != uuid.Nil {
		query := `SELECT id, name_ar, base_currency FROM organization.legal_entities WHERE id = $1`
		err := r.pool.QueryRow(ctx, query, legalEntityID).Scan(&id, &name, &currency)
		if err == nil {
			return id, name, currency, nil
		}
	}

	fallbackQuery := `SELECT id, name_ar, base_currency FROM organization.legal_entities WHERE is_active = true ORDER BY created_at ASC LIMIT 1`
	err := r.pool.QueryRow(ctx, fallbackQuery).Scan(&id, &name, &currency)
	if err != nil {
		return uuid.Nil, "الشركة الرئيسية", "SAR", nil
	}
	return id, name, currency, nil
}

func (r *dashboardRepository) GetFinancialKPIs(ctx context.Context, legalEntityID uuid.UUID) (*domain.FinancialKPIs, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN a.account_type = 'REVENUE' THEN jl.base_credit - jl.base_debit ELSE 0 END), 0) AS total_revenue,
			COALESCE(SUM(CASE WHEN a.account_type = 'EXPENSE' THEN jl.base_debit - jl.base_credit ELSE 0 END), 0) AS total_expense
		FROM accounting.journal_lines jl
		JOIN accounting.accounts a ON a.id = jl.account_id
		JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
		WHERE jl.legal_entity_id = $1 AND je.status = 'POSTED'
	`

	var revDec, expDec decimal.Decimal
	err := r.pool.QueryRow(ctx, query, legalEntityID).Scan(&revDec, &expDec)
	if err != nil {
		return &domain.FinancialKPIs{Currency: "SAR"}, nil
	}

	rev := revDec.InexactFloat64()
	exp := expDec.InexactFloat64()
	netProfit := rev - exp

	margin := 0.0
	if rev > 0 {
		margin = math.Round((netProfit/rev)*10000) / 100 // 2 decimal places
	}

	return &domain.FinancialKPIs{
		TotalRevenue: math.Round(rev*100) / 100,
		TotalExpense: math.Round(exp*100) / 100,
		NetProfit:    math.Round(netProfit*100) / 100,
		ProfitMargin: margin,
		Currency:     "SAR",
	}, nil
}

func (r *dashboardRepository) GetCashPosition(ctx context.Context, legalEntityID uuid.UUID) (*domain.CashPositionKPI, error) {
	query := `
		SELECT 
			(SELECT COALESCE(SUM(current_balance), 0) FROM treasury.cash_accounts WHERE legal_entity_id = $1 AND status = 'ACTIVE') AS cash_balance,
			(SELECT COALESCE(SUM(current_balance), 0) FROM banking.bank_accounts WHERE legal_entity_id = $1 AND status = 'ACTIVE') AS bank_balance,
			(SELECT COUNT(*) FROM treasury.cash_accounts WHERE legal_entity_id = $1 AND status = 'ACTIVE') +
			(SELECT COUNT(*) FROM banking.bank_accounts WHERE legal_entity_id = $1 AND status = 'ACTIVE') AS total_accounts
	`

	var cashDec, bankDec decimal.Decimal
	var accountsCount int
	err := r.pool.QueryRow(ctx, query, legalEntityID).Scan(&cashDec, &bankDec, &accountsCount)
	if err != nil {
		return &domain.CashPositionKPI{Currency: "SAR"}, nil
	}

	cash := cashDec.InexactFloat64()
	bank := bankDec.InexactFloat64()

	return &domain.CashPositionKPI{
		TotalCashBalance: math.Round(cash*100) / 100,
		TotalBankBalance: math.Round(bank*100) / 100,
		TotalLiquidFunds: math.Round((cash+bank)*100) / 100,
		AccountsCount:    accountsCount,
		Currency:         "SAR",
	}, nil
}

func (r *dashboardRepository) GetARAPSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ARAPSummary, error) {
	arQuery := `
		SELECT 
			COALESCE(SUM(remaining_amount), 0) AS total_receivables,
			COALESCE(SUM(CASE WHEN due_date < CURRENT_DATE THEN remaining_amount ELSE 0 END), 0) AS overdue_receivables,
			COUNT(*) AS open_docs
		FROM receivables.open_items
		WHERE legal_entity_id = $1 AND status IN ('OPEN', 'PARTIALLY_PAID')
	`

	apQuery := `
		SELECT 
			COALESCE(SUM(remaining_amount), 0) AS total_payables,
			COALESCE(SUM(CASE WHEN due_date < CURRENT_DATE THEN remaining_amount ELSE 0 END), 0) AS overdue_payables,
			COUNT(*) AS open_docs
		FROM payables.open_items
		WHERE legal_entity_id = $1 AND status IN ('OPEN', 'PARTIALLY_PAID')
	`

	var totalArDec, overdueArDec decimal.Decimal
	var openArDocs int
	_ = r.pool.QueryRow(ctx, arQuery, legalEntityID).Scan(&totalArDec, &overdueArDec, &openArDocs)

	var totalApDec, overdueApDec decimal.Decimal
	var openApDocs int
	_ = r.pool.QueryRow(ctx, apQuery, legalEntityID).Scan(&totalApDec, &overdueApDec, &openApDocs)

	return &domain.ARAPSummary{
		TotalReceivables:   math.Round(totalArDec.InexactFloat64()*100) / 100,
		OverdueReceivables: math.Round(overdueArDec.InexactFloat64()*100) / 100,
		OpenReceivableDocs: openArDocs,
		TotalPayables:      math.Round(totalApDec.InexactFloat64()*100) / 100,
		OverduePayables:    math.Round(overdueApDec.InexactFloat64()*100) / 100,
		OpenPayableDocs:    openApDocs,
		Currency:           "SAR",
	}, nil
}

func (r *dashboardRepository) GetStockWarnings(ctx context.Context, legalEntityID uuid.UUID, limit int) ([]domain.StockWarning, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			sb.item_id,
			p.code,
			p.name_en,
			p.name_ar,
			w.name_en,
			w.name_ar,
			sb.quantity_on_hand,
			sb.quantity_reserved,
			sb.quantity_available
		FROM inventory.stock_balances sb
		JOIN products.items p ON p.id = sb.item_id
		JOIN warehouse.warehouses w ON w.id = sb.warehouse_id
		WHERE sb.legal_entity_id = $1 AND sb.quantity_available <= 10
		ORDER BY sb.quantity_available ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, legalEntityID, limit)
	if err != nil {
		return []domain.StockWarning{}, nil
	}
	defer rows.Close()

	var warnings []domain.StockWarning
	for rows.Next() {
		var itemID uuid.UUID
		var code, nameEn, nameAr, whEn, whAr string
		var onHand, reserved, available decimal.Decimal

		if err := rows.Scan(&itemID, &code, &nameEn, &nameAr, &whEn, &whAr, &onHand, &reserved, &available); err != nil {
			continue
		}

		availFloat := available.InexactFloat64()
		severity := "LOW"
		if availFloat <= 0 {
			severity = "CRITICAL"
		} else if availFloat <= 5 {
			severity = "WARNING"
		}

		warnings = append(warnings, domain.StockWarning{
			ItemID:            itemID,
			ItemCode:          code,
			ItemNameEn:        nameEn,
			ItemNameAr:        nameAr,
			WarehouseNameEn:   whEn,
			WarehouseNameAr:   whAr,
			QuantityOnHand:    onHand.InexactFloat64(),
			QuantityReserved:  reserved.InexactFloat64(),
			QuantityAvailable: availFloat,
			ReorderPoint:      10.0,
			Severity:          severity,
		})
	}

	if warnings == nil {
		warnings = []domain.StockWarning{}
	}
	return warnings, nil
}

func (r *dashboardRepository) GetPendingApprovals(ctx context.Context, legalEntityID uuid.UUID, limit int) (*domain.ApprovalsSummary, error) {
	if limit <= 0 {
		limit = 10
	}

	// 1. Group counts by resource_type
	countQuery := `
		SELECT wi.resource_type, COUNT(*)
		FROM workflow.approval_requests ar
		JOIN workflow.instance_steps step ON step.id = ar.instance_step_id
		JOIN workflow.instances wi ON wi.id = step.instance_id
		WHERE wi.legal_entity_id = $1 AND ar.status = 'PENDING'
		GROUP BY wi.resource_type
	`
	byType := make(map[string]int)
	totalPending := 0

	rows, err := r.pool.Query(ctx, countQuery, legalEntityID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rType string
			var cnt int
			if err := rows.Scan(&rType, &cnt); err == nil {
				byType[rType] = cnt
				totalPending += cnt
			}
		}
	}

	// 2. Recent pending requests
	itemsQuery := `
		SELECT 
			ar.id,
			wi.resource_type,
			wi.resource_id,
			wi.initiator_id,
			ar.assigned_user_id,
			ar.status,
			ar.created_at
		FROM workflow.approval_requests ar
		JOIN workflow.instance_steps step ON step.id = ar.instance_step_id
		JOIN workflow.instances wi ON wi.id = step.instance_id
		WHERE wi.legal_entity_id = $1 AND ar.status = 'PENDING'
		ORDER BY ar.created_at DESC
		LIMIT $2
	`

	var recent []domain.PendingApprovalItem
	reqRows, err := r.pool.Query(ctx, itemsQuery, legalEntityID, limit)
	if err == nil {
		defer reqRows.Close()
		for reqRows.Next() {
			var item domain.PendingApprovalItem
			if err := reqRows.Scan(
				&item.ID,
				&item.ResourceType,
				&item.ResourceID,
				&item.InitiatorID,
				&item.AssignedUserID,
				&item.Status,
				&item.CreatedAt,
			); err == nil {
				recent = append(recent, item)
			}
		}
	}

	if recent == nil {
		recent = []domain.PendingApprovalItem{}
	}

	return &domain.ApprovalsSummary{
		TotalPending:   totalPending,
		ByType:         byType,
		RecentRequests: recent,
	}, nil
}

func (r *dashboardRepository) GetZATCASummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ZATCAStatusSummary, error) {
	query := `
		SELECT
			COUNT(*) AS total_documents,
			COALESCE(COUNT(CASE WHEN status IN ('CLEARED', 'REPORTED') THEN 1 END), 0) AS cleared_count,
			COALESCE(COUNT(CASE WHEN status IN ('GENERATED', 'VALIDATED', 'SIGNED', 'QUEUED', 'SUBMITTING') THEN 1 END), 0) AS pending_count,
			COALESCE(COUNT(CASE WHEN status = 'WARNING' THEN 1 END), 0) AS warning_count,
			COALESCE(COUNT(CASE WHEN status IN ('REJECTED', 'FAILED') THEN 1 END), 0) AS rejected_count
		FROM zatca.documents
		WHERE legal_entity_id = $1
	`

	var total, cleared, pending, warning, rejected int
	err := r.pool.QueryRow(ctx, query, legalEntityID).Scan(&total, &cleared, &pending, &warning, &rejected)
	if err != nil {
		return &domain.ZATCAStatusSummary{ComplianceRate: 100.0}, nil
	}

	compliance := 100.0
	if total > 0 {
		compliance = math.Round((float64(cleared)/float64(total))*10000) / 100
	}

	return &domain.ZATCAStatusSummary{
		TotalDocuments: total,
		ClearedCount:   cleared,
		PendingCount:   pending,
		WarningCount:   warning,
		RejectedCount:  rejected,
		ComplianceRate: compliance,
	}, nil
}

func (r *dashboardRepository) GetRecentDocuments(ctx context.Context, legalEntityID uuid.UUID, limit int) ([]domain.RecentDocument, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			f.id,
			f.title,
			f.category,
			COALESCE(fv.original_filename, f.title) AS filename,
			COALESCE(fv.mime_type, 'application/octet-stream') AS mime_type,
			COALESCE(fv.byte_size, 0) AS byte_size,
			f.created_at
		FROM documents.files f
		LEFT JOIN documents.file_versions fv ON fv.id = f.current_version_id
		WHERE f.legal_entity_id = $1 AND f.is_archived = false
		ORDER BY f.created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, legalEntityID, limit)
	if err != nil {
		return []domain.RecentDocument{}, nil
	}
	defer rows.Close()

	var docs []domain.RecentDocument
	for rows.Next() {
		var doc domain.RecentDocument
		if err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&doc.Category,
			&doc.Filename,
			&doc.MimeType,
			&doc.ByteSize,
			&doc.CreatedAt,
		); err == nil {
			docs = append(docs, doc)
		}
	}

	if docs == nil {
		docs = []domain.RecentDocument{}
	}
	return docs, nil
}

func (r *dashboardRepository) GetRecentActivities(ctx context.Context, legalEntityID uuid.UUID, limit int) ([]domain.ActivityFeedItem, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			ae.id,
			ae.user_id,
			COALESCE(u.display_name, 'System') AS user_name,
			ae.action,
			ae.resource_type,
			ae.resource_id,
			ae.occurred_at
		FROM audit.events ae
		LEFT JOIN iam.users u ON u.id = ae.user_id
		WHERE (ae.legal_entity_id = $1 OR ae.legal_entity_id IS NULL)
		ORDER BY ae.occurred_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, legalEntityID, limit)
	if err != nil {
		return []domain.ActivityFeedItem{}, nil
	}
	defer rows.Close()

	var items []domain.ActivityFeedItem
	for rows.Next() {
		var item domain.ActivityFeedItem
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserName,
			&item.Action,
			&item.ResourceType,
			&item.ResourceID,
			&item.OccurredAt,
		); err == nil {
			items = append(items, item)
		}
	}

	if items == nil {
		items = []domain.ActivityFeedItem{}
	}
	return items, nil
}
