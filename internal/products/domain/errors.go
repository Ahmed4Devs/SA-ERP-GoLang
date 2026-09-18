package domain

import "errors"

var (
	ErrItemNotFound         = errors.New("product item not found")
	ErrCategoryNotFound     = errors.New("product category not found")
	ErrUOMNotFound          = errors.New("unit of measure not found")
	ErrVariantNotFound      = errors.New("item variant not found")
	ErrSKUNotFound          = errors.New("sku not found")
	ErrBarcodeNotFound      = errors.New("barcode not found")
	ErrDuplicateItemCode    = errors.New("duplicate item code for legal entity")
	ErrDuplicateSKU         = errors.New("duplicate sku code for legal entity")
	ErrDuplicateBarcode     = errors.New("duplicate barcode for legal entity")
	ErrInvalidUOMConversion = errors.New("invalid unit of measure conversion")
	ErrInvalidTrackingType  = errors.New("invalid tracking type: must be NONE, LOT, or SERIAL")
	ErrInvalidItemType      = errors.New("invalid item type: must be STOCK, SERVICE, NON_STOCK, or ASSET")
)
