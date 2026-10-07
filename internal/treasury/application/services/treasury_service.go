package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"erp/internal/accounting/application/posting"
	accountingdomain "erp/internal/accounting/domain"
	accountingports "erp/internal/accounting/ports"
	"erp/internal/platform/identifiers"
	"erp/internal/treasury/domain"
	"erp/internal/treasury/ports"

	"github.com/google/uuid"
)

type treasuryService struct {
	repo   ports.TreasuryRepository
	ledger *posting.SourceService
	audit  accountingports.PostingAuditRecorder
}

func NewTreasuryService(repo ports.TreasuryRepository, ledger *posting.SourceService, audit accountingports.PostingAuditRecorder) ports.TreasuryService {
	return &treasuryService{
		repo:   repo,
		ledger: ledger,
		audit:  audit,
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
	if r.CreatedByUserID == nil || *r.CreatedByUserID == uuid.Nil {
		return nil, fmt.Errorf("created_by_user_id is required")
	}
	r.JournalEntryID = nil
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

func (s *treasuryService) PostReceipt(ctx context.Context, legalEntityID, id uuid.UUID, actorID uuid.UUID) (*domain.Receipt, error) {
	if actorID == uuid.Nil {
		return nil, fmt.Errorf("posting actor is required")
	}
	var posted *domain.Receipt
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		rc, err := s.repo.LockReceiptByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if rc.Status == domain.ReceiptPosted {
			posted = rc
			return nil
		}
		if err := validateChecker(rc.CreatedByUserID, actorID); err != nil {
			return err
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
		cashGL, err := s.requiredAccount(txCtx, legalEntityID, account.GLAccountID)
		if err != nil {
			return err
		}
		offsetGL, err := s.requiredAccount(txCtx, legalEntityID, rc.GLOffsetAccountID)
		if err != nil {
			return err
		}
		journalID, err := s.ledger.Post(txCtx, "CASH", accountingdomain.PostingRequest{
			LegalEntityID: legalEntityID, SourceModule: "TREASURY", SourceType: "RECEIPT", SourceID: rc.ID,
			AccountingDate: rc.ReceiptDate, CurrencyCode: rc.CurrencyCode, ActorID: actorID,
			Reference: rc.ReceiptNumber, Description: "Treasury receipt " + rc.ReceiptNumber,
			Lines: []accountingdomain.PostingLine{{AccountID: cashGL, TransactionDebit: rc.Amount, BranchID: account.BranchID}, {AccountID: offsetGL, TransactionCredit: rc.Amount}},
		})
		if err != nil {
			return err
		}
		posted, err = s.repo.UpdateReceiptStatus(txCtx, legalEntityID, id, domain.ReceiptPosted, &journalID)
		if err != nil {
			return err
		}
		return s.recordTransition(txCtx, legalEntityID, actorID, "receipt.post", "receipts", rc.ID, rc, posted)
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
	if p.CreatedByUserID == nil || *p.CreatedByUserID == uuid.Nil {
		return nil, fmt.Errorf("created_by_user_id is required")
	}
	p.JournalEntryID, p.ApprovedByUserID = nil, nil
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
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		payment, err := s.repo.LockPaymentByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if err := validateChecker(payment.CreatedByUserID, approverID); err != nil {
			return err
		}
		if payment.Status != domain.PaymentDraft {
			return domain.ErrInvalidStatusTransition
		}
		approved, err = s.repo.UpdatePaymentStatus(txCtx, legalEntityID, id, domain.PaymentApproved, nil, &approverID)
		if err != nil {
			return err
		}
		return s.recordTransition(txCtx, legalEntityID, approverID, "payment.approve", "payments", payment.ID, payment, approved)
	})
	if err != nil {
		return nil, err
	}
	return approved, nil
}

func (s *treasuryService) PayPayment(ctx context.Context, legalEntityID, id uuid.UUID, actorID uuid.UUID) (*domain.Payment, error) {
	if actorID == uuid.Nil {
		return nil, fmt.Errorf("posting actor is required")
	}
	var paid *domain.Payment
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		payment, err := s.repo.LockPaymentByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if payment.Status == domain.PaymentPaid {
			paid = payment
			return nil
		}
		if err := validateChecker(payment.CreatedByUserID, actorID); err != nil {
			return err
		}
		if payment.Status != domain.PaymentApproved || payment.ApprovedByUserID == nil || *payment.ApprovedByUserID == uuid.Nil {
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
		cashGL, err := s.requiredAccount(txCtx, legalEntityID, account.GLAccountID)
		if err != nil {
			return err
		}
		offsetGL, err := s.requiredAccount(txCtx, legalEntityID, payment.GLOffsetAccountID)
		if err != nil {
			return err
		}
		journalID, err := s.ledger.Post(txCtx, "CASH", accountingdomain.PostingRequest{
			LegalEntityID: legalEntityID, SourceModule: "TREASURY", SourceType: "PAYMENT", SourceID: payment.ID,
			AccountingDate: payment.PaymentDate, CurrencyCode: payment.CurrencyCode, ActorID: actorID,
			Reference: payment.PaymentNumber, Description: "Treasury payment " + payment.PaymentNumber,
			Lines: []accountingdomain.PostingLine{{AccountID: offsetGL, TransactionDebit: payment.Amount}, {AccountID: cashGL, TransactionCredit: payment.Amount, BranchID: account.BranchID}},
		})
		if err != nil {
			return err
		}
		paid, err = s.repo.UpdatePaymentStatus(txCtx, legalEntityID, id, domain.PaymentPaid, &journalID, nil)
		if err != nil {
			return err
		}
		return s.recordTransition(txCtx, legalEntityID, actorID, "payment.pay", "payments", payment.ID, payment, paid)
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
	if t.InitiatedByUserID == nil || *t.InitiatedByUserID == uuid.Nil {
		return nil, fmt.Errorf("initiated_by_user_id is required")
	}
	t.JournalEntryID, t.CompletedByUserID = nil, nil
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

func (s *treasuryService) CompleteTransfer(ctx context.Context, legalEntityID, id uuid.UUID, completedBy uuid.UUID) (*domain.Transfer, error) {
	if completedBy == uuid.Nil {
		return nil, fmt.Errorf("completed_by is required")
	}
	var completed *domain.Transfer
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
		transfer, err := s.repo.LockTransferByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if transfer.Status == domain.TransferCompleted {
			completed = transfer
			return nil
		}
		if err := validateChecker(transfer.InitiatedByUserID, completedBy); err != nil {
			return err
		}
		if transfer.Status != domain.TransferDraft && transfer.Status != domain.TransferInTransit {
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

		var sourceID, destinationID *uuid.UUID
		if transfer.SourceType == domain.TransferSourceCash {
			sourceID = cashAccounts[*transfer.SourceCashAccountID].GLAccountID
		} else {
			sourceID = bankAccounts[*transfer.SourceBankAccountID].GLAccountID
		}
		if transfer.DestinationType == domain.TransferDestinationCash {
			destinationID = cashAccounts[*transfer.DestinationCashAccountID].GLAccountID
		} else {
			destinationID = bankAccounts[*transfer.DestinationBankAccountID].GLAccountID
		}
		sourceGL, err := s.requiredAccount(txCtx, legalEntityID, sourceID)
		if err != nil {
			return err
		}
		destinationGL, err := s.requiredAccount(txCtx, legalEntityID, destinationID)
		if err != nil {
			return err
		}
		lines := []accountingdomain.PostingLine{{AccountID: destinationGL, TransactionDebit: transfer.Amount}, {AccountID: sourceGL, TransactionCredit: totalDeduction}}
		if transfer.FeeAmount.IsPositive() {
			feeGL, err := s.requiredAccount(txCtx, legalEntityID, transfer.FeeGLAccountID)
			if err != nil {
				return err
			}
			lines = append(lines, accountingdomain.PostingLine{AccountID: feeGL, TransactionDebit: transfer.FeeAmount})
		}
		journalType := "CASH"
		if transfer.SourceType == domain.TransferSourceBank || transfer.DestinationType == domain.TransferDestinationBank {
			journalType = "BANK"
		}
		journalID, err := s.ledger.Post(txCtx, journalType, accountingdomain.PostingRequest{
			LegalEntityID: legalEntityID, SourceModule: "TREASURY", SourceType: "TRANSFER", SourceID: transfer.ID,
			AccountingDate: transfer.TransferDate, CurrencyCode: transfer.CurrencyCode, ActorID: completedBy,
			Reference: transfer.TransferNumber, Description: "Treasury transfer " + transfer.TransferNumber, Lines: lines,
		})
		if err != nil {
			return err
		}
		completed, err = s.repo.UpdateTransferStatus(txCtx, legalEntityID, id, domain.TransferCompleted, &journalID, &completedBy)
		if err != nil {
			return err
		}
		return s.recordTransition(txCtx, legalEntityID, completedBy, "transfer.complete", "transfers", transfer.ID, transfer, completed)
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
