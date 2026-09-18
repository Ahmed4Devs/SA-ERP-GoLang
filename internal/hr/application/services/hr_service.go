package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/hr/domain"
	"erp/internal/hr/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type hrService struct {
	repo ports.HRRepository
	db   *database.DB
	tx   transaction.Manager
}

func NewHRService(repo ports.HRRepository, db *database.DB) ports.HRService {
	return &hrService{
		repo: repo,
		db:   db,
		tx:   transaction.NewManager(db.Pool),
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

	var createdEmp *domain.Employee
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdEmp, err = s.repo.CreateEmployee(txCtx, emp)
		if err != nil {
			return err
		}

		if identifier != nil && identifier.IdentifierValue != "" {
			identifier.EmployeeID = createdEmp.ID
			if identifier.IdentifierType == "" {
				if createdEmp.IsSaudiNational {
					identifier.IdentifierType = domain.IdentifierNationalID
				} else {
					identifier.IdentifierType = domain.IdentifierIqama
				}
			}
			if _, err := s.repo.CreateIdentifier(txCtx, identifier); err != nil {
				return fmt.Errorf("failed to create employee identifier: %w", err)
			}
		}

		if contact != nil && contact.MobilePhone != "" {
			contact.EmployeeID = createdEmp.ID
			if _, err := s.repo.CreateContact(txCtx, contact); err != nil {
				return fmt.Errorf("failed to create employee contact: %w", err)
			}
		}

		if bankAccount != nil && bankAccount.IBAN != "" {
			bankAccount.EmployeeID = createdEmp.ID
			if _, err := s.repo.CreateBankAccount(txCtx, bankAccount); err != nil {
				return fmt.Errorf("failed to create employee bank account: %w", err)
			}
		}

		if createdEmp.PrimaryPositionID != nil {
			if _, err := s.repo.UpdatePositionStatus(txCtx, createdEmp.LegalEntityID, *createdEmp.PrimaryPositionID, domain.PositionFilled); err != nil {
				return fmt.Errorf("failed to mark employee position as filled: %w", err)
			}
		}
		return nil
	})
	return createdEmp, err
}

func (s *hrService) GetEmployee(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error) {
	return s.repo.GetEmployeeByID(ctx, legalEntityID, id)
}

func (s *hrService) ListEmployees(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Employee, error) {
	return s.repo.ListEmployees(ctx, legalEntityID)
}

func (s *hrService) TerminateEmployee(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error) {
	var terminated *domain.Employee
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		emp, err := s.repo.GetEmployeeByID(txCtx, legalEntityID, id)
		if err != nil {
			return err
		}
		if emp.EmploymentStatus == domain.EmploymentStatusTerminated {
			terminated = emp
			return nil
		}
		if emp.PrimaryPositionID != nil {
			if _, err := s.repo.UpdatePositionStatus(txCtx, legalEntityID, *emp.PrimaryPositionID, domain.PositionVacant); err != nil {
				return fmt.Errorf("failed to mark terminated employee position as vacant: %w", err)
			}
		}
		terminated, err = s.repo.UpdateEmployeeStatus(txCtx, legalEntityID, id, domain.EmploymentStatusTerminated)
		return err
	})
	return terminated, err
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

	var createdContract *domain.Contract
	var createdVer *domain.ContractVersion
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdContract, err = s.repo.CreateContract(txCtx, contract)
		if err != nil {
			return err
		}
		if initialVersion == nil {
			return nil
		}
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
		createdVer, err = s.repo.CreateContractVersion(txCtx, initialVersion)
		if err != nil {
			return fmt.Errorf("failed to create initial contract version: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return createdContract, createdVer, nil
}

func (s *hrService) AmendContractVersion(ctx context.Context, legalEntityID uuid.UUID, contractID uuid.UUID, newVersion *domain.ContractVersion) (*domain.ContractVersion, error) {
	var created *domain.ContractVersion
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.db.DBTX(txCtx).QueryRow(txCtx,
			"SELECT id FROM hr.contracts WHERE id = $1 AND legal_entity_id = $2 FOR UPDATE",
			contractID, legalEntityID,
		).Scan(&contractID); err != nil {
			return domain.ErrContractNotFound
		}
		contract, err := s.repo.GetContractByID(txCtx, legalEntityID, contractID)
		if err != nil {
			return err
		}
		if contract.Status == domain.ContractTerminated || contract.Status == domain.ContractExpired {
			return domain.ErrInvalidStatusTransition
		}

		latestVer, err := s.repo.GetLatestContractVersion(txCtx, contractID)
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
		created, err = s.repo.CreateContractVersion(txCtx, newVersion)
		return err
	})
	return created, err
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
	version, err := s.repo.GetLatestContractVersion(ctx, contract.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get latest contract version: %w", err)
	}
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
