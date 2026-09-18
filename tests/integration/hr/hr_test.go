package hr_test

import (
	"context"
	"testing"
	"time"

	"erp/internal/hr"
	"erp/internal/hr/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupHRTest(t *testing.T) (*hr.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("PostgreSQL not accessible: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	hrMod := hr.NewModule(cfg, db)

	// 1. Seed org
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Admin user
	_, err = iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	if err != nil {
		t.Logf("admin lookup note: %v", err)
	}

	// 3. Obtain or create Department
	depts, err := orgMod.OrgService.ListDepartments(ctx, le.ID)
	var deptID uuid.UUID
	if err == nil && len(depts) > 0 {
		deptID = depts[0].ID
	} else {
		dept, err := orgMod.OrgService.CreateDepartment(ctx, le.ID, nil, "HR-TEST-"+uuid.New().String()[:6], "الموارد البشرية", "Human Resources")
		if err != nil {
			t.Fatalf("failed to create department: %v", err)
		}
		deptID = dept.ID
	}

	return hrMod, db, le.ID, deptID, ctx
}

func TestEmployee_OnboardingAndIdentifiers(t *testing.T) {
	hrMod, _, leID, deptID, ctx := setupHRTest(t)
	svc := hrMod.HRService
	repo := hrMod.Repo

	// 1. Create Job & Position
	job, err := svc.CreateJob(ctx, &domain.Job{
		LegalEntityID: leID,
		Code:          "JOB-HR-" + uuid.New().String()[:6],
		TitleEn:       "HR Specialist",
		TitleAr:       "أخصائي موارد بشرية",
		JobFamily:     domain.JobFamilyHR,
		JobLevel:      domain.JobLevelSenior,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	pos, err := svc.CreatePosition(ctx, &domain.Position{
		LegalEntityID: leID,
		Code:          "POS-HR-" + uuid.New().String()[:6],
		JobID:         job.ID,
		DepartmentID:  deptID,
		Headcount:     1,
		Status:        domain.PositionVacant,
	})
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if pos.Status != domain.PositionVacant {
		t.Fatalf("expected position status VACANT, got %s", pos.Status)
	}

	// 2. Onboard Saudi Employee
	dob := time.Date(1993, 4, 15, 0, 0, 0, 0, time.UTC)
	hireDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	empNumber := "EMP-SA-" + uuid.New().String()[:6]

	emp := &domain.Employee{
		LegalEntityID:          leID,
		EmployeeNumber:         empNumber,
		FirstNameEn:            "Mohammed",
		LastNameEn:             "Al-Otaibi",
		FirstNameAr:            "محمد",
		LastNameAr:             "العتيبي",
		DateOfBirth:            dob,
		Gender:                 domain.GenderMale,
		MaritalStatus:          domain.MaritalStatusMarried,
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               hireDate,
		EmploymentStatus:       domain.EmploymentStatusActive,
		PrimaryPositionID:      &pos.ID,
	}

	natID := "1087654321"
	ident := &domain.EmployeeIdentifier{
		IdentifierType:  domain.IdentifierNationalID,
		IdentifierValue: natID,
		IsPrimary:       true,
		IsVerified:      true,
	}

	bNo := "4251"
	stName := "King Fahd Road"
	dist := "Al Olaya"
	pCode := "12214"
	addNo := "8421"
	workEmail := "mohammed.otaibi@example.sa"
	contact := &domain.EmployeeContact{
		WorkEmail:    &workEmail,
		MobilePhone:  "+966501234567",
		BuildingNo:   &bNo,
		StreetName:   &stName,
		District:     &dist,
		City:         "Riyadh",
		PostalCode:   &pCode,
		AdditionalNo: &addNo,
		CountryCode:  "SA",
	}

	bankAccount := &domain.EmployeeBankAccount{
		BankNameEn:    "Al Rajhi Bank",
		BankNameAr:    "مصرف الراجحي",
		IBAN:          "SA4420000001234567890123",
		SwiftBIC:      "RJHI",
		IsPrimary:     true,
		AccountStatus: "ACTIVE",
	}

	createdEmp, err := svc.OnboardEmployee(ctx, emp, ident, contact, bankAccount)
	if err != nil {
		t.Fatalf("failed to onboard employee: %v", err)
	}
	if createdEmp.ID == uuid.Nil {
		t.Fatalf("expected valid employee id, got nil")
	}

	// 3. Verify Position Transitioned to FILLED
	updatedPos, err := repo.GetPositionByID(ctx, leID, pos.ID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}
	if updatedPos.Status != domain.PositionFilled {
		t.Errorf("expected position status FILLED after hire, got %s", updatedPos.Status)
	}

	// 4. Verify Identity Masking
	identifiers, err := repo.ListIdentifiers(ctx, createdEmp.ID)
	if err != nil {
		t.Fatalf("failed to list identifiers: %v", err)
	}
	if len(identifiers) == 0 {
		t.Fatalf("expected at least 1 identifier")
	}
	if identifiers[0].IdentifierValueMasked != "1087****21" {
		t.Errorf("expected masked ID '1087****21', got '%s'", identifiers[0].IdentifierValueMasked)
	}

	// 5. Verify Saudi National Address Contact
	savedContact, err := repo.GetContact(ctx, createdEmp.ID)
	if err != nil {
		t.Fatalf("failed to get contact: %v", err)
	}
	if savedContact.City != "Riyadh" || *savedContact.BuildingNo != "4251" {
		t.Errorf("expected Riyadh building 4251, got %s building %v", savedContact.City, savedContact.BuildingNo)
	}

	// 6. Verify WPS Bank Account & IBAN Masking
	banks, err := repo.ListBankAccounts(ctx, createdEmp.ID)
	if err != nil {
		t.Fatalf("failed to list bank accounts: %v", err)
	}
	if len(banks) == 0 {
		t.Fatalf("expected bank account to be saved")
	}
	if banks[0].IBANMasked != "SA4420000001****0123" {
		t.Errorf("expected masked IBAN 'SA4420000001****0123', got '%s'", banks[0].IBANMasked)
	}
}

func TestContract_CreationAndVersioning(t *testing.T) {
	hrMod, _, leID, _, ctx := setupHRTest(t)
	svc := hrMod.HRService
	repo := hrMod.Repo

	// 1. Onboard Employee for Contract
	emp := &domain.Employee{
		LegalEntityID:          leID,
		EmployeeNumber:         "EMP-CNT-" + uuid.New().String()[:6],
		FirstNameEn:            "Sarah",
		LastNameEn:             "Al-Ghamdi",
		FirstNameAr:            "سارة",
		LastNameAr:             "الغامدي",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               time.Now(),
		EmploymentStatus:       domain.EmploymentStatusActive,
	}
	createdEmp, err := svc.OnboardEmployee(ctx, emp, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create employee: %v", err)
	}

	// 2. Create Qiwa-compliant Contract
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	reason1 := "INITIAL_HIRE"

	contract, version, err := svc.CreateContract(ctx, &domain.Contract{
		LegalEntityID:       leID,
		EmployeeID:          createdEmp.ID,
		ContractNumber:      "QIWA-TEST-" + uuid.New().String()[:8],
		ContractType:        domain.ContractUnspecifiedTerm,
		StartDate:           startDate,
		ProbationDays:       90,
		NoticePeriodDays:    60,
		WorkLocation:        "Riyadh HQ",
		WorkingHoursPerWeek: decimal.NewFromInt(40),
		Status:              domain.ContractActive,
	}, &domain.ContractVersion{
		EffectiveFrom:      startDate,
		BasicSalary:        decimal.NewFromInt(12000),
		HousingAllowance:   decimal.NewFromInt(3000),
		TransportAllowance: decimal.NewFromInt(1000),
		OtherAllowances:    decimal.NewFromInt(500),
		ChangeReason:       &reason1,
	})
	if err != nil {
		t.Fatalf("failed to create contract: %v", err)
	}

	// Verify Gross Salary Calculation (12000 + 3000 + 1000 + 500 = 16500)
	expectedGross := decimal.NewFromInt(16500)
	if !version.GrossSalary.Equal(expectedGross) {
		t.Errorf("expected gross salary %s, got %s", expectedGross.String(), version.GrossSalary.String())
	}
	if version.VersionNumber != 1 {
		t.Errorf("expected version 1, got %d", version.VersionNumber)
	}

	// 3. Amend Contract with Salary Increment (Version 2)
	amendDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	reason2 := "ANNUAL_MERIT_INCREASE"

	v2, err := svc.AmendContractVersion(ctx, leID, contract.ID, &domain.ContractVersion{
		EffectiveFrom:      amendDate,
		BasicSalary:        decimal.NewFromInt(15000),
		HousingAllowance:   decimal.NewFromInt(3750),
		TransportAllowance: decimal.NewFromInt(1250),
		OtherAllowances:    decimal.NewFromInt(500),
		ChangeReason:       &reason2,
	})
	if err != nil {
		t.Fatalf("failed to amend contract version: %v", err)
	}

	expectedGrossV2 := decimal.NewFromInt(20500)
	if !v2.GrossSalary.Equal(expectedGrossV2) {
		t.Errorf("expected v2 gross salary %s, got %s", expectedGrossV2.String(), v2.GrossSalary.String())
	}
	if v2.VersionNumber != 2 {
		t.Errorf("expected version 2, got %d", v2.VersionNumber)
	}

	// 4. Verify Version History
	versions, err := repo.ListContractVersions(ctx, contract.ID)
	if err != nil {
		t.Fatalf("failed to list contract versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions in history, got %d", len(versions))
	}

	// 5. Verify Employee Contract Query
	fetchedContract, latestVer, err := svc.GetEmployeeContract(ctx, leID, createdEmp.ID)
	if err != nil {
		t.Fatalf("failed to get employee contract: %v", err)
	}
	if fetchedContract.ID != contract.ID {
		t.Errorf("expected contract id %s, got %s", contract.ID, fetchedContract.ID)
	}
	if latestVer.VersionNumber != 2 {
		t.Errorf("expected latest version 2, got %d", latestVer.VersionNumber)
	}
}

func TestPosition_HeadcountAndTermination(t *testing.T) {
	hrMod, _, leID, deptID, ctx := setupHRTest(t)
	svc := hrMod.HRService
	repo := hrMod.Repo

	// 1. Create Job & Position
	job, err := svc.CreateJob(ctx, &domain.Job{
		LegalEntityID: leID,
		Code:          "JOB-FIN-" + uuid.New().String()[:6],
		TitleEn:       "Senior Financial Analyst",
		TitleAr:       "محلل مالي أول",
		JobFamily:     domain.JobFamilyFinance,
		JobLevel:      domain.JobLevelSenior,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	pos, err := svc.CreatePosition(ctx, &domain.Position{
		LegalEntityID: leID,
		Code:          "POS-FIN-" + uuid.New().String()[:6],
		JobID:         job.ID,
		DepartmentID:  deptID,
		Headcount:     1,
		Status:        domain.PositionVacant,
	})
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// 2. Hire employee into position
	emp, err := svc.OnboardEmployee(ctx, &domain.Employee{
		LegalEntityID:          leID,
		EmployeeNumber:         "EMP-FIN-" + uuid.New().String()[:6],
		FirstNameEn:            "Tariq",
		LastNameEn:             "Mansour",
		FirstNameAr:            "طارق",
		LastNameAr:             "منصور",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               time.Now(),
		EmploymentStatus:       domain.EmploymentStatusActive,
		PrimaryPositionID:      &pos.ID,
	}, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to onboard employee: %v", err)
	}

	filledPos, err := repo.GetPositionByID(ctx, leID, pos.ID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}
	if filledPos.Status != domain.PositionFilled {
		t.Errorf("expected position FILLED, got %s", filledPos.Status)
	}

	// 3. Terminate employee -> Position reverts to VACANT
	termEmp, err := svc.TerminateEmployee(ctx, leID, emp.ID)
	if err != nil {
		t.Fatalf("failed to terminate employee: %v", err)
	}
	if termEmp.EmploymentStatus != domain.EmploymentStatusTerminated {
		t.Errorf("expected employee status TERMINATED, got %s", termEmp.EmploymentStatus)
	}

	vacantPos, err := repo.GetPositionByID(ctx, leID, pos.ID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}
	if vacantPos.Status != domain.PositionVacant {
		t.Errorf("expected position status to revert to VACANT, got %s", vacantPos.Status)
	}
}

func TestSaudization_Calculation(t *testing.T) {
	hrMod, _, leID, _, ctx := setupHRTest(t)
	svc := hrMod.HRService

	// 1. Onboard 3 Saudis
	for i := 0; i < 3; i++ {
		_, err := svc.OnboardEmployee(ctx, &domain.Employee{
			LegalEntityID:          leID,
			EmployeeNumber:         "SA-" + uuid.New().String()[:6],
			FirstNameEn:            "Saudi",
			LastNameEn:             "Colleague",
			FirstNameAr:            "زميل",
			LastNameAr:             "سعودي",
			NationalityCountryCode: "SA",
			IsSaudiNational:        true,
			HireDate:               time.Now(),
			EmploymentStatus:       domain.EmploymentStatusActive,
		}, nil, nil, nil)
		if err != nil {
			t.Fatalf("failed to create saudi employee: %v", err)
		}
	}

	// 2. Onboard 1 Expatriate
	_, err := svc.OnboardEmployee(ctx, &domain.Employee{
		LegalEntityID:          leID,
		EmployeeNumber:         "EXP-" + uuid.New().String()[:6],
		FirstNameEn:            "Expat",
		LastNameEn:             "Colleague",
		FirstNameAr:            "زميل",
		LastNameAr:             "وافد",
		NationalityCountryCode: "EG",
		IsSaudiNational:        false,
		HireDate:               time.Now(),
		EmploymentStatus:       domain.EmploymentStatusActive,
	}, nil, nil, nil)
	if err != nil {
		t.Fatalf("failed to create expat employee: %v", err)
	}

	// 3. Compute Saudization
	metrics, err := svc.GetSaudizationMetrics(ctx, leID)
	if err != nil {
		t.Fatalf("failed to get saudization metrics: %v", err)
	}

	if metrics.SaudiCount < 3 {
		t.Errorf("expected at least 3 saudi employees, got %d", metrics.SaudiCount)
	}
	if metrics.ExpatCount < 1 {
		t.Errorf("expected at least 1 expat employee, got %d", metrics.ExpatCount)
	}
	if metrics.TotalCount < 4 {
		t.Errorf("expected at least 4 total employees, got %d", metrics.TotalCount)
	}
	if metrics.SaudizationRatioPercent.LessThan(decimal.NewFromFloat(50.0)) {
		t.Errorf("expected >= 50%% saudization ratio, got %s%%", metrics.SaudizationRatioPercent.String())
	}
	if metrics.NitaqatTier == "" {
		t.Errorf("expected valid Nitaqat tier, got empty")
	}
}

func TestHR_SeedDefaultHRData(t *testing.T) {
	hrMod, _, leID, _, ctx := setupHRTest(t)

	err := hrMod.SeedDefaultHRData(ctx, leID)
	if err != nil {
		t.Fatalf("failed to seed default HR data: %v", err)
	}

	// Verify seeded jobs
	jobs, err := hrMod.HRService.ListJobs(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list jobs: %v", err)
	}
	if len(jobs) == 0 {
		t.Fatalf("expected seeded jobs")
	}

	// Verify seeded positions
	positions, err := hrMod.HRService.ListPositions(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list positions: %v", err)
	}
	if len(positions) == 0 {
		t.Fatalf("expected seeded positions")
	}

	// Verify seeded employees
	employees, err := hrMod.HRService.ListEmployees(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list employees: %v", err)
	}
	if len(employees) == 0 {
		t.Fatalf("expected seeded employees")
	}

	// Verify seeded contracts
	contracts, err := hrMod.HRService.ListContracts(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list contracts: %v", err)
	}
	if len(contracts) == 0 {
		t.Fatalf("expected seeded contracts")
	}

	// Verify Saudization metrics
	metrics, err := hrMod.HRService.GetSaudizationMetrics(ctx, leID)
	if err != nil {
		t.Fatalf("failed to get saudization metrics: %v", err)
	}
	if metrics.TotalCount == 0 {
		t.Fatalf("expected positive total count")
	}
}

func TestEmployee_OnboardingRollsBackWhenIdentifierFails(t *testing.T) {
	hrMod, db, leID, _, ctx := setupHRTest(t)
	defer db.Close()

	employeeNumber := "ROLLBACK-EMP-" + uuid.New().String()[:8]
	_, err := hrMod.HRService.OnboardEmployee(ctx, &domain.Employee{
		LegalEntityID:          leID,
		EmployeeNumber:         employeeNumber,
		FirstNameEn:            "Rollback",
		LastNameEn:             "Employee",
		FirstNameAr:            "موظف",
		LastNameAr:             "تراجع",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               time.Now(),
		EmploymentStatus:       domain.EmploymentStatusActive,
	}, &domain.EmployeeIdentifier{
		IdentifierType:  domain.IdentifierType("INVALID_TYPE"),
		IdentifierValue: "ROLLBACK-ID-" + uuid.New().String(),
		IsPrimary:       true,
	}, nil, nil)
	if err == nil {
		t.Fatal("expected employee identifier constraint failure")
	}

	var count int
	if err := db.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM hr.employees WHERE legal_entity_id = $1 AND employee_number = $2",
		leID, employeeNumber,
	).Scan(&count); err != nil {
		t.Fatalf("failed to verify employee rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected employee rollback, found %d rows", count)
	}
}
