package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"
	"erp/internal/sales/domain"
	"erp/internal/sales/ports"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type salesService struct {
	repo ports.Repository
	db   *database.DB
	tx   transaction.Manager
}

func NewSalesService(repo ports.Repository, db *database.DB) ports.SalesService {
	return &salesService{
		repo: repo,
		db:   db,
		tx:   transaction.NewManager(db.Pool),
	}
}

// ----------------------------------------------------------------------------
// Quotations
// ----------------------------------------------------------------------------

func (s *salesService) CreateQuotation(ctx context.Context, cmd ports.CreateQuotationCommand) (*domain.Quotation, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	quoteNumber := identifiers.NewDocumentNumber("QT", time.Now())
	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}
	rate := cmd.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	quotation := &domain.Quotation{
		ID:              uuid.New(),
		LegalEntityID:   cmd.LegalEntityID,
		QuotationNumber: quoteNumber,
		CustomerID:      cmd.CustomerID,
		CustomerNameEn:  cmd.CustomerNameEn,
		CustomerNameAr:  cmd.CustomerNameAr,
		QuoteDate:       cmd.QuoteDate,
		ValidUntil:      cmd.ValidUntil,
		CurrencyCode:    currency,
		ExchangeRate:    rate,
		Status:          domain.QuotationStatusDraft,
		Notes:           cmd.Notes,
		CreatedByUserID: cmd.CreatedByUserID,
		Subtotal:        decimal.Zero,
		DiscountAmount:  decimal.Zero,
		TaxableAmount:   decimal.Zero,
		TaxAmount:       decimal.Zero,
		TotalAmount:     decimal.Zero,
	}

	for i, l := range cmd.Lines {
		gross, disc, taxable, tax, total := domain.CalculateLineAmounts(l.Quantity, l.UnitPrice, l.DiscountPercent, l.TaxRatePercent)
		quotation.Subtotal = quotation.Subtotal.Add(gross)
		quotation.DiscountAmount = quotation.DiscountAmount.Add(disc)
		quotation.TaxableAmount = quotation.TaxableAmount.Add(taxable)
		quotation.TaxAmount = quotation.TaxAmount.Add(tax)
		quotation.TotalAmount = quotation.TotalAmount.Add(total)

		quotation.Lines = append(quotation.Lines, domain.QuotationLine{
			ID:              uuid.New(),
			LineNumber:      i + 1,
			ItemID:          l.ItemID,
			DescriptionEn:   l.DescriptionEn,
			DescriptionAr:   l.DescriptionAr,
			UOM:             l.UOM,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			DiscountPercent: l.DiscountPercent,
			DiscountAmount:  disc,
			TaxableAmount:   taxable,
			TaxCode:         l.TaxCode,
			TaxRatePercent:  l.TaxRatePercent,
			TaxAmount:       tax,
			TotalAmount:     total,
		})
	}

	return s.repo.CreateQuotation(ctx, quotation)
}

func (s *salesService) GetQuotation(ctx context.Context, id uuid.UUID) (*domain.Quotation, error) {
	return s.repo.GetQuotation(ctx, id)
}

func (s *salesService) ListQuotations(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.QuotationStatus, limit, offset int) ([]*domain.Quotation, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListQuotations(ctx, legalEntityID, customerID, status, limit, offset)
}

