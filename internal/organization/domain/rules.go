package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	codeRegex   = regexp.MustCompile(`^[A-Za-z0-9_-]{2,32}$`)
	crRegex     = regexp.MustCompile(`^[0-9]{10}$`)
	vatRegex    = regexp.MustCompile(`^3[0-9]{13}3$`)
	bldgRegex   = regexp.MustCompile(`^[0-9]{4}$`)
	postalRegex = regexp.MustCompile(`^[0-9]{5}$`)
)

func ValidateCode(code string) error {
	trimmed := strings.TrimSpace(code)
	if !codeRegex.MatchString(trimmed) {
		return ErrInvalidCode
	}
	return nil
}

func ValidateCRNumber(cr string) error {
	if cr == "" {
		return nil
	}
	if !crRegex.MatchString(strings.TrimSpace(cr)) {
		return ErrInvalidCRNumber
	}
	return nil
}

func ValidateVATNumber(vat string) error {
	if vat == "" {
		return nil
	}
	if !vatRegex.MatchString(strings.TrimSpace(vat)) {
		return ErrInvalidVATNumber
	}
	return nil
}

func ValidateNationalAddress(bldg, postal, addl string) error {
	if !bldgRegex.MatchString(strings.TrimSpace(bldg)) {
		return fmt.Errorf("%w: building number must be 4 digits", ErrInvalidNationalAddress)
	}
	if !postalRegex.MatchString(strings.TrimSpace(postal)) {
		return fmt.Errorf("%w: postal code must be 5 digits", ErrInvalidNationalAddress)
	}
	if addl != "" && !bldgRegex.MatchString(strings.TrimSpace(addl)) {
		return fmt.Errorf("%w: additional number must be 4 digits", ErrInvalidNationalAddress)
	}
	return nil
}

// FormatSequence generates the formatted document number according to sequence definition.
func (s *NumberSequence) FormatValue(val int64, now time.Time) string {
	padFormat := fmt.Sprintf("%%0%dd", s.PaddingLength)
	paddedNum := fmt.Sprintf(padFormat, val)

	var sb strings.Builder
	if s.Prefix != "" {
		sb.WriteString(s.Prefix)
		if !strings.HasSuffix(s.Prefix, "-") && !strings.HasSuffix(s.Prefix, "/") {
			sb.WriteString("-")
		}
	}

	if s.ResetInterval == ResetYearly {
		sb.WriteString(fmt.Sprintf("%d-", now.Year()))
	} else if s.ResetInterval == ResetMonthly {
		sb.WriteString(fmt.Sprintf("%d%02d-", now.Year(), int(now.Month())))
	}

	sb.WriteString(paddedNum)

	if s.Suffix != "" {
		if !strings.HasPrefix(s.Suffix, "-") && !strings.HasPrefix(s.Suffix, "/") {
			sb.WriteString("-")
		}
		sb.WriteString(s.Suffix)
	}

	return sb.String()
}
