package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/treasury/domain"
	"erp/internal/treasury/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type treasuryRepository struct {
	db *database.DB
}

func NewTreasuryRepository(db *database.DB) ports.TreasuryRepository {
	return &treasuryRepository{db: db}
}

func (r *treasuryRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// ============================================================================
// CASH ACCOUNTS
// ============================================================================

func (r *treasuryRepository) CreateCashAccount(ctx context.Context, a *domain.CashAccount) (*domain.CashAccount, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateCashAccount(ctx, sqlc.CreateCashAccountParams{
		ID:              toPgUUID(a.ID),
		LegalEntityID:   toPgUUID(a.LegalEntityID),
		BranchID:        toPgUUIDPtr(a.BranchID),
		AccountCode:     a.AccountCode,
		NameEn:          a.NameEn,
		NameAr:          a.NameAr,
		AccountType:     string(a.AccountType),
		GlAccountID:     toPgUUIDPtr(a.GLAccountID),
		CurrencyCode:    a.CurrencyCode,
		CustodianUserID: toPgUUIDPtr(a.CustodianUserID),
		MaxLimit:        toPgNumeric(a.MaxLimit),
		CurrentBalance:  toPgNumeric(a.CurrentBalance),
		Status:          string(a.Status),
	})
	if err != nil {
		return nil, err
	}
	return mapCashAccount(row), nil
}

func (r *treasuryRepository) GetCashAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.CashAccount, error) {
	row, err := r.queries(ctx).GetCashAccountByID(ctx, sqlc.GetCashAccountByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCashAccountNotFound
		}
		return nil, err
	}
	return mapCashAccount(row), nil
}

func (r *treasuryRepository) LockCashAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.CashAccount, error) {
	row, err := r.queries(ctx).LockCashAccountByID(ctx, sqlc.LockCashAccountByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCashAccountNotFound
		}
		return nil, err
	}
	return mapCashAccount(row), nil
}

func (r *treasuryRepository) GetCashAccountByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.CashAccount, error) {
	row, err := r.queries(ctx).GetCashAccountByCode(ctx, sqlc.GetCashAccountByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		AccountCode:   code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCashAccountNotFound
		}
		return nil, err
	}
	return mapCashAccount(row), nil
}

func (r *treasuryRepository) ListCashAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.CashAccount, error) {
	rows, err := r.queries(ctx).ListCashAccounts(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.CashAccount, len(rows))
	for i, row := range rows {
		res[i] = mapCashAccount(row)
	}
	return res, nil
}

func (r *treasuryRepository) UpdateCashAccountBalance(ctx context.Context, legalEntityID, id uuid.UUID, delta decimal.Decimal) (*domain.CashAccount, error) {
	row, err := r.queries(ctx).UpdateCashAccountBalance(ctx, sqlc.UpdateCashAccountBalanceParams{
		ID:             toPgUUID(id),
		LegalEntityID:  toPgUUID(legalEntityID),
		CurrentBalance: toPgNumeric(delta),
	})
	if err != nil {
		return nil, err
	}
	return mapCashAccount(row), nil
}

