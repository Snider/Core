package gocmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/qa"
)

// QA command flags - comprehensive options for all agents
var (
	qaFix               bool
	qaChanged           bool
	qaAll               bool
	qaSkip              string
	qaOnly              string
	qaCoverage          bool
	qaThreshold         float64
	qaDocblockThreshold float64
	qaJSON              bool
	qaVerbose           bool
	qaQuiet             bool
	qaTimeout           time.Duration
	qaShort             bool
	qaRace              bool
	qaBench             bool
	qaFailFast          bool
	qaMod               bool
	qaCI                bool
)

func addGoQACommand(parent *cli.Command) {
	qaCmd := &cli.Command{
		Use:   "qa",
		Short: "Run QA checks",
		Long: `Run comprehensive code quality checks for Go projects.

Checks available: fmt, vet, lint, test, race, vuln, sec, bench, docblock

Examples:
  core go qa                    # Default: fmt, lint, test
  core go qa --fix              # Auto-fix formatting and lint issues
  core go qa --only=test        # Only run tests
  core go qa --skip=vuln,sec    # Skip vulnerability and security scans
  core go qa --coverage --threshold=80  # Require 80% coverage
  core go qa --changed          # Only check changed files (git-aware)
  core go qa --ci               # CI mode: strict, coverage, fail-fast
  core go qa --race --short     # Quick tests with race detection
  core go qa --json             # Output results as JSON`,
		RunE: runGoQA,
	}

	// Fix and modification flags (persistent so subcommands inherit them)
	qaCmd.PersistentFlags().BoolVar(&qaFix, "fix", false, "Auto-fix issues where possible")
	qaCmd.PersistentFlags().BoolVar(&qaMod, "mod", false, "Run go mod tidy before checks")

	// Scope flags
	qaCmd.PersistentFlags().BoolVar(&qaChanged, "changed", false, "Only check changed files (git-aware)")
	qaCmd.PersistentFlags().BoolVar(&qaAll, "all", false, "Check all files (override git-aware)")
	qaCmd.PersistentFlags().StringVar(&qaSkip, "skip", "", "Skip checks (comma-separated: fmt,vet,lint,test,race,vuln,sec,bench)")
	qaCmd.PersistentFlags().StringVar(&qaOnly, "only", "", "Only run these checks (comma-separated)")

	// Coverage flags
	qaCmd.PersistentFlags().BoolVar(&qaCoverage, "coverage", false, "Include coverage reporting")
	qaCmd.PersistentFlags().BoolVarP(&qaCoverage, "cov", "c", false, "Include coverage reporting (shorthand)")
	qaCmd.PersistentFlags().Float64Var(&qaThreshold, "threshold", 0, "Minimum coverage threshold (0-100), fail if below")
	qaCmd.PersistentFlags().Float64Var(&qaDocblockThreshold, "docblock-threshold", 80, "Minimum docblock coverage threshold (0-100)")

	// Test flags
	qaCmd.PersistentFlags().BoolVar(&qaShort, "short", false, "Run tests with -short flag")
	qaCmd.PersistentFlags().BoolVar(&qaRace, "race", false, "Include race detection in tests")
	qaCmd.PersistentFlags().BoolVar(&qaBench, "bench", false, "Include benchmarks")

	// Output flags
	qaCmd.PersistentFlags().BoolVar(&qaJSON, "json", false, "Output results as JSON")
	qaCmd.PersistentFlags().BoolVarP(&qaVerbose, "verbose", "v", false, "Show verbose output")
	qaCmd.PersistentFlags().BoolVarP(&qaQuiet, "quiet", "q", false, "Only show errors")

	// Control flags
	qaCmd.PersistentFlags().DurationVar(&qaTimeout, "timeout", 10*time.Minute, "Timeout for all checks")
	qaCmd.PersistentFlags().BoolVar(&qaFailFast, "fail-fast", false, "Stop on first failure")
	qaCmd.PersistentFlags().BoolVar(&qaCI, "ci", false, "CI mode: strict checks, coverage required, fail-fast")

	// Preset subcommands for convenience
	qaCmd.AddCommand(&cli.Command{
		Use:   "quick",
		Short: "Quick QA: fmt, vet, lint (no tests)",
		RunE:  func(cmd *cli.Command, args []string) error { qaOnly = "fmt,vet,lint"; return runGoQA(cmd, args) },
	})

	qaCmd.AddCommand(&cli.Command{
		Use:   "full",
		Short: "Full QA: all checks including race, vuln, sec",
		RunE: func(cmd *cli.Command, args []string) error {
			qaOnly = "fmt,vet,lint,test,race,vuln,sec"
			return runGoQA(cmd, args)
		},
	})

	qaCmd.AddCommand(&cli.Command{
		Use:   "pre-commit",
		Short: "Pre-commit checks: fmt --fix, lint --fix, test --short",
		RunE: func(cmd *cli.Command, args []string) error {
			qaFix = true
			qaShort = true
			qaOnly = "fmt,lint,test"
			return runGoQA(cmd, args)
		},
	})

	qaCmd.AddCommand(&cli.Command{
		Use:   "pr",
		Short: "PR checks: full QA with coverage threshold",
		RunE: func(cmd *cli.Command, args []string) error {
			qaCoverage = true
			if qaThreshold == 0 {
				qaThreshold = 50 // Default PR threshold
			}
			qaOnly = "fmt,vet,lint,test"
			return runGoQA(cmd, args)
		},
	})

	parent.AddCommand(qaCmd)
}

