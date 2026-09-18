package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/fixedassets/domain"
	"erp/internal/fixedassets/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type FixedAssetsService struct {
	repo ports.Repository
	db   *database.DB
}

func NewFixedAssetsService(repo ports.Repository, db *database.DB) ports.Service {
	return &FixedAssetsService{
		repo: repo,
		db:   db,
	}
}

// ----------------------------------------------------------------------
// 1. Categories & Locations & Books
// ----------------------------------------------------------------------

func (s *FixedAssetsService) CreateAssetCategory(ctx context.Context, cat domain.AssetCategory) (*domain.AssetCategory, error) {
	if cat.UsefulLifeMonths <= 0 {
		return nil, fmt.Errorf("useful life months must be greater than 0")
	}
	if cat.DepreciationMethod == "" {
		cat.DepreciationMethod = "STRAIGHT_LINE"
	}
	cat.IsActive = true
	return s.repo.CreateAssetCategory(ctx, cat)
}

func (s *FixedAssetsService) ListAssetCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetCategory, error) {
	return s.repo.ListAssetCategories(ctx, legalEntityID)
}

func (s *FixedAssetsService) CreateAssetLocation(ctx context.Context, loc domain.AssetLocation) (*domain.AssetLocation, error) {
	loc.IsActive = true
	return s.repo.CreateAssetLocation(ctx, loc)
}

func (s *FixedAssetsService) ListAssetLocations(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetLocation, error) {
	return s.repo.ListAssetLocations(ctx, legalEntityID)
}

func (s *FixedAssetsService) ListDepreciationBooks(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationBook, error) {
	return s.repo.ListDepreciationBooks(ctx, legalEntityID)
}

// ----------------------------------------------------------------------
// 2. Fixed Assets
// ----------------------------------------------------------------------

func (s *FixedAssetsService) RegisterAsset(ctx context.Context, asset domain.Asset) (*domain.Asset, error) {
	if asset.AcquisitionCost.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("acquisition cost must be greater than zero")
	}
	if asset.UsefulLifeMonths <= 0 {
		// Inherit from category if not explicitly provided
		cat, err := s.repo.GetAssetCategoryByID(ctx, asset.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch category: %w", err)
		}
		asset.UsefulLifeMonths = cat.UsefulLifeMonths
		if asset.SalvageValue.IsZero() && cat.SalvageValueRate.GreaterThan(decimal.Zero) {
			asset.SalvageValue = asset.AcquisitionCost.Mul(cat.SalvageValueRate).Round(2)
		}
	}
	if asset.DepreciationMethod == "" {
		asset.DepreciationMethod = "STRAIGHT_LINE"
	}
	if asset.Status == "" {
		asset.Status = "ACTIVE"
	}

	asset.CurrentNetBookValue = asset.AcquisitionCost
	asset.AccumulatedDepreciation = decimal.Zero
	asset.ImpairmentTotal = decimal.Zero

	if asset.InServiceDate.IsZero() {
		asset.InServiceDate = asset.AcquisitionDate
	}

	savedAsset, err := s.repo.CreateFixedAsset(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to register fixed asset: %w", err)
	}

	// Auto-generate depreciation schedule for primary COMMERCIAL book
	books, err := s.repo.ListDepreciationBooks(ctx, asset.LegalEntityID)
	if err == nil && len(books) > 0 {
		var primaryBook *domain.DepreciationBook
		for _, b := range books {
			if b.IsPrimary {
				primaryBook = &b
				break
			}
		}
		if primaryBook == nil {
			primaryBook = &books[0]
		}

		schedules := domain.GenerateDepreciationSchedule(*savedAsset, primaryBook.ID, savedAsset.InServiceDate)
		for _, sched := range schedules {
			_, _ = s.repo.CreateDepreciationSchedule(ctx, sched)
		}
	}

	return savedAsset, nil
}

func (s *FixedAssetsService) GetAsset(ctx context.Context, id uuid.UUID) (*domain.Asset, error) {
	return s.repo.GetFixedAssetByID(ctx, id)
}

func (s *FixedAssetsService) ListAssets(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Asset, error) {
	return s.repo.ListFixedAssets(ctx, legalEntityID)
}

func (s *FixedAssetsService) AssignAsset(ctx context.Context, assignment domain.AssetAssignment) (*domain.AssetAssignment, error) {
	// End any active assignments for this asset
	_ = s.repo.EndAssetAssignment(ctx, assignment.AssetID, assignment.AssignedDate)

	assignment.IsCurrent = true
	return s.repo.CreateAssetAssignment(ctx, assignment)
}

// ----------------------------------------------------------------------
// 3. Depreciation Engine
// ----------------------------------------------------------------------

