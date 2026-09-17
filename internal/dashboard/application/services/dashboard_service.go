package services

import (
	"context"
	"time"

	"erp/internal/dashboard/domain"
	"erp/internal/dashboard/ports"

	"github.com/google/uuid"
)

type dashboardService struct {
	repo ports.DashboardRepository
}

func NewDashboardService(repo ports.DashboardRepository) ports.DashboardService {
	return &dashboardService{
		repo: repo,
	}
}

func (s *dashboardService) GetDashboardData(ctx context.Context, legalEntityID uuid.UUID) (*domain.DashboardData, error) {
	// 1. Resolve Legal Entity info (and fallback if legalEntityID is uuid.Nil)
	resolvedID, entityName, currency, err := s.repo.GetLegalEntityInfo(ctx, legalEntityID)
	if err != nil {
		resolvedID = legalEntityID
		entityName = "الشركة الرئيسية"
		currency = "SAR"
	}

	// 2. Fetch Financials
	fin, err := s.repo.GetFinancialKPIs(ctx, resolvedID)
	if err != nil {
		fin = &domain.FinancialKPIs{Currency: currency}
	}
	fin.Currency = currency

	// 3. Fetch Cash Position
	cash, err := s.repo.GetCashPosition(ctx, resolvedID)
	if err != nil {
		cash = &domain.CashPositionKPI{Currency: currency}
	}
	cash.Currency = currency

	// 4. Fetch AR/AP Summary
	arap, err := s.repo.GetARAPSummary(ctx, resolvedID)
	if err != nil {
		arap = &domain.ARAPSummary{Currency: currency}
	}
	arap.Currency = currency

	// 5. Fetch Stock Warnings
	warnings, err := s.repo.GetStockWarnings(ctx, resolvedID, 10)
	if err != nil {
		warnings = []domain.StockWarning{}
	}

	// 6. Fetch Pending Approvals
	approvals, err := s.repo.GetPendingApprovals(ctx, resolvedID, 10)
	if err != nil {
		approvals = &domain.ApprovalsSummary{
			ByType:         make(map[string]int),
			RecentRequests: []domain.PendingApprovalItem{},
		}
	}

	// 7. Fetch ZATCA Summary
	zatca, err := s.repo.GetZATCASummary(ctx, resolvedID)
	if err != nil {
		zatca = &domain.ZATCAStatusSummary{ComplianceRate: 100.0}
	}

	// 8. Fetch Recent Documents
	docs, err := s.repo.GetRecentDocuments(ctx, resolvedID, 10)
	if err != nil {
		docs = []domain.RecentDocument{}
	}

	// 9. Fetch Activity Feed
	activities, err := s.repo.GetRecentActivities(ctx, resolvedID, 15)
	if err != nil {
		activities = []domain.ActivityFeedItem{}
	}

	data := &domain.DashboardData{
		LegalEntityID:   resolvedID,
		LegalEntityName: entityName,
		Currency:        currency,
		GeneratedAt:     time.Now().UTC(),
		Financials:      *fin,
		CashPosition:    *cash,
		ARAP:            *arap,
		StockWarnings:   warnings,
		Approvals:       *approvals,
		ZATCA:           *zatca,
		RecentDocs:      docs,
		ActivityFeed:    activities,
	}

	return data, nil
}
