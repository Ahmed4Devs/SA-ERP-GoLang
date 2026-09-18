package services

import (
	"context"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"
	"erp/internal/purchasing/domain"
	"erp/internal/purchasing/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type purchasingService struct {
	repo ports.Repository
	db   *database.DB
	tx   transaction.Manager
}

func NewPurchasingService(repo ports.Repository, db *database.DB) ports.PurchasingService {
	return &purchasingService{
		repo: repo,
		db:   db,
		tx:   transaction.NewManager(db.Pool),
	}
}

// ----------------------------------------------------------------------------
// Purchase Requests
// ----------------------------------------------------------------------------

func (s *purchasingService) CreatePurchaseRequest(ctx context.Context, cmd ports.CreatePRCmd) (*domain.PurchaseRequest, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	prNumber := identifiers.NewDocumentNumber("PR", time.Now())
	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}

	pr := &domain.PurchaseRequest{
		ID:                uuid.New(),
		LegalEntityID:     cmd.LegalEntityID,
		PRNumber:          prNumber,
		RequestedByUserID: cmd.RequestedByUserID,
		DepartmentID:      cmd.DepartmentID,
		RequestDate:       cmd.RequestDate,
		RequiredByDate:    cmd.RequiredByDate,
		Status:            domain.PRStatusDraft,
		Purpose:           cmd.Purpose,
		Notes:             cmd.Notes,
		EstimatedTotal:    decimal.Zero,
		CurrencyCode:      currency,
	}

	for i, l := range cmd.Lines {
		total := l.Quantity.Mul(l.EstimatedUnitCost).Round(4)
		pr.EstimatedTotal = pr.EstimatedTotal.Add(total)

		uom := l.UOM
		if uom == "" {
			uom = "EA"
		}

		pr.Lines = append(pr.Lines, domain.PurchaseRequestLine{
			ID:                  uuid.New(),
			LineNumber:          i + 1,
			ItemID:              l.ItemID,
			Description:         l.Description,
			UOM:                 uom,
			Quantity:            l.Quantity,
			OrderedQuantity:     decimal.Zero,
			EstimatedUnitCost:   l.EstimatedUnitCost,
			EstimatedTotalCost:  total,
			SuggestedSupplierID: l.SuggestedSupplierID,
		})
	}

	var created *domain.PurchaseRequest
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreatePurchaseRequest(txCtx, pr)
		return err
	})
	return created, err
}

func (s *purchasingService) GetPurchaseRequest(ctx context.Context, id uuid.UUID) (*domain.PurchaseRequest, error) {
	return s.repo.GetPurchaseRequest(ctx, id)
}

func (s *purchasingService) ListPurchaseRequests(ctx context.Context, legalEntityID uuid.UUID, status *domain.PRStatus, limit, offset int) ([]*domain.PurchaseRequest, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListPurchaseRequests(ctx, legalEntityID, status, limit, offset)
}

func (s *purchasingService) ApprovePurchaseRequest(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*domain.PurchaseRequest, error) {
	pr, err := s.repo.GetPurchaseRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if pr.Status != domain.PRStatusDraft && pr.Status != domain.PRStatusSubmitted {
		return nil, domain.ErrInvalidPOStatus
	}
	return s.repo.UpdatePurchaseRequestStatus(ctx, id, domain.PRStatusApproved)
}

// ----------------------------------------------------------------------------
// RFQs & Supplier Quotes
// ----------------------------------------------------------------------------

