// Package gocmd provides Go development commands.
//
// Note: Package named gocmd because 'go' is a reserved keyword.
package gocmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases for shared styles
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
)

// AddGoCommands adds Go development commands.
func AddGoCommands(parent *clir.Cli) {
	goCmd := parent.NewSubCommand("go", "Go development tools")
	goCmd.LongDescription("Go development tools with enhanced output and environment setup.\n\n" +
		"Commands:\n" +
		"  test     Run tests\n" +
		"  cov      Run tests with coverage report\n" +
		"  fmt      Format Go code\n" +
		"  lint     Run golangci-lint\n" +
		"  install  Install Go binary\n" +
		"  mod      Module management (tidy, download, verify)\n" +
		"  work     Workspace management")

	addGoTestCommand(goCmd)
	addGoCovCommand(goCmd)
	addGoFmtCommand(goCmd)
	addGoLintCommand(goCmd)
	addGoInstallCommand(goCmd)
	addGoModCommand(goCmd)
	addGoWorkCommand(goCmd)
}

func addGoTestCommand(parent *clir.Command) {
	var (
		coverage bool
		pkg      string
		run      string
		short    bool
		race     bool
		json     bool
		verbose  bool
	)

	testCmd := parent.NewSubCommand("test", "Run tests with coverage")
	testCmd.LongDescription("Run Go tests with coverage reporting.\n\n" +
		"Sets MACOSX_DEPLOYMENT_TARGET=26.0 to suppress linker warnings.\n" +
		"Filters noisy output and provides colour-coded coverage.\n\n" +
		"Examples:\n" +
		"  core go test\n" +
		"  core go test --coverage\n" +
		"  core go test --pkg ./pkg/crypt\n" +
		"  core go test --run TestHash")

	testCmd.BoolFlag("coverage", "Show detailed per-package coverage", &coverage)
	testCmd.StringFlag("pkg", "Package to test (default: ./...)", &pkg)
	testCmd.StringFlag("run", "Run only tests matching regexp", &run)
	testCmd.BoolFlag("short", "Run only short tests", &short)
	testCmd.BoolFlag("race", "Enable race detector", &race)
	testCmd.BoolFlag("json", "Output JSON results", &json)
	testCmd.BoolFlag("v", "Verbose output", &verbose)

	testCmd.Action(func() error {
		return runGoTest(coverage, pkg, run, short, race, json, verbose)
	})
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
		covStyle := successStyle
		if cov < 50 {
			covStyle = errorStyle
		} else if cov < 80 {
			covStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))
		}
		fmt.Printf("\n  %s %s\n", dimStyle.Render("Coverage:"), covStyle.Render(fmt.Sprintf("%.1f%%", cov)))
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

