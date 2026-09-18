package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
	"github.com/google/uuid"
)

// SimulatorClient provides deterministic, hermetic ZATCA API responses for testing and development.
type SimulatorClient struct {
	mu                  sync.RWMutex
	shouldFailClearance bool
	shouldFailReporting bool
	shouldFailNetwork   bool
	warningMessages     []domain.ValidationMessage
	errorMessages       []domain.ValidationMessage
}

func NewSimulatorClient() *SimulatorClient {
	return &SimulatorClient{}
}

func (s *SimulatorClient) SetFailureMode(network, clearance, reporting bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shouldFailNetwork = network
	s.shouldFailClearance = clearance
	s.shouldFailReporting = reporting
}

func (s *SimulatorClient) SetCustomMessages(warnings, errors []domain.ValidationMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.warningMessages = warnings
	s.errorMessages = errors
}

// ----------------------------------------------------------------------------
// ports.ZATCAClient Implementation
// ----------------------------------------------------------------------------

func (s *SimulatorClient) Clearance(ctx context.Context, req domain.ClearanceRequest, csid *domain.CSID) (domain.ClearanceResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shouldFailNetwork {
		return domain.ClearanceResponse{}, fmt.Errorf("simulated network connection timeout to ZATCA gateway")
	}

	requestID := uuid.New().String()

	if s.shouldFailClearance {
		return domain.ClearanceResponse{}, fmt.Errorf("ZATCA clearance rejected: invoice structure invalid")
	}

	// Produce cleared invoice XML with simulated ZATCA cryptographic stamp
	clearedXML := fmt.Sprintf(`<!-- ZATCA Phase 2 Clearance Stamp: %s -->%s`, requestID, req.Invoice)
	clearedBase64 := base64.StdEncoding.EncodeToString([]byte(clearedXML))

	return domain.ClearanceResponse{
		ClearanceStatus: "CLEARED",
		ClearedInvoice:  clearedBase64,
		ValidationResults: &domain.ValidationResult{
			OverallStatus: "PASSED",
			RuleVersion:   "BR-KSA-2.1",
			ValidatedAt:   time.Now(),
		},
		WarningMessages: s.warningMessages,
		ZATCARequestID:  requestID,
	}, nil
}

func (s *SimulatorClient) Report(ctx context.Context, req domain.ReportingRequest, csid *domain.CSID) (domain.ReportingResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shouldFailNetwork {
		return domain.ReportingResponse{}, fmt.Errorf("simulated network connection timeout to ZATCA gateway")
	}

	requestID := uuid.New().String()

	if s.shouldFailReporting {
		return domain.ReportingResponse{
			ReportingStatus: "NOT_REPORTED",
			ErrorMessages: []domain.ValidationMessage{
				{
					Code:        "BR-KSA-SIM-ERR",
					MessageType: "ERROR",
					Category:    "SIMULATION",
					MessageText: "Reporting rejected due to simulated compliance constraint",
				},
			},
			ZATCARequestID: requestID,
		}, nil
	}

	return domain.ReportingResponse{
		ReportingStatus: "REPORTED",
		ValidationResults: &domain.ValidationResult{
			OverallStatus: "PASSED",
			RuleVersion:   "BR-KSA-2.1",
			ValidatedAt:   time.Now(),
		},
		WarningMessages: s.warningMessages,
		ZATCARequestID:  requestID,
	}, nil
}

// ----------------------------------------------------------------------------
// ports.ZATCAOnboardingClient Implementation
// ----------------------------------------------------------------------------

func (s *SimulatorClient) RequestComplianceCSID(ctx context.Context, req domain.ComplianceCSIDRequest) (domain.ComplianceCSIDResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shouldFailNetwork {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("simulated network failure connecting to ZATCA onboarding service")
	}

	if len(req.OTP) < 4 {
		return domain.ComplianceCSIDResponse{
			DispositionMessage: "INVALID_OTP",
			Errors:             []string{"Provided OTP is invalid or expired"},
		}, nil
	}

	tokenBytes := make([]byte, 64)
	_, _ = rand.Read(tokenBytes)
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	secretBytes := make([]byte, 32)
	_, _ = rand.Read(secretBytes)
	secret := base64.StdEncoding.EncodeToString(secretBytes)

	requestID := fmt.Sprintf("REQ-COMP-%d", time.Now().Unix())

	return domain.ComplianceCSIDResponse{
		RequestID:           requestID,
		DispositionMessage:  "ISSUED",
		BinarySecurityToken: token,
		Secret:              secret,
	}, nil
}

func (s *SimulatorClient) RunComplianceChecks(ctx context.Context, req domain.ComplianceCheckRequest, csid *domain.CSID) (domain.ComplianceCheckResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shouldFailNetwork {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("simulated network timeout during compliance check")
	}

	return domain.ComplianceCheckResponse{
		Status: "PASS",
		ValidationResults: &domain.ValidationResult{
			OverallStatus: "PASSED",
			RuleVersion:   "BR-KSA-2.1",
			ValidatedAt:   time.Now(),
		},
		WarningMessages: s.warningMessages,
	}, nil
}

func (s *SimulatorClient) RequestProductionCSID(ctx context.Context, req domain.ProductionCSIDRequest, csid *domain.CSID) (domain.ProductionCSIDResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shouldFailNetwork {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("simulated network failure during production CSID request")
	}

	tokenBytes := make([]byte, 64)
	_, _ = rand.Read(tokenBytes)
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	secretBytes := make([]byte, 32)
	_, _ = rand.Read(secretBytes)
	secret := base64.StdEncoding.EncodeToString(secretBytes)

	requestID := fmt.Sprintf("REQ-PROD-%d", time.Now().Unix())

	return domain.ProductionCSIDResponse{
		RequestID:           requestID,
		DispositionMessage:  "ISSUED",
		BinarySecurityToken: token,
		Secret:              secret,
	}, nil
}

func (s *SimulatorClient) RenewProductionCSID(ctx context.Context, req domain.RenewalRequest, currentCSID *domain.CSID) (domain.RenewalResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tokenBytes := make([]byte, 64)
	_, _ = rand.Read(tokenBytes)
	token := base64.StdEncoding.EncodeToString(tokenBytes)

	secretBytes := make([]byte, 32)
	_, _ = rand.Read(secretBytes)
	secret := base64.StdEncoding.EncodeToString(secretBytes)

	return domain.RenewalResponse{
		RequestID:           fmt.Sprintf("REQ-RENEW-%d", time.Now().Unix()),
		DispositionMessage:  "RENEWED",
		BinarySecurityToken: token,
		Secret:              secret,
	}, nil
}

// Ensure interface adherence
var _ ports.ZATCAClient = (*SimulatorClient)(nil)
var _ ports.ZATCAOnboardingClient = (*SimulatorClient)(nil)
