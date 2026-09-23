package tax_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/tax"
	"erp/internal/tax/domain"
	"erp/internal/tax/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupTaxTest(t *testing.T) (*tax.Module, *accounting.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
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
	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	taxMod := tax.NewModule(cfg, db)

	// Isolate every test run so period and return uniqueness constraints remain repeatable.
	org, err := orgMod.OrgService.CreateOrganization(ctx, "ORG-TAX-"+uuid.NewString()[:8], "منظمة اختبار الضريبة", "Tax Test Organization")
	if err != nil {
		t.Fatalf("failed to create test organization: %v", err)
	}
	le, err := orgMod.OrgService.CreateLegalEntity(ctx, org.ID, "LE-TAX", "كيان اختبار الضريبة", "Tax Test Entity", "1010111111", "SAR")
	if err != nil {
		t.Fatalf("failed to create test legal entity: %v", err)
	}

	// Seed accounting COA & tax configuration
	if err := accMod.SeedDefaultCOA(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed COA: %v", err)
	}
	if err := taxMod.SeedDefaultTaxData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed tax data: %v", err)
	}

	// Get or create test user
	adminUser, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var userID uuid.UUID
	if err != nil || adminUser == nil {
		user, err := iamMod.AuthService.SeedAdmin(ctx, "tax_admin_"+uuid.NewString()[:6], "tax_admin@erp.sa", "Tax Admin", "Password@123")
		if err != nil {
			t.Fatalf("failed to seed admin user: %v", err)
		}
		userID = user.ID
	} else {
		userID = adminUser.ID
	}

	return taxMod, accMod, db, le.ID, userID, ctx
}

// 1. Test Effective-Dated Rate Lookup
func TestTaxEngine_EffectiveDatedRates(t *testing.T) {
	taxMod, _, _, leID, _, ctx := setupTaxTest(t)

	codes, err := taxMod.TaxService.ListTaxCodes(ctx, leID)
	if err != nil || len(codes) == 0 {
		t.Fatalf("failed to list tax codes: %v", err)
	}

	var stdCode *domain.TaxCode
	for i := range codes {
		if codes[i].Code == "VAT-STD-15" {
			stdCode = &codes[i]
			break
		}
	}
	if stdCode == nil {
		t.Fatal("standard tax code VAT-STD-15 not found")
	}
	if _, err := taxMod.TaxService.GetTaxCode(ctx, uuid.New(), stdCode.ID); !errors.Is(err, domain.ErrTaxCodeNotFound) {
		t.Fatalf("expected tenant-scoped tax code lookup, got %v", err)
	}

	// Historical Date: 2019-05-15 (before July 2020 VAT increase to 15%)
	histDate := time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC)
	histRate, err := taxMod.TaxService.GetActiveRate(ctx, leID, stdCode.ID, histDate)
	if err != nil {
		t.Fatalf("failed to get historical rate for 2019: %v", err)
	}
	if !histRate.RatePercent.Equal(decimal.NewFromInt(5)) {
		t.Errorf("expected historical rate 5%%, got: %s", histRate.RatePercent.String())
	}

	// Current Date: 2026-03-15 (after July 2020)
	currentDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	currentRate, err := taxMod.TaxService.GetActiveRate(ctx, leID, stdCode.ID, currentDate)
	if err != nil {
		t.Fatalf("failed to get current rate for 2026: %v", err)
	}
	if !currentRate.RatePercent.Equal(decimal.NewFromInt(15)) {
		t.Errorf("expected current rate 15%%, got: %s", currentRate.RatePercent.String())
	}
	if _, err := taxMod.TaxService.GetActiveRate(ctx, uuid.New(), stdCode.ID, currentDate); !errors.Is(err, domain.ErrTaxCodeNotFound) {
		t.Fatalf("expected tenant-scoped rate lookup, got %v", err)
	}
}

func TestTaxEngine_TaxCodeCreationRollsBackWhenInitialRateFails(t *testing.T) {
	taxMod, _, _, leID, _, ctx := setupTaxTest(t)
	code := "VAT-ROLLBACK-" + uuid.NewString()[:8]
	missingActorID := uuid.New()

	_, err := taxMod.TaxService.CreateTaxCode(ctx, domain.TaxCode{
		LegalEntityID:   leID,
		Code:            code,
		NameEn:          "Rollback Test Tax",
		NameAr:          "ضريبة اختبار التراجع",
		TaxCategoryCode: "S",
		TaxType:         "BOTH",
		IsRecoverable:   true,
		IsActive:        true,
	}, decimal.NewFromInt(15), time.Now(), &missingActorID)
	if err == nil {
		t.Fatal("expected the initial rate insert to fail for a missing actor")
	}

	codes, err := taxMod.TaxService.ListTaxCodes(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list tax codes: %v", err)
	}
	for _, taxCode := range codes {
		if taxCode.Code == code {
			t.Fatalf("tax code %s remained after its initial rate insert failed", code)
		}
	}
}

