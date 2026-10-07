package gosi_test

import (
	"context"
	"erp/internal/gosi/application/services"
	"errors"
	"strings"
	"testing"
	"time"

	"erp/internal/gosi"
	"erp/internal/gosi/domain"
	"erp/internal/hr"
	hrdomain "erp/internal/hr/domain"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupGOSITest(t *testing.T) (*gosi.Module, *hr.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("integration database setup failed: %v", err)
	}
	t.Cleanup(db.Close)

	txManager := transaction.NewManager(db.Pool)
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	hrMod := hr.NewModule(cfg, db)
	gosiMod := gosi.NewModule(cfg, db)

	// 1. Create an isolated organization and legal entity for each test.
	org, err := orgMod.OrgService.CreateOrganization(ctx, "ORG-GOSI-"+uuid.NewString()[:8], "منظمة اختبار التأمينات", "GOSI Test Organization")
	if err != nil {
		t.Fatalf("failed to create test organization: %v", err)
	}
	le, err := orgMod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-GOSI", "كيان اختبار التأمينات", "GOSI Test Entity", "1010222222", "SAR")
	if err != nil {
		t.Fatalf("failed to create test legal entity: %v", err)
	}

	// 2. Seed default GOSI rules & versions
	if err := gosiMod.SeedDefaultGOSIData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed default GOSI data: %v", err)
	}

	// 3. Admin User
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	if err != nil || admin == nil {
		_, _ = iamMod.AuthService.SeedAdmin(ctx, "admin", "admin@erp.sa", "System Administrator", "Admin@ERP123!")
	}

	// 4. Onboard Saudi Employee
	hireDate := time.Now().AddDate(-2, 0, 0)
	saudiEmp, err := hrMod.HRService.OnboardEmployee(ctx, &hrdomain.Employee{
		LegalEntityID:          le.ID,
		EmployeeNumber:         "EMP-SA-" + uuid.New().String()[:6],
		FirstNameEn:            "Abdullah",
		LastNameEn:             "Al-Otaibi",
		FirstNameAr:            "عبدالله",
		LastNameAr:             "العتيبي",
		NationalityCountryCode: "SA",
		IsSaudiNational:        true,
		HireDate:               hireDate,
		EmploymentStatus:       hrdomain.EmploymentStatusActive,
	}, &hrdomain.EmployeeIdentifier{
		IdentifierType:  hrdomain.IdentifierNationalID,
		IdentifierValue: "10" + uuid.New().String()[:8],
		IsPrimary:       true,
		IsVerified:      true,
	}, nil, nil)
	if err != nil {
		t.Fatalf("failed to onboard saudi employee: %v", err)
	}

	// Register Saudi employee in GOSI
	override := decimal.NewFromInt(10000)
	_, err = gosiMod.GosiService.RegisterEmployee(ctx, domain.InsuranceProfile{
		LegalEntityID:             le.ID,
		EmployeeID:                saudiEmp.ID,
		GosiNumber:                "GOSI-" + uuid.New().String()[:8],
		NationalityScope:          "SAUDI",
		ContributionBasisOverride: &override,
		Notes:                     "Standard Saudi enrollee",
	})
	if err != nil {
		t.Fatalf("failed to register saudi insurance profile: %v", err)
	}

	// 5. Onboard Non-Saudi Employee
	nonSaudiEmp, err := hrMod.HRService.OnboardEmployee(ctx, &hrdomain.Employee{
		LegalEntityID:          le.ID,
		EmployeeNumber:         "EMP-EXP-" + uuid.New().String()[:6],
		FirstNameEn:            "Tariq",
		LastNameEn:             "Mahmoud",
		FirstNameAr:            "طارق",
		LastNameAr:             "محمود",
		NationalityCountryCode: "EG",
		IsSaudiNational:        false,
		HireDate:               hireDate,
		EmploymentStatus:       hrdomain.EmploymentStatusActive,
	}, &hrdomain.EmployeeIdentifier{
		IdentifierType:  hrdomain.IdentifierIqama,
		IdentifierValue: "20" + uuid.New().String()[:8],
		IsPrimary:       true,
		IsVerified:      true,
	}, nil, nil)
	if err != nil {
		t.Fatalf("failed to onboard non-saudi employee: %v", err)
	}

	// Register Non-Saudi employee in GOSI
	overrideNonSaudi := decimal.NewFromInt(8000)
	_, err = gosiMod.GosiService.RegisterEmployee(ctx, domain.InsuranceProfile{
		LegalEntityID:             le.ID,
		EmployeeID:                nonSaudiEmp.ID,
		GosiNumber:                "GOSI-" + uuid.New().String()[:8],
		NationalityScope:          "NON_SAUDI",
		ContributionBasisOverride: &overrideNonSaudi,
		Notes:                     "Standard Expat enrollee",
	})
	if err != nil {
		t.Fatalf("failed to register non-saudi insurance profile: %v", err)
	}

	return gosiMod, hrMod, db, le.ID, saudiEmp.ID, nonSaudiEmp.ID, ctx
}

