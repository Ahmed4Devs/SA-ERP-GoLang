package services

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"erp/internal/gosi/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	ModeManual     = "manual"
	ModeFile       = "file"
	ModeSimulation = "simulation"
	ModeOfficial   = "official"
)

var (
	ErrManualSubmissionRequired       = errors.New("GOSI automatic submission is unavailable in manual/file mode; export the report and retain the actual portal receipt")
	ErrOfficialIntegrationUnavailable = errors.New("GOSI official integration is unavailable: a verified API contract, credentials, and acceptance tests are required")
	ErrStatusUnavailable              = errors.New("GOSI portal status is unavailable without an official integration and an actual portal receipt")
)

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}
type SubmissionResult struct {
	SubmissionID       string    `json:"submission_id"`
	Status             string    `json:"status"`
	Mode               string    `json:"mode"`
	ExternalSubmission bool      `json:"external_submission"`
	SubmittedAt        time.Time `json:"submitted_at"`
}
type SubmissionStatus struct {
	SubmissionID string    `json:"submission_id"`
	Status       string    `json:"status"`
	Mode         string    `json:"mode"`
	Message      string    `json:"message"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type GOSIAdapter interface {
	Validate(context.Context, []byte) (ValidationResult, error)
	Export(context.Context, uuid.UUID) ([]byte, error)
	Submit(context.Context, uuid.UUID) (SubmissionResult, error)
	Status(context.Context, string) (SubmissionStatus, error)
}

// DefaultGOSIAdapter defaults to local file export. No local action represents
// acceptance by GOSI. Simulation never changes a real batch's submission state.
type DefaultGOSIAdapter struct {
	repo ports.Repository
	mode string
}

func NewDefaultGOSIAdapter(repo ports.Repository) GOSIAdapter {
	return NewGOSIAdapter(repo, ModeFile)
}
func NewGOSIAdapter(repo ports.Repository, mode string) GOSIAdapter {
	if strings.TrimSpace(mode) == "" {
		mode = ModeFile
	}
	return &DefaultGOSIAdapter{repo: repo, mode: strings.ToLower(strings.TrimSpace(mode))}
}

// This validates the ERP exchange format, not a certified GOSI portal schema.
func (a *DefaultGOSIAdapter) Validate(ctx context.Context, file []byte) (ValidationResult, error) {
	if err := ctx.Err(); err != nil {
		return ValidationResult{}, err
	}
	result := ValidationResult{Errors: []string{}, Warnings: []string{"Local ERP export validation only; GOSI portal acceptance is not verified."}}
	if len(file) == 0 || len(file) > 16*1024*1024 {
		result.Errors = append(result.Errors, "report must contain 1 byte to 16 MiB")
		return result, nil
	}
	// Reject directives/DTDs and multiple documents before decoding values.
	decoder := xml.NewDecoder(bytes.NewReader(file))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, "malformed XML: "+err.Error())
			return result, nil
		}
		switch token := token.(type) {
		case xml.Directive:
			result.Errors = append(result.Errors, "XML directives and DTDs are not allowed")
		case xml.StartElement:
			if depth == 0 {
				roots++
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(token)) != "" {
				result.Errors = append(result.Errors, "unexpected content outside the root element")
			}
		}
	}
	if roots != 1 {
		result.Errors = append(result.Errors, "exactly one report root is required")
	}
	var report contributionReport
	if err := xml.Unmarshal(file, &report); err != nil {
		result.Errors = append(result.Errors, "invalid report: "+err.Error())
		return result, nil
	}
	if report.XMLName.Space != "urn:erp:gosi:export:v1" {
		result.Errors = append(result.Errors, "unsupported report namespace")
	}
	if report.Year < 2000 || report.Year > 9999 || report.Month < 1 || report.Month > 12 {
		result.Errors = append(result.Errors, "invalid reporting period")
	}
	if id, err := uuid.Parse(report.LegalEntity.ID); err != nil || id == uuid.Nil {
		result.Errors = append(result.Errors, "valid LegalEntity id is required")
	}
	if _, err := time.Parse(time.RFC3339, report.GeneratedAt); err != nil {
		result.Errors = append(result.Errors, "valid GeneratedAt is required")
	}
	emp, er := decimal.Zero, decimal.Zero
	seen := map[string]bool{}
	for _, employee := range report.Employees {
		if id, err := uuid.Parse(employee.ID); err != nil || id == uuid.Nil || seen[employee.ID] {
			result.Errors = append(result.Errors, "employee IDs must be valid and unique")
		}
		seen[employee.ID] = true
		employeeSum, employerSum := decimal.Zero, decimal.Zero
		if len(employee.Rules) == 0 {
			result.Errors = append(result.Errors, "each employee requires contribution rule snapshots")
		}
		for _, rule := range employee.Rules {
			if id, err := uuid.Parse(rule.VersionID); err != nil || id == uuid.Nil {
				result.Errors = append(result.Errors, "valid rule version id is required")
			}
			for _, value := range []string{rule.Basis, rule.EmployeeRate, rule.EmployerRate, rule.EmployeeAmount, rule.EmployerAmount} {
				if d, err := decimal.NewFromString(value); err != nil || d.IsNegative() {
					result.Errors = append(result.Errors, "rule values must be nonnegative decimal numbers")
				}
			}
			employeeAmount, _ := decimal.NewFromString(rule.EmployeeAmount)
			employerAmount, _ := decimal.NewFromString(rule.EmployerAmount)
			employeeSum = employeeSum.Add(employeeAmount)
			employerSum = employerSum.Add(employerAmount)
		}
		if amount, err := decimal.NewFromString(employee.EmployeeAmount); err != nil || !amount.Equal(employeeSum) {
			result.Errors = append(result.Errors, "employee contribution does not match rule totals")
		}
		if amount, err := decimal.NewFromString(employee.EmployerAmount); err != nil || !amount.Equal(employerSum) {
			result.Errors = append(result.Errors, "employer contribution does not match rule totals")
		}
		emp, er = emp.Add(employeeSum), er.Add(employerSum)
	}
	if report.Summary.SubscriberCount <= 0 || report.Summary.SubscriberCount != len(report.Employees) {
		result.Errors = append(result.Errors, "subscriber count does not match employee details")
	}
	for _, total := range []struct {
		raw      string
		expected decimal.Decimal
	}{{report.Summary.EmployeeTotal, emp}, {report.Summary.EmployerTotal, er}, {report.Summary.GrandTotal, emp.Add(er)}} {
		if amount, err := decimal.NewFromString(total.raw); err != nil || !amount.Equal(total.expected) {
			result.Errors = append(result.Errors, "summary totals do not match employee details")
		}
	}
	result.Valid = len(result.Errors) == 0
	return result, nil
}
func (a *DefaultGOSIAdapter) Export(ctx context.Context, runID uuid.UUID) ([]byte, error) {
	batch, err := a.repo.GetExportBatchByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("fetch export batch: %w", err)
	}
	if batch.FileContent == nil {
		return nil, fmt.Errorf("batch %s has no file content", runID)
	}
	file := []byte(*batch.FileContent)
	result, err := a.Validate(ctx, file)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return nil, fmt.Errorf("invalid local GOSI report: %s", strings.Join(result.Errors, "; "))
	}
	return file, nil
}
func (a *DefaultGOSIAdapter) Submit(ctx context.Context, runID uuid.UUID) (SubmissionResult, error) {
	switch a.mode {
	case ModeManual, ModeFile:
		return SubmissionResult{}, ErrManualSubmissionRequired
	case ModeOfficial:
		return SubmissionResult{}, ErrOfficialIntegrationUnavailable
	case ModeSimulation:
		if _, err := a.Export(ctx, runID); err != nil {
			return SubmissionResult{}, err
		}
		return SubmissionResult{SubmissionID: "SIMULATED:" + runID.String(), Status: "SIMULATED_SUBMITTED", Mode: a.mode, ExternalSubmission: false, SubmittedAt: time.Now().UTC()}, nil
	default:
		return SubmissionResult{}, fmt.Errorf("unsupported GOSI integration mode %q", a.mode)
	}
}
func (a *DefaultGOSIAdapter) Status(ctx context.Context, submissionID string) (SubmissionStatus, error) {
	if a.mode == ModeOfficial {
		return SubmissionStatus{}, ErrOfficialIntegrationUnavailable
	}
	if a.mode != ModeSimulation {
		return SubmissionStatus{}, ErrStatusUnavailable
	}
	if !strings.HasPrefix(submissionID, "SIMULATED:") {
		return SubmissionStatus{}, ErrStatusUnavailable
	}
	id, err := uuid.Parse(strings.TrimPrefix(submissionID, "SIMULATED:"))
	if err != nil {
		return SubmissionStatus{}, ErrStatusUnavailable
	}
	if _, err := a.Export(ctx, id); err != nil {
		return SubmissionStatus{}, err
	}
	return SubmissionStatus{SubmissionID: submissionID, Status: "SIMULATED_SUBMITTED", Mode: a.mode, Message: "Simulation only; no report was sent to or confirmed by GOSI.", UpdatedAt: time.Now().UTC()}, nil
}
