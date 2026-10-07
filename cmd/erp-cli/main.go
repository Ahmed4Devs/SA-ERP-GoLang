package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"erp/internal/platform/migrationcli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	var err error
	switch os.Args[1] {
	case "version":
		version := os.Getenv("APP_VERSION")
		if version == "" {
			version = "1.0.0"
		}
		fmt.Printf("Saudi ERP CLI v%s\n", version)
	case "migrate":
		err = migrationcli.Run(ctx, os.Args[2:], os.Stdout, os.Stderr)
	case "seed":
		// Reference records are versioned with the schema. Demo business data
		// is controlled separately by explicit API bootstrap configuration.
		if len(os.Args) > 2 {
			err = fmt.Errorf("seed accepts no options; use migrate up --path for a custom migration directory")
			break
		}
		err = migrationcli.Run(ctx, []string{"up"}, os.Stdout, os.Stderr)
		if err == nil {
			fmt.Println("Reference data migrations are current. Demo records require BOOTSTRAP_DEMO_DATA_ENABLED=true in development.")
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
		printUsage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Saudi ERP CLI Utility")
	fmt.Println("Usage: erp-cli <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  version                  Show application version")
	fmt.Println("  migrate [up|down|status]  Run database migrations (--path, --timeout)")
	fmt.Println("  seed                     Apply schema and reference-data migrations")
}