// 2. Test Exact VAT Arithmetic & Halala Rounding
func TestTaxEngine_VATCalculationExactRounding(t *testing.T) {
	taxMod, _, _, leID, _, ctx := setupTaxTest(t)

	codes, _ := taxMod.TaxService.ListTaxCodes(ctx, leID)
	var stdCodeID uuid.UUID
	for _, c := range codes {
		if c.Code == "VAT-STD-15" {
			stdCodeID = c.ID
			break
		}
	}

	// Test calculation:
	// Line 1: Qty 3 * UnitPrice 125.33 - Discount 10.00 = 365.99 Taxable
	// Tax (15%) = 365.99 * 0.15 = 54.8985 -> round(2) = 54.90
	// Gross = 365.99 + 54.90 = 420.89
	res, err := taxMod.TaxService.CalculateTax(ctx, ports.CalculateTaxRequest{
		LegalEntityID:   leID,
		TransactionDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Lines: []domain.TaxLineCalculationInput{
			{
				LineNumber:     1,
				TaxCodeID:      stdCodeID,
				Quantity:       decimal.NewFromInt(3),
				UnitPrice:      decimal.NewFromFloat(125.33),
				DiscountAmount: decimal.NewFromInt(10),
			},
		},
	})
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	expectedTaxable := decimal.NewFromFloat(365.99)
	expectedTax := decimal.NewFromFloat(54.90)
	expectedGross := decimal.NewFromFloat(420.89)

	if !res.TotalTaxableAmount.Equal(expectedTaxable) {
		t.Errorf("expected taxable %s, got %s", expectedTaxable, res.TotalTaxableAmount)
	}
	if !res.TotalTaxAmount.Equal(expectedTax) {
		t.Errorf("expected tax %s, got %s", expectedTax, res.TotalTaxAmount)
	}
	if !res.TotalGrossAmount.Equal(expectedGross) {
		t.Errorf("expected gross %s, got %s", expectedGross, res.TotalGrossAmount)
	}
}

// 3. Test ZATCA Validation Rules (Exempt, Zero-Rated, Out of Scope)
func TestTaxEngine_ZATCAValidationRules(t *testing.T) {
	// Rule A: Category 'E' (Exempt) with rate > 0% must be rejected
	_, err := domain.CalculateVATLine(domain.TaxLineCalculationInput{
		LineNumber:      1,
		TaxCategoryCode: "E",
		TaxRatePercent:  decimal.NewFromInt(15),
		Quantity:        decimal.NewFromInt(1),
		UnitPrice:       decimal.NewFromInt(100),
	})
	if !errors.Is(err, domain.ErrInvalidTaxRate) {
		t.Errorf("expected ErrInvalidTaxRate for exempt supply with rate > 0, got: %v", err)
	}

	// Rule B: Category 'E' without ExemptionReasonCode must be rejected
	_, err = domain.CalculateVATLine(domain.TaxLineCalculationInput{
		LineNumber:      1,
		TaxCategoryCode: "E",
		TaxRatePercent:  decimal.Zero,
		Quantity:        decimal.NewFromInt(1),
		UnitPrice:       decimal.NewFromInt(100),
	})
	if !errors.Is(err, domain.ErrExemptionReasonRequired) {
		t.Errorf("expected ErrExemptionReasonRequired for exempt supply without reason code, got: %v", err)
	}

	// Rule C: Category 'Z' (Zero-rated) without ZeroRateReasonCode must be rejected
	_, err = domain.CalculateVATLine(domain.TaxLineCalculationInput{
		LineNumber:      1,
		TaxCategoryCode: "Z",
		TaxRatePercent:  decimal.Zero,
		Quantity:        decimal.NewFromInt(1),
		UnitPrice:       decimal.NewFromInt(100),
	})
	if !errors.Is(err, domain.ErrZeroRateReasonRequired) {
		t.Errorf("expected ErrZeroRateReasonRequired for zero-rated supply without reason code, got: %v", err)
	}

	// Rule D: Valid Zero-rated supply with reason code
	validReason := "VATEX-SA-32"
	lineRes, err := domain.CalculateVATLine(domain.TaxLineCalculationInput{
		LineNumber:         1,
		TaxCategoryCode:    "Z",
		TaxRatePercent:     decimal.Zero,
		Quantity:           decimal.NewFromInt(10),
		UnitPrice:          decimal.NewFromInt(50),
		ZeroRateReasonCode: &validReason,
	})
	if err != nil {
		t.Fatalf("unexpected error for valid zero-rated line: %v", err)
	}
	if !lineRes.TaxAmount.IsZero() {
		t.Errorf("expected 0.00 tax amount for zero-rated supply, got: %s", lineRes.TaxAmount)
	}
}

