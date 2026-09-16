package services

import (
	"context"
	"strings"

	"erp/internal/products/domain"
	"erp/internal/products/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductService struct {
	repo ports.ProductRepository
}

func NewProductService(repo ports.ProductRepository) ports.ProductService {
	return &ProductService{repo: repo}
}

// Units of Measure
func (s *ProductService) ListUOMs(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return s.repo.ListUnitsOfMeasure(ctx)
}

func (s *ProductService) CreateUOM(ctx context.Context, uom domain.UnitOfMeasure) (*domain.UnitOfMeasure, error) {
	if uom.ID == uuid.Nil {
		uom.ID = uuid.New()
	}
	uom.Code = strings.ToUpper(strings.TrimSpace(uom.Code))
	uom.IsActive = true
	return s.repo.CreateUnitOfMeasure(ctx, &uom)
}

func (s *ProductService) ConvertQuantity(ctx context.Context, fromUOMID, toUOMID uuid.UUID, quantity decimal.Decimal) (decimal.Decimal, error) {
	if fromUOMID == toUOMID {
		return quantity, nil
	}

	// Try direct conversion
	rate, err := s.repo.GetUOMConversionRate(ctx, fromUOMID, toUOMID)
	if err == nil && !rate.IsZero() {
		return quantity.Mul(rate), nil
	}

	// Try inverse conversion
	rateInv, errInv := s.repo.GetUOMConversionRate(ctx, toUOMID, fromUOMID)
	if errInv == nil && !rateInv.IsZero() {
		return quantity.Div(rateInv), nil
	}

	return decimal.Zero, domain.ErrInvalidUOMConversion
}

// Categories
func (s *ProductService) ListCategories(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx, legalEntityID)
}

func (s *ProductService) CreateCategory(ctx context.Context, cat domain.Category) (*domain.Category, error) {
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	cat.Code = strings.ToUpper(strings.TrimSpace(cat.Code))
	cat.IsActive = true
	return s.repo.CreateCategory(ctx, &cat)
}

// Items Master
func (s *ProductService) ListItems(ctx context.Context, legalEntityID uuid.UUID, itemType *string, categoryID *uuid.UUID, limit, offset int32) ([]domain.Item, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListItems(ctx, legalEntityID, itemType, categoryID, limit, offset)
}

func (s *ProductService) GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	return s.repo.GetItemByID(ctx, id)
}

func (s *ProductService) GetItemByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Item, error) {
	return s.repo.GetItemByCode(ctx, legalEntityID, strings.TrimSpace(code))
}

