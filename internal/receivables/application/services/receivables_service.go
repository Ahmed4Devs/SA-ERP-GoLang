package services

import (
	"context"
	postingapp "erp/internal/accounting/application/posting"
	accountingdomain "erp/internal/accounting/domain"
	"fmt"
	"sort"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"
	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type receivablesService struct {
	ledger          *postingapp.SourceService
	customerRepo    ports.CustomerRepository
	receivablesRepo ports.ReceivablesRepository
	db              *database.DB
	tx              transaction.Manager
}

func NewReceivablesService(
	customerRepo ports.CustomerRepository,
	receivablesRepo ports.ReceivablesRepository,
	db *database.DB,
	ledger *postingapp.SourceService,
) ports.ReceivablesService {
	return &receivablesService{
		ledger:          ledger,
		customerRepo:    customerRepo,
		receivablesRepo: receivablesRepo,
		db:              db,
		tx:              transaction.NewManager(db.Pool),
	}
}

func (s *receivablesService) CreateReceipt(ctx context.Context, cmd ports.CreateReceiptCommand) (*domain.Receipt, error) {
	var receipt *domain.Receipt
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		receipt, err = s.createReceipt(txCtx, cmd)
		return err
	})
	return receipt, err
}

func (s *receivablesService) createReceipt(ctx context.Context, cmd ports.CreateReceiptCommand) (*domain.Receipt, error) {
	if cmd.CreatedByUserID == nil || *cmd.CreatedByUserID == uuid.Nil {
		return nil, fmt.Errorf("receipt actor is required")
	}
	if cmd.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("receipt amount must be greater than zero")
	}

	customer, err := s.customerRepo.GetCustomerByID(ctx, cmd.CustomerID)
	if err != nil {
		return nil, err
	}
	if customer.LegalEntityID != cmd.LegalEntityID {
		return nil, fmt.Errorf("customer does not belong to receipt legal entity")
	}
	if customer.Status == domain.CustomerStatusBlocked {
		return nil, fmt.Errorf("cannot issue receipt for blocked customer: %s", customer.CustomerNumber)
	}

	var receiptDate time.Time
	if cmd.ReceiptDate != "" {
		parsed, err := time.Parse("2006-01-02", cmd.ReceiptDate)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, cmd.ReceiptDate)
		}
		if err == nil {
			receiptDate = parsed
		}
	}
	if receiptDate.IsZero() {
		receiptDate = time.Now()
	}

	receiptNumber := identifiers.NewDocumentNumber("REC", time.Now())

	ordered, totalAllocated, err := validateAllocations(cmd.Allocations)
	if err != nil {
		return nil, err
	}
	cmd.Allocations = ordered

	if totalAllocated.GreaterThan(cmd.Amount) {
		return nil, domain.ErrOverAllocation
	}

	unallocated := cmd.Amount.Sub(totalAllocated)

	currencyCode := cmd.CurrencyCode
	if currencyCode == "" {
		currencyCode = customer.CurrencyCode
		if currencyCode == "" {
			currencyCode = "SAR"
		}
	}

	receipt := &domain.Receipt{
		LegalEntityID:     cmd.LegalEntityID,
		ReceiptNumber:     receiptNumber,
		CustomerID:        cmd.CustomerID,
		ReceiptDate:       receiptDate,
		PaymentMethod:     cmd.PaymentMethod,
		DepositAccountID:  cmd.DepositAccountID,
		Amount:            cmd.Amount,
		CurrencyCode:      currencyCode,
		ReferenceNumber:   cmd.ReferenceNumber,
		Status:            domain.ReceiptStatusPosted,
		Notes:             cmd.Notes,
		AllocatedAmount:   totalAllocated,
		UnallocatedAmount: unallocated,
		CreatedByUserID:   cmd.CreatedByUserID,
	}

	createdReceipt, err := s.receivablesRepo.CreateReceipt(ctx, receipt)
	if err != nil {
		return nil, err
	}

	// Process allocations
	for _, alloc := range cmd.Allocations {
		if alloc.AllocatedAmount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		openItem, err := s.receivablesRepo.LockOpenItem(ctx, alloc.OpenItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to lock open item %s: %w", alloc.OpenItemID, err)
		}

		if openItem.LegalEntityID != cmd.LegalEntityID || openItem.CustomerID != cmd.CustomerID || openItem.CurrencyCode != currencyCode {
			return nil, fmt.Errorf("allocation must match receipt legal entity, customer and currency")
		}
		if openItem.Status != domain.OpenItemStatusOpen && openItem.Status != domain.OpenItemStatusPartiallyPaid {
			return nil, fmt.Errorf("allocation requires an open invoice")
		}
		if alloc.AllocatedAmount.GreaterThan(openItem.RemainingAmount) {
			return nil, fmt.Errorf("allocation amount %s exceeds remaining balance %s for invoice %s: %w",
				alloc.AllocatedAmount, openItem.RemainingAmount, openItem.DocumentNumber, domain.ErrOverAllocation)
		}

		newRemaining := openItem.RemainingAmount.Sub(alloc.AllocatedAmount)
		newAllocated := openItem.AllocatedAmount.Add(alloc.AllocatedAmount)
		newStatus := domain.OpenItemStatusPartiallyPaid
		if newRemaining.IsZero() {
			newStatus = domain.OpenItemStatusPaid
		}

		_, err = s.receivablesRepo.UpdateOpenItemAllocation(ctx, openItem.ID, newAllocated, newRemaining, newStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to update open item: %w", err)
		}

		_, err = s.receivablesRepo.CreateReceiptAllocation(ctx, &domain.ReceiptAllocation{
			ReceiptID:       createdReceipt.ID,
			OpenItemID:      openItem.ID,
			AllocatedAmount: alloc.AllocatedAmount,
			CreatedByUserID: cmd.CreatedByUserID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to record receipt allocation: %w", err)
		}
	}

	// Reduce customer current balance if allocations occurred
	if totalAllocated.GreaterThan(decimal.Zero) {
		if _, err := s.customerRepo.UpdateCustomerBalance(ctx, cmd.CustomerID, totalAllocated.Neg()); err != nil {
			return nil, fmt.Errorf("failed to update customer balance: %w", err)
		}
	}

	arID, err := s.ledger.Account(ctx, cmd.LegalEntityID, customer.DefaultAccountID, "112000")
	if err != nil {
		return nil, err
	}
	journalID, err := s.ledger.Post(ctx, "BANK", accountingdomain.PostingRequest{LegalEntityID: cmd.LegalEntityID, SourceModule: "RECEIVABLES", SourceType: "RECEIPT", SourceID: createdReceipt.ID, AccountingDate: receiptDate, CurrencyCode: currencyCode, Reference: receiptNumber, ActorID: *cmd.CreatedByUserID, Lines: []accountingdomain.PostingLine{
		{AccountID: cmd.DepositAccountID, TransactionDebit: cmd.Amount}, {AccountID: arID, TransactionCredit: cmd.Amount},
	}})
	if err != nil {
		return nil, err
	}
	if _, err = s.db.DBTX(ctx).Exec(ctx, "UPDATE receivables.receipts SET journal_entry_id=$1 WHERE id=$2", journalID, createdReceipt.ID); err != nil {
		return nil, err
	}
	return s.receivablesRepo.GetReceiptByID(ctx, createdReceipt.ID)
}

