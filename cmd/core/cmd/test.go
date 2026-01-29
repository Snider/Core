package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

// Test command styles
var (
	testHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3b82f6")) // blue-500

	testPassStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22c55e")). // green-500
			Bold(true)

	testFailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")). // red-500
			Bold(true)

	testSkipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	testDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")) // gray-500

	testCovHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")) // green-500

	testCovMedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	testCovLowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")) // red-500
)

// AddTestCommand adds the 'test' command to the given parent command.
func AddTestCommand(parent *clir.Cli) {
	var verbose bool
	var coverage bool
	var short bool
	var pkg string
	var run string
	var race bool
	var json bool

	testCmd := parent.NewSubCommand("test", "Run tests with coverage")
	testCmd.LongDescription("Runs Go tests with coverage reporting.\n\n" +
		"Sets MACOSX_DEPLOYMENT_TARGET=26.0 to suppress linker warnings on macOS.\n\n" +
		"Examples:\n" +
		"  core test                     # Run all tests with coverage summary\n" +
		"  core test --verbose           # Show test output as it runs\n" +
		"  core test --coverage          # Show detailed per-package coverage\n" +
		"  core test --pkg ./pkg/...     # Test specific packages\n" +
		"  core test --run TestName      # Run specific test by name\n" +
		"  core test --short             # Skip long-running tests\n" +
		"  core test --race              # Enable race detector\n" +
		"  core test --json              # Output JSON for CI/agents")

	testCmd.BoolFlag("verbose", "Show test output as it runs (-v)", &verbose)
	testCmd.BoolFlag("coverage", "Show detailed per-package coverage", &coverage)
	testCmd.BoolFlag("short", "Skip long-running tests (-short)", &short)
	testCmd.StringFlag("pkg", "Package pattern to test (default: ./...)", &pkg)
	testCmd.StringFlag("run", "Run only tests matching this regex (-run)", &run)
	testCmd.BoolFlag("race", "Enable race detector (-race)", &race)
	testCmd.BoolFlag("json", "Output JSON for CI/agents", &json)

	testCmd.Action(func() error {
		return runTest(verbose, coverage, short, pkg, run, race, json)
	})
}

type packageCoverage struct {
	name     string
	coverage float64
	hasCov   bool
}

func runTest(verbose, coverage, short bool, pkg, run string, race, jsonOutput bool) error {
	// Detect if we're in a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("no go.mod found - run from a Go project directory")
	}

	// Build command arguments
	args := []string{"test"}

	// Default to ./... if no package specified
	if pkg == "" {
		pkg = "./..."
	}

	// Add flags
	if verbose {
		args = append(args, "-v")
	}
	if short {
		args = append(args, "-short")
	}
	if run != "" {
		args = append(args, "-run", run)
	}
	if race {
		args = append(args, "-race")
	}

	// Always add coverage
	args = append(args, "-cover")

	// Add package pattern
	args = append(args, pkg)

	// Create command
	cmd := exec.Command("go", args...)
	cmd.Dir, _ = os.Getwd()

	// Set environment to suppress macOS linker warnings
	cmd.Env = append(os.Environ(), getMacOSDeploymentTarget())

	if !jsonOutput {
		fmt.Printf("%s Running tests\n", testHeaderStyle.Render("Test:"))
		fmt.Printf("  Package: %s\n", testDimStyle.Render(pkg))
		if run != "" {
			fmt.Printf("  Filter:  %s\n", testDimStyle.Render(run))
		}
		fmt.Println()
	}

	// Capture output for parsing
	var stdout, stderr strings.Builder

	if verbose && !jsonOutput {
		// Stream output in verbose mode, but also capture for parsing
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	} else {
		// Capture output for parsing
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	// Combine stdout and stderr for parsing, filtering linker warnings
	combined := filterLinkerWarnings(stdout.String() + "\n" + stderr.String())

	// Parse results
	results := parseTestOutput(combined)

	if jsonOutput {
		// JSON output for CI/agents
		printJSONResults(results, exitCode)
		if exitCode != 0 {
			return fmt.Errorf("tests failed")
		}
		return nil
	}

	// Print summary
	if !verbose {
		printTestSummary(results, coverage)
	} else if coverage {
		// In verbose mode, still show coverage summary at end
		fmt.Println()
		printCoverageSummary(results)
	}

	if exitCode != 0 {
		fmt.Printf("\n%s Tests failed\n", testFailStyle.Render("FAIL"))
		return fmt.Errorf("tests failed")
	}

	fmt.Printf("\n%s All tests passed\n", testPassStyle.Render("PASS"))
	return nil
}

func getMacOSDeploymentTarget() string {
	if runtime.GOOS == "darwin" {
		// Use deployment target matching current macOS to suppress linker warnings
		return "MACOSX_DEPLOYMENT_TARGET=26.0"
	}
	return ""
}