func (s *salesService) ConvertQuotationToOrder(ctx context.Context, quoteID uuid.UUID, userID *uuid.UUID) (*domain.SalesOrder, error) {
	quote, err := s.repo.GetQuotation(ctx, quoteID)
	if err != nil {
		return nil, err
	}
	if quote.Status != domain.QuotationStatusAccepted && quote.Status != domain.QuotationStatusSent && quote.Status != domain.QuotationStatusDraft {
		return nil, domain.ErrQuotationNotAccepted
	}

	var orderLines []ports.CreateOrderLineCommand
	for _, ql := range quote.Lines {
		orderLines = append(orderLines, ports.CreateOrderLineCommand{
			ItemID:          ql.ItemID,
			DescriptionEn:   ql.DescriptionEn,
			DescriptionAr:   ql.DescriptionAr,
			UOM:             ql.UOM,
			Quantity:        ql.Quantity,
			UnitPrice:       ql.UnitPrice,
			DiscountPercent: ql.DiscountPercent,
			TaxCode:         ql.TaxCode,
			TaxRatePercent:  ql.TaxRatePercent,
		})
	}

	orderCmd := ports.CreateOrderCommand{
		LegalEntityID:   quote.LegalEntityID,
		CustomerID:      quote.CustomerID,
		QuotationID:     &quote.ID,
		OrderDate:       time.Now(),
		CurrencyCode:    quote.CurrencyCode,
		ExchangeRate:    quote.ExchangeRate,
		Notes:           quote.Notes,
		CreatedByUserID: userID,
		Lines:           orderLines,
	}

	order, err := s.CreateOrder(ctx, orderCmd)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateQuotationStatus(ctx, quote.ID, domain.QuotationStatusConverted, &order.ID)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// ----------------------------------------------------------------------------
// Orders
// ----------------------------------------------------------------------------

func (s *salesService) CreateOrder(ctx context.Context, cmd ports.CreateOrderCommand) (*domain.SalesOrder, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	orderNumber := identifiers.NewDocumentNumber("SO", time.Now())
	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}
	rate := cmd.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	order := &domain.SalesOrder{
		ID:                   uuid.New(),
		LegalEntityID:        cmd.LegalEntityID,
		OrderNumber:          orderNumber,
		CustomerID:           cmd.CustomerID,
		QuotationID:          cmd.QuotationID,
		CustomerPONumber:     cmd.CustomerPONumber,
		OrderDate:            cmd.OrderDate,
		ExpectedDeliveryDate: cmd.ExpectedDeliveryDate,
		PaymentTermsID:       cmd.PaymentTermsID,
		CurrencyCode:         currency,
		ExchangeRate:         rate,
		Status:               domain.OrderStatusDraft,
		Notes:                cmd.Notes,
		CreatedByUserID:      cmd.CreatedByUserID,
		Subtotal:             decimal.Zero,
		DiscountAmount:       decimal.Zero,
		TaxableAmount:        decimal.Zero,
		TaxAmount:            decimal.Zero,
		TotalAmount:          decimal.Zero,
	}

	for i, l := range cmd.Lines {
		gross, disc, taxable, tax, total := domain.CalculateLineAmounts(l.Quantity, l.UnitPrice, l.DiscountPercent, l.TaxRatePercent)
		order.Subtotal = order.Subtotal.Add(gross)
		order.DiscountAmount = order.DiscountAmount.Add(disc)
		order.TaxableAmount = order.TaxableAmount.Add(taxable)
		order.TaxAmount = order.TaxAmount.Add(tax)
		order.TotalAmount = order.TotalAmount.Add(total)

		order.Lines = append(order.Lines, domain.SalesOrderLine{
			ID:                uuid.New(),
			LineNumber:        i + 1,
			ItemID:            l.ItemID,
			DescriptionEn:     l.DescriptionEn,
			DescriptionAr:     l.DescriptionAr,
			UOM:               l.UOM,
			Quantity:          l.Quantity,
			DeliveredQuantity: decimal.Zero,
			InvoicedQuantity:  decimal.Zero,
			UnitPrice:         l.UnitPrice,
			DiscountPercent:   l.DiscountPercent,
			DiscountAmount:    disc,
			TaxableAmount:     taxable,
			TaxCode:           l.TaxCode,
			TaxRatePercent:    l.TaxRatePercent,
			TaxAmount:         tax,
			TotalAmount:       total,
		})
	}

	return s.repo.CreateOrder(ctx, order)
}

func (s *salesService) ConfirmOrder(ctx context.Context, orderID uuid.UUID) (*domain.SalesOrder, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.OrderStatusDraft {
		return nil, domain.ErrInvalidOrderStatus
	}

	// 1. Credit Limit Check
	var isCreditHold bool
	var creditLimit, currentBalance decimal.Decimal
	var blockAction string
	err = s.db.Pool.QueryRow(ctx, `
		SELECT is_credit_hold, credit_limit, current_balance, block_action
		FROM receivables.customer_credit_limits
		WHERE customer_id = $1
	`, order.CustomerID).Scan(&isCreditHold, &creditLimit, &currentBalance, &blockAction)
	if err == nil {
		if isCreditHold {
			return nil, domain.ErrCustomerCreditHold
		}
		if creditLimit.GreaterThan(decimal.Zero) && blockAction == "BLOCK" {
			newPotentialBalance := currentBalance.Add(order.TotalAmount)
			if newPotentialBalance.GreaterThan(creditLimit) {
				return nil, domain.ErrCustomerCreditLimitExceeded
			}
		}
	}

	return s.repo.UpdateOrderStatus(ctx, order.ID, domain.OrderStatusConfirmed)
}

func (s *salesService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.SalesOrder, error) {
	return s.repo.GetOrder(ctx, id)
}

