package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/inventory/application/costing"
	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"
	"erp/internal/platform/database"
	"erp/internal/platform/identifiers"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InventoryService struct {
	repo    ports.StockRepository
	costing *costing.CostingService
	db      *database.DB
	tx      transaction.Manager
}

func NewInventoryService(repo ports.StockRepository, db *database.DB) ports.InventoryService {
	return &InventoryService{
		repo:    repo,
		costing: costing.NewCostingService(repo),
		db:      db,
		tx:      transaction.NewManager(db.Pool),
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
	var createdWh *domain.Warehouse
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		createdWh, err = s.repo.CreateWarehouse(txCtx, &wh)
		if err != nil {
			return err
		}

		defaultBin := domain.BinLocation{
			ID:          uuid.New(),
			WarehouseID: createdWh.ID,
			Code:        "DEFAULT",
			IsDefault:   true,
			IsActive:    true,
		}
		if _, err := s.repo.CreateBin(txCtx, &defaultBin); err != nil {
			return fmt.Errorf("failed to create default warehouse bin: %w", err)
		}
		return nil
	})
	return createdWh, err
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
	var movement *domain.StockMovement
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		movement, err = s.postStockMovement(txCtx, req)
		return err
	})
	return movement, err
}

func (s *InventoryService) postStockMovement(ctx context.Context, req ports.PostMovementRequest) (*domain.StockMovement, error) {
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
	if req.Quantity.IsZero() {
		return nil, fmt.Errorf("movement quantity must not be zero")
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
	if isIncoming && req.UnitCost.IsNegative() {
		return nil, domain.ErrNegativeUnitCost
	}

	stockKey := domain.StockKey{
		LegalEntityID: req.LegalEntityID,
		ItemID:        req.ItemID,
		WarehouseID:   req.WarehouseID,
		BinID:         binID,
	}
	if err := s.lockStock(ctx, stockKey.LegalEntityID, stockKey.ItemID, stockKey.WarehouseID); err != nil {
		return nil, err
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
		if err != nil {
			return nil, fmt.Errorf("failed to calculate issue cost: %w", err)
		}
		unitCost := bal.AverageCost
		if costResult.TotalCost.IsPositive() {
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
	movNum := identifiers.NewDocumentNumber("MOV", time.Now())

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
		if err := s.costing.RecordReceiptCost(ctx, req.LegalEntityID, req.ItemID, req.WarehouseID, quantityDelta, req.UnitCost, createdMov.ID); err != nil {
			return nil, fmt.Errorf("failed to record receipt cost: %w", err)
		}
	}

	// Outbound issue: persist FIFO layer consumptions if any
	for _, cons := range consumptions {
		cons.IssueMovementID = createdMov.ID
		if _, err := s.repo.InsertCostLayerConsumption(ctx, &cons); err != nil {
			return nil, fmt.Errorf("failed to record cost layer consumption: %w", err)
		}
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
	var reservation *domain.StockReservation
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockStock(txCtx, req.LegalEntityID, req.ItemID, req.WarehouseID); err != nil {
			return err
		}
		var err error
		reservation, err = s.repo.Reserve(txCtx, &res)
		return err
	})
	return reservation, err
}

func (s *InventoryService) ReleaseStock(ctx context.Context, reservationID uuid.UUID) error {
	return s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.db.DBTX(txCtx).QueryRow(txCtx,
			"SELECT id FROM inventory.stock_reservations WHERE id = $1 FOR UPDATE", reservationID,
		).Scan(&reservationID); err != nil {
			return fmt.Errorf("failed to lock stock reservation: %w", err)
		}
		return s.repo.ReleaseReservation(txCtx, reservationID)
	})
}

