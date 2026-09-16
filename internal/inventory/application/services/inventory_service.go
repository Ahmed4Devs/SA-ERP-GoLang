package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/inventory/application/costing"
	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InventoryService struct {
	repo    ports.StockRepository
	costing *costing.CostingService
}

func NewInventoryService(repo ports.StockRepository) ports.InventoryService {
	return &InventoryService{
		repo:    repo,
		costing: costing.NewCostingService(repo),
	}
}

// 1. Warehouses & Bins
func (s *InventoryService) ListWarehouses(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Warehouse, error) {
	return s.repo.ListWarehouses(ctx, legalEntityID)
}

func (s *InventoryService) GetWarehouse(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	return s.repo.GetWarehouseByID(ctx, id)
}

func (s *InventoryService) CreateWarehouse(ctx context.Context, wh domain.Warehouse) (*domain.Warehouse, error) {
	if wh.ID == uuid.Nil {
		wh.ID = uuid.New()
	}
	wh.Code = strings.ToUpper(strings.TrimSpace(wh.Code))
	wh.IsActive = true
	createdWh, err := s.repo.CreateWarehouse(ctx, &wh)
	if err != nil {
		return nil, err
	}

	// Auto-create default system bin for this warehouse
	defaultBin := domain.BinLocation{
		ID:          uuid.New(),
		WarehouseID: createdWh.ID,
		Code:        "DEFAULT",
		IsDefault:   true,
		IsActive:    true,
	}
	_, _ = s.repo.CreateBin(ctx, &defaultBin)

	return createdWh, nil
}

func (s *InventoryService) ListBins(ctx context.Context, warehouseID uuid.UUID) ([]domain.BinLocation, error) {
	return s.repo.ListBins(ctx, warehouseID)
}

func (s *InventoryService) CreateBin(ctx context.Context, bin domain.BinLocation) (*domain.BinLocation, error) {
	if bin.ID == uuid.Nil {
		bin.ID = uuid.New()
	}
	bin.Code = strings.ToUpper(strings.TrimSpace(bin.Code))
	bin.IsActive = true
	return s.repo.CreateBin(ctx, &bin)
}

