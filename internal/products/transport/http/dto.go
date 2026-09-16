package http

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateUOMRequest struct {
	Code     string `json:"code"`
	NameEn   string `json:"name_en"`
	NameAr   string `json:"name_ar"`
	SymbolEn string `json:"symbol_en,omitempty"`
	SymbolAr string `json:"symbol_ar,omitempty"`
	UOMType  string `json:"uom_type"`
}

type ConvertUOMRequest struct {
	FromUOMID uuid.UUID       `json:"from_uom_id"`
	ToUOMID   uuid.UUID       `json:"to_uom_id"`
	Quantity  decimal.Decimal `json:"quantity"`
}

type CreateCategoryRequest struct {
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Code        string     `json:"code"`
	NameEn      string     `json:"name_en"`
	NameAr      string     `json:"name_ar"`
	Description string     `json:"description,omitempty"`
}

type CreateItemRequestDTO struct {
	CategoryID        *uuid.UUID      `json:"category_id,omitempty"`
	Code              string          `json:"code"`
	NameEn            string          `json:"name_en"`
	NameAr            string          `json:"name_ar"`
	DescriptionEn     string          `json:"description_en,omitempty"`
	DescriptionAr     string          `json:"description_ar,omitempty"`
	ItemType          string          `json:"item_type"`
	TrackingType      string          `json:"tracking_type"`
	BaseUOMID         uuid.UUID       `json:"base_uom_id"`
	PurchaseUOMID     *uuid.UUID      `json:"purchase_uom_id,omitempty"`
	SalesUOMID        *uuid.UUID      `json:"sales_uom_id,omitempty"`
	DefaultTaxCodeID  *uuid.UUID      `json:"default_tax_code_id,omitempty"`
	CostPrice         decimal.Decimal `json:"cost_price"`
	ListPrice         decimal.Decimal `json:"list_price"`
	IsSalesEnabled    bool            `json:"is_sales_enabled"`
	IsPurchaseEnabled bool            `json:"is_purchase_enabled"`
	DefaultBarcode    string          `json:"default_barcode,omitempty"`
}

type UpdateItemRequestDTO struct {
	CategoryID        *uuid.UUID      `json:"category_id,omitempty"`
	NameEn            string          `json:"name_en"`
	NameAr            string          `json:"name_ar"`
	DescriptionEn     string          `json:"description_en,omitempty"`
	DescriptionAr     string          `json:"description_ar,omitempty"`
	ItemType          string          `json:"item_type"`
	TrackingType      string          `json:"tracking_type"`
	BaseUOMID         uuid.UUID       `json:"base_uom_id"`
	PurchaseUOMID     *uuid.UUID      `json:"purchase_uom_id,omitempty"`
	SalesUOMID        *uuid.UUID      `json:"sales_uom_id,omitempty"`
	DefaultTaxCodeID  *uuid.UUID      `json:"default_tax_code_id,omitempty"`
	CostPrice         decimal.Decimal `json:"cost_price"`
	ListPrice         decimal.Decimal `json:"list_price"`
	IsSalesEnabled    bool            `json:"is_sales_enabled"`
	IsPurchaseEnabled bool            `json:"is_purchase_enabled"`
	IsActive          bool            `json:"is_active"`
}

type CreateVariantRequestDTO struct {
	VariantCode  string          `json:"variant_code"`
	NameEn       string          `json:"name_en"`
	NameAr       string          `json:"name_ar"`
	Attributes   map[string]any  `json:"attributes"`
	CostPrice    decimal.Decimal `json:"cost_price"`
	SellingPrice decimal.Decimal `json:"selling_price"`
}

type AddBarcodeRequestDTO struct {
	SKUID         *uuid.UUID `json:"sku_id,omitempty"`
	UOMID         *uuid.UUID `json:"uom_id,omitempty"`
	Barcode       string     `json:"barcode"`
	BarcodeFormat string     `json:"barcode_format,omitempty"`
	IsPrimary     bool       `json:"is_primary"`
}

type AddItemUOMRequestDTO struct {
	UOMID          uuid.UUID       `json:"uom_id"`
	ConversionRate decimal.Decimal `json:"conversion_rate"`
	IsSalesUOM     bool            `json:"is_sales_uom"`
	IsPurchaseUOM  bool            `json:"is_purchase_uom"`
}

type SetTaxProfileRequestDTO struct {
	TaxCodeID           uuid.UUID `json:"tax_code_id"`
	ExemptionReasonCode *string   `json:"exemption_reason_code,omitempty"`
	ZeroRateReasonCode  *string   `json:"zero_rate_reason_code,omitempty"`
}
