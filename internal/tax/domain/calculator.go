package domain

import (
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TaxLineCalculationInput struct {
	LineNumber          int             `json:"line_number"`
	TaxCodeID           uuid.UUID       `json:"tax_code_id"`
	TaxCode             string          `json:"tax_code"`
	TaxCategoryCode     string          `json:"tax_category_code"`
	TaxRatePercent      decimal.Decimal `json:"tax_rate_percent"`
	Quantity            decimal.Decimal `json:"quantity"`
	UnitPrice           decimal.Decimal `json:"unit_price"`
	DiscountAmount      decimal.Decimal `json:"discount_amount"`
	ExemptionReasonCode *string         `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string         `json:"zero_rate_reason_code,omitempty"`
	IsRecoverable       bool            `json:"is_recoverable"`
}

type TaxLineCalculationResult struct {
	LineNumber          int             `json:"line_number"`
	TaxCodeID           uuid.UUID       `json:"tax_code_id"`
	TaxCode             string          `json:"tax_code"`
	TaxCategoryCode     string          `json:"tax_category_code"`
	TaxRatePercent      decimal.Decimal `json:"tax_rate_percent"`
	TaxableAmount       decimal.Decimal `json:"taxable_amount"`
	TaxAmount           decimal.Decimal `json:"tax_amount"`
	GrossAmount         decimal.Decimal `json:"gross_amount"`
	ExemptionReasonCode *string         `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string         `json:"zero_rate_reason_code,omitempty"`
	IsRecoverable       bool            `json:"is_recoverable"`
}

type TaxCalculationResult struct {
	TotalTaxableAmount decimal.Decimal            `json:"total_taxable_amount"`
	TotalTaxAmount     decimal.Decimal            `json:"total_tax_amount"`
	TotalGrossAmount   decimal.Decimal            `json:"total_gross_amount"`
	Lines              []TaxLineCalculationResult `json:"lines"`
}

// CalculateVATLine validates ZATCA regulatory invariants and computes exact tax values.
func CalculateVATLine(input TaxLineCalculationInput) (*TaxLineCalculationResult, error) {
	catCode := strings.ToUpper(strings.TrimSpace(input.TaxCategoryCode))

	// 1. Validate Category and Invariants
	switch catCode {
	case "E": // Exempt
		if !input.TaxRatePercent.IsZero() {
			return nil, ErrInvalidTaxRate
		}
		if input.ExemptionReasonCode == nil || strings.TrimSpace(*input.ExemptionReasonCode) == "" {
			return nil, ErrExemptionReasonRequired
		}
	case "Z": // Zero Rated
		if !input.TaxRatePercent.IsZero() {
			return nil, ErrInvalidTaxRate
		}
		if input.ZeroRateReasonCode == nil || strings.TrimSpace(*input.ZeroRateReasonCode) == "" {
			return nil, ErrZeroRateReasonRequired
		}
	case "O": // Out of Scope
		if !input.TaxRatePercent.IsZero() {
			return nil, ErrInvalidTaxRate
		}
	case "S": // Standard Rate
		if input.TaxRatePercent.IsNegative() || input.TaxRatePercent.IsZero() {
			return nil, ErrInvalidTaxRate
		}
	default:
		// Unknown category defaults to standard check
		if input.TaxRatePercent.IsNegative() {
			return nil, ErrInvalidTaxRate
		}
	}

	// 2. Exact Decimal Arithmetic
	// Line Gross before discount = Quantity * UnitPrice
	subtotal := input.Quantity.Mul(input.UnitPrice)
	taxable := subtotal.Sub(input.DiscountAmount)
	if taxable.IsNegative() {
		return nil, ErrInvalidTaxTransaction
	}

	// Tax amount = Taxable * (Rate / 100), rounded to 2 decimal places (SAR Halalas)
	hundred := decimal.NewFromInt(100)
	taxAmount := taxable.Mul(input.TaxRatePercent).Div(hundred).Round(2)
	gross := taxable.Add(taxAmount)

	return &TaxLineCalculationResult{
		LineNumber:          input.LineNumber,
		TaxCodeID:           input.TaxCodeID,
		TaxCode:             input.TaxCode,
		TaxCategoryCode:     catCode,
		TaxRatePercent:      input.TaxRatePercent,
		TaxableAmount:       taxable.Round(2),
		TaxAmount:           taxAmount,
		GrossAmount:         gross.Round(2),
		ExemptionReasonCode: input.ExemptionReasonCode,
		ZeroRateReasonCode:  input.ZeroRateReasonCode,
		IsRecoverable:       input.IsRecoverable,
	}, nil
}

// CalculateVATDocument computes line items and accumulates summary totals.
func CalculateVATDocument(lines []TaxLineCalculationInput) (*TaxCalculationResult, error) {
	if len(lines) == 0 {
		return nil, ErrInvalidTaxTransaction
	}

	res := &TaxCalculationResult{
		TotalTaxableAmount: decimal.Zero,
		TotalTaxAmount:     decimal.Zero,
		TotalGrossAmount:   decimal.Zero,
		Lines:              make([]TaxLineCalculationResult, 0, len(lines)),
	}

	for i, lineInput := range lines {
		if lineInput.LineNumber == 0 {
			lineInput.LineNumber = i + 1
		}
		lineRes, err := CalculateVATLine(lineInput)
		if err != nil {
			return nil, err
		}

		res.Lines = append(res.Lines, *lineRes)
		res.TotalTaxableAmount = res.TotalTaxableAmount.Add(lineRes.TaxableAmount)
		res.TotalTaxAmount = res.TotalTaxAmount.Add(lineRes.TaxAmount)
		res.TotalGrossAmount = res.TotalGrossAmount.Add(lineRes.GrossAmount)
	}

	res.TotalTaxableAmount = res.TotalTaxableAmount.Round(2)
	res.TotalTaxAmount = res.TotalTaxAmount.Round(2)
	res.TotalGrossAmount = res.TotalGrossAmount.Round(2)

	return res, nil
}
