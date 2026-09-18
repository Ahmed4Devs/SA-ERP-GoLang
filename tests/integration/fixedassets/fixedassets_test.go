package fixedassets_test

import (
	"context"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/fixedassets"
	"erp/internal/fixedassets/domain"
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

func setupFixedAssetsTest(t *testing.T) (*fixedassets.Module, *hr.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	accMod := accounting.NewModule(cfg, db)
	hrMod := hr.NewModule(cfg, db)
	faMod := fixedassets.NewModule(cfg, db)

	// 1. Seed Organization
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Seed default Chart of Accounts and Fiscal Calendar
	if err := accMod.SeedDefaultCOA(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed default COA: %v", err)
	}

	// 3. Seed default Fixed Assets data (categories, locations, books)
	if err := faMod.SeedDefaultFixedAssetsData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed default fixed assets data: %v", err)
	}

	// 3. Admin User
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	} else {
		createdAdmin, _ := iamMod.AuthService.SeedAdmin(ctx, "admin", "admin@erp.sa", "System Administrator", "Admin@ERP123!")
		if createdAdmin != nil {
			adminID = createdAdmin.ID
		}
	}

	// 4. Onboard an Employee for Custody Assignment
	hireDate := time.Now().AddDate(-1, 0, 0)
	emp, err := hrMod.HRService.OnboardEmployee(ctx, &hrdomain.Employee{
		LegalEntityID:          le.ID,
		EmployeeNumber:         "EMP-FA-" + uuid.New().String()[:6],
		FirstNameEn:            "Nasser",
		LastNameEn:             "Al-Ghamdi",
		FirstNameAr:            "ناصر",
		LastNameAr:             "الغامدي",
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
		t.Fatalf("failed to onboard employee: %v", err)
	}

	return faMod, hrMod, db, le.ID, adminID, emp.ID, ctx
}

// Test 1: Asset Registration and Automatic Schedule Projection
func TestFixedAssets_AssetRegistrationAndScheduleProjection(t *testing.T) {
	faMod, _, _, leID, _, _, ctx := setupFixedAssetsTest(t)

	cats, err := faMod.Service.ListAssetCategories(ctx, leID)
	if err != nil || len(cats) == 0 {
		t.Fatalf("failed to list categories: %v", err)
	}

	books, err := faMod.Service.ListDepreciationBooks(ctx, leID)
	if err != nil || len(books) == 0 {
		t.Fatalf("failed to list books: %v", err)
	}

	assetNumber := "TEST-AST-" + uuid.New().String()[:6]
	cost := decimal.NewFromInt(120000)
	salvage := decimal.NewFromInt(12000)
	months := 48

	asset := domain.Asset{
		LegalEntityID:      leID,
		CategoryID:         cats[0].ID,
		AssetNumber:        assetNumber,
		NameEn:             "Fleet Delivery Truck",
		NameAr:             "شاحنة توصيل الأسطول",
		AcquisitionDate:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		InServiceDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AcquisitionCost:    cost,
		SalvageValue:       salvage,
		UsefulLifeMonths:   months,
		DepreciationMethod: "STRAIGHT_LINE",
	}

	created, err := faMod.Service.RegisterAsset(ctx, asset)
	if err != nil {
		t.Fatalf("failed to register asset: %v", err)
	}

	if created.Status != "ACTIVE" {
		t.Errorf("expected status ACTIVE, got %s", created.Status)
	}
	if !created.CurrentNetBookValue.Equal(cost) {
		t.Errorf("expected initial NBV %s, got %s", cost, created.CurrentNetBookValue)
	}
	if !created.AccumulatedDepreciation.Equal(decimal.Zero) {
		t.Errorf("expected initial accumulated depreciation 0, got %s", created.AccumulatedDepreciation)
	}

	// Verify pre-computed schedule
	scheds, err := faMod.Service.GetAssetSchedule(ctx, created.ID, books[0].ID)
	if err != nil {
		t.Fatalf("failed to get asset schedule: %v", err)
	}
	if len(scheds) != months {
		t.Fatalf("expected %d monthly schedules, got %d", months, len(scheds))
	}

	// Monthly depr = (120,000 - 12,000) / 48 = 2,250 SAR
	expectedMonthly := decimal.NewFromInt(2250)
	if !scheds[0].DepreciationAmount.Equal(expectedMonthly) {
		t.Errorf("expected monthly depr %s, got %s", expectedMonthly, scheds[0].DepreciationAmount)
	}

	lastSched := scheds[len(scheds)-1]
	if !lastSched.ClosingNetBookValue.Equal(salvage) {
		t.Errorf("expected final closing NBV equal to salvage %s, got %s", salvage, lastSched.ClosingNetBookValue)
	}
}