func (r *treasuryRepository) UpdateCashAccountStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.CashAccountStatus) (*domain.CashAccount, error) {
	row, err := r.queries(ctx).UpdateCashAccountStatus(ctx, sqlc.UpdateCashAccountStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapCashAccount(row), nil
}

// ============================================================================
// RECEIPTS
// ============================================================================

func (r *treasuryRepository) CreateReceipt(ctx context.Context, rc *domain.Receipt) (*domain.Receipt, error) {
	if rc.ID == uuid.Nil {
		rc.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateTreasuryReceipt(ctx, sqlc.CreateTreasuryReceiptParams{
		ID:            toPgUUID(rc.ID),
		LegalEntityID: toPgUUID(rc.LegalEntityID),
		ReceiptNumber: rc.ReceiptNumber,
		CashAccountID: toPgUUID(rc.CashAccountID),
		ReceiptDate:   toPgDate(rc.ReceiptDate),
		ReceiptType:   string(rc.ReceiptType),
		ReceivedFrom:  rc.ReceivedFrom,
		Amount:        toPgNumeric(rc.Amount),
		CurrencyCode:  rc.CurrencyCode,
		PaymentMethod: string(rc.PaymentMethod),
		ChequeNumber:  toPgTextPtr(rc.ChequeNumber),
		ChequeDate: toPgDate(func() time.Time {
			if rc.ChequeDate != nil {
				return *rc.ChequeDate
			}
			return time.Time{}
		}()),
		ReferenceNumber:   toPgTextPtr(rc.ReferenceNumber),
		Description:       toPgTextPtr(rc.Description),
		Status:            string(rc.Status),
		GlOffsetAccountID: toPgUUIDPtr(rc.GLOffsetAccountID),
		JournalEntryID:    toPgUUIDPtr(rc.JournalEntryID),
		CreatedByUserID:   toPgUUIDPtr(rc.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapReceipt(row), nil
}

func (r *treasuryRepository) GetReceiptByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Receipt, error) {
	row, err := r.queries(ctx).GetTreasuryReceiptByID(ctx, sqlc.GetTreasuryReceiptByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReceiptNotFound
		}
		return nil, err
	}
	return mapReceipt(row), nil
}

func (r *treasuryRepository) LockReceiptByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Receipt, error) {
	row, err := r.queries(ctx).LockTreasuryReceiptByID(ctx, sqlc.LockTreasuryReceiptByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReceiptNotFound
		}
		return nil, err
	}
	return mapReceipt(row), nil
}

func (r *treasuryRepository) GetReceiptByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Receipt, error) {
	row, err := r.queries(ctx).GetTreasuryReceiptByNumber(ctx, sqlc.GetTreasuryReceiptByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		ReceiptNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReceiptNotFound
		}
		return nil, err
	}
	return mapReceipt(row), nil
}

func (r *treasuryRepository) ListReceipts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Receipt, error) {
	rows, err := r.queries(ctx).ListTreasuryReceipts(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Receipt, len(rows))
	for i, row := range rows {
		res[i] = mapReceipt(row)
	}
	return res, nil
}

func (r *treasuryRepository) UpdateReceiptStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.ReceiptStatus, journalEntryID *uuid.UUID) (*domain.Receipt, error) {
	row, err := r.queries(ctx).UpdateTreasuryReceiptStatus(ctx, sqlc.UpdateTreasuryReceiptStatusParams{
		ID:             toPgUUID(id),
		LegalEntityID:  toPgUUID(legalEntityID),
		Status:         string(status),
		JournalEntryID: toPgUUIDPtr(journalEntryID),
	})
	if err != nil {
		return nil, err
	}
	return mapReceipt(row), nil
}

// ============================================================================
// PAYMENTS
// ============================================================================

