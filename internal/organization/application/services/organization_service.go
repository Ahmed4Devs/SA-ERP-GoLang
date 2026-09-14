package services

import (
	"context"
	"fmt"
	"strings"

	"erp/internal/organization/domain"
	"erp/internal/organization/ports"
	"github.com/google/uuid"
)

type OrganizationService struct {
	orgRepo        ports.OrganizationRepository
	legalEntityRepo ports.LegalEntityRepository
	branchRepo     ports.BranchRepository
	deptRepo       ports.DepartmentRepository
	costCenterRepo ports.CostCenterRepository
	addrRepo       ports.AddressRepository
}

func NewOrganizationService(
	orgRepo ports.OrganizationRepository,
	legalEntityRepo ports.LegalEntityRepository,
	branchRepo ports.BranchRepository,
	deptRepo ports.DepartmentRepository,
	costCenterRepo ports.CostCenterRepository,
	addrRepo ports.AddressRepository,
) ports.OrganizationService {
	return &OrganizationService{
		orgRepo:        orgRepo,
		legalEntityRepo: legalEntityRepo,
		branchRepo:     branchRepo,
		deptRepo:       deptRepo,
		costCenterRepo: costCenterRepo,
		addrRepo:       addrRepo,
	}
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, code, nameAr, nameEn string) (*domain.Organization, error) {
	if err := domain.ValidateCode(code); err != nil {
		return nil, err
	}
	if strings.TrimSpace(nameAr) == "" || strings.TrimSpace(nameEn) == "" {
		return nil, fmt.Errorf("both Arabic and English organization names are required")
	}

	org := &domain.Organization{
		Code:   strings.ToUpper(strings.TrimSpace(code)),
		NameAr: strings.TrimSpace(nameAr),
		NameEn: strings.TrimSpace(nameEn),
		Status: domain.OrgStatusActive,
	}

	return s.orgRepo.Create(ctx, org)
}

func (s *OrganizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return s.orgRepo.GetByID(ctx, id)
}

func (s *OrganizationService) ListOrganizations(ctx context.Context) ([]*domain.Organization, error) {
	return s.orgRepo.List(ctx)
}

func (s *OrganizationService) CreateLegalEntity(ctx context.Context, orgID uuid.UUID, code, nameAr, nameEn, crNumber, baseCurrency string) (*domain.LegalEntity, error) {
	if err := domain.ValidateCode(code); err != nil {
		return nil, err
	}
	if err := domain.ValidateCRNumber(crNumber); err != nil {
		return nil, err
	}
	if strings.TrimSpace(nameAr) == "" || strings.TrimSpace(nameEn) == "" {
		return nil, fmt.Errorf("both Arabic and English legal entity names are required")
	}
	if strings.TrimSpace(baseCurrency) == "" {
		baseCurrency = "SAR"
	}

	entity := &domain.LegalEntity{
		OrganizationID: orgID,
		Code:           strings.ToUpper(strings.TrimSpace(code)),
		NameAr:         strings.TrimSpace(nameAr),
		NameEn:         strings.TrimSpace(nameEn),
		CRNumber:       strings.TrimSpace(crNumber),
		BaseCurrency:   strings.ToUpper(strings.TrimSpace(baseCurrency)),
		IsActive:       true,
	}

	return s.legalEntityRepo.Create(ctx, entity)
}

func (s *OrganizationService) GetLegalEntity(ctx context.Context, id uuid.UUID) (*domain.LegalEntity, error) {
	return s.legalEntityRepo.GetByID(ctx, id)
}

func (s *OrganizationService) ListLegalEntities(ctx context.Context, orgID uuid.UUID) ([]*domain.LegalEntity, error) {
	return s.legalEntityRepo.ListByOrg(ctx, orgID)
}

