package ports

import (
	"context"

	"erp/internal/organization/domain"
	"github.com/google/uuid"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	GetByCode(ctx context.Context, code string) (*domain.Organization, error)
	List(ctx context.Context) ([]*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error)
}

type LegalEntityRepository interface {
	Create(ctx context.Context, entity *domain.LegalEntity) (*domain.LegalEntity, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.LegalEntity, error)
	GetByOrgAndCode(ctx context.Context, orgID uuid.UUID, code string) (*domain.LegalEntity, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.LegalEntity, error)
	ListActive(ctx context.Context) ([]*domain.LegalEntity, error)
	Update(ctx context.Context, entity *domain.LegalEntity) (*domain.LegalEntity, error)
}

type BranchRepository interface {
	Create(ctx context.Context, branch *domain.Branch) (*domain.Branch, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error)
	GetByEntityAndCode(ctx context.Context, entityID uuid.UUID, code string) (*domain.Branch, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.Branch, error)
	Update(ctx context.Context, branch *domain.Branch) (*domain.Branch, error)
}

type DepartmentRepository interface {
	Create(ctx context.Context, dept *domain.Department) (*domain.Department, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Department, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.Department, error)
	Update(ctx context.Context, dept *domain.Department) (*domain.Department, error)
}

type CostCenterRepository interface {
	Create(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CostCenter, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.CostCenter, error)
	Update(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error)
}

type AddressRepository interface {
	Create(ctx context.Context, addr *domain.Address) (*domain.Address, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Address, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.Address, error)
	ListByBranch(ctx context.Context, branchID uuid.UUID) ([]*domain.Address, error)
}

type NumberSequenceRepository interface {
	Create(ctx context.Context, seq *domain.NumberSequence) (*domain.NumberSequence, error)
	GetByEntityAndCode(ctx context.Context, entityID uuid.UUID, code string) (*domain.NumberSequence, error)
	GetForUpdate(ctx context.Context, entityID uuid.UUID, code string) (*domain.NumberSequence, error)
	Increment(ctx context.Context, entityID uuid.UUID, code string, lastResetYear int32) (*domain.NumberSequence, error)
	ResetAndIncrement(ctx context.Context, entityID uuid.UUID, code string, currentYear int32) (*domain.NumberSequence, error)
	ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.NumberSequence, error)
}
