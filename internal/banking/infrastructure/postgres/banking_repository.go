package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/banking/domain"
	"erp/internal/banking/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type bankingRepository struct {
	db *database.DB
}

func NewBankingRepository(db *database.DB) ports.BankingRepository {
	return &bankingRepository{db: db}
}

func (r *bankingRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// ============================================================================
// BANK ACCOUNTS
// ============================================================================

func (r *bankingRepository) CreateBankAccount(ctx context.Context, a *domain.BankAccount) (*domain.BankAccount, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.AccountNumberMasked == "" {
		a.AccountNumberMasked = domain.MaskAccountNumber(a.AccountNumber)
	}
	row, err := r.queries(ctx).CreateBankAccount(ctx, sqlc.CreateBankAccountParams{
		ID:                  toPgUUID(a.ID),
		LegalEntityID:       toPgUUID(a.LegalEntityID),
		AccountCode:         a.AccountCode,
		BankNameEn:          a.BankNameEn,
		BankNameAr:          a.BankNameAr,
		BranchName:          toPgTextPtr(a.BranchName),
		AccountNumber:       a.AccountNumber,
		AccountNumberMasked: a.AccountNumberMasked,
		Iban:                a.IBAN,
		SwiftBic:            a.SwiftBIC,
		CurrencyCode:        a.CurrencyCode,
		GlAccountID:         toPgUUIDPtr(a.GLAccountID),
		OpeningBalance:      toPgNumeric(a.OpeningBalance),
		CurrentBalance:      toPgNumeric(a.CurrentBalance),
		AvailableBalance:    toPgNumeric(a.AvailableBalance),
		OverdraftLimit:      toPgNumeric(a.OverdraftLimit),
		Status:              string(a.Status),
		IsPrimary:           a.IsPrimary,
		ContactPerson:       toPgTextPtr(a.ContactPerson),
		ContactPhone:        toPgTextPtr(a.ContactPhone),
	})
	if err != nil {
		return nil, err
	}
	return mapBankAccount(row), nil
}

func (r *bankingRepository) GetBankAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.BankAccount, error) {
	row, err := r.queries(ctx).GetBankAccountByID(ctx, sqlc.GetBankAccountByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBankAccountNotFound
		}
		return nil, err
	}
	return mapBankAccount(row), nil
}

func (r *bankingRepository) GetBankAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.BankAccount, error) {
	row, err := r.queries(ctx).GetBankAccountByCode(ctx, sqlc.GetBankAccountByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		AccountCode:   code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBankAccountNotFound
		}
		return nil, err
	}
	return mapBankAccount(row), nil
}

func (r *bankingRepository) GetBankAccountByIBAN(ctx context.Context, legalEntityID uuid.UUID, iban string) (*domain.BankAccount, error) {
	row, err := r.queries(ctx).GetBankAccountByIBAN(ctx, sqlc.GetBankAccountByIBANParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Iban:          iban,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBankAccountNotFound
		}
		return nil, err
	}
	return mapBankAccount(row), nil
}

func (r *bankingRepository) ListBankAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BankAccount, error) {
	rows, err := r.queries(ctx).ListBankAccounts(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.BankAccount, len(rows))
	for i, row := range rows {
		res[i] = mapBankAccount(row)
	}
	return res, nil
}

func (r *bankingRepository) UpdateBankAccountBalance(ctx context.Context, legalEntityID, id uuid.UUID, delta decimal.Decimal) (*domain.BankAccount, error) {
	row, err := r.queries(ctx).UpdateBankAccountBalance(ctx, sqlc.UpdateBankAccountBalanceParams{
		ID:             toPgUUID(id),
		LegalEntityID:  toPgUUID(legalEntityID),
		CurrentBalance: toPgNumeric(delta),
	})
	if err != nil {
		return nil, err
	}
	return mapBankAccount(row), nil
}

