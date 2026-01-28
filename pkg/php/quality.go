package php

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// FormatOptions configures PHP code formatting.
type FormatOptions struct {
	// Dir is the project directory (defaults to current working directory).
	Dir string

	// Fix automatically fixes formatting issues.
	Fix bool

	// Diff shows a diff of changes instead of modifying files.
	Diff bool

	// Paths limits formatting to specific paths.
	Paths []string

	// Output is the writer for output (defaults to os.Stdout).
	Output io.Writer
}

// AnalyseOptions configures PHP static analysis.
type AnalyseOptions struct {
	// Dir is the project directory (defaults to current working directory).
	Dir string

	// Level is the PHPStan analysis level (0-9).
	Level int

	// Paths limits analysis to specific paths.
	Paths []string

	// Memory is the memory limit for analysis (e.g., "2G").
	Memory string

	// Output is the writer for output (defaults to os.Stdout).
	Output io.Writer
}

// FormatterType represents the detected formatter.
type FormatterType string

const (
	FormatterPint FormatterType = "pint"
)

// AnalyserType represents the detected static analyser.
type AnalyserType string

const (
	AnalyserPHPStan   AnalyserType = "phpstan"
	AnalyserLarastan  AnalyserType = "larastan"
)

// DetectFormatter detects which formatter is available in the project.
func DetectFormatter(dir string) (FormatterType, bool) {
	// Check for Pint config
	pintConfig := filepath.Join(dir, "pint.json")
	if _, err := os.Stat(pintConfig); err == nil {
		return FormatterPint, true
	}

	// Check for vendor binary
	pintBin := filepath.Join(dir, "vendor", "bin", "pint")
	if _, err := os.Stat(pintBin); err == nil {
		return FormatterPint, true
	}

	return "", false
}

// DetectAnalyser detects which static analyser is available in the project.
func DetectAnalyser(dir string) (AnalyserType, bool) {
	// Check for PHPStan config
	phpstanConfig := filepath.Join(dir, "phpstan.neon")
	phpstanDistConfig := filepath.Join(dir, "phpstan.neon.dist")

	hasConfig := false
	if _, err := os.Stat(phpstanConfig); err == nil {
		hasConfig = true
	}
	if _, err := os.Stat(phpstanDistConfig); err == nil {
		hasConfig = true
	}

	// Check for vendor binary
	phpstanBin := filepath.Join(dir, "vendor", "bin", "phpstan")
	hasBin := false
	if _, err := os.Stat(phpstanBin); err == nil {
		hasBin = true
	}

	if hasConfig || hasBin {
		// Check if it's Larastan (Laravel-specific PHPStan)
		larastanPath := filepath.Join(dir, "vendor", "larastan", "larastan")
		if _, err := os.Stat(larastanPath); err == nil {
			return AnalyserLarastan, true
		}
		// Also check nunomaduro/larastan
		larastanPath2 := filepath.Join(dir, "vendor", "nunomaduro", "larastan")
		if _, err := os.Stat(larastanPath2); err == nil {
			return AnalyserLarastan, true
		}
		return AnalyserPHPStan, true
	}

	return "", false
}

// Format runs Laravel Pint to format PHP code.
func Format(ctx context.Context, opts FormatOptions) error {
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

	// Check if formatter is available
	formatter, found := DetectFormatter(opts.Dir)
	if !found {
		return fmt.Errorf("no formatter found (install Laravel Pint: composer require laravel/pint --dev)")
	}

	var cmdName string
	var args []string

	switch formatter {
	case FormatterPint:
		cmdName, args = buildPintCommand(opts)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// Analyse runs PHPStan or Larastan for static analysis.
func Analyse(ctx context.Context, opts AnalyseOptions) error {
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

	// Check if analyser is available
	analyser, found := DetectAnalyser(opts.Dir)
	if !found {
		return fmt.Errorf("no static analyser found (install PHPStan: composer require phpstan/phpstan --dev)")
	}

	var cmdName string
	var args []string

	switch analyser {
	case AnalyserPHPStan, AnalyserLarastan:
		cmdName, args = buildPHPStanCommand(opts)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// buildPintCommand builds the command for running Laravel Pint.
func buildPintCommand(opts FormatOptions) (string, []string) {
	// Check for vendor binary first
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "pint")
	cmdName := "pint"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	var args []string

	if !opts.Fix {
		args = append(args, "--test")
	}

	if opts.Diff {
		args = append(args, "--diff")
	}

	// Add specific paths if provided
	args = append(args, opts.Paths...)

	return cmdName, args
}

// buildPHPStanCommand builds the command for running PHPStan.
func buildPHPStanCommand(opts AnalyseOptions) (string, []string) {
	// Check for vendor binary first
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "phpstan")
	cmdName := "phpstan"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	args := []string{"analyse"}

	if opts.Level > 0 {
		args = append(args, "--level", fmt.Sprintf("%d", opts.Level))
	}

	if opts.Memory != "" {
		args = append(args, "--memory-limit", opts.Memory)
	}

	// Add specific paths if provided
	args = append(args, opts.Paths...)

	return cmdName, args
}
