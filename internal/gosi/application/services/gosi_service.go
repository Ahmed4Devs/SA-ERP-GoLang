package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/gosi/domain"
	"erp/internal/gosi/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GosiService struct {
	repo ports.Repository
	db   *database.DB
}

func NewGosiService(repo ports.Repository, db *database.DB) ports.Service {
	return &GosiService{
		repo: repo,
		db:   db,
	}
}

// ----------------------------------------------------------------------
// 1. Contribution Rules
// ----------------------------------------------------------------------

func (s *GosiService) ListContributionRules(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ContributionRule, error) {
	return s.repo.ListContributionRules(ctx, legalEntityID)
}

func (s *GosiService) GetActiveRuleVersions(ctx context.Context, legalEntityID uuid.UUID, asOfDate time.Time) ([]domain.RuleVersion, error) {
	if asOfDate.IsZero() {
		asOfDate = time.Now()
	}
	return s.repo.GetActiveRuleVersions(ctx, legalEntityID, asOfDate)
}

// ----------------------------------------------------------------------
// 2. Profiles
// ----------------------------------------------------------------------

func (s *GosiService) RegisterEmployee(ctx context.Context, profile domain.InsuranceProfile) (*domain.InsuranceProfile, error) {
	// Check if an insurance profile already exists for this employee
	existing, err := s.repo.GetInsuranceProfileByEmployee(ctx, profile.LegalEntityID, profile.EmployeeID)
	if err == nil && existing != nil {
		// Update existing profile
		profile.ID = existing.ID
		return s.repo.UpdateInsuranceProfile(ctx, profile)
	}

	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	if profile.RegistrationDate == nil {
		now := time.Now()
		profile.RegistrationDate = &now
	}
	profile.IsActive = true

	return s.repo.CreateInsuranceProfile(ctx, profile)
}

func (s *GosiService) ListInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error) {
	return s.repo.ListInsuranceProfiles(ctx, legalEntityID)
}

// ----------------------------------------------------------------------
// 3. Calculations (Section 26 GOSIService interface)
// ----------------------------------------------------------------------

// Calculate calculates contribution amounts for a single employee using active database rule versions.
// Regulatory rates MUST come from versioned database rules — never hard-coded.
func (s *GosiService) Calculate(ctx context.Context, input domain.ContributionInput) ([]domain.SingleRuleResult, error) {
	asOfDate := input.PayrollDate
	if asOfDate.IsZero() {
		asOfDate = time.Now()
	}

	activeRules, err := s.repo.GetActiveRuleVersions(ctx, input.LegalEntityID, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active rule versions: %w", err)
	}
	if len(activeRules) == 0 {
		return nil, fmt.Errorf("no active GOSI rule versions found for legal entity %s as of %s", input.LegalEntityID, asOfDate.Format("2006-01-02"))
	}

	return domain.CalculateContribution(input, activeRules)
}

