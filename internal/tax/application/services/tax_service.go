package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/tax/domain"
	"erp/internal/tax/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type taxService struct {
	repo ports.TaxRepository
}

func NewTaxService(repo ports.TaxRepository) ports.TaxService {
	return &taxService{
		repo: repo,
	}
}

// 1. Reference & Configuration
func (s *taxService) ListCategories(ctx context.Context) ([]domain.TaxCategory, error) {
	return s.repo.ListTaxCategories(ctx)
}

func (s *taxService) ListExemptionReasons(ctx context.Context) ([]domain.ExemptionReason, error) {
	return s.repo.ListExemptionReasons(ctx)
}

func (s *taxService) ListZeroRateReasons(ctx context.Context) ([]domain.ZeroRateReason, error) {
	return s.repo.ListZeroRateReasons(ctx)
}

func (s *taxService) ListTaxCodes(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxCode, error) {
	return s.repo.ListTaxCodesByEntity(ctx, legalEntityID)
}

func (s *taxService) GetTaxCode(ctx context.Context, id uuid.UUID) (*domain.TaxCode, error) {
	return s.repo.GetTaxCodeByID(ctx, id)
}

func (s *taxService) CreateTaxCode(ctx context.Context, req domain.TaxCode, initialRate decimal.Decimal, effectiveFrom time.Time, createdBy *uuid.UUID) (*domain.TaxCode, error) {
	cat, err := s.repo.GetTaxCategoryByCode(ctx, req.TaxCategoryCode)
	if err != nil || cat == nil {
		return nil, fmt.Errorf("invalid tax category: %s", req.TaxCategoryCode)
	}

	// Validate rate according to category
	catUpper := strings.ToUpper(req.TaxCategoryCode)
	if (catUpper == "E" || catUpper == "Z" || catUpper == "O") && !initialRate.IsZero() {
		return nil, domain.ErrInvalidTaxRate
	}

	createdCode, err := s.repo.CreateTaxCode(ctx, &req)
	if err != nil {
		return nil, err
	}

	if effectiveFrom.IsZero() {
		effectiveFrom = time.Now().UTC().Truncate(24 * time.Hour)
	}

	_, err = s.repo.CreateTaxRateVersion(ctx, &domain.TaxRateVersion{
		TaxCodeID:     createdCode.ID,
		RatePercent:   initialRate,
		EffectiveFrom: effectiveFrom,
		Description:   "Initial tax rate version",
		CreatedBy:     createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create initial tax rate version: %w", err)
	}

	createdCode.CurrentRatePercent = &initialRate
	return createdCode, nil
}

func (s *taxService) CreateRateVersion(ctx context.Context, rate domain.TaxRateVersion) (*domain.TaxRateVersion, error) {
	taxCode, err := s.repo.GetTaxCodeByID(ctx, rate.TaxCodeID)
	if err != nil {
		return nil, err
	}

	catUpper := strings.ToUpper(taxCode.TaxCategoryCode)
	if (catUpper == "E" || catUpper == "Z" || catUpper == "O") && !rate.RatePercent.IsZero() {
		return nil, domain.ErrInvalidTaxRate
	}

	return s.repo.CreateTaxRateVersion(ctx, &rate)
}

func (s *taxService) GetActiveRate(ctx context.Context, taxCodeID uuid.UUID, targetDate time.Time) (*domain.TaxRateVersion, error) {
	if targetDate.IsZero() {
		targetDate = time.Now().UTC()
	}
	return s.repo.GetActiveRateVersion(ctx, taxCodeID, targetDate)
}

// 2. Financial Calculation Engine
func (s *taxService) CalculateTax(ctx context.Context, req ports.CalculateTaxRequest) (*domain.TaxCalculationResult, error) {
	if len(req.Lines) == 0 {
		return nil, domain.ErrInvalidTaxTransaction
	}

	date := req.TransactionDate
	if date.IsZero() {
		date = time.Now().UTC()
	}

	resolvedLines := make([]domain.TaxLineCalculationInput, 0, len(req.Lines))
	for i, line := range req.Lines {
		if line.LineNumber == 0 {
			line.LineNumber = i + 1
		}

		// Lookup rate version if rate is not explicitly provided or if taxCodeID is given
		if line.TaxCodeID != uuid.Nil {
			taxCode, err := s.repo.GetTaxCodeByID(ctx, line.TaxCodeID)
			if err != nil {
				return nil, err
			}
			line.TaxCode = taxCode.Code
			line.TaxCategoryCode = taxCode.TaxCategoryCode

			rateVer, err := s.repo.GetActiveRateVersion(ctx, line.TaxCodeID, date)
			if err != nil {
				return nil, err
			}
			line.TaxRatePercent = rateVer.RatePercent
		}

		resolvedLines = append(resolvedLines, line)
	}

	return domain.CalculateVATDocument(resolvedLines)
}

// 3. Tax Transaction Engine
func (s *taxService) PostTaxTransaction(ctx context.Context, req ports.PostTaxTransactionRequest) (*domain.TaxTransaction, error) {
	// 1. Idempotency Check
	existing, err := s.repo.GetTaxTransactionBySource(ctx, req.LegalEntityID, req.SourceModule, req.SourceType, req.SourceID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrDuplicateTaxTransaction
	}

	// 2. Resolve Tax Period
	date := req.TransactionDate
	if date.IsZero() {
		date = time.Now().UTC()
	}

	period, err := s.repo.GetOpenPeriodForDate(ctx, req.LegalEntityID, date)
	if err != nil {
		return nil, domain.ErrTaxPeriodNotFound
	}
	if period.Status == domain.TaxPeriodStatusClosed {
		return nil, domain.ErrTaxPeriodClosed
	}

	// 3. Calculate and Validate Lines
	if len(req.Lines) == 0 {
		return nil, domain.ErrInvalidTaxTransaction
	}

	calcInputs := make([]domain.TaxLineCalculationInput, 0, len(req.Lines))
	for i, l := range req.Lines {
		rateVer, err := s.repo.GetActiveRateVersion(ctx, l.TaxCodeID, date)
		if err != nil {
			return nil, err
		}
		taxCode, err := s.repo.GetTaxCodeByID(ctx, l.TaxCodeID)
		if err != nil {
			return nil, err
		}

		calcInputs = append(calcInputs, domain.TaxLineCalculationInput{
			LineNumber:          i + 1,
			TaxCodeID:           l.TaxCodeID,
			TaxCode:             taxCode.Code,
			TaxCategoryCode:     taxCode.TaxCategoryCode,
			TaxRatePercent:      rateVer.RatePercent,
			Quantity:            decimal.NewFromInt(1), // Base multiplier
			UnitPrice:           l.TaxableAmount,
			DiscountAmount:      decimal.Zero,
			ExemptionReasonCode: l.ExemptionReasonCode,
			ZeroRateReasonCode:  l.ZeroRateReasonCode,
			IsRecoverable:       taxCode.IsRecoverable,
		})
	}

	calcDoc, err := domain.CalculateVATDocument(calcInputs)
	if err != nil {
		return nil, err
	}

	// 4. Construct Transaction Record
	txNumber := req.TransactionNumber
	if txNumber == "" {
		txNumber = fmt.Sprintf("TX-%s-%s", date.Format("20060102"), uuid.NewString()[:8])
	}

	txRecord := &domain.TaxTransaction{
		LegalEntityID:      req.LegalEntityID,
		TaxPeriodID:        &period.ID,
		SourceModule:       req.SourceModule,
		SourceType:         req.SourceType,
		SourceID:           req.SourceID,
		TransactionNumber:  txNumber,
		TransactionDate:    date,
		TransactionType:    req.TransactionType,
		PartnerID:          req.PartnerID,
		PartnerName:        req.PartnerName,
		PartnerTaxNumber:   req.PartnerTaxNumber,
		TotalTaxableAmount: calcDoc.TotalTaxableAmount,
		TotalTaxAmount:     calcDoc.TotalTaxAmount,
		TotalGrossAmount:   calcDoc.TotalGrossAmount,
		JournalEntryID:     req.JournalEntryID,
		Status:             "POSTED",
		CreatedBy:          req.CreatedBy,
		Lines:              make([]domain.TaxTransactionLine, 0, len(calcDoc.Lines)),
	}

	for _, l := range calcDoc.Lines {
		txRecord.Lines = append(txRecord.Lines, domain.TaxTransactionLine{
			LineNumber:          l.LineNumber,
			TaxCodeID:           l.TaxCodeID,
			TaxCategoryCode:     l.TaxCategoryCode,
			TaxRatePercent:      l.TaxRatePercent,
			TaxableAmount:       l.TaxableAmount,
			TaxAmount:           l.TaxAmount,
			GrossAmount:         l.GrossAmount,
			ExemptionReasonCode: l.ExemptionReasonCode,
			ZeroRateReasonCode:  l.ZeroRateReasonCode,
			IsRecoverable:       l.IsRecoverable,
		})
	}

	return s.repo.CreateTaxTransaction(ctx, txRecord)
}

func (s *taxService) GetTaxTransaction(ctx context.Context, id uuid.UUID) (*domain.TaxTransaction, error) {
	return s.repo.GetTaxTransactionByID(ctx, id)
}

func (s *taxService) ListTaxTransactions(ctx context.Context, legalEntityID uuid.UUID, limit, offset int32) ([]domain.TaxTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListTaxTransactionsByEntity(ctx, legalEntityID, limit, offset)
}

// 4. Periods & VAT Returns (ZATCA Form)
func (s *taxService) ListPeriods(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxPeriod, error) {
	return s.repo.ListTaxPeriods(ctx, legalEntityID)
}

func (s *taxService) CreatePeriod(ctx context.Context, period domain.TaxPeriod) (*domain.TaxPeriod, error) {
	return s.repo.CreateTaxPeriod(ctx, &period)
}

func (s *taxService) ClosePeriod(ctx context.Context, periodID uuid.UUID) (*domain.TaxPeriod, error) {
	return s.repo.UpdateTaxPeriodStatus(ctx, periodID, domain.TaxPeriodStatusClosed)
}

func (s *taxService) GenerateVATReturnDraft(ctx context.Context, legalEntityID, periodID uuid.UUID) (*domain.TaxReturn, error) {
	period, err := s.repo.GetTaxPeriodByID(ctx, periodID)
	if err != nil {
		return nil, err
	}

	existingReturn, err := s.repo.GetTaxReturnByPeriod(ctx, legalEntityID, periodID)
	if err != nil {
		return nil, err
	}
	if existingReturn != nil && existingReturn.Status == "SUBMITTED" {
		return nil, domain.ErrTaxReturnAlreadySubmitted
	}

	summaryRows, err := s.repo.AggregatePeriodSummary(ctx, periodID)
	if err != nil {
		return nil, err
	}

	boxes := domain.DeclarationBoxes{}

	for _, row := range summaryRows {
		cat := strings.ToUpper(row.TaxCategoryCode)

		if row.TransactionType == "OUTPUT" {
			// Sales
			switch cat {
			case "S":
				boxes.Box1StandardRatedSales.TaxableAmount = boxes.Box1StandardRatedSales.TaxableAmount.Add(row.TotalTaxable)
				boxes.Box1StandardRatedSales.VATAmount = boxes.Box1StandardRatedSales.VATAmount.Add(row.TotalTax)
			case "Z":
				reason := ""
				if row.ZeroRateReasonCode != nil {
					reason = *row.ZeroRateReasonCode
				}
				if reason == "VATEX-SA-EDU" || reason == "VATEX-SA-HEA" {
					boxes.Box2PrivateHealthcareEdu.TaxableAmount = boxes.Box2PrivateHealthcareEdu.TaxableAmount.Add(row.TotalTaxable)
					boxes.Box2PrivateHealthcareEdu.VATAmount = boxes.Box2PrivateHealthcareEdu.VATAmount.Add(row.TotalTax)
				} else if reason == "VATEX-SA-32" || reason == "VATEX-SA-33" {
					boxes.Box4Exports.TaxableAmount = boxes.Box4Exports.TaxableAmount.Add(row.TotalTaxable)
					boxes.Box4Exports.VATAmount = boxes.Box4Exports.VATAmount.Add(row.TotalTax)
				} else {
					boxes.Box3ZeroRatedDomestic.TaxableAmount = boxes.Box3ZeroRatedDomestic.TaxableAmount.Add(row.TotalTaxable)
					boxes.Box3ZeroRatedDomestic.VATAmount = boxes.Box3ZeroRatedDomestic.VATAmount.Add(row.TotalTax)
				}
			case "E":
				boxes.Box5ExemptSales.TaxableAmount = boxes.Box5ExemptSales.TaxableAmount.Add(row.TotalTaxable)
				boxes.Box5ExemptSales.VATAmount = boxes.Box5ExemptSales.VATAmount.Add(row.TotalTax)
			}
		} else if row.TransactionType == "INPUT" {
			// Purchases
			switch cat {
			case "S":
				boxes.Box7StandardRatedPurchases.TaxableAmount = boxes.Box7StandardRatedPurchases.TaxableAmount.Add(row.TotalTaxable)
				boxes.Box7StandardRatedPurchases.VATAmount = boxes.Box7StandardRatedPurchases.VATAmount.Add(row.TotalTax)
			case "Z":
				boxes.Box10ZeroRatedPurchases.TaxableAmount = boxes.Box10ZeroRatedPurchases.TaxableAmount.Add(row.TotalTaxable)
				boxes.Box10ZeroRatedPurchases.VATAmount = boxes.Box10ZeroRatedPurchases.VATAmount.Add(row.TotalTax)
			case "E":
				boxes.Box11ExemptPurchases.TaxableAmount = boxes.Box11ExemptPurchases.TaxableAmount.Add(row.TotalTaxable)
				boxes.Box11ExemptPurchases.VATAmount = boxes.Box11ExemptPurchases.VATAmount.Add(row.TotalTax)
			}
		}
	}

	// Accumulate Sales Total (Box 6)
	boxes.Box6TotalSales.TaxableAmount = boxes.Box1StandardRatedSales.TaxableAmount.
		Add(boxes.Box2PrivateHealthcareEdu.TaxableAmount).
		Add(boxes.Box3ZeroRatedDomestic.TaxableAmount).
		Add(boxes.Box4Exports.TaxableAmount).
		Add(boxes.Box5ExemptSales.TaxableAmount)

	boxes.Box6TotalSales.VATAmount = boxes.Box1StandardRatedSales.VATAmount.
		Add(boxes.Box2PrivateHealthcareEdu.VATAmount).
		Add(boxes.Box3ZeroRatedDomestic.VATAmount).
		Add(boxes.Box4Exports.VATAmount).
		Add(boxes.Box5ExemptSales.VATAmount)

	// Accumulate Purchases Total (Box 12)
	boxes.Box12TotalPurchases.TaxableAmount = boxes.Box7StandardRatedPurchases.TaxableAmount.
		Add(boxes.Box8ImportsSubjectToVAT.TaxableAmount).
		Add(boxes.Box9ImportsReverseCharge.TaxableAmount).
		Add(boxes.Box10ZeroRatedPurchases.TaxableAmount).
		Add(boxes.Box11ExemptPurchases.TaxableAmount)

	boxes.Box12TotalPurchases.VATAmount = boxes.Box7StandardRatedPurchases.VATAmount.
		Add(boxes.Box8ImportsSubjectToVAT.VATAmount).
		Add(boxes.Box9ImportsReverseCharge.VATAmount).
		Add(boxes.Box10ZeroRatedPurchases.VATAmount).
		Add(boxes.Box11ExemptPurchases.VATAmount)

	// Net VAT Due (Box 13 & Box 16)
	boxes.Box13TotalVATDueCurrentPeriod = boxes.Box6TotalSales.VATAmount.Sub(boxes.Box12TotalPurchases.VATAmount)
	boxes.Box16NetVATDue = boxes.Box13TotalVATDueCurrentPeriod

	retNumber := fmt.Sprintf("VAT-RET-%d-P%02d", period.FiscalYear, period.PeriodNumber)

	draft := &domain.TaxReturn{
		LegalEntityID:         legalEntityID,
		TaxPeriodID:           periodID,
		ReturnNumber:          retNumber,
		Status:                "DRAFT",
		TotalSalesTaxable:     boxes.Box6TotalSales.TaxableAmount,
		TotalSalesVAT:         boxes.Box6TotalSales.VATAmount,
		TotalPurchasesTaxable: boxes.Box12TotalPurchases.TaxableAmount,
		TotalPurchasesVAT:     boxes.Box12TotalPurchases.VATAmount,
		NetVATDue:             boxes.Box16NetVATDue,
		DeclarationBoxes:      boxes,
	}

	if existingReturn != nil {
		draft.ID = existingReturn.ID
		return s.repo.UpdateTaxReturn(ctx, draft)
	}

	return s.repo.CreateTaxReturn(ctx, draft)
}

func (s *taxService) SubmitVATReturn(ctx context.Context, legalEntityID, returnID uuid.UUID, actorID uuid.UUID) (*domain.TaxReturn, error) {
	ret, err := s.repo.GetTaxReturnByID(ctx, returnID)
	if err != nil || ret == nil {
		return nil, fmt.Errorf("tax return not found")
	}

	if ret.Status == "SUBMITTED" {
		return nil, domain.ErrTaxReturnAlreadySubmitted
	}

	now := time.Now().UTC()
	ret.Status = "SUBMITTED"
	ret.FilingDate = &now
	ret.SubmittedAt = &now
	ret.SubmittedBy = &actorID

	return s.repo.UpdateTaxReturn(ctx, ret)
}

func (s *taxService) ListVATReturns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.TaxReturn, error) {
	return s.repo.ListTaxReturnsByEntity(ctx, legalEntityID)
}