func (s *purchasingService) CreateRFQ(ctx context.Context, cmd ports.CreateRFQCmd) (*domain.RFQ, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	rfqNumber := identifiers.NewDocumentNumber("RFQ", time.Now())

	rfq := &domain.RFQ{
		ID:                uuid.New(),
		LegalEntityID:     cmd.LegalEntityID,
		RFQNumber:         rfqNumber,
		PurchaseRequestID: cmd.PurchaseRequestID,
		Title:             cmd.Title,
		RFQDate:           cmd.RFQDate,
		ClosingDate:       cmd.ClosingDate,
		Status:            domain.RFQStatusDraft,
		Notes:             cmd.Notes,
		CreatedByUserID:   cmd.CreatedByUserID,
	}

	for i, l := range cmd.Lines {
		uom := l.UOM
		if uom == "" {
			uom = "EA"
		}
		rfq.Lines = append(rfq.Lines, domain.RFQLine{
			ID:              uuid.New(),
			LineNumber:      i + 1,
			PRLineID:        l.PRLineID,
			ItemID:          l.ItemID,
			Description:     l.Description,
			UOM:             uom,
			Quantity:        l.Quantity,
			TargetUnitPrice: l.TargetUnitPrice,
		})
	}

	var created *domain.RFQ
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreateRFQ(txCtx, rfq)
		return err
	})
	return created, err
}

func (s *purchasingService) GetRFQ(ctx context.Context, id uuid.UUID) (*domain.RFQ, error) {
	return s.repo.GetRFQ(ctx, id)
}

func (s *purchasingService) ListRFQs(ctx context.Context, legalEntityID uuid.UUID, status *domain.RFQStatus, limit, offset int) ([]*domain.RFQ, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListRFQs(ctx, legalEntityID, status, limit, offset)
}

func (s *purchasingService) RecordSupplierQuote(ctx context.Context, cmd ports.RecordQuoteCmd) (*domain.SupplierQuote, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}
	rate := cmd.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	quote := &domain.SupplierQuote{
		ID:                     uuid.New(),
		LegalEntityID:          cmd.LegalEntityID,
		RFQID:                  cmd.RFQID,
		SupplierID:             cmd.SupplierID,
		SupplierQuoteReference: cmd.SupplierQuoteReference,
		QuoteDate:              cmd.QuoteDate,
		ValidUntil:             cmd.ValidUntil,
		CurrencyCode:           currency,
		ExchangeRate:           rate,
		PaymentTermsID:         cmd.PaymentTermsID,
		Subtotal:               decimal.Zero,
		TaxAmount:              decimal.Zero,
		TotalAmount:            decimal.Zero,
		Status:                 domain.QuoteStatusPending,
		Notes:                  cmd.Notes,
	}

	for i, l := range cmd.Lines {
		_, _, taxable, tax, total := domain.CalculateLineAmounts(l.Quantity, l.UnitPrice, l.DiscountPercent, l.TaxRatePercent)
		quote.Subtotal = quote.Subtotal.Add(taxable)
		quote.TaxAmount = quote.TaxAmount.Add(tax)
		quote.TotalAmount = quote.TotalAmount.Add(total)

		quote.Lines = append(quote.Lines, domain.SupplierQuoteLine{
			ID:              uuid.New(),
			LineNumber:      i + 1,
			RFQLineID:       l.RFQLineID,
			ItemID:          l.ItemID,
			Description:     l.Description,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			DiscountPercent: l.DiscountPercent,
			TaxRatePercent:  l.TaxRatePercent,
			TaxAmount:       tax,
			TotalAmount:     total,
			IsAwarded:       false,
		})
	}

	var created *domain.SupplierQuote
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreateSupplierQuote(txCtx, quote)
		return err
	})
	return created, err
}

func (s *purchasingService) GetSupplierQuote(ctx context.Context, id uuid.UUID) (*domain.SupplierQuote, error) {
	return s.repo.GetSupplierQuote(ctx, id)
}

func (s *purchasingService) ListSupplierQuotes(ctx context.Context, legalEntityID uuid.UUID, rfqID, supplierID *uuid.UUID, status *domain.QuoteStatus, limit, offset int) ([]*domain.SupplierQuote, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListSupplierQuotes(ctx, legalEntityID, rfqID, supplierID, status, limit, offset)
}

func (s *purchasingService) AwardSupplierQuote(ctx context.Context, quoteID uuid.UUID) (*domain.SupplierQuote, error) {
	var awarded *domain.SupplierQuote
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		quote, err := s.repo.GetSupplierQuote(txCtx, quoteID)
		if err != nil {
			return err
		}
		for _, line := range quote.Lines {
			if err := s.repo.AwardSupplierQuoteLine(txCtx, line.ID, true); err != nil {
				return fmt.Errorf("failed to award supplier quote line %s: %w", line.ID, err)
			}
		}
		awarded, err = s.repo.UpdateSupplierQuoteStatus(txCtx, quoteID, domain.QuoteStatusAccepted)
		return err
	})
	return awarded, err
}

