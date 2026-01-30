package php

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/i18n"
	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/host-uk/core/pkg/process"
)

// QARunner orchestrates PHP QA checks using pkg/process.
type QARunner struct {
	dir     string
	fix     bool
	service *process.Service
	core    *framework.Core

	// Output tracking
	outputMu     sync.Mutex
	checkOutputs map[string][]string
}

// NewQARunner creates a QA runner for the given directory.
func NewQARunner(dir string, fix bool) (*QARunner, error) {
	// Create a Core with process service for the QA session
	core, err := framework.New(
		framework.WithName("process", process.NewService(process.Options{})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create process service: %w", err)
	}

	svc, err := framework.ServiceFor[*process.Service](core, "process")
	if err != nil {
		return nil, fmt.Errorf("failed to get process service: %w", err)
	}

	runner := &QARunner{
		dir:          dir,
		fix:          fix,
		service:      svc,
		core:         core,
		checkOutputs: make(map[string][]string),
	}

	return runner, nil
}

// BuildSpecs creates RunSpecs for the given QA checks.
func (r *QARunner) BuildSpecs(checks []string) []process.RunSpec {
	specs := make([]process.RunSpec, 0, len(checks))

	for _, check := range checks {
		spec := r.buildSpec(check)
		if spec != nil {
			specs = append(specs, *spec)
		}
	}

	return specs
}

// buildSpec creates a RunSpec for a single check.
func (r *QARunner) buildSpec(check string) *process.RunSpec {
	switch check {
	case "audit":
		return &process.RunSpec{
			Name:    "audit",
			Command: "composer",
			Args:    []string{"audit", "--format=summary"},
			Dir:     r.dir,
		}

	case "fmt":
		formatter, found := phppkg.DetectFormatter(r.dir)
		if !found {
			return nil
		}
		if formatter == phppkg.FormatterPint {
			vendorBin := filepath.Join(r.dir, "vendor", "bin", "pint")
			cmd := "pint"
			if _, err := os.Stat(vendorBin); err == nil {
				cmd = vendorBin
			}
			args := []string{}
			if !r.fix {
				args = append(args, "--test")
			}
			return &process.RunSpec{
				Name:    "fmt",
				Command: cmd,
				Args:    args,
				Dir:     r.dir,
				After:   []string{"audit"},
			}
		}
		return nil

	case "stan":
		_, found := phppkg.DetectAnalyser(r.dir)
		if !found {
			return nil
		}
		vendorBin := filepath.Join(r.dir, "vendor", "bin", "phpstan")
		cmd := "phpstan"
		if _, err := os.Stat(vendorBin); err == nil {
			cmd = vendorBin
		}
		return &process.RunSpec{
			Name:    "stan",
			Command: cmd,
			Args:    []string{"analyse", "--no-progress"},
			Dir:     r.dir,
			After:   []string{"fmt"},
		}

	case "psalm":
		_, found := phppkg.DetectPsalm(r.dir)
		if !found {
			return nil
		}
		vendorBin := filepath.Join(r.dir, "vendor", "bin", "psalm")
		cmd := "psalm"
		if _, err := os.Stat(vendorBin); err == nil {
			cmd = vendorBin
		}
		args := []string{"--no-progress"}
		if r.fix {
			args = append(args, "--alter", "--issues=all")
		}
		return &process.RunSpec{
			Name:    "psalm",
			Command: cmd,
			Args:    args,
			Dir:     r.dir,
			After:   []string{"stan"},
		}

	case "test":
		// Check for Pest first, fall back to PHPUnit
		pestBin := filepath.Join(r.dir, "vendor", "bin", "pest")
		phpunitBin := filepath.Join(r.dir, "vendor", "bin", "phpunit")

		cmd := "pest"
		if _, err := os.Stat(pestBin); err == nil {
			cmd = pestBin
		} else if _, err := os.Stat(phpunitBin); err == nil {
			cmd = phpunitBin
		} else {
			return nil
		}

		// Tests depend on stan (or psalm if available)
		after := []string{"stan"}
		if _, found := phppkg.DetectPsalm(r.dir); found {
			after = []string{"psalm"}
		}

		return &process.RunSpec{
			Name:    "test",
			Command: cmd,
			Args:    []string{},
			Dir:     r.dir,
			After:   after,
		}

	case "rector":
		if !phppkg.DetectRector(r.dir) {
			return nil
		}
		vendorBin := filepath.Join(r.dir, "vendor", "bin", "rector")
		cmd := "rector"
		if _, err := os.Stat(vendorBin); err == nil {
			cmd = vendorBin
		}
		args := []string{"process"}
		if !r.fix {
			args = append(args, "--dry-run")
		}
		return &process.RunSpec{
			Name:         "rector",
			Command:      cmd,
			Args:         args,
			Dir:          r.dir,
			After:        []string{"test"},
			AllowFailure: true, // Dry-run returns non-zero if changes would be made
		}

	case "infection":
		if !phppkg.DetectInfection(r.dir) {
			return nil
		}
		vendorBin := filepath.Join(r.dir, "vendor", "bin", "infection")
		cmd := "infection"
		if _, err := os.Stat(vendorBin); err == nil {
			cmd = vendorBin
		}
		return &process.RunSpec{
			Name:         "infection",
			Command:      cmd,
			Args:         []string{"--min-msi=50", "--min-covered-msi=70", "--threads=4"},
			Dir:          r.dir,
			After:        []string{"test"},
			AllowFailure: true,
		}
	}

	return nil
}

// Run executes all QA checks and returns the results.
func (r *QARunner) Run(ctx context.Context, stages []phppkg.QAStage) (*QARunResult, error) {
	// Collect all checks from all stages
	var allChecks []string
	for _, stage := range stages {
		checks := phppkg.GetQAChecks(r.dir, stage)
		allChecks = append(allChecks, checks...)
	}

	if len(allChecks) == 0 {
		return &QARunResult{Passed: true}, nil
	}

	// Build specs
	specs := r.BuildSpecs(allChecks)
	if len(specs) == 0 {
		return &QARunResult{Passed: true}, nil
	}

	// Register output handler
	r.core.RegisterAction(func(c *framework.Core, msg framework.Message) error {
		switch m := msg.(type) {
		case process.ActionProcessOutput:
			r.outputMu.Lock()
			// Extract check name from process ID mapping
			for _, spec := range specs {
				if strings.Contains(m.ID, spec.Name) || m.ID != "" {
					// Store output for later display if needed
					r.checkOutputs[spec.Name] = append(r.checkOutputs[spec.Name], m.Line)
					break
				}
			}
			r.outputMu.Unlock()
		}
		return nil
	})

	// Create runner and execute
	runner := process.NewRunner(r.service)
	result, err := runner.RunAll(ctx, specs)
	if err != nil {
		return nil, err
	}

	// Convert to QA result
	qaResult := &QARunResult{
		Passed:   result.Success(),
		Duration: result.Duration.String(),
		Results:  make([]QACheckRunResult, 0, len(result.Results)),
	}

	for _, res := range result.Results {
		qaResult.Results = append(qaResult.Results, QACheckRunResult{
			Name:     res.Name,
			Passed:   res.Passed(),
			Skipped:  res.Skipped,
			ExitCode: res.ExitCode,
			Duration: res.Duration.String(),
			Output:   res.Output,
		})
		if res.Passed() {
			qaResult.PassedCount++
		} else if res.Skipped {
			qaResult.SkippedCount++
		} else {
			qaResult.FailedCount++
		}
	}

	return qaResult, nil
}

// GetCheckOutput returns captured output for a check.
func (r *QARunner) GetCheckOutput(check string) []string {
	r.outputMu.Lock()
	defer r.outputMu.Unlock()
	return r.checkOutputs[check]
}

// QARunResult holds the results of running QA checks.
type QARunResult struct {
	Passed       bool
	Duration     string
	Results      []QACheckRunResult
	PassedCount  int
	FailedCount  int
	SkippedCount int
}

// QACheckRunResult holds the result of a single QA check.
type QACheckRunResult struct {
	Name     string
	Passed   bool
	Skipped  bool
	ExitCode int
	Duration string
	Output   string
}

// GetIssueMessage returns a localized issue message for a check.
func (r QACheckRunResult) GetIssueMessage() string {
	if r.Passed || r.Skipped {
		return ""
	}
	switch r.Name {
	case "audit":
		return i18n.T("cmd.php.qa.issue_audit")
	case "fmt":
		return i18n.T("cmd.php.qa.issue_style")
	case "stan":
		return i18n.T("cmd.php.qa.issue_analysis")
	case "psalm":
		return i18n.T("cmd.php.qa.issue_types")
	case "test":
		return i18n.T("cmd.php.qa.issue_tests")
	case "rector":
		return i18n.T("cmd.php.qa.issue_rector")
	case "infection":
		return i18n.T("cmd.php.qa.issue_mutation")
	default:
		return "issues found"
	}
}