// 5. Transfers
func (s *InventoryService) CreateTransferOrder(ctx context.Context, req ports.CreateTransferRequest) (*domain.TransferOrder, error) {
	if req.FromWarehouseID == req.ToWarehouseID {
		return nil, fmt.Errorf("source and destination warehouses must be different")
	}

	orderNum := identifiers.NewDocumentNumber("TRF", time.Now())
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
	var dispatched *domain.TransferOrder
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockRecord(txCtx, "inventory.transfer_orders", orderID); err != nil {
			return fmt.Errorf("failed to lock transfer order: %w", err)
		}
		order, err := s.repo.GetTransferOrder(txCtx, orderID)
		if err != nil {
			return err
		}
		if order.Status != domain.TransferStatusDraft && order.Status != domain.TransferStatusRequested && order.Status != domain.TransferStatusApproved {
			return domain.ErrInvalidTransferState
		}

		docType := "TRANSFER_ORDER"
		for _, l := range order.Lines {
			_, err := s.postStockMovement(txCtx, ports.PostMovementRequest{
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
				return fmt.Errorf("failed to dispatch line %s: %w", l.ItemCode, err)
			}
			if err := s.repo.UpdateTransferLineQuantities(txCtx, l.ID, l.RequestedQuantity, l.ReceivedQuantity); err != nil {
				return fmt.Errorf("failed to update dispatched quantity for line %s: %w", l.ItemCode, err)
			}
		}
		dispatched, err = s.repo.UpdateTransferOrderStatus(txCtx, orderID, domain.TransferStatusDispatched)
		return err
	})
	return dispatched, err
}

func (s *InventoryService) ReceiveTransfer(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*domain.TransferOrder, error) {
	var received *domain.TransferOrder
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockRecord(txCtx, "inventory.transfer_orders", orderID); err != nil {
			return fmt.Errorf("failed to lock transfer order: %w", err)
		}
		order, err := s.repo.GetTransferOrder(txCtx, orderID)
		if err != nil {
			return err
		}
		if order.Status != domain.TransferStatusDispatched {
			return domain.ErrInvalidTransferState
		}

		docType := "TRANSFER_ORDER"
		for _, l := range order.Lines {
			_, err := s.postStockMovement(txCtx, ports.PostMovementRequest{
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
				return fmt.Errorf("failed to receive line %s: %w", l.ItemCode, err)
			}
			if err := s.repo.UpdateTransferLineQuantities(txCtx, l.ID, l.DispatchedQuantity, l.RequestedQuantity); err != nil {
				return fmt.Errorf("failed to update received quantity for line %s: %w", l.ItemCode, err)
			}
		}
		received, err = s.repo.UpdateTransferOrderStatus(txCtx, orderID, domain.TransferStatusReceived)
		return err
	})
	return received, err
}

func (s *InventoryService) ListTransferOrders(ctx context.Context, legalEntityID uuid.UUID, status *string, limit, offset int32) ([]domain.TransferOrder, error) {
	return s.repo.ListTransferOrders(ctx, legalEntityID, status, limit, offset)
}

func (s *InventoryService) GetTransferOrder(ctx context.Context, id uuid.UUID) (*domain.TransferOrder, error) {
	return s.repo.GetTransferOrder(ctx, id)
}

// 6. Adjustments
func (s *InventoryService) CreateAdjustment(ctx context.Context, req ports.CreateAdjustmentRequest) (*domain.StockAdjustment, error) {
	adjNum := identifiers.NewDocumentNumber("ADJ", time.Now())
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
	var posted *domain.StockAdjustment
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockRecord(txCtx, "inventory.stock_adjustments", adjustmentID); err != nil {
			return fmt.Errorf("failed to lock stock adjustment: %w", err)
		}
		adj, err := s.repo.GetStockAdjustment(txCtx, adjustmentID)
		if err != nil {
			return err
		}
		if adj.Status == domain.AdjustmentStatusPosted {
			posted = adj
			return nil
		}

		docType := "STOCK_ADJUSTMENT"
		movType := domain.MovementTypeAdjustmentIn
		if adj.AdjustmentType == domain.AdjustmentTypeDecrease || adj.AdjustmentType == domain.AdjustmentTypeScrap || adj.AdjustmentType == domain.AdjustmentTypeDamaged {
			movType = domain.MovementTypeAdjustmentOut
		}

		for _, l := range adj.Lines {
			_, err := s.postStockMovement(txCtx, ports.PostMovementRequest{
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
				return fmt.Errorf("failed to post adjustment line: %w", err)
			}
		}
		posted, err = s.repo.UpdateStockAdjustmentStatus(txCtx, adjustmentID, domain.AdjustmentStatusPosted, &userID)
		return err
	})
	return posted, err
}

