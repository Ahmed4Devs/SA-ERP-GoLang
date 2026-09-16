package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/hr/domain"
	"erp/internal/hr/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type hrRepository struct {
	db *database.DB
	q  *sqlc.Queries
}

func NewHRRepository(db *database.DB) ports.HRRepository {
	return &hrRepository{
		db: db,
		q:  sqlc.New(db.Pool),
	}
}

// ============================================================================
// JOBS
// ============================================================================

func (r *hrRepository) CreateJob(ctx context.Context, j *domain.Job) (*domain.Job, error) {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	row, err := r.q.CreateJob(ctx, sqlc.CreateJobParams{
		ID:            toPgUUID(j.ID),
		LegalEntityID: toPgUUID(j.LegalEntityID),
		Code:          j.Code,
		TitleEn:       j.TitleEn,
		TitleAr:       j.TitleAr,
		JobFamily:     string(j.JobFamily),
		JobLevel:      string(j.JobLevel),
		Description:   toPgTextPtr(j.Description),
		IsActive:      j.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return mapJob(row), nil
}

func (r *hrRepository) GetJobByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Job, error) {
	row, err := r.q.GetJobByID(ctx, sqlc.GetJobByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, err
	}
	return mapJob(row), nil
}

func (r *hrRepository) GetJobByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Job, error) {
	row, err := r.q.GetJobByCode(ctx, sqlc.GetJobByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrJobNotFound
		}
		return nil, err
	}
	return mapJob(row), nil
}

func (r *hrRepository) ListJobs(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Job, error) {
	rows, err := r.q.ListJobs(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Job, len(rows))
	for i, row := range rows {
		res[i] = mapJob(row)
	}
	return res, nil
}

// ============================================================================
// POSITIONS
// ============================================================================

func (r *hrRepository) CreatePosition(ctx context.Context, p *domain.Position) (*domain.Position, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	row, err := r.q.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:                  toPgUUID(p.ID),
		LegalEntityID:       toPgUUID(p.LegalEntityID),
		Code:                p.Code,
		JobID:               toPgUUID(p.JobID),
		DepartmentID:        toPgUUID(p.DepartmentID),
		BranchID:            toPgUUIDPtr(p.BranchID),
		ReportsToPositionID: toPgUUIDPtr(p.ReportsToPositionID),
		Headcount:           p.Headcount,
		Status:              string(p.Status),
	})
	if err != nil {
		return nil, err
	}
	return mapPosition(row), nil
}

func (r *hrRepository) GetPositionByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Position, error) {
	row, err := r.q.GetPositionByID(ctx, sqlc.GetPositionByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPositionNotFound
		}
		return nil, err
	}
	return mapPosition(row), nil
}

func (r *hrRepository) GetPositionByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Position, error) {
	row, err := r.q.GetPositionByCode(ctx, sqlc.GetPositionByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPositionNotFound
		}
		return nil, err
	}
	return mapPosition(row), nil
}

func (r *hrRepository) ListPositions(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Position, error) {
	rows, err := r.q.ListPositions(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Position, len(rows))
	for i, row := range rows {
		res[i] = mapPosition(row)
	}
	return res, nil
}

