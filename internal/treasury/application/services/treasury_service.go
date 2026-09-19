package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"
	"erp/internal/treasury/domain"
	"erp/internal/treasury/ports"

	"github.com/google/uuid"
)

type treasuryService struct {
	repo ports.TreasuryRepository
	tx   transaction.Manager
}

func NewTreasuryService(repo ports.TreasuryRepository, db *database.DB) ports.TreasuryService {
	return &treasuryService{
		repo: repo,
		tx:   transaction.NewManager(db.Pool),
	}
}

// ============================================================================
// CASH ACCOUNTS
// ============================================================================

func (s *treasuryService) CreateCashAccount(ctx context.Context, a *domain.CashAccount) (*domain.CashAccount, error) {
	if a.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	a.AccountCode = strings.TrimSpace(a.AccountCode)
	a.NameEn = strings.TrimSpace(a.NameEn)
	a.NameAr = strings.TrimSpace(a.NameAr)
	if a.AccountCode == "" {
		return nil, fmt.Errorf("account_code is required")
	}
	if a.NameEn == "" || a.NameAr == "" {
		return nil, fmt.Errorf("name_en and name_ar are required")
	}
	if a.MaxLimit.IsNegative() || a.CurrentBalance.IsNegative() {
		return nil, fmt.Errorf("cash account limits and balances cannot be negative")
	}
	if a.MaxLimit.IsPositive() && a.CurrentBalance.GreaterThan(a.MaxLimit) {
		return nil, domain.ErrCashAccountLimitExceeded
	}
	if a.Status == "" {
		a.Status = domain.CashAccountActive
	}
	if a.CurrencyCode == "" {
		a.CurrencyCode = "SAR"
	}
	a.CurrencyCode = strings.ToUpper(strings.TrimSpace(a.CurrencyCode))
	return s.repo.CreateCashAccount(ctx, a)
}

func (s *treasuryService) GetCashAccount(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.CashAccount, error) {
	return s.repo.GetCashAccountByID(ctx, legalEntityID, id)
}

func (s *treasuryService) ListCashAccounts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.CashAccount, error) {
	return s.repo.ListCashAccounts(ctx, legalEntityID)
}

// ============================================================================
// RECEIPTS
// ============================================================================

func (s *treasuryService) CreateReceipt(ctx context.Context, r *domain.Receipt) (*domain.Receipt, error) {
	if r.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if r.CashAccountID == uuid.Nil {
		return nil, fmt.Errorf("cash_account_id is required")
	}
	if !r.Amount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}
	if r.ReceiptNumber == "" {
		r.ReceiptNumber = identifiers.NewDocumentNumber("REC", time.Now())
	}
	if r.ReceiptDate.IsZero() {
		r.ReceiptDate = time.Now()
	}
	if r.CurrencyCode == "" {
		r.CurrencyCode = "SAR"
	}
	r.CurrencyCode = strings.ToUpper(strings.TrimSpace(r.CurrencyCode))
	if r.PaymentMethod == "" {
		r.PaymentMethod = domain.PaymentMethodCash
	}
	if r.Status == "" {
		r.Status = domain.ReceiptDraft
	} else if r.Status != domain.ReceiptDraft {
		return nil, domain.ErrInvalidStatusTransition
	}
	return s.repo.CreateReceipt(ctx, r)
}

func (s *treasuryService) PostReceipt(ctx context.Context, legalEntityID, id uuid.UUID, journalEntryID *uuid.UUID) (*domain.Receipt, error) {
	var posted *domain.Receipt
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		rc, err := s.repo.LockReceiptByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if rc.Status == domain.ReceiptPosted {
			posted = rc
			return nil
		}
		if rc.Status != domain.ReceiptDraft {
			return domain.ErrInvalidStatusTransition
		}

		account, err := s.repo.LockCashAccountByID(txCtx, legalEntityID, rc.CashAccountID)
		if err != nil {
			return err
		}
		if err := validateCashAccountForTransaction(account, rc.CurrencyCode); err != nil {
			return err
		}
		newBalance := account.CurrentBalance.Add(rc.Amount)
		if account.MaxLimit.IsPositive() && newBalance.GreaterThan(account.MaxLimit) {
			return domain.ErrCashAccountLimitExceeded
		}
		if _, err := s.repo.UpdateCashAccountBalance(txCtx, legalEntityID, rc.CashAccountID, rc.Amount); err != nil {
			return err
		}
		posted, err = s.repo.UpdateReceiptStatus(txCtx, legalEntityID, id, domain.ReceiptPosted, journalEntryID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return posted, nil
}

func (s *treasuryService) GetReceipt(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Receipt, error) {
	return s.repo.GetReceiptByID(ctx, legalEntityID, id)
}

func (s *treasuryService) ListReceipts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Receipt, error) {
	return s.repo.ListReceipts(ctx, legalEntityID)
}

