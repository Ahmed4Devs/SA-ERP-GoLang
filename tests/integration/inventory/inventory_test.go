package inventory_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"erp/internal/iam"
	"erp/internal/inventory"
	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/products"
	prodDomain "erp/internal/products/domain"
	prodPorts "erp/internal/products/ports"
	"erp/internal/tax"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupInventoryTest(t *testing.T) (*inventory.Module, *products.Module, *database.DB, uuid.UUID, uuid.UUID, context.Context) {
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
	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	taxMod := tax.NewModule(cfg, db)
	prodMod := products.NewModule(cfg, db)
	invMod := inventory.NewModule(cfg, db)

	// Ensure organization & legal entity exist
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// Seed tax, products, and inventory defaults
	_ = taxMod.SeedDefaultTaxData(ctx, le.ID)
	_ = prodMod.SeedDefaultProductsData(ctx, le.ID)
	if err := invMod.SeedDefaultInventoryData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed inventory defaults: %v", err)
	}

	// Get admin user
	adminUser, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var userID uuid.UUID
	if err != nil || adminUser == nil {
		user, err := iamMod.AuthService.SeedAdmin(ctx, "inv_admin_"+uuid.NewString()[:6], "inv_admin@erp.sa", "Inventory Admin", "Password@123")
		if err != nil {
			t.Fatalf("failed to seed admin user: %v", err)
		}
		userID = user.ID
	} else {
		userID = adminUser.ID
	}

	return invMod, prodMod, db, le.ID, userID, ctx
}

// 1. Warehouses & Bins
func TestInventory_WarehouseAndBins(t *testing.T) {
	invMod, _, _, leID, _, ctx := setupInventoryTest(t)

	whs, err := invMod.InventoryService.ListWarehouses(ctx, leID)
	if err != nil {
		t.Fatalf("failed to list warehouses: %v", err)
	}
	if len(whs) < 2 {
		t.Fatalf("expected at least 2 seeded warehouses (WH-MAIN & WH-TRANSIT), got %d", len(whs))
	}

	var mainWH *domain.Warehouse
	for i := range whs {
		if whs[i].Code == "WH-MAIN" {
			mainWH = &whs[i]
			break
		}
	}
	if mainWH == nil {
		t.Fatal("WH-MAIN warehouse not found")
	}

	// List Bins for WH-MAIN
	bins, err := invMod.InventoryService.ListBins(ctx, mainWH.ID)
	if err != nil || len(bins) == 0 {
		t.Fatalf("expected at least default bin for WH-MAIN, got: %v", err)
	}
	if !bins[0].IsDefault {
		t.Errorf("expected bin to be default system bin")
	}

	// Create New Regional Warehouse: WH-DMM
	dmmWH, err := invMod.InventoryService.CreateWarehouse(ctx, domain.Warehouse{
		LegalEntityID: leID,
		Code:          "WH-DMM-" + strings.ToUpper(uuid.NewString()[:4]),
		NameEn:        "Dammam Regional Distribution Center",
		NameAr:        "مركز توزيع الدمام الإقليمي",
		WarehouseType: domain.WarehouseTypePhysical,
		City:          "Dammam",
		CountryCode:   "SA",
	})
	if err != nil {
		t.Fatalf("failed to create regional warehouse: %v", err)
	}

	// Verify auto-created default bin for new warehouse
	dmmBins, err := invMod.InventoryService.ListBins(ctx, dmmWH.ID)
	if err != nil || len(dmmBins) == 0 {
		t.Fatalf("expected auto-created DEFAULT bin for new warehouse, got: %v", err)
	}
}