func (r *treasuryRepository) CreatePayment(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateTreasuryPayment(ctx, sqlc.CreateTreasuryPaymentParams{
		ID:                toPgUUID(p.ID),
		LegalEntityID:     toPgUUID(p.LegalEntityID),
		PaymentNumber:     p.PaymentNumber,
		CashAccountID:     toPgUUID(p.CashAccountID),
		PaymentDate:       toPgDate(p.PaymentDate),
		PaymentType:       string(p.PaymentType),
		PaidTo:            p.PaidTo,
		Amount:            toPgNumeric(p.Amount),
		CurrencyCode:      p.CurrencyCode,
		ReferenceNumber:   toPgTextPtr(p.ReferenceNumber),
		Description:       toPgTextPtr(p.Description),
		Status:            string(p.Status),
		GlOffsetAccountID: toPgUUIDPtr(p.GLOffsetAccountID),
		JournalEntryID:    toPgUUIDPtr(p.JournalEntryID),
		CreatedByUserID:   toPgUUIDPtr(p.CreatedByUserID),
		ApprovedByUserID:  toPgUUIDPtr(p.ApprovedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapPayment(row), nil
}

func (r *treasuryRepository) GetPaymentByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Payment, error) {
	row, err := r.queries(ctx).GetTreasuryPaymentByID(ctx, sqlc.GetTreasuryPaymentByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return mapPayment(row), nil
}

func (r *treasuryRepository) LockPaymentByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Payment, error) {
	row, err := r.queries(ctx).LockTreasuryPaymentByID(ctx, sqlc.LockTreasuryPaymentByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return mapPayment(row), nil
}

func (r *treasuryRepository) GetPaymentByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Payment, error) {
	row, err := r.queries(ctx).GetTreasuryPaymentByNumber(ctx, sqlc.GetTreasuryPaymentByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		PaymentNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return mapPayment(row), nil
}

func (r *treasuryRepository) ListPayments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Payment, error) {
	rows, err := r.queries(ctx).ListTreasuryPayments(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Payment, len(rows))
	for i, row := range rows {
		res[i] = mapPayment(row)
	}
	return res, nil
}

func (r *treasuryRepository) UpdatePaymentStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.PaymentStatus, journalEntryID, approvedBy *uuid.UUID) (*domain.Payment, error) {
	row, err := r.queries(ctx).UpdateTreasuryPaymentStatus(ctx, sqlc.UpdateTreasuryPaymentStatusParams{
		ID:               toPgUUID(id),
		LegalEntityID:    toPgUUID(legalEntityID),
		Status:           string(status),
		JournalEntryID:   toPgUUIDPtr(journalEntryID),
		ApprovedByUserID: toPgUUIDPtr(approvedBy),
	})
	if err != nil {
		return nil, err
	}
	return mapPayment(row), nil
}

// ============================================================================
// TRANSFERS
// ============================================================================

func (r *treasuryRepository) CreateTransfer(ctx context.Context, t *domain.Transfer) (*domain.Transfer, error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	row, err := r.queries(ctx).CreateTreasuryTransfer(ctx, sqlc.CreateTreasuryTransferParams{
		ID:                       toPgUUID(t.ID),
		LegalEntityID:            toPgUUID(t.LegalEntityID),
		TransferNumber:           t.TransferNumber,
		TransferDate:             toPgDate(t.TransferDate),
		SourceType:               string(t.SourceType),
		SourceCashAccountID:      toPgUUIDPtr(t.SourceCashAccountID),
		SourceBankAccountID:      toPgUUIDPtr(t.SourceBankAccountID),
		DestinationType:          string(t.DestinationType),
		DestinationCashAccountID: toPgUUIDPtr(t.DestinationCashAccountID),
		DestinationBankAccountID: toPgUUIDPtr(t.DestinationBankAccountID),
		Amount:                   toPgNumeric(t.Amount),
		CurrencyCode:             t.CurrencyCode,
		FeeAmount:                toPgNumeric(t.FeeAmount),
		FeeGlAccountID:           toPgUUIDPtr(t.FeeGLAccountID),
		Status:                   string(t.Status),
		ReferenceNumber:          toPgTextPtr(t.ReferenceNumber),
		Notes:                    toPgTextPtr(t.Notes),
		JournalEntryID:           toPgUUIDPtr(t.JournalEntryID),
		InitiatedByUserID:        toPgUUIDPtr(t.InitiatedByUserID),
		CompletedByUserID:        toPgUUIDPtr(t.CompletedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapTransfer(row), nil
}

func (r *treasuryRepository) GetTransferByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error) {
	row, err := r.queries(ctx).GetTreasuryTransferByID(ctx, sqlc.GetTreasuryTransferByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransferNotFound
		}
		return nil, err
	}
	return mapTransfer(row), nil
}

func (r *treasuryRepository) LockTransferByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error) {
	row, err := r.queries(ctx).LockTreasuryTransferByID(ctx, sqlc.LockTreasuryTransferByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransferNotFound
		}
		return nil, err
	}
	return mapTransfer(row), nil
}