func (s *salesService) ListOrders(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.OrderStatus, limit, offset int) ([]*domain.SalesOrder, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListOrders(ctx, legalEntityID, customerID, status, limit, offset)
}

// ----------------------------------------------------------------------------
// Deliveries
// ----------------------------------------------------------------------------

func (s *salesService) CreateDelivery(ctx context.Context, cmd ports.CreateDeliveryCommand) (*domain.Delivery, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	deliveryNumber := identifiers.NewDocumentNumber("DN", time.Now())

	delivery := &domain.Delivery{
		ID:                uuid.New(),
		LegalEntityID:     cmd.LegalEntityID,
		DeliveryNumber:    deliveryNumber,
		OrderID:           cmd.OrderID,
		CustomerID:        cmd.CustomerID,
		WarehouseID:       cmd.WarehouseID,
		DeliveryDate:      cmd.DeliveryDate,
		Status:            domain.DeliveryStatusDraft,
		ShippingReference: cmd.ShippingReference,
		Notes:             cmd.Notes,
		CreatedByUserID:   cmd.CreatedByUserID,
	}

	for i, l := range cmd.Lines {
		delivery.Lines = append(delivery.Lines, domain.DeliveryLine{
			ID:          uuid.New(),
			LineNumber:  i + 1,
			OrderLineID: l.OrderLineID,
			ItemID:      l.ItemID,
			UOM:         l.UOM,
			Quantity:    l.Quantity,
			WarehouseID: l.WarehouseID,
			BinID:       l.BinID,
		})
	}

	return s.repo.CreateDelivery(ctx, delivery)
}

func (s *salesService) DispatchDelivery(ctx context.Context, deliveryID uuid.UUID) (*domain.Delivery, error) {
	var delivery *domain.Delivery
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		delivery, err = s.dispatchDelivery(txCtx, deliveryID)
		return err
	})
	return delivery, err
}

func (s *salesService) dispatchDelivery(ctx context.Context, deliveryID uuid.UUID) (*domain.Delivery, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM sales.deliveries WHERE id = $1 FOR UPDATE", deliveryID).Scan(&deliveryID); err != nil {
		return nil, fmt.Errorf("failed to lock delivery: %w", err)
	}
	delivery, err := s.repo.GetDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}
	if delivery.Status != domain.DeliveryStatusDraft {
		return nil, errors.New("delivery is already dispatched or cancelled")
	}

	// For each line, post stock movement and update order delivered quantity
	for _, line := range delivery.Lines {
		// 1. Get default bin if bin is not specified
		binID := line.BinID
		if binID == nil {
			var defBin uuid.UUID
			err := dbtx.QueryRow(ctx, "SELECT id FROM warehouse.bin_locations WHERE warehouse_id = $1 LIMIT 1", line.WarehouseID).Scan(&defBin)
			if err == nil {
				binID = &defBin
			}
		}

		// 2. Get UOM ID
		var uomID uuid.UUID
		if err := dbtx.QueryRow(ctx, "SELECT id FROM products.units_of_measure LIMIT 1").Scan(&uomID); err != nil {
			return nil, fmt.Errorf("failed to resolve inventory unit of measure: %w", err)
		}

		// 3. Insert into inventory.stock_movements (Negative quantity for SALES_ISSUE)
		movementNumber := identifiers.NewDocumentNumber("SM-OUT", time.Now())
		negQty := line.Quantity.Neg()
		creatorID := delivery.CreatedByUserID
		if creatorID == nil {
			var adminID uuid.UUID
			if err := dbtx.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID); err != nil {
				return nil, fmt.Errorf("failed to resolve delivery creator: %w", err)
			}
			creatorID = &adminID
		}

		if _, err := dbtx.Exec(ctx, `
			INSERT INTO inventory.stock_movements (
				legal_entity_id, movement_number, movement_type, item_id,
				warehouse_id, bin_id, quantity, uom_id, unit_cost, total_cost,
				source_doc_type, source_doc_id, created_by
			) VALUES ($1, $2, 'SALES_ISSUE', $3, $4, $5, $6, $7, $8, $9, 'DELIVERY', $10, $11)
		`, delivery.LegalEntityID, movementNumber, line.ItemID, line.WarehouseID, binID, negQty, uomID, line.UnitCost, line.TotalCost, delivery.ID, creatorID); err != nil {
			return nil, fmt.Errorf("failed to create sales stock movement: %w", err)
		}

		// 4. Update order line delivered quantity if linked
		if line.OrderLineID != nil {
			if err := s.repo.UpdateOrderLineDeliveredQty(ctx, *line.OrderLineID, line.Quantity); err != nil {
				return nil, fmt.Errorf("failed to update delivered quantity: %w", err)
			}
		}
	}

	return s.repo.UpdateDeliveryStatus(ctx, delivery.ID, domain.DeliveryStatusDelivered)
}