// ----------------------------------------------------------------------------
// Purchase Orders
// ----------------------------------------------------------------------------

func (s *purchasingService) CreatePurchaseOrder(ctx context.Context, cmd ports.CreatePOCmd) (*domain.PurchaseOrder, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	poNumber := identifiers.NewDocumentNumber("PO", time.Now())
	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}
	rate := cmd.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	po := &domain.PurchaseOrder{
		ID:                   uuid.New(),
		LegalEntityID:        cmd.LegalEntityID,
		PONumber:             poNumber,
		PurchaseRequestID:    cmd.PurchaseRequestID,
		RFQID:                cmd.RFQID,
		SupplierQuoteID:      cmd.SupplierQuoteID,
		SupplierID:           cmd.SupplierID,
		OrderDate:            cmd.OrderDate,
		ExpectedDeliveryDate: cmd.ExpectedDeliveryDate,
		PaymentTermsID:       cmd.PaymentTermsID,
		DeliveryWarehouseID:  cmd.DeliveryWarehouseID,
		CurrencyCode:         currency,
		ExchangeRate:         rate,
		Subtotal:             decimal.Zero,
		DiscountAmount:       decimal.Zero,
		TaxableAmount:        decimal.Zero,
		TaxAmount:            decimal.Zero,
		TotalAmount:          decimal.Zero,
		Status:               domain.POStatusDraft,
		Notes:                cmd.Notes,
		CreatedByUserID:      cmd.CreatedByUserID,
	}

	for i, l := range cmd.Lines {
		gross, disc, taxable, tax, total := domain.CalculateLineAmounts(l.Quantity, l.UnitPrice, l.DiscountPercent, l.TaxRatePercent)
		po.Subtotal = po.Subtotal.Add(gross)
		po.DiscountAmount = po.DiscountAmount.Add(disc)
		po.TaxableAmount = po.TaxableAmount.Add(taxable)
		po.TaxAmount = po.TaxAmount.Add(tax)
		po.TotalAmount = po.TotalAmount.Add(total)

		uom := l.UOM
		if uom == "" {
			uom = "EA"
		}
		taxCode := l.TaxCode
		if taxCode == "" {
			taxCode = "VAT-15"
		}

		po.Lines = append(po.Lines, domain.PurchaseOrderLine{
			ID:               uuid.New(),
			LineNumber:       i + 1,
			PRLineID:         l.PRLineID,
			ItemID:           l.ItemID,
			Description:      l.Description,
			UOM:              uom,
			Quantity:         l.Quantity,
			ReceivedQuantity: decimal.Zero,
			BilledQuantity:   decimal.Zero,
			UnitPrice:        l.UnitPrice,
			DiscountPercent:  l.DiscountPercent,
			DiscountAmount:   disc,
			TaxableAmount:    taxable,
			TaxCode:          taxCode,
			TaxRatePercent:   l.TaxRatePercent,
			TaxAmount:        tax,
			TotalAmount:      total,
			ExpenseAccountID: l.ExpenseAccountID,
		})

	}

	var created *domain.PurchaseOrder
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		for i, line := range po.Lines {
			if line.PRLineID != nil {
				if err := s.repo.UpdatePRLineOrderedQty(txCtx, *line.PRLineID, cmd.Lines[i].Quantity); err != nil {
					return fmt.Errorf("failed to update purchase-request ordered quantity: %w", err)
				}
			}
		}
		var err error
		created, err = s.repo.CreatePurchaseOrder(txCtx, po)
		return err
	})
	return created, err
}

func (s *purchasingService) GetPurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error) {
	return s.repo.GetPurchaseOrder(ctx, id)
}

func (s *purchasingService) ListPurchaseOrders(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID, status *domain.POStatus, limit, offset int) ([]*domain.PurchaseOrder, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListPurchaseOrders(ctx, legalEntityID, supplierID, status, limit, offset)
}

