package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txContextKey struct{}

// Manager defines the canonical transactional boundary contract per Section 20.
type Manager interface {
	WithinTx(
		ctx context.Context,
		fn func(ctx context.Context) error,
	) error
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// ContextWithTx injects an active pgx.Tx into the context.
func ContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext retrieves an active pgx.Tx from the context if present.
func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx, ok
}

// WithinTx executes fn within an atomic transaction.
// If an active transaction already exists in ctx, it reuses it.
// If fn returns an error or panics, the transaction is rolled back.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	// 1. Check if a transaction is already active in this context
	if _, ok := TxFromContext(ctx); ok {
		return fn(ctx)
	}

	// 2. Begin a new transaction
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				err = errors.Join(err, fmt.Errorf("transaction rollback failed: %w", rbErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = fmt.Errorf("transaction commit failed: %w", commitErr)
			}
		}
	}()

	// 3. Inject transaction into context and invoke fn
	txCtx := ContextWithTx(ctx, tx)
	err = fn(txCtx)
	return err
}