func (s *ProductService) CreateItem(ctx context.Context, req ports.CreateItemRequest) (*domain.Item, error) {
	// Validate item type
	switch req.ItemType {
	case domain.ItemTypeStock, domain.ItemTypeService, domain.ItemTypeNonStock, domain.ItemTypeAsset:
	default:
		return nil, domain.ErrInvalidItemType
	}

	// Validate tracking type
	switch req.TrackingType {
	case domain.TrackingTypeNone, domain.TrackingTypeLot, domain.TrackingTypeSerial:
	default:
		return nil, domain.ErrInvalidTrackingType
	}

	item := domain.Item{
		ID:                uuid.New(),
		LegalEntityID:     req.LegalEntityID,
		CategoryID:        req.CategoryID,
		Code:              strings.ToUpper(strings.TrimSpace(req.Code)),
		NameEn:            strings.TrimSpace(req.NameEn),
		NameAr:            strings.TrimSpace(req.NameAr),
		DescriptionEn:     strings.TrimSpace(req.DescriptionEn),
		DescriptionAr:     strings.TrimSpace(req.DescriptionAr),
		ItemType:          req.ItemType,
		TrackingType:      req.TrackingType,
		BaseUOMID:         req.BaseUOMID,
		PurchaseUOMID:     req.PurchaseUOMID,
		SalesUOMID:        req.SalesUOMID,
		DefaultTaxCodeID:  req.DefaultTaxCodeID,
		CostPrice:         req.CostPrice,
		ListPrice:         req.ListPrice,
		IsSalesEnabled:    req.IsSalesEnabled,
		IsPurchaseEnabled: req.IsPurchaseEnabled,
		IsActive:          true,
	}

	created, err := s.repo.CreateItem(ctx, &item)
	if err != nil {
		return nil, err
	}

	// 1. Automatically register Base UOM packaging entry
	baseUOM := domain.ItemUOM{
		ID:             uuid.New(),
		ItemID:         created.ID,
		UOMID:          req.BaseUOMID,
		ConversionRate: decimal.NewFromInt(1),
		IsBaseUOM:      true,
		IsSalesUOM:     req.SalesUOMID == nil || *req.SalesUOMID == req.BaseUOMID,
		IsPurchaseUOM:  req.PurchaseUOMID == nil || *req.PurchaseUOMID == req.BaseUOMID,
	}
	_, _ = s.repo.UpsertItemUOM(ctx, &baseUOM)

	// 2. Register Purchase UOM if distinct from Base
	if req.PurchaseUOMID != nil && *req.PurchaseUOMID != req.BaseUOMID {
		purUOM := domain.ItemUOM{
			ID:             uuid.New(),
			ItemID:         created.ID,
			UOMID:          *req.PurchaseUOMID,
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUOM:      false,
			IsSalesUOM:     false,
			IsPurchaseUOM:  true,
		}
		_, _ = s.repo.UpsertItemUOM(ctx, &purUOM)
	}

	// 3. Register Sales UOM if distinct from Base and Purchase
	if req.SalesUOMID != nil && *req.SalesUOMID != req.BaseUOMID && (req.PurchaseUOMID == nil || *req.SalesUOMID != *req.PurchaseUOMID) {
		saleUOM := domain.ItemUOM{
			ID:             uuid.New(),
			ItemID:         created.ID,
			UOMID:          *req.SalesUOMID,
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUOM:      false,
			IsSalesUOM:     true,
			IsPurchaseUOM:  false,
		}
		_, _ = s.repo.UpsertItemUOM(ctx, &saleUOM)
	}

	// 4. Create default root SKU for this item
	sku := domain.SKU{
		ID:            uuid.New(),
		LegalEntityID: req.LegalEntityID,
		ItemID:        created.ID,
		VariantID:     nil,
		SKUCode:       created.Code,
		CostPrice:     req.CostPrice,
		SellingPrice:  req.ListPrice,
		IsActive:      true,
	}
	createdSKU, _ := s.repo.CreateSKU(ctx, &sku)

	// 5. Default barcode if provided
	if trimmedBarcode := strings.TrimSpace(req.DefaultBarcode); trimmedBarcode != "" {
		bc := domain.Barcode{
			ID:            uuid.New(),
			LegalEntityID: req.LegalEntityID,
			ItemID:        created.ID,
			Barcode:       trimmedBarcode,
			BarcodeFormat: "CUSTOM",
			IsPrimary:     true,
		}
		if createdSKU != nil {
			bc.SKUID = &createdSKU.ID
		}
		bc.UOMID = &req.BaseUOMID
		_, _ = s.repo.CreateBarcode(ctx, &bc)
	}

	// 6. Tax Profile if default tax code provided
	if req.DefaultTaxCodeID != nil {
		itp := domain.ItemTaxProfile{
			ID:        uuid.New(),
			ItemID:    created.ID,
			TaxCodeID: *req.DefaultTaxCodeID,
			IsActive:  true,
		}
		_, _ = s.repo.UpsertItemTaxProfile(ctx, &itp)
	}

	return created, nil
}

func (s *ProductService) UpdateItem(ctx context.Context, item domain.Item) (*domain.Item, error) {
	return s.repo.UpdateItem(ctx, &item)
}