// 2. Immutable Movements & Moving Average Cost Balance
func TestInventory_ImmutableMovementsAndBalanceDelta(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	// Create test item
	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	testItem, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-COST-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Industrial Air Filter",
		NameAr:         "فلتر هواء صناعي",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(100),
		ListPrice:      decimal.NewFromInt(160),
		IsSalesEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create test item: %v", err)
	}

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// 1. Post Opening Movement: 10 units @ 100 SAR = 1000 SAR
	mov1, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        testItem.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(100),
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post opening movement: %v", err)
	}
	if mov1.Quantity.Cmp(decimal.NewFromInt(10)) != 0 {
		t.Fatalf("expected quantity 10, got: %s", mov1.Quantity.String())
	}

	// Verify projected balance
	bal1, err := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        testItem.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mov1.BinID,
	})
	if err != nil || bal1 == nil {
		t.Fatalf("failed to get balance: %v", err)
	}
	if !bal1.QuantityOnHand.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected on hand 10, got: %s", bal1.QuantityOnHand.String())
	}
	if !bal1.AverageCost.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected avg cost 100, got: %s", bal1.AverageCost.String())
	}
	if !bal1.TotalValue.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("expected total value 1000, got: %s", bal1.TotalValue.String())
	}

	// 2. Post Purchase Receipt: 10 units @ 200 SAR = 2000 SAR
	// Resulting balance: 20 units, Total value: 3000 SAR, Average cost: 150 SAR!
	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypePurchaseReceipt,
		ItemID:        testItem.ID,
		WarehouseID:   mainWH.ID,
		BinID:         &mov1.BinID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(200),
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post purchase receipt: %v", err)
	}

	bal2, err := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        testItem.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mov1.BinID,
	})
	if err != nil || bal2 == nil {
		t.Fatalf("failed to get updated balance: %v", err)
	}
	if !bal2.QuantityOnHand.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("expected on hand 20, got: %s", bal2.QuantityOnHand.String())
	}
	if !bal2.AverageCost.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("expected moving average cost 150, got: %s", bal2.AverageCost.String())
	}
	if !bal2.TotalValue.Equal(decimal.NewFromInt(3000)) {
		t.Fatalf("expected total value 3000, got: %s", bal2.TotalValue.String())
	}
}

// 3. Pessimistic Lock & Insufficient Stock Invariant
func TestInventory_PessimisticLockAndInsufficientStock(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-LOCK-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Precision Gear",
		NameAr:         "ترس ميكانيكي دقيق",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(50),
		ListPrice:      decimal.NewFromInt(80),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Add 5 units
	mov, _ := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(5),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(50),
		CreatedBy:     userID,
	})

	// Try to issue 10 units -> must fail with ErrInsufficientStock
	_, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeSalesIssue,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         &mov.BinID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		CreatedBy:     userID,
	})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got: %v", err)
	}

	// Valid issue of 2 units
	_, validErr := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeSalesIssue,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         &mov.BinID,
		Quantity:      decimal.NewFromInt(2),
		UOMID:         pcsUOM.ID,
		CreatedBy:     userID,
	})
	if validErr != nil {
		t.Fatalf("valid sales issue failed: %v", validErr)
	}

	// Verify remaining balance: 3 units
	bal, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mov.BinID,
	})
	if !bal.QuantityOnHand.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("expected 3 remaining units, got: %s", bal.QuantityOnHand.String())
	}
}

