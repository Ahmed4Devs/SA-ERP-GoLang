package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ItemType string

const (
	ItemTypeStock    ItemType = "STOCK"
	ItemTypeService  ItemType = "SERVICE"
	ItemTypeNonStock ItemType = "NON_STOCK"
	ItemTypeAsset    ItemType = "ASSET"
)

type TrackingType string

const (
	TrackingTypeNone   TrackingType = "NONE"
	TrackingTypeLot    TrackingType = "LOT"
	TrackingTypeSerial TrackingType = "SERIAL"
)

type UnitOfMeasure struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	NameEn    string    `json:"name_en"`
	NameAr    string    `json:"name_ar"`
	SymbolEn  string    `json:"symbol_en,omitempty"`
	SymbolAr  string    `json:"symbol_ar,omitempty"`
	UOMType   string    `json:"uom_type"` // COUNT, WEIGHT, LENGTH, VOLUME, AREA, TIME
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UOMConversion struct {
	ID             uuid.UUID       `json:"id"`
	FromUOMID      uuid.UUID       `json:"from_uom_id"`
	ToUOMID        uuid.UUID       `json:"to_uom_id"`
	ConversionRate decimal.Decimal `json:"conversion_rate"`
	FromUOMCode    string          `json:"from_uom_code,omitempty"`
	FromUOMNameAr  string          `json:"from_uom_name_ar,omitempty"`
	ToUOMCode      string          `json:"to_uom_code,omitempty"`
	ToUOMNameAr    string          `json:"to_uom_name_ar,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Category struct {
	ID            uuid.UUID  `json:"id"`
	LegalEntityID uuid.UUID  `json:"legal_entity_id"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	Code          string     `json:"code"`
	NameEn        string     `json:"name_en"`
	NameAr        string     `json:"name_ar"`
	Description   string     `json:"description,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ParentNameAr  string     `json:"parent_name_ar,omitempty"`
	ParentNameEn  string     `json:"parent_name_en,omitempty"`
}

type Item struct {
	ID                uuid.UUID       `json:"id"`
	LegalEntityID     uuid.UUID       `json:"legal_entity_id"`
	CategoryID        *uuid.UUID      `json:"category_id,omitempty"`
	Code              string          `json:"code"`
	NameEn            string          `json:"name_en"`
	NameAr            string          `json:"name_ar"`
	DescriptionEn     string          `json:"description_en,omitempty"`
	DescriptionAr     string          `json:"description_ar,omitempty"`
	ItemType          ItemType        `json:"item_type"`
	TrackingType      TrackingType    `json:"tracking_type"`
	BaseUOMID         uuid.UUID       `json:"base_uom_id"`
	PurchaseUOMID     *uuid.UUID      `json:"purchase_uom_id,omitempty"`
	SalesUOMID        *uuid.UUID      `json:"sales_uom_id,omitempty"`
	DefaultTaxCodeID  *uuid.UUID      `json:"default_tax_code_id,omitempty"`
	CostPrice         decimal.Decimal `json:"cost_price"`
	ListPrice         decimal.Decimal `json:"list_price"`
	IsSalesEnabled    bool            `json:"is_sales_enabled"`
	IsPurchaseEnabled bool            `json:"is_purchase_enabled"`
	IsActive          bool            `json:"is_active"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Version           int64           `json:"version"`
	CategoryNameAr    string          `json:"category_name_ar,omitempty"`
	CategoryNameEn    string          `json:"category_name_en,omitempty"`
	BaseUOMCode       string          `json:"base_uom_code,omitempty"`
	BaseUOMNameAr     string          `json:"base_uom_name_ar,omitempty"`
	TaxCode           string          `json:"tax_code,omitempty"`
	TaxNameAr         string          `json:"tax_name_ar,omitempty"`
}

type ItemVariant struct {
	ID          uuid.UUID      `json:"id"`
	ItemID      uuid.UUID      `json:"item_id"`
	VariantCode string         `json:"variant_code"`
	NameEn      string         `json:"name_en"`
	NameAr      string         `json:"name_ar"`
	Attributes  map[string]any `json:"attributes"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type SKU struct {
	ID            uuid.UUID       `json:"id"`
	LegalEntityID uuid.UUID       `json:"legal_entity_id"`
	ItemID        uuid.UUID       `json:"item_id"`
	VariantID     *uuid.UUID      `json:"variant_id,omitempty"`
	SKUCode       string          `json:"sku_code"`
	CostPrice     decimal.Decimal `json:"cost_price"`
	SellingPrice  decimal.Decimal `json:"selling_price"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	VariantCode   string          `json:"variant_code,omitempty"`
	VariantNameAr string          `json:"variant_name_ar,omitempty"`
}

type Barcode struct {
	ID            uuid.UUID       `json:"id"`
	LegalEntityID uuid.UUID       `json:"legal_entity_id"`
	ItemID        uuid.UUID       `json:"item_id"`
	SKUID         *uuid.UUID      `json:"sku_id,omitempty"`
	UOMID         *uuid.UUID      `json:"uom_id,omitempty"`
	Barcode       string          `json:"barcode"`
	BarcodeFormat string          `json:"barcode_format"` // EAN13, EAN8, UPCA, CODE128, QR, CUSTOM
	IsPrimary     bool            `json:"is_primary"`
	CreatedAt     time.Time       `json:"created_at"`
	ItemCode      string          `json:"item_code,omitempty"`
	ItemNameAr    string          `json:"item_name_ar,omitempty"`
	ItemNameEn    string          `json:"item_name_en,omitempty"`
	ItemType      ItemType        `json:"item_type,omitempty"`
	ListPrice     decimal.Decimal `json:"list_price,omitempty"`
	CostPrice     decimal.Decimal `json:"cost_price,omitempty"`
	UOMCode       string          `json:"uom_code,omitempty"`
	UOMNameAr     string          `json:"uom_name_ar,omitempty"`
	SKUCode       string          `json:"sku_code,omitempty"`
}

type ItemUOM struct {
	ID             uuid.UUID       `json:"id"`
	ItemID         uuid.UUID       `json:"item_id"`
	UOMID          uuid.UUID       `json:"uom_id"`
	ConversionRate decimal.Decimal `json:"conversion_rate"`
	IsBaseUOM      bool            `json:"is_base_uom"`
	IsSalesUOM     bool            `json:"is_sales_uom"`
	IsPurchaseUOM  bool            `json:"is_purchase_uom"`
	CreatedAt      time.Time       `json:"created_at"`
	UOMCode        string          `json:"uom_code,omitempty"`
	UOMNameAr      string          `json:"uom_name_ar,omitempty"`
	UOMNameEn      string          `json:"uom_name_en,omitempty"`
	UOMSymbolAr    string          `json:"uom_symbol_ar,omitempty"`
}

type ItemTaxProfile struct {
	ID                  uuid.UUID `json:"id"`
	ItemID              uuid.UUID `json:"item_id"`
	TaxCodeID           uuid.UUID `json:"tax_code_id"`
	ExemptionReasonCode *string   `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string   `json:"zero_rate_reason_code,omitempty"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	TaxCode             string    `json:"tax_code,omitempty"`
	TaxNameAr           string    `json:"tax_name_ar,omitempty"`
	TaxCategoryCode     string    `json:"tax_category_code,omitempty"`
}
