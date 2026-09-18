package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"erp/internal/platform/config"
	"github.com/jackc/pgx/v5"
)

// Integration test packages are separate processes. Holding this session-level
// advisory lock for the lifetime of each test binary prevents packages that
// share TEST_DATABASE_URL from truncating or seeding the same database at once.
const integrationTestPackageLockID int64 = 675270348994611998

var (
	testDatabaseOnce    sync.Once
	testDatabaseInitErr error
	testDatabaseLock    *pgx.Conn
)

func prepareIntegrationTestDatabase(cfg *config.Config) error {
	if cfg.Environment != config.EnvTesting {
		return nil
	}

	testDatabaseOnce.Do(func() {
		setupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		lockConn, err := pgx.Connect(setupCtx, cfg.Database.URL)
		if err != nil {
			testDatabaseInitErr = fmt.Errorf("failed to connect for integration test isolation: %w", err)
			return
		}
		if _, err := lockConn.Exec(setupCtx, "SELECT pg_advisory_lock($1)", integrationTestPackageLockID); err != nil {
			_ = lockConn.Close(context.Background())
			testDatabaseInitErr = fmt.Errorf("failed to acquire integration test package lock: %w", err)
			return
		}
		testDatabaseLock = lockConn

		migrationsDir, err := findMigrationsDir()
		if err != nil {
			testDatabaseInitErr = err
			return
		}
		if err := MigrateUp(setupCtx, cfg.Database.URL, migrationsDir); err != nil {
			testDatabaseInitErr = fmt.Errorf("failed to prepare integration test schema: %w", err)
		}
	})

	return testDatabaseInitErr
}

func findMigrationsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to resolve working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "database", "migrations")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("database/migrations not found from %s", dir)
		}
		dir = parent
	}
}
