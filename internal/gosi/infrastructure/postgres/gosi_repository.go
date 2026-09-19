package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"erp/internal/gosi/domain"
	"erp/internal/gosi/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.DB
	q  *sqlc.Queries
}

func NewRepository(db *database.DB) ports.Repository {
	return &Repository{
		db: db,
		q:  sqlc.New(db.Pool),
	}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// ----------------------------------------------------------------------
// 1. Contribution Rules
// ----------------------------------------------------------------------

func (r *Repository) CreateContributionRule(ctx context.Context, rule domain.ContributionRule) (*domain.ContributionRule, error) {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}
	row, err := r.q.CreateContributionRule(ctx, sqlc.CreateContributionRuleParams{
		ID:            toPgUUID(rule.ID),
		LegalEntityID: toPgUUID(rule.LegalEntityID),
		Code:          rule.Code,
		NameEn:        rule.NameEn,
		NameAr:        rule.NameAr,
		DescriptionEn: toPgText(rule.DescriptionEn),
		DescriptionAr: toPgText(rule.DescriptionAr),
		IsActive:      rule.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create contribution rule: %w", err)
	}

	return &domain.ContributionRule{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		DescriptionEn: fromPgText(row.DescriptionEn),
		DescriptionAr: fromPgText(row.DescriptionAr),
		IsActive:      row.IsActive,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:     fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) ListContributionRules(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ContributionRule, error) {
	rows, err := r.q.ListContributionRules(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list contribution rules: %w", err)
	}

	rules := make([]domain.ContributionRule, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, domain.ContributionRule{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			DescriptionEn: fromPgText(row.DescriptionEn),
			DescriptionAr: fromPgText(row.DescriptionAr),
			IsActive:      row.IsActive,
			CreatedAt:     fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:     fromPgTimestamptz(row.UpdatedAt),
		})
	}
	return rules, nil
}

func (r *Repository) GetContributionRuleByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.ContributionRule, error) {
	row, err := r.q.GetContributionRuleByCode(ctx, sqlc.GetContributionRuleByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get contribution rule by code: %w", err)
	}

	return &domain.ContributionRule{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		DescriptionEn: fromPgText(row.DescriptionEn),
		DescriptionAr: fromPgText(row.DescriptionAr),
		IsActive:      row.IsActive,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:     fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

// ----------------------------------------------------------------------
// 2. Rule Versions
// ----------------------------------------------------------------------

func (r *Repository) CreateRuleVersion(ctx context.Context, rv domain.RuleVersion) (*domain.RuleVersion, error) {
	if rv.ID == uuid.Nil {
		rv.ID = uuid.New()
	}

	configBytes := []byte("{}")
	if rv.CalculationConfig != nil {
		var err error
		configBytes, err = json.Marshal(rv.CalculationConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal calculation config: %w", err)
		}
	}

	row, err := r.q.CreateRuleVersion(ctx, sqlc.CreateRuleVersionParams{
		ID:                toPgUUID(rv.ID),
		RuleID:            toPgUUID(rv.RuleID),
		LegalEntityID:     toPgUUID(rv.LegalEntityID),
		VersionNumber:     int32(rv.VersionNumber),
		EffectiveFrom:     toPgDate(rv.EffectiveFrom),
		EffectiveTo:       toPgDatePtr(rv.EffectiveTo),
		NationalityScope:  rv.NationalityScope,
		EmployeeRate:      toPgNumeric(rv.EmployeeRate),
		EmployerRate:      toPgNumeric(rv.EmployerRate),
		MinimumBasis:      toPgNumeric(rv.MinimumBasis),
		MaximumBasis:      toPgNumeric(rv.MaximumBasis),
		CalculationConfig: configBytes,
		LegalReference:    toPgText(rv.LegalReference),
		Notes:             toPgText(rv.Notes),
		IsActive:          rv.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create rule version: %w", err)
	}

	return &domain.RuleVersion{
		ID:                toUUID(row.ID),
		RuleID:            toUUID(row.RuleID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		VersionNumber:     int(row.VersionNumber),
		EffectiveFrom:     fromPgDate(row.EffectiveFrom),
		EffectiveTo:       fromPgDatePtr(row.EffectiveTo),
		NationalityScope:  row.NationalityScope,
		EmployeeRate:      fromPgNumeric(row.EmployeeRate),
		EmployerRate:      fromPgNumeric(row.EmployerRate),
		MinimumBasis:      fromPgNumeric(row.MinimumBasis),
		MaximumBasis:      fromPgNumeric(row.MaximumBasis),
		CalculationConfig: rv.CalculationConfig,
		LegalReference:    fromPgText(row.LegalReference),
		Notes:             fromPgText(row.Notes),
		IsActive:          row.IsActive,
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetActiveRuleVersions(ctx context.Context, legalEntityID uuid.UUID, asOfDate time.Time) ([]domain.RuleVersion, error) {
	rows, err := r.queries(ctx).GetActiveRuleVersions(ctx, sqlc.GetActiveRuleVersionsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Column2:       toPgDate(asOfDate),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active rule versions: %w", err)
	}

	result := make([]domain.RuleVersion, 0, len(rows))
	for _, row := range rows {
		var cfg map[string]any
		if len(row.CalculationConfig) > 0 {
			_ = json.Unmarshal(row.CalculationConfig, &cfg)
		}

		result = append(result, domain.RuleVersion{
			ID:                toUUID(row.ID),
			RuleID:            toUUID(row.RuleID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			VersionNumber:     int(row.VersionNumber),
			EffectiveFrom:     fromPgDate(row.EffectiveFrom),
			EffectiveTo:       fromPgDatePtr(row.EffectiveTo),
			NationalityScope:  row.NationalityScope,
			EmployeeRate:      fromPgNumeric(row.EmployeeRate),
			EmployerRate:      fromPgNumeric(row.EmployerRate),
			MinimumBasis:      fromPgNumeric(row.MinimumBasis),
			MaximumBasis:      fromPgNumeric(row.MaximumBasis),
			CalculationConfig: cfg,
			LegalReference:    fromPgText(row.LegalReference),
			Notes:             fromPgText(row.Notes),
			IsActive:          row.IsActive,
			CreatedAt:         fromPgTimestamptz(row.CreatedAt),
			RuleCode:          row.RuleCode,
			RuleNameEn:        row.RuleNameEn,
			RuleNameAr:        row.RuleNameAr,
		})
	}
	return result, nil
}

func (r *Repository) GetRuleVersionByID(ctx context.Context, id uuid.UUID) (*domain.RuleVersion, error) {
	row, err := r.q.GetRuleVersionByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get rule version by id: %w", err)
	}

	var cfg map[string]any
	if len(row.CalculationConfig) > 0 {
		_ = json.Unmarshal(row.CalculationConfig, &cfg)
	}

	return &domain.RuleVersion{
		ID:                toUUID(row.ID),
		RuleID:            toUUID(row.RuleID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		VersionNumber:     int(row.VersionNumber),
		EffectiveFrom:     fromPgDate(row.EffectiveFrom),
		EffectiveTo:       fromPgDatePtr(row.EffectiveTo),
		NationalityScope:  row.NationalityScope,
		EmployeeRate:      fromPgNumeric(row.EmployeeRate),
		EmployerRate:      fromPgNumeric(row.EmployerRate),
		MinimumBasis:      fromPgNumeric(row.MinimumBasis),
		MaximumBasis:      fromPgNumeric(row.MaximumBasis),
		CalculationConfig: cfg,
		LegalReference:    fromPgText(row.LegalReference),
		Notes:             fromPgText(row.Notes),
		IsActive:          row.IsActive,
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListRuleVersions(ctx context.Context, ruleID uuid.UUID) ([]domain.RuleVersion, error) {
	rows, err := r.q.ListRuleVersions(ctx, toPgUUID(ruleID))
	if err != nil {
		return nil, fmt.Errorf("failed to list rule versions: %w", err)
	}

	result := make([]domain.RuleVersion, 0, len(rows))
	for _, row := range rows {
		var cfg map[string]any
		if len(row.CalculationConfig) > 0 {
			_ = json.Unmarshal(row.CalculationConfig, &cfg)
		}

		result = append(result, domain.RuleVersion{
			ID:                toUUID(row.ID),
			RuleID:            toUUID(row.RuleID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			VersionNumber:     int(row.VersionNumber),
			EffectiveFrom:     fromPgDate(row.EffectiveFrom),
			EffectiveTo:       fromPgDatePtr(row.EffectiveTo),
			NationalityScope:  row.NationalityScope,
			EmployeeRate:      fromPgNumeric(row.EmployeeRate),
			EmployerRate:      fromPgNumeric(row.EmployerRate),
			MinimumBasis:      fromPgNumeric(row.MinimumBasis),
			MaximumBasis:      fromPgNumeric(row.MaximumBasis),
			CalculationConfig: cfg,
			LegalReference:    fromPgText(row.LegalReference),
			Notes:             fromPgText(row.Notes),
			IsActive:          row.IsActive,
			CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

// ----------------------------------------------------------------------
// 3. Profiles
// ----------------------------------------------------------------------

func (r *Repository) CreateInsuranceProfile(ctx context.Context, profile domain.InsuranceProfile) (*domain.InsuranceProfile, error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}

	row, err := r.q.CreateInsuranceProfile(ctx, sqlc.CreateInsuranceProfileParams{
		ID:                        toPgUUID(profile.ID),
		LegalEntityID:             toPgUUID(profile.LegalEntityID),
		EmployeeID:                toPgUUID(profile.EmployeeID),
		GosiNumber:                toPgText(profile.GosiNumber),
		RegistrationDate:          toPgDatePtr(profile.RegistrationDate),
		NationalityScope:          profile.NationalityScope,
		IsActive:                  profile.IsActive,
		ContributionBasisOverride: toPgNumericPtr(profile.ContributionBasisOverride),
		Notes:                     toPgText(profile.Notes),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create insurance profile: %w", err)
	}

	return &domain.InsuranceProfile{
		ID:                        toUUID(row.ID),
		LegalEntityID:             toUUID(row.LegalEntityID),
		EmployeeID:                toUUID(row.EmployeeID),
		GosiNumber:                fromPgText(row.GosiNumber),
		RegistrationDate:          fromPgDatePtr(row.RegistrationDate),
		DeregistrationDate:        fromPgDatePtr(row.DeregistrationDate),
		NationalityScope:          row.NationalityScope,
		IsActive:                  row.IsActive,
		ContributionBasisOverride: fromPgNumericPtr(row.ContributionBasisOverride),
		Notes:                     fromPgText(row.Notes),
		CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                 fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) GetInsuranceProfileByEmployee(ctx context.Context, legalEntityID, employeeID uuid.UUID) (*domain.InsuranceProfile, error) {
	row, err := r.q.GetInsuranceProfileByEmployee(ctx, sqlc.GetInsuranceProfileByEmployeeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		EmployeeID:    toPgUUID(employeeID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInsuranceProfileNotFound
		}
		return nil, fmt.Errorf("failed to get insurance profile: %w", err)
	}

	return &domain.InsuranceProfile{
		ID:                        toUUID(row.ID),
		LegalEntityID:             toUUID(row.LegalEntityID),
		EmployeeID:                toUUID(row.EmployeeID),
		GosiNumber:                fromPgText(row.GosiNumber),
		RegistrationDate:          fromPgDatePtr(row.RegistrationDate),
		DeregistrationDate:        fromPgDatePtr(row.DeregistrationDate),
		NationalityScope:          row.NationalityScope,
		IsActive:                  row.IsActive,
		ContributionBasisOverride: fromPgNumericPtr(row.ContributionBasisOverride),
		Notes:                     fromPgText(row.Notes),
		CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                 fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) ListInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error) {
	rows, err := r.q.ListInsuranceProfiles(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list insurance profiles: %w", err)
	}

	result := make([]domain.InsuranceProfile, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InsuranceProfile{
			ID:                        toUUID(row.ID),
			LegalEntityID:             toUUID(row.LegalEntityID),
			EmployeeID:                toUUID(row.EmployeeID),
			GosiNumber:                fromPgText(row.GosiNumber),
			RegistrationDate:          fromPgDatePtr(row.RegistrationDate),
			DeregistrationDate:        fromPgDatePtr(row.DeregistrationDate),
			NationalityScope:          row.NationalityScope,
			IsActive:                  row.IsActive,
			ContributionBasisOverride: fromPgNumericPtr(row.ContributionBasisOverride),
			Notes:                     fromPgText(row.Notes),
			CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                 fromPgTimestamptz(row.UpdatedAt),
			FirstNameEn:               row.FirstNameEn,
			LastNameEn:                row.LastNameEn,
			FirstNameAr:               row.FirstNameAr,
			LastNameAr:                row.LastNameAr,
			EmployeeNumber:            row.EmployeeNumber,
		})
	}
	return result, nil
}

func (r *Repository) UpdateInsuranceProfile(ctx context.Context, profile domain.InsuranceProfile) (*domain.InsuranceProfile, error) {
	row, err := r.q.UpdateInsuranceProfile(ctx, sqlc.UpdateInsuranceProfileParams{
		ID:                        toPgUUID(profile.ID),
		GosiNumber:                toPgText(profile.GosiNumber),
		IsActive:                  profile.IsActive,
		ContributionBasisOverride: toPgNumericPtr(profile.ContributionBasisOverride),
		DeregistrationDate:        toPgDatePtr(profile.DeregistrationDate),
		Notes:                     toPgText(profile.Notes),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update insurance profile: %w", err)
	}

	return &domain.InsuranceProfile{
		ID:                        toUUID(row.ID),
		LegalEntityID:             toUUID(row.LegalEntityID),
		EmployeeID:                toUUID(row.EmployeeID),
		GosiNumber:                fromPgText(row.GosiNumber),
		RegistrationDate:          fromPgDatePtr(row.RegistrationDate),
		DeregistrationDate:        fromPgDatePtr(row.DeregistrationDate),
		NationalityScope:          row.NationalityScope,
		IsActive:                  row.IsActive,
		ContributionBasisOverride: fromPgNumericPtr(row.ContributionBasisOverride),
		Notes:                     fromPgText(row.Notes),
		CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                 fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) ListActiveInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error) {
	rows, err := r.q.ListActiveInsuranceProfiles(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list active insurance profiles: %w", err)
	}

	result := make([]domain.InsuranceProfile, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InsuranceProfile{
			ID:                        toUUID(row.ID),
			LegalEntityID:             toUUID(row.LegalEntityID),
			EmployeeID:                toUUID(row.EmployeeID),
			GosiNumber:                fromPgText(row.GosiNumber),
			RegistrationDate:          fromPgDatePtr(row.RegistrationDate),
			DeregistrationDate:        fromPgDatePtr(row.DeregistrationDate),
			NationalityScope:          row.NationalityScope,
			IsActive:                  row.IsActive,
			ContributionBasisOverride: fromPgNumericPtr(row.ContributionBasisOverride),
			Notes:                     fromPgText(row.Notes),
			CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                 fromPgTimestamptz(row.UpdatedAt),
			FirstNameEn:               row.FirstNameEn,
			LastNameEn:                row.LastNameEn,
			FirstNameAr:               row.FirstNameAr,
			LastNameAr:                row.LastNameAr,
			EmployeeNumber:            row.EmployeeNumber,
			NationalityCode:           row.NationalityCountryCode,
		})
	}
	return result, nil
}

func (r *Repository) LockActiveInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error) {
	rows, err := r.queries(ctx).LockActiveInsuranceProfiles(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to lock active insurance profiles: %w", err)
	}

	result := make([]domain.InsuranceProfile, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.InsuranceProfile{
			ID:                        toUUID(row.ID),
			LegalEntityID:             toUUID(row.LegalEntityID),
			EmployeeID:                toUUID(row.EmployeeID),
			GosiNumber:                fromPgText(row.GosiNumber),
			RegistrationDate:          fromPgDatePtr(row.RegistrationDate),
			DeregistrationDate:        fromPgDatePtr(row.DeregistrationDate),
			NationalityScope:          row.NationalityScope,
			IsActive:                  row.IsActive,
			ContributionBasisOverride: fromPgNumericPtr(row.ContributionBasisOverride),
			Notes:                     fromPgText(row.Notes),
			CreatedAt:                 fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                 fromPgTimestamptz(row.UpdatedAt),
			FirstNameEn:               row.FirstNameEn,
			LastNameEn:                row.LastNameEn,
			FirstNameAr:               row.FirstNameAr,
			LastNameAr:                row.LastNameAr,
			EmployeeNumber:            row.EmployeeNumber,
			NationalityCode:           row.NationalityCountryCode,
		})
	}
	return result, nil
}

func (r *Repository) EmployeeExistsInLegalEntity(ctx context.Context, legalEntityID, employeeID uuid.UUID) (bool, error) {
	exists, err := r.queries(ctx).EmployeeExistsInLegalEntity(ctx, sqlc.EmployeeExistsInLegalEntityParams{
		ID:            toPgUUID(employeeID),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to verify employee legal entity: %w", err)
	}
	return exists, nil
}

// ----------------------------------------------------------------------
// 4. Calculations
// ----------------------------------------------------------------------

func (r *Repository) CreateContributionCalculation(ctx context.Context, calc domain.ContributionCalculation) (*domain.ContributionCalculation, error) {
	if calc.ID == uuid.Nil {
		calc.ID = uuid.New()
	}

	snapBytes := []byte("{}")
	if calc.CalculationSnapshot != nil {
		var err error
		snapBytes, err = json.Marshal(calc.CalculationSnapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal calculation snapshot: %w", err)
		}
	}

	row, err := r.queries(ctx).CreateContributionCalculation(ctx, sqlc.CreateContributionCalculationParams{
		ID:                  toPgUUID(calc.ID),
		LegalEntityID:       toPgUUID(calc.LegalEntityID),
		EmployeeID:          toPgUUID(calc.EmployeeID),
		ProfileID:           toPgUUID(calc.ProfileID),
		CalculationYear:     int32(calc.CalculationYear),
		CalculationMonth:    int32(calc.CalculationMonth),
		ContributionBasis:   toPgNumeric(calc.ContributionBasis),
		TotalEmployeeAmount: toPgNumeric(calc.TotalEmployeeAmount),
		TotalEmployerAmount: toPgNumeric(calc.TotalEmployerAmount),
		TotalAmount:         toPgNumeric(calc.TotalAmount),
		CalculationSnapshot: snapBytes,
		Status:              calc.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create contribution calculation: %w", err)
	}

	return &domain.ContributionCalculation{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		EmployeeID:          toUUID(row.EmployeeID),
		ProfileID:           toUUID(row.ProfileID),
		CalculationYear:     int(row.CalculationYear),
		CalculationMonth:    int(row.CalculationMonth),
		ContributionBasis:   fromPgNumeric(row.ContributionBasis),
		TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
		TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
		TotalAmount:         fromPgNumeric(row.TotalAmount),
		CalculationSnapshot: calc.CalculationSnapshot,
		ExportBatchID:       toUUIDPtr(row.ExportBatchID),
		Status:              row.Status,
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) CreateCalculationLine(ctx context.Context, line domain.CalculationLine) (*domain.CalculationLine, error) {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateGosiCalculationLine(ctx, sqlc.CreateGosiCalculationLineParams{
		ID:             toPgUUID(line.ID),
		CalculationID:  toPgUUID(line.CalculationID),
		RuleVersionID:  toPgUUID(line.RuleVersionID),
		RuleCode:       line.RuleCode,
		RuleNameEn:     toPgText(line.RuleNameEn),
		RuleNameAr:     toPgText(line.RuleNameAr),
		BasisAmount:    toPgNumeric(line.BasisAmount),
		EmployeeRate:   toPgNumeric(line.EmployeeRate),
		EmployerRate:   toPgNumeric(line.EmployerRate),
		EmployeeAmount: toPgNumeric(line.EmployeeAmount),
		EmployerAmount: toPgNumeric(line.EmployerAmount),
		FormulaNote:    toPgText(line.FormulaNote),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create calculation line: %w", err)
	}

	return &domain.CalculationLine{
		ID:             toUUID(row.ID),
		CalculationID:  toUUID(row.CalculationID),
		RuleVersionID:  toUUID(row.RuleVersionID),
		RuleCode:       row.RuleCode,
		RuleNameEn:     fromPgText(row.RuleNameEn),
		RuleNameAr:     fromPgText(row.RuleNameAr),
		BasisAmount:    fromPgNumeric(row.BasisAmount),
		EmployeeRate:   fromPgNumeric(row.EmployeeRate),
		EmployerRate:   fromPgNumeric(row.EmployerRate),
		EmployeeAmount: fromPgNumeric(row.EmployeeAmount),
		EmployerAmount: fromPgNumeric(row.EmployerAmount),
		FormulaNote:    fromPgText(row.FormulaNote),
		CreatedAt:      fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) DeleteCalculationLines(ctx context.Context, calculationID uuid.UUID) error {
	if err := r.queries(ctx).DeleteGosiCalculationLines(ctx, toPgUUID(calculationID)); err != nil {
		return fmt.Errorf("failed to replace calculation lines: %w", err)
	}
	return nil
}

func (r *Repository) GetContributionCalculation(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.ContributionCalculation, error) {
	row, err := r.queries(ctx).GetGosiContributionCalculation(ctx, sqlc.GetGosiContributionCalculationParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCalculationNotFound
		}
		return nil, fmt.Errorf("failed to get contribution calculation: %w", err)
	}

	var snap map[string]any
	if len(row.CalculationSnapshot) > 0 {
		_ = json.Unmarshal(row.CalculationSnapshot, &snap)
	}

	return &domain.ContributionCalculation{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		EmployeeID:          toUUID(row.EmployeeID),
		ProfileID:           toUUID(row.ProfileID),
		CalculationYear:     int(row.CalculationYear),
		CalculationMonth:    int(row.CalculationMonth),
		ContributionBasis:   fromPgNumeric(row.ContributionBasis),
		TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
		TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
		TotalAmount:         fromPgNumeric(row.TotalAmount),
		CalculationSnapshot: snap,
		ExportBatchID:       toUUIDPtr(row.ExportBatchID),
		Status:              row.Status,
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListCalculationsByPeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error) {
	rows, err := r.queries(ctx).ListGosiCalculationsByPeriod(ctx, sqlc.ListGosiCalculationsByPeriodParams{
		LegalEntityID:    toPgUUID(legalEntityID),
		CalculationYear:  int32(year),
		CalculationMonth: int32(month),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list calculations by period: %w", err)
	}

	result := make([]domain.ContributionCalculation, 0, len(rows))
	for _, row := range rows {
		var snap map[string]any
		if len(row.CalculationSnapshot) > 0 {
			_ = json.Unmarshal(row.CalculationSnapshot, &snap)
		}

		result = append(result, domain.ContributionCalculation{
			ID:                  toUUID(row.ID),
			LegalEntityID:       toUUID(row.LegalEntityID),
			EmployeeID:          toUUID(row.EmployeeID),
			ProfileID:           toUUID(row.ProfileID),
			CalculationYear:     int(row.CalculationYear),
			CalculationMonth:    int(row.CalculationMonth),
			ContributionBasis:   fromPgNumeric(row.ContributionBasis),
			TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
			TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
			TotalAmount:         fromPgNumeric(row.TotalAmount),
			CalculationSnapshot: snap,
			ExportBatchID:       toUUIDPtr(row.ExportBatchID),
			Status:              row.Status,
			CreatedAt:           fromPgTimestamptz(row.CreatedAt),
			FirstNameEn:         row.FirstNameEn,
			LastNameEn:          row.LastNameEn,
			FirstNameAr:         row.FirstNameAr,
			LastNameAr:          row.LastNameAr,
			EmployeeNumber:      row.EmployeeNumber,
		})
	}
	return result, nil
}

func (r *Repository) LockCalculationsByPeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error) {
	rows, err := r.queries(ctx).LockGosiCalculationsByPeriod(ctx, sqlc.LockGosiCalculationsByPeriodParams{
		LegalEntityID:    toPgUUID(legalEntityID),
		CalculationYear:  int32(year),
		CalculationMonth: int32(month),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to lock calculations by period: %w", err)
	}

	result := make([]domain.ContributionCalculation, 0, len(rows))
	for _, row := range rows {
		var snapshot map[string]any
		if len(row.CalculationSnapshot) > 0 {
			_ = json.Unmarshal(row.CalculationSnapshot, &snapshot)
		}
		result = append(result, domain.ContributionCalculation{
			ID:                  toUUID(row.ID),
			LegalEntityID:       toUUID(row.LegalEntityID),
			EmployeeID:          toUUID(row.EmployeeID),
			ProfileID:           toUUID(row.ProfileID),
			CalculationYear:     int(row.CalculationYear),
			CalculationMonth:    int(row.CalculationMonth),
			ContributionBasis:   fromPgNumeric(row.ContributionBasis),
			TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
			TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
			TotalAmount:         fromPgNumeric(row.TotalAmount),
			CalculationSnapshot: snapshot,
			ExportBatchID:       toUUIDPtr(row.ExportBatchID),
			Status:              row.Status,
			CreatedAt:           fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) ListCalculationLines(ctx context.Context, calcID uuid.UUID) ([]domain.CalculationLine, error) {
	rows, err := r.q.ListGosiCalculationLines(ctx, toPgUUID(calcID))
	if err != nil {
		return nil, fmt.Errorf("failed to list calculation lines: %w", err)
	}

	result := make([]domain.CalculationLine, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.CalculationLine{
			ID:             toUUID(row.ID),
			CalculationID:  toUUID(row.CalculationID),
			RuleVersionID:  toUUID(row.RuleVersionID),
			RuleCode:       row.RuleCode,
			RuleNameEn:     fromPgText(row.RuleNameEn),
			RuleNameAr:     fromPgText(row.RuleNameAr),
			BasisAmount:    fromPgNumeric(row.BasisAmount),
			EmployeeRate:   fromPgNumeric(row.EmployeeRate),
			EmployerRate:   fromPgNumeric(row.EmployerRate),
			EmployeeAmount: fromPgNumeric(row.EmployeeAmount),
			EmployerAmount: fromPgNumeric(row.EmployerAmount),
			FormulaNote:    fromPgText(row.FormulaNote),
			CreatedAt:      fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

// ----------------------------------------------------------------------
// 5. Export Batches
// ----------------------------------------------------------------------

func (r *Repository) CreateExportBatch(ctx context.Context, batch domain.ExportBatch) (*domain.ExportBatch, error) {
	if batch.ID == uuid.Nil {
		batch.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateExportBatch(ctx, sqlc.CreateExportBatchParams{
		ID:                  toPgUUID(batch.ID),
		LegalEntityID:       toPgUUID(batch.LegalEntityID),
		BatchNumber:         batch.BatchNumber,
		ExportYear:          int32(batch.ExportYear),
		ExportMonth:         int32(batch.ExportMonth),
		EmployeeCount:       int32(batch.EmployeeCount),
		TotalEmployeeAmount: toPgNumeric(batch.TotalEmployeeAmount),
		TotalEmployerAmount: toPgNumeric(batch.TotalEmployerAmount),
		TotalAmount:         toPgNumeric(batch.TotalAmount),
		FileContent:         toPgTextPtr(batch.FileContent),
		FileFormat:          batch.FileFormat,
		Status:              batch.Status,
		CreatedBy:           toPgUUIDPtr(batch.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create export batch: %w", err)
	}

	return &domain.ExportBatch{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		BatchNumber:         row.BatchNumber,
		ExportYear:          int(row.ExportYear),
		ExportMonth:         int(row.ExportMonth),
		EmployeeCount:       int(row.EmployeeCount),
		TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
		TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
		TotalAmount:         fromPgNumeric(row.TotalAmount),
		FileContent:         fromPgTextPtr(row.FileContent),
		FileFormat:          row.FileFormat,
		Status:              row.Status,
		SubmittedAt:         fromPgTimestamptzPtr(row.SubmittedAt),
		ConfirmedAt:         fromPgTimestamptzPtr(row.ConfirmedAt),
		CreatedBy:           toUUIDPtr(row.CreatedBy),
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetExportBatchByID(ctx context.Context, id uuid.UUID) (*domain.ExportBatch, error) {
	row, err := r.q.GetExportBatchByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get export batch by id: %w", err)
	}

	return &domain.ExportBatch{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		BatchNumber:         row.BatchNumber,
		ExportYear:          int(row.ExportYear),
		ExportMonth:         int(row.ExportMonth),
		EmployeeCount:       int(row.EmployeeCount),
		TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
		TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
		TotalAmount:         fromPgNumeric(row.TotalAmount),
		FileContent:         fromPgTextPtr(row.FileContent),
		FileFormat:          row.FileFormat,
		Status:              row.Status,
		SubmittedAt:         fromPgTimestamptzPtr(row.SubmittedAt),
		ConfirmedAt:         fromPgTimestamptzPtr(row.ConfirmedAt),
		CreatedBy:           toUUIDPtr(row.CreatedBy),
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListExportBatches(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ExportBatch, error) {
	rows, err := r.q.ListExportBatches(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list export batches: %w", err)
	}

	result := make([]domain.ExportBatch, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.ExportBatch{
			ID:                  toUUID(row.ID),
			LegalEntityID:       toUUID(row.LegalEntityID),
			BatchNumber:         row.BatchNumber,
			ExportYear:          int(row.ExportYear),
			ExportMonth:         int(row.ExportMonth),
			EmployeeCount:       int(row.EmployeeCount),
			TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
			TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
			TotalAmount:         fromPgNumeric(row.TotalAmount),
			FileContent:         fromPgTextPtr(row.FileContent),
			FileFormat:          row.FileFormat,
			Status:              row.Status,
			SubmittedAt:         fromPgTimestamptzPtr(row.SubmittedAt),
			ConfirmedAt:         fromPgTimestamptzPtr(row.ConfirmedAt),
			CreatedBy:           toUUIDPtr(row.CreatedBy),
			CreatedAt:           fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) UpdateExportBatchStatus(ctx context.Context, id uuid.UUID, status string) (*domain.ExportBatch, error) {
	row, err := r.q.UpdateExportBatchStatus(ctx, sqlc.UpdateExportBatchStatusParams{
		ID:     toPgUUID(id),
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update export batch status: %w", err)
	}

	return &domain.ExportBatch{
		ID:                  toUUID(row.ID),
		LegalEntityID:       toUUID(row.LegalEntityID),
		BatchNumber:         row.BatchNumber,
		ExportYear:          int(row.ExportYear),
		ExportMonth:         int(row.ExportMonth),
		EmployeeCount:       int(row.EmployeeCount),
		TotalEmployeeAmount: fromPgNumeric(row.TotalEmployeeAmount),
		TotalEmployerAmount: fromPgNumeric(row.TotalEmployerAmount),
		TotalAmount:         fromPgNumeric(row.TotalAmount),
		FileContent:         fromPgTextPtr(row.FileContent),
		FileFormat:          row.FileFormat,
		Status:              row.Status,
		SubmittedAt:         fromPgTimestamptzPtr(row.SubmittedAt),
		ConfirmedAt:         fromPgTimestamptzPtr(row.ConfirmedAt),
		CreatedBy:           toUUIDPtr(row.CreatedBy),
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) LinkCalculationsToExportBatch(ctx context.Context, batchID, legalEntityID uuid.UUID, year, month int) (int64, error) {
	count, err := r.queries(ctx).LinkCalculationsToExportBatch(ctx, sqlc.LinkCalculationsToExportBatchParams{
		ExportBatchID:    toPgUUID(batchID),
		LegalEntityID:    toPgUUID(legalEntityID),
		CalculationYear:  int32(year),
		CalculationMonth: int32(month),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to link calculations to export batch: %w", err)
	}
	return count, nil
}

// ----------------------------------------------------------------------
// 6. Integration Attempts
// ----------------------------------------------------------------------

func (r *Repository) CreateIntegrationAttempt(ctx context.Context, attempt domain.IntegrationAttempt) (*domain.IntegrationAttempt, error) {
	if attempt.ID == uuid.Nil {
		attempt.ID = uuid.New()
	}

	reqBytes := []byte("{}")
	if attempt.RequestPayload != nil {
		var err error
		reqBytes, err = json.Marshal(attempt.RequestPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
	}

	respBytes := []byte("{}")
	if attempt.ResponsePayload != nil {
		var err error
		respBytes, err = json.Marshal(attempt.ResponsePayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response payload: %w", err)
		}
	}

	row, err := r.queries(ctx).CreateIntegrationAttempt(ctx, sqlc.CreateIntegrationAttemptParams{
		ID:              toPgUUID(attempt.ID),
		LegalEntityID:   toPgUUID(attempt.LegalEntityID),
		BatchID:         toPgUUID(attempt.BatchID),
		AttemptNumber:   int32(attempt.AttemptNumber),
		AdapterType:     attempt.AdapterType,
		RequestPayload:  reqBytes,
		ResponsePayload: respBytes,
		Status:          attempt.Status,
		ErrorMessage:    toPgText(attempt.ErrorMessage),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create integration attempt: %w", err)
	}

	return &domain.IntegrationAttempt{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		BatchID:         toUUID(row.BatchID),
		AttemptNumber:   int(row.AttemptNumber),
		AdapterType:     row.AdapterType,
		RequestPayload:  attempt.RequestPayload,
		ResponsePayload: attempt.ResponsePayload,
		Status:          row.Status,
		ErrorMessage:    fromPgText(row.ErrorMessage),
		StartedAt:       fromPgTimestamptz(row.StartedAt),
		CompletedAt:     fromPgTimestamptzPtr(row.CompletedAt),
	}, nil
}

func (r *Repository) ListIntegrationAttempts(ctx context.Context, batchID uuid.UUID) ([]domain.IntegrationAttempt, error) {
	rows, err := r.q.ListIntegrationAttempts(ctx, toPgUUID(batchID))
	if err != nil {
		return nil, fmt.Errorf("failed to list integration attempts: %w", err)
	}

	result := make([]domain.IntegrationAttempt, 0, len(rows))
	for _, row := range rows {
		var req, resp map[string]any
		if len(row.RequestPayload) > 0 {
			_ = json.Unmarshal(row.RequestPayload, &req)
		}
		if len(row.ResponsePayload) > 0 {
			_ = json.Unmarshal(row.ResponsePayload, &resp)
		}

		result = append(result, domain.IntegrationAttempt{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			BatchID:         toUUID(row.BatchID),
			AttemptNumber:   int(row.AttemptNumber),
			AdapterType:     row.AdapterType,
			RequestPayload:  req,
			ResponsePayload: resp,
			Status:          row.Status,
			ErrorMessage:    fromPgText(row.ErrorMessage),
			StartedAt:       fromPgTimestamptz(row.StartedAt),
			CompletedAt:     fromPgTimestamptzPtr(row.CompletedAt),
		})
	}
	return result, nil
}
