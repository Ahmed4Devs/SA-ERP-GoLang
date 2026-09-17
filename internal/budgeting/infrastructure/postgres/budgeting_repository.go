package postgres

import (
	"context"
	"time"

	"erp/internal/budgeting/domain"
	"erp/internal/budgeting/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *database.DB
	q  *sqlc.Queries
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
		q:  sqlc.New(db.Pool),
	}
}

func (r *Repository) GetBudgetByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.Budget, error) {
	row, err := r.q.GetBudgetByID(ctx, sqlc.GetBudgetByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetToDomain(row), nil
}

func (r *Repository) GetBudgetByCode(ctx context.Context, code string, legalEntityID uuid.UUID) (*domain.Budget, error) {
	row, err := r.q.GetBudgetByCode(ctx, sqlc.GetBudgetByCodeParams{
		Code:          code,
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetToDomain(row), nil
}

func (r *Repository) ListBudgets(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Budget, error) {
	rows, err := r.q.ListBudgets(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Budget, len(rows))
	for i, row := range rows {
		result[i] = mapBudgetToDomain(row)
	}
	return result, nil
}

func (r *Repository) CreateBudget(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	row, err := r.q.CreateBudget(ctx, sqlc.CreateBudgetParams{
		ID:                   toPgUUID(b.ID),
		LegalEntityID:        toPgUUID(b.LegalEntityID),
		Code:                 b.Code,
		NameEn:               b.NameEn,
		NameAr:               b.NameAr,
		DescriptionEn:        toPgText(b.DescriptionEn),
		DescriptionAr:        toPgText(b.DescriptionAr),
		FiscalYearID:         toPgUUID(b.FiscalYearID),
		CurrencyCode:         b.CurrencyCode,
		BudgetType:           string(b.BudgetType),
		Status:               string(b.Status),
		TotalBudgetedAmount:  toPgNumeric(b.TotalBudgetedAmount),
		TotalActualAmount:    toPgNumeric(b.TotalActualAmount),
		TotalCommittedAmount: toPgNumeric(b.TotalCommittedAmount),
		CreatedByUserID:      toPgUUIDPtr(b.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetToDomain(row), nil
}

func (r *Repository) UpdateBudgetStatus(ctx context.Context, id, legalEntityID uuid.UUID, status domain.BudgetStatus, approverID *uuid.UUID, approvedAt *time.Time) (*domain.Budget, error) {
	var appAt time.Time
	if approvedAt != nil {
		appAt = *approvedAt
	}
	row, err := r.q.UpdateBudgetStatus(ctx, sqlc.UpdateBudgetStatusParams{
		ID:               toPgUUID(id),
		LegalEntityID:    toPgUUID(legalEntityID),
		Status:           string(status),
		ApprovedByUserID: toPgUUIDPtr(approverID),
		ApprovedAt:       toPgTimestamptz(appAt),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetToDomain(row), nil
}

func (r *Repository) UpdateBudgetTotals(ctx context.Context, id, legalEntityID uuid.UUID, budgeted, actual, committed decimal.Decimal) (*domain.Budget, error) {
	row, err := r.q.UpdateBudgetTotals(ctx, sqlc.UpdateBudgetTotalsParams{
		ID:                   toPgUUID(id),
		LegalEntityID:        toPgUUID(legalEntityID),
		TotalBudgetedAmount:  toPgNumeric(budgeted),
		TotalActualAmount:    toPgNumeric(actual),
		TotalCommittedAmount: toPgNumeric(committed),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetToDomain(row), nil
}

func (r *Repository) DeleteBudget(ctx context.Context, id, legalEntityID uuid.UUID) error {
	return r.q.DeleteBudget(ctx, sqlc.DeleteBudgetParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
}

func (r *Repository) GetBudgetVersions(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetVersion, error) {
	rows, err := r.q.GetBudgetVersions(ctx, toPgUUID(budgetID))
	if err != nil {
		return nil, err
	}
	result := make([]*domain.BudgetVersion, len(rows))
	for i, row := range rows {
		result[i] = mapBudgetVersionToDomain(row)
	}
	return result, nil
}

func (r *Repository) GetBudgetVersionByID(ctx context.Context, id uuid.UUID) (*domain.BudgetVersion, error) {
	row, err := r.q.GetBudgetVersionByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return mapBudgetVersionToDomain(row), nil
}

func (r *Repository) GetActiveBudgetVersion(ctx context.Context, budgetID uuid.UUID) (*domain.BudgetVersion, error) {
	row, err := r.q.GetActiveBudgetVersion(ctx, toPgUUID(budgetID))
	if err != nil {
		return nil, err
	}
	return mapBudgetVersionToDomain(row), nil
}

func (r *Repository) CreateBudgetVersion(ctx context.Context, v *domain.BudgetVersion) (*domain.BudgetVersion, error) {
	row, err := r.q.CreateBudgetVersion(ctx, sqlc.CreateBudgetVersionParams{
		ID:              toPgUUID(v.ID),
		BudgetID:        toPgUUID(v.BudgetID),
		VersionNumber:   v.VersionNumber,
		VersionName:     v.VersionName,
		Status:          string(v.Status),
		Notes:           toPgText(v.Notes),
		CreatedByUserID: toPgUUIDPtr(v.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetVersionToDomain(row), nil
}

func (r *Repository) UpdateBudgetVersionStatus(ctx context.Context, id uuid.UUID, status domain.VersionStatus, approverID *uuid.UUID, approvedAt *time.Time) (*domain.BudgetVersion, error) {
	var appAt time.Time
	if approvedAt != nil {
		appAt = *approvedAt
	}
	row, err := r.q.UpdateBudgetVersionStatus(ctx, sqlc.UpdateBudgetVersionStatusParams{
		ID:               toPgUUID(id),
		Status:           string(status),
		ApprovedByUserID: toPgUUIDPtr(approverID),
		ApprovedAt:       toPgTimestamptz(appAt),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetVersionToDomain(row), nil
}

func (r *Repository) ListBudgetLines(ctx context.Context, versionID uuid.UUID) ([]*domain.BudgetLine, error) {
	rows, err := r.q.ListBudgetLines(ctx, toPgUUID(versionID))
	if err != nil {
		return nil, err
	}
	result := make([]*domain.BudgetLine, len(rows))
	for i, row := range rows {
		result[i] = mapBudgetLineRowToDomain(row)
	}
	return result, nil
}

func (r *Repository) GetBudgetLineByID(ctx context.Context, id uuid.UUID) (*domain.BudgetLine, error) {
	row, err := r.q.GetBudgetLineByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return mapBudgetLineToDomain(row), nil
}

func (r *Repository) CreateBudgetLine(ctx context.Context, l *domain.BudgetLine) (*domain.BudgetLine, error) {
	row, err := r.q.CreateBudgetLine(ctx, sqlc.CreateBudgetLineParams{
		ID:                toPgUUID(l.ID),
		BudgetVersionID:   toPgUUID(l.BudgetVersionID),
		AccountID:         toPgUUID(l.AccountID),
		DepartmentID:      toPgUUIDPtr(l.DepartmentID),
		CostCenterID:      toPgUUIDPtr(l.CostCenterID),
		ProjectCode:       toPgText(l.ProjectCode),
		FiscalPeriodID:    toPgUUIDPtr(l.FiscalPeriodID),
		Amount:            toPgNumeric(l.Amount),
		ActualAmount:      toPgNumeric(l.ActualAmount),
		CommittedAmount:   toPgNumeric(l.CommittedAmount),
		LineDescriptionEn: toPgText(l.LineDescriptionEn),
		LineDescriptionAr: toPgText(l.LineDescriptionAr),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetLineToDomain(row), nil
}

func (r *Repository) UpdateBudgetLine(ctx context.Context, l *domain.BudgetLine) (*domain.BudgetLine, error) {
	row, err := r.q.UpdateBudgetLine(ctx, sqlc.UpdateBudgetLineParams{
		ID:                toPgUUID(l.ID),
		Amount:            toPgNumeric(l.Amount),
		ActualAmount:      toPgNumeric(l.ActualAmount),
		CommittedAmount:   toPgNumeric(l.CommittedAmount),
		LineDescriptionEn: toPgText(l.LineDescriptionEn),
		LineDescriptionAr: toPgText(l.LineDescriptionAr),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetLineToDomain(row), nil
}

func (r *Repository) DeleteBudgetLine(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteBudgetLine(ctx, toPgUUID(id))
}

func (r *Repository) RecalculateBudgetTotals(ctx context.Context, budgetID, legalEntityID uuid.UUID, versionID uuid.UUID) error {
	totals, err := r.q.SumBudgetLinesByVersion(ctx, toPgUUID(versionID))
	if err != nil {
		return err
	}

	budgeted := fromPgNumeric(totals.TotalBudgeted)
	actual := fromPgNumeric(totals.TotalActual)
	committed := fromPgNumeric(totals.TotalCommitted)

	_, err = r.UpdateBudgetTotals(ctx, budgetID, legalEntityID, budgeted, actual, committed)
	return err
}

func (r *Repository) CreateBudgetTransfer(ctx context.Context, t *domain.BudgetTransfer) (*domain.BudgetTransfer, error) {
	row, err := r.q.CreateBudgetTransfer(ctx, sqlc.CreateBudgetTransferParams{
		ID:                  toPgUUID(t.ID),
		LegalEntityID:       toPgUUID(t.LegalEntityID),
		TransferNumber:      t.TransferNumber,
		SourceLineID:        toPgUUID(t.SourceLineID),
		TargetLineID:        toPgUUID(t.TargetLineID),
		Amount:              toPgNumeric(t.Amount),
		Reason:              t.Reason,
		Status:              string(t.Status),
		TransferredByUserID: toPgUUIDPtr(t.TransferredByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapBudgetTransferToDomain(row), nil
}

func (r *Repository) ListBudgetTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BudgetTransfer, error) {
	rows, err := r.q.ListBudgetTransfers(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	result := make([]*domain.BudgetTransfer, len(rows))
	for i, row := range rows {
		result[i] = &domain.BudgetTransfer{
			ID:                  toUUID(row.ID),
			LegalEntityID:       toUUID(row.LegalEntityID),
			TransferNumber:      row.TransferNumber,
			SourceLineID:        toUUID(row.SourceLineID),
			TargetLineID:        toUUID(row.TargetLineID),
			Amount:              fromPgNumeric(row.Amount),
			Reason:              row.Reason,
			Status:              domain.TransferStatus(row.Status),
			TransferredByUserID: toUUIDPtr(row.TransferredByUserID),
			CreatedAt:           fromPgTimestamptz(row.CreatedAt),
			SourceAccountCode:   row.SourceAccountCode,
			SourceAccountNameEn: row.SourceAccountNameEn,
			SourceAccountNameAr: row.SourceAccountNameAr,
			TargetAccountCode:   row.TargetAccountCode,
			TargetAccountNameEn: row.TargetAccountNameEn,
			TargetAccountNameAr: row.TargetAccountNameAr,
		}
	}
	return result, nil
}

// Helpers for model mapping
func mapBudgetToDomain(row sqlc.BudgetingBudgets) *domain.Budget {
	return &domain.Budget{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		Code:                 row.Code,
		NameEn:               row.NameEn,
		NameAr:               row.NameAr,
		DescriptionEn:        fromPgText(row.DescriptionEn),
		DescriptionAr:        fromPgText(row.DescriptionAr),
		FiscalYearID:         toUUID(row.FiscalYearID),
		CurrencyCode:         row.CurrencyCode,
		BudgetType:           domain.BudgetType(row.BudgetType),
		Status:               domain.BudgetStatus(row.Status),
		TotalBudgetedAmount:  fromPgNumeric(row.TotalBudgetedAmount),
		TotalActualAmount:    fromPgNumeric(row.TotalActualAmount),
		TotalCommittedAmount: fromPgNumeric(row.TotalCommittedAmount),
		CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
		ApprovedByUserID:     toUUIDPtr(row.ApprovedByUserID),
		ApprovedAt:           fromPgTimestamptzPtr(row.ApprovedAt),
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:            fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapBudgetVersionToDomain(row sqlc.BudgetingBudgetVersions) *domain.BudgetVersion {
	return &domain.BudgetVersion{
		ID:               toUUID(row.ID),
		BudgetID:         toUUID(row.BudgetID),
		VersionNumber:    row.VersionNumber,
		VersionName:      row.VersionName,
		Status:           domain.VersionStatus(row.Status),
		Notes:            fromPgText(row.Notes),
		CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
		ApprovedByUserID: toUUIDPtr(row.ApprovedByUserID),
		ApprovedAt:       fromPgTimestamptzPtr(row.ApprovedAt),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:        fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapBudgetLineToDomain(row sqlc.BudgetingBudgetLines) *domain.BudgetLine {
	return &domain.BudgetLine{
		ID:                toUUID(row.ID),
		BudgetVersionID:   toUUID(row.BudgetVersionID),
		AccountID:         toUUID(row.AccountID),
		DepartmentID:      toUUIDPtr(row.DepartmentID),
		CostCenterID:      toUUIDPtr(row.CostCenterID),
		ProjectCode:       fromPgText(row.ProjectCode),
		FiscalPeriodID:    toUUIDPtr(row.FiscalPeriodID),
		Amount:            fromPgNumeric(row.Amount),
		ActualAmount:      fromPgNumeric(row.ActualAmount),
		CommittedAmount:   fromPgNumeric(row.CommittedAmount),
		LineDescriptionEn: fromPgText(row.LineDescriptionEn),
		LineDescriptionAr: fromPgText(row.LineDescriptionAr),
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapBudgetLineRowToDomain(row sqlc.ListBudgetLinesRow) *domain.BudgetLine {
	var pNum *int16
	if row.PeriodNumber.Valid {
		v := row.PeriodNumber.Int16
		pNum = &v
	}
	return &domain.BudgetLine{
		ID:                toUUID(row.ID),
		BudgetVersionID:   toUUID(row.BudgetVersionID),
		AccountID:         toUUID(row.AccountID),
		DepartmentID:      toUUIDPtr(row.DepartmentID),
		CostCenterID:      toUUIDPtr(row.CostCenterID),
		ProjectCode:       fromPgText(row.ProjectCode),
		FiscalPeriodID:    toUUIDPtr(row.FiscalPeriodID),
		Amount:            fromPgNumeric(row.Amount),
		ActualAmount:      fromPgNumeric(row.ActualAmount),
		CommittedAmount:   fromPgNumeric(row.CommittedAmount),
		LineDescriptionEn: fromPgText(row.LineDescriptionEn),
		LineDescriptionAr: fromPgText(row.LineDescriptionAr),
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptz(row.UpdatedAt),
		AccountCode:       row.AccountCode,
		AccountNameEn:     row.AccountNameEn,
		AccountNameAr:     row.AccountNameAr,
		DepartmentNameEn:  fromPgText(row.DepartmentNameEn),
		DepartmentNameAr:  fromPgText(row.DepartmentNameAr),
		CostCenterNameEn:  fromPgText(row.CostCenterNameEn),
		CostCenterNameAr:  fromPgText(row.CostCenterNameAr),
		PeriodNumber:      pNum,
		PeriodNameEn:      fromPgText(row.PeriodNameEn),
		PeriodNameAr:      fromPgText(row.PeriodNameAr),
	}
}

func mapBudgetTransferToDomain(row sqlc.BudgetingBudgetTransfers) *domain.BudgetTransfer {
	return &domain.BudgetTransfer{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		TransferNumber:      row.TransferNumber,
		SourceLineID:        toUUID(row.SourceLineID),
		TargetLineID:        toUUID(row.TargetLineID),
		Amount:              fromPgNumeric(row.Amount),
		Reason:              row.Reason,
		Status:              domain.TransferStatus(row.Status),
		TransferredByUserID: toUUIDPtr(row.TransferredByUserID),
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}
}

// Ensure interface compliance
var _ ports.Repository = (*Repository)(nil)