func (r *bankingRepository) UpdateBankAccountStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.BankAccountStatus) (*domain.BankAccount, error) {
	row, err := r.queries(ctx).UpdateBankAccountStatus(ctx, sqlc.UpdateBankAccountStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapBankAccount(row), nil
}

// ============================================================================
// STATEMENTS
// ============================================================================

func (r *bankingRepository) CreateBankStatement(ctx context.Context, s *domain.Statement) (*domain.Statement, error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateBankStatement(ctx, sqlc.CreateBankStatementParams{
		ID:               toPgUUID(s.ID),
		LegalEntityID:    toPgUUID(s.LegalEntityID),
		BankAccountID:    toPgUUID(s.BankAccountID),
		StatementNumber:  s.StatementNumber,
		StatementDate:    toPgDate(s.StatementDate),
		StartDate:        toPgDate(s.StartDate),
		EndDate:          toPgDate(s.EndDate),
		OpeningBalance:   toPgNumeric(s.OpeningBalance),
		ClosingBalance:   toPgNumeric(s.ClosingBalance),
		TotalCredits:     toPgNumeric(s.TotalCredits),
		TotalDebits:      toPgNumeric(s.TotalDebits),
		CurrencyCode:     s.CurrencyCode,
		FileFormat:       string(s.FileFormat),
		Status:           string(s.Status),
		SourceFilename:   toPgTextPtr(s.SourceFilename),
		ImportedByUserID: toPgUUIDPtr(s.ImportedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapStatement(row), nil
}

func (r *bankingRepository) GetBankStatementByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Statement, error) {
	row, err := r.queries(ctx).GetBankStatementByID(ctx, sqlc.GetBankStatementByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStatementNotFound
		}
		return nil, err
	}
	return mapStatement(row), nil
}

func (r *bankingRepository) ListBankStatements(ctx context.Context, bankAccountID uuid.UUID) ([]*domain.Statement, error) {
	rows, err := r.queries(ctx).ListBankStatements(ctx, toPgUUID(bankAccountID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Statement, len(rows))
	for i, row := range rows {
		res[i] = mapStatement(row)
	}
	return res, nil
}

func (r *bankingRepository) UpdateBankStatementStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.StatementStatus) (*domain.Statement, error) {
	row, err := r.queries(ctx).UpdateBankStatementStatus(ctx, sqlc.UpdateBankStatementStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapStatement(row), nil
}

// ============================================================================
// STATEMENT LINES
// ============================================================================

func (r *bankingRepository) CreateStatementLine(ctx context.Context, l *domain.StatementLine) (*domain.StatementLine, error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	var valDate time.Time
	if l.ValueDate != nil {
		valDate = *l.ValueDate
	}
	row, err := r.queries(ctx).CreateStatementLine(ctx, sqlc.CreateStatementLineParams{
		ID:               toPgUUID(l.ID),
		StatementID:      toPgUUID(l.StatementID),
		LineNumber:       l.LineNumber,
		TransactionDate:  toPgDate(l.TransactionDate),
		ValueDate:        toPgDate(valDate),
		Description:      l.Description,
		ReferenceNumber:  toPgTextPtr(l.ReferenceNumber),
		TransactionType:  string(l.TransactionType),
		Amount:           toPgNumeric(l.Amount),
		CurrencyCode:     l.CurrencyCode,
		BalanceAfter:     toPgNumericPtr(l.BalanceAfter),
		ReconciledStatus: string(l.ReconciledStatus),
	})
	if err != nil {
		return nil, err
	}
	return mapStatementLine(row), nil
}

func (r *bankingRepository) GetStatementLineByID(ctx context.Context, id uuid.UUID) (*domain.StatementLine, error) {
	row, err := r.queries(ctx).GetStatementLineByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStatementLineNotFound
		}
		return nil, err
	}
	return mapStatementLine(row), nil
}

func (r *bankingRepository) LockStatementLineByID(ctx context.Context, id uuid.UUID) (*domain.StatementLine, error) {
	row, err := r.queries(ctx).LockStatementLineByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStatementLineNotFound
		}
		return nil, err
	}
	return mapStatementLine(row), nil
}

