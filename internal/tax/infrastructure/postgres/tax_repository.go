package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"erp/internal/platform/database/sqlc"
	"erp/internal/tax/domain"
	"erp/internal/tax/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taxRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewTaxRepository(pool *pgxpool.Pool) ports.TaxRepository {
	return &taxRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// 1. Tax Categories & Reasons
func (r *taxRepository) ListTaxCategories(ctx context.Context) ([]domain.TaxCategory, error) {
	rows, err := r.queries.ListTaxCategories(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxCategory, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TaxCategory{
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			DescriptionEn: fromPgText(row.DescriptionEn),
			DescriptionAr: fromPgText(row.DescriptionAr),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return res, nil
}

func (r *taxRepository) GetTaxCategoryByCode(ctx context.Context, code string) (*domain.TaxCategory, error) {
	row, err := r.queries.GetTaxCategoryByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.TaxCategory{
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		DescriptionEn: fromPgText(row.DescriptionEn),
		DescriptionAr: fromPgText(row.DescriptionAr),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (r *taxRepository) ListExemptionReasons(ctx context.Context) ([]domain.ExemptionReason, error) {
	rows, err := r.queries.ListExemptionReasons(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.ExemptionReason, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ExemptionReason{
			Code:          row.Code,
			CategoryCode:  row.CategoryCode,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			DescriptionEn: fromPgText(row.DescriptionEn),
			DescriptionAr: fromPgText(row.DescriptionAr),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return res, nil
}

func (r *taxRepository) ListZeroRateReasons(ctx context.Context) ([]domain.ZeroRateReason, error) {
	rows, err := r.queries.ListZeroRateReasons(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.ZeroRateReason, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ZeroRateReason{
			Code:          row.Code,
			CategoryCode:  row.CategoryCode,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			DescriptionEn: fromPgText(row.DescriptionEn),
			DescriptionAr: fromPgText(row.DescriptionAr),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return res, nil
}

// 2. Tax Codes & Rates
func (r *taxRepository) ListTaxCodesByEntity(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxCode, error) {
	rows, err := r.queries.ListTaxCodesByEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxCode, 0, len(rows))
	for _, row := range rows {
		var ratePtr *domain.TaxRateVersion
		_ = ratePtr
		rateDec := fromPgNumeric(row.CurrentRatePercent)
		res = append(res, domain.TaxCode{
			ID:                   toUUID(row.ID),
			LegalEntityID:        toUUID(row.LegalEntityID),
			Code:                 row.Code,
			NameEn:               row.NameEn,
			NameAr:               row.NameAr,
			TaxCategoryCode:      row.TaxCategoryCode,
			TaxType:              row.TaxType,
			SalesTaxAccountID:    toUUIDPtr(row.SalesTaxAccountID),
			PurchaseTaxAccountID: toUUIDPtr(row.PurchaseTaxAccountID),
			IsRecoverable:        row.IsRecoverable,
			IsActive:             row.IsActive,
			CreatedAt:            row.CreatedAt.Time,
			UpdatedAt:            row.UpdatedAt.Time,
			Version:              row.Version,
			CategoryNameEn:       row.CategoryNameEn,
			CategoryNameAr:       row.CategoryNameAr,
			CurrentRatePercent:   &rateDec,
		})
	}
	return res, nil
}

func (r *taxRepository) GetTaxCodeByID(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error) {
	row, err := r.queries.GetTaxCodeByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaxCodeNotFound
		}
		return nil, err
	}
	return &domain.TaxCode{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		Code:                 row.Code,
		NameEn:               row.NameEn,
		NameAr:               row.NameAr,
		TaxCategoryCode:      row.TaxCategoryCode,
		TaxType:              row.TaxType,
		SalesTaxAccountID:    toUUIDPtr(row.SalesTaxAccountID),
		PurchaseTaxAccountID: toUUIDPtr(row.PurchaseTaxAccountID),
		IsRecoverable:        row.IsRecoverable,
		IsActive:             row.IsActive,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		Version:              row.Version,
	}, nil
}

func (r *taxRepository) GetTaxCodeByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.TaxCode, error) {
	row, err := r.queries.GetTaxCodeByEntityAndCode(ctx, sqlc.GetTaxCodeByEntityAndCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaxCodeNotFound
		}
		return nil, err
	}
	return &domain.TaxCode{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		Code:                 row.Code,
		NameEn:               row.NameEn,
		NameAr:               row.NameAr,
		TaxCategoryCode:      row.TaxCategoryCode,
		TaxType:              row.TaxType,
		SalesTaxAccountID:    toUUIDPtr(row.SalesTaxAccountID),
		PurchaseTaxAccountID: toUUIDPtr(row.PurchaseTaxAccountID),
		IsRecoverable:        row.IsRecoverable,
		IsActive:             row.IsActive,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		Version:              row.Version,
	}, nil
}

func (r *taxRepository) CreateTaxCode(ctx context.Context, code *domain.TaxCode) (*domain.TaxCode, error) {
	row, err := r.queries.CreateTaxCode(ctx, sqlc.CreateTaxCodeParams{
		LegalEntityID:        toPgUUID(code.LegalEntityID),
		Code:                 code.Code,
		NameEn:               code.NameEn,
		NameAr:               code.NameAr,
		TaxCategoryCode:      code.TaxCategoryCode,
		TaxType:              code.TaxType,
		SalesTaxAccountID:    toPgUUIDPtr(code.SalesTaxAccountID),
		PurchaseTaxAccountID: toPgUUIDPtr(code.PurchaseTaxAccountID),
		IsRecoverable:        code.IsRecoverable,
		IsActive:             code.IsActive,
	})
	if err != nil {
		return nil, err
	}
	code.ID = toUUID(row.ID)
	code.CreatedAt = row.CreatedAt.Time
	code.UpdatedAt = row.UpdatedAt.Time
	code.Version = row.Version
	return code, nil
}

func (r *taxRepository) UpdateTaxCode(ctx context.Context, code *domain.TaxCode) (*domain.TaxCode, error) {
	row, err := r.queries.UpdateTaxCode(ctx, sqlc.UpdateTaxCodeParams{
		ID:                   toPgUUID(code.ID),
		NameEn:               code.NameEn,
		NameAr:               code.NameAr,
		TaxType:              code.TaxType,
		SalesTaxAccountID:    toPgUUIDPtr(code.SalesTaxAccountID),
		PurchaseTaxAccountID: toPgUUIDPtr(code.PurchaseTaxAccountID),
		IsRecoverable:        code.IsRecoverable,
		IsActive:             code.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaxCodeNotFound
		}
		return nil, err
	}
	code.UpdatedAt = row.UpdatedAt.Time
	code.Version = row.Version
	return code, nil
}

func (r *taxRepository) CreateTaxRateVersion(ctx context.Context, rate *domain.TaxRateVersion) (*domain.TaxRateVersion, error) {
	row, err := r.queries.CreateTaxRateVersion(ctx, sqlc.CreateTaxRateVersionParams{
		TaxCodeID:     toPgUUID(rate.TaxCodeID),
		RatePercent:   toPgNumeric(rate.RatePercent),
		EffectiveFrom: toPgDate(rate.EffectiveFrom),
		EffectiveTo:   toPgDatePtr(rate.EffectiveTo),
		Description:   toPgText(rate.Description),
		CreatedBy:     toPgUUIDPtr(rate.CreatedBy),
	})
	if err != nil {
		return nil, err
	}
	rate.ID = toUUID(row.ID)
	rate.CreatedAt = row.CreatedAt.Time
	return rate, nil
}

func (r *taxRepository) GetActiveRateVersion(ctx context.Context, taxCodeID uuid.UUID, targetDate time.Time) (*domain.TaxRateVersion, error) {
	row, err := r.queries.GetActiveTaxRateVersion(ctx, sqlc.GetActiveTaxRateVersionParams{
		TaxCodeID:     toPgUUID(taxCodeID),
		EffectiveFrom: toPgDate(targetDate),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaxRateVersionNotFound
		}
		return nil, err
	}
	return &domain.TaxRateVersion{
		ID:            toUUID(row.ID),
		TaxCodeID:     toUUID(row.TaxCodeID),
		RatePercent:   fromPgNumeric(row.RatePercent),
		EffectiveFrom: fromPgDate(row.EffectiveFrom),
		EffectiveTo:   fromPgDatePtr(row.EffectiveTo),
		Description:   fromPgText(row.Description),
		CreatedBy:     toUUIDPtr(row.CreatedBy),
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (r *taxRepository) ListRateVersions(ctx context.Context, taxCodeID uuid.UUID) ([]domain.TaxRateVersion, error) {
	rows, err := r.queries.ListTaxRateVersionsByCode(ctx, toPgUUID(taxCodeID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxRateVersion, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TaxRateVersion{
			ID:            toUUID(row.ID),
			TaxCodeID:     toUUID(row.TaxCodeID),
			RatePercent:   fromPgNumeric(row.RatePercent),
			EffectiveFrom: fromPgDate(row.EffectiveFrom),
			EffectiveTo:   fromPgDatePtr(row.EffectiveTo),
			Description:   fromPgText(row.Description),
			CreatedBy:     toUUIDPtr(row.CreatedBy),
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return res, nil
}

// 3. Tax Periods
func (r *taxRepository) ListTaxPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxPeriod, error) {
	rows, err := r.queries.ListTaxPeriodsByEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxPeriod, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TaxPeriod{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			PeriodNumber:  int(row.PeriodNumber),
			FiscalYear:    int(row.FiscalYear),
			PeriodType:    row.PeriodType,
			StartDate:     fromPgDate(row.StartDate),
			EndDate:       fromPgDate(row.EndDate),
			DueDate:       fromPgDate(row.DueDate),
			Status:        domain.TaxPeriodStatus(row.Status),
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		})
	}
	return res, nil
}

func (r *taxRepository) GetTaxPeriodByID(ctx context.Context, id uuid.UUID) (*domain.TaxPeriod, error) {
	row, err := r.queries.GetTaxPeriodByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaxPeriodNotFound
		}
		return nil, err
	}
	return &domain.TaxPeriod{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		PeriodNumber:  int(row.PeriodNumber),
		FiscalYear:    int(row.FiscalYear),
		PeriodType:    row.PeriodType,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		DueDate:       fromPgDate(row.DueDate),
		Status:        domain.TaxPeriodStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *taxRepository) GetOpenPeriodForDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.TaxPeriod, error) {
	row, err := r.queries.GetOpenTaxPeriodForDate(ctx, sqlc.GetOpenTaxPeriodForDateParams{
		LegalEntityID: toPgUUID(legalEntityID),
		StartDate:     toPgDate(date),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaxPeriodNotFound
		}
		return nil, err
	}
	return &domain.TaxPeriod{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		PeriodNumber:  int(row.PeriodNumber),
		FiscalYear:    int(row.FiscalYear),
		PeriodType:    row.PeriodType,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		DueDate:       fromPgDate(row.DueDate),
		Status:        domain.TaxPeriodStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *taxRepository) CreateTaxPeriod(ctx context.Context, period *domain.TaxPeriod) (*domain.TaxPeriod, error) {
	row, err := r.queries.CreateTaxPeriod(ctx, sqlc.CreateTaxPeriodParams{
		LegalEntityID: toPgUUID(period.LegalEntityID),
		PeriodNumber:  int32(period.PeriodNumber),
		FiscalYear:    int32(period.FiscalYear),
		PeriodType:    period.PeriodType,
		StartDate:     toPgDate(period.StartDate),
		EndDate:       toPgDate(period.EndDate),
		DueDate:       toPgDate(period.DueDate),
		Status:        string(period.Status),
	})
	if err != nil {
		return nil, err
	}
	period.ID = toUUID(row.ID)
	period.CreatedAt = row.CreatedAt.Time
	period.UpdatedAt = row.UpdatedAt.Time
	return period, nil
}

func (r *taxRepository) UpdateTaxPeriodStatus(ctx context.Context, id uuid.UUID, status domain.TaxPeriodStatus) (*domain.TaxPeriod, error) {
	row, err := r.queries.UpdateTaxPeriodStatus(ctx, sqlc.UpdateTaxPeriodStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return &domain.TaxPeriod{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		PeriodNumber:  int(row.PeriodNumber),
		FiscalYear:    int(row.FiscalYear),
		PeriodType:    row.PeriodType,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		DueDate:       fromPgDate(row.DueDate),
		Status:        domain.TaxPeriodStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

// 4. Tax Transactions
func (r *taxRepository) CreateTaxTransaction(ctx context.Context, txRecord *domain.TaxTransaction) (*domain.TaxTransaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// 1. Insert Header
	headerRow, err := qtx.CreateTaxTransaction(ctx, sqlc.CreateTaxTransactionParams{
		LegalEntityID:      toPgUUID(txRecord.LegalEntityID),
		TaxPeriodID:        toPgUUIDPtr(txRecord.TaxPeriodID),
		SourceModule:       txRecord.SourceModule,
		SourceType:         txRecord.SourceType,
		SourceID:           toPgUUID(txRecord.SourceID),
		TransactionNumber:  txRecord.TransactionNumber,
		TransactionDate:    toPgDate(txRecord.TransactionDate),
		TransactionType:    txRecord.TransactionType,
		PartnerID:          toPgUUIDPtr(txRecord.PartnerID),
		PartnerName:        toPgTextPtr(txRecord.PartnerName),
		PartnerTaxNumber:   toPgTextPtr(txRecord.PartnerTaxNumber),
		TotalTaxableAmount: toPgNumeric(txRecord.TotalTaxableAmount),
		TotalTaxAmount:     toPgNumeric(txRecord.TotalTaxAmount),
		TotalGrossAmount:   toPgNumeric(txRecord.TotalGrossAmount),
		JournalEntryID:     toPgUUIDPtr(txRecord.JournalEntryID),
		Status:             txRecord.Status,
		CreatedBy:          toPgUUIDPtr(txRecord.CreatedBy),
	})
	if err != nil {
		return nil, err
	}

	txID := toUUID(headerRow.ID)
	txRecord.ID = txID
	txRecord.CreatedAt = headerRow.CreatedAt.Time

	// 2. Insert Lines
	createdLines := make([]domain.TaxTransactionLine, 0, len(txRecord.Lines))
	for i, line := range txRecord.Lines {
		lineNum := line.LineNumber
		if lineNum == 0 {
			lineNum = i + 1
		}
		lineRow, err := qtx.CreateTaxTransactionLine(ctx, sqlc.CreateTaxTransactionLineParams{
			TaxTransactionID:    toPgUUID(txID),
			LineNumber:          int32(lineNum),
			TaxCodeID:           toPgUUID(line.TaxCodeID),
			TaxCategoryCode:     line.TaxCategoryCode,
			TaxRatePercent:      toPgNumeric(line.TaxRatePercent),
			TaxableAmount:       toPgNumeric(line.TaxableAmount),
			TaxAmount:           toPgNumeric(line.TaxAmount),
			GrossAmount:         toPgNumeric(line.GrossAmount),
			ExemptionReasonCode: toPgTextPtr(line.ExemptionReasonCode),
			ZeroRateReasonCode:  toPgTextPtr(line.ZeroRateReasonCode),
			IsRecoverable:       line.IsRecoverable,
		})
		if err != nil {
			return nil, err
		}
		line.ID = toUUID(lineRow.ID)
		line.TaxTransactionID = txID
		line.LineNumber = lineNum
		line.CreatedAt = lineRow.CreatedAt.Time
		createdLines = append(createdLines, line)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	txRecord.Lines = createdLines
	return txRecord, nil
}

func (r *taxRepository) GetTaxTransactionByID(ctx context.Context, id uuid.UUID) (*domain.TaxTransaction, error) {
	row, err := r.queries.GetTaxTransactionByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	linesRows, err := r.queries.ListTaxTransactionLines(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}

	lines := make([]domain.TaxTransactionLine, 0, len(linesRows))
	for _, l := range linesRows {
		lines = append(lines, domain.TaxTransactionLine{
			ID:                  toUUID(l.ID),
			TaxTransactionID:    toUUID(l.TaxTransactionID),
			LineNumber:          int(l.LineNumber),
			TaxCodeID:           toUUID(l.TaxCodeID),
			TaxCode:             l.TaxCode,
			TaxNameAr:           l.TaxNameAr,
			TaxNameEn:           l.TaxNameEn,
			TaxCategoryCode:     l.TaxCategoryCode,
			TaxRatePercent:      fromPgNumeric(l.TaxRatePercent),
			TaxableAmount:       fromPgNumeric(l.TaxableAmount),
			TaxAmount:           fromPgNumeric(l.TaxAmount),
			GrossAmount:         fromPgNumeric(l.GrossAmount),
			ExemptionReasonCode: fromPgTextPtr(l.ExemptionReasonCode),
			ZeroRateReasonCode:  fromPgTextPtr(l.ZeroRateReasonCode),
			IsRecoverable:       l.IsRecoverable,
			CreatedAt:           l.CreatedAt.Time,
		})
	}

	return &domain.TaxTransaction{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		TaxPeriodID:        toUUIDPtr(row.TaxPeriodID),
		SourceModule:       row.SourceModule,
		SourceType:         row.SourceType,
		SourceID:           toUUID(row.SourceID),
		TransactionNumber:  row.TransactionNumber,
		TransactionDate:    fromPgDate(row.TransactionDate),
		TransactionType:    row.TransactionType,
		PartnerID:          toUUIDPtr(row.PartnerID),
		PartnerName:        fromPgTextPtr(row.PartnerName),
		PartnerTaxNumber:   fromPgTextPtr(row.PartnerTaxNumber),
		TotalTaxableAmount: fromPgNumeric(row.TotalTaxableAmount),
		TotalTaxAmount:     fromPgNumeric(row.TotalTaxAmount),
		TotalGrossAmount:   fromPgNumeric(row.TotalGrossAmount),
		JournalEntryID:     toUUIDPtr(row.JournalEntryID),
		Status:             row.Status,
		CreatedBy:          toUUIDPtr(row.CreatedBy),
		CreatedAt:          row.CreatedAt.Time,
		Lines:              lines,
	}, nil
}

func (r *taxRepository) GetTaxTransactionBySource(ctx context.Context, legalEntityID uuid.UUID, sourceModule, sourceType string, sourceID uuid.UUID) (*domain.TaxTransaction, error) {
	row, err := r.queries.GetTaxTransactionBySource(ctx, sqlc.GetTaxTransactionBySourceParams{
		LegalEntityID: toPgUUID(legalEntityID),
		SourceModule:  sourceModule,
		SourceType:    sourceType,
		SourceID:      toPgUUID(sourceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return r.GetTaxTransactionByID(ctx, toUUID(row.ID))
}

func (r *taxRepository) ListTaxTransactionsByEntity(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.TaxTransaction, error) {
	rows, err := r.queries.ListTaxTransactionsByEntity(ctx, sqlc.ListTaxTransactionsByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxTransaction, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TaxTransaction{
			ID:                 toUUID(row.ID),
			LegalEntityID:      toUUID(row.LegalEntityID),
			TaxPeriodID:        toUUIDPtr(row.TaxPeriodID),
			SourceModule:       row.SourceModule,
			SourceType:         row.SourceType,
			SourceID:           toUUID(row.SourceID),
			TransactionNumber:  row.TransactionNumber,
			TransactionDate:    fromPgDate(row.TransactionDate),
			TransactionType:    row.TransactionType,
			PartnerID:          toUUIDPtr(row.PartnerID),
			PartnerName:        fromPgTextPtr(row.PartnerName),
			PartnerTaxNumber:   fromPgTextPtr(row.PartnerTaxNumber),
			TotalTaxableAmount: fromPgNumeric(row.TotalTaxableAmount),
			TotalTaxAmount:     fromPgNumeric(row.TotalTaxAmount),
			TotalGrossAmount:   fromPgNumeric(row.TotalGrossAmount),
			JournalEntryID:     toUUIDPtr(row.JournalEntryID),
			Status:             row.Status,
			CreatedBy:          toUUIDPtr(row.CreatedBy),
			CreatedAt:          row.CreatedAt.Time,
		})
	}
	return res, nil
}

func (r *taxRepository) ListTaxTransactionsByPeriod(ctx context.Context, periodID uuid.UUID) ([]domain.TaxTransaction, error) {
	rows, err := r.queries.ListTaxTransactionsByPeriod(ctx, toPgUUIDPtr(&periodID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxTransaction, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TaxTransaction{
			ID:                 toUUID(row.ID),
			LegalEntityID:      toUUID(row.LegalEntityID),
			TaxPeriodID:        toUUIDPtr(row.TaxPeriodID),
			SourceModule:       row.SourceModule,
			SourceType:         row.SourceType,
			SourceID:           toUUID(row.SourceID),
			TransactionNumber:  row.TransactionNumber,
			TransactionDate:    fromPgDate(row.TransactionDate),
			TransactionType:    row.TransactionType,
			PartnerID:          toUUIDPtr(row.PartnerID),
			PartnerName:        fromPgTextPtr(row.PartnerName),
			PartnerTaxNumber:   fromPgTextPtr(row.PartnerTaxNumber),
			TotalTaxableAmount: fromPgNumeric(row.TotalTaxableAmount),
			TotalTaxAmount:     fromPgNumeric(row.TotalTaxAmount),
			TotalGrossAmount:   fromPgNumeric(row.TotalGrossAmount),
			JournalEntryID:     toUUIDPtr(row.JournalEntryID),
			Status:             row.Status,
			CreatedBy:          toUUIDPtr(row.CreatedBy),
			CreatedAt:          row.CreatedAt.Time,
		})
	}
	return res, nil
}

// 5. Tax Returns & Aggregation
func (r *taxRepository) CreateTaxReturn(ctx context.Context, ret *domain.TaxReturn) (*domain.TaxReturn, error) {
	boxesBytes, err := json.Marshal(ret.DeclarationBoxes)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.CreateTaxReturn(ctx, sqlc.CreateTaxReturnParams{
		LegalEntityID:         toPgUUID(ret.LegalEntityID),
		TaxPeriodID:           toPgUUID(ret.TaxPeriodID),
		ReturnNumber:          ret.ReturnNumber,
		FilingDate:            toPgDatePtr(ret.FilingDate),
		Status:                ret.Status,
		TotalSalesTaxable:     toPgNumeric(ret.TotalSalesTaxable),
		TotalSalesVat:         toPgNumeric(ret.TotalSalesVAT),
		TotalPurchasesTaxable: toPgNumeric(ret.TotalPurchasesTaxable),
		TotalPurchasesVat:     toPgNumeric(ret.TotalPurchasesVAT),
		NetVatDue:             toPgNumeric(ret.NetVATDue),
		DeclarationBoxes:      boxesBytes,
		SubmittedBy:           toPgUUIDPtr(ret.SubmittedBy),
	})
	if err != nil {
		return nil, err
	}
	ret.ID = toUUID(row.ID)
	ret.CreatedAt = row.CreatedAt.Time
	ret.UpdatedAt = row.UpdatedAt.Time
	ret.Version = row.Version
	return ret, nil
}

func (r *taxRepository) GetTaxReturnByID(ctx context.Context, id uuid.UUID) (*domain.TaxReturn, error) {
	row, err := r.queries.GetTaxReturnByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapSqlcTaxReturnToDomain(&row)
}

func (r *taxRepository) GetTaxReturnByPeriod(ctx context.Context, legalEntityID, periodID uuid.UUID) (*domain.TaxReturn, error) {
	row, err := r.queries.GetTaxReturnByPeriod(ctx, sqlc.GetTaxReturnByPeriodParams{
		LegalEntityID: toPgUUID(legalEntityID),
		TaxPeriodID:   toPgUUID(periodID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapSqlcTaxReturnToDomain(&row)
}

func (r *taxRepository) UpdateTaxReturn(ctx context.Context, ret *domain.TaxReturn) (*domain.TaxReturn, error) {
	boxesBytes, err := json.Marshal(ret.DeclarationBoxes)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.UpdateTaxReturn(ctx, sqlc.UpdateTaxReturnParams{
		ID:                    toPgUUID(ret.ID),
		Status:                ret.Status,
		FilingDate:            toPgDatePtr(ret.FilingDate),
		TotalSalesTaxable:     toPgNumeric(ret.TotalSalesTaxable),
		TotalSalesVat:         toPgNumeric(ret.TotalSalesVAT),
		TotalPurchasesTaxable: toPgNumeric(ret.TotalPurchasesTaxable),
		TotalPurchasesVat:     toPgNumeric(ret.TotalPurchasesVAT),
		NetVatDue:             toPgNumeric(ret.NetVATDue),
		DeclarationBoxes:      boxesBytes,
		SubmittedAt:           toPgTimestamptz(ret.SubmittedAt),
		SubmittedBy:           toPgUUIDPtr(ret.SubmittedBy),
	})
	if err != nil {
		return nil, err
	}
	ret.UpdatedAt = row.UpdatedAt.Time
	ret.Version = row.Version
	return ret, nil
}

func (r *taxRepository) ListTaxReturnsByEntity(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxReturn, error) {
	rows, err := r.queries.ListTaxReturnsByEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.TaxReturn, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TaxReturn{
			ID:                    toUUID(row.ID),
			LegalEntityID:         toUUID(row.LegalEntityID),
			TaxPeriodID:           toUUID(row.TaxPeriodID),
			ReturnNumber:          row.ReturnNumber,
			FilingDate:            fromPgDatePtr(row.FilingDate),
			Status:                row.Status,
			TotalSalesTaxable:     fromPgNumeric(row.TotalSalesTaxable),
			TotalSalesVAT:         fromPgNumeric(row.TotalSalesVat),
			TotalPurchasesTaxable: fromPgNumeric(row.TotalPurchasesTaxable),
			TotalPurchasesVAT:     fromPgNumeric(row.TotalPurchasesVat),
			NetVATDue:             fromPgNumeric(row.NetVatDue),
			SubmittedAt:           fromPgTimestamptz(row.SubmittedAt),
			CreatedAt:             row.CreatedAt.Time,
		})
	}
	return res, nil
}

func (r *taxRepository) AggregatePeriodSummary(ctx context.Context, periodID uuid.UUID) ([]ports.TaxPeriodAggregationRow, error) {
	rows, err := r.queries.AggregateTaxPeriodSummary(ctx, toPgUUIDPtr(&periodID))
	if err != nil {
		return nil, err
	}
	res := make([]ports.TaxPeriodAggregationRow, 0, len(rows))
	for _, row := range rows {
		res = append(res, ports.TaxPeriodAggregationRow{
			TransactionType:     row.TransactionType,
			TaxCategoryCode:     row.TaxCategoryCode,
			TaxRatePercent:      fromPgNumeric(row.TaxRatePercent),
			ExemptionReasonCode: fromPgTextPtr(row.ExemptionReasonCode),
			ZeroRateReasonCode:  fromPgTextPtr(row.ZeroRateReasonCode),
			TotalTaxable:        fromPgNumeric(row.TotalTaxable),
			TotalTax:            fromPgNumeric(row.TotalTax),
			TotalGross:          fromPgNumeric(row.TotalGross),
		})
	}
	return res, nil
}

func mapSqlcTaxReturnToDomain(row *sqlc.TaxTaxReturns) (*domain.TaxReturn, error) {
	var boxes domain.DeclarationBoxes
	if len(row.DeclarationBoxes) > 0 {
		_ = json.Unmarshal(row.DeclarationBoxes, &boxes)
	}

	return &domain.TaxReturn{
		ID:                    toUUID(row.ID),
		LegalEntityID:         toUUID(row.LegalEntityID),
		TaxPeriodID:           toUUID(row.TaxPeriodID),
		ReturnNumber:          row.ReturnNumber,
		FilingDate:            fromPgDatePtr(row.FilingDate),
		Status:                row.Status,
		TotalSalesTaxable:     fromPgNumeric(row.TotalSalesTaxable),
		TotalSalesVAT:         fromPgNumeric(row.TotalSalesVat),
		TotalPurchasesTaxable: fromPgNumeric(row.TotalPurchasesTaxable),
		TotalPurchasesVAT:     fromPgNumeric(row.TotalPurchasesVat),
		NetVATDue:             fromPgNumeric(row.NetVatDue),
		DeclarationBoxes:      boxes,
		SubmittedAt:           fromPgTimestamptz(row.SubmittedAt),
		SubmittedBy:           toUUIDPtr(row.SubmittedBy),
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
		Version:               row.Version,
	}, nil
}
