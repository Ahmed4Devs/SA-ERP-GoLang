package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	postingapp "erp/internal/accounting/application/posting"
	accountingdomain "erp/internal/accounting/domain"
	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type payableService struct {
	ledger       *postingapp.SourceService
	supplierRepo ports.SupplierRepository
	payableRepo  ports.PayableRepository
	db           *database.DB
	tx           transaction.Manager
}

func NewPayableService(
	supplierRepo ports.SupplierRepository,
	payableRepo ports.PayableRepository,
	db *database.DB,
	ledger *postingapp.SourceService,
) ports.PayableService {
	return &payableService{
		ledger:       ledger,
		supplierRepo: supplierRepo,
		payableRepo:  payableRepo,
		db:           db,
		tx:           transaction.NewManager(db.Pool),
	}
}

// Supplier Invoices

func (s *payableService) CreateSupplierInvoice(ctx context.Context, cmd ports.CreateSupplierInvoiceCmd) (*domain.SupplierInvoice, error) {
	var invoice *domain.SupplierInvoice
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		invoice, err = s.createSupplierInvoice(txCtx, cmd)
		return err
	})
	return invoice, err
}

func (s *payableService) createSupplierInvoice(ctx context.Context, cmd ports.CreateSupplierInvoiceCmd) (*domain.SupplierInvoice, error) {
	if len(cmd.Lines) == 0 {
		return nil, errors.New("supplier invoice must contain at least one line")
	}

	supp, err := s.supplierRepo.GetSupplierByID(ctx, cmd.SupplierID)
	if err != nil {
		return nil, err
	}
	if supp.LegalEntityID != cmd.LegalEntityID {
		return nil, errors.New("supplier legal entity mismatch")
	}
	if supp.Status == domain.SupplierStatusBlocked {
		return nil, domain.ErrSupplierBlocked
	}

	invDate := time.Now()
	if cmd.InvoiceDate != "" {
		invDate, err = time.Parse("2006-01-02", cmd.InvoiceDate)
		if err != nil {
			return nil, errors.New("invalid invoice date")
		}
	}

	dueDate := invDate.AddDate(0, 0, 30)
	if cmd.DueDate != "" {
		dueDate, err = time.Parse("2006-01-02", cmd.DueDate)
		if err != nil {
			return nil, errors.New("invalid due date")
		}
	}

	invNumber := cmd.InvoiceNumber
	if invNumber == "" {
		invNumber = identifiers.NewDocumentNumber("PINV", time.Now())
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
		var base string
		if err := s.db.DBTX(ctx).QueryRow(ctx, "SELECT base_currency FROM organization.legal_entities WHERE id=$1", cmd.LegalEntityID).Scan(&base); err != nil {
			return nil, err
		}
		if curr != base {
			return nil, errors.New("foreign-currency invoice requires an exchange rate")
		}
		rate = decimal.NewFromInt(1)
	}
	if !rate.IsPositive() {
		return nil, errors.New("exchange rate must be positive")
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
		if l.DiscountPercent.IsNegative() || l.DiscountPercent.GreaterThan(decimal.NewFromInt(100)) {
			return nil, errors.New("discount percent must be between zero and 100")
		}

		taxCode, taxRate, err := s.resolveInputTax(ctx, cmd.LegalEntityID, l.TaxCodeID, invDate, l.TaxRate)
		if err != nil {
			return nil, err
		}
		l.TaxCodeID = &taxCode
		l.TaxRate = taxRate
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
			TaxCodeID:        l.TaxCodeID,
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
	if userID == nil || *userID == uuid.Nil {
		return nil, errors.New("posting actor is required")
	}
	var invoice *domain.SupplierInvoice
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		invoice, err = s.postSupplierInvoice(txCtx, id, userID)
		return err
	})
	return invoice, err
}