func addGoCovCommand(parent *clir.Command) {
	var (
		pkg     string
		html    bool
		open    bool
		threshold float64
	)

	covCmd := parent.NewSubCommand("cov", "Run tests with coverage report")
	covCmd.LongDescription("Run tests and generate coverage report.\n\n" +
		"Examples:\n" +
		"  core go cov                  # Run with coverage summary\n" +
		"  core go cov --html           # Generate HTML report\n" +
		"  core go cov --open           # Generate and open HTML report\n" +
		"  core go cov --threshold 80   # Fail if coverage < 80%")

	covCmd.StringFlag("pkg", "Package to test (default: ./...)", &pkg)
	covCmd.BoolFlag("html", "Generate HTML coverage report", &html)
	covCmd.BoolFlag("open", "Generate and open HTML report in browser", &open)
	covCmd.Float64Flag("threshold", "Minimum coverage percentage (exit 1 if below)", &threshold)

	covCmd.Action(func() error {
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
		args := append([]string{"test", "-coverprofile=" + covPath, "-covermode=atomic"}, pkgArgs...)
		
		cmd := exec.Command("go", args...)
		cmd.Env = append(os.Environ(), "MACOSX_DEPLOYMENT_TARGET=26.0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		testErr := cmd.Run()

		// Get coverage percentage
		covCmd := exec.Command("go", "tool", "cover", "-func="+covPath)
		covOutput, err := covCmd.Output()
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
		covStyle := successStyle
		if totalCov < 50 {
			covStyle = errorStyle
		} else if totalCov < 80 {
			covStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b"))
		}
		fmt.Printf("  %s %s\n", dimStyle.Render("Total:"), covStyle.Render(fmt.Sprintf("%.1f%%", totalCov)))

		// Generate HTML if requested
		if html || open {
			htmlPath := "coverage.html"
			htmlCmd := exec.Command("go", "tool", "cover", "-html="+covPath, "-o="+htmlPath)
			if err := htmlCmd.Run(); err != nil {
				return fmt.Errorf("failed to generate HTML: %w", err)
			}
			fmt.Printf("  %s %s\n", dimStyle.Render("HTML:"), htmlPath)

			if open {
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
		if threshold > 0 && totalCov < threshold {
			fmt.Printf("\n%s Coverage %.1f%% is below threshold %.1f%%\n",
				errorStyle.Render("FAIL"), totalCov, threshold)
			return fmt.Errorf("coverage below threshold")
		}

		if testErr != nil {
			return testErr
		}

		fmt.Printf("\n%s\n", successStyle.Render("OK"))
		return nil
	})
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

func addGoFmtCommand(parent *clir.Command) {
	var (
		fix   bool
		diff  bool
		check bool
	)

	fmtCmd := parent.NewSubCommand("fmt", "Format Go code")
	fmtCmd.LongDescription("Format Go code using gofmt or goimports.\n\n" +
		"Examples:\n" +
		"  core go fmt              # Check formatting\n" +
		"  core go fmt --fix        # Fix formatting\n" +
		"  core go fmt --diff       # Show diff")

	fmtCmd.BoolFlag("fix", "Fix formatting in place", &fix)
	fmtCmd.BoolFlag("diff", "Show diff of changes", &diff)
	fmtCmd.BoolFlag("check", "Check only, exit 1 if not formatted", &check)

	fmtCmd.Action(func() error {
		args := []string{}
		if fix {
			args = append(args, "-w")
		}
		if diff {
			args = append(args, "-d")
		}
		if !fix && !diff {
			args = append(args, "-l")
		}
		args = append(args, ".")

		// Try goimports first, fall back to gofmt
		var cmd *exec.Cmd
		if _, err := exec.LookPath("goimports"); err == nil {
			cmd = exec.Command("goimports", args...)
		} else {
			cmd = exec.Command("gofmt", args...)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func addGoLintCommand(parent *clir.Command) {
	var fix bool

	lintCmd := parent.NewSubCommand("lint", "Run golangci-lint")
	lintCmd.LongDescription("Run golangci-lint on the codebase.\n\n" +
		"Examples:\n" +
		"  core go lint\n" +
		"  core go lint --fix")

	lintCmd.BoolFlag("fix", "Fix issues automatically", &fix)

	lintCmd.Action(func() error {
		args := []string{"run"}
		if fix {
			args = append(args, "--fix")
		}

		cmd := exec.Command("golangci-lint", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func addGoInstallCommand(parent *clir.Command) {
	var verbose bool
	var noCgo bool

	installCmd := parent.NewSubCommand("install", "Install Go binary")
	installCmd.LongDescription("Install Go binary to $GOPATH/bin.\n\n" +
		"Examples:\n" +
		"  core go install              # Install current module\n" +
		"  core go install ./cmd/core   # Install specific path\n" +
		"  core go install --no-cgo     # Pure Go (no C dependencies)\n" +
		"  core go install -v           # Verbose output")

	installCmd.BoolFlag("v", "Verbose output", &verbose)
	installCmd.BoolFlag("no-cgo", "Disable CGO (CGO_ENABLED=0)", &noCgo)

	installCmd.Action(func() error {
		// Get install path from args or default to current dir
		args := installCmd.OtherArgs()
		installPath := "./..."
		if len(args) > 0 {
			installPath = args[0]
		}

		// Detect if we're in a module with cmd/ subdirectories or a root main.go
		if installPath == "./..." {
			if _, err := os.Stat("core.go"); err == nil {
				installPath = "."
			} else if entries, err := os.ReadDir("cmd"); err == nil && len(entries) > 0 {
				installPath = "./cmd/..."
			} else if _, err := os.Stat("main.go"); err == nil {
				installPath = "."
			}
		}

		fmt.Printf("%s Installing\n", dimStyle.Render("Install:"))
		fmt.Printf("  %s %s\n", dimStyle.Render("Path:"), installPath)
		if noCgo {
			fmt.Printf("  %s %s\n", dimStyle.Render("CGO:"), "disabled")
		}

		cmdArgs := []string{"install"}
		if verbose {
			cmdArgs = append(cmdArgs, "-v")
		}
		cmdArgs = append(cmdArgs, installPath)

		cmd := exec.Command("go", cmdArgs...)
		if noCgo {
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("\n%s\n", errorStyle.Render("FAIL Install failed"))
			return err
		}

		// Show where it was installed
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		binDir := filepath.Join(gopath, "bin")

		fmt.Printf("\n%s Installed to %s\n", successStyle.Render("OK"), binDir)
		return nil
	})
}

func addGoModCommand(parent *clir.Command) {
	modCmd := parent.NewSubCommand("mod", "Module management")
	modCmd.LongDescription("Go module management commands.\n\n" +
		"Commands:\n" +
		"  tidy      Add missing and remove unused modules\n" +
		"  download  Download modules to local cache\n" +
		"  verify    Verify dependencies\n" +
		"  graph     Print module dependency graph")

	// tidy
	tidyCmd := modCmd.NewSubCommand("tidy", "Tidy go.mod")
	tidyCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// download
	downloadCmd := modCmd.NewSubCommand("download", "Download modules")
	downloadCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "download")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// verify
	verifyCmd := modCmd.NewSubCommand("verify", "Verify dependencies")
	verifyCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "verify")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// graph
	graphCmd := modCmd.NewSubCommand("graph", "Print dependency graph")
	graphCmd.Action(func() error {
		cmd := exec.Command("go", "mod", "graph")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func addGoWorkCommand(parent *clir.Command) {
	workCmd := parent.NewSubCommand("work", "Workspace management")
	workCmd.LongDescription("Go workspace management commands.\n\n" +
		"Commands:\n" +
		"  sync    Sync go.work with modules\n" +
		"  init    Initialize go.work\n" +
		"  use     Add module to workspace")

	// sync
	syncCmd := workCmd.NewSubCommand("sync", "Sync workspace")
	syncCmd.Action(func() error {
		cmd := exec.Command("go", "work", "sync")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})

	// init
	initCmd := workCmd.NewSubCommand("init", "Initialize workspace")
	initCmd.Action(func() error {
		cmd := exec.Command("go", "work", "init")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		// Auto-add current module if go.mod exists
		if _, err := os.Stat("go.mod"); err == nil {
			cmd = exec.Command("go", "work", "use", ".")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return nil
	})

	// use
	useCmd := workCmd.NewSubCommand("use", "Add module to workspace")
	useCmd.Action(func() error {
		args := useCmd.OtherArgs()
		if len(args) == 0 {
			// Auto-detect modules
			modules := findGoModules(".")
			if len(modules) == 0 {
				return fmt.Errorf("no go.mod files found")
			}
			for _, mod := range modules {
				cmd := exec.Command("go", "work", "use", mod)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return err
				}
				fmt.Printf("Added %s\n", mod)
			}
			return nil
		}

		cmdArgs := append([]string{"work", "use"}, args...)
		cmd := exec.Command("go", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

func findGoModules(root string) []string {
	var modules []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == "go.mod" && path != "go.mod" {
			modules = append(modules, filepath.Dir(path))
		}
		return nil
	})
	return modules
}