func (r *treasuryRepository) GetTransferByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Transfer, error) {
	row, err := r.queries(ctx).GetTreasuryTransferByNumber(ctx, sqlc.GetTreasuryTransferByNumberParams{
		LegalEntityID:  toPgUUID(legalEntityID),
		TransferNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransferNotFound
		}
		return nil, err
	}
	return mapTransfer(row), nil
}

func (r *treasuryRepository) ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Transfer, error) {
	rows, err := r.queries(ctx).ListTreasuryTransfers(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Transfer, len(rows))
	for i, row := range rows {
		res[i] = mapTransfer(row)
	}
	return res, nil
}

func (r *treasuryRepository) UpdateTransferStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.TransferStatus, journalEntryID, completedBy *uuid.UUID) (*domain.Transfer, error) {
	row, err := r.queries(ctx).UpdateTreasuryTransferStatus(ctx, sqlc.UpdateTreasuryTransferStatusParams{
		ID:                toPgUUID(id),
		LegalEntityID:     toPgUUID(legalEntityID),
		Status:            string(status),
		JournalEntryID:    toPgUUIDPtr(journalEntryID),
		CompletedByUserID: toPgUUIDPtr(completedBy),
	})
	if err != nil {
		return nil, err
	}
	return mapTransfer(row), nil
}

func (r *treasuryRepository) LockBankAccountByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.BankAccountBalance, error) {
	row, err := r.queries(ctx).LockTreasuryBankAccountByID(ctx, sqlc.LockTreasuryBankAccountByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBankAccountNotFound
		}
		return nil, err
	}
	return mapBankAccountBalance(row), nil
}

func (r *treasuryRepository) UpdateBankAccountBalance(ctx context.Context, legalEntityID, id uuid.UUID, delta decimal.Decimal) (*domain.BankAccountBalance, error) {
	row, err := r.queries(ctx).UpdateTreasuryBankAccountBalance(ctx, sqlc.UpdateTreasuryBankAccountBalanceParams{
		ID:             toPgUUID(id),
		LegalEntityID:  toPgUUID(legalEntityID),
		CurrentBalance: toPgNumeric(delta),
	})
	if err != nil {
		return nil, err
	}
	return mapBankAccountBalance(row), nil
}

// ============================================================================
// MAPPERS
// ============================================================================

