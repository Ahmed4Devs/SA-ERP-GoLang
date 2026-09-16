package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"erp/internal/platform/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type payableService struct {
	supplierRepo ports.SupplierRepository
	payableRepo  ports.PayableRepository
	db           *database.DB
}

func NewPayableService(
	supplierRepo ports.SupplierRepository,
	payableRepo ports.PayableRepository,
	db *database.DB,
) ports.PayableService {
	return &payableService{
		supplierRepo: supplierRepo,
		payableRepo:  payableRepo,
		db:           db,
	}
}

// Supplier Invoices

func (s *payableService) CreateSupplierInvoice(ctx context.Context, cmd ports.CreateSupplierInvoiceCmd) (*domain.SupplierInvoice, error) {
	if len(cmd.Lines) == 0 {
		return nil, errors.New("supplier invoice must contain at least one line")
	}

	supp, err := s.supplierRepo.GetSupplierByID(ctx, cmd.SupplierID)
	if err != nil {
		return nil, err
	}
	if supp.Status == domain.SupplierStatusBlocked {
		return nil, domain.ErrSupplierBlocked
	}

	invDate := time.Now()
	if cmd.InvoiceDate != "" {
		if parsed, err := time.Parse("2006-01-02", cmd.InvoiceDate); err == nil {
			invDate = parsed
		}
	}

	dueDate := invDate.AddDate(0, 0, 30)
	if cmd.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", cmd.DueDate); err == nil {
			dueDate = parsed
		}
	}

	invNumber := cmd.InvoiceNumber
	if invNumber == "" {
		invNumber = fmt.Sprintf("PINV-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)
	}

	curr := cmd.CurrencyCode
	if curr == "" {
		curr = supp.CurrencyCode
		if curr == "" {
			curr = "SAR"
		}
	}

	rate := cmd.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	var subtotal, totalDiscount, taxableAmt, taxAmt, grandTotal decimal.Decimal
	var lines []domain.SupplierInvoiceLine

	for i, l := range cmd.Lines {
		if l.Quantity.LessThanOrEqual(decimal.Zero) {
			return nil, errors.New("line quantity must be greater than zero")
		}
		if l.UnitPrice.LessThan(decimal.Zero) {
			return nil, errors.New("line unit price cannot be negative")
		}

		lSub, lDisc, lTax, lTot := domain.CalculateLineAmounts(l.Quantity, l.UnitPrice, l.DiscountPercent, l.TaxRate)

		subtotal = subtotal.Add(lSub)
		totalDiscount = totalDiscount.Add(lDisc)
		taxableAmt = taxableAmt.Add(lSub)
		taxAmt = taxAmt.Add(lTax)
		grandTotal = grandTotal.Add(lTot)

		expAccID := l.ExpenseAccountID
		if expAccID == uuid.Nil && supp.DefaultExpenseAccountID != nil {
			expAccID = *supp.DefaultExpenseAccountID
		}

		lines = append(lines, domain.SupplierInvoiceLine{
			LineNumber:       int32(i + 1),
			ItemID:           l.ItemID,
			ExpenseAccountID: expAccID,
			Description:      l.Description,
			Quantity:         l.Quantity,
			UnitPrice:        l.UnitPrice,
			DiscountPercent:  l.DiscountPercent,
			DiscountAmount:   lDisc,
			TaxRate:          l.TaxRate,
			TaxAmount:        lTax,
			Subtotal:         lSub,
			TotalAmount:      lTot,
		})
	}

	invoice := &domain.SupplierInvoice{
		LegalEntityID:         cmd.LegalEntityID,
		InvoiceNumber:         invNumber,
		SupplierInvoiceNumber: cmd.SupplierInvoiceNumber,
		SupplierID:            cmd.SupplierID,
		InvoiceDate:           invDate,
		DueDate:               dueDate,
		CurrencyCode:          curr,
		ExchangeRate:          rate,
		Subtotal:              subtotal,
		DiscountAmount:        totalDiscount,
		TaxableAmount:         taxableAmt,
		TaxAmount:             taxAmt,
		TotalAmount:           grandTotal,
		Status:                domain.InvoiceStatusDraft,
		PaymentStatus:         domain.PaymentStatusUnpaid,
		Notes:                 cmd.Notes,
		CreatedByUserID:       cmd.CreatedByUserID,
		Lines:                 lines,
	}

	return s.payableRepo.CreateSupplierInvoice(ctx, invoice)
}

