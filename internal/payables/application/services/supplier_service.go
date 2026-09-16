package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"erp/internal/payables/domain"
	"erp/internal/payables/ports"
	"github.com/google/uuid"
)

type supplierService struct {
	supplierRepo ports.SupplierRepository
}

func NewSupplierService(supplierRepo ports.SupplierRepository) ports.SupplierService {
	return &supplierService{supplierRepo: supplierRepo}
}

func (s *supplierService) CreateSupplier(ctx context.Context, cmd ports.CreateSupplierCmd) (*domain.Supplier, error) {
	suppNumber := cmd.SupplierNumber
	if suppNumber == "" {
		suppNumber = fmt.Sprintf("SUP-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)
	}

	curr := cmd.CurrencyCode
	if curr == "" {
		curr = "SAR"
	}

	country := cmd.Address.CountryCode
	if country == "" {
		country = "SA"
	}
	cmd.Address.CountryCode = country

	supp := &domain.Supplier{
		LegalEntityID:             cmd.LegalEntityID,
		SupplierNumber:            suppNumber,
		SupplierType:              cmd.SupplierType,
		NameEn:                    cmd.NameEn,
		NameAr:                    cmd.NameAr,
		TaxID:                     cmd.TaxID,
		CRNumber:                  cmd.CRNumber,
		NationalID:                cmd.NationalID,
		IBAN:                      cmd.IBAN,
		BankName:                  cmd.BankName,
		SwiftCode:                 cmd.SwiftCode,
		Address:                   cmd.Address,
		Email:                     cmd.Email,
		Phone:                     cmd.Phone,
		ContactPerson:             cmd.ContactPerson,
		Website:                   cmd.Website,
		PaymentTermsID:            cmd.PaymentTermsID,
		CurrencyCode:              curr,
		DefaultExpenseAccountID:   cmd.DefaultExpenseAccountID,
		DefaultLiabilityAccountID: cmd.DefaultLiabilityAccountID,
		Status:                    domain.SupplierStatusActive,
	}

	if err := supp.Validate(); err != nil {
		return nil, err
	}

	return s.supplierRepo.CreateSupplier(ctx, supp)
}

func (s *supplierService) GetSupplier(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	return s.supplierRepo.GetSupplierByID(ctx, id)
}

func (s *supplierService) ListSuppliers(ctx context.Context, filter ports.SupplierFilter) ([]*domain.Supplier, int64, error) {
	return s.supplierRepo.ListSuppliers(ctx, filter)
}

func (s *supplierService) UpdateSupplier(ctx context.Context, supp *domain.Supplier) (*domain.Supplier, error) {
	if err := supp.Validate(); err != nil {
		return nil, err
	}
	return s.supplierRepo.UpdateSupplier(ctx, supp)
}

func (s *supplierService) SetSupplierStatus(ctx context.Context, id uuid.UUID, status domain.SupplierStatus) error {
	return s.supplierRepo.SetSupplierStatus(ctx, id, status)
}

func (s *supplierService) CreatePaymentTerm(ctx context.Context, term *domain.PaymentTerm) (*domain.PaymentTerm, error) {
	return s.supplierRepo.CreatePaymentTerm(ctx, term)
}

func (s *supplierService) ListPaymentTerms(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.PaymentTerm, error) {
	return s.supplierRepo.ListPaymentTerms(ctx, legalEntityID, false)
}
