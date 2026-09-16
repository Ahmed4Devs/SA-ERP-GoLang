package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type receivablesService struct {
	customerRepo    ports.CustomerRepository
	receivablesRepo ports.ReceivablesRepository
}

func NewReceivablesService(
	customerRepo ports.CustomerRepository,
	receivablesRepo ports.ReceivablesRepository,
) ports.ReceivablesService {
	return &receivablesService{
		customerRepo:    customerRepo,
		receivablesRepo: receivablesRepo,
	}
}

func (s *receivablesService) CreateReceipt(ctx context.Context, cmd ports.CreateReceiptCommand) (*domain.Receipt, error) {
	if cmd.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("receipt amount must be greater than zero")
	}

	customer, err := s.customerRepo.GetCustomerByID(ctx, cmd.CustomerID)
	if err != nil {
		return nil, err
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

	receiptNumber := fmt.Sprintf("REC-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)

	totalAllocated := decimal.Zero
	for _, alloc := range cmd.Allocations {
		if alloc.AllocatedAmount.LessThanOrEqual(decimal.Zero) {
			continue
		}
		totalAllocated = totalAllocated.Add(alloc.AllocatedAmount)
	}

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
		_, _ = s.customerRepo.UpdateCustomerBalance(ctx, cmd.CustomerID, totalAllocated.Neg())
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
	receipt, err := s.receivablesRepo.GetReceiptByID(ctx, receiptID)
	if err != nil {
		return err
	}

	totalToAllocate := decimal.Zero
	for _, a := range allocations {
		totalToAllocate = totalToAllocate.Add(a.AllocatedAmount)
	}

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

	_, _ = s.customerRepo.UpdateCustomerBalance(ctx, receipt.CustomerID, totalToAllocate.Neg())
	return nil
}

func (s *receivablesService) CreateOpenItem(ctx context.Context, item *domain.OpenItem) (*domain.OpenItem, error) {
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
		_, _ = s.customerRepo.UpdateCustomerBalance(ctx, item.CustomerID, item.OriginalAmount)
	}

	return created, nil
}

func (s *receivablesService) ListOpenItems(ctx context.Context, customerID uuid.UUID, openOnly bool) ([]*domain.OpenItem, error) {
	return s.receivablesRepo.ListOpenItemsByCustomer(ctx, customerID, openOnly)
}

func (s *receivablesService) GetCustomerAging(ctx context.Context, customerID uuid.UUID) (*domain.CustomerAgingSummary, error) {
	return s.receivablesRepo.GetCustomerAgingSummary(ctx, customerID)
}