func (s *salesService) GetDelivery(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	return s.repo.GetDelivery(ctx, id)
}

func (s *salesService) ListDeliveries(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, orderID *uuid.UUID, status *domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListDeliveries(ctx, legalEntityID, customerID, orderID, status, limit, offset)
}

// ----------------------------------------------------------------------------
// Invoices
// ----------------------------------------------------------------------------

func (s *salesService) CreateInvoice(ctx context.Context, cmd ports.CreateInvoiceCommand) (*domain.SalesInvoice, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	// Check ZATCA requirement: STANDARD invoice requires valid 15-digit VAT
	if cmd.InvoiceType == domain.InvoiceTypeStandard {
		if err := domain.ValidateStandardInvoice(cmd.CustomerTaxID); err != nil {
			return nil, err
		}
	}

	invNumber := identifiers.NewDocumentNumber("INV", time.Now())
	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}
	rate := cmd.ExchangeRate
	if rate.IsZero() {
		rate = decimal.NewFromInt(1)
	}

	invoice := &domain.SalesInvoice{
		ID:                  uuid.New(),
		LegalEntityID:       cmd.LegalEntityID,
		InvoiceNumber:       invNumber,
		InvoiceType:         cmd.InvoiceType,
		OrderID:             cmd.OrderID,
		DeliveryID:          cmd.DeliveryID,
		CustomerID:          cmd.CustomerID,
		CustomerNameAr:      cmd.CustomerNameAr,
		CustomerNameEn:      cmd.CustomerNameEn,
		CustomerTaxID:       cmd.CustomerTaxID,
		CustomerCRNumber:    cmd.CustomerCRNumber,
		CustomerAddressJSON: cmd.CustomerAddressJSON,
		IssueDate:           cmd.IssueDate,
		DueDate:             cmd.DueDate,
		PaymentTermsID:      cmd.PaymentTermsID,
		CurrencyCode:        currency,
		ExchangeRate:        rate,
		Status:              domain.InvoiceStatusDraft,
		Notes:               cmd.Notes,
		CreatedByUserID:     cmd.CreatedByUserID,
		Subtotal:            decimal.Zero,
		DiscountAmount:      decimal.Zero,
		TaxableAmount:       decimal.Zero,
		TaxAmount:           decimal.Zero,
		TotalAmount:         decimal.Zero,
	}

	for i, l := range cmd.Lines {
		gross, disc, taxable, tax, total := domain.CalculateLineAmounts(l.Quantity, l.UnitPrice, l.DiscountPercent, l.TaxRatePercent)
		invoice.Subtotal = invoice.Subtotal.Add(gross)
		invoice.DiscountAmount = invoice.DiscountAmount.Add(disc)
		invoice.TaxableAmount = invoice.TaxableAmount.Add(taxable)
		invoice.TaxAmount = invoice.TaxAmount.Add(tax)
		invoice.TotalAmount = invoice.TotalAmount.Add(total)

		invoice.Lines = append(invoice.Lines, domain.SalesInvoiceLine{
			ID:              uuid.New(),
			LineNumber:      i + 1,
			ItemID:          l.ItemID,
			Description:     l.Description,
			UOM:             l.UOM,
			Quantity:        l.Quantity,
			UnitPrice:       l.UnitPrice,
			GrossAmount:     gross,
			DiscountAmount:  disc,
			TaxableAmount:   taxable,
			TaxCode:         l.TaxCode,
			TaxRatePercent:  l.TaxRatePercent,
			TaxAmount:       tax,
			NetAmount:       total,
			ExemptionReason: l.ExemptionReason,
		})
	}

	return s.repo.CreateInvoice(ctx, invoice)
}

