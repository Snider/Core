package main

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/cli"

	// Build variants import commands via self-registration.
	// See internal/variants/ for available variants: full, ci, php, minimal.
	_ "github.com/host-uk/core/internal/variants"
)

func main() {
	if err := cli.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
