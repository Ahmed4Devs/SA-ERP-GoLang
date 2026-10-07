package migrationcli

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestMigrationCLIRejectsInvalidInputsBeforeConnecting(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "MIGRATION_DATABASE_URL", "GOOSE_DBSTRING"} {
		t.Setenv(key, "")
	}
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"credentials required", []string{"up"}, "set MIGRATION_DATABASE_URL"},
		{"invalid command", []string{"reset"}, "unknown migration command"},
		{"extra arguments", []string{"up", "extra"}, "unexpected arguments"},
		{"invalid duration", []string{"up", "--timeout", "0s"}, "timeout must be positive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Run(context.Background(), tc.args, io.Discard, io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v; want error containing %q", err, tc.want)
			}
		})
	}
}

func TestMigrationCLIHelpNeedsNoRuntimeSecrets(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("MFA_ENCRYPTION_KEY", "")
	if err := Run(context.Background(), []string{"--help"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
}
