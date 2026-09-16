package products_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"erp/internal/iam"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/products"
	"erp/internal/products/domain"
	"erp/internal/products/ports"
	"erp/internal/tax"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupProductTest(t *testing.T) (*products.Module, *tax.Module, *database.DB, uuid.UUID, context.Context) {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewPool(ctx, cfg)
	if err != nil {
		t.Skipf("PostgreSQL not accessible: %v", err)
	}

	txManager := transaction.NewManager(db.Pool)
	_ = iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	taxMod := tax.NewModule(cfg, db)
	prodMod := products.NewModule(cfg, db)

	// Ensure organization & legal entity exist
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// Seed tax & products default catalogs
	if err := taxMod.SeedDefaultTaxData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed tax data: %v", err)
	}
	if err := prodMod.SeedDefaultProductsData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed default products data: %v", err)
	}

	return prodMod, taxMod, db, le.ID, ctx
}

// 1. Units of Measure & Conversions
func TestProducts_UOMAndConversions(t *testing.T) {
	prodMod, _, _, _, ctx := setupProductTest(t)

	uoms, err := prodMod.ProductService.ListUOMs(ctx)
	if err != nil {
		t.Fatalf("failed to list UOMs: %v", err)
	}
	if len(uoms) < 5 {
		t.Fatalf("expected at least 5 seeded UOMs, got %d", len(uoms))
	}

	// Find KG and G
	var kgUOM, gUOM *domain.UnitOfMeasure
	for i := range uoms {
		if uoms[i].Code == "KG" {
			kgUOM = &uoms[i]
		} else if uoms[i].Code == "G" {
			gUOM = &uoms[i]
		}
	}
	if kgUOM == nil || gUOM == nil {
		t.Fatal("KG or G UOM not found in seeded data")
	}

	// Forward conversion: 2.5 KG -> G (1 KG = 1000 G)
	qty, err := prodMod.ProductService.ConvertQuantity(ctx, kgUOM.ID, gUOM.ID, decimal.NewFromFloat(2.5))
	if err != nil {
		t.Fatalf("failed to convert KG to G: %v", err)
	}
	if !qty.Equal(decimal.NewFromInt(2500)) {
		t.Fatalf("expected 2500 G for 2.5 KG, got: %s", qty.String())
	}

	// Inverse conversion: 5000 G -> KG (5000 / 1000 = 5 KG)
	qtyInv, err := prodMod.ProductService.ConvertQuantity(ctx, gUOM.ID, kgUOM.ID, decimal.NewFromInt(5000))
	if err != nil {
		t.Fatalf("failed inverse conversion G to KG: %v", err)
	}
	if !qtyInv.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("expected 5 KG for 5000 G, got: %s", qtyInv.String())
	}

	// Incompatible conversion returns ErrInvalidUOMConversion
	var mUOM *domain.UnitOfMeasure
	for i := range uoms {
		if uoms[i].Code == "M" {
			mUOM = &uoms[i]
			break
		}
	}
	if mUOM != nil {
		_, errIncomp := prodMod.ProductService.ConvertQuantity(ctx, kgUOM.ID, mUOM.ID, decimal.NewFromInt(1))
		if !errors.Is(errIncomp, domain.ErrInvalidUOMConversion) {
			t.Fatalf("expected ErrInvalidUOMConversion between KG and M, got: %v", errIncomp)
		}
	}

	// Create custom UOM: DOZEN
	dozen, err := prodMod.ProductService.CreateUOM(ctx, domain.UnitOfMeasure{
		Code:     "DZN_" + uuid.NewString()[:4],
		NameEn:   "Dozen",
		NameAr:   "درزن",
		SymbolEn: "dzn",
		SymbolAr: "درزن",
		UOMType:  "COUNT",
	})
	if err != nil {
		t.Fatalf("failed to create custom UOM: %v", err)
	}
	if dozen.Code == "" {
		t.Fatal("expected non-empty UOM code")
	}
}

