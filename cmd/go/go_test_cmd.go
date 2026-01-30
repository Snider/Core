package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

var (
	testCoverage bool
	testPkg      string
	testRun      string
	testShort    bool
	testRace     bool
	testJSON     bool
	testVerbose  bool
)

func addGoTestCommand(parent *cobra.Command) {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests with coverage",
		Long: "Run Go tests with coverage reporting.\n\n" +
			"Sets MACOSX_DEPLOYMENT_TARGET=26.0 to suppress linker warnings.\n" +
			"Filters noisy output and provides colour-coded coverage.\n\n" +
			"Examples:\n" +
			"  core go test\n" +
			"  core go test --coverage\n" +
			"  core go test --pkg ./pkg/crypt\n" +
			"  core go test --run TestHash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoTest(testCoverage, testPkg, testRun, testShort, testRace, testJSON, testVerbose)
		},
	}

	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, "Show detailed per-package coverage")
	testCmd.Flags().StringVar(&testPkg, "pkg", "", "Package to test (default: ./...)")
	testCmd.Flags().StringVar(&testRun, "run", "", "Run only tests matching regexp")
	testCmd.Flags().BoolVar(&testShort, "short", false, "Run only short tests")
	testCmd.Flags().BoolVar(&testRace, "race", false, "Enable race detector")
	testCmd.Flags().BoolVar(&testJSON, "json", false, "Output JSON results")
	testCmd.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Verbose output")

	parent.AddCommand(testCmd)
}

func runGoTest(coverage bool, pkg, run string, short, race, jsonOut, verbose bool) error {
	if pkg == "" {
		pkg = "./..."
	}

	args := []string{"test"}

	if coverage {
		args = append(args, "-cover")
	} else {
		args = append(args, "-cover")
	}

	if run != "" {
		args = append(args, "-run", run)
	}
	if short {
		args = append(args, "-short")
	}
	if race {
		args = append(args, "-race")
	}
	if verbose {
		args = append(args, "-v")
	}

	args = append(args, pkg)

	if !jsonOut {
		fmt.Printf("%s Running tests\n", dimStyle.Render("Test:"))
		fmt.Printf("  %s %s\n", dimStyle.Render("Package:"), pkg)
		fmt.Println()
	}

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "MACOSX_DEPLOYMENT_TARGET=26.0", "CGO_ENABLED=0")
	cmd.Dir, _ = os.Getwd()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Filter linker warnings
	lines := strings.Split(outputStr, "\n")
	var filtered []string
	for _, line := range lines {
		if !strings.Contains(line, "ld: warning:") {
			filtered = append(filtered, line)
		}
	}
	outputStr = strings.Join(filtered, "\n")

	// Parse results
	passed, failed, skipped := parseTestResults(outputStr)
	cov := parseOverallCoverage(outputStr)

	if jsonOut {
		fmt.Printf(`{"passed":%d,"failed":%d,"skipped":%d,"coverage":%.1f,"exit_code":%d}`,
			passed, failed, skipped, cov, cmd.ProcessState.ExitCode())
		fmt.Println()
		return err
	}

	// Print filtered output if verbose or failed
	if verbose || err != nil {
		fmt.Println(outputStr)
	}

	// Summary
	if err == nil {
		fmt.Printf("  %s %d passed\n", successStyle.Render("✓"), passed)
	} else {
		fmt.Printf("  %s %d passed, %d failed\n", errorStyle.Render("✗"), passed, failed)
	}

	if cov > 0 {
		fmt.Printf("\n  %s %s\n", shared.ProgressLabel("Coverage"), shared.FormatCoverage(cov))
	}

	if err == nil {
		fmt.Printf("\n%s\n", successStyle.Render("PASS All tests passed"))
	} else {
		fmt.Printf("\n%s\n", errorStyle.Render("FAIL Some tests failed"))
	}

	return err
}

func parseTestResults(output string) (passed, failed, skipped int) {
	passRe := regexp.MustCompile(`(?m)^ok\s+`)
	failRe := regexp.MustCompile(`(?m)^FAIL\s+`)
	skipRe := regexp.MustCompile(`(?m)^\?\s+`)

	passed = len(passRe.FindAllString(output, -1))
	failed = len(failRe.FindAllString(output, -1))
	skipped = len(skipRe.FindAllString(output, -1))
	return
}

func parseOverallCoverage(output string) float64 {
	re := regexp.MustCompile(`coverage:\s+([\d.]+)%`)
	matches := re.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return 0
	}

	var total float64
	for _, m := range matches {
		var cov float64
		fmt.Sscanf(m[1], "%f", &cov)
		total += cov
	}
	return total / float64(len(matches))
}

var (
	covPkg       string
	covHTML      bool
	covOpen      bool
	covThreshold float64
)

