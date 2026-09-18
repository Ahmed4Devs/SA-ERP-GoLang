package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/hr/domain"
	"erp/internal/hr/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type hrService struct {
	repo ports.HRRepository
	db   *database.DB
}

func NewHRService(repo ports.HRRepository, db *database.DB) ports.HRService {
	return &hrService{
		repo: repo,
		db:   db,
	}
}

// ============================================================================
// JOBS & POSITIONS
// ============================================================================

func (s *hrService) CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	if job.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if job.Code == "" {
		return nil, fmt.Errorf("job code is required")
	}
	if job.TitleEn == "" || job.TitleAr == "" {
		return nil, fmt.Errorf("title_en and title_ar are required")
	}
	if job.JobFamily == "" {
		job.JobFamily = domain.JobFamilyOperations
	}
	if job.JobLevel == "" {
		job.JobLevel = domain.JobLevelMid
	}
	return s.repo.CreateJob(ctx, job)
}

func (s *hrService) ListJobs(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Job, error) {
	return s.repo.ListJobs(ctx, legalEntityID)
}

func (s *hrService) CreatePosition(ctx context.Context, pos *domain.Position) (*domain.Position, error) {
	if pos.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if pos.Code == "" {
		return nil, fmt.Errorf("position code is required")
	}
	if pos.JobID == uuid.Nil {
		return nil, fmt.Errorf("job_id is required")
	}
	if pos.DepartmentID == uuid.Nil {
		return nil, fmt.Errorf("department_id is required")
	}
	if pos.Headcount <= 0 {
		pos.Headcount = 1
	}
	if pos.Status == "" {
		pos.Status = domain.PositionVacant
	}
	return s.repo.CreatePosition(ctx, pos)
}

func (s *hrService) ListPositions(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Position, error) {
	return s.repo.ListPositions(ctx, legalEntityID)
}

// ============================================================================
// EMPLOYEES
// ============================================================================

func (s *hrService) OnboardEmployee(
	ctx context.Context,
	emp *domain.Employee,
	identifier *domain.EmployeeIdentifier,
	contact *domain.EmployeeContact,
	bankAccount *domain.EmployeeBankAccount,
) (*domain.Employee, error) {
	if emp.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if emp.FirstNameEn == "" || emp.LastNameEn == "" || emp.FirstNameAr == "" || emp.LastNameAr == "" {
		return nil, fmt.Errorf("employee bilingual names are required")
	}
	if emp.EmployeeNumber == "" {
		emp.EmployeeNumber = identifiers.NewDocumentNumber("EMP", time.Now())
	}
	if emp.DateOfBirth.IsZero() {
		emp.DateOfBirth = time.Now().AddDate(-25, 0, 0)
	}
	if emp.HireDate.IsZero() {
		emp.HireDate = time.Now()
	}
	if emp.EmploymentStatus == "" {
		emp.EmploymentStatus = domain.EmploymentStatusActive
	}
	if emp.NationalityCountryCode == "" {
		emp.NationalityCountryCode = "SA"
	}
	if emp.Gender == "" {
		emp.Gender = domain.GenderMale
	}
	if emp.MaritalStatus == "" {
		emp.MaritalStatus = domain.MaritalStatusSingle
	}

	createdEmp, err := s.repo.CreateEmployee(ctx, emp)
	if err != nil {
		return nil, err
	}

	// 1. Identifiers
	if identifier != nil && identifier.IdentifierValue != "" {
		identifier.EmployeeID = createdEmp.ID
		if identifier.IdentifierType == "" {
			if createdEmp.IsSaudiNational {
				identifier.IdentifierType = domain.IdentifierNationalID
			} else {
				identifier.IdentifierType = domain.IdentifierIqama
			}
		}
		if _, err := s.repo.CreateIdentifier(ctx, identifier); err != nil {
			return nil, err
		}
	}

	// 2. Contacts
	if contact != nil && contact.MobilePhone != "" {
		contact.EmployeeID = createdEmp.ID
		if _, err := s.repo.CreateContact(ctx, contact); err != nil {
			return nil, err
		}
	}

	// 3. Bank Account
	if bankAccount != nil && bankAccount.IBAN != "" {
		bankAccount.EmployeeID = createdEmp.ID
		if _, err := s.repo.CreateBankAccount(ctx, bankAccount); err != nil {
			return nil, err
		}
	}

	// 4. Update Position Status
	if createdEmp.PrimaryPositionID != nil {
		_, _ = s.repo.UpdatePositionStatus(ctx, createdEmp.LegalEntityID, *createdEmp.PrimaryPositionID, domain.PositionFilled)
	}

	return createdEmp, nil
}

func (s *hrService) GetEmployee(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error) {
	return s.repo.GetEmployeeByID(ctx, legalEntityID, id)
}

func (s *hrService) ListEmployees(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Employee, error) {
	return s.repo.ListEmployees(ctx, legalEntityID)
}

func (s *hrService) TerminateEmployee(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error) {
	emp, err := s.repo.GetEmployeeByID(ctx, legalEntityID, id)
	if err != nil {
		return nil, err
	}
	if emp.EmploymentStatus == domain.EmploymentStatusTerminated {
		return emp, nil
	}

	// Mark position as VACANT if primary
	if emp.PrimaryPositionID != nil {
		_, _ = s.repo.UpdatePositionStatus(ctx, legalEntityID, *emp.PrimaryPositionID, domain.PositionVacant)
	}

	return s.repo.UpdateEmployeeStatus(ctx, legalEntityID, id, domain.EmploymentStatusTerminated)
}

