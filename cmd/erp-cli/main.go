package main

import (
	"flag"
	"fmt"
	"os"

	"erp/internal/platform/config"
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
			fmt.Printf("Saudi ERP CLI v0.1.0\n")
			return
		}
		fmt.Printf("Saudi ERP Platform v%s (%s)\n", cfg.AppVersion, cfg.Environment)

	case "migrate":
		migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
		direction := migrateCmd.String("dir", "up", "Migration direction: up or down")
		_ = migrateCmd.Parse(os.Args[2:])
		fmt.Printf("Database migration requested: direction=%s (Goose engine will be wired in Phase 1)\n", *direction)

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
