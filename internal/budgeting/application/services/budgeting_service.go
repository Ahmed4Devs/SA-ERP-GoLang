package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/budgeting/domain"
	"erp/internal/budgeting/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetingService struct {
	repo ports.Repository
	db   *database.DB
}

func NewBudgetingService(repo ports.Repository, db *database.DB) *BudgetingService {
	return &BudgetingService{
		repo: repo,
		db:   db,
	}
}

func (s *BudgetingService) CreateBudget(ctx context.Context, req ports.CreateBudgetRequest) (*domain.Budget, error) {
	if req.Code == "" || req.NameEn == "" || req.NameAr == "" {
		return nil, domain.ErrInvalidDimensions
	}
	if req.FiscalYearID == uuid.Nil {
		return nil, fmt.Errorf("fiscal_year_id is required")
	}

	existing, _ := s.repo.GetBudgetByCode(ctx, req.Code, req.LegalEntityID)
	if existing != nil {
		return nil, domain.ErrBudgetCodeExists
	}

	curr := req.CurrencyCode
	if curr == "" {
		curr = "SAR"
	}

	bType := req.BudgetType
	if bType == "" {
		bType = domain.BudgetTypeOperating
	}

	b := &domain.Budget{
		ID:                   uuid.New(),
		LegalEntityID:        req.LegalEntityID,
		Code:                 req.Code,
		NameEn:               req.NameEn,
		NameAr:               req.NameAr,
		DescriptionEn:        req.DescriptionEn,
		DescriptionAr:        req.DescriptionAr,
		FiscalYearID:         req.FiscalYearID,
		CurrencyCode:         curr,
		BudgetType:           bType,
		Status:               domain.BudgetStatusDraft,
		TotalBudgetedAmount:  decimal.Zero,
		TotalActualAmount:    decimal.Zero,
		TotalCommittedAmount: decimal.Zero,
		CreatedByUserID:      req.CreatedBy,
	}

	createdBudget, err := s.repo.CreateBudget(ctx, b)
	if err != nil {
		return nil, err
	}

	// Automatically initialize Version 1
	v1 := &domain.BudgetVersion{
		ID:              uuid.New(),
		BudgetID:        createdBudget.ID,
		VersionNumber:   1,
		VersionName:     "Initial Working Budget v1.0",
		Status:          domain.VersionStatusDraft,
		Notes:           "Baseline draft version initialized with budget creation",
		CreatedByUserID: req.CreatedBy,
	}
	_, _ = s.repo.CreateBudgetVersion(ctx, v1)

	return createdBudget, nil
}

func (s *BudgetingService) GetBudget(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.Budget, error) {
	b, err := s.repo.GetBudgetByID(ctx, id, legalEntityID)
	if err != nil {
		return nil, domain.ErrBudgetNotFound
	}
	return b, nil
}

func (s *BudgetingService) ListBudgets(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Budget, error) {
	return s.repo.ListBudgets(ctx, legalEntityID)
}

func (s *BudgetingService) SubmitBudget(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.Budget, error) {
	b, err := s.repo.GetBudgetByID(ctx, id, legalEntityID)
	if err != nil {
		return nil, domain.ErrBudgetNotFound
	}
	if b.Status != domain.BudgetStatusDraft {
		return nil, domain.ErrInvalidBudgetStatus
	}

	return s.repo.UpdateBudgetStatus(ctx, id, legalEntityID, domain.BudgetStatusSubmitted, nil, nil)
}

func (s *BudgetingService) ApproveBudget(ctx context.Context, id, legalEntityID, approverID uuid.UUID) (*domain.Budget, error) {
	b, err := s.repo.GetBudgetByID(ctx, id, legalEntityID)
	if err != nil {
		return nil, domain.ErrBudgetNotFound
	}
	if b.Status == domain.BudgetStatusClosed {
		return nil, domain.ErrCannotModifyClosed
	}

	now := time.Now()
	// Approve active version or version 1
	versions, err := s.repo.GetBudgetVersions(ctx, id)
	if err == nil && len(versions) > 0 {
		latest := versions[len(versions)-1]
		_, _ = s.repo.UpdateBudgetVersionStatus(ctx, latest.ID, domain.VersionStatusActive, &approverID, &now)
	}

	return s.repo.UpdateBudgetStatus(ctx, id, legalEntityID, domain.BudgetStatusActive, &approverID, &now)
}

