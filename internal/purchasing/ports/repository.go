package ports

import (
	"context"

	"erp/internal/purchasing/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Repository defines data access operations for purchasing documents
type Repository interface {
	// Purchase Requests
	CreatePurchaseRequest(ctx context.Context, pr *domain.PurchaseRequest) (*domain.PurchaseRequest, error)
	GetPurchaseRequest(ctx context.Context, id uuid.UUID) (*domain.PurchaseRequest, error)
	ListPurchaseRequests(ctx context.Context, legalEntityID uuid.UUID, status *domain.PRStatus, limit, offset int) ([]*domain.PurchaseRequest, error)
	UpdatePurchaseRequestStatus(ctx context.Context, id uuid.UUID, status domain.PRStatus) (*domain.PurchaseRequest, error)
	UpdatePRLineOrderedQty(ctx context.Context, lineID uuid.UUID, delta decimal.Decimal) error

	// RFQs
	CreateRFQ(ctx context.Context, rfq *domain.RFQ) (*domain.RFQ, error)
	GetRFQ(ctx context.Context, id uuid.UUID) (*domain.RFQ, error)
	ListRFQs(ctx context.Context, legalEntityID uuid.UUID, status *domain.RFQStatus, limit, offset int) ([]*domain.RFQ, error)
	UpdateRFQStatus(ctx context.Context, id uuid.UUID, status domain.RFQStatus) (*domain.RFQ, error)

	// Supplier Quotes
	CreateSupplierQuote(ctx context.Context, sq *domain.SupplierQuote) (*domain.SupplierQuote, error)
	GetSupplierQuote(ctx context.Context, id uuid.UUID) (*domain.SupplierQuote, error)
	ListSupplierQuotes(ctx context.Context, legalEntityID uuid.UUID, rfqID, supplierID *uuid.UUID, status *domain.QuoteStatus, limit, offset int) ([]*domain.SupplierQuote, error)
	UpdateSupplierQuoteStatus(ctx context.Context, id uuid.UUID, status domain.QuoteStatus) (*domain.SupplierQuote, error)
	AwardSupplierQuoteLine(ctx context.Context, lineID uuid.UUID, awarded bool) error

	// Purchase Orders
	CreatePurchaseOrder(ctx context.Context, po *domain.PurchaseOrder) (*domain.PurchaseOrder, error)
	GetPurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error)
	GetPurchaseOrderByNumber(ctx context.Context, legalEntityID uuid.UUID, poNumber string) (*domain.PurchaseOrder, error)
	ListPurchaseOrders(ctx context.Context, legalEntityID uuid.UUID, supplierID *uuid.UUID, status *domain.POStatus, limit, offset int) ([]*domain.PurchaseOrder, error)
	UpdatePurchaseOrderStatus(ctx context.Context, id uuid.UUID, status domain.POStatus) (*domain.PurchaseOrder, error)
	ApprovePurchaseOrder(ctx context.Context, id uuid.UUID, approverID uuid.UUID) (*domain.PurchaseOrder, error)
	IssuePurchaseOrder(ctx context.Context, id uuid.UUID) (*domain.PurchaseOrder, error)
	UpdatePOLineReceivedQty(ctx context.Context, lineID uuid.UUID, delta decimal.Decimal) error
	UpdatePOLineBilledQty(ctx context.Context, lineID uuid.UUID, delta decimal.Decimal) error

	// Goods Receipts
	CreateGoodsReceipt(ctx context.Context, gr *domain.GoodsReceipt) (*domain.GoodsReceipt, error)
	GetGoodsReceipt(ctx context.Context, id uuid.UUID) (*domain.GoodsReceipt, error)
	GetGoodsReceiptByNumber(ctx context.Context, legalEntityID uuid.UUID, receiptNumber string) (*domain.GoodsReceipt, error)
	ListGoodsReceipts(ctx context.Context, legalEntityID uuid.UUID, poID, supplierID *uuid.UUID, status *domain.GRNStatus, limit, offset int) ([]*domain.GoodsReceipt, error)
	UpdateGoodsReceiptStatus(ctx context.Context, id uuid.UUID, status domain.GRNStatus) (*domain.GoodsReceipt, error)
	UpdateGoodsReceiptLineMovement(ctx context.Context, lineID uuid.UUID, movementID uuid.UUID) error
}