// 2. Movements & Ledger
func (s *InventoryService) PostStockMovement(ctx context.Context, req ports.PostMovementRequest) (*domain.StockMovement, error) {
	// Validate movement type
	switch req.MovementType {
	case domain.MovementTypeOpening, domain.MovementTypePurchaseReceipt, domain.MovementTypeSalesIssue,
		domain.MovementTypeCustomerReturn, domain.MovementTypeSupplierReturn,
		domain.MovementTypeTransferOut, domain.MovementTypeTransferIn,
		domain.MovementTypeAdjustmentIn, domain.MovementTypeAdjustmentOut,
		domain.MovementTypeProductionIn, domain.MovementTypeProductionOut,
		domain.MovementTypeReversal:
	default:
		return nil, domain.ErrInvalidMovementType
	}

	// Resolve Bin
	binID := uuid.Nil
	if req.BinID != nil && *req.BinID != uuid.Nil {
		binID = *req.BinID
	} else {
		defBin, err := s.repo.GetDefaultBin(ctx, req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get default warehouse bin: %w", err)
		}
		binID = defBin.ID
	}

	// Determine direction & quantity delta
	var quantityDelta decimal.Decimal
	var costDelta decimal.Decimal
	isIncoming := false

	switch req.MovementType {
	case domain.MovementTypeOpening, domain.MovementTypePurchaseReceipt, domain.MovementTypeCustomerReturn,
		domain.MovementTypeTransferIn, domain.MovementTypeAdjustmentIn, domain.MovementTypeProductionIn:
		isIncoming = true
		quantityDelta = req.Quantity.Abs()
		costDelta = quantityDelta.Mul(req.UnitCost)
	case domain.MovementTypeSalesIssue, domain.MovementTypeSupplierReturn,
		domain.MovementTypeTransferOut, domain.MovementTypeAdjustmentOut, domain.MovementTypeProductionOut:
		isIncoming = false
		quantityDelta = req.Quantity.Abs().Neg()
	case domain.MovementTypeReversal:
		if req.Quantity.IsPositive() {
			isIncoming = true
			quantityDelta = req.Quantity.Abs()
			costDelta = quantityDelta.Mul(req.UnitCost)
		} else {
			isIncoming = false
			quantityDelta = req.Quantity.Abs().Neg()
		}
	}

	stockKey := domain.StockKey{
		LegalEntityID: req.LegalEntityID,
		ItemID:        req.ItemID,
		WarehouseID:   req.WarehouseID,
		BinID:         binID,
	}

	var consumptions []domain.CostLayerConsumption

	// Lock balance and check availability for outgoing movements
	if !isIncoming {
		bal, err := s.repo.LockBalance(ctx, stockKey.LegalEntityID, stockKey.ItemID, stockKey.WarehouseID, stockKey.BinID)
		if err != nil {
			return nil, err
		}
		if bal == nil || bal.QuantityAvailable.LessThan(req.Quantity.Abs()) {
			return nil, domain.ErrInsufficientStock
		}
		// Calculate outgoing cost delta based on costing engine (Weighted Avg / FIFO / Standard)
		costResult, err := s.costing.CalculateIssueCost(ctx, req.LegalEntityID, req.ItemID, req.WarehouseID, req.Quantity.Abs())
		unitCost := bal.AverageCost
		if err == nil && costResult.TotalCost.IsPositive() {
			unitCost = costResult.UnitCost
			costDelta = costResult.TotalCost.Neg()
			consumptions = costResult.Consumptions
		} else {
			if req.UnitCost.IsPositive() {
				unitCost = req.UnitCost
			}
			costDelta = quantityDelta.Mul(unitCost) // Negative delta
		}
		req.UnitCost = unitCost
	}

	uomID := req.UOMID
	if uomID == uuid.Nil {
		var err error
		uomID, err = s.repo.GetItemBaseUOM(ctx, req.ItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup base uom for item %s: %w", req.ItemID, err)
		}
	}

	// Generate Movement Number
	movNum := fmt.Sprintf("MOV-%s-%d", strings.ToUpper(uuid.NewString()[:8]), time.Now().UnixNano()%1000000)

	m := domain.StockMovement{
		LegalEntityID:   req.LegalEntityID,
		MovementNumber:  movNum,
		MovementType:    req.MovementType,
		ItemID:          req.ItemID,
		SKUID:           req.SKUID,
		WarehouseID:     req.WarehouseID,
		BinID:           binID,
		LotNumber:       req.LotNumber,
		SerialNumber:    req.SerialNumber,
		Quantity:        quantityDelta,
		UOMID:           uomID,
		UnitCost:        req.UnitCost,
		TotalCost:       costDelta.Abs(),
		SourceDocType:   req.SourceDocType,
		SourceDocID:     req.SourceDocID,
		SourceDocLineID: req.SourceDocLineID,
		OccurredAt:      time.Now().UTC(),
		CreatedBy:       req.CreatedBy,
	}

	// Insert immutable movement record
	createdMov, err := s.repo.InsertMovement(ctx, &m)
	if err != nil {
		return nil, err
	}

	// Atomically apply projection delta to stock_balances
	_, err = s.repo.ApplyBalanceDelta(ctx, stockKey, quantityDelta, costDelta)
	if err != nil {
		return nil, err
	}

	// Inbound receipt: record cost layer
	if isIncoming && req.UnitCost.IsPositive() {
		_ = s.costing.RecordReceiptCost(ctx, req.LegalEntityID, req.ItemID, req.WarehouseID, quantityDelta, req.UnitCost, createdMov.ID)
	}

	// Outbound issue: persist FIFO layer consumptions if any
	for _, cons := range consumptions {
		cons.IssueMovementID = createdMov.ID
		_, _ = s.repo.InsertCostLayerConsumption(ctx, &cons)
	}

	return createdMov, nil
}

func (s *InventoryService) ListMovements(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, movType *string, limit, offset int32) ([]domain.StockMovement, error) {
	return s.repo.ListMovements(ctx, legalEntityID, itemID, warehouseID, movType, limit, offset)
}

func (s *InventoryService) GetMovement(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error) {
	return s.repo.GetMovementByID(ctx, id)
}

// 3. Balances
func (s *InventoryService) GetBalance(ctx context.Context, key domain.StockKey) (*domain.StockBalance, error) {
	return s.repo.GetBalance(ctx, key)
}

