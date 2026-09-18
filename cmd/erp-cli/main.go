package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"erp/internal/platform/config"
	"erp/internal/platform/database"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version":
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Saudi ERP CLI v1.0.0\n")
			return
		}
		fmt.Printf("Saudi ERP Platform v%s (%s)\n", cfg.AppVersion, cfg.Environment)

	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		direction := migrateCmd.String("dir", "up", "Migration direction: up, down, or status")
		migrationsDir := migrateCmd.String("path", "database/migrations", "Path to migrations directory")
		_ = migrateCmd.Parse(os.Args[2:])

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}

		ctx := context.Background()
		switch *direction {
		case "up":
			if err := database.MigrateUp(ctx, cfg.Database.URL, *migrationsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Database migrations applied successfully")
		case "down":
			if err := database.MigrateDown(ctx, cfg.Database.URL, *migrationsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Database rollback applied successfully")
		case "status":
			if err := database.MigrateStatus(ctx, cfg.Database.URL, *migrationsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Status check failed: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown migration direction: %s (must be 'up', 'down', or 'status')\n", *direction)
			os.Exit(1)
		}

	case "seed":
		fmt.Println("Database seed requested: reference data seeds will execute in Phase 1")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Saudi ERP CLI Utility")
	fmt.Println("Usage: erp-cli <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  version     Show application and CLI version")
	fmt.Println("  migrate     Run database migrations")
	fmt.Println("  seed        Seed reference database records")
}
