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

type BranchRepository struct {
	db *database.DB
}

func NewBranchRepository(db *database.DB) ports.BranchRepository {
	return &BranchRepository{db: db}
}

func (r *BranchRepository) Create(ctx context.Context, b *domain.Branch) (*domain.Branch, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateBranch(ctx, sqlc.CreateBranchParams{
		LegalEntityID: toPgUUID(b.LegalEntityID),
		Code:          b.Code,
		NameAr:        b.NameAr,
		NameEn:        b.NameEn,
		IsActive:      b.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return toDomainBranch(row), nil
}

func (r *BranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Branch, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetBranchByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to get branch by id: %w", err)
	}

	return toDomainBranch(row), nil
}

func (r *BranchRepository) GetByEntityAndCode(ctx context.Context, entityID uuid.UUID, code string) (*domain.Branch, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetBranchByEntityAndCode(ctx, sqlc.GetBranchByEntityAndCodeParams{
		LegalEntityID: toPgUUID(entityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to get branch by entity and code: %w", err)
	}

	return toDomainBranch(row), nil
}

func (r *BranchRepository) ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.Branch, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListBranchesByEntity(ctx, toPgUUID(entityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	res := make([]*domain.Branch, len(rows))
	for i, row := range rows {
		res[i] = toDomainBranch(row)
	}
	return res, nil
}

func (r *BranchRepository) Update(ctx context.Context, b *domain.Branch) (*domain.Branch, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateBranch(ctx, sqlc.UpdateBranchParams{
		ID:       toPgUUID(b.ID),
		NameAr:   b.NameAr,
		NameEn:   b.NameEn,
		IsActive: b.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBranchNotFound
		}
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}

	return toDomainBranch(row), nil
}

func toDomainBranch(row sqlc.OrganizationBranches) *domain.Branch {
	return &domain.Branch{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameAr:        row.NameAr,
		NameEn:        row.NameEn,
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}
}
