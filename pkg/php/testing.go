package php

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// TestOptions configures PHP test execution.
type TestOptions struct {
	// Dir is the project directory (defaults to current working directory).
	Dir string

	// Filter filters tests by name pattern.
	Filter string

	// Parallel runs tests in parallel.
	Parallel bool

	// Coverage generates code coverage.
	Coverage bool

	// CoverageFormat is the coverage output format (text, html, clover).
	CoverageFormat string

	// Groups runs only tests in the specified groups.
	Groups []string

	// Output is the writer for test output (defaults to os.Stdout).
	Output io.Writer
}

// TestRunner represents the detected test runner.
type TestRunner string

const (
	TestRunnerPest    TestRunner = "pest"
	TestRunnerPHPUnit TestRunner = "phpunit"
)

// DetectTestRunner detects which test runner is available in the project.
// Returns Pest if tests/Pest.php exists, otherwise PHPUnit.
func DetectTestRunner(dir string) TestRunner {
	// Check for Pest
	pestFile := filepath.Join(dir, "tests", "Pest.php")
	if _, err := os.Stat(pestFile); err == nil {
		return TestRunnerPest
	}

	return TestRunnerPHPUnit
}

// RunTests runs PHPUnit or Pest tests.
func RunTests(ctx context.Context, opts TestOptions) error {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Detect test runner
	runner := DetectTestRunner(opts.Dir)

	// Build command based on runner
	var cmdName string
	var args []string

	switch runner {
	case TestRunnerPest:
		cmdName, args = buildPestCommand(opts)
	default:
		cmdName, args = buildPHPUnitCommand(opts)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// RunParallel runs tests in parallel using the appropriate runner.
func RunParallel(ctx context.Context, opts TestOptions) error {
	opts.Parallel = true
	return RunTests(ctx, opts)
}

// buildPestCommand builds the command for running Pest tests.
func buildPestCommand(opts TestOptions) (string, []string) {
	// Check for vendor binary first
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "pest")
	cmdName := "pest"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	var args []string

	if opts.Filter != "" {
		args = append(args, "--filter", opts.Filter)
	}

	if opts.Parallel {
		args = append(args, "--parallel")
	}

	if opts.Coverage {
		switch opts.CoverageFormat {
		case "html":
			args = append(args, "--coverage-html", "coverage")
		case "clover":
			args = append(args, "--coverage-clover", "coverage.xml")
		default:
			args = append(args, "--coverage")
		}
	}

	for _, group := range opts.Groups {
		args = append(args, "--group", group)
	}

	return cmdName, args
}

// buildPHPUnitCommand builds the command for running PHPUnit tests.
func buildPHPUnitCommand(opts TestOptions) (string, []string) {
	// Check for vendor binary first
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "phpunit")
	cmdName := "phpunit"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	var args []string

	if opts.Filter != "" {
		args = append(args, "--filter", opts.Filter)
	}

	if opts.Parallel {
		// PHPUnit uses paratest for parallel execution
		paratestBin := filepath.Join(opts.Dir, "vendor", "bin", "paratest")
		if _, err := os.Stat(paratestBin); err == nil {
			cmdName = paratestBin
		}
	}

	if opts.Coverage {
		switch opts.CoverageFormat {
		case "html":
			args = append(args, "--coverage-html", "coverage")
		case "clover":
			args = append(args, "--coverage-clover", "coverage.xml")
		default:
			args = append(args, "--coverage-text")
		}
	}

	for _, group := range opts.Groups {
		args = append(args, "--group", group)
	}

	return cmdName, args
}
