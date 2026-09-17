package ports

import (
	"context"
	"time"

	"erp/internal/fixedassets/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository defines data access for the Fixed Assets module.
type Repository interface {
	// Categories & Locations
	CreateAssetCategory(ctx context.Context, cat domain.AssetCategory) (*domain.AssetCategory, error)
	ListAssetCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetCategory, error)
	GetAssetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.AssetCategory, error)

	CreateAssetLocation(ctx context.Context, loc domain.AssetLocation) (*domain.AssetLocation, error)
	ListAssetLocations(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetLocation, error)
	GetAssetLocationByID(ctx context.Context, id uuid.UUID) (*domain.AssetLocation, error)

	// Depreciation Books
	CreateDepreciationBook(ctx context.Context, book domain.DepreciationBook) (*domain.DepreciationBook, error)
	ListDepreciationBooks(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationBook, error)
	GetDepreciationBookByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.DepreciationBook, error)

	// Assets
	CreateFixedAsset(ctx context.Context, asset domain.Asset) (*domain.Asset, error)
	GetFixedAssetByID(ctx context.Context, id uuid.UUID) (*domain.Asset, error)
	GetFixedAssetByNumber(ctx context.Context, legalEntityID uuid.UUID, assetNumber string) (*domain.Asset, error)
	ListFixedAssets(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Asset, error)
	ListActiveFixedAssets(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Asset, error)
	UpdateFixedAssetStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Asset, error)
	UpdateFixedAssetBookValue(ctx context.Context, id uuid.UUID, currentNBV, accumulatedDepr decimal.Decimal) (*domain.Asset, error)
	UpdateFixedAssetLocation(ctx context.Context, id, locationID uuid.UUID) (*domain.Asset, error)

	// Assignments
	CreateAssetAssignment(ctx context.Context, assignment domain.AssetAssignment) (*domain.AssetAssignment, error)
	ListAssignmentsByAsset(ctx context.Context, assetID uuid.UUID) ([]domain.AssetAssignment, error)
	EndAssetAssignment(ctx context.Context, assetID uuid.UUID, returnDate time.Time) error

	// Schedules
	CreateDepreciationSchedule(ctx context.Context, sched domain.DepreciationSchedule) (*domain.DepreciationSchedule, error)
	ListDepreciationSchedulesByAsset(ctx context.Context, assetID, bookID uuid.UUID) ([]domain.DepreciationSchedule, error)
	MarkSchedulePosted(ctx context.Context, assetID, bookID, runID uuid.UUID, year, month int) error

	// Depreciation Runs
	CreateDepreciationRun(ctx context.Context, run domain.DepreciationRun) (*domain.DepreciationRun, error)
	GetDepreciationRunByID(ctx context.Context, id uuid.UUID) (*domain.DepreciationRun, error)
	ListDepreciationRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationRun, error)
	UpdateDepreciationRunStatus(ctx context.Context, id uuid.UUID, status string, journalEntryID, approvedBy *uuid.UUID) (*domain.DepreciationRun, error)

	// Depreciation Lines
	CreateDepreciationLine(ctx context.Context, line domain.DepreciationLine) (*domain.DepreciationLine, error)
	ListDepreciationLinesByRun(ctx context.Context, runID uuid.UUID) ([]domain.DepreciationLine, error)

	// Transfers
	CreateAssetTransfer(ctx context.Context, transfer domain.Transfer) (*domain.Transfer, error)
	ListAssetTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Transfer, error)

	// Impairments
	CreateAssetImpairment(ctx context.Context, impairment domain.Impairment) (*domain.Impairment, error)

	// Disposals
	CreateAssetDisposal(ctx context.Context, disposal domain.Disposal) (*domain.Disposal, error)
	GetAssetDisposalByID(ctx context.Context, id uuid.UUID) (*domain.Disposal, error)
}

// Service defines business operations for the Fixed Assets module.
type Service interface {
	// Categories & Locations
	CreateAssetCategory(ctx context.Context, cat domain.AssetCategory) (*domain.AssetCategory, error)
	ListAssetCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetCategory, error)
	CreateAssetLocation(ctx context.Context, loc domain.AssetLocation) (*domain.AssetLocation, error)
	ListAssetLocations(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetLocation, error)
	ListDepreciationBooks(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationBook, error)

	// Assets
	RegisterAsset(ctx context.Context, asset domain.Asset) (*domain.Asset, error)
	GetAsset(ctx context.Context, id uuid.UUID) (*domain.Asset, error)
	ListAssets(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Asset, error)
	AssignAsset(ctx context.Context, assignment domain.AssetAssignment) (*domain.AssetAssignment, error)

	// Depreciation
	RunMonthlyDepreciation(ctx context.Context, legalEntityID, bookID uuid.UUID, year, month int, runBy *uuid.UUID) (*domain.DepreciationRun, []domain.DepreciationLine, error)
	PostDepreciationRun(ctx context.Context, runID uuid.UUID, approvedBy *uuid.UUID) (*domain.DepreciationRun, error)
	GetDepreciationRun(ctx context.Context, runID uuid.UUID) (*domain.DepreciationRun, []domain.DepreciationLine, error)
	ListDepreciationRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationRun, error)
	GetAssetSchedule(ctx context.Context, assetID, bookID uuid.UUID) ([]domain.DepreciationSchedule, error)

	// Lifecycle
	TransferAsset(ctx context.Context, transfer domain.Transfer) (*domain.Transfer, error)
	ListTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Transfer, error)
	DisposeAsset(ctx context.Context, disposal domain.Disposal) (*domain.Disposal, error)

	// Dashboard
	GetDashboardSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.DashboardSummary, error)
}