// Test 2: Straight-Line Depreciation Math & Boundary Conditions
func TestFixedAssets_StraightLineCalculation(t *testing.T) {
	cost := decimal.NewFromInt(120000)
	salvage := decimal.NewFromInt(12000)
	lifeMonths := 48

	// Standard monthly calculation: (120,000 - 12,000) / 48 = 2,250
	nbv1 := decimal.NewFromInt(120000)
	depr1 := domain.CalculateStraightLineDepreciation(cost, salvage, lifeMonths, nbv1)
	if !depr1.Equal(decimal.NewFromInt(2250)) {
		t.Errorf("expected depr 2250, got %s", depr1)
	}

	// Capping at salvage value: NBV is 13,000, salvage is 12,000 -> max allowed is 1,000
	nbvNearSalvage := decimal.NewFromInt(13000)
	deprCapped := domain.CalculateStraightLineDepreciation(cost, salvage, lifeMonths, nbvNearSalvage)
	if !deprCapped.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("expected capped depr 1000, got %s", deprCapped)
	}

	// Already at or below salvage value -> 0
	nbvAtSalvage := decimal.NewFromInt(12000)
	deprZero := domain.CalculateStraightLineDepreciation(cost, salvage, lifeMonths, nbvAtSalvage)
	if !deprZero.Equal(decimal.Zero) {
		t.Errorf("expected depr 0 at salvage, got %s", deprZero)
	}
}

// Test 3: Monthly Depreciation Run and Posting to Ledger
func TestFixedAssets_MonthlyDepreciationRunAndPosting(t *testing.T) {
	faMod, _, _, leID, adminID, _, ctx := setupFixedAssetsTest(t)

	books, err := faMod.Service.ListDepreciationBooks(ctx, leID)
	if err != nil || len(books) == 0 {
		t.Fatalf("failed to list books: %v", err)
	}
	bookID := books[0].ID

	year := 2026
	month := 9

	// 1. Run Monthly Depreciation
	run, lines, err := faMod.Service.RunMonthlyDepreciation(ctx, leID, bookID, year, month, &adminID)
	if err != nil {
		t.Fatalf("RunMonthlyDepreciation failed: %v", err)
	}

	if run.Status != "CALCULATED" {
		t.Errorf("expected status CALCULATED, got %s", run.Status)
	}
	if len(lines) == 0 {
		t.Fatalf("expected at least 1 depreciation line for seeded assets")
	}
	if !run.TotalDepreciationAmount.GreaterThan(decimal.Zero) {
		t.Errorf("expected total depreciation amount > 0")
	}

	// 2. Post Depreciation Run
	postedRun, err := faMod.Service.PostDepreciationRun(ctx, run.ID, &adminID)
	if err != nil {
		t.Fatalf("PostDepreciationRun failed: %v", err)
	}
	if postedRun.Status != "POSTED" {
		t.Errorf("expected status POSTED, got %s", postedRun.Status)
	}

	// 3. Verify asset's NBV decreased
	assetID := lines[0].AssetID
	updatedAsset, err := faMod.Service.GetAsset(ctx, assetID)
	if err != nil {
		t.Fatalf("failed to fetch updated asset: %v", err)
	}
	if !updatedAsset.CurrentNetBookValue.Equal(lines[0].NetBookValueAfter) {
		t.Errorf("asset NBV mismatch: expected %s, got %s", lines[0].NetBookValueAfter, updatedAsset.CurrentNetBookValue)
	}
	if !updatedAsset.AccumulatedDepreciation.Equal(lines[0].AccumulatedAfter) {
		t.Errorf("asset accumulated depreciation mismatch: expected %s, got %s", lines[0].AccumulatedAfter, updatedAsset.AccumulatedDepreciation)
	}
}

