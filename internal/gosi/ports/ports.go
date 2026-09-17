package ports

import (
	"context"
	"time"

	"erp/internal/gosi/domain"

	"github.com/google/uuid"
)

// Repository defines the data access interface for the GOSI module.
type Repository interface {
	// Rules
	CreateContributionRule(ctx context.Context, rule domain.ContributionRule) (*domain.ContributionRule, error)
	ListContributionRules(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ContributionRule, error)
	GetContributionRuleByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.ContributionRule, error)

	// Rule Versions
	CreateRuleVersion(ctx context.Context, rv domain.RuleVersion) (*domain.RuleVersion, error)
	GetActiveRuleVersions(ctx context.Context, legalEntityID uuid.UUID, asOfDate time.Time) ([]domain.RuleVersion, error)
	GetRuleVersionByID(ctx context.Context, id uuid.UUID) (*domain.RuleVersion, error)
	ListRuleVersions(ctx context.Context, ruleID uuid.UUID) ([]domain.RuleVersion, error)

	// Profiles
	CreateInsuranceProfile(ctx context.Context, profile domain.InsuranceProfile) (*domain.InsuranceProfile, error)
	GetInsuranceProfileByEmployee(ctx context.Context, legalEntityID, employeeID uuid.UUID) (*domain.InsuranceProfile, error)
	ListInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error)
	UpdateInsuranceProfile(ctx context.Context, profile domain.InsuranceProfile) (*domain.InsuranceProfile, error)
	ListActiveInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error)

	// Calculations
	CreateContributionCalculation(ctx context.Context, calc domain.ContributionCalculation) (*domain.ContributionCalculation, error)
	CreateCalculationLine(ctx context.Context, line domain.CalculationLine) (*domain.CalculationLine, error)
	GetContributionCalculation(ctx context.Context, id uuid.UUID) (*domain.ContributionCalculation, error)
	ListCalculationsByPeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error)
	ListCalculationLines(ctx context.Context, calcID uuid.UUID) ([]domain.CalculationLine, error)

	// Export Batches
	CreateExportBatch(ctx context.Context, batch domain.ExportBatch) (*domain.ExportBatch, error)
	GetExportBatchByID(ctx context.Context, id uuid.UUID) (*domain.ExportBatch, error)
	ListExportBatches(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ExportBatch, error)
	UpdateExportBatchStatus(ctx context.Context, id uuid.UUID, status string) (*domain.ExportBatch, error)
	LinkCalculationsToExportBatch(ctx context.Context, batchID, legalEntityID uuid.UUID, year, month int) error

	// Integration Attempts
	CreateIntegrationAttempt(ctx context.Context, attempt domain.IntegrationAttempt) (*domain.IntegrationAttempt, error)
	ListIntegrationAttempts(ctx context.Context, batchID uuid.UUID) ([]domain.IntegrationAttempt, error)
}

// Service defines the business logic interface for the GOSI module.
type Service interface {
	// Rules
	ListContributionRules(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ContributionRule, error)
	GetActiveRuleVersions(ctx context.Context, legalEntityID uuid.UUID, asOfDate time.Time) ([]domain.RuleVersion, error)

	// Profiles
	RegisterEmployee(ctx context.Context, profile domain.InsuranceProfile) (*domain.InsuranceProfile, error)
	ListInsuranceProfiles(ctx context.Context, legalEntityID uuid.UUID) ([]domain.InsuranceProfile, error)

	// Calculations (Section 26 GOSIService interface)
	Calculate(ctx context.Context, input domain.ContributionInput) ([]domain.SingleRuleResult, error)
	RunMonthlyCalculation(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error)
	ListCalculationsByPeriod(ctx context.Context, legalEntityID uuid.UUID, year, month int) ([]domain.ContributionCalculation, error)
	GetCalculationDetails(ctx context.Context, calcID uuid.UUID) (*domain.ContributionCalculation, []domain.CalculationLine, error)

	// Export
	CreateExportBatch(ctx context.Context, legalEntityID uuid.UUID, year, month int, createdBy *uuid.UUID) (*domain.ExportBatch, error)
	ListExportBatches(ctx context.Context, legalEntityID uuid.UUID) ([]domain.ExportBatch, error)

	// Dashboard
	GetDashboardSummary(ctx context.Context, legalEntityID uuid.UUID, year, month int) (*domain.DashboardSummary, error)
}