func (r *hrRepository) UpdatePositionStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.PositionStatus) (*domain.Position, error) {
	row, err := r.q.UpdatePositionStatus(ctx, sqlc.UpdatePositionStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapPosition(row), nil
}

// ============================================================================
// EMPLOYEES
// ============================================================================

func (r *hrRepository) CreateEmployee(ctx context.Context, e *domain.Employee) (*domain.Employee, error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	row, err := r.q.CreateEmployee(ctx, sqlc.CreateEmployeeParams{
		ID:                     toPgUUID(e.ID),
		LegalEntityID:          toPgUUID(e.LegalEntityID),
		EmployeeNumber:         e.EmployeeNumber,
		UserID:                 toPgUUIDPtr(e.UserID),
		FirstNameEn:            e.FirstNameEn,
		LastNameEn:             e.LastNameEn,
		FirstNameAr:            e.FirstNameAr,
		LastNameAr:             e.LastNameAr,
		Gender:                 string(e.Gender),
		MaritalStatus:          string(e.MaritalStatus),
		DateOfBirth:            toPgDate(e.DateOfBirth),
		NationalityCountryCode: e.NationalityCountryCode,
		IsSaudiNational:        e.IsSaudiNational,
		HireDate:               toPgDate(e.HireDate),
		ProbationEndDate:       toPgDatePtr(e.ProbationEndDate),
		TerminationDate:        toPgDatePtr(e.TerminationDate),
		EmploymentStatus:       string(e.EmploymentStatus),
		PrimaryPositionID:      toPgUUIDPtr(e.PrimaryPositionID),
		DepartmentID:           toPgUUIDPtr(e.DepartmentID),
		BranchID:               toPgUUIDPtr(e.BranchID),
		ManagerEmployeeID:      toPgUUIDPtr(e.ManagerEmployeeID),
	})
	if err != nil {
		return nil, err
	}
	return mapEmployee(row), nil
}

func (r *hrRepository) GetEmployeeByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Employee, error) {
	row, err := r.q.GetEmployeeByID(ctx, sqlc.GetEmployeeByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}
	return mapEmployee(row), nil
}

func (r *hrRepository) GetEmployeeByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Employee, error) {
	row, err := r.q.GetEmployeeByNumber(ctx, sqlc.GetEmployeeByNumberParams{
		LegalEntityID:  toPgUUID(legalEntityID),
		EmployeeNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEmployeeNotFound
		}
		return nil, err
	}
	return mapEmployee(row), nil
}

func (r *hrRepository) ListEmployees(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Employee, error) {
	rows, err := r.q.ListEmployees(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Employee, len(rows))
	for i, row := range rows {
		res[i] = mapEmployee(row)
	}
	return res, nil
}

func (r *hrRepository) UpdateEmployeeStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.EmploymentStatus) (*domain.Employee, error) {
	var termDate *time.Time
	if status == domain.EmploymentStatusTerminated {
		now := time.Now()
		termDate = &now
	}
	row, err := r.q.UpdateEmployeeStatus(ctx, sqlc.UpdateEmployeeStatusParams{
		ID:               toPgUUID(id),
		LegalEntityID:    toPgUUID(legalEntityID),
		EmploymentStatus: string(status),
		TerminationDate:  toPgDatePtr(termDate),
	})
	if err != nil {
		return nil, err
	}
	return mapEmployee(row), nil
}

func (r *hrRepository) CountEmployeesByNationality(ctx context.Context, legalEntityID uuid.UUID) (saudiCount int64, expatCount int64, err error) {
	rows, err := r.q.CountEmployeesByNationality(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return 0, 0, err
	}
	for _, row := range rows {
		if row.IsSaudiNational {
			saudiCount = row.Count
		} else {
			expatCount = row.Count
		}
	}
	return saudiCount, expatCount, nil
}

// ============================================================================
// IDENTIFIERS & CONTACTS
// ============================================================================

func (r *hrRepository) CreateIdentifier(ctx context.Context, iden *domain.EmployeeIdentifier) (*domain.EmployeeIdentifier, error) {
	if iden.ID == uuid.Nil {
		iden.ID = uuid.New()
	}
	if iden.IdentifierValueMasked == "" {
		iden.IdentifierValueMasked = domain.MaskIdentifier(iden.IdentifierValue)
	}
	row, err := r.q.CreateEmployeeIdentifier(ctx, sqlc.CreateEmployeeIdentifierParams{
		ID:                    toPgUUID(iden.ID),
		EmployeeID:            toPgUUID(iden.EmployeeID),
		IdentifierType:        string(iden.IdentifierType),
		IdentifierValue:       iden.IdentifierValue,
		IdentifierValueMasked: iden.IdentifierValueMasked,
		IssueDate:             toPgDatePtr(iden.IssueDate),
		ExpiryDate:            toPgDatePtr(iden.ExpiryDate),
		IssuePlace:            toPgTextPtr(iden.IssuePlace),
		IsPrimary:             iden.IsPrimary,
		IsVerified:            iden.IsVerified,
	})
	if err != nil {
		return nil, err
	}
	return mapIdentifier(row), nil
}

