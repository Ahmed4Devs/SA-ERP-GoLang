package postgres

import (
	"context"
	"fmt"
	"time"

	"erp/internal/fixedassets/domain"
	"erp/internal/fixedassets/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) ports.Repository {
	return &Repository{db: db}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// ----------------------------------------------------------------------
// 1. Categories & Locations
// ----------------------------------------------------------------------

func (r *Repository) CreateAssetCategory(ctx context.Context, cat domain.AssetCategory) (*domain.AssetCategory, error) {
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateAssetCategory(ctx, sqlc.CreateAssetCategoryParams{
		ID:                               toPgUUID(cat.ID),
		LegalEntityID:                    toPgUUID(cat.LegalEntityID),
		Code:                             cat.Code,
		NameEn:                           cat.NameEn,
		NameAr:                           cat.NameAr,
		DescriptionEn:                    toPgText(cat.DescriptionEn),
		DescriptionAr:                    toPgText(cat.DescriptionAr),
		DepreciationMethod:               cat.DepreciationMethod,
		UsefulLifeMonths:                 int32(cat.UsefulLifeMonths),
		SalvageValueRate:                 toPgNumeric(cat.SalvageValueRate),
		AssetAccountID:                   toPgUUIDPtr(cat.AssetAccountID),
		AccumulatedDepreciationAccountID: toPgUUIDPtr(cat.AccumulatedDepreciationAccountID),
		DepreciationExpenseAccountID:     toPgUUIDPtr(cat.DepreciationExpenseAccountID),
		IsActive:                         cat.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset category: %w", err)
	}

	return &domain.AssetCategory{
		ID:                               toUUID(row.ID),
		LegalEntityID:                    toUUID(row.LegalEntityID),
		Code:                             row.Code,
		NameEn:                           row.NameEn,
		NameAr:                           row.NameAr,
		DescriptionEn:                    fromPgText(row.DescriptionEn),
		DescriptionAr:                    fromPgText(row.DescriptionAr),
		DepreciationMethod:               row.DepreciationMethod,
		UsefulLifeMonths:                 int(row.UsefulLifeMonths),
		SalvageValueRate:                 fromPgNumeric(row.SalvageValueRate),
		AssetAccountID:                   toUUIDPtr(row.AssetAccountID),
		AccumulatedDepreciationAccountID: toUUIDPtr(row.AccumulatedDepreciationAccountID),
		DepreciationExpenseAccountID:     toUUIDPtr(row.DepreciationExpenseAccountID),
		IsActive:                         row.IsActive,
		CreatedAt:                        fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                        fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) ListAssetCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetCategory, error) {
	rows, err := r.queries(ctx).ListAssetCategories(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list asset categories: %w", err)
	}

	result := make([]domain.AssetCategory, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.AssetCategory{
			ID:                               toUUID(row.ID),
			LegalEntityID:                    toUUID(row.LegalEntityID),
			Code:                             row.Code,
			NameEn:                           row.NameEn,
			NameAr:                           row.NameAr,
			DescriptionEn:                    fromPgText(row.DescriptionEn),
			DescriptionAr:                    fromPgText(row.DescriptionAr),
			DepreciationMethod:               row.DepreciationMethod,
			UsefulLifeMonths:                 int(row.UsefulLifeMonths),
			SalvageValueRate:                 fromPgNumeric(row.SalvageValueRate),
			AssetAccountID:                   toUUIDPtr(row.AssetAccountID),
			AccumulatedDepreciationAccountID: toUUIDPtr(row.AccumulatedDepreciationAccountID),
			DepreciationExpenseAccountID:     toUUIDPtr(row.DepreciationExpenseAccountID),
			IsActive:                         row.IsActive,
			CreatedAt:                        fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                        fromPgTimestamptz(row.UpdatedAt),
		})
	}
	return result, nil
}

func (r *Repository) GetAssetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.AssetCategory, error) {
	row, err := r.queries(ctx).GetAssetCategoryByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset category by id: %w", err)
	}

	return &domain.AssetCategory{
		ID:                               toUUID(row.ID),
		LegalEntityID:                    toUUID(row.LegalEntityID),
		Code:                             row.Code,
		NameEn:                           row.NameEn,
		NameAr:                           row.NameAr,
		DescriptionEn:                    fromPgText(row.DescriptionEn),
		DescriptionAr:                    fromPgText(row.DescriptionAr),
		DepreciationMethod:               row.DepreciationMethod,
		UsefulLifeMonths:                 int(row.UsefulLifeMonths),
		SalvageValueRate:                 fromPgNumeric(row.SalvageValueRate),
		AssetAccountID:                   toUUIDPtr(row.AssetAccountID),
		AccumulatedDepreciationAccountID: toUUIDPtr(row.AccumulatedDepreciationAccountID),
		DepreciationExpenseAccountID:     toUUIDPtr(row.DepreciationExpenseAccountID),
		IsActive:                         row.IsActive,
		CreatedAt:                        fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                        fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) CreateAssetLocation(ctx context.Context, loc domain.AssetLocation) (*domain.AssetLocation, error) {
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateAssetLocation(ctx, sqlc.CreateAssetLocationParams{
		ID:            toPgUUID(loc.ID),
		LegalEntityID: toPgUUID(loc.LegalEntityID),
		Code:          loc.Code,
		NameEn:        loc.NameEn,
		NameAr:        loc.NameAr,
		Building:      toPgText(loc.Building),
		Floor:         toPgText(loc.Floor),
		Room:          toPgText(loc.Room),
		IsActive:      loc.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset location: %w", err)
	}

	return &domain.AssetLocation{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		Building:      fromPgText(row.Building),
		Floor:         fromPgText(row.Floor),
		Room:          fromPgText(row.Room),
		IsActive:      row.IsActive,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListAssetLocations(ctx context.Context, legalEntityID uuid.UUID) ([]domain.AssetLocation, error) {
	rows, err := r.queries(ctx).ListAssetLocations(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list asset locations: %w", err)
	}

	result := make([]domain.AssetLocation, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.AssetLocation{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			Building:      fromPgText(row.Building),
			Floor:         fromPgText(row.Floor),
			Room:          fromPgText(row.Room),
			IsActive:      row.IsActive,
			CreatedAt:     fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) GetAssetLocationByID(ctx context.Context, id uuid.UUID) (*domain.AssetLocation, error) {
	row, err := r.queries(ctx).GetAssetLocationByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset location by id: %w", err)
	}

	return &domain.AssetLocation{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		Building:      fromPgText(row.Building),
		Floor:         fromPgText(row.Floor),
		Room:          fromPgText(row.Room),
		IsActive:      row.IsActive,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

// ----------------------------------------------------------------------
// 2. Depreciation Books
// ----------------------------------------------------------------------

func (r *Repository) CreateDepreciationBook(ctx context.Context, book domain.DepreciationBook) (*domain.DepreciationBook, error) {
	if book.ID == uuid.Nil {
		book.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateDepreciationBook(ctx, sqlc.CreateDepreciationBookParams{
		ID:            toPgUUID(book.ID),
		LegalEntityID: toPgUUID(book.LegalEntityID),
		Code:          book.Code,
		NameEn:        book.NameEn,
		NameAr:        book.NameAr,
		IsPrimary:     book.IsPrimary,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create depreciation book: %w", err)
	}

	return &domain.DepreciationBook{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		IsPrimary:     row.IsPrimary,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListDepreciationBooks(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationBook, error) {
	rows, err := r.queries(ctx).ListDepreciationBooks(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list depreciation books: %w", err)
	}

	result := make([]domain.DepreciationBook, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.DepreciationBook{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			IsPrimary:     row.IsPrimary,
			CreatedAt:     fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) GetDepreciationBookByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.DepreciationBook, error) {
	row, err := r.queries(ctx).GetDepreciationBookByCode(ctx, sqlc.GetDepreciationBookByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get depreciation book by code: %w", err)
	}

	return &domain.DepreciationBook{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		IsPrimary:     row.IsPrimary,
		CreatedAt:     fromPgTimestamptz(row.CreatedAt),
	}, nil
}

// ----------------------------------------------------------------------
// 3. Fixed Assets
// ----------------------------------------------------------------------

func (r *Repository) CreateFixedAsset(ctx context.Context, asset domain.Asset) (*domain.Asset, error) {
	if asset.ID == uuid.Nil {
		asset.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateFixedAsset(ctx, sqlc.CreateFixedAssetParams{
		ID:                      toPgUUID(asset.ID),
		LegalEntityID:           toPgUUID(asset.LegalEntityID),
		CategoryID:              toPgUUID(asset.CategoryID),
		LocationID:              toPgUUIDPtr(asset.LocationID),
		AssetNumber:             asset.AssetNumber,
		SerialNumber:            toPgText(asset.SerialNumber),
		Barcode:                 toPgText(asset.Barcode),
		NameEn:                  asset.NameEn,
		NameAr:                  asset.NameAr,
		DescriptionEn:           toPgText(asset.DescriptionEn),
		DescriptionAr:           toPgText(asset.DescriptionAr),
		AcquisitionDate:         toPgDate(asset.AcquisitionDate),
		InServiceDate:           toPgDate(asset.InServiceDate),
		AcquisitionCost:         toPgNumeric(asset.AcquisitionCost),
		SalvageValue:            toPgNumeric(asset.SalvageValue),
		UsefulLifeMonths:        int32(asset.UsefulLifeMonths),
		DepreciationMethod:      asset.DepreciationMethod,
		Status:                  asset.Status,
		CurrentNetBookValue:     toPgNumeric(asset.CurrentNetBookValue),
		AccumulatedDepreciation: toPgNumeric(asset.AccumulatedDepreciation),
		ImpairmentTotal:         toPgNumeric(asset.ImpairmentTotal),
		SupplierID:              toPgUUIDPtr(asset.SupplierID),
		PurchaseInvoiceID:       toPgUUIDPtr(asset.PurchaseInvoiceID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create fixed asset: %w", err)
	}

	return &domain.Asset{
		ID:                      toUUID(row.ID),
		LegalEntityID:           toUUID(row.LegalEntityID),
		CategoryID:              toUUID(row.CategoryID),
		LocationID:              toUUIDPtr(row.LocationID),
		AssetNumber:             row.AssetNumber,
		SerialNumber:            fromPgText(row.SerialNumber),
		Barcode:                 fromPgText(row.Barcode),
		NameEn:                  row.NameEn,
		NameAr:                  row.NameAr,
		DescriptionEn:           fromPgText(row.DescriptionEn),
		DescriptionAr:           fromPgText(row.DescriptionAr),
		AcquisitionDate:         fromPgDate(row.AcquisitionDate),
		InServiceDate:           fromPgDate(row.InServiceDate),
		AcquisitionCost:         fromPgNumeric(row.AcquisitionCost),
		SalvageValue:            fromPgNumeric(row.SalvageValue),
		UsefulLifeMonths:        int(row.UsefulLifeMonths),
		DepreciationMethod:      row.DepreciationMethod,
		Status:                  row.Status,
		CurrentNetBookValue:     fromPgNumeric(row.CurrentNetBookValue),
		AccumulatedDepreciation: fromPgNumeric(row.AccumulatedDepreciation),
		ImpairmentTotal:         fromPgNumeric(row.ImpairmentTotal),
		SupplierID:              toUUIDPtr(row.SupplierID),
		PurchaseInvoiceID:       toUUIDPtr(row.PurchaseInvoiceID),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) GetFixedAssetByID(ctx context.Context, id uuid.UUID) (*domain.Asset, error) {
	row, err := r.queries(ctx).GetFixedAssetByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get fixed asset by id: %w", err)
	}

	return &domain.Asset{
		ID:                               toUUID(row.ID),
		LegalEntityID:                    toUUID(row.LegalEntityID),
		CategoryID:                       toUUID(row.CategoryID),
		LocationID:                       toUUIDPtr(row.LocationID),
		AssetNumber:                      row.AssetNumber,
		SerialNumber:                     fromPgText(row.SerialNumber),
		Barcode:                          fromPgText(row.Barcode),
		NameEn:                           row.NameEn,
		NameAr:                           row.NameAr,
		DescriptionEn:                    fromPgText(row.DescriptionEn),
		DescriptionAr:                    fromPgText(row.DescriptionAr),
		AcquisitionDate:                  fromPgDate(row.AcquisitionDate),
		InServiceDate:                    fromPgDate(row.InServiceDate),
		AcquisitionCost:                  fromPgNumeric(row.AcquisitionCost),
		SalvageValue:                     fromPgNumeric(row.SalvageValue),
		UsefulLifeMonths:                 int(row.UsefulLifeMonths),
		DepreciationMethod:               row.DepreciationMethod,
		Status:                           row.Status,
		CurrentNetBookValue:              fromPgNumeric(row.CurrentNetBookValue),
		AccumulatedDepreciation:          fromPgNumeric(row.AccumulatedDepreciation),
		ImpairmentTotal:                  fromPgNumeric(row.ImpairmentTotal),
		SupplierID:                       toUUIDPtr(row.SupplierID),
		PurchaseInvoiceID:                toUUIDPtr(row.PurchaseInvoiceID),
		CreatedAt:                        fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                        fromPgTimestamptz(row.UpdatedAt),
		CategoryCode:                     row.CategoryCode,
		CategoryNameEn:                   row.CategoryNameEn,
		CategoryNameAr:                   row.CategoryNameAr,
		DepreciationExpenseAccountID:     toUUIDPtr(row.DepreciationExpenseAccountID),
		AccumulatedDepreciationAccountID: toUUIDPtr(row.AccumulatedDepreciationAccountID),
		AssetAccountID:                   toUUIDPtr(row.AssetAccountID),
		LocationCode:                     fromPgText(row.LocationCode),
		LocationNameEn:                   fromPgText(row.LocationNameEn),
		LocationNameAr:                   fromPgText(row.LocationNameAr),
	}, nil
}

func (r *Repository) GetFixedAssetByNumber(ctx context.Context, legalEntityID uuid.UUID, assetNumber string) (*domain.Asset, error) {
	row, err := r.queries(ctx).GetFixedAssetByNumber(ctx, sqlc.GetFixedAssetByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		AssetNumber:   assetNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get fixed asset by number: %w", err)
	}

	return &domain.Asset{
		ID:                               toUUID(row.ID),
		LegalEntityID:                    toUUID(row.LegalEntityID),
		CategoryID:                       toUUID(row.CategoryID),
		LocationID:                       toUUIDPtr(row.LocationID),
		AssetNumber:                      row.AssetNumber,
		SerialNumber:                     fromPgText(row.SerialNumber),
		Barcode:                          fromPgText(row.Barcode),
		NameEn:                           row.NameEn,
		NameAr:                           row.NameAr,
		DescriptionEn:                    fromPgText(row.DescriptionEn),
		DescriptionAr:                    fromPgText(row.DescriptionAr),
		AcquisitionDate:                  fromPgDate(row.AcquisitionDate),
		InServiceDate:                    fromPgDate(row.InServiceDate),
		AcquisitionCost:                  fromPgNumeric(row.AcquisitionCost),
		SalvageValue:                     fromPgNumeric(row.SalvageValue),
		UsefulLifeMonths:                 int(row.UsefulLifeMonths),
		DepreciationMethod:               row.DepreciationMethod,
		Status:                           row.Status,
		CurrentNetBookValue:              fromPgNumeric(row.CurrentNetBookValue),
		AccumulatedDepreciation:          fromPgNumeric(row.AccumulatedDepreciation),
		ImpairmentTotal:                  fromPgNumeric(row.ImpairmentTotal),
		SupplierID:                       toUUIDPtr(row.SupplierID),
		PurchaseInvoiceID:                toUUIDPtr(row.PurchaseInvoiceID),
		CreatedAt:                        fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:                        fromPgTimestamptz(row.UpdatedAt),
		CategoryCode:                     row.CategoryCode,
		CategoryNameEn:                   row.CategoryNameEn,
		CategoryNameAr:                   row.CategoryNameAr,
		DepreciationExpenseAccountID:     toUUIDPtr(row.DepreciationExpenseAccountID),
		AccumulatedDepreciationAccountID: toUUIDPtr(row.AccumulatedDepreciationAccountID),
		AssetAccountID:                   toUUIDPtr(row.AssetAccountID),
		LocationCode:                     fromPgText(row.LocationCode),
		LocationNameEn:                   fromPgText(row.LocationNameEn),
		LocationNameAr:                   fromPgText(row.LocationNameAr),
	}, nil
}

func (r *Repository) ListFixedAssets(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Asset, error) {
	rows, err := r.queries(ctx).ListFixedAssets(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list fixed assets: %w", err)
	}

	result := make([]domain.Asset, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Asset{
			ID:                      toUUID(row.ID),
			LegalEntityID:           toUUID(row.LegalEntityID),
			CategoryID:              toUUID(row.CategoryID),
			LocationID:              toUUIDPtr(row.LocationID),
			AssetNumber:             row.AssetNumber,
			SerialNumber:            fromPgText(row.SerialNumber),
			Barcode:                 fromPgText(row.Barcode),
			NameEn:                  row.NameEn,
			NameAr:                  row.NameAr,
			DescriptionEn:           fromPgText(row.DescriptionEn),
			DescriptionAr:           fromPgText(row.DescriptionAr),
			AcquisitionDate:         fromPgDate(row.AcquisitionDate),
			InServiceDate:           fromPgDate(row.InServiceDate),
			AcquisitionCost:         fromPgNumeric(row.AcquisitionCost),
			SalvageValue:            fromPgNumeric(row.SalvageValue),
			UsefulLifeMonths:        int(row.UsefulLifeMonths),
			DepreciationMethod:      row.DepreciationMethod,
			Status:                  row.Status,
			CurrentNetBookValue:     fromPgNumeric(row.CurrentNetBookValue),
			AccumulatedDepreciation: fromPgNumeric(row.AccumulatedDepreciation),
			ImpairmentTotal:         fromPgNumeric(row.ImpairmentTotal),
			SupplierID:              toUUIDPtr(row.SupplierID),
			PurchaseInvoiceID:       toUUIDPtr(row.PurchaseInvoiceID),
			CreatedAt:               fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
			CategoryCode:            row.CategoryCode,
			CategoryNameEn:          row.CategoryNameEn,
			CategoryNameAr:          row.CategoryNameAr,
			LocationCode:            fromPgText(row.LocationCode),
			LocationNameEn:          fromPgText(row.LocationNameEn),
			LocationNameAr:          fromPgText(row.LocationNameAr),
		})
	}
	return result, nil
}

func (r *Repository) ListActiveFixedAssets(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Asset, error) {
	rows, err := r.queries(ctx).ListActiveFixedAssets(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list active fixed assets: %w", err)
	}

	result := make([]domain.Asset, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Asset{
			ID:                               toUUID(row.ID),
			LegalEntityID:                    toUUID(row.LegalEntityID),
			CategoryID:                       toUUID(row.CategoryID),
			LocationID:                       toUUIDPtr(row.LocationID),
			AssetNumber:                      row.AssetNumber,
			SerialNumber:                     fromPgText(row.SerialNumber),
			Barcode:                          fromPgText(row.Barcode),
			NameEn:                           row.NameEn,
			NameAr:                           row.NameAr,
			DescriptionEn:                    fromPgText(row.DescriptionEn),
			DescriptionAr:                    fromPgText(row.DescriptionAr),
			AcquisitionDate:                  fromPgDate(row.AcquisitionDate),
			InServiceDate:                    fromPgDate(row.InServiceDate),
			AcquisitionCost:                  fromPgNumeric(row.AcquisitionCost),
			SalvageValue:                     fromPgNumeric(row.SalvageValue),
			UsefulLifeMonths:                 int(row.UsefulLifeMonths),
			DepreciationMethod:               row.DepreciationMethod,
			Status:                           row.Status,
			CurrentNetBookValue:              fromPgNumeric(row.CurrentNetBookValue),
			AccumulatedDepreciation:          fromPgNumeric(row.AccumulatedDepreciation),
			ImpairmentTotal:                  fromPgNumeric(row.ImpairmentTotal),
			SupplierID:                       toUUIDPtr(row.SupplierID),
			PurchaseInvoiceID:                toUUIDPtr(row.PurchaseInvoiceID),
			CreatedAt:                        fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:                        fromPgTimestamptz(row.UpdatedAt),
			CategoryCode:                     row.CategoryCode,
			CategoryNameEn:                   row.CategoryNameEn,
			CategoryNameAr:                   row.CategoryNameAr,
			DepreciationExpenseAccountID:     toUUIDPtr(row.DepreciationExpenseAccountID),
			AccumulatedDepreciationAccountID: toUUIDPtr(row.AccumulatedDepreciationAccountID),
			AssetAccountID:                   toUUIDPtr(row.AssetAccountID),
			LocationCode:                     fromPgText(row.LocationCode),
			LocationNameEn:                   fromPgText(row.LocationNameEn),
			LocationNameAr:                   fromPgText(row.LocationNameAr),
		})
	}
	return result, nil
}

func (r *Repository) UpdateFixedAssetStatus(ctx context.Context, id uuid.UUID, status string) (*domain.Asset, error) {
	row, err := r.queries(ctx).UpdateFixedAssetStatus(ctx, sqlc.UpdateFixedAssetStatusParams{
		ID:     toPgUUID(id),
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update fixed asset status: %w", err)
	}

	return &domain.Asset{
		ID:                      toUUID(row.ID),
		LegalEntityID:           toUUID(row.LegalEntityID),
		CategoryID:              toUUID(row.CategoryID),
		LocationID:              toUUIDPtr(row.LocationID),
		AssetNumber:             row.AssetNumber,
		SerialNumber:            fromPgText(row.SerialNumber),
		Barcode:                 fromPgText(row.Barcode),
		NameEn:                  row.NameEn,
		NameAr:                  row.NameAr,
		DescriptionEn:           fromPgText(row.DescriptionEn),
		DescriptionAr:           fromPgText(row.DescriptionAr),
		AcquisitionDate:         fromPgDate(row.AcquisitionDate),
		InServiceDate:           fromPgDate(row.InServiceDate),
		AcquisitionCost:         fromPgNumeric(row.AcquisitionCost),
		SalvageValue:            fromPgNumeric(row.SalvageValue),
		UsefulLifeMonths:        int(row.UsefulLifeMonths),
		DepreciationMethod:      row.DepreciationMethod,
		Status:                  row.Status,
		CurrentNetBookValue:     fromPgNumeric(row.CurrentNetBookValue),
		AccumulatedDepreciation: fromPgNumeric(row.AccumulatedDepreciation),
		ImpairmentTotal:         fromPgNumeric(row.ImpairmentTotal),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) UpdateFixedAssetBookValue(ctx context.Context, id uuid.UUID, currentNBV, accumulatedDepr decimal.Decimal) (*domain.Asset, error) {
	row, err := r.queries(ctx).UpdateFixedAssetBookValue(ctx, sqlc.UpdateFixedAssetBookValueParams{
		ID:                      toPgUUID(id),
		CurrentNetBookValue:     toPgNumeric(currentNBV),
		AccumulatedDepreciation: toPgNumeric(accumulatedDepr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update fixed asset book value: %w", err)
	}

	return &domain.Asset{
		ID:                      toUUID(row.ID),
		LegalEntityID:           toUUID(row.LegalEntityID),
		CategoryID:              toUUID(row.CategoryID),
		LocationID:              toUUIDPtr(row.LocationID),
		AssetNumber:             row.AssetNumber,
		SerialNumber:            fromPgText(row.SerialNumber),
		Barcode:                 fromPgText(row.Barcode),
		NameEn:                  row.NameEn,
		NameAr:                  row.NameAr,
		DescriptionEn:           fromPgText(row.DescriptionEn),
		DescriptionAr:           fromPgText(row.DescriptionAr),
		AcquisitionDate:         fromPgDate(row.AcquisitionDate),
		InServiceDate:           fromPgDate(row.InServiceDate),
		AcquisitionCost:         fromPgNumeric(row.AcquisitionCost),
		SalvageValue:            fromPgNumeric(row.SalvageValue),
		UsefulLifeMonths:        int(row.UsefulLifeMonths),
		DepreciationMethod:      row.DepreciationMethod,
		Status:                  row.Status,
		CurrentNetBookValue:     fromPgNumeric(row.CurrentNetBookValue),
		AccumulatedDepreciation: fromPgNumeric(row.AccumulatedDepreciation),
		ImpairmentTotal:         fromPgNumeric(row.ImpairmentTotal),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) UpdateFixedAssetLocation(ctx context.Context, id, locationID uuid.UUID) (*domain.Asset, error) {
	row, err := r.queries(ctx).UpdateFixedAssetLocation(ctx, sqlc.UpdateFixedAssetLocationParams{
		ID:         toPgUUID(id),
		LocationID: toPgUUID(locationID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update fixed asset location: %w", err)
	}

	return &domain.Asset{
		ID:         toUUID(row.ID),
		LocationID: toUUIDPtr(row.LocationID),
		UpdatedAt:  fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

// ----------------------------------------------------------------------
// 4. Assignments
// ----------------------------------------------------------------------

func (r *Repository) CreateAssetAssignment(ctx context.Context, assignment domain.AssetAssignment) (*domain.AssetAssignment, error) {
	if assignment.ID == uuid.Nil {
		assignment.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateAssetAssignment(ctx, sqlc.CreateAssetAssignmentParams{
		ID:             toPgUUID(assignment.ID),
		AssetID:        toPgUUID(assignment.AssetID),
		AssignedToType: assignment.AssignedToType,
		EmployeeID:     toPgUUIDPtr(assignment.EmployeeID),
		DepartmentID:   toPgUUIDPtr(assignment.DepartmentID),
		AssignedDate:   toPgDate(assignment.AssignedDate),
		ReturnDate:     toPgDatePtr(assignment.ReturnDate),
		Notes:          toPgText(assignment.Notes),
		IsCurrent:      assignment.IsCurrent,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset assignment: %w", err)
	}

	return &domain.AssetAssignment{
		ID:             toUUID(row.ID),
		AssetID:        toUUID(row.AssetID),
		AssignedToType: row.AssignedToType,
		EmployeeID:     toUUIDPtr(row.EmployeeID),
		DepartmentID:   toUUIDPtr(row.DepartmentID),
		AssignedDate:   fromPgDate(row.AssignedDate),
		ReturnDate:     fromPgDatePtr(row.ReturnDate),
		Notes:          fromPgText(row.Notes),
		IsCurrent:      row.IsCurrent,
		CreatedAt:      fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListAssignmentsByAsset(ctx context.Context, assetID uuid.UUID) ([]domain.AssetAssignment, error) {
	rows, err := r.queries(ctx).ListAssignmentsByAsset(ctx, toPgUUID(assetID))
	if err != nil {
		return nil, fmt.Errorf("failed to list assignments by asset: %w", err)
	}

	result := make([]domain.AssetAssignment, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.AssetAssignment{
			ID:             toUUID(row.ID),
			AssetID:        toUUID(row.AssetID),
			AssignedToType: row.AssignedToType,
			EmployeeID:     toUUIDPtr(row.EmployeeID),
			DepartmentID:   toUUIDPtr(row.DepartmentID),
			AssignedDate:   fromPgDate(row.AssignedDate),
			ReturnDate:     fromPgDatePtr(row.ReturnDate),
			Notes:          fromPgText(row.Notes),
			IsCurrent:      row.IsCurrent,
			CreatedAt:      fromPgTimestamptz(row.CreatedAt),
			FirstNameEn:    fromPgText(row.FirstNameEn),
			LastNameEn:     fromPgText(row.LastNameEn),
			FirstNameAr:    fromPgText(row.FirstNameAr),
			LastNameAr:     fromPgText(row.LastNameAr),
			EmployeeNumber: fromPgText(row.EmployeeNumber),
		})
	}
	return result, nil
}

func (r *Repository) EndAssetAssignment(ctx context.Context, assetID uuid.UUID, returnDate time.Time) error {
	return r.queries(ctx).EndAssetAssignment(ctx, sqlc.EndAssetAssignmentParams{
		AssetID:    toPgUUID(assetID),
		ReturnDate: toPgDate(returnDate),
	})
}

// ----------------------------------------------------------------------
// 5. Schedules
// ----------------------------------------------------------------------

func (r *Repository) CreateDepreciationSchedule(ctx context.Context, sched domain.DepreciationSchedule) (*domain.DepreciationSchedule, error) {
	if sched.ID == uuid.Nil {
		sched.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateDepreciationSchedule(ctx, sqlc.CreateDepreciationScheduleParams{
		ID:                      toPgUUID(sched.ID),
		AssetID:                 toPgUUID(sched.AssetID),
		BookID:                  toPgUUID(sched.BookID),
		PeriodYear:              int32(sched.PeriodYear),
		PeriodMonth:             int32(sched.PeriodMonth),
		OpeningNetBookValue:     toPgNumeric(sched.OpeningNetBookValue),
		DepreciationAmount:      toPgNumeric(sched.DepreciationAmount),
		ClosingNetBookValue:     toPgNumeric(sched.ClosingNetBookValue),
		AccumulatedDepreciation: toPgNumeric(sched.AccumulatedDepreciation),
		IsPosted:                sched.IsPosted,
		DepreciationRunID:       toPgUUIDPtr(sched.DepreciationRunID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create depreciation schedule: %w", err)
	}

	return &domain.DepreciationSchedule{
		ID:                      toUUID(row.ID),
		AssetID:                 toUUID(row.AssetID),
		BookID:                  toUUID(row.BookID),
		PeriodYear:              int(row.PeriodYear),
		PeriodMonth:             int(row.PeriodMonth),
		OpeningNetBookValue:     fromPgNumeric(row.OpeningNetBookValue),
		DepreciationAmount:      fromPgNumeric(row.DepreciationAmount),
		ClosingNetBookValue:     fromPgNumeric(row.ClosingNetBookValue),
		AccumulatedDepreciation: fromPgNumeric(row.AccumulatedDepreciation),
		IsPosted:                row.IsPosted,
		DepreciationRunID:       toUUIDPtr(row.DepreciationRunID),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListDepreciationSchedulesByAsset(ctx context.Context, assetID, bookID uuid.UUID) ([]domain.DepreciationSchedule, error) {
	rows, err := r.queries(ctx).ListDepreciationSchedulesByAsset(ctx, sqlc.ListDepreciationSchedulesByAssetParams{
		AssetID: toPgUUID(assetID),
		BookID:  toPgUUID(bookID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list depreciation schedules: %w", err)
	}

	result := make([]domain.DepreciationSchedule, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.DepreciationSchedule{
			ID:                      toUUID(row.ID),
			AssetID:                 toUUID(row.AssetID),
			BookID:                  toUUID(row.BookID),
			PeriodYear:              int(row.PeriodYear),
			PeriodMonth:             int(row.PeriodMonth),
			OpeningNetBookValue:     fromPgNumeric(row.OpeningNetBookValue),
			DepreciationAmount:      fromPgNumeric(row.DepreciationAmount),
			ClosingNetBookValue:     fromPgNumeric(row.ClosingNetBookValue),
			AccumulatedDepreciation: fromPgNumeric(row.AccumulatedDepreciation),
			IsPosted:                row.IsPosted,
			DepreciationRunID:       toUUIDPtr(row.DepreciationRunID),
			CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) MarkSchedulePosted(ctx context.Context, assetID, bookID, runID uuid.UUID, year, month int) error {
	return r.queries(ctx).MarkSchedulePosted(ctx, sqlc.MarkSchedulePostedParams{
		AssetID:           toPgUUID(assetID),
		BookID:            toPgUUID(bookID),
		DepreciationRunID: toPgUUID(runID),
		PeriodYear:        int32(year),
		PeriodMonth:       int32(month),
	})
}

// ----------------------------------------------------------------------
// 6. Depreciation Runs & Lines
// ----------------------------------------------------------------------

func (r *Repository) CreateDepreciationRun(ctx context.Context, run domain.DepreciationRun) (*domain.DepreciationRun, error) {
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateDepreciationRun(ctx, sqlc.CreateDepreciationRunParams{
		ID:                      toPgUUID(run.ID),
		LegalEntityID:           toPgUUID(run.LegalEntityID),
		BookID:                  toPgUUID(run.BookID),
		RunNumber:               run.RunNumber,
		PeriodYear:              int32(run.PeriodYear),
		PeriodMonth:             int32(run.PeriodMonth),
		RunDate:                 toPgDate(run.RunDate),
		TotalAssetsProcessed:    int32(run.TotalAssetsProcessed),
		TotalDepreciationAmount: toPgNumeric(run.TotalDepreciationAmount),
		Status:                  run.Status,
		JournalEntryID:          toPgUUIDPtr(run.JournalEntryID),
		CreatedBy:               toPgUUIDPtr(run.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create depreciation run: %w", err)
	}

	return &domain.DepreciationRun{
		ID:                      toUUID(row.ID),
		LegalEntityID:           toUUID(row.LegalEntityID),
		BookID:                  toUUID(row.BookID),
		RunNumber:               row.RunNumber,
		PeriodYear:              int(row.PeriodYear),
		PeriodMonth:             int(row.PeriodMonth),
		RunDate:                 fromPgDate(row.RunDate),
		TotalAssetsProcessed:    int(row.TotalAssetsProcessed),
		TotalDepreciationAmount: fromPgNumeric(row.TotalDepreciationAmount),
		Status:                  row.Status,
		JournalEntryID:          toUUIDPtr(row.JournalEntryID),
		CreatedBy:               toUUIDPtr(row.CreatedBy),
		ApprovedBy:              toUUIDPtr(row.ApprovedBy),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) GetDepreciationRunByID(ctx context.Context, id uuid.UUID) (*domain.DepreciationRun, error) {
	row, err := r.queries(ctx).GetDepreciationRunByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get depreciation run by id: %w", err)
	}

	return &domain.DepreciationRun{
		ID:                      toUUID(row.ID),
		LegalEntityID:           toUUID(row.LegalEntityID),
		BookID:                  toUUID(row.BookID),
		RunNumber:               row.RunNumber,
		PeriodYear:              int(row.PeriodYear),
		PeriodMonth:             int(row.PeriodMonth),
		RunDate:                 fromPgDate(row.RunDate),
		TotalAssetsProcessed:    int(row.TotalAssetsProcessed),
		TotalDepreciationAmount: fromPgNumeric(row.TotalDepreciationAmount),
		Status:                  row.Status,
		JournalEntryID:          toUUIDPtr(row.JournalEntryID),
		CreatedBy:               toUUIDPtr(row.CreatedBy),
		ApprovedBy:              toUUIDPtr(row.ApprovedBy),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
		BookCode:                row.BookCode,
		BookNameEn:              row.BookNameEn,
		BookNameAr:              row.BookNameAr,
	}, nil
}

func (r *Repository) ListDepreciationRuns(ctx context.Context, legalEntityID uuid.UUID) ([]domain.DepreciationRun, error) {
	rows, err := r.queries(ctx).ListDepreciationRuns(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list depreciation runs: %w", err)
	}

	result := make([]domain.DepreciationRun, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.DepreciationRun{
			ID:                      toUUID(row.ID),
			LegalEntityID:           toUUID(row.LegalEntityID),
			BookID:                  toUUID(row.BookID),
			RunNumber:               row.RunNumber,
			PeriodYear:              int(row.PeriodYear),
			PeriodMonth:             int(row.PeriodMonth),
			RunDate:                 fromPgDate(row.RunDate),
			TotalAssetsProcessed:    int(row.TotalAssetsProcessed),
			TotalDepreciationAmount: fromPgNumeric(row.TotalDepreciationAmount),
			Status:                  row.Status,
			JournalEntryID:          toUUIDPtr(row.JournalEntryID),
			CreatedBy:               toUUIDPtr(row.CreatedBy),
			ApprovedBy:              toUUIDPtr(row.ApprovedBy),
			CreatedAt:               fromPgTimestamptz(row.CreatedAt),
			UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
			BookCode:                row.BookCode,
			BookNameEn:              row.BookNameEn,
			BookNameAr:              row.BookNameAr,
		})
	}
	return result, nil
}

func (r *Repository) UpdateDepreciationRunStatus(ctx context.Context, id uuid.UUID, status string, journalEntryID, approvedBy *uuid.UUID) (*domain.DepreciationRun, error) {
	row, err := r.queries(ctx).UpdateDepreciationRunStatus(ctx, sqlc.UpdateDepreciationRunStatusParams{
		Status:         status,
		JournalEntryID: toPgUUIDPtr(journalEntryID),
		ApprovedBy:     toPgUUIDPtr(approvedBy),
		ID:             toPgUUID(id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update depreciation run status: %w", err)
	}

	return &domain.DepreciationRun{
		ID:                      toUUID(row.ID),
		LegalEntityID:           toUUID(row.LegalEntityID),
		BookID:                  toUUID(row.BookID),
		RunNumber:               row.RunNumber,
		PeriodYear:              int(row.PeriodYear),
		PeriodMonth:             int(row.PeriodMonth),
		RunDate:                 fromPgDate(row.RunDate),
		TotalAssetsProcessed:    int(row.TotalAssetsProcessed),
		TotalDepreciationAmount: fromPgNumeric(row.TotalDepreciationAmount),
		Status:                  row.Status,
		JournalEntryID:          toUUIDPtr(row.JournalEntryID),
		CreatedBy:               toUUIDPtr(row.CreatedBy),
		ApprovedBy:              toUUIDPtr(row.ApprovedBy),
		CreatedAt:               fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:               fromPgTimestamptz(row.UpdatedAt),
	}, nil
}

func (r *Repository) CreateDepreciationLine(ctx context.Context, line domain.DepreciationLine) (*domain.DepreciationLine, error) {
	if line.ID == uuid.Nil {
		line.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateDepreciationLine(ctx, sqlc.CreateDepreciationLineParams{
		ID:                   toPgUUID(line.ID),
		RunID:                toPgUUID(line.RunID),
		AssetID:              toPgUUID(line.AssetID),
		AssetNumber:          line.AssetNumber,
		CostBasis:            toPgNumeric(line.CostBasis),
		DepreciationAmount:   toPgNumeric(line.DepreciationAmount),
		AccumulatedBefore:    toPgNumeric(line.AccumulatedBefore),
		AccumulatedAfter:     toPgNumeric(line.AccumulatedAfter),
		NetBookValueAfter:    toPgNumeric(line.NetBookValueAfter),
		ExpenseAccountID:     toPgUUIDPtr(line.ExpenseAccountID),
		AccumulatedAccountID: toPgUUIDPtr(line.AccumulatedAccountID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create depreciation line: %w", err)
	}

	return &domain.DepreciationLine{
		ID:                   toUUID(row.ID),
		RunID:                toUUID(row.RunID),
		AssetID:              toUUID(row.AssetID),
		AssetNumber:          row.AssetNumber,
		CostBasis:            fromPgNumeric(row.CostBasis),
		DepreciationAmount:   fromPgNumeric(row.DepreciationAmount),
		AccumulatedBefore:    fromPgNumeric(row.AccumulatedBefore),
		AccumulatedAfter:     fromPgNumeric(row.AccumulatedAfter),
		NetBookValueAfter:    fromPgNumeric(row.NetBookValueAfter),
		ExpenseAccountID:     toUUIDPtr(row.ExpenseAccountID),
		AccumulatedAccountID: toUUIDPtr(row.AccumulatedAccountID),
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListDepreciationLinesByRun(ctx context.Context, runID uuid.UUID) ([]domain.DepreciationLine, error) {
	rows, err := r.queries(ctx).ListDepreciationLinesByRun(ctx, toPgUUID(runID))
	if err != nil {
		return nil, fmt.Errorf("failed to list depreciation lines by run: %w", err)
	}

	result := make([]domain.DepreciationLine, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.DepreciationLine{
			ID:                   toUUID(row.ID),
			RunID:                toUUID(row.RunID),
			AssetID:              toUUID(row.AssetID),
			AssetNumber:          row.AssetNumber,
			CostBasis:            fromPgNumeric(row.CostBasis),
			DepreciationAmount:   fromPgNumeric(row.DepreciationAmount),
			AccumulatedBefore:    fromPgNumeric(row.AccumulatedBefore),
			AccumulatedAfter:     fromPgNumeric(row.AccumulatedAfter),
			NetBookValueAfter:    fromPgNumeric(row.NetBookValueAfter),
			ExpenseAccountID:     toUUIDPtr(row.ExpenseAccountID),
			AccumulatedAccountID: toUUIDPtr(row.AccumulatedAccountID),
			CreatedAt:            fromPgTimestamptz(row.CreatedAt),
			AssetNameEn:          row.AssetNameEn,
			AssetNameAr:          row.AssetNameAr,
		})
	}
	return result, nil
}

// ----------------------------------------------------------------------
// 7. Transfers
// ----------------------------------------------------------------------

func (r *Repository) CreateAssetTransfer(ctx context.Context, transfer domain.Transfer) (*domain.Transfer, error) {
	if transfer.ID == uuid.Nil {
		transfer.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateAssetTransfer(ctx, sqlc.CreateAssetTransferParams{
		ID:             toPgUUID(transfer.ID),
		LegalEntityID:  toPgUUID(transfer.LegalEntityID),
		AssetID:        toPgUUID(transfer.AssetID),
		TransferNumber: transfer.TransferNumber,
		TransferDate:   toPgDate(transfer.TransferDate),
		FromLocationID: toPgUUIDPtr(transfer.FromLocationID),
		ToLocationID:   toPgUUID(transfer.ToLocationID),
		FromEmployeeID: toPgUUIDPtr(transfer.FromEmployeeID),
		ToEmployeeID:   toPgUUIDPtr(transfer.ToEmployeeID),
		Reason:         toPgText(transfer.Reason),
		Status:         transfer.Status,
		TransferredBy:  toPgUUIDPtr(transfer.TransferredBy),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset transfer: %w", err)
	}

	return &domain.Transfer{
		ID:             toUUID(row.ID),
		LegalEntityID:  toUUID(row.LegalEntityID),
		AssetID:        toUUID(row.AssetID),
		TransferNumber: row.TransferNumber,
		TransferDate:   fromPgDate(row.TransferDate),
		FromLocationID: toUUIDPtr(row.FromLocationID),
		ToLocationID:   toUUID(row.ToLocationID),
		FromEmployeeID: toUUIDPtr(row.FromEmployeeID),
		ToEmployeeID:   toUUIDPtr(row.ToEmployeeID),
		Reason:         fromPgText(row.Reason),
		Status:         row.Status,
		TransferredBy:  toUUIDPtr(row.TransferredBy),
		CreatedAt:      fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) ListAssetTransfers(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Transfer, error) {
	rows, err := r.queries(ctx).ListAssetTransfers(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, fmt.Errorf("failed to list asset transfers: %w", err)
	}

	result := make([]domain.Transfer, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.Transfer{
			ID:               toUUID(row.ID),
			LegalEntityID:    toUUID(row.LegalEntityID),
			AssetID:          toUUID(row.AssetID),
			TransferNumber:   row.TransferNumber,
			TransferDate:     fromPgDate(row.TransferDate),
			FromLocationID:   toUUIDPtr(row.FromLocationID),
			ToLocationID:     toUUID(row.ToLocationID),
			FromEmployeeID:   toUUIDPtr(row.FromEmployeeID),
			ToEmployeeID:     toUUIDPtr(row.ToEmployeeID),
			Reason:           fromPgText(row.Reason),
			Status:           row.Status,
			TransferredBy:    toUUIDPtr(row.TransferredBy),
			CreatedAt:        fromPgTimestamptz(row.CreatedAt),
			FromLocationCode: fromPgText(row.FromLocationCode),
			FromLocationName: fromPgText(row.FromLocationNameEn),
			ToLocationCode:   row.ToLocationCode,
			ToLocationName:   row.ToLocationNameEn,
		})
	}
	return result, nil
}

// ----------------------------------------------------------------------
// 8. Impairments
// ----------------------------------------------------------------------

func (r *Repository) CreateAssetImpairment(ctx context.Context, impairment domain.Impairment) (*domain.Impairment, error) {
	if impairment.ID == uuid.Nil {
		impairment.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateAssetImpairment(ctx, sqlc.CreateAssetImpairmentParams{
		ID:                   toPgUUID(impairment.ID),
		LegalEntityID:        toPgUUID(impairment.LegalEntityID),
		AssetID:              toPgUUID(impairment.AssetID),
		ImpairmentNumber:     impairment.ImpairmentNumber,
		ImpairmentDate:       toPgDate(impairment.ImpairmentDate),
		CarryingAmountBefore: toPgNumeric(impairment.CarryingAmountBefore),
		RecoverableAmount:    toPgNumeric(impairment.RecoverableAmount),
		ImpairmentLoss:       toPgNumeric(impairment.ImpairmentLoss),
		Reason:               impairment.Reason,
		JournalEntryID:       toPgUUIDPtr(impairment.JournalEntryID),
		CreatedBy:            toPgUUIDPtr(impairment.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset impairment: %w", err)
	}

	return &domain.Impairment{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		AssetID:              toUUID(row.AssetID),
		ImpairmentNumber:     row.ImpairmentNumber,
		ImpairmentDate:       fromPgDate(row.ImpairmentDate),
		CarryingAmountBefore: fromPgNumeric(row.CarryingAmountBefore),
		RecoverableAmount:    fromPgNumeric(row.RecoverableAmount),
		ImpairmentLoss:       fromPgNumeric(row.ImpairmentLoss),
		Reason:               row.Reason,
		JournalEntryID:       toUUIDPtr(row.JournalEntryID),
		CreatedBy:            toUUIDPtr(row.CreatedBy),
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
	}, nil
}

// ----------------------------------------------------------------------
// 9. Disposals
// ----------------------------------------------------------------------

func (r *Repository) CreateAssetDisposal(ctx context.Context, disposal domain.Disposal) (*domain.Disposal, error) {
	if disposal.ID == uuid.Nil {
		disposal.ID = uuid.New()
	}

	row, err := r.queries(ctx).CreateAssetDisposal(ctx, sqlc.CreateAssetDisposalParams{
		ID:                                toPgUUID(disposal.ID),
		LegalEntityID:                     toPgUUID(disposal.LegalEntityID),
		AssetID:                           toPgUUID(disposal.AssetID),
		DisposalNumber:                    disposal.DisposalNumber,
		DisposalDate:                      toPgDate(disposal.DisposalDate),
		DisposalType:                      disposal.DisposalType,
		ProceedsAmount:                    toPgNumeric(disposal.ProceedsAmount),
		CostAtDisposal:                    toPgNumeric(disposal.CostAtDisposal),
		AccumulatedDepreciationAtDisposal: toPgNumeric(disposal.AccumulatedDepreciationAtDisposal),
		NetBookValueAtDisposal:            toPgNumeric(disposal.NetBookValueAtDisposal),
		GainOrLossAmount:                  toPgNumeric(disposal.GainOrLossAmount),
		JournalEntryID:                    toPgUUIDPtr(disposal.JournalEntryID),
		CustomerID:                        toPgUUIDPtr(disposal.CustomerID),
		Reason:                            toPgText(disposal.Reason),
		Status:                            disposal.Status,
		CreatedBy:                         toPgUUIDPtr(disposal.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset disposal: %w", err)
	}

	return &domain.Disposal{
		ID:                                toUUID(row.ID),
		LegalEntityID:                     toUUID(row.LegalEntityID),
		AssetID:                           toUUID(row.AssetID),
		DisposalNumber:                    row.DisposalNumber,
		DisposalDate:                      fromPgDate(row.DisposalDate),
		DisposalType:                      row.DisposalType,
		ProceedsAmount:                    fromPgNumeric(row.ProceedsAmount),
		CostAtDisposal:                    fromPgNumeric(row.CostAtDisposal),
		AccumulatedDepreciationAtDisposal: fromPgNumeric(row.AccumulatedDepreciationAtDisposal),
		NetBookValueAtDisposal:            fromPgNumeric(row.NetBookValueAtDisposal),
		GainOrLossAmount:                  fromPgNumeric(row.GainOrLossAmount),
		JournalEntryID:                    toUUIDPtr(row.JournalEntryID),
		CustomerID:                        toUUIDPtr(row.CustomerID),
		Reason:                            fromPgText(row.Reason),
		Status:                            row.Status,
		CreatedBy:                         toUUIDPtr(row.CreatedBy),
		CreatedAt:                         fromPgTimestamptz(row.CreatedAt),
	}, nil
}

func (r *Repository) GetAssetDisposalByID(ctx context.Context, id uuid.UUID) (*domain.Disposal, error) {
	row, err := r.queries(ctx).GetAssetDisposalByID(ctx, toPgUUID(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset disposal by id: %w", err)
	}

	return &domain.Disposal{
		ID:                                toUUID(row.ID),
		LegalEntityID:                     toUUID(row.LegalEntityID),
		AssetID:                           toUUID(row.AssetID),
		DisposalNumber:                    row.DisposalNumber,
		DisposalDate:                      fromPgDate(row.DisposalDate),
		DisposalType:                      row.DisposalType,
		ProceedsAmount:                    fromPgNumeric(row.ProceedsAmount),
		CostAtDisposal:                    fromPgNumeric(row.CostAtDisposal),
		AccumulatedDepreciationAtDisposal: fromPgNumeric(row.AccumulatedDepreciationAtDisposal),
		NetBookValueAtDisposal:            fromPgNumeric(row.NetBookValueAtDisposal),
		GainOrLossAmount:                  fromPgNumeric(row.GainOrLossAmount),
		JournalEntryID:                    toUUIDPtr(row.JournalEntryID),
		CustomerID:                        toUUIDPtr(row.CustomerID),
		Reason:                            fromPgText(row.Reason),
		Status:                            row.Status,
		CreatedBy:                         toUUIDPtr(row.CreatedBy),
		CreatedAt:                         fromPgTimestamptz(row.CreatedAt),
	}, nil
}