func (s *purchasingService) ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*domain.PurchaseOrder, error) {
	po, err := s.repo.GetPurchaseOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.Status != domain.POStatusDraft {
		return nil, domain.ErrInvalidPOStatus
	}
	return s.repo.ApprovePurchaseOrder(ctx, id, approverID)
}

func (s *purchasingService) IssuePurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error) {
	po, err := s.repo.GetPurchaseOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.Status != domain.POStatusApproved {
		return nil, domain.ErrPONotApproved
	}
	return s.repo.IssuePurchaseOrder(ctx, id)
}

// ----------------------------------------------------------------------------
// Goods Receipts & Inspection
// ----------------------------------------------------------------------------

func (s *purchasingService) CreateGoodsReceipt(ctx context.Context, cmd ports.CreateReceiptCmd) (*domain.GoodsReceipt, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	po, err := s.repo.GetPurchaseOrder(ctx, cmd.PurchaseOrderID)
	if err != nil {
		return nil, err
	}
	if po.Status != domain.POStatusIssued && po.Status != domain.POStatusPartiallyReceived {
		return nil, domain.ErrPONotIssued
	}

	receiptNumber := identifiers.NewDocumentNumber("GRN", time.Now())

	gr := &domain.GoodsReceipt{
		ID:                 uuid.New(),
		LegalEntityID:      cmd.LegalEntityID,
		ReceiptNumber:      receiptNumber,
		PurchaseOrderID:    cmd.PurchaseOrderID,
		SupplierID:         cmd.SupplierID,
		WarehouseID:        cmd.WarehouseID,
		ReceiptDate:        cmd.ReceiptDate,
		DeliveryNoteNumber: cmd.DeliveryNoteNumber,
		Status:             domain.GRNStatusDraft,
		Notes:              cmd.Notes,
		ReceivedByUserID:   cmd.ReceivedByUserID,
	}

	for i, l := range cmd.Lines {
		totalCost := l.AcceptedQuantity.Mul(l.UnitCost).Round(4)
		gr.Lines = append(gr.Lines, domain.GoodsReceiptLine{
			ID:               uuid.New(),
			LineNumber:       i + 1,
			POLineID:         l.POLineID,
			ItemID:           l.ItemID,
			WarehouseID:      l.WarehouseID,
			BinID:            l.BinID,
			ReceivedQuantity: l.ReceivedQuantity,
			AcceptedQuantity: l.AcceptedQuantity,
			RejectedQuantity: l.RejectedQuantity,
			UnitCost:         l.UnitCost,
			TotalCost:        totalCost,
			RejectionReason:  l.RejectionReason,
		})
	}

	var created *domain.GoodsReceipt
	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.repo.CreateGoodsReceipt(txCtx, gr)
		return err
	})
	return created, err
}

func (s *purchasingService) GetGoodsReceipt(ctx context.Context, id uuid.UUID) (*domain.GoodsReceipt, error) {
	return s.repo.GetGoodsReceipt(ctx, id)
}

func (s *purchasingService) ListGoodsReceipts(ctx context.Context, legalEntityID uuid.UUID, poID, supplierID *uuid.UUID, status *domain.GRNStatus, limit, offset int) ([]*domain.GoodsReceipt, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListGoodsReceipts(ctx, legalEntityID, poID, supplierID, status, limit, offset)
}

func (s *purchasingService) InspectAndAcceptGoodsReceipt(ctx context.Context, id uuid.UUID, inspectorID uuid.UUID) (*domain.GoodsReceipt, error) {
	var receipt *domain.GoodsReceipt
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		receipt, err = s.inspectAndAcceptGoodsReceipt(txCtx, id, inspectorID)
		return err
	})
	return receipt, err
}

