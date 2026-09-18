package postgres

import (
	"context"
	"errors"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type receivablesRepository struct {
	db *database.DB
}

func NewReceivablesRepository(db *database.DB) ports.ReceivablesRepository {
	return &receivablesRepository{db: db}
}

func (r *receivablesRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

func (r *receivablesRepository) CreateReceipt(ctx context.Context, receipt *domain.Receipt) (*domain.Receipt, error) {
	row, err := r.queries(ctx).CreateReceipt(ctx, sqlc.CreateReceiptParams{
		LegalEntityID:     toPgUUID(receipt.LegalEntityID),
		ReceiptNumber:     receipt.ReceiptNumber,
		CustomerID:        toPgUUID(receipt.CustomerID),
		ReceiptDate:       toPgDate(receipt.ReceiptDate),
		PaymentMethod:     string(receipt.PaymentMethod),
		DepositAccountID:  toPgUUID(receipt.DepositAccountID),
		Amount:            toPgNumeric(receipt.Amount),
		CurrencyCode:      receipt.CurrencyCode,
		ReferenceNumber:   toPgTextPtr(receipt.ReferenceNumber),
		Status:            string(receipt.Status),
		Notes:             toPgTextPtr(receipt.Notes),
		AllocatedAmount:   toPgNumeric(receipt.AllocatedAmount),
		UnallocatedAmount: toPgNumeric(receipt.UnallocatedAmount),
		JournalEntryID:    toPgUUIDPtr(receipt.JournalEntryID),
		CreatedByUserID:   toPgUUIDPtr(receipt.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	receipt.ID = toUUID(row.ID)
	receipt.CreatedAt = row.CreatedAt.Time
	return receipt, nil
}

func (r *receivablesRepository) GetReceiptByID(ctx context.Context, id uuid.UUID) (*domain.Receipt, error) {
	row, err := r.queries(ctx).GetReceiptByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReceiptNotFound
		}
		return nil, err
	}
	return &domain.Receipt{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		ReceiptNumber:        row.ReceiptNumber,
		CustomerID:           toUUID(row.CustomerID),
		ReceiptDate:          fromPgDate(row.ReceiptDate),
		PaymentMethod:        domain.PaymentMethod(row.PaymentMethod),
		DepositAccountID:     toUUID(row.DepositAccountID),
		Amount:               fromPgNumeric(row.Amount),
		CurrencyCode:         row.CurrencyCode,
		ReferenceNumber:      fromPgTextPtr(row.ReferenceNumber),
		Status:               domain.ReceiptStatus(row.Status),
		Notes:                fromPgTextPtr(row.Notes),
		AllocatedAmount:      fromPgNumeric(row.AllocatedAmount),
		UnallocatedAmount:    fromPgNumeric(row.UnallocatedAmount),
		JournalEntryID:       toUUIDPtr(row.JournalEntryID),
		CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
		CustomerNumber:       row.CustomerNumber,
		CustomerNameAr:       row.CustomerNameAr,
		CustomerNameEn:       row.CustomerNameEn,
		DepositAccountCode:   row.DepositAccountCode,
		DepositAccountNameAr: row.DepositAccountNameAr,
		CreatedAt:            row.CreatedAt.Time,
	}, nil
}

func (r *receivablesRepository) ListReceipts(ctx context.Context, filter ports.ReceiptFilter) ([]*domain.Receipt, int64, error) {
	var custID pgtype.UUID
	if filter.CustomerID != nil && *filter.CustomerID != uuid.Nil {
		custID = toPgUUID(*filter.CustomerID)
	}

	total, err := r.queries(ctx).CountReceipts(ctx, sqlc.CountReceiptsParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       custID,
		Column3:       filter.Status,
		Column4:       filter.PaymentMethod,
		Column5:       filter.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries(ctx).ListReceipts(ctx, sqlc.ListReceiptsParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       custID,
		Column3:       filter.Status,
		Column4:       filter.PaymentMethod,
		Column5:       filter.Search,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	receipts := make([]*domain.Receipt, len(rows))
	for i, row := range rows {
		receipts[i] = &domain.Receipt{
			ID:                   toUUID(row.ID),
			LegalEntityID:        toUUID(row.LegalEntityID),
			ReceiptNumber:        row.ReceiptNumber,
			CustomerID:           toUUID(row.CustomerID),
			ReceiptDate:          fromPgDate(row.ReceiptDate),
			PaymentMethod:        domain.PaymentMethod(row.PaymentMethod),
			DepositAccountID:     toUUID(row.DepositAccountID),
			Amount:               fromPgNumeric(row.Amount),
			CurrencyCode:         row.CurrencyCode,
			ReferenceNumber:      fromPgTextPtr(row.ReferenceNumber),
			Status:               domain.ReceiptStatus(row.Status),
			Notes:                fromPgTextPtr(row.Notes),
			AllocatedAmount:      fromPgNumeric(row.AllocatedAmount),
			UnallocatedAmount:    fromPgNumeric(row.UnallocatedAmount),
			JournalEntryID:       toUUIDPtr(row.JournalEntryID),
			CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
			CustomerNumber:       row.CustomerNumber,
			CustomerNameAr:       row.CustomerNameAr,
			CustomerNameEn:       row.CustomerNameEn,
			DepositAccountCode:   row.DepositAccountCode,
			DepositAccountNameAr: row.DepositAccountNameAr,
			CreatedAt:            row.CreatedAt.Time,
		}
	}
	return receipts, total, nil
}

func (r *receivablesRepository) UpdateReceiptStatus(ctx context.Context, id uuid.UUID, status domain.ReceiptStatus) error {
	_, err := r.queries(ctx).UpdateReceiptStatus(ctx, sqlc.UpdateReceiptStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	return err
}

func (r *receivablesRepository) UpdateReceiptAllocationAmounts(ctx context.Context, id uuid.UUID, allocated, unallocated decimal.Decimal) error {
	_, err := r.queries(ctx).UpdateReceiptAllocationAmounts(ctx, sqlc.UpdateReceiptAllocationAmountsParams{
		ID:                toPgUUID(id),
		AllocatedAmount:   toPgNumeric(allocated),
		UnallocatedAmount: toPgNumeric(unallocated),
	})
	return err
}

func (r *receivablesRepository) CreateReceiptAllocation(ctx context.Context, allocation *domain.ReceiptAllocation) (*domain.ReceiptAllocation, error) {
	row, err := r.queries(ctx).CreateReceiptAllocation(ctx, sqlc.CreateReceiptAllocationParams{
		ReceiptID:       toPgUUID(allocation.ReceiptID),
		OpenItemID:      toPgUUID(allocation.OpenItemID),
		AllocatedAmount: toPgNumeric(allocation.AllocatedAmount),
		CreatedByUserID: toPgUUIDPtr(allocation.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	allocation.ID = toUUID(row.ID)
	allocation.AllocatedAt = row.AllocatedAt.Time
	return allocation, nil
}

func (r *receivablesRepository) ListReceiptAllocations(ctx context.Context, receiptID uuid.UUID) ([]*domain.ReceiptAllocation, error) {
	rows, err := r.queries(ctx).ListReceiptAllocations(ctx, toPgUUID(receiptID))
	if err != nil {
		return nil, err
	}
	allocations := make([]*domain.ReceiptAllocation, len(rows))
	for i, row := range rows {
		allocations[i] = &domain.ReceiptAllocation{
			ID:              toUUID(row.ID),
			ReceiptID:       toUUID(row.ReceiptID),
			OpenItemID:      toUUID(row.OpenItemID),
			AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
			AllocatedAt:     row.AllocatedAt.Time,
			CreatedByUserID: toUUIDPtr(row.CreatedByUserID),
			DocumentType:    domain.DocumentType(row.DocumentType),
			DocumentNumber:  row.DocumentNumber,
			DueDate:         fromPgDate(row.DueDate),
			OriginalAmount:  fromPgNumeric(row.OriginalAmount),
			RemainingAmount: fromPgNumeric(row.RemainingAmount),
		}
	}
	return allocations, nil
}

func (r *receivablesRepository) CreateOpenItem(ctx context.Context, item *domain.OpenItem) (*domain.OpenItem, error) {
	row, err := r.queries(ctx).CreateOpenItem(ctx, sqlc.CreateOpenItemParams{
		LegalEntityID:   toPgUUID(item.LegalEntityID),
		CustomerID:      toPgUUID(item.CustomerID),
		DocumentType:    string(item.DocumentType),
		DocumentNumber:  item.DocumentNumber,
		DocumentID:      toPgUUIDPtr(item.DocumentID),
		DocumentDate:    toPgDate(item.DocumentDate),
		DueDate:         toPgDate(item.DueDate),
		CurrencyCode:    item.CurrencyCode,
		OriginalAmount:  toPgNumeric(item.OriginalAmount),
		AllocatedAmount: toPgNumeric(item.AllocatedAmount),
		RemainingAmount: toPgNumeric(item.RemainingAmount),
		Status:          string(item.Status),
	})
	if err != nil {
		return nil, err
	}
	item.ID = toUUID(row.ID)
	item.CreatedAt = row.CreatedAt.Time
	item.UpdatedAt = row.UpdatedAt.Time
	return item, nil
}

func (r *receivablesRepository) GetOpenItemByID(ctx context.Context, id uuid.UUID) (*domain.OpenItem, error) {
	row, err := r.queries(ctx).GetOpenItemByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOpenItemNotFound
		}
		return nil, err
	}
	return &domain.OpenItem{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		CustomerID:      toUUID(row.CustomerID),
		DocumentType:    domain.DocumentType(row.DocumentType),
		DocumentNumber:  row.DocumentNumber,
		DocumentID:      toUUIDPtr(row.DocumentID),
		DocumentDate:    fromPgDate(row.DocumentDate),
		DueDate:         fromPgDate(row.DueDate),
		CurrencyCode:    row.CurrencyCode,
		OriginalAmount:  fromPgNumeric(row.OriginalAmount),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		RemainingAmount: fromPgNumeric(row.RemainingAmount),
		Status:          domain.OpenItemStatus(row.Status),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *receivablesRepository) LockOpenItem(ctx context.Context, id uuid.UUID) (*domain.OpenItem, error) {
	row, err := r.queries(ctx).LockOpenItem(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOpenItemNotFound
		}
		return nil, err
	}
	return &domain.OpenItem{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		CustomerID:      toUUID(row.CustomerID),
		DocumentType:    domain.DocumentType(row.DocumentType),
		DocumentNumber:  row.DocumentNumber,
		DocumentID:      toUUIDPtr(row.DocumentID),
		DocumentDate:    fromPgDate(row.DocumentDate),
		DueDate:         fromPgDate(row.DueDate),
		CurrencyCode:    row.CurrencyCode,
		OriginalAmount:  fromPgNumeric(row.OriginalAmount),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		RemainingAmount: fromPgNumeric(row.RemainingAmount),
		Status:          domain.OpenItemStatus(row.Status),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *receivablesRepository) ListOpenItemsByCustomer(ctx context.Context, customerID uuid.UUID, openOnly bool) ([]*domain.OpenItem, error) {
	rows, err := r.queries(ctx).ListOpenItemsByCustomer(ctx, sqlc.ListOpenItemsByCustomerParams{
		CustomerID: toPgUUID(customerID),
		Column2:    openOnly,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*domain.OpenItem, len(rows))
	for i, row := range rows {
		items[i] = &domain.OpenItem{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			CustomerID:      toUUID(row.CustomerID),
			DocumentType:    domain.DocumentType(row.DocumentType),
			DocumentNumber:  row.DocumentNumber,
			DocumentID:      toUUIDPtr(row.DocumentID),
			DocumentDate:    fromPgDate(row.DocumentDate),
			DueDate:         fromPgDate(row.DueDate),
			CurrencyCode:    row.CurrencyCode,
			OriginalAmount:  fromPgNumeric(row.OriginalAmount),
			AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
			RemainingAmount: fromPgNumeric(row.RemainingAmount),
			Status:          domain.OpenItemStatus(row.Status),
			CustomerNumber:  row.CustomerNumber,
			CustomerNameAr:  row.CustomerNameAr,
			DaysOverdue:     row.DaysOverdue,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		}
	}
	return items, nil
}

func (r *receivablesRepository) ListOpenItemsByLegalEntity(ctx context.Context, legalEntityID uuid.UUID, openOnly bool, limit, offset int32) ([]*domain.OpenItem, error) {
	rows, err := r.queries(ctx).ListOpenItemsByLegalEntity(ctx, sqlc.ListOpenItemsByLegalEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       openOnly,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*domain.OpenItem, len(rows))
	for i, row := range rows {
		items[i] = &domain.OpenItem{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			CustomerID:      toUUID(row.CustomerID),
			DocumentType:    domain.DocumentType(row.DocumentType),
			DocumentNumber:  row.DocumentNumber,
			DocumentID:      toUUIDPtr(row.DocumentID),
			DocumentDate:    fromPgDate(row.DocumentDate),
			DueDate:         fromPgDate(row.DueDate),
			CurrencyCode:    row.CurrencyCode,
			OriginalAmount:  fromPgNumeric(row.OriginalAmount),
			AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
			RemainingAmount: fromPgNumeric(row.RemainingAmount),
			Status:          domain.OpenItemStatus(row.Status),
			CustomerNumber:  row.CustomerNumber,
			CustomerNameAr:  row.CustomerNameAr,
			DaysOverdue:     row.DaysOverdue,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		}
	}
	return items, nil
}

func (r *receivablesRepository) UpdateOpenItemAllocation(ctx context.Context, id uuid.UUID, allocated, remaining decimal.Decimal, status domain.OpenItemStatus) (*domain.OpenItem, error) {
	row, err := r.queries(ctx).UpdateOpenItemAllocation(ctx, sqlc.UpdateOpenItemAllocationParams{
		ID:              toPgUUID(id),
		AllocatedAmount: toPgNumeric(allocated),
		RemainingAmount: toPgNumeric(remaining),
		Status:          string(status),
	})
	if err != nil {
		return nil, err
	}
	return &domain.OpenItem{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		CustomerID:      toUUID(row.CustomerID),
		DocumentType:    domain.DocumentType(row.DocumentType),
		DocumentNumber:  row.DocumentNumber,
		DocumentID:      toUUIDPtr(row.DocumentID),
		DocumentDate:    fromPgDate(row.DocumentDate),
		DueDate:         fromPgDate(row.DueDate),
		CurrencyCode:    row.CurrencyCode,
		OriginalAmount:  fromPgNumeric(row.OriginalAmount),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		RemainingAmount: fromPgNumeric(row.RemainingAmount),
		Status:          domain.OpenItemStatus(row.Status),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *receivablesRepository) GetCustomerAgingSummary(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAgingSummary, error) {
	row, err := r.queries(ctx).GetCustomerAgingSummary(ctx, toPgUUID(customerID))
	if err != nil {
		return nil, err
	}
	return &domain.CustomerAgingSummary{
		TotalOutstanding: fromPgNumeric(row.TotalOutstanding),
		CurrentAmount:    fromPgNumeric(row.CurrentAmount),
		Overdue1to30:     fromPgNumeric(row.Overdue130),
		Overdue31to60:    fromPgNumeric(row.Overdue3160),
		Overdue61to90:    fromPgNumeric(row.Overdue6190),
		OverdueOver90:    fromPgNumeric(row.OverdueOver90),
	}, nil
}