func filterLinkerWarnings(output string) string {
	// Filter out ld: warning lines that pollute the output
	var filtered []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		// Skip linker warnings
		if strings.HasPrefix(line, "ld: warning:") {
			continue
		}
		// Skip test binary build comments
		if strings.HasPrefix(line, "# ") && strings.HasSuffix(line, ".test") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

type testResults struct {
	packages   []packageCoverage
	passed     int
	failed     int
	skipped    int
	totalCov   float64
	covCount   int
	failedPkgs []string
}

func parseTestOutput(output string) testResults {
	results := testResults{}

	// Regex patterns - handle both timed and cached test results
	// Example: ok  	github.com/host-uk/core/pkg/crypt	0.015s	coverage: 91.2% of statements
	// Example: ok  	github.com/host-uk/core/pkg/crypt	(cached)	coverage: 91.2% of statements
	okPattern := regexp.MustCompile(`^ok\s+(\S+)\s+(?:[\d.]+s|\(cached\))(?:\s+coverage:\s+([\d.]+)%)?`)
	failPattern := regexp.MustCompile(`^FAIL\s+(\S+)`)
	skipPattern := regexp.MustCompile(`^\?\s+(\S+)\s+\[no test files\]`)
	coverPattern := regexp.MustCompile(`coverage:\s+([\d.]+)%`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := okPattern.FindStringSubmatch(line); matches != nil {
			pkg := packageCoverage{name: matches[1]}
			if len(matches) > 2 && matches[2] != "" {
				cov, _ := strconv.ParseFloat(matches[2], 64)
				pkg.coverage = cov
				pkg.hasCov = true
				results.totalCov += cov
				results.covCount++
			}
			results.packages = append(results.packages, pkg)
			results.passed++
		} else if matches := failPattern.FindStringSubmatch(line); matches != nil {
			results.failed++
			results.failedPkgs = append(results.failedPkgs, matches[1])
		} else if matches := skipPattern.FindStringSubmatch(line); matches != nil {
			results.skipped++
		} else if matches := coverPattern.FindStringSubmatch(line); matches != nil {
			// Catch any additional coverage lines
			cov, _ := strconv.ParseFloat(matches[1], 64)
			if cov > 0 {
				// Find the last package without coverage and update it
				for i := len(results.packages) - 1; i >= 0; i-- {
					if !results.packages[i].hasCov {
						results.packages[i].coverage = cov
						results.packages[i].hasCov = true
						results.totalCov += cov
						results.covCount++
						break
					}
				}
			}
		}
	}

	return results
}

func printTestSummary(results testResults, showCoverage bool) {
	// Print pass/fail summary
	total := results.passed + results.failed
	if total > 0 {
		fmt.Printf("  %s %d passed", testPassStyle.Render("✓"), results.passed)
		if results.failed > 0 {
			fmt.Printf("  %s %d failed", testFailStyle.Render("✗"), results.failed)
		}
		if results.skipped > 0 {
			fmt.Printf("  %s %d skipped", testSkipStyle.Render("○"), results.skipped)
		}
		fmt.Println()
	}

	// Print failed packages
	if len(results.failedPkgs) > 0 {
		fmt.Printf("\n  Failed packages:\n")
		for _, pkg := range results.failedPkgs {
			fmt.Printf("    %s %s\n", testFailStyle.Render("✗"), pkg)
		}
	}

	// Print coverage
	if showCoverage {
		printCoverageSummary(results)
	} else if results.covCount > 0 {
		avgCov := results.totalCov / float64(results.covCount)
		fmt.Printf("\n  Coverage: %s\n", formatCoverage(avgCov))
	}
}

func printCoverageSummary(results testResults) {
	if len(results.packages) == 0 {
		return
	}

	fmt.Printf("\n  %s\n", testHeaderStyle.Render("Coverage by package:"))

	// Sort packages by name
	sort.Slice(results.packages, func(i, j int) bool {
		return results.packages[i].name < results.packages[j].name
	})

	// Find max package name length for alignment
	maxLen := 0
	for _, pkg := range results.packages {
		name := shortenPackageName(pkg.name)
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// Print each package
	for _, pkg := range results.packages {
		if !pkg.hasCov {
			continue
		}
		name := shortenPackageName(pkg.name)
		padding := strings.Repeat(" ", maxLen-len(name)+2)
		fmt.Printf("    %s%s%s\n", name, padding, formatCoverage(pkg.coverage))
	}

	// Print average
	if results.covCount > 0 {
		avgCov := results.totalCov / float64(results.covCount)
		padding := strings.Repeat(" ", maxLen-7+2)
		fmt.Printf("\n    %s%s%s\n", testHeaderStyle.Render("Average"), padding, formatCoverage(avgCov))
	}
}

func formatCoverage(cov float64) string {
	var style lipgloss.Style
	switch {
	case cov >= 80:
		style = testCovHighStyle
	case cov >= 50:
		style = testCovMedStyle
	default:
		style = testCovLowStyle
	}
	return style.Render(fmt.Sprintf("%.1f%%", cov))
}

func shortenPackageName(name string) string {
	// Remove common prefixes
	prefixes := []string{
		"github.com/host-uk/core/",
		"github.com/host-uk/",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return strings.TrimPrefix(name, prefix)
		}
	}
	return filepath.Base(name)
}

func printJSONResults(results testResults, exitCode int) {
	// Simple JSON output for agents
	fmt.Printf("{\n")
	fmt.Printf("  \"passed\": %d,\n", results.passed)
	fmt.Printf("  \"failed\": %d,\n", results.failed)
	fmt.Printf("  \"skipped\": %d,\n", results.skipped)
	if results.covCount > 0 {
		avgCov := results.totalCov / float64(results.covCount)
		fmt.Printf("  \"coverage\": %.1f,\n", avgCov)
	}
	fmt.Printf("  \"exit_code\": %d,\n", exitCode)
	if len(results.failedPkgs) > 0 {
		fmt.Printf("  \"failed_packages\": [\n")
		for i, pkg := range results.failedPkgs {
			comma := ","
			if i == len(results.failedPkgs)-1 {
				comma = ""
			}
			fmt.Printf("    %q%s\n", pkg, comma)
		}
		fmt.Printf("  ]\n")
	} else {
		fmt.Printf("  \"failed_packages\": []\n")
	}
	fmt.Printf("}\n")
}
