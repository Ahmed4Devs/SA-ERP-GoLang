package services

import (
	"context"
	"fmt"
	"strings"

	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/products/domain"
	"erp/internal/products/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductService struct {
	repo ports.ProductRepository
	tx   transaction.Manager
}

func NewProductService(repo ports.ProductRepository, db *database.DB) ports.ProductService {
	return &ProductService{repo: repo, tx: transaction.NewManager(db.Pool)}
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
	if req.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.NameEn) == "" || strings.TrimSpace(req.NameAr) == "" {
		return nil, fmt.Errorf("code and bilingual names are required")
	}
	if req.BaseUOMID == uuid.Nil {
		return nil, fmt.Errorf("base_uom_id is required")
	}
	if req.CostPrice.IsNegative() || req.ListPrice.IsNegative() {
		return nil, fmt.Errorf("item prices cannot be negative")
	}
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

	var created *domain.Item
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreateItem(txCtx, &item)
		if err != nil {
			return err
		}

		uoms := []domain.ItemUOM{{
			ID:             uuid.New(),
			ItemID:         created.ID,
			UOMID:          req.BaseUOMID,
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUOM:      true,
			IsSalesUOM:     req.SalesUOMID == nil || *req.SalesUOMID == req.BaseUOMID,
			IsPurchaseUOM:  req.PurchaseUOMID == nil || *req.PurchaseUOMID == req.BaseUOMID,
		}}
		if req.PurchaseUOMID != nil && *req.PurchaseUOMID != req.BaseUOMID {
			uoms = append(uoms, domain.ItemUOM{
				ID:             uuid.New(),
				ItemID:         created.ID,
				UOMID:          *req.PurchaseUOMID,
				ConversionRate: decimal.NewFromInt(1),
				IsPurchaseUOM:  true,
				IsSalesUOM:     req.SalesUOMID != nil && *req.SalesUOMID == *req.PurchaseUOMID,
			})
		}
		if req.SalesUOMID != nil && *req.SalesUOMID != req.BaseUOMID && (req.PurchaseUOMID == nil || *req.SalesUOMID != *req.PurchaseUOMID) {
			uoms = append(uoms, domain.ItemUOM{
				ID:             uuid.New(),
				ItemID:         created.ID,
				UOMID:          *req.SalesUOMID,
				ConversionRate: decimal.NewFromInt(1),
				IsSalesUOM:     true,
			})
		}
		for i := range uoms {
			if _, err := s.repo.UpsertItemUOM(txCtx, &uoms[i]); err != nil {
				return fmt.Errorf("failed to register item UOM: %w", err)
			}
		}

		sku := domain.SKU{
			ID:            uuid.New(),
			LegalEntityID: req.LegalEntityID,
			ItemID:        created.ID,
			SKUCode:       created.Code,
			CostPrice:     req.CostPrice,
			SellingPrice:  req.ListPrice,
			IsActive:      true,
		}
		createdSKU, err := s.repo.CreateSKU(txCtx, &sku)
		if err != nil {
			return fmt.Errorf("failed to create root SKU: %w", err)
		}

		if trimmedBarcode := strings.TrimSpace(req.DefaultBarcode); trimmedBarcode != "" {
			bc := domain.Barcode{
				ID:            uuid.New(),
				LegalEntityID: req.LegalEntityID,
				ItemID:        created.ID,
				SKUID:         &createdSKU.ID,
				UOMID:         &req.BaseUOMID,
				Barcode:       trimmedBarcode,
				BarcodeFormat: "CUSTOM",
				IsPrimary:     true,
			}
			if _, err := s.repo.CreateBarcode(txCtx, &bc); err != nil {
				return fmt.Errorf("failed to create default barcode: %w", err)
			}
		}

		if req.DefaultTaxCodeID != nil {
			itp := domain.ItemTaxProfile{
				ID:        uuid.New(),
				ItemID:    created.ID,
				TaxCodeID: *req.DefaultTaxCodeID,
				IsActive:  true,
			}
			if _, err := s.repo.UpsertItemTaxProfile(txCtx, &itp); err != nil {
				return fmt.Errorf("failed to create item tax profile: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *ProductService) UpdateItem(ctx context.Context, item domain.Item) (*domain.Item, error) {
	if item.ID == uuid.Nil {
		return nil, fmt.Errorf("item id is required")
	}
	if strings.TrimSpace(item.NameEn) == "" || strings.TrimSpace(item.NameAr) == "" {
		return nil, fmt.Errorf("bilingual item names are required")
	}
	if item.CostPrice.IsNegative() || item.ListPrice.IsNegative() {
		return nil, fmt.Errorf("item prices cannot be negative")
	}
	switch item.ItemType {
	case domain.ItemTypeStock, domain.ItemTypeService, domain.ItemTypeNonStock, domain.ItemTypeAsset:
	default:
		return nil, domain.ErrInvalidItemType
	}
	switch item.TrackingType {
	case domain.TrackingTypeNone, domain.TrackingTypeLot, domain.TrackingTypeSerial:
	default:
		return nil, domain.ErrInvalidTrackingType
	}
	return s.repo.UpdateItem(ctx, &item)
}

func (s *ProductService) SearchItems(ctx context.Context, legalEntityID uuid.UUID, query string) ([]domain.Item, error) {
	return s.repo.SearchItems(ctx, legalEntityID, strings.TrimSpace(query))
}

// Variants & SKUs
func (s *ProductService) CreateVariantWithSKU(ctx context.Context, legalEntityID uuid.UUID, itemID uuid.UUID, variantCode, nameEn, nameAr string, attrs map[string]any, cost, selling decimal.Decimal) (*domain.ItemVariant, *domain.SKU, error) {
	if legalEntityID == uuid.Nil || itemID == uuid.Nil {
		return nil, nil, fmt.Errorf("legal_entity_id and item_id are required")
	}
	if strings.TrimSpace(variantCode) == "" || strings.TrimSpace(nameEn) == "" || strings.TrimSpace(nameAr) == "" {
		return nil, nil, fmt.Errorf("variant code and bilingual names are required")
	}
	if cost.IsNegative() || selling.IsNegative() {
		return nil, nil, fmt.Errorf("variant prices cannot be negative")
	}
	v := domain.ItemVariant{
		ID:          uuid.New(),
		ItemID:      itemID,
		VariantCode: strings.ToUpper(strings.TrimSpace(variantCode)),
		NameEn:      strings.TrimSpace(nameEn),
		NameAr:      strings.TrimSpace(nameAr),
		Attributes:  attrs,
		IsActive:    true,
	}
	var createdV *domain.ItemVariant
	var createdSKU *domain.SKU
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdV, err = s.repo.CreateVariant(txCtx, &v)
		if err != nil {
			return err
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
		createdSKU, err = s.repo.CreateSKU(txCtx, &sku)
		if err != nil {
			return fmt.Errorf("failed to create variant SKU: %w", err)
		}
		return nil
	})
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
	if bc.LegalEntityID == uuid.Nil || bc.ItemID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id and item_id are required")
	}
	if bc.ID == uuid.Nil {
		bc.ID = uuid.New()
	}
	bc.Barcode = strings.TrimSpace(bc.Barcode)
	if bc.Barcode == "" {
		return nil, fmt.Errorf("barcode is required")
	}
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
	if iu.ItemID == uuid.Nil || iu.UOMID == uuid.Nil {
		return nil, fmt.Errorf("item_id and uom_id are required")
	}
	if !iu.ConversionRate.IsPositive() {
		return nil, fmt.Errorf("conversion rate must be greater than zero")
	}
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
	if itp.ItemID == uuid.Nil || itp.TaxCodeID == uuid.Nil {
		return nil, fmt.Errorf("item_id and tax_code_id are required")
	}
	if itp.ID == uuid.Nil {
		itp.ID = uuid.New()
	}
	itp.IsActive = true
	return s.repo.UpsertItemTaxProfile(ctx, &itp)
}
