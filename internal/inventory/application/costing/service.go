package costing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExecuteRevaluationRequest struct {
	LegalEntityID  uuid.UUID
	WarehouseID    uuid.UUID
	ItemID         uuid.UUID
	NewUnitCost    decimal.Decimal
	Reason         string
	JournalEntryID *uuid.UUID
	CreatedBy      uuid.UUID
}

type CostingService struct {
	repo            ports.StockRepository
	weightedAvgCalc *WeightedAverageCalculator
	fifoCalc        *FIFOCalculator
	standardCalc    *StandardCostCalculator
}

func NewCostingService(repo ports.StockRepository) *CostingService {
	return &CostingService{
		repo:            repo,
		weightedAvgCalc: NewWeightedAverageCalculator(repo),
		fifoCalc:        NewFIFOCalculator(repo),
		standardCalc:    NewStandardCostCalculator(repo),
	}
}

func (s *CostingService) GetCalculator(method domain.CostingMethod) CostCalculator {
	switch method {
	case domain.CostingMethodFIFO:
		return s.fifoCalc
	case domain.CostingMethodStandard:
		return s.standardCalc
	case domain.CostingMethodWeightedAverage:
		fallthrough
	default:
		return s.weightedAvgCalc
	}
}

// CalculateIssueCost calculates the cost for an issue using the item's configured costing method.
func (s *CostingService) CalculateIssueCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	issueQty decimal.Decimal,
) (domain.CostCalculationResult, error) {
	itemInfo, err := s.repo.GetItemCostingDetails(ctx, itemID)
	if err != nil {
		// Fallback to weighted average if item lookup fails
		return s.weightedAvgCalc.CalculateOutboundCost(ctx, entityID, itemID, warehouseID, issueQty)
	}

	calc := s.GetCalculator(itemInfo.CostingMethod)
	return calc.CalculateOutboundCost(ctx, entityID, itemID, warehouseID, issueQty)
}

// RecordReceiptCost records the cost layer or updates the moving average upon inventory receipt.
func (s *CostingService) RecordReceiptCost(
	ctx context.Context,
	entityID, itemID, warehouseID uuid.UUID,
	qty, unitCost decimal.Decimal,
	sourceMovementID uuid.UUID,
) error {
	itemInfo, err := s.repo.GetItemCostingDetails(ctx, itemID)
	method := domain.CostingMethodWeightedAverage
	if err == nil && itemInfo != nil {
		method = itemInfo.CostingMethod
	}

	calc := s.GetCalculator(method)
	return calc.RecordInboundCost(ctx, entityID, itemID, warehouseID, qty, unitCost, sourceMovementID)
}

// GetValuationReport returns the inventory valuation report across warehouses and items.
func (s *CostingService) GetValuationReport(
	ctx context.Context,
	entityID uuid.UUID,
	warehouseID, itemID *uuid.UUID,
) ([]domain.ValuationSummaryItem, error) {
	return s.repo.GetValuationSummary(ctx, entityID, warehouseID, itemID)
}

// ListCostLayers lists discrete FIFO cost layers.
func (s *CostingService) ListCostLayers(
	ctx context.Context,
	entityID uuid.UUID,
	itemID, warehouseID *uuid.UUID,
	onlyActive *bool,
	limit, offset int32,
) ([]domain.CostLayer, error) {
	return s.repo.ListCostLayers(ctx, entityID, itemID, warehouseID, onlyActive, limit, offset)
}

// ExecuteRevaluation executes an inventory revaluation, updating the moving average cost and total value.
func (s *CostingService) ExecuteRevaluation(
	ctx context.Context,
	req ExecuteRevaluationRequest,
) (*domain.InventoryRevaluation, error) {
	if req.NewUnitCost.IsNegative() {
		return nil, domain.ErrNegativeUnitCost
	}

	bal, err := s.repo.GetBalance(ctx, domain.StockKey{
		LegalEntityID: req.LegalEntityID,
		ItemID:        req.ItemID,
		WarehouseID:   req.WarehouseID,
		BinID:         uuid.Nil,
	})
	if err != nil {
		return nil, err
	}
	if bal == nil || bal.QuantityOnHand.IsZero() {
		return nil, domain.ErrZeroQuantityRevaluation
	}

	oldUnitCost := bal.AverageCost
	oldTotalValue := bal.TotalValue
	newTotalValue := bal.QuantityOnHand.Mul(req.NewUnitCost).Round(4)
	varianceAmount := newTotalValue.Sub(oldTotalValue)

	revNum := fmt.Sprintf("REV-%s-%d", strings.ToUpper(uuid.NewString()[:8]), time.Now().UnixNano()%1000000)
	rev := &domain.InventoryRevaluation{
		LegalEntityID:     req.LegalEntityID,
		RevaluationNumber: revNum,
		WarehouseID:       req.WarehouseID,
		ItemID:            req.ItemID,
		QuantityOnHand:    bal.QuantityOnHand,
		OldUnitCost:       oldUnitCost,
		NewUnitCost:       req.NewUnitCost,
		OldTotalValue:     oldTotalValue,
		NewTotalValue:     newTotalValue,
		VarianceAmount:    varianceAmount,
		Reason:            req.Reason,
		JournalEntryID:    req.JournalEntryID,
		CreatedBy:         req.CreatedBy,
	}

	createdRev, err := s.repo.CreateInventoryRevaluation(ctx, rev)
	if err != nil {
		return nil, fmt.Errorf("failed to create revaluation record: %w", err)
	}

	// Update stock balance atomically
	_, err = s.repo.UpdateStockBalanceCostAndValue(ctx, req.LegalEntityID, req.ItemID, req.WarehouseID, req.NewUnitCost, newTotalValue)
	if err != nil {
		return nil, fmt.Errorf("failed to update stock balance: %w", err)
	}

	return createdRev, nil
}

// ListRevaluations lists past inventory revaluations.
func (s *CostingService) ListRevaluations(
	ctx context.Context,
	entityID uuid.UUID,
	warehouseID, itemID *uuid.UUID,
	limit, offset int32,
) ([]domain.InventoryRevaluation, error) {
	return s.repo.ListInventoryRevaluations(ctx, entityID, warehouseID, itemID, limit, offset)
}