// 4. Stock Reservations & Availability Holds
func TestInventory_StockReservations(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-RES-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Hydraulic Valve",
		NameAr:         "صمام هيدروليكي",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(300),
		ListPrice:      decimal.NewFromInt(450),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Add 10 units
	mov, _ := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(300),
		CreatedBy:     userID,
	})

	// Reserve 4 units for Sales Order
	soID := uuid.New()
	res, err := invMod.InventoryService.ReserveStock(ctx, ports.ReserveStockRequest{
		LegalEntityID:    leID,
		ItemID:           item.ID,
		WarehouseID:      mainWH.ID,
		BinID:            &mov.BinID,
		Quantity:         decimal.NewFromInt(4),
		DemandSourceType: "SALES_ORDER",
		DemandSourceID:   soID,
	})
	if err != nil {
		t.Fatalf("failed to reserve stock: %v", err)
	}

	// Verify balance projection: on_hand=10, reserved=4, available=6
	balAfterRes, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mov.BinID,
	})
	if !balAfterRes.QuantityOnHand.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected on_hand 10, got: %s", balAfterRes.QuantityOnHand.String())
	}
	if !balAfterRes.QuantityReserved.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("expected reserved 4, got: %s", balAfterRes.QuantityReserved.String())
	}
	if !balAfterRes.QuantityAvailable.Equal(decimal.NewFromInt(6)) {
		t.Fatalf("expected available 6, got: %s", balAfterRes.QuantityAvailable.String())
	}

	// Attempting to reserve 8 units fails because only 6 are available!
	_, overResErr := invMod.InventoryService.ReserveStock(ctx, ports.ReserveStockRequest{
		LegalEntityID:    leID,
		ItemID:           item.ID,
		WarehouseID:      mainWH.ID,
		BinID:            &mov.BinID,
		Quantity:         decimal.NewFromInt(8),
		DemandSourceType: "SALES_ORDER",
		DemandSourceID:   uuid.New(),
	})
	if !errors.Is(overResErr, domain.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock for over-reservation, got: %v", overResErr)
	}

	// Release reservation
	if err := invMod.InventoryService.ReleaseStock(ctx, res.ID); err != nil {
		t.Fatalf("failed to release reservation: %v", err)
	}

	// Balance restored to available=10, reserved=0
	balRestored, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mov.BinID,
	})
	if !balRestored.QuantityAvailable.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected available restored to 10, got: %s", balRestored.QuantityAvailable.String())
	}
}

// 5. Transfer Order Dispatch & Receipt Workflow
func TestInventory_TransferOrderWorkflow(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-TRF-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Industrial Conveyor Roller",
		NameAr:         "أسطوانة سير نقل صناعي",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(80),
		ListPrice:      decimal.NewFromInt(130),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	var mainWH, transitWH *domain.Warehouse
	for i := range whs {
		if whs[i].Code == "WH-MAIN" {
			mainWH = &whs[i]
		} else if whs[i].Code == "WH-TRANSIT" {
			transitWH = &whs[i]
		}
	}
	if mainWH == nil || transitWH == nil {
		t.Fatal("WH-MAIN or WH-TRANSIT warehouse not found")
	}

	// Initial stock at WH-MAIN: 20 units
	_, _ = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(20),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(80),
		CreatedBy:     userID,
	})

	// 1. Create Transfer Order for 8 units
	order, err := invMod.InventoryService.CreateTransferOrder(ctx, ports.CreateTransferRequest{
		LegalEntityID:   leID,
		FromWarehouseID: mainWH.ID,
		ToWarehouseID:   transitWH.ID,
		Notes:           "Inter-warehouse replenishment",
		CreatedBy:       userID,
		Lines: []ports.CreateTransferLineRequest{
			{
				ItemID:            item.ID,
				RequestedQuantity: decimal.NewFromInt(8),
				UOMID:             pcsUOM.ID,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create transfer order: %v", err)
	}
	if order.Status != domain.TransferStatusDraft {
		t.Fatalf("expected DRAFT status, got %s", order.Status)
	}

	// 2. Dispatch Transfer Order
	dispatchedOrder, err := invMod.InventoryService.DispatchTransfer(ctx, order.ID, userID)
	if err != nil {
		t.Fatalf("failed to dispatch transfer order: %v", err)
	}
	if dispatchedOrder.Status != domain.TransferStatusDispatched {
		t.Fatalf("expected DISPATCHED status, got %s", dispatchedOrder.Status)
	}
	if len(dispatchedOrder.Lines) != 1 || !dispatchedOrder.Lines[0].DispatchedQuantity.Equal(decimal.NewFromInt(8)) {
		t.Fatalf("expected dispatched line quantity 8, got %+v", dispatchedOrder.Lines)
	}

	// Verify WH-MAIN on-hand dropped from 20 to 12
	mainBins, _ := invMod.InventoryService.ListBins(ctx, mainWH.ID)
	balMain, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mainBins[0].ID,
	})
	if !balMain.QuantityOnHand.Equal(decimal.NewFromInt(12)) {
		t.Fatalf("expected WH-MAIN on_hand 12, got: %s", balMain.QuantityOnHand.String())
	}

	// 3. Receive Transfer Order
	receivedOrder, err := invMod.InventoryService.ReceiveTransfer(ctx, order.ID, userID)
	if err != nil {
		t.Fatalf("failed to receive transfer order: %v", err)
	}
	if receivedOrder.Status != domain.TransferStatusReceived {
		t.Fatalf("expected RECEIVED status, got %s", receivedOrder.Status)
	}
	if len(receivedOrder.Lines) != 1 || !receivedOrder.Lines[0].ReceivedQuantity.Equal(decimal.NewFromInt(8)) {
		t.Fatalf("expected received line quantity 8, got %+v", receivedOrder.Lines)
	}

	// Verify WH-TRANSIT on-hand increased by 8
	transitBins, _ := invMod.InventoryService.ListBins(ctx, transitWH.ID)
	balTransit, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   transitWH.ID,
		BinID:         transitBins[0].ID,
	})
	if balTransit == nil || !balTransit.QuantityOnHand.Equal(decimal.NewFromInt(8)) {
		t.Fatalf("expected WH-TRANSIT on_hand 8, got: %v", balTransit)
	}
}

