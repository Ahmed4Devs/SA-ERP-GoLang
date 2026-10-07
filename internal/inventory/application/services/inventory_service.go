package services

import (
	"context"
	postingapp "erp/internal/accounting/application/posting"
	accountingdomain "erp/internal/accounting/domain"
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
	ledger  *postingapp.SourceService
	repo    ports.StockRepository
	costing *costing.CostingService
	db      *database.DB
	tx      transaction.Manager
}

func NewInventoryService(repo ports.StockRepository, db *database.DB, ledger *postingapp.SourceService) ports.InventoryService {
	return &InventoryService{
		repo:    repo,
		ledger:  ledger,
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

	var scope bool
	if err := s.db.DBTX(ctx).QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM products.items i JOIN warehouse.warehouses w ON w.legal_entity_id=i.legal_entity_id JOIN warehouse.bin_locations b ON b.warehouse_id=w.id WHERE i.id=$1 AND w.id=$2 AND b.id=$3 AND i.legal_entity_id=$4)`, req.ItemID, req.WarehouseID, binID, req.LegalEntityID).Scan(&scope); err != nil {
		return nil, err
	}
	if !scope {
		return nil, fmt.Errorf("item, warehouse and bin must belong to the movement legal entity")
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
	if req.CreatedBy == uuid.Nil || len(req.Lines) == 0 {
		return nil, fmt.Errorf("adjustment requires a creator and lines")
	}
	for _, line := range req.Lines {
		if !line.Quantity.IsPositive() || line.UnitCost.IsNegative() {
			return nil, fmt.Errorf("adjustment requires positive quantities and nonnegative costs")
		}
	}
	adjNum := identifiers.NewDocumentNumber("ADJ", time.Now())
	adj := domain.StockAdjustment{
		LegalEntityID:      req.LegalEntityID,
		AdjustmentNumber:   adjNum,
		InventoryAccountID: req.InventoryAccountID, OffsetAccountID: req.OffsetAccountID,
		WarehouseID:    req.WarehouseID,
		AdjustmentType: req.AdjustmentType,
		ReasonCode:     req.ReasonCode,
		Status:         domain.AdjustmentStatusDraft,
		AdjustmentDate: time.Now().UTC(),
		Notes:          req.Notes,
		CreatedBy:      req.CreatedBy,
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
	if userID == uuid.Nil {
		return nil, fmt.Errorf("posting actor is required")
	}
	var posted *domain.StockAdjustment
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
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

		if adj.CreatedBy == userID || (adj.Status != domain.AdjustmentStatusDraft && adj.Status != domain.AdjustmentStatusApproved) {
			return fmt.Errorf("adjustment requires an independent posting actor and an open state")
		}
		var variance decimal.Decimal
		docType := "STOCK_ADJUSTMENT"
		movType := domain.MovementTypeAdjustmentIn
		if adj.AdjustmentType == domain.AdjustmentTypeDecrease || adj.AdjustmentType == domain.AdjustmentTypeScrap || adj.AdjustmentType == domain.AdjustmentTypeDamaged {
			movType = domain.MovementTypeAdjustmentOut
		}

		for _, l := range adj.Lines {
			movement, err := s.postStockMovement(txCtx, ports.PostMovementRequest{
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
			if movement.Quantity.IsPositive() {
				variance = variance.Add(movement.TotalCost)
			} else {
				variance = variance.Sub(movement.TotalCost)
			}
		}
		journal, err := s.postVariance(txCtx, adj.LegalEntityID, adj.ID, userID, "ADJUSTMENT", adj.AdjustmentNumber, adj.AdjustmentDate, adj.InventoryAccountID, adj.OffsetAccountID, variance)
		if err != nil {
			return err
		}
		if _, err := s.db.DBTX(txCtx).Exec(txCtx, "UPDATE inventory.stock_adjustments SET journal_entry_id=$2 WHERE id=$1", adj.ID, journal); err != nil {
			return err
		}
		posted, err = s.repo.UpdateStockAdjustmentStatus(txCtx, adjustmentID, domain.AdjustmentStatusPosted, &userID)
		if err != nil {
			return err
		}
		return s.ledger.RecordSource(txCtx, adj.LegalEntityID, userID, adj.ID, "inventory.adjustment.post", posted)
	})
	return posted, err
}

func (s *InventoryService) ListAdjustments(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockAdjustment, error) {
	return s.repo.ListStockAdjustments(ctx, legalEntityID, warehouseID, status, limit, offset)
}

// 7. Counts
func (s *InventoryService) CreateStockCount(ctx context.Context, req ports.CreateCountRequest) (*domain.StockCount, error) {
	if req.CreatedBy == uuid.Nil || len(req.Lines) == 0 {
		return nil, fmt.Errorf("count requires a creator and lines")
	}
	for _, line := range req.Lines {
		if line.CountedQuantity.IsNegative() || line.UnitCost.IsNegative() {
			return nil, fmt.Errorf("count quantities and costs cannot be negative")
		}
	}
	cntNum := identifiers.NewDocumentNumber("CNT", time.Now())
	count := domain.StockCount{
		LegalEntityID: req.LegalEntityID,
		CountNumber:   cntNum,
		CreatedBy:     &req.CreatedBy, InventoryAccountID: req.InventoryAccountID, OffsetAccountID: req.OffsetAccountID,
		WarehouseID: req.WarehouseID,
		CountType:   req.CountType,
		Status:      domain.CountStatusPlanned,
		CountDate:   time.Now().UTC(),
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
			CountedQuantity: l.CountedQuantity, UnitCost: l.UnitCost,
			Notes: l.Notes,
		})
	}

	return s.repo.CreateStockCount(ctx, &count, lines)
}

func (s *InventoryService) ReconcileStockCount(ctx context.Context, countID uuid.UUID, userID uuid.UUID) (*domain.StockCount, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("posting actor is required")
	}
	var reconciled *domain.StockCount
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
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

		if cnt.CreatedBy == nil || *cnt.CreatedBy == userID || cnt.Status == domain.CountStatusCancelled {
			return fmt.Errorf("count requires an independent posting actor and an open state")
		}
		var variance decimal.Decimal
		docType := "STOCK_COUNT"
		for _, l := range cnt.Lines {
			if l.Variance.IsZero() {
				continue
			}

			if err := s.lockStock(txCtx, cnt.LegalEntityID, l.ItemID, cnt.WarehouseID); err != nil {
				return err
			}
			bin := uuid.Nil
			if l.BinID != nil {
				bin = *l.BinID
			}
			balance, err := s.repo.GetBalance(txCtx, domain.StockKey{LegalEntityID: cnt.LegalEntityID, ItemID: l.ItemID, WarehouseID: cnt.WarehouseID, BinID: bin})
			if err != nil {
				return err
			}
			quantity, cost := decimal.Zero, l.UnitCost
			if balance != nil {
				quantity = balance.QuantityOnHand
				if balance.AverageCost.IsPositive() {
					cost = balance.AverageCost
				}
			}
			if !quantity.Equal(l.SystemQuantity) {
				return fmt.Errorf("stock changed after count; recount before posting")
			}
			if l.Variance.IsPositive() && !cost.IsPositive() {
				return fmt.Errorf("positive count variance requires a unit cost")
			}
			movType := domain.MovementTypeAdjustmentIn
			qty := l.Variance.Abs()
			if l.Variance.IsNegative() {
				movType = domain.MovementTypeAdjustmentOut
			}

			movement, err := s.postStockMovement(txCtx, ports.PostMovementRequest{
				LegalEntityID: cnt.LegalEntityID,
				MovementType:  movType,
				ItemID:        l.ItemID,
				WarehouseID:   cnt.WarehouseID,
				BinID:         l.BinID,
				Quantity:      qty, UnitCost: cost,
				UOMID:           l.UOMID,
				SourceDocType:   &docType,
				SourceDocID:     &cnt.ID,
				SourceDocLineID: &l.ID,
				CreatedBy:       userID,
			})
			if err != nil {
				return fmt.Errorf("failed to reconcile count variance for item %s: %w", l.ItemCode, err)
			}
			if movement.Quantity.IsPositive() {
				variance = variance.Add(movement.TotalCost)
			} else {
				variance = variance.Sub(movement.TotalCost)
			}
		}
		journal, err := s.postVariance(txCtx, cnt.LegalEntityID, cnt.ID, userID, "COUNT", cnt.CountNumber, cnt.CountDate, cnt.InventoryAccountID, cnt.OffsetAccountID, variance)
		if err != nil {
			return err
		}
		if _, err := s.db.DBTX(txCtx).Exec(txCtx, "UPDATE inventory.stock_counts SET journal_entry_id=$2 WHERE id=$1", cnt.ID, journal); err != nil {
			return err
		}
		reconciled, err = s.repo.UpdateStockCountStatus(txCtx, countID, domain.CountStatusPosted, &userID)
		if err != nil {
			return err
		}
		return s.ledger.RecordSource(txCtx, cnt.LegalEntityID, userID, cnt.ID, "inventory.count.post", reconciled)
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
	if req.CreatedBy == uuid.Nil {
		return nil, fmt.Errorf("revaluation actor is required")
	}
	var revaluation *domain.InventoryRevaluation
	err := s.ledger.WithinTx(ctx, func(txCtx context.Context) error {
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
			JournalEntryID: nil,
			CreatedBy:      req.CreatedBy,
		})
		if err != nil {
			return err
		}
		journal, err := s.postVariance(txCtx, req.LegalEntityID, revaluation.ID, req.CreatedBy, "REVALUATION", revaluation.RevaluationNumber, time.Now(), req.InventoryAccountID, req.OffsetAccountID, revaluation.VarianceAmount)
		if err != nil {
			return err
		}
		if _, err := s.db.DBTX(txCtx).Exec(txCtx, "UPDATE inventory.inventory_revaluations SET journal_entry_id=$2 WHERE id=$1", revaluation.ID, journal); err != nil {
			return err
		}
		revaluation.JournalEntryID = journal
		return s.ledger.RecordSource(txCtx, req.LegalEntityID, req.CreatedBy, revaluation.ID, "inventory.revaluation.post", revaluation)
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

func (s *InventoryService) postVariance(ctx context.Context, entity, source, actor uuid.UUID, kind, reference string, date time.Time, inventory, offset *uuid.UUID, variance decimal.Decimal) (*uuid.UUID, error) {
	if variance.IsZero() {
		return nil, nil
	}
	if inventory == nil || offset == nil {
		return nil, fmt.Errorf("inventory and variance GL accounts must be selected")
	}
	stock, err := s.ledger.Account(ctx, entity, inventory, "")
	if err != nil {
		return nil, err
	}
	other, err := s.ledger.Account(ctx, entity, offset, "")
	if err != nil {
		return nil, err
	}
	if stock == other {
		return nil, fmt.Errorf("inventory and variance accounts must differ")
	}
	debit, credit := stock, other
	if variance.IsNegative() {
		debit, credit = other, stock
	}
	journal, err := s.ledger.Post(ctx, "GENERAL", accountingdomain.PostingRequest{LegalEntityID: entity, SourceModule: "INVENTORY", SourceType: kind, SourceID: source, AccountingDate: date, Reference: reference, ActorID: actor, Lines: []accountingdomain.PostingLine{{AccountID: debit, TransactionDebit: variance.Abs()}, {AccountID: credit, TransactionCredit: variance.Abs()}}})
	if err != nil {
		return nil, err
	}
	return &journal, nil
}
