package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/purchasing/domain"
	"erp/internal/purchasing/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type PurchasingRepository struct {
	db *database.DB
}

func NewPurchasingRepository(db *database.DB) ports.Repository {
	return &PurchasingRepository{db: db}
}

func (r *PurchasingRepository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

// ----------------------------------------------------------------------------
// Purchase Requests
// ----------------------------------------------------------------------------

func (r *PurchasingRepository) CreatePurchaseRequest(ctx context.Context, pr *domain.PurchaseRequest) (*domain.PurchaseRequest, error) {
	q := r.queries(ctx)

	row, err := q.CreatePurchaseRequest(ctx, sqlc.CreatePurchaseRequestParams{
		ID:                toPgUUID(pr.ID),
		LegalEntityID:     toPgUUID(pr.LegalEntityID),
		PrNumber:          pr.PRNumber,
		RequestedByUserID: toPgUUID(pr.RequestedByUserID),
		DepartmentID:      toPgUUIDPtr(pr.DepartmentID),
		RequestDate:       toPgDate(pr.RequestDate),
		RequiredByDate: toPgDate(func() time.Time {
			if pr.RequiredByDate != nil {
				return *pr.RequiredByDate
			}
			return time.Time{}
		}()),
		Status:         string(pr.Status),
		Purpose:        toPgTextPtr(pr.Purpose),
		Notes:          toPgTextPtr(pr.Notes),
		EstimatedTotal: toPgNumeric(pr.EstimatedTotal),
		CurrencyCode:   pr.CurrencyCode,
	})
	if err != nil {
		return nil, err
	}

	created := mapPR(row)
	for _, l := range pr.Lines {
		lineRow, err := q.CreatePurchaseRequestLine(ctx, sqlc.CreatePurchaseRequestLineParams{
			ID:                  toPgUUID(l.ID),
			PurchaseRequestID:   row.ID,
			LineNumber:          int32(l.LineNumber),
			ItemID:              toPgUUIDPtr(l.ItemID),
			Description:         l.Description,
			Uom:                 l.UOM,
			Quantity:            toPgNumeric(l.Quantity),
			EstimatedUnitCost:   toPgNumeric(l.EstimatedUnitCost),
			EstimatedTotalCost:  toPgNumeric(l.EstimatedTotalCost),
			SuggestedSupplierID: toPgUUIDPtr(l.SuggestedSupplierID),
		})
		if err != nil {
			return nil, err
		}
		created.Lines = append(created.Lines, mapPRLine(lineRow))
	}

	return created, nil
}

func (r *PurchasingRepository) GetPurchaseRequest(ctx context.Context, id uuid.UUID) (*domain.PurchaseRequest, error) {
	q := r.queries(ctx)
	row, err := q.GetPurchaseRequest(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseRequestNotFound
		}
		return nil, err
	}
	pr := mapPR(row)
	lines, err := q.ListPurchaseRequestLines(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		pr.Lines = append(pr.Lines, mapPRLine(l))
	}
	return pr, nil
}

