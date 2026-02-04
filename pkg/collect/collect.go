// Package collect provides a data collection subsystem for gathering information
// from multiple sources including GitHub, BitcoinTalk, CoinGecko, and academic
// paper repositories. It supports rate limiting, incremental state tracking,
// and event-driven progress reporting.
package collect

import (
	"context"
	"path/filepath"

	"github.com/host-uk/core/pkg/io"
)

// Collector is the interface all collection sources implement.
type Collector interface {
	// Name returns a human-readable name for this collector.
	Name() string

	// Collect gathers data from the source and writes it to the configured output.
	Collect(ctx context.Context, cfg *Config) (*Result, error)
}

// Config holds shared configuration for all collectors.
type Config struct {
	// Output is the storage medium for writing collected data.
	Output io.Medium

	// OutputDir is the base directory for all collected data.
	OutputDir string

	// Limiter provides per-source rate limiting.
	Limiter *RateLimiter

	// State tracks collection progress for incremental runs.
	State *State

	// Dispatcher manages event dispatch for progress reporting.
	Dispatcher *Dispatcher

	// Verbose enables detailed logging output.
	Verbose bool

	// DryRun simulates collection without writing files.
	DryRun bool
}

// Result holds the output of a collection run.
type Result struct {
	// Source identifies which collector produced this result.
	Source string

	// Items is the number of items successfully collected.
	Items int

	// Errors is the number of errors encountered during collection.
	Errors int

	// Skipped is the number of items skipped (e.g. already collected).
	Skipped int

	// Files lists the paths of all files written.
	Files []string
}

// NewConfig creates a Config with sensible defaults.
// It initialises a MockMedium for output if none is provided,
// sets up a rate limiter, state tracker, and event dispatcher.
func NewConfig(outputDir string) *Config {
	m := io.NewMockMedium()
	return &Config{
		Output:     m,
		OutputDir:  outputDir,
		Limiter:    NewRateLimiter(),
		State:      NewState(m, filepath.Join(outputDir, ".collect-state.json")),
		Dispatcher: NewDispatcher(),
	}
}

// NewConfigWithMedium creates a Config using the specified storage medium.
func NewConfigWithMedium(m io.Medium, outputDir string) *Config {
	return &Config{
		Output:     m,
		OutputDir:  outputDir,
		Limiter:    NewRateLimiter(),
		State:      NewState(m, filepath.Join(outputDir, ".collect-state.json")),
		Dispatcher: NewDispatcher(),
	}
}

// MergeResults combines multiple results into a single aggregated result.
func MergeResults(source string, results ...*Result) *Result {
	merged := &Result{Source: source}
	for _, r := range results {
		if r == nil {
			continue
		}
		merged.Items += r.Items
		merged.Errors += r.Errors
		merged.Skipped += r.Skipped
		merged.Files = append(merged.Files, r.Files...)
	}
	return merged
}