// 2. Hierarchical Categories
func TestProducts_HierarchicalCategories(t *testing.T) {
	prodMod, _, _, leID, ctx := setupProductTest(t)

	cats, err := prodMod.ProductService.ListCategories(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list categories: %v", err)
	}
	if len(cats) < 2 {
		t.Fatalf("expected at least 2 seeded categories, got %d", len(cats))
	}

	var itCategory *domain.Category
	for i := range cats {
		if cats[i].Code == "CAT-IT" {
			itCategory = &cats[i]
			break
		}
	}
	if itCategory == nil {
		t.Fatal("CAT-IT root category not found")
	}

	// Create Child Category
	childCode := "LAPTOPS_" + uuid.NewString()[:4]
	child, err := prodMod.ProductService.CreateCategory(ctx, domain.Category{
		LegalEntityID: leID,
		ParentID:      &itCategory.ID,
		Code:          childCode,
		NameEn:        "Laptops & Notebooks",
		NameAr:        "حواسيب محمولة",
		Description:   "Business and enterprise grade laptops",
	})
	if err != nil {
		t.Fatalf("failed to create child category: %v", err)
	}
	if child.ParentID == nil || *child.ParentID != itCategory.ID {
		t.Fatalf("expected parent ID %v, got %v", itCategory.ID, child.ParentID)
	}
}

// 3. Item Master Constraints & Invariants
func TestProducts_ItemMasterConstraints(t *testing.T) {
	prodMod, taxMod, _, leID, ctx := setupProductTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsID := uoms[0].ID

	taxCodes, _ := taxMod.TaxService.ListTaxCodes(ctx, leID)
	var vat15ID *uuid.UUID
	if len(taxCodes) > 0 {
		vat15ID = &taxCodes[0].ID
	}

	itemCode := "SRV-TEST-" + uuid.NewString()[:6]

	// Successful Creation
	item, err := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID:     leID,
		Code:              itemCode,
		NameEn:            "Cloud Deployment Service",
		NameAr:            "خدمة نشر سحابي",
		ItemType:          domain.ItemTypeService,
		TrackingType:      domain.TrackingTypeNone,
		BaseUOMID:         pcsID,
		DefaultTaxCodeID:  vat15ID,
		CostPrice:         decimal.NewFromInt(500),
		ListPrice:         decimal.NewFromInt(1200),
		IsSalesEnabled:    true,
		IsPurchaseEnabled: true,
		DefaultBarcode:    "BC-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if item.Code != strings.ToUpper(itemCode) {
		t.Fatalf("expected code %s, got %s", strings.ToUpper(itemCode), item.Code)
	}

	// Duplicate code constraint
	_, dupErr := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           itemCode,
		NameEn:         "Duplicate Item",
		NameAr:         "صنف مكرر",
		ItemType:       domain.ItemTypeService,
		TrackingType:   domain.TrackingTypeNone,
		BaseUOMID:      pcsID,
		CostPrice:      decimal.NewFromInt(10),
		ListPrice:      decimal.NewFromInt(20),
		IsSalesEnabled: true,
	})
	if dupErr == nil {
		t.Fatal("expected duplicate item code to fail, got nil")
	}

	// Invalid Item Type
	_, invalidTypeErr := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID: leID,
		Code:          "INV-TYPE-" + uuid.NewString()[:4],
		NameEn:        "Invalid Type",
		NameAr:        "نوع خاطئ",
		ItemType:      "INVALID_TYPE",
		TrackingType:  domain.TrackingTypeNone,
		BaseUOMID:     pcsID,
	})
	if !errors.Is(invalidTypeErr, domain.ErrInvalidItemType) {
		t.Fatalf("expected ErrInvalidItemType, got %v", invalidTypeErr)
	}

	// Invalid Tracking Type
	_, invalidTrackErr := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID: leID,
		Code:          "INV-TRACK-" + uuid.NewString()[:4],
		NameEn:        "Invalid Tracking",
		NameAr:        "تتبع خاطئ",
		ItemType:      domain.ItemTypeStock,
		TrackingType:  "GPS_TRACK",
		BaseUOMID:     pcsID,
	})
	if !errors.Is(invalidTrackErr, domain.ErrInvalidTrackingType) {
		t.Fatalf("expected ErrInvalidTrackingType, got %v", invalidTrackErr)
	}

	// Search Items
	searchResults, err := prodMod.ProductService.SearchItems(ctx, leID, "Cloud Deployment")
	if err != nil {
		t.Fatalf("failed to search items: %v", err)
	}
	if len(searchResults) == 0 {
		t.Fatal("expected at least 1 search result for 'Cloud Deployment'")
	}
}