// 4. Test Tax Transaction Posting and Idempotency Enforcement
func TestTaxEngine_TaxTransactionPostingAndIdempotency(t *testing.T) {
	taxMod, _, _, leID, userID, ctx := setupTaxTest(t)

	codes, _ := taxMod.TaxService.ListTaxCodes(ctx, leID)
	var stdCodeID uuid.UUID
	for _, c := range codes {
		if c.Code == "VAT-STD-15" {
			stdCodeID = c.ID
			break
		}
	}

	sourceID := uuid.New()
	partnerName := "شركة الأفق للتجارة المحدودة"
	partnerVat := "310123456700003"
	txDate := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

	req := ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      "SALES",
		SourceType:        "INVOICE",
		SourceID:          sourceID,
		TransactionNumber: "TX-INV-2026-0001",
		TransactionDate:   txDate,
		TransactionType:   "OUTPUT",
		PartnerName:       &partnerName,
		PartnerTaxNumber:  &partnerVat,
		CreatedBy:         &userID,
		Lines: []domain.TaxTransactionLine{
			{
				TaxCodeID:     stdCodeID,
				TaxableAmount: decimal.NewFromInt(1000), // 1000 SAR taxable -> 150 SAR VAT
			},
		},
	}

	// 1. Post transaction
	txRecord, err := taxMod.TaxService.PostTaxTransaction(ctx, req)
	if err != nil {
		t.Fatalf("post tax transaction failed: %v", err)
	}
	if txRecord.ID == uuid.Nil {
		t.Fatal("expected non-nil tax transaction ID")
	}
	if !txRecord.TotalTaxAmount.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected 150 SAR tax amount, got: %s", txRecord.TotalTaxAmount)
	}
	if len(txRecord.Lines) != 1 {
		t.Fatalf("expected 1 line item, got: %d", len(txRecord.Lines))
	}
	if _, err := taxMod.TaxService.GetTaxTransaction(ctx, uuid.New(), txRecord.ID); !errors.Is(err, domain.ErrTaxTransactionNotFound) {
		t.Fatalf("expected tenant-scoped tax transaction lookup, got %v", err)
	}

	// 2. Duplicate posting must fail with ErrDuplicateTaxTransaction
	_, err = taxMod.TaxService.PostTaxTransaction(ctx, req)
	if !errors.Is(err, domain.ErrDuplicateTaxTransaction) {
		t.Fatalf("expected ErrDuplicateTaxTransaction on duplicate post, got: %v", err)
	}
}

// 5. Test Closed Tax Period Enforcement
func TestTaxEngine_PeriodLocking(t *testing.T) {
	taxMod, _, _, leID, userID, ctx := setupTaxTest(t)

	codes, _ := taxMod.TaxService.ListTaxCodes(ctx, leID)
	var stdCodeID uuid.UUID
	for _, c := range codes {
		if c.Code == "VAT-STD-15" {
			stdCodeID = c.ID
			break
		}
	}

	// Create a closed historical tax period (e.g., Period 1 2025)
	pStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pEnd := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	pDue := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)

	period, err := taxMod.TaxService.CreatePeriod(ctx, domain.TaxPeriod{
		LegalEntityID: leID,
		PeriodNumber:  1,
		FiscalYear:    2025,
		PeriodType:    "MONTHLY",
		StartDate:     pStart,
		EndDate:       pEnd,
		DueDate:       pDue,
		Status:        domain.TaxPeriodStatusClosed,
	})
	if err != nil {
		t.Fatalf("failed to create historical period: %v", err)
	}
	if _, err := taxMod.TaxService.ClosePeriod(ctx, uuid.New(), period.ID); !errors.Is(err, domain.ErrTaxPeriodNotFound) {
		t.Fatalf("expected tenant-scoped period close, got %v", err)
	}

	// Ensure it is closed
	_, err = taxMod.TaxService.ClosePeriod(ctx, leID, period.ID)
	if err != nil {
		t.Fatalf("failed to close period: %v", err)
	}

	// Attempt to post a transaction in January 2025 (closed period)
	req := ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      "PURCHASES",
		SourceType:        "BILL",
		SourceID:          uuid.New(),
		TransactionNumber: "TX-BILL-2025-0001",
		TransactionDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		TransactionType:   "INPUT",
		CreatedBy:         &userID,
		Lines: []domain.TaxTransactionLine{
			{
				TaxCodeID:     stdCodeID,
				TaxableAmount: decimal.NewFromInt(500),
			},
		},
	}

	_, err = taxMod.TaxService.PostTaxTransaction(ctx, req)
	if !errors.Is(err, domain.ErrTaxPeriodClosed) {
		t.Fatalf("expected ErrTaxPeriodClosed when posting to closed period, got: %v", err)
	}
}

