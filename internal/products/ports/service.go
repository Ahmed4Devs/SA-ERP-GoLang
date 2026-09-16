package ports

import (
	"context"

	"erp/internal/products/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateItemRequest struct {
	LegalEntityID     uuid.UUID           `json:"legal_entity_id"`
	CategoryID        *uuid.UUID          `json:"category_id,omitempty"`
	Code              string              `json:"code"`
	NameEn            string              `json:"name_en"`
	NameAr            string              `json:"name_ar"`
	DescriptionEn     string              `json:"description_en,omitempty"`
	DescriptionAr     string              `json:"description_ar,omitempty"`
	ItemType          domain.ItemType     `json:"item_type"`
	TrackingType      domain.TrackingType `json:"tracking_type"`
	BaseUOMID         uuid.UUID           `json:"base_uom_id"`
	PurchaseUOMID     *uuid.UUID          `json:"purchase_uom_id,omitempty"`
	SalesUOMID        *uuid.UUID          `json:"sales_uom_id,omitempty"`
	DefaultTaxCodeID  *uuid.UUID          `json:"default_tax_code_id,omitempty"`
	CostPrice         decimal.Decimal     `json:"cost_price"`
	ListPrice         decimal.Decimal     `json:"list_price"`
	IsSalesEnabled    bool                `json:"is_sales_enabled"`
	IsPurchaseEnabled bool                `json:"is_purchase_enabled"`
	DefaultBarcode    string              `json:"default_barcode,omitempty"`
}

type ProductService interface {
	// Units of Measure
	ListUOMs(ctx context.Context) ([]domain.UnitOfMeasure, error)
	CreateUOM(ctx context.Context, uom domain.UnitOfMeasure) (*domain.UnitOfMeasure, error)
	ConvertQuantity(ctx context.Context, fromUOMID, toUOMID uuid.UUID, quantity decimal.Decimal) (decimal.Decimal, error)

	// Categories
	ListCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Category, error)
	CreateCategory(ctx context.Context, cat domain.Category) (*domain.Category, error)

	// Items Master
	ListItems(ctx context.Context, legalEntityID uuid.UUID, itemType *string, categoryID *uuid.UUID, limit, offset int32) ([]domain.Item, error)
	GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	GetItemByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Item, error)
	CreateItem(ctx context.Context, req CreateItemRequest) (*domain.Item, error)
	UpdateItem(ctx context.Context, item domain.Item) (*domain.Item, error)
	SearchItems(ctx context.Context, legalEntityID uuid.UUID, query string) ([]domain.Item, error)

	// Variants & SKUs
	CreateVariantWithSKU(ctx context.Context, legalEntityID uuid.UUID, itemID uuid.UUID, variantCode, nameEn, nameAr string, attrs map[string]any, cost, selling decimal.Decimal) (*domain.ItemVariant, *domain.SKU, error)
	ListVariants(ctx context.Context, itemID uuid.UUID) ([]domain.ItemVariant, error)
	ListSKUs(ctx context.Context, itemID uuid.UUID) ([]domain.SKU, error)

	// Barcodes
	LookupBarcode(ctx context.Context, legalEntityID uuid.UUID, barcode string) (*domain.Barcode, error)
	AddBarcode(ctx context.Context, bc domain.Barcode) (*domain.Barcode, error)
	ListBarcodes(ctx context.Context, itemID uuid.UUID) ([]domain.Barcode, error)

	// Item Packaging UOMs
	ListItemUOMs(ctx context.Context, itemID uuid.UUID) ([]domain.ItemUOM, error)
	AddItemUOM(ctx context.Context, iu domain.ItemUOM) (*domain.ItemUOM, error)

	// Item Tax Profiles
	GetItemTaxProfile(ctx context.Context, itemID uuid.UUID) (*domain.ItemTaxProfile, error)
	SetItemTaxProfile(ctx context.Context, itp domain.ItemTaxProfile) (*domain.ItemTaxProfile, error)
}