func (s *payableService) GetSupplierInvoice(ctx context.Context, id uuid.UUID) (*domain.SupplierInvoice, error) {
	return s.payableRepo.GetSupplierInvoiceByID(ctx, id)
}

func (s *payableService) ListSupplierInvoices(ctx context.Context, filter ports.SupplierInvoiceFilter) ([]*domain.SupplierInvoice, int64, error) {
	return s.payableRepo.ListSupplierInvoices(ctx, filter)
}

func (s *payableService) PostSupplierInvoice(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.SupplierInvoice, error) {
	inv, err := s.payableRepo.GetSupplierInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status == domain.InvoiceStatusPosted {
		return nil, domain.ErrInvoiceAlreadyPosted
	}

	// 1. Create AP Open Item in payables.open_items
	openItemID := uuid.New()
	_, err = s.payableRepo.CreateOpenItem(ctx, &domain.PayablesOpenItem{
		ID:              openItemID,
		LegalEntityID:   inv.LegalEntityID,
		SupplierID:      inv.SupplierID,
		DocumentType:    domain.DocumentTypeInvoice,
		DocumentNumber:  inv.InvoiceNumber,
		DocumentID:      &inv.ID,
		DocumentDate:    inv.InvoiceDate,
		DueDate:         inv.DueDate,
		CurrencyCode:    inv.CurrencyCode,
		OriginalAmount:  inv.TotalAmount,
		AllocatedAmount: decimal.Zero,
		RemainingAmount: inv.TotalAmount,
		Status:          domain.OpenItemStatusOpen,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AP open item: %w", err)
	}

	// 2. Post GL Entry (Double-Entry Bookkeeping):
	// Debits: Expense Accounts (line.ExpenseAccountID) for line subtotals
	// Debit: Input VAT (114000) for tax amount
	// Credit: Trade Payables AP (211000) for total amount
	var jeIDPtr *uuid.UUID
	var journalID, periodID, apAccountID, vatAccountID uuid.UUID

	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 AND code = 'PURCHASE' LIMIT 1", inv.LegalEntityID).Scan(&journalID)
	if journalID == uuid.Nil {
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 LIMIT 1", inv.LegalEntityID).Scan(&journalID)
	}
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", inv.LegalEntityID).Scan(&periodID)
	
	// Liability account: check supplier default or 211000
	if inv.SupplierLiabilityAccID != nil {
		apAccountID = *inv.SupplierLiabilityAccID
	} else {
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '211000' LIMIT 1", inv.LegalEntityID).Scan(&apAccountID)
	}
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '114000' LIMIT 1", inv.LegalEntityID).Scan(&vatAccountID)

	creator := userID
	if creator == nil {
		creator = inv.CreatedByUserID
	}
	if creator == nil {
		var adminID uuid.UUID
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID)
		creator = &adminID
	}

	if journalID != uuid.Nil && periodID != uuid.Nil && apAccountID != uuid.Nil {
		journalEntryID := uuid.New()
		entryNumber := fmt.Sprintf("JE-PINV-%d", time.Now().UnixNano()%1000000)

		_, err = s.db.Pool.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by, posted_by, posted_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'POSTED', 'PAYABLES', 'SUPPLIER_INVOICE', $7,
				$8, $9, $10, 1.0, $11, $11, $12, $12, now()
			)
		`, journalEntryID, inv.LegalEntityID, journalID, periodID, entryNumber, inv.InvoiceDate, inv.ID,
			inv.InvoiceNumber, "إثبات فاتورة مشتريات مورد - "+inv.SupplierInvoiceNumber, inv.CurrencyCode, inv.TotalAmount, creator)

		if err == nil {
			jeIDPtr = &journalEntryID
			lineNum := 1

			// Line 1..N: Debit Expense/Asset accounts
			for _, line := range inv.Lines {
				expAcc := line.ExpenseAccountID
				if expAcc == uuid.Nil {
					_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '520000' LIMIT 1", inv.LegalEntityID).Scan(&expAcc)
				}
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $7, 0)
				`, journalEntryID, inv.LegalEntityID, lineNum, expAcc, line.Description, inv.CurrencyCode, line.Subtotal)
				lineNum++
			}

			// Tax Line: Debit Input VAT (114000) if > 0
			if inv.TaxAmount.GreaterThan(decimal.Zero) && vatAccountID != uuid.Nil {
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, 'ضريبة القيمة المضافة على المدخلات', $5, $6, 0, $6, 0)
				`, journalEntryID, inv.LegalEntityID, lineNum, vatAccountID, inv.CurrencyCode, inv.TaxAmount)
				lineNum++
			}

			// AP Line: Credit Trade Payables (211000)
			_, _ = s.db.Pool.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, $3, $4, $5, $6, 0, $7, 0, $7)
			`, journalEntryID, inv.LegalEntityID, lineNum, apAccountID, "الموردون والدائنون التجاريون - "+inv.SupplierNameAr, inv.CurrencyCode, inv.TotalAmount)
		}
	}

	err = s.payableRepo.UpdateSupplierInvoiceStatus(ctx, id, domain.InvoiceStatusPosted, jeIDPtr)
	if err != nil {
		return nil, err
	}

	return s.payableRepo.GetSupplierInvoiceByID(ctx, id)
}