func (s *salesService) CreateInvoiceFromOrder(ctx context.Context, orderID uuid.UUID, invoiceType domain.InvoiceType, userID *uuid.UUID) (*domain.SalesInvoice, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Fetch customer details
	var nameAr, nameEn string
	var taxID, crNum *string
	err = s.db.Pool.QueryRow(ctx, `
		SELECT name_ar, name_en, tax_id, cr_number
		FROM receivables.customers
		WHERE id = $1
	`, order.CustomerID).Scan(&nameAr, &nameEn, &taxID, &crNum)
	if err != nil {
		return nil, err
	}

	var invoiceLines []ports.CreateInvoiceLineCommand
	for _, ol := range order.Lines {
		invoiceLines = append(invoiceLines, ports.CreateInvoiceLineCommand{
			ItemID:          ol.ItemID,
			Description:     ol.DescriptionAr,
			UOM:             ol.UOM,
			Quantity:        ol.Quantity,
			UnitPrice:       ol.UnitPrice,
			DiscountPercent: ol.DiscountPercent,
			TaxCode:         ol.TaxCode,
			TaxRatePercent:  ol.TaxRatePercent,
		})
	}

	dueDate := order.OrderDate.AddDate(0, 0, 30)

	return s.CreateInvoice(ctx, ports.CreateInvoiceCommand{
		LegalEntityID:    order.LegalEntityID,
		InvoiceType:      invoiceType,
		OrderID:          &order.ID,
		CustomerID:       order.CustomerID,
		CustomerNameAr:   nameAr,
		CustomerNameEn:   nameEn,
		CustomerTaxID:    taxID,
		CustomerCRNumber: crNum,
		IssueDate:        order.OrderDate,
		DueDate:          dueDate,
		PaymentTermsID:   order.PaymentTermsID,
		CurrencyCode:     order.CurrencyCode,
		ExchangeRate:     order.ExchangeRate,
		Notes:            order.Notes,
		CreatedByUserID:  userID,
		Lines:            invoiceLines,
	})
}

func (s *salesService) PostInvoice(ctx context.Context, invoiceID uuid.UUID) (*domain.SalesInvoice, error) {
	var invoice *domain.SalesInvoice
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		invoice, err = s.postInvoice(txCtx, invoiceID)
		return err
	})
	return invoice, err
}

