package database

import (
	"context"
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

func NewPool(ctx context.Context, cfg *config.Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	// Verify connectivity
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	return db.Pool.Ping(ctx)
}

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
