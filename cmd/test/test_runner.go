package testcmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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