// Test 1: Verify versioned rules are in DB and never hard-coded
func TestGOSI_RuleVersionsFromDB_NeverHardcoded(t *testing.T) {
	gosiMod, _, _, leID, _, _, ctx := setupGOSITest(t)

	rules, err := gosiMod.GosiService.ListContributionRules(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list contribution rules: %v", err)
	}
	if len(rules) < 3 {
		t.Fatalf("expected at least 3 rules (Annuity, OHRSI, SANED), got %d", len(rules))
	}

	activeVersions, err := gosiMod.GosiService.GetActiveRuleVersions(ctx, leID, time.Now())
	if err != nil {
		t.Fatalf("failed to get active rule versions: %v", err)
	}
	if len(activeVersions) < 3 {
		t.Fatalf("expected at least 3 active versions, got %d", len(activeVersions))
	}

	ruleCodes := make(map[string]bool)
	for _, v := range activeVersions {
		ruleCodes[v.RuleCode] = true
		if v.EmployerRate.IsZero() && v.EmployeeRate.IsZero() {
			t.Errorf("rule %s has zero for both employee and employer rates", v.RuleCode)
		}
	}

	if !ruleCodes["ANNUITY"] || !ruleCodes["OHRSI"] || !ruleCodes["SANED"] {
		t.Fatalf("missing required statutory rules, found: %+v", ruleCodes)
	}
}

// Test 2: Saudi national calculation with 10,000 SAR basis
// Annuity: 9% emp (900) + 9% er (900)
// OHRSI: 0% emp (0) + 2% er (200)
// SANED: 0.75% emp (75) + 0.75% er (75)
// Total Emp = 975 SAR, Total Er = 1,175 SAR, Total = 2,150 SAR
func TestGOSI_SaudiNationalCalculation(t *testing.T) {
	gosiMod, _, _, leID, saudiEmpID, _, ctx := setupGOSITest(t)

	input := domain.ContributionInput{
		LegalEntityID:     leID,
		EmployeeID:        saudiEmpID,
		PayrollDate:       time.Now(),
		NationalityCode:   "SA",
		ContributionBasis: decimal.NewFromInt(10000),
	}

	results, err := gosiMod.GosiService.Calculate(ctx, input)
	if err != nil {
		t.Fatalf("calculate failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 rules applied for Saudi national, got %d", len(results))
	}

	input.LegalEntityID = uuid.New()
	if _, err := gosiMod.GosiService.Calculate(ctx, input); !errors.Is(err, domain.ErrEmployeeScopeMismatch) {
		t.Fatalf("expected cross-entity employee calculation to fail with scope mismatch, got %v", err)
	}

	totalEmp := decimal.Zero
	totalEr := decimal.Zero
	for _, r := range results {
		totalEmp = totalEmp.Add(r.EmployeeAmount)
		totalEr = totalEr.Add(r.EmployerAmount)

		switch r.RuleVersion.RuleCode {
		case "ANNUITY":
			if !r.EmployeeAmount.Equal(decimal.NewFromInt(900)) {
				t.Errorf("ANNUITY emp amount expected 900, got %s", r.EmployeeAmount)
			}
			if !r.EmployerAmount.Equal(decimal.NewFromInt(900)) {
				t.Errorf("ANNUITY er amount expected 900, got %s", r.EmployerAmount)
			}
		case "OHRSI":
			if !r.EmployeeAmount.Equal(decimal.Zero) {
				t.Errorf("OHRSI emp amount expected 0, got %s", r.EmployeeAmount)
			}
			if !r.EmployerAmount.Equal(decimal.NewFromInt(200)) {
				t.Errorf("OHRSI er amount expected 200, got %s", r.EmployerAmount)
			}
		case "SANED":
			if !r.EmployeeAmount.Equal(decimal.NewFromFloat(75.00)) {
				t.Errorf("SANED emp amount expected 75, got %s", r.EmployeeAmount)
			}
			if !r.EmployerAmount.Equal(decimal.NewFromFloat(75.00)) {
				t.Errorf("SANED er amount expected 75, got %s", r.EmployerAmount)
			}
		}
	}

	expectedEmpTotal := decimal.NewFromFloat(975.00)
	expectedErTotal := decimal.NewFromFloat(1175.00)
	expectedGrandTotal := decimal.NewFromFloat(2150.00)

	if !totalEmp.Equal(expectedEmpTotal) {
		t.Errorf("total employee contribution expected %s, got %s", expectedEmpTotal, totalEmp)
	}
	if !totalEr.Equal(expectedErTotal) {
		t.Errorf("total employer contribution expected %s, got %s", expectedErTotal, totalEr)
	}
	if !totalEmp.Add(totalEr).Equal(expectedGrandTotal) {
		t.Errorf("grand total contribution expected %s, got %s", expectedGrandTotal, totalEmp.Add(totalEr))
	}
}