// 6. Stock Adjustments
func TestInventory_StockAdjustment(t *testing.T) {
	invMod, prodMod, _, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-ADJ-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Safety Helmet Pro",
		NameAr:         "خوذة حماية صناعية",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(45),
		ListPrice:      decimal.NewFromInt(70),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	// Create Stock Adjustment: Found 5 extra helmets
	adj, err := invMod.InventoryService.CreateAdjustment(ctx, ports.CreateAdjustmentRequest{
		LegalEntityID:  leID,
		WarehouseID:    mainWH.ID,
		AdjustmentType: domain.AdjustmentTypeFound,
		ReasonCode:     "FOUND_IN_CYCLE_COUNT",
		Notes:          "Physical count discovered untracked stock",
		CreatedBy:      userID,
		Lines: []ports.CreateAdjustmentLineRequest{
			{
				ItemID:   item.ID,
				Quantity: decimal.NewFromInt(5),
				UnitCost: decimal.NewFromInt(45),
				Reason:   "Unrecorded replenishment",
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create adjustment: %v", err)
	}

	// Post Adjustment
	postedAdj, err := invMod.InventoryService.PostAdjustment(ctx, adj.ID, userID)
	if err != nil {
		t.Fatalf("failed to post adjustment: %v", err)
	}
	if postedAdj.Status != domain.AdjustmentStatusPosted {
		t.Fatalf("expected POSTED status, got: %s", postedAdj.Status)
	}

	// Verify balance incremented by 5
	mainBins, _ := invMod.InventoryService.ListBins(ctx, mainWH.ID)
	bal, _ := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		BinID:         mainBins[0].ID,
	})
	if !bal.QuantityOnHand.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("expected on_hand 5, got: %s", bal.QuantityOnHand.String())
	}
}

