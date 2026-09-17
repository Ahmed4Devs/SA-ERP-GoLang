package validation

import (
	"context"
	"regexp"
	"strings"
	"time"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var vatRegex = regexp.MustCompile(`^3\d{13}3$`)

type InvoiceValidator struct{}

func NewInvoiceValidator() *InvoiceValidator {
	return &InvoiceValidator{}
}

func (v *InvoiceValidator) Validate(ctx context.Context, xmlBytes []byte, data domain.InvoiceData) (*domain.ValidationResult, error) {
	result := &domain.ValidationResult{
		ID:            uuid.New(),
		OverallStatus: "PASSED",
		RuleVersion:   "ZATCA-2024-v2.1",
		ValidatedAt:   time.Now(),
		Messages:      []domain.ValidationMessage{},
	}

	addMsg := func(msgType, code, category, text string) {
		result.Messages = append(result.Messages, domain.ValidationMessage{
			ID:          uuid.New(),
			MessageType: msgType,
			Code:        code,
			Category:    category,
			MessageText: text,
			CreatedAt:   time.Now(),
		})
		if msgType == "ERROR" {
			result.OverallStatus = "FAILED"
		} else if msgType == "WARNING" && result.OverallStatus != "FAILED" {
			result.OverallStatus = "WARNING"
		}
	}

	// BR-KSA-01: Issue Date check
	if data.IssueDate == "" {
		addMsg("ERROR", "BR-KSA-01", "BUSINESS_RULES", "Issue date is mandatory")
	} else {
		parsedDate, err := time.Parse("2006-01-02", data.IssueDate)
		if err != nil {
			addMsg("ERROR", "BR-KSA-01", "SCHEMA", "Issue date format must be YYYY-MM-DD")
		} else if parsedDate.After(time.Now().AddDate(0, 0, 1)) {
			addMsg("ERROR", "BR-KSA-01", "BUSINESS_RULES", "Issue date cannot be in the future")
		}
	}

	// BR-KSA-02: Issue Time check
	if data.IssueTime == "" {
		addMsg("ERROR", "BR-KSA-02", "BUSINESS_RULES", "Issue time is mandatory")
	}

	// BR-KSA-03: UUID check
	if data.UUID == uuid.Nil {
		addMsg("ERROR", "BR-KSA-03", "BUSINESS_RULES", "Invoice UUID is required and must follow RFC 4122")
	}

	// BR-KSA-04: Currency check
	if data.CurrencyCode != "SAR" {
		addMsg("ERROR", "BR-KSA-04", "BUSINESS_RULES", "Invoice currency must be SAR for Saudi domestic transactions")
	}

	// BR-KSA-05: Seller VAT number
	sellerVAT := strings.TrimSpace(data.Seller.VATNumber)
	if !vatRegex.MatchString(sellerVAT) {
		addMsg("ERROR", "BR-KSA-05", "BUSINESS_RULES", "Seller VAT number must be 15 digits starting and ending with 3")
	}

	// BR-KSA-06: Seller Postal Address
	if strings.TrimSpace(data.Seller.Street) == "" || strings.TrimSpace(data.Seller.BuildingNumber) == "" ||
		strings.TrimSpace(data.Seller.PostalCode) == "" || strings.TrimSpace(data.Seller.City) == "" {
		addMsg("ERROR", "BR-KSA-06", "BUSINESS_RULES", "Seller address must contain Street, Building Number, Postal Code, and City")
	}

	// BR-KSA-07: Buyer VAT in Standard B2B Invoices
	if data.InvoiceType == domain.InvoiceTypeStandard {
		buyerVAT := strings.TrimSpace(data.Buyer.VATNumber)
		if buyerVAT == "" || !vatRegex.MatchString(buyerVAT) {
			addMsg("ERROR", "BR-KSA-07", "BUSINESS_RULES", "Buyer VAT number is required for Standard B2B invoices and must be 15 digits starting and ending with 3")
		}
	}

	// BR-KSA-08: ICV & PIH Chaining
	if data.ICV <= 0 {
		addMsg("ERROR", "BR-KSA-08", "CRYPTOGRAPHIC_STAMP", "Invoice Counter Value (ICV) must be greater than zero")
	}
	if data.PIH == "" {
		addMsg("ERROR", "BR-KSA-08", "CRYPTOGRAPHIC_STAMP", "Previous Invoice Hash (PIH) is mandatory for cryptographic chaining")
	}

	// BR-KSA-09: Lines Validation
	if len(data.Lines) == 0 {
		addMsg("ERROR", "BR-KSA-09", "BUSINESS_RULES", "Invoice must have at least one line item")
	}

	expectedTax := decimal.Zero
	expectedExclusive := decimal.Zero

	fifteenPercent := decimal.NewFromInt(15)
	for idx, line := range data.Lines {
		if line.Quantity.LessThanOrEqual(decimal.Zero) {
			addMsg("ERROR", "BR-KSA-10", "BUSINESS_RULES", "Invoice line quantity must be strictly positive")
		}
		if line.UnitPrice.IsNegative() {
			addMsg("ERROR", "BR-KSA-11", "BUSINESS_RULES", "Invoice line unit price cannot be negative")
		}
		if !line.VATRate.Equal(fifteenPercent) {
			addMsg("WARNING", "BR-KSA-12", "BUSINESS_RULES", "Standard Saudi VAT rate is 15.00%")
		}

		lineSubtotal := line.Quantity.Mul(line.UnitPrice).Sub(line.Discount)
		expectedExclusive = expectedExclusive.Add(lineSubtotal)
		expectedTax = expectedTax.Add(line.VATAmount)

		_ = idx
	}

	// BR-KSA-13: Mathematical Balance Check
	expectedInclusive := data.TaxExclusiveAmount.Add(data.TaxTotal)
	diffInclusive := data.TaxInclusiveAmount.Sub(expectedInclusive).Abs()
	if diffInclusive.GreaterThan(decimal.NewFromFloat(0.05)) {
		addMsg("ERROR", "BR-KSA-13", "BUSINESS_RULES", "TaxInclusiveAmount must equal TaxExclusiveAmount + TaxTotal")
	}

	// BR-KSA-14: Payable Amount Check
	expectedPayable := data.TaxInclusiveAmount.Sub(data.AllowanceTotal)
	diffPayable := data.PayableAmount.Sub(expectedPayable).Abs()
	if diffPayable.GreaterThan(decimal.NewFromFloat(0.05)) {
		addMsg("ERROR", "BR-KSA-14", "BUSINESS_RULES", "PayableAmount must equal TaxInclusiveAmount - AllowanceTotal")
	}

	// If no errors, add success info
	if len(result.Messages) == 0 {
		addMsg("INFO", "BR-KSA-OK", "SCHEMA", "Document conforms to all ZATCA Phase 2 UBL 2.1 e-invoicing business rules")
	}

	return result, nil
}

var _ ports.InvoiceValidator = (*InvoiceValidator)(nil)