// ============================================================================
// PAYMENTS
// ============================================================================

func (s *treasuryService) CreatePayment(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	if p.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if p.CashAccountID == uuid.Nil {
		return nil, fmt.Errorf("cash_account_id is required")
	}
	if !p.Amount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}
	if p.PaymentNumber == "" {
		p.PaymentNumber = identifiers.NewDocumentNumber("PMT", time.Now())
	}
	if p.PaymentDate.IsZero() {
		p.PaymentDate = time.Now()
	}
	if p.CurrencyCode == "" {
		p.CurrencyCode = "SAR"
	}
	p.CurrencyCode = strings.ToUpper(strings.TrimSpace(p.CurrencyCode))
	if p.Status == "" {
		p.Status = domain.PaymentDraft
	} else if p.Status != domain.PaymentDraft {
		return nil, domain.ErrInvalidStatusTransition
	}
	return s.repo.CreatePayment(ctx, p)
}

func (s *treasuryService) ApprovePayment(ctx context.Context, legalEntityID, id uuid.UUID, approverID uuid.UUID) (*domain.Payment, error) {
	if approverID == uuid.Nil {
		return nil, fmt.Errorf("approver_id is required")
	}
	var approved *domain.Payment
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		payment, err := s.repo.LockPaymentByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if payment.Status != domain.PaymentDraft {
			return domain.ErrInvalidStatusTransition
		}
		approved, err = s.repo.UpdatePaymentStatus(txCtx, legalEntityID, id, domain.PaymentApproved, nil, &approverID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return approved, nil
}

func (s *treasuryService) PayPayment(ctx context.Context, legalEntityID, id uuid.UUID, journalEntryID *uuid.UUID) (*domain.Payment, error) {
	var paid *domain.Payment
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		payment, err := s.repo.LockPaymentByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if payment.Status == domain.PaymentPaid {
			paid = payment
			return nil
		}
		if payment.Status != domain.PaymentDraft && payment.Status != domain.PaymentApproved {
			return domain.ErrInvalidStatusTransition
		}

		account, err := s.repo.LockCashAccountByID(txCtx, legalEntityID, payment.CashAccountID)
		if err != nil {
			return err
		}
		if err := validateCashAccountForTransaction(account, payment.CurrencyCode); err != nil {
			return err
		}
		if account.CurrentBalance.LessThan(payment.Amount) {
			return domain.ErrInsufficientCashBalance
		}
		if _, err := s.repo.UpdateCashAccountBalance(txCtx, legalEntityID, payment.CashAccountID, payment.Amount.Neg()); err != nil {
			return err
		}
		paid, err = s.repo.UpdatePaymentStatus(txCtx, legalEntityID, id, domain.PaymentPaid, journalEntryID, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return paid, nil
}

func (s *treasuryService) GetPayment(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Payment, error) {
	return s.repo.GetPaymentByID(ctx, legalEntityID, id)
}

func (s *treasuryService) ListPayments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Payment, error) {
	return s.repo.ListPayments(ctx, legalEntityID)
}

// ============================================================================
// TRANSFERS
// ============================================================================

func (s *treasuryService) CreateTransfer(ctx context.Context, t *domain.Transfer) (*domain.Transfer, error) {
	if t.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if !t.Amount.IsPositive() {
		return nil, domain.ErrInvalidAmount
	}
	if t.FeeAmount.IsNegative() {
		return nil, domain.ErrInvalidAmount
	}
	if err := validateTransferAccounts(t); err != nil {
		return nil, err
	}
	if t.TransferNumber == "" {
		t.TransferNumber = identifiers.NewDocumentNumber("TRF", time.Now())
	}
	if t.TransferDate.IsZero() {
		t.TransferDate = time.Now()
	}
	if t.CurrencyCode == "" {
		t.CurrencyCode = "SAR"
	}
	t.CurrencyCode = strings.ToUpper(strings.TrimSpace(t.CurrencyCode))
	if t.Status == "" {
		t.Status = domain.TransferDraft
	} else if t.Status != domain.TransferDraft {
		return nil, domain.ErrInvalidStatusTransition
	}
	return s.repo.CreateTransfer(ctx, t)
}

func (s *treasuryService) CompleteTransfer(ctx context.Context, legalEntityID, id uuid.UUID, completedBy uuid.UUID, journalEntryID *uuid.UUID) (*domain.Transfer, error) {
	if completedBy == uuid.Nil {
		return nil, fmt.Errorf("completed_by is required")
	}
	var completed *domain.Transfer
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		transfer, err := s.repo.LockTransferByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if transfer.Status == domain.TransferCompleted {
			completed = transfer
			return nil
		}
		if transfer.Status == domain.TransferCancelled {
			return domain.ErrInvalidStatusTransition
		}
		if err := validateTransferAccounts(transfer); err != nil {
			return err
		}

		cashAccounts, bankAccounts, err := s.lockTransferAccounts(txCtx, legalEntityID, transfer)
		if err != nil {
			return err
		}
		totalDeduction := transfer.Amount.Add(transfer.FeeAmount)

		if transfer.SourceType == domain.TransferSourceCash {
			source := cashAccounts[*transfer.SourceCashAccountID]
			if err := validateCashAccountForTransaction(source, transfer.CurrencyCode); err != nil {
				return err
			}
			if source.CurrentBalance.LessThan(totalDeduction) {
				return domain.ErrInsufficientCashBalance
			}
		} else {
			source := bankAccounts[*transfer.SourceBankAccountID]
			if err := validateBankAccountForTransaction(source, transfer.CurrencyCode); err != nil {
				return err
			}
			if source.AvailableBalance.Add(source.OverdraftLimit).LessThan(totalDeduction) {
				return domain.ErrInsufficientCashBalance
			}
		}

		if transfer.DestinationType == domain.TransferDestinationCash {
			destination := cashAccounts[*transfer.DestinationCashAccountID]
			if err := validateCashAccountForTransaction(destination, transfer.CurrencyCode); err != nil {
				return err
			}
			if destination.MaxLimit.IsPositive() && destination.CurrentBalance.Add(transfer.Amount).GreaterThan(destination.MaxLimit) {
				return domain.ErrCashAccountLimitExceeded
			}
		} else if err := validateBankAccountForTransaction(bankAccounts[*transfer.DestinationBankAccountID], transfer.CurrencyCode); err != nil {
			return err
		}

		if transfer.SourceType == domain.TransferSourceCash {
			if _, err := s.repo.UpdateCashAccountBalance(txCtx, legalEntityID, *transfer.SourceCashAccountID, totalDeduction.Neg()); err != nil {
				return err
			}
		} else if _, err := s.repo.UpdateBankAccountBalance(txCtx, legalEntityID, *transfer.SourceBankAccountID, totalDeduction.Neg()); err != nil {
			return err
		}

		if transfer.DestinationType == domain.TransferDestinationCash {
			if _, err := s.repo.UpdateCashAccountBalance(txCtx, legalEntityID, *transfer.DestinationCashAccountID, transfer.Amount); err != nil {
				return err
			}
		} else if _, err := s.repo.UpdateBankAccountBalance(txCtx, legalEntityID, *transfer.DestinationBankAccountID, transfer.Amount); err != nil {
			return err
		}

		completed, err = s.repo.UpdateTransferStatus(txCtx, legalEntityID, id, domain.TransferCompleted, journalEntryID, &completedBy)
		return err
	})
	if err != nil {
		return nil, err
	}
	return completed, nil
}

