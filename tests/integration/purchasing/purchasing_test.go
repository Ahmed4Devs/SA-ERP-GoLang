package purchasing_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/iam"
	"erp/internal/inventory"
	"erp/internal/organization"
	"erp/internal/payables"
	payDomain "erp/internal/payables/domain"
	payPorts "erp/internal/payables/ports"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/products"
	"erp/internal/purchasing"
	"erp/internal/purchasing/domain"
	"erp/internal/purchasing/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupPurchasingTest(t *testing.T) (*purchasing.Module, *payables.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	if err := database.MigrateUp(ctx, cfg.Database.URL, "../../../database/migrations"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	iamMod := iam.NewModule(cfg, db, txManager)
	orgMod := organization.NewModule(cfg, db, txManager)
	accMod := accounting.NewModule(cfg, db)
	prodMod := products.NewModule(cfg, db)
	invMod := inventory.NewModule(cfg, db)
	payMod := payables.NewModule(cfg, db)
	purchMod := purchasing.NewModule(cfg, db)

	// 1. Organization & Legal Entity
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// 2. Foundation modules defaults
	_ = accMod.SeedDefaultCOA(ctx, le.ID)
	_ = prodMod.SeedDefaultProductsData(ctx, le.ID)
	_ = invMod.SeedDefaultInventoryData(ctx, le.ID)
	if err := payMod.SeedDefaultPayablesData(ctx, le.ID); err != nil {
		t.Fatalf("failed to seed payables: %v", err)
	}
	_ = purchMod.SeedDefaultPurchasingData(ctx, le.ID)

	// 3. Admin user
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	}

	// 4. Sample product item, supplier, warehouse
	var itemID, supplierID, warehouseID uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT id FROM products.items WHERE legal_entity_id = $1 LIMIT 1", le.ID).Scan(&itemID)
	if err != nil {
		t.Fatalf("failed to find product item: %v", err)
	}

	err = db.Pool.QueryRow(ctx, "SELECT id FROM payables.suppliers WHERE legal_entity_id = $1 LIMIT 1", le.ID).Scan(&supplierID)
	if err != nil {
		t.Fatalf("failed to find supplier: %v", err)
	}

	err = db.Pool.QueryRow(ctx, "SELECT id FROM warehouse.warehouses WHERE legal_entity_id = $1 LIMIT 1", le.ID).Scan(&warehouseID)
	if err != nil {
		t.Fatalf("failed to find warehouse: %v", err)
	}

	return purchMod, payMod, db, le.ID, adminID, supplierID, itemID, warehouseID, ctx
}

func TestPurchaseRequest_Lifecycle(t *testing.T) {
	purchMod, _, db, leID, adminID, suppID, itemID, _, ctx := setupPurchasingTest(t)
	defer db.Close()

	svc := purchMod.PurchasingService

	// 1. Create Purchase Requisition (PR)
	cmd := ports.CreatePRCmd{
		LegalEntityID:     leID,
		RequestedByUserID: adminID,
		RequestDate:       time.Now(),
		Purpose:           strPtr("تجهيزات تشغيلية عاجلة للقسم الفني"),
		Notes:             strPtr("طلب معتمد من مدير القسم"),
		CurrencyCode:      "SAR",
		Lines: []ports.CreatePRLineCmd{
			{
				ItemID:              &itemID,
				Description:         "قواطع كهربائية 16 أمبير",
				UOM:                 "EA",
				Quantity:            decimal.NewFromInt(50),
				EstimatedUnitCost:   decimal.NewFromInt(25),
				SuggestedSupplierID: &suppID,
			},
			{
				Description:         "كابلات تمديد نحاسية معزولة",
				UOM:                 "MTR",
				Quantity:            decimal.NewFromInt(200),
				EstimatedUnitCost:   decimal.NewFromFloat(12.50),
				SuggestedSupplierID: &suppID,
			},
		},
	}

	pr, err := svc.CreatePurchaseRequest(ctx, cmd)
	if err != nil {
		t.Fatalf("expected successful PR creation, got: %v", err)
	}

	if pr.Status != domain.PRStatusDraft {
		t.Errorf("expected DRAFT status, got %s", pr.Status)
	}
	// Total: 50*25 + 200*12.50 = 1250 + 2500 = 3750
	expectedTotal := decimal.NewFromInt(3750)
	if !pr.EstimatedTotal.Equal(expectedTotal) {
		t.Errorf("expected estimated total %s, got %s", expectedTotal, pr.EstimatedTotal)
	}
	if len(pr.Lines) != 2 {
		t.Fatalf("expected 2 PR lines, got %d", len(pr.Lines))
	}

	// 2. Approve PR
	approvedPR, err := svc.ApprovePurchaseRequest(ctx, pr.ID, adminID)
	if err != nil {
		t.Fatalf("failed to approve PR: %v", err)
	}
	if approvedPR.Status != domain.PRStatusApproved {
		t.Errorf("expected APPROVED status, got %s", approvedPR.Status)
	}
}

