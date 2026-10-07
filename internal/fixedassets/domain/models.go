package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AssetCategory represents an asset class/category with depreciation rules and account mappings.
type AssetCategory struct {
	ID                               uuid.UUID
	LegalEntityID                    uuid.UUID
	Code                             string
	NameEn                           string
	NameAr                           string
	DescriptionEn                    string
	DescriptionAr                    string
	DepreciationMethod               string // STRAIGHT_LINE
	UsefulLifeMonths                 int
	SalvageValueRate                 decimal.Decimal
	AssetAccountID                   *uuid.UUID
	AccumulatedDepreciationAccountID *uuid.UUID
	DepreciationExpenseAccountID     *uuid.UUID
	IsActive                         bool
	CreatedAt                        time.Time
	UpdatedAt                        time.Time
}

// AssetLocation represents a physical premises, building, floor, or room.
type AssetLocation struct {
	ID            uuid.UUID
	LegalEntityID uuid.UUID
	Code          string
	NameEn        string
	NameAr        string
	Building      string
	Floor         string
	Room          string
	IsActive      bool
	CreatedAt     time.Time
}

// DepreciationBook represents a depreciation book (COMMERCIAL for financial, TAX for ZATCA).
type DepreciationBook struct {
	ID            uuid.UUID
	LegalEntityID uuid.UUID
	Code          string
	NameEn        string
	NameAr        string
	IsPrimary     bool
	CreatedAt     time.Time
}

// Asset represents a fixed asset master record.
type Asset struct {
	ID                      uuid.UUID
	LegalEntityID           uuid.UUID
	CategoryID              uuid.UUID
	LocationID              *uuid.UUID
	AssetNumber             string
	SerialNumber            string
	Barcode                 string
	NameEn                  string
	NameAr                  string
	DescriptionEn           string
	DescriptionAr           string
	AcquisitionDate         time.Time
	InServiceDate           time.Time
	AcquisitionCost         decimal.Decimal
	SalvageValue            decimal.Decimal
	UsefulLifeMonths        int
	DepreciationMethod      string
	Status                  string // DRAFT, ACTIVE, FULLY_DEPRECIATED, DISPOSED, WRITTEN_OFF
	CurrentNetBookValue     decimal.Decimal
	AccumulatedDepreciation decimal.Decimal
	ImpairmentTotal         decimal.Decimal
	SupplierID              *uuid.UUID
	PurchaseInvoiceID       *uuid.UUID
	CreatedAt               time.Time
	UpdatedAt               time.Time

	// Joined fields
	CategoryCode                     string
	CategoryNameEn                   string
	CategoryNameAr                   string
	DepreciationExpenseAccountID     *uuid.UUID
	AccumulatedDepreciationAccountID *uuid.UUID
	AssetAccountID                   *uuid.UUID
	LocationCode                     string
	LocationNameEn                   string
	LocationNameAr                   string
}

// AssetAssignment represents custody assignment to an employee, department, or project.
type AssetAssignment struct {
	ID             uuid.UUID
	AssetID        uuid.UUID
	AssignedToType string // EMPLOYEE, DEPARTMENT, PROJECT
	EmployeeID     *uuid.UUID
	DepartmentID   *uuid.UUID
	AssignedDate   time.Time
	ReturnDate     *time.Time
	Notes          string
	IsCurrent      bool
	CreatedAt      time.Time

	// Joined employee fields
	FirstNameEn    string
	LastNameEn     string
	FirstNameAr    string
	LastNameAr     string
	EmployeeNumber string
}

// DepreciationSchedule represents an individual period projection.
type DepreciationSchedule struct {
	ID                      uuid.UUID
	AssetID                 uuid.UUID
	BookID                  uuid.UUID
	PeriodYear              int
	PeriodMonth             int
	OpeningNetBookValue     decimal.Decimal
	DepreciationAmount      decimal.Decimal
	ClosingNetBookValue     decimal.Decimal
	AccumulatedDepreciation decimal.Decimal
	IsPosted                bool
	DepreciationRunID       *uuid.UUID
	CreatedAt               time.Time
}

// DepreciationRun represents a monthly batch calculation and GL posting run.
type DepreciationRun struct {
	ID                      uuid.UUID
	LegalEntityID           uuid.UUID
	BookID                  uuid.UUID
	RunNumber               string
	PeriodYear              int
	PeriodMonth             int
	RunDate                 time.Time
	TotalAssetsProcessed    int
	TotalDepreciationAmount decimal.Decimal
	Status                  string // DRAFT, CALCULATED, APPROVED, POSTED
	JournalEntryID          *uuid.UUID
	CreatedBy               *uuid.UUID
	ApprovedBy              *uuid.UUID
	CreatedAt               time.Time
	UpdatedAt               time.Time

	// Joined fields
	BookCode   string
	BookNameEn string
	BookNameAr string
}

// DepreciationLine represents an individual asset's depreciation within a run.
type DepreciationLine struct {
	ID                   uuid.UUID
	RunID                uuid.UUID
	AssetID              uuid.UUID
	AssetNumber          string
	CostBasis            decimal.Decimal
	DepreciationAmount   decimal.Decimal
	AccumulatedBefore    decimal.Decimal
	AccumulatedAfter     decimal.Decimal
	NetBookValueAfter    decimal.Decimal
	ExpenseAccountID     *uuid.UUID
	AccumulatedAccountID *uuid.UUID
	CreatedAt            time.Time

	// Joined fields
	AssetNameEn string
	AssetNameAr string
}