func (s *payableService) postSupplierInvoice(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.SupplierInvoice, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM payables.supplier_invoices WHERE id = $1 FOR UPDATE", id).Scan(&id); err != nil {
		return nil, fmt.Errorf("failed to lock supplier invoice: %w", err)
	}
	inv, err := s.payableRepo.GetSupplierInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status == domain.InvoiceStatusPosted {
		return nil, domain.ErrInvoiceAlreadyPosted
	}
	if inv.Status != domain.InvoiceStatusDraft {
		return nil, errors.New("only draft supplier invoices can be posted")
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

	apID, err := s.ledger.AccountForDate(ctx, inv.LegalEntityID, inv.SupplierLiabilityAccID, "PAYABLES", "AP_CONTROL", "211000", inv.InvoiceDate)
	if err != nil {
		return nil, err
	}
	lines := make([]accountingdomain.PostingLine, 0, len(inv.Lines)+2)
	for _, line := range inv.Lines {
		account, err := s.ledger.AccountForDate(ctx, inv.LegalEntityID, &line.ExpenseAccountID, "PAYABLES", "PURCHASE_EXPENSE", "523000", inv.InvoiceDate)
		if err != nil {
			return nil, err
		}
		if line.Subtotal.IsPositive() {
			lines = append(lines, accountingdomain.PostingLine{AccountID: account, TransactionDebit: line.Subtotal, Description: line.Description})
		}
	}
	if inv.TaxAmount.IsPositive() {
		vat, err := s.ledger.AccountForDate(ctx, inv.LegalEntityID, nil, "PAYABLES", "INPUT_VAT", "114000", inv.InvoiceDate)
		if err != nil {
			return nil, err
		}
		lines = append(lines, accountingdomain.PostingLine{AccountID: vat, TransactionDebit: inv.TaxAmount})
	}
	lines = append(lines, accountingdomain.PostingLine{AccountID: apID, TransactionCredit: inv.TotalAmount})
	journalID, err := s.ledger.Post(ctx, "PURCHASE", accountingdomain.PostingRequest{LegalEntityID: inv.LegalEntityID, SourceModule: "PAYABLES", SourceType: "SUPPLIER_INVOICE", SourceID: inv.ID, AccountingDate: inv.InvoiceDate, CurrencyCode: inv.CurrencyCode, ExchangeRate: inv.ExchangeRate, Reference: inv.InvoiceNumber, ActorID: *userID, Lines: lines})
	if err != nil {
		return nil, err
	}
	if err := s.postInputTax(ctx, inv, journalID, *userID); err != nil {
		return nil, err
	}
	jeIDPtr := &journalID

	err = s.payableRepo.UpdateSupplierInvoiceStatus(ctx, id, domain.InvoiceStatusPosted, jeIDPtr)
	if err != nil {
		return nil, err
	}

	return s.payableRepo.GetSupplierInvoiceByID(ctx, id)
}

// Supplier Payments

func (s *payableService) CreateSupplierPayment(ctx context.Context, cmd ports.CreateSupplierPaymentCmd) (*domain.SupplierPayment, error) {
	var payment *domain.SupplierPayment
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		payment, err = s.createSupplierPayment(txCtx, cmd)
		return err
	})
	return payment, err
}

func (s *payableService) createSupplierPayment(ctx context.Context, cmd ports.CreateSupplierPaymentCmd) (*domain.SupplierPayment, error) {
	if cmd.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("payment amount must be greater than zero")
	}

	supp, err := s.supplierRepo.GetSupplierByID(ctx, cmd.SupplierID)
	if err != nil {
		return nil, err
	}
	if supp.LegalEntityID != cmd.LegalEntityID {
		return nil, errors.New("supplier does not belong to payment legal entity")
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
		pNumber = identifiers.NewDocumentNumber("PAY", time.Now())
	}

	curr := cmd.CurrencyCode
	if curr == "" {
		curr = supp.CurrencyCode
		if curr == "" {
			curr = "SAR"
		}
	}

	totalAlloc := decimal.Zero
	cmd.Allocations = append([]ports.PaymentAllocationCmd(nil), cmd.Allocations...)
	seen := make(map[uuid.UUID]bool)
	for _, allocation := range cmd.Allocations {
		if allocation.OpenItemID == uuid.Nil || !allocation.Amount.IsPositive() {
			return nil, errors.New("allocation requires an invoice and a positive amount")
		}
		if seen[allocation.OpenItemID] {
			return nil, errors.New("duplicate invoice allocation")
		}
		seen[allocation.OpenItemID] = true
		totalAlloc = totalAlloc.Add(allocation.Amount)
	}
	sort.Slice(cmd.Allocations, func(i, j int) bool {
		return cmd.Allocations[i].OpenItemID.String() < cmd.Allocations[j].OpenItemID.String()
	})
	for _, allocation := range cmd.Allocations {
		oi, err := s.payableRepo.LockOpenItem(ctx, allocation.OpenItemID)
		if err != nil {
			return nil, err
		}
		if oi.LegalEntityID != cmd.LegalEntityID || oi.SupplierID != cmd.SupplierID || oi.CurrencyCode != curr {
			return nil, errors.New("allocation must match payment legal entity, supplier and currency")
		}
		if allocation.Amount.GreaterThan(oi.RemainingAmount) {
			return nil, domain.ErrOverAllocation
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
	if userID == nil || *userID == uuid.Nil {
		return nil, errors.New("posting actor is required")
	}
	var payment *domain.SupplierPayment
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		payment, err = s.postSupplierPayment(txCtx, id, userID)
		return err
	})
	return payment, err
}