// Supplier Payments

func (s *payableService) CreateSupplierPayment(ctx context.Context, cmd ports.CreateSupplierPaymentCmd) (*domain.SupplierPayment, error) {
	if cmd.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("payment amount must be greater than zero")
	}

	supp, err := s.supplierRepo.GetSupplierByID(ctx, cmd.SupplierID)
	if err != nil {
		return nil, err
	}
	if supp.Status == domain.SupplierStatusBlocked {
		return nil, domain.ErrSupplierBlocked
	}

	payDate := time.Now()
	if cmd.PaymentDate != "" {
		if parsed, err := time.Parse("2006-01-02", cmd.PaymentDate); err == nil {
			payDate = parsed
		}
	}

	pNumber := cmd.PaymentNumber
	if pNumber == "" {
		pNumber = fmt.Sprintf("PAY-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)
	}

	curr := cmd.CurrencyCode
	if curr == "" {
		curr = supp.CurrencyCode
		if curr == "" {
			curr = "SAR"
		}
	}

	totalAlloc := decimal.Zero
	for _, a := range cmd.Allocations {
		if a.Amount.GreaterThan(decimal.Zero) {
			totalAlloc = totalAlloc.Add(a.Amount)
		}
	}

	if totalAlloc.GreaterThan(cmd.Amount) {
		return nil, domain.ErrOverAllocation
	}

	unalloc := cmd.Amount.Sub(totalAlloc)

	payment := &domain.SupplierPayment{
		LegalEntityID:         cmd.LegalEntityID,
		PaymentNumber:         pNumber,
		SupplierID:            cmd.SupplierID,
		PaymentDate:           payDate,
		PaymentMethod:         cmd.PaymentMethod,
		DisbursementAccountID: cmd.DisbursementAccountID,
		Amount:                cmd.Amount,
		CurrencyCode:          curr,
		ReferenceNumber:       cmd.ReferenceNumber,
		Status:                "DRAFT",
		Notes:                 cmd.Notes,
		AllocatedAmount:       totalAlloc,
		UnallocatedAmount:     unalloc,
		CreatedByUserID:       cmd.CreatedByUserID,
	}

	created, err := s.payableRepo.CreateSupplierPayment(ctx, payment)
	if err != nil {
		return nil, err
	}

	for _, a := range cmd.Allocations {
		if a.Amount.GreaterThan(decimal.Zero) {
			_, err = s.payableRepo.CreatePaymentAllocation(ctx, &domain.PaymentAllocation{
				PaymentID:       created.ID,
				OpenItemID:      a.OpenItemID,
				AllocatedAmount: a.Amount,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return s.payableRepo.GetSupplierPaymentByID(ctx, created.ID)
}

func (s *payableService) GetSupplierPayment(ctx context.Context, id uuid.UUID) (*domain.SupplierPayment, error) {
	return s.payableRepo.GetSupplierPaymentByID(ctx, id)
}

func (s *payableService) ListSupplierPayments(ctx context.Context, filter ports.SupplierPaymentFilter) ([]*domain.SupplierPayment, int64, error) {
	return s.payableRepo.ListSupplierPayments(ctx, filter)
}

func (s *payableService) PostSupplierPayment(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.SupplierPayment, error) {
	payment, err := s.payableRepo.GetSupplierPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if payment.Status == "POSTED" {
		return nil, domain.ErrPaymentAlreadyPosted
	}

	supp, err := s.supplierRepo.GetSupplierByID(ctx, payment.SupplierID)
	if err != nil {
		return nil, err
	}

	// 1. Process allocations on open items
	for _, alloc := range payment.Allocations {
		oi, err := s.payableRepo.LockOpenItem(ctx, alloc.OpenItemID)
		if err != nil {
			return nil, err
		}

		if alloc.AllocatedAmount.GreaterThan(oi.RemainingAmount) {
			return nil, domain.ErrOverAllocation
		}

		newAllocated := oi.AllocatedAmount.Add(alloc.AllocatedAmount)
		newRemaining := oi.RemainingAmount.Sub(alloc.AllocatedAmount)

		newStatus := domain.OpenItemStatusPartiallyPaid
		if newRemaining.IsZero() {
			newStatus = domain.OpenItemStatusPaid
		}

		_, err = s.payableRepo.UpdateOpenItemAllocation(ctx, oi.ID, newAllocated, newRemaining, newStatus)
		if err != nil {
			return nil, err
		}

		// Update linked supplier invoice payment status if document is an invoice
		if oi.DocumentType == domain.DocumentTypeInvoice && oi.DocumentID != nil {
			pStat := domain.PaymentStatusPartiallyPaid
			if newRemaining.IsZero() {
				pStat = domain.PaymentStatusPaid
			}
			_ = s.payableRepo.UpdateSupplierInvoicePaymentStatus(ctx, *oi.DocumentID, pStat)
		}
	}

	// 2. Double-Entry GL Journal Entry:
	// Debit: Trade Payables AP (211000) for payment amount
	// Credit: Disbursement Bank/Cash Account (payment.DisbursementAccountID) for payment amount
	var jeIDPtr *uuid.UUID
	var journalID, periodID, apAccountID uuid.UUID

	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 AND code = 'PAYMENT' LIMIT 1", payment.LegalEntityID).Scan(&journalID)
	if journalID == uuid.Nil {
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 LIMIT 1", payment.LegalEntityID).Scan(&journalID)
	}
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", payment.LegalEntityID).Scan(&periodID)

	if supp.DefaultLiabilityAccountID != nil {
		apAccountID = *supp.DefaultLiabilityAccountID
	} else {
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '211000' LIMIT 1", payment.LegalEntityID).Scan(&apAccountID)
	}

	creator := userID
	if creator == nil {
		creator = payment.CreatedByUserID
	}
	if creator == nil {
		var adminID uuid.UUID
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID)
		creator = &adminID
	}

	if journalID != uuid.Nil && periodID != uuid.Nil && apAccountID != uuid.Nil {
		journalEntryID := uuid.New()
		entryNumber := fmt.Sprintf("JE-PAY-%d", time.Now().UnixNano()%1000000)

		_, err = s.db.Pool.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by, posted_by, posted_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'POSTED', 'PAYABLES', 'PAYMENT', $7,
				$8, $9, $10, 1.0, $11, $11, $12, $12, now()
			)
		`, journalEntryID, payment.LegalEntityID, journalID, periodID, entryNumber, payment.PaymentDate, payment.ID,
			payment.PaymentNumber, "سند صرف مورد - "+payment.PaymentNumber, payment.CurrencyCode, payment.Amount, creator)

		if err == nil {
			jeIDPtr = &journalEntryID
			// Line 1: Debit AP (211000)
			_, _ = s.db.Pool.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 1, $3, $4, $5, $6, 0, $6, 0)
			`, journalEntryID, payment.LegalEntityID, apAccountID, "سداد للمورد - "+supp.NameAr, payment.CurrencyCode, payment.Amount)

			// Line 2: Credit Disbursement Bank/Cash Account
			_, _ = s.db.Pool.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 2, $3, 'حساب الصرف / البنك', $4, 0, $5, 0, $5)
			`, journalEntryID, payment.LegalEntityID, payment.DisbursementAccountID, payment.CurrencyCode, payment.Amount)
		}
	}

	err = s.payableRepo.UpdateSupplierPaymentStatus(ctx, id, "POSTED", payment.AllocatedAmount, payment.UnallocatedAmount, jeIDPtr)
	if err != nil {
		return nil, err
	}

	return s.payableRepo.GetSupplierPaymentByID(ctx, id)
}

// Supplier Credit Notes

func (s *payableService) CreateSupplierCreditNote(ctx context.Context, cmd ports.CreateCreditNoteCmd) (*domain.SupplierCreditNote, error) {
	if len(cmd.Lines) == 0 {
		return nil, errors.New("credit note must contain at least one line")
	}

	supp, err := s.supplierRepo.GetSupplierByID(ctx, cmd.SupplierID)
	if err != nil {
		return nil, err
	}

	cnDate := time.Now()
	if cmd.CreditNoteDate != "" {
		if parsed, err := time.Parse("2006-01-02", cmd.CreditNoteDate); err == nil {
			cnDate = parsed
		}
	}

	cnNum := cmd.CreditNoteNumber
	if cnNum == "" {
		cnNum = fmt.Sprintf("PCN-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)
	}

	curr := cmd.CurrencyCode
	if curr == "" {
		curr = supp.CurrencyCode
		if curr == "" {
			curr = "SAR"
		}
	}

	var subtotal, taxAmt, totalAmt decimal.Decimal
	var lines []domain.SupplierCreditNoteLine

	for i, l := range cmd.Lines {
		lSub := l.Quantity.Mul(l.UnitPrice).Round(4)
		lTax := lSub.Mul(l.TaxRate).Div(decimal.NewFromInt(100)).Round(4)
		lTot := lSub.Add(lTax).Round(4)

		subtotal = subtotal.Add(lSub)
		taxAmt = taxAmt.Add(lTax)
		totalAmt = totalAmt.Add(lTot)

		expAcc := l.ExpenseAccountID
		if expAcc == uuid.Nil && supp.DefaultExpenseAccountID != nil {
			expAcc = *supp.DefaultExpenseAccountID
		}

		lines = append(lines, domain.SupplierCreditNoteLine{
			LineNumber:       int32(i + 1),
			ItemID:           l.ItemID,
			ExpenseAccountID: expAcc,
			Description:      l.Description,
			Quantity:         l.Quantity,
			UnitPrice:        l.UnitPrice,
			TaxRate:          l.TaxRate,
			TaxAmount:        lTax,
			Subtotal:         lSub,
			TotalAmount:      lTot,
		})
	}

	cn := &domain.SupplierCreditNote{
		LegalEntityID:            cmd.LegalEntityID,
		CreditNoteNumber:         cnNum,
		SupplierCreditNoteNumber: cmd.SupplierCreditNoteNumber,
		SupplierID:               cmd.SupplierID,
		OriginalInvoiceID:        cmd.OriginalInvoiceID,
		CreditNoteDate:           cnDate,
		CurrencyCode:             curr,
		Subtotal:                 subtotal,
		TaxAmount:                taxAmt,
		TotalAmount:              totalAmt,
		Reason:                   cmd.Reason,
		Status:                   "POSTED",
		CreatedByUserID:          cmd.CreatedByUserID,
		Lines:                    lines,
	}

	created, err := s.payableRepo.CreateSupplierCreditNote(ctx, cn)
	if err != nil {
		return nil, err
	}

	// Double-entry GL for Credit Note:
	// Debit: Trade Payables AP (211000) for total amount (reversal)
	// Credit: Expense Account for subtotal
	// Credit: Input VAT Recoverable (114000) for tax amount
	var journalID, periodID, apAccountID, vatAccountID uuid.UUID
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 LIMIT 1", cmd.LegalEntityID).Scan(&journalID)
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", cmd.LegalEntityID).Scan(&periodID)
	if supp.DefaultLiabilityAccountID != nil {
		apAccountID = *supp.DefaultLiabilityAccountID
	} else {
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '211000' LIMIT 1", cmd.LegalEntityID).Scan(&apAccountID)
	}
	_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '114000' LIMIT 1", cmd.LegalEntityID).Scan(&vatAccountID)

	creator := cmd.CreatedByUserID
	if creator == nil {
		var adminID uuid.UUID
		_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID)
		creator = &adminID
	}

	if journalID != uuid.Nil && periodID != uuid.Nil && apAccountID != uuid.Nil {
		journalEntryID := uuid.New()
		entryNumber := fmt.Sprintf("JE-PCN-%d", time.Now().UnixNano()%1000000)

		_, err = s.db.Pool.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by, posted_by, posted_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'POSTED', 'PAYABLES', 'CREDIT_NOTE', $7,
				$8, $9, $10, 1.0, $11, $11, $12, $12, now()
			)
		`, journalEntryID, cmd.LegalEntityID, journalID, periodID, entryNumber, cnDate, created.ID,
			cnNum, "إشعار خصم دائن من مورد - "+cnNum, curr, totalAmt, creator)

		if err == nil {
			lineNum := 1
			// Line 1: Debit AP (211000)
			_, _ = s.db.Pool.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 1, $3, $4, $5, $6, 0, $6, 0)
			`, journalEntryID, cmd.LegalEntityID, apAccountID, "تخفيض حساب المورد - "+supp.NameAr, curr, totalAmt)
			lineNum++

			// Line 2..N: Credit Expense accounts
			for _, line := range lines {
				expAcc := line.ExpenseAccountID
				if expAcc == uuid.Nil {
					_ = s.db.Pool.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '520000' LIMIT 1", cmd.LegalEntityID).Scan(&expAcc)
				}
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, $5, $6, 0, $7, 0, $7)
				`, journalEntryID, cmd.LegalEntityID, lineNum, expAcc, line.Description, curr, line.Subtotal)
				lineNum++
			}

			// Credit Input VAT (114000) if > 0
			if taxAmt.GreaterThan(decimal.Zero) && vatAccountID != uuid.Nil {
				_, _ = s.db.Pool.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, 'عكس ضريبة القيمة المضافة على المدخلات', $5, 0, $6, 0, $6)
				`, journalEntryID, cmd.LegalEntityID, lineNum, vatAccountID, curr, taxAmt)
			}
		}
	}

	return s.payableRepo.GetSupplierCreditNoteByID(ctx, created.ID)
}

func (s *payableService) GetSupplierCreditNote(ctx context.Context, id uuid.UUID) (*domain.SupplierCreditNote, error) {
	return s.payableRepo.GetSupplierCreditNoteByID(ctx, id)
}

func (s *payableService) ListSupplierCreditNotes(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID) ([]*domain.SupplierCreditNote, error) {
	return s.payableRepo.ListSupplierCreditNotes(ctx, legalEntityID, supplierID)
}

// Open Items & Aging

func (s *payableService) ListOpenItems(ctx context.Context, supplierID uuid.UUID, openOnly bool) ([]*domain.PayablesOpenItem, error) {
	return s.payableRepo.ListOpenItemsBySupplier(ctx, supplierID, openOnly)
}

func (s *payableService) GetSupplierAgingSummary(ctx context.Context, supplierID uuid.UUID) (*domain.SupplierAgingSummary, error) {
	return s.payableRepo.GetSupplierAgingSummary(ctx, supplierID)
}