func addGoCovCommand(parent *cobra.Command) {
	covCmd := &cobra.Command{
		Use:   "cov",
		Short: "Run tests with coverage report",
		Long: "Run tests and generate coverage report.\n\n" +
			"Examples:\n" +
			"  core go cov                  # Run with coverage summary\n" +
			"  core go cov --html           # Generate HTML report\n" +
			"  core go cov --open           # Generate and open HTML report\n" +
			"  core go cov --threshold 80   # Fail if coverage < 80%",
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg := covPkg
			if pkg == "" {
				// Auto-discover packages with tests
				pkgs, err := findTestPackages(".")
				if err != nil {
					return fmt.Errorf("failed to discover test packages: %w", err)
				}
				if len(pkgs) == 0 {
					return fmt.Errorf("no test packages found")
				}
				pkg = strings.Join(pkgs, " ")
			}

			// Create temp file for coverage data
			covFile, err := os.CreateTemp("", "coverage-*.out")
			if err != nil {
				return fmt.Errorf("failed to create coverage file: %w", err)
			}
			covPath := covFile.Name()
			covFile.Close()
			defer os.Remove(covPath)

			fmt.Printf("%s Running tests with coverage\n", dimStyle.Render("Coverage:"))
			// Truncate package list if too long for display
			displayPkg := pkg
			if len(displayPkg) > 60 {
				displayPkg = displayPkg[:57] + "..."
			}
			fmt.Printf("  %s %s\n", dimStyle.Render("Package:"), displayPkg)
			fmt.Println()

			// Run tests with coverage
			// We need to split pkg into individual arguments if it contains spaces
			pkgArgs := strings.Fields(pkg)
			cmdArgs := append([]string{"test", "-coverprofile=" + covPath, "-covermode=atomic"}, pkgArgs...)

			goCmd := exec.Command("go", cmdArgs...)
			goCmd.Env = append(os.Environ(), "MACOSX_DEPLOYMENT_TARGET=26.0")
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr

			testErr := goCmd.Run()

			// Get coverage percentage
			coverCmd := exec.Command("go", "tool", "cover", "-func="+covPath)
			covOutput, err := coverCmd.Output()
			if err != nil {
				if testErr != nil {
					return testErr
				}
				return fmt.Errorf("failed to get coverage: %w", err)
			}

			// Parse total coverage from last line
			lines := strings.Split(strings.TrimSpace(string(covOutput)), "\n")
			var totalCov float64
			if len(lines) > 0 {
				lastLine := lines[len(lines)-1]
				// Format: "total:    (statements)    XX.X%"
				if strings.Contains(lastLine, "total:") {
					parts := strings.Fields(lastLine)
					if len(parts) >= 3 {
						covStr := strings.TrimSuffix(parts[len(parts)-1], "%")
						fmt.Sscanf(covStr, "%f", &totalCov)
					}
				}
			}

			// Print coverage summary
			fmt.Println()
			fmt.Printf("  %s %s\n", shared.ProgressLabel("Total"), shared.FormatCoverage(totalCov))

			// Generate HTML if requested
			if covHTML || covOpen {
				htmlPath := "coverage.html"
				htmlCmd := exec.Command("go", "tool", "cover", "-html="+covPath, "-o="+htmlPath)
				if err := htmlCmd.Run(); err != nil {
					return fmt.Errorf("failed to generate HTML: %w", err)
				}
				fmt.Printf("  %s %s\n", dimStyle.Render("HTML:"), htmlPath)

				if covOpen {
					// Open in browser
					var openCmd *exec.Cmd
					switch {
					case exec.Command("which", "open").Run() == nil:
						openCmd = exec.Command("open", htmlPath)
					case exec.Command("which", "xdg-open").Run() == nil:
						openCmd = exec.Command("xdg-open", htmlPath)
					default:
						fmt.Printf("  %s\n", dimStyle.Render("(open manually)"))
					}
					if openCmd != nil {
						openCmd.Run()
					}
				}
			}

			// Check threshold
			if covThreshold > 0 && totalCov < covThreshold {
				fmt.Printf("\n%s Coverage %.1f%% is below threshold %.1f%%\n",
					errorStyle.Render("FAIL"), totalCov, covThreshold)
				return fmt.Errorf("coverage below threshold")
			}

			if testErr != nil {
				return testErr
			}

			fmt.Printf("\n%s\n", successStyle.Render("OK"))
			return nil
		},
	}

	covCmd.Flags().StringVar(&covPkg, "pkg", "", "Package to test (default: ./...)")
	covCmd.Flags().BoolVar(&covHTML, "html", false, "Generate HTML coverage report")
	covCmd.Flags().BoolVar(&covOpen, "open", false, "Generate and open HTML report in browser")
	covCmd.Flags().Float64Var(&covThreshold, "threshold", 0, "Minimum coverage percentage (exit 1 if below)")

	parent.AddCommand(covCmd)
}

func findTestPackages(root string) ([]string, error) {
	pkgMap := make(map[string]bool)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			dir := filepath.Dir(path)
			if !strings.HasPrefix(dir, ".") {
				dir = "./" + dir
			}
			pkgMap[dir] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var pkgs []string
	for pkg := range pkgMap {
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}
