package postgres

import (
	"context"
	"errors"

	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type PayableRepository struct {
	db *database.DB
}

func NewPayableRepository(db *database.DB) *PayableRepository {
	return &PayableRepository{db: db}
}

func (r *PayableRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// Supplier Invoices

func (r *PayableRepository) CreateSupplierInvoice(ctx context.Context, inv *domain.SupplierInvoice) (*domain.SupplierInvoice, error) {
	q := r.queries(ctx)

	rate := inv.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}
	curr := inv.CurrencyCode
	if curr == "" {
		curr = "SAR"
	}
	pStatus := inv.PaymentStatus
	if pStatus == "" {
		pStatus = domain.PaymentStatusUnpaid
	}
	status := inv.Status
	if status == "" {
		status = domain.InvoiceStatusDraft
	}

	row, err := q.CreateSupplierInvoice(ctx, sqlc.CreateSupplierInvoiceParams{
		LegalEntityID:         toPgUUID(inv.LegalEntityID),
		InvoiceNumber:         inv.InvoiceNumber,
		SupplierInvoiceNumber: inv.SupplierInvoiceNumber,
		SupplierID:            toPgUUID(inv.SupplierID),
		InvoiceDate:           toPgDate(inv.InvoiceDate),
		DueDate:               toPgDate(inv.DueDate),
		CurrencyCode:          curr,
		ExchangeRate:          toPgNumeric(rate),
		Subtotal:              toPgNumeric(inv.Subtotal),
		DiscountAmount:        toPgNumeric(inv.DiscountAmount),
		TaxableAmount:         toPgNumeric(inv.TaxableAmount),
		TaxAmount:             toPgNumeric(inv.TaxAmount),
		TotalAmount:           toPgNumeric(inv.TotalAmount),
		Status:                string(status),
		PaymentStatus:         string(pStatus),
		Notes:                 toPgTextPtr(inv.Notes),
		CreatedByUserID:       toPgUUIDPtr(inv.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	invID := toUUID(row.ID)

	for _, line := range inv.Lines {
		_, err := q.CreateSupplierInvoiceLine(ctx, sqlc.CreateSupplierInvoiceLineParams{
			InvoiceID:        toPgUUID(invID),
			LineNumber:       line.LineNumber,
			ItemID:           toPgUUIDPtr(line.ItemID),
			ExpenseAccountID: toPgUUID(line.ExpenseAccountID),
			Description:      line.Description,
			Quantity:         toPgNumeric(line.Quantity),
			UnitPrice:        toPgNumeric(line.UnitPrice),
			DiscountPercent:  toPgNumeric(line.DiscountPercent),
			DiscountAmount:   toPgNumeric(line.DiscountAmount),
			TaxRate:          toPgNumeric(line.TaxRate),
			TaxCodeID:        toPgUUIDPtr(line.TaxCodeID),
			TaxAmount:        toPgNumeric(line.TaxAmount),
			Subtotal:         toPgNumeric(line.Subtotal),
			TotalAmount:      toPgNumeric(line.TotalAmount),
		})
		if err != nil {
			return nil, err
		}
	}

	return r.GetSupplierInvoiceByID(ctx, invID)
}

func (r *PayableRepository) GetSupplierInvoiceByID(ctx context.Context, id uuid.UUID) (*domain.SupplierInvoice, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierInvoiceByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, err
	}

	inv := &domain.SupplierInvoice{
		ID:                     toUUID(row.ID),
		LegalEntityID:          toUUID(row.LegalEntityID),
		InvoiceNumber:          row.InvoiceNumber,
		SupplierInvoiceNumber:  row.SupplierInvoiceNumber,
		SupplierID:             toUUID(row.SupplierID),
		SupplierNameEn:         row.SupplierNameEn,
		SupplierNameAr:         row.SupplierNameAr,
		SupplierTaxID:          fromPgTextPtr(row.SupplierTaxID),
		SupplierCRNumber:       fromPgTextPtr(row.SupplierCrNumber),
		SupplierIBAN:           fromPgTextPtr(row.SupplierIban),
		SupplierLiabilityAccID: toUUIDPtr(row.SupplierLiabilityAccountID),
		InvoiceDate:            fromPgDate(row.InvoiceDate),
		DueDate:                fromPgDate(row.DueDate),
		CurrencyCode:           row.CurrencyCode,
		ExchangeRate:           fromPgNumeric(row.ExchangeRate),
		Subtotal:               fromPgNumeric(row.Subtotal),
		DiscountAmount:         fromPgNumeric(row.DiscountAmount),
		TaxableAmount:          fromPgNumeric(row.TaxableAmount),
		TaxAmount:              fromPgNumeric(row.TaxAmount),
		TotalAmount:            fromPgNumeric(row.TotalAmount),
		Status:                 domain.SupplierInvoiceStatus(row.Status),
		PaymentStatus:          domain.PaymentStatus(row.PaymentStatus),
		JournalEntryID:         toUUIDPtr(row.JournalEntryID),
		Notes:                  fromPgTextPtr(row.Notes),
		CreatedByUserID:        toUUIDPtr(row.CreatedByUserID),
		CreatedAt:              fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:              fromPgTimestamptz(row.UpdatedAt),
	}

	lines, err := r.ListSupplierInvoiceLines(ctx, id)
	if err != nil {
		return nil, err
	}
	inv.Lines = lines

	return inv, nil
}

func (r *PayableRepository) GetSupplierInvoiceByNumber(ctx context.Context, legalEntityID uuid.UUID, invoiceNumber string) (*domain.SupplierInvoice, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierInvoiceByNumber(ctx, sqlc.GetSupplierInvoiceByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		InvoiceNumber: invoiceNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, err
	}
	return r.GetSupplierInvoiceByID(ctx, toUUID(row.ID))
}

func (r *PayableRepository) ListSupplierInvoices(ctx context.Context, filter ports.SupplierInvoiceFilter) ([]*domain.SupplierInvoice, int64, error) {
	q := r.queries(ctx)

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	status := filter.Status
	if status == "" {
		status = "ALL"
	}
	pStatus := filter.PaymentStatus
	if pStatus == "" {
		pStatus = "ALL"
	}

	var suppID pgtype.UUID
	if filter.SupplierID != nil {
		suppID = toPgUUID(*filter.SupplierID)
	}

	rows, err := q.ListSupplierInvoices(ctx, sqlc.ListSupplierInvoicesParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       filter.Search,
		Column3:       suppID,
		Column4:       status,
		Column5:       pStatus,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := q.CountSupplierInvoices(ctx, sqlc.CountSupplierInvoicesParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       filter.Search,
		Column3:       suppID,
		Column4:       status,
		Column5:       pStatus,
	})
	if err != nil {
		return nil, 0, err
	}

	var invoices []*domain.SupplierInvoice
	for _, row := range rows {
		invoices = append(invoices, &domain.SupplierInvoice{
			ID:                    toUUID(row.ID),
			LegalEntityID:         toUUID(row.LegalEntityID),
			InvoiceNumber:         row.InvoiceNumber,
			SupplierInvoiceNumber: row.SupplierInvoiceNumber,
			SupplierID:            toUUID(row.SupplierID),
			SupplierNameEn:        row.SupplierNameEn,
			SupplierNameAr:        row.SupplierNameAr,
			SupplierTaxID:         fromPgTextPtr(row.SupplierTaxID),
			InvoiceDate:           fromPgDate(row.InvoiceDate),
			DueDate:               fromPgDate(row.DueDate),
			CurrencyCode:          row.CurrencyCode,
			ExchangeRate:          fromPgNumeric(row.ExchangeRate),
			Subtotal:              fromPgNumeric(row.Subtotal),
			DiscountAmount:        fromPgNumeric(row.DiscountAmount),
			TaxableAmount:         fromPgNumeric(row.TaxableAmount),
			TaxAmount:             fromPgNumeric(row.TaxAmount),
			TotalAmount:           fromPgNumeric(row.TotalAmount),
			Status:                domain.SupplierInvoiceStatus(row.Status),
			PaymentStatus:         domain.PaymentStatus(row.PaymentStatus),
			JournalEntryID:        toUUIDPtr(row.JournalEntryID),
			Notes:                 fromPgTextPtr(row.Notes),
			CreatedByUserID:       toUUIDPtr(row.CreatedByUserID),
			CreatedAt:             fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:             fromPgTimestamptz(row.UpdatedAt),
		})
	}
	return invoices, total, nil
}

func (r *PayableRepository) UpdateSupplierInvoiceStatus(ctx context.Context, id uuid.UUID, status domain.SupplierInvoiceStatus, journalEntryID *uuid.UUID) error {
	q := r.queries(ctx)
	return q.UpdateSupplierInvoiceStatus(ctx, sqlc.UpdateSupplierInvoiceStatusParams{
		ID:             toPgUUID(id),
		Status:         string(status),
		JournalEntryID: toPgUUIDPtr(journalEntryID),
	})
}

func (r *PayableRepository) UpdateSupplierInvoicePaymentStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	q := r.queries(ctx)
	return q.UpdateSupplierInvoicePaymentStatus(ctx, sqlc.UpdateSupplierInvoicePaymentStatusParams{
		ID:            toPgUUID(id),
		PaymentStatus: string(status),
	})
}

func (r *PayableRepository) ListSupplierInvoiceLines(ctx context.Context, invoiceID uuid.UUID) ([]domain.SupplierInvoiceLine, error) {
	q := r.queries(ctx)
	rows, err := q.ListSupplierInvoiceLines(ctx, toPgUUID(invoiceID))
	if err != nil {
		return nil, err
	}

	var lines []domain.SupplierInvoiceLine
	for _, row := range rows {
		lines = append(lines, domain.SupplierInvoiceLine{
			ID:                   toUUID(row.ID),
			InvoiceID:            toUUID(row.InvoiceID),
			LineNumber:           row.LineNumber,
			ItemID:               toUUIDPtr(row.ItemID),
			ItemCode:             fromPgText(row.ItemCode),
			ItemNameEn:           fromPgText(row.ItemNameEn),
			ItemNameAr:           fromPgText(row.ItemNameAr),
			ExpenseAccountID:     toUUID(row.ExpenseAccountID),
			ExpenseAccountNumber: row.ExpenseAccountNumber,
			ExpenseAccountNameAr: row.ExpenseAccountNameAr,
			Description:          row.Description,
			Quantity:             fromPgNumeric(row.Quantity),
			UnitPrice:            fromPgNumeric(row.UnitPrice),
			DiscountPercent:      fromPgNumeric(row.DiscountPercent),
			DiscountAmount:       fromPgNumeric(row.DiscountAmount),
			TaxRate:              fromPgNumeric(row.TaxRate),
			TaxCodeID:            toUUIDPtr(row.TaxCodeID),
			TaxAmount:            fromPgNumeric(row.TaxAmount),
			Subtotal:             fromPgNumeric(row.Subtotal),
			TotalAmount:          fromPgNumeric(row.TotalAmount),
			CreatedAt:            fromPgTimestamptz(row.CreatedAt),
		})
	}
	return lines, nil
}

// Supplier Credit Notes

func (r *PayableRepository) CreateSupplierCreditNote(ctx context.Context, cn *domain.SupplierCreditNote) (*domain.SupplierCreditNote, error) {
	q := r.queries(ctx)
	row, err := q.CreateSupplierCreditNote(ctx, sqlc.CreateSupplierCreditNoteParams{
		LegalEntityID:            toPgUUID(cn.LegalEntityID),
		CreditNoteNumber:         cn.CreditNoteNumber,
		SupplierCreditNoteNumber: toPgTextPtr(cn.SupplierCreditNoteNumber),
		SupplierID:               toPgUUID(cn.SupplierID),
		OriginalInvoiceID:        toPgUUIDPtr(cn.OriginalInvoiceID),
		CreditNoteDate:           toPgDate(cn.CreditNoteDate),
		CurrencyCode:             cn.CurrencyCode,
		Subtotal:                 toPgNumeric(cn.Subtotal),
		TaxAmount:                toPgNumeric(cn.TaxAmount),
		TotalAmount:              toPgNumeric(cn.TotalAmount),
		Reason:                   toPgTextPtr(cn.Reason),
		Status:                   cn.Status,
		JournalEntryID:           toPgUUIDPtr(cn.JournalEntryID),
		CreatedByUserID:          toPgUUIDPtr(cn.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	cnID := toUUID(row.ID)
	for _, line := range cn.Lines {
		_, err := q.CreateSupplierCreditNoteLine(ctx, sqlc.CreateSupplierCreditNoteLineParams{
			CreditNoteID:     toPgUUID(cnID),
			LineNumber:       line.LineNumber,
			ItemID:           toPgUUIDPtr(line.ItemID),
			ExpenseAccountID: toPgUUID(line.ExpenseAccountID),
			Description:      line.Description,
			Quantity:         toPgNumeric(line.Quantity),
			UnitPrice:        toPgNumeric(line.UnitPrice),
			TaxRate:          toPgNumeric(line.TaxRate),
			TaxAmount:        toPgNumeric(line.TaxAmount),
			Subtotal:         toPgNumeric(line.Subtotal),
			TotalAmount:      toPgNumeric(line.TotalAmount),
		})
		if err != nil {
			return nil, err
		}
	}

	return r.GetSupplierCreditNoteByID(ctx, cnID)
}

func (r *PayableRepository) GetSupplierCreditNoteByID(ctx context.Context, id uuid.UUID) (*domain.SupplierCreditNote, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierCreditNoteByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCreditNoteNotFound
		}
		return nil, err
	}

	cn := &domain.SupplierCreditNote{
		ID:                       toUUID(row.ID),
		LegalEntityID:            toUUID(row.LegalEntityID),
		CreditNoteNumber:         row.CreditNoteNumber,
		SupplierCreditNoteNumber: fromPgTextPtr(row.SupplierCreditNoteNumber),
		SupplierID:               toUUID(row.SupplierID),
		SupplierNameEn:           row.SupplierNameEn,
		SupplierNameAr:           row.SupplierNameAr,
		OriginalInvoiceID:        toUUIDPtr(row.OriginalInvoiceID),
		CreditNoteDate:           fromPgDate(row.CreditNoteDate),
		CurrencyCode:             row.CurrencyCode,
		Subtotal:                 fromPgNumeric(row.Subtotal),
		TaxAmount:                fromPgNumeric(row.TaxAmount),
		TotalAmount:              fromPgNumeric(row.TotalAmount),
		Reason:                   fromPgTextPtr(row.Reason),
		Status:                   row.Status,
		JournalEntryID:           toUUIDPtr(row.JournalEntryID),
		CreatedByUserID:          toUUIDPtr(row.CreatedByUserID),
		CreatedAt:                fromPgTimestamptz(row.CreatedAt),
	}

	lineRows, err := q.ListSupplierCreditNoteLines(ctx, toPgUUID(id))
	if err == nil {
		for _, lr := range lineRows {
			cn.Lines = append(cn.Lines, domain.SupplierCreditNoteLine{
				ID:               toUUID(lr.ID),
				CreditNoteID:     toUUID(lr.CreditNoteID),
				LineNumber:       lr.LineNumber,
				ItemID:           toUUIDPtr(lr.ItemID),
				ExpenseAccountID: toUUID(lr.ExpenseAccountID),
				Description:      lr.Description,
				Quantity:         fromPgNumeric(lr.Quantity),
				UnitPrice:        fromPgNumeric(lr.UnitPrice),
				TaxRate:          fromPgNumeric(lr.TaxRate),
				TaxAmount:        fromPgNumeric(lr.TaxAmount),
				Subtotal:         fromPgNumeric(lr.Subtotal),
				TotalAmount:      fromPgNumeric(lr.TotalAmount),
				CreatedAt:        fromPgTimestamptz(lr.CreatedAt),
			})
		}
	}

	return cn, nil
}

func (r *PayableRepository) ListSupplierCreditNotes(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID) ([]*domain.SupplierCreditNote, error) {
	q := r.queries(ctx)
	var sID pgtype.UUID
	if supplierID != nil {
		sID = toPgUUID(*supplierID)
	}

	rows, err := q.ListSupplierCreditNotes(ctx, sqlc.ListSupplierCreditNotesParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       sID,
	})
	if err != nil {
		return nil, err
	}

	var notes []*domain.SupplierCreditNote
	for _, row := range rows {
		notes = append(notes, &domain.SupplierCreditNote{
			ID:                       toUUID(row.ID),
			LegalEntityID:            toUUID(row.LegalEntityID),
			CreditNoteNumber:         row.CreditNoteNumber,
			SupplierCreditNoteNumber: fromPgTextPtr(row.SupplierCreditNoteNumber),
			SupplierID:               toUUID(row.SupplierID),
			SupplierNameEn:           row.SupplierNameEn,
			SupplierNameAr:           row.SupplierNameAr,
			OriginalInvoiceID:        toUUIDPtr(row.OriginalInvoiceID),
			CreditNoteDate:           fromPgDate(row.CreditNoteDate),
			CurrencyCode:             row.CurrencyCode,
			Subtotal:                 fromPgNumeric(row.Subtotal),
			TaxAmount:                fromPgNumeric(row.TaxAmount),
			TotalAmount:              fromPgNumeric(row.TotalAmount),
			Reason:                   fromPgTextPtr(row.Reason),
			Status:                   row.Status,
			JournalEntryID:           toUUIDPtr(row.JournalEntryID),
			CreatedByUserID:          toUUIDPtr(row.CreatedByUserID),
			CreatedAt:                fromPgTimestamptz(row.CreatedAt),
		})
	}
	return notes, nil
}

// Supplier Payments

func (r *PayableRepository) CreateSupplierPayment(ctx context.Context, p *domain.SupplierPayment) (*domain.SupplierPayment, error) {
	q := r.queries(ctx)
	status := p.Status
	if status == "" {
		status = "POSTED"
	}

	row, err := q.CreateSupplierPayment(ctx, sqlc.CreateSupplierPaymentParams{
		LegalEntityID:         toPgUUID(p.LegalEntityID),
		PaymentNumber:         p.PaymentNumber,
		SupplierID:            toPgUUID(p.SupplierID),
		PaymentDate:           toPgDate(p.PaymentDate),
		PaymentMethod:         string(p.PaymentMethod),
		DisbursementAccountID: toPgUUID(p.DisbursementAccountID),
		Amount:                toPgNumeric(p.Amount),
		CurrencyCode:          p.CurrencyCode,
		ReferenceNumber:       toPgTextPtr(p.ReferenceNumber),
		Status:                status,
		Notes:                 toPgTextPtr(p.Notes),
		AllocatedAmount:       toPgNumeric(p.AllocatedAmount),
		UnallocatedAmount:     toPgNumeric(p.UnallocatedAmount),
		JournalEntryID:        toPgUUIDPtr(p.JournalEntryID),
		CreatedByUserID:       toPgUUIDPtr(p.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return r.GetSupplierPaymentByID(ctx, toUUID(row.ID))
}

func (r *PayableRepository) GetSupplierPaymentByID(ctx context.Context, id uuid.UUID) (*domain.SupplierPayment, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierPaymentByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}

	p := &domain.SupplierPayment{
		ID:                        toUUID(row.ID),
		LegalEntityID:             toUUID(row.LegalEntityID),
		PaymentNumber:             row.PaymentNumber,
		SupplierID:                toUUID(row.SupplierID),
		SupplierNameEn:            row.SupplierNameEn,
		SupplierNameAr:            row.SupplierNameAr,
		SupplierNumber:            row.SupplierNumber,
		PaymentDate:               fromPgDate(row.PaymentDate),
		PaymentMethod:             domain.PaymentMethod(row.PaymentMethod),
		DisbursementAccountID:     toUUID(row.DisbursementAccountID),
		DisbursementAccountNumber: row.DisbursementAccountNumber,
		DisbursementAccountNameAr: row.DisbursementAccountNameAr,
		Amount:                    fromPgNumeric(row.Amount),
		CurrencyCode:              row.CurrencyCode,
		ReferenceNumber:           fromPgTextPtr(row.ReferenceNumber),
		Status:                    row.Status,
		Notes:                     fromPgTextPtr(row.Notes),
		AllocatedAmount:           fromPgNumeric(row.AllocatedAmount),
		UnallocatedAmount:         fromPgNumeric(row.UnallocatedAmount),
		JournalEntryID:            toUUIDPtr(row.JournalEntryID),
		CreatedByUserID:           toUUIDPtr(row.CreatedByUserID),
		CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
	}

	allocs, err := r.ListPaymentAllocations(ctx, id)
	if err == nil {
		p.Allocations = allocs
	}

	return p, nil
}

func (r *PayableRepository) ListSupplierPayments(ctx context.Context, filter ports.SupplierPaymentFilter) ([]*domain.SupplierPayment, int64, error) {
	q := r.queries(ctx)

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var suppID pgtype.UUID
	if filter.SupplierID != nil {
		suppID = toPgUUID(*filter.SupplierID)
	}

	rows, err := q.ListSupplierPayments(ctx, sqlc.ListSupplierPaymentsParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       suppID,
		Column3:       filter.Search,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := q.CountSupplierPayments(ctx, sqlc.CountSupplierPaymentsParams{
		LegalEntityID: toPgUUID(filter.LegalEntityID),
		Column2:       suppID,
		Column3:       filter.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	var payments []*domain.SupplierPayment
	for _, row := range rows {
		payments = append(payments, &domain.SupplierPayment{
			ID:                    toUUID(row.ID),
			LegalEntityID:         toUUID(row.LegalEntityID),
			PaymentNumber:         row.PaymentNumber,
			SupplierID:            toUUID(row.SupplierID),
			SupplierNameEn:        row.SupplierNameEn,
			SupplierNameAr:        row.SupplierNameAr,
			SupplierNumber:        row.SupplierNumber,
			PaymentDate:           fromPgDate(row.PaymentDate),
			PaymentMethod:         domain.PaymentMethod(row.PaymentMethod),
			DisbursementAccountID: toUUID(row.DisbursementAccountID),
			Amount:                fromPgNumeric(row.Amount),
			CurrencyCode:          row.CurrencyCode,
			ReferenceNumber:       fromPgTextPtr(row.ReferenceNumber),
			Status:                row.Status,
			Notes:                 fromPgTextPtr(row.Notes),
			AllocatedAmount:       fromPgNumeric(row.AllocatedAmount),
			UnallocatedAmount:     fromPgNumeric(row.UnallocatedAmount),
			JournalEntryID:        toUUIDPtr(row.JournalEntryID),
			CreatedByUserID:       toUUIDPtr(row.CreatedByUserID),
			CreatedAt:             fromPgTimestamptz(row.CreatedAt),
		})
	}
	return payments, total, nil
}

func (r *PayableRepository) UpdateSupplierPaymentStatus(ctx context.Context, id uuid.UUID, status string, allocated, unallocated decimal.Decimal, journalEntryID *uuid.UUID) error {
	q := r.queries(ctx)
	return q.UpdateSupplierPaymentStatus(ctx, sqlc.UpdateSupplierPaymentStatusParams{
		ID:                toPgUUID(id),
		Status:            status,
		AllocatedAmount:   toPgNumeric(allocated),
		UnallocatedAmount: toPgNumeric(unallocated),
		JournalEntryID:    toPgUUIDPtr(journalEntryID),
	})
}

// Payment Allocations

func (r *PayableRepository) CreatePaymentAllocation(ctx context.Context, alloc *domain.PaymentAllocation) (*domain.PaymentAllocation, error) {
	q := r.queries(ctx)
	row, err := q.CreateSupplierPaymentAllocation(ctx, sqlc.CreateSupplierPaymentAllocationParams{
		PaymentID:       toPgUUID(alloc.PaymentID),
		OpenItemID:      toPgUUID(alloc.OpenItemID),
		AllocatedAmount: toPgNumeric(alloc.AllocatedAmount),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PaymentAllocation{
		ID:              toUUID(row.ID),
		PaymentID:       toUUID(row.PaymentID),
		OpenItemID:      toUUID(row.OpenItemID),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		AllocatedAt:     fromPgTimestamptz(row.AllocatedAt),
	}, nil
}

func (r *PayableRepository) ListPaymentAllocations(ctx context.Context, paymentID uuid.UUID) ([]domain.PaymentAllocation, error) {
	q := r.queries(ctx)
	rows, err := q.ListSupplierPaymentAllocations(ctx, toPgUUID(paymentID))
	if err != nil {
		return nil, err
	}
	var allocs []domain.PaymentAllocation
	for _, row := range rows {
		allocs = append(allocs, domain.PaymentAllocation{
			ID:              toUUID(row.ID),
			PaymentID:       toUUID(row.PaymentID),
			OpenItemID:      toUUID(row.OpenItemID),
			DocumentType:    domain.OpenItemDocumentType(row.DocumentType),
			DocumentNumber:  row.DocumentNumber,
			DueDate:         fromPgDate(row.DueDate),
			OriginalAmount:  fromPgNumeric(row.OriginalAmount),
			RemainingAmount: fromPgNumeric(row.RemainingAmount),
			AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
			AllocatedAt:     fromPgTimestamptz(row.AllocatedAt),
		})
	}
	return allocs, nil
}

// Open Items

func (r *PayableRepository) CreateOpenItem(ctx context.Context, item *domain.PayablesOpenItem) (*domain.PayablesOpenItem, error) {
	q := r.queries(ctx)
	status := item.Status
	if status == "" {
		status = domain.OpenItemStatusOpen
	}

	row, err := q.CreatePayablesOpenItem(ctx, sqlc.CreatePayablesOpenItemParams{
		LegalEntityID:   toPgUUID(item.LegalEntityID),
		SupplierID:      toPgUUID(item.SupplierID),
		DocumentType:    string(item.DocumentType),
		DocumentNumber:  item.DocumentNumber,
		DocumentID:      toPgUUIDPtr(item.DocumentID),
		DocumentDate:    toPgDate(item.DocumentDate),
		DueDate:         toPgDate(item.DueDate),
		CurrencyCode:    item.CurrencyCode,
		OriginalAmount:  toPgNumeric(item.OriginalAmount),
		AllocatedAmount: toPgNumeric(item.AllocatedAmount),
		RemainingAmount: toPgNumeric(item.RemainingAmount),
		Status:          string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetOpenItemByID(ctx, toUUID(row.ID))
}

func (r *PayableRepository) GetOpenItemByID(ctx context.Context, id uuid.UUID) (*domain.PayablesOpenItem, error) {
	q := r.queries(ctx)
	row, err := q.GetPayablesOpenItemByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOpenItemNotFound
		}
		return nil, err
	}
	return &domain.PayablesOpenItem{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		SupplierID:      toUUID(row.SupplierID),
		DocumentType:    domain.OpenItemDocumentType(row.DocumentType),
		DocumentNumber:  row.DocumentNumber,
		DocumentID:      toUUIDPtr(row.DocumentID),
		DocumentDate:    fromPgDate(row.DocumentDate),
		DueDate:         fromPgDate(row.DueDate),
		CurrencyCode:    row.CurrencyCode,
		OriginalAmount:  fromPgNumeric(row.OriginalAmount),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		RemainingAmount: fromPgNumeric(row.RemainingAmount),
		Status:          domain.OpenItemStatus(row.Status),
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:       fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *PayableRepository) LockOpenItem(ctx context.Context, id uuid.UUID) (*domain.PayablesOpenItem, error) {
	q := r.queries(ctx)
	row, err := q.LockPayablesOpenItem(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOpenItemNotFound
		}
		return nil, err
	}
	return &domain.PayablesOpenItem{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		SupplierID:      toUUID(row.SupplierID),
		DocumentType:    domain.OpenItemDocumentType(row.DocumentType),
		DocumentNumber:  row.DocumentNumber,
		DocumentID:      toUUIDPtr(row.DocumentID),
		DocumentDate:    fromPgDate(row.DocumentDate),
		DueDate:         fromPgDate(row.DueDate),
		CurrencyCode:    row.CurrencyCode,
		OriginalAmount:  fromPgNumeric(row.OriginalAmount),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		RemainingAmount: fromPgNumeric(row.RemainingAmount),
		Status:          domain.OpenItemStatus(row.Status),
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:       fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *PayableRepository) ListOpenItemsBySupplier(ctx context.Context, supplierID uuid.UUID, openOnly bool) ([]*domain.PayablesOpenItem, error) {
	q := r.queries(ctx)
	rows, err := q.ListOpenItemsBySupplier(ctx, sqlc.ListOpenItemsBySupplierParams{
		SupplierID: toPgUUID(supplierID),
		Column2:    openOnly,
	})
	if err != nil {
		return nil, err
	}
	var items []*domain.PayablesOpenItem
	for _, row := range rows {
		items = append(items, &domain.PayablesOpenItem{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			SupplierID:      toUUID(row.SupplierID),
			SupplierNumber:  row.SupplierNumber,
			SupplierNameAr:  row.SupplierNameAr,
			DocumentType:    domain.OpenItemDocumentType(row.DocumentType),
			DocumentNumber:  row.DocumentNumber,
			DocumentID:      toUUIDPtr(row.DocumentID),
			DocumentDate:    fromPgDate(row.DocumentDate),
			DueDate:         fromPgDate(row.DueDate),
			CurrencyCode:    row.CurrencyCode,
			OriginalAmount:  fromPgNumeric(row.OriginalAmount),
			AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
			RemainingAmount: fromPgNumeric(row.RemainingAmount),
			Status:          domain.OpenItemStatus(row.Status),
			DaysOverdue:     row.DaysOverdue,
			CreatedAt:       fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:       fromPgTimestamptz(row.UpdatedAt),
		})
	}
	return items, nil
}

func (r *PayableRepository) ListOpenItemsByLegalEntity(ctx context.Context, legalEntityID uuid.UUID, openOnly bool, limit, offset int32) ([]*domain.PayablesOpenItem, error) {
	q := r.queries(ctx)
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := q.ListOpenItemsByLegalEntityPayables(ctx, sqlc.ListOpenItemsByLegalEntityPayablesParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       openOnly,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	var items []*domain.PayablesOpenItem
	for _, row := range rows {
		items = append(items, &domain.PayablesOpenItem{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			SupplierID:      toUUID(row.SupplierID),
			SupplierNumber:  row.SupplierNumber,
			SupplierNameAr:  row.SupplierNameAr,
			DocumentType:    domain.OpenItemDocumentType(row.DocumentType),
			DocumentNumber:  row.DocumentNumber,
			DocumentID:      toUUIDPtr(row.DocumentID),
			DocumentDate:    fromPgDate(row.DocumentDate),
			DueDate:         fromPgDate(row.DueDate),
			CurrencyCode:    row.CurrencyCode,
			OriginalAmount:  fromPgNumeric(row.OriginalAmount),
			AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
			RemainingAmount: fromPgNumeric(row.RemainingAmount),
			Status:          domain.OpenItemStatus(row.Status),
			DaysOverdue:     row.DaysOverdue,
			CreatedAt:       fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:       fromPgTimestamptz(row.UpdatedAt),
		})
	}
	return items, nil
}

func (r *PayableRepository) UpdateOpenItemAllocation(ctx context.Context, id uuid.UUID, allocated, remaining decimal.Decimal, status domain.OpenItemStatus) (*domain.PayablesOpenItem, error) {
	q := r.queries(ctx)
	row, err := q.UpdatePayablesOpenItemAllocation(ctx, sqlc.UpdatePayablesOpenItemAllocationParams{
		ID:              toPgUUID(id),
		AllocatedAmount: toPgNumeric(allocated),
		RemainingAmount: toPgNumeric(remaining),
		Status:          string(status),
	})
	if err != nil {
		return nil, err
	}
	return &domain.PayablesOpenItem{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		SupplierID:      toUUID(row.SupplierID),
		DocumentType:    domain.OpenItemDocumentType(row.DocumentType),
		DocumentNumber:  row.DocumentNumber,
		DocumentID:      toUUIDPtr(row.DocumentID),
		DocumentDate:    fromPgDate(row.DocumentDate),
		DueDate:         fromPgDate(row.DueDate),
		CurrencyCode:    row.CurrencyCode,
		OriginalAmount:  fromPgNumeric(row.OriginalAmount),
		AllocatedAmount: fromPgNumeric(row.AllocatedAmount),
		RemainingAmount: fromPgNumeric(row.RemainingAmount),
		Status:          domain.OpenItemStatus(row.Status),
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:       fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

// Aging

func (r *PayableRepository) GetSupplierAgingSummary(ctx context.Context, supplierID uuid.UUID) (*domain.SupplierAgingSummary, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierAgingSummary(ctx, toPgUUID(supplierID))
	if err != nil {
		return nil, err
	}
	return &domain.SupplierAgingSummary{
		TotalOutstanding: fromPgNumeric(row.TotalOutstanding),
		CurrentAmount:    fromPgNumeric(row.CurrentAmount),
		Overdue1to30:     fromPgNumeric(row.Overdue130),
		Overdue31to60:    fromPgNumeric(row.Overdue3160),
		Overdue61to90:    fromPgNumeric(row.Overdue6190),
		OverdueOver90:    fromPgNumeric(row.OverdueOver90),
	}, nil
}
