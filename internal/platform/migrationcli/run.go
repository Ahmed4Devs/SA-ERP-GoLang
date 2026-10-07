package migrationcli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"erp/internal/platform/database"
)

// Run accepts both `up --path dir` and the legacy `--dir up --path dir`.
// Migration credentials are separate from runtime credentials in production.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("erp-migrate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	command := "up"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command, args = args[0], args[1:]
	}
	direction := flags.String("dir", command, "Migration command: up, down, or status")
	path := flags.String("path", envOr("GOOSE_MIGRATION_DIR", "database/migrations"), "Migration directory")
	timeout := flags.Duration("timeout", 10*time.Minute, "Maximum migration duration")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *direction != "up" && *direction != "down" && *direction != "status" {
		return fmt.Errorf("unknown migration command %q: use up, down, or status", *direction)
	}
	if *timeout <= 0 {
		return errors.New("migration timeout must be positive")
	}
	databaseURL := envOr("MIGRATION_DATABASE_URL", envOr("GOOSE_DBSTRING", envOr("DATABASE_URL", "")))
	if databaseURL == "" {
		return errors.New("set MIGRATION_DATABASE_URL, GOOSE_DBSTRING, or DATABASE_URL before running migrations")
	}
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	var err error
	switch *direction {
	case "up":
		err = database.MigrateUp(ctx, databaseURL, *path)
	case "down":
		err = database.MigrateDown(ctx, databaseURL, *path)
	case "status":
		err = database.MigrateStatus(ctx, databaseURL, *path)
	}
	if err != nil {
		return fmt.Errorf("migration %s failed: %w", *direction, err)
	}
	fmt.Fprintf(stdout, "Migration %s completed successfully\n", *direction)
	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
