package ports

import (
	"context"
	"time"

	"erp/internal/budgeting/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateBudgetRequest struct {
	LegalEntityID uuid.UUID         `json:"legal_entity_id"`
	Code          string            `json:"code"`
	NameEn        string            `json:"name_en"`
	NameAr        string            `json:"name_ar"`
	DescriptionEn string            `json:"description_en"`
	DescriptionAr string            `json:"description_ar"`
	FiscalYearID  uuid.UUID         `json:"fiscal_year_id"`
	CurrencyCode  string            `json:"currency_code"`
	BudgetType    domain.BudgetType `json:"budget_type"`
	CreatedBy     *uuid.UUID        `json:"created_by,omitempty"`
}

type CreateVersionRequest struct {
	BudgetID    uuid.UUID  `json:"budget_id"`
	VersionName string     `json:"version_name"`
	Notes       string     `json:"notes"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
}

type AddBudgetLineRequest struct {
	BudgetID          uuid.UUID       `json:"budget_id"`
	BudgetVersionID   uuid.UUID       `json:"budget_version_id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	AccountID         uuid.UUID       `json:"account_id"`
	DepartmentID      *uuid.UUID      `json:"department_id,omitempty"`
	CostCenterID      *uuid.UUID      `json:"cost_center_id,omitempty"`
	ProjectCode       string          `json:"project_code,omitempty"`
	FiscalPeriodID    *uuid.UUID      `json:"fiscal_period_id,omitempty"`
	Amount            decimal.Decimal `json:"amount"`
	LineDescriptionEn string          `json:"line_description_en"`
	LineDescriptionAr string          `json:"line_description_ar"`
}

type TransferBudgetRequest struct {
	LegalEntityID uuid.UUID       `json:"legal_entity_id"`
	SourceLineID  uuid.UUID       `json:"source_line_id"`
	TargetLineID  uuid.UUID       `json:"target_line_id"`
	Amount        decimal.Decimal `json:"amount"`
	Reason        string          `json:"reason"`
	TransferredBy *uuid.UUID      `json:"transferred_by,omitempty"`
}

type Repository interface {
	GetBudgetByID(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.Budget, error)
	GetBudgetByCode(ctx context.Context, code string, legalEntityID uuid.UUID) (*domain.Budget, error)
	ListBudgets(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Budget, error)
	CreateBudget(ctx context.Context, b *domain.Budget) (*domain.Budget, error)
	UpdateBudgetStatus(ctx context.Context, id, legalEntityID uuid.UUID, status domain.BudgetStatus, approverID *uuid.UUID, approvedAt *time.Time) (*domain.Budget, error)
	UpdateBudgetTotals(ctx context.Context, id, legalEntityID uuid.UUID, budgeted, actual, committed decimal.Decimal) (*domain.Budget, error)
	DeleteBudget(ctx context.Context, id, legalEntityID uuid.UUID) error

	GetBudgetVersions(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetVersion, error)
	GetBudgetVersionByID(ctx context.Context, id uuid.UUID) (*domain.BudgetVersion, error)
	GetActiveBudgetVersion(ctx context.Context, budgetID uuid.UUID) (*domain.BudgetVersion, error)
	CreateBudgetVersion(ctx context.Context, v *domain.BudgetVersion) (*domain.BudgetVersion, error)
	UpdateBudgetVersionStatus(ctx context.Context, id uuid.UUID, status domain.VersionStatus, approverID *uuid.UUID, approvedAt *time.Time) (*domain.BudgetVersion, error)

	ListBudgetLines(ctx context.Context, versionID uuid.UUID) ([]*domain.BudgetLine, error)
	GetBudgetLineByID(ctx context.Context, id uuid.UUID) (*domain.BudgetLine, error)
	CreateBudgetLine(ctx context.Context, l *domain.BudgetLine) (*domain.BudgetLine, error)
	UpdateBudgetLine(ctx context.Context, l *domain.BudgetLine) (*domain.BudgetLine, error)
	DeleteBudgetLine(ctx context.Context, id uuid.UUID) error
	RecalculateBudgetTotals(ctx context.Context, budgetID, legalEntityID uuid.UUID, versionID uuid.UUID) error

	CreateBudgetTransfer(ctx context.Context, t *domain.BudgetTransfer) (*domain.BudgetTransfer, error)
	ListBudgetTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BudgetTransfer, error)
}

type Service interface {
	CreateBudget(ctx context.Context, req CreateBudgetRequest) (*domain.Budget, error)
	GetBudget(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.Budget, error)
	ListBudgets(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.Budget, error)
	SubmitBudget(ctx context.Context, id, legalEntityID uuid.UUID) (*domain.Budget, error)
	ApproveBudget(ctx context.Context, id, legalEntityID, approverID uuid.UUID) (*domain.Budget, error)

	CreateVersion(ctx context.Context, req CreateVersionRequest) (*domain.BudgetVersion, error)
	ListVersions(ctx context.Context, budgetID uuid.UUID) ([]*domain.BudgetVersion, error)

	AddBudgetLine(ctx context.Context, req AddBudgetLineRequest) (*domain.BudgetLine, error)
	ListBudgetLines(ctx context.Context, versionID uuid.UUID) ([]*domain.BudgetLine, error)

	TransferBudget(ctx context.Context, req TransferBudgetRequest) (*domain.BudgetTransfer, error)
	ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]*domain.BudgetTransfer, error)

	GetBudgetVsActual(ctx context.Context, budgetID, legalEntityID uuid.UUID) ([]*domain.BudgetVsActualItem, error)
	GetBudgetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.BudgetSummary, error)
}
