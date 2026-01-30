package process

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Runner orchestrates multiple processes with dependencies.
type Runner struct {
	service *Service
}

// NewRunner creates a runner for the given service.
func NewRunner(svc *Service) *Runner {
	return &Runner{service: svc}
}

// RunSpec defines a process to run with optional dependencies.
type RunSpec struct {
	// Name is a friendly identifier (e.g., "lint", "test").
	Name string
	// Command is the executable to run.
	Command string
	// Args are the command arguments.
	Args []string
	// Dir is the working directory.
	Dir string
	// Env are additional environment variables.
	Env []string
	// After lists spec names that must complete successfully first.
	After []string
	// AllowFailure if true, continues pipeline even if this spec fails.
	AllowFailure bool
}

// RunResult captures the outcome of a single process.
type RunResult struct {
	Name     string
	Spec     RunSpec
	ExitCode int
	Duration time.Duration
	Output   string
	Error    error
	Skipped  bool
}

// Passed returns true if the process succeeded.
func (r RunResult) Passed() bool {
	return !r.Skipped && r.Error == nil && r.ExitCode == 0
}

// RunAllResult is the aggregate result of running multiple specs.
type RunAllResult struct {
	Results  []RunResult
	Duration time.Duration
	Passed   int
	Failed   int
	Skipped  int
}

// Success returns true if all non-skipped specs passed.
func (r RunAllResult) Success() bool {
	return r.Failed == 0
}

// RunAll executes specs respecting dependencies, parallelising where possible.
func (r *Runner) RunAll(ctx context.Context, specs []RunSpec) (*RunAllResult, error) {
	start := time.Now()

	// Build dependency graph
	specMap := make(map[string]RunSpec)
	for _, spec := range specs {
		specMap[spec.Name] = spec
	}

	// Track completion
	completed := make(map[string]*RunResult)
	var completedMu sync.Mutex

	results := make([]RunResult, 0, len(specs))
	var resultsMu sync.Mutex

	// Process specs in waves
	remaining := make(map[string]RunSpec)
	for _, spec := range specs {
		remaining[spec.Name] = spec
	}

	for len(remaining) > 0 {
		// Find specs ready to run (all dependencies satisfied)
		ready := make([]RunSpec, 0)
		for _, spec := range remaining {
			if r.canRun(spec, completed) {
				ready = append(ready, spec)
			}
		}

		if len(ready) == 0 && len(remaining) > 0 {
			// Deadlock - circular dependency or missing specs
			for name := range remaining {
				results = append(results, RunResult{
					Name:    name,
					Spec:    remaining[name],
					Skipped: true,
					Error:   fmt.Errorf("circular dependency or missing dependency"),
				})
			}
			break
		}

		// Run ready specs in parallel
		var wg sync.WaitGroup
		for _, spec := range ready {
			wg.Add(1)
			go func(spec RunSpec) {
				defer wg.Done()

				// Check if dependencies failed
				completedMu.Lock()
				shouldSkip := false
				for _, dep := range spec.After {
					if result, ok := completed[dep]; ok {
						if !result.Passed() && !specMap[dep].AllowFailure {
							shouldSkip = true
							break
						}
					}
				}
				completedMu.Unlock()

				var result RunResult
				if shouldSkip {
					result = RunResult{
						Name:    spec.Name,
						Spec:    spec,
						Skipped: true,
						Error:   fmt.Errorf("skipped due to dependency failure"),
					}
				} else {
					result = r.runSpec(ctx, spec)
				}

				completedMu.Lock()
				completed[spec.Name] = &result
				completedMu.Unlock()

				resultsMu.Lock()
				results = append(results, result)
				resultsMu.Unlock()
			}(spec)
		}
		wg.Wait()

		// Remove completed from remaining
		for _, spec := range ready {
			delete(remaining, spec.Name)
		}
	}

	// Build aggregate result
	aggResult := &RunAllResult{
		Results:  results,
		Duration: time.Since(start),
	}

	for _, res := range results {
		if res.Skipped {
			aggResult.Skipped++
		} else if res.Passed() {
			aggResult.Passed++
		} else {
			aggResult.Failed++
		}
	}

	return aggResult, nil
}

// canRun checks if all dependencies are completed.
func (r *Runner) canRun(spec RunSpec, completed map[string]*RunResult) bool {
	for _, dep := range spec.After {
		if _, ok := completed[dep]; !ok {
			return false
		}
	}
	return true
}

// runSpec executes a single spec.
func (r *Runner) runSpec(ctx context.Context, spec RunSpec) RunResult {
	start := time.Now()

	proc, err := r.service.StartWithOptions(ctx, RunOptions{
		Command: spec.Command,
		Args:    spec.Args,
		Dir:     spec.Dir,
		Env:     spec.Env,
	})
	if err != nil {
		return RunResult{
			Name:     spec.Name,
			Spec:     spec,
			Duration: time.Since(start),
			Error:    err,
		}
	}

	<-proc.Done()

	return RunResult{
		Name:     spec.Name,
		Spec:     spec,
		ExitCode: proc.ExitCode,
		Duration: proc.Duration,
		Output:   proc.Output(),
		Error:    nil,
	}
}

// RunSequential executes specs one after another, stopping on first failure.
func (r *Runner) RunSequential(ctx context.Context, specs []RunSpec) (*RunAllResult, error) {
	start := time.Now()
	results := make([]RunResult, 0, len(specs))

	for _, spec := range specs {
		result := r.runSpec(ctx, spec)
		results = append(results, result)

		if !result.Passed() && !spec.AllowFailure {
			// Mark remaining as skipped
			for i := len(results); i < len(specs); i++ {
				results = append(results, RunResult{
					Name:    specs[i].Name,
					Spec:    specs[i],
					Skipped: true,
				})
			}
			break
		}
	}

	aggResult := &RunAllResult{
		Results:  results,
		Duration: time.Since(start),
	}

	for _, res := range results {
		if res.Skipped {
			aggResult.Skipped++
		} else if res.Passed() {
			aggResult.Passed++
		} else {
			aggResult.Failed++
		}
	}

	return aggResult, nil
}

// RunParallel executes all specs concurrently, regardless of dependencies.
func (r *Runner) RunParallel(ctx context.Context, specs []RunSpec) (*RunAllResult, error) {
	start := time.Now()
	results := make([]RunResult, len(specs))

	var wg sync.WaitGroup
	for i, spec := range specs {
		wg.Add(1)
		go func(i int, spec RunSpec) {
			defer wg.Done()
			results[i] = r.runSpec(ctx, spec)
		}(i, spec)
	}
	wg.Wait()

	aggResult := &RunAllResult{
		Results:  results,
		Duration: time.Since(start),
	}

	for _, res := range results {
		if res.Skipped {
			aggResult.Skipped++
		} else if res.Passed() {
			aggResult.Passed++
		} else {
			aggResult.Failed++
		}
	}

	return aggResult, nil
}