// 7. Immutability of Posted Movements
func TestInventory_MovementImmutability(t *testing.T) {
	invMod, prodMod, db, leID, userID, ctx := setupInventoryTest(t)

	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	pcsUOM := uoms[0]

	item, _ := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID:  leID,
		Code:           "ITEM-IMM-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:         "Immutable Stock Test",
		NameAr:         "صنف اختبار حظر التعديل",
		ItemType:       prodDomain.ItemTypeStock,
		TrackingType:   prodDomain.TrackingTypeNone,
		BaseUOMID:      pcsUOM.ID,
		CostPrice:      decimal.NewFromInt(10),
		ListPrice:      decimal.NewFromInt(20),
		IsSalesEnabled: true,
	})

	whs, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	mainWH := whs[0]

	mov, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   mainWH.ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         pcsUOM.ID,
		UnitCost:      decimal.NewFromInt(10),
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to post movement: %v", err)
	}

	// Attempting SQL UPDATE directly on inventory.stock_movements MUST fail
	_, updateErr := db.Pool.Exec(ctx, "UPDATE inventory.stock_movements SET quantity = 999 WHERE id = $1", mov.ID)
	if updateErr == nil {
		t.Fatal("expected update to fail on immutable stock_movements, but it succeeded")
	}
	if !strings.Contains(updateErr.Error(), "immutable audit records") {
		t.Fatalf("expected immutability error message, got: %v", updateErr)
	}

	// Attempting SQL DELETE directly on inventory.stock_movements MUST fail
	_, deleteErr := db.Pool.Exec(ctx, "DELETE FROM inventory.stock_movements WHERE id = $1", mov.ID)
	if deleteErr == nil {
		t.Fatal("expected delete to fail on immutable stock_movements, but it succeeded")
	}
	if !strings.Contains(deleteErr.Error(), "immutable audit records") {
		t.Fatalf("expected immutability error message, got: %v", deleteErr)
	}
}

func TestInventory_WarehouseCreationRollsBackWhenDefaultBinFails(t *testing.T) {
	invMod, _, db, leID, _, ctx := setupInventoryTest(t)
	code := "WH-ROLLBACK-" + strings.ToUpper(uuid.NewString()[:6])

	_, err := db.Pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION warehouse.fail_test_default_bin()
		RETURNS trigger AS $$
		BEGIN
			IF NEW.code = 'DEFAULT' THEN
				RAISE EXCEPTION 'forced default bin failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		DROP TRIGGER IF EXISTS trg_fail_test_default_bin ON warehouse.bin_locations;
		CREATE TRIGGER trg_fail_test_default_bin
		BEFORE INSERT ON warehouse.bin_locations
		FOR EACH ROW EXECUTE FUNCTION warehouse.fail_test_default_bin();
	`)
	if err != nil {
		t.Fatalf("failed to install rollback test trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DROP TRIGGER IF EXISTS trg_fail_test_default_bin ON warehouse.bin_locations;
			DROP FUNCTION IF EXISTS warehouse.fail_test_default_bin();
		`)
	})

	_, err = invMod.InventoryService.CreateWarehouse(ctx, domain.Warehouse{
		LegalEntityID: leID,
		Code:          code,
		NameEn:        "Rollback warehouse",
		NameAr:        "مستودع اختبار التراجع",
		WarehouseType: domain.WarehouseTypePhysical,
		CountryCode:   "SA",
	})
	if err == nil {
		t.Fatal("expected forced default-bin failure")
	}

	var count int
	if err := db.Pool.QueryRow(ctx,
		"SELECT count(*) FROM warehouse.warehouses WHERE legal_entity_id = $1 AND code = $2", leID, code,
	).Scan(&count); err != nil {
		t.Fatalf("failed to verify warehouse rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected warehouse header rollback, found %d rows", count)
	}
}

