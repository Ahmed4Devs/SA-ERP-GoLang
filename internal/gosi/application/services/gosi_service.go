package services

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"erp/internal/gosi/domain"
	"erp/internal/gosi/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GosiService struct {
	repo ports.Repository
	tx   transaction.Manager
}

func NewGosiService(repo ports.Repository, db *database.DB) ports.Service {
	return &GosiService{
		repo: repo,
		tx:   transaction.NewManager(db.Pool),
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
	if profile.LegalEntityID == uuid.Nil || profile.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id and employee_id are required")
	}
	exists, err := s.repo.EmployeeExistsInLegalEntity(ctx, profile.LegalEntityID, profile.EmployeeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrEmployeeScopeMismatch
	}
	profile.NationalityScope = strings.ToUpper(strings.TrimSpace(profile.NationalityScope))
	if profile.NationalityScope == "" {
		profile.NationalityScope = "SAUDI"
	}
	if profile.NationalityScope != "SAUDI" && profile.NationalityScope != "NON_SAUDI" {
		return nil, fmt.Errorf("nationality_scope must be SAUDI or NON_SAUDI")
	}
	if profile.ContributionBasisOverride != nil && !profile.ContributionBasisOverride.IsPositive() {
		return nil, fmt.Errorf("contribution_basis_override must be greater than zero")
	}

	// Check if an insurance profile already exists for this employee
	existing, err := s.repo.GetInsuranceProfileByEmployee(ctx, profile.LegalEntityID, profile.EmployeeID)
	if err == nil && existing != nil {
		// Update existing profile
		profile.ID = existing.ID
		return s.repo.UpdateInsuranceProfile(ctx, profile)
	}
	if err != nil && !errors.Is(err, domain.ErrInsuranceProfileNotFound) {
		return nil, err
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
	if input.LegalEntityID == uuid.Nil || input.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id and employee_id are required")
	}
	if !input.ContributionBasis.IsPositive() {
		return nil, fmt.Errorf("contribution_basis must be greater than zero")
	}
	exists, err := s.repo.EmployeeExistsInLegalEntity(ctx, input.LegalEntityID, input.EmployeeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrEmployeeScopeMismatch
	}
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
	if legalEntityID == uuid.Nil || year < 2000 || year > 9999 || month < 1 || month > 12 {
		return nil, domain.ErrInvalidPeriod
	}
	asOfDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	var calculations []domain.ContributionCalculation
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		existingCalculations, err := s.repo.LockCalculationsByPeriod(txCtx, legalEntityID, year, month)
		if err != nil {
			return err
		}
		for _, calculation := range existingCalculations {
			if calculation.Status != "CALCULATED" || calculation.ExportBatchID != nil {
				return domain.ErrPeriodAlreadyExported
			}
		}

		activeRules, err := s.repo.GetActiveRuleVersions(txCtx, legalEntityID, asOfDate)
		if err != nil {
			return fmt.Errorf("failed to fetch active rule versions: %w", err)
		}
		if len(activeRules) == 0 {
			return fmt.Errorf("no active GOSI rule versions found for period %d-%02d", year, month)
		}

		profiles, err := s.repo.LockActiveInsuranceProfiles(txCtx, legalEntityID)
		if err != nil {
			return fmt.Errorf("failed to list active insurance profiles: %w", err)
		}
		calculations = make([]domain.ContributionCalculation, 0, len(profiles))

		for _, profile := range profiles {
			if profile.ContributionBasisOverride == nil || !profile.ContributionBasisOverride.IsPositive() {
				return fmt.Errorf("employee %s requires an explicit approved contribution basis before monthly calculation", profile.EmployeeID)
			}
			basis := *profile.ContributionBasisOverride

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
				return fmt.Errorf("failed calculating contribution for employee %s: %w", profile.EmployeeID, err)
			}

			totalEmployee := decimal.Zero
			totalEmployer := decimal.Zero
			for _, lineResult := range lineResults {
				totalEmployee = totalEmployee.Add(lineResult.EmployeeAmount)
				totalEmployer = totalEmployer.Add(lineResult.EmployerAmount)
			}

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
				TotalAmount:         totalEmployee.Add(totalEmployer),
				CalculationSnapshot: map[string]any{
					"nationality_scope": profile.NationalityScope,
					"nationality_code":  natCode,
					"rules_applied":     len(lineResults),
					"calculated_at":     time.Now().UTC().Format(time.RFC3339),
				},
				Status: "CALCULATED",
			}

			savedCalc, err := s.repo.CreateContributionCalculation(txCtx, calc)
			if err != nil {
				return fmt.Errorf("failed to persist contribution calculation: %w", err)
			}
			if err := s.repo.DeleteCalculationLines(txCtx, savedCalc.ID); err != nil {
				return err
			}

			for _, lineResult := range lineResults {
				line := domain.CalculationLine{
					ID:             uuid.New(),
					CalculationID:  savedCalc.ID,
					RuleVersionID:  lineResult.RuleVersion.ID,
					RuleCode:       lineResult.RuleVersion.RuleCode,
					RuleNameEn:     lineResult.RuleVersion.RuleNameEn,
					RuleNameAr:     lineResult.RuleVersion.RuleNameAr,
					BasisAmount:    lineResult.BasisApplied,
					EmployeeRate:   lineResult.RuleVersion.EmployeeRate,
					EmployerRate:   lineResult.RuleVersion.EmployerRate,
					EmployeeAmount: lineResult.EmployeeAmount,
					EmployerAmount: lineResult.EmployerAmount,
					FormulaNote:    lineResult.FormulaNote,
				}
				if _, err := s.repo.CreateCalculationLine(txCtx, line); err != nil {
					return fmt.Errorf("failed to persist calculation line: %w", err)
				}
			}

			savedCalc.FirstNameEn = profile.FirstNameEn
			savedCalc.LastNameEn = profile.LastNameEn
			savedCalc.FirstNameAr = profile.FirstNameAr
			savedCalc.LastNameAr = profile.LastNameAr
			savedCalc.EmployeeNumber = profile.EmployeeNumber
			calculations = append(calculations, *savedCalc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return calculations, nil
}

