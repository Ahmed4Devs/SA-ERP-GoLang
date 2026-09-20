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

const resetIntegrationSchemasSQL = `
DROP SCHEMA IF EXISTS reporting, notifications, documents, zatca, budgeting,
fixedassets, gosi, payroll, leave, attendance, hr, banking, treasury,
purchasing, payables, sales, receivables, inventory, warehouse, products,
tax, accounting, workflow, audit, authz, organization, iam, platform, ref CASCADE;
DROP TABLE IF EXISTS public.goose_db_version CASCADE;
`

const seedIntegrationPrincipalSQL = `
INSERT INTO iam.users (
    id, username, email, display_name, preferred_locale, preferred_timezone,
    status, created_at, updated_at, version
) VALUES (
    '00000000-0000-7000-8000-000000000001', 'admin', 'admin@erp.test',
    'Integration Test Administrator', 'ar-SA', 'Asia/Riyadh',
    'ACTIVE', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 1
)
ON CONFLICT DO NOTHING;
`

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

		// Each integration package is a separate process. Rebuild the application
		// schemas while holding the package-wide lock so no package can inherit
		// fixtures from another package or race its migrations.
		migrationsDir, err := findMigrationsDir()
		if err != nil {
			testDatabaseInitErr = err
			return
		}
		if _, err := testDatabaseLock.Exec(setupCtx, resetIntegrationSchemasSQL); err != nil {
			testDatabaseInitErr = fmt.Errorf("failed to reset integration test schemas: %w", err)
			return
		}
		if err := MigrateUp(setupCtx, cfg.Database.URL, migrationsDir); err != nil {
			testDatabaseInitErr = fmt.Errorf("failed to prepare integration test schema: %w", err)
			return
		}
		if _, err := testDatabaseLock.Exec(setupCtx, seedIntegrationPrincipalSQL); err != nil {
			testDatabaseInitErr = fmt.Errorf("failed to seed integration test principal: %w", err)
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
