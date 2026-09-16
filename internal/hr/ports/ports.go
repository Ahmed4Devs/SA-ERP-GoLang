package ports

import (
	"context"
	"erp/internal/hr/domain"

	"github.com/google/uuid"
)

type HRRepository interface {
	// Jobs
	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
	GetJobByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Job, error)
	GetJobByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Job, error)
	ListJobs(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Job, error)

	// Positions
	CreatePosition(ctx context.Context, pos *domain.Position) (*domain.Position, error)
	GetPositionByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Position, error)
	GetPositionByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Position, error)
	ListPositions(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Position, error)
	UpdatePositionStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.PositionStatus) (*domain.Position, error)

	// Employees
	CreateEmployee(ctx context.Context, emp *domain.Employee) (*domain.Employee, error)
	GetEmployeeByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error)
	GetEmployeeByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Employee, error)
	ListEmployees(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Employee, error)
	UpdateEmployeeStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.EmploymentStatus) (*domain.Employee, error)
	CountEmployeesByNationality(ctx context.Context, legalEntityID uuid.UUID) (saudiCount int64, expatCount int64, err error)

	// Identifiers & Contacts
	CreateIdentifier(ctx context.Context, identifier *domain.EmployeeIdentifier) (*domain.EmployeeIdentifier, error)
	ListIdentifiers(ctx context.Context, employeeID uuid.UUID) ([]*domain.EmployeeIdentifier, error)
	CreateContact(ctx context.Context, contact *domain.EmployeeContact) (*domain.EmployeeContact, error)
	GetContact(ctx context.Context, employeeID uuid.UUID) (*domain.EmployeeContact, error)

	// Contracts & Versions
	CreateContract(ctx context.Context, contract *domain.Contract) (*domain.Contract, error)
	GetContractByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Contract, error)
	GetContractByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Contract, error)
	GetContractByEmployeeID(ctx context.Context, legalEntityID, employeeID uuid.UUID) (*domain.Contract, error)
	ListContracts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Contract, error)
	UpdateContractStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.ContractStatus) (*domain.Contract, error)
	CreateContractVersion(ctx context.Context, version *domain.ContractVersion) (*domain.ContractVersion, error)
	ListContractVersions(ctx context.Context, contractID uuid.UUID) ([]*domain.ContractVersion, error)
	GetLatestContractVersion(ctx context.Context, contractID uuid.UUID) (*domain.ContractVersion, error)

	// Bank Accounts
	CreateBankAccount(ctx context.Context, account *domain.EmployeeBankAccount) (*domain.EmployeeBankAccount, error)
	ListBankAccounts(ctx context.Context, employeeID uuid.UUID) ([]*domain.EmployeeBankAccount, error)
}

type HRService interface {
	// Jobs & Positions
	CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error)
	ListJobs(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Job, error)
	CreatePosition(ctx context.Context, pos *domain.Position) (*domain.Position, error)
	ListPositions(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Position, error)

	// Employees
	OnboardEmployee(
		ctx context.Context,
		emp *domain.Employee,
		identifier *domain.EmployeeIdentifier,
		contact *domain.EmployeeContact,
		bankAccount *domain.EmployeeBankAccount,
	) (*domain.Employee, error)
	GetEmployee(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error)
	ListEmployees(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Employee, error)
	TerminateEmployee(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error)

	// Contracts
	CreateContract(ctx context.Context, contract *domain.Contract, initialVersion *domain.ContractVersion) (*domain.Contract, *domain.ContractVersion, error)
	AmendContractVersion(ctx context.Context, legalEntityID uuid.UUID, contractID uuid.UUID, newVersion *domain.ContractVersion) (*domain.ContractVersion, error)
	ActivateContract(ctx context.Context, legalEntityID, contractID uuid.UUID) (*domain.Contract, error)
	GetContract(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Contract, error)
	ListContracts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Contract, error)
	GetEmployeeContract(ctx context.Context, legalEntityID, employeeID uuid.UUID) (*domain.Contract, *domain.ContractVersion, error)

	// Saudization & Analytics
	GetSaudizationMetrics(ctx context.Context, legalEntityID uuid.UUID) (*domain.SaudizationMetric, error)
}