// RunMonthlyCalculation runs the GOSI calculation for all active enrolled employees for a given month.
func (s *GosiService) RunMonthlyCalculation(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error) {
	asOfDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	activeRules, err := s.repo.GetActiveRuleVersions(ctx, legalEntityID, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active rule versions: %w", err)
	}
	if len(activeRules) == 0 {
		return nil, fmt.Errorf("no active GOSI rule versions found for period %d-%02d", year, month)
	}

	profiles, err := s.repo.ListActiveInsuranceProfiles(ctx, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active insurance profiles: %w", err)
	}

	var calculations []domain.ContributionCalculation

	for _, profile := range profiles {
		// Determine contribution basis: override takes precedence, otherwise fallback to 10,000 SAR default
		basis := decimal.NewFromInt(10000)
		if profile.ContributionBasisOverride != nil && profile.ContributionBasisOverride.GreaterThan(decimal.Zero) {
			basis = *profile.ContributionBasisOverride
		}

		natCode := profile.NationalityCode
		if natCode == "" {
			if profile.NationalityScope == "SAUDI" {
				natCode = "SA"
			} else {
				natCode = "OTHER"
			}
		}

		input := domain.ContributionInput{
			LegalEntityID:     legalEntityID,
			EmployeeID:        profile.EmployeeID,
			PayrollDate:       asOfDate,
			NationalityCode:   natCode,
			ContributionBasis: basis,
		}

		lineResults, err := domain.CalculateContribution(input, activeRules)
		if err != nil {
			return nil, fmt.Errorf("failed calculating contribution for employee %s: %w", profile.EmployeeID, err)
		}

		totalEmployee := decimal.Zero
		totalEmployer := decimal.Zero

		for _, lr := range lineResults {
			totalEmployee = totalEmployee.Add(lr.EmployeeAmount)
			totalEmployer = totalEmployer.Add(lr.EmployerAmount)
		}
		totalAmount := totalEmployee.Add(totalEmployer)

		calc := domain.ContributionCalculation{
			ID:                  uuid.New(),
			LegalEntityID:       legalEntityID,
			EmployeeID:          profile.EmployeeID,
			ProfileID:           profile.ID,
			CalculationYear:     year,
			CalculationMonth:    month,
			ContributionBasis:   basis,
			TotalEmployeeAmount: totalEmployee,
			TotalEmployerAmount: totalEmployer,
			TotalAmount:         totalAmount,
			CalculationSnapshot: map[string]any{
				"nationality_scope": profile.NationalityScope,
				"nationality_code":  natCode,
				"rules_applied":     len(lineResults),
				"calculated_at":     time.Now().Format(time.RFC3339),
			},
			Status: "CALCULATED",
		}

		savedCalc, err := s.repo.CreateContributionCalculation(ctx, calc)
		if err != nil {
			return nil, fmt.Errorf("failed to persist contribution calculation: %w", err)
		}

		// Insert calculation lines
		for _, lr := range lineResults {
			line := domain.CalculationLine{
				ID:             uuid.New(),
				CalculationID:  savedCalc.ID,
				RuleVersionID:  lr.RuleVersion.ID,
				RuleCode:       lr.RuleVersion.RuleCode,
				RuleNameEn:     lr.RuleVersion.RuleNameEn,
				RuleNameAr:     lr.RuleVersion.RuleNameAr,
				BasisAmount:    lr.BasisApplied,
				EmployeeRate:   lr.RuleVersion.EmployeeRate,
				EmployerRate:   lr.RuleVersion.EmployerRate,
				EmployeeAmount: lr.EmployeeAmount,
				EmployerAmount: lr.EmployerAmount,
				FormulaNote:    lr.FormulaNote,
			}
			if _, err := s.repo.CreateCalculationLine(ctx, line); err != nil {
				return nil, fmt.Errorf("failed to persist calculation line: %w", err)
			}
		}

		savedCalc.FirstNameEn = profile.FirstNameEn
		savedCalc.LastNameEn = profile.LastNameEn
		savedCalc.FirstNameAr = profile.FirstNameAr
		savedCalc.LastNameAr = profile.LastNameAr
		savedCalc.EmployeeNumber = profile.EmployeeNumber
		calculations = append(calculations, *savedCalc)
	}

	return calculations, nil
}

func (s *GosiService) ListCalculationsByPeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error) {
	return s.repo.ListCalculationsByPeriod(ctx, legalEntityID, year, month)
}

func (s *GosiService) GetCalculationDetails(ctx context.Context, calcID uuid.UUID) (*domain.ContributionCalculation, []domain.CalculationLine, error) {
	calc, err := s.repo.GetContributionCalculation(ctx, calcID)
	if err != nil {
		return nil, nil, err
	}
	lines, err := s.repo.ListCalculationLines(ctx, calcID)
	if err != nil {
		return nil, nil, err
	}
	return calc, lines, nil
}

// ----------------------------------------------------------------------
// 4. Export Batches
// ----------------------------------------------------------------------