func TestRFQ_And_SupplierQuoteAward(t *testing.T) {
	purchMod, payMod, db, leID, adminID, suppID, itemID, _, ctx := setupPurchasingTest(t)
	defer db.Close()

	svc := purchMod.PurchasingService

	// 1. Create RFQ
	rfqCmd := ports.CreateRFQCmd{
		LegalEntityID:   leID,
		Title:           "RFQ for Factory Electrical Parts",
		RFQDate:         time.Now(),
		ClosingDate:     time.Now().AddDate(0, 0, 7),
		Notes:           strPtr("تسليم مستودع الرياض الرئيسي"),
		CreatedByUserID: &adminID,
		Lines: []ports.CreateRFQLineCmd{
			{
				ItemID:          &itemID,
				Description:     "قواطع كهربائية صناعية",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(100),
				TargetUnitPrice: decPtr(decimal.NewFromInt(24)),
			},
		},
	}

	rfq, err := svc.CreateRFQ(ctx, rfqCmd)
	if err != nil {
		t.Fatalf("failed to create RFQ: %v", err)
	}
	if rfq.Status != domain.RFQStatusDraft {
		t.Errorf("expected RFQ status DRAFT, got %s", rfq.Status)
	}

	// Create second supplier for quote comparison
	validTaxID2 := "310987654300003"
	supp2, _ := payMod.SupplierService.CreateSupplier(ctx, payPorts.CreateSupplierCmd{
		LegalEntityID: leID,
		SupplierType:  payDomain.SupplierTypeLocal,
		NameEn:        "Binladen Electrical Supplies",
		NameAr:        "شركة بن لادن للتجهيزات الكهربائية",
		TaxID:         &validTaxID2,
		Address: payDomain.SaudiAddress{
			City:        "Riyadh",
			CountryCode: "SA",
		},
	})
	supp2ID := suppID
	if supp2 != nil {
		supp2ID = supp2.ID
	}

	// 2. Record Supplier Quote 1 (Award winner: 22 SAR/unit)
	quote1, err := svc.RecordSupplierQuote(ctx, ports.RecordQuoteCmd{
		LegalEntityID:          leID,
		RFQID:                  &rfq.ID,
		SupplierID:             suppID,
		SupplierQuoteReference: fmt.Sprintf("SQ-FOZAN-%s", uuid.New().String()[:8]),
		QuoteDate:              time.Now(),
		ValidUntil:             time.Now().AddDate(0, 1, 0),
		CurrencyCode:           "SAR",
		ExchangeRate:           decimal.NewFromInt(1),
		Lines: []ports.RecordQuoteLineCmd{
			{
				RFQLineID:       &rfq.Lines[0].ID,
				ItemID:          &itemID,
				Description:     "قواطع كهربائية صناعية",
				Quantity:        decimal.NewFromInt(100),
				UnitPrice:       decimal.NewFromInt(22),
				DiscountPercent: decimal.Zero,
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to record quote 1: %v", err)
	}
	// Subtotal: 100 * 22 = 2200, Tax: 330, Total: 2530
	expectedQuoteTotal := decimal.NewFromInt(2530)
	if !quote1.TotalAmount.Equal(expectedQuoteTotal) {
		t.Errorf("expected quote total %s, got %s", expectedQuoteTotal, quote1.TotalAmount)
	}

	// 3. Record Supplier Quote 2 (Higher price: 26 SAR/unit)
	_, err = svc.RecordSupplierQuote(ctx, ports.RecordQuoteCmd{
		LegalEntityID:          leID,
		RFQID:                  &rfq.ID,
		SupplierID:             supp2ID,
		SupplierQuoteReference: fmt.Sprintf("SQ-BINLADEN-%s", uuid.New().String()[:8]),
		QuoteDate:              time.Now(),
		ValidUntil:             time.Now().AddDate(0, 1, 0),
		CurrencyCode:           "SAR",
		ExchangeRate:           decimal.NewFromInt(1),
		Lines: []ports.RecordQuoteLineCmd{
			{
				RFQLineID:       &rfq.Lines[0].ID,
				ItemID:          &itemID,
				Description:     "قواطع كهربائية صناعية",
				Quantity:        decimal.NewFromInt(100),
				UnitPrice:       decimal.NewFromInt(26),
				DiscountPercent: decimal.Zero,
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to record quote 2: %v", err)
	}

	// 4. Award Supplier Quote 1
	awardedQuote, err := svc.AwardSupplierQuote(ctx, quote1.ID)
	if err != nil {
		t.Fatalf("failed to award supplier quote: %v", err)
	}
	if awardedQuote.Status != domain.QuoteStatusAccepted {
		t.Errorf("expected ACCEPTED status on awarded quote, got %s", awardedQuote.Status)
	}
}

func TestPurchaseOrder_CreationAndApproval(t *testing.T) {
	purchMod, _, db, leID, adminID, suppID, itemID, whID, ctx := setupPurchasingTest(t)
	defer db.Close()

	svc := purchMod.PurchasingService

	// 1. Create PO
	poCmd := ports.CreatePOCmd{
		LegalEntityID:        leID,
		SupplierID:           suppID,
		OrderDate:            time.Now(),
		ExpectedDeliveryDate: timePtr(time.Now().AddDate(0, 0, 10)),
		DeliveryWarehouseID:  &whID,
		CurrencyCode:         "SAR",
		ExchangeRate:         decimal.NewFromInt(1),
		Notes:                strPtr("أمر شراء عاجل ومعتمد"),
		CreatedByUserID:      &adminID,
		Lines: []ports.CreatePOLineCmd{
			{
				ItemID:          itemID,
				Description:     "مواد بناء قياسية ومستلزمات إنشائية",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(80),
				UnitPrice:       decimal.NewFromInt(50),
				DiscountPercent: decimal.NewFromInt(5), // 5% discount
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	}

	po, err := svc.CreatePurchaseOrder(ctx, poCmd)
	if err != nil {
		t.Fatalf("failed to create PO: %v", err)
	}

	// 80 * 50 = 4000 gross
	// Disc 5% = 200
	// Taxable = 3800
	// Tax 15% = 570
	// Total = 4370
	if !po.Subtotal.Equal(decimal.NewFromInt(4000)) {
		t.Errorf("expected subtotal 4000, got %s", po.Subtotal)
	}
	if !po.DiscountAmount.Equal(decimal.NewFromInt(200)) {
		t.Errorf("expected discount 200, got %s", po.DiscountAmount)
	}
	if !po.TaxableAmount.Equal(decimal.NewFromInt(3800)) {
		t.Errorf("expected taxable 3800, got %s", po.TaxableAmount)
	}
	if !po.TaxAmount.Equal(decimal.NewFromInt(570)) {
		t.Errorf("expected tax 570, got %s", po.TaxAmount)
	}
	if !po.TotalAmount.Equal(decimal.NewFromInt(4370)) {
		t.Errorf("expected total 4370, got %s", po.TotalAmount)
	}

	// 2. Negative: Issuing Draft PO directly should fail
	_, err = svc.IssuePurchaseOrder(ctx, po.ID)
	if !errors.Is(err, domain.ErrPONotApproved) {
		t.Errorf("expected ErrPONotApproved when issuing draft PO, got: %v", err)
	}

	// 3. Approve PO
	approvedPO, err := svc.ApprovePurchaseOrder(ctx, po.ID, adminID)
	if err != nil {
		t.Fatalf("failed to approve PO: %v", err)
	}
	if approvedPO.Status != domain.POStatusApproved {
		t.Errorf("expected PO status APPROVED, got %s", approvedPO.Status)
	}

	// 4. Issue PO
	issuedPO, err := svc.IssuePurchaseOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("failed to issue PO: %v", err)
	}
	if issuedPO.Status != domain.POStatusIssued {
		t.Errorf("expected PO status ISSUED, got %s", issuedPO.Status)
	}
}

func TestPurchaseOrder_RollsBackPRQuantityWhenLineInsertFails(t *testing.T) {
	purchMod, _, db, leID, adminID, suppID, itemID, whID, ctx := setupPurchasingTest(t)
	defer db.Close()

	pr, err := purchMod.PurchasingService.CreatePurchaseRequest(ctx, ports.CreatePRCmd{
		LegalEntityID:     leID,
		RequestedByUserID: adminID,
		RequestDate:       time.Now(),
		CurrencyCode:      "SAR",
		Lines: []ports.CreatePRLineCmd{{
			ItemID:            &itemID,
			Description:       "Rollback purchase request line",
			UOM:               "EA",
			Quantity:          decimal.NewFromInt(10),
			EstimatedUnitCost: decimal.NewFromInt(5),
		}},
	})
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	var poCountBefore int
	if err := db.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM purchasing.purchase_orders WHERE legal_entity_id = $1",
		leID,
	).Scan(&poCountBefore); err != nil {
		t.Fatalf("failed to count purchase orders: %v", err)
	}

	_, err = purchMod.PurchasingService.CreatePurchaseOrder(ctx, ports.CreatePOCmd{
		LegalEntityID:       leID,
		SupplierID:          suppID,
		OrderDate:           time.Now(),
		DeliveryWarehouseID: &whID,
		CurrencyCode:        "SAR",
		ExchangeRate:        decimal.NewFromInt(1),
		CreatedByUserID:     &adminID,
		Lines: []ports.CreatePOLineCmd{{
			PRLineID:        &pr.Lines[0].ID,
			ItemID:          uuid.New(), // Forces the PO line foreign-key failure.
			Description:     "Invalid item rollback test",
			UOM:             "EA",
			Quantity:        decimal.NewFromInt(3),
			UnitPrice:       decimal.NewFromInt(5),
			TaxCode:         "VAT-15",
			TaxRatePercent:  decimal.NewFromInt(15),
			DiscountPercent: decimal.Zero,
		}},
	})
	if err == nil {
		t.Fatal("expected purchase-order line foreign-key failure")
	}

	var ordered decimal.Decimal
	if err := db.Pool.QueryRow(ctx,
		"SELECT ordered_quantity FROM purchasing.purchase_request_lines WHERE id = $1",
		pr.Lines[0].ID,
	).Scan(&ordered); err != nil {
		t.Fatalf("failed to load purchase-request quantity: %v", err)
	}
	if !ordered.IsZero() {
		t.Fatalf("expected ordered quantity rollback to zero, got %s", ordered)
	}

	var poCountAfter int
	if err := db.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM purchasing.purchase_orders WHERE legal_entity_id = $1",
		leID,
	).Scan(&poCountAfter); err != nil {
		t.Fatalf("failed to recount purchase orders: %v", err)
	}
	if poCountAfter != poCountBefore {
		t.Fatalf("expected purchase-order header rollback: before=%d after=%d", poCountBefore, poCountAfter)
	}
}

func TestGoodsReceipt_InspectionAndStockIncrement(t *testing.T) {
	purchMod, _, db, leID, adminID, suppID, itemID, whID, ctx := setupPurchasingTest(t)
	defer db.Close()

	svc := purchMod.PurchasingService

	// 1. Create and issue PO for 100 units @ 50 SAR
	po, err := svc.CreatePurchaseOrder(ctx, ports.CreatePOCmd{
		LegalEntityID:       leID,
		SupplierID:          suppID,
		OrderDate:           time.Now(),
		DeliveryWarehouseID: &whID,
		CurrencyCode:        "SAR",
		ExchangeRate:        decimal.NewFromInt(1),
		CreatedByUserID:     &adminID,
		Lines: []ports.CreatePOLineCmd{
			{
				ItemID:          itemID,
				Description:     "توريد مواد خام",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(100),
				UnitPrice:       decimal.NewFromInt(50),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create PO: %v", err)
	}
	_, _ = svc.ApprovePurchaseOrder(ctx, po.ID, adminID)
	_, _ = svc.IssuePurchaseOrder(ctx, po.ID)

	// 2. Query initial stock balance for item in warehouse
	var initialQty decimal.Decimal
	_ = db.Pool.QueryRow(ctx, "SELECT COALESCE(SUM(quantity_on_hand), 0) FROM inventory.stock_balances WHERE legal_entity_id = $1 AND item_id = $2 AND warehouse_id = $3", leID, itemID, whID).Scan(&initialQty)

	// 3. Create Goods Receipt Note (GRN) for 100 units
	grnCmd := ports.CreateReceiptCmd{
		LegalEntityID:      leID,
		PurchaseOrderID:    po.ID,
		SupplierID:         suppID,
		WarehouseID:        whID,
		ReceiptDate:        time.Now(),
		DeliveryNoteNumber: strPtr("DN-RAWABI-77441"),
		Notes:              strPtr("استلام ومطابقة الفحص الفني"),
		ReceivedByUserID:   &adminID,
		Lines: []ports.CreateReceiptLineCmd{
			{
				POLineID:         po.Lines[0].ID,
				ItemID:           itemID,
				WarehouseID:      whID,
				ReceivedQuantity: decimal.NewFromInt(100),
				AcceptedQuantity: decimal.NewFromInt(100),
				RejectedQuantity: decimal.Zero,
				UnitCost:         decimal.NewFromInt(50),
			},
		},
	}

	grn, err := svc.CreateGoodsReceipt(ctx, grnCmd)
	if err != nil {
		t.Fatalf("failed to create goods receipt: %v", err)
	}
	if grn.Status != domain.GRNStatusDraft {
		t.Errorf("expected GRN status DRAFT, got %s", grn.Status)
	}

	// 4. Inspect and Accept Goods Receipt
	acceptedGRN, err := svc.InspectAndAcceptGoodsReceipt(ctx, grn.ID, adminID)
	if err != nil {
		t.Fatalf("failed to inspect and accept goods receipt: %v", err)
	}
	if acceptedGRN.Status != domain.GRNStatusAccepted {
		t.Errorf("expected GRN status ACCEPTED, got %s", acceptedGRN.Status)
	}

	// 5. Verify immutable inventory stock movement entry
	var movementCount int
	var movementType string
	var movementQty decimal.Decimal
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*), movement_type, quantity
		FROM inventory.stock_movements
		WHERE source_doc_id = $1
		GROUP BY movement_type, quantity
	`, grn.ID).Scan(&movementCount, &movementType, &movementQty)
	if err != nil {
		t.Fatalf("expected stock movement entry for GRN, got: %v", err)
	}
	if movementType != "PURCHASE_RECEIPT" {
		t.Errorf("expected movement_type PURCHASE_RECEIPT, got %s", movementType)
	}
	if !movementQty.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected positive quantity 100 in stock movement, got %s", movementQty)
	}

	// 6. Verify stock_balances increased by accepted quantity
	var finalQty decimal.Decimal
	_ = db.Pool.QueryRow(ctx, "SELECT COALESCE(SUM(quantity_on_hand), 0) FROM inventory.stock_balances WHERE legal_entity_id = $1 AND item_id = $2 AND warehouse_id = $3", leID, itemID, whID).Scan(&finalQty)

	expectedQty := initialQty.Add(decimal.NewFromInt(100))
	if !finalQty.Equal(expectedQty) {
		t.Errorf("expected final stock %s, got %s", expectedQty, finalQty)
	}

	// 7. Verify PO status transitioned to RECEIVED and line received_quantity updated
	refreshedPO, err := svc.GetPurchaseOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("failed to refresh PO: %v", err)
	}
	if refreshedPO.Status != domain.POStatusReceived {
		t.Errorf("expected PO status RECEIVED, got %s", refreshedPO.Status)
	}
	if !refreshedPO.Lines[0].ReceivedQuantity.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected PO line received quantity 100, got %s", refreshedPO.Lines[0].ReceivedQuantity)
	}
}

func TestThreeWayMatching_Verification(t *testing.T) {
	purchMod, _, db, leID, adminID, suppID, itemID, whID, ctx := setupPurchasingTest(t)
	defer db.Close()

	svc := purchMod.PurchasingService

	// Setup PO and Accepted GRN for 50 units @ 100 SAR
	po, _ := svc.CreatePurchaseOrder(ctx, ports.CreatePOCmd{
		LegalEntityID:       leID,
		SupplierID:          suppID,
		OrderDate:           time.Now(),
		DeliveryWarehouseID: &whID,
		CurrencyCode:        "SAR",
		ExchangeRate:        decimal.NewFromInt(1),
		CreatedByUserID:     &adminID,
		Lines: []ports.CreatePOLineCmd{
			{
				ItemID:          itemID,
				Description:     "قطع غيار دقيقة",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(50),
				UnitPrice:       decimal.NewFromInt(100),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	})
	_, _ = svc.ApprovePurchaseOrder(ctx, po.ID, adminID)
	_, _ = svc.IssuePurchaseOrder(ctx, po.ID)

	grn, err := svc.CreateGoodsReceipt(ctx, ports.CreateReceiptCmd{
		LegalEntityID:   leID,
		PurchaseOrderID: po.ID,
		SupplierID:      suppID,
		WarehouseID:     whID,
		ReceiptDate:     time.Now(),
		Lines: []ports.CreateReceiptLineCmd{
			{
				POLineID:         po.Lines[0].ID,
				ItemID:           itemID,
				WarehouseID:      whID,
				ReceivedQuantity: decimal.NewFromInt(50),
				AcceptedQuantity: decimal.NewFromInt(50),
				RejectedQuantity: decimal.Zero,
				UnitCost:         decimal.NewFromInt(100),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create GRN: %v", err)
	}
	acceptedGRN, err := svc.InspectAndAcceptGoodsReceipt(ctx, grn.ID, adminID)
	if err != nil {
		t.Fatalf("failed to accept GRN: %v", err)
	}

	poLineID := po.Lines[0].ID
	grnLineID := acceptedGRN.Lines[0].ID
	dummyInvLineID := uuid.New()

	// Case 1: Perfect Match (Invoiced 50 @ 100 SAR)
	res1, err := svc.VerifyThreeWayMatch(ctx, ports.ThreeWayMatchReq{
		POLineID:              poLineID,
		GRNLineID:             grnLineID,
		InvoiceLineID:         dummyInvLineID,
		InvoicedQuantity:      decimal.NewFromInt(50),
		InvoicedUnitPrice:     decimal.NewFromInt(100),
		PriceTolerancePercent: decimal.Zero,
		QtyTolerancePercent:   decimal.Zero,
	})
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if res1.MatchStatus != domain.MatchStatusPerfectMatch {
		t.Errorf("expected PERFECT_MATCH, got %s", res1.MatchStatus)
	}
	if !res1.PassedTolerance {
		t.Errorf("expected PassedTolerance = true")
	}

	// Case 2: Quantity Variance (Invoiced 60 vs Received 50)
	res2, err := svc.VerifyThreeWayMatch(ctx, ports.ThreeWayMatchReq{
		POLineID:              poLineID,
		GRNLineID:             grnLineID,
		InvoiceLineID:         dummyInvLineID,
		InvoicedQuantity:      decimal.NewFromInt(60),
		InvoicedUnitPrice:     decimal.NewFromInt(100),
		PriceTolerancePercent: decimal.Zero,
		QtyTolerancePercent:   decimal.Zero,
	})
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if res2.MatchStatus != domain.MatchStatusQuantityVariance {
		t.Errorf("expected QUANTITY_VARIANCE, got %s", res2.MatchStatus)
	}
	if res2.PassedTolerance {
		t.Errorf("expected PassedTolerance = false for 20%% quantity variance without tolerance")
	}

	// Case 3: Price Variance (Invoiced 105 SAR vs PO 100 SAR)
	res3, err := svc.VerifyThreeWayMatch(ctx, ports.ThreeWayMatchReq{
		POLineID:              poLineID,
		GRNLineID:             grnLineID,
		InvoiceLineID:         dummyInvLineID,
		InvoicedQuantity:      decimal.NewFromInt(50),
		InvoicedUnitPrice:     decimal.NewFromInt(105),
		PriceTolerancePercent: decimal.Zero,
		QtyTolerancePercent:   decimal.Zero,
	})
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if res3.MatchStatus != domain.MatchStatusPriceVariance {
		t.Errorf("expected PRICE_VARIANCE, got %s", res3.MatchStatus)
	}

	// Case 4: Price Tolerance Acceptance (Invoiced 101 SAR with 2% tolerance allowed)
	res4, err := svc.VerifyThreeWayMatch(ctx, ports.ThreeWayMatchReq{
		POLineID:              poLineID,
		GRNLineID:             grnLineID,
		InvoiceLineID:         dummyInvLineID,
		InvoicedQuantity:      decimal.NewFromInt(50),
		InvoicedUnitPrice:     decimal.NewFromInt(101),
		PriceTolerancePercent: decimal.NewFromInt(2), // 2% tolerance allows up to 102
		QtyTolerancePercent:   decimal.Zero,
	})
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if !res4.PassedTolerance {
		t.Errorf("expected PassedTolerance = true within 2%% tolerance")
	}
}

// Helpers
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func decPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
