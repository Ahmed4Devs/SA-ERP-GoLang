package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"erp/internal/openapi"
)

func main() {
	output := flag.String("output", "api/openapi/openapi.yaml", "generated OpenAPI output path")
	flag.Parse()

	workingDirectory, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	contract, err := openapi.Generate(workingDirectory)
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, contract, 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
