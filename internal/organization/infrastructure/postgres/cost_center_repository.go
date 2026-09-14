package postgres

import (
	"context"
	"errors"
	"fmt"

	"erp/internal/organization/domain"
	"erp/internal/organization/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CostCenterRepository struct {
	db *database.DB
}

func NewCostCenterRepository(db *database.DB) ports.CostCenterRepository {
	return &CostCenterRepository{db: db}
}

func (r *CostCenterRepository) Create(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateCostCenter(ctx, sqlc.CreateCostCenterParams{
		LegalEntityID: toPgUUID(cc.LegalEntityID),
		ParentID:      toPgUUIDPtr(cc.ParentID),
		Code:          cc.Code,
		NameAr:        cc.NameAr,
		NameEn:        cc.NameEn,
		IsActive:      cc.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cost center: %w", err)
	}

	return toDomainCostCenter(row), nil
}

func (r *CostCenterRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CostCenter, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetCostCenterByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCostCenterNotFound
		}
		return nil, fmt.Errorf("failed to get cost center by id: %w", err)
	}

	return toDomainCostCenter(row), nil
}

func (r *CostCenterRepository) ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.CostCenter, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListCostCentersByEntity(ctx, toPgUUID(entityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list cost centers: %w", err)
	}

	res := make([]*domain.CostCenter, len(rows))
	for i, row := range rows {
		res[i] = toDomainCostCenter(row)
	}
	return res, nil
}

func (r *CostCenterRepository) Update(ctx context.Context, cc *domain.CostCenter) (*domain.CostCenter, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateCostCenter(ctx, sqlc.UpdateCostCenterParams{
		ID:       toPgUUID(cc.ID),
		ParentID: toPgUUIDPtr(cc.ParentID),
		NameAr:   cc.NameAr,
		NameEn:   cc.NameEn,
		IsActive: cc.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCostCenterNotFound
		}
		return nil, fmt.Errorf("failed to update cost center: %w", err)
	}

	return toDomainCostCenter(row), nil
}

func toDomainCostCenter(row sqlc.OrganizationCostCenters) *domain.CostCenter {
	return &domain.CostCenter{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ParentID:      toUUIDPtr(row.ParentID),
		Code:          row.Code,
		NameAr:        row.NameAr,
		NameEn:        row.NameEn,
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}
}