// 4. Variants & SKUs Generation
func TestProducts_VariantsAndSKUs(t *testing.T) {
	prodMod, _, _, leID, ctx := setupProductTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsID := uoms[0].ID

	baseItem, err := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID:     leID,
		Code:              "SHIRT-" + uuid.NewString()[:4],
		NameEn:            "Corporate Polo Shirt",
		NameAr:            "قميص بولو رسمي",
		ItemType:          domain.ItemTypeStock,
		TrackingType:      domain.TrackingTypeNone,
		BaseUOMID:         pcsID,
		CostPrice:         decimal.NewFromInt(40),
		ListPrice:         decimal.NewFromInt(95),
		IsSalesEnabled:    true,
		IsPurchaseEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create base item: %v", err)
	}

	// Add Variant: Blue / L
	vCode := baseItem.Code + "-BLU-L"
	variant, sku, err := prodMod.ProductService.CreateVariantWithSKU(
		ctx,
		leID,
		baseItem.ID,
		vCode,
		"Corporate Polo - Blue Large",
		"قميص بولو - أزرق كبير",
		map[string]any{"color": "Navy Blue", "size": "L"},
		decimal.NewFromInt(42),
		decimal.NewFromInt(100),
	)
	if err != nil {
		t.Fatalf("failed to create variant with SKU: %v", err)
	}
	if variant.VariantCode != vCode {
		t.Fatalf("expected variant code %s, got %s", vCode, variant.VariantCode)
	}
	if sku.SKUCode != vCode {
		t.Fatalf("expected SKU code %s, got %s", vCode, sku.SKUCode)
	}

	// Verify Listing Variants & SKUs
	variants, err := prodMod.ProductService.ListVariants(ctx, baseItem.ID)
	if err != nil || len(variants) != 1 {
		t.Fatalf("expected 1 variant, got %d, err: %v", len(variants), err)
	}

	skus, err := prodMod.ProductService.ListSKUs(ctx, baseItem.ID)
	if err != nil || len(skus) < 2 { // root SKU + variant SKU
		t.Fatalf("expected at least 2 SKUs (root + variant), got %d, err: %v", len(skus), err)
	}
}

// 5. Barcode Lookup & Scanner Resolution
func TestProducts_BarcodeLookup(t *testing.T) {
	prodMod, _, _, leID, ctx := setupProductTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsID := uoms[0].ID

	barcodeVal := "SCAN-" + uuid.NewString()[:8]

	item, err := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID:     leID,
		Code:              "SCAN-ITEM-" + uuid.NewString()[:4],
		NameEn:            "Handheld Barcode Scanner 2D",
		NameAr:            "قارئ باركود ثنائي الأبعاد",
		ItemType:          domain.ItemTypeStock,
		TrackingType:      domain.TrackingTypeSerial,
		BaseUOMID:         pcsID,
		CostPrice:         decimal.NewFromInt(150),
		ListPrice:         decimal.NewFromInt(320),
		IsSalesEnabled:    true,
		IsPurchaseEnabled: true,
		DefaultBarcode:    barcodeVal,
	})
	if err != nil {
		t.Fatalf("failed to create item with barcode: %v", err)
	}

	// Scanner Lookup Test
	foundBC, err := prodMod.ProductService.LookupBarcode(ctx, leID, barcodeVal)
	if err != nil {
		t.Fatalf("failed to lookup barcode: %v", err)
	}
	if foundBC.ItemID != item.ID {
		t.Fatalf("expected item ID %v, got %v", item.ID, foundBC.ItemID)
	}
	if !foundBC.ListPrice.Equal(decimal.NewFromInt(320)) {
		t.Fatalf("expected list price 320, got: %s", foundBC.ListPrice.String())
	}

	// Non-existent Barcode Lookup
	_, notFoundErr := prodMod.ProductService.LookupBarcode(ctx, leID, "NON-EXISTENT-CODE-999")
	if !errors.Is(notFoundErr, domain.ErrBarcodeNotFound) {
		t.Fatalf("expected ErrBarcodeNotFound, got: %v", notFoundErr)
	}
}

