package ports

import (
	"context"

	"erp/internal/products/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductRepository interface {
	// Units of Measure
	ListUnitsOfMeasure(ctx context.Context) ([]domain.UnitOfMeasure, error)
	GetUnitOfMeasureByID(ctx context.Context, id uuid.UUID) (*domain.UnitOfMeasure, error)
	GetUnitOfMeasureByCode(ctx context.Context, code string) (*domain.UnitOfMeasure, error)
	CreateUnitOfMeasure(ctx context.Context, uom *domain.UnitOfMeasure) (*domain.UnitOfMeasure, error)

	// UOM Conversions
	ListUOMConversions(ctx context.Context) ([]domain.UOMConversion, error)
	GetUOMConversionRate(ctx context.Context, fromUOMID, toUOMID uuid.UUID) (decimal.Decimal, error)
	CreateUOMConversion(ctx context.Context, conv *domain.UOMConversion) (*domain.UOMConversion, error)

	// Categories
	ListCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Category, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetCategoryByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Category, error)
	CreateCategory(ctx context.Context, cat *domain.Category) (*domain.Category, error)
	UpdateCategory(ctx context.Context, cat *domain.Category) (*domain.Category, error)

	// Items Master
	ListItems(ctx context.Context, legalEntityID uuid.UUID, itemType *string, categoryID *uuid.UUID, limit, offset int32) ([]domain.Item, error)
	GetItemByID(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	GetItemByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Item, error)
	CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
	SearchItems(ctx context.Context, legalEntityID uuid.UUID, query string) ([]domain.Item, error)

	// Item Variants
	ListVariants(ctx context.Context, itemID uuid.UUID) ([]domain.ItemVariant, error)
	GetVariantByID(ctx context.Context, id uuid.UUID) (*domain.ItemVariant, error)
	CreateVariant(ctx context.Context, variant *domain.ItemVariant) (*domain.ItemVariant, error)

	// SKUs
	ListSKUs(ctx context.Context, itemID uuid.UUID) ([]domain.SKU, error)
	GetSKUByID(ctx context.Context, id uuid.UUID) (*domain.SKU, error)
	GetSKUByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.SKU, error)
	CreateSKU(ctx context.Context, sku *domain.SKU) (*domain.SKU, error)
	UpdateSKUPrices(ctx context.Context, id uuid.UUID, cost, selling decimal.Decimal, isActive bool) (*domain.SKU, error)

	// Barcodes
	ListBarcodes(ctx context.Context, itemID uuid.UUID) ([]domain.Barcode, error)
	GetBarcode(ctx context.Context, legalEntityID uuid.UUID, barcode string) (*domain.Barcode, error)
	CreateBarcode(ctx context.Context, bc *domain.Barcode) (*domain.Barcode, error)
	DeleteBarcode(ctx context.Context, id uuid.UUID) error

	// Item UOMs
	ListItemUOMs(ctx context.Context, itemID uuid.UUID) ([]domain.ItemUOM, error)
	UpsertItemUOM(ctx context.Context, iu *domain.ItemUOM) (*domain.ItemUOM, error)

	// Item Tax Profiles
	GetItemTaxProfile(ctx context.Context, itemID uuid.UUID) (*domain.ItemTaxProfile, error)
	UpsertItemTaxProfile(ctx context.Context, itp *domain.ItemTaxProfile) (*domain.ItemTaxProfile, error)
}