func (s *FixedAssetsService) RunMonthlyDepreciation(
	ctx context.Context,
	legalEntityID, bookID uuid.UUID,
	year, month int,
	runBy *uuid.UUID,
) (*domain.DepreciationRun, []domain.DepreciationLine, error) {
	activeAssets, err := s.repo.ListActiveFixedAssets(ctx, legalEntityID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list active assets: %w", err)
	}

	runDate := time.Date(year, time.Month(month), 28, 0, 0, 0, 0, time.UTC)
	runNumber := identifiers.NewDocumentNumber(fmt.Sprintf("DEP-%d%02d", year, month), time.Now())

	var linesToCreate []domain.DepreciationLine
	totalDepr := decimal.Zero

	for _, asset := range activeAssets {
		// Calculate straight-line depreciation for this month
		monthlyDepr := domain.CalculateStraightLineDepreciation(
			asset.AcquisitionCost,
			asset.SalvageValue,
			asset.UsefulLifeMonths,
			asset.CurrentNetBookValue,
		)

		if monthlyDepr.GreaterThan(decimal.Zero) {
			accumBefore := asset.AccumulatedDepreciation
			accumAfter := accumBefore.Add(monthlyDepr)
			nbvAfter := asset.AcquisitionCost.Sub(accumAfter)

			linesToCreate = append(linesToCreate, domain.DepreciationLine{
				ID:                   uuid.New(),
				AssetID:              asset.ID,
				AssetNumber:          asset.AssetNumber,
				CostBasis:            asset.AcquisitionCost,
				DepreciationAmount:   monthlyDepr,
				AccumulatedBefore:    accumBefore,
				AccumulatedAfter:     accumAfter,
				NetBookValueAfter:    nbvAfter,
				ExpenseAccountID:     asset.DepreciationExpenseAccountID,
				AccumulatedAccountID: asset.AccumulatedDepreciationAccountID,
				AssetNameEn:          asset.NameEn,
				AssetNameAr:          asset.NameAr,
			})
			totalDepr = totalDepr.Add(monthlyDepr)
		}
	}

	run := domain.DepreciationRun{
		ID:                      uuid.New(),
		LegalEntityID:           legalEntityID,
		BookID:                  bookID,
		RunNumber:               runNumber,
		PeriodYear:              year,
		PeriodMonth:             month,
		RunDate:                 runDate,
		TotalAssetsProcessed:    len(linesToCreate),
		TotalDepreciationAmount: totalDepr,
		Status:                  "CALCULATED",
		CreatedBy:               runBy,
	}

	savedRun, err := s.repo.CreateDepreciationRun(ctx, run)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create depreciation run: %w", err)
	}

	var savedLines []domain.DepreciationLine
	for _, l := range linesToCreate {
		l.RunID = savedRun.ID
		savedLine, err := s.repo.CreateDepreciationLine(ctx, l)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create depreciation line: %w", err)
		}
		savedLine.AssetNameEn = l.AssetNameEn
		savedLine.AssetNameAr = l.AssetNameAr
		savedLines = append(savedLines, *savedLine)
	}

	return savedRun, savedLines, nil
}

func (s *FixedAssetsService) PostDepreciationRun(ctx context.Context, runID uuid.UUID, approvedBy *uuid.UUID) (*domain.DepreciationRun, error) {
	run, err := s.repo.GetDepreciationRunByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch depreciation run: %w", err)
	}
	if run.Status == "POSTED" {
		return run, nil
	}

	lines, err := s.repo.ListDepreciationLinesByRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch depreciation lines: %w", err)
	}

	// Update each asset's NBV, accumulated depreciation, and schedule
	for _, line := range lines {
		_, err := s.repo.UpdateFixedAssetBookValue(ctx, line.AssetID, line.NetBookValueAfter, line.AccumulatedAfter)
		if err != nil {
			return nil, fmt.Errorf("failed updating asset %s book value: %w", line.AssetID, err)
		}

		_ = s.repo.MarkSchedulePosted(ctx, line.AssetID, run.BookID, run.ID, run.PeriodYear, run.PeriodMonth)
	}

	// Update run status to POSTED
	postedRun, err := s.repo.UpdateDepreciationRunStatus(ctx, runID, "POSTED", nil, approvedBy)
	if err != nil {
		return nil, fmt.Errorf("failed updating depreciation run status to POSTED: %w", err)
	}

	return postedRun, nil
}

func (s *FixedAssetsService) GetDepreciationRun(ctx context.Context, runID uuid.UUID) (*domain.DepreciationRun, []domain.DepreciationLine, error) {
	run, err := s.repo.GetDepreciationRunByID(ctx, runID)
	if err != nil {
		return nil, nil, err
	}
	lines, err := s.repo.ListDepreciationLinesByRun(ctx, runID)
	if err != nil {
		return nil, nil, err
	}
	return run, lines, nil
}

func (s *FixedAssetsService) ListDepreciationRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationRun, error) {
	return s.repo.ListDepreciationRuns(ctx, legalEntityID)
}

