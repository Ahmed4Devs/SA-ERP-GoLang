package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/accounting/domain"
	"erp/internal/accounting/ports"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type accountingRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewAccountingRepository(pool *pgxpool.Pool) ports.AccountingRepository {
	return &accountingRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// NewPostingRepository returns the same repository implementation typed as
// PostingRepository, which adds GetJournalEntryBySource and UpsertAccountBalance
// on top of the base AccountingRepository interface.
func NewPostingRepository(pool *pgxpool.Pool) ports.PostingRepository {
	return &accountingRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}


// Chart of Accounts
func (r *accountingRepository) ListChartOfAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ChartOfAccounts, error) {
	rows, err := r.queries.ListChartOfAccountsByLegalEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.ChartOfAccounts, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ChartOfAccounts{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			Description:   fromPgText(row.Description),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			Version:       row.Version,
		})
	}
	return result, nil
}

func (r *accountingRepository) GetChartOfAccountsByID(ctx context.Context, id uuid.UUID) (*domain.ChartOfAccounts, error) {
	row, err := r.queries.GetChartOfAccountsByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return &domain.ChartOfAccounts{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		Description:   fromPgText(row.Description),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

// Accounts
func (r *accountingRepository) CreateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error) {
	row, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		LegalEntityID:    toPgUUID(acc.LegalEntityID),
		ChartID:          toPgUUID(acc.ChartID),
		ParentID:         toPgUUIDPtr(acc.ParentID),
		Code:             acc.Code,
		NameEn:           acc.NameEn,
		NameAr:           acc.NameAr,
		AccountType:      string(acc.AccountType),
		NormalBalance:    string(acc.NormalBalance),
		AllowPosting:     acc.AllowPosting,
		IsControlAccount: acc.IsControlAccount,
		IsActive:         acc.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return toDomainAccount(&row), nil
}

func (r *accountingRepository) UpdateAccount(ctx context.Context, acc domain.Account) (*domain.Account, error) {
	row, err := r.queries.UpdateAccount(ctx, sqlc.UpdateAccountParams{
		ID:               toPgUUID(acc.ID),
		NameEn:           acc.NameEn,
		NameAr:           acc.NameAr,
		AllowPosting:     acc.AllowPosting,
		IsControlAccount: acc.IsControlAccount,
		IsActive:         acc.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(&row), nil
}

func (r *accountingRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	row, err := r.queries.GetAccountByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(&row), nil
}

func (r *accountingRepository) GetAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Account, error) {
	row, err := r.queries.GetAccountByCode(ctx, sqlc.GetAccountByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return toDomainAccount(&row), nil
}

func (r *accountingRepository) ListAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Account, error) {
	rows, err := r.queries.ListAccountsByLegalEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.Account, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainAccount(&row))
	}
	return result, nil
}

func toDomainAccount(row *sqlc.AccountingAccounts) *domain.Account {
	return &domain.Account{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		ChartID:          toUUID(row.ChartID),
		ParentID:         toUUIDPtr(row.ParentID),
		Code:             row.Code,
		NameEn:           row.NameEn,
		NameAr:           row.NameAr,
		AccountType:      domain.AccountType(row.AccountType),
		NormalBalance:    domain.NormalBalance(row.NormalBalance),
		AllowPosting:     row.AllowPosting,
		IsControlAccount: row.IsControlAccount,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		Version:          row.Version,
	}
}

// Fiscal Years & Periods
func (r *accountingRepository) ListFiscalYears(ctx context.Context, legalEntityID uuid.UUID) ([]domain.FiscalYear, error) {
	rows, err := r.queries.ListFiscalYearsByLegalEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.FiscalYear, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.FiscalYear{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			StartDate:     fromPgDate(row.StartDate),
			EndDate:       fromPgDate(row.EndDate),
			Status:        row.Status,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			Version:       row.Version,
		})
	}
	return result, nil
}