// 6. Packaging Units of Measure
func TestProducts_ItemPackagingUOMs(t *testing.T) {
	prodMod, _, _, leID, ctx := setupProductTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	var pcsID, boxID uuid.UUID
	for _, u := range uoms {
		if u.Code == "PCS" {
			pcsID = u.ID
		} else if u.Code == "BOX" {
			boxID = u.ID
		}
	}

	item, err := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID:     leID,
		Code:              "PKG-ITEM-" + uuid.NewString()[:4],
		NameEn:            "Thermal Receipt Paper Rolls",
		NameAr:            "لفائف ورق حراري للمحاسبة",
		ItemType:          domain.ItemTypeStock,
		TrackingType:      domain.TrackingTypeNone,
		BaseUOMID:         pcsID,
		CostPrice:         decimal.NewFromInt(2),
		ListPrice:         decimal.NewFromInt(5),
		IsSalesEnabled:    true,
		IsPurchaseEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add Packaging UOM: Box of 50 Rolls
	iu, err := prodMod.ProductService.AddItemUOM(ctx, domain.ItemUOM{
		ItemID:         item.ID,
		UOMID:          boxID,
		ConversionRate: decimal.NewFromInt(50),
		IsBaseUOM:      false,
		IsSalesUOM:     true,
		IsPurchaseUOM:  true,
	})
	if err != nil {
		t.Fatalf("failed to add item packaging UOM: %v", err)
	}
	if !iu.ConversionRate.Equal(decimal.NewFromInt(50)) {
		t.Fatalf("expected conversion rate 50, got: %s", iu.ConversionRate.String())
	}

	// List Item UOMs
	itemUOMs, err := prodMod.ProductService.ListItemUOMs(ctx, item.ID)
	if err != nil || len(itemUOMs) < 2 {
		t.Fatalf("expected at least 2 item UOMs (base PCS + packaging BOX), got %d, err: %v", len(itemUOMs), err)
	}
}

// 7. ZATCA Item Tax Profile
func TestProducts_ItemTaxProfile(t *testing.T) {
	prodMod, taxMod, _, leID, ctx := setupProductTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsID := uoms[0].ID

	taxCodes, _ := taxMod.TaxService.ListTaxCodes(ctx, leID)
	if len(taxCodes) == 0 {
		t.Fatal("no tax codes found")
	}
	taxCode := taxCodes[0]

	item, err := prodMod.ProductService.CreateItem(ctx, ports.CreateItemRequest{
		LegalEntityID:     leID,
		Code:              "TAX-ITEM-" + uuid.NewString()[:4],
		NameEn:            "Taxable Electronic Terminal",
		NameAr:            "جهاز نقطة بيع إلكتروني",
		ItemType:          domain.ItemTypeStock,
		TrackingType:      domain.TrackingTypeSerial,
		BaseUOMID:         pcsID,
		CostPrice:         decimal.NewFromInt(800),
		ListPrice:         decimal.NewFromInt(1500),
		IsSalesEnabled:    true,
		IsPurchaseEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Set Tax Profile
	itp, err := prodMod.ProductService.SetItemTaxProfile(ctx, domain.ItemTaxProfile{
		ItemID:    item.ID,
		TaxCodeID: taxCode.ID,
	})
	if err != nil {
		t.Fatalf("failed to set tax profile: %v", err)
	}
	if itp.TaxCodeID != taxCode.ID {
		t.Fatalf("expected tax code ID %v, got %v", taxCode.ID, itp.TaxCodeID)
	}

	// Retrieve Tax Profile
	fetched, err := prodMod.ProductService.GetItemTaxProfile(ctx, item.ID)
	if err != nil || fetched == nil {
		t.Fatalf("failed to get item tax profile: %v", err)
	}
	if fetched.TaxCode != taxCode.Code {
		t.Fatalf("expected tax code %s, got %s", taxCode.Code, fetched.TaxCode)
	}
}
