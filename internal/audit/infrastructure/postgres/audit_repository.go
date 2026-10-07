package postgres

import (
	"context"
	"errors"
	"time"

	"erp/internal/audit/domain"
	"erp/internal/audit/ports"
	"erp/internal/platform/database/sqlc"
	"erp/internal/platform/transaction"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type auditRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewAuditRepository(pool *pgxpool.Pool) ports.AuditRepository {
	return &auditRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *auditRepository) queriesFor(ctx context.Context) *sqlc.Queries {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *auditRepository) LockChain(ctx context.Context) error {
	tx, ok := transaction.TxFromContext(ctx)
	if !ok {
		return errors.New("audit chain lock requires a transaction")
	}
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended('audit.events.hash-chain', 0))")
	return err
}

// AppendEvent serializes the global hash chain with a PostgreSQL transaction
// advisory lock. The event and all associated changes commit atomically, so
// multiple API replicas cannot fork the chain or leave partial state diffs.
func (r *auditRepository) AppendEvent(ctx context.Context, event *domain.AuditEvent, changes []domain.EntityChange) (*domain.AuditEvent, error) {
	tx, ok := transaction.TxFromContext(ctx)
	if !ok {
		var inserted *domain.AuditEvent
		err := transaction.NewManager(r.pool).WithinTx(ctx, func(txCtx context.Context) error {
			var err error
			inserted, err = r.AppendEvent(txCtx, event, changes)
			return err
		})
		return inserted, err
	}

	if err := r.LockChain(ctx); err != nil {
		return nil, err
	}

	qtx := r.queriesFor(ctx).WithTx(tx)
	previousHash := domain.GenesisHash
	latest, err := qtx.GetLatestAuditEvent(ctx)
	if err == nil {
		previousHash = latest.RecordHash
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	event.Payload = domain.CanonicalizeJSON(event.Payload)
	event.PreviousHash = previousHash
	event.OccurredAt = time.Now().UTC().Truncate(time.Microsecond)
	// Keep chronological verification stable across clock corrections and
	// multiple records created within one timestamp resolution.
	if latest.OccurredAt.Valid && !event.OccurredAt.After(latest.OccurredAt.Time) {
		event.OccurredAt = latest.OccurredAt.Time.Add(time.Microsecond)
	}
	event.RecordHash = domain.ComputeRecordHash(
		event.PreviousHash,
		event.UserID,
		event.LegalEntityID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		event.Payload,
		event.OccurredAt,
	)

	row, err := qtx.InsertAuditEvent(ctx, sqlc.InsertAuditEventParams{
		UserID:        toPgUUIDPtr(event.UserID),
		LegalEntityID: toPgUUIDPtr(event.LegalEntityID),
		Action:        event.Action,
		ResourceType:  event.ResourceType,
		ResourceID:    event.ResourceID,
		Payload:       event.Payload,
		IpAddress:     toPgText(event.IPAddress),
		UserAgent:     toPgText(event.UserAgent),
		PreviousHash:  toPgText(event.PreviousHash),
		RecordHash:    event.RecordHash,
		OccurredAt:    toPgTimestamptz(&event.OccurredAt),
	})
	if err != nil {
		return nil, err
	}

	inserted := toDomainAuditEvent(&row)
	for i := range changes {
		changes[i].EventID = inserted.ID
		changeRow, err := qtx.InsertEntityChange(ctx, sqlc.InsertEntityChangeParams{
			EventID:     toPgUUID(changes[i].EventID),
			TableName:   changes[i].TableName,
			RecordID:    changes[i].RecordID,
			BeforeState: changes[i].BeforeState,
			AfterState:  changes[i].AfterState,
		})
		if err != nil {
			return nil, err
		}
		inserted.Changes = append(inserted.Changes, *toDomainEntityChange(&changeRow))
	}

	return inserted, nil
}

func (r *auditRepository) InsertEvent(ctx context.Context, event *domain.AuditEvent) (*domain.AuditEvent, error) {
	payload := domain.CanonicalizeJSON(event.Payload)

	row, err := r.queriesFor(ctx).InsertAuditEvent(ctx, sqlc.InsertAuditEventParams{
		UserID:        toPgUUIDPtr(event.UserID),
		LegalEntityID: toPgUUIDPtr(event.LegalEntityID),
		Action:        event.Action,
		ResourceType:  event.ResourceType,
		ResourceID:    event.ResourceID,
		Payload:       payload,
		IpAddress:     toPgText(event.IPAddress),
		UserAgent:     toPgText(event.UserAgent),
		PreviousHash:  toPgText(event.PreviousHash),
		RecordHash:    event.RecordHash,
		OccurredAt:    toPgTimestamptz(&event.OccurredAt),
	})
	if err != nil {
		return nil, err
	}

	return toDomainAuditEvent(&row), nil
}

func (r *auditRepository) GetLatestEvent(ctx context.Context) (*domain.AuditEvent, error) {
	row, err := r.queriesFor(ctx).GetLatestAuditEvent(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainAuditEvent(&row), nil
}

func (r *auditRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*domain.AuditEvent, error) {
	row, err := r.queriesFor(ctx).GetAuditEventByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}
	return toDomainAuditEvent(&row), nil
}

func (r *auditRepository) ListEvents(ctx context.Context, filter domain.AuditFilter) ([]domain.AuditEvent, int64, error) {
	limit := int32(50)
	if filter.Limit > 0 {
		limit = int32(filter.Limit)
	}
	offset := int32(0)
	if filter.Offset > 0 {
		offset = int32(filter.Offset)
	}

	count, err := r.queriesFor(ctx).CountAuditEvents(ctx, sqlc.CountAuditEventsParams{
		LegalEntityID: toPgUUIDPtr(filter.LegalEntityID),
		UserID:        toPgUUIDPtr(filter.UserID),
		Action:        toPgText(filter.Action),
		ResourceType:  toPgText(filter.ResourceType),
		ResourceID:    toPgText(filter.ResourceID),
		FromDate:      toPgTimestamptz(filter.FromDate),
		ToDate:        toPgTimestamptz(filter.ToDate),
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queriesFor(ctx).ListAuditEvents(ctx, sqlc.ListAuditEventsParams{
		Limit:         limit,
		Offset:        offset,
		LegalEntityID: toPgUUIDPtr(filter.LegalEntityID),
		UserID:        toPgUUIDPtr(filter.UserID),
		Action:        toPgText(filter.Action),
		ResourceType:  toPgText(filter.ResourceType),
		ResourceID:    toPgText(filter.ResourceID),
		FromDate:      toPgTimestamptz(filter.FromDate),
		ToDate:        toPgTimestamptz(filter.ToDate),
	})
	if err != nil {
		return nil, 0, err
	}

	result := make([]domain.AuditEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainAuditEvent(&row))
	}

	return result, count, nil
}

func (r *auditRepository) ListEventsChronological(ctx context.Context, limit, offset int32) ([]domain.AuditEvent, error) {
	rows, err := r.queriesFor(ctx).ListAuditEventsChronological(ctx, sqlc.ListAuditEventsChronologicalParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.AuditEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainAuditEvent(&row))
	}
	return result, nil
}

func (r *auditRepository) InsertEntityChange(ctx context.Context, change *domain.EntityChange) (*domain.EntityChange, error) {
	row, err := r.queriesFor(ctx).InsertEntityChange(ctx, sqlc.InsertEntityChangeParams{
		EventID:     toPgUUID(change.EventID),
		TableName:   change.TableName,
		RecordID:    change.RecordID,
		BeforeState: change.BeforeState,
		AfterState:  change.AfterState,
	})
	if err != nil {
		return nil, err
	}

	return toDomainEntityChange(&row), nil
}

func (r *auditRepository) ListEntityChanges(ctx context.Context, eventID uuid.UUID) ([]domain.EntityChange, error) {
	rows, err := r.queriesFor(ctx).ListEntityChangesByEventID(ctx, toPgUUID(eventID))
	if err != nil {
		return nil, err
	}

	result := make([]domain.EntityChange, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainEntityChange(&row))
	}
	return result, nil
}

