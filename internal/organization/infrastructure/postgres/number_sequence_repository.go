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

type NumberSequenceRepository struct {
	db *database.DB
}

func NewNumberSequenceRepository(db *database.DB) ports.NumberSequenceRepository {
	return &NumberSequenceRepository{db: db}
}

func (r *NumberSequenceRepository) Create(ctx context.Context, seq *domain.NumberSequence) (*domain.NumberSequence, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateNumberSequence(ctx, sqlc.CreateNumberSequenceParams{
		LegalEntityID: toPgUUID(seq.LegalEntityID),
		SequenceCode:  seq.SequenceCode,
		Prefix:        seq.Prefix,
		Suffix:        seq.Suffix,
		NextValue:     seq.NextValue,
		Step:          seq.Step,
		PaddingLength: seq.PaddingLength,
		ResetInterval: string(seq.ResetInterval),
		LastResetYear: seq.LastResetYear,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create number sequence: %w", err)
	}

	return toDomainNumberSequence(row), nil
}

func (r *NumberSequenceRepository) GetByEntityAndCode(ctx context.Context, entityID uuid.UUID, code string) (*domain.NumberSequence, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetNumberSequenceByEntityAndCode(ctx, sqlc.GetNumberSequenceByEntityAndCodeParams{
		LegalEntityID: toPgUUID(entityID),
		SequenceCode:  code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNumberSequenceNotFound
		}
		return nil, fmt.Errorf("failed to get number sequence: %w", err)
	}

	return toDomainNumberSequence(row), nil
}

func (r *NumberSequenceRepository) GetForUpdate(ctx context.Context, entityID uuid.UUID, code string) (*domain.NumberSequence, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetNumberSequenceForUpdate(ctx, sqlc.GetNumberSequenceForUpdateParams{
		LegalEntityID: toPgUUID(entityID),
		SequenceCode:  code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNumberSequenceNotFound
		}
		return nil, fmt.Errorf("failed to lock number sequence for update: %w", err)
	}

	return toDomainNumberSequence(row), nil
}

func (r *NumberSequenceRepository) Increment(ctx context.Context, entityID uuid.UUID, code string, lastResetYear int32) (*domain.NumberSequence, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.IncrementNumberSequence(ctx, sqlc.IncrementNumberSequenceParams{
		LegalEntityID: toPgUUID(entityID),
		SequenceCode:  code,
		LastResetYear: lastResetYear,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNumberSequenceNotFound
		}
		return nil, fmt.Errorf("failed to increment number sequence: %w", err)
	}

	return toDomainNumberSequence(row), nil
}

func (r *NumberSequenceRepository) ResetAndIncrement(ctx context.Context, entityID uuid.UUID, code string, currentYear int32) (*domain.NumberSequence, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.ResetAndIncrementNumberSequence(ctx, sqlc.ResetAndIncrementNumberSequenceParams{
		LegalEntityID: toPgUUID(entityID),
		SequenceCode:  code,
		LastResetYear: currentYear,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNumberSequenceNotFound
		}
		return nil, fmt.Errorf("failed to reset and increment number sequence: %w", err)
	}

	return toDomainNumberSequence(row), nil
}

func (r *NumberSequenceRepository) ListByEntity(ctx context.Context, entityID uuid.UUID) ([]*domain.NumberSequence, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	rows, err := queries.ListNumberSequencesByEntity(ctx, toPgUUID(entityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list number sequences: %w", err)
	}

	res := make([]*domain.NumberSequence, len(rows))
	for i, row := range rows {
		res[i] = toDomainNumberSequence(row)
	}
	return res, nil
}

func toDomainNumberSequence(row sqlc.OrganizationNumberSequences) *domain.NumberSequence {
	return &domain.NumberSequence{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		SequenceCode:  row.SequenceCode,
		Prefix:        row.Prefix,
		Suffix:        row.Suffix,
		NextValue:     row.NextValue,
		Step:          row.Step,
		PaddingLength: row.PaddingLength,
		ResetInterval: domain.ResetInterval(row.ResetInterval),
		LastResetYear: row.LastResetYear,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		Version:       row.Version,
	}
}
