package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/products/domain"
	"erp/internal/products/ports"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type productRepository struct {
	db *database.DB
}

func NewProductRepository(db *database.DB) ports.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// 1. Units of Measure
func (r *productRepository) ListUnitsOfMeasure(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	rows, err := r.queries(ctx).ListUnitsOfMeasure(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.UnitOfMeasure, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.UnitOfMeasure{
			ID:        toUUID(row.ID),
			Code:      row.Code,
			NameEn:    row.NameEn,
			NameAr:    row.NameAr,
			SymbolEn:  fromPgText(row.SymbolEn),
			SymbolAr:  fromPgText(row.SymbolAr),
			UOMType:   row.UomType,
			IsActive:  row.IsActive,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return res, nil
}

func (r *productRepository) GetUnitOfMeasureByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error) {
	row, err := r.queries(ctx).GetUnitOfMeasureByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUOMNotFound
		}
		return nil, err
	}
	return &domain.UnitOfMeasure{
		ID:        toUUID(row.ID),
		Code:      row.Code,
		NameEn:    row.NameEn,
		NameAr:    row.NameAr,
		SymbolEn:  fromPgText(row.SymbolEn),
		SymbolAr:  fromPgText(row.SymbolAr),
		UOMType:   row.UomType,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *productRepository) GetUnitOfMeasureByCode(ctx context.Context, code string) (*domain.UnitOfMeasure, error) {
	row, err := r.queries(ctx).GetUnitOfMeasureByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUOMNotFound
		}
		return nil, err
	}
	return &domain.UnitOfMeasure{
		ID:        toUUID(row.ID),
		Code:      row.Code,
		NameEn:    row.NameEn,
		NameAr:    row.NameAr,
		SymbolEn:  fromPgText(row.SymbolEn),
		SymbolAr:  fromPgText(row.SymbolAr),
		UOMType:   row.UomType,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *productRepository) CreateUnitOfMeasure(ctx context.Context, uom *domain.UnitOfMeasure) (*domain.UnitOfMeasure, error) {
	row, err := r.queries(ctx).CreateUnitOfMeasure(ctx, sqlc.CreateUnitOfMeasureParams{
		Code:     uom.Code,
		NameEn:   uom.NameEn,
		NameAr:   uom.NameAr,
		SymbolEn: toPgText(uom.SymbolEn),
		SymbolAr: toPgText(uom.SymbolAr),
		UomType:  uom.UOMType,
	})
	if err != nil {
		return nil, err
	}
	uom.ID = toUUID(row.ID)
	uom.CreatedAt = row.CreatedAt.Time
	uom.IsActive = row.IsActive
	return uom, nil
}

