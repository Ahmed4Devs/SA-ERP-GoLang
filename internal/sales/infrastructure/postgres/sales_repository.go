package postgres

import (
	"context"
	"errors"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"
	"erp/internal/sales/domain"
	"erp/internal/sales/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type salesRepository struct {
	db *database.DB
}

func NewSalesRepository(db *database.DB) ports.Repository {
	return &salesRepository{db: db}
}

// ----------------------------------------------------------------------------
// Quotations
// ----------------------------------------------------------------------------

func (r *salesRepository) CreateQuotation(ctx context.Context, q *domain.Quotation) (*domain.Quotation, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateQuotation(ctx, sqlc.CreateQuotationParams{
		ID:              toPgUUID(q.ID),
		LegalEntityID:   toPgUUID(q.LegalEntityID),
		QuotationNumber: q.QuotationNumber,
		CustomerID:      toPgUUID(q.CustomerID),
		CustomerNameEn:  q.CustomerNameEn,
		CustomerNameAr:  q.CustomerNameAr,
		QuoteDate:       toPgDate(q.QuoteDate),
		ValidUntil:      toPgDate(q.ValidUntil),
		CurrencyCode:    q.CurrencyCode,
		ExchangeRate:    toPgNumeric(q.ExchangeRate),
		Subtotal:        toPgNumeric(q.Subtotal),
		DiscountAmount:  toPgNumeric(q.DiscountAmount),
		TaxableAmount:   toPgNumeric(q.TaxableAmount),
		TaxAmount:       toPgNumeric(q.TaxAmount),
		TotalAmount:     toPgNumeric(q.TotalAmount),
		Status:          string(q.Status),
		Notes:           toPgTextPtr(q.Notes),
		CreatedByUserID: toPgUUIDPtr(q.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	result := &domain.Quotation{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		QuotationNumber:  row.QuotationNumber,
		CustomerID:       toUUID(row.CustomerID),
		CustomerNameEn:   row.CustomerNameEn,
		CustomerNameAr:   row.CustomerNameAr,
		QuoteDate:        fromPgDate(row.QuoteDate),
		ValidUntil:       fromPgDate(row.ValidUntil),
		CurrencyCode:     row.CurrencyCode,
		ExchangeRate:     fromPgNumeric(row.ExchangeRate),
		Subtotal:         fromPgNumeric(row.Subtotal),
		DiscountAmount:   fromPgNumeric(row.DiscountAmount),
		TaxableAmount:    fromPgNumeric(row.TaxableAmount),
		TaxAmount:        fromPgNumeric(row.TaxAmount),
		TotalAmount:      fromPgNumeric(row.TotalAmount),
		Status:           domain.QuotationStatus(row.Status),
		ConvertedOrderID: toUUIDPtr(row.ConvertedOrderID),
		Notes:            fromPgTextPtr(row.Notes),
		CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
		CreatedAt:        fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:        fromPgTimestamptzVal(row.UpdatedAt),
	}

	for _, l := range q.Lines {
		lRow, err := queries.CreateQuotationLine(ctx, sqlc.CreateQuotationLineParams{
			ID:              toPgUUID(l.ID),
			QuotationID:     row.ID,
			LineNumber:      int32(l.LineNumber),
			ItemID:          toPgUUID(l.ItemID),
			DescriptionEn:   l.DescriptionEn,
			DescriptionAr:   l.DescriptionAr,
			Uom:             l.UOM,
			Quantity:        toPgNumeric(l.Quantity),
			UnitPrice:       toPgNumeric(l.UnitPrice),
			DiscountPercent: toPgNumeric(l.DiscountPercent),
			DiscountAmount:  toPgNumeric(l.DiscountAmount),
			TaxableAmount:   toPgNumeric(l.TaxableAmount),
			TaxCode:         l.TaxCode,
			TaxRatePercent:  toPgNumeric(l.TaxRatePercent),
			TaxAmount:       toPgNumeric(l.TaxAmount),
			TotalAmount:     toPgNumeric(l.TotalAmount),
		})
		if err != nil {
			return nil, err
		}

		result.Lines = append(result.Lines, domain.QuotationLine{
			ID:              toUUID(lRow.ID),
			QuotationID:     toUUID(lRow.QuotationID),
			LineNumber:      int(lRow.LineNumber),
			ItemID:          toUUID(lRow.ItemID),
			DescriptionEn:   lRow.DescriptionEn,
			DescriptionAr:   lRow.DescriptionAr,
			UOM:             lRow.Uom,
			Quantity:        fromPgNumeric(lRow.Quantity),
			UnitPrice:       fromPgNumeric(lRow.UnitPrice),
			DiscountPercent: fromPgNumeric(lRow.DiscountPercent),
			DiscountAmount:  fromPgNumeric(lRow.DiscountAmount),
			TaxableAmount:   fromPgNumeric(lRow.TaxableAmount),
			TaxCode:         lRow.TaxCode,
			TaxRatePercent:  fromPgNumeric(lRow.TaxRatePercent),
			TaxAmount:       fromPgNumeric(lRow.TaxAmount),
			TotalAmount:     fromPgNumeric(lRow.TotalAmount),
		})
	}

	return result, nil
}

func (r *salesRepository) GetQuotation(ctx context.Context, id uuid.UUID) (*domain.Quotation, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetQuotation(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrQuotationNotFound
		}
		return nil, err
	}

	q := &domain.Quotation{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		QuotationNumber:  row.QuotationNumber,
		CustomerID:       toUUID(row.CustomerID),
		CustomerNameEn:   row.CustomerNameEn,
		CustomerNameAr:   row.CustomerNameAr,
		QuoteDate:        fromPgDate(row.QuoteDate),
		ValidUntil:       fromPgDate(row.ValidUntil),
		CurrencyCode:     row.CurrencyCode,
		ExchangeRate:     fromPgNumeric(row.ExchangeRate),
		Subtotal:         fromPgNumeric(row.Subtotal),
		DiscountAmount:   fromPgNumeric(row.DiscountAmount),
		TaxableAmount:    fromPgNumeric(row.TaxableAmount),
		TaxAmount:        fromPgNumeric(row.TaxAmount),
		TotalAmount:      fromPgNumeric(row.TotalAmount),
		Status:           domain.QuotationStatus(row.Status),
		ConvertedOrderID: toUUIDPtr(row.ConvertedOrderID),
		Notes:            fromPgTextPtr(row.Notes),
		CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
		CreatedAt:        fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:        fromPgTimestamptzVal(row.UpdatedAt),
	}

	lines, err := queries.ListQuotationLines(ctx, row.ID)
	if err == nil {
		for _, l := range lines {
			q.Lines = append(q.Lines, domain.QuotationLine{
				ID:              toUUID(l.ID),
				QuotationID:     toUUID(l.QuotationID),
				LineNumber:      int(l.LineNumber),
				ItemID:          toUUID(l.ItemID),
				DescriptionEn:   l.DescriptionEn,
				DescriptionAr:   l.DescriptionAr,
				UOM:             l.Uom,
				Quantity:        fromPgNumeric(l.Quantity),
				UnitPrice:       fromPgNumeric(l.UnitPrice),
				DiscountPercent: fromPgNumeric(l.DiscountPercent),
				DiscountAmount:  fromPgNumeric(l.DiscountAmount),
				TaxableAmount:   fromPgNumeric(l.TaxableAmount),
				TaxCode:         l.TaxCode,
				TaxRatePercent:  fromPgNumeric(l.TaxRatePercent),
				TaxAmount:       fromPgNumeric(l.TaxAmount),
				TotalAmount:     fromPgNumeric(l.TotalAmount),
			})
		}
	}

	return q, nil
}

func (r *salesRepository) ListQuotations(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.QuotationStatus, limit, offset int) ([]*domain.Quotation, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var statusText pgtype.Text
	if status != nil {
		statusText = toPgText(string(*status))
	}

	rows, err := queries.ListQuotations(ctx, sqlc.ListQuotationsParams{
		LegalEntityID: toPgUUID(legalEntityID),
		CustomerID:    toPgUUIDPtr(customerID),
		Status:        statusText,
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []*domain.Quotation
	for _, row := range rows {
		res = append(res, &domain.Quotation{
			ID:               toUUID(row.ID),
			LegalEntityID:    toUUID(row.LegalEntityID),
			QuotationNumber:  row.QuotationNumber,
			CustomerID:       toUUID(row.CustomerID),
			CustomerNameEn:   row.CustomerNameEn,
			CustomerNameAr:   row.CustomerNameAr,
			QuoteDate:        fromPgDate(row.QuoteDate),
			ValidUntil:       fromPgDate(row.ValidUntil),
			CurrencyCode:     row.CurrencyCode,
			ExchangeRate:     fromPgNumeric(row.ExchangeRate),
			Subtotal:         fromPgNumeric(row.Subtotal),
			DiscountAmount:   fromPgNumeric(row.DiscountAmount),
			TaxableAmount:    fromPgNumeric(row.TaxableAmount),
			TaxAmount:        fromPgNumeric(row.TaxAmount),
			TotalAmount:      fromPgNumeric(row.TotalAmount),
			Status:           domain.QuotationStatus(row.Status),
			ConvertedOrderID: toUUIDPtr(row.ConvertedOrderID),
			Notes:            fromPgTextPtr(row.Notes),
			CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
			CreatedAt:        fromPgTimestamptzVal(row.CreatedAt),
			UpdatedAt:        fromPgTimestamptzVal(row.UpdatedAt),
		})
	}
	return res, nil
}

func (r *salesRepository) UpdateQuotationStatus(ctx context.Context, id uuid.UUID, status domain.QuotationStatus, convertedOrderID *uuid.UUID) (*domain.Quotation, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateQuotationStatus(ctx, sqlc.UpdateQuotationStatusParams{
		ID:               toPgUUID(id),
		Status:           string(status),
		ConvertedOrderID: toPgUUIDPtr(convertedOrderID),
	})
	if err != nil {
		return nil, err
	}

	return r.GetQuotation(ctx, toUUID(row.ID))
}

// ----------------------------------------------------------------------------
// Orders
// ----------------------------------------------------------------------------

func (r *salesRepository) CreateOrder(ctx context.Context, o *domain.SalesOrder) (*domain.SalesOrder, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateOrder(ctx, sqlc.CreateOrderParams{
		ID:                   toPgUUID(o.ID),
		LegalEntityID:        toPgUUID(o.LegalEntityID),
		OrderNumber:          o.OrderNumber,
		CustomerID:           toPgUUID(o.CustomerID),
		QuotationID:          toPgUUIDPtr(o.QuotationID),
		CustomerPoNumber:     toPgTextPtr(o.CustomerPONumber),
		OrderDate:            toPgDate(o.OrderDate),
		ExpectedDeliveryDate: toPgDatePtr(o.ExpectedDeliveryDate),
		PaymentTermsID:       toPgUUIDPtr(o.PaymentTermsID),
		CurrencyCode:         o.CurrencyCode,
		ExchangeRate:         toPgNumeric(o.ExchangeRate),
		Subtotal:             toPgNumeric(o.Subtotal),
		DiscountAmount:       toPgNumeric(o.DiscountAmount),
		TaxableAmount:        toPgNumeric(o.TaxableAmount),
		TaxAmount:            toPgNumeric(o.TaxAmount),
		TotalAmount:          toPgNumeric(o.TotalAmount),
		Status:               string(o.Status),
		Notes:                toPgTextPtr(o.Notes),
		CreatedByUserID:      toPgUUIDPtr(o.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	result := &domain.SalesOrder{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		OrderNumber:          row.OrderNumber,
		CustomerID:           toUUID(row.CustomerID),
		QuotationID:          toUUIDPtr(row.QuotationID),
		CustomerPONumber:     fromPgTextPtr(row.CustomerPoNumber),
		OrderDate:            fromPgDate(row.OrderDate),
		ExpectedDeliveryDate: fromPgDatePtr(row.ExpectedDeliveryDate),
		PaymentTermsID:       toUUIDPtr(row.PaymentTermsID),
		CurrencyCode:         row.CurrencyCode,
		ExchangeRate:         fromPgNumeric(row.ExchangeRate),
		Subtotal:             fromPgNumeric(row.Subtotal),
		DiscountAmount:       fromPgNumeric(row.DiscountAmount),
		TaxableAmount:        fromPgNumeric(row.TaxableAmount),
		TaxAmount:            fromPgNumeric(row.TaxAmount),
		TotalAmount:          fromPgNumeric(row.TotalAmount),
		Status:               domain.OrderStatus(row.Status),
		Notes:                fromPgTextPtr(row.Notes),
		CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
		CreatedAt:            fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:            fromPgTimestamptzVal(row.UpdatedAt),
	}

	for _, l := range o.Lines {
		lRow, err := queries.CreateOrderLine(ctx, sqlc.CreateOrderLineParams{
			ID:              toPgUUID(l.ID),
			OrderID:         row.ID,
			LineNumber:      int32(l.LineNumber),
			ItemID:          toPgUUID(l.ItemID),
			DescriptionEn:   l.DescriptionEn,
			DescriptionAr:   l.DescriptionAr,
			Uom:             l.UOM,
			Quantity:        toPgNumeric(l.Quantity),
			UnitPrice:       toPgNumeric(l.UnitPrice),
			DiscountPercent: toPgNumeric(l.DiscountPercent),
			DiscountAmount:  toPgNumeric(l.DiscountAmount),
			TaxableAmount:   toPgNumeric(l.TaxableAmount),
			TaxCode:         l.TaxCode,
			TaxRatePercent:  toPgNumeric(l.TaxRatePercent),
			TaxAmount:       toPgNumeric(l.TaxAmount),
			TotalAmount:     toPgNumeric(l.TotalAmount),
		})
		if err != nil {
			return nil, err
		}

		result.Lines = append(result.Lines, domain.SalesOrderLine{
			ID:                toUUID(lRow.ID),
			OrderID:           toUUID(lRow.OrderID),
			LineNumber:        int(lRow.LineNumber),
			ItemID:            toUUID(lRow.ItemID),
			DescriptionEn:     lRow.DescriptionEn,
			DescriptionAr:     lRow.DescriptionAr,
			UOM:               lRow.Uom,
			Quantity:          fromPgNumeric(lRow.Quantity),
			DeliveredQuantity: fromPgNumeric(lRow.DeliveredQuantity),
			InvoicedQuantity:  fromPgNumeric(lRow.InvoicedQuantity),
			UnitPrice:         fromPgNumeric(lRow.UnitPrice),
			DiscountPercent:   fromPgNumeric(lRow.DiscountPercent),
			DiscountAmount:    fromPgNumeric(lRow.DiscountAmount),
			TaxableAmount:     fromPgNumeric(lRow.TaxableAmount),
			TaxCode:           lRow.TaxCode,
			TaxRatePercent:    fromPgNumeric(lRow.TaxRatePercent),
			TaxAmount:         fromPgNumeric(lRow.TaxAmount),
			TotalAmount:       fromPgNumeric(lRow.TotalAmount),
		})
	}

	return result, nil
}

func (r *salesRepository) GetOrder(ctx context.Context, id uuid.UUID) (*domain.SalesOrder, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetOrder(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	order := &domain.SalesOrder{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		OrderNumber:          row.OrderNumber,
		CustomerID:           toUUID(row.CustomerID),
		QuotationID:          toUUIDPtr(row.QuotationID),
		CustomerPONumber:     fromPgTextPtr(row.CustomerPoNumber),
		OrderDate:            fromPgDate(row.OrderDate),
		ExpectedDeliveryDate: fromPgDatePtr(row.ExpectedDeliveryDate),
		PaymentTermsID:       toUUIDPtr(row.PaymentTermsID),
		CurrencyCode:         row.CurrencyCode,
		ExchangeRate:         fromPgNumeric(row.ExchangeRate),
		Subtotal:             fromPgNumeric(row.Subtotal),
		DiscountAmount:       fromPgNumeric(row.DiscountAmount),
		TaxableAmount:        fromPgNumeric(row.TaxableAmount),
		TaxAmount:            fromPgNumeric(row.TaxAmount),
		TotalAmount:          fromPgNumeric(row.TotalAmount),
		Status:               domain.OrderStatus(row.Status),
		Notes:                fromPgTextPtr(row.Notes),
		CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
		CreatedAt:            fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:            fromPgTimestamptzVal(row.UpdatedAt),
	}

	lines, err := queries.ListOrderLines(ctx, row.ID)
	if err == nil {
		for _, l := range lines {
			order.Lines = append(order.Lines, domain.SalesOrderLine{
				ID:                toUUID(l.ID),
				OrderID:           toUUID(l.OrderID),
				LineNumber:        int(l.LineNumber),
				ItemID:            toUUID(l.ItemID),
				DescriptionEn:     l.DescriptionEn,
				DescriptionAr:     l.DescriptionAr,
				UOM:               l.Uom,
				Quantity:          fromPgNumeric(l.Quantity),
				DeliveredQuantity: fromPgNumeric(l.DeliveredQuantity),
				InvoicedQuantity:  fromPgNumeric(l.InvoicedQuantity),
				UnitPrice:         fromPgNumeric(l.UnitPrice),
				DiscountPercent:   fromPgNumeric(l.DiscountPercent),
				DiscountAmount:    fromPgNumeric(l.DiscountAmount),
				TaxableAmount:     fromPgNumeric(l.TaxableAmount),
				TaxCode:           l.TaxCode,
				TaxRatePercent:    fromPgNumeric(l.TaxRatePercent),
				TaxAmount:         fromPgNumeric(l.TaxAmount),
				TotalAmount:       fromPgNumeric(l.TotalAmount),
			})
		}
	}

	return order, nil
}

func (r *salesRepository) ListOrders(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.OrderStatus, limit, offset int) ([]*domain.SalesOrder, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var statusText pgtype.Text
	if status != nil {
		statusText = toPgText(string(*status))
	}

	rows, err := queries.ListOrders(ctx, sqlc.ListOrdersParams{
		LegalEntityID: toPgUUID(legalEntityID),
		CustomerID:    toPgUUIDPtr(customerID),
		Status:        statusText,
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []*domain.SalesOrder
	for _, row := range rows {
		res = append(res, &domain.SalesOrder{
			ID:                   toUUID(row.ID),
			LegalEntityID:        toUUID(row.LegalEntityID),
			OrderNumber:          row.OrderNumber,
			CustomerID:           toUUID(row.CustomerID),
			QuotationID:          toUUIDPtr(row.QuotationID),
			CustomerPONumber:     fromPgTextPtr(row.CustomerPoNumber),
			OrderDate:            fromPgDate(row.OrderDate),
			ExpectedDeliveryDate: fromPgDatePtr(row.ExpectedDeliveryDate),
			PaymentTermsID:       toUUIDPtr(row.PaymentTermsID),
			CurrencyCode:         row.CurrencyCode,
			ExchangeRate:         fromPgNumeric(row.ExchangeRate),
			Subtotal:             fromPgNumeric(row.Subtotal),
			DiscountAmount:       fromPgNumeric(row.DiscountAmount),
			TaxableAmount:        fromPgNumeric(row.TaxableAmount),
			TaxAmount:            fromPgNumeric(row.TaxAmount),
			TotalAmount:          fromPgNumeric(row.TotalAmount),
			Status:               domain.OrderStatus(row.Status),
			Notes:                fromPgTextPtr(row.Notes),
			CreatedByUserID:      toUUIDPtr(row.CreatedByUserID),
			CreatedAt:            fromPgTimestamptzVal(row.CreatedAt),
			UpdatedAt:            fromPgTimestamptzVal(row.UpdatedAt),
		})
	}
	return res, nil
}

func (r *salesRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) (*domain.SalesOrder, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}

	return r.GetOrder(ctx, toUUID(row.ID))
}

func (r *salesRepository) UpdateOrderLineDeliveredQty(ctx context.Context, lineID uuid.UUID, qtyDelta decimal.Decimal) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	_, err := queries.UpdateOrderLineDeliveredQty(ctx, sqlc.UpdateOrderLineDeliveredQtyParams{
		ID:                toPgUUID(lineID),
		DeliveredQuantity: toPgNumeric(qtyDelta),
	})
	return err
}

func (r *salesRepository) UpdateOrderLineInvoicedQty(ctx context.Context, lineID uuid.UUID, qtyDelta decimal.Decimal) error {
	queries := sqlc.New(r.db.DBTX(ctx))
	_, err := queries.UpdateOrderLineInvoicedQty(ctx, sqlc.UpdateOrderLineInvoicedQtyParams{
		ID:               toPgUUID(lineID),
		InvoicedQuantity: toPgNumeric(qtyDelta),
	})
	return err
}

// ----------------------------------------------------------------------------
// Deliveries
// ----------------------------------------------------------------------------

func (r *salesRepository) CreateDelivery(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateDelivery(ctx, sqlc.CreateDeliveryParams{
		ID:                toPgUUID(d.ID),
		LegalEntityID:     toPgUUID(d.LegalEntityID),
		DeliveryNumber:    d.DeliveryNumber,
		OrderID:           toPgUUIDPtr(d.OrderID),
		CustomerID:        toPgUUID(d.CustomerID),
		WarehouseID:       toPgUUID(d.WarehouseID),
		DeliveryDate:      toPgDate(d.DeliveryDate),
		Status:            string(d.Status),
		ShippingReference: toPgTextPtr(d.ShippingReference),
		Notes:             toPgTextPtr(d.Notes),
		CreatedByUserID:   toPgUUIDPtr(d.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	result := &domain.Delivery{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		DeliveryNumber:    row.DeliveryNumber,
		OrderID:           toUUIDPtr(row.OrderID),
		CustomerID:        toUUID(row.CustomerID),
		WarehouseID:       toUUID(row.WarehouseID),
		DeliveryDate:      fromPgDate(row.DeliveryDate),
		Status:            domain.DeliveryStatus(row.Status),
		ShippingReference: fromPgTextPtr(row.ShippingReference),
		Notes:             fromPgTextPtr(row.Notes),
		CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
		CreatedAt:         fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptzVal(row.UpdatedAt),
	}

	for _, l := range d.Lines {
		lRow, err := queries.CreateDeliveryLine(ctx, sqlc.CreateDeliveryLineParams{
			ID:                  toPgUUID(l.ID),
			DeliveryID:          row.ID,
			LineNumber:          int32(l.LineNumber),
			OrderLineID:         toPgUUIDPtr(l.OrderLineID),
			ItemID:              toPgUUID(l.ItemID),
			Uom:                 l.UOM,
			Quantity:            toPgNumeric(l.Quantity),
			WarehouseID:         toPgUUID(l.WarehouseID),
			BinID:               toPgUUIDPtr(l.BinID),
			InventoryMovementID: toPgUUIDPtr(l.InventoryMovementID),
			UnitCost:            toPgNumeric(l.UnitCost),
			TotalCost:           toPgNumeric(l.TotalCost),
		})
		if err != nil {
			return nil, err
		}

		result.Lines = append(result.Lines, domain.DeliveryLine{
			ID:                  toUUID(lRow.ID),
			DeliveryID:          toUUID(lRow.DeliveryID),
			LineNumber:          int(lRow.LineNumber),
			OrderLineID:         toUUIDPtr(lRow.OrderLineID),
			ItemID:              toUUID(lRow.ItemID),
			UOM:                 lRow.Uom,
			Quantity:            fromPgNumeric(lRow.Quantity),
			WarehouseID:         toUUID(lRow.WarehouseID),
			BinID:               toUUIDPtr(lRow.BinID),
			InventoryMovementID: toUUIDPtr(lRow.InventoryMovementID),
			UnitCost:            fromPgNumeric(lRow.UnitCost),
			TotalCost:           fromPgNumeric(lRow.TotalCost),
		})
	}

	return result, nil
}

func (r *salesRepository) GetDelivery(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetDelivery(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDeliveryNotFound
		}
		return nil, err
	}

	d := &domain.Delivery{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		DeliveryNumber:    row.DeliveryNumber,
		OrderID:           toUUIDPtr(row.OrderID),
		CustomerID:        toUUID(row.CustomerID),
		WarehouseID:       toUUID(row.WarehouseID),
		DeliveryDate:      fromPgDate(row.DeliveryDate),
		Status:            domain.DeliveryStatus(row.Status),
		ShippingReference: fromPgTextPtr(row.ShippingReference),
		Notes:             fromPgTextPtr(row.Notes),
		CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
		CreatedAt:         fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptzVal(row.UpdatedAt),
	}

	lines, err := queries.ListDeliveryLines(ctx, row.ID)
	if err == nil {
		for _, l := range lines {
			d.Lines = append(d.Lines, domain.DeliveryLine{
				ID:                  toUUID(l.ID),
				DeliveryID:          toUUID(l.DeliveryID),
				LineNumber:          int(l.LineNumber),
				OrderLineID:         toUUIDPtr(l.OrderLineID),
				ItemID:              toUUID(l.ItemID),
				UOM:                 l.Uom,
				Quantity:            fromPgNumeric(l.Quantity),
				WarehouseID:         toUUID(l.WarehouseID),
				BinID:               toUUIDPtr(l.BinID),
				InventoryMovementID: toUUIDPtr(l.InventoryMovementID),
				UnitCost:            fromPgNumeric(l.UnitCost),
				TotalCost:           fromPgNumeric(l.TotalCost),
			})
		}
	}

	return d, nil
}

func (r *salesRepository) ListDeliveries(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, orderID *uuid.UUID, status *domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var statusText pgtype.Text
	if status != nil {
		statusText = toPgText(string(*status))
	}

	rows, err := queries.ListDeliveries(ctx, sqlc.ListDeliveriesParams{
		LegalEntityID: toPgUUID(legalEntityID),
		CustomerID:    toPgUUIDPtr(customerID),
		OrderID:       toPgUUIDPtr(orderID),
		Status:        statusText,
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []*domain.Delivery
	for _, row := range rows {
		res = append(res, &domain.Delivery{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			DeliveryNumber:    row.DeliveryNumber,
			OrderID:           toUUIDPtr(row.OrderID),
			CustomerID:        toUUID(row.CustomerID),
			WarehouseID:       toUUID(row.WarehouseID),
			DeliveryDate:      fromPgDate(row.DeliveryDate),
			Status:            domain.DeliveryStatus(row.Status),
			ShippingReference: fromPgTextPtr(row.ShippingReference),
			Notes:             fromPgTextPtr(row.Notes),
			CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
			CreatedAt:         fromPgTimestamptzVal(row.CreatedAt),
			UpdatedAt:         fromPgTimestamptzVal(row.UpdatedAt),
		})
	}
	return res, nil
}

func (r *salesRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status domain.DeliveryStatus) (*domain.Delivery, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateDeliveryStatus(ctx, sqlc.UpdateDeliveryStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}

	return r.GetDelivery(ctx, toUUID(row.ID))
}

// ----------------------------------------------------------------------------
// Invoices
// ----------------------------------------------------------------------------

func (r *salesRepository) CreateInvoice(ctx context.Context, inv *domain.SalesInvoice) (*domain.SalesInvoice, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var addrJson []byte
	if inv.CustomerAddressJSON != nil && *inv.CustomerAddressJSON != "" {
		addrJson = []byte(*inv.CustomerAddressJSON)
	}

	row, err := queries.CreateInvoice(ctx, sqlc.CreateInvoiceParams{
		ID:                  toPgUUID(inv.ID),
		LegalEntityID:       toPgUUID(inv.LegalEntityID),
		InvoiceNumber:       inv.InvoiceNumber,
		InvoiceType:         string(inv.InvoiceType),
		OrderID:             toPgUUIDPtr(inv.OrderID),
		DeliveryID:          toPgUUIDPtr(inv.DeliveryID),
		CustomerID:          toPgUUID(inv.CustomerID),
		CustomerNameAr:      inv.CustomerNameAr,
		CustomerNameEn:      inv.CustomerNameEn,
		CustomerTaxID:       toPgTextPtr(inv.CustomerTaxID),
		CustomerCrNumber:    toPgTextPtr(inv.CustomerCRNumber),
		CustomerAddressJson: addrJson,
		IssueDate:           toPgDate(inv.IssueDate),
		DueDate:             toPgDate(inv.DueDate),
		PaymentTermsID:      toPgUUIDPtr(inv.PaymentTermsID),
		CurrencyCode:        inv.CurrencyCode,
		ExchangeRate:        toPgNumeric(inv.ExchangeRate),
		Subtotal:            toPgNumeric(inv.Subtotal),
		DiscountAmount:      toPgNumeric(inv.DiscountAmount),
		TaxableAmount:       toPgNumeric(inv.TaxableAmount),
		TaxAmount:           toPgNumeric(inv.TaxAmount),
		TotalAmount:         toPgNumeric(inv.TotalAmount),
		Status:              string(inv.Status),
		Notes:               toPgTextPtr(inv.Notes),
		CreatedByUserID:     toPgUUIDPtr(inv.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	result := &domain.SalesInvoice{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		InvoiceNumber:    row.InvoiceNumber,
		InvoiceType:      domain.InvoiceType(row.InvoiceType),
		OrderID:          toUUIDPtr(row.OrderID),
		DeliveryID:       toUUIDPtr(row.DeliveryID),
		CustomerID:       toUUID(row.CustomerID),
		CustomerNameAr:   row.CustomerNameAr,
		CustomerNameEn:   row.CustomerNameEn,
		CustomerTaxID:    fromPgTextPtr(row.CustomerTaxID),
		CustomerCRNumber: fromPgTextPtr(row.CustomerCrNumber),
		IssueDate:        fromPgDate(row.IssueDate),
		DueDate:          fromPgDate(row.DueDate),
		PaymentTermsID:   toUUIDPtr(row.PaymentTermsID),
		CurrencyCode:     row.CurrencyCode,
		ExchangeRate:     fromPgNumeric(row.ExchangeRate),
		Subtotal:         fromPgNumeric(row.Subtotal),
		DiscountAmount:   fromPgNumeric(row.DiscountAmount),
		TaxableAmount:    fromPgNumeric(row.TaxableAmount),
		TaxAmount:        fromPgNumeric(row.TaxAmount),
		TotalAmount:      fromPgNumeric(row.TotalAmount),
		Status:           domain.InvoiceStatus(row.Status),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		OpenItemID:       toUUIDPtr(row.OpenItemID),
		QRCodeData:       fromPgTextPtr(row.QrCodeData),
		Notes:            fromPgTextPtr(row.Notes),
		CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
		PostedAt:         fromPgTimestamptz(row.PostedAt),
		CreatedAt:        fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:        fromPgTimestamptzVal(row.UpdatedAt),
	}

	for _, l := range inv.Lines {
		lRow, err := queries.CreateInvoiceLine(ctx, sqlc.CreateInvoiceLineParams{
			ID:              toPgUUID(l.ID),
			InvoiceID:       row.ID,
			LineNumber:      int32(l.LineNumber),
			ItemID:          toPgUUID(l.ItemID),
			Description:     l.Description,
			Uom:             l.UOM,
			Quantity:        toPgNumeric(l.Quantity),
			UnitPrice:       toPgNumeric(l.UnitPrice),
			GrossAmount:     toPgNumeric(l.GrossAmount),
			DiscountAmount:  toPgNumeric(l.DiscountAmount),
			TaxableAmount:   toPgNumeric(l.TaxableAmount),
			TaxCode:         l.TaxCode,
			TaxRatePercent:  toPgNumeric(l.TaxRatePercent),
			TaxAmount:       toPgNumeric(l.TaxAmount),
			NetAmount:       toPgNumeric(l.NetAmount),
			ExemptionReason: toPgTextPtr(l.ExemptionReason),
		})
		if err != nil {
			return nil, err
		}

		result.Lines = append(result.Lines, domain.SalesInvoiceLine{
			ID:              toUUID(lRow.ID),
			InvoiceID:       toUUID(lRow.InvoiceID),
			LineNumber:      int(lRow.LineNumber),
			ItemID:          toUUID(lRow.ItemID),
			Description:     lRow.Description,
			UOM:             lRow.Uom,
			Quantity:        fromPgNumeric(lRow.Quantity),
			UnitPrice:       fromPgNumeric(lRow.UnitPrice),
			GrossAmount:     fromPgNumeric(lRow.GrossAmount),
			DiscountAmount:  fromPgNumeric(lRow.DiscountAmount),
			TaxableAmount:   fromPgNumeric(lRow.TaxableAmount),
			TaxCode:         lRow.TaxCode,
			TaxRatePercent:  fromPgNumeric(lRow.TaxRatePercent),
			TaxAmount:       fromPgNumeric(lRow.TaxAmount),
			NetAmount:       fromPgNumeric(lRow.NetAmount),
			ExemptionReason: fromPgTextPtr(lRow.ExemptionReason),
		})
	}

	return result, nil
}

func (r *salesRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.SalesInvoice, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetInvoice(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvoiceNotFound
		}
		return nil, err
	}

	inv := &domain.SalesInvoice{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		InvoiceNumber:    row.InvoiceNumber,
		InvoiceType:      domain.InvoiceType(row.InvoiceType),
		OrderID:          toUUIDPtr(row.OrderID),
		DeliveryID:       toUUIDPtr(row.DeliveryID),
		CustomerID:       toUUID(row.CustomerID),
		CustomerNameAr:   row.CustomerNameAr,
		CustomerNameEn:   row.CustomerNameEn,
		CustomerTaxID:    fromPgTextPtr(row.CustomerTaxID),
		CustomerCRNumber: fromPgTextPtr(row.CustomerCrNumber),
		IssueDate:        fromPgDate(row.IssueDate),
		DueDate:          fromPgDate(row.DueDate),
		PaymentTermsID:   toUUIDPtr(row.PaymentTermsID),
		CurrencyCode:     row.CurrencyCode,
		ExchangeRate:     fromPgNumeric(row.ExchangeRate),
		Subtotal:         fromPgNumeric(row.Subtotal),
		DiscountAmount:   fromPgNumeric(row.DiscountAmount),
		TaxableAmount:    fromPgNumeric(row.TaxableAmount),
		TaxAmount:        fromPgNumeric(row.TaxAmount),
		TotalAmount:      fromPgNumeric(row.TotalAmount),
		Status:           domain.InvoiceStatus(row.Status),
		JournalEntryID:   toUUIDPtr(row.JournalEntryID),
		OpenItemID:       toUUIDPtr(row.OpenItemID),
		QRCodeData:       fromPgTextPtr(row.QrCodeData),
		Notes:            fromPgTextPtr(row.Notes),
		CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
		PostedAt:         fromPgTimestamptz(row.PostedAt),
		CreatedAt:        fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:        fromPgTimestamptzVal(row.UpdatedAt),
	}

	lines, err := queries.ListInvoiceLines(ctx, row.ID)
	if err == nil {
		for _, l := range lines {
			inv.Lines = append(inv.Lines, domain.SalesInvoiceLine{
				ID:              toUUID(l.ID),
				InvoiceID:       toUUID(l.InvoiceID),
				LineNumber:      int(l.LineNumber),
				ItemID:          toUUID(l.ItemID),
				Description:     l.Description,
				UOM:             l.Uom,
				Quantity:        fromPgNumeric(l.Quantity),
				UnitPrice:       fromPgNumeric(l.UnitPrice),
				GrossAmount:     fromPgNumeric(l.GrossAmount),
				DiscountAmount:  fromPgNumeric(l.DiscountAmount),
				TaxableAmount:   fromPgNumeric(l.TaxableAmount),
				TaxCode:         l.TaxCode,
				TaxRatePercent:  fromPgNumeric(l.TaxRatePercent),
				TaxAmount:       fromPgNumeric(l.TaxAmount),
				NetAmount:       fromPgNumeric(l.NetAmount),
				ExemptionReason: fromPgTextPtr(l.ExemptionReason),
			})
		}
	}

	return inv, nil
}

func (r *salesRepository) ListInvoices(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, invoiceType *domain.InvoiceType, status *domain.InvoiceStatus, limit, offset int) ([]*domain.SalesInvoice, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var typeText, statusText pgtype.Text
	if invoiceType != nil {
		typeText = toPgText(string(*invoiceType))
	}
	if status != nil {
		statusText = toPgText(string(*status))
	}

	rows, err := queries.ListInvoices(ctx, sqlc.ListInvoicesParams{
		LegalEntityID: toPgUUID(legalEntityID),
		CustomerID:    toPgUUIDPtr(customerID),
		InvoiceType:   typeText,
		Status:        statusText,
		LimitCount:    int32(limit),
		OffsetCount:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []*domain.SalesInvoice
	for _, row := range rows {
		res = append(res, &domain.SalesInvoice{
			ID:               toUUID(row.ID),
			LegalEntityID:    toUUID(row.LegalEntityID),
			InvoiceNumber:    row.InvoiceNumber,
			InvoiceType:      domain.InvoiceType(row.InvoiceType),
			OrderID:          toUUIDPtr(row.OrderID),
			DeliveryID:       toUUIDPtr(row.DeliveryID),
			CustomerID:       toUUID(row.CustomerID),
			CustomerNameAr:   row.CustomerNameAr,
			CustomerNameEn:   row.CustomerNameEn,
			CustomerTaxID:    fromPgTextPtr(row.CustomerTaxID),
			CustomerCRNumber: fromPgTextPtr(row.CustomerCrNumber),
			IssueDate:        fromPgDate(row.IssueDate),
			DueDate:          fromPgDate(row.DueDate),
			PaymentTermsID:   toUUIDPtr(row.PaymentTermsID),
			CurrencyCode:     row.CurrencyCode,
			ExchangeRate:     fromPgNumeric(row.ExchangeRate),
			Subtotal:         fromPgNumeric(row.Subtotal),
			DiscountAmount:   fromPgNumeric(row.DiscountAmount),
			TaxableAmount:    fromPgNumeric(row.TaxableAmount),
			TaxAmount:        fromPgNumeric(row.TaxAmount),
			TotalAmount:      fromPgNumeric(row.TotalAmount),
			Status:           domain.InvoiceStatus(row.Status),
			JournalEntryID:   toUUIDPtr(row.JournalEntryID),
			OpenItemID:       toUUIDPtr(row.OpenItemID),
			QRCodeData:       fromPgTextPtr(row.QrCodeData),
			Notes:            fromPgTextPtr(row.Notes),
			CreatedByUserID:  toUUIDPtr(row.CreatedByUserID),
			PostedAt:         fromPgTimestamptz(row.PostedAt),
			CreatedAt:        fromPgTimestamptzVal(row.CreatedAt),
			UpdatedAt:        fromPgTimestamptzVal(row.UpdatedAt),
		})
	}
	return res, nil
}

func (r *salesRepository) PostInvoice(ctx context.Context, id uuid.UUID, journalEntryID, openItemID *uuid.UUID, qrCodeData *string) (*domain.SalesInvoice, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.PostInvoice(ctx, sqlc.PostInvoiceParams{
		ID:             toPgUUID(id),
		JournalEntryID: toPgUUIDPtr(journalEntryID),
		OpenItemID:     toPgUUIDPtr(openItemID),
		QrCodeData:     toPgTextPtr(qrCodeData),
	})
	if err != nil {
		return nil, err
	}

	return r.GetInvoice(ctx, toUUID(row.ID))
}

func (r *salesRepository) UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) (*domain.SalesInvoice, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.UpdateInvoiceStatus(ctx, sqlc.UpdateInvoiceStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}

	return r.GetInvoice(ctx, toUUID(row.ID))
}

// ----------------------------------------------------------------------------
// Credit Notes
// ----------------------------------------------------------------------------

func (r *salesRepository) CreateCreditNote(ctx context.Context, cn *domain.CreditNote) (*domain.CreditNote, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.CreateCreditNote(ctx, sqlc.CreateCreditNoteParams{
		ID:                toPgUUID(cn.ID),
		LegalEntityID:     toPgUUID(cn.LegalEntityID),
		CreditNoteNumber:  cn.CreditNoteNumber,
		OriginalInvoiceID: toPgUUID(cn.OriginalInvoiceID),
		CustomerID:        toPgUUID(cn.CustomerID),
		IssueDate:         toPgDate(cn.IssueDate),
		ReasonCode:        string(cn.ReasonCode),
		ReasonDetails:     toPgTextPtr(cn.ReasonDetails),
		CurrencyCode:      cn.CurrencyCode,
		ExchangeRate:      toPgNumeric(cn.ExchangeRate),
		Subtotal:          toPgNumeric(cn.Subtotal),
		TaxAmount:         toPgNumeric(cn.TaxAmount),
		TotalAmount:       toPgNumeric(cn.TotalAmount),
		Status:            string(cn.Status),
		CreatedByUserID:   toPgUUIDPtr(cn.CreatedByUserID),
	})
	if err != nil {
		return nil, err
	}

	result := &domain.CreditNote{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		CreditNoteNumber:  row.CreditNoteNumber,
		OriginalInvoiceID: toUUID(row.OriginalInvoiceID),
		CustomerID:        toUUID(row.CustomerID),
		IssueDate:         fromPgDate(row.IssueDate),
		ReasonCode:        domain.CreditNoteReason(row.ReasonCode),
		ReasonDetails:     fromPgTextPtr(row.ReasonDetails),
		CurrencyCode:      row.CurrencyCode,
		ExchangeRate:      fromPgNumeric(row.ExchangeRate),
		Subtotal:          fromPgNumeric(row.Subtotal),
		TaxAmount:         fromPgNumeric(row.TaxAmount),
		TotalAmount:       fromPgNumeric(row.TotalAmount),
		Status:            domain.CreditNoteStatus(row.Status),
		JournalEntryID:    toUUIDPtr(row.JournalEntryID),
		OpenItemID:        toUUIDPtr(row.OpenItemID),
		CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
		PostedAt:          fromPgTimestamptz(row.PostedAt),
		CreatedAt:         fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptzVal(row.UpdatedAt),
	}

	for _, l := range cn.Lines {
		lRow, err := queries.CreateCreditNoteLine(ctx, sqlc.CreateCreditNoteLineParams{
			ID:             toPgUUID(l.ID),
			CreditNoteID:   row.ID,
			LineNumber:     int32(l.LineNumber),
			ItemID:         toPgUUID(l.ItemID),
			Description:    l.Description,
			Quantity:       toPgNumeric(l.Quantity),
			UnitPrice:      toPgNumeric(l.UnitPrice),
			TaxableAmount:  toPgNumeric(l.TaxableAmount),
			TaxCode:        l.TaxCode,
			TaxRatePercent: toPgNumeric(l.TaxRatePercent),
			TaxAmount:      toPgNumeric(l.TaxAmount),
			TotalAmount:    toPgNumeric(l.TotalAmount),
		})
		if err != nil {
			return nil, err
		}

		result.Lines = append(result.Lines, domain.CreditNoteLine{
			ID:             toUUID(lRow.ID),
			CreditNoteID:   toUUID(lRow.CreditNoteID),
			LineNumber:     int(lRow.LineNumber),
			ItemID:         toUUID(lRow.ItemID),
			Description:    lRow.Description,
			Quantity:       fromPgNumeric(lRow.Quantity),
			UnitPrice:      fromPgNumeric(lRow.UnitPrice),
			TaxableAmount:  fromPgNumeric(lRow.TaxableAmount),
			TaxCode:        lRow.TaxCode,
			TaxRatePercent: fromPgNumeric(lRow.TaxRatePercent),
			TaxAmount:      fromPgNumeric(lRow.TaxAmount),
			TotalAmount:    fromPgNumeric(lRow.TotalAmount),
		})
	}

	return result, nil
}

func (r *salesRepository) GetCreditNote(ctx context.Context, id uuid.UUID) (*domain.CreditNote, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.GetCreditNote(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCreditNoteNotFound
		}
		return nil, err
	}

	cn := &domain.CreditNote{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		CreditNoteNumber:  row.CreditNoteNumber,
		OriginalInvoiceID: toUUID(row.OriginalInvoiceID),
		CustomerID:        toUUID(row.CustomerID),
		IssueDate:         fromPgDate(row.IssueDate),
		ReasonCode:        domain.CreditNoteReason(row.ReasonCode),
		ReasonDetails:     fromPgTextPtr(row.ReasonDetails),
		CurrencyCode:      row.CurrencyCode,
		ExchangeRate:      fromPgNumeric(row.ExchangeRate),
		Subtotal:          fromPgNumeric(row.Subtotal),
		TaxAmount:         fromPgNumeric(row.TaxAmount),
		TotalAmount:       fromPgNumeric(row.TotalAmount),
		Status:            domain.CreditNoteStatus(row.Status),
		JournalEntryID:    toUUIDPtr(row.JournalEntryID),
		OpenItemID:        toUUIDPtr(row.OpenItemID),
		CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
		PostedAt:          fromPgTimestamptz(row.PostedAt),
		CreatedAt:         fromPgTimestamptzVal(row.CreatedAt),
		UpdatedAt:         fromPgTimestamptzVal(row.UpdatedAt),
	}

	lines, err := queries.ListCreditNoteLines(ctx, row.ID)
	if err == nil {
		for _, l := range lines {
			cn.Lines = append(cn.Lines, domain.CreditNoteLine{
				ID:             toUUID(l.ID),
				CreditNoteID:   toUUID(l.CreditNoteID),
				LineNumber:     int(l.LineNumber),
				ItemID:         toUUID(l.ItemID),
				Description:    l.Description,
				Quantity:       fromPgNumeric(l.Quantity),
				UnitPrice:      fromPgNumeric(l.UnitPrice),
				TaxableAmount:  fromPgNumeric(l.TaxableAmount),
				TaxCode:        l.TaxCode,
				TaxRatePercent: fromPgNumeric(l.TaxRatePercent),
				TaxAmount:      fromPgNumeric(l.TaxAmount),
				TotalAmount:    fromPgNumeric(l.TotalAmount),
			})
		}
	}

	return cn, nil
}

func (r *salesRepository) ListCreditNotes(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, originalInvoiceID *uuid.UUID, status *domain.CreditNoteStatus, limit, offset int) ([]*domain.CreditNote, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	var statusText pgtype.Text
	if status != nil {
		statusText = toPgText(string(*status))
	}

	rows, err := queries.ListCreditNotes(ctx, sqlc.ListCreditNotesParams{
		LegalEntityID:     toPgUUID(legalEntityID),
		CustomerID:        toPgUUIDPtr(customerID),
		OriginalInvoiceID: toPgUUIDPtr(originalInvoiceID),
		Status:            statusText,
		LimitCount:        int32(limit),
		OffsetCount:       int32(offset),
	})
	if err != nil {
		return nil, err
	}

	var res []*domain.CreditNote
	for _, row := range rows {
		res = append(res, &domain.CreditNote{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			CreditNoteNumber:  row.CreditNoteNumber,
			OriginalInvoiceID: toUUID(row.OriginalInvoiceID),
			CustomerID:        toUUID(row.CustomerID),
			IssueDate:         fromPgDate(row.IssueDate),
			ReasonCode:        domain.CreditNoteReason(row.ReasonCode),
			ReasonDetails:     fromPgTextPtr(row.ReasonDetails),
			CurrencyCode:      row.CurrencyCode,
			ExchangeRate:      fromPgNumeric(row.ExchangeRate),
			Subtotal:          fromPgNumeric(row.Subtotal),
			TaxAmount:         fromPgNumeric(row.TaxAmount),
			TotalAmount:       fromPgNumeric(row.TotalAmount),
			Status:            domain.CreditNoteStatus(row.Status),
			JournalEntryID:    toUUIDPtr(row.JournalEntryID),
			OpenItemID:        toUUIDPtr(row.OpenItemID),
			CreatedByUserID:   toUUIDPtr(row.CreatedByUserID),
			PostedAt:          fromPgTimestamptz(row.PostedAt),
			CreatedAt:         fromPgTimestamptzVal(row.CreatedAt),
			UpdatedAt:         fromPgTimestamptzVal(row.UpdatedAt),
		})
	}
	return res, nil
}

func (r *salesRepository) PostCreditNote(ctx context.Context, id uuid.UUID, journalEntryID, openItemID *uuid.UUID) (*domain.CreditNote, error) {
	queries := sqlc.New(r.db.DBTX(ctx))

	row, err := queries.PostCreditNote(ctx, sqlc.PostCreditNoteParams{
		ID:             toPgUUID(id),
		JournalEntryID: toPgUUIDPtr(journalEntryID),
		OpenItemID:     toPgUUIDPtr(openItemID),
	})
	if err != nil {
		return nil, err
	}

	return r.GetCreditNote(ctx, toUUID(row.ID))
}
