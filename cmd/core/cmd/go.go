package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leaanthony/clir"
)

// AddGoCommands adds Go development commands.
func AddGoCommands(parent *clir.Cli) {
	goCmd := parent.NewSubCommand("go", "Go development tools")
	goCmd.LongDescription("Go development tools with enhanced output and environment setup.\n\n" +
		"Commands:\n" +
		"  test     Run tests with coverage\n" +
		"  fmt      Format Go code\n" +
		"  lint     Run golangci-lint\n" +
		"  mod      Module management (tidy, download, verify)\n" +
		"  work     Workspace management")

	addGoTestCommand(goCmd)
	addGoFmtCommand(goCmd)
	addGoLintCommand(goCmd)
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