func (s *GosiService) ListCalculationsByPeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error) {
	return s.repo.ListCalculationsByPeriod(ctx, legalEntityID, year, month)
}

func (s *GosiService) GetCalculationDetails(ctx context.Context, legalEntityID, calcID uuid.UUID) (*domain.ContributionCalculation, []domain.CalculationLine, error) {
	if legalEntityID == uuid.Nil || calcID == uuid.Nil {
		return nil, nil, fmt.Errorf("legal_entity_id and calculation id are required")
	}
	calc, err := s.repo.GetContributionCalculation(ctx, legalEntityID, calcID)
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
	if legalEntityID == uuid.Nil || year < 2000 || year > 9999 || month < 1 || month > 12 {
		return nil, domain.ErrInvalidPeriod
	}

	var savedBatch *domain.ExportBatch
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		calcs, err := s.repo.LockCalculationsByPeriod(txCtx, legalEntityID, year, month)
		if err != nil {
			return fmt.Errorf("failed to lock calculations for export batch: %w", err)
		}
		if len(calcs) == 0 {
			return fmt.Errorf("no calculations found for period %d-%02d to export", year, month)
		}
		for _, calc := range calcs {
			if calc.Status != "CALCULATED" || calc.ExportBatchID != nil {
				return domain.ErrPeriodAlreadyExported
			}
		}

		totalEmp := decimal.Zero
		totalEr := decimal.Zero
		for _, calc := range calcs {
			totalEmp = totalEmp.Add(calc.TotalEmployeeAmount)
			totalEr = totalEr.Add(calc.TotalEmployerAmount)
		}
		totalAmount := totalEmp.Add(totalEr)
		generatedAt := time.Now().UTC()
		batchNumber := identifiers.NewDocumentNumber(fmt.Sprintf("GOSI-%d%02d", year, month), generatedAt)
		report := contributionReport{
			Year: year, Month: month, LegalEntity: reportEntity{ID: legalEntityID.String()},
			Summary:     reportSummary{SubscriberCount: len(calcs), EmployeeTotal: totalEmp.String(), EmployerTotal: totalEr.String(), GrandTotal: totalAmount.String()},
			GeneratedAt: generatedAt.Format(time.RFC3339),
		}
		for _, calc := range calcs {
			lines, err := s.repo.ListCalculationLines(txCtx, calc.ID)
			if err != nil {
				return fmt.Errorf("load contribution rule snapshots: %w", err)
			}
			report.Employees = append(report.Employees, employeeReport(calc, lines))
		}
		reportXML, err := xml.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("encode local contribution report: %w", err)
		}
		xmlContent := xml.Header + string(reportXML)
		validation, err := NewDefaultGOSIAdapter(s.repo).Validate(txCtx, []byte(xmlContent))
		if err != nil {
			return err
		}
		if !validation.Valid {
			return fmt.Errorf("invalid contribution export: %s", strings.Join(validation.Errors, "; "))
		}

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

		savedBatch, err = s.repo.CreateExportBatch(txCtx, batch)
		if err != nil {
			return fmt.Errorf("failed to create export batch: %w", err)
		}

		linked, err := s.repo.LinkCalculationsToExportBatch(txCtx, savedBatch.ID, legalEntityID, year, month)
		if err != nil {
			return fmt.Errorf("failed to link calculations to export batch: %w", err)
		}
		if linked != int64(len(calcs)) {
			return fmt.Errorf("linked %d of %d calculations to export batch", linked, len(calcs))
		}

		_, err = s.repo.CreateIntegrationAttempt(txCtx, domain.IntegrationAttempt{
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
				"status":              "LOCAL_EXPORT_CREATED",
				"external_submission": false,
				"schema":              "urn:erp:gosi:export:v1",
				"format":              "XML",
				"filesize":            len(xmlContent),
			},
			Status:    "SUCCESS",
			StartedAt: generatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to record export integration attempt: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

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