func (s *treasuryService) lockTransferAccounts(ctx context.Context, legalEntityID uuid.UUID, transfer *domain.Transfer) (map[uuid.UUID]*domain.CashAccount, map[uuid.UUID]*domain.BankAccountBalance, error) {
	cashIDs := make(map[uuid.UUID]struct{})
	if transfer.SourceCashAccountID != nil {
		cashIDs[*transfer.SourceCashAccountID] = struct{}{}
	}
	if transfer.DestinationCashAccountID != nil {
		cashIDs[*transfer.DestinationCashAccountID] = struct{}{}
	}
	orderedCashIDs := make([]uuid.UUID, 0, len(cashIDs))
	for accountID := range cashIDs {
		orderedCashIDs = append(orderedCashIDs, accountID)
	}
	sort.Slice(orderedCashIDs, func(i, j int) bool { return orderedCashIDs[i].String() < orderedCashIDs[j].String() })

	lockedCash := make(map[uuid.UUID]*domain.CashAccount, len(orderedCashIDs))
	for _, accountID := range orderedCashIDs {
		account, err := s.repo.LockCashAccountByID(ctx, legalEntityID, accountID)
		if err != nil {
			return nil, nil, err
		}
		lockedCash[accountID] = account
	}

	bankIDs := make(map[uuid.UUID]struct{})
	if transfer.SourceBankAccountID != nil {
		bankIDs[*transfer.SourceBankAccountID] = struct{}{}
	}
	if transfer.DestinationBankAccountID != nil {
		bankIDs[*transfer.DestinationBankAccountID] = struct{}{}
	}
	orderedBankIDs := make([]uuid.UUID, 0, len(bankIDs))
	for accountID := range bankIDs {
		orderedBankIDs = append(orderedBankIDs, accountID)
	}
	sort.Slice(orderedBankIDs, func(i, j int) bool { return orderedBankIDs[i].String() < orderedBankIDs[j].String() })

	lockedBanks := make(map[uuid.UUID]*domain.BankAccountBalance, len(orderedBankIDs))
	for _, accountID := range orderedBankIDs {
		account, err := s.repo.LockBankAccountByID(ctx, legalEntityID, accountID)
		if err != nil {
			return nil, nil, err
		}
		lockedBanks[accountID] = account
	}
	return lockedCash, lockedBanks, nil
}

