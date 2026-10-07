package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"erp/internal/platform/config"
	"erp/internal/platform/database/sqlc"
	"erp/internal/platform/transaction"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
	cfg  *config.Config
}

const LatestSchemaVersion int64 = 350

func NewPool(ctx context.Context, cfg *config.Config) (*DB, error) {
	if err := prepareIntegrationTestDatabase(cfg); err != nil {
		return nil, err
	}
	connectCtx := ctx
	if cfg.Environment == config.EnvTesting {
		// Test packages may wait for the cross-process database lock longer than
		// the short setup context created by an individual test. Use a fresh
		// bounded context once that lock is acquired.
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	// Verify connectivity
	pingCtx, cancel := context.WithTimeout(connectCtx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		Pool: pool,
		cfg:  cfg,
	}, nil
}

func (db *DB) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.Pool.Ping(ctx); err != nil {
		return err
	}
	var version int64
	err := db.Pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version_id), 0)
		FROM goose_db_version
		WHERE is_applied = TRUE`).Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}
	if version != LatestSchemaVersion {
		return fmt.Errorf("%w: database=%d application=%d", ErrSchemaVersionMismatch, version, LatestSchemaVersion)
	}
	return nil
}

var ErrSchemaVersionMismatch = errors.New("database schema version does not match application")

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// DBTX returns the active transaction from context if present; otherwise, returns the pool.
func (db *DB) DBTX(ctx context.Context) sqlc.DBTX {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return db.Pool
}
