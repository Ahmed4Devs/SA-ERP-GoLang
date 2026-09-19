package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"erp/internal/platform/database"
	"erp/internal/platform/database/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPollInterval  = 5 * time.Second
	defaultLeaseDuration = time.Minute
	defaultBatchSize     = 50
)

// Event is a durable message waiting to be delivered to an in-process handler.
type Event struct {
	ID            uuid.UUID
	EventType     string
	AggregateType string
	AggregateID   *uuid.UUID
	Payload       json.RawMessage
	OccurredAt    time.Time
	AvailableAt   time.Time
	AttemptCount  int32
}

// Enqueuer is the minimal contract used by transactional event producers.
type Enqueuer interface {
	Enqueue(ctx context.Context, eventType, aggregateType string, aggregateID *uuid.UUID, payload any) (*Event, error)
}

// Store persists and leases outbox events.
type Store interface {
	Enqueuer
	Claim(ctx context.Context, batchSize int, leaseUntil time.Time) ([]Event, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	RecordFailure(ctx context.Context, id uuid.UUID, message string, availableAt time.Time) error
	DeadLetter(ctx context.Context, id uuid.UUID, message string) error
}

// Repository implements Store on PostgreSQL. It honours transactions carried in ctx.
type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db.DBTX(ctx))
}

func (r *Repository) Enqueue(ctx context.Context, eventType, aggregateType string, aggregateID *uuid.UUID, payload any) (*Event, error) {
	if eventType == "" || aggregateType == "" {
		return nil, errors.New("outbox event_type and aggregate_type are required")
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}
	now := time.Now().UTC()
	row, err := r.queries(ctx).InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   toPGUUID(aggregateID),
		Payload:       payloadBytes,
		OccurredAt:    pgtype.Timestamptz{Time: now, Valid: true},
		AvailableAt:   pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("enqueue outbox event: %w", err)
	}
	event := fromRow(row)
	return &event, nil
}

func (r *Repository) Claim(ctx context.Context, batchSize int, leaseUntil time.Time) ([]Event, error) {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	rows, err := r.queries(ctx).ClaimPendingOutboxEvents(ctx, sqlc.ClaimPendingOutboxEventsParams{
		LeaseUntil: pgtype.Timestamptz{Time: leaseUntil.UTC(), Valid: true},
		BatchSize:  int32(batchSize),
	})
	if err != nil {
		return nil, fmt.Errorf("claim outbox events: %w", err)
	}
	events := make([]Event, 0, len(rows))
	for _, row := range rows {
		events = append(events, fromRow(row))
	}
	return events, nil
}

func (r *Repository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	if err := r.queries(ctx).MarkOutboxEventPublished(ctx, pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		return fmt.Errorf("mark outbox event published: %w", err)
	}
	return nil
}

func (r *Repository) RecordFailure(ctx context.Context, id uuid.UUID, message string, availableAt time.Time) error {
	if err := r.queries(ctx).RecordOutboxEventAttempt(ctx, sqlc.RecordOutboxEventAttemptParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		LastError:   pgtype.Text{String: message, Valid: message != ""},
		AvailableAt: pgtype.Timestamptz{Time: availableAt.UTC(), Valid: true},
	}); err != nil {
		return fmt.Errorf("record outbox event failure: %w", err)
	}
	return nil
}

func (r *Repository) DeadLetter(ctx context.Context, id uuid.UUID, message string) error {
	if err := r.queries(ctx).MarkOutboxEventDeadLettered(ctx, sqlc.MarkOutboxEventDeadLetteredParams{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		LastError: pgtype.Text{String: message, Valid: message != ""},
	}); err != nil {
		return fmt.Errorf("dead-letter outbox event: %w", err)
	}
	return nil
}

func toPGUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func fromRow(row sqlc.PlatformOutboxEvents) Event {
	event := Event{
		ID:            uuid.UUID(row.ID.Bytes),
		EventType:     row.EventType,
		AggregateType: row.AggregateType,
		Payload:       append(json.RawMessage(nil), row.Payload...),
		OccurredAt:    row.OccurredAt.Time,
		AvailableAt:   row.AvailableAt.Time,
		AttemptCount:  row.AttemptCount,
	}
	if row.AggregateID.Valid {
		id := uuid.UUID(row.AggregateID.Bytes)
		event.AggregateID = &id
	}
	return event
}

