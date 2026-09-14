package ports

import (
	"context"

	"erp/internal/organization/domain"
	"github.com/google/uuid"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, code, nameAr, nameEn string) (*domain.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	ListOrganizations(ctx context.Context) ([]*domain.Organization, error)

	CreateLegalEntity(ctx context.Context, orgID uuid.UUID, code, nameAr, nameEn, crNumber, baseCurrency string) (*domain.LegalEntity, error)
	GetLegalEntity(ctx context.Context, id uuid.UUID) (*domain.LegalEntity, error)
	ListLegalEntities(ctx context.Context, orgID uuid.UUID) ([]*domain.LegalEntity, error)

	CreateBranch(ctx context.Context, legalEntityID uuid.UUID, code, nameAr, nameEn string) (*domain.Branch, error)
	ListBranches(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Branch, error)

	CreateDepartment(ctx context.Context, legalEntityID uuid.UUID, parentID *uuid.UUID, code, nameAr, nameEn string) (*domain.Department, error)
	ListDepartments(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Department, error)

	CreateCostCenter(ctx context.Context, legalEntityID uuid.UUID, parentID *uuid.UUID, code, nameAr, nameEn string) (*domain.CostCenter, error)
	ListCostCenters(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.CostCenter, error)

	CreateNationalAddress(ctx context.Context, legalEntityID, branchID *uuid.UUID, bldg, streetAr, streetEn, distAr, distEn, cityAr, cityEn, postal, addl string) (*domain.Address, error)
}

type NumberSequenceService interface {
	DefineSequence(ctx context.Context, legalEntityID uuid.UUID, code, prefix, suffix string, startVal int64, step, padding int32, reset domain.ResetInterval) (*domain.NumberSequence, error)
	GenerateNext(ctx context.Context, legalEntityID uuid.UUID, code string) (string, error)
	ListSequences(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.NumberSequence, error)
}
