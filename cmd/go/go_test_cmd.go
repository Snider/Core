package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.go.test.short"),
		Long:  i18n.T("cmd.go.test.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoTest(testCoverage, testPkg, testRun, testShort, testRace, testJSON, testVerbose)
		},
	}

	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, i18n.T("cmd.go.test.flag.coverage"))
	testCmd.Flags().StringVar(&testPkg, "pkg", "", i18n.T("cmd.go.test.flag.pkg"))
	testCmd.Flags().StringVar(&testRun, "run", "", i18n.T("cmd.go.test.flag.run"))
	testCmd.Flags().BoolVar(&testShort, "short", false, i18n.T("cmd.go.test.flag.short"))
	testCmd.Flags().BoolVar(&testRace, "race", false, i18n.T("cmd.go.test.flag.race"))
	testCmd.Flags().BoolVar(&testJSON, "json", false, i18n.T("cmd.go.test.flag.json"))
	testCmd.Flags().BoolVarP(&testVerbose, "verbose", "v", false, i18n.T("cmd.go.test.flag.verbose"))

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
		fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.go.test.label")), i18n.T("cmd.go.test.running"))
		fmt.Printf("  %s %s\n", dimStyle.Render(i18n.T("cmd.go.test.package_label")), pkg)
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
		fmt.Printf("  %s %s\n", successStyle.Render("✓"), i18n.T("cmd.go.test.passed", map[string]interface{}{"Count": passed}))
	} else {
		fmt.Printf("  %s %s\n", errorStyle.Render("✗"), i18n.T("cmd.go.test.passed_failed", map[string]interface{}{"Passed": passed, "Failed": failed}))
	}

	if cov > 0 {
		fmt.Printf("\n  %s %s\n", cli.ProgressLabel(i18n.T("cmd.go.test.coverage")), cli.FormatCoverage(cov))
	}

	if err == nil {
		fmt.Printf("\n%s\n", successStyle.Render(i18n.T("cmd.go.test.all_passed")))
	} else {
		fmt.Printf("\n%s\n", errorStyle.Render(i18n.T("cmd.go.test.some_failed")))
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
		Short: i18n.T("cmd.go.cov.short"),
		Long:  i18n.T("cmd.go.cov.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg := covPkg
			if pkg == "" {
				// Auto-discover packages with tests
				pkgs, err := findTestPackages(".")
				if err != nil {
					return fmt.Errorf("%s: %w", i18n.T("cmd.go.cov.error.discover"), err)
				}
				if len(pkgs) == 0 {
					return fmt.Errorf(i18n.T("cmd.go.cov.error.no_packages"))
				}
				pkg = strings.Join(pkgs, " ")
			}

			// Create temp file for coverage data
			covFile, err := os.CreateTemp("", "coverage-*.out")
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.go.cov.error.create_file"), err)
			}
			covPath := covFile.Name()
			covFile.Close()
			defer os.Remove(covPath)

			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.go.cov.label")), i18n.T("cmd.go.cov.running"))
			// Truncate package list if too long for display
			displayPkg := pkg
			if len(displayPkg) > 60 {
				displayPkg = displayPkg[:57] + "..."
			}
			fmt.Printf("  %s %s\n", dimStyle.Render(i18n.T("cmd.go.test.package_label")), displayPkg)
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.go.cov.error.get_coverage"), err)
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
			fmt.Printf("  %s %s\n", cli.ProgressLabel(i18n.T("label.total")), cli.FormatCoverage(totalCov))

			// Generate HTML if requested
			if covHTML || covOpen {
				htmlPath := "coverage.html"
				htmlCmd := exec.Command("go", "tool", "cover", "-html="+covPath, "-o="+htmlPath)
				if err := htmlCmd.Run(); err != nil {
					return fmt.Errorf("%s: %w", i18n.T("cmd.go.cov.error.generate_html"), err)
				}
				fmt.Printf("  %s %s\n", dimStyle.Render(i18n.T("cmd.go.cov.html_label")), htmlPath)

				if covOpen {
					// Open in browser
					var openCmd *exec.Cmd
					switch {
					case exec.Command("which", "open").Run() == nil:
						openCmd = exec.Command("open", htmlPath)
					case exec.Command("which", "xdg-open").Run() == nil:
						openCmd = exec.Command("xdg-open", htmlPath)
					default:
						fmt.Printf("  %s\n", dimStyle.Render(i18n.T("cmd.go.cov.open_manually")))
					}
					if openCmd != nil {
						openCmd.Run()
					}
				}
			}

			// Check threshold
			if covThreshold > 0 && totalCov < covThreshold {
				fmt.Printf("\n%s\n", errorStyle.Render(i18n.T("cmd.go.cov.below_threshold", map[string]interface{}{
					"Actual":    fmt.Sprintf("%.1f", totalCov),
					"Threshold": fmt.Sprintf("%.1f", covThreshold),
				})))
				return fmt.Errorf(i18n.T("cmd.go.cov.error.below_threshold"))
			}

			if testErr != nil {
				return testErr
			}

			fmt.Printf("\n%s\n", successStyle.Render(i18n.T("cli.ok")))
			return nil
		},
	}

	covCmd.Flags().StringVar(&covPkg, "pkg", "", i18n.T("cmd.go.cov.flag.pkg"))
	covCmd.Flags().BoolVar(&covHTML, "html", false, i18n.T("cmd.go.cov.flag.html"))
	covCmd.Flags().BoolVar(&covOpen, "open", false, i18n.T("cmd.go.cov.flag.open"))
	covCmd.Flags().Float64Var(&covThreshold, "threshold", 0, i18n.T("cmd.go.cov.flag.threshold"))

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