func (s *purchasingService) inspectAndAcceptGoodsReceipt(ctx context.Context, id uuid.UUID, inspectorID uuid.UUID) (*domain.GoodsReceipt, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM purchasing.goods_receipts WHERE id = $1 FOR UPDATE", id).Scan(&id); err != nil {
		return nil, fmt.Errorf("failed to lock goods receipt: %w", err)
	}
	gr, err := s.repo.GetGoodsReceipt(ctx, id)
	if err != nil {
		return nil, err
	}
	if gr.Status != domain.GRNStatusDraft {
		return nil, domain.ErrInvalidGRNStatus
	}

	po, err := s.repo.GetPurchaseOrder(ctx, gr.PurchaseOrderID)
	if err != nil {
		return nil, err
	}

	poLinesMap := make(map[uuid.UUID]domain.PurchaseOrderLine)
	for _, l := range po.Lines {
		poLinesMap[l.ID] = l
	}

	// 1. Process inventory increment and stock movements
	for _, line := range gr.Lines {
		if line.AcceptedQuantity.GreaterThan(decimal.Zero) {
			poLine, exists := poLinesMap[line.POLineID]
			if !exists {
				continue
			}

			// Validate remaining quantity
			remaining := poLine.Quantity.Sub(poLine.ReceivedQuantity)
			if line.AcceptedQuantity.GreaterThan(remaining) {
				return nil, domain.ErrQuantityExceedsOrdered
			}

			// Lookup bin if not specified
			binID := line.BinID
			if binID == nil {
				var defBin uuid.UUID
				err := dbtx.QueryRow(ctx, "SELECT id FROM warehouse.bin_locations WHERE warehouse_id = $1 LIMIT 1", line.WarehouseID).Scan(&defBin)
				if err == nil {
					binID = &defBin
				}
			}

			// Lookup UOM ID
			var uomID uuid.UUID
			if err := dbtx.QueryRow(ctx, "SELECT id FROM products.units_of_measure LIMIT 1").Scan(&uomID); err != nil {
				return nil, fmt.Errorf("failed to resolve inventory unit of measure: %w", err)
			}

			movementNumber := identifiers.NewDocumentNumber("SM-IN", time.Now())
			movementID := uuid.New()

			// Insert into inventory.stock_movements (POSITIVE quantity for PURCHASE_RECEIPT)
			_, err = dbtx.Exec(ctx, `
				INSERT INTO inventory.stock_movements (
					id, legal_entity_id, movement_number, movement_type, item_id,
					warehouse_id, bin_id, quantity, uom_id, unit_cost, total_cost,
					source_doc_type, source_doc_id, source_doc_line_id, created_by
				) VALUES ($1, $2, $3, 'PURCHASE_RECEIPT', $4, $5, $6, $7, $8, $9, $10, 'PURCHASE_RECEIPT', $11, $12, $13)
			`, movementID, gr.LegalEntityID, movementNumber, line.ItemID, line.WarehouseID, binID, line.AcceptedQuantity, uomID, line.UnitCost, line.TotalCost, gr.ID, line.ID, inspectorID)

			if err != nil {
				return nil, fmt.Errorf("failed to create purchase stock movement: %w", err)
			}
			if err := s.repo.UpdateGoodsReceiptLineMovement(ctx, line.ID, movementID); err != nil {
				return nil, fmt.Errorf("failed to link goods receipt movement: %w", err)
			}

			if binID != nil {
				if _, err = dbtx.Exec(ctx, `
						INSERT INTO inventory.stock_balances (
							id, legal_entity_id, item_id, warehouse_id, bin_id, quantity_on_hand, average_cost, total_value
						) VALUES (
							uuidv7(), $1, $2, $3, $4, $5, $6, $7
						) ON CONFLICT (legal_entity_id, item_id, warehouse_id, bin_id)
						DO UPDATE SET
							quantity_on_hand = inventory.stock_balances.quantity_on_hand + EXCLUDED.quantity_on_hand,
							total_value = inventory.stock_balances.total_value + EXCLUDED.total_value,
							updated_at = now()
				`, gr.LegalEntityID, line.ItemID, line.WarehouseID, *binID, line.AcceptedQuantity, line.UnitCost, line.TotalCost); err != nil {
					return nil, fmt.Errorf("failed to update stock balance: %w", err)
				}
			}

			// Update PO Line received quantity
			if err := s.repo.UpdatePOLineReceivedQty(ctx, line.POLineID, line.AcceptedQuantity); err != nil {
				return nil, fmt.Errorf("failed to update purchase-order received quantity: %w", err)
			}
		}
	}

	// 2. Update Goods Receipt status
	acceptedGR, err := s.repo.UpdateGoodsReceiptStatus(ctx, id, domain.GRNStatusAccepted)
	if err != nil {
		return nil, err
	}

	// 3. Update PO Status (Check if fully or partially received)
	refreshedPO, err := s.repo.GetPurchaseOrder(ctx, po.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload purchase order: %w", err)
	}
	allReceived := true
	for _, l := range refreshedPO.Lines {
		if l.ReceivedQuantity.LessThan(l.Quantity) {
			allReceived = false
			break
		}
	}
	newStatus := domain.POStatusPartiallyReceived
	if allReceived {
		newStatus = domain.POStatusReceived
	}
	if _, err := s.repo.UpdatePurchaseOrderStatus(ctx, po.ID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update purchase-order status: %w", err)
	}

	return acceptedGR, nil
}

