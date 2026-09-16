package inventory_test

import (
	"strings"
	"testing"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"
	prodDomain "erp/internal/products/domain"
	prodPorts "erp/internal/products/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// 1. Moving Weighted Average Invariant Test
func TestCosting_MovingWeightedAverageCalculation(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	// Create Item
	item, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-AVG-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Industrial Sensor",
		NameAr:         "حساس صناعي رقمي",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(100),
		ListPrice:      decimal.NewFromInt(160),
		IsSalesEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Step 1: Receipt 10 units @ 100 SAR = 1000 SAR
	doc1 := "PO-COST-001"
	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypePurchaseReceipt,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(100),
		SourceDocType: &doc1,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post receipt 1: %v", err)
	}

	bal1, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
	})
	if !bal1.QuantityOnHand.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected on-hand 10, got %s", bal1.QuantityOnHand)
	}
	if !bal1.AverageCost.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected avg cost 100, got %s", bal1.AverageCost)
	}
	if !bal1.TotalValue.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("expected total value 1000, got %s", bal1.TotalValue)
	}

	// Step 2: Receipt 10 units @ 200 SAR = 2000 SAR
	// New total qty = 20, New total val = 3000 SAR, New avg = 150 SAR
	doc2 := "PO-COST-002"
	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypePurchaseReceipt,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(200),
		SourceDocType: &doc2,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post receipt 2: %v", err)
	}

	bal2, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
	})
	if !bal2.QuantityOnHand.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("expected on-hand 20, got %s", bal2.QuantityOnHand)
	}
	if !bal2.AverageCost.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("expected avg cost 150, got %s", bal2.AverageCost)
	}
	if !bal2.TotalValue.Equal(decimal.NewFromInt(3000)) {
		t.Fatalf("expected total value 3000, got %s", bal2.TotalValue)
	}

	// Step 3: Issue 5 units. Cost should be at current moving average (150 SAR).
	// Total cost = 750 SAR. Remaining on-hand = 15, Remaining val = 2250 SAR.
	doc3 := "SO-COST-001"
	movOut, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeSalesIssue,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(5),
		UOMID:         pcsUOM.ID,
		SourceDocType: &doc3,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post sales issue: %v", err)
	}
	if !movOut.UnitCost.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("expected issue unit cost 150, got %s", movOut.UnitCost)
	}
	if !movOut.TotalCost.Equal(decimal.NewFromInt(750)) {
		t.Fatalf("expected issue total cost 750, got %s", movOut.TotalCost)
	}

	bal3, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
	})
	if !bal3.QuantityOnHand.Equal(decimal.NewFromInt(15)) {
		t.Fatalf("expected on-hand 15, got %s", bal3.QuantityOnHand)
	}
	if !bal3.AverageCost.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("expected avg cost to remain 150, got %s", bal3.AverageCost)
	}
	if !bal3.TotalValue.Equal(decimal.NewFromInt(2250)) {
		t.Fatalf("expected total value 2250, got %s", bal3.TotalValue)
	}
}

// 2. FIFO Layer Creation & Consumption Invariant Test
func TestCosting_FIFOLayerConsumption(t *testing.T) {
	invMod, prodMod, db, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	// Create Item
	item, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-FIFO-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Precision Bearing",
		NameAr:         "محمل كروي دقيق",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(50),
		ListPrice:      decimal.NewFromInt(90),
		IsSalesEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Update item costing method to FIFO
	_, err = db.Pool.Exec(ctx, "UPDATE products.items SET costing_method = 'FIFO' WHERE id = $1", item.ID)
	if err != nil {
		t.Fatalf("failed to set item costing_method to FIFO: %v", err)
	}

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Receipt 1: 10 units @ 100 SAR
	doc1 := "PO-FIFO-001"
	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypePurchaseReceipt,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(100),
		SourceDocType: &doc1,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post receipt 1: %v", err)
	}

	// Receipt 2: 10 units @ 150 SAR
	doc2 := "PO-FIFO-002"
	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypePurchaseReceipt,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(150),
		SourceDocType: &doc2,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post receipt 2: %v", err)
	}

	// Verify 2 cost layers exist
	activeOnly := true
	layers, err := invMod.InventoryService.ListCostLayers(ctx, leID, &item.ID, &mainWH.ID, &activeOnly, 10, 0)
	if err != nil {
		t.Fatalf("failed to list cost layers: %v", err)
	}
	if len(layers) != 2 {
		t.Fatalf("expected 2 active cost layers, got %d", len(layers))
	}
	if !layers[0].RemainingQuantity.Equal(decimal.NewFromInt(10)) || !layers[0].UnitCost.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("unexpected layer 0: %+v", layers[0])
	}
	if !layers[1].RemainingQuantity.Equal(decimal.NewFromInt(10)) || !layers[1].UnitCost.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("unexpected layer 1: %+v", layers[1])
	}

	// Issue 15 units: Should consume ALL 10 from Layer 1 (1000 SAR) + 5 from Layer 2 (750 SAR)
	// Total cost = 1750 SAR, Blended unit cost = 1750 / 15 = 116.6667 SAR
	doc3 := "SO-FIFO-001"
	movOut, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeSalesIssue,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(15),
		UOMID:         pcsUOM.ID,
		SourceDocType: &doc3,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post FIFO sales issue: %v", err)
	}
	if !movOut.TotalCost.Equal(decimal.NewFromInt(1750)) {
		t.Fatalf("expected total cost 1750, got %s", movOut.TotalCost)
	}

	// Verify remaining layers: Layer 1 should have 0 remaining, Layer 2 should have 5 remaining
	activeLayers, _ := invMod.InventoryService.ListCostLayers(ctx, leID, &item.ID, &mainWH.ID, &activeOnly, 10, 0)
	if len(activeLayers) != 1 {
		t.Fatalf("expected 1 remaining active layer, got %d", len(activeLayers))
	}
	if !activeLayers[0].RemainingQuantity.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("expected layer 2 remaining qty 5, got %s", activeLayers[0].RemainingQuantity)
	}
}

