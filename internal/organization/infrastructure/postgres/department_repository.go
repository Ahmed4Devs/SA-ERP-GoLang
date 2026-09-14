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

type DepartmentRepository struct {
	db *database.DB
}

func NewDepartmentRepository(db *database.DB) ports.DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) Create(ctx context.Context, d *domain.Department) (*domain.Department, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateDepartment(ctx, sqlc.CreateDepartmentParams{
		LegalEntityID: toPgUUID(d.LegalEntityID),
		ParentID:      toPgUUIDPtr(d.ParentID),
		Code:          d.Code,
		NameAr:        d.NameAr,
		NameEn:        d.NameEn,
		IsActive:      d.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return toDomainDepartment(row), nil
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Department, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetDepartmentByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDepartmentNotFound
		}
		return nil, fmt.Errorf("failed to get department by id: %w", err)
	}

	return toDomainDepartment(row), nil
}

func (r *DepartmentRepository) ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.Department, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListDepartmentsByEntity(ctx, toPgUUID(entityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	res := make([]*domain.Department, len(rows))
	for i, row := range rows {
		res[i] = toDomainDepartment(row)
	}
	return res, nil
}

func (r *DepartmentRepository) Update(ctx context.Context, d *domain.Department) (*domain.Department, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateDepartment(ctx, sqlc.UpdateDepartmentParams{
		ID:       toPgUUID(d.ID),
		ParentID: toPgUUIDPtr(d.ParentID),
		NameAr:   d.NameAr,
		NameEn:   d.NameEn,
		IsActive: d.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDepartmentNotFound
		}
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	return toDomainDepartment(row), nil
}

func toDomainDepartment(row sqlc.OrganizationDepartments) *domain.Department {
	return &domain.Department{
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