func (s *salesService) postInvoice(ctx context.Context, invoiceID uuid.UUID) (*domain.SalesInvoice, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM sales.invoices WHERE id = $1 FOR UPDATE", invoiceID).Scan(&invoiceID); err != nil {
		return nil, fmt.Errorf("failed to lock sales invoice: %w", err)
	}
	inv, err := s.repo.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.Status == domain.InvoiceStatusPosted {
		return nil, domain.ErrInvoiceAlreadyPosted
	}

	// 1. Create AR Open Item in receivables.open_items
	openItemID := uuid.New()
	_, err = dbtx.Exec(ctx, `
		INSERT INTO receivables.open_items (
			id, legal_entity_id, customer_id, document_type, document_number,
			document_id, document_date, due_date, currency_code, original_amount,
			allocated_amount, remaining_amount, status
		) VALUES ($1, $2, $3, 'INVOICE', $4, $5, $6, $7, $8, $9, 0, $9, 'OPEN')
	`, openItemID, inv.LegalEntityID, inv.CustomerID, inv.InvoiceNumber, inv.ID, inv.IssueDate, inv.DueDate, inv.CurrencyCode, inv.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to create AR open item: %w", err)
	}

	// 2. Update Customer Balance in receivables.customer_credit_limits
	if _, err = dbtx.Exec(ctx, `
		UPDATE receivables.customer_credit_limits
		SET current_balance = current_balance + $2, updated_at = now()
		WHERE customer_id = $1
	`, inv.CustomerID, inv.TotalAmount); err != nil {
		return nil, fmt.Errorf("failed to update customer balance: %w", err)
	}

	// 3. Create Double-Entry GL Journal Entry in accounting.journal_entries
	var journalEntryID uuid.UUID
	var jeIDPtr *uuid.UUID
	var journalID, periodID, arAccountID, revAccountID, vatAccountID uuid.UUID
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 ORDER BY (code = 'SALES') DESC LIMIT 1", inv.LegalEntityID).Scan(&journalID); err != nil {
		return nil, fmt.Errorf("sales journal is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", inv.LegalEntityID).Scan(&periodID); err != nil {
		return nil, fmt.Errorf("open fiscal period is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '112000' LIMIT 1", inv.LegalEntityID).Scan(&arAccountID); err != nil {
		return nil, fmt.Errorf("accounts receivable account 112000 is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '410000' LIMIT 1", inv.LegalEntityID).Scan(&revAccountID); err != nil {
		return nil, fmt.Errorf("sales revenue account 410000 is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '213000' LIMIT 1", inv.LegalEntityID).Scan(&vatAccountID); err != nil {
		return nil, fmt.Errorf("output VAT account 213000 is not configured: %w", err)
	}

	creatorID := inv.CreatedByUserID
	if creatorID == nil {
		var adminID uuid.UUID
		if err := dbtx.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID); err != nil {
			return nil, fmt.Errorf("failed to resolve invoice creator: %w", err)
		}
		creatorID = &adminID
	}

	journalEntryID = uuid.New()
	entryNumber := identifiers.NewDocumentNumber("JE-INV", time.Now())
	_, err = dbtx.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by, posted_by, posted_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'DRAFT', 'SALES', 'INVOICE', $7,
				$8, $9, $10, 1.0, $11, $11, $12, $12, now()
			)
	`, journalEntryID, inv.LegalEntityID, journalID, periodID, entryNumber, inv.IssueDate, inv.ID,
		inv.InvoiceNumber, "إثبات فاتورة مبيعات - "+inv.InvoiceNumber, inv.CurrencyCode, inv.TotalAmount, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales journal entry: %w", err)
	}
	jeIDPtr = &journalEntryID

	if _, err = dbtx.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 1, $3, $4, $5, $6, 0, $6, 0)
	`, journalEntryID, inv.LegalEntityID, arAccountID, "العملاء والمدينون - "+inv.CustomerNameAr, inv.CurrencyCode, inv.TotalAmount); err != nil {
		return nil, fmt.Errorf("failed to create accounts receivable journal line: %w", err)
	}

	if _, err = dbtx.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 2, $3, 'إيرادات المبيعات', $4, 0, $5, 0, $5)
	`, journalEntryID, inv.LegalEntityID, revAccountID, inv.CurrencyCode, inv.TaxableAmount); err != nil {
		return nil, fmt.Errorf("failed to create sales revenue journal line: %w", err)
	}

	if inv.TaxAmount.GreaterThan(decimal.Zero) {
		if _, err = dbtx.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, 3, $3, 'ضريبة القيمة المضافة المستحقة (15%)', $4, 0, $5, 0, $5)
		`, journalEntryID, inv.LegalEntityID, vatAccountID, inv.CurrencyCode, inv.TaxAmount); err != nil {
			return nil, fmt.Errorf("failed to create output VAT journal line: %w", err)
		}
	}
	if _, err = dbtx.Exec(ctx, `
		UPDATE accounting.journal_entries
		SET status = 'POSTED', posted_by = $1, posted_at = now()
		WHERE id = $2
	`, creatorID, journalEntryID); err != nil {
		return nil, fmt.Errorf("failed to post sales journal entry: %w", err)
	}

	// 4. Update Order Invoiced Quantities if linked
	if inv.OrderID != nil {
		for _, l := range inv.Lines {
			if _, err = dbtx.Exec(ctx, `
				UPDATE sales.order_lines
				SET invoiced_quantity = invoiced_quantity + $2
				WHERE order_id = $1 AND item_id = $3
			`, inv.OrderID, l.Quantity, l.ItemID); err != nil {
				return nil, fmt.Errorf("failed to update invoiced quantity: %w", err)
			}
		}
	}

	// A QR is attached only by the real ZATCA signing/clearance flow. Creating a
	// locally fabricated value would misrepresent the invoice's compliance state.
	return s.repo.PostInvoice(ctx, inv.ID, jeIDPtr, &openItemID, nil)
}

func (s *salesService) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.SalesInvoice, error) {
	return s.repo.GetInvoice(ctx, id)
}

func (s *salesService) ListInvoices(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, invoiceType *domain.InvoiceType, status *domain.InvoiceStatus, limit, offset int) ([]*domain.SalesInvoice, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListInvoices(ctx, legalEntityID, customerID, invoiceType, status, limit, offset)
}

// ----------------------------------------------------------------------------
// Credit Notes
// ----------------------------------------------------------------------------