// 3. Inventory Revaluation Test
func TestCosting_InventoryRevaluation(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-REV-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Heavy Transformer",
		NameAr:         "محول كهربائي ضغط عالي",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(2000),
		ListPrice:      decimal.NewFromInt(2800),
		IsSalesEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Post opening balance: 10 units @ 2000 SAR = 20,000 SAR
	doc := "INIT-REV-01"
	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(2000),
		SourceDocType: &doc,
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post opening balance: %v", err)
	}

	// Revalue item from 2000 to 2200 SAR (+2000 SAR variance)
	rev, err := invMod.InventoryService.ExecuteRevaluation(ctx, ports.ExecuteRevaluationRequest{
		LegalEntityID: leID,
		WarehouseID:   mainWH.ID,
		ItemID:        item.ID,
		NewUnitCost:   decimal.NewFromInt(2200),
		Reason:        "Supplier price adjustment & currency fluctuation",
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to execute revaluation: %v", err)
	}

	if !rev.OldUnitCost.Equal(decimal.NewFromInt(2000)) {
		t.Fatalf("expected old unit cost 2000, got %s", rev.OldUnitCost)
	}
	if !rev.NewUnitCost.Equal(decimal.NewFromInt(2200)) {
		t.Fatalf("expected new unit cost 2200, got %s", rev.NewUnitCost)
	}
	if !rev.VarianceAmount.Equal(decimal.NewFromInt(2000)) {
		t.Fatalf("expected variance amount 2000, got %s", rev.VarianceAmount)
	}

	// Verify balance updated
	bal, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
	})
	if !bal.AverageCost.Equal(decimal.NewFromInt(2200)) {
		t.Fatalf("expected new average cost in balance 2200, got %s", bal.AverageCost)
	}
	if !bal.TotalValue.Equal(decimal.NewFromInt(22000)) {
		t.Fatalf("expected new total value in balance 22000, got %s", bal.TotalValue)
	}

	// Verify listed in revaluations
	revs, err := invMod.InventoryService.ListRevaluations(ctx, leID, &mainWH.ID, &item.ID, 10, 0)
	if err != nil {
		t.Fatalf("failed to list revaluations: %v", err)
	}
	if len(revs) != 1 {
		t.Fatalf("expected 1 revaluation in list, got %d", len(revs))
	}
}

// 4. Valuation Report Reconciliation Invariant Test
func TestCosting_ValuationReportReconciliation(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-VAL-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Industrial Pipe 2m",
		NameAr:         "أنبوب صناعي 2 متر",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(80),
		ListPrice:      decimal.NewFromInt(120),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	doc := "INIT-VAL-01"
	_, _ = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(50),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(80),
		SourceDocType: &doc,
		CreatedBy:     userID,
	})

	report, err := invMod.InventoryService.GetValuationReport(ctx, leID, &mainWH.ID, &item.ID)
	if err != nil {
		t.Fatalf("failed to get valuation report: %v", err)
	}
	if len(report) != 1 {
		t.Fatalf("expected 1 item in valuation report, got %d", len(report))
	}

	valItem := report[0]
	expectedTotal := valItem.QuantityOnHand.Mul(valItem.AverageCost)
	if !valItem.TotalValue.Equal(expectedTotal) {
		t.Fatalf("valuation does not reconcile: onHand (%s) * avgCost (%s) != totalVal (%s)",
			valItem.QuantityOnHand, valItem.AverageCost, valItem.TotalValue)
	}
}

// 5. Zero Quantity Revaluation Rejection Invariant Test
func TestCosting_ZeroQuantityRevaluationBlocked(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-ZERO-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Zero Stock Item",
		NameAr:         "صنف برصيد صفري",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(30),
		ListPrice:      decimal.NewFromInt(50),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Revaluation on zero stock must fail with ErrZeroQuantityRevaluation
	_, err := invMod.InventoryService.ExecuteRevaluation(ctx, ports.ExecuteRevaluationRequest{
		LegalEntityID: leID,
		WarehouseID:   mainWH.ID,
		ItemID:        item.ID,
		NewUnitCost:   decimal.NewFromInt(45),
		Reason:        "Should fail",
		CreatedBy:     userID,
	})
	if err == nil {
		t.Fatal("expected error when revaluing item with zero stock, got nil")
	}
}
