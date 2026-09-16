package postgres

import (
	"context"
	"errors"

	"erp/internal/inventory/domain"
	"erp/internal/inventory/ports"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type inventoryRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewInventoryRepository(pool *pgxpool.Pool) ports.StockRepository {
	return &inventoryRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

// 1. Canonical StockRepository Contract (Section 25)
func (r *inventoryRepository) LockBalance(
	ctx context.Context,
	legalEntityID uuid.UUID,
	itemID uuid.UUID,
	warehouseID uuid.UUID,
	binID uuid.UUID,
) (*domain.StockBalance, error) {
	if binID == uuid.Nil {
		defBin, err := r.GetDefaultBin(ctx, warehouseID)
		if err == nil && defBin != nil {
			binID = defBin.ID
		}
	}
	row, err := r.queries.LockBalance(ctx, sqlc.LockBalanceParams{
		LegalEntityID: toPgUUID(legalEntityID),
		ItemID:        toPgUUID(itemID),
		WarehouseID:   toPgUUID(warehouseID),
		BinID:         toPgUUID(binID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.StockBalance{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		ItemID:            toUUID(row.ItemID),
		WarehouseID:       toUUID(row.WarehouseID),
		BinID:             toUUID(row.BinID),
		QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
		QuantityReserved:  fromPgNumeric(row.QuantityReserved),
		QuantityAvailable: fromPgNumeric(row.QuantityAvailable),
		AverageCost:       fromPgNumeric(row.AverageCost),
		TotalValue:        fromPgNumeric(row.TotalValue),
		UpdatedAt:         row.UpdatedAt.Time,
		Version:           row.Version,
	}, nil
}

func (r *inventoryRepository) InsertMovement(
	ctx context.Context,
	m *domain.StockMovement,
) (*domain.StockMovement, error) {
	row, err := r.queries.InsertStockMovement(ctx, sqlc.InsertStockMovementParams{
		LegalEntityID:   toPgUUID(m.LegalEntityID),
		MovementNumber:  m.MovementNumber,
		MovementType:    string(m.MovementType),
		ItemID:          toPgUUID(m.ItemID),
		SkuID:           toPgUUIDPtr(m.SKUID),
		WarehouseID:     toPgUUID(m.WarehouseID),
		BinID:           toPgUUID(m.BinID),
		LotNumber:       toPgTextPtr(m.LotNumber),
		SerialNumber:    toPgTextPtr(m.SerialNumber),
		Quantity:        toPgNumeric(m.Quantity),
		UomID:           toPgUUID(m.UOMID),
		UnitCost:        toPgNumeric(m.UnitCost),
		TotalCost:       toPgNumeric(m.TotalCost),
		SourceDocType:   toPgTextPtr(m.SourceDocType),
		SourceDocID:     toPgUUIDPtr(m.SourceDocID),
		SourceDocLineID: toPgUUIDPtr(m.SourceDocLineID),
		OccurredAt:      toPgTimestamptzVal(m.OccurredAt),
		CreatedBy:       toPgUUID(m.CreatedBy),
	})
	if err != nil {
		return nil, err
	}
	m.ID = toUUID(row.ID)
	m.CreatedAt = row.CreatedAt.Time
	m.OccurredAt = row.OccurredAt.Time
	return m, nil
}

func (r *inventoryRepository) GetItemBaseUOM(ctx context.Context, itemID uuid.UUID) (uuid.UUID, error) {
	uomID, err := r.queries.GetItemBaseUOM(ctx, toPgUUID(itemID))
	if err != nil {
		return uuid.Nil, err
	}
	return toUUID(uomID), nil
}

func (r *inventoryRepository) ApplyBalanceDelta(
	ctx context.Context,
	key domain.StockKey,
	quantityDelta decimal.Decimal,
	costDelta decimal.Decimal,
) (*domain.StockBalance, error) {
	current, err := r.LockBalance(ctx, key.LegalEntityID, key.ItemID, key.WarehouseID, key.BinID)
	if err != nil {
		return nil, err
	}

	var newOnHand, newReserved, newVal, avgCost decimal.Decimal
	if current == nil {
		newOnHand = quantityDelta
		newReserved = decimal.Zero
		newVal = costDelta
		if newOnHand.IsPositive() {
			avgCost = newVal.Div(newOnHand)
		} else {
			avgCost = decimal.Zero
		}
	} else {
		newOnHand = current.QuantityOnHand.Add(quantityDelta)
		newReserved = current.QuantityReserved
		newVal = current.TotalValue.Add(costDelta)
		if newOnHand.IsPositive() {
			avgCost = newVal.Div(newOnHand)
		} else if newOnHand.IsZero() {
			newVal = decimal.Zero
			avgCost = decimal.Zero
		} else {
			avgCost = current.AverageCost
		}
	}

	row, err := r.queries.UpsertBalance(ctx, sqlc.UpsertBalanceParams{
		LegalEntityID:    toPgUUID(key.LegalEntityID),
		ItemID:           toPgUUID(key.ItemID),
		WarehouseID:      toPgUUID(key.WarehouseID),
		BinID:            toPgUUID(key.BinID),
		QuantityOnHand:   toPgNumeric(newOnHand),
		QuantityReserved: toPgNumeric(newReserved),
		AverageCost:      toPgNumeric(avgCost),
		TotalValue:       toPgNumeric(newVal),
	})
	if err != nil {
		return nil, err
	}

	return &domain.StockBalance{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		ItemID:            toUUID(row.ItemID),
		WarehouseID:       toUUID(row.WarehouseID),
		BinID:             toUUID(row.BinID),
		QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
		QuantityReserved:  fromPgNumeric(row.QuantityReserved),
		QuantityAvailable: fromPgNumeric(row.QuantityAvailable),
		AverageCost:       fromPgNumeric(row.AverageCost),
		TotalValue:        fromPgNumeric(row.TotalValue),
		UpdatedAt:         row.UpdatedAt.Time,
		Version:           row.Version,
	}, nil
}

func (r *inventoryRepository) Reserve(
	ctx context.Context,
	res *domain.StockReservation,
) (*domain.StockReservation, error) {
	binID := uuid.Nil
	if res.BinID != nil {
		binID = *res.BinID
	} else {
		defBin, err := r.GetDefaultBin(ctx, res.WarehouseID)
		if err != nil {
			return nil, err
		}
		binID = defBin.ID
		res.BinID = &binID
	}

	bal, err := r.LockBalance(ctx, res.LegalEntityID, res.ItemID, res.WarehouseID, binID)
	if err != nil {
		return nil, err
	}
	if bal == nil || bal.QuantityAvailable.LessThan(res.Quantity) {
		return nil, domain.ErrInsufficientStock
	}

	// Update balance reservation
	newReserved := bal.QuantityReserved.Add(res.Quantity)
	_, err = r.queries.UpsertBalance(ctx, sqlc.UpsertBalanceParams{
		LegalEntityID:    toPgUUID(res.LegalEntityID),
		ItemID:           toPgUUID(res.ItemID),
		WarehouseID:      toPgUUID(res.WarehouseID),
		BinID:            toPgUUID(binID),
		QuantityOnHand:   toPgNumeric(bal.QuantityOnHand),
		QuantityReserved: toPgNumeric(newReserved),
		AverageCost:      toPgNumeric(bal.AverageCost),
		TotalValue:       toPgNumeric(bal.TotalValue),
	})
	if err != nil {
		return nil, err
	}

	// Insert reservation row
	row, err := r.queries.CreateStockReservation(ctx, sqlc.CreateStockReservationParams{
		LegalEntityID:     toPgUUID(res.LegalEntityID),
		ItemID:            toPgUUID(res.ItemID),
		WarehouseID:       toPgUUID(res.WarehouseID),
		BinID:             toPgUUIDPtr(res.BinID),
		Quantity:          toPgNumeric(res.Quantity),
		ReservationStatus: string(domain.ReservationStatusActive),
		DemandSourceType:  res.DemandSourceType,
		DemandSourceID:    toPgUUID(res.DemandSourceID),
		ExpiresAt:         toPgTimestamptz(res.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}

	res.ID = toUUID(row.ID)
	res.ReservationStatus = domain.ReservationStatusActive
	res.CreatedAt = row.CreatedAt.Time
	res.UpdatedAt = row.UpdatedAt.Time
	return res, nil
}

func (r *inventoryRepository) ReleaseReservation(
	ctx context.Context,
	reservationID uuid.UUID,
) error {
	row, err := r.queries.GetReservationByID(ctx, toPgUUID(reservationID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrReservationNotFound
		}
		return err
	}
	if row.ReservationStatus != string(domain.ReservationStatusActive) {
		return domain.ErrReservationNotFound
	}

	// Decrement balance reservation
	binID := toUUIDPtr(row.BinID)
	if binID != nil {
		bal, err := r.LockBalance(ctx, toUUID(row.LegalEntityID), toUUID(row.ItemID), toUUID(row.WarehouseID), *binID)
		if err == nil && bal != nil {
			newReserved := bal.QuantityReserved.Sub(fromPgNumeric(row.Quantity))
			if newReserved.IsNegative() {
				newReserved = decimal.Zero
			}
			_, _ = r.queries.UpsertBalance(ctx, sqlc.UpsertBalanceParams{
				LegalEntityID:    toPgUUID(bal.LegalEntityID),
				ItemID:           toPgUUID(bal.ItemID),
				WarehouseID:      toPgUUID(bal.WarehouseID),
				BinID:            toPgUUID(bal.BinID),
				QuantityOnHand:   toPgNumeric(bal.QuantityOnHand),
				QuantityReserved: toPgNumeric(newReserved),
				AverageCost:      toPgNumeric(bal.AverageCost),
				TotalValue:       toPgNumeric(bal.TotalValue),
			})
		}
	}

	_, err = r.queries.ReleaseStockReservation(ctx, toPgUUID(reservationID))
	return err
}

// 2. Balances & Movements Queries
func (r *inventoryRepository) GetBalance(ctx context.Context, key domain.StockKey) (*domain.StockBalance, error) {
	binID := key.BinID
	if binID == uuid.Nil {
		defBin, err := r.GetDefaultBin(ctx, key.WarehouseID)
		if err == nil && defBin != nil {
			binID = defBin.ID
		}
	}
	row, err := r.queries.GetBalance(ctx, sqlc.GetBalanceParams{
		LegalEntityID: toPgUUID(key.LegalEntityID),
		ItemID:        toPgUUID(key.ItemID),
		WarehouseID:   toPgUUID(key.WarehouseID),
		BinID:         toPgUUID(binID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.StockBalance{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		ItemID:            toUUID(row.ItemID),
		WarehouseID:       toUUID(row.WarehouseID),
		BinID:             toUUID(row.BinID),
		QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
		QuantityReserved:  fromPgNumeric(row.QuantityReserved),
		QuantityAvailable: fromPgNumeric(row.QuantityAvailable),
		AverageCost:       fromPgNumeric(row.AverageCost),
		TotalValue:        fromPgNumeric(row.TotalValue),
		UpdatedAt:         row.UpdatedAt.Time,
		Version:           row.Version,
		ItemCode:          row.ItemCode,
		ItemNameEn:        row.ItemNameEn,
		ItemNameAr:        row.ItemNameAr,
		WarehouseCode:     row.WarehouseCode,
		WarehouseNameAr:   row.WarehouseNameAr,
		BinCode:           row.BinCode,
	}, nil
}

func (r *inventoryRepository) ListBalances(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, limit, offset int32) ([]domain.StockBalance, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries.ListBalancesByEntity(ctx, sqlc.ListBalancesByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		ItemID:        toPgUUIDPtr(itemID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.StockBalance, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.StockBalance{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			ItemID:            toUUID(row.ItemID),
			WarehouseID:       toUUID(row.WarehouseID),
			BinID:             toUUID(row.BinID),
			QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
			QuantityReserved:  fromPgNumeric(row.QuantityReserved),
			QuantityAvailable: fromPgNumeric(row.QuantityAvailable),
			AverageCost:       fromPgNumeric(row.AverageCost),
			TotalValue:        fromPgNumeric(row.TotalValue),
			UpdatedAt:         row.UpdatedAt.Time,
			Version:           row.Version,
			ItemCode:          row.ItemCode,
			ItemNameEn:        row.ItemNameEn,
			ItemNameAr:        row.ItemNameAr,
			BaseUOMCode:       row.BaseUomCode,
			WarehouseCode:     row.WarehouseCode,
			WarehouseNameAr:   row.WarehouseNameAr,
			BinCode:           row.BinCode,
		})
	}
	return res, nil
}

func (r *inventoryRepository) ListMovements(ctx context.Context, legalEntityID uuid.UUID, itemID, warehouseID *uuid.UUID, movType *string, limit, offset int32) ([]domain.StockMovement, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries.ListMovementsByEntity(ctx, sqlc.ListMovementsByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		ItemID:        toPgUUIDPtr(itemID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		MovementType:  toPgTextPtr(movType),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.StockMovement, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.StockMovement{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			MovementNumber:  row.MovementNumber,
			MovementType:    domain.MovementType(row.MovementType),
			ItemID:          toUUID(row.ItemID),
			SKUID:           toUUIDPtr(row.SkuID),
			WarehouseID:     toUUID(row.WarehouseID),
			BinID:           toUUID(row.BinID),
			LotNumber:       fromPgTextPtr(row.LotNumber),
			SerialNumber:    fromPgTextPtr(row.SerialNumber),
			Quantity:        fromPgNumeric(row.Quantity),
			UOMID:           toUUID(row.UomID),
			UnitCost:        fromPgNumeric(row.UnitCost),
			TotalCost:       fromPgNumeric(row.TotalCost),
			SourceDocType:   fromPgTextPtr(row.SourceDocType),
			SourceDocID:     toUUIDPtr(row.SourceDocID),
			OccurredAt:      row.OccurredAt.Time,
			CreatedAt:       row.CreatedAt.Time,
			ItemCode:        row.ItemCode,
			ItemNameEn:      row.ItemNameEn,
			ItemNameAr:      row.ItemNameAr,
			WarehouseCode:   row.WarehouseCode,
			WarehouseNameAr: row.WarehouseNameAr,
			BinCode:         row.BinCode,
			UOMCode:         row.UomCode,
		})
	}
	return res, nil
}

func (r *inventoryRepository) GetMovementByID(ctx context.Context, id uuid.UUID) (*domain.StockMovement, error) {
	row, err := r.queries.GetMovementByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.StockMovement{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		MovementNumber:  row.MovementNumber,
		MovementType:    domain.MovementType(row.MovementType),
		ItemID:          toUUID(row.ItemID),
		SKUID:           toUUIDPtr(row.SkuID),
		WarehouseID:     toUUID(row.WarehouseID),
		BinID:           toUUID(row.BinID),
		LotNumber:       fromPgTextPtr(row.LotNumber),
		SerialNumber:    fromPgTextPtr(row.SerialNumber),
		Quantity:        fromPgNumeric(row.Quantity),
		UOMID:           toUUID(row.UomID),
		UnitCost:        fromPgNumeric(row.UnitCost),
		TotalCost:       fromPgNumeric(row.TotalCost),
		SourceDocType:   fromPgTextPtr(row.SourceDocType),
		SourceDocID:     toUUIDPtr(row.SourceDocID),
		SourceDocLineID: toUUIDPtr(row.SourceDocLineID),
		OccurredAt:      row.OccurredAt.Time,
		CreatedBy:       toUUID(row.CreatedBy),
		CreatedAt:       row.CreatedAt.Time,
		ItemCode:        row.ItemCode,
		ItemNameEn:      row.ItemNameEn,
		ItemNameAr:      row.ItemNameAr,
		WarehouseCode:   row.WarehouseCode,
		WarehouseNameAr: row.WarehouseNameAr,
		BinCode:         row.BinCode,
		UOMCode:         row.UomCode,
	}, nil
}

// 3. Warehouses & Bins
func (r *inventoryRepository) ListWarehouses(ctx context.Context, legalEntityID uuid.UUID) ([]domain.Warehouse, error) {
	rows, err := r.queries.ListWarehousesByEntity(ctx, toPgUUID(legalEntityID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.Warehouse, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.Warehouse{
			ID:            toUUID(row.ID),
			LegalEntityID: toUUID(row.LegalEntityID),
			Code:          row.Code,
			NameEn:        row.NameEn,
			NameAr:        row.NameAr,
			WarehouseType: domain.WarehouseType(row.WarehouseType),
			IsActive:      row.IsActive,
			AddressLine1:  fromPgText(row.AddressLine1),
			City:          fromPgText(row.City),
			PostalCode:    fromPgText(row.PostalCode),
			CountryCode:   fromPgText(row.CountryCode),
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		})
	}
	return res, nil
}

func (r *inventoryRepository) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*domain.Warehouse, error) {
	row, err := r.queries.GetWarehouseByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWarehouseNotFound
		}
		return nil, err
	}
	return &domain.Warehouse{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		WarehouseType: domain.WarehouseType(row.WarehouseType),
		IsActive:      row.IsActive,
		AddressLine1:  fromPgText(row.AddressLine1),
		City:          fromPgText(row.City),
		PostalCode:    fromPgText(row.PostalCode),
		CountryCode:   fromPgText(row.CountryCode),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *inventoryRepository) GetWarehouseByCode(ctx context.Context, legalEntityID uuid.UUID, code string) (*domain.Warehouse, error) {
	row, err := r.queries.GetWarehouseByCode(ctx, sqlc.GetWarehouseByCodeParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Code:          code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWarehouseNotFound
		}
		return nil, err
	}
	return &domain.Warehouse{
		ID:            toUUID(row.ID),
		LegalEntityID: toUUID(row.LegalEntityID),
		Code:          row.Code,
		NameEn:        row.NameEn,
		NameAr:        row.NameAr,
		WarehouseType: domain.WarehouseType(row.WarehouseType),
		IsActive:      row.IsActive,
		AddressLine1:  fromPgText(row.AddressLine1),
		City:          fromPgText(row.City),
		PostalCode:    fromPgText(row.PostalCode),
		CountryCode:   fromPgText(row.CountryCode),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *inventoryRepository) CreateWarehouse(ctx context.Context, wh *domain.Warehouse) (*domain.Warehouse, error) {
	row, err := r.queries.CreateWarehouse(ctx, sqlc.CreateWarehouseParams{
		LegalEntityID: toPgUUID(wh.LegalEntityID),
		Code:          wh.Code,
		NameEn:        wh.NameEn,
		NameAr:        wh.NameAr,
		WarehouseType: string(wh.WarehouseType),
		IsActive:      wh.IsActive,
		AddressLine1:  toPgText(wh.AddressLine1),
		City:          toPgText(wh.City),
		PostalCode:    toPgText(wh.PostalCode),
		CountryCode:   toPgText(wh.CountryCode),
	})
	if err != nil {
		return nil, err
	}
	wh.ID = toUUID(row.ID)
	wh.CreatedAt = row.CreatedAt.Time
	wh.UpdatedAt = row.UpdatedAt.Time
	return wh, nil
}

func (r *inventoryRepository) UpdateWarehouse(ctx context.Context, wh *domain.Warehouse) (*domain.Warehouse, error) {
	row, err := r.queries.UpdateWarehouse(ctx, sqlc.UpdateWarehouseParams{
		ID:            toPgUUID(wh.ID),
		NameEn:        wh.NameEn,
		NameAr:        wh.NameAr,
		WarehouseType: string(wh.WarehouseType),
		IsActive:      wh.IsActive,
		AddressLine1:  toPgText(wh.AddressLine1),
		City:          toPgText(wh.City),
		PostalCode:    toPgText(wh.PostalCode),
		CountryCode:   toPgText(wh.CountryCode),
	})
	if err != nil {
		return nil, err
	}
	wh.UpdatedAt = row.UpdatedAt.Time
	return wh, nil
}

func (r *inventoryRepository) ListBins(ctx context.Context, warehouseID uuid.UUID) ([]domain.BinLocation, error) {
	rows, err := r.queries.ListBinsByWarehouse(ctx, toPgUUID(warehouseID))
	if err != nil {
		return nil, err
	}
	res := make([]domain.BinLocation, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.BinLocation{
			ID:          toUUID(row.ID),
			WarehouseID: toUUID(row.WarehouseID),
			Code:        row.Code,
			Aisle:       fromPgText(row.Aisle),
			Rack:        fromPgText(row.Rack),
			Shelf:       fromPgText(row.Shelf),
			Bin:         fromPgText(row.Bin),
			IsDefault:   row.IsDefault,
			IsActive:    row.IsActive,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return res, nil
}

func (r *inventoryRepository) GetBinByID(ctx context.Context, id uuid.UUID) (*domain.BinLocation, error) {
	row, err := r.queries.GetBinByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBinNotFound
		}
		return nil, err
	}
	return &domain.BinLocation{
		ID:          toUUID(row.ID),
		WarehouseID: toUUID(row.WarehouseID),
		Code:        row.Code,
		Aisle:       fromPgText(row.Aisle),
		Rack:        fromPgText(row.Rack),
		Shelf:       fromPgText(row.Shelf),
		Bin:         fromPgText(row.Bin),
		IsDefault:   row.IsDefault,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *inventoryRepository) GetDefaultBin(ctx context.Context, warehouseID uuid.UUID) (*domain.BinLocation, error) {
	row, err := r.queries.GetDefaultBinByWarehouse(ctx, toPgUUID(warehouseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBinNotFound
		}
		return nil, err
	}
	return &domain.BinLocation{
		ID:          toUUID(row.ID),
		WarehouseID: toUUID(row.WarehouseID),
		Code:        row.Code,
		Aisle:       fromPgText(row.Aisle),
		Rack:        fromPgText(row.Rack),
		Shelf:       fromPgText(row.Shelf),
		Bin:         fromPgText(row.Bin),
		IsDefault:   row.IsDefault,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}, nil
}

func (r *inventoryRepository) CreateBin(ctx context.Context, bin *domain.BinLocation) (*domain.BinLocation, error) {
	row, err := r.queries.CreateBin(ctx, sqlc.CreateBinParams{
		WarehouseID: toPgUUID(bin.WarehouseID),
		Code:        bin.Code,
		Aisle:       toPgText(bin.Aisle),
		Rack:        toPgText(bin.Rack),
		Shelf:       toPgText(bin.Shelf),
		Bin:         toPgText(bin.Bin),
		IsDefault:   bin.IsDefault,
		IsActive:    bin.IsActive,
	})
	if err != nil {
		return nil, err
	}
	bin.ID = toUUID(row.ID)
	bin.CreatedAt = row.CreatedAt.Time
	bin.UpdatedAt = row.UpdatedAt.Time
	return bin, nil
}

// 4. Transfers
func (r *inventoryRepository) CreateTransferOrder(ctx context.Context, order *domain.TransferOrder, lines []domain.TransferOrderLine) (*domain.TransferOrder, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.CreateTransferOrder(ctx, sqlc.CreateTransferOrderParams{
		LegalEntityID:        toPgUUID(order.LegalEntityID),
		OrderNumber:          order.OrderNumber,
		FromWarehouseID:      toPgUUID(order.FromWarehouseID),
		ToWarehouseID:        toPgUUID(order.ToWarehouseID),
		Status:               string(order.Status),
		ShippingDate:         toPgTimestamptz(order.ShippingDate),
		ExpectedDeliveryDate: toPgTimestamptz(order.ExpectedDeliveryDate),
		Notes:                toPgText(order.Notes),
		CreatedBy:            toPgUUID(order.CreatedBy),
	})
	if err != nil {
		return nil, err
	}

	order.ID = toUUID(row.ID)
	order.CreatedAt = row.CreatedAt.Time
	order.UpdatedAt = row.UpdatedAt.Time

	createdLines := make([]domain.TransferOrderLine, 0, len(lines))
	for _, l := range lines {
		lRow, lErr := qtx.CreateTransferOrderLine(ctx, sqlc.CreateTransferOrderLineParams{
			TransferOrderID:    toPgUUID(order.ID),
			ItemID:             toPgUUID(l.ItemID),
			SkuID:              toPgUUIDPtr(l.SKUID),
			RequestedQuantity:  toPgNumeric(l.RequestedQuantity),
			DispatchedQuantity: toPgNumeric(l.DispatchedQuantity),
			ReceivedQuantity:   toPgNumeric(l.ReceivedQuantity),
			UomID:              toPgUUID(l.UOMID),
			Notes:              toPgText(l.Notes),
		})
		if lErr != nil {
			return nil, lErr
		}
		l.ID = toUUID(lRow.ID)
		l.TransferOrderID = order.ID
		createdLines = append(createdLines, l)
	}
	order.Lines = createdLines

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return order, nil
}

func (r *inventoryRepository) GetTransferOrder(ctx context.Context, id uuid.UUID) (*domain.TransferOrder, error) {
	row, err := r.queries.GetTransferOrderByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransferOrderNotFound
		}
		return nil, err
	}

	lineRows, err := r.queries.ListTransferOrderLines(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	lines := make([]domain.TransferOrderLine, 0, len(lineRows))
	for _, lr := range lineRows {
		lines = append(lines, domain.TransferOrderLine{
			ID:                 toUUID(lr.ID),
			TransferOrderID:    toUUID(lr.TransferOrderID),
			ItemID:             toUUID(lr.ItemID),
			SKUID:              toUUIDPtr(lr.SkuID),
			RequestedQuantity:  fromPgNumeric(lr.RequestedQuantity),
			DispatchedQuantity: fromPgNumeric(lr.DispatchedQuantity),
			ReceivedQuantity:   fromPgNumeric(lr.ReceivedQuantity),
			UOMID:              toUUID(lr.UomID),
			Notes:              fromPgText(lr.Notes),
			ItemCode:           lr.ItemCode,
			ItemNameEn:         lr.ItemNameEn,
			ItemNameAr:         lr.ItemNameAr,
			UOMCode:            lr.UomCode,
		})
	}

	return &domain.TransferOrder{
		ID:                   toUUID(row.ID),
		LegalEntityID:        toUUID(row.LegalEntityID),
		OrderNumber:          row.OrderNumber,
		FromWarehouseID:      toUUID(row.FromWarehouseID),
		ToWarehouseID:        toUUID(row.ToWarehouseID),
		Status:               domain.TransferStatus(row.Status),
		ShippingDate:         fromPgTimestamptz(row.ShippingDate),
		ExpectedDeliveryDate: fromPgTimestamptz(row.ExpectedDeliveryDate),
		ReceivedDate:         fromPgTimestamptz(row.ReceivedDate),
		Notes:                fromPgText(row.Notes),
		CreatedBy:            toUUID(row.CreatedBy),
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		FromWarehouseCode:    row.FromWarehouseCode,
		FromWarehouseNameAr:  row.FromWarehouseNameAr,
		ToWarehouseCode:      row.ToWarehouseCode,
		ToWarehouseNameAr:    row.ToWarehouseNameAr,
		Lines:                lines,
	}, nil
}

func (r *inventoryRepository) ListTransferOrders(ctx context.Context, legalEntityID uuid.UUID, status *string, limit, offset int32) ([]domain.TransferOrder, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries.ListTransferOrdersByEntity(ctx, sqlc.ListTransferOrdersByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		Status:        toPgTextPtr(status),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.TransferOrder, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.TransferOrder{
			ID:                   toUUID(row.ID),
			LegalEntityID:        toUUID(row.LegalEntityID),
			OrderNumber:          row.OrderNumber,
			FromWarehouseID:      toUUID(row.FromWarehouseID),
			ToWarehouseID:        toUUID(row.ToWarehouseID),
			Status:               domain.TransferStatus(row.Status),
			ShippingDate:         fromPgTimestamptz(row.ShippingDate),
			ExpectedDeliveryDate: fromPgTimestamptz(row.ExpectedDeliveryDate),
			ReceivedDate:         fromPgTimestamptz(row.ReceivedDate),
			Notes:                fromPgText(row.Notes),
			CreatedAt:            row.CreatedAt.Time,
			FromWarehouseCode:    row.FromWarehouseCode,
			FromWarehouseNameAr:  row.FromWarehouseNameAr,
			ToWarehouseCode:      row.ToWarehouseCode,
			ToWarehouseNameAr:    row.ToWarehouseNameAr,
		})
	}
	return res, nil
}

func (r *inventoryRepository) UpdateTransferOrderStatus(ctx context.Context, id uuid.UUID, status domain.TransferStatus) (*domain.TransferOrder, error) {
	row, err := r.queries.UpdateTransferOrderStatus(ctx, sqlc.UpdateTransferOrderStatusParams{
		ID:     toPgUUID(id),
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return r.GetTransferOrder(ctx, toUUID(row.ID))
}

// 5. Adjustments
func (r *inventoryRepository) CreateStockAdjustment(ctx context.Context, adj *domain.StockAdjustment, lines []domain.StockAdjustmentLine) (*domain.StockAdjustment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.CreateStockAdjustment(ctx, sqlc.CreateStockAdjustmentParams{
		LegalEntityID:    toPgUUID(adj.LegalEntityID),
		AdjustmentNumber: adj.AdjustmentNumber,
		WarehouseID:      toPgUUID(adj.WarehouseID),
		AdjustmentType:   string(adj.AdjustmentType),
		ReasonCode:       adj.ReasonCode,
		Status:           string(adj.Status),
		AdjustmentDate:   toPgDate(adj.AdjustmentDate),
		Notes:            toPgText(adj.Notes),
		CreatedBy:        toPgUUID(adj.CreatedBy),
	})
	if err != nil {
		return nil, err
	}

	adj.ID = toUUID(row.ID)
	adj.CreatedAt = row.CreatedAt.Time
	adj.UpdatedAt = row.UpdatedAt.Time

	createdLines := make([]domain.StockAdjustmentLine, 0, len(lines))
	for _, l := range lines {
		lRow, lErr := qtx.CreateStockAdjustmentLine(ctx, sqlc.CreateStockAdjustmentLineParams{
			AdjustmentID: toPgUUID(adj.ID),
			ItemID:       toPgUUID(l.ItemID),
			BinID:        toPgUUIDPtr(l.BinID),
			LotNumber:    toPgTextPtr(l.LotNumber),
			SerialNumber: toPgTextPtr(l.SerialNumber),
			Quantity:     toPgNumeric(l.Quantity),
			UnitCost:     toPgNumeric(l.UnitCost),
			TotalCost:    toPgNumeric(l.TotalCost),
			Reason:       toPgText(l.Reason),
		})
		if lErr != nil {
			return nil, lErr
		}
		l.ID = toUUID(lRow.ID)
		l.AdjustmentID = adj.ID
		createdLines = append(createdLines, l)
	}
	adj.Lines = createdLines

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return adj, nil
}

func (r *inventoryRepository) GetStockAdjustment(ctx context.Context, id uuid.UUID) (*domain.StockAdjustment, error) {
	row, err := r.queries.GetStockAdjustmentByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStockAdjustmentNotFound
		}
		return nil, err
	}

	lineRows, err := r.queries.ListStockAdjustmentLines(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	lines := make([]domain.StockAdjustmentLine, 0, len(lineRows))
	for _, lr := range lineRows {
		lines = append(lines, domain.StockAdjustmentLine{
			ID:           toUUID(lr.ID),
			AdjustmentID: toUUID(lr.AdjustmentID),
			ItemID:       toUUID(lr.ItemID),
			UOMID:        toUUID(lr.UomID),
			BinID:        toUUIDPtr(lr.BinID),
			LotNumber:    fromPgTextPtr(lr.LotNumber),
			SerialNumber: fromPgTextPtr(lr.SerialNumber),
			Quantity:     fromPgNumeric(lr.Quantity),
			UnitCost:     fromPgNumeric(lr.UnitCost),
			TotalCost:    fromPgNumeric(lr.TotalCost),
			Reason:       fromPgText(lr.Reason),
			ItemCode:     lr.ItemCode,
			ItemNameEn:   lr.ItemNameEn,
			ItemNameAr:   lr.ItemNameAr,
			BinCode:      fromPgText(lr.BinCode),
		})
	}

	return &domain.StockAdjustment{
		ID:               toUUID(row.ID),
		LegalEntityID:    toUUID(row.LegalEntityID),
		AdjustmentNumber: row.AdjustmentNumber,
		WarehouseID:      toUUID(row.WarehouseID),
		AdjustmentType:   domain.AdjustmentType(row.AdjustmentType),
		ReasonCode:       row.ReasonCode,
		Status:           domain.AdjustmentStatus(row.Status),
		AdjustmentDate:   row.AdjustmentDate.Time,
		PostedAt:         fromPgTimestamptz(row.PostedAt),
		PostedBy:         toUUIDPtr(row.PostedBy),
		Notes:            fromPgText(row.Notes),
		CreatedBy:        toUUID(row.CreatedBy),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
		WarehouseCode:    row.WarehouseCode,
		WarehouseNameAr:  row.WarehouseNameAr,
		Lines:            lines,
	}, nil
}

func (r *inventoryRepository) ListStockAdjustments(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockAdjustment, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries.ListStockAdjustmentsByEntity(ctx, sqlc.ListStockAdjustmentsByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		Status:        toPgTextPtr(status),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.StockAdjustment, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.StockAdjustment{
			ID:               toUUID(row.ID),
			LegalEntityID:    toUUID(row.LegalEntityID),
			AdjustmentNumber: row.AdjustmentNumber,
			WarehouseID:      toUUID(row.WarehouseID),
			AdjustmentType:   domain.AdjustmentType(row.AdjustmentType),
			ReasonCode:       row.ReasonCode,
			Status:           domain.AdjustmentStatus(row.Status),
			AdjustmentDate:   row.AdjustmentDate.Time,
			PostedAt:         fromPgTimestamptz(row.PostedAt),
			PostedBy:         toUUIDPtr(row.PostedBy),
			Notes:            fromPgText(row.Notes),
			CreatedAt:        row.CreatedAt.Time,
			WarehouseCode:    row.WarehouseCode,
			WarehouseNameAr:  row.WarehouseNameAr,
		})
	}
	return res, nil
}

func (r *inventoryRepository) UpdateStockAdjustmentStatus(ctx context.Context, id uuid.UUID, status domain.AdjustmentStatus, postedBy *uuid.UUID) (*domain.StockAdjustment, error) {
	row, err := r.queries.UpdateStockAdjustmentStatus(ctx, sqlc.UpdateStockAdjustmentStatusParams{
		ID:       toPgUUID(id),
		Status:   string(status),
		PostedBy: toPgUUIDPtr(postedBy),
	})
	if err != nil {
		return nil, err
	}
	return r.GetStockAdjustment(ctx, toUUID(row.ID))
}

// 6. Counts
func (r *inventoryRepository) CreateStockCount(ctx context.Context, count *domain.StockCount, lines []domain.StockCountLine) (*domain.StockCount, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.CreateStockCount(ctx, sqlc.CreateStockCountParams{
		LegalEntityID: toPgUUID(count.LegalEntityID),
		CountNumber:   count.CountNumber,
		WarehouseID:   toPgUUID(count.WarehouseID),
		CountType:     string(count.CountType),
		Status:        string(count.Status),
		CountDate:     toPgDate(count.CountDate),
	})
	if err != nil {
		return nil, err
	}

	count.ID = toUUID(row.ID)
	count.CreatedAt = row.CreatedAt.Time
	count.UpdatedAt = row.UpdatedAt.Time

	createdLines := make([]domain.StockCountLine, 0, len(lines))
	for _, l := range lines {
		lRow, lErr := qtx.CreateStockCountLine(ctx, sqlc.CreateStockCountLineParams{
			CountID:         toPgUUID(count.ID),
			ItemID:          toPgUUID(l.ItemID),
			BinID:           toPgUUIDPtr(l.BinID),
			SystemQuantity:  toPgNumeric(l.SystemQuantity),
			CountedQuantity: toPgNumeric(l.CountedQuantity),
			Notes:           toPgText(l.Notes),
		})
		if lErr != nil {
			return nil, lErr
		}
		l.ID = toUUID(lRow.ID)
		l.CountID = count.ID
		l.Variance = fromPgNumeric(lRow.Variance)
		createdLines = append(createdLines, l)
	}
	count.Lines = createdLines

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return count, nil
}

func (r *inventoryRepository) GetStockCount(ctx context.Context, id uuid.UUID) (*domain.StockCount, error) {
	row, err := r.queries.GetStockCountByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStockCountNotFound
		}
		return nil, err
	}

	lineRows, err := r.queries.ListStockCountLines(ctx, toPgUUID(id))
	if err != nil {
		return nil, err
	}
	lines := make([]domain.StockCountLine, 0, len(lineRows))
	for _, lr := range lineRows {
		lines = append(lines, domain.StockCountLine{
			ID:              toUUID(lr.ID),
			CountID:         toUUID(lr.CountID),
			ItemID:          toUUID(lr.ItemID),
			UOMID:           toUUID(lr.UomID),
			BinID:           toUUIDPtr(lr.BinID),
			SystemQuantity:  fromPgNumeric(lr.SystemQuantity),
			CountedQuantity: fromPgNumeric(lr.CountedQuantity),
			Variance:        fromPgNumeric(lr.Variance),
			Notes:           fromPgText(lr.Notes),
			ItemCode:        lr.ItemCode,
			ItemNameEn:      lr.ItemNameEn,
			ItemNameAr:      lr.ItemNameAr,
			BinCode:         fromPgText(lr.BinCode),
		})
	}

	return &domain.StockCount{
		ID:              toUUID(row.ID),
		LegalEntityID:   toUUID(row.LegalEntityID),
		CountNumber:     row.CountNumber,
		WarehouseID:     toUUID(row.WarehouseID),
		CountType:       domain.CountType(row.CountType),
		Status:          domain.CountStatus(row.Status),
		CountDate:       row.CountDate.Time,
		ReconciledAt:    fromPgTimestamptz(row.ReconciledAt),
		ReconciledBy:    toUUIDPtr(row.ReconciledBy),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		WarehouseCode:   row.WarehouseCode,
		WarehouseNameAr: row.WarehouseNameAr,
		Lines:           lines,
	}, nil
}

func (r *inventoryRepository) ListStockCounts(ctx context.Context, legalEntityID uuid.UUID, warehouseID *uuid.UUID, status *string, limit, offset int32) ([]domain.StockCount, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries.ListStockCountsByEntity(ctx, sqlc.ListStockCountsByEntityParams{
		LegalEntityID: toPgUUID(legalEntityID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		Status:        toPgTextPtr(status),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	res := make([]domain.StockCount, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.StockCount{
			ID:              toUUID(row.ID),
			LegalEntityID:   toUUID(row.LegalEntityID),
			CountNumber:     row.CountNumber,
			WarehouseID:     toUUID(row.WarehouseID),
			CountType:       domain.CountType(row.CountType),
			Status:          domain.CountStatus(row.Status),
			CountDate:       row.CountDate.Time,
			ReconciledAt:    fromPgTimestamptz(row.ReconciledAt),
			ReconciledBy:    toUUIDPtr(row.ReconciledBy),
			CreatedAt:       row.CreatedAt.Time,
			WarehouseCode:   row.WarehouseCode,
			WarehouseNameAr: row.WarehouseNameAr,
		})
	}
	return res, nil
}

func (r *inventoryRepository) UpdateStockCountStatus(ctx context.Context, id uuid.UUID, status domain.CountStatus, reconciledBy *uuid.UUID) (*domain.StockCount, error) {
	row, err := r.queries.UpdateStockCountStatus(ctx, sqlc.UpdateStockCountStatusParams{
		ID:           toPgUUID(id),
		Status:       string(status),
		ReconciledBy: toPgUUIDPtr(reconciledBy),
	})
	if err != nil {
		return nil, err
	}
	return r.GetStockCount(ctx, toUUID(row.ID))
}

// 7. Costing & Valuation

func (r *inventoryRepository) InsertCostLayer(ctx context.Context, layer *domain.CostLayer) (*domain.CostLayer, error) {
	row, err := r.queries.InsertCostLayer(ctx, sqlc.InsertCostLayerParams{
		LegalEntityID:     toPgUUID(layer.LegalEntityID),
		ItemID:            toPgUUID(layer.ItemID),
		WarehouseID:       toPgUUID(layer.WarehouseID),
		SourceMovementID:  toPgUUID(layer.SourceMovementID),
		OriginalQuantity:  toPgNumeric(layer.OriginalQuantity),
		RemainingQuantity: toPgNumeric(layer.RemainingQuantity),
		UnitCost:          toPgNumeric(layer.UnitCost),
		TotalCost:         toPgNumeric(layer.TotalCost),
	})
	if err != nil {
		return nil, err
	}
	layer.ID = toUUID(row.ID)
	layer.CreatedAt = row.CreatedAt.Time
	layer.UpdatedAt = row.UpdatedAt.Time
	return layer, nil
}

func (r *inventoryRepository) GetActiveCostLayersForUpdate(ctx context.Context, entityID, itemID, warehouseID uuid.UUID) ([]domain.CostLayer, error) {
	rows, err := r.queries.GetActiveCostLayersForUpdate(ctx, sqlc.GetActiveCostLayersForUpdateParams{
		LegalEntityID: toPgUUID(entityID),
		ItemID:        toPgUUID(itemID),
		WarehouseID:   toPgUUID(warehouseID),
	})
	if err != nil {
		return nil, err
	}
	layers := make([]domain.CostLayer, 0, len(rows))
	for _, row := range rows {
		layers = append(layers, domain.CostLayer{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			ItemID:            toUUID(row.ItemID),
			WarehouseID:       toUUID(row.WarehouseID),
			SourceMovementID:  toUUID(row.SourceMovementID),
			OriginalQuantity:  fromPgNumeric(row.OriginalQuantity),
			RemainingQuantity: fromPgNumeric(row.RemainingQuantity),
			UnitCost:          fromPgNumeric(row.UnitCost),
			TotalCost:         fromPgNumeric(row.TotalCost),
			CreatedAt:         row.CreatedAt.Time,
			UpdatedAt:         row.UpdatedAt.Time,
		})
	}
	return layers, nil
}

func (r *inventoryRepository) UpdateCostLayerRemaining(ctx context.Context, layerID uuid.UUID, remainingQty decimal.Decimal) (*domain.CostLayer, error) {
	row, err := r.queries.UpdateCostLayerRemaining(ctx, sqlc.UpdateCostLayerRemainingParams{
		ID:                toPgUUID(layerID),
		RemainingQuantity: toPgNumeric(remainingQty),
	})
	if err != nil {
		return nil, err
	}
	return &domain.CostLayer{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		ItemID:            toUUID(row.ItemID),
		WarehouseID:       toUUID(row.WarehouseID),
		SourceMovementID:  toUUID(row.SourceMovementID),
		OriginalQuantity:  fromPgNumeric(row.OriginalQuantity),
		RemainingQuantity: fromPgNumeric(row.RemainingQuantity),
		UnitCost:          fromPgNumeric(row.UnitCost),
		TotalCost:         fromPgNumeric(row.TotalCost),
		CreatedAt:         row.CreatedAt.Time,
		UpdatedAt:         row.UpdatedAt.Time,
	}, nil
}

func (r *inventoryRepository) InsertCostLayerConsumption(ctx context.Context, c *domain.CostLayerConsumption) (*domain.CostLayerConsumption, error) {
	row, err := r.queries.InsertCostLayerConsumption(ctx, sqlc.InsertCostLayerConsumptionParams{
		LegalEntityID:    toPgUUID(c.LegalEntityID),
		CostLayerID:      toPgUUID(c.CostLayerID),
		IssueMovementID:  toPgUUID(c.IssueMovementID),
		QuantityConsumed: toPgNumeric(c.QuantityConsumed),
		UnitCost:         toPgNumeric(c.UnitCost),
		TotalCost:        toPgNumeric(c.TotalCost),
	})
	if err != nil {
		return nil, err
	}
	c.ID = toUUID(row.ID)
	c.ConsumedAt = row.ConsumedAt.Time
	return c, nil
}

func (r *inventoryRepository) ListCostLayers(ctx context.Context, entityID uuid.UUID, itemID, warehouseID *uuid.UUID, onlyActive *bool, limit, offset int32) ([]domain.CostLayer, error) {
	if limit <= 0 {
		limit = 50
	}
	var activeParam pgtype.Bool
	if onlyActive != nil {
		activeParam = pgtype.Bool{Bool: *onlyActive, Valid: true}
	}
	rows, err := r.queries.ListCostLayers(ctx, sqlc.ListCostLayersParams{
		LegalEntityID: toPgUUID(entityID),
		ItemID:        toPgUUIDPtr(itemID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		OnlyActive:    activeParam,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	layers := make([]domain.CostLayer, 0, len(rows))
	for _, row := range rows {
		layers = append(layers, domain.CostLayer{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			ItemID:            toUUID(row.ItemID),
			WarehouseID:       toUUID(row.WarehouseID),
			SourceMovementID:  toUUID(row.SourceMovementID),
			OriginalQuantity:  fromPgNumeric(row.OriginalQuantity),
			RemainingQuantity: fromPgNumeric(row.RemainingQuantity),
			UnitCost:          fromPgNumeric(row.UnitCost),
			TotalCost:         fromPgNumeric(row.TotalCost),
			CreatedAt:         row.CreatedAt.Time,
			UpdatedAt:         row.UpdatedAt.Time,
			ItemCode:          row.ItemCode,
			ItemNameEn:        row.ItemNameEn,
			ItemNameAr:        row.ItemNameAr,
			WarehouseCode:     row.WarehouseCode,
			WarehouseNameAr:   row.WarehouseNameAr,
		})
	}
	return layers, nil
}

func (r *inventoryRepository) CreateInventoryRevaluation(ctx context.Context, rev *domain.InventoryRevaluation) (*domain.InventoryRevaluation, error) {
	row, err := r.queries.CreateInventoryRevaluation(ctx, sqlc.CreateInventoryRevaluationParams{
		LegalEntityID:     toPgUUID(rev.LegalEntityID),
		RevaluationNumber: rev.RevaluationNumber,
		WarehouseID:       toPgUUID(rev.WarehouseID),
		ItemID:            toPgUUID(rev.ItemID),
		QuantityOnHand:    toPgNumeric(rev.QuantityOnHand),
		OldUnitCost:       toPgNumeric(rev.OldUnitCost),
		NewUnitCost:       toPgNumeric(rev.NewUnitCost),
		OldTotalValue:     toPgNumeric(rev.OldTotalValue),
		NewTotalValue:     toPgNumeric(rev.NewTotalValue),
		VarianceAmount:    toPgNumeric(rev.VarianceAmount),
		Reason:            rev.Reason,
		JournalEntryID:    toPgUUIDPtr(rev.JournalEntryID),
		CreatedBy:         toPgUUID(rev.CreatedBy),
	})
	if err != nil {
		return nil, err
	}
	rev.ID = toUUID(row.ID)
	rev.CreatedAt = row.CreatedAt.Time
	return rev, nil
}

func (r *inventoryRepository) ListInventoryRevaluations(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID, limit, offset int32) ([]domain.InventoryRevaluation, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queries.ListInventoryRevaluations(ctx, sqlc.ListInventoryRevaluationsParams{
		LegalEntityID: toPgUUID(entityID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		ItemID:        toPgUUIDPtr(itemID),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}
	revs := make([]domain.InventoryRevaluation, 0, len(rows))
	for _, row := range rows {
		revs = append(revs, domain.InventoryRevaluation{
			ID:                toUUID(row.ID),
			LegalEntityID:     toUUID(row.LegalEntityID),
			RevaluationNumber: row.RevaluationNumber,
			WarehouseID:       toUUID(row.WarehouseID),
			ItemID:            toUUID(row.ItemID),
			QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
			OldUnitCost:       fromPgNumeric(row.OldUnitCost),
			NewUnitCost:       fromPgNumeric(row.NewUnitCost),
			OldTotalValue:     fromPgNumeric(row.OldTotalValue),
			NewTotalValue:     fromPgNumeric(row.NewTotalValue),
			VarianceAmount:    fromPgNumeric(row.VarianceAmount),
			Reason:            row.Reason,
			JournalEntryID:    toUUIDPtr(row.JournalEntryID),
			CreatedBy:         toUUID(row.CreatedBy),
			CreatedAt:         row.CreatedAt.Time,
			ItemCode:          row.ItemCode,
			ItemNameEn:        row.ItemNameEn,
			ItemNameAr:        row.ItemNameAr,
			WarehouseCode:     row.WarehouseCode,
			WarehouseNameAr:   row.WarehouseNameAr,
		})
	}
	return revs, nil
}

func (r *inventoryRepository) GetValuationSummary(ctx context.Context, entityID uuid.UUID, warehouseID, itemID *uuid.UUID) ([]domain.ValuationSummaryItem, error) {
	rows, err := r.queries.GetValuationSummary(ctx, sqlc.GetValuationSummaryParams{
		LegalEntityID: toPgUUID(entityID),
		WarehouseID:   toPgUUIDPtr(warehouseID),
		ItemID:        toPgUUIDPtr(itemID),
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.ValuationSummaryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.ValuationSummaryItem{
			LegalEntityID:     toUUID(row.LegalEntityID),
			ItemID:            toUUID(row.ItemID),
			WarehouseID:       toUUID(row.WarehouseID),
			QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
			QuantityReserved:  fromPgNumeric(row.QuantityReserved),
			QuantityAvailable: fromPgNumeric(row.QuantityAvailable),
			AverageCost:       fromPgNumeric(row.AverageCost),
			TotalValue:        fromPgNumeric(row.TotalValue),
			ItemCode:          row.ItemCode,
			ItemNameEn:        row.ItemNameEn,
			ItemNameAr:        row.ItemNameAr,
			CostingMethod:     domain.CostingMethod(row.CostingMethod),
			StandardCost:      fromPgNumeric(row.StandardCost),
			WarehouseCode:     row.WarehouseCode,
			WarehouseNameAr:   row.WarehouseNameAr,
		})
	}
	return items, nil
}

func (r *inventoryRepository) UpdateStockBalanceCostAndValue(ctx context.Context, entityID, itemID, warehouseID uuid.UUID, avgCost, totalValue decimal.Decimal) (*domain.StockBalance, error) {
	row, err := r.queries.UpdateStockBalanceCostAndValue(ctx, sqlc.UpdateStockBalanceCostAndValueParams{
		LegalEntityID: toPgUUID(entityID),
		ItemID:        toPgUUID(itemID),
		WarehouseID:   toPgUUID(warehouseID),
		AverageCost:   toPgNumeric(avgCost),
		TotalValue:    toPgNumeric(totalValue),
	})
	if err != nil {
		return nil, err
	}
	return &domain.StockBalance{
		ID:                toUUID(row.ID),
		LegalEntityID:     toUUID(row.LegalEntityID),
		ItemID:            toUUID(row.ItemID),
		WarehouseID:       toUUID(row.WarehouseID),
		BinID:             toUUID(row.BinID),
		QuantityOnHand:    fromPgNumeric(row.QuantityOnHand),
		QuantityReserved:  fromPgNumeric(row.QuantityReserved),
		QuantityAvailable: fromPgNumeric(row.QuantityAvailable),
		AverageCost:       fromPgNumeric(row.AverageCost),
		TotalValue:        fromPgNumeric(row.TotalValue),
		UpdatedAt:         row.UpdatedAt.Time,
		Version:           row.Version,
	}, nil
}

func (r *inventoryRepository) GetItemCostingDetails(ctx context.Context, itemID uuid.UUID) (*domain.ItemCostingInfo, error) {
	row, err := r.queries.GetItemCostingDetails(ctx, toPgUUID(itemID))
	if err != nil {
		return nil, err
	}
	return &domain.ItemCostingInfo{
		ID:            toUUID(row.ID),
		CostingMethod: domain.CostingMethod(row.CostingMethod),
		CostPrice:     fromPgNumeric(row.CostPrice),
		StandardCost:  fromPgNumeric(row.StandardCost),
	}, nil
}