func validateTransferAccounts(transfer *domain.Transfer) error {
	switch transfer.SourceType {
	case domain.TransferSourceCash:
		if transfer.SourceCashAccountID == nil || *transfer.SourceCashAccountID == uuid.Nil || transfer.SourceBankAccountID != nil {
			return domain.ErrInvalidTransferAccounts
		}
	case domain.TransferSourceBank:
		if transfer.SourceBankAccountID == nil || *transfer.SourceBankAccountID == uuid.Nil || transfer.SourceCashAccountID != nil {
			return domain.ErrInvalidTransferAccounts
		}
	default:
		return domain.ErrInvalidTransferAccounts
	}

	switch transfer.DestinationType {
	case domain.TransferDestinationCash:
		if transfer.DestinationCashAccountID == nil || *transfer.DestinationCashAccountID == uuid.Nil || transfer.DestinationBankAccountID != nil {
			return domain.ErrInvalidTransferAccounts
		}
	case domain.TransferDestinationBank:
		if transfer.DestinationBankAccountID == nil || *transfer.DestinationBankAccountID == uuid.Nil || transfer.DestinationCashAccountID != nil {
			return domain.ErrInvalidTransferAccounts
		}
	default:
		return domain.ErrInvalidTransferAccounts
	}

	if transfer.SourceType == domain.TransferSourceCash && transfer.DestinationType == domain.TransferDestinationCash && *transfer.SourceCashAccountID == *transfer.DestinationCashAccountID {
		return domain.ErrInvalidTransferAccounts
	}
	if transfer.SourceType == domain.TransferSourceBank && transfer.DestinationType == domain.TransferDestinationBank && *transfer.SourceBankAccountID == *transfer.DestinationBankAccountID {
		return domain.ErrInvalidTransferAccounts
	}
	return nil
}

func validateCashAccountForTransaction(account *domain.CashAccount, currencyCode string) error {
	if account == nil {
		return domain.ErrCashAccountNotFound
	}
	if account.Status != domain.CashAccountActive {
		return domain.ErrInactiveTreasuryAccount
	}
	if !strings.EqualFold(account.CurrencyCode, currencyCode) {
		return domain.ErrCurrencyMismatch
	}
	return nil
}

func validateBankAccountForTransaction(account *domain.BankAccountBalance, currencyCode string) error {
	if account == nil {
		return domain.ErrBankAccountNotFound
	}
	if account.Status != "ACTIVE" {
		return domain.ErrInactiveTreasuryAccount
	}
	if !strings.EqualFold(account.CurrencyCode, currencyCode) {
		return domain.ErrCurrencyMismatch
	}
	return nil
}

func (s *treasuryService) GetTransfer(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error) {
	return s.repo.GetTransferByID(ctx, legalEntityID, id)
}

func (s *treasuryService) ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Transfer, error) {
	return s.repo.ListTransfers(ctx, legalEntityID)
}