func (r *auditRepository) InsertSecurityEvent(ctx context.Context, event *domain.SecurityEvent) (*domain.SecurityEvent, error) {
	details := event.Details
	if len(details) == 0 {
		details = []byte("{}")
	}

	row, err := r.queriesFor(ctx).InsertSecurityEvent(ctx, sqlc.InsertSecurityEventParams{
		UserID:    toPgUUIDPtr(event.UserID),
		EventType: event.EventType,
		Status:    event.Status,
		Details:   details,
		IpAddress: toPgText(event.IPAddress),
	})
	if err != nil {
		return nil, err
	}

	return toDomainSecurityEvent(&row), nil
}

func (r *auditRepository) ListSecurityEvents(ctx context.Context, userID *uuid.UUID, eventType, status string, limit, offset int32) ([]domain.SecurityEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queriesFor(ctx).ListSecurityEvents(ctx, sqlc.ListSecurityEventsParams{
		UserID:    toPgUUIDPtr(userID),
		EventType: toPgText(eventType),
		Status:    toPgText(status),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.SecurityEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainSecurityEvent(&row))
	}
	return result, nil
}

func (r *auditRepository) InsertAccessEvent(ctx context.Context, event *domain.AccessEvent) (*domain.AccessEvent, error) {
	row, err := r.queriesFor(ctx).InsertAccessEvent(ctx, sqlc.InsertAccessEventParams{
		UserID:       toPgUUIDPtr(event.UserID),
		ResourcePath: event.ResourcePath,
		Method:       event.Method,
		StatusCode:   int32(event.StatusCode),
		IpAddress:    toPgText(event.IPAddress),
		LatencyMs:    int32(event.LatencyMS),
	})
	if err != nil {
		return nil, err
	}

	return toDomainAccessEvent(&row), nil
}