func (r *PurchasingRepository) ListPurchaseRequests(ctx context.Context, legalEntityID uuid.UUID, status *domain.PRStatus, limit, offset int) ([]*domain.PurchaseRequest, error) {
	q := r.queries(ctx)
	var statusParam *string
	if status != nil {
		s := string(*status)
		statusParam = &s
	}
	rows, err := q.ListPurchaseRequests(ctx, sqlc.ListPurchaseRequestsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        toPgTextPtr(statusParam),
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	res := make([]*domain.PurchaseRequest, 0, len(rows))
	for _, row := range rows {
		pr := mapPR(row)
		lines, _ := q.ListPurchaseRequestLines(ctx, row.ID)
		for _, l := range lines {
			pr.Lines = append(pr.Lines, mapPRLine(l))
		}
		res = append(res, pr)
	}
	return res, nil
}

func (r *PurchasingRepository) UpdatePurchaseRequestStatus(ctx context.Context, id uuid.UUID, status domain.PRStatus) (*domain.PurchaseRequest, error) {
	q := r.queries(ctx)
	row, err := q.UpdatePurchaseRequestStatus(ctx, sqlc.UpdatePurchaseRequestStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetPurchaseRequest(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) UpdatePRLineOrderedQty(ctx context.Context, lineID uuid.UUID, delta decimal.Decimal) error {
	q := r.queries(ctx)
	_, err := q.UpdatePRLineOrderedQty(ctx, sqlc.UpdatePRLineOrderedQtyParams{
		ID:              toPgUUID(lineID),
		OrderedQuantity: toPgNumeric(delta),
	})
	return err
}

// ----------------------------------------------------------------------------
// RFQs & Supplier Quotes
// ----------------------------------------------------------------------------

func (r *PurchasingRepository) CreateRFQ(ctx context.Context, rfq *domain.RFQ) (*domain.RFQ, error) {
	q := r.queries(ctx)
	row, err := q.CreateRFQ(ctx, sqlc.CreateRFQParams{
		ID:                toPgUUID(rfq.ID),
		LegalEntityID:     toPgUUID(rfq.LegalEntityID),
		RfqNumber:         rfq.RFQNumber,
		PurchaseRequestID: toPgUUIDPtr(rfq.PurchaseRequestID),
		Title:             rfq.Title,
		RfqDate:           toPgDate(rfq.RFQDate),
		ClosingDate:       toPgDate(rfq.ClosingDate),
		Status:            string(rfq.Status),
		Notes:             toPgTextPtr(rfq.Notes),
		CreatedByUserID:   toPgUUIDPtr(rfq.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	created := mapRFQ(row)
	for _, l := range rfq.Lines {
		lineRow, err := q.CreateRFQLine(ctx, sqlc.CreateRFQLineParams{
			ID:              toPgUUID(l.ID),
			RfqID:           row.ID,
			LineNumber:      int32(l.LineNumber),
			PrLineID:        toPgUUIDPtr(l.PRLineID),
			ItemID:          toPgUUIDPtr(l.ItemID),
			Description:     l.Description,
			Uom:             l.UOM,
			Quantity:        toPgNumeric(l.Quantity),
			TargetUnitPrice: toPgNumericPtr(l.TargetUnitPrice),
		})
		if err != nil {
			return nil, err
		}
		created.Lines = append(created.Lines, mapRFQLine(lineRow))
	}
	return created, nil
}

func (r *PurchasingRepository) GetRFQ(ctx context.Context, id uuid.UUID) (*domain.RFQ, error) {
	q := r.queries(ctx)
	row, err := q.GetRFQ(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRFQNotFound
		}
		return nil, err
	}
	rfq := mapRFQ(row)
	lines, err := q.ListRFQLines(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		rfq.Lines = append(rfq.Lines, mapRFQLine(l))
	}
	return rfq, nil
}

func (r *PurchasingRepository) ListRFQs(ctx context.Context, legalEntityID uuid.UUID, status *domain.RFQStatus, limit, offset int) ([]*domain.RFQ, error) {
	q := r.queries(ctx)
	var s *string
	if status != nil {
		str := string(*status)
		s = &str
	}
	rows, err := q.ListRFQs(ctx, sqlc.ListRFQsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        toPgTextPtr(s),
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.RFQ, 0, len(rows))
	for _, row := range rows {
		rfq := mapRFQ(row)
		lines, _ := q.ListRFQLines(ctx, row.ID)
		for _, l := range lines {
			rfq.Lines = append(rfq.Lines, mapRFQLine(l))
		}
		res = append(res, rfq)
	}
	return res, nil
}

func (r *PurchasingRepository) UpdateRFQStatus(ctx context.Context, id uuid.UUID, status domain.RFQStatus) (*domain.RFQ, error) {
	q := r.queries(ctx)
	row, err := q.UpdateRFQStatus(ctx, sqlc.UpdateRFQStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetRFQ(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) CreateSupplierQuote(ctx context.Context, sq *domain.SupplierQuote) (*domain.SupplierQuote, error) {
	q := r.queries(ctx)
	row, err := q.CreateSupplierQuote(ctx, sqlc.CreateSupplierQuoteParams{
		ID:                     toPgUUID(sq.ID),
		LegalEntityID:          toPgUUID(sq.LegalEntityID),
		RfqID:                  toPgUUIDPtr(sq.RFQID),
		SupplierID:             toPgUUID(sq.SupplierID),
		SupplierQuoteReference: sq.SupplierQuoteReference,
		QuoteDate:              toPgDate(sq.QuoteDate),
		ValidUntil:             toPgDate(sq.ValidUntil),
		CurrencyCode:           sq.CurrencyCode,
		ExchangeRate:           toPgNumeric(sq.ExchangeRate),
		PaymentTermsID:         toPgUUIDPtr(sq.PaymentTermsID),
		Subtotal:               toPgNumeric(sq.Subtotal),
		TaxAmount:              toPgNumeric(sq.TaxAmount),
		TotalAmount:            toPgNumeric(sq.TotalAmount),
		Status:                 string(sq.Status),
		Notes:                  toPgTextPtr(sq.Notes),
	})
	if err != nil {
		return nil, err
	}

	created := mapSupplierQuote(row)
	for _, l := range sq.Lines {
		lineRow, err := q.CreateSupplierQuoteLine(ctx, sqlc.CreateSupplierQuoteLineParams{
			ID:              toPgUUID(l.ID),
			QuoteID:         row.ID,
			LineNumber:      int32(l.LineNumber),
			RfqLineID:       toPgUUIDPtr(l.RFQLineID),
			ItemID:          toPgUUIDPtr(l.ItemID),
			Description:     l.Description,
			Quantity:        toPgNumeric(l.Quantity),
			UnitPrice:       toPgNumeric(l.UnitPrice),
			DiscountPercent: toPgNumeric(l.DiscountPercent),
			TaxRatePercent:  toPgNumeric(l.TaxRatePercent),
			TaxAmount:       toPgNumeric(l.TaxAmount),
			TotalAmount:     toPgNumeric(l.TotalAmount),
			IsAwarded:       l.IsAwarded,
		})
		if err != nil {
			return nil, err
		}
		created.Lines = append(created.Lines, mapSupplierQuoteLine(lineRow))
	}
	return created, nil
}

func (r *PurchasingRepository) GetSupplierQuote(ctx context.Context, id uuid.UUID) (*domain.SupplierQuote, error) {
	q := r.queries(ctx)
	row, err := q.GetSupplierQuote(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSupplierQuoteNotFound
		}
		return nil, err
	}
	quote := mapSupplierQuote(row)
	lines, err := q.ListSupplierQuoteLines(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		quote.Lines = append(quote.Lines, mapSupplierQuoteLine(l))
	}
	return quote, nil
}

func (r *PurchasingRepository) ListSupplierQuotes(ctx context.Context, legalEntityID uuid.UUID, rfqID, supplierID *uuid.UUID, status *domain.QuoteStatus, limit, offset int) ([]*domain.SupplierQuote, error) {
	q := r.queries(ctx)
	var s *string
	if status != nil {
		str := string(*status)
		s = &str
	}
	rows, err := q.ListSupplierQuotes(ctx, sqlc.ListSupplierQuotesParams{
		LegalEntityID: toPgUUID(legalEntityID),
		RfqID:         toPgUUIDPtr(rfqID),
		SupplierID:    toPgUUIDPtr(supplierID),
		Status:        toPgTextPtr(s),
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.SupplierQuote, 0, len(rows))
	for _, row := range rows {
		sq := mapSupplierQuote(row)
		lines, _ := q.ListSupplierQuoteLines(ctx, row.ID)
		for _, l := range lines {
			sq.Lines = append(sq.Lines, mapSupplierQuoteLine(l))
		}
		res = append(res, sq)
	}
	return res, nil
}

func (r *PurchasingRepository) UpdateSupplierQuoteStatus(ctx context.Context, id uuid.UUID, status domain.QuoteStatus) (*domain.SupplierQuote, error) {
	q := r.queries(ctx)
	row, err := q.UpdateSupplierQuoteStatus(ctx, sqlc.UpdateSupplierQuoteStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetSupplierQuote(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) AwardSupplierQuoteLine(ctx context.Context, lineID uuid.UUID, awarded bool) error {
	q := r.queries(ctx)
	_, err := q.AwardSupplierQuoteLine(ctx, sqlc.AwardSupplierQuoteLineParams{
		ID:        toPgUUID(lineID),
		IsAwarded: awarded,
	})
	return err
}

// ----------------------------------------------------------------------------
// Purchase Orders
// ----------------------------------------------------------------------------

func (r *PurchasingRepository) CreatePurchaseOrder(ctx context.Context, po *domain.PurchaseOrder) (*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	row, err := q.CreatePurchaseOrder(ctx, sqlc.CreatePurchaseOrderParams{
		ID:                toPgUUID(po.ID),
		LegalEntityID:     toPgUUID(po.LegalEntityID),
		PoNumber:          po.PONumber,
		PurchaseRequestID: toPgUUIDPtr(po.PurchaseRequestID),
		RfqID:             toPgUUIDPtr(po.RFQID),
		SupplierQuoteID:   toPgUUIDPtr(po.SupplierQuoteID),
		SupplierID:        toPgUUID(po.SupplierID),
		OrderDate:         toPgDate(po.OrderDate),
		ExpectedDeliveryDate: toPgDate(func() time.Time {
			if po.ExpectedDeliveryDate != nil {
				return *po.ExpectedDeliveryDate
			}
			return time.Time{}
		}()),
		PaymentTermsID:      toPgUUIDPtr(po.PaymentTermsID),
		DeliveryWarehouseID: toPgUUIDPtr(po.DeliveryWarehouseID),
		CurrencyCode:        po.CurrencyCode,
		ExchangeRate:        toPgNumeric(po.ExchangeRate),
		Subtotal:            toPgNumeric(po.Subtotal),
		DiscountAmount:      toPgNumeric(po.DiscountAmount),
		TaxableAmount:       toPgNumeric(po.TaxableAmount),
		TaxAmount:           toPgNumeric(po.TaxAmount),
		TotalAmount:         toPgNumeric(po.TotalAmount),
		Status:              string(po.Status),
		Notes:               toPgTextPtr(po.Notes),
		CreatedByUserID:     toPgUUIDPtr(po.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	created := mapPO(row)
	for _, l := range po.Lines {
		lineRow, err := q.CreatePurchaseOrderLine(ctx, sqlc.CreatePurchaseOrderLineParams{
			ID:               toPgUUID(l.ID),
			PurchaseOrderID:  row.ID,
			LineNumber:       int32(l.LineNumber),
			PrLineID:         toPgUUIDPtr(l.PRLineID),
			ItemID:           toPgUUID(l.ItemID),
			Description:      l.Description,
			Uom:              l.UOM,
			Quantity:         toPgNumeric(l.Quantity),
			UnitPrice:        toPgNumeric(l.UnitPrice),
			DiscountPercent:  toPgNumeric(l.DiscountPercent),
			DiscountAmount:   toPgNumeric(l.DiscountAmount),
			TaxableAmount:    toPgNumeric(l.TaxableAmount),
			TaxCode:          l.TaxCode,
			TaxRatePercent:   toPgNumeric(l.TaxRatePercent),
			TaxAmount:        toPgNumeric(l.TaxAmount),
			TotalAmount:      toPgNumeric(l.TotalAmount),
			ExpenseAccountID: toPgUUIDPtr(l.ExpenseAccountID),
		})
		if err != nil {
			return nil, err
		}
		created.Lines = append(created.Lines, mapPOLine(lineRow))
	}
	return created, nil
}

func (r *PurchasingRepository) GetPurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	row, err := q.GetPurchaseOrder(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	po := mapPO(row)
	lines, err := q.ListPurchaseOrderLines(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		po.Lines = append(po.Lines, mapPOLine(l))
	}
	return po, nil
}

func (r *PurchasingRepository) GetPurchaseOrderByNumber(ctx context.Context, legalEntityID uuid.UUID, poNumber string) (*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	row, err := q.GetPurchaseOrderByNumber(ctx, sqlc.GetPurchaseOrderByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		PoNumber:      poNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseOrderNotFound
		}
		return nil, err
	}
	return r.GetPurchaseOrder(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) ListPurchaseOrders(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID, status *domain.POStatus, limit, offset int) ([]*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	var s *string
	if status != nil {
		str := string(*status)
		s = &str
	}
	rows, err := q.ListPurchaseOrders(ctx, sqlc.ListPurchaseOrdersParams{
		LegalEntityID: toPgUUID(legalEntityID),
		SupplierID:    toPgUUIDPtr(supplierID),
		Status:        toPgTextPtr(s),
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.PurchaseOrder, 0, len(rows))
	for _, row := range rows {
		po := mapPO(row)
		lines, _ := q.ListPurchaseOrderLines(ctx, row.ID)
		for _, l := range lines {
			po.Lines = append(po.Lines, mapPOLine(l))
		}
		res = append(res, po)
	}
	return res, nil
}

func (r *PurchasingRepository) UpdatePurchaseOrderStatus(ctx context.Context, id uuid.UUID, status domain.POStatus) (*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	row, err := q.UpdatePurchaseOrderStatus(ctx, sqlc.UpdatePurchaseOrderStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetPurchaseOrder(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	row, err := q.ApprovePurchaseOrder(ctx, sqlc.ApprovePurchaseOrderParams{
		ID:               toPgUUID(id),
		ApprovedByUserID: toPgUUID(approverID),
	})
	if err != nil {
		return nil, err
	}
	return r.GetPurchaseOrder(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) IssuePurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error) {
	q := r.queries(ctx)
	row, err := q.IssuePurchaseOrder(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	return r.GetPurchaseOrder(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) UpdatePOLineReceivedQty(ctx context.Context, lineID uuid.UUID, delta decimal.Decimal) error {
	q := r.queries(ctx)
	_, err := q.UpdatePOLineReceivedQty(ctx, sqlc.UpdatePOLineReceivedQtyParams{
		ID:               toPgUUID(lineID),
		ReceivedQuantity: toPgNumeric(delta),
	})
	return err
}

func (r *PurchasingRepository) UpdatePOLineBilledQty(ctx context.Context, lineID uuid.UUID, delta decimal.Decimal) error {
	q := r.queries(ctx)
	_, err := q.UpdatePOLineBilledQty(ctx, sqlc.UpdatePOLineBilledQtyParams{
		ID:             toPgUUID(lineID),
		BilledQuantity: toPgNumeric(delta),
	})
	return err
}

// ----------------------------------------------------------------------------
// Goods Receipts
// ----------------------------------------------------------------------------

func (r *PurchasingRepository) CreateGoodsReceipt(ctx context.Context, gr *domain.GoodsReceipt) (*domain.GoodsReceipt, error) {
	q := r.queries(ctx)
	row, err := q.CreateGoodsReceipt(ctx, sqlc.CreateGoodsReceiptParams{
		ID:                 toPgUUID(gr.ID),
		LegalEntityID:      toPgUUID(gr.LegalEntityID),
		ReceiptNumber:      gr.ReceiptNumber,
		PurchaseOrderID:    toPgUUID(gr.PurchaseOrderID),
		SupplierID:         toPgUUID(gr.SupplierID),
		WarehouseID:        toPgUUID(gr.WarehouseID),
		ReceiptDate:        toPgDate(gr.ReceiptDate),
		DeliveryNoteNumber: toPgTextPtr(gr.DeliveryNoteNumber),
		Status:             string(gr.Status),
		Notes:              toPgTextPtr(gr.Notes),
		ReceivedByUserID:   toPgUUIDPtr(gr.ReceivedByUserID),
		InspectedByUserID:  toPgUUIDPtr(gr.InspectedByUserID),
	})
	if err != nil {
		return nil, err
	}

	created := mapGR(row)
	for _, l := range gr.Lines {
		lineRow, err := q.CreateGoodsReceiptLine(ctx, sqlc.CreateGoodsReceiptLineParams{
			ID:                  toPgUUID(l.ID),
			GoodsReceiptID:      row.ID,
			LineNumber:          int32(l.LineNumber),
			PoLineID:            toPgUUID(l.POLineID),
			ItemID:              toPgUUID(l.ItemID),
			WarehouseID:         toPgUUID(l.WarehouseID),
			BinID:               toPgUUIDPtr(l.BinID),
			ReceivedQuantity:    toPgNumeric(l.ReceivedQuantity),
			AcceptedQuantity:    toPgNumeric(l.AcceptedQuantity),
			RejectedQuantity:    toPgNumeric(l.RejectedQuantity),
			UnitCost:            toPgNumeric(l.UnitCost),
			TotalCost:           toPgNumeric(l.TotalCost),
			InventoryMovementID: toPgUUIDPtr(l.InventoryMovementID),
			RejectionReason:     toPgTextPtr(l.RejectionReason),
		})
		if err != nil {
			return nil, err
		}
		created.Lines = append(created.Lines, mapGRLine(lineRow))
	}
	return created, nil
}

func (r *PurchasingRepository) GetGoodsReceipt(ctx context.Context, id uuid.UUID) (*domain.GoodsReceipt, error) {
	q := r.queries(ctx)
	row, err := q.GetGoodsReceipt(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	gr := mapGR(row)
	lines, err := q.ListGoodsReceiptLines(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	for _, l := range lines {
		gr.Lines = append(gr.Lines, mapGRLine(l))
	}
	return gr, nil
}

func (r *PurchasingRepository) GetGoodsReceiptByNumber(ctx context.Context, legalEntityID uuid.UUID, receiptNumber string) (*domain.GoodsReceipt, error) {
	q := r.queries(ctx)
	row, err := q.GetGoodsReceiptByNumber(ctx, sqlc.GetGoodsReceiptByNumberParams{
		LegalEntityID: toPgUUID(legalEntityID),
		ReceiptNumber: receiptNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGoodsReceiptNotFound
		}
		return nil, err
	}
	return r.GetGoodsReceipt(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) ListGoodsReceipts(ctx context.Context, legalEntityID uuid.UUID, poID, supplierID *uuid.UUID, status *domain.GRNStatus, limit, offset int) ([]*domain.GoodsReceipt, error) {
	q := r.queries(ctx)
	var s *string
	if status != nil {
		str := string(*status)
		s = &str
	}
	rows, err := q.ListGoodsReceipts(ctx, sqlc.ListGoodsReceiptsParams{
		LegalEntityID:   toPgUUID(legalEntityID),
		PurchaseOrderID: toPgUUIDPtr(poID),
		SupplierID:      toPgUUIDPtr(supplierID),
		Status:          toPgTextPtr(s),
		LimitCount:      int32(limit),
		OffsetCount:     int32(offset),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*domain.GoodsReceipt, 0, len(rows))
	for _, row := range rows {
		gr := mapGR(row)
		lines, _ := q.ListGoodsReceiptLines(ctx, row.ID)
		for _, l := range lines {
			gr.Lines = append(gr.Lines, mapGRLine(l))
		}
		res = append(res, gr)
	}
	return res, nil
}

func (r *PurchasingRepository) UpdateGoodsReceiptStatus(ctx context.Context, id uuid.UUID, status domain.GRNStatus) (*domain.GoodsReceipt, error) {
	q := r.queries(ctx)
	row, err := q.UpdateGoodsReceiptStatus(ctx, sqlc.UpdateGoodsReceiptStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetGoodsReceipt(ctx, toUUID(row.ID))
}

func (r *PurchasingRepository) UpdateGoodsReceiptLineMovement(ctx context.Context, lineID uuid.UUID, movementID uuid.UUID) error {
	q := r.queries(ctx)
	_, err := q.UpdateGoodsReceiptLineMovement(ctx, sqlc.UpdateGoodsReceiptLineMovementParams{
		ID:                  toPgUUID(lineID),
		InventoryMovementID: toPgUUID(movementID),
	})
	return err
}

// ----------------------------------------------------------------------------
// Internal Mapping Helpers
// ----------------------------------------------------------------------------

func mapPR(row sqlc.PurchasingPurchaseRequests) *domain.PurchaseRequest {
	var reqDate *time.Time
	if row.RequiredByDate.Valid {
		t := row.RequiredByDate.Time
		reqDate = &t
	}
	return &domain.PurchaseRequest{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		PRNumber:          row.PrNumber,
		RequestedByUserID: toUUID(row.RequestedByUserID),
		DepartmentID:      toUUIDPtr(row.DepartmentID),
		RequestDate:       fromPgDate(row.RequestDate),
		RequiredByDate:    reqDate,
		Status:            domain.PRStatus(row.Status),
		Purpose:           fromPgTextPtr(row.Purpose),
		Notes:             fromPgTextPtr(row.Notes),
		EstimatedTotal:    fromPgNumeric(row.EstimatedTotal),
		CurrencyCode:      row.CurrencyCode,
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapPRLine(row sqlc.PurchasingPurchaseRequestLines) domain.PurchaseRequestLine {
	return domain.PurchaseRequestLine{
		ID:                  toUUID(row.ID),
		PurchaseRequestID:   toUUID(row.PurchaseRequestID),
		LineNumber:          int(row.LineNumber),
		ItemID:              toUUIDPtr(row.ItemID),
		Description:         row.Description,
		UOM:                 row.Uom,
		Quantity:            fromPgNumeric(row.Quantity),
		OrderedQuantity:     fromPgNumeric(row.OrderedQuantity),
		EstimatedUnitCost:   fromPgNumeric(row.EstimatedUnitCost),
		EstimatedTotalCost:  fromPgNumeric(row.EstimatedTotalCost),
		SuggestedSupplierID: toUUIDPtr(row.SuggestedSupplierID),
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}
}

func mapRFQ(row sqlc.PurchasingRfqs) *domain.RFQ {
	return &domain.RFQ{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		RFQNumber:         row.RfqNumber,
		PurchaseRequestID: toUUIDPtr(row.PurchaseRequestID),
		Title:             row.Title,
		RFQDate:           fromPgDate(row.RfqDate),
		ClosingDate:       fromPgDate(row.ClosingDate),
		Status:            domain.RFQStatus(row.Status),
		Notes:             fromPgTextPtr(row.Notes),
		CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
		CreatedAt:         fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapRFQLine(row sqlc.PurchasingRfqLines) domain.RFQLine {
	return domain.RFQLine{
		ID:              toUUID(row.ID),
		RFQID:           toUUID(row.RfqID),
		LineNumber:      int(row.LineNumber),
		PRLineID:        toUUIDPtr(row.PrLineID),
		ItemID:          toUUIDPtr(row.ItemID),
		Description:     row.Description,
		UOM:             row.Uom,
		Quantity:        fromPgNumeric(row.Quantity),
		TargetUnitPrice: fromPgNumericPtr(row.TargetUnitPrice),
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
	}
}

func mapSupplierQuote(row sqlc.PurchasingSupplierQuotes) *domain.SupplierQuote {
	return &domain.SupplierQuote{
		ID:                     toUUID(row.ID),
		LegalEntityID:          toUUID(row.LegalEntityID),
		RFQID:                  toUUIDPtr(row.RfqID),
		SupplierID:             toUUID(row.SupplierID),
		SupplierQuoteReference: row.SupplierQuoteReference,
		QuoteDate:              fromPgDate(row.QuoteDate),
		ValidUntil:             fromPgDate(row.ValidUntil),
		CurrencyCode:           row.CurrencyCode,
		ExchangeRate:           fromPgNumeric(row.ExchangeRate),
		PaymentTermsID:         toUUIDPtr(row.PaymentTermsID),
		Subtotal:               fromPgNumeric(row.Subtotal),
		TaxAmount:              fromPgNumeric(row.TaxAmount),
		TotalAmount:            fromPgNumeric(row.TotalAmount),
		Status:                 domain.QuoteStatus(row.Status),
		Notes:                  fromPgTextPtr(row.Notes),
		CreatedAt:              fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:              fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapSupplierQuoteLine(row sqlc.PurchasingSupplierQuoteLines) domain.SupplierQuoteLine {
	return domain.SupplierQuoteLine{
		ID:              toUUID(row.ID),
		QuoteID:         toUUID(row.QuoteID),
		LineNumber:      int(row.LineNumber),
		RFQLineID:       toUUIDPtr(row.RfqLineID),
		ItemID:          toUUIDPtr(row.ItemID),
		Description:     row.Description,
		Quantity:        fromPgNumeric(row.Quantity),
		UnitPrice:       fromPgNumeric(row.UnitPrice),
		DiscountPercent: fromPgNumeric(row.DiscountPercent),
		TaxRatePercent:  fromPgNumeric(row.TaxRatePercent),
		TaxAmount:       fromPgNumeric(row.TaxAmount),
		TotalAmount:     fromPgNumeric(row.TotalAmount),
		IsAwarded:       row.IsAwarded,
		CreatedAt:       fromPgTimestamptz(row.CreatedAt),
	}
}

func mapPO(row sqlc.PurchasingPurchaseOrders) *domain.PurchaseOrder {
	var expDate *time.Time
	if row.ExpectedDeliveryDate.Valid {
		t := row.ExpectedDeliveryDate.Time
		expDate = &t
	}
	var appAt, issAt *time.Time
	if row.ApprovedAt.Valid {
		t := row.ApprovedAt.Time
		appAt = &t
	}
	if row.IssuedAt.Valid {
		t := row.IssuedAt.Time
		issAt = &t
	}
	return &domain.PurchaseOrder{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		PONumber:             row.PoNumber,
		PurchaseRequestID:    toUUIDPtr(row.PurchaseRequestID),
		RFQID:                toUUIDPtr(row.RfqID),
		SupplierQuoteID:      toUUIDPtr(row.SupplierQuoteID),
		SupplierID:           toUUID(row.SupplierID),
		OrderDate:            fromPgDate(row.OrderDate),
		ExpectedDeliveryDate: expDate,
		PaymentTermsID:       toUUIDPtr(row.PaymentTermsID),
		DeliveryWarehouseID:  toUUIDPtr(row.DeliveryWarehouseID),
		CurrencyCode:         row.CurrencyCode,
		ExchangeRate:         fromPgNumeric(row.ExchangeRate),
		Subtotal:             fromPgNumeric(row.Subtotal),
		DiscountAmount:       fromPgNumeric(row.DiscountAmount),
		TaxableAmount:        fromPgNumeric(row.TaxableAmount),
		TaxAmount:            fromPgNumeric(row.TaxAmount),
		TotalAmount:          fromPgNumeric(row.TotalAmount),
		Status:               domain.POStatus(row.Status),
		Notes:                fromPgTextPtr(row.Notes),
		CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
		ApprovedByUserID:     toUUIDPtr(row.ApprovedByUserID),
		ApprovedAt:           appAt,
		IssuedAt:             issAt,
		CreatedAt:            fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:            fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapPOLine(row sqlc.PurchasingPurchaseOrderLines) domain.PurchaseOrderLine {
	return domain.PurchaseOrderLine{
		ID:               toUUID(row.ID),
		PurchaseOrderID:  toUUID(row.PurchaseOrderID),
		LineNumber:       int(row.LineNumber),
		PRLineID:         toUUIDPtr(row.PrLineID),
		ItemID:           toUUID(row.ItemID),
		Description:      row.Description,
		UOM:              row.Uom,
		Quantity:         fromPgNumeric(row.Quantity),
		ReceivedQuantity: fromPgNumeric(row.ReceivedQuantity),
		BilledQuantity:   fromPgNumeric(row.BilledQuantity),
		UnitPrice:        fromPgNumeric(row.UnitPrice),
		DiscountPercent:  fromPgNumeric(row.DiscountPercent),
		DiscountAmount:   fromPgNumeric(row.DiscountAmount),
		TaxableAmount:    fromPgNumeric(row.TaxableAmount),
		TaxCode:          row.TaxCode,
		TaxRatePercent:   fromPgNumeric(row.TaxRatePercent),
		TaxAmount:        fromPgNumeric(row.TaxAmount),
		TotalAmount:      fromPgNumeric(row.TotalAmount),
		ExpenseAccountID: toUUIDPtr(row.ExpenseAccountID),
		CreatedAt:        fromPgTimestamptz(row.CreatedAt),
	}
}

func mapGR(row sqlc.PurchasingGoodsReceipts) *domain.GoodsReceipt {
	return &domain.GoodsReceipt{
		ID:                 toUUID(row.ID),
		LegalEntityID:      toUUID(row.LegalEntityID),
		ReceiptNumber:      row.ReceiptNumber,
		PurchaseOrderID:    toUUID(row.PurchaseOrderID),
		SupplierID:         toUUID(row.SupplierID),
		WarehouseID:        toUUID(row.WarehouseID),
		ReceiptDate:        fromPgDate(row.ReceiptDate),
		DeliveryNoteNumber: fromPgTextPtr(row.DeliveryNoteNumber),
		Status:             domain.GRNStatus(row.Status),
		JournalEntryID:     toUUIDPtr(row.JournalEntryID),
		Notes:              fromPgTextPtr(row.Notes),
		ReceivedByUserID:   toUUIDPtr(row.ReceivedByUserID),
		InspectedByUserID:  toUUIDPtr(row.InspectedByUserID),
		CreatedAt:          fromPgTimestamptz(row.CreatedAt),
		UpdatedAt:          fromPgTimestamptz(row.UpdatedAt),
	}
}

func mapGRLine(row sqlc.PurchasingGoodsReceiptLines) domain.GoodsReceiptLine {
	return domain.GoodsReceiptLine{
		ID:                  toUUID(row.ID),
		GoodsReceiptID:      toUUID(row.GoodsReceiptID),
		LineNumber:          int(row.LineNumber),
		POLineID:            toUUID(row.PoLineID),
		ItemID:              toUUID(row.ItemID),
		WarehouseID:         toUUID(row.WarehouseID),
		BinID:               toUUIDPtr(row.BinID),
		ReceivedQuantity:    fromPgNumeric(row.ReceivedQuantity),
		AcceptedQuantity:    fromPgNumeric(row.AcceptedQuantity),
		RejectedQuantity:    fromPgNumeric(row.RejectedQuantity),
		UnitCost:            fromPgNumeric(row.UnitCost),
		TotalCost:           fromPgNumeric(row.TotalCost),
		InventoryMovementID: toUUIDPtr(row.InventoryMovementID),
		RejectionReason:     fromPgTextPtr(row.RejectionReason),
		CreatedAt:           fromPgTimestamptz(row.CreatedAt),
	}
}
