package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func MigrateUp(ctx context.Context, databaseURL string, migrationsDir string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.UpContext(ctx, db, migrationsDir)
}

func MigrateDown(ctx context.Context, databaseURL string, migrationsDir string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.DownContext(ctx, db, migrationsDir)
}

func MigrateStatus(ctx context.Context, databaseURL string, migrationsDir string) error {
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