func (r *hrRepository) ListIdentifiers(ctx context.Context, employeeID uuid.UUID) ([]*domain.EmployeeIdentifier, error) {
	rows, err := r.q.ListEmployeeIdentifiers(ctx, toPgUUID(employeeID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.EmployeeIdentifier, len(rows))
	for i, row := range rows {
		res[i] = mapIdentifier(row)
	}
	return res, nil
}

func (r *hrRepository) CreateContact(ctx context.Context, c *domain.EmployeeContact) (*domain.EmployeeContact, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	row, err := r.q.CreateEmployeeContact(ctx, sqlc.CreateEmployeeContactParams{
		ID:            toPgUUID(c.ID),
		EmployeeID:    toPgUUID(c.EmployeeID),
		WorkEmail:     toPgTextPtr(c.WorkEmail),
		PersonalEmail: toPgTextPtr(c.PersonalEmail),
		MobilePhone:   c.MobilePhone,
		WorkPhone:     toPgTextPtr(c.WorkPhone),
		Extension:     toPgTextPtr(c.Extension),
		BuildingNo:    toPgTextPtr(c.BuildingNo),
		StreetName:    toPgTextPtr(c.StreetName),
		District:      toPgTextPtr(c.District),
		City:          c.City,
		PostalCode:    toPgTextPtr(c.PostalCode),
		AdditionalNo:  toPgTextPtr(c.AdditionalNo),
		CountryCode:   c.CountryCode,
	})
	if err != nil {
		return nil, err
	}
	return mapContact(row), nil
}

func (r *hrRepository) GetContact(ctx context.Context, employeeID uuid.UUID) (*domain.EmployeeContact, error) {
	row, err := r.q.GetEmployeeContact(ctx, toPgUUID(employeeID))
	if err != nil {
		return nil, err
	}
	return mapContact(row), nil
}

// ============================================================================
// CONTRACTS & VERSIONS
// ============================================================================

func (r *hrRepository) CreateContract(ctx context.Context, c *domain.Contract) (*domain.Contract, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	row, err := r.q.CreateContract(ctx, sqlc.CreateContractParams{
		ID:                  toPgUUID(c.ID),
		LegalEntityID:       toPgUUID(c.LegalEntityID),
		EmployeeID:          toPgUUID(c.EmployeeID),
		ContractNumber:      c.ContractNumber,
		ContractType:        string(c.ContractType),
		StartDate:           toPgDate(c.StartDate),
		EndDate:             toPgDatePtr(c.EndDate),
		ProbationDays:       c.ProbationDays,
		NoticePeriodDays:    c.NoticePeriodDays,
		WorkingHoursPerWeek: toPgNumeric(c.WorkingHoursPerWeek),
		WorkLocation:        c.WorkLocation,
		Status:              string(c.Status),
		Notes:               toPgTextPtr(c.Notes),
	})
	if err != nil {
		return nil, err
	}
	return mapContract(row), nil
}

func (r *hrRepository) GetContractByID(ctx context.Context, legalEntityID, id uuid.UUID) (*domain.Contract, error) {
	row, err := r.q.GetContractByID(ctx, sqlc.GetContractByIDParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, err
	}
	return mapContract(row), nil
}

func (r *hrRepository) GetContractByNumber(ctx context.Context, legalEntityID uuid.UUID, number string) (*domain.Contract, error) {
	row, err := r.q.GetContractByNumber(ctx, sqlc.GetContractByNumberParams{
		LegalEntityID:  toPgUUID(legalEntityID),
		ContractNumber: number,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, err
	}
	return mapContract(row), nil
}

func (r *hrRepository) GetContractByEmployeeID(ctx context.Context, legalEntityID, employeeID uuid.UUID) (*domain.Contract, error) {
	row, err := r.q.GetContractByEmployeeID(ctx, sqlc.GetContractByEmployeeIDParams{
		EmployeeID:    toPgUUID(employeeID),
		LegalEntityID: toPgUUID(legalEntityID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrContractNotFound
		}
		return nil, err
	}
	return mapContract(row), nil
}

func (r *hrRepository) ListContracts(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Contract, error) {
	rows, err := r.q.ListContracts(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.Contract, len(rows))
	for i, row := range rows {
		res[i] = mapContract(row)
	}
	return res, nil
}

func (r *hrRepository) UpdateContractStatus(ctx context.Context, legalEntityID, id uuid.UUID, status domain.ContractStatus) (*domain.Contract, error) {
	row, err := r.q.UpdateContractStatus(ctx, sqlc.UpdateContractStatusParams{
		ID:            toPgUUID(id),
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapContract(row), nil
}

func (r *hrRepository) CreateContractVersion(ctx context.Context, v *domain.ContractVersion) (*domain.ContractVersion, error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.GrossSalary.IsZero() {
		v.GrossSalary = domain.CalculateGrossSalary(v.BasicSalary, v.HousingAllowance, v.TransportAllowance, v.OtherAllowances)
	}
	row, err := r.q.CreateContractVersion(ctx, sqlc.CreateContractVersionParams{
		ID:                 toPgUUID(v.ID),
		ContractID:         toPgUUID(v.ContractID),
		VersionNumber:      v.VersionNumber,
		EffectiveFrom:      toPgDate(v.EffectiveFrom),
		EffectiveTo:        toPgDatePtr(v.EffectiveTo),
		BasicSalary:        toPgNumeric(v.BasicSalary),
		HousingAllowance:   toPgNumeric(v.HousingAllowance),
		TransportAllowance: toPgNumeric(v.TransportAllowance),
		OtherAllowances:    toPgNumeric(v.OtherAllowances),
		GrossSalary:        toPgNumeric(v.GrossSalary),
		ChangeReason:       toPgTextPtr(v.ChangeReason),
		ApprovedByUserID:   toPgUUIDPtr(v.ApprovedByUserID),
	})
	if err != nil {
		return nil, err
	}
	return mapContractVersion(row), nil
}

func (r *hrRepository) ListContractVersions(ctx context.Context, contractID uuid.UUID) ([]*domain.ContractVersion, error) {
	rows, err := r.q.ListContractVersions(ctx, toPgUUID(contractID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.ContractVersion, len(rows))
	for i, row := range rows {
		res[i] = mapContractVersion(row)
	}
	return res, nil
}

func (r *hrRepository) GetLatestContractVersion(ctx context.Context, contractID uuid.UUID) (*domain.ContractVersion, error) {
	row, err := r.q.GetLatestContractVersion(ctx, toPgUUID(contractID))
	if err != nil {
		return nil, err
	}
	return mapContractVersion(row), nil
}

// ============================================================================
// BANK ACCOUNTS
// ============================================================================

func (r *hrRepository) CreateBankAccount(ctx context.Context, a *domain.EmployeeBankAccount) (*domain.EmployeeBankAccount, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.IBANMasked == "" {
		a.IBANMasked = domain.MaskIBAN(a.IBAN)
	}
	row, err := r.q.CreateEmployeeBankAccount(ctx, sqlc.CreateEmployeeBankAccountParams{
		ID:            toPgUUID(a.ID),
		EmployeeID:    toPgUUID(a.EmployeeID),
		BankNameEn:    a.BankNameEn,
		BankNameAr:    a.BankNameAr,
		Iban:          a.IBAN,
		IbanMasked:    a.IBANMasked,
		SwiftBic:      a.SwiftBIC,
		IsPrimary:     a.IsPrimary,
		AccountStatus: a.AccountStatus,
	})
	if err != nil {
		return nil, err
	}
	return mapEmployeeBankAccount(row), nil
}

func (r *hrRepository) ListBankAccounts(ctx context.Context, employeeID uuid.UUID) ([]*domain.EmployeeBankAccount, error) {
	rows, err := r.q.ListEmployeeBankAccounts(ctx, toPgUUID(employeeID))
	if err != nil {
		return nil, err
	}
	res := make([]*domain.EmployeeBankAccount, len(rows))
	for i, row := range rows {
		res[i] = mapEmployeeBankAccount(row)
	}
	return res, nil
}

// ============================================================================
// MAPPERS
// ============================================================================

func mapJob(r sqlc.HrJobs) *domain.Job {
	return &domain.Job{
		ID:            toUUID(r.ID),
		LegalEntityID: toUUID(r.LegalEntityID),
		Code:          r.Code,
		TitleEn:       r.TitleEn,
		TitleAr:       r.TitleAr,
		JobFamily:     domain.JobFamily(r.JobFamily),
		JobLevel:      domain.JobLevel(r.JobLevel),
		Description:   fromPgTextPtr(r.Description),
		IsActive:      r.IsActive,
		CreatedAt:     fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:     fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapPosition(r sqlc.HrPositions) *domain.Position {
	return &domain.Position{
		ID:                  toUUID(r.ID),
		LegalEntityID:       toUUID(r.LegalEntityID),
		Code:                r.Code,
		JobID:               toUUID(r.JobID),
		DepartmentID:        toUUID(r.DepartmentID),
		BranchID:            toUUIDPtr(r.BranchID),
		ReportsToPositionID: toUUIDPtr(r.ReportsToPositionID),
		Headcount:           r.Headcount,
		Status:              domain.PositionStatus(r.Status),
		CreatedAt:           fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:           fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapEmployee(r sqlc.HrEmployees) *domain.Employee {
	return &domain.Employee{
		ID:                     toUUID(r.ID),
		LegalEntityID:          toUUID(r.LegalEntityID),
		EmployeeNumber:         r.EmployeeNumber,
		UserID:                 toUUIDPtr(r.UserID),
		FirstNameEn:            r.FirstNameEn,
		LastNameEn:             r.LastNameEn,
		FirstNameAr:            r.FirstNameAr,
		LastNameAr:             r.LastNameAr,
		Gender:                 domain.Gender(r.Gender),
		MaritalStatus:          domain.MaritalStatus(r.MaritalStatus),
		DateOfBirth:            fromPgDate(r.DateOfBirth),
		NationalityCountryCode: r.NationalityCountryCode,
		IsSaudiNational:        r.IsSaudiNational,
		HireDate:               fromPgDate(r.HireDate),
		ProbationEndDate:       fromPgDatePtr(r.ProbationEndDate),
		TerminationDate:        fromPgDatePtr(r.TerminationDate),
		EmploymentStatus:       domain.EmploymentStatus(r.EmploymentStatus),
		PrimaryPositionID:      toUUIDPtr(r.PrimaryPositionID),
		DepartmentID:           toUUIDPtr(r.DepartmentID),
		BranchID:               toUUIDPtr(r.BranchID),
		ManagerEmployeeID:      toUUIDPtr(r.ManagerEmployeeID),
		CreatedAt:              fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:              fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapIdentifier(r sqlc.HrEmployeeIdentifiers) *domain.EmployeeIdentifier {
	return &domain.EmployeeIdentifier{
		ID:                    toUUID(r.ID),
		EmployeeID:            toUUID(r.EmployeeID),
		IdentifierType:        domain.IdentifierType(r.IdentifierType),
		IdentifierValue:       r.IdentifierValue,
		IdentifierValueMasked: r.IdentifierValueMasked,
		IssueDate:             fromPgDatePtr(r.IssueDate),
		ExpiryDate:            fromPgDatePtr(r.ExpiryDate),
		IssuePlace:            fromPgTextPtr(r.IssuePlace),
		IsPrimary:             r.IsPrimary,
		IsVerified:            r.IsVerified,
		CreatedAt:             fromPgTimestamptz(r.CreatedAt),
	}
}

func mapContact(r sqlc.HrEmployeeContacts) *domain.EmployeeContact {
	return &domain.EmployeeContact{
		ID:            toUUID(r.ID),
		EmployeeID:    toUUID(r.EmployeeID),
		WorkEmail:     fromPgTextPtr(r.WorkEmail),
		PersonalEmail: fromPgTextPtr(r.PersonalEmail),
		MobilePhone:   r.MobilePhone,
		WorkPhone:     fromPgTextPtr(r.WorkPhone),
		Extension:     fromPgTextPtr(r.Extension),
		BuildingNo:    fromPgTextPtr(r.BuildingNo),
		StreetName:    fromPgTextPtr(r.StreetName),
		District:      fromPgTextPtr(r.District),
		City:          r.City,
		PostalCode:    fromPgTextPtr(r.PostalCode),
		AdditionalNo:  fromPgTextPtr(r.AdditionalNo),
		CountryCode:   r.CountryCode,
		CreatedAt:     fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:     fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapContract(r sqlc.HrContracts) *domain.Contract {
	return &domain.Contract{
		ID:                  toUUID(r.ID),
		LegalEntityID:       toUUID(r.LegalEntityID),
		EmployeeID:          toUUID(r.EmployeeID),
		ContractNumber:      r.ContractNumber,
		ContractType:        domain.ContractType(r.ContractType),
		StartDate:           fromPgDate(r.StartDate),
		EndDate:             fromPgDatePtr(r.EndDate),
		ProbationDays:       r.ProbationDays,
		NoticePeriodDays:    r.NoticePeriodDays,
		WorkingHoursPerWeek: fromPgNumeric(r.WorkingHoursPerWeek),
		WorkLocation:        r.WorkLocation,
		Status:              domain.ContractStatus(r.Status),
		Notes:               fromPgTextPtr(r.Notes),
		CreatedAt:           fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:           fromPgTimestamptz(r.UpdatedAt),
	}
}

func mapContractVersion(r sqlc.HrContractVersions) *domain.ContractVersion {
	return &domain.ContractVersion{
		ID:                 toUUID(r.ID),
		ContractID:         toUUID(r.ContractID),
		VersionNumber:      r.VersionNumber,
		EffectiveFrom:      fromPgDate(r.EffectiveFrom),
		EffectiveTo:        fromPgDatePtr(r.EffectiveTo),
		BasicSalary:        fromPgNumeric(r.BasicSalary),
		HousingAllowance:   fromPgNumeric(r.HousingAllowance),
		TransportAllowance: fromPgNumeric(r.TransportAllowance),
		OtherAllowances:    fromPgNumeric(r.OtherAllowances),
		GrossSalary:        fromPgNumeric(r.GrossSalary),
		ChangeReason:       fromPgTextPtr(r.ChangeReason),
		ApprovedByUserID:   toUUIDPtr(r.ApprovedByUserID),
		CreatedAt:          fromPgTimestamptz(r.CreatedAt),
	}
}

func mapEmployeeBankAccount(r sqlc.HrBankAccounts) *domain.EmployeeBankAccount {
	return &domain.EmployeeBankAccount{
		ID:            toUUID(r.ID),
		EmployeeID:    toUUID(r.EmployeeID),
		BankNameEn:    r.BankNameEn,
		BankNameAr:    r.BankNameAr,
		IBAN:          r.Iban,
		IBANMasked:    r.IbanMasked,
		SwiftBIC:      r.SwiftBic,
		IsPrimary:     r.IsPrimary,
		AccountStatus: r.AccountStatus,
		CreatedAt:     fromPgTimestamptz(r.CreatedAt),
		UpdatedAt:     fromPgTimestamptz(r.UpdatedAt),
	}
}