// Test 3: Non-Saudi national calculation with 8,000 SAR basis
// Annuity: excluded (0 SAR)
// SANED: excluded (0 SAR)
// OHRSI: 0% emp (0) + 2% er (160 SAR)
// Total Emp = 0 SAR, Total Er = 160 SAR, Total = 160 SAR
func TestGOSI_NonSaudiNationalCalculation(t *testing.T) {
	gosiMod, _, _, leID, _, nonSaudiEmpID, ctx := setupGOSITest(t)

	input := domain.ContributionInput{
		LegalEntityID:     leID,
		EmployeeID:        nonSaudiEmpID,
		PayrollDate:       time.Now(),
		NationalityCode:   "EG",
		ContributionBasis: decimal.NewFromInt(8000),
	}

	results, err := gosiMod.GosiService.Calculate(ctx, input)
	if err != nil {
		t.Fatalf("calculate failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 rule (OHRSI) applied for Non-Saudi, got %d", len(results))
	}

	res := results[0]
	if res.RuleVersion.RuleCode != "OHRSI" {
		t.Errorf("expected rule code OHRSI, got %s", res.RuleVersion.RuleCode)
	}
	if !res.EmployeeAmount.Equal(decimal.Zero) {
		t.Errorf("non-saudi employee amount expected 0, got %s", res.EmployeeAmount)
	}
	if !res.EmployerAmount.Equal(decimal.NewFromInt(160)) {
		t.Errorf("non-saudi employer amount expected 160, got %s", res.EmployerAmount)
	}
}

// Test 4: Basis Capping Invariants (1,500 SAR Min, 45,000 SAR Max)
func TestGOSI_BasisCapping_MinMax(t *testing.T) {
	gosiMod, _, _, leID, saudiEmpID, _, ctx := setupGOSITest(t)

	// Case A: Below Minimum (1,000 SAR basis -> capped to 1,500 SAR)
	inputBelowMin := domain.ContributionInput{
		LegalEntityID:     leID,
		EmployeeID:        saudiEmpID,
		PayrollDate:       time.Now(),
		NationalityCode:   "SA",
		ContributionBasis: decimal.NewFromInt(1000),
	}
	resBelow, err := gosiMod.GosiService.Calculate(ctx, inputBelowMin)
	if err != nil {
		t.Fatalf("calculate below min failed: %v", err)
	}
	for _, r := range resBelow {
		if !r.BasisApplied.Equal(decimal.NewFromInt(1500)) {
			t.Errorf("expected basis capped to min 1500, got %s for rule %s", r.BasisApplied, r.RuleVersion.RuleCode)
		}
	}

	// Case B: Above Maximum (60,000 SAR basis -> capped to 45,000 SAR)
	inputAboveMax := domain.ContributionInput{
		LegalEntityID:     leID,
		EmployeeID:        saudiEmpID,
		PayrollDate:       time.Now(),
		NationalityCode:   "SA",
		ContributionBasis: decimal.NewFromInt(60000),
	}
	resAbove, err := gosiMod.GosiService.Calculate(ctx, inputAboveMax)
	if err != nil {
		t.Fatalf("calculate above max failed: %v", err)
	}
	for _, r := range resAbove {
		if !r.BasisApplied.Equal(decimal.NewFromInt(45000)) {
			t.Errorf("expected basis capped to max 45000, got %s for rule %s", r.BasisApplied, r.RuleVersion.RuleCode)
		}
	}
}

