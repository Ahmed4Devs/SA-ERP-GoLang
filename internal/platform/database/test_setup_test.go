package database

import (
	"context"
	"strings"
	"testing"

	"erp/internal/platform/config"
)

func TestDatabaseTestsRequireExplicitIsolatedURL(t *testing.T) {
	t.Setenv("TEST_DATABASE_URL", "")
	cfg := &config.Config{
		Environment: config.EnvDevelopment,
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:toor@localhost:5432/erp_dev?sslmode=disable",
			MaxConns: 5,
		},
	}

	if _, err := NewPool(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "TEST_DATABASE_URL") {
		t.Fatalf("NewPool must reject the development database in a test binary, got %v", err)
	}
	if err := MigrateUp(context.Background(), cfg.Database.URL, "."); err == nil || !strings.Contains(err.Error(), "TEST_DATABASE_URL") {
		t.Fatalf("MigrateUp must reject the development database in a test binary, got %v", err)
	}
}

func TestDatabaseTestsRejectMismatchedURLAndEnvironment(t *testing.T) {
	const testURL = "postgres://test:test@localhost:5432/erp_test?sslmode=disable"
	t.Setenv("TEST_DATABASE_URL", testURL)

	if err := requireIsolatedTestDatabaseURL("postgres://postgres:toor@localhost:5432/erp_dev?sslmode=disable"); err == nil {
		t.Fatal("expected mismatched database URL to be rejected")
	}
	if err := prepareIntegrationTestDatabase(&config.Config{
		Environment: config.EnvDevelopment,
		Database:    config.DatabaseConfig{URL: testURL},
	}); err == nil || !strings.Contains(err.Error(), "APP_ENV=testing") {
		t.Fatalf("expected explicit testing environment, got %v", err)
	}
}
