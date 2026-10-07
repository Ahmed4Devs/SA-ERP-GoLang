package services

import (
	"context"
	"encoding/json"
	"errors"

	"erp/internal/accounting/application/posting"
	"erp/internal/platform/database"
	"github.com/google/uuid"
)

type ValuationService struct {
	db     *database.DB
	ledger *posting.SourceService
}

func NewValuationService(db *database.DB, ledger *posting.SourceService) *ValuationService {
	return &ValuationService{db: db, ledger: ledger}
}

// CreateSnapshot captures one MVCC snapshot of the current warehouse cost
// projection. It does not claim to reconstruct arbitrary historical dates.
func (s *ValuationService) CreateSnapshot(ctx context.Context, entity, actor uuid.UUID) (uuid.UUID, int64, error) {
	if entity == uuid.Nil || actor == uuid.Nil {
		return uuid.Nil, 0, errors.New("entity and authenticated actor required")
	}
	batch := uuid.New()
	var count int64
	err := s.ledger.WithinTx(ctx, func(ctx context.Context) error {
		result, err := s.db.DBTX(ctx).Exec(ctx, `INSERT INTO inventory.valuation_snapshots(legal_entity_id,snapshot_batch_id,item_id,warehouse_id,costing_method,quantity_on_hand,unit_cost,total_value,created_by) SELECT legal_entity_id,$2,item_id,warehouse_id,costing_method,quantity_on_hand,unit_cost,total_value,$3 FROM inventory.item_costs WHERE legal_entity_id=$1`, entity, batch, actor)
		if err != nil {
			return err
		}
		count = result.RowsAffected()
		if count == 0 {
			return errors.New("no inventory costs available to snapshot")
		}
		return s.ledger.RecordSource(ctx, entity, actor, batch, "INVENTORY_VALUATION_SNAPSHOT", map[string]any{"records": count})
	})
	return batch, count, err
}

func (s *ValuationService) List(ctx context.Context, entity uuid.UUID, snapshots bool, batch *uuid.UUID, limit, offset int) (json.RawMessage, int64, error) {
	if entity == uuid.Nil {
		return nil, 0, errors.New("entity is required")
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id,legal_entity_id,item_id,warehouse_id,costing_method,quantity_on_hand::text,unit_cost::text,total_value::text,updated_at AS snapshot_at FROM inventory.item_costs WHERE legal_entity_id=$1`
	if snapshots {
		query = `SELECT id,legal_entity_id,snapshot_batch_id,snapshot_at,item_id,warehouse_id,costing_method,quantity_on_hand::text,unit_cost::text,total_value::text,created_by FROM inventory.valuation_snapshots WHERE legal_entity_id=$1 AND ($4::uuid IS NULL OR snapshot_batch_id=$4)`
	}
	args := []any{entity, limit, offset}
	if snapshots {
		args = append(args, batch)
	}
	var data json.RawMessage
	var total int64
	err := s.db.DBTX(ctx).QueryRow(ctx, `WITH records AS (`+query+`) SELECT (SELECT count(*) FROM records),coalesce((SELECT jsonb_agg(to_jsonb(p)) FROM (SELECT * FROM records ORDER BY snapshot_at DESC,id DESC LIMIT $2 OFFSET $3)p),'[]'::jsonb)`, args...).Scan(&total, &data)
	return data, total, err
}