// Test 5: Run monthly calculation for multiple employees and create export batch
func TestGOSI_MonthlyCalculationAndExportBatch(t *testing.T) {
	gosiMod, _, _, leID, _, _, ctx := setupGOSITest(t)

	year := time.Now().Year()
	month := int(time.Now().Month())

	// 1. Run monthly calculation
	calcs, err := gosiMod.GosiService.RunMonthlyCalculation(ctx, leID, year, month)
	if err != nil {
		t.Fatalf("RunMonthlyCalculation failed: %v", err)
	}
	if len(calcs) < 2 {
		t.Fatalf("expected at least 2 calculations (1 Saudi, 1 Expat), got %d", len(calcs))
	}

	// Re-running an open period replaces line details instead of duplicating them.
	recalcs, err := gosiMod.GosiService.RunMonthlyCalculation(ctx, leID, year, month)
	if err != nil {
		t.Fatalf("idempotent RunMonthlyCalculation failed: %v", err)
	}
	if len(recalcs) != len(calcs) {
		t.Fatalf("expected %d calculations after rerun, got %d", len(calcs), len(recalcs))
	}

	// 2. Fetch calculation details and lines
	calc0 := calcs[0]
	detCalc, lines, err := gosiMod.GosiService.GetCalculationDetails(ctx, leID, calc0.ID)
	if err != nil {
		t.Fatalf("GetCalculationDetails failed: %v", err)
	}
	if detCalc == nil || len(lines) == 0 {
		t.Fatalf("expected non-empty calculation and lines")
	}
	if len(lines) > 3 {
		t.Fatalf("monthly rerun duplicated calculation lines: got %d", len(lines))
	}
	if _, _, err := gosiMod.GosiService.GetCalculationDetails(ctx, uuid.New(), calc0.ID); !errors.Is(err, domain.ErrCalculationNotFound) {
		t.Fatalf("expected cross-entity calculation lookup to be hidden, got %v", err)
	}

	// 3. Create Export Batch
	batch, err := gosiMod.GosiService.CreateExportBatch(ctx, leID, year, month, nil)
	if err != nil {
		t.Fatalf("CreateExportBatch failed: %v", err)
	}
	if batch.BatchNumber == "" {
		t.Errorf("expected non-empty batch number")
	}
	if batch.EmployeeCount != len(calcs) {
		t.Errorf("batch subscriber count mismatch: expected %d, got %d", len(calcs), batch.EmployeeCount)
	}
	if batch.FileContent == nil || !strings.Contains(*batch.FileContent, "GOSIContributionReport") {
		t.Errorf("batch XML file content invalid or missing root tag")
	}
	if _, err := gosiMod.GosiService.CreateExportBatch(ctx, leID, year, month, nil); !errors.Is(err, domain.ErrPeriodAlreadyExported) {
		t.Fatalf("expected duplicate export to fail, got %v", err)
	}
	if _, err := gosiMod.GosiService.RunMonthlyCalculation(ctx, leID, year, month); !errors.Is(err, domain.ErrPeriodAlreadyExported) {
		t.Fatalf("expected exported period recalculation to fail, got %v", err)
	}

	// 4. Verify Dashboard Summary
	dash, err := gosiMod.GosiService.GetDashboardSummary(ctx, leID, year, month)
	if err != nil {
		t.Fatalf("GetDashboardSummary failed: %v", err)
	}
	if dash.SubscriberCount < 2 {
		t.Errorf("expected at least 2 subscribers in dashboard, got %d", dash.SubscriberCount)
	}
	if !dash.TotalContributions.GreaterThan(decimal.Zero) {
		t.Errorf("expected total contributions > 0")
	}
}

