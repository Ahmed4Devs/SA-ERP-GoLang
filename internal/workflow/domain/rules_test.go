package domain

import (
	"encoding/json"
	"testing"
)

func TestApprovalThresholdUsesExactDecimal(t *testing.T) {
	for _, value := range []any{"9007199254740992.000001", json.Number("9007199254740992.000001")} {
		met, err := EvaluateConditions([]StepCondition{{FieldName: "amount", Operator: "GT", FieldValue: "9007199254740992.000000"}}, map[string]any{"amount": value})
		if err != nil || !met {
			t.Fatalf("approval threshold rounded away a financial difference: %v %v", met, err)
		}
	}
	if _, err := EvaluateConditions([]StepCondition{{FieldName: "amount", Operator: "EQ", FieldValue: "1"}}, map[string]any{"amount": float64(1)}); err == nil {
		t.Fatal("binary floating point must not decide a financial approval")
	}
}
