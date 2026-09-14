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

type OrganizationRepository struct {
	db *database.DB
}

func NewOrganizationRepository(db *database.DB) ports.OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
		Code:   org.Code,
		NameAr: org.NameAr,
		NameEn: org.NameEn,
		Status: string(org.Status),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return toDomainOrg(row), nil
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetOrganizationByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by id: %w", err)
	}

	return toDomainOrg(row), nil
}

func (r *OrganizationRepository) GetByCode(ctx context.Context, code string) (*domain.Organization, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetOrganizationByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by code: %w", err)
	}

	return toDomainOrg(row), nil
}

func (r *OrganizationRepository) List(ctx context.Context) ([]*domain.Organization, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	res := make([]*domain.Organization, len(rows))
	for i, row := range rows {
		res[i] = toDomainOrg(row)
	}
	return res, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateOrganization(ctx, sqlc.UpdateOrganizationParams{
		ID:     toPgUUID(org.ID),
		NameAr: org.NameAr,
		NameEn: org.NameEn,
		Status: string(org.Status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return toDomainOrg(row), nil
}

func toDomainOrg(row sqlc.OrganizationOrganizations) *domain.Organization {
	return &domain.Organization{
		ID:        toUUID(row.ID),
		Code:      row.Code,
		NameAr:    row.NameAr,
		NameEn:    row.NameEn,
		Status:    domain.OrganizationStatus(row.Status),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		Version:   row.Version,
	}
}