func (s *InventoryService) ListBalances(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, limit, offset int32) ([]domain.StockBalance, error) {
	return s.repo.ListBalances(ctx, legalEntityID, itemID, warehouseID, limit, offset)
}

// 4. Reservations
func (s *InventoryService) ReserveStock(ctx context.Context, req ports.ReserveStockRequest) (*domain.StockReservation, error) {
	res := domain.StockReservation{
		LegalEntityID:    req.LegalEntityID,
		ItemID:           req.ItemID,
		WarehouseID:      req.WarehouseID,
		BinID:            req.BinID,
		Quantity:         req.Quantity,
		DemandSourceType: req.DemandSourceType,
		DemandSourceID:   req.DemandSourceID,
	}
	return s.repo.Reserve(ctx, &res)
}

func (s *InventoryService) ReleaseStock(ctx context.Context, reservationID uuid.UUID) error {
	return s.repo.ReleaseReservation(ctx, reservationID)
}

// 5. Transfers
func (s *InventoryService) CreateTransferOrder(ctx context.Context, req ports.CreateTransferRequest) (*domain.TransferOrder, error) {
	if req.FromWarehouseID == req.ToWarehouseID {
		return nil, fmt.Errorf("source and destination warehouses must be different")
	}

	orderNum := fmt.Sprintf("TRF-%s-%d", strings.ToUpper(uuid.NewString()[:8]), time.Now().UnixNano()%1000000)
	order := domain.TransferOrder{
		LegalEntityID:   req.LegalEntityID,
		OrderNumber:     orderNum,
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Status:          domain.TransferStatusDraft,
		Notes:           req.Notes,
		CreatedBy:       req.CreatedBy,
	}

	lines := make([]domain.TransferOrderLine, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, domain.TransferOrderLine{
			ItemID:            l.ItemID,
			SKUID:             l.SKUID,
			RequestedQuantity: l.RequestedQuantity,
			UOMID:             l.UOMID,
			Notes:             l.Notes,
		})
	}

	return s.repo.CreateTransferOrder(ctx, &order, lines)
}