func (r *accountingRepository) GetFiscalYearByID(ctx context.Context, id uuid.UUID) (*domain.FiscalYear, error) {
	row, err := r.queries.GetFiscalYearByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFiscalPeriodNotFound
		}
		return nil, err
	}
	return &domain.FiscalYear{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		Status:        row.Status,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

func (r *accountingRepository) GetFiscalYearByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.FiscalYear, error) {
	row, err := r.queries.GetFiscalYearByDate(ctx, sqlc.GetFiscalYearByDateParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       toPgDate(date),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFiscalPeriodNotFound
		}
		return nil, err
	}
	return &domain.FiscalYear{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		Status:        row.Status,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

func (r *accountingRepository) ListFiscalPeriods(ctx context.Context, fiscalYearID uuid.UUID) ([]domain.FiscalPeriod, error) {
	rows, err := r.queries.ListFiscalPeriodsByYear(ctx, toPgUUID(fiscalYearID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.FiscalPeriod, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.FiscalPeriod{
			ID:            toUUID(row.ID),
			FiscalYearID:  toUUID(row.FiscalYearID),
			LegalEntityID: toUUID(row.LegalEntityID),
			PeriodNumber:  row.PeriodNumber,
			PeriodNameEn:  row.PeriodNameEn,
			PeriodNameAr:  row.PeriodNameAr,
			Quarter:       row.Quarter,
			StartDate:     fromPgDate(row.StartDate),
			EndDate:       fromPgDate(row.EndDate),
			Status:        domain.FiscalPeriodStatus(row.Status),
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			Version:       row.Version,
		})
	}
	return result, nil
}

func (r *accountingRepository) GetFiscalPeriodByID(ctx context.Context, id uuid.UUID) (*domain.FiscalPeriod, error) {
	row, err := r.queries.GetFiscalPeriodByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFiscalPeriodNotFound
		}
		return nil, err
	}
	return &domain.FiscalPeriod{
		ID:            toUUID(row.ID),
		FiscalYearID:  toUUID(row.FiscalYearID),
		LegalEntityID: toUUID(row.LegalEntityID),
		PeriodNumber:  row.PeriodNumber,
		PeriodNameEn:  row.PeriodNameEn,
		PeriodNameAr:  row.PeriodNameAr,
		Quarter:       row.Quarter,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		Status:        domain.FiscalPeriodStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

func (r *accountingRepository) GetFiscalPeriodByDate(ctx context.Context, legalEntityID uuid.UUID, date time.Time) (*domain.FiscalPeriod, error) {
	row, err := r.queries.GetFiscalPeriodByDate(ctx, sqlc.GetFiscalPeriodByDateParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       toPgDate(date),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFiscalPeriodNotFound
		}
		return nil, err
	}
	return &domain.FiscalPeriod{
		ID:            toUUID(row.ID),
		FiscalYearID:  toUUID(row.FiscalYearID),
		LegalEntityID: toUUID(row.LegalEntityID),
		PeriodNumber:  row.PeriodNumber,
		PeriodNameEn:  row.PeriodNameEn,
		PeriodNameAr:  row.PeriodNameAr,
		Quarter:       row.Quarter,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		Status:        domain.FiscalPeriodStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

func (r *accountingRepository) UpdateFiscalPeriodStatus(ctx context.Context, id uuid.UUID, status domain.FiscalPeriodStatus) (*domain.FiscalPeriod, error) {
	row, err := r.queries.UpdateFiscalPeriodStatus(ctx, sqlc.UpdateFiscalPeriodStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFiscalPeriodNotFound
		}
		return nil, err
	}
	return &domain.FiscalPeriod{
		ID:            toUUID(row.ID),
		FiscalYearID:  toUUID(row.FiscalYearID),
		LegalEntityID: toUUID(row.LegalEntityID),
		PeriodNumber:  row.PeriodNumber,
		PeriodNameEn:  row.PeriodNameEn,
		PeriodNameAr:  row.PeriodNameAr,
		Quarter:       row.Quarter,
		StartDate:     fromPgDate(row.StartDate),
		EndDate:       fromPgDate(row.EndDate),
		Status:        domain.FiscalPeriodStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

// Journals
func (r *accountingRepository) ListJournals(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Journal, error) {
	rows, err := r.queries.ListJournalsByLegalEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.Journal, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Journal{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			JournalType:   domain.JournalType(row.JournalType),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			Version:       row.Version,
		})
	}
	return result, nil
}

func (r *accountingRepository) GetJournalByID(ctx context.Context, id uuid.UUID) (*domain.Journal, error) {
	row, err := r.queries.GetJournalByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJournalNotFound
		}
		return nil, err
	}
	return &domain.Journal{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		JournalType:   domain.JournalType(row.JournalType),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

func (r *accountingRepository) GetJournalByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Journal, error) {
	row, err := r.queries.GetJournalByCode(ctx, sqlc.GetJournalByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJournalNotFound
		}
		return nil, err
	}
	return &domain.Journal{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		JournalType:   domain.JournalType(row.JournalType),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

func (r *accountingRepository) CreateJournal(ctx context.Context, j domain.Journal) (*domain.Journal, error) {
	row, err := r.queries.CreateJournal(ctx, sqlc.CreateJournalParams{
		LegalEntityID: toPgUUID(j.LegalEntityID),
		Code:          j.Code,
		NameEn:        j.NameEn,
		NameAr:        j.NameAr,
		JournalType:   string(j.JournalType),
		IsActive:      j.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Journal{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		JournalType:   domain.JournalType(row.JournalType),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}, nil
}

// Journal Entries & Lines
func (r *accountingRepository) CreateJournalEntry(ctx context.Context, entry domain.JournalEntry, lines []domain.JournalLine) (*domain.JournalEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	rate := entry.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}
	currency := entry.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}

	entryRow, err := qtx.CreateJournalEntry(ctx, sqlc.CreateJournalEntryParams{
		LegalEntityID:  toPgUUID(entry.LegalEntityID),
		JournalID:      toPgUUID(entry.JournalID),
		FiscalPeriodID: toPgUUID(entry.FiscalPeriodID),
		EntryNumber:    entry.EntryNumber,
		AccountingDate: toPgDate(entry.AccountingDate),
		Status:         string(entry.Status),
		SourceModule:   entry.SourceModule,
		SourceType:     toPgTextPtr(entry.SourceType),
		SourceID:       toPgUUIDPtr(entry.SourceID),
		Reference:      toPgTextPtr(entry.Reference),
		Description:    entry.Description,
		CurrencyCode:   currency,
		ExchangeRate:   toPgNumeric(rate),
		TotalDebit:     toPgNumeric(entry.TotalDebit),
		TotalCredit:    toPgNumeric(entry.TotalCredit),
		CreatedBy:      toPgUUID(entry.CreatedBy),
	})
	if err != nil {
		return nil, err
	}

	createdLines := make([]domain.JournalLine, 0, len(lines))
	for _, l := range lines {
		lRate := l.ExchangeRate
		if lRate.IsZero() {
			lRate = rate
		}
		lCurr := l.CurrencyCode
		if lCurr == "" {
			lCurr = currency
		}

		lineRow, err := qtx.CreateJournalLine(ctx, sqlc.CreateJournalLineParams{
			JournalEntryID:    entryRow.ID,
			LegalEntityID:     toPgUUID(entry.LegalEntityID),
			LineNumber:        l.LineNumber,
			AccountID:         toPgUUID(l.AccountID),
			Description:       toPgTextPtr(l.Description),
			CurrencyCode:      lCurr,
			ExchangeRate:      toPgNumeric(lRate),
			TransactionDebit:  toPgNumeric(l.TransactionDebit),
			TransactionCredit: toPgNumeric(l.TransactionCredit),
			BaseDebit:         toPgNumeric(l.BaseDebit),
			BaseCredit:        toPgNumeric(l.BaseCredit),
			BranchID:          toPgUUIDPtr(l.BranchID),
			DepartmentID:      toPgUUIDPtr(l.DepartmentID),
			CostCenterID:      toPgUUIDPtr(l.CostCenterID),
		})
		if err != nil {
			return nil, err
		}

		createdLines = append(createdLines, domain.JournalLine{
			ID:                toUUID(lineRow.ID),
			JournalEntryID:    toUUID(lineRow.JournalEntryID),
			LegalEntityID:     toUUID(lineRow.LegalEntityID),
			LineNumber:        lineRow.LineNumber,
			AccountID:         toUUID(lineRow.AccountID),
			Description:       fromPgTextPtr(lineRow.Description),
			CurrencyCode:      lineRow.CurrencyCode,
			ExchangeRate:      fromPgNumeric(lineRow.ExchangeRate),
			TransactionDebit:  fromPgNumeric(lineRow.TransactionDebit),
			TransactionCredit: fromPgNumeric(lineRow.TransactionCredit),
			BaseDebit:         fromPgNumeric(lineRow.BaseDebit),
			BaseCredit:        fromPgNumeric(lineRow.BaseCredit),
			BranchID:          toUUIDPtr(lineRow.BranchID),
			DepartmentID:      toUUIDPtr(lineRow.DepartmentID),
			CostCenterID:      toUUIDPtr(lineRow.CostCenterID),
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	res := toDomainJournalEntry(&entryRow)
	res.Lines = createdLines
	return res, nil
}

func (r *accountingRepository) GetJournalEntryByID(ctx context.Context, id uuid.UUID) (*domain.JournalEntry, error) {
	row, err := r.queries.GetJournalEntryByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJournalEntryNotFound
		}
		return nil, err
	}

	entry := toDomainJournalEntry(&row)

	lineRows, err := r.queries.ListJournalLinesByEntryID(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}

	lines := make([]domain.JournalLine, 0, len(lineRows))
	for _, l := range lineRows {
		lines = append(lines, domain.JournalLine{
			ID:                toUUID(l.ID),
			JournalEntryID:    toUUID(l.JournalEntryID),
			LegalEntityID:     toUUID(l.LegalEntityID),
			LineNumber:        l.LineNumber,
			AccountID:         toUUID(l.AccountID),
			AccountCode:       l.AccountCode,
			AccountNameEn:     l.AccountNameEn,
			AccountNameAr:     l.AccountNameAr,
			Description:       fromPgTextPtr(l.Description),
			CurrencyCode:      l.CurrencyCode,
			ExchangeRate:      fromPgNumeric(l.ExchangeRate),
			TransactionDebit:  fromPgNumeric(l.TransactionDebit),
			TransactionCredit: fromPgNumeric(l.TransactionCredit),
			BaseDebit:         fromPgNumeric(l.BaseDebit),
			BaseCredit:        fromPgNumeric(l.BaseCredit),
			BranchID:          toUUIDPtr(l.BranchID),
			DepartmentID:      toUUIDPtr(l.DepartmentID),
			CostCenterID:      toUUIDPtr(l.CostCenterID),
		})
	}
	entry.Lines = lines
	return entry, nil
}

func (r *accountingRepository) GetJournalEntryByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.JournalEntry, error) {
	row, err := r.queries.GetJournalEntryByNumber(ctx, sqlc.GetJournalEntryByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		EntryNumber:   number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJournalEntryNotFound
		}
		return nil, err
	}
	return r.GetJournalEntryByID(ctx, toUUID(row.ID))
}