func (s *BudgetingService) CreateVersion(ctx context.Context, req ports.CreateVersionRequest) (*domain.BudgetVersion, error) {
	versions, err := s.repo.GetBudgetVersions(ctx, req.BudgetID)
	if err != nil {
		return nil, err
	}

	var nextVersionNum int32 = 1
	if len(versions) > 0 {
		nextVersionNum = versions[len(versions)-1].VersionNumber + 1
	}

	v := &domain.BudgetVersion{
		ID:              uuid.New(),
		BudgetID:        req.BudgetID,
		VersionNumber:   nextVersionNum,
		VersionName:     req.VersionName,
		Status:          domain.VersionStatusDraft,
		Notes:           req.Notes,
		CreatedByUserID: req.CreatedBy,
	}

	return s.repo.CreateBudgetVersion(ctx, v)
}

func (s *BudgetingService) ListVersions(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetVersion, error) {
	return s.repo.GetBudgetVersions(ctx, budgetID)
}

func (s *BudgetingService) AddBudgetLine(ctx context.Context, req ports.AddBudgetLineRequest) (*domain.BudgetLine, error) {
	if req.Amount.IsNegative() {
		return nil, domain.ErrNegativeAmount
	}
	if req.AccountID == uuid.Nil {
		return nil, domain.ErrInvalidDimensions
	}

	v, err := s.repo.GetBudgetVersionByID(ctx, req.BudgetVersionID)
	if err != nil {
		return nil, domain.ErrBudgetVersionNotFound
	}
	if v.Status == domain.VersionStatusSuperseded {
		return nil, domain.ErrCannotModifyClosed
	}

	line := &domain.BudgetLine{
		ID:                uuid.New(),
		BudgetVersionID:   req.BudgetVersionID,
		AccountID:         req.AccountID,
		DepartmentID:      req.DepartmentID,
		CostCenterID:      req.CostCenterID,
		ProjectCode:       req.ProjectCode,
		FiscalPeriodID:    req.FiscalPeriodID,
		Amount:            req.Amount,
		ActualAmount:      decimal.Zero,
		CommittedAmount:   decimal.Zero,
		LineDescriptionEn: req.LineDescriptionEn,
		LineDescriptionAr: req.LineDescriptionAr,
	}

	created, err := s.repo.CreateBudgetLine(ctx, line)
	if err != nil {
		return nil, err
	}

	// Recalculate totals
	_ = s.repo.RecalculateBudgetTotals(ctx, req.BudgetID, req.LegalEntityID, req.BudgetVersionID)

	return created, nil
}

func (s *BudgetingService) ListBudgetLines(ctx context.Context, versionID uuid.UUID) ([]*domain.BudgetLine, error) {
	return s.repo.ListBudgetLines(ctx, versionID)
}

func (s *BudgetingService) TransferBudget(ctx context.Context, req ports.TransferBudgetRequest) (*domain.BudgetTransfer, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrNegativeAmount
	}

	src, err := s.repo.GetBudgetLineByID(ctx, req.SourceLineID)
	if err != nil {
		return nil, domain.ErrBudgetLineNotFound
	}

	tgt, err := s.repo.GetBudgetLineByID(ctx, req.TargetLineID)
	if err != nil {
		return nil, domain.ErrBudgetLineNotFound
	}

	if src.BudgetVersionID != tgt.BudgetVersionID {
		return nil, domain.ErrInvalidTransferLines
	}

	// Verify source has enough allocated amount
	if src.Amount.LessThan(req.Amount) {
		return nil, domain.ErrInsufficientBudgetFunds
	}

	// Execute transfer
	src.Amount = src.Amount.Sub(req.Amount)
	tgt.Amount = tgt.Amount.Add(req.Amount)

	if _, err := s.repo.UpdateBudgetLine(ctx, src); err != nil {
		return nil, err
	}
	if _, err := s.repo.UpdateBudgetLine(ctx, tgt); err != nil {
		return nil, err
	}

	trfNumber := identifiers.NewDocumentNumber("TRF", time.Now())
	transfer := &domain.BudgetTransfer{
		ID:                  uuid.New(),
		LegalEntityID:       req.LegalEntityID,
		TransferNumber:      trfNumber,
		SourceLineID:        req.SourceLineID,
		TargetLineID:        req.TargetLineID,
		Amount:              req.Amount,
		Reason:              req.Reason,
		Status:              domain.TransferStatusCompleted,
		TransferredByUserID: req.TransferredBy,
	}

	return s.repo.CreateBudgetTransfer(ctx, transfer)
}