// Handler processes one event. It must be safe to retry.
type Handler func(context.Context, Event) error

type DispatcherConfig struct {
	PollInterval  time.Duration
	LeaseDuration time.Duration
	BatchSize     int
	MaxAttempts   int32
	MaxBackoff    time.Duration
}

func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		PollInterval:  defaultPollInterval,
		LeaseDuration: defaultLeaseDuration,
		BatchSize:     defaultBatchSize,
		MaxAttempts:   10,
		MaxBackoff:    30 * time.Minute,
	}
}

// Dispatcher leases events so multiple worker replicas can process the queue safely.
type Dispatcher struct {
	store    Store
	logger   *slog.Logger
	cfg      DispatcherConfig
	mu       sync.RWMutex
	handlers map[string]Handler
}

func NewDispatcher(store Store, logger *slog.Logger, cfg DispatcherConfig) *Dispatcher {
	defaults := DefaultDispatcherConfig()
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaults.PollInterval
	}
	if cfg.LeaseDuration <= 0 {
		cfg.LeaseDuration = defaults.LeaseDuration
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaults.BatchSize
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaults.MaxAttempts
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = defaults.MaxBackoff
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{store: store, logger: logger, cfg: cfg, handlers: make(map[string]Handler)}
}

func (d *Dispatcher) Register(eventType string, handler Handler) error {
	if eventType == "" || handler == nil {
		return errors.New("outbox event type and handler are required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.handlers[eventType]; exists {
		return fmt.Errorf("outbox handler already registered for %q", eventType)
	}
	d.handlers[eventType] = handler
	return nil
}

func (d *Dispatcher) RunOnce(ctx context.Context) (int, error) {
	events, err := d.store.Claim(ctx, d.cfg.BatchSize, time.Now().UTC().Add(d.cfg.LeaseDuration))
	if err != nil {
		return 0, err
	}

	processed := 0
	var errs []error
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return processed, errors.Join(append(errs, err)...)
		}

		d.mu.RLock()
		handler := d.handlers[event.EventType]
		d.mu.RUnlock()
		if handler == nil {
			err = fmt.Errorf("no outbox handler registered for %q", event.EventType)
		} else {
			err = handler(ctx, event)
		}

		if err != nil {
			nextAttempt := event.AttemptCount + 1
			var recordErr error
			if nextAttempt >= d.cfg.MaxAttempts {
				recordErr = d.store.DeadLetter(ctx, event.ID, err.Error())
			} else {
				retryAt := time.Now().UTC().Add(d.backoff(nextAttempt))
				recordErr = d.store.RecordFailure(ctx, event.ID, err.Error(), retryAt)
			}
			if recordErr != nil {
				err = errors.Join(err, recordErr)
			}
			errs = append(errs, fmt.Errorf("process outbox event %s: %w", event.ID, err))
			continue
		}
		if err := d.store.MarkPublished(ctx, event.ID); err != nil {
			errs = append(errs, err)
			continue
		}
		processed++
	}
	return processed, errors.Join(errs...)
}

func (d *Dispatcher) Run(ctx context.Context) {
	d.runAndLog(ctx)
	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.runAndLog(ctx)
		}
	}
}

func (d *Dispatcher) runAndLog(ctx context.Context) {
	processed, err := d.RunOnce(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		d.logger.Error("outbox dispatch pass failed", "error", err)
	}
	if processed > 0 {
		d.logger.Info("outbox dispatch pass completed", "processed", processed)
	}
}

func (d *Dispatcher) backoff(attempt int32) time.Duration {
	seconds := math.Pow(2, float64(min(attempt, 10)))
	delay := time.Duration(seconds) * time.Second
	if delay > d.cfg.MaxBackoff {
		return d.cfg.MaxBackoff
	}
	return delay
}
