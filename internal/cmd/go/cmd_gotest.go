package gocmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
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

func addGoTestCommand(parent *cli.Command) {
	testCmd := &cli.Command{
		Use:   "test",
		Short: "Run Go tests",
		Long:  "Run Go tests with optional coverage, filtering, and race detection",
		RunE: func(cmd *cli.Command, args []string) error {
			return runGoTest(testCoverage, testPkg, testRun, testShort, testRace, testJSON, testVerbose)
		},
	}

	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, "Generate coverage report")
	testCmd.Flags().StringVar(&testPkg, "pkg", "", "Package to test")
	testCmd.Flags().StringVar(&testRun, "run", "", "Run only tests matching pattern")
	testCmd.Flags().BoolVar(&testShort, "short", false, "Run only short tests")
	testCmd.Flags().BoolVar(&testRace, "race", false, "Enable race detector")
	testCmd.Flags().BoolVar(&testJSON, "json", false, "Output as JSON")
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
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("test")), i18n.ProgressSubject("run", "tests"))
		cli.Print("  %s %s\n", dimStyle.Render(i18n.Label("package")), pkg)
		cli.Blank()
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
		cli.Print(`{"passed":%d,"failed":%d,"skipped":%d,"coverage":%.1f,"exit_code":%d}`,
			passed, failed, skipped, cov, cmd.ProcessState.ExitCode())
		cli.Blank()
		return err
	}

	// Print filtered output if verbose or failed
	if verbose || err != nil {
		cli.Text(outputStr)
	}

	// Summary
	if err == nil {
		cli.Print("  %s %s\n", successStyle.Render(cli.Glyph(":check:")), i18n.T("i18n.count.test", passed)+" "+i18n.T("i18n.done.pass"))
	} else {
		cli.Print("  %s %s, %s\n", errorStyle.Render(cli.Glyph(":cross:")),
			i18n.T("i18n.count.test", passed)+" "+i18n.T("i18n.done.pass"),
			i18n.T("i18n.count.test", failed)+" "+i18n.T("i18n.done.fail"))
	}

	if cov > 0 {
		cli.Print("\n  %s %s\n", cli.KeyStyle.Render(i18n.Label("coverage")), formatCoverage(cov))
	}

	if err == nil {
		cli.Print("\n%s\n", successStyle.Render(i18n.T("i18n.done.pass")))
	} else {
		cli.Print("\n%s\n", errorStyle.Render(i18n.T("i18n.done.fail")))
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
		_, _ = fmt.Sscanf(m[1], "%f", &cov)
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

func addGoCovCommand(parent *cli.Command) {
	covCmd := &cli.Command{
		Use:   "cov",
		Short: "Run tests with coverage report",
		Long:  "Run tests with detailed coverage reports, HTML output, and threshold checking",
		RunE: func(cmd *cli.Command, args []string) error {
			pkg := covPkg
			if pkg == "" {
				// Auto-discover packages with tests
				pkgs, err := findTestPackages(".")
				if err != nil {
					return cli.Wrap(err, i18n.T("i18n.fail.find", "test packages"))
				}
				if len(pkgs) == 0 {
					return errors.New("no test packages found")
				}
				pkg = strings.Join(pkgs, " ")
			}

			// Create temp file for coverage data
			covFile, err := os.CreateTemp("", "coverage-*.out")
			if err != nil {
				return cli.Wrap(err, i18n.T("i18n.fail.create", "coverage file"))
			}
			covPath := covFile.Name()
			_ = covFile.Close()
			defer func() { _ = os.Remove(covPath) }()

			cli.Print("%s %s\n", dimStyle.Render(i18n.Label("coverage")), i18n.ProgressSubject("run", "tests"))
			// Truncate package list if too long for display
			displayPkg := pkg
			if len(displayPkg) > 60 {
				displayPkg = displayPkg[:57] + "..."
			}
			cli.Print("  %s %s\n", dimStyle.Render(i18n.Label("package")), displayPkg)
			cli.Blank()

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
				return cli.Wrap(err, i18n.T("i18n.fail.get", "coverage"))
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
						_, _ = fmt.Sscanf(covStr, "%f", &totalCov)
					}
				}
			}

			// Print coverage summary
			cli.Blank()
			cli.Print("  %s %s\n", cli.KeyStyle.Render(i18n.Label("total")), formatCoverage(totalCov))

			// Generate HTML if requested
			if covHTML || covOpen {
				htmlPath := "coverage.html"
				htmlCmd := exec.Command("go", "tool", "cover", "-html="+covPath, "-o="+htmlPath)
				if err := htmlCmd.Run(); err != nil {
					return cli.Wrap(err, i18n.T("i18n.fail.generate", "HTML"))
				}
				cli.Print("  %s %s\n", dimStyle.Render(i18n.Label("html")), htmlPath)

				if covOpen {
					// Open in browser
					var openCmd *exec.Cmd
					switch {
					case exec.Command("which", "open").Run() == nil:
						openCmd = exec.Command("open", htmlPath)
					case exec.Command("which", "xdg-open").Run() == nil:
						openCmd = exec.Command("xdg-open", htmlPath)
					default:
						cli.Print("  %s\n", dimStyle.Render("Open coverage.html in your browser"))
					}
					if openCmd != nil {
						_ = openCmd.Run()
					}
				}
			}

			// Check threshold
			if covThreshold > 0 && totalCov < covThreshold {
				cli.Print("\n%s %.1f%% < %.1f%%\n", errorStyle.Render(i18n.T("i18n.fail.meet", "threshold")), totalCov, covThreshold)
				return errors.New("coverage below threshold")
			}

			if testErr != nil {
				return testErr
			}

			cli.Print("\n%s\n", successStyle.Render(i18n.T("i18n.done.pass")))
			return nil
		},
	}

	covCmd.Flags().StringVar(&covPkg, "pkg", "", "Package to test")
	covCmd.Flags().BoolVar(&covHTML, "html", false, "Generate HTML report")
	covCmd.Flags().BoolVar(&covOpen, "open", false, "Open HTML report in browser")
	covCmd.Flags().Float64Var(&covThreshold, "threshold", 0, "Minimum coverage percentage")

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

func formatCoverage(cov float64) string {
	s := fmt.Sprintf("%.1f%%", cov)
	if cov >= 80 {
		return cli.SuccessStyle.Render(s)
	} else if cov >= 50 {
		return cli.WarningStyle.Render(s)
	}
	return cli.ErrorStyle.Render(s)
}