// ============================================================================
// CONTRACTS & VERSIONS
// ============================================================================

func (s *hrService) CreateContract(ctx context.Context, contract *domain.Contract, initialVersion *domain.ContractVersion) (*domain.Contract, *domain.ContractVersion, error) {
	if contract.LegalEntityID == uuid.Nil {
		return nil, nil, fmt.Errorf("legal_entity_id is required")
	}
	if contract.EmployeeID == uuid.Nil {
		return nil, nil, fmt.Errorf("employee_id is required")
	}
	if contract.ContractNumber == "" {
		contract.ContractNumber = identifiers.NewDocumentNumber("QIWA", time.Now())
	}
	if contract.StartDate.IsZero() {
		contract.StartDate = time.Now()
	}
	if contract.ProbationDays > 180 {
		return nil, nil, domain.ErrInvalidProbationPeriod
	}
	if contract.WorkingHoursPerWeek.IsZero() {
		contract.WorkingHoursPerWeek = decimal.NewFromInt(40)
	}
	if contract.Status == "" {
		contract.Status = domain.ContractActive
	}

	createdContract, err := s.repo.CreateContract(ctx, contract)
	if err != nil {
		return nil, nil, err
	}

	var createdVer *domain.ContractVersion
	if initialVersion != nil {
		initialVersion.ContractID = createdContract.ID
		initialVersion.VersionNumber = 1
		if initialVersion.EffectiveFrom.IsZero() {
			initialVersion.EffectiveFrom = createdContract.StartDate
		}
		initialVersion.GrossSalary = domain.CalculateGrossSalary(
			initialVersion.BasicSalary,
			initialVersion.HousingAllowance,
			initialVersion.TransportAllowance,
			initialVersion.OtherAllowances,
		)
		createdVer, err = s.repo.CreateContractVersion(ctx, initialVersion)
		if err != nil {
			return nil, nil, err
		}
	}

	return createdContract, createdVer, nil
}

func (s *hrService) AmendContractVersion(ctx context.Context, legalEntityID uuid.UUID, contractID uuid.UUID, newVersion *domain.ContractVersion) (*domain.ContractVersion, error) {
	contract, err := s.repo.GetContractByID(ctx, legalEntityID, contractID)
	if err != nil {
		return nil, err
	}
	if contract.Status == domain.ContractTerminated || contract.Status == domain.ContractExpired {
		return nil, domain.ErrInvalidStatusTransition
	}

	latestVer, err := s.repo.GetLatestContractVersion(ctx, contractID)
	nextVerNum := int32(1)
	if err == nil && latestVer != nil {
		nextVerNum = latestVer.VersionNumber + 1
	}

	newVersion.ContractID = contractID
	newVersion.VersionNumber = nextVerNum
	if newVersion.EffectiveFrom.IsZero() {
		newVersion.EffectiveFrom = time.Now()
	}
	newVersion.GrossSalary = domain.CalculateGrossSalary(
		newVersion.BasicSalary,
		newVersion.HousingAllowance,
		newVersion.TransportAllowance,
		newVersion.OtherAllowances,
	)

	return s.repo.CreateContractVersion(ctx, newVersion)
}

func (s *hrService) ActivateContract(ctx context.Context, legalEntityID, contractID uuid.UUID) (*domain.Contract, error) {
	return s.repo.UpdateContractStatus(ctx, legalEntityID, contractID, domain.ContractActive)
}

func (s *hrService) GetContract(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Contract, error) {
	return s.repo.GetContractByID(ctx, legalEntityID, id)
}

func (s *hrService) ListContracts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Contract, error) {
	return s.repo.ListContracts(ctx, legalEntityID)
}

func (s *hrService) GetEmployeeContract(ctx context.Context, legalEntityID, employeeID uuid.UUID) (*domain.Contract, *domain.ContractVersion, error) {
	contract, err := s.repo.GetContractByEmployeeID(ctx, legalEntityID, employeeID)
	if err != nil {
		return nil, nil, err
	}
	version, _ := s.repo.GetLatestContractVersion(ctx, contract.ID)
	return contract, version, nil
}

// ============================================================================
// SAUDIZATION & ANALYTICS
// ============================================================================

func (s *hrService) GetSaudizationMetrics(ctx context.Context, legalEntityID uuid.UUID) (*domain.SaudizationMetric, error) {
	saudiCount, expatCount, err := s.repo.CountEmployeesByNationality(ctx, legalEntityID)
	if err != nil {
		return nil, err
	}

	total := saudiCount + expatCount
	ratio := decimal.Zero
	tier := "RED"

	if total > 0 {
		ratio = decimal.NewFromInt(saudiCount).Div(decimal.NewFromInt(total)).Mul(decimal.NewFromInt(100)).Round(2)
		val, _ := ratio.Float64()
		if val >= 50.0 {
			tier = "PLATINUM"
		} else if val >= 40.0 {
			tier = "HIGH_GREEN"
		} else if val >= 25.0 {
			tier = "MID_GREEN"
		} else if val >= 10.0 {
			tier = "LOW_GREEN"
		} else {
			tier = "RED"
		}
	}

	return &domain.SaudizationMetric{
		SaudiCount:              saudiCount,
		ExpatCount:              expatCount,
		TotalCount:              total,
		SaudizationRatioPercent: ratio,
		NitaqatTier:             tier,
	}, nil
}