func (r *accountingRepository) ListJournalEntries(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.JournalEntry, int64, error) {
	rows, err := r.queries.ListJournalEntriesByLegalEntity(ctx, sqlc.ListJournalEntriesByLegalEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	totalCount, err := r.queries.CountJournalEntriesByLegalEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, 0, err
	}

	result := make([]domain.JournalEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainJournalEntry(&row))
	}
	return result, totalCount, nil
}

func (r *accountingRepository) PostJournalEntry(ctx context.Context, id uuid.UUID, postedBy uuid.UUID) (*domain.JournalEntry, error) {
	row, err := r.queries.PostJournalEntry(ctx, sqlc.PostJournalEntryParams{
		ID:       toPgUUID(id),
		PostedBy: toPgUUID(postedBy),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJournalEntryNotFound
		}
		return nil, err
	}
	return r.GetJournalEntryByID(ctx, toUUID(row.ID))
}

func (r *accountingRepository) ReverseJournalEntry(
	ctx context.Context,
	originalID uuid.UUID,
	reversalEntry domain.JournalEntry,
	reversalLines []domain.JournalLine,
	reversedBy uuid.UUID,
) (*domain.JournalEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// 1. Create reversal journal entry
	rate := reversalEntry.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}
	curr := reversalEntry.CurrencyCode
	if curr == "" {
		curr = "SAR"
	}

	origIDPg := toPgUUID(originalID)

	revRow, err := qtx.CreateJournalEntry(ctx, sqlc.CreateJournalEntryParams{
		LegalEntityID:  toPgUUID(reversalEntry.LegalEntityID),
		JournalID:      toPgUUID(reversalEntry.JournalID),
		FiscalPeriodID: toPgUUID(reversalEntry.FiscalPeriodID),
		EntryNumber:    reversalEntry.EntryNumber,
		AccountingDate: toPgDate(reversalEntry.AccountingDate),
		Status:         "DRAFT", // Initialized as DRAFT so lines can be inserted
		SourceModule:   reversalEntry.SourceModule,
		SourceType:     toPgTextPtr(reversalEntry.SourceType),
		SourceID:       toPgUUIDPtr(reversalEntry.SourceID),
		Reference:      toPgTextPtr(reversalEntry.Reference),
		Description:    reversalEntry.Description,
		CurrencyCode:   curr,
		ExchangeRate:   toPgNumeric(rate),
		TotalDebit:     toPgNumeric(reversalEntry.TotalDebit),
		TotalCredit:    toPgNumeric(reversalEntry.TotalCredit),
		CreatedBy:      toPgUUID(reversedBy),
	})
	if err != nil {
		return nil, err
	}

	// 2. Insert reversed lines (swap debit and credit)
	for _, l := range reversalLines {
		_, err := qtx.CreateJournalLine(ctx, sqlc.CreateJournalLineParams{
			JournalEntryID:    revRow.ID,
			LegalEntityID:     toPgUUID(reversalEntry.LegalEntityID),
			LineNumber:        l.LineNumber,
			AccountID:         toPgUUID(l.AccountID),
			Description:       toPgTextPtr(l.Description),
			CurrencyCode:      curr,
			ExchangeRate:      toPgNumeric(rate),
			TransactionDebit:  toPgNumeric(l.TransactionDebit),
			TransactionCredit: toPgNumeric(l.TransactionCredit),
			BaseDebit:         toPgNumeric(l.BaseDebit),
			BaseCredit:        toPgNumeric(l.BaseCredit),
			BranchID:          toPgUUIDPtr(l.BranchID),
			DepartmentID:      toPgUUIDPtr(l.DepartmentID),
			CostCenterID:      toPgUUIDPtr(l.CostCenterID),
		})
		if err != nil {
			return nil, err
		}
	}

	// 3. Post the reversal entry
	_, err = qtx.PostJournalEntry(ctx, sqlc.PostJournalEntryParams{
		ID:       revRow.ID,
		PostedBy: toPgUUID(reversedBy),
	})
	if err != nil {
		return nil, err
	}

	// 4. Mark original entry as REVERSED and set reversal_of_id to the reversal entry ID
	_, err = qtx.MarkJournalEntryReversed(ctx, sqlc.MarkJournalEntryReversedParams{
		ID:           origIDPg,
		ReversalOfID: toPgUUID(toUUID(revRow.ID)),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetJournalEntryByID(ctx, toUUID(revRow.ID))
}

func (r *accountingRepository) GetTrialBalance(ctx context.Context, legalEntityID uuid.UUID, asOfDate *time.Time) (*domain.TrialBalanceReport, error) {
	var dateParam pgtype.Date
	var reportDate time.Time
	if asOfDate != nil && !asOfDate.IsZero() {
		dateParam = toPgDate(*asOfDate)
		reportDate = *asOfDate
	} else {
		reportDate = time.Now().UTC()
		dateParam = toPgDate(reportDate)
	}

	rows, err := r.queries.GetTrialBalance(ctx, sqlc.GetTrialBalanceParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       dateParam,
	})
	if err != nil {
		return nil, err
	}

	summary, err := r.queries.GetTrialBalanceSummary(ctx, sqlc.GetTrialBalanceSummaryParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       dateParam,
	})
	if err != nil {
		return nil, err
	}

	resultRows := make([]domain.TrialBalanceRow, 0, len(rows))
	for _, row := range rows {
		resultRows = append(resultRows, domain.TrialBalanceRow{
			AccountID:     toUUID(row.AccountID),
			AccountCode:   row.AccountCode,
			AccountNameEn: row.AccountNameEn,
			AccountNameAr: row.AccountNameAr,
			AccountType:   domain.AccountType(row.AccountType),
			NormalBalance: domain.NormalBalance(row.NormalBalance),
			TotalDebit:    fromPgNumeric(row.TotalDebit),
			TotalCredit:   fromPgNumeric(row.TotalCredit),
			NetBalance:    fromPgNumeric(row.NetBalance),
		})
	}

	isBalanced := summary.IsBalanced

	return &domain.TrialBalanceReport{
		LegalEntityID:  legalEntityID,
		AsOfDate:       reportDate,
		Rows:           resultRows,
		TotalDebitSum:  fromPgNumeric(summary.SumDebit),
		TotalCreditSum: fromPgNumeric(summary.SumCredit),
		IsBalanced:     isBalanced,
	}, nil
}

