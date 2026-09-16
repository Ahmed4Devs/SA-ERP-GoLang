package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"erp/internal/receivables/domain"
	"erp/internal/receivables/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type customerService struct {
	customerRepo ports.CustomerRepository
}

func NewCustomerService(customerRepo ports.CustomerRepository) ports.CustomerService {
	return &customerService{
		customerRepo: customerRepo,
	}
}

func (s *customerService) CreateCustomer(ctx context.Context, cmd ports.CreateCustomerCommand) (*domain.Customer, error) {
	currencyCode := cmd.CurrencyCode
	if currencyCode == "" {
		currencyCode = "SAR"
	}

	custNum := fmt.Sprintf("CUST-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)

	customer := &domain.Customer{
		LegalEntityID:    cmd.LegalEntityID,
		CustomerNumber:   custNum,
		CustomerType:     cmd.CustomerType,
		NameEn:           cmd.NameEn,
		NameAr:           cmd.NameAr,
		TaxID:            cmd.TaxID,
		CRNumber:         cmd.CRNumber,
		NationalID:       cmd.NationalID,
		Address:          cmd.Address,
		Email:            cmd.Email,
		Phone:            cmd.Phone,
		ContactPerson:    cmd.ContactPerson,
		PaymentTermsID:   cmd.PaymentTermsID,
		CurrencyCode:     currencyCode,
		DefaultAccountID: cmd.DefaultAccountID,
		Status:           domain.CustomerStatusActive,
	}

	if err := customer.Validate(); err != nil {
		return nil, err
	}

	created, err := s.customerRepo.CreateCustomer(ctx, customer)
	if err != nil {
		return nil, err
	}

	// Setup Credit Limit
	blockAction := cmd.BlockAction
	if blockAction == "" {
		blockAction = domain.BlockActionWarn
	}
	creditLimit := cmd.CreditLimit
	if creditLimit.IsNegative() {
		creditLimit = decimal.Zero
	}

	_, err = s.customerRepo.UpsertCreditLimit(ctx, &domain.CreditLimit{
		CustomerID:     created.ID,
		CreditLimit:    creditLimit,
		CurrentBalance: decimal.Zero,
		MaxOverdueDays: cmd.MaxOverdueDays,
		BlockAction:    blockAction,
		IsCreditHold:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup customer credit limit: %w", err)
	}

	// Link GL Account if specified
	if cmd.DefaultAccountID != nil && *cmd.DefaultAccountID != uuid.Nil {
		_, _ = s.customerRepo.UpsertCustomerAccount(ctx, &domain.CustomerAccount{
			CustomerID:   created.ID,
			GLAccountID:  *cmd.DefaultAccountID,
			CurrencyCode: currencyCode,
			IsPrimary:    true,
		})
	}

	return created, nil
}

func (s *customerService) GetCustomer(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	return s.customerRepo.GetCustomerByID(ctx, id)
}

func (s *customerService) ListCustomers(ctx context.Context, filter ports.CustomerFilter) ([]*domain.Customer, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.CustomerType == "" {
		filter.CustomerType = "ALL"
	}
	if filter.Status == "" {
		filter.Status = "ALL"
	}
	return s.customerRepo.ListCustomers(ctx, filter)
}

func (s *customerService) UpdateCustomer(ctx context.Context, cmd ports.UpdateCustomerCommand) (*domain.Customer, error) {
	existing, err := s.customerRepo.GetCustomerByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	existing.NameEn = cmd.NameEn
	existing.NameAr = cmd.NameAr
	existing.TaxID = cmd.TaxID
	existing.CRNumber = cmd.CRNumber
	existing.NationalID = cmd.NationalID
	existing.Address = cmd.Address
	existing.Email = cmd.Email
	existing.Phone = cmd.Phone
	existing.ContactPerson = cmd.ContactPerson
	existing.PaymentTermsID = cmd.PaymentTermsID
	if cmd.CurrencyCode != "" {
		existing.CurrencyCode = cmd.CurrencyCode
	}
	existing.DefaultAccountID = cmd.DefaultAccountID

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	return s.customerRepo.UpdateCustomer(ctx, existing)
}

func (s *customerService) SetCustomerStatus(ctx context.Context, id uuid.UUID, status domain.CustomerStatus) error {
	return s.customerRepo.UpdateCustomerStatus(ctx, id, status)
}

func (s *customerService) SetCreditHold(ctx context.Context, customerID uuid.UUID, isHold bool, reason *string) error {
	return s.customerRepo.SetCreditHold(ctx, customerID, isHold, reason)
}

func (s *customerService) CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error) {
	term.Code = strings.ToUpper(strings.TrimSpace(term.Code))
	return s.customerRepo.CreatePaymentTerm(ctx, term)
}

func (s *customerService) ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.PaymentTerm, error) {
	return s.customerRepo.ListPaymentTerms(ctx, legalEntityID, false)
}
