// erp-migrate applies Goose migrations without requiring API runtime secrets.
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
	if err := migrationcli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