// 2. UOM Conversions
func (r *productRepository) ListUOMConversions(ctx context.Context) ([]domain.UOMConversion, error) {
	rows, err := r.queries(ctx).ListUOMConversions(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.UOMConversion, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.UOMConversion{
			ID:             toUUID(row.ID),
			FromUOMID:      toUUID(row.FromUomID),
			ToUOMID:        toUUID(row.ToUomID),
			ConversionRate: fromPgNumeric(row.ConversionRate),
			FromUOMCode:    row.FromUomCode,
			FromUOMNameAr:  row.FromUomNameAr,
			ToUOMCode:      row.ToUomCode,
			ToUOMNameAr:    row.ToUomNameAr,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return res, nil
}

func (r *productRepository) GetUOMConversionRate(ctx context.Context, fromUOMID, toUOMID uuid.UUID) (decimal.Decimal, error) {
	if fromUOMID == toUOMID {
		return decimal.NewFromInt(1), nil
	}
	rate, err := r.queries(ctx).GetUOMConversionRate(ctx, sqlc.GetUOMConversionRateParams{
		FromUomID: toPgUUID(fromUOMID),
		ToUomID:   toPgUUID(toUOMID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, domain.ErrInvalidUOMConversion
		}
		return decimal.Zero, err
	}
	return fromPgNumeric(rate), nil
}

func (r *productRepository) CreateUOMConversion(ctx context.Context, conv *domain.UOMConversion) (*domain.UOMConversion, error) {
	row, err := r.queries(ctx).CreateUOMConversion(ctx, sqlc.CreateUOMConversionParams{
		FromUomID:      toPgUUID(conv.FromUOMID),
		ToUomID:        toPgUUID(conv.ToUOMID),
		ConversionRate: toPgNumeric(conv.ConversionRate),
	})
	if err != nil {
		return nil, err
	}
	conv.ID = toUUID(row.ID)
	conv.CreatedAt = row.CreatedAt.Time
	return conv, nil
}

// 3. Categories
func (r *productRepository) ListCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Category, error) {
	rows, err := r.queries(ctx).ListCategoriesByEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.Category{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			ParentID:      toUUIDPtr(row.ParentID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			Description:   fromPgText(row.Description),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			ParentNameAr:  fromPgText(row.ParentNameAr),
			ParentNameEn:  fromPgText(row.ParentNameEn),
		})
	}
	return res, nil
}

func (r *productRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	row, err := r.queries(ctx).GetCategoryByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return &domain.Category{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ParentID:      toUUIDPtr(row.ParentID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		Description:   fromPgText(row.Description),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *productRepository) GetCategoryByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Category, error) {
	row, err := r.queries(ctx).GetCategoryByCode(ctx, sqlc.GetCategoryByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return &domain.Category{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ParentID:      toUUIDPtr(row.ParentID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		Description:   fromPgText(row.Description),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *productRepository) CreateCategory(ctx context.Context, cat *domain.Category) (*domain.Category, error) {
	row, err := r.queries(ctx).CreateCategory(ctx, sqlc.CreateCategoryParams{
		LegalEntityID: toPgUUID(cat.LegalEntityID),
		ParentID:      toPgUUIDPtr(cat.ParentID),
		Code:          cat.Code,
		NameEn:        cat.NameEn,
		NameAr:        cat.NameAr,
		Description:   toPgText(cat.Description),
	})
	if err != nil {
		return nil, err
	}
	cat.ID = toUUID(row.ID)
	cat.CreatedAt = row.CreatedAt.Time
	cat.UpdatedAt = row.UpdatedAt.Time
	cat.IsActive = row.IsActive
	return cat, nil
}

func (r *productRepository) UpdateCategory(ctx context.Context, cat *domain.Category) (*domain.Category, error) {
	row, err := r.queries(ctx).UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:          toPgUUID(cat.ID),
		ParentID:    toPgUUIDPtr(cat.ParentID),
		NameEn:      cat.NameEn,
		NameAr:      cat.NameAr,
		Description: toPgText(cat.Description),
		IsActive:    cat.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	cat.UpdatedAt = row.UpdatedAt.Time
	return cat, nil
}

// 4. Items Master
func (r *productRepository) ListItems(ctx context.Context, legalEntityID uuid.UUID, itemType *string, categoryID *uuid.UUID, limit, offset int32) ([]domain.Item, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries(ctx).ListItemsByEntity(ctx, sqlc.ListItemsByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		ItemType:      toPgTextPtr(itemType),
		CategoryID:    toPgUUIDPtr(categoryID),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.Item{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			CategoryID:        toUUIDPtr(row.CategoryID),
			Code:              row.Code,
			NameEn:            row.NameEn,
			NameAr:            row.NameAr,
			DescriptionEn:     fromPgText(row.DescriptionEn),
			DescriptionAr:     fromPgText(row.DescriptionAr),
			ItemType:          domain.ItemType(row.ItemType),
			TrackingType:      domain.TrackingType(row.TrackingType),
			BaseUOMID:         toUUID(row.BaseUomID),
			PurchaseUOMID:     toUUIDPtr(row.PurchaseUomID),
			SalesUOMID:        toUUIDPtr(row.SalesUomID),
			DefaultTaxCodeID:  toUUIDPtr(row.DefaultTaxCodeID),
			CostPrice:         fromPgNumeric(row.CostPrice),
			ListPrice:         fromPgNumeric(row.ListPrice),
			IsSalesEnabled:    row.IsSalesEnabled,
			IsPurchaseEnabled: row.IsPurchaseEnabled,
			IsActive:          row.IsActive,
			CreatedAt:         row.CreatedAt.Time,
			UpdatedAt:         row.UpdatedAt.Time,
			Version:           row.Version,
			CategoryNameAr:    fromPgText(row.CategoryNameAr),
			CategoryNameEn:    fromPgText(row.CategoryNameEn),
			BaseUOMCode:       row.BaseUomCode,
			BaseUOMNameAr:     row.BaseUomNameAr,
			TaxCode:           fromPgText(row.TaxCode),
			TaxNameAr:         fromPgText(row.TaxNameAr),
		})
	}
	return res, nil
}

func (r *productRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	row, err := r.queries(ctx).GetItemByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	return mapGetItemByIDToDomain(&row), nil
}

func (r *productRepository) GetItemByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Item, error) {
	row, err := r.queries(ctx).GetItemByCode(ctx, sqlc.GetItemByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	return mapGetItemByCodeToDomain(&row), nil
}

func (r *productRepository) CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	row, err := r.queries(ctx).CreateItem(ctx, sqlc.CreateItemParams{
		LegalEntityID:     toPgUUID(item.LegalEntityID),
		CategoryID:        toPgUUIDPtr(item.CategoryID),
		Code:              item.Code,
		NameEn:            item.NameEn,
		NameAr:            item.NameAr,
		DescriptionEn:     toPgText(item.DescriptionEn),
		DescriptionAr:     toPgText(item.DescriptionAr),
		ItemType:          string(item.ItemType),
		TrackingType:      string(item.TrackingType),
		BaseUomID:         toPgUUID(item.BaseUOMID),
		PurchaseUomID:     toPgUUIDPtr(item.PurchaseUOMID),
		SalesUomID:        toPgUUIDPtr(item.SalesUOMID),
		DefaultTaxCodeID:  toPgUUIDPtr(item.DefaultTaxCodeID),
		CostPrice:         toPgNumeric(item.CostPrice),
		ListPrice:         toPgNumeric(item.ListPrice),
		IsSalesEnabled:    item.IsSalesEnabled,
		IsPurchaseEnabled: item.IsPurchaseEnabled,
		IsActive:          item.IsActive,
	})
	if err != nil {
		return nil, err
	}
	item.ID = toUUID(row.ID)
	item.CreatedAt = row.CreatedAt.Time
	item.UpdatedAt = row.UpdatedAt.Time
	item.Version = row.Version
	return item, nil
}

func (r *productRepository) UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	row, err := r.queries(ctx).UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:                toPgUUID(item.ID),
		CategoryID:        toPgUUIDPtr(item.CategoryID),
		NameEn:            item.NameEn,
		NameAr:            item.NameAr,
		DescriptionEn:     toPgText(item.DescriptionEn),
		DescriptionAr:     toPgText(item.DescriptionAr),
		ItemType:          string(item.ItemType),
		TrackingType:      string(item.TrackingType),
		PurchaseUomID:     toPgUUIDPtr(item.PurchaseUOMID),
		SalesUomID:        toPgUUIDPtr(item.SalesUOMID),
		DefaultTaxCodeID:  toPgUUIDPtr(item.DefaultTaxCodeID),
		CostPrice:         toPgNumeric(item.CostPrice),
		ListPrice:         toPgNumeric(item.ListPrice),
		IsSalesEnabled:    item.IsSalesEnabled,
		IsPurchaseEnabled: item.IsPurchaseEnabled,
		IsActive:          item.IsActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, err
	}
	item.UpdatedAt = row.UpdatedAt.Time
	item.Version = row.Version
	return item, nil
}

func (r *productRepository) SearchItems(ctx context.Context, legalEntityID uuid.UUID, query string) ([]domain.Item, error) {
	rows, err := r.queries(ctx).SearchItems(ctx, sqlc.SearchItemsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Query:         toPgText(query),
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.Item, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.Item{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			ItemType:      domain.ItemType(row.ItemType),
			CostPrice:     fromPgNumeric(row.CostPrice),
			ListPrice:     fromPgNumeric(row.ListPrice),
			BaseUOMCode:   row.BaseUomCode,
		})
	}
	return res, nil
}

