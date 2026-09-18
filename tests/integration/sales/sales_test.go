package sales_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"erp/internal/accounting"
	"erp/internal/iam"
	"erp/internal/inventory"
	"erp/internal/organization"
	"erp/internal/platform/config"
	"erp/internal/platform/database"
	"erp/internal/platform/transaction"
	"erp/internal/products"
	"erp/internal/receivables"
	recDomain "erp/internal/receivables/domain"
	recPorts "erp/internal/receivables/ports"
	"erp/internal/sales"
	"erp/internal/sales/domain"
	"erp/internal/sales/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func setupSalesTest(t *testing.T) (*sales.Module, *receivables.Module, *database.DB, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, context.Context) {
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
	accMod := accounting.NewModule(cfg, db)
	prodMod := products.NewModule(cfg, db)
	invMod := inventory.NewModule(cfg, db)
	recMod := receivables.NewModule(cfg, db)
	salesMod := sales.NewModule(cfg, db)

	// Ensure organization & legal entity exist
	le, err := orgMod.SeedDefaultStructure(ctx)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	// Seed accounting, products, inventory, receivables defaults
	_ = accMod.SeedDefaultCOA(ctx, le.ID)
	_ = prodMod.SeedDefaultProductsData(ctx, le.ID)
	_ = invMod.SeedDefaultInventoryData(ctx, le.ID)
	_ = recMod.SeedDefaultReceivablesData(ctx, le.ID)

	// Get admin user
	admin, err := iamMod.AuthService.GetUserByIdentifier(ctx, "admin")
	var adminID uuid.UUID
	if err == nil && admin != nil {
		adminID = admin.ID
	}

	// Lookup a sample product item and warehouse
	var itemID, warehouseID uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT id FROM products.items WHERE legal_entity_id = $1 LIMIT 1", le.ID).Scan(&itemID)
	if err != nil {
		t.Fatalf("failed to find product item: %v", err)
	}
	err = db.Pool.QueryRow(ctx, "SELECT id FROM warehouse.warehouses WHERE legal_entity_id = $1 LIMIT 1", le.ID).Scan(&warehouseID)
	if err != nil {
		t.Fatalf("failed to find warehouse: %v", err)
	}

	return salesMod, recMod, db, le.ID, adminID, itemID, warehouseID, ctx
}

func createTestCustomer(t *testing.T, recMod *receivables.Module, leID uuid.UUID, custType recDomain.CustomerType, nameAr, nameEn string, taxID *string, limit decimal.Decimal, ctx context.Context) *recDomain.Customer {
	custSvc := recMod.CustomerService
	cmd := recPorts.CreateCustomerCommand{
		LegalEntityID:  leID,
		CustomerType:   custType,
		NameEn:         nameEn,
		NameAr:         nameAr,
		TaxID:          taxID,
		Address:        recDomain.SaudiAddress{City: "Riyadh", CountryCode: "SA"},
		CreditLimit:    limit,
		BlockAction:    recDomain.BlockActionBlock,
		CurrencyCode:   "SAR",
		MaxOverdueDays: 30,
	}

	customer, err := custSvc.CreateCustomer(ctx, cmd)
	if err != nil {
		t.Fatalf("failed to create test customer: %v", err)
	}
	return customer
}