func (r *bankingRepository) ListStatementLines(ctx context.Context, statementID uuid.UUID) ([]*domain.StatementLine, error) {
	rows, err := r.queries(ctx).ListStatementLines(ctx, toPgUUID(statementID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.StatementLine, len(rows))
	for i, row := range rows {
		res[i] = mapStatementLine(row)
	}
	return res, nil
}

func (r *bankingRepository) UpdateStatementLineReconciledStatus(ctx context.Context, id uuid.UUID, status domain.ReconciledStatus) (*domain.StatementLine, error) {
	row, err := r.queries(ctx).UpdateStatementLineReconciledStatus(ctx, sqlc.UpdateStatementLineReconciledStatusParams{
		ID:               toPgUUID(id),
		ReconciledStatus: string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapStatementLine(row), nil
}

// ============================================================================
// RECONCILIATIONS
// ============================================================================

func (r *bankingRepository) CreateBankReconciliation(ctx context.Context, recon *domain.Reconciliation) (*domain.Reconciliation, error) {
	if recon.ID == uuid.Nil {
		recon.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateBankReconciliation(ctx, sqlc.CreateBankReconciliationParams{
		ID:                     toPgUUID(recon.ID),
		LegalEntityID:          toPgUUID(recon.LegalEntityID),
		BankAccountID:          toPgUUID(recon.BankAccountID),
		StatementID:            toPgUUIDPtr(recon.StatementID),
		ReconciliationNumber:   recon.ReconciliationNumber,
		AsOfDate:               toPgDate(recon.AsOfDate),
		StatementEndingBalance: toPgNumeric(recon.StatementEndingBalance),
		GlBookBalance:          toPgNumeric(recon.GLBookBalance),
		UnreconciledDifference: toPgNumeric(recon.UnreconciledDifference),
		Status:                 string(recon.Status),
		Notes:                  toPgTextPtr(recon.Notes),
		ReconciledByUserID:     toPgUUIDPtr(recon.ReconciledByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapReconciliation(row), nil
}

func (r *bankingRepository) GetBankReconciliationByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error) {
	row, err := r.queries(ctx).GetBankReconciliationByID(ctx, sqlc.GetBankReconciliationByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReconciliationNotFound
		}
		return nil, err
	}
	return mapReconciliation(row), nil
}

func (r *bankingRepository) LockBankReconciliationByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Reconciliation, error) {
	row, err := r.queries(ctx).LockBankReconciliationByID(ctx, sqlc.LockBankReconciliationByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReconciliationNotFound
		}
		return nil, err
	}
	return mapReconciliation(row), nil
}

func (r *bankingRepository) ListBankReconciliations(ctx context.Context, bankAccountID uuid.UUID) ([]*domain.Reconciliation, error) {
	rows, err := r.queries(ctx).ListBankReconciliations(ctx, toPgUUID(bankAccountID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Reconciliation, len(rows))
	for i, row := range rows {
		res[i] = mapReconciliation(row)
	}
	return res, nil
}

func (r *bankingRepository) UpdateBankReconciliation(ctx context.Context, recon *domain.Reconciliation) (*domain.Reconciliation, error) {
	row, err := r.queries(ctx).UpdateBankReconciliation(ctx, sqlc.UpdateBankReconciliationParams{
		ID:                     toPgUUID(recon.ID),
		LegalEntityID:          toPgUUID(recon.LegalEntityID),
		Status:                 string(recon.Status),
		StatementEndingBalance: toPgNumeric(recon.StatementEndingBalance),
		GlBookBalance:          toPgNumeric(recon.GLBookBalance),
		UnreconciledDifference: toPgNumeric(recon.UnreconciledDifference),
		Notes:                  toPgTextPtr(recon.Notes),
		ReconciledByUserID:     toPgUUIDPtr(recon.ReconciledByUserID),
		ReconciledAt:           toPgTimestamptzPtr(recon.ReconciledAt),
	})
	if err != nil {
		return nil, err
	}
	return mapReconciliation(row), nil
}

// ============================================================================
// MATCHES
// ============================================================================

func (r *bankingRepository) CreateReconciliationMatch(ctx context.Context, m *domain.ReconciliationMatch) (*domain.ReconciliationMatch, error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateReconciliationMatch(ctx, sqlc.CreateReconciliationMatchParams{
		ID:                toPgUUID(m.ID),
		ReconciliationID:  toPgUUID(m.ReconciliationID),
		StatementLineID:   toPgUUID(m.StatementLineID),
		MatchType:         string(m.MatchType),
		MatchedEntityType: string(m.MatchedEntityType),
		MatchedEntityID:   toPgUUIDPtr(m.MatchedEntityID),
		MatchedAmount:     toPgNumeric(m.MatchedAmount),
		DifferenceAmount:  toPgNumeric(m.DifferenceAmount),
		Notes:             toPgTextPtr(m.Notes),
		MatchedByUserID:   toPgUUIDPtr(m.MatchedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapMatch(row), nil
}

func (r *bankingRepository) ListReconciliationMatches(ctx context.Context, reconciliationID uuid.UUID) ([]*domain.ReconciliationMatch, error) {
	rows, err := r.queries(ctx).ListReconciliationMatches(ctx, toPgUUID(reconciliationID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.ReconciliationMatch, len(rows))
	for i, row := range rows {
		res[i] = mapMatch(row)
	}
	return res, nil
}

func (r *bankingRepository) DeleteReconciliationMatch(ctx context.Context, id uuid.UUID) error {
	return r.queries(ctx).DeleteReconciliationMatch(ctx, toPgUUID(id))
}

// ============================================================================
// MAPPERS
// ============================================================================

func mapBankAccount(r sqlc.BankingBankAccounts) *domain.BankAccount {
	return &domain.BankAccount{
		ID:                  toUUID(r.ID),
		LegalEntityID:       toUUID(r.LegalEntityID),
		AccountCode:         r.AccountCode,
		BankNameEn:          r.BankNameEn,
		BankNameAr:          r.BankNameAr,
		BranchName:          fromPgTextPtr(r.BranchName),
		AccountNumber:       r.AccountNumber,
		AccountNumberMasked: r.AccountNumberMasked,
		IBAN:                r.Iban,
		SwiftBIC:            r.SwiftBic,
		CurrencyCode:        r.CurrencyCode,
		GLAccountID:         toUUIDPtr(r.GlAccountID),
		OpeningBalance:      fromPgNumeric(r.OpeningBalance),
		CurrentBalance:      fromPgNumeric(r.CurrentBalance),
		AvailableBalance:    fromPgNumeric(r.AvailableBalance),
		OverdraftLimit:      fromPgNumeric(r.OverdraftLimit),
		Status:              domain.BankAccountStatus(r.Status),
		IsPrimary:           r.IsPrimary,
		ContactPerson:       fromPgTextPtr(r.ContactPerson),
		ContactPhone:        fromPgTextPtr(r.ContactPhone),
		CreatedAt:           fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:           fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapStatement(r sqlc.BankingStatements) *domain.Statement {
	return &domain.Statement{
		ID:               toUUID(r.ID),
		LegalEntityID:    toUUID(r.LegalEntityID),
		BankAccountID:    toUUID(r.BankAccountID),
		StatementNumber:  r.StatementNumber,
		StatementDate:    fromPgDate(r.StatementDate),
		StartDate:        fromPgDate(r.StartDate),
		EndDate:          fromPgDate(r.EndDate),
		OpeningBalance:   fromPgNumeric(r.OpeningBalance),
		ClosingBalance:   fromPgNumeric(r.ClosingBalance),
		TotalCredits:     fromPgNumeric(r.TotalCredits),
		TotalDebits:      fromPgNumeric(r.TotalDebits),
		CurrencyCode:     r.CurrencyCode,
		FileFormat:       domain.StatementFormat(r.FileFormat),
		Status:           domain.StatementStatus(r.Status),
		SourceFilename:   fromPgTextPtr(r.SourceFilename),
		ImportedByUserID: toUUIDPtr(r.ImportedByUserID),
		CreatedAt:        fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:        fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapStatementLine(r sqlc.BankingStatementLines) *domain.StatementLine {
	var valDate *time.Time
	if r.ValueDate.Valid {
		t := r.ValueDate.Time
		valDate = &t
	}
	return &domain.StatementLine{
		ID:               toUUID(r.ID),
		StatementID:      toUUID(r.StatementID),
		LineNumber:       r.LineNumber,
		TransactionDate:  fromPgDate(r.TransactionDate),
		ValueDate:        valDate,
		Description:      r.Description,
		ReferenceNumber:  fromPgTextPtr(r.ReferenceNumber),
		TransactionType:  domain.TransactionType(r.TransactionType),
		Amount:           fromPgNumeric(r.Amount),
		CurrencyCode:     r.CurrencyCode,
		BalanceAfter:     fromPgNumericPtr(r.BalanceAfter),
		ReconciledStatus: domain.ReconciledStatus(r.ReconciledStatus),
		CreatedAt:        fromPgTimestamptz(r.CreatedAt),
	}
}

func mapReconciliation(r sqlc.BankingReconciliations) *domain.Reconciliation {
	return &domain.Reconciliation{
		ID:                     toUUID(r.ID),
		LegalEntityID:          toUUID(r.LegalEntityID),
		BankAccountID:          toUUID(r.BankAccountID),
		StatementID:            toUUIDPtr(r.StatementID),
		ReconciliationNumber:   r.ReconciliationNumber,
		AsOfDate:               fromPgDate(r.AsOfDate),
		StatementEndingBalance: fromPgNumeric(r.StatementEndingBalance),
		GLBookBalance:          fromPgNumeric(r.GlBookBalance),
		UnreconciledDifference: fromPgNumeric(r.UnreconciledDifference),
		Status:                 domain.ReconciliationStatus(r.Status),
		Notes:                  fromPgTextPtr(r.Notes),
		ReconciledByUserID:     toUUIDPtr(r.ReconciledByUserID),
		ReconciledAt:           fromPgTimestamptzPtr(r.ReconciledAt),
		CreatedAt:              fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:              fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapMatch(r sqlc.BankingReconciliationMatches) *domain.ReconciliationMatch {
	return &domain.ReconciliationMatch{
		ID:                toUUID(r.ID),
		ReconciliationID:  toUUID(r.ReconciliationID),
		StatementLineID:   toUUID(r.StatementLineID),
		MatchType:         domain.MatchType(r.MatchType),
		MatchedEntityType: domain.MatchedEntityType(r.MatchedEntityType),
		MatchedEntityID:   toUUIDPtr(r.MatchedEntityID),
		MatchedAmount:     fromPgNumeric(r.MatchedAmount),
		DifferenceAmount:  fromPgNumeric(r.DifferenceAmount),
		Notes:             fromPgTextPtr(r.Notes),
		MatchedByUserID:   toUUIDPtr(r.MatchedByUserID),
		MatchedAt:         fromPgTimestamptz(r.MatchedAt),
	}
}