func (s *salesService) CreateCreditNote(ctx context.Context, cmd ports.CreateCreditNoteCommand) (*domain.CreditNote, error) {
	if len(cmd.Lines) == 0 {
		return nil, domain.ErrEmptyLineItems
	}

	origInv, err := s.repo.GetInvoice(ctx, cmd.OriginalInvoiceID)
	if err != nil {
		return nil, err
	}
	if origInv.Status != domain.InvoiceStatusPosted {
		return nil, domain.ErrOriginalInvoiceNotPosted
	}

	cnNumber := identifiers.NewDocumentNumber("CN", time.Now())
	currency := cmd.CurrencyCode
	if currency == "" {
		currency = "SAR"
	}

	cn := &domain.CreditNote{
		ID:                uuid.New(),
		LegalEntityID:     cmd.LegalEntityID,
		CreditNoteNumber:  cnNumber,
		OriginalInvoiceID: cmd.OriginalInvoiceID,
		CustomerID:        cmd.CustomerID,
		IssueDate:         cmd.IssueDate,
		ReasonCode:        cmd.ReasonCode,
		ReasonDetails:     cmd.ReasonDetails,
		CurrencyCode:      currency,
		ExchangeRate:      decimal.NewFromInt(1),
		Status:            domain.CreditNoteStatusDraft,
		CreatedByUserID:   cmd.CreatedByUserID,
		Subtotal:          decimal.Zero,
		TaxAmount:         decimal.Zero,
		TotalAmount:       decimal.Zero,
	}

	for i, l := range cmd.Lines {
		taxable := l.Quantity.Mul(l.UnitPrice).Round(4)
		tax := decimal.Zero
		if l.TaxRatePercent.GreaterThan(decimal.Zero) {
			tax = taxable.Mul(l.TaxRatePercent).Div(decimal.NewFromInt(100)).Round(4)
		}
		total := taxable.Add(tax).Round(4)

		cn.Subtotal = cn.Subtotal.Add(taxable)
		cn.TaxAmount = cn.TaxAmount.Add(tax)
		cn.TotalAmount = cn.TotalAmount.Add(total)

		cn.Lines = append(cn.Lines, domain.CreditNoteLine{
			ID:             uuid.New(),
			LineNumber:     i + 1,
			ItemID:         l.ItemID,
			Description:    l.Description,
			Quantity:       l.Quantity,
			UnitPrice:      l.UnitPrice,
			TaxableAmount:  taxable,
			TaxCode:        l.TaxCode,
			TaxRatePercent: l.TaxRatePercent,
			TaxAmount:      tax,
			TotalAmount:    total,
		})
	}

	return s.repo.CreateCreditNote(ctx, cn)
}

func (s *salesService) PostCreditNote(ctx context.Context, creditNoteID uuid.UUID) (*domain.CreditNote, error) {
	var creditNote *domain.CreditNote
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		creditNote, err = s.postCreditNote(txCtx, creditNoteID)
		return err
	})
	return creditNote, err
}