func (s *BudgetingService) ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BudgetTransfer, error) {
	return s.repo.ListBudgetTransfers(ctx, legalEntityID)
}

func (s *BudgetingService) GetBudgetVsActual(ctx context.Context, budgetID, legalEntityID uuid.UUID) ([]*domain.BudgetVsActualItem, error) {
	// Verify budget exists
	b, err := s.repo.GetBudgetByID(ctx, budgetID, legalEntityID)
	if err != nil {
		return nil, domain.ErrBudgetNotFound
	}

	// Get active or first version
	activeVer, err := s.repo.GetActiveBudgetVersion(ctx, b.ID)
	if err != nil || activeVer == nil {
		versions, errVer := s.repo.GetBudgetVersions(ctx, b.ID)
		if errVer != nil || len(versions) == 0 {
			return []*domain.BudgetVsActualItem{}, nil
		}
		activeVer = versions[len(versions)-1]
	}

	lines, err := s.repo.ListBudgetLines(ctx, activeVer.ID)
	if err != nil {
		return nil, err
	}

	items := make([]*domain.BudgetVsActualItem, 0, len(lines))
	for _, l := range lines {
		rem := l.Amount.Sub(l.ActualAmount)
		variance := l.Amount.Sub(l.ActualAmount)

		rate := decimal.Zero
		if !l.Amount.IsZero() {
			rate = l.ActualAmount.Div(l.Amount).Mul(decimal.NewFromInt(100)).Round(2)
		}

		vStatus := "ON_TRACK"
		if rate.GreaterThan(decimal.NewFromInt(100)) {
			vStatus = "OVER_BUDGET"
		} else if rate.GreaterThanOrEqual(decimal.NewFromInt(80)) {
			vStatus = "WARNING"
		}

		items = append(items, &domain.BudgetVsActualItem{
			LineID:           l.ID,
			AccountID:        l.AccountID,
			AccountCode:      l.AccountCode,
			AccountNameEn:    l.AccountNameEn,
			AccountNameAr:    l.AccountNameAr,
			DepartmentNameEn: l.DepartmentNameEn,
			DepartmentNameAr: l.DepartmentNameAr,
			CostCenterNameEn: l.CostCenterNameEn,
			CostCenterNameAr: l.CostCenterNameAr,
			PeriodNameEn:     l.PeriodNameEn,
			PeriodNameAr:     l.PeriodNameAr,
			ProjectCode:      l.ProjectCode,
			BudgetedAmount:   l.Amount,
			ActualAmount:     l.ActualAmount,
			CommittedAmount:  l.CommittedAmount,
			RemainingAmount:  rem,
			VarianceAmount:   variance,
			ConsumptionRate:  rate,
			VarianceStatus:   vStatus,
		})
	}

	return items, nil
}

func (s *BudgetingService) GetBudgetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.BudgetSummary, error) {
	budgets, err := s.repo.ListBudgets(ctx, legalEntityID)
	if err != nil {
		return nil, err
	}

	summary := &domain.BudgetSummary{
		TotalBudgeted:  decimal.Zero,
		TotalActual:    decimal.Zero,
		TotalCommitted: decimal.Zero,
		TotalRemaining: decimal.Zero,
	}

	activeCount := 0
	for _, b := range budgets {
		summary.TotalBudgeted = summary.TotalBudgeted.Add(b.TotalBudgetedAmount)
		summary.TotalActual = summary.TotalActual.Add(b.TotalActualAmount)
		summary.TotalCommitted = summary.TotalCommitted.Add(b.TotalCommittedAmount)
		if b.Status == domain.BudgetStatusActive {
			activeCount++
		}
	}

	summary.ActiveBudgetCount = activeCount
	summary.TotalRemaining = summary.TotalBudgeted.Sub(summary.TotalActual)

	if !summary.TotalBudgeted.IsZero() {
		summary.AverageUsageRate = summary.TotalActual.Div(summary.TotalBudgeted).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return summary, nil
}

var _ ports.Service = (*BudgetingService)(nil)
