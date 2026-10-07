package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"erp/internal/zatca/domain"
	"erp/internal/zatca/ports"
)

type Config struct {
	Environment domain.Environment
	BaseURL     string
	Timeout     time.Duration
}

func DefaultConfig(env domain.Environment) Config {
	var baseURL string
	switch env {
	case domain.EnvironmentProduction:
		baseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
	case domain.EnvironmentSimulation:
		baseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/simulation"
	default:
		baseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/developer-portal"
	}

	return Config{
		Environment: env,
		BaseURL:     baseURL,
		Timeout:     30 * time.Second,
	}
}

type HTTPClient struct {
	cfg        Config
	httpClient *http.Client
}

func NewHTTPClient(cfg Config) *HTTPClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &HTTPClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// ----------------------------------------------------------------------------
// ports.ZATCAClient Implementation (Clearance & Reporting)
// ----------------------------------------------------------------------------

func (c *HTTPClient) Clearance(ctx context.Context, req domain.ClearanceRequest, csid *domain.CSID) (domain.ClearanceResponse, error) {
	url := fmt.Sprintf("%s/invoices/clearance/single", c.cfg.BaseURL)
	payload, err := json.Marshal(req)
	if err != nil {
		return domain.ClearanceResponse{}, fmt.Errorf("failed to marshal clearance request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ClearanceResponse{}, fmt.Errorf("failed to create clearance request: %w", err)
	}

	c.setHeaders(httpReq, csid)
	httpReq.Header.Set("Clearance-Status", "1")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ClearanceResponse{}, fmt.Errorf("clearance request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ClearanceResponse{}, fmt.Errorf("ZATCA HTTP request rejected with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ClearanceResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var clearanceResp domain.ClearanceResponse
	if err := json.Unmarshal(body, &clearanceResp); err != nil {
		return domain.ClearanceResponse{}, fmt.Errorf("failed to decode clearance response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		if clearanceResp.ClearanceStatus == "" {
			clearanceResp.ClearanceStatus = "NOT_CLEARED"
		}
	}

	return clearanceResp, nil
}

func (c *HTTPClient) Report(ctx context.Context, req domain.ReportingRequest, csid *domain.CSID) (domain.ReportingResponse, error) {
	url := fmt.Sprintf("%s/invoices/reporting/single", c.cfg.BaseURL)
	payload, err := json.Marshal(req)
	if err != nil {
		return domain.ReportingResponse{}, fmt.Errorf("failed to marshal reporting request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ReportingResponse{}, fmt.Errorf("failed to create reporting request: %w", err)
	}

	c.setHeaders(httpReq, csid)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ReportingResponse{}, fmt.Errorf("reporting request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ReportingResponse{}, fmt.Errorf("ZATCA HTTP request rejected with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ReportingResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var reportResp domain.ReportingResponse
	if err := json.Unmarshal(body, &reportResp); err != nil {
		return domain.ReportingResponse{}, fmt.Errorf("failed to decode reporting response: %w", err)
	}

	return reportResp, nil
}

// ----------------------------------------------------------------------------
// ports.ZATCAOnboardingClient Implementation
// ----------------------------------------------------------------------------

func (c *HTTPClient) RequestComplianceCSID(ctx context.Context, req domain.ComplianceCSIDRequest) (domain.ComplianceCSIDResponse, error) {
	url := fmt.Sprintf("%s/compliance", c.cfg.BaseURL)
	payload, err := json.Marshal(map[string]string{"csr": req.CSR})
	if err != nil {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("failed to marshal CSR request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("failed to create compliance CSID request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept-Version", "V2")
	httpReq.Header.Set("OTP", req.OTP)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("compliance CSID request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("ZATCA HTTP request rejected with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("failed to read compliance CSID response: %w", err)
	}

	var res domain.ComplianceCSIDResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return domain.ComplianceCSIDResponse{}, fmt.Errorf("failed to decode compliance CSID response: %w", err)
	}

	return res, nil
}

func (c *HTTPClient) RunComplianceChecks(ctx context.Context, req domain.ComplianceCheckRequest, csid *domain.CSID) (domain.ComplianceCheckResponse, error) {
	url := fmt.Sprintf("%s/compliance/invoices", c.cfg.BaseURL)
	payload, err := json.Marshal(req)
	if err != nil {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("failed to marshal compliance check request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("failed to create compliance check request: %w", err)
	}

	c.setHeaders(httpReq, csid)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("compliance check request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("ZATCA HTTP request rejected with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("failed to read compliance check response: %w", err)
	}

	var checkResp domain.ComplianceCheckResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return domain.ComplianceCheckResponse{}, fmt.Errorf("failed to decode compliance check response: %w", err)
	}

	return checkResp, nil
}

func (c *HTTPClient) RequestProductionCSID(ctx context.Context, req domain.ProductionCSIDRequest, csid *domain.CSID) (domain.ProductionCSIDResponse, error) {
	url := fmt.Sprintf("%s/production/csids", c.cfg.BaseURL)
	payload, err := json.Marshal(map[string]string{"compliance_request_id": req.ComplianceRequestID})
	if err != nil {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("failed to marshal production CSID request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("failed to create production CSID request: %w", err)
	}

	c.setHeaders(httpReq, csid)
	httpReq.Header.Set("Compliance-Request-ID", req.ComplianceRequestID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("production CSID request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("ZATCA HTTP request rejected with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("failed to read production CSID response: %w", err)
	}

	var prodResp domain.ProductionCSIDResponse
	if err := json.Unmarshal(body, &prodResp); err != nil {
		return domain.ProductionCSIDResponse{}, fmt.Errorf("failed to decode production CSID response: %w", err)
	}

	return prodResp, nil
}

func (c *HTTPClient) RenewProductionCSID(ctx context.Context, req domain.RenewalRequest, currentCSID *domain.CSID) (domain.RenewalResponse, error) {
	url := fmt.Sprintf("%s/production/csids", c.cfg.BaseURL)
	payload, err := json.Marshal(map[string]string{"csr": req.CSR})
	if err != nil {
		return domain.RenewalResponse{}, fmt.Errorf("failed to marshal renewal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(payload))
	if err != nil {
		return domain.RenewalResponse{}, fmt.Errorf("failed to create renewal request: %w", err)
	}

	c.setHeaders(httpReq, currentCSID)
	httpReq.Header.Set("OTP", req.OTP)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.RenewalResponse{}, fmt.Errorf("renewal request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.RenewalResponse{}, fmt.Errorf("ZATCA HTTP request rejected with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.RenewalResponse{}, fmt.Errorf("failed to read renewal response: %w", err)
	}

	var renewalResp domain.RenewalResponse
	if err := json.Unmarshal(body, &renewalResp); err != nil {
		return domain.RenewalResponse{}, fmt.Errorf("failed to decode renewal response: %w", err)
	}

	return renewalResp, nil
}

// setHeaders injects mandatory ZATCA headers and HTTP Basic Auth.
func (c *HTTPClient) setHeaders(req *http.Request, csid *domain.CSID) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Version", "V2")
	req.Header.Set("Accept-Language", "ar")

	if csid != nil && csid.BinarySecurityToken != "" && csid.Secret != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", csid.BinarySecurityToken, csid.Secret)))
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	}
}

// Ensure interface adherence
var _ ports.ZATCAClient = (*HTTPClient)(nil)
var _ ports.ZATCAOnboardingClient = (*HTTPClient)(nil)
