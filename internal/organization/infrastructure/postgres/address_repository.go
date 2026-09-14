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

type AddressRepository struct {
	db *database.DB
}

func NewAddressRepository(db *database.DB) ports.AddressRepository {
	return &AddressRepository{db: db}
}

func (r *AddressRepository) Create(ctx context.Context, addr *domain.Address) (*domain.Address, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateAddress(ctx, sqlc.CreateAddressParams{
		LegalEntityID:    toPgUUIDPtr(addr.LegalEntityID),
		BranchID:         toPgUUIDPtr(addr.BranchID),
		AddressType:      addr.AddressType,
		BuildingNumber:   addr.BuildingNumber,
		StreetNameAr:     addr.StreetNameAr,
		StreetNameEn:     addr.StreetNameEn,
		DistrictAr:       addr.DistrictAr,
		DistrictEn:       addr.DistrictEn,
		CityAr:           addr.CityAr,
		CityEn:           addr.CityEn,
		PostalCode:       addr.PostalCode,
		AdditionalNumber: toPgText(addr.AdditionalNumber),
		CountryCode:      addr.CountryCode,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return toDomainAddress(row), nil
}

func (r *AddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Address, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetAddressByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidNationalAddress
		}
		return nil, fmt.Errorf("failed to get address by id: %w", err)
	}

	return toDomainAddress(row), nil
}

func (r *AddressRepository) ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.Address, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListAddressesByLegalEntity(ctx, toPgUUID(entityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}

	res := make([]*domain.Address, len(rows))
	for i, row := range rows {
		res[i] = toDomainAddress(row)
	}
	return res, nil
}

func (r *AddressRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]*domain.Address, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListAddressesByBranch(ctx, toPgUUID(branchID))
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses by branch: %w", err)
	}

	res := make([]*domain.Address, len(rows))
	for i, row := range rows {
		res[i] = toDomainAddress(row)
	}
	return res, nil
}

func toDomainAddress(row sqlc.OrganizationAddresses) *domain.Address {
	return &domain.Address{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUIDPtr(row.LegalEntityID),
		BranchID:         toUUIDPtr(row.BranchID),
		AddressType:      row.AddressType,
		BuildingNumber:   row.BuildingNumber,
		StreetNameAr:     row.StreetNameAr,
		StreetNameEn:     row.StreetNameEn,
		DistrictAr:       row.DistrictAr,
		DistrictEn:       row.DistrictEn,
		CityAr:           row.CityAr,
		CityEn:           row.CityEn,
		PostalCode:       row.PostalCode,
		AdditionalNumber: fromPgText(row.AdditionalNumber),
		CountryCode:      row.CountryCode,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