func (s *payableService) postSupplierPayment(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*domain.SupplierPayment, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM payables.payments WHERE id = $1 FOR UPDATE", id).Scan(&id); err != nil {
		return nil, fmt.Errorf("failed to lock supplier payment: %w", err)
	}
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

	sort.Slice(payment.Allocations, func(i, j int) bool {
		return payment.Allocations[i].OpenItemID.String() < payment.Allocations[j].OpenItemID.String()
	})
	// 1. Process allocations on open items
	for _, alloc := range payment.Allocations {
		oi, err := s.payableRepo.LockOpenItem(ctx, alloc.OpenItemID)
		if err != nil {
			return nil, err
		}

		if !alloc.AllocatedAmount.IsPositive() || oi.LegalEntityID != payment.LegalEntityID || oi.SupplierID != payment.SupplierID || oi.CurrencyCode != payment.CurrencyCode {
			return nil, errors.New("allocation must be positive and match payment legal entity, supplier and currency")
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
			if err := s.payableRepo.UpdateSupplierInvoicePaymentStatus(ctx, *oi.DocumentID, pStat); err != nil {
				return nil, fmt.Errorf("failed to update supplier invoice payment status: %w", err)
			}
		}
	}

	apID, err := s.ledger.AccountForDate(ctx, payment.LegalEntityID, supp.DefaultLiabilityAccountID, "PAYABLES", "AP_CONTROL", "211000", payment.PaymentDate)
	if err != nil {
		return nil, err
	}
	journalID, err := s.ledger.Post(ctx, "BANK", accountingdomain.PostingRequest{LegalEntityID: payment.LegalEntityID, SourceModule: "PAYABLES", SourceType: "PAYMENT", SourceID: payment.ID, AccountingDate: payment.PaymentDate, CurrencyCode: payment.CurrencyCode, Reference: payment.PaymentNumber, ActorID: *userID, Lines: []accountingdomain.PostingLine{
		{AccountID: apID, TransactionDebit: payment.Amount}, {AccountID: payment.DisbursementAccountID, TransactionCredit: payment.Amount},
	}})
	if err != nil {
		return nil, err
	}
	jeIDPtr := &journalID

	err = s.payableRepo.UpdateSupplierPaymentStatus(ctx, id, "POSTED", payment.AllocatedAmount, payment.UnallocatedAmount, jeIDPtr)
	if err != nil {
		return nil, err
	}

	return s.payableRepo.GetSupplierPaymentByID(ctx, id)
}

// Supplier Credit Notes

func (s *payableService) CreateSupplierCreditNote(ctx context.Context, cmd ports.CreateCreditNoteCmd) (*domain.SupplierCreditNote, error) {
	var creditNote *domain.SupplierCreditNote
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		creditNote, err = s.createSupplierCreditNote(txCtx, cmd)
		return err
	})
	return creditNote, err
}