func (s *salesService) postCreditNote(ctx context.Context, creditNoteID uuid.UUID) (*domain.CreditNote, error) {
	dbtx := s.db.DBTX(ctx)
	if err := dbtx.QueryRow(ctx, "SELECT id FROM sales.credit_notes WHERE id = $1 FOR UPDATE", creditNoteID).Scan(&creditNoteID); err != nil {
		return nil, fmt.Errorf("failed to lock sales credit note: %w", err)
	}
	cn, err := s.repo.GetCreditNote(ctx, creditNoteID)
	if err != nil {
		return nil, err
	}
	if cn.Status == domain.CreditNoteStatusPosted {
		return nil, errors.New("credit note is already posted")
	}

	// 1. Create AR Open Item in receivables.open_items
	openItemID := uuid.New()
	_, err = dbtx.Exec(ctx, `
		INSERT INTO receivables.open_items (
			id, legal_entity_id, customer_id, document_type, document_number,
			document_id, document_date, due_date, currency_code, original_amount,
			allocated_amount, remaining_amount, status
		) VALUES ($1, $2, $3, 'CREDIT_NOTE', $4, $5, $6, $6, $7, $8, 0, $8, 'OPEN')
	`, openItemID, cn.LegalEntityID, cn.CustomerID, cn.CreditNoteNumber, cn.ID, cn.IssueDate, cn.CurrencyCode, cn.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit note open item: %w", err)
	}

	// 2. Reduce Customer Balance in receivables.customer_credit_limits
	if _, err = dbtx.Exec(ctx, `
		UPDATE receivables.customer_credit_limits
		SET current_balance = current_balance - $2, updated_at = now()
		WHERE customer_id = $1
	`, cn.CustomerID, cn.TotalAmount); err != nil {
		return nil, fmt.Errorf("failed to reduce customer balance: %w", err)
	}

	// 3. Post Reverse Journal Entry
	var journalEntryID uuid.UUID
	var jeIDPtr *uuid.UUID
	var journalID, periodID, arAccountID, revAccountID, vatAccountID uuid.UUID
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.journals WHERE legal_entity_id = $1 ORDER BY (code = 'SALES') DESC LIMIT 1", cn.LegalEntityID).Scan(&journalID); err != nil {
		return nil, fmt.Errorf("sales journal is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.fiscal_periods WHERE legal_entity_id = $1 AND status = 'OPEN' LIMIT 1", cn.LegalEntityID).Scan(&periodID); err != nil {
		return nil, fmt.Errorf("open fiscal period is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '112000' LIMIT 1", cn.LegalEntityID).Scan(&arAccountID); err != nil {
		return nil, fmt.Errorf("accounts receivable account 112000 is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '410000' LIMIT 1", cn.LegalEntityID).Scan(&revAccountID); err != nil {
		return nil, fmt.Errorf("sales revenue account 410000 is not configured: %w", err)
	}
	if err := dbtx.QueryRow(ctx, "SELECT id FROM accounting.accounts WHERE legal_entity_id = $1 AND code = '213000' LIMIT 1", cn.LegalEntityID).Scan(&vatAccountID); err != nil {
		return nil, fmt.Errorf("output VAT account 213000 is not configured: %w", err)
	}

	creatorID := cn.CreatedByUserID
	if creatorID == nil {
		var adminID uuid.UUID
		if err := dbtx.QueryRow(ctx, "SELECT id FROM iam.users LIMIT 1").Scan(&adminID); err != nil {
			return nil, fmt.Errorf("failed to resolve credit-note creator: %w", err)
		}
		creatorID = &adminID
	}

	journalEntryID = uuid.New()
	entryNumber := identifiers.NewDocumentNumber("JE-CN", time.Now())
	_, err = dbtx.Exec(ctx, `
			INSERT INTO accounting.journal_entries (
				id, legal_entity_id, journal_id, fiscal_period_id, entry_number,
				accounting_date, status, source_module, source_type, source_id,
				reference, description, currency_code, exchange_rate,
				total_debit, total_credit, created_by, posted_by, posted_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'DRAFT', 'SALES', 'CREDIT_NOTE', $7,
				$8, $9, $10, 1.0, $11, $11, $12, $12, now()
			)
	`, journalEntryID, cn.LegalEntityID, journalID, periodID, entryNumber, cn.IssueDate, cn.ID,
		cn.CreditNoteNumber, "إثبات إشعار دائن - "+cn.CreditNoteNumber, cn.CurrencyCode, cn.TotalAmount, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit-note journal entry: %w", err)
	}

	jeIDPtr = &journalEntryID
	_, err = dbtx.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 1, $3, 'تسوية مردودات ومسموحات مبيعات', $4, $5, 0, $5, 0)
	`, journalEntryID, cn.LegalEntityID, revAccountID, cn.CurrencyCode, cn.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("failed to create revenue reversal journal line: %w", err)
	}

	if cn.TaxAmount.GreaterThan(decimal.Zero) {
		_, err = dbtx.Exec(ctx, `
					INSERT INTO accounting.journal_lines (
						journal_entry_id, legal_entity_id, line_number, account_id,
						description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
					) VALUES ($1, $2, 2, $3, 'تسوية ضريبة القيمة المضافة', $4, $5, 0, $5, 0)
		`, journalEntryID, cn.LegalEntityID, vatAccountID, cn.CurrencyCode, cn.TaxAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to create VAT reversal journal line: %w", err)
		}
	}

	// Line 3: Credit AR (112000)
	_, err = dbtx.Exec(ctx, `
				INSERT INTO accounting.journal_lines (
					journal_entry_id, legal_entity_id, line_number, account_id,
					description, currency_code, transaction_debit, transaction_credit, base_debit, base_credit
				) VALUES ($1, $2, 3, $3, 'تخفيض حساب العميل', $4, 0, $5, 0, $5)
	`, journalEntryID, cn.LegalEntityID, arAccountID, cn.CurrencyCode, cn.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to create receivable reversal journal line: %w", err)
	}
	if _, err = dbtx.Exec(ctx, `
		UPDATE accounting.journal_entries
		SET status = 'POSTED', posted_by = $1, posted_at = now()
		WHERE id = $2
	`, creatorID, journalEntryID); err != nil {
		return nil, fmt.Errorf("failed to post credit-note journal entry: %w", err)
	}

	return s.repo.PostCreditNote(ctx, cn.ID, jeIDPtr, &openItemID)
}

func (s *salesService) GetCreditNote(ctx context.Context, id uuid.UUID) (*domain.CreditNote, error) {
	return s.repo.GetCreditNote(ctx, id)
}

func (s *salesService) ListCreditNotes(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, originalInvoiceID *uuid.UUID, status *domain.CreditNoteStatus, limit, offset int) ([]*domain.CreditNote, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListCreditNotes(ctx, legalEntityID, customerID, originalInvoiceID, status, limit, offset)
}
