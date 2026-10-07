package client

import (
	"context"
	"erp/internal/zatca/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPFailureCannotBeReportedAsRegulatorySuccess(t *testing.T) {
	for _, status := range []int{400, 401, 429, 500, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"clearanceStatus":"CLEARED","reportingStatus":"REPORTED","dispositionMessage":"ISSUED"}`))
			}))
			defer server.Close()
			client := NewHTTPClient(Config{BaseURL: server.URL, Environment: domain.EnvironmentSandbox})
			ctx := context.Background()
			csid := &domain.CSID{}
			if _, err := client.Clearance(ctx, domain.ClearanceRequest{}, csid); err == nil {
				t.Fatal("failed clearance accepted")
			}
			if _, err := client.Report(ctx, domain.ReportingRequest{}, csid); err == nil {
				t.Fatal("failed reporting accepted")
			}
			if _, err := client.RequestComplianceCSID(ctx, domain.ComplianceCSIDRequest{}); err == nil {
				t.Fatal("failed onboarding accepted")
			}
			if _, err := client.RunComplianceChecks(ctx, domain.ComplianceCheckRequest{}, csid); err == nil {
				t.Fatal("failed compliance accepted")
			}
			if _, err := client.RequestProductionCSID(ctx, domain.ProductionCSIDRequest{}, csid); err == nil {
				t.Fatal("failed production request accepted")
			}
			if _, err := client.RenewProductionCSID(ctx, domain.RenewalRequest{}, csid); err == nil {
				t.Fatal("failed renewal accepted")
			}
		})
	}
}
