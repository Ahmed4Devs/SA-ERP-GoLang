package domain

import "errors"

var (
	ErrReportNotFound        = errors.New("report not found")
	ErrLegalEntityRequired   = errors.New("legal_entity_id is required")
	ErrInvalidDateRange      = errors.New("invalid date range: start_date must be before end_date")
	ErrEntityIDRequired      = errors.New("customer_id or supplier_id is required for statement report")
	ErrUnsupportedExportType = errors.New("unsupported export format")
)