func TestInventory_MovementRollsBackWhenCostLayerFails(t *testing.T) {
	invMod, prodMod, db, leID, userID, ctx := setupInventoryTest(t)
	uoms, err := prodMod.ProductService.ListUOMs(ctx)
	if err != nil || len(uoms) == 0 {
		t.Fatalf("failed to load UOMs: %v", err)
	}
	item, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID: leID,
		Code:          "ITEM-COST-RB-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:        "Cost rollback item",
		NameAr:        "صنف اختبار تراجع التكلفة",
		ItemType:      prodDomain.ItemTypeStock,
		TrackingType:  prodDomain.TrackingTypeNone,
		BaseUOMID:     uoms[0].ID,
		CostPrice:     decimal.NewFromInt(25),
	})
	if err != nil {
		t.Fatalf("failed to create test item: %v", err)
	}
	warehouses, err := invMod.InventoryService.ListWarehouses(ctx, leID)
	if err != nil || len(warehouses) == 0 {
		t.Fatalf("failed to load warehouses: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION inventory.fail_test_cost_layer()
		RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced cost layer failure';
		END;
		$$ LANGUAGE plpgsql;
		DROP TRIGGER IF EXISTS trg_fail_test_cost_layer ON inventory.cost_layers;
		CREATE TRIGGER trg_fail_test_cost_layer
		BEFORE INSERT ON inventory.cost_layers
		FOR EACH ROW EXECUTE FUNCTION inventory.fail_test_cost_layer();
	`)
	if err != nil {
		t.Fatalf("failed to install rollback test trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DROP TRIGGER IF EXISTS trg_fail_test_cost_layer ON inventory.cost_layers;
			DROP FUNCTION IF EXISTS inventory.fail_test_cost_layer();
		`)
	})

	_, err = invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypePurchaseReceipt,
		ItemID:        item.ID,
		WarehouseID:   warehouses[0].ID,
		Quantity:      decimal.NewFromInt(3),
		UOMID:         uoms[0].ID,
		UnitCost:      decimal.NewFromInt(25),
		CreatedBy:     userID,
	})
	if err == nil {
		t.Fatal("expected forced cost-layer failure")
	}

	var movements, balances int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM inventory.stock_movements WHERE item_id = $1", item.ID).Scan(&movements); err != nil {
		t.Fatalf("failed to verify movement rollback: %v", err)
	}
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM inventory.stock_balances WHERE item_id = $1", item.ID).Scan(&balances); err != nil {
		t.Fatalf("failed to verify balance rollback: %v", err)
	}
	if movements != 0 || balances != 0 {
		t.Fatalf("expected movement and balance rollback, got movements=%d balances=%d", movements, balances)
	}
}

func TestInventory_ReservationRollsBackWhenRecordInsertFails(t *testing.T) {
	invMod, prodMod, db, leID, userID, ctx := setupInventoryTest(t)
	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	item, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
		LegalEntityID: leID,
		Code:          "ITEM-RES-RB-" + strings.ToUpper(uuid.NewString()[:6]),
		NameEn:        "Reservation rollback item",
		NameAr:        "صنف اختبار تراجع الحجز",
		ItemType:      prodDomain.ItemTypeStock,
		TrackingType:  prodDomain.TrackingTypeNone,
		BaseUOMID:     uoms[0].ID,
		CostPrice:     decimal.NewFromInt(10),
	})
	if err != nil {
		t.Fatalf("failed to create test item: %v", err)
	}
	warehouses, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	movement, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        item.ID,
		WarehouseID:   warehouses[0].ID,
		Quantity:      decimal.NewFromInt(10),
		UOMID:         uoms[0].ID,
		UnitCost:      decimal.NewFromInt(10),
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to seed stock: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION inventory.fail_test_reservation_insert()
		RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced reservation failure';
		END;
		$$ LANGUAGE plpgsql;
		DROP TRIGGER IF EXISTS trg_fail_test_reservation_insert ON inventory.stock_reservations;
		CREATE TRIGGER trg_fail_test_reservation_insert
		BEFORE INSERT ON inventory.stock_reservations
		FOR EACH ROW EXECUTE FUNCTION inventory.fail_test_reservation_insert();
	`)
	if err != nil {
		t.Fatalf("failed to install rollback test trigger: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DROP TRIGGER IF EXISTS trg_fail_test_reservation_insert ON inventory.stock_reservations;
			DROP FUNCTION IF EXISTS inventory.fail_test_reservation_insert();
		`)
	})

	_, err = invMod.InventoryService.ReserveStock(ctx, ports.ReserveStockRequest{
		LegalEntityID:    leID,
		ItemID:           item.ID,
		WarehouseID:      warehouses[0].ID,
		BinID:            &movement.BinID,
		Quantity:         decimal.NewFromInt(4),
		DemandSourceType: "SALES_ORDER",
		DemandSourceID:   uuid.New(),
	})
	if err == nil {
		t.Fatal("expected forced reservation failure")
	}

	bal, err := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        item.ID,
		WarehouseID:   warehouses[0].ID,
		BinID:         movement.BinID,
	})
	if err != nil {
		t.Fatalf("failed to reload balance: %v", err)
	}
	if !bal.QuantityReserved.IsZero() || !bal.QuantityAvailable.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected reservation projection rollback, got reserved=%s available=%s", bal.QuantityReserved, bal.QuantityAvailable)
	}
}

