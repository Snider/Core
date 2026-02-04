package collect

import (
	"context"
	"fmt"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
)

// Excavator runs multiple collectors as a coordinated operation.
// It provides sequential execution with rate limit respect, state tracking
// for resume support, and aggregated results.
type Excavator struct {
	// Collectors is the list of collectors to run.
	Collectors []Collector

	// ScanOnly reports what would be collected without performing collection.
	ScanOnly bool

	// Resume enables incremental collection using saved state.
	Resume bool
}

// Name returns the orchestrator name.
func (e *Excavator) Name() string {
	return "excavator"
}

// Run executes all collectors sequentially, respecting rate limits and
// using state for resume support. Results are aggregated from all collectors.
func (e *Excavator) Run(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: e.Name()}

	if len(e.Collectors) == 0 {
		return result, nil
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(e.Name(), fmt.Sprintf("Starting excavation with %d collectors", len(e.Collectors)))
	}

	// Load state if resuming
	if e.Resume && cfg.State != nil {
		if err := cfg.State.Load(); err != nil {
			return result, core.E("collect.Excavator.Run", "failed to load state", err)
		}
	}

	// If scan-only, just report what would be collected
	if e.ScanOnly {
		for _, c := range e.Collectors {
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitProgress(e.Name(), fmt.Sprintf("[scan] Would run collector: %s", c.Name()), nil)
			}
		}
		return result, nil
	}

	for i, c := range e.Collectors {
		if ctx.Err() != nil {
			return result, core.E("collect.Excavator.Run", "context cancelled", ctx.Err())
		}

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(e.Name(),
				fmt.Sprintf("Running collector %d/%d: %s", i+1, len(e.Collectors), c.Name()), nil)
		}

		// Check if we should skip (already completed in a previous run)
		if e.Resume && cfg.State != nil {
			if entry, ok := cfg.State.Get(c.Name()); ok {
				if entry.Items > 0 && !entry.LastRun.IsZero() {
					if cfg.Dispatcher != nil {
						cfg.Dispatcher.EmitProgress(e.Name(),
							fmt.Sprintf("Skipping %s (already collected %d items on %s)",
								c.Name(), entry.Items, entry.LastRun.Format(time.RFC3339)), nil)
					}
					result.Skipped++
					continue
				}
			}
		}

		collectorResult, err := c.Collect(ctx, cfg)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(e.Name(),
					fmt.Sprintf("Collector %s failed: %v", c.Name(), err), nil)
			}
			continue
		}

		if collectorResult != nil {
			result.Items += collectorResult.Items
			result.Errors += collectorResult.Errors
			result.Skipped += collectorResult.Skipped
			result.Files = append(result.Files, collectorResult.Files...)

			// Update state
			if cfg.State != nil {
				cfg.State.Set(c.Name(), &StateEntry{
					Source:  c.Name(),
					LastRun: time.Now(),
					Items:   collectorResult.Items,
				})
			}
		}
	}

	// Save state
	if cfg.State != nil {
		if err := cfg.State.Save(); err != nil {
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(e.Name(), fmt.Sprintf("Failed to save state: %v", err), nil)
			}
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(e.Name(),
			fmt.Sprintf("Excavation complete: %d items, %d errors, %d skipped",
				result.Items, result.Errors, result.Skipped), result)
	}

	return result, nil
}
