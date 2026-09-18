package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"erp/internal/payroll/domain"
	"erp/internal/payroll/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) ports.Repository {
	return &Repository{db: db}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// 1. Components
func (r *Repository) CreateComponent(ctx context.Context, comp domain.SalaryComponent) (*domain.SalaryComponent, error) {
	row, err := r.queries(ctx).CreateSalaryComponent(ctx, sqlc.CreateSalaryComponentParams{
		ID:                   toPgUUID(comp.ID),
		LegalEntityID:        toPgUUID(comp.LegalEntityID),
		Code:                 comp.Code,
		NameEn:               comp.NameEn,
		NameAr:               comp.NameAr,
		ComponentType:        string(comp.ComponentType),
		IsTaxableOrGosiBasis: comp.IsTaxableOrGosiBasis,
		CalculationType:      comp.CalculationType,
		FormulaExpression:    toPgTextPtr(comp.FormulaExpression),
		DefaultAmount:        toPgNumeric(comp.DefaultAmount),
		IsActive:             comp.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	return &domain.SalaryComponent{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		Code:                 row.Code,
		NameEn:               row.NameEn,
		NameAr:               row.NameAr,
		ComponentType:        domain.ComponentType(row.ComponentType),
		IsTaxableOrGosiBasis: row.IsTaxableOrGosiBasis,
		CalculationType:      row.CalculationType,
		FormulaExpression:    fromPgTextPtr(row.FormulaExpression),
		DefaultAmount:        fromPgNumeric(row.DefaultAmount),
		IsActive:             row.IsActive,
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetComponentByID(ctx context.Context, id uuid.UUID) (*domain.SalaryComponent, error) {
	row, err := r.queries(ctx).GetSalaryComponentByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return &domain.SalaryComponent{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		Code:                 row.Code,
		NameEn:               row.NameEn,
		NameAr:               row.NameAr,
		ComponentType:        domain.ComponentType(row.ComponentType),
		IsTaxableOrGosiBasis: row.IsTaxableOrGosiBasis,
		CalculationType:      row.CalculationType,
		FormulaExpression:    fromPgTextPtr(row.FormulaExpression),
		DefaultAmount:        fromPgNumeric(row.DefaultAmount),
		IsActive:             row.IsActive,
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListComponents(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryComponent, error) {
	rows, err := r.queries(ctx).ListSalaryComponents(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	var res []domain.SalaryComponent
	for _, row := range rows {
		res = append(res, domain.SalaryComponent{
			ID:                   toUUID(row.ID),
			LegalEntityID:        toUUID(row.LegalEntityID),
			Code:                 row.Code,
			NameEn:               row.NameEn,
			NameAr:               row.NameAr,
			ComponentType:        domain.ComponentType(row.ComponentType),
			IsTaxableOrGosiBasis: row.IsTaxableOrGosiBasis,
			CalculationType:      row.CalculationType,
			FormulaExpression:    fromPgTextPtr(row.FormulaExpression),
			DefaultAmount:        fromPgNumeric(row.DefaultAmount),
			IsActive:             row.IsActive,
			CreatedAt:            fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

// 2. Structures
func (r *Repository) CreateStructure(ctx context.Context, structObj domain.SalaryStructure) (*domain.SalaryStructure, error) {
	row, err := r.queries(ctx).CreateSalaryStructure(ctx, sqlc.CreateSalaryStructureParams{
		ID:            toPgUUID(structObj.ID),
		LegalEntityID: toPgUUID(structObj.LegalEntityID),
		Code:          structObj.Code,
		NameEn:        structObj.NameEn,
		NameAr:        structObj.NameAr,
		CurrencyCode:  structObj.CurrencyCode,
		IsActive:      structObj.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create structure: %w", err)
	}

	return &domain.SalaryStructure{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		CurrencyCode:  row.CurrencyCode,
		IsActive:      row.IsActive,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetStructureByID(ctx context.Context, id uuid.UUID) (*domain.SalaryStructure, error) {
	row, err := r.queries(ctx).GetSalaryStructureByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return &domain.SalaryStructure{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		CurrencyCode:  row.CurrencyCode,
		IsActive:      row.IsActive,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListStructures(ctx context.Context, legalEntityID uuid.UUID) ([]domain.SalaryStructure, error) {
	rows, err := r.queries(ctx).ListSalaryStructures(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	var res []domain.SalaryStructure
	for _, row := range rows {
		res = append(res, domain.SalaryStructure{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			CurrencyCode:  row.CurrencyCode,
			IsActive:      row.IsActive,
			CreatedAt:     fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) CreateStructureLine(ctx context.Context, line domain.SalaryStructureLine) (*domain.SalaryStructureLine, error) {
	row, err := r.queries(ctx).CreateSalaryStructureLine(ctx, sqlc.CreateSalaryStructureLineParams{
		ID:                toPgUUID(line.ID),
		SalaryStructureID: toPgUUID(line.SalaryStructureID),
		ComponentID:       toPgUUID(line.ComponentID),
		LineNumber:        int32(line.LineNumber),
		RatePercentage:    toPgNumeric(line.RatePercentage),
		FixedAmount:       toPgNumeric(line.FixedAmount),
	})
	if err != nil {
		return nil, err
	}
	return &domain.SalaryStructureLine{
		ID:                toUUID(row.ID),
		SalaryStructureID: toUUID(row.SalaryStructureID),
		ComponentID:       toUUID(row.ComponentID),
		LineNumber:        int(row.LineNumber),
		RatePercentage:    fromPgNumeric(row.RatePercentage),
		FixedAmount:       fromPgNumeric(row.FixedAmount),
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListStructureLines(ctx context.Context, structureID uuid.UUID) ([]domain.SalaryStructureLine, error) {
	rows, err := r.queries(ctx).ListSalaryStructureLines(ctx, toPgUUID(structureID))
	if err != nil {
		return nil, err
	}
	var res []domain.SalaryStructureLine
	for _, row := range rows {
		res = append(res, domain.SalaryStructureLine{
			ID:                toUUID(row.ID),
			SalaryStructureID: toUUID(row.SalaryStructureID),
			ComponentID:       toUUID(row.ComponentID),
			LineNumber:        int(row.LineNumber),
			RatePercentage:    fromPgNumeric(row.RatePercentage),
			FixedAmount:       fromPgNumeric(row.FixedAmount),
			CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

// 3. Salary Assignment
func (r *Repository) CreateSalaryAssignment(ctx context.Context, assign domain.EmployeeSalaryAssignment) (*domain.EmployeeSalaryAssignment, error) {
	row, err := r.queries(ctx).CreateEmployeeSalaryAssignment(ctx, sqlc.CreateEmployeeSalaryAssignmentParams{
		ID:                 toPgUUID(assign.ID),
		LegalEntityID:      toPgUUID(assign.LegalEntityID),
		EmployeeID:         toPgUUID(assign.EmployeeID),
		SalaryStructureID:  toPgUUID(assign.SalaryStructureID),
		EffectiveFrom:      toPgDate(assign.EffectiveFrom),
		EffectiveTo:        toPgDatePtr(assign.EffectiveTo),
		BasicSalary:        toPgNumeric(assign.BasicSalary),
		HousingAllowance:   toPgNumeric(assign.HousingAllowance),
		TransportAllowance: toPgNumeric(assign.TransportAllowance),
		OtherAllowances:    toPgNumeric(assign.OtherAllowances),
		GrossSalary:        toPgNumeric(assign.GrossSalary),
	})
	if err != nil {
		return nil, err
	}
	return &domain.EmployeeSalaryAssignment{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		EmployeeID:         toUUID(row.EmployeeID),
		SalaryStructureID:  toUUID(row.SalaryStructureID),
		EffectiveFrom:      fromPgDate(row.EffectiveFrom),
		EffectiveTo:        fromPgDatePtr(row.EffectiveTo),
		BasicSalary:        fromPgNumeric(row.BasicSalary),
		HousingAllowance:   fromPgNumeric(row.HousingAllowance),
		TransportAllowance: fromPgNumeric(row.TransportAllowance),
		OtherAllowances:    fromPgNumeric(row.OtherAllowances),
		GrossSalary:        fromPgNumeric(row.GrossSalary),
		CreatedAt:          fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetActiveSalaryAssignment(ctx context.Context, employeeID uuid.UUID, asOfDate time.Time) (*domain.EmployeeSalaryAssignment, error) {
	row, err := r.queries(ctx).GetActiveEmployeeSalaryAssignment(ctx, sqlc.GetActiveEmployeeSalaryAssignmentParams{
		EmployeeID:  toPgUUID(employeeID),
		EffectiveTo: toPgDate(asOfDate),
	})
	if err != nil {
		return nil, err
	}
	return &domain.EmployeeSalaryAssignment{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		EmployeeID:         toUUID(row.EmployeeID),
		SalaryStructureID:  toUUID(row.SalaryStructureID),
		EffectiveFrom:      fromPgDate(row.EffectiveFrom),
		EffectiveTo:        fromPgDatePtr(row.EffectiveTo),
		BasicSalary:        fromPgNumeric(row.BasicSalary),
		HousingAllowance:   fromPgNumeric(row.HousingAllowance),
		TransportAllowance: fromPgNumeric(row.TransportAllowance),
		OtherAllowances:    fromPgNumeric(row.OtherAllowances),
		GrossSalary:        fromPgNumeric(row.GrossSalary),
		CreatedAt:          fromPgTimestamptz(row.CreatedAt),
	}, nil
}

// 4. Periods & Runs
func (r *Repository) CreatePeriod(ctx context.Context, period domain.PayrollPeriod) (*domain.PayrollPeriod, error) {
	row, err := r.queries(ctx).CreatePayrollPeriod(ctx, sqlc.CreatePayrollPeriodParams{
		ID:             toPgUUID(period.ID),
		LegalEntityID:  toPgUUID(period.LegalEntityID),
		FiscalPeriodID: toPgUUIDPtr(period.FiscalPeriodID),
		PeriodNameEn:   period.PeriodNameEn,
		PeriodNameAr:   period.PeriodNameAr,
		PeriodYear:     int32(period.PeriodYear),
		PeriodMonth:    int32(period.PeriodMonth),
		StartDate:      toPgDate(period.StartDate),
		EndDate:        toPgDate(period.EndDate),
		PaymentDueDate: toPgDate(period.PaymentDueDate),
		IsClosed:       period.IsClosed,
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayrollPeriod{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		FiscalPeriodID: toUUIDPtr(row.FiscalPeriodID),
		PeriodNameEn:   row.PeriodNameEn,
		PeriodNameAr:   row.PeriodNameAr,
		PeriodYear:     int(row.PeriodYear),
		PeriodMonth:    int(row.PeriodMonth),
		StartDate:      fromPgDate(row.StartDate),
		EndDate:        fromPgDate(row.EndDate),
		PaymentDueDate: fromPgDate(row.PaymentDueDate),
		IsClosed:       row.IsClosed,
		CreatedAt:      fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetPeriodByID(ctx context.Context, id uuid.UUID) (*domain.PayrollPeriod, error) {
	row, err := r.queries(ctx).GetPayrollPeriodByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return &domain.PayrollPeriod{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		FiscalPeriodID: toUUIDPtr(row.FiscalPeriodID),
		PeriodNameEn:   row.PeriodNameEn,
		PeriodNameAr:   row.PeriodNameAr,
		PeriodYear:     int(row.PeriodYear),
		PeriodMonth:    int(row.PeriodMonth),
		StartDate:      fromPgDate(row.StartDate),
		EndDate:        fromPgDate(row.EndDate),
		PaymentDueDate: fromPgDate(row.PaymentDueDate),
		IsClosed:       row.IsClosed,
		CreatedAt:      fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollPeriod, error) {
	rows, err := r.queries(ctx).ListPayrollPeriods(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	var res []domain.PayrollPeriod
	for _, row := range rows {
		res = append(res, domain.PayrollPeriod{
			ID:             toUUID(row.ID),
			LegalEntityID:  toUUID(row.LegalEntityID),
			FiscalPeriodID: toUUIDPtr(row.FiscalPeriodID),
			PeriodNameEn:   row.PeriodNameEn,
			PeriodNameAr:   row.PeriodNameAr,
			PeriodYear:     int(row.PeriodYear),
			PeriodMonth:    int(row.PeriodMonth),
			StartDate:      fromPgDate(row.StartDate),
			EndDate:        fromPgDate(row.EndDate),
			PaymentDueDate: fromPgDate(row.PaymentDueDate),
			IsClosed:       row.IsClosed,
			CreatedAt:      fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) CreateRun(ctx context.Context, run domain.PayrollRun) (*domain.PayrollRun, error) {
	row, err := r.queries(ctx).CreatePayrollRun(ctx, sqlc.CreatePayrollRunParams{
		ID:              toPgUUID(run.ID),
		LegalEntityID:   toPgUUID(run.LegalEntityID),
		PeriodID:        toPgUUID(run.PeriodID),
		RunNumber:       run.RunNumber,
		RunDate:         toPgDate(run.RunDate),
		CutoffDate:      toPgDate(run.CutoffDate),
		Status:          string(run.Status),
		TotalBasic:      toPgNumeric(run.TotalBasic),
		TotalAllowances: toPgNumeric(run.TotalAllowances),
		TotalGross:      toPgNumeric(run.TotalGross),
		TotalDeductions: toPgNumeric(run.TotalDeductions),
		TotalNet:        toPgNumeric(run.TotalNet),
		EmployeeCount:   int32(run.EmployeeCount),
		CreatedBy:       toPgUUIDPtr(run.CreatedBy),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayrollRun{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		PeriodID:         toUUID(row.PeriodID),
		RunNumber:        row.RunNumber,
		RunDate:          fromPgDate(row.RunDate),
		CutoffDate:       fromPgDate(row.CutoffDate),
		Status:           domain.RunStatus(row.Status),
		TotalBasic:       fromPgNumeric(row.TotalBasic),
		TotalAllowances:  fromPgNumeric(row.TotalAllowances),
		TotalGross:       fromPgNumeric(row.TotalGross),
		TotalDeductions:  fromPgNumeric(row.TotalDeductions),
		TotalNet:         fromPgNumeric(row.TotalNet),
		EmployeeCount:    int(row.EmployeeCount),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		WpsFileGenerated: row.WpsFileGenerated,
		WpsFileName:      fromPgTextPtr(row.WpsFileName),
		PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		CreatedBy:        toUUIDPtr(row.CreatedBy),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetRunByID(ctx context.Context, id uuid.UUID) (*domain.PayrollRun, error) {
	row, err := r.queries(ctx).GetPayrollRunByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return &domain.PayrollRun{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		PeriodID:         toUUID(row.PeriodID),
		RunNumber:        row.RunNumber,
		RunDate:          fromPgDate(row.RunDate),
		CutoffDate:       fromPgDate(row.CutoffDate),
		Status:           domain.RunStatus(row.Status),
		TotalBasic:       fromPgNumeric(row.TotalBasic),
		TotalAllowances:  fromPgNumeric(row.TotalAllowances),
		TotalGross:       fromPgNumeric(row.TotalGross),
		TotalDeductions:  fromPgNumeric(row.TotalDeductions),
		TotalNet:         fromPgNumeric(row.TotalNet),
		EmployeeCount:    int(row.EmployeeCount),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		WpsFileGenerated: row.WpsFileGenerated,
		WpsFileName:      fromPgTextPtr(row.WpsFileName),
		PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		CreatedBy:        toUUIDPtr(row.CreatedBy),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.PayrollRun, error) {
	rows, err := r.queries(ctx).ListPayrollRuns(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	var res []domain.PayrollRun
	for _, row := range rows {
		res = append(res, domain.PayrollRun{
			ID:               toUUID(row.ID),
			LegalEntityID:    toUUID(row.LegalEntityID),
			PeriodID:         toUUID(row.PeriodID),
			RunNumber:        row.RunNumber,
			RunDate:          fromPgDate(row.RunDate),
			CutoffDate:       fromPgDate(row.CutoffDate),
			Status:           domain.RunStatus(row.Status),
			TotalBasic:       fromPgNumeric(row.TotalBasic),
			TotalAllowances:  fromPgNumeric(row.TotalAllowances),
			TotalGross:       fromPgNumeric(row.TotalGross),
			TotalDeductions:  fromPgNumeric(row.TotalDeductions),
			TotalNet:         fromPgNumeric(row.TotalNet),
			EmployeeCount:    int(row.EmployeeCount),
			JournalEntryID:   toUUIDPtr(row.JournalEntryID),
			WpsFileGenerated: row.WpsFileGenerated,
			WpsFileName:      fromPgTextPtr(row.WpsFileName),
			PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
			PostedBy:         toUUIDPtr(row.PostedBy),
			CreatedBy:        toUUIDPtr(row.CreatedBy),
			CreatedAt:        fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) UpdateRunStatus(ctx context.Context, id uuid.UUID, status domain.RunStatus, postedBy *uuid.UUID) (*domain.PayrollRun, error) {
	row, err := r.queries(ctx).UpdatePayrollRunStatus(ctx, sqlc.UpdatePayrollRunStatusParams{
		ID:       toPgUUID(id),
		Status:   string(status),
		PostedBy: toPgUUIDPtr(postedBy),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayrollRun{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		PeriodID:         toUUID(row.PeriodID),
		RunNumber:        row.RunNumber,
		RunDate:          fromPgDate(row.RunDate),
		CutoffDate:       fromPgDate(row.CutoffDate),
		Status:           domain.RunStatus(row.Status),
		TotalBasic:       fromPgNumeric(row.TotalBasic),
		TotalAllowances:  fromPgNumeric(row.TotalAllowances),
		TotalGross:       fromPgNumeric(row.TotalGross),
		TotalDeductions:  fromPgNumeric(row.TotalDeductions),
		TotalNet:         fromPgNumeric(row.TotalNet),
		EmployeeCount:    int(row.EmployeeCount),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		WpsFileGenerated: row.WpsFileGenerated,
		WpsFileName:      fromPgTextPtr(row.WpsFileName),
		PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		CreatedBy:        toUUIDPtr(row.CreatedBy),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) UpdateRunTotals(ctx context.Context, id uuid.UUID, basic, allow, gross, ded, net decimal.Decimal, count int) (*domain.PayrollRun, error) {
	row, err := r.queries(ctx).UpdatePayrollRunTotals(ctx, sqlc.UpdatePayrollRunTotalsParams{
		ID:              toPgUUID(id),
		TotalBasic:      toPgNumeric(basic),
		TotalAllowances: toPgNumeric(allow),
		TotalGross:      toPgNumeric(gross),
		TotalDeductions: toPgNumeric(ded),
		TotalNet:        toPgNumeric(net),
		EmployeeCount:   int32(count),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayrollRun{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		PeriodID:         toUUID(row.PeriodID),
		RunNumber:        row.RunNumber,
		RunDate:          fromPgDate(row.RunDate),
		CutoffDate:       fromPgDate(row.CutoffDate),
		Status:           domain.RunStatus(row.Status),
		TotalBasic:       fromPgNumeric(row.TotalBasic),
		TotalAllowances:  fromPgNumeric(row.TotalAllowances),
		TotalGross:       fromPgNumeric(row.TotalGross),
		TotalDeductions:  fromPgNumeric(row.TotalDeductions),
		TotalNet:         fromPgNumeric(row.TotalNet),
		EmployeeCount:    int(row.EmployeeCount),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		WpsFileGenerated: row.WpsFileGenerated,
		WpsFileName:      fromPgTextPtr(row.WpsFileName),
		PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		CreatedBy:        toUUIDPtr(row.CreatedBy),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) LinkRunJournalEntry(ctx context.Context, id uuid.UUID, journalEntryID uuid.UUID) (*domain.PayrollRun, error) {
	row, err := r.queries(ctx).LinkPayrollRunJournalEntry(ctx, sqlc.LinkPayrollRunJournalEntryParams{
		ID:             toPgUUID(id),
		JournalEntryID: toPgUUID(journalEntryID),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayrollRun{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		PeriodID:         toUUID(row.PeriodID),
		RunNumber:        row.RunNumber,
		RunDate:          fromPgDate(row.RunDate),
		CutoffDate:       fromPgDate(row.CutoffDate),
		Status:           domain.RunStatus(row.Status),
		TotalBasic:       fromPgNumeric(row.TotalBasic),
		TotalAllowances:  fromPgNumeric(row.TotalAllowances),
		TotalGross:       fromPgNumeric(row.TotalGross),
		TotalDeductions:  fromPgNumeric(row.TotalDeductions),
		TotalNet:         fromPgNumeric(row.TotalNet),
		EmployeeCount:    int(row.EmployeeCount),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		WpsFileGenerated: row.WpsFileGenerated,
		WpsFileName:      fromPgTextPtr(row.WpsFileName),
		PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		CreatedBy:        toUUIDPtr(row.CreatedBy),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) UpdateRunWpsStatus(ctx context.Context, id uuid.UUID, fileName string) (*domain.PayrollRun, error) {
	row, err := r.queries(ctx).UpdatePayrollRunWpsStatus(ctx, sqlc.UpdatePayrollRunWpsStatusParams{
		ID:          toPgUUID(id),
		WpsFileName: toPgText(fileName),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayrollRun{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		PeriodID:         toUUID(row.PeriodID),
		RunNumber:        row.RunNumber,
		RunDate:          fromPgDate(row.RunDate),
		CutoffDate:       fromPgDate(row.CutoffDate),
		Status:           domain.RunStatus(row.Status),
		TotalBasic:       fromPgNumeric(row.TotalBasic),
		TotalAllowances:  fromPgNumeric(row.TotalAllowances),
		TotalGross:       fromPgNumeric(row.TotalGross),
		TotalDeductions:  fromPgNumeric(row.TotalDeductions),
		TotalNet:         fromPgNumeric(row.TotalNet),
		EmployeeCount:    int(row.EmployeeCount),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		WpsFileGenerated: row.WpsFileGenerated,
		WpsFileName:      fromPgTextPtr(row.WpsFileName),
		PostedAt:         fromPgTimestamptzPtr(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		CreatedBy:        toUUIDPtr(row.CreatedBy),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}, nil
}

// 5. Run Employees & Lines
func (r *Repository) CreateRunEmployee(ctx context.Context, emp domain.RunEmployee) (*domain.RunEmployee, error) {
	row, err := r.queries(ctx).CreatePayrollRunEmployee(ctx, sqlc.CreatePayrollRunEmployeeParams{
		ID:                    toPgUUID(emp.ID),
		RunID:                 toPgUUID(emp.RunID),
		EmployeeID:            toPgUUID(emp.EmployeeID),
		ContractID:            toPgUUIDPtr(emp.ContractID),
		BasicSalary:           toPgNumeric(emp.BasicSalary),
		HousingAllowance:      toPgNumeric(emp.HousingAllowance),
		TransportAllowance:    toPgNumeric(emp.TransportAllowance),
		OtherAllowances:       toPgNumeric(emp.OtherAllowances),
		OvertimeAmount:        toPgNumeric(emp.OvertimeAmount),
		GrossSalary:           toPgNumeric(emp.GrossSalary),
		GosiEmployeeDeduction: toPgNumeric(emp.GosiEmployeeDeduction),
		LoanDeduction:         toPgNumeric(emp.LoanDeduction),
		AbsenceDeduction:      toPgNumeric(emp.AbsenceDeduction),
		OtherDeductions:       toPgNumeric(emp.OtherDeductions),
		TotalDeductions:       toPgNumeric(emp.TotalDeductions),
		NetSalary:             toPgNumeric(emp.NetSalary),
		PaymentStatus:         emp.PaymentStatus,
	})
	if err != nil {
		return nil, err
	}
	return &domain.RunEmployee{
		ID:                    toUUID(row.ID),
		RunID:                 toUUID(row.RunID),
		EmployeeID:            toUUID(row.EmployeeID),
		ContractID:            toUUIDPtr(row.ContractID),
		BasicSalary:           fromPgNumeric(row.BasicSalary),
		HousingAllowance:      fromPgNumeric(row.HousingAllowance),
		TransportAllowance:    fromPgNumeric(row.TransportAllowance),
		OtherAllowances:       fromPgNumeric(row.OtherAllowances),
		OvertimeAmount:        fromPgNumeric(row.OvertimeAmount),
		GrossSalary:           fromPgNumeric(row.GrossSalary),
		GosiEmployeeDeduction: fromPgNumeric(row.GosiEmployeeDeduction),
		LoanDeduction:         fromPgNumeric(row.LoanDeduction),
		AbsenceDeduction:      fromPgNumeric(row.AbsenceDeduction),
		OtherDeductions:       fromPgNumeric(row.OtherDeductions),
		TotalDeductions:       fromPgNumeric(row.TotalDeductions),
		NetSalary:             fromPgNumeric(row.NetSalary),
		PaymentStatus:         row.PaymentStatus,
		CreatedAt:             fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListRunEmployees(ctx context.Context, runID uuid.UUID) ([]domain.RunEmployee, error) {
	rows, err := r.queries(ctx).ListPayrollRunEmployees(ctx, toPgUUID(runID))
	if err != nil {
		return nil, err
	}
	var res []domain.RunEmployee
	for _, row := range rows {
		res = append(res, domain.RunEmployee{
			ID:                    toUUID(row.ID),
			RunID:                 toUUID(row.RunID),
			EmployeeID:            toUUID(row.EmployeeID),
			ContractID:            toUUIDPtr(row.ContractID),
			BasicSalary:           fromPgNumeric(row.BasicSalary),
			HousingAllowance:      fromPgNumeric(row.HousingAllowance),
			TransportAllowance:    fromPgNumeric(row.TransportAllowance),
			OtherAllowances:       fromPgNumeric(row.OtherAllowances),
			OvertimeAmount:        fromPgNumeric(row.OvertimeAmount),
			GrossSalary:           fromPgNumeric(row.GrossSalary),
			GosiEmployeeDeduction: fromPgNumeric(row.GosiEmployeeDeduction),
			LoanDeduction:         fromPgNumeric(row.LoanDeduction),
			AbsenceDeduction:      fromPgNumeric(row.AbsenceDeduction),
			OtherDeductions:       fromPgNumeric(row.OtherDeductions),
			TotalDeductions:       fromPgNumeric(row.TotalDeductions),
			NetSalary:             fromPgNumeric(row.NetSalary),
			PaymentStatus:         row.PaymentStatus,
			CreatedAt:             fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) GetRunEmployeeByID(ctx context.Context, id uuid.UUID) (*domain.RunEmployee, error) {
	row, err := r.queries(ctx).GetPayrollRunEmployeeByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return &domain.RunEmployee{
		ID:                    toUUID(row.ID),
		RunID:                 toUUID(row.RunID),
		EmployeeID:            toUUID(row.EmployeeID),
		ContractID:            toUUIDPtr(row.ContractID),
		BasicSalary:           fromPgNumeric(row.BasicSalary),
		HousingAllowance:      fromPgNumeric(row.HousingAllowance),
		TransportAllowance:    fromPgNumeric(row.TransportAllowance),
		OtherAllowances:       fromPgNumeric(row.OtherAllowances),
		OvertimeAmount:        fromPgNumeric(row.OvertimeAmount),
		GrossSalary:           fromPgNumeric(row.GrossSalary),
		GosiEmployeeDeduction: fromPgNumeric(row.GosiEmployeeDeduction),
		LoanDeduction:         fromPgNumeric(row.LoanDeduction),
		AbsenceDeduction:      fromPgNumeric(row.AbsenceDeduction),
		OtherDeductions:       fromPgNumeric(row.OtherDeductions),
		TotalDeductions:       fromPgNumeric(row.TotalDeductions),
		NetSalary:             fromPgNumeric(row.NetSalary),
		PaymentStatus:         row.PaymentStatus,
		CreatedAt:             fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) DeleteRunEmployees(ctx context.Context, runID uuid.UUID) error {
	return r.queries(ctx).DeletePayrollRunEmployees(ctx, toPgUUID(runID))
}

func (r *Repository) CreateCalculationLine(ctx context.Context, line domain.CalculationLine) (*domain.CalculationLine, error) {
	row, err := r.queries(ctx).CreateCalculationLine(ctx, sqlc.CreateCalculationLineParams{
		ID:            toPgUUID(line.ID),
		RunEmployeeID: toPgUUID(line.RunEmployeeID),
		ComponentID:   toPgUUIDPtr(line.ComponentID),
		LineType:      string(line.LineType),
		Code:          line.Code,
		NameEn:        line.NameEn,
		NameAr:        line.NameAr,
		Rate:          toPgNumeric(line.Rate),
		BasisAmount:   toPgNumeric(line.BasisAmount),
		Amount:        toPgNumeric(line.Amount),
		FormulaNote:   toPgText(line.FormulaNote),
	})
	if err != nil {
		return nil, err
	}
	return &domain.CalculationLine{
		ID:            toUUID(row.ID),
		RunEmployeeID: toUUID(row.RunEmployeeID),
		ComponentID:   toUUIDPtr(row.ComponentID),
		LineType:      domain.ComponentType(row.LineType),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		Rate:          fromPgNumeric(row.Rate),
		BasisAmount:   fromPgNumeric(row.BasisAmount),
		Amount:        fromPgNumeric(row.Amount),
		FormulaNote:   fromPgText(row.FormulaNote),
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListCalculationLines(ctx context.Context, runEmployeeID uuid.UUID) ([]domain.CalculationLine, error) {
	rows, err := r.queries(ctx).ListCalculationLinesByRunEmployee(ctx, toPgUUID(runEmployeeID))
	if err != nil {
		return nil, err
	}
	var res []domain.CalculationLine
	for _, row := range rows {
		res = append(res, domain.CalculationLine{
			ID:            toUUID(row.ID),
			RunEmployeeID: toUUID(row.RunEmployeeID),
			ComponentID:   toUUIDPtr(row.ComponentID),
			LineType:      domain.ComponentType(row.LineType),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			Rate:          fromPgNumeric(row.Rate),
			BasisAmount:   fromPgNumeric(row.BasisAmount),
			Amount:        fromPgNumeric(row.Amount),
			FormulaNote:   fromPgText(row.FormulaNote),
			CreatedAt:     fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

// 6. Payslips
func (r *Repository) CreatePayslip(ctx context.Context, slip domain.Payslip) (*domain.Payslip, error) {
	snapBytes, _ := json.Marshal(slip.CalculationSnapshot)
	row, err := r.queries(ctx).CreatePayslip(ctx, sqlc.CreatePayslipParams{
		ID:                  toPgUUID(slip.ID),
		LegalEntityID:       toPgUUID(slip.LegalEntityID),
		RunEmployeeID:       toPgUUID(slip.RunEmployeeID),
		PayslipNumber:       slip.PayslipNumber,
		EmployeeID:          toPgUUID(slip.EmployeeID),
		PeriodID:            toPgUUID(slip.PeriodID),
		IssueDate:           toPgDate(slip.IssueDate),
		GrossSalary:         toPgNumeric(slip.GrossSalary),
		TotalDeductions:     toPgNumeric(slip.TotalDeductions),
		NetSalary:           toPgNumeric(slip.NetSalary),
		CalculationSnapshot: snapBytes,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Payslip{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		RunEmployeeID:       toUUID(row.RunEmployeeID),
		PayslipNumber:       row.PayslipNumber,
		EmployeeID:          toUUID(row.EmployeeID),
		PeriodID:            toUUID(row.PeriodID),
		IssueDate:           fromPgDate(row.IssueDate),
		GrossSalary:         fromPgNumeric(row.GrossSalary),
		TotalDeductions:     fromPgNumeric(row.TotalDeductions),
		NetSalary:           fromPgNumeric(row.NetSalary),
		CalculationSnapshot: slip.CalculationSnapshot,
		IsViewedByEmployee:  row.IsViewedByEmployee,
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetPayslipByID(ctx context.Context, id uuid.UUID) (*domain.Payslip, error) {
	row, err := r.queries(ctx).GetPayslipByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	var snap map[string]any
	_ = json.Unmarshal(row.CalculationSnapshot, &snap)
	return &domain.Payslip{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		RunEmployeeID:       toUUID(row.RunEmployeeID),
		PayslipNumber:       row.PayslipNumber,
		EmployeeID:          toUUID(row.EmployeeID),
		PeriodID:            toUUID(row.PeriodID),
		IssueDate:           fromPgDate(row.IssueDate),
		GrossSalary:         fromPgNumeric(row.GrossSalary),
		TotalDeductions:     fromPgNumeric(row.TotalDeductions),
		NetSalary:           fromPgNumeric(row.NetSalary),
		CalculationSnapshot: snap,
		IsViewedByEmployee:  row.IsViewedByEmployee,
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListPayslipsByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.Payslip, error) {
	rows, err := r.queries(ctx).ListPayslipsByPeriod(ctx, toPgUUID(periodID))
	if err != nil {
		return nil, err
	}
	var res []domain.Payslip
	for _, row := range rows {
		var snap map[string]any
		_ = json.Unmarshal(row.CalculationSnapshot, &snap)
		res = append(res, domain.Payslip{
			ID:                  toUUID(row.ID),
			LegalEntityID:       toUUID(row.LegalEntityID),
			RunEmployeeID:       toUUID(row.RunEmployeeID),
			PayslipNumber:       row.PayslipNumber,
			EmployeeID:          toUUID(row.EmployeeID),
			PeriodID:            toUUID(row.PeriodID),
			IssueDate:           fromPgDate(row.IssueDate),
			GrossSalary:         fromPgNumeric(row.GrossSalary),
			TotalDeductions:     fromPgNumeric(row.TotalDeductions),
			NetSalary:           fromPgNumeric(row.NetSalary),
			CalculationSnapshot: snap,
			IsViewedByEmployee:  row.IsViewedByEmployee,
			CreatedAt:           fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) ListPayslipsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]domain.Payslip, error) {
	rows, err := r.queries(ctx).ListPayslipsByEmployee(ctx, toPgUUID(employeeID))
	if err != nil {
		return nil, err
	}
	var res []domain.Payslip
	for _, row := range rows {
		var snap map[string]any
		_ = json.Unmarshal(row.CalculationSnapshot, &snap)
		res = append(res, domain.Payslip{
			ID:                  toUUID(row.ID),
			LegalEntityID:       toUUID(row.LegalEntityID),
			RunEmployeeID:       toUUID(row.RunEmployeeID),
			PayslipNumber:       row.PayslipNumber,
			EmployeeID:          toUUID(row.EmployeeID),
			PeriodID:            toUUID(row.PeriodID),
			IssueDate:           fromPgDate(row.IssueDate),
			GrossSalary:         fromPgNumeric(row.GrossSalary),
			TotalDeductions:     fromPgNumeric(row.TotalDeductions),
			NetSalary:           fromPgNumeric(row.NetSalary),
			CalculationSnapshot: snap,
			IsViewedByEmployee:  row.IsViewedByEmployee,
			CreatedAt:           fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

// 7. Loans
func (r *Repository) CreateLoan(ctx context.Context, loan domain.EmployeeLoan) (*domain.EmployeeLoan, error) {
	row, err := r.queries(ctx).CreateEmployeeLoan(ctx, sqlc.CreateEmployeeLoanParams{
		ID:                 toPgUUID(loan.ID),
		LegalEntityID:      toPgUUID(loan.LegalEntityID),
		EmployeeID:         toPgUUID(loan.EmployeeID),
		LoanNumber:         loan.LoanNumber,
		PrincipalAmount:    toPgNumeric(loan.PrincipalAmount),
		MonthlyInstallment: toPgNumeric(loan.MonthlyInstallment),
		RemainingBalance:   toPgNumeric(loan.RemainingBalance),
		DisbursementDate:   toPgDate(loan.DisbursementDate),
		Status:             loan.Status,
		Reason:             toPgTextPtr(loan.Reason),
	})
	if err != nil {
		return nil, err
	}
	return &domain.EmployeeLoan{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		EmployeeID:         toUUID(row.EmployeeID),
		LoanNumber:         row.LoanNumber,
		PrincipalAmount:    fromPgNumeric(row.PrincipalAmount),
		MonthlyInstallment: fromPgNumeric(row.MonthlyInstallment),
		RemainingBalance:   fromPgNumeric(row.RemainingBalance),
		DisbursementDate:   fromPgDate(row.DisbursementDate),
		Status:             row.Status,
		Reason:             fromPgTextPtr(row.Reason),
		CreatedAt:          fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetLoanByID(ctx context.Context, id uuid.UUID) (*domain.EmployeeLoan, error) {
	row, err := r.queries(ctx).GetEmployeeLoanByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return &domain.EmployeeLoan{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		EmployeeID:         toUUID(row.EmployeeID),
		LoanNumber:         row.LoanNumber,
		PrincipalAmount:    fromPgNumeric(row.PrincipalAmount),
		MonthlyInstallment: fromPgNumeric(row.MonthlyInstallment),
		RemainingBalance:   fromPgNumeric(row.RemainingBalance),
		DisbursementDate:   fromPgDate(row.DisbursementDate),
		Status:             row.Status,
		Reason:             fromPgTextPtr(row.Reason),
		CreatedAt:          fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListLoans(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EmployeeLoan, error) {
	rows, err := r.queries(ctx).ListEmployeeLoans(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	var res []domain.EmployeeLoan
	for _, row := range rows {
		res = append(res, domain.EmployeeLoan{
			ID:                 toUUID(row.ID),
			LegalEntityID:      toUUID(row.LegalEntityID),
			EmployeeID:         toUUID(row.EmployeeID),
			LoanNumber:         row.LoanNumber,
			PrincipalAmount:    fromPgNumeric(row.PrincipalAmount),
			MonthlyInstallment: fromPgNumeric(row.MonthlyInstallment),
			RemainingBalance:   fromPgNumeric(row.RemainingBalance),
			DisbursementDate:   fromPgDate(row.DisbursementDate),
			Status:             row.Status,
			Reason:             fromPgTextPtr(row.Reason),
			CreatedAt:          fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) CreateLoanInstallment(ctx context.Context, inst domain.LoanInstallment) (*domain.LoanInstallment, error) {
	row, err := r.queries(ctx).CreateLoanInstallment(ctx, sqlc.CreateLoanInstallmentParams{
		ID:                toPgUUID(inst.ID),
		LoanID:            toPgUUID(inst.LoanID),
		PayrollRunID:      toPgUUIDPtr(inst.PayrollRunID),
		InstallmentNumber: int32(inst.InstallmentNumber),
		DueDate:           toPgDate(inst.DueDate),
		Amount:            toPgNumeric(inst.Amount),
		Status:            inst.Status,
	})
	if err != nil {
		return nil, err
	}
	return &domain.LoanInstallment{
		ID:                toUUID(row.ID),
		LoanID:            toUUID(row.LoanID),
		PayrollRunID:      toUUIDPtr(row.PayrollRunID),
		InstallmentNumber: int(row.InstallmentNumber),
		DueDate:           fromPgDate(row.DueDate),
		Amount:            fromPgNumeric(row.Amount),
		Status:            row.Status,
		DeductedAt:        fromPgTimestamptzPtr(row.DeductedAt),
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListPendingLoanInstallmentsByEmployee(ctx context.Context, employeeID uuid.UUID, dueDate time.Time) ([]domain.LoanInstallment, error) {
	rows, err := r.queries(ctx).ListPendingLoanInstallmentsByEmployee(ctx, sqlc.ListPendingLoanInstallmentsByEmployeeParams{
		EmployeeID: toPgUUID(employeeID),
		DueDate:    toPgDate(dueDate),
	})
	if err != nil {
		return nil, err
	}
	var res []domain.LoanInstallment
	for _, row := range rows {
		res = append(res, domain.LoanInstallment{
			ID:                toUUID(row.ID),
			LoanID:            toUUID(row.LoanID),
			PayrollRunID:      toUUIDPtr(row.PayrollRunID),
			InstallmentNumber: int(row.InstallmentNumber),
			DueDate:           fromPgDate(row.DueDate),
			Amount:            fromPgNumeric(row.Amount),
			Status:            row.Status,
			DeductedAt:        fromPgTimestamptzPtr(row.DeductedAt),
			CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

func (r *Repository) UpdateLoanInstallmentStatus(ctx context.Context, id uuid.UUID, status string, runID *uuid.UUID) (*domain.LoanInstallment, error) {
	row, err := r.queries(ctx).UpdateLoanInstallmentStatus(ctx, sqlc.UpdateLoanInstallmentStatusParams{
		ID:           toPgUUID(id),
		Status:       status,
		PayrollRunID: toPgUUIDPtr(runID),
	})
	if err != nil {
		return nil, err
	}
	return &domain.LoanInstallment{
		ID:                toUUID(row.ID),
		LoanID:            toUUID(row.LoanID),
		PayrollRunID:      toUUIDPtr(row.PayrollRunID),
		InstallmentNumber: int(row.InstallmentNumber),
		DueDate:           fromPgDate(row.DueDate),
		Amount:            fromPgNumeric(row.Amount),
		Status:            row.Status,
		DeductedAt:        fromPgTimestamptzPtr(row.DeductedAt),
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) UpdateLoanRemainingBalance(ctx context.Context, id uuid.UUID, deductionAmount decimal.Decimal) (*domain.EmployeeLoan, error) {
	row, err := r.queries(ctx).UpdateLoanRemainingBalance(ctx, sqlc.UpdateLoanRemainingBalanceParams{
		ID:               toPgUUID(id),
		RemainingBalance: toPgNumeric(deductionAmount),
	})
	if err != nil {
		return nil, err
	}
	return &domain.EmployeeLoan{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		EmployeeID:         toUUID(row.EmployeeID),
		LoanNumber:         row.LoanNumber,
		PrincipalAmount:    fromPgNumeric(row.PrincipalAmount),
		MonthlyInstallment: fromPgNumeric(row.MonthlyInstallment),
		RemainingBalance:   fromPgNumeric(row.RemainingBalance),
		DisbursementDate:   fromPgDate(row.DisbursementDate),
		Status:             row.Status,
		Reason:             fromPgTextPtr(row.Reason),
		CreatedAt:          fromPgTimestamptz(row.CreatedAt),
	}, nil
}

// 8. End of Service
func (r *Repository) CreateEndOfService(ctx context.Context, eos domain.EndOfServiceCalculation) (*domain.EndOfServiceCalculation, error) {
	snapBytes, _ := json.Marshal(eos.CalculationBreakdown)
	row, err := r.queries(ctx).CreateEndOfServiceCalculation(ctx, sqlc.CreateEndOfServiceCalculationParams{
		ID:                    toPgUUID(eos.ID),
		LegalEntityID:         toPgUUID(eos.LegalEntityID),
		EmployeeID:            toPgUUID(eos.EmployeeID),
		CalculationNumber:     eos.CalculationNumber,
		HireDate:              toPgDate(eos.HireDate),
		TerminationDate:       toPgDate(eos.TerminationDate),
		TerminationReason:     string(eos.TerminationReason),
		ServiceYears:          toPgNumeric(eos.ServiceYears),
		LastGrossSalary:       toPgNumeric(eos.LastGrossSalary),
		AwardAmount:           toPgNumeric(eos.AwardAmount),
		StatutoryRateFraction: toPgNumeric(eos.StatutoryRateFraction),
		FinalSettlementAmount: toPgNumeric(eos.FinalSettlementAmount),
		CalculationBreakdown:  snapBytes,
		Status:                eos.Status,
	})
	if err != nil {
		return nil, err
	}
	return &domain.EndOfServiceCalculation{
		ID:                    toUUID(row.ID),
		LegalEntityID:         toUUID(row.LegalEntityID),
		EmployeeID:            toUUID(row.EmployeeID),
		CalculationNumber:     row.CalculationNumber,
		HireDate:              fromPgDate(row.HireDate),
		TerminationDate:       fromPgDate(row.TerminationDate),
		TerminationReason:     domain.TerminationReason(row.TerminationReason),
		ServiceYears:          fromPgNumeric(row.ServiceYears),
		LastGrossSalary:       fromPgNumeric(row.LastGrossSalary),
		AwardAmount:           fromPgNumeric(row.AwardAmount),
		StatutoryRateFraction: fromPgNumeric(row.StatutoryRateFraction),
		FinalSettlementAmount: fromPgNumeric(row.FinalSettlementAmount),
		CalculationBreakdown:  eos.CalculationBreakdown,
		Status:                row.Status,
		CreatedAt:             fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetEndOfServiceByID(ctx context.Context, id uuid.UUID) (*domain.EndOfServiceCalculation, error) {
	row, err := r.queries(ctx).GetEndOfServiceCalculationByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	var snap map[string]any
	_ = json.Unmarshal(row.CalculationBreakdown, &snap)
	return &domain.EndOfServiceCalculation{
		ID:                    toUUID(row.ID),
		LegalEntityID:         toUUID(row.LegalEntityID),
		EmployeeID:            toUUID(row.EmployeeID),
		CalculationNumber:     row.CalculationNumber,
		HireDate:              fromPgDate(row.HireDate),
		TerminationDate:       fromPgDate(row.TerminationDate),
		TerminationReason:     domain.TerminationReason(row.TerminationReason),
		ServiceYears:          fromPgNumeric(row.ServiceYears),
		LastGrossSalary:       fromPgNumeric(row.LastGrossSalary),
		AwardAmount:           fromPgNumeric(row.AwardAmount),
		StatutoryRateFraction: fromPgNumeric(row.StatutoryRateFraction),
		FinalSettlementAmount: fromPgNumeric(row.FinalSettlementAmount),
		CalculationBreakdown:  snap,
		Status:                row.Status,
		CreatedAt:             fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListEndOfService(ctx context.Context, legalEntityID uuid.UUID) ([]domain.EndOfServiceCalculation, error) {
	rows, err := r.queries(ctx).ListEndOfServiceCalculations(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	var res []domain.EndOfServiceCalculation
	for _, row := range rows {
		var snap map[string]any
		_ = json.Unmarshal(row.CalculationBreakdown, &snap)
		res = append(res, domain.EndOfServiceCalculation{
			ID:                    toUUID(row.ID),
			LegalEntityID:         toUUID(row.LegalEntityID),
			EmployeeID:            toUUID(row.EmployeeID),
			CalculationNumber:     row.CalculationNumber,
			HireDate:              fromPgDate(row.HireDate),
			TerminationDate:       fromPgDate(row.TerminationDate),
			TerminationReason:     domain.TerminationReason(row.TerminationReason),
			ServiceYears:          fromPgNumeric(row.ServiceYears),
			LastGrossSalary:       fromPgNumeric(row.LastGrossSalary),
			AwardAmount:           fromPgNumeric(row.AwardAmount),
			StatutoryRateFraction: fromPgNumeric(row.StatutoryRateFraction),
			FinalSettlementAmount: fromPgNumeric(row.FinalSettlementAmount),
			CalculationBreakdown:  snap,
			Status:                row.Status,
			CreatedAt:             fromPgTimestamptz(row.CreatedAt),
		})
	}
	return res, nil
}

// 9. Seed
func (r *Repository) SeedDefaultPayrollData(ctx context.Context, legalEntityID uuid.UUID) error {
	_, err := r.db.Pool.Exec(ctx, "CALL payroll.seed_default_payroll_data($1)", legalEntityID)
	return err
}