func (s *receivablesService) GetReceipt(ctx context.Context, id uuid.UUID) (*domain.Receipt, error) {
	return s.receivablesRepo.GetReceiptByID(ctx, id)
}

func (s *receivablesService) ListReceipts(ctx context.Context, filter ports.ReceiptFilter) ([]*domain.Receipt, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	if filter.PaymentMethod == "" {
		filter.PaymentMethod = "ALL"
	}
	return s.receivablesRepo.ListReceipts(ctx, filter)
}

func (s *receivablesService) AllocateReceipt(ctx context.Context, receiptID uuid.UUID, allocations []ports.AllocationCommand, userID *uuid.UUID) error {
	return s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		return s.allocateReceipt(txCtx, receiptID, allocations, userID)
	})
}

func (s *receivablesService) allocateReceipt(ctx context.Context, receiptID uuid.UUID, allocations []ports.AllocationCommand, userID *uuid.UUID) error {
	if err := s.db.DBTX(ctx).QueryRow(ctx, "SELECT id FROM receivables.receipts WHERE id = $1 FOR UPDATE", receiptID).Scan(&receiptID); err != nil {
		return fmt.Errorf("failed to lock receipt: %w", err)
	}
	receipt, err := s.receivablesRepo.GetReceiptByID(ctx, receiptID)
	if err != nil {
		return err
	}

	if receipt.Status != domain.ReceiptStatusPosted {
		return fmt.Errorf("only posted receipts may be allocated")
	}
	ordered, totalToAllocate, err := validateAllocations(allocations)
	if err != nil {
		return err
	}
	allocations = ordered

	if totalToAllocate.GreaterThan(receipt.UnallocatedAmount) {
		return fmt.Errorf("allocation exceeds receipt unallocated amount %s: %w", receipt.UnallocatedAmount, domain.ErrOverAllocation)
	}

	for _, alloc := range allocations {
		if alloc.AllocatedAmount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		openItem, err := s.receivablesRepo.LockOpenItem(ctx, alloc.OpenItemID)
		if err != nil {
			return err
		}

		if openItem.LegalEntityID != receipt.LegalEntityID || openItem.CustomerID != receipt.CustomerID || openItem.CurrencyCode != receipt.CurrencyCode {
			return fmt.Errorf("allocation must match receipt legal entity, customer and currency")
		}
		if openItem.Status != domain.OpenItemStatusOpen && openItem.Status != domain.OpenItemStatusPartiallyPaid {
			return fmt.Errorf("allocation requires an open invoice")
		}
		if alloc.AllocatedAmount.GreaterThan(openItem.RemainingAmount) {
			return domain.ErrOverAllocation
		}

		newRemaining := openItem.RemainingAmount.Sub(alloc.AllocatedAmount)
		newAllocated := openItem.AllocatedAmount.Add(alloc.AllocatedAmount)
		newStatus := domain.OpenItemStatusPartiallyPaid
		if newRemaining.IsZero() {
			newStatus = domain.OpenItemStatusPaid
		}

		_, err = s.receivablesRepo.UpdateOpenItemAllocation(ctx, openItem.ID, newAllocated, newRemaining, newStatus)
		if err != nil {
			return err
		}

		_, err = s.receivablesRepo.CreateReceiptAllocation(ctx, &domain.ReceiptAllocation{
			ReceiptID:       receipt.ID,
			OpenItemID:      openItem.ID,
			AllocatedAmount: alloc.AllocatedAmount,
			CreatedByUserID: userID,
		})
		if err != nil {
			return err
		}
	}

	newReceiptAllocated := receipt.AllocatedAmount.Add(totalToAllocate)
	newReceiptUnallocated := receipt.UnallocatedAmount.Sub(totalToAllocate)
	if err := s.receivablesRepo.UpdateReceiptAllocationAmounts(ctx, receipt.ID, newReceiptAllocated, newReceiptUnallocated); err != nil {
		return err
	}

	_, err = s.customerRepo.UpdateCustomerBalance(ctx, receipt.CustomerID, totalToAllocate.Neg())
	return err
}