func TestInventory_TransferDispatchRollsBackAllLines(t *testing.T) {
	invMod, prodMod, db, leID, userID, ctx := setupInventoryTest(t)
	uoms, _ := prodMod.ProductService.ListUOMs(ctx)
	createItem := func(code string) *prodDomain.Item {
		item, err := prodMod.ProductService.CreateItem(ctx, prodPorts.CreateItemRequest{
			LegalEntityID: leID,
			Code:          code + strings.ToUpper(uuid.NewString()[:6]),
			NameEn:        "Transfer rollback item",
			NameAr:        "صنف اختبار تراجع التحويل",
			ItemType:      prodDomain.ItemTypeStock,
			TrackingType:  prodDomain.TrackingTypeNone,
			BaseUOMID:     uoms[0].ID,
			CostPrice:     decimal.NewFromInt(15),
		})
		if err != nil {
			t.Fatalf("failed to create transfer item: %v", err)
		}
		return item
	}
	first := createItem("ITEM-TRF-RB-A-")
	second := createItem("ITEM-TRF-RB-B-")
	warehouses, _ := invMod.InventoryService.ListWarehouses(ctx, leID)
	if len(warehouses) < 2 {
		t.Fatal("expected at least two warehouses")
	}
	opening, err := invMod.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{
		LegalEntityID: leID,
		MovementType:  domain.MovementTypeOpening,
		ItemID:        first.ID,
		WarehouseID:   warehouses[0].ID,
		Quantity:      decimal.NewFromInt(5),
		UOMID:         uoms[0].ID,
		UnitCost:      decimal.NewFromInt(15),
		CreatedBy:     userID,
	})
	if err != nil {
		t.Fatalf("failed to seed first item stock: %v", err)
	}
	order, err := invMod.InventoryService.CreateTransferOrder(ctx, ports.CreateTransferRequest{
		LegalEntityID:   leID,
		FromWarehouseID: warehouses[0].ID,
		ToWarehouseID:   warehouses[1].ID,
		CreatedBy:       userID,
		Lines: []ports.CreateTransferLineRequest{
			{ItemID: first.ID, RequestedQuantity: decimal.NewFromInt(2), UOMID: uoms[0].ID},
			{ItemID: second.ID, RequestedQuantity: decimal.NewFromInt(2), UOMID: uoms[0].ID},
		},
	})
	if err != nil {
		t.Fatalf("failed to create transfer: %v", err)
	}
	if _, err := invMod.InventoryService.DispatchTransfer(ctx, order.ID, userID); !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("expected insufficient stock on second line, got %v", err)
	}

	bal, err := invMod.InventoryService.GetBalance(ctx, domain.StockKey{
		LegalEntityID: leID,
		ItemID:        first.ID,
		WarehouseID:   warehouses[0].ID,
		BinID:         opening.BinID,
	})
	if err != nil || !bal.QuantityOnHand.Equal(decimal.NewFromInt(5)) {
		t.Fatalf("expected first line balance rollback to 5, got balance=%v err=%v", bal, err)
	}
	reloaded, err := invMod.InventoryService.GetTransferOrder(ctx, order.ID)
	if err != nil || reloaded.Status != domain.TransferStatusDraft {
		t.Fatalf("expected transfer status rollback to DRAFT, got order=%v err=%v", reloaded, err)
	}
	var movements int
	if err := db.Pool.QueryRow(ctx,
		"SELECT count(*) FROM inventory.stock_movements WHERE source_doc_id = $1", order.ID,
	).Scan(&movements); err != nil {
		t.Fatalf("failed to count transfer movements: %v", err)
	}
	if movements != 0 {
		t.Fatalf("expected all transfer movements rolled back, found %d", movements)
	}
}