func (s *ProductService) SearchItems(ctx context.Context, legalEntityID uuid.UUID, query string) ([]domain.Item, error) {
	return s.repo.SearchItems(ctx, legalEntityID, strings.TrimSpace(query))
}

// Variants & SKUs
func (s *ProductService) CreateVariantWithSKU(ctx context.Context, legalEntityID uuid.UUID, itemID uuid.UUID, variantCode, nameEn, nameAr string, attrs map[string]any, cost, selling decimal.Decimal) (*domain.ItemVariant, *domain.SKU, error) {
	v := domain.ItemVariant{
		ID:          uuid.New(),
		ItemID:      itemID,
		VariantCode: strings.ToUpper(strings.TrimSpace(variantCode)),
		NameEn:      strings.TrimSpace(nameEn),
		NameAr:      strings.TrimSpace(nameAr),
		Attributes:  attrs,
		IsActive:    true,
	}
	createdV, err := s.repo.CreateVariant(ctx, &v)
	if err != nil {
		return nil, nil, err
	}

	sku := domain.SKU{
		ID:            uuid.New(),
		LegalEntityID: legalEntityID,
		ItemID:        itemID,
		VariantID:     &createdV.ID,
		SKUCode:       createdV.VariantCode,
		CostPrice:     cost,
		SellingPrice:  selling,
		IsActive:      true,
	}
	createdSKU, err := s.repo.CreateSKU(ctx, &sku)
	if err != nil {
		return nil, nil, err
	}

	return createdV, createdSKU, nil
}

func (s *ProductService) ListVariants(ctx context.Context, itemID uuid.UUID) ([]domain.ItemVariant, error) {
	return s.repo.ListVariants(ctx, itemID)
}

func (s *ProductService) ListSKUs(ctx context.Context, itemID uuid.UUID) ([]domain.SKU, error) {
	return s.repo.ListSKUs(ctx, itemID)
}

// Barcodes
func (s *ProductService) LookupBarcode(ctx context.Context, legalEntityID uuid.UUID, barcode string) (*domain.Barcode, error) {
	return s.repo.GetBarcode(ctx, legalEntityID, strings.TrimSpace(barcode))
}

func (s *ProductService) AddBarcode(ctx context.Context, bc domain.Barcode) (*domain.Barcode, error) {
	if bc.ID == uuid.Nil {
		bc.ID = uuid.New()
	}
	bc.Barcode = strings.TrimSpace(bc.Barcode)
	if bc.BarcodeFormat == "" {
		bc.BarcodeFormat = "CUSTOM"
	}
	return s.repo.CreateBarcode(ctx, &bc)
}

func (s *ProductService) ListBarcodes(ctx context.Context, itemID uuid.UUID) ([]domain.Barcode, error) {
	return s.repo.ListBarcodes(ctx, itemID)
}

// Item Packaging UOMs
func (s *ProductService) ListItemUOMs(ctx context.Context, itemID uuid.UUID) ([]domain.ItemUOM, error) {
	return s.repo.ListItemUOMs(ctx, itemID)
}

func (s *ProductService) AddItemUOM(ctx context.Context, iu domain.ItemUOM) (*domain.ItemUOM, error) {
	if iu.ID == uuid.Nil {
		iu.ID = uuid.New()
	}
	return s.repo.UpsertItemUOM(ctx, &iu)
}

// Item Tax Profiles
func (s *ProductService) GetItemTaxProfile(ctx context.Context, itemID uuid.UUID) (*domain.ItemTaxProfile, error) {
	return s.repo.GetItemTaxProfile(ctx, itemID)
}

func (s *ProductService) SetItemTaxProfile(ctx context.Context, itp domain.ItemTaxProfile) (*domain.ItemTaxProfile, error) {
	if itp.ID == uuid.Nil {
		itp.ID = uuid.New()
	}
	itp.IsActive = true
	return s.repo.UpsertItemTaxProfile(ctx, &itp)
}