func toDomainJournalEntry(row *sqlc.AccountingJournalEntries) *domain.JournalEntry {
	return &domain.JournalEntry{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		JournalID:      toUUID(row.JournalID),
		FiscalPeriodID: toUUID(row.FiscalPeriodID),
		EntryNumber:    row.EntryNumber,
		AccountingDate: fromPgDate(row.AccountingDate),
		Status:         domain.JournalEntryStatus(row.Status),
		SourceModule:   row.SourceModule,
		SourceType:     fromPgTextPtr(row.SourceType),
		SourceID:       toUUIDPtr(row.SourceID),
		Reference:      fromPgTextPtr(row.Reference),
		Description:    row.Description,
		CurrencyCode:   row.CurrencyCode,
		ExchangeRate:   fromPgNumeric(row.ExchangeRate),
		TotalDebit:     fromPgNumeric(row.TotalDebit),
		TotalCredit:    fromPgNumeric(row.TotalCredit),
		CreatedBy:      toUUID(row.CreatedBy),
		ApprovedBy:     toUUIDPtr(row.ApprovedBy),
		PostedBy:       toUUIDPtr(row.PostedBy),
		CreatedAt:      row.CreatedAt.Time,
		ApprovedAt:     fromPgTimestamptz(row.ApprovedAt),
		PostedAt:       fromPgTimestamptz(row.PostedAt),
		ReversedAt:     fromPgTimestamptz(row.ReversedAt),
		ReversalOfID:   toUUIDPtr(row.ReversalOfID),
		Version:        row.Version,
	}
}