func (s *FixedAssetsService) GetAssetSchedule(ctx context.Context, assetID, bookID uuid.UUID) ([]domain.DepreciationSchedule, error) {
	return s.repo.ListDepreciationSchedulesByAsset(ctx, assetID, bookID)
}

// ----------------------------------------------------------------------
// 4. Transfers & Disposals
// ----------------------------------------------------------------------

func (s *FixedAssetsService) TransferAsset(ctx context.Context, transfer domain.Transfer) (*domain.Transfer, error) {
	if transfer.TransferNumber == "" {
		transfer.TransferNumber = identifiers.NewDocumentNumber("TRF", time.Now())
	}
	if transfer.TransferDate.IsZero() {
		transfer.TransferDate = time.Now()
	}
	transfer.Status = "COMPLETED"

	saved, err := s.repo.CreateAssetTransfer(ctx, transfer)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset transfer: %w", err)
	}

	// Update asset location
	_, err = s.repo.UpdateFixedAssetLocation(ctx, transfer.AssetID, transfer.ToLocationID)
	if err != nil {
		return nil, fmt.Errorf("failed to update asset location: %w", err)
	}

	// Reassign custodian if changed
	if transfer.ToEmployeeID != nil {
		_ = s.repo.EndAssetAssignment(ctx, transfer.AssetID, transfer.TransferDate)
		_, _ = s.repo.CreateAssetAssignment(ctx, domain.AssetAssignment{
			ID:             uuid.New(),
			AssetID:        transfer.AssetID,
			AssignedToType: "EMPLOYEE",
			EmployeeID:     transfer.ToEmployeeID,
			AssignedDate:   transfer.TransferDate,
			Notes:          "Transferred via " + transfer.TransferNumber,
			IsCurrent:      true,
		})
	}

	return saved, nil
}

func (s *FixedAssetsService) ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Transfer, error) {
	return s.repo.ListAssetTransfers(ctx, legalEntityID)
}

func (s *FixedAssetsService) DisposeAsset(ctx context.Context, disposal domain.Disposal) (*domain.Disposal, error) {
	asset, err := s.repo.GetFixedAssetByID(ctx, disposal.AssetID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch asset for disposal: %w", err)
	}

	if disposal.DisposalNumber == "" {
		disposal.DisposalNumber = identifiers.NewDocumentNumber("DISP", time.Now())
	}
	if disposal.DisposalDate.IsZero() {
		disposal.DisposalDate = time.Now()
	}

	disposal.CostAtDisposal = asset.AcquisitionCost
	disposal.AccumulatedDepreciationAtDisposal = asset.AccumulatedDepreciation
	disposal.NetBookValueAtDisposal = asset.CurrentNetBookValue
	disposal.GainOrLossAmount = disposal.ProceedsAmount.Sub(asset.CurrentNetBookValue)
	disposal.Status = "POSTED"

	saved, err := s.repo.CreateAssetDisposal(ctx, disposal)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset disposal: %w", err)
	}

	// Update asset book value and set status to DISPOSED
	_, _ = s.repo.UpdateFixedAssetBookValue(ctx, asset.ID, decimal.Zero, asset.AccumulatedDepreciation)
	_, err = s.repo.UpdateFixedAssetStatus(ctx, asset.ID, "DISPOSED")
	if err != nil {
		return nil, fmt.Errorf("failed to update asset status to DISPOSED: %w", err)
	}

	return saved, nil
}

// ----------------------------------------------------------------------
// 5. Dashboard Summary
// ----------------------------------------------------------------------

func (s *FixedAssetsService) GetDashboardSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.DashboardSummary, error) {
	assets, err := s.repo.ListFixedAssets(ctx, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets for dashboard: %w", err)
	}

	totalCost := decimal.Zero
	totalNBV := decimal.Zero
	totalAccum := decimal.Zero
	activeCount := 0

	for _, a := range assets {
		if a.Status == "ACTIVE" || a.Status == "FULLY_DEPRECIATED" {
			totalCost = totalCost.Add(a.AcquisitionCost)
			totalNBV = totalNBV.Add(a.CurrentNetBookValue)
			totalAccum = totalAccum.Add(a.AccumulatedDepreciation)
		}
		if a.Status == "ACTIVE" {
			activeCount++
		}
	}

	rate := "100%"
	if len(assets) > 0 {
		pct := decimal.NewFromInt(int64(activeCount)).Div(decimal.NewFromInt(int64(len(assets)))).Mul(decimal.NewFromInt(100)).Round(1)
		rate = fmt.Sprintf("%s%%", pct.String())
	}

	return &domain.DashboardSummary{
		TotalHistoricalCost:  totalCost,
		TotalNetBookValue:    totalNBV,
		TotalAccumulatedDepr: totalAccum,
		TotalActiveAssets:    activeCount,
		InServiceRate:        rate,
	}, nil
}