func (s *GosiService) CreateExportBatch(ctx context.Context, legalEntityID uuid.UUID, year, month int, createdBy *uuid.UUID) (*domain.ExportBatch, error) {
	calcs, err := s.repo.ListCalculationsByPeriod(ctx, legalEntityID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to list calculations for export batch: %w", err)
	}
	if len(calcs) == 0 {
		return nil, fmt.Errorf("no calculations found for period %d-%02d to export", year, month)
	}

	totalEmp := decimal.Zero
	totalEr := decimal.Zero
	for _, c := range calcs {
		totalEmp = totalEmp.Add(c.TotalEmployeeAmount)
		totalEr = totalEr.Add(c.TotalEmployerAmount)
	}
	totalAmount := totalEmp.Add(totalEr)

	batchNumber := identifiers.NewDocumentNumber(fmt.Sprintf("GOSI-%d%02d", year, month), time.Now())
	xmlContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<GOSIContributionReport xmlns="urn:sa:gosi:v1" year="%d" month="%02d">
  <LegalEntity id="%s"/>
  <Summary subscriberCount="%d" employeeTotal="%s" employerTotal="%s" grandTotal="%s"/>
  <GeneratedAt>%s</GeneratedAt>
</GOSIContributionReport>`, year, month, legalEntityID.String(), len(calcs), totalEmp.StringFixed(2), totalEr.StringFixed(2), totalAmount.StringFixed(2), time.Now().UTC().Format(time.RFC3339))

	batch := domain.ExportBatch{
		ID:                  uuid.New(),
		LegalEntityID:       legalEntityID,
		BatchNumber:         batchNumber,
		ExportYear:          year,
		ExportMonth:         month,
		EmployeeCount:       len(calcs),
		TotalEmployeeAmount: totalEmp,
		TotalEmployerAmount: totalEr,
		TotalAmount:         totalAmount,
		FileContent:         &xmlContent,
		FileFormat:          "XML",
		Status:              "GENERATED",
		CreatedBy:           createdBy,
	}

	savedBatch, err := s.repo.CreateExportBatch(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("failed to create export batch: %w", err)
	}

	// Link calculations to this export batch
	if err := s.repo.LinkCalculationsToExportBatch(ctx, savedBatch.ID, legalEntityID, year, month); err != nil {
		return nil, fmt.Errorf("failed to link calculations to export batch: %w", err)
	}

	// Record initial integration attempt
	_, _ = s.repo.CreateIntegrationAttempt(ctx, domain.IntegrationAttempt{
		ID:            uuid.New(),
		LegalEntityID: legalEntityID,
		BatchID:       savedBatch.ID,
		AttemptNumber: 1,
		AdapterType:   "FILE_EXPORT",
		RequestPayload: map[string]any{
			"batch_number": batchNumber,
			"year":         year,
			"month":        month,
			"subscribers":  len(calcs),
		},
		ResponsePayload: map[string]any{
			"status":   "SUCCESS",
			"format":   "XML",
			"filesize": len(xmlContent),
		},
		Status:    "SUCCESS",
		StartedAt: time.Now(),
	})

	return savedBatch, nil
}

func (s *GosiService) ListExportBatches(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ExportBatch, error) {
	return s.repo.ListExportBatches(ctx, legalEntityID)
}

// ----------------------------------------------------------------------
// 5. Dashboard Summary
// ----------------------------------------------------------------------

func (s *GosiService) GetDashboardSummary(ctx context.Context, legalEntityID uuid.UUID, year, month int) (*domain.DashboardSummary, error) {
	calcs, err := s.repo.ListCalculationsByPeriod(ctx, legalEntityID, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calculations for dashboard: %w", err)
	}

	activeProfiles, err := s.repo.ListActiveInsuranceProfiles(ctx, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active profiles: %w", err)
	}

	totalEmp := decimal.Zero
	totalEr := decimal.Zero

	for _, c := range calcs {
		totalEmp = totalEmp.Add(c.TotalEmployeeAmount)
		totalEr = totalEr.Add(c.TotalEmployerAmount)
	}

	complianceRate := "100%"
	if len(activeProfiles) > 0 {
		rate := decimal.NewFromInt(int64(len(calcs))).Div(decimal.NewFromInt(int64(len(activeProfiles)))).Mul(decimal.NewFromInt(100))
		if rate.GreaterThan(decimal.NewFromInt(100)) {
			rate = decimal.NewFromInt(100)
		}
		complianceRate = fmt.Sprintf("%s%%", rate.Round(1).String())
	}

	return &domain.DashboardSummary{
		TotalEmployeeContributions: totalEmp,
		TotalEmployerContributions: totalEr,
		TotalContributions:         totalEmp.Add(totalEr),
		SubscriberCount:            len(activeProfiles),
		ComplianceRate:             complianceRate,
	}, nil
}
