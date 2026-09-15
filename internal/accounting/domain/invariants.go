package domain

import (
	"github.com/shopspring/decimal"
)

// ValidateJournalEntryBalance verifies double-entry bookkeeping invariants:
// 1. Must contain at least two lines.
// 2. Each line must have debit XOR credit > 0.
// 3. Sum of all base debits must exactly equal sum of all base credits.
func ValidateJournalEntryBalance(lines []JournalLine) (totalDebit decimal.Decimal, totalCredit decimal.Decimal, err error) {
	if len(lines) < 2 {
		return decimal.Zero, decimal.Zero, ErrNoLinesProvided
	}

	seenLineNumbers := make(map[int16]bool)
	sumDebit := decimal.Zero
	sumCredit := decimal.Zero

	for _, line := range lines {
		if line.LineNumber <= 0 || seenLineNumbers[line.LineNumber] {
			return decimal.Zero, decimal.Zero, ErrDuplicateLineNumber
		}
		seenLineNumbers[line.LineNumber] = true

		hasDebit := line.BaseDebit.GreaterThan(decimal.Zero)
		hasCredit := line.BaseCredit.GreaterThan(decimal.Zero)

		// Debit XOR Credit rule
		if (hasDebit && hasCredit) || (!hasDebit && !hasCredit) {
			return decimal.Zero, decimal.Zero, ErrInvalidLineAmounts
		}

		if line.BaseDebit.IsNegative() || line.BaseCredit.IsNegative() {
			return decimal.Zero, decimal.Zero, ErrInvalidLineAmounts
		}

		sumDebit = sumDebit.Add(line.BaseDebit)
		sumCredit = sumCredit.Add(line.BaseCredit)
	}

	if !sumDebit.Equal(sumCredit) {
		return sumDebit, sumCredit, ErrUnbalancedJournalEntry
	}

	return sumDebit, sumCredit, nil
}

// ValidatePeriodCanPost ensures that financial movements can only be posted into OPEN fiscal periods.
func ValidatePeriodCanPost(status FiscalPeriodStatus) error {
	if status != FiscalPeriodStatusOpen {
		return ErrPeriodNotOpen
	}
	return nil
}

// ValidateAccountCanPost verifies that an account is active and designated for posting (not a parent category).
func ValidateAccountCanPost(allowPosting bool, isActive bool) error {
	if !isActive {
		return ErrAccountInactive
	}
	if !allowPosting {
		return ErrAccountNotPostable
	}
	return nil
}
