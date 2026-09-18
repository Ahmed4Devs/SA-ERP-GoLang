package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationAdvisoryLockID int64 = 675270348994611997

func MigrateUp(ctx context.Context, databaseURL string, migrationsDir string) error {
	ctx, cancel := migrationContext(ctx)
	defer cancel()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()
	if err := lockMigrations(ctx, db); err != nil {
		return err
	}
	defer unlockMigrations(ctx, db)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.UpContext(ctx, db, migrationsDir)
}

func MigrateDown(ctx context.Context, databaseURL string, migrationsDir string) error {
	ctx, cancel := migrationContext(ctx)
	defer cancel()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()
	if err := lockMigrations(ctx, db); err != nil {
		return err
	}
	defer unlockMigrations(ctx, db)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.DownContext(ctx, db, migrationsDir)
}

func MigrateStatus(ctx context.Context, databaseURL string, migrationsDir string) error {
	ctx, cancel := migrationContext(ctx)
	defer cancel()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.StatusContext(ctx, db, migrationsDir)
}

func lockMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockID); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	return nil
}

func unlockMigrations(ctx context.Context, db *sql.DB) {
	_, _ = db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockID)
}

func migrationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")) != "" {
		return context.WithTimeout(context.Background(), 15*time.Minute)
	}
	return ctx, func() {}
}