func mapCashAccount(r sqlc.TreasuryCashAccounts) *domain.CashAccount {
	return &domain.CashAccount{
		ID:              toUUID(r.ID),
		LegalEntityID:   toUUID(r.LegalEntityID),
		BranchID:        toUUIDPtr(r.BranchID),
		AccountCode:     r.AccountCode,
		NameEn:          r.NameEn,
		NameAr:          r.NameAr,
		AccountType:     domain.CashAccountType(r.AccountType),
		GLAccountID:     toUUIDPtr(r.GlAccountID),
		CurrencyCode:    r.CurrencyCode,
		CustodianUserID: toUUIDPtr(r.CustodianUserID),
		MaxLimit:        fromPgNumeric(r.MaxLimit),
		CurrentBalance:  fromPgNumeric(r.CurrentBalance),
		Status:          domain.CashAccountStatus(r.Status),
		CreatedAt:       fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:       fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapBankAccountBalance(r sqlc.BankingBankAccounts) *domain.BankAccountBalance {
	return &domain.BankAccountBalance{
		GLAccountID:      toUUIDPtr(r.GlAccountID),
		ID:               toUUID(r.ID),
		LegalEntityID:    toUUID(r.LegalEntityID),
		CurrencyCode:     r.CurrencyCode,
		CurrentBalance:   fromPgNumeric(r.CurrentBalance),
		AvailableBalance: fromPgNumeric(r.AvailableBalance),
		OverdraftLimit:   fromPgNumeric(r.OverdraftLimit),
		Status:           r.Status,
	}
}

func mapReceipt(r sqlc.TreasuryReceipts) *domain.Receipt {
	var chkDate *time.Time
	if r.ChequeDate.Valid {
		t := r.ChequeDate.Time
		chkDate = &t
	}
	return &domain.Receipt{
		ID:                toUUID(r.ID),
		LegalEntityID:     toUUID(r.LegalEntityID),
		ReceiptNumber:     r.ReceiptNumber,
		CashAccountID:     toUUID(r.CashAccountID),
		ReceiptDate:       fromPgDate(r.ReceiptDate),
		ReceiptType:       domain.ReceiptType(r.ReceiptType),
		ReceivedFrom:      r.ReceivedFrom,
		Amount:            fromPgNumeric(r.Amount),
		CurrencyCode:      r.CurrencyCode,
		PaymentMethod:     domain.PaymentMethod(r.PaymentMethod),
		ChequeNumber:      fromPgTextPtr(r.ChequeNumber),
		ChequeDate:        chkDate,
		ReferenceNumber:   fromPgTextPtr(r.ReferenceNumber),
		Description:       fromPgTextPtr(r.Description),
		Status:            domain.ReceiptStatus(r.Status),
		GLOffsetAccountID: toUUIDPtr(r.GlOffsetAccountID),
		JournalEntryID:    toUUIDPtr(r.JournalEntryID),
		CreatedByUserID:   toUUIDPtr(r.CreatedByUserID),
		CreatedAt:         fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:         fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapPayment(r sqlc.TreasuryPayments) *domain.Payment {
	return &domain.Payment{
		ID:                toUUID(r.ID),
		LegalEntityID:     toUUID(r.LegalEntityID),
		PaymentNumber:     r.PaymentNumber,
		CashAccountID:     toUUID(r.CashAccountID),
		PaymentDate:       fromPgDate(r.PaymentDate),
		PaymentType:       domain.PaymentType(r.PaymentType),
		PaidTo:            r.PaidTo,
		Amount:            fromPgNumeric(r.Amount),
		CurrencyCode:      r.CurrencyCode,
		ReferenceNumber:   fromPgTextPtr(r.ReferenceNumber),
		Description:       fromPgTextPtr(r.Description),
		Status:            domain.PaymentStatus(r.Status),
		GLOffsetAccountID: toUUIDPtr(r.GlOffsetAccountID),
		JournalEntryID:    toUUIDPtr(r.JournalEntryID),
		CreatedByUserID:   toUUIDPtr(r.CreatedByUserID),
		ApprovedByUserID:  toUUIDPtr(r.ApprovedByUserID),
		CreatedAt:         fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:         fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapTransfer(r sqlc.TreasuryTransfers) *domain.Transfer {
	return &domain.Transfer{
		ID:                       toUUID(r.ID),
		LegalEntityID:            toUUID(r.LegalEntityID),
		TransferNumber:           r.TransferNumber,
		TransferDate:             fromPgDate(r.TransferDate),
		SourceType:               domain.TransferSourceType(r.SourceType),
		SourceCashAccountID:      toUUIDPtr(r.SourceCashAccountID),
		SourceBankAccountID:      toUUIDPtr(r.SourceBankAccountID),
		DestinationType:          domain.TransferDestinationType(r.DestinationType),
		DestinationCashAccountID: toUUIDPtr(r.DestinationCashAccountID),
		DestinationBankAccountID: toUUIDPtr(r.DestinationBankAccountID),
		Amount:                   fromPgNumeric(r.Amount),
		CurrencyCode:             r.CurrencyCode,
		FeeAmount:                fromPgNumeric(r.FeeAmount),
		FeeGLAccountID:           toUUIDPtr(r.FeeGlAccountID),
		Status:                   domain.TransferStatus(r.Status),
		ReferenceNumber:          fromPgTextPtr(r.ReferenceNumber),
		Notes:                    fromPgTextPtr(r.Notes),
		JournalEntryID:           toUUIDPtr(r.JournalEntryID),
		InitiatedByUserID:        toUUIDPtr(r.InitiatedByUserID),
		CompletedByUserID:        toUUIDPtr(r.CompletedByUserID),
		CreatedAt:                fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:                fromPgTimestamptz(r.UpdatedAt),
	}
}