func (s *OrganizationService) CreateBranch(ctx context.Context, legalEntityID uuid.UUID, code, nameAr, nameEn string) (*domain.Branch, error) {
	if err := domain.ValidateCode(code); err != nil {
		return nil, err
	}
	if strings.TrimSpace(nameAr) == "" || strings.TrimSpace(nameEn) == "" {
		return nil, fmt.Errorf("both Arabic and English branch names are required")
	}

	branch := &domain.Branch{
		LegalEntityID: legalEntityID,
		Code:          strings.ToUpper(strings.TrimSpace(code)),
		NameAr:        strings.TrimSpace(nameAr),
		NameEn:        strings.TrimSpace(nameEn),
		IsActive:      true,
	}

	return s.branchRepo.Create(ctx, branch)
}

func (s *OrganizationService) ListBranches(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Branch, error) {
	return s.branchRepo.ListByEntity(ctx, legalEntityID)
}

func (s *OrganizationService) CreateDepartment(ctx context.Context, legalEntityID uuid.UUID, parentID *uuid.UUID, code, nameAr, nameEn string) (*domain.Department, error) {
	if err := domain.ValidateCode(code); err != nil {
		return nil, err
	}
	if strings.TrimSpace(nameAr) == "" || strings.TrimSpace(nameEn) == "" {
		return nil, fmt.Errorf("both Arabic and English department names are required")
	}

	dept := &domain.Department{
		LegalEntityID: legalEntityID,
		ParentID:      parentID,
		Code:          strings.ToUpper(strings.TrimSpace(code)),
		NameAr:        strings.TrimSpace(nameAr),
		NameEn:        strings.TrimSpace(nameEn),
		IsActive:      true,
	}

	return s.deptRepo.Create(ctx, dept)
}

func (s *OrganizationService) ListDepartments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Department, error) {
	return s.deptRepo.ListByEntity(ctx, legalEntityID)
}

func (s *OrganizationService) CreateCostCenter(ctx context.Context, legalEntityID uuid.UUID, parentID *uuid.UUID, code, nameAr, nameEn string) (*domain.CostCenter, error) {
	if err := domain.ValidateCode(code); err != nil {
		return nil, err
	}
	if strings.TrimSpace(nameAr) == "" || strings.TrimSpace(nameEn) == "" {
		return nil, fmt.Errorf("both Arabic and English cost center names are required")
	}

	cc := &domain.CostCenter{
		LegalEntityID: legalEntityID,
		ParentID:      parentID,
		Code:          strings.ToUpper(strings.TrimSpace(code)),
		NameAr:        strings.TrimSpace(nameAr),
		NameEn:        strings.TrimSpace(nameEn),
		IsActive:      true,
	}

	return s.costCenterRepo.Create(ctx, cc)
}

func (s *OrganizationService) ListCostCenters(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.CostCenter, error) {
	return s.costCenterRepo.ListByEntity(ctx, legalEntityID)
}

func (s *OrganizationService) CreateNationalAddress(
	ctx context.Context,
	legalEntityID, branchID *uuid.UUID,
	bldg, streetAr, streetEn, distAr, distEn, cityAr, cityEn, postal, addl string,
) (*domain.Address, error) {
	if legalEntityID == nil && branchID == nil {
		return nil, fmt.Errorf("address must be attached to either a legal entity or a branch")
	}

	if err := domain.ValidateNationalAddress(bldg, postal, addl); err != nil {
		return nil, err
	}

	addr := &domain.Address{
		LegalEntityID:    legalEntityID,
		BranchID:         branchID,
		AddressType:      "NATIONAL_ADDRESS",
		BuildingNumber:   strings.TrimSpace(bldg),
		StreetNameAr:     strings.TrimSpace(streetAr),
		StreetNameEn:     strings.TrimSpace(streetEn),
		DistrictAr:       strings.TrimSpace(distAr),
		DistrictEn:       strings.TrimSpace(distEn),
		CityAr:           strings.TrimSpace(cityAr),
		CityEn:           strings.TrimSpace(cityEn),
		PostalCode:       strings.TrimSpace(postal),
		AdditionalNumber: strings.TrimSpace(addl),
		CountryCode:      "SA",
	}

	return s.addrRepo.Create(ctx, addr)
}
