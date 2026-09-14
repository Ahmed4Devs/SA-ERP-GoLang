package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/organization/domain"
	"erp/internal/organization/ports"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
)

type NumberSequenceService struct {
	repo      ports.NumberSequenceRepository
	txManager transaction.Manager
}

func NewNumberSequenceService(repo ports.NumberSequenceRepository, txManager transaction.Manager) ports.NumberSequenceService {
	return &NumberSequenceService{
		repo:      repo,
		txManager: txManager,
	}
}

func (s *NumberSequenceService) DefineSequence(
	ctx context.Context,
	legalEntityID uuid.UUID,
	code, prefix, suffix string,
	startVal int64,
	step, padding int32,
	reset domain.ResetInterval,
) (*domain.NumberSequence, error) {
	if err := domain.ValidateCode(code); err != nil {
		return nil, err
	}
	if startVal <= 0 {
		startVal = 1
	}
	if step <= 0 {
		step = 1
	}
	if padding <= 0 {
		padding = 6
	}
	if reset == "" {
		reset = domain.ResetYearly
	}

	seq := &domain.NumberSequence{
		LegalEntityID: legalEntityID,
		SequenceCode:  strings.ToUpper(strings.TrimSpace(code)),
		Prefix:        strings.ToUpper(strings.TrimSpace(prefix)),
		Suffix:        strings.ToUpper(strings.TrimSpace(suffix)),
		NextValue:     startVal,
		Step:          step,
		PaddingLength: padding,
		ResetInterval: reset,
		LastResetYear: int32(time.Now().Year()),
	}

	return s.repo.Create(ctx, seq)
}

func (s *NumberSequenceService) GenerateNext(ctx context.Context, legalEntityID uuid.UUID, code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	var formattedNumber string

	err := s.txManager.WithinTx(ctx, func(txCtx context.Context) error {
		// 1. Pessimistic lock row for update
		seq, err := s.repo.GetForUpdate(txCtx, legalEntityID, code)
		if err != nil {
			return err
		}

		now := time.Now()
		currentYear := int32(now.Year())
		allocatedVal := seq.NextValue

		// 2. Check reset condition
		if seq.ResetInterval == domain.ResetYearly && seq.LastResetYear < currentYear {
			allocatedVal = 1
			_, err = s.repo.ResetAndIncrement(txCtx, legalEntityID, code, currentYear)
			if err != nil {
				return fmt.Errorf("failed to reset number sequence: %w", err)
			}
		} else {
			_, err = s.repo.Increment(txCtx, legalEntityID, code, currentYear)
			if err != nil {
				return fmt.Errorf("failed to increment number sequence: %w", err)
			}
		}

		formattedNumber = seq.FormatValue(allocatedVal, now)
		return nil
	})

	if err != nil {
		return "", err
	}

	return formattedNumber, nil
}

func (s *NumberSequenceService) ListSequences(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.NumberSequence, error) {
	return s.repo.ListByEntity(ctx, legalEntityID)
}