// Test 4: Floor at Salvage Value and Auto-Transition to FULLY_DEPRECIATED
func TestFixedAssets_DepreciationFloorAtSalvage(t *testing.T) {
	faMod, _, _, leID, adminID, _, ctx := setupFixedAssetsTest(t)

	cats, _ := faMod.Service.ListAssetCategories(ctx, leID)
	books, _ := faMod.Service.ListDepreciationBooks(ctx, leID)

	// Create an asset that is 500 SAR away from salvage value (Salvage: 5,000, NBV: 5,500)
	asset := domain.Asset{
		LegalEntityID:      leID,
		CategoryID:         cats[0].ID,
		AssetNumber:        "NEAR-SALVAGE-" + uuid.New().String()[:6],
		NameEn:             "Aging Machine",
		NameAr:             "آلة منتهية تقريباً",
		AcquisitionDate:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		InServiceDate:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		AcquisitionCost:    decimal.NewFromInt(50000),
		SalvageValue:       decimal.NewFromInt(5000),
		UsefulLifeMonths:   48,
		DepreciationMethod: "STRAIGHT_LINE",
	}

	created, err := faMod.Service.RegisterAsset(ctx, asset)
	if err != nil {
		t.Fatalf("failed to register asset: %v", err)
	}

	// Artificially update NBV to 5,500 (just 500 SAR above salvage)
	_, err = faMod.Repo.UpdateFixedAssetBookValue(ctx, created.ID, decimal.NewFromInt(5500), decimal.NewFromInt(44500))
	if err != nil {
		t.Fatalf("failed to set NBV near salvage: %v", err)
	}

	// Run depreciation for next month
	run, lines, err := faMod.Service.RunMonthlyDepreciation(ctx, leID, books[0].ID, 2026, 10, &adminID)
	if err != nil {
		t.Fatalf("RunMonthlyDepreciation failed: %v", err)
	}

	var foundLine *domain.DepreciationLine
	for _, l := range lines {
		if l.AssetID == created.ID {
			foundLine = &l
			break
		}
	}
	if foundLine == nil {
		t.Fatalf("created asset was not included in depreciation run")
	}

	// Should be capped exactly to 500 SAR to meet salvage value 5,000 SAR
	if !foundLine.DepreciationAmount.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected capped depreciation amount 500, got %s", foundLine.DepreciationAmount)
	}
	if !foundLine.NetBookValueAfter.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected NBV after 5000, got %s", foundLine.NetBookValueAfter)
	}

	// Post the run and check transition to FULLY_DEPRECIATED
	_, err = faMod.Service.PostDepreciationRun(ctx, run.ID, &adminID)
	if err != nil {
		t.Fatalf("post run failed: %v", err)
	}

	postAsset, _ := faMod.Service.GetAsset(ctx, created.ID)
	if postAsset.Status != "FULLY_DEPRECIATED" {
		t.Errorf("expected asset status FULLY_DEPRECIATED, got %s", postAsset.Status)
	}
}

// Test 5: Asset Location and Custodian Transfer
func TestFixedAssets_AssetTransfer(t *testing.T) {
	faMod, _, _, leID, adminID, empID, ctx := setupFixedAssetsTest(t)

	cats, _ := faMod.Service.ListAssetCategories(ctx, leID)
	locs, _ := faMod.Service.ListAssetLocations(ctx, leID)

	if len(locs) < 2 {
		t.Fatalf("need at least 2 locations for transfer test")
	}

	asset := domain.Asset{
		LegalEntityID:      leID,
		CategoryID:         cats[0].ID,
		LocationID:         &locs[0].ID,
		AssetNumber:        "TRF-AST-" + uuid.New().String()[:6],
		NameEn:             "Laptop MacBook Pro",
		NameAr:             "حاسب ماك بوك برو",
		AcquisitionDate:    time.Now(),
		InServiceDate:      time.Now(),
		AcquisitionCost:    decimal.NewFromInt(12000),
		SalvageValue:       decimal.Zero,
		UsefulLifeMonths:   36,
		DepreciationMethod: "STRAIGHT_LINE",
	}

	created, err := faMod.Service.RegisterAsset(ctx, asset)
	if err != nil {
		t.Fatalf("failed to register asset: %v", err)
	}

	// Transfer to location 2 and assign to employee
	transfer := domain.Transfer{
		LegalEntityID:  leID,
		AssetID:        created.ID,
		TransferDate:   time.Now(),
		FromLocationID: &locs[0].ID,
		ToLocationID:   locs[1].ID,
		ToEmployeeID:   &empID,
		Reason:         "Employee re-assigned to Jeddah office",
		TransferredBy:  &adminID,
	}

	savedTransfer, err := faMod.Service.TransferAsset(ctx, transfer)
	if err != nil {
		t.Fatalf("TransferAsset failed: %v", err)
	}
	if savedTransfer.Status != "COMPLETED" {
		t.Errorf("expected transfer status COMPLETED, got %s", savedTransfer.Status)
	}

	// Verify asset has new location
	updated, _ := faMod.Service.GetAsset(ctx, created.ID)
	if updated.LocationID == nil || *updated.LocationID != locs[1].ID {
		t.Errorf("asset location was not updated to location 2")
	}
}