// ----------------------------------------------------------------------------
// Three-Way Matching
// ----------------------------------------------------------------------------

func (s *purchasingService) VerifyThreeWayMatch(ctx context.Context, req ports.ThreeWayMatchReq) (*domain.ThreeWayMatchResult, error) {
	var poLine domain.PurchaseOrderLine
	var grnLine domain.GoodsReceiptLine

	// 1. Fetch PO Line
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, purchase_order_id, line_number, item_id, description, uom, quantity,
		       received_quantity, billed_quantity, unit_price, discount_percent, discount_amount,
		       taxable_amount, tax_code, tax_rate_percent, tax_amount, total_amount, expense_account_id, created_at
		FROM purchasing.purchase_order_lines
		WHERE id = $1
	`, req.POLineID).Scan(
		&poLine.ID, &poLine.PurchaseOrderID, &poLine.LineNumber, &poLine.ItemID, &poLine.Description,
		&poLine.UOM, &poLine.Quantity, &poLine.ReceivedQuantity, &poLine.BilledQuantity, &poLine.UnitPrice,
		&poLine.DiscountPercent, &poLine.DiscountAmount, &poLine.TaxableAmount, &poLine.TaxCode,
		&poLine.TaxRatePercent, &poLine.TaxAmount, &poLine.TotalAmount, &poLine.ExpenseAccountID, &poLine.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("po line not found: %w", err)
	}

	// 2. Fetch GRN Line
	err = s.db.Pool.QueryRow(ctx, `
		SELECT id, goods_receipt_id, line_number, po_line_id, item_id, warehouse_id, bin_id,
		       received_quantity, accepted_quantity, rejected_quantity, unit_cost, total_cost,
		       inventory_movement_id, rejection_reason, created_at
		FROM purchasing.goods_receipt_lines
		WHERE id = $1
	`, req.GRNLineID).Scan(
		&grnLine.ID, &grnLine.GoodsReceiptID, &grnLine.LineNumber, &grnLine.POLineID, &grnLine.ItemID,
		&grnLine.WarehouseID, &grnLine.BinID, &grnLine.ReceivedQuantity, &grnLine.AcceptedQuantity,
		&grnLine.RejectedQuantity, &grnLine.UnitCost, &grnLine.TotalCost, &grnLine.InventoryMovementID,
		&grnLine.RejectionReason, &grnLine.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("grn line not found: %w", err)
	}

	// 3. Evaluate match
	matchResult := domain.PerformThreeWayMatch(
		poLine,
		grnLine,
		req.InvoicedQuantity,
		req.InvoicedUnitPrice,
		req.PriceTolerancePercent,
		req.QtyTolerancePercent,
	)
	matchResult.InvoiceLineID = req.InvoiceLineID

	// If match passed, optionally increment billed quantity on PO line
	if matchResult.PassedTolerance {
		_ = s.repo.UpdatePOLineBilledQty(ctx, poLine.ID, req.InvoicedQuantity)
	}

	return &matchResult, nil
}
