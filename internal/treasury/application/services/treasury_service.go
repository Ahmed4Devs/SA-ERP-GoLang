package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/treasury/domain"
	"erp/internal/treasury/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type treasuryService struct {
	repo ports.TreasuryRepository
	db   *database.DB
}

func NewTreasuryService(repo ports.TreasuryRepository, db *database.DB) ports.TreasuryService {
	return &treasuryService{
		repo: repo,
		db:   db,
	}
}

// ============================================================================
// CASH ACCOUNTS
// ============================================================================

func (s *treasuryService) CreateCashAccount(ctx context.Context, a *domain.CashAccount) (*domain.CashAccount, error) {
	if a.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if a.AccountCode == "" {
		return nil, fmt.Errorf("account_code is required")
	}
	if a.NameEn == "" || a.NameAr == "" {
		return nil, fmt.Errorf("name_en and name_ar are required")
	}
	if a.MaxLimit.IsNegative() {
		return nil, fmt.Errorf("max_limit cannot be negative")
	}
	if a.CurrentBalance.IsZero() {
		a.CurrentBalance = decimal.Zero
	}
	if a.Status == "" {
		a.Status = domain.CashAccountActive
	}
	if a.CurrencyCode == "" {
		a.CurrencyCode = "SAR"
	}
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
	if r.PaymentMethod == "" {
		r.PaymentMethod = domain.PaymentMethodCash
	}
	if r.Status == "" {
		r.Status = domain.ReceiptDraft
	}
	return s.repo.CreateReceipt(ctx, r)
}

func (s *treasuryService) PostReceipt(ctx context.Context, legalEntityID, id uuid.UUID, journalEntryID *uuid.UUID) (*domain.Receipt, error) {
	rc, err := s.repo.GetReceiptByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if rc.Status == domain.ReceiptPosted {
		return rc, nil
	}
	if rc.Status == domain.ReceiptVoided {
		return nil, domain.ErrInvalidStatusTransition
	}

	// Update cash account balance
	acc, err := s.repo.GetCashAccountByID(ctx, legalEntityID, rc.CashAccountID)
	if err != nil {
		return nil, err
	}
	newBalance := acc.CurrentBalance.Add(rc.Amount)
	if acc.MaxLimit.IsPositive() && newBalance.GreaterThan(acc.MaxLimit) {
		return nil, domain.ErrCashAccountLimitExceeded
	}

	if _, err := s.repo.UpdateCashAccountBalance(ctx, legalEntityID, rc.CashAccountID, rc.Amount); err != nil {
		return nil, err
	}

	return s.repo.UpdateReceiptStatus(ctx, legalEntityID, id, domain.ReceiptPosted, journalEntryID)
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
	if p.Status == "" {
		p.Status = domain.PaymentDraft
	}
	return s.repo.CreatePayment(ctx, p)
}

func (s *treasuryService) ApprovePayment(ctx context.Context, legalEntityID, id uuid.UUID, approverID uuid.UUID) (*domain.Payment, error) {
	p, err := s.repo.GetPaymentByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.PaymentDraft {
		return nil, domain.ErrInvalidStatusTransition
	}
	return s.repo.UpdatePaymentStatus(ctx, legalEntityID, id, domain.PaymentApproved, nil, &approverID)
}

func (s *treasuryService) PayPayment(ctx context.Context, legalEntityID, id uuid.UUID, journalEntryID *uuid.UUID) (*domain.Payment, error) {
	p, err := s.repo.GetPaymentByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if p.Status == domain.PaymentPaid {
		return p, nil
	}
	if p.Status != domain.PaymentDraft && p.Status != domain.PaymentApproved {
		return nil, domain.ErrInvalidStatusTransition
	}

	// Verify balance
	acc, err := s.repo.GetCashAccountByID(ctx, legalEntityID, p.CashAccountID)
	if err != nil {
		return nil, err
	}
	if acc.CurrentBalance.LessThan(p.Amount) {
		return nil, domain.ErrInsufficientCashBalance
	}

	// Decrement cash balance
	negAmount := p.Amount.Neg()
	if _, err := s.repo.UpdateCashAccountBalance(ctx, legalEntityID, p.CashAccountID, negAmount); err != nil {
		return nil, err
	}

	return s.repo.UpdatePaymentStatus(ctx, legalEntityID, id, domain.PaymentPaid, journalEntryID, nil)
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
	if t.TransferNumber == "" {
		t.TransferNumber = identifiers.NewDocumentNumber("TRF", time.Now())
	}
	if t.TransferDate.IsZero() {
		t.TransferDate = time.Now()
	}
	if t.CurrencyCode == "" {
		t.CurrencyCode = "SAR"
	}
	if t.Status == "" {
		t.Status = domain.TransferDraft
	}
	return s.repo.CreateTransfer(ctx, t)
}

func (s *treasuryService) CompleteTransfer(ctx context.Context, legalEntityID, id uuid.UUID, completedBy uuid.UUID, journalEntryID *uuid.UUID) (*domain.Transfer, error) {
	t, err := s.repo.GetTransferByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.TransferCompleted {
		return t, nil
	}
	if t.Status == domain.TransferCancelled {
		return nil, domain.ErrInvalidStatusTransition
	}

	// If source is Cash, check and deduct
	if t.SourceType == domain.TransferSourceCash && t.SourceCashAccountID != nil {
		sourceAcc, err := s.repo.GetCashAccountByID(ctx, legalEntityID, *t.SourceCashAccountID)
		if err != nil {
			return nil, err
		}
		totalDeduction := t.Amount.Add(t.FeeAmount)
		if sourceAcc.CurrentBalance.LessThan(totalDeduction) {
			return nil, domain.ErrInsufficientCashBalance
		}
		if _, err := s.repo.UpdateCashAccountBalance(ctx, legalEntityID, *t.SourceCashAccountID, totalDeduction.Neg()); err != nil {
			return nil, err
		}
	}

	// If destination is Cash, increment
	if t.DestinationType == domain.TransferDestinationCash && t.DestinationCashAccountID != nil {
		if _, err := s.repo.UpdateCashAccountBalance(ctx, legalEntityID, *t.DestinationCashAccountID, t.Amount); err != nil {
			return nil, err
		}
	}

	return s.repo.UpdateTransferStatus(ctx, legalEntityID, id, domain.TransferCompleted, journalEntryID, &completedBy)
}

func (s *treasuryService) GetTransfer(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Transfer, error) {
	return s.repo.GetTransferByID(ctx, legalEntityID, id)
}

func (s *treasuryService) ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Transfer, error) {
	return s.repo.ListTransfers(ctx, legalEntityID)
}
