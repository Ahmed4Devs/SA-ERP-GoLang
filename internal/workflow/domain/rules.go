package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ValidateMakerChecker enforces Section 55: cannot_approve_own_document.
// The document initiator cannot approve or decide their own workflow step.
func ValidateMakerChecker(initiatorID, actorID uuid.UUID) error {
	if initiatorID == actorID {
		return ErrMakerCannotApprove
	}
	return nil
}

// EvaluateConditions checks whether a step's activation conditions are met by context data.
func EvaluateConditions(conditions []StepCondition, contextData map[string]any) (bool, error) {
	if len(conditions) == 0 {
		return true, nil
	}

	for _, cond := range conditions {
		val, exists := contextData[cond.FieldName]
		if !exists {
			// Field missing in context -> condition not met
			return false, nil
		}

		met, err := compareValues(val, cond.Operator, cond.FieldValue)
		if err != nil {
			return false, fmt.Errorf("failed evaluating condition on field '%s': %w", cond.FieldName, err)
		}
		if !met {
			return false, nil
		}
	}

	return true, nil
}

func compareValues(actual any, operator string, targetStr string) (bool, error) {
	switch actual.(type) {
	case float32, float64:
		return false, fmt.Errorf("workflow numbers must be exact decimal strings or JSON numbers")
	}
	// 1. Try numeric comparison if applicable
	actualNum, actualIsNum := toDecimal(actual)
	targetNum, targetErr := decimal.NewFromString(targetStr)

	if actualIsNum && targetErr == nil {
		switch operator {
		case "GT":
			return actualNum.GreaterThan(targetNum), nil
		case "GTE":
			return actualNum.GreaterThanOrEqual(targetNum), nil
		case "LT":
			return actualNum.LessThan(targetNum), nil
		case "LTE":
			return actualNum.LessThanOrEqual(targetNum), nil
		case "EQ":
			return actualNum.Equal(targetNum), nil
		case "NEQ":
			return !actualNum.Equal(targetNum), nil
		default:
			return false, fmt.Errorf("unsupported operator: %s", operator)
		}
	}

	// 2. String comparison
	actualStr := fmt.Sprintf("%v", actual)
	switch operator {
	case "EQ":
		return actualStr == targetStr, nil
	case "NEQ":
		return actualStr != targetStr, nil
	default:
		return false, fmt.Errorf("operator %s not supported for non-numeric values", operator)
	}
}

func toDecimal(v any) (decimal.Decimal, bool) {
	switch n := v.(type) {
	case decimal.Decimal:
		return n, true
	case json.Number:
		d, err := decimal.NewFromString(n.String())
		return d, err == nil
	case int:
		return decimal.NewFromInt(int64(n)), true
	case int64:
		return decimal.NewFromInt(n), true
	case string:
		d, err := decimal.NewFromString(n)
		return d, err == nil
	default:
		return decimal.Zero, false
	}
}

// IsDelegationActive checks whether a delegation is valid at the given time and scope.
func IsDelegationActive(d *Delegation, legalEntityID uuid.UUID, at time.Time) bool {
	if !d.IsActive {
		return false
	}
	if at.Before(d.StartDate) || at.After(d.EndDate) {
		return false
	}
	if d.LegalEntityID != nil && *d.LegalEntityID != legalEntityID {
		return false
	}
	return true
}
