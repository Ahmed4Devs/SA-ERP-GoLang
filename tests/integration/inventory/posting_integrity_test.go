package inventory_test

import (
	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"
	productdomain "erp/internal/products/domain"
	productports "erp/internal/products/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"testing"
	"time"
)

func TestCountRollbackAndMultiBinRevaluation(t *testing.T) {
	mod, products, db, entity, maker, ctx := setupInventoryTest(t)
	defer db.Close()
	svc := mod.InventoryService
	checker := inventoryChecker(t, db, ctx)
	uoms, err := products.ProductService.ListUOMs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	item, err := products.ProductService.CreateItem(ctx, productports.CreateItemRequest{LegalEntityID: entity, Code: uuid.NewString(), NameEn: "Posting item", NameAr: "صنف الترحيل", ItemType: productdomain.ItemTypeStock, TrackingType: productdomain.TrackingTypeNone, BaseUOMID: uoms[0].ID, CostPrice: decimal.NewFromInt(10)})
	if err != nil {
		t.Fatal(err)
	}
	warehouses, err := svc.ListWarehouses(ctx, entity)
	if err != nil {
		t.Fatal(err)
	}
	wh := warehouses[0].ID
	bins, err := svc.ListBins(ctx, wh)
	if err != nil {
		t.Fatal(err)
	}
	var first uuid.UUID
	for _, bin := range bins {
		if bin.IsDefault {
			first = bin.ID
		}
	}
	second, err := svc.CreateBin(ctx, domain.BinLocation{WarehouseID: wh, Code: uuid.NewString()})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []struct {
		bin            uuid.UUID
		quantity, cost int64
	}{{first, 100, 10}, {second.ID, 10, 20}} {
		if _, err := svc.PostStockMovement(ctx, ports.PostMovementRequest{LegalEntityID: entity, WarehouseID: wh, BinID: &line.bin, ItemID: item.ID, UOMID: uoms[0].ID, MovementType: domain.MovementTypeOpening, Quantity: decimal.NewFromInt(line.quantity), UnitCost: decimal.NewFromInt(line.cost), CreatedBy: maker}); err != nil {
			t.Fatal(err)
		}
	}
	stock, offset := inventoryAccount(t, db, ctx, entity, "113000"), inventoryAccount(t, db, ctx, entity, "510000")
	rev, err := svc.ExecuteRevaluation(ctx, ports.ExecuteRevaluationRequest{LegalEntityID: entity, WarehouseID: wh, ItemID: item.ID, NewUnitCost: decimal.NewFromInt(30), Reason: "Revalue all bins", CreatedBy: maker, InventoryAccountID: stock, OffsetAccountID: offset})
	if err != nil {
		t.Fatal(err)
	}
	if rev.JournalEntryID == nil || !rev.OldTotalValue.Equal(decimal.NewFromInt(1200)) || !rev.NewTotalValue.Equal(decimal.NewFromInt(3300)) || !rev.VarianceAmount.Equal(decimal.NewFromInt(2100)) {
		t.Fatalf("incorrect revaluation: %+v", rev)
	}
	var balanceValue, journalValue decimal.Decimal
	if err := db.Pool.QueryRow(ctx, "SELECT SUM(total_value) FROM inventory.stock_balances WHERE item_id=$1 AND warehouse_id=$2", item.ID, wh).Scan(&balanceValue); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(ctx, "SELECT total_debit FROM accounting.journal_entries WHERE id=$1", rev.JournalEntryID).Scan(&journalValue); err != nil {
		t.Fatal(err)
	}
	if !balanceValue.Equal(rev.NewTotalValue) || !journalValue.Equal(rev.VarianceAmount) {
		t.Fatal("bin valuation and GL disagree")
	}
	count, err := svc.CreateStockCount(ctx, ports.CreateCountRequest{LegalEntityID: entity, WarehouseID: wh, CountType: domain.CountTypeCycle, CreatedBy: maker, InventoryAccountID: stock, OffsetAccountID: offset, Lines: []ports.CreateCountLineRequest{{ItemID: item.ID, BinID: &first, CountedQuantity: decimal.NewFromInt(95)}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReconcileStockCount(ctx, count.ID, maker); err == nil {
		t.Fatal("maker reconciled own count")
	}
	var period uuid.UUID
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id=$1 AND $2 BETWEEN start_date AND end_date", entity, time.Now()).Scan(&period); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='CLOSED' WHERE id=$1", period); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReconcileStockCount(ctx, count.ID, checker); err == nil {
		t.Fatal("count posted into closed period")
	}
	if _, err := db.Pool.Exec(ctx, "UPDATE accounting.fiscal_periods SET status='OPEN' WHERE id=$1", period); err != nil {
		t.Fatal(err)
	}
	balance, err := svc.GetBalance(ctx, domain.StockKey{LegalEntityID: entity, ItemID: item.ID, WarehouseID: wh, BinID: first})
	if err != nil {
		t.Fatal(err)
	}
	if !balance.QuantityOnHand.Equal(decimal.NewFromInt(100)) {
		t.Fatal("failed count changed stock")
	}
	posted, err := svc.ReconcileStockCount(ctx, count.ID, checker)
	if err != nil {
		t.Fatal(err)
	}
	if posted.JournalEntryID == nil {
		t.Fatal("count did not create GL")
	}
	if err := db.Pool.QueryRow(ctx, "SELECT total_debit FROM accounting.journal_entries WHERE id=$1", posted.JournalEntryID).Scan(&journalValue); err != nil {
		t.Fatal(err)
	}
	if !journalValue.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("count GL should reflect actual issue cost: %s", journalValue)
	}
	if _, err := svc.ReconcileStockCount(ctx, count.ID, checker); err != nil {
		t.Fatal(err)
	}
	var movements int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM inventory.stock_movements WHERE source_doc_id=$1", count.ID).Scan(&movements); err != nil || movements != 1 {
		t.Fatalf("duplicate count movements: %d %v", movements, err)
	}
}
