package testcmd

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

type packageCoverage struct {
	name     string
	coverage float64
	hasCov   bool
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
		fmt.Printf("  %s %s", testPassStyle.Render("✓"), i18n.T("i18n.count.passed", results.passed))
		if results.failed > 0 {
			fmt.Printf("  %s %s", testFailStyle.Render("✗"), i18n.T("i18n.count.failed", results.failed))
		}
		if results.skipped > 0 {
			fmt.Printf("  %s %s", testSkipStyle.Render("○"), i18n.T("i18n.count.skipped", results.skipped))
		}
		fmt.Println()
	}

	// Print failed packages
	if len(results.failedPkgs) > 0 {
		fmt.Printf("\n  %s\n", i18n.T("cmd.test.failed_packages"))
		for _, pkg := range results.failedPkgs {
			fmt.Printf("    %s %s\n", testFailStyle.Render("✗"), pkg)
		}
	}

	// Print coverage
	if showCoverage {
		printCoverageSummary(results)
	} else if results.covCount > 0 {
		avgCov := results.totalCov / float64(results.covCount)
		fmt.Printf("\n  %s %s\n", i18n.Label("coverage"), formatCoverage(avgCov))
	}
}

func printCoverageSummary(results testResults) {
	if len(results.packages) == 0 {
		return
	}

	fmt.Printf("\n  %s\n", testHeaderStyle.Render(i18n.T("cmd.test.coverage_by_package")))

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
		avgLabel := i18n.T("cmd.test.label.average")
		padding := strings.Repeat(" ", maxLen-len(avgLabel)+2)
		fmt.Printf("\n    %s%s%s\n", testHeaderStyle.Render(avgLabel), padding, formatCoverage(avgCov))
	}
}

func formatCoverage(cov float64) string {
	return cli.FormatCoverage(cov)
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
