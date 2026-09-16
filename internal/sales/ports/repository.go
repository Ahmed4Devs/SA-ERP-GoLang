package ports

import (
	"context"

	"erp/internal/sales/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository defines data access operations for sales documents
type Repository interface {
	// Quotations
	CreateQuotation(ctx context.Context, q *domain.Quotation) (*domain.Quotation, error)
	GetQuotation(ctx context.Context, id uuid.UUID) (*domain.Quotation, error)
	ListQuotations(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.QuotationStatus, limit, offset int) ([]*domain.Quotation, error)
	UpdateQuotationStatus(ctx context.Context, id uuid.UUID, status domain.QuotationStatus, convertedOrderID *uuid.UUID) (*domain.Quotation, error)

	// Orders
	CreateOrder(ctx context.Context, o *domain.SalesOrder) (*domain.SalesOrder, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.SalesOrder, error)
	ListOrders(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, status *domain.OrderStatus, limit, offset int) ([]*domain.SalesOrder, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) (*domain.SalesOrder, error)
	UpdateOrderLineDeliveredQty(ctx context.Context, lineID uuid.UUID, qtyDelta decimal.Decimal) error
	UpdateOrderLineInvoicedQty(ctx context.Context, lineID uuid.UUID, qtyDelta decimal.Decimal) error

	// Deliveries
	CreateDelivery(ctx context.Context, d *domain.Delivery) (*domain.Delivery, error)
	GetDelivery(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	ListDeliveries(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, orderID *uuid.UUID, status *domain.DeliveryStatus, limit, offset int) ([]*domain.Delivery, error)
	UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status domain.DeliveryStatus) (*domain.Delivery, error)

	// Invoices
	CreateInvoice(ctx context.Context, inv *domain.SalesInvoice) (*domain.SalesInvoice, error)
	GetInvoice(ctx context.Context, id uuid.UUID) (*domain.SalesInvoice, error)
	ListInvoices(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, invoiceType *domain.InvoiceType, status *domain.InvoiceStatus, limit, offset int) ([]*domain.SalesInvoice, error)
	PostInvoice(ctx context.Context, id uuid.UUID, journalEntryID, openItemID *uuid.UUID, qrCodeData *string) (*domain.SalesInvoice, error)
	UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status domain.InvoiceStatus) (*domain.SalesInvoice, error)

	// Credit Notes
	CreateCreditNote(ctx context.Context, cn *domain.CreditNote) (*domain.CreditNote, error)
	GetCreditNote(ctx context.Context, id uuid.UUID) (*domain.CreditNote, error)
	ListCreditNotes(ctx context.Context, legalEntityID uuid.UUID, customerID *uuid.UUID, originalInvoiceID *uuid.UUID, status *domain.CreditNoteStatus, limit, offset int) ([]*domain.CreditNote, error)
	PostCreditNote(ctx context.Context, id uuid.UUID, journalEntryID, openItemID *uuid.UUID) (*domain.CreditNote, error)
}