func (r *auditRepository) ListAccessEvents(ctx context.Context, userID *uuid.UUID, resourcePath string, limit, offset int32) ([]domain.AccessEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.queriesFor(ctx).ListAccessEvents(ctx, sqlc.ListAccessEventsParams{
		UserID:       toPgUUIDPtr(userID),
		ResourcePath: toPgText(resourcePath),
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.AccessEvent, 0, len(rows))
	for _, row := range rows {
		result = append(result, *toDomainAccessEvent(&row))
	}
	return result, nil
}

func toDomainAuditEvent(row *sqlc.AuditEvents) *domain.AuditEvent {
	var prevHash string
	if row.PreviousHash.Valid {
		prevHash = row.PreviousHash.String
	}

	return &domain.AuditEvent{
		ID:            toUUID(row.ID),
		UserID:        toUUIDPtr(row.UserID),
		LegalEntityID: toUUIDPtr(row.LegalEntityID),
		Action:        row.Action,
		ResourceType:  row.ResourceType,
		ResourceID:    row.ResourceID,
		Payload:       row.Payload,
		IPAddress:     fromPgText(row.IpAddress),
		UserAgent:     fromPgText(row.UserAgent),
		PreviousHash:  prevHash,
		RecordHash:    row.RecordHash,
		OccurredAt:    row.OccurredAt.Time,
	}
}

func toDomainEntityChange(row *sqlc.AuditEntityChanges) *domain.EntityChange {
	return &domain.EntityChange{
		ID:          toUUID(row.ID),
		EventID:     toUUID(row.EventID),
		TableName:   row.TableName,
		RecordID:    row.RecordID,
		BeforeState: row.BeforeState,
		AfterState:  row.AfterState,
		CreatedAt:   row.CreatedAt.Time,
	}
}

func toDomainSecurityEvent(row *sqlc.AuditSecurityEvents) *domain.SecurityEvent {
	return &domain.SecurityEvent{
		ID:         toUUID(row.ID),
		UserID:     toUUIDPtr(row.UserID),
		EventType:  row.EventType,
		Status:     row.Status,
		Details:    row.Details,
		IPAddress:  fromPgText(row.IpAddress),
		OccurredAt: row.OccurredAt.Time,
	}
}

func toDomainAccessEvent(row *sqlc.AuditAccessEvents) *domain.AccessEvent {
	return &domain.AccessEvent{
		ID:           toUUID(row.ID),
		UserID:       toUUIDPtr(row.UserID),
		ResourcePath: row.ResourcePath,
		Method:       row.Method,
		StatusCode:   int(row.StatusCode),
		IPAddress:    fromPgText(row.IpAddress),
		LatencyMS:    int(row.LatencyMs),
		OccurredAt:   row.OccurredAt.Time,
	}
}
