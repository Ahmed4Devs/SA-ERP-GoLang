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

type LegalEntityRepository struct {
	db *database.DB
}

func NewLegalEntityRepository(db *database.DB) ports.LegalEntityRepository {
	return &LegalEntityRepository{db: db}
}

func (r *LegalEntityRepository) Create(ctx context.Context, entity *domain.LegalEntity) (*domain.LegalEntity, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateLegalEntity(ctx, sqlc.CreateLegalEntityParams{
		OrganizationID: toPgUUID(entity.OrganizationID),
		Code:           entity.Code,
		NameAr:         entity.NameAr,
		NameEn:         entity.NameEn,
		CrNumber:       toPgText(entity.CRNumber),
		BaseCurrency:   entity.BaseCurrency,
		IsActive:       entity.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create legal entity: %w", err)
	}

	return toDomainLegalEntity(row), nil
}

func (r *LegalEntityRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.LegalEntity, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetLegalEntityByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLegalEntityNotFound
		}
		return nil, fmt.Errorf("failed to get legal entity by id: %w", err)
	}

	return toDomainLegalEntity(row), nil
}

func (r *LegalEntityRepository) GetByOrgAndCode(ctx context.Context, orgID uuid.UUID, code string) (*domain.LegalEntity, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetLegalEntityByOrgAndCode(ctx, sqlc.GetLegalEntityByOrgAndCodeParams{
		OrganizationID: toPgUUID(orgID),
		Code:           code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLegalEntityNotFound
		}
		return nil, fmt.Errorf("failed to get legal entity by org and code: %w", err)
	}

	return toDomainLegalEntity(row), nil
}

func (r *LegalEntityRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.LegalEntity, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListLegalEntitiesByOrg(ctx, toPgUUID(orgID))
	if err != nil {
		return nil, fmt.Errorf("failed to list legal entities: %w", err)
	}

	res := make([]*domain.LegalEntity, len(rows))
	for i, row := range rows {
		res[i] = toDomainLegalEntity(row)
	}
	return res, nil
}

func (r *LegalEntityRepository) ListActive(ctx context.Context) ([]*domain.LegalEntity, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListAllActiveLegalEntities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active legal entities: %w", err)
	}

	res := make([]*domain.LegalEntity, len(rows))
	for i, row := range rows {
		res[i] = toDomainLegalEntity(row)
	}
	return res, nil
}

func (r *LegalEntityRepository) Update(ctx context.Context, entity *domain.LegalEntity) (*domain.LegalEntity, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateLegalEntity(ctx, sqlc.UpdateLegalEntityParams{
		ID:           toPgUUID(entity.ID),
		NameAr:       entity.NameAr,
		NameEn:       entity.NameEn,
		CrNumber:     toPgText(entity.CRNumber),
		BaseCurrency: entity.BaseCurrency,
		IsActive:     entity.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLegalEntityNotFound
		}
		return nil, fmt.Errorf("failed to update legal entity: %w", err)
	}

	return toDomainLegalEntity(row), nil
}

func toDomainLegalEntity(row sqlc.OrganizationLegalEntities) *domain.LegalEntity {
	return &domain.LegalEntity{
		ID:             toUUID(row.ID),
		OrganizationID: toUUID(row.OrganizationID),
		Code:           row.Code,
		NameAr:         row.NameAr,
		NameEn:         row.NameEn,
		CRNumber:       fromPgText(row.CrNumber),
		BaseCurrency:   row.BaseCurrency,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Version:        row.Version,
	}
}
