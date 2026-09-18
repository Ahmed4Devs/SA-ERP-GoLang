package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/budgeting/domain"
	"erp/internal/budgeting/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetingService struct {
	repo ports.Repository
	db   *database.DB
	tx   transaction.Manager
}

func NewBudgetingService(repo ports.Repository, db *database.DB) *BudgetingService {
	return &BudgetingService{
		repo: repo,
		db:   db,
		tx:   transaction.NewManager(db.Pool),
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

	var createdBudget *domain.Budget
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdBudget, err = s.repo.CreateBudget(txCtx, b)
		if err != nil {
			return err
		}
		_, err = s.repo.CreateBudgetVersion(txCtx, &domain.BudgetVersion{
			ID:              uuid.New(),
			BudgetID:        createdBudget.ID,
			VersionNumber:   1,
			VersionName:     "Initial Working Budget v1.0",
			Status:          domain.VersionStatusDraft,
			Notes:           "Baseline draft version initialized with budget creation",
			CreatedByUserID: req.CreatedBy,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize budget version: %w", err)
		}
		return nil
	})
	return createdBudget, err
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
	var approved *domain.Budget
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		b, err := s.repo.GetBudgetByID(txCtx, id, legalEntityID)
		if err != nil {
			return domain.ErrBudgetNotFound
		}
		if b.Status == domain.BudgetStatusClosed {
			return domain.ErrCannotModifyClosed
		}

		now := time.Now()
		versions, err := s.repo.GetBudgetVersions(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to list budget versions: %w", err)
		}
		if len(versions) == 0 {
			return domain.ErrBudgetVersionNotFound
		}
		latest := versions[len(versions)-1]
		if _, err := s.repo.UpdateBudgetVersionStatus(txCtx, latest.ID, domain.VersionStatusActive, &approverID, &now); err != nil {
			return fmt.Errorf("failed to activate budget version: %w", err)
		}
		approved, err = s.repo.UpdateBudgetStatus(txCtx, id, legalEntityID, domain.BudgetStatusActive, &approverID, &now)
		return err
	})
	return approved, err
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

	var created *domain.BudgetLine
	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreateBudgetLine(txCtx, line)
		if err != nil {
			return err
		}
		if err := s.repo.RecalculateBudgetTotals(txCtx, req.BudgetID, req.LegalEntityID, req.BudgetVersionID); err != nil {
			return fmt.Errorf("failed to recalculate budget totals: %w", err)
		}
		return nil
	})
	return created, err
}

func (s *BudgetingService) ListBudgetLines(ctx context.Context, versionID uuid.UUID) ([]*domain.BudgetLine, error) {
	return s.repo.ListBudgetLines(ctx, versionID)
}

func (s *BudgetingService) TransferBudget(ctx context.Context, req ports.TransferBudgetRequest) (*domain.BudgetTransfer, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrNegativeAmount
	}
	if req.SourceLineID == req.TargetLineID {
		return nil, domain.ErrInvalidTransferLines
	}

	var created *domain.BudgetTransfer
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		// Lock in deterministic order to prevent concurrent transfers from
		// overspending a source line or deadlocking each other.
		first, second := req.SourceLineID, req.TargetLineID
		if first.String() > second.String() {
			first, second = second, first
		}
		for _, id := range []uuid.UUID{first, second} {
			if err := s.db.DBTX(txCtx).QueryRow(txCtx, "SELECT id FROM budgeting.budget_lines WHERE id = $1 FOR UPDATE", id).Scan(&id); err != nil {
				return domain.ErrBudgetLineNotFound
			}
		}

		src, err := s.repo.GetBudgetLineByID(txCtx, req.SourceLineID)
		if err != nil {
			return domain.ErrBudgetLineNotFound
		}
		tgt, err := s.repo.GetBudgetLineByID(txCtx, req.TargetLineID)
		if err != nil {
			return domain.ErrBudgetLineNotFound
		}
		if src.BudgetVersionID != tgt.BudgetVersionID {
			return domain.ErrInvalidTransferLines
		}
		if src.Amount.LessThan(req.Amount) {
			return domain.ErrInsufficientBudgetFunds
		}

		src.Amount = src.Amount.Sub(req.Amount)
		tgt.Amount = tgt.Amount.Add(req.Amount)
		if _, err := s.repo.UpdateBudgetLine(txCtx, src); err != nil {
			return err
		}
		if _, err := s.repo.UpdateBudgetLine(txCtx, tgt); err != nil {
			return err
		}
		created, err = s.repo.CreateBudgetTransfer(txCtx, &domain.BudgetTransfer{
			ID:                  uuid.New(),
			LegalEntityID:       req.LegalEntityID,
			TransferNumber:      identifiers.NewDocumentNumber("TRF", time.Now()),
			SourceLineID:        req.SourceLineID,
			TargetLineID:        req.TargetLineID,
			Amount:              req.Amount,
			Reason:              req.Reason,
			Status:              domain.TransferStatusCompleted,
			TransferredByUserID: req.TransferredBy,
		})
		return err
	})
	return created, err
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