// 6. Test ZATCA VAT Return 16-Box Declaration Summary & Submission
func TestTaxEngine_VATReturnDeclarationBoxes(t *testing.T) {
	taxMod, _, _, leID, userID, ctx := setupTaxTest(t)

	codes, _ := taxMod.TaxService.ListTaxCodes(ctx, leID)
	var stdCode, zeroCode, exemptCode *domain.TaxCode
	for i := range codes {
		switch codes[i].Code {
		case "VAT-STD-15":
			stdCode = &codes[i]
		case "VAT-ZERO-0":
			zeroCode = &codes[i]
		case "VAT-EXEMPT-0":
			exemptCode = &codes[i]
		}
	}

	// Post in July 2026 (Period 7)
	julyDate := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	period, err := taxMod.TaxService.ListPeriods(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list periods: %v", err)
	}
	var julyPeriod *domain.TaxPeriod
	for i := range period {
		if period[i].FiscalYear == 2026 && period[i].PeriodNumber == 7 {
			julyPeriod = &period[i]
			break
		}
	}
	if julyPeriod == nil {
		t.Fatal("July 2026 tax period not found")
	}

	// A. Standard Rated Sales: 10,000 SAR taxable -> 1,500 SAR VAT (Box 1)
	_, err = taxMod.TaxService.PostTaxTransaction(ctx, ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      "SALES",
		SourceType:        "INVOICE",
		SourceID:          uuid.New(),
		TransactionNumber: "INV-JUL-001",
		TransactionDate:   julyDate,
		TransactionType:   "OUTPUT",
		Lines: []domain.TaxTransactionLine{
			{
				TaxCodeID:     stdCode.ID,
				TaxableAmount: decimal.NewFromInt(10000),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to post sales transaction: %v", err)
	}

	// B. Export Sales: 4,000 SAR (Box 4, Zero-rated)
	exportReason := "VATEX-SA-32"
	_, err = taxMod.TaxService.PostTaxTransaction(ctx, ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      "SALES",
		SourceType:        "INVOICE",
		SourceID:          uuid.New(),
		TransactionNumber: "INV-JUL-002",
		TransactionDate:   julyDate,
		TransactionType:   "OUTPUT",
		Lines: []domain.TaxTransactionLine{
			{
				TaxCodeID:          zeroCode.ID,
				TaxableAmount:      decimal.NewFromInt(4000),
				ZeroRateReasonCode: &exportReason,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to post export sales: %v", err)
	}

	// C. Exempt Sales: 2,000 SAR (Box 5, Exempt)
	exemptReason := "VATEX-SA-29"
	_, err = taxMod.TaxService.PostTaxTransaction(ctx, ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      "SALES",
		SourceType:        "INVOICE",
		SourceID:          uuid.New(),
		TransactionNumber: "INV-JUL-003",
		TransactionDate:   julyDate,
		TransactionType:   "OUTPUT",
		Lines: []domain.TaxTransactionLine{
			{
				TaxCodeID:           exemptCode.ID,
				TaxableAmount:       decimal.NewFromInt(2000),
				ExemptionReasonCode: &exemptReason,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to post exempt sales: %v", err)
	}

	// D. Standard Rated Purchases: 6,000 SAR taxable -> 900 SAR VAT (Box 7)
	_, err = taxMod.TaxService.PostTaxTransaction(ctx, ports.PostTaxTransactionRequest{
		LegalEntityID:     leID,
		SourceModule:      "PURCHASES",
		SourceType:        "BILL",
		SourceID:          uuid.New(),
		TransactionNumber: "BILL-JUL-001",
		TransactionDate:   julyDate,
		TransactionType:   "INPUT",
		Lines: []domain.TaxTransactionLine{
			{
				TaxCodeID:     stdCode.ID,
				TaxableAmount: decimal.NewFromInt(6000),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to post purchases: %v", err)
	}

	// Generate Draft VAT Return
	retDraft, err := taxMod.TaxService.GenerateVATReturnDraft(ctx, leID, julyPeriod.ID)
	if err != nil {
		t.Fatalf("failed to generate VAT return draft: %v", err)
	}
	if _, err := taxMod.TaxService.GenerateVATReturnDraft(ctx, uuid.New(), julyPeriod.ID); !errors.Is(err, domain.ErrTaxPeriodNotFound) {
		t.Fatalf("expected tenant-scoped VAT return generation, got %v", err)
	}

	// Verification of ZATCA 16-Boxes:
	// Box 1: Taxable 10,000, VAT 1,500
	if !retDraft.DeclarationBoxes.Box1StandardRatedSales.TaxableAmount.Equal(decimal.NewFromInt(10000)) {
		t.Errorf("Box 1 Taxable expected 10,000, got: %s", retDraft.DeclarationBoxes.Box1StandardRatedSales.TaxableAmount)
	}
	if !retDraft.DeclarationBoxes.Box1StandardRatedSales.VATAmount.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("Box 1 VAT expected 1,500, got: %s", retDraft.DeclarationBoxes.Box1StandardRatedSales.VATAmount)
	}

	// Box 4: Exports 4,000 SAR
	if !retDraft.DeclarationBoxes.Box4Exports.TaxableAmount.Equal(decimal.NewFromInt(4000)) {
		t.Errorf("Box 4 expected 4,000, got: %s", retDraft.DeclarationBoxes.Box4Exports.TaxableAmount)
	}

	// Box 5: Exempt 2,000 SAR
	if !retDraft.DeclarationBoxes.Box5ExemptSales.TaxableAmount.Equal(decimal.NewFromInt(2000)) {
		t.Errorf("Box 5 expected 2,000, got: %s", retDraft.DeclarationBoxes.Box5ExemptSales.TaxableAmount)
	}

	// Box 6: Total Sales = 10,000 + 4,000 + 2,000 = 16,000 SAR Taxable, 1,500 SAR VAT
	if !retDraft.TotalSalesTaxable.Equal(decimal.NewFromInt(16000)) {
		t.Errorf("Total Sales Taxable expected 16,000, got: %s", retDraft.TotalSalesTaxable)
	}
	if !retDraft.TotalSalesVAT.Equal(decimal.NewFromInt(1500)) {
		t.Errorf("Total Sales VAT expected 1,500, got: %s", retDraft.TotalSalesVAT)
	}

	// Box 7 & Box 12: Total Purchases = 6,000 SAR Taxable, 900 SAR VAT
	if !retDraft.TotalPurchasesTaxable.Equal(decimal.NewFromInt(6000)) {
		t.Errorf("Total Purchases Taxable expected 6,000, got: %s", retDraft.TotalPurchasesTaxable)
	}
	if !retDraft.TotalPurchasesVAT.Equal(decimal.NewFromInt(900)) {
		t.Errorf("Total Purchases VAT expected 900, got: %s", retDraft.TotalPurchasesVAT)
	}

	// Net VAT Due = Output VAT (1,500) - Input VAT (900) = 600 SAR
	if !retDraft.NetVATDue.Equal(decimal.NewFromInt(600)) {
		t.Errorf("Net VAT Due expected 600, got: %s", retDraft.NetVATDue)
	}

	// Submit VAT Return
	if _, err := taxMod.TaxService.SubmitVATReturn(ctx, uuid.New(), retDraft.ID, userID); !errors.Is(err, domain.ErrTaxReturnNotFound) {
		t.Fatalf("expected tenant-scoped VAT return submission, got %v", err)
	}
	submitted, err := taxMod.TaxService.SubmitVATReturn(ctx, leID, retDraft.ID, userID)
	if err != nil {
		t.Fatalf("failed to submit VAT return: %v", err)
	}
	if submitted.Status != "SUBMITTED" {
		t.Errorf("expected status SUBMITTED, got: %s", submitted.Status)
	}

	// Resubmitting must fail with ErrTaxReturnAlreadySubmitted
	_, err = taxMod.TaxService.GenerateVATReturnDraft(ctx, leID, julyPeriod.ID)
	if !errors.Is(err, domain.ErrTaxReturnAlreadySubmitted) {
		t.Fatalf("expected ErrTaxReturnAlreadySubmitted on resubmission, got: %v", err)
	}
}