// Test 6: Asset Disposal with Gain/Loss Calculation
func TestFixedAssets_AssetDisposal(t *testing.T) {
	faMod, _, _, leID, adminID, _, ctx := setupFixedAssetsTest(t)

	cats, _ := faMod.Service.ListAssetCategories(ctx, leID)

	// Asset with Cost = 100,000 SAR, Accumulated Depr = 30,000 SAR, NBV = 70,000 SAR
	asset := domain.Asset{
		LegalEntityID:      leID,
		CategoryID:         cats[0].ID,
		AssetNumber:        "DISP-AST-" + uuid.New().String()[:6],
		NameEn:             "Delivery Van",
		NameAr:             "فان توزيع",
		AcquisitionDate:    time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		InServiceDate:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		AcquisitionCost:    decimal.NewFromInt(100000),
		SalvageValue:       decimal.NewFromInt(10000),
		UsefulLifeMonths:   48,
		DepreciationMethod: "STRAIGHT_LINE",
	}

	created, err := faMod.Service.RegisterAsset(ctx, asset)
	if err != nil {
		t.Fatalf("failed to register asset: %v", err)
	}

	_, err = faMod.Repo.UpdateFixedAssetBookValue(ctx, created.ID, decimal.NewFromInt(70000), decimal.NewFromInt(30000))
	if err != nil {
		t.Fatalf("failed to update book value: %v", err)
	}

	// Case: Sold for 85,000 SAR -> Gain = 85,000 - 70,000 = 15,000 SAR
	proceeds := decimal.NewFromInt(85000)
	disposal := domain.Disposal{
		LegalEntityID:  leID,
		AssetID:        created.ID,
		DisposalType:   "SALE",
		ProceedsAmount: proceeds,
		Reason:         "Fleet modernization",
		CreatedBy:      &adminID,
	}

	savedDisposal, err := faMod.Service.DisposeAsset(ctx, disposal)
	if err != nil {
		t.Fatalf("DisposeAsset failed: %v", err)
	}

	expectedGain := decimal.NewFromInt(15000)
	if !savedDisposal.GainOrLossAmount.Equal(expectedGain) {
		t.Errorf("expected gain of %s, got %s", expectedGain, savedDisposal.GainOrLossAmount)
	}

	// Verify asset is DISPOSED and NBV is 0
	disposedAsset, _ := faMod.Service.GetAsset(ctx, created.ID)
	if disposedAsset.Status != "DISPOSED" {
		t.Errorf("expected status DISPOSED, got %s", disposedAsset.Status)
	}
	if !disposedAsset.CurrentNetBookValue.Equal(decimal.Zero) {
		t.Errorf("expected current NBV 0, got %s", disposedAsset.CurrentNetBookValue)
	}

	// Verify dashboard reflects metrics
	dash, err := faMod.Service.GetDashboardSummary(ctx, leID)
	if err != nil {
		t.Fatalf("GetDashboardSummary failed: %v", err)
	}
	if dash.TotalHistoricalCost.LessThanOrEqual(decimal.Zero) {
		t.Errorf("expected TotalHistoricalCost > 0")
	}
}