func TestSales_QuotationToOrderConversion(t *testing.T) {
	salesMod, recMod, db, leID, adminID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()

	salesSvc := salesMod.SalesService
	taxID := "310123456700003"
	customer := createTestCustomer(t, recMod, leID, recDomain.CustomerTypeB2B, "شركة النخبة للتوريدات", "Elite Supplies Co.", &taxID, decimal.NewFromInt(100000), ctx)

	// 1. Create Quotation: 10 units @ 500 SAR, 10% discount, 15% VAT
	// Gross = 5000, Disc = 500, Taxable = 4500, Tax = 675, Total = 5175
	qCmd := ports.CreateQuotationCommand{
		LegalEntityID:   leID,
		CustomerID:      customer.ID,
		CustomerNameEn:  customer.NameEn,
		CustomerNameAr:  customer.NameAr,
		QuoteDate:       time.Now(),
		ValidUntil:      time.Now().AddDate(0, 0, 30),
		CurrencyCode:    "SAR",
		ExchangeRate:    decimal.NewFromInt(1),
		CreatedByUserID: &adminID,
		Lines: []ports.CreateQuotationLineCommand{
			{
				ItemID:          itemID,
				DescriptionEn:   "Industrial Pumps",
				DescriptionAr:   "مضخات صناعية",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(10),
				UnitPrice:       decimal.NewFromInt(500),
				DiscountPercent: decimal.NewFromInt(10),
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	}

	quote, err := salesSvc.CreateQuotation(ctx, qCmd)
	if err != nil {
		t.Fatalf("failed to create quotation: %v", err)
	}

	if quote.ID == uuid.Nil {
		t.Errorf("expected valid quotation ID")
	}
	if !quote.Subtotal.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("expected subtotal 5000, got %s", quote.Subtotal)
	}
	if !quote.DiscountAmount.Equal(decimal.NewFromInt(500)) {
		t.Errorf("expected discount 500, got %s", quote.DiscountAmount)
	}
	if !quote.TaxableAmount.Equal(decimal.NewFromInt(4500)) {
		t.Errorf("expected taxable 4500, got %s", quote.TaxableAmount)
	}
	if !quote.TaxAmount.Equal(decimal.NewFromInt(675)) {
		t.Errorf("expected tax 675, got %s", quote.TaxAmount)
	}
	if !quote.TotalAmount.Equal(decimal.NewFromInt(5175)) {
		t.Errorf("expected total 5175, got %s", quote.TotalAmount)
	}

	// 2. Convert Quotation to Order
	order, err := salesSvc.ConvertQuotationToOrder(ctx, quote.ID, &adminID)
	if err != nil {
		t.Fatalf("failed to convert quotation to order: %v", err)
	}

	if order.ID == uuid.Nil {
		t.Errorf("expected valid order ID")
	}
	if order.QuotationID == nil || *order.QuotationID != quote.ID {
		t.Errorf("expected quotation ID reference on order")
	}
	if !order.TotalAmount.Equal(decimal.NewFromInt(5175)) {
		t.Errorf("expected order total 5175, got %s", order.TotalAmount)
	}
	if len(order.Lines) != 1 {
		t.Fatalf("expected 1 order line, got %d", len(order.Lines))
	}

	// Verify quotation status updated to CONVERTED
	qUpdated, err := salesSvc.GetQuotation(ctx, quote.ID)
	if err != nil {
		t.Fatalf("failed to get updated quotation: %v", err)
	}
	if qUpdated.Status != domain.QuotationStatusConverted {
		t.Errorf("expected quotation status CONVERTED, got %s", qUpdated.Status)
	}
}

func TestSales_OrderCreditLimitValidation(t *testing.T) {
	salesMod, recMod, db, leID, adminID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()

	salesSvc := salesMod.SalesService
	custSvc := recMod.CustomerService

	taxID := "310123456700003"
	// Customer with 20,000 SAR credit limit
	customer := createTestCustomer(t, recMod, leID, recDomain.CustomerTypeB2B, "مؤسسة الرواد للخدمات", "Al-Rowad Services Est.", &taxID, decimal.NewFromInt(20000), ctx)

	// 1. Order 1: 15,000 SAR -> Confirm succeeds (15,000 <= 20,000)
	order1, err := salesSvc.CreateOrder(ctx, ports.CreateOrderCommand{
		LegalEntityID:   leID,
		CustomerID:      customer.ID,
		OrderDate:       time.Now(),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
		Lines: []ports.CreateOrderLineCommand{
			{
				ItemID:          itemID,
				DescriptionEn:   "Standard Equipment",
				DescriptionAr:   "معدات معيارية",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(15),
				UnitPrice:       decimal.NewFromInt(1000),
				DiscountPercent: decimal.Zero,
				TaxCode:         "EXEMPT",
				TaxRatePercent:  decimal.Zero,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create order 1: %v", err)
	}

	confirmedOrder1, err := salesSvc.ConfirmOrder(ctx, order1.ID)
	if err != nil {
		t.Fatalf("expected order 1 confirmation to succeed: %v", err)
	}
	if confirmedOrder1.Status != domain.OrderStatusConfirmed {
		t.Errorf("expected CONFIRMED status, got %s", confirmedOrder1.Status)
	}

	// 2. Post an invoice for 15,000 SAR to reflect balance in customer_credit_limits
	inv1, err := salesSvc.CreateInvoiceFromOrder(ctx, confirmedOrder1.ID, domain.InvoiceTypeStandard, &adminID)
	if err != nil {
		t.Fatalf("failed to create invoice from order 1: %v", err)
	}
	_, err = salesSvc.PostInvoice(ctx, inv1.ID)
	if err != nil {
		t.Fatalf("failed to post invoice 1: %v", err)
	}

	// Verify customer balance is now 15,000
	custCheck, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !custCheck.CurrentBalance.Equal(decimal.NewFromInt(15000)) {
		t.Errorf("expected customer balance 15000, got %s", custCheck.CurrentBalance)
	}

	// 3. Order 2: 10,000 SAR -> Total balance would be 25,000 > 20,000 limit -> Confirm MUST FAIL!
	order2, err := salesSvc.CreateOrder(ctx, ports.CreateOrderCommand{
		LegalEntityID:   leID,
		CustomerID:      customer.ID,
		OrderDate:       time.Now(),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
		Lines: []ports.CreateOrderLineCommand{
			{
				ItemID:          itemID,
				DescriptionEn:   "Heavy Machinery",
				DescriptionAr:   "آلات ثقيلة",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(10),
				UnitPrice:       decimal.NewFromInt(1000),
				DiscountPercent: decimal.Zero,
				TaxCode:         "EXEMPT",
				TaxRatePercent:  decimal.Zero,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create order 2: %v", err)
	}

	_, err = salesSvc.ConfirmOrder(ctx, order2.ID)
	if !errors.Is(err, domain.ErrCustomerCreditLimitExceeded) {
		t.Errorf("expected ErrCustomerCreditLimitExceeded, got: %v", err)
	}

	// 4. Test Credit Hold: If customer is on hold, even small orders are blocked
	reason := "Hold for review"
	_ = custSvc.SetCreditHold(ctx, customer.ID, true, &reason)

	order3, err := salesSvc.CreateOrder(ctx, ports.CreateOrderCommand{
		LegalEntityID:   leID,
		CustomerID:      customer.ID,
		OrderDate:       time.Now(),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
		Lines: []ports.CreateOrderLineCommand{
			{
				ItemID:          itemID,
				DescriptionEn:   "Small Accessory",
				DescriptionAr:   "ملحق صغير",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(1),
				UnitPrice:       decimal.NewFromInt(100),
				DiscountPercent: decimal.Zero,
				TaxCode:         "EXEMPT",
				TaxRatePercent:  decimal.Zero,
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create order 3: %v", err)
	}

	_, err = salesSvc.ConfirmOrder(ctx, order3.ID)
	if !errors.Is(err, domain.ErrCustomerCreditHold) {
		t.Errorf("expected ErrCustomerCreditHold, got: %v", err)
	}
}

func TestSales_DeliveryAndStockDeduction(t *testing.T) {
	salesMod, recMod, db, leID, adminID, itemID, warehouseID, ctx := setupSalesTest(t)
	defer db.Close()

	salesSvc := salesMod.SalesService
	customer := createTestCustomer(t, recMod, leID, recDomain.CustomerTypeB2C, "سالم الدوسري", "Salem Al-Dawsari", nil, decimal.NewFromInt(50000), ctx)

	// 1. Create and confirm order for 20 units
	order, err := salesSvc.CreateOrder(ctx, ports.CreateOrderCommand{
		LegalEntityID:   leID,
		CustomerID:      customer.ID,
		OrderDate:       time.Now(),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
		Lines: []ports.CreateOrderLineCommand{
			{
				ItemID:          itemID,
				DescriptionEn:   "Commercial Filters",
				DescriptionAr:   "فلاتر تجارية",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(20),
				UnitPrice:       decimal.NewFromInt(100),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	_, _ = salesSvc.ConfirmOrder(ctx, order.ID)

	// 2. Create Delivery for 15 units (partial delivery)
	delCmd := ports.CreateDeliveryCommand{
		LegalEntityID:     leID,
		OrderID:           &order.ID,
		CustomerID:        customer.ID,
		WarehouseID:       warehouseID,
		DeliveryDate:      time.Now(),
		ShippingReference: strPtr("ARAMEX-EXP-8899"),
		Notes:             strPtr("إرسالية أولى 15 قطعة"),
		CreatedByUserID:   &adminID,
		Lines: []ports.CreateDeliveryLineCommand{
			{
				OrderLineID: &order.Lines[0].ID,
				ItemID:      itemID,
				UOM:         "EA",
				Quantity:    decimal.NewFromInt(15),
				WarehouseID: warehouseID,
			},
		},
	}

	delivery, err := salesSvc.CreateDelivery(ctx, delCmd)
	if err != nil {
		t.Fatalf("failed to create delivery: %v", err)
	}

	if delivery.Status != domain.DeliveryStatusDraft {
		t.Errorf("expected DRAFT delivery status, got %s", delivery.Status)
	}

	// 3. Dispatch Delivery
	dispatched, err := salesSvc.DispatchDelivery(ctx, delivery.ID)
	if err != nil {
		t.Fatalf("failed to dispatch delivery: %v", err)
	}

	if dispatched.Status != domain.DeliveryStatusDelivered {
		t.Errorf("expected DELIVERED status, got %s", dispatched.Status)
	}

	// 4. Verify Stock Movement was recorded in inventory.stock_movements with negative quantity
	var smCount int
	var moveType string
	var smQty decimal.Decimal
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(movement_type), ''), COALESCE(SUM(quantity), 0)
		FROM inventory.stock_movements
		WHERE source_doc_id = $1
	`, delivery.ID).Scan(&smCount, &moveType, &smQty)
	if err != nil {
		t.Fatalf("failed to query stock movements: %v", err)
	}

	if smCount < 1 {
		t.Errorf("expected at least 1 stock movement, got %d", smCount)
	}
	if moveType != "SALES_ISSUE" {
		t.Errorf("expected movement_type SALES_ISSUE, got %s", moveType)
	}
	if !smQty.Equal(decimal.NewFromInt(-15)) {
		t.Errorf("expected quantity -15, got %s", smQty)
	}

	// 5. Verify Order Line delivered_quantity was updated
	updatedOrder, _ := salesSvc.GetOrder(ctx, order.ID)
	if !updatedOrder.Lines[0].DeliveredQuantity.Equal(decimal.NewFromInt(15)) {
		t.Errorf("expected delivered_quantity 15, got %s", updatedOrder.Lines[0].DeliveredQuantity)
	}
}

func TestSales_StandardInvoicePosting(t *testing.T) {
	salesMod, recMod, db, leID, adminID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()

	salesSvc := salesMod.SalesService
	custSvc := recMod.CustomerService

	taxID := "310123456700003"
	cr := "1010998877"
	customer := createTestCustomer(t, recMod, leID, recDomain.CustomerTypeB2B, "شركة الإمداد الوطنية المحدودة", "National Supply Co. Ltd.", &taxID, decimal.NewFromInt(500000), ctx)

	// 1. Negative Test: Standard invoice without valid 15-digit VAT number must fail
	badCmd := ports.CreateInvoiceCommand{
		LegalEntityID:  leID,
		InvoiceType:    domain.InvoiceTypeStandard,
		CustomerID:     customer.ID,
		CustomerNameAr: customer.NameAr,
		CustomerNameEn: customer.NameEn,
		CustomerTaxID:  strPtr("12345"), // invalid tax id
		IssueDate:      time.Now(),
		DueDate:        time.Now().AddDate(0, 0, 30),
		CurrencyCode:   "SAR",
		Lines: []ports.CreateInvoiceLineCommand{
			{
				ItemID:          itemID,
				Description:     "Medical Devices",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(1),
				UnitPrice:       decimal.NewFromInt(5000),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	}
	_, err := salesSvc.CreateInvoice(ctx, badCmd)
	if !errors.Is(err, domain.ErrInvalidTaxIDForStandardInvoice) {
		t.Errorf("expected ErrInvalidTaxIDForStandardInvoice, got: %v", err)
	}

	// 2. Positive Test: Create Standard Tax Invoice: 2 units @ 10,000 SAR + 15% VAT
	// Taxable = 20,000, Tax = 3,000, Total = 23,000 SAR
	invCmd := ports.CreateInvoiceCommand{
		LegalEntityID:       leID,
		InvoiceType:         domain.InvoiceTypeStandard,
		CustomerID:          customer.ID,
		CustomerNameAr:      customer.NameAr,
		CustomerNameEn:      customer.NameEn,
		CustomerTaxID:       &taxID,
		CustomerCRNumber:    &cr,
		CustomerAddressJSON: strPtr(`{"city":"Riyadh","building_no":"4455","postal_code":"12233"}`),
		IssueDate:           time.Now(),
		DueDate:             time.Now().AddDate(0, 0, 30),
		CurrencyCode:        "SAR",
		ExchangeRate:        decimal.NewFromInt(1),
		Notes:               strPtr("فاتورة ضريبية قياسية معتمدة"),
		CreatedByUserID:     &adminID,
		Lines: []ports.CreateInvoiceLineCommand{
			{
				ItemID:          itemID,
				Description:     "أجهزة طبية متقدمة",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(2),
				UnitPrice:       decimal.NewFromInt(10000),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	}

	invoice, err := salesSvc.CreateInvoice(ctx, invCmd)
	if err != nil {
		t.Fatalf("failed to create standard invoice: %v", err)
	}

	if invoice.Status != domain.InvoiceStatusDraft {
		t.Errorf("expected DRAFT invoice, got %s", invoice.Status)
	}
	if !invoice.TotalAmount.Equal(decimal.NewFromInt(23000)) {
		t.Errorf("expected total 23000, got %s", invoice.TotalAmount)
	}

	// 3. Post Invoice
	postedInv, err := salesSvc.PostInvoice(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("failed to post invoice: %v", err)
	}

	if postedInv.Status != domain.InvoiceStatusPosted {
		t.Errorf("expected POSTED invoice status, got %s", postedInv.Status)
	}
	if postedInv.PostedAt == nil {
		t.Errorf("expected posted_at timestamp to be set")
	}
	if postedInv.QRCodeData != nil {
		t.Errorf("expected QR code to remain unset until the invoice is signed by the ZATCA flow")
	}
	if postedInv.OpenItemID == nil {
		t.Errorf("expected open_item_id to be populated")
	}

	// 4. Verify Open Item in receivables.open_items
	var openItemDoc string
	var openRemaining decimal.Decimal
	err = db.Pool.QueryRow(ctx, `
		SELECT document_number, remaining_amount
		FROM receivables.open_items
		WHERE id = $1
	`, *postedInv.OpenItemID).Scan(&openItemDoc, &openRemaining)
	if err != nil {
		t.Fatalf("failed to query open item: %v", err)
	}
	if openItemDoc != invoice.InvoiceNumber {
		t.Errorf("expected doc number %s, got %s", invoice.InvoiceNumber, openItemDoc)
	}
	if !openRemaining.Equal(decimal.NewFromInt(23000)) {
		t.Errorf("expected remaining amount 23000, got %s", openRemaining)
	}

	// 5. Verify Customer Balance Updated in receivables.customer_credit_limits
	cCheck, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !cCheck.CurrentBalance.Equal(decimal.NewFromInt(23000)) {
		t.Errorf("expected customer balance 23000, got %s", cCheck.CurrentBalance)
	}

	// 6. Verify GL Journal Entry in accounting.journal_entries
	if postedInv.JournalEntryID != nil {
		var debits, credits decimal.Decimal
		var jeStatus string
		err = db.Pool.QueryRow(ctx, `
			SELECT status, total_debit, total_credit
			FROM accounting.journal_entries
			WHERE id = $1
		`, *postedInv.JournalEntryID).Scan(&jeStatus, &debits, &credits)
		if err != nil {
			t.Fatalf("failed to query journal entry: %v", err)
		}
		if jeStatus != "POSTED" {
			t.Errorf("expected GL entry status POSTED, got %s", jeStatus)
		}
		if !debits.Equal(decimal.NewFromInt(23000)) || !credits.Equal(decimal.NewFromInt(23000)) {
			t.Errorf("expected balanced entry 23000, got Dr: %s, Cr: %s", debits, credits)
		}
	}
}

func TestSales_SimplifiedInvoicePosting(t *testing.T) {
	salesMod, recMod, db, leID, adminID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()

	salesSvc := salesMod.SalesService
	// B2C Customer without tax ID
	customer := createTestCustomer(t, recMod, leID, recDomain.CustomerTypeB2C, "عبدالله الشمري", "Abdullah Al-Shammari", nil, decimal.NewFromInt(10000), ctx)

	invCmd := ports.CreateInvoiceCommand{
		LegalEntityID:   leID,
		InvoiceType:     domain.InvoiceTypeSimplified,
		CustomerID:      customer.ID,
		CustomerNameAr:  customer.NameAr,
		CustomerNameEn:  customer.NameEn,
		IssueDate:       time.Now(),
		DueDate:         time.Now(),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
		Lines: []ports.CreateInvoiceLineCommand{
			{
				ItemID:          itemID,
				Description:     "بطاقة خدمة مستهلك",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(3),
				UnitPrice:       decimal.NewFromInt(200),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	}

	invoice, err := salesSvc.CreateInvoice(ctx, invCmd)
	if err != nil {
		t.Fatalf("failed to create simplified invoice: %v", err)
	}

	if invoice.InvoiceType != domain.InvoiceTypeSimplified {
		t.Errorf("expected SIMPLIFIED type, got %s", invoice.InvoiceType)
	}

	posted, err := salesSvc.PostInvoice(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("failed to post simplified invoice: %v", err)
	}

	if posted.Status != domain.InvoiceStatusPosted {
		t.Errorf("expected POSTED status, got %s", posted.Status)
	}

	// Verify line snapshot attributes
	line := posted.Lines[0]
	if line.TaxCode != "VAT-15" {
		t.Errorf("expected snapshot tax_code VAT-15, got %s", line.TaxCode)
	}
	if !line.TaxRatePercent.Equal(decimal.NewFromInt(15)) {
		t.Errorf("expected snapshot tax_rate_percent 15, got %s", line.TaxRatePercent)
	}
	if !line.TaxAmount.Equal(decimal.NewFromInt(90)) {
		t.Errorf("expected snapshot tax_amount 90, got %s", line.TaxAmount)
	}
}

func TestSales_CreditNoteLifecycle(t *testing.T) {
	salesMod, recMod, db, leID, adminID, itemID, _, ctx := setupSalesTest(t)
	defer db.Close()

	salesSvc := salesMod.SalesService
	custSvc := recMod.CustomerService

	taxID := "310123456700003"
	customer := createTestCustomer(t, recMod, leID, recDomain.CustomerTypeB2B, "شركة البناء المتطورة", "Advanced Construction Co.", &taxID, decimal.NewFromInt(100000), ctx)

	// 1. Post original invoice: 10,000 + 15% VAT = 11,500 SAR
	invCmd := ports.CreateInvoiceCommand{
		LegalEntityID:   leID,
		InvoiceType:     domain.InvoiceTypeStandard,
		CustomerID:      customer.ID,
		CustomerNameAr:  customer.NameAr,
		CustomerNameEn:  customer.NameEn,
		CustomerTaxID:   &taxID,
		IssueDate:       time.Now(),
		DueDate:         time.Now().AddDate(0, 0, 30),
		CurrencyCode:    "SAR",
		CreatedByUserID: &adminID,
		Lines: []ports.CreateInvoiceLineCommand{
			{
				ItemID:          itemID,
				Description:     "حديد تسليح قياسي",
				UOM:             "EA",
				Quantity:        decimal.NewFromInt(10),
				UnitPrice:       decimal.NewFromInt(1000),
				DiscountPercent: decimal.Zero,
				TaxCode:         "VAT-15",
				TaxRatePercent:  decimal.NewFromInt(15),
			},
		},
	}

	invoice, err := salesSvc.CreateInvoice(ctx, invCmd)
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}
	postedInv, err := salesSvc.PostInvoice(ctx, invoice.ID)
	if err != nil {
		t.Fatalf("failed to post invoice: %v", err)
	}

	// Customer balance is now 11,500 SAR
	c1, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !c1.CurrentBalance.Equal(decimal.NewFromInt(11500)) {
		t.Errorf("expected customer balance 11500, got %s", c1.CurrentBalance)
	}

	// 2. Issue Credit Note for 2 returned units: 2,000 + 15% VAT = 2,300 SAR
	cnCmd := ports.CreateCreditNoteCommand{
		LegalEntityID:     leID,
		OriginalInvoiceID: postedInv.ID,
		CustomerID:        customer.ID,
		IssueDate:         time.Now(),
		ReasonCode:        domain.CreditNoteReasonReturn,
		ReasonDetails:     strPtr("إرجاع وحدتين بسبب فائض المشروع"),
		CurrencyCode:      "SAR",
		CreatedByUserID:   &adminID,
		Lines: []ports.CreateCreditNoteLineCommand{
			{
				ItemID:         itemID,
				Description:    "إرجاع حديد تسليح",
				Quantity:       decimal.NewFromInt(2),
				UnitPrice:      decimal.NewFromInt(1000),
				TaxCode:        "VAT-15",
				TaxRatePercent: decimal.NewFromInt(15),
			},
		},
	}

	cn, err := salesSvc.CreateCreditNote(ctx, cnCmd)
	if err != nil {
		t.Fatalf("failed to create credit note: %v", err)
	}

	if !cn.TotalAmount.Equal(decimal.NewFromInt(2300)) {
		t.Errorf("expected credit note total 2300, got %s", cn.TotalAmount)
	}

	// 3. Post Credit Note
	postedCN, err := salesSvc.PostCreditNote(ctx, cn.ID)
	if err != nil {
		t.Fatalf("failed to post credit note: %v", err)
	}

	if postedCN.Status != domain.CreditNoteStatusPosted {
		t.Errorf("expected POSTED credit note, got %s", postedCN.Status)
	}

	// 4. Verify Customer Balance decreased: 11,500 - 2,300 = 9,200 SAR
	c2, _ := custSvc.GetCustomer(ctx, customer.ID)
	if !c2.CurrentBalance.Equal(decimal.NewFromInt(9200)) {
		t.Errorf("expected customer balance 9200, got %s", c2.CurrentBalance)
	}

	// 5. Verify Credit Note Open Item in receivables.open_items
	var cnOpenCount int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM receivables.open_items
		WHERE document_type = 'CREDIT_NOTE' AND document_id = $1
	`, postedCN.ID).Scan(&cnOpenCount)
	if err != nil {
		t.Fatalf("failed to query credit note open item: %v", err)
	}
	if cnOpenCount != 1 {
		t.Errorf("expected 1 credit note open item, got %d", cnOpenCount)
	}
}

func strPtr(s string) *string {
	return &s
}