// 5. Item Variants
func (r *productRepository) ListVariants(ctx context.Context, itemID uuid.UUID) ([]domain.ItemVariant, error) {
	rows, err := r.queries(ctx).ListVariantsByItem(ctx, toPgUUID(itemID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.ItemVariant, 0, len(rows))
	for _, row := range rows {
		var attrs map[string]any
		if err := json.Unmarshal(row.Attributes, &attrs); err != nil {
			return nil, fmt.Errorf("failed to decode attributes for variant %s: %w", toUUID(row.ID), err)
		}
		res = append(res, domain.ItemVariant{
			ID:          toUUID(row.ID),
			ItemID:      toUUID(row.ItemID),
			VariantCode: row.VariantCode,
			NameEn:      row.NameEn,
			NameAr:      row.NameAr,
			Attributes:  attrs,
			IsActive:    row.IsActive,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return res, nil
}

func (r *productRepository) GetVariantByID(ctx context.Context, id uuid.UUID) (*domain.ItemVariant, error) {
	row, err := r.queries(ctx).GetVariantByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVariantNotFound
		}
		return nil, err
	}
	var attrs map[string]any
	if err := json.Unmarshal(row.Attributes, &attrs); err != nil {
		return nil, fmt.Errorf("failed to decode attributes for variant %s: %w", id, err)
	}
	return &domain.ItemVariant{
		ID:          toUUID(row.ID),
		ItemID:      toUUID(row.ItemID),
		VariantCode: row.VariantCode,
		NameEn:      row.NameEn,
		NameAr:      row.NameAr,
		Attributes:  attrs,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *productRepository) CreateVariant(ctx context.Context, variant *domain.ItemVariant) (*domain.ItemVariant, error) {
	attrsBytes, err := json.Marshal(variant.Attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to encode variant attributes: %w", err)
	}

	row, err := r.queries(ctx).CreateVariant(ctx, sqlc.CreateVariantParams{
		ItemID:      toPgUUID(variant.ItemID),
		VariantCode: variant.VariantCode,
		NameEn:      variant.NameEn,
		NameAr:      variant.NameAr,
		Attributes:  attrsBytes,
		IsActive:    variant.IsActive,
	})
	if err != nil {
		return nil, err
	}
	variant.ID = toUUID(row.ID)
	variant.CreatedAt = row.CreatedAt.Time
	variant.UpdatedAt = row.UpdatedAt.Time
	return variant, nil
}

// 6. SKUs
func (r *productRepository) ListSKUs(ctx context.Context, itemID uuid.UUID) ([]domain.SKU, error) {
	rows, err := r.queries(ctx).ListSKUsByItem(ctx, toPgUUID(itemID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.SKU, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.SKU{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			ItemID:        toUUID(row.ItemID),
			VariantID:     toUUIDPtr(row.VariantID),
			SKUCode:       row.SkuCode,
			CostPrice:     fromPgNumeric(row.CostPrice),
			SellingPrice:  fromPgNumeric(row.SellingPrice),
			IsActive:      row.IsActive,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			VariantCode:   fromPgText(row.VariantCode),
			VariantNameAr: fromPgText(row.VariantNameAr),
		})
	}
	return res, nil
}

func (r *productRepository) GetSKUByID(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	row, err := r.queries(ctx).GetSKUByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSKUNotFound
		}
		return nil, err
	}
	return &domain.SKU{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ItemID:        toUUID(row.ItemID),
		VariantID:     toUUIDPtr(row.VariantID),
		SKUCode:       row.SkuCode,
		CostPrice:     fromPgNumeric(row.CostPrice),
		SellingPrice:  fromPgNumeric(row.SellingPrice),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *productRepository) GetSKUByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.SKU, error) {
	row, err := r.queries(ctx).GetSKUByCode(ctx, sqlc.GetSKUByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		SkuCode:       code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSKUNotFound
		}
		return nil, err
	}
	return &domain.SKU{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ItemID:        toUUID(row.ItemID),
		VariantID:     toUUIDPtr(row.VariantID),
		SKUCode:       row.SkuCode,
		CostPrice:     fromPgNumeric(row.CostPrice),
		SellingPrice:  fromPgNumeric(row.SellingPrice),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *productRepository) CreateSKU(ctx context.Context, sku *domain.SKU) (*domain.SKU, error) {
	row, err := r.queries(ctx).CreateSKU(ctx, sqlc.CreateSKUParams{
		LegalEntityID: toPgUUID(sku.LegalEntityID),
		ItemID:        toPgUUID(sku.ItemID),
		VariantID:     toPgUUIDPtr(sku.VariantID),
		SkuCode:       sku.SKUCode,
		CostPrice:     toPgNumeric(sku.CostPrice),
		SellingPrice:  toPgNumeric(sku.SellingPrice),
		IsActive:      sku.IsActive,
	})
	if err != nil {
		return nil, err
	}
	sku.ID = toUUID(row.ID)
	sku.CreatedAt = row.CreatedAt.Time
	sku.UpdatedAt = row.UpdatedAt.Time
	return sku, nil
}

func (r *productRepository) UpdateSKUPrices(ctx context.Context, id uuid.UUID, cost, selling decimal.Decimal, isActive bool) (*domain.SKU, error) {
	row, err := r.queries(ctx).UpdateSKUPrices(ctx, sqlc.UpdateSKUPricesParams{
		ID:           toPgUUID(id),
		CostPrice:    toPgNumeric(cost),
		SellingPrice: toPgNumeric(selling),
		IsActive:     isActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSKUNotFound
		}
		return nil, err
	}
	return &domain.SKU{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ItemID:        toUUID(row.ItemID),
		VariantID:     toUUIDPtr(row.VariantID),
		SKUCode:       row.SkuCode,
		CostPrice:     fromPgNumeric(row.CostPrice),
		SellingPrice:  fromPgNumeric(row.SellingPrice),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

// 7. Barcodes
func (r *productRepository) ListBarcodes(ctx context.Context, itemID uuid.UUID) ([]domain.Barcode, error) {
	rows, err := r.queries(ctx).ListBarcodesByItem(ctx, toPgUUID(itemID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.Barcode, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.Barcode{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			ItemID:        toUUID(row.ItemID),
			SKUID:         toUUIDPtr(row.SkuID),
			UOMID:         toUUIDPtr(row.UomID),
			Barcode:       row.Barcode,
			BarcodeFormat: row.BarcodeFormat,
			IsPrimary:     row.IsPrimary,
			CreatedAt:     row.CreatedAt.Time,
			UOMCode:       fromPgText(row.UomCode),
			UOMNameAr:     fromPgText(row.UomNameAr),
			SKUCode:       fromPgText(row.SkuCode),
		})
	}
	return res, nil
}

func (r *productRepository) GetBarcode(ctx context.Context, legalEntityID uuid.UUID, barcode string) (*domain.Barcode, error) {
	row, err := r.queries(ctx).GetBarcodeByCode(ctx, sqlc.GetBarcodeByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Barcode:       barcode,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBarcodeNotFound
		}
		return nil, err
	}
	return &domain.Barcode{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		ItemID:        toUUID(row.ItemID),
		SKUID:         toUUIDPtr(row.SkuID),
		UOMID:         toUUIDPtr(row.UomID),
		Barcode:       row.Barcode,
		BarcodeFormat: row.BarcodeFormat,
		IsPrimary:     row.IsPrimary,
		CreatedAt:     row.CreatedAt.Time,
		ItemCode:      row.ItemCode,
		ItemNameAr:    row.ItemNameAr,
		ItemNameEn:    row.ItemNameEn,
		ItemType:      domain.ItemType(row.ItemType),
		ListPrice:     fromPgNumeric(row.ListPrice),
		CostPrice:     fromPgNumeric(row.CostPrice),
		UOMCode:       fromPgText(row.UomCode),
		UOMNameAr:     fromPgText(row.UomNameAr),
		SKUCode:       fromPgText(row.SkuCode),
	}, nil
}

func (r *productRepository) CreateBarcode(ctx context.Context, bc *domain.Barcode) (*domain.Barcode, error) {
	row, err := r.queries(ctx).CreateBarcode(ctx, sqlc.CreateBarcodeParams{
		LegalEntityID: toPgUUID(bc.LegalEntityID),
		ItemID:        toPgUUID(bc.ItemID),
		SkuID:         toPgUUIDPtr(bc.SKUID),
		UomID:         toPgUUIDPtr(bc.UOMID),
		Barcode:       bc.Barcode,
		BarcodeFormat: bc.BarcodeFormat,
		IsPrimary:     bc.IsPrimary,
	})
	if err != nil {
		return nil, err
	}
	bc.ID = toUUID(row.ID)
	bc.CreatedAt = row.CreatedAt.Time
	return bc, nil
}

func (r *productRepository) DeleteBarcode(ctx context.Context, id uuid.UUID) error {
	return r.queries(ctx).DeleteBarcode(ctx, toPgUUID(id))
}

// 8. Item UOMs
func (r *productRepository) ListItemUOMs(ctx context.Context, itemID uuid.UUID) ([]domain.ItemUOM, error) {
	rows, err := r.queries(ctx).ListItemUOMsByItem(ctx, toPgUUID(itemID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.ItemUOM, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ItemUOM{
			ID:             toUUID(row.ID),
			ItemID:         toUUID(row.ItemID),
			UOMID:          toUUID(row.UomID),
			ConversionRate: fromPgNumeric(row.ConversionRate),
			IsBaseUOM:      row.IsBaseUom,
			IsSalesUOM:     row.IsSalesUom,
			IsPurchaseUOM:  row.IsPurchaseUom,
			CreatedAt:      row.CreatedAt.Time,
			UOMCode:        row.UomCode,
			UOMNameAr:      row.UomNameAr,
			UOMNameEn:      row.UomNameEn,
			UOMSymbolAr:    fromPgText(row.UomSymbolAr),
		})
	}
	return res, nil
}

func (r *productRepository) UpsertItemUOM(ctx context.Context, iu *domain.ItemUOM) (*domain.ItemUOM, error) {
	row, err := r.queries(ctx).UpsertItemUOM(ctx, sqlc.UpsertItemUOMParams{
		ItemID:         toPgUUID(iu.ItemID),
		UomID:          toPgUUID(iu.UOMID),
		ConversionRate: toPgNumeric(iu.ConversionRate),
		IsBaseUom:      iu.IsBaseUOM,
		IsSalesUom:     iu.IsSalesUOM,
		IsPurchaseUom:  iu.IsPurchaseUOM,
	})
	if err != nil {
		return nil, err
	}
	iu.ID = toUUID(row.ID)
	iu.CreatedAt = row.CreatedAt.Time
	return iu, nil
}

// 9. Item Tax Profiles
func (r *productRepository) GetItemTaxProfile(ctx context.Context, itemID uuid.UUID) (*domain.ItemTaxProfile, error) {
	row, err := r.queries(ctx).GetItemTaxProfile(ctx, toPgUUID(itemID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.ItemTaxProfile{
		ID:                  toUUID(row.ID),
		ItemID:              toUUID(row.ItemID),
		TaxCodeID:           toUUID(row.TaxCodeID),
		ExemptionReasonCode: fromPgTextPtr(row.ExemptionReasonCode),
		ZeroRateReasonCode:  fromPgTextPtr(row.ZeroRateReasonCode),
		IsActive:            row.IsActive,
		CreatedAt:           row.CreatedAt.Time,
		TaxCode:             row.TaxCode,
		TaxNameAr:           row.TaxNameAr,
		TaxCategoryCode:     row.TaxCategoryCode,
	}, nil
}

func (r *productRepository) UpsertItemTaxProfile(ctx context.Context, itp *domain.ItemTaxProfile) (*domain.ItemTaxProfile, error) {
	row, err := r.queries(ctx).UpsertItemTaxProfile(ctx, sqlc.UpsertItemTaxProfileParams{
		ItemID:              toPgUUID(itp.ItemID),
		TaxCodeID:           toPgUUID(itp.TaxCodeID),
		ExemptionReasonCode: toPgTextPtr(itp.ExemptionReasonCode),
		ZeroRateReasonCode:  toPgTextPtr(itp.ZeroRateReasonCode),
		IsActive:            itp.IsActive,
	})
	if err != nil {
		return nil, err
	}
	itp.ID = toUUID(row.ID)
	itp.CreatedAt = row.CreatedAt.Time
	return itp, nil
}

func mapGetItemByIDToDomain(row *sqlc.GetItemByIDRow) *domain.Item {
	return &domain.Item{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		CategoryID:        toUUIDPtr(row.CategoryID),
		Code:              row.Code,
		NameEn:            row.NameEn,
		NameAr:            row.NameAr,
		DescriptionEn:     fromPgText(row.DescriptionEn),
		DescriptionAr:     fromPgText(row.DescriptionAr),
		ItemType:          domain.ItemType(row.ItemType),
		TrackingType:      domain.TrackingType(row.TrackingType),
		BaseUOMID:         toUUID(row.BaseUomID),
		PurchaseUOMID:     toUUIDPtr(row.PurchaseUomID),
		SalesUOMID:        toUUIDPtr(row.SalesUomID),
		DefaultTaxCodeID:  toUUIDPtr(row.DefaultTaxCodeID),
		CostPrice:         fromPgNumeric(row.CostPrice),
		ListPrice:         fromPgNumeric(row.ListPrice),
		IsSalesEnabled:    row.IsSalesEnabled,
		IsPurchaseEnabled: row.IsPurchaseEnabled,
		IsActive:          row.IsActive,
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
		Version:           row.Version,
		CategoryNameAr:    fromPgText(row.CategoryNameAr),
		CategoryNameEn:    fromPgText(row.CategoryNameEn),
		BaseUOMCode:       row.BaseUomCode,
		BaseUOMNameAr:     row.BaseUomNameAr,
		TaxCode:           fromPgText(row.TaxCode),
		TaxNameAr:         fromPgText(row.TaxNameAr),
	}
}

func mapGetItemByCodeToDomain(row *sqlc.GetItemByCodeRow) *domain.Item {
	return &domain.Item{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		CategoryID:        toUUIDPtr(row.CategoryID),
		Code:              row.Code,
		NameEn:            row.NameEn,
		NameAr:            row.NameAr,
		DescriptionEn:     fromPgText(row.DescriptionEn),
		DescriptionAr:     fromPgText(row.DescriptionAr),
		ItemType:          domain.ItemType(row.ItemType),
		TrackingType:      domain.TrackingType(row.TrackingType),
		BaseUOMID:         toUUID(row.BaseUomID),
		PurchaseUOMID:     toUUIDPtr(row.PurchaseUomID),
		SalesUOMID:        toUUIDPtr(row.SalesUomID),
		DefaultTaxCodeID:  toUUIDPtr(row.DefaultTaxCodeID),
		CostPrice:         fromPgNumeric(row.CostPrice),
		ListPrice:         fromPgNumeric(row.ListPrice),
		IsSalesEnabled:    row.IsSalesEnabled,
		IsPurchaseEnabled: row.IsPurchaseEnabled,
		IsActive:          row.IsActive,
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
		Version:           row.Version,
		CategoryNameAr:    fromPgText(row.CategoryNameAr),
		CategoryNameEn:    fromPgText(row.CategoryNameEn),
		BaseUOMCode:       row.BaseUomCode,
		BaseUOMNameAr:     row.BaseUomNameAr,
		TaxCode:           fromPgText(row.TaxCode),
		TaxNameAr:         fromPgText(row.TaxNameAr),
	}
}
