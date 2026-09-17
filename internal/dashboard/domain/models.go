package domain

import (
	"time"

	"github.com/google/uuid"
)

// FinancialKPIs represents executive financial metrics.
type FinancialKPIs struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalExpense float64 `json:"total_expense"`
	NetProfit    float64 `json:"net_profit"`
	ProfitMargin float64 `json:"profit_margin"` // percentage
	Currency     string  `json:"currency"`
}

// CashPositionKPI represents bank and cash liquid balance.
type CashPositionKPI struct {
	TotalCashBalance float64 `json:"total_cash_balance"`
	TotalBankBalance float64 `json:"total_bank_balance"`
	TotalLiquidFunds float64 `json:"total_liquid_funds"`
	AccountsCount    int     `json:"accounts_count"`
	Currency         string  `json:"currency"`
}

// ARAPSummary represents accounts receivable and payables summary.
type ARAPSummary struct {
	TotalReceivables   float64 `json:"total_receivables"`
	OverdueReceivables float64 `json:"overdue_receivables"`
	OpenReceivableDocs int     `json:"open_receivable_docs"`
	TotalPayables      float64 `json:"total_payables"`
	OverduePayables    float64 `json:"overdue_payables"`
	OpenPayableDocs    int     `json:"open_payable_docs"`
	Currency           string  `json:"currency"`
}

// StockWarning represents an item below safety/reorder threshold.
type StockWarning struct {
	ItemID            uuid.UUID `json:"item_id"`
	ItemCode          string    `json:"item_code"`
	ItemNameEn        string    `json:"item_name_en"`
	ItemNameAr        string    `json:"item_name_ar"`
	WarehouseNameEn   string    `json:"warehouse_name_en"`
	WarehouseNameAr   string    `json:"warehouse_name_ar"`
	QuantityOnHand    float64   `json:"quantity_on_hand"`
	QuantityReserved  float64   `json:"quantity_reserved"`
	QuantityAvailable float64   `json:"quantity_available"`
	ReorderPoint      float64   `json:"reorder_point"`
	Severity          string    `json:"severity"` // 'CRITICAL', 'WARNING', 'LOW'
}

// PendingApprovalItem represents an awaiting approval task.
type PendingApprovalItem struct {
	ID             uuid.UUID `json:"id"`
	ResourceType   string    `json:"resource_type"`
	ResourceID     string    `json:"resource_id"`
	InitiatorID    uuid.UUID `json:"initiator_id"`
	AssignedUserID uuid.UUID `json:"assigned_user_id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// ApprovalsSummary represents pending approvals count and recent requests.
type ApprovalsSummary struct {
	TotalPending   int                   `json:"total_pending"`
	ByType         map[string]int        `json:"by_type"`
	RecentRequests []PendingApprovalItem `json:"recent_requests"`
}

// ZATCAStatusSummary represents e-invoicing compliance metrics.
type ZATCAStatusSummary struct {
	TotalDocuments int     `json:"total_documents"`
	ClearedCount   int     `json:"cleared_count"`
	PendingCount   int     `json:"pending_count"`
	WarningCount   int     `json:"warning_count"`
	RejectedCount  int     `json:"rejected_count"`
	ComplianceRate float64 `json:"compliance_rate"` // percentage
}

// RecentDocument represents recently added or modified files.
type RecentDocument struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	ByteSize  int64     `json:"byte_size"`
	CreatedAt time.Time `json:"created_at"`
}

// ActivityFeedItem represents a recent audit trail log.
type ActivityFeedItem struct {
	ID           uuid.UUID  `json:"id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	UserName     string     `json:"user_name"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   string     `json:"resource_id"`
	OccurredAt   time.Time  `json:"occurred_at"`
}

// DashboardData is the consolidated top-level dashboard aggregate.
type DashboardData struct {
	LegalEntityID   uuid.UUID          `json:"legal_entity_id"`
	LegalEntityName string             `json:"legal_entity_name"`
	Currency        string             `json:"currency"`
	GeneratedAt     time.Time          `json:"generated_at"`
	Financials      FinancialKPIs      `json:"financials"`
	CashPosition    CashPositionKPI    `json:"cash_position"`
	ARAP            ARAPSummary        `json:"arap"`
	StockWarnings   []StockWarning     `json:"stock_warnings"`
	Approvals       ApprovalsSummary   `json:"approvals"`
	ZATCA           ZATCAStatusSummary `json:"zatca"`
	RecentDocs      []RecentDocument   `json:"recent_documents"`
	ActivityFeed    []ActivityFeedItem `json:"activity_feed"`
}