// QAResult holds the result of a QA run for JSON output
type QAResult struct {
	Success   bool          `json:"success"`
	Duration  string        `json:"duration"`
	Checks    []CheckResult `json:"checks"`
	Coverage  *float64      `json:"coverage,omitempty"`
	Threshold *float64      `json:"threshold,omitempty"`
}

// CheckResult holds the result of a single check
type CheckResult struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
	Output   string `json:"output,omitempty"`
}

func runGoQA(cmd *cli.Command, args []string) error {
	// Apply CI mode defaults
	if qaCI {
		qaCoverage = true
		qaFailFast = true
		if qaThreshold == 0 {
			qaThreshold = 50
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return cli.Wrap(err, i18n.T("i18n.fail.get", "working directory"))
	}

	// Detect if this is a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return cli.Err("not a Go project (no go.mod found)")
	}

	// Determine which checks to run
	checkNames := determineChecks()

	if !qaJSON && !qaQuiet {
		cli.Print("%s %s\n\n", cli.DimStyle.Render(i18n.Label("qa")), i18n.ProgressSubject("run", "Go QA"))
	}

	// Run go mod tidy if requested
	if qaMod {
		if !qaQuiet {
			cli.Print("%s %s\n", cli.DimStyle.Render("→"), "Running go mod tidy...")
		}
		modCmd := exec.Command("go", "mod", "tidy")
		modCmd.Dir = cwd
		if err := modCmd.Run(); err != nil {
			return cli.Wrap(err, "go mod tidy failed")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), qaTimeout)
	defer cancel()

	startTime := time.Now()
	checks := buildChecks(checkNames)
	results := make([]CheckResult, 0, len(checks))
	passed := 0
	failed := 0

	for _, check := range checks {
		checkStart := time.Now()

		if !qaJSON && !qaQuiet {
			cli.Print("%s %s\n", cli.DimStyle.Render("→"), i18n.Progress(check.Name))
		}

		output, err := runCheckCapture(ctx, cwd, check)
		checkDuration := time.Since(checkStart)

		result := CheckResult{
			Name:     check.Name,
			Duration: checkDuration.Round(time.Millisecond).String(),
		}

		if err != nil {
			result.Passed = false
			result.Error = err.Error()
			if qaVerbose {
				result.Output = output
			}
			failed++

			if !qaJSON && !qaQuiet {
				cli.Print("  %s %s\n", cli.ErrorStyle.Render(cli.Glyph(":cross:")), err.Error())
				if qaVerbose && output != "" {
					cli.Text(output)
				}
			}

			if qaFailFast {
				results = append(results, result)
				break
			}
		} else {
			result.Passed = true
			if qaVerbose {
				result.Output = output
			}
			passed++

			if !qaJSON && !qaQuiet {
				cli.Print("  %s %s\n", cli.SuccessStyle.Render(cli.Glyph(":check:")), i18n.T("i18n.done.pass"))
			}
		}

		results = append(results, result)
	}

	// Run coverage if requested
	var coverageVal *float64
	if qaCoverage && !qaFailFast || (qaCoverage && failed == 0) {
		cov, err := runCoverage(ctx, cwd)
		if err == nil {
			coverageVal = &cov
			if !qaJSON && !qaQuiet {
				cli.Print("\n%s %.1f%%\n", cli.DimStyle.Render("Coverage:"), cov)
			}
			if qaThreshold > 0 && cov < qaThreshold {
				failed++
				if !qaJSON && !qaQuiet {
					cli.Print("  %s Coverage %.1f%% below threshold %.1f%%\n",
						cli.ErrorStyle.Render(cli.Glyph(":cross:")), cov, qaThreshold)
				}
			}
		}
	}

	duration := time.Since(startTime).Round(time.Millisecond)

	// JSON output
	if qaJSON {
		qaResult := QAResult{
			Success:  failed == 0,
			Duration: duration.String(),
			Checks:   results,
			Coverage: coverageVal,
		}
		if qaThreshold > 0 {
			qaResult.Threshold = &qaThreshold
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(qaResult)
	}

	// Summary
	if !qaQuiet {
		cli.Blank()
		if failed > 0 {
			cli.Print("%s %s, %s (%s)\n",
				cli.ErrorStyle.Render(cli.Glyph(":cross:")),
				i18n.T("i18n.count.check", passed)+" "+i18n.T("i18n.done.pass"),
				i18n.T("i18n.count.check", failed)+" "+i18n.T("i18n.done.fail"),
				duration)
		} else {
			cli.Print("%s %s (%s)\n",
				cli.SuccessStyle.Render(cli.Glyph(":check:")),
				i18n.T("i18n.count.check", passed)+" "+i18n.T("i18n.done.pass"),
				duration)
		}
	}

	if failed > 0 {
		os.Exit(1)
	}
	return nil
}

func determineChecks() []string {
	// If --only is specified, use those
	if qaOnly != "" {
		return strings.Split(qaOnly, ",")
	}

	// Default checks
	checks := []string{"fmt", "lint", "test", "docblock"}

	// Add race if requested
	if qaRace {
		// Replace test with race (which includes test)
		for i, c := range checks {
			if c == "test" {
				checks[i] = "race"
				break
			}
		}
	}

	// Add bench if requested
	if qaBench {
		checks = append(checks, "bench")
	}

	// Remove skipped checks
	if qaSkip != "" {
		skipMap := make(map[string]bool)
		for _, s := range strings.Split(qaSkip, ",") {
			skipMap[strings.TrimSpace(s)] = true
		}
		filtered := make([]string, 0, len(checks))
		for _, c := range checks {
			if !skipMap[c] {
				filtered = append(filtered, c)
			}
		}
		checks = filtered
	}

	return checks
}

// QACheck represents a single QA check.
type QACheck struct {
	Name    string
	Command string
	Args    []string
}

func buildChecks(names []string) []QACheck {
	var checks []QACheck
	for _, name := range names {
		name = strings.TrimSpace(name)
		check := buildCheck(name)
		if check.Command != "" {
			checks = append(checks, check)
		}
	}
	return checks
}

func buildCheck(name string) QACheck {
	switch name {
	case "fmt", "format":
		args := []string{"-l", "."}
		if qaFix {
			args = []string{"-w", "."}
		}
		return QACheck{Name: "format", Command: "gofmt", Args: args}

	case "vet":
		return QACheck{Name: "vet", Command: "go", Args: []string{"vet", "./..."}}

	case "lint":
		args := []string{"run"}
		if qaFix {
			args = append(args, "--fix")
		}
		if qaChanged && !qaAll {
			args = append(args, "--new-from-rev=HEAD")
		}
		args = append(args, "./...")
		return QACheck{Name: "lint", Command: "golangci-lint", Args: args}

	case "test":
		args := []string{"test"}
		if qaShort {
			args = append(args, "-short")
		}
		if qaVerbose {
			args = append(args, "-v")
		}
		args = append(args, "./...")
		return QACheck{Name: "test", Command: "go", Args: args}

	case "race":
		args := []string{"test", "-race"}
		if qaShort {
			args = append(args, "-short")
		}
		if qaVerbose {
			args = append(args, "-v")
		}
		args = append(args, "./...")
		return QACheck{Name: "race", Command: "go", Args: args}

	case "bench":
		args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}
		args = append(args, "./...")
		return QACheck{Name: "bench", Command: "go", Args: args}

	case "vuln":
		return QACheck{Name: "vuln", Command: "govulncheck", Args: []string{"./..."}}

	case "sec":
		return QACheck{Name: "sec", Command: "gosec", Args: []string{"-quiet", "./..."}}

	case "docblock":
		// Special internal check - handled separately
		return QACheck{Name: "docblock", Command: "_internal_"}

	default:
		return QACheck{}
	}
}

