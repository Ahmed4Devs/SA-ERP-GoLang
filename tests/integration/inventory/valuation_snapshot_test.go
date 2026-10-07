package inventory_test

import (
	"encoding/json"
	"sync"
	"testing"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"
	productdomain "erp/internal/products/domain"
	productports "erp/internal/products/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestValuationSnapshot_PrecisionConcurrentBinsAndImmutableHistory(t *testing.T) {
	inv, products, db, entity, actor, ctx := setupInventoryTest(t)
	defer db.Close()
	uoms, err := products.ProductService.ListUOMs(ctx)
	if err != nil || len(uoms) == 0 {
		t.Fatal(err)
	}
	item, err := products.ProductService.CreateItem(ctx, productports.CreateItemRequest{LegalEntityID: entity, Code: "SNAP-" + uuid.NewString(), NameEn: "Snapshot item", NameAr: "صنف", ItemType: productdomain.ItemTypeStock, TrackingType: productdomain.TrackingTypeNone, BaseUOMID: uoms[0].ID, CostPrice: decimal.RequireFromString("12.123456")})
	if err != nil {
		t.Fatal(err)
	}
	warehouses, err := inv.InventoryService.ListWarehouses(ctx, entity)
	if err != nil || len(warehouses) == 0 {
		t.Fatal(err)
	}
	wh := warehouses[0]
	bin, err := inv.InventoryService.CreateBin(ctx, domain.BinLocation{WarehouseID: wh.ID, Code: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	first, err := inv.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{LegalEntityID: entity, ItemID: item.ID, WarehouseID: wh.ID, UOMID: uoms[0].ID, MovementType: domain.MovementTypePurchaseReceipt, Quantity: decimal.NewFromInt(1), UnitCost: decimal.RequireFromString("12.123456"), CreatedBy: actor})
	if err != nil {
		t.Fatal(err)
	}
	_, err = inv.InventoryService.PostStockMovement(ctx, ports.PostMovementRequest{LegalEntityID: entity, ItemID: item.ID, WarehouseID: wh.ID, BinID: &bin.ID, UOMID: uoms[0].ID, MovementType: domain.MovementTypePurchaseReceipt, Quantity: decimal.NewFromInt(2), UnitCost: decimal.RequireFromString("12.123456"), CreatedBy: actor})
	if err != nil {
		t.Fatal(err)
	}
	var quantity, value string
	if err = db.Pool.QueryRow(ctx, `SELECT quantity_on_hand::text,total_value::text FROM inventory.item_costs WHERE legal_entity_id=$1 AND item_id=$2 AND warehouse_id=$3`, entity, item.ID, wh.ID).Scan(&quantity, &value); err != nil || quantity != "3.000000" || value != "36.370368" {
		t.Fatalf("projection %s %s: %v", quantity, value, err)
	}
	batch, count, err := inv.ValuationService.CreateSnapshot(ctx, entity, actor)
	if err != nil || count < 1 {
		t.Fatalf("snapshot %d: %v", count, err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, binID := range []uuid.UUID{first.BinID, bin.ID} {
		wg.Add(1)
		go func(binID uuid.UUID) {
			defer wg.Done()
			_, err := db.Pool.Exec(ctx, `UPDATE inventory.stock_balances SET quantity_on_hand=quantity_on_hand+1,total_value=total_value+12.123456 WHERE legal_entity_id=$1 AND item_id=$2 AND bin_id=$3`, entity, item.ID, binID)
			errs <- err
		}(binID)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err = db.Pool.QueryRow(ctx, `SELECT quantity_on_hand::text,total_value::text FROM inventory.item_costs WHERE legal_entity_id=$1 AND item_id=$2 AND warehouse_id=$3`, entity, item.ID, wh.ID).Scan(&quantity, &value); err != nil || quantity != "5.000000" || value != "60.617280" {
		t.Fatalf("concurrent projection %s %s: %v", quantity, value, err)
	}
	if err = db.Pool.QueryRow(ctx, `SELECT quantity_on_hand::text,total_value::text FROM inventory.valuation_snapshots WHERE snapshot_batch_id=$1 AND item_id=$2`, batch, item.ID).Scan(&quantity, &value); err != nil || quantity != "3.000000" || value != "36.370368" {
		t.Fatalf("snapshot changed: %s %s %v", quantity, value, err)
	}
	if _, err = db.Pool.Exec(ctx, `UPDATE inventory.valuation_snapshots SET total_value=0 WHERE snapshot_batch_id=$1`, batch); err == nil {
		t.Fatal("snapshot mutable")
	}
	data, total, err := inv.ValuationService.List(ctx, entity, true, &batch, 1, 0)
	if err != nil || total != count {
		t.Fatalf("snapshot page: %d %v", total, err)
	}
	var rows []map[string]any
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 1 {
		t.Fatalf("snapshot data %s: %v", data, err)
	}
	_, total, err = inv.ValuationService.List(ctx, uuid.New(), true, &batch, 50, 0)
	if err != nil || total != 0 {
		t.Fatal("valuation snapshot crossed entity boundary")
	}
}
