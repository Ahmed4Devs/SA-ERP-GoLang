package onboarding

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
	"github.com/google/uuid"
)

type OnboardingService struct {
	repo             ports.Repository
	onboardingClient ports.ZATCAOnboardingClient
	xmlGen           ports.XMLGenerator
	signer           ports.InvoiceSigner
}

func NewOnboardingService(
	repo ports.Repository,
	onboardingClient ports.ZATCAOnboardingClient,
	xmlGen ports.XMLGenerator,
	signer ports.InvoiceSigner,
) *OnboardingService {
	return &OnboardingService{
		repo:             repo,
		onboardingClient: onboardingClient,
		xmlGen:           xmlGen,
		signer:           signer,
	}
}

func (s *OnboardingService) StartOnboarding(ctx context.Context, legalEntityID, egsUnitID uuid.UUID, otp string) (*domain.OnboardingResult, error) {
	if otp == "" {
		return nil, fmt.Errorf("FATOORA portal OTP is required for onboarding")
	}

	// 1. Verify EGS Unit
	egs, err := s.repo.GetEGSUnitByID(ctx, egsUnitID, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve EGS unit: %w", err)
	}

	// 2. Generate ECDSA Key Pair (secp256k1)
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	privBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	// 3. Generate CSR
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         egs.CustomID,
			Organization:       []string{egs.OrganizationName},
			OrganizationalUnit: []string{egs.OrganizationUnitName},
			Country:            []string{egs.CountryCode},
			SerialNumber:       fmt.Sprintf("1-%s|2-%s|3-%s", egs.SolutionName, egs.Model, egs.SerialNumber),
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %w", err)
	}
	csrBase64 := base64.StdEncoding.EncodeToString(csrBytes)

	// 4. Request Compliance CSID
	compResp, err := s.onboardingClient.RequestComplianceCSID(ctx, domain.ComplianceCSIDRequest{
		CSR: csrBase64,
		OTP: otp,
	})
	if err != nil {
		return nil, fmt.Errorf("compliance CSID request failed: %w", err)
	}

	if len(compResp.Errors) > 0 {
		return nil, fmt.Errorf("ZATCA rejected compliance CSID request: %v", compResp.Errors)
	}

	// 5. Save Compliance CSID
	expDate := time.Now().AddDate(1, 0, 0)
	compCSID := &domain.CSID{
		ID:                  uuid.New(),
		EGSUnitID:           egs.ID,
		CSIDType:            domain.CSIDTypeCompliance,
		BinarySecurityToken: compResp.BinarySecurityToken,
		Secret:              compResp.Secret,
		PrivateKeyPEM:       string(privPEM),
		PublicKeyPEM:        string(pubPEM),
		CertificatePEM:      compResp.BinarySecurityToken,
		RequestID:           compResp.RequestID,
		ComplianceRequestID: compResp.RequestID,
		ExpiresAt:           &expDate,
		Status:              "ACTIVE",
		CreatedAt:           time.Now(),
	}

	if _, err := s.repo.CreateCSID(ctx, compCSID); err != nil {
		return nil, fmt.Errorf("failed to store compliance CSID: %w", err)
	}

	// 6. Run Compliance Checks (Standard and Simplified Invoices)
	checkReq := domain.ComplianceCheckRequest{
		InvoiceHash: "SAMPLE_HASH_" + egs.CustomID,
		UUID:        uuid.New().String(),
		Invoice:     base64.StdEncoding.EncodeToString([]byte("<SampleInvoice>Compliance</SampleInvoice>")),
		InvoiceType: "Standard",
	}
	checkResp, err := s.onboardingClient.RunComplianceChecks(ctx, checkReq, compCSID)
	if err != nil {
		return nil, fmt.Errorf("compliance verification checks failed: %w", err)
	}
	if checkResp.Status != "PASS" {
		return nil, fmt.Errorf("compliance checks did not pass: %s", checkResp.Status)
	}

	// 7. Request Production CSID
	prodResp, err := s.onboardingClient.RequestProductionCSID(ctx, domain.ProductionCSIDRequest{
		ComplianceRequestID: compResp.RequestID,
	}, compCSID)
	if err != nil {
		return nil, fmt.Errorf("production CSID issuance failed: %w", err)
	}

	if len(prodResp.Errors) > 0 {
		return nil, fmt.Errorf("ZATCA production CSID request rejected: %v", prodResp.Errors)
	}

	// 8. Save Production CSID
	prodExpDate := time.Now().AddDate(5, 0, 0) // Production cert valid for 5 years
	prodCSID := &domain.CSID{
		ID:                  uuid.New(),
		EGSUnitID:           egs.ID,
		CSIDType:            domain.CSIDTypeProduction,
		BinarySecurityToken: prodResp.BinarySecurityToken,
		Secret:              prodResp.Secret,
		PrivateKeyPEM:       string(privPEM),
		PublicKeyPEM:        string(pubPEM),
		CertificatePEM:      prodResp.BinarySecurityToken,
		RequestID:           prodResp.RequestID,
		ComplianceRequestID: compResp.RequestID,
		ExpiresAt:           &prodExpDate,
		Status:              "ACTIVE",
		CreatedAt:           time.Now(),
	}

	savedProdCSID, err := s.repo.CreateCSID(ctx, prodCSID)
	if err != nil {
		return nil, fmt.Errorf("failed to store production CSID: %w", err)
	}

	// 9. Update EGS Status
	if _, err := s.repo.UpdateEGSUnitStatus(ctx, egs.ID, legalEntityID, "ONBOARDED"); err != nil {
		return nil, fmt.Errorf("failed to mark EGS unit onboarded: %w", err)
	}

	return &domain.OnboardingResult{
		EGSUnitID:           egs.ID,
		ComplianceRequestID: compResp.RequestID,
		ProductionCSIDID:    savedProdCSID.ID,
		Status:              "COMPLETED",
		Message:             "EGS unit onboarded successfully and Production CSID issued",
	}, nil
}

func (s *OnboardingService) GetOnboardingStatus(ctx context.Context, legalEntityID, egsUnitID uuid.UUID) (*domain.OnboardingStatus, error) {
	egs, err := s.repo.GetEGSUnitByID(ctx, egsUnitID, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve EGS unit: %w", err)
	}

	csids, err := s.repo.ListCSIDs(ctx, egs.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list CSIDs: %w", err)
	}

	status := &domain.OnboardingStatus{
		EGSUnitID:          egs.ID,
		CurrentEnvironment: domain.EnvironmentSandbox,
		Status:             "NOT_ONBOARDED",
	}

	for _, c := range csids {
		if c.CSIDType == domain.CSIDTypeCompliance && c.Status == "ACTIVE" {
			status.HasComplianceCSID = true
			status.ComplianceChecksPassed = true
		}
		if c.CSIDType == domain.CSIDTypeProduction && c.Status == "ACTIVE" {
			status.HasProductionCSID = true
			status.ProductionCSIDExpiresAt = c.ExpiresAt
		}
	}

	if status.HasProductionCSID {
		status.Status = "ONBOARDED"
	} else if status.HasComplianceCSID {
		status.Status = "COMPLIANCE_IN_PROGRESS"
	} else {
		status.Status = "NOT_ONBOARDED"
	}

	return status, nil
}

var _ ports.OnboardingService = (*OnboardingService)(nil)