// GetJournalEntryBySource checks for an existing journal entry by the
// (legal_entity_id, source_module, source_type, source_id) triple.
// Used by the PostingEngine for idempotency enforcement.
func (r *accountingRepository) GetJournalEntryBySource(ctx context.Context, legalEntityID uuid.UUID, sourceModule, sourceType string, sourceID uuid.UUID) (*domain.JournalEntry, error) {
	row, err := r.queries.GetJournalEntryBySource(ctx, sqlc.GetJournalEntryBySourceParams{
		LegalEntityID: toPgUUID(legalEntityID),
		SourceModule:  sourceModule,
		SourceType:    toPgText(sourceType),
		SourceID:      toPgUUID(sourceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No existing entry — not an error, enables idempotency check
		}
		return nil, err
	}
	return toDomainJournalEntry(&row), nil
}

// UpsertAccountBalance atomically inserts or accumulates account period balances.
// Uses PostgreSQL ON CONFLICT DO UPDATE for thread-safe accumulation.
func (r *accountingRepository) UpsertAccountBalance(ctx context.Context, balance domain.AccountBalance) (*domain.AccountBalance, error) {
	row, err := r.queries.UpsertAccountBalance(ctx, sqlc.UpsertAccountBalanceParams{
		LegalEntityID:  toPgUUID(balance.LegalEntityID),
		AccountID:      toPgUUID(balance.AccountID),
		FiscalPeriodID: toPgUUID(balance.FiscalPeriodID),
		OpeningBalance: toPgNumeric(balance.OpeningBalance),
		PeriodDebit:    toPgNumeric(balance.PeriodDebit),
		PeriodCredit:   toPgNumeric(balance.PeriodCredit),
		ClosingBalance: toPgNumeric(balance.ClosingBalance),
	})
	if err != nil {
		return nil, err
	}
	return &domain.AccountBalance{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		AccountID:      toUUID(row.AccountID),
		FiscalPeriodID: toUUID(row.FiscalPeriodID),
		OpeningBalance: fromPgNumeric(row.OpeningBalance),
		PeriodDebit:    fromPgNumeric(row.PeriodDebit),
		PeriodCredit:   fromPgNumeric(row.PeriodCredit),
		ClosingBalance: fromPgNumeric(row.ClosingBalance),
		UpdatedAt:      row.UpdatedAt.Time,
	}, nil
}