// Test 6: GOSIAdapter Interface per Section 27
func TestGOSI_AdapterInterfaceValidation(t *testing.T) {
	gosiMod, _, _, leID, _, _, ctx := setupGOSITest(t)

	year := time.Now().Year()
	month := int(time.Now().Month())

	_, err := gosiMod.GosiService.RunMonthlyCalculation(ctx, leID, year, month)
	if err != nil {
		t.Fatalf("run monthly calc failed: %v", err)
	}

	batch, err := gosiMod.GosiService.CreateExportBatch(ctx, leID, year, month, nil)
	if err != nil {
		t.Fatalf("create batch failed: %v", err)
	}

	adapter := gosiMod.Adapter

	// Test Export
	fileBytes, err := adapter.Export(ctx, batch.ID)
	if err != nil {
		t.Fatalf("adapter.Export failed: %v", err)
	}

	// Test Validate
	valRes, err := adapter.Validate(ctx, fileBytes)
	if err != nil {
		t.Fatalf("adapter.Validate failed: %v", err)
	}
	if !valRes.Valid {
		t.Errorf("expected valid XML file, errors: %+v", valRes.Errors)
	}

	// Local export never implies portal submission or acceptance.
	if _, err := adapter.Submit(ctx, batch.ID); !errors.Is(err, services.ErrManualSubmissionRequired) {
		t.Fatalf("expected manual submission required, got %v", err)
	}
	if _, err := adapter.Status(ctx, "invented-receipt"); !errors.Is(err, services.ErrStatusUnavailable) {
		t.Fatalf("expected unknown external status, got %v", err)
	}
	simulation := services.NewGOSIAdapter(gosiMod.Repo, services.ModeSimulation)
	result, err := simulation.Submit(ctx, batch.ID)
	if err != nil || result.Status != "SIMULATED_SUBMITTED" || result.ExternalSubmission {
		t.Fatalf("simulation result: %+v, %v", result, err)
	}
	status, err := simulation.Status(ctx, result.SubmissionID)
	if err != nil || status.Status != "SIMULATED_SUBMITTED" || status.Mode != "simulation" {
		t.Fatalf("simulation status: %+v, %v", status, err)
	}
	if _, err := simulation.Status(ctx, "SIMULATED:"+uuid.NewString()); err == nil {
		t.Fatal("unknown batch accepted")
	}
	stored, err := gosiMod.Repo.GetExportBatchByID(ctx, batch.ID)
	if err != nil || stored.Status != "GENERATED" || stored.SubmittedAt != nil || stored.ConfirmedAt != nil {
		t.Fatalf("simulation changed actual state: %+v, %v", stored, err)
	}
	official := services.NewGOSIAdapter(gosiMod.Repo, services.ModeOfficial)
	if _, err := official.Submit(ctx, batch.ID); !errors.Is(err, services.ErrOfficialIntegrationUnavailable) {
		t.Fatalf("official mode did not fail closed: %v", err)
	}
	for _, invalid := range [][]byte{
		[]byte("GOSIContributionReport LegalEntity"),
		[]byte("<GOSIContributionReport>"),
		append(append([]byte{}, fileBytes...), []byte("<Other/>")...),
		[]byte(strings.Replace(string(fileBytes), `grandTotal="`, `grandTotal="9`, 1)),
		[]byte(strings.Replace(string(fileBytes), `month="`, `month="99`, 1)),
		[]byte(strings.Replace(string(fileBytes), `urn:erp:gosi:export:v1`, `urn:sa:gosi:official`, 1)),
		append([]byte("<!DOCTYPE report>"), fileBytes...),
	} {
		result, err := adapter.Validate(ctx, invalid)
		if err != nil || result.Valid {
			t.Fatalf("invalid report accepted: %+v, %v", result, err)
		}
	}
}

func TestGOSI_MissingContributionBasisDoesNotInventSalary(t *testing.T) {
	mod, _, _, leID, saudiID, _, ctx := setupGOSITest(t)
	profile, err := mod.Repo.GetInsuranceProfileByEmployee(ctx, leID, saudiID)
	if err != nil {
		t.Fatal(err)
	}
	profile.ContributionBasisOverride = nil
	if _, err := mod.Repo.UpdateInsuranceProfile(ctx, *profile); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.GosiService.RunMonthlyCalculation(ctx, leID, time.Now().Year(), int(time.Now().Month())); err == nil {
		t.Fatal("missing contribution basis was replaced with fabricated salary")
	}
	calcs, err := mod.GosiService.ListCalculationsByPeriod(ctx, leID, time.Now().Year(), int(time.Now().Month()))
	if err != nil || len(calcs) != 0 {
		t.Fatalf("failed monthly calculation must roll back: %d, %v", len(calcs), err)
	}
}