func (s *receivablesService) CreateOpenItem(ctx context.Context, item *domain.OpenItem) (*domain.OpenItem, error) {
	var openItem *domain.OpenItem
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		openItem, err = s.createOpenItem(txCtx, item)
		return err
	})
	return openItem, err
}

func (s *receivablesService) createOpenItem(ctx context.Context, item *domain.OpenItem) (*domain.OpenItem, error) {
	if item.OriginalAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("open item original amount must be greater than zero")
	}

	item.AllocatedAmount = decimal.Zero
	item.RemainingAmount = item.OriginalAmount
	item.Status = domain.OpenItemStatusOpen

	created, err := s.receivablesRepo.CreateOpenItem(ctx, item)
	if err != nil {
		return nil, err
	}

	// Invoices increase customer outstanding balance
	if item.DocumentType == domain.DocumentTypeInvoice || item.DocumentType == domain.DocumentTypeDebitNote {
		if _, err := s.customerRepo.UpdateCustomerBalance(ctx, item.CustomerID, item.OriginalAmount); err != nil {
			return nil, fmt.Errorf("failed to update customer balance: %w", err)
		}
	}

	return created, nil
}

func (s *receivablesService) ListOpenItems(ctx context.Context, customerID uuid.UUID, openOnly bool) ([]*domain.OpenItem, error) {
	return s.receivablesRepo.ListOpenItemsByCustomer(ctx, customerID, openOnly)
}

func (s *receivablesService) GetCustomerAging(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAgingSummary, error) {
	return s.receivablesRepo.GetCustomerAgingSummary(ctx, customerID)
}

// Validate once so header totals and applied lines describe the same allocation.
func validateAllocations(input []ports.AllocationCommand) ([]ports.AllocationCommand, decimal.Decimal, error) {
	result := append([]ports.AllocationCommand(nil), input...)
	total := decimal.Zero
	seen := make(map[uuid.UUID]bool, len(result))
	for _, item := range result {
		if item.OpenItemID == uuid.Nil || !item.AllocatedAmount.IsPositive() {
			return nil, total, fmt.Errorf("allocation requires an invoice and a positive amount")
		}
		if seen[item.OpenItemID] {
			return nil, total, fmt.Errorf("duplicate invoice allocation")
		}
		seen[item.OpenItemID] = true
		total = total.Add(item.AllocatedAmount)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].OpenItemID.String() < result[j].OpenItemID.String() })
	return result, total, nil
}