func (s *InventoryService) ListAdjustments(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockAdjustment, error) {
	return s.repo.ListStockAdjustments(ctx, legalEntityID, warehouseID, status, limit, offset)
}

// 7. Counts
func (s *InventoryService) CreateStockCount(ctx context.Context, req ports.CreateCountRequest) (*domain.StockCount, error) {
	cntNum := identifiers.NewDocumentNumber("CNT", time.Now())
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
			if err != nil {
				return nil, fmt.Errorf("failed to get default warehouse bin: %w", err)
			}
			binID = defBin.ID
		}

		sysQty := decimal.Zero
		bal, err := s.repo.GetBalance(ctx, domain.StockKey{
			LegalEntityID: req.LegalEntityID,
			ItemID:        l.ItemID,
			WarehouseID:   req.WarehouseID,
			BinID:         binID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get stock balance for count: %w", err)
		}
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
	var reconciled *domain.StockCount
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockRecord(txCtx, "inventory.stock_counts", countID); err != nil {
			return fmt.Errorf("failed to lock stock count: %w", err)
		}
		cnt, err := s.repo.GetStockCount(txCtx, countID)
		if err != nil {
			return err
		}
		if cnt.Status == domain.CountStatusPosted {
			reconciled = cnt
			return nil
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

			_, err := s.postStockMovement(txCtx, ports.PostMovementRequest{
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
				return fmt.Errorf("failed to reconcile count variance for item %s: %w", l.ItemCode, err)
			}
		}
		reconciled, err = s.repo.UpdateStockCountStatus(txCtx, countID, domain.CountStatusPosted, &userID)
		return err
	})
	return reconciled, err
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
	var revaluation *domain.InventoryRevaluation
	err := s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := s.lockStock(txCtx, req.LegalEntityID, req.ItemID, req.WarehouseID); err != nil {
			return err
		}
		var err error
		revaluation, err = s.costing.ExecuteRevaluation(txCtx, costing.ExecuteRevaluationRequest{
			LegalEntityID:  req.LegalEntityID,
			WarehouseID:    req.WarehouseID,
			ItemID:         req.ItemID,
			NewUnitCost:    req.NewUnitCost,
			Reason:         req.Reason,
			JournalEntryID: req.JournalEntryID,
			CreatedBy:      req.CreatedBy,
		})
		return err
	})
	return revaluation, err
}

func (s *InventoryService) lockStock(ctx context.Context, entityID, itemID, warehouseID uuid.UUID) error {
	key := entityID.String() + ":" + itemID.String() + ":" + warehouseID.String()
	if _, err := s.db.DBTX(ctx).Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", key); err != nil {
		return fmt.Errorf("failed to lock stock position: %w", err)
	}
	return nil
}

func (s *InventoryService) lockRecord(ctx context.Context, table string, id uuid.UUID) error {
	var query string
	switch table {
	case "inventory.transfer_orders":
		query = "SELECT id FROM inventory.transfer_orders WHERE id = $1 FOR UPDATE"
	case "inventory.stock_adjustments":
		query = "SELECT id FROM inventory.stock_adjustments WHERE id = $1 FOR UPDATE"
	case "inventory.stock_counts":
		query = "SELECT id FROM inventory.stock_counts WHERE id = $1 FOR UPDATE"
	default:
		return fmt.Errorf("unsupported inventory lock target %q", table)
	}
	return s.db.DBTX(ctx).QueryRow(ctx, query, id).Scan(&id)
}

func (s *InventoryService) ListRevaluations(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID, limit, offset int32) ([]domain.InventoryRevaluation, error) {
	return s.costing.ListRevaluations(ctx, entityID, warehouseID, itemID, limit, offset)
}