func (s *payableService) createSupplierCreditNote(ctx context.Context, cmd ports.CreateCreditNoteCmd) (*domain.SupplierCreditNote, error) {
	dbtx := s.db.DBTX(ctx)
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
		cnNum = identifiers.NewDocumentNumber("PCN", time.Now())
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
		lSub := l.Quantity.Mul(l.UnitPrice).Round(6)
		lTax := lSub.Mul(l.TaxRate).Div(decimal.NewFromInt(100)).Round(6)
		lTot := lSub.Add(lTax).Round(6)

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
		Status:                   "DRAFT",
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
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 ORDER BY (code = 'PURCHASE') DESC LIMIT 1", cmd.LegalEntityID).Scan(&journalID); err != nil {
		return nil, fmt.Errorf("purchase journal is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", cmd.LegalEntityID).Scan(&periodID); err != nil {
		return nil, fmt.Errorf("open fiscal period is not configured: %w", err)
	}
	if supp.DefaultLiabilityAccountID != nil {
		apAccountID = *supp.DefaultLiabilityAccountID
	} else {
		if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '211000' LIMIT 1", cmd.LegalEntityID).Scan(&apAccountID); err != nil {
			return nil, fmt.Errorf("trade payables account 211000 is not configured: %w", err)
		}
	}
	if taxAmt.GreaterThan(decimal.Zero) {
		if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '114000' LIMIT 1", cmd.LegalEntityID).Scan(&vatAccountID); err != nil {
			return nil, fmt.Errorf("input VAT account 114000 is not configured: %w", err)
		}
	}

	creator := cmd.CreatedByUserID
	if creator == nil {
		var adminID uuid.UUID
		if err := dbtx.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID); err != nil {
			return nil, fmt.Errorf("failed to resolve credit-note creator: %w", err)
		}
		creator = &adminID
	}

	journalEntryID := uuid.New()
	entryNumber := identifiers.NewDocumentNumber("JE-PCN", time.Now())

	_, err = dbtx.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by, posted_by, posted_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'DRAFT', 'PAYABLES', 'CREDIT_NOTE', $7,
				$8, $9, $10, 1.0, $11, $11, $12, $12, now()
			)
	`, journalEntryID, cmd.LegalEntityID, journalID, periodID, entryNumber, cnDate, created.ID,
		cnNum, "إشعار خصم دائن من مورد - "+cnNum, curr, totalAmt, creator)

	if err != nil {
		return nil, fmt.Errorf("failed to create supplier credit-note journal entry: %w", err)
	}
	{
		lineNum := 1
		// Line 1: Debit AP (211000)
		if _, err = dbtx.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 1, $3, $4, $5, $6, 0, $6, 0)
		`, journalEntryID, cmd.LegalEntityID, apAccountID, "تخفيض حساب المورد - "+supp.NameAr, curr, totalAmt); err != nil {
			return nil, fmt.Errorf("failed to create payables reversal journal line: %w", err)
		}
		lineNum++

		// Line 2..N: Credit Expense accounts
		for _, line := range lines {
			expAcc := line.ExpenseAccountID
			if expAcc == uuid.Nil {
				if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '523000' AND is_active AND allow_posting LIMIT 1", cmd.LegalEntityID).Scan(&expAcc); err != nil {
					return nil, fmt.Errorf("posting expense account 523000 is not configured: %w", err)
				}
			}
			if _, err = dbtx.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, $5, $6, 0, $7, 0, $7)
				`, journalEntryID, cmd.LegalEntityID, lineNum, expAcc, line.Description, curr, line.Subtotal); err != nil {
				return nil, fmt.Errorf("failed to create expense reversal journal line: %w", err)
			}
			lineNum++
		}

		// Credit Input VAT (114000) if > 0
		if taxAmt.GreaterThan(decimal.Zero) && vatAccountID != uuid.Nil {
			if _, err = dbtx.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, $3, $4, 'عكس ضريبة القيمة المضافة على المدخلات', $5, 0, $6, 0, $6)
				`, journalEntryID, cmd.LegalEntityID, lineNum, vatAccountID, curr, taxAmt); err != nil {
				return nil, fmt.Errorf("failed to create VAT reversal journal line: %w", err)
			}
		}
		if _, err = dbtx.Exec(ctx, "UPDATE accounting.journal_entries SET status = 'POSTED', posted_by = $1, posted_at = now() WHERE id = $2", creator, journalEntryID); err != nil {
			return nil, fmt.Errorf("failed to post supplier credit-note journal entry: %w", err)
		}
	}
	if _, err = dbtx.Exec(ctx, `
		UPDATE payables.supplier_credit_notes
		SET status = 'POSTED', journal_entry_id = $2
		WHERE id = $1
	`, created.ID, journalEntryID); err != nil {
		return nil, fmt.Errorf("failed to finalize supplier credit note: %w", err)
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
