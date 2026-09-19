package domain

import "errors"

var (
	ErrTaxCodeNotFound           = errors.New("tax code not found")
	ErrTaxTransactionNotFound    = errors.New("tax transaction not found")
	ErrTaxReturnNotFound         = errors.New("tax return not found")
	ErrTaxRateVersionNotFound    = errors.New("active tax rate version not found for date")
	ErrExemptionReasonRequired   = errors.New("exemption reason code is mandatory for exempt tax category (E)")
	ErrZeroRateReasonRequired    = errors.New("zero rate reason code is mandatory for zero-rated tax category (Z)")
	ErrInvalidTaxRate            = errors.New("invalid tax rate percent")
	ErrTaxPeriodNotFound         = errors.New("no matching tax reporting period found for date")
	ErrTaxPeriodClosed           = errors.New("tax reporting period is closed; posting new tax transactions is prohibited")
	ErrDuplicateTaxTransaction   = errors.New("duplicate tax transaction for this source document")
	ErrTaxReturnAlreadySubmitted = errors.New("tax return for this period has already been submitted")
	ErrInvalidTaxTransaction     = errors.New("invalid tax transaction: missing required lines or negative taxable amount")
)