// Transfer records asset relocation or custodian change.
type Transfer struct {
	ID             uuid.UUID
	LegalEntityID  uuid.UUID
	AssetID        uuid.UUID
	TransferNumber string
	TransferDate   time.Time
	FromLocationID *uuid.UUID
	ToLocationID   uuid.UUID
	FromEmployeeID *uuid.UUID
	ToEmployeeID   *uuid.UUID
	Reason         string
	Status         string
	TransferredBy  *uuid.UUID
	CreatedAt      time.Time

	// Joined fields
	FromLocationCode string
	FromLocationName string
	ToLocationCode   string
	ToLocationName   string
	FromEmployeeName string
	ToEmployeeName   string
}

// Impairment represents an extraordinary carrying value reduction.
type Impairment struct {
	ID                   uuid.UUID
	LegalEntityID        uuid.UUID
	AssetID              uuid.UUID
	ImpairmentNumber     string
	ImpairmentDate       time.Time
	CarryingAmountBefore decimal.Decimal
	RecoverableAmount    decimal.Decimal
	ImpairmentLoss       decimal.Decimal
	Reason               string
	JournalEntryID       *uuid.UUID
	CreatedBy            *uuid.UUID
	CreatedAt            time.Time
}

// Disposal represents retirement, scrap, or sale of an asset.
type Disposal struct {
	ImpairmentAccountID *uuid.UUID
	// Posting account selections are captured on the linked immutable journal.
	ProceedsAccountID                 *uuid.UUID
	GainLossAccountID                 *uuid.UUID
	ID                                uuid.UUID
	LegalEntityID                     uuid.UUID
	AssetID                           uuid.UUID
	DisposalNumber                    string
	DisposalDate                      time.Time
	DisposalType                      string // SALE, SCRAP, WRITE_OFF, DONATION
	ProceedsAmount                    decimal.Decimal
	CostAtDisposal                    decimal.Decimal
	AccumulatedDepreciationAtDisposal decimal.Decimal
	NetBookValueAtDisposal            decimal.Decimal
	GainOrLossAmount                  decimal.Decimal // Proceeds - Net Book Value
	JournalEntryID                    *uuid.UUID
	CustomerID                        *uuid.UUID
	Reason                            string
	Status                            string
	CreatedBy                         *uuid.UUID
	CreatedAt                         time.Time
}

// DashboardSummary aggregates KPI metrics for fixed assets.
type DashboardSummary struct {
	TotalHistoricalCost  decimal.Decimal
	TotalNetBookValue    decimal.Decimal
	TotalAccumulatedDepr decimal.Decimal
	TotalActiveAssets    int
	InServiceRate        string
}

// CalculateStraightLineDepreciation computes one month of straight-line depreciation.
// Depreciable Base = Cost - Salvage.
// Monthly Depreciation = Depreciable Base / UsefulLifeMonths.
// Depreciation cannot reduce carrying value below Salvage Value.
func CalculateStraightLineDepreciation(cost, salvage decimal.Decimal, usefulLifeMonths int, currentNBV decimal.Decimal) decimal.Decimal {
	if usefulLifeMonths <= 0 || currentNBV.LessThanOrEqual(salvage) {
		return decimal.Zero
	}

	depreciableBase := cost.Sub(salvage)
	if depreciableBase.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	monthlyDepr := depreciableBase.Div(decimal.NewFromInt(int64(usefulLifeMonths))).Round(2)
	remainingToSalvage := currentNBV.Sub(salvage)

	if monthlyDepr.GreaterThan(remainingToSalvage) {
		monthlyDepr = remainingToSalvage
	}

	if monthlyDepr.LessThan(decimal.Zero) {
		return decimal.Zero
	}

	return monthlyDepr
}

// GenerateDepreciationSchedule generates full monthly schedule from in-service date.
func GenerateDepreciationSchedule(asset Asset, bookID uuid.UUID, startDate time.Time) []DepreciationSchedule {
	if asset.UsefulLifeMonths <= 0 {
		return nil
	}

	schedules := make([]DepreciationSchedule, 0, asset.UsefulLifeMonths)
	depreciableBase := asset.AcquisitionCost.Sub(asset.SalvageValue)
	monthlyDepr := depreciableBase.Div(decimal.NewFromInt(int64(asset.UsefulLifeMonths))).Round(2)

	currentNBV := asset.AcquisitionCost
	accumulated := decimal.Zero

	currYear := startDate.Year()
	currMonth := int(startDate.Month())

	for i := 1; i <= asset.UsefulLifeMonths; i++ {
		openingNBV := currentNBV
		depr := monthlyDepr

		remaining := openingNBV.Sub(asset.SalvageValue)
		if depr.GreaterThan(remaining) {
			depr = remaining
		}
		if depr.LessThan(decimal.Zero) {
			depr = decimal.Zero
		}

		closingNBV := openingNBV.Sub(depr)
		accumulated = accumulated.Add(depr)

		schedules = append(schedules, DepreciationSchedule{
			ID:                      uuid.New(),
			AssetID:                 asset.ID,
			BookID:                  bookID,
			PeriodYear:              currYear,
			PeriodMonth:             currMonth,
			OpeningNetBookValue:     openingNBV,
			DepreciationAmount:      depr,
			ClosingNetBookValue:     closingNBV,
			AccumulatedDepreciation: accumulated,
			IsPosted:                false,
		})

		currentNBV = closingNBV
		if currentNBV.LessThanOrEqual(asset.SalvageValue) {
			break
		}

		currMonth++
		if currMonth > 12 {
			currMonth = 1
			currYear++
		}
	}

	return schedules
}
