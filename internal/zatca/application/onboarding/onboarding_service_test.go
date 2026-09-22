package onboarding

import (
	"context"
	"strings"
	"testing"

	"erp/internal/zatca/domain"

	"github.com/google/uuid"
)

func TestLiveOnboardingRejectsSimulatorComplianceSample(t *testing.T) {
	service := NewOnboardingService(nil, nil, nil, nil, domain.EnvironmentSandbox)
	_, err := service.StartOnboarding(context.Background(), uuid.New(), uuid.New(), "123456")
	if err == nil || !strings.Contains(err.Error(), "official signed compliance sample suite") {
		t.Fatalf("expected live onboarding to fail closed, got %v", err)
	}
}