func runCheckCapture(ctx context.Context, dir string, check QACheck) (string, error) {
	// Handle internal checks
	if check.Command == "_internal_" {
		return runInternalCheck(check)
	}

	// Check if command exists
	if _, err := exec.LookPath(check.Command); err != nil {
		return "", cli.Err("%s: not installed", check.Command)
	}

	cmd := exec.CommandContext(ctx, check.Command, check.Args...)
	cmd.Dir = dir

	// For gofmt -l, capture output to check if files need formatting
	if check.Name == "format" && len(check.Args) > 0 && check.Args[0] == "-l" {
		output, err := cmd.Output()
		if err != nil {
			return string(output), err
		}
		if len(output) > 0 {
			// Show files that need formatting
			if !qaQuiet && !qaJSON {
				cli.Text(string(output))
			}
			return string(output), cli.Err("files need formatting (use --fix)")
		}
		return "", nil
	}

	// For other commands, stream or capture based on quiet mode
	if qaQuiet || qaJSON {
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return "", cmd.Run()
}

func runCoverage(ctx context.Context, dir string) (float64, error) {
	args := []string{"test", "-cover", "-coverprofile=/tmp/coverage.out"}
	if qaShort {
		args = append(args, "-short")
	}
	args = append(args, "./...")

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	if !qaQuiet && !qaJSON {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Parse coverage
	coverCmd := exec.CommandContext(ctx, "go", "tool", "cover", "-func=/tmp/coverage.out")
	output, err := coverCmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse last line for total coverage
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return 0, nil
	}

	lastLine := lines[len(lines)-1]
	fields := strings.Fields(lastLine)
	if len(fields) < 3 {
		return 0, nil
	}

	// Parse percentage (e.g., "45.6%")
	pctStr := strings.TrimSuffix(fields[len(fields)-1], "%")
	var pct float64
	if _, err := fmt.Sscanf(pctStr, "%f", &pct); err == nil {
		return pct, nil
	}

	return 0, nil
}

// runInternalCheck runs internal Go-based checks (not external commands).
func runInternalCheck(check QACheck) (string, error) {
	switch check.Name {
	case "docblock":
		result, err := qa.CheckDocblockCoverage([]string{"./..."})
		if err != nil {
			return "", err
		}
		result.Threshold = qaDocblockThreshold
		result.Passed = result.Coverage >= qaDocblockThreshold

		if !result.Passed {
			var output strings.Builder
			output.WriteString(fmt.Sprintf("Docblock coverage: %.1f%% (threshold: %.1f%%)\n",
				result.Coverage, qaDocblockThreshold))
			for _, m := range result.Missing {
				output.WriteString(fmt.Sprintf("%s:%d\n", m.File, m.Line))
			}
			return output.String(), cli.Err("docblock coverage %.1f%% below threshold %.1f%%",
				result.Coverage, qaDocblockThreshold)
		}
		return fmt.Sprintf("Docblock coverage: %.1f%%", result.Coverage), nil

	default:
		return "", cli.Err("unknown internal check: %s", check.Name)
	}
}