func (s *InventoryService) DispatchTransfer(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*domain.TransferOrder, error) {
	order, err := s.repo.GetTransferOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.TransferStatusDraft && order.Status != domain.TransferStatusRequested && order.Status != domain.TransferStatusApproved {
		return nil, domain.ErrInvalidTransferState
	}

	// Post TRANSFER_OUT movement for each line from the source warehouse
	docType := "TRANSFER_ORDER"
	for _, l := range order.Lines {
		_, err := s.PostStockMovement(ctx, ports.PostMovementRequest{
			LegalEntityID:   order.LegalEntityID,
			MovementType:    domain.MovementTypeTransferOut,
			ItemID:          l.ItemID,
			SKUID:           l.SKUID,
			WarehouseID:     order.FromWarehouseID,
			Quantity:        l.RequestedQuantity,
			UOMID:           l.UOMID,
			SourceDocType:   &docType,
			SourceDocID:     &order.ID,
			SourceDocLineID: &l.ID,
			CreatedBy:       userID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to dispatch line %s: %w", l.ItemCode, err)
		}
	}

	return s.repo.UpdateTransferOrderStatus(ctx, orderID, domain.TransferStatusDispatched)
}

func (s *InventoryService) ReceiveTransfer(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*domain.TransferOrder, error) {
	order, err := s.repo.GetTransferOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.TransferStatusDispatched {
		return nil, domain.ErrInvalidTransferState
	}

	// Post TRANSFER_IN movement for each line into the destination warehouse
	docType := "TRANSFER_ORDER"
	for _, l := range order.Lines {
		_, err := s.PostStockMovement(ctx, ports.PostMovementRequest{
			LegalEntityID:   order.LegalEntityID,
			MovementType:    domain.MovementTypeTransferIn,
			ItemID:          l.ItemID,
			SKUID:           l.SKUID,
			WarehouseID:     order.ToWarehouseID,
			Quantity:        l.RequestedQuantity,
			UOMID:           l.UOMID,
			SourceDocType:   &docType,
			SourceDocID:     &order.ID,
			SourceDocLineID: &l.ID,
			CreatedBy:       userID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to receive line %s: %w", l.ItemCode, err)
		}
	}

	return s.repo.UpdateTransferOrderStatus(ctx, orderID, domain.TransferStatusReceived)
}

func (s *InventoryService) ListTransferOrders(ctx context.Context, legalEntityID uuid.UUID, status *string, limit, offset int32) ([]domain.TransferOrder, error) {
	return s.repo.ListTransferOrders(ctx, legalEntityID, status, limit, offset)
}

func (s *InventoryService) GetTransferOrder(ctx context.Context, id uuid.UUID) (*domain.TransferOrder, error) {
	return s.repo.GetTransferOrder(ctx, id)
}

// 6. Adjustments
func (s *InventoryService) CreateAdjustment(ctx context.Context, req ports.CreateAdjustmentRequest) (*domain.StockAdjustment, error) {
	adjNum := fmt.Sprintf("ADJ-%s-%d", strings.ToUpper(uuid.NewString()[:8]), time.Now().UnixNano()%1000000)
	adj := domain.StockAdjustment{
		LegalEntityID:    req.LegalEntityID,
		AdjustmentNumber: adjNum,
		WarehouseID:      req.WarehouseID,
		AdjustmentType:   req.AdjustmentType,
		ReasonCode:       req.ReasonCode,
		Status:           domain.AdjustmentStatusDraft,
		AdjustmentDate:   time.Now().UTC(),
		Notes:            req.Notes,
		CreatedBy:        req.CreatedBy,
	}

	lines := make([]domain.StockAdjustmentLine, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, domain.StockAdjustmentLine{
			ItemID:       l.ItemID,
			BinID:        l.BinID,
			LotNumber:    l.LotNumber,
			SerialNumber: l.SerialNumber,
			Quantity:     l.Quantity,
			UnitCost:     l.UnitCost,
			TotalCost:    l.Quantity.Mul(l.UnitCost),
			Reason:       l.Reason,
		})
	}

	return s.repo.CreateStockAdjustment(ctx, &adj, lines)
}

func (s *InventoryService) PostAdjustment(ctx context.Context, adjustmentID uuid.UUID, userID uuid.UUID) (*domain.StockAdjustment, error) {
	adj, err := s.repo.GetStockAdjustment(ctx, adjustmentID)
	if err != nil {
		return nil, err
	}
	if adj.Status == domain.AdjustmentStatusPosted {
		return adj, nil
	}

	docType := "STOCK_ADJUSTMENT"
	movType := domain.MovementTypeAdjustmentIn
	if adj.AdjustmentType == domain.AdjustmentTypeDecrease || adj.AdjustmentType == domain.AdjustmentTypeScrap || adj.AdjustmentType == domain.AdjustmentTypeDamaged {
		movType = domain.MovementTypeAdjustmentOut
	}

	for _, l := range adj.Lines {
		_, err := s.PostStockMovement(ctx, ports.PostMovementRequest{
			LegalEntityID:   adj.LegalEntityID,
			MovementType:    movType,
			ItemID:          l.ItemID,
			WarehouseID:     adj.WarehouseID,
			BinID:           l.BinID,
			LotNumber:       l.LotNumber,
			SerialNumber:    l.SerialNumber,
			Quantity:        l.Quantity,
			UOMID:           l.UOMID,
			UnitCost:        l.UnitCost,
			SourceDocType:   &docType,
			SourceDocID:     &adj.ID,
			SourceDocLineID: &l.ID,
			CreatedBy:       userID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to post adjustment line: %w", err)
		}
	}

	return s.repo.UpdateStockAdjustmentStatus(ctx, adjustmentID, domain.AdjustmentStatusPosted, &userID)
}

func (s *InventoryService) ListAdjustments(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockAdjustment, error) {
	return s.repo.ListStockAdjustments(ctx, legalEntityID, warehouseID, status, limit, offset)
}

// 7. Counts
func (s *InventoryService) CreateStockCount(ctx context.Context, req ports.CreateCountRequest) (*domain.StockCount, error) {
	cntNum := fmt.Sprintf("CNT-%s-%d", strings.ToUpper(uuid.NewString()[:8]), time.Now().UnixNano()%1000000)
	count := domain.StockCount{
		LegalEntityID: req.LegalEntityID,
		CountNumber:   cntNum,
		WarehouseID:   req.WarehouseID,
		CountType:     req.CountType,
		Status:        domain.CountStatusPlanned,
		CountDate:     time.Now().UTC(),
	}

	lines := make([]domain.StockCountLine, 0, len(req.Lines))
	for _, l := range req.Lines {
		// Lookup current system quantity on hand
		binID := uuid.Nil
		if l.BinID != nil {
			binID = *l.BinID
		} else {
			defBin, err := s.repo.GetDefaultBin(ctx, req.WarehouseID)
			if err == nil && defBin != nil {
				binID = defBin.ID
			}
		}

		sysQty := decimal.Zero
		bal, _ := s.repo.GetBalance(ctx, domain.StockKey{
			LegalEntityID: req.LegalEntityID,
			ItemID:        l.ItemID,
			WarehouseID:   req.WarehouseID,
			BinID:         binID,
		})
		if bal != nil {
			sysQty = bal.QuantityOnHand
		}

		lines = append(lines, domain.StockCountLine{
			ItemID:          l.ItemID,
			BinID:           &binID,
			SystemQuantity:  sysQty,
			CountedQuantity: l.CountedQuantity,
			Notes:           l.Notes,
		})
	}

	return s.repo.CreateStockCount(ctx, &count, lines)
}

func (s *InventoryService) ReconcileStockCount(ctx context.Context, countID uuid.UUID, userID uuid.UUID) (*domain.StockCount, error) {
	cnt, err := s.repo.GetStockCount(ctx, countID)
	if err != nil {
		return nil, err
	}
	if cnt.Status == domain.CountStatusPosted {
		return cnt, nil
	}

	docType := "STOCK_COUNT"
	for _, l := range cnt.Lines {
		if l.Variance.IsZero() {
			continue
		}

		movType := domain.MovementTypeAdjustmentIn
		qty := l.Variance.Abs()
		if l.Variance.IsNegative() {
			movType = domain.MovementTypeAdjustmentOut
		}

		_, err := s.PostStockMovement(ctx, ports.PostMovementRequest{
			LegalEntityID:   cnt.LegalEntityID,
			MovementType:    movType,
			ItemID:          l.ItemID,
			WarehouseID:     cnt.WarehouseID,
			BinID:           l.BinID,
			Quantity:        qty,
			UOMID:           l.UOMID,
			SourceDocType:   &docType,
			SourceDocID:     &cnt.ID,
			SourceDocLineID: &l.ID,
			CreatedBy:       userID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to reconcile count variance for item %s: %w", l.ItemCode, err)
		}
	}

	return s.repo.UpdateStockCountStatus(ctx, countID, domain.CountStatusPosted, &userID)
}

func (s *InventoryService) ListStockCounts(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockCount, error) {
	return s.repo.ListStockCounts(ctx, legalEntityID, warehouseID, status, limit, offset)
}

// 8. Costing & Valuation
func (s *InventoryService) GetValuationReport(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID) ([]domain.ValuationSummaryItem, error) {
	return s.costing.GetValuationReport(ctx, entityID, warehouseID, itemID)
}

func (s *InventoryService) ListCostLayers(ctx context.Context, entityID uuid.UUID, itemID, warehouseID *uuid.UUID, onlyActive *bool, limit, offset int32) ([]domain.CostLayer, error) {
	return s.costing.ListCostLayers(ctx, entityID, itemID, warehouseID, onlyActive, limit, offset)
}

func (s *InventoryService) ExecuteRevaluation(ctx context.Context, req ports.ExecuteRevaluationRequest) (*domain.InventoryRevaluation, error) {
	return s.costing.ExecuteRevaluation(ctx, costing.ExecuteRevaluationRequest{
		LegalEntityID:  req.LegalEntityID,
		WarehouseID:    req.WarehouseID,
		ItemID:         req.ItemID,
		NewUnitCost:    req.NewUnitCost,
		Reason:         req.Reason,
		JournalEntryID: req.JournalEntryID,
		CreatedBy:      req.CreatedBy,
	})
}

func (s *InventoryService) ListRevaluations(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID, limit, offset int32) ([]domain.InventoryRevaluation, error) {
	return s.costing.ListRevaluations(ctx, entityID, warehouseID, itemID, limit, offset)
}
