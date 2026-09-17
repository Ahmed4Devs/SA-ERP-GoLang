package ports

import (
	"context"

	"erp/internal/dashboard/domain"

	"github.com/google/uuid"
)

type DashboardRepository interface {
	GetLegalEntityInfo(ctx context.Context, legalEntityID uuid.UUID) (id uuid.UUID, name string, currency string, err error)
	GetFinancialKPIs(ctx context.Context, legalEntityID uuid.UUID) (*domain.FinancialKPIs, error)
	GetCashPosition(ctx context.Context, legalEntityID uuid.UUID) (*domain.CashPositionKPI, error)
	GetARAPSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ARAPSummary, error)
	GetStockWarnings(ctx context.Context, legalEntityID uuid.UUID, limit int) ([]domain.StockWarning, error)
	GetPendingApprovals(ctx context.Context, legalEntityID uuid.UUID, limit int) (*domain.ApprovalsSummary, error)
	GetZATCASummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.ZATCAStatusSummary, error)
	GetRecentDocuments(ctx context.Context, legalEntityID uuid.UUID, limit int) ([]domain.RecentDocument, error)
	GetRecentActivities(ctx context.Context, legalEntityID uuid.UUID, limit int) ([]domain.ActivityFeedItem, error)
}

type DashboardService interface {
	GetDashboardData(ctx context.Context, legalEntityID uuid.UUID) (*domain.DashboardData, error)
}
