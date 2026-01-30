package gocmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

var qaFix bool

func addGoQACommand(parent *cobra.Command) {
	qaCmd := &cobra.Command{
		Use:   "qa",
		Short: "Run QA checks",
		Long:  "Run code quality checks: formatting, vetting, linting, and testing",
		RunE:  runGoQADefault,
	}

	qaCmd.PersistentFlags().BoolVar(&qaFix, "fix", false, i18n.T("common.flag.fix"))

	// Subcommands for individual checks
	qaCmd.AddCommand(&cobra.Command{
		Use:   "fmt",
		Short: "Check/fix code formatting",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"fmt"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "vet",
		Short: "Run go vet",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"vet"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "lint",
		Short: "Run golangci-lint",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"lint"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Run tests",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"test"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "race",
		Short: "Run tests with race detector",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"race"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "vuln",
		Short: "Check for vulnerabilities",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"vuln"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "sec",
		Short: "Run security scanner",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"sec"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "quick",
		Short: "Quick QA: fmt, vet, lint",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"fmt", "vet", "lint"}) },
	})

	qaCmd.AddCommand(&cobra.Command{
		Use:   "full",
		Short: "Full QA: all checks including race, vuln, sec",
		RunE:  func(cmd *cobra.Command, args []string) error { return runQAChecks([]string{"fmt", "vet", "lint", "test", "race", "vuln", "sec"}) },
	})

	parent.AddCommand(qaCmd)
}

// runGoQADefault runs the default QA checks (fmt, vet, lint, test)
func runGoQADefault(cmd *cobra.Command, args []string) error {
	return runQAChecks([]string{"fmt", "vet", "lint", "test"})
}

// QACheck represents a single QA check.
type QACheck struct {
	Name    string
	Command string
	Args    []string
}

func runQAChecks(checkNames []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
	}

	// Detect if this is a Go project
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("not a Go project (no %s found)", i18n.T("gram.word.go_mod"))
	}

	fmt.Printf("%s %s\n\n", cli.DimStyle.Render(i18n.Label("qa")), i18n.ProgressSubject("run", "Go QA"))

	checks := buildChecksForNames(checkNames)

	ctx := context.Background()
	startTime := time.Now()
	passed := 0
	failed := 0

	for _, check := range checks {
		fmt.Printf("%s %s\n", cli.DimStyle.Render("→"), i18n.Progress(check.Name))

		if err := runCheck(ctx, cwd, check); err != nil {
			fmt.Printf("  %s %s\n", cli.ErrorStyle.Render(cli.SymbolCross), err.Error())
			failed++
		} else {
			fmt.Printf("  %s %s\n", cli.SuccessStyle.Render(cli.SymbolCheck), i18n.T("i18n.done.pass"))
			passed++
		}
	}

	// Summary
	fmt.Println()
	duration := time.Since(startTime).Round(time.Millisecond)

	if failed > 0 {
		fmt.Printf("%s %s, %s (%s)\n",
			cli.ErrorStyle.Render(cli.SymbolCross),
			i18n.T("i18n.count.check", passed)+" "+i18n.T("i18n.done.pass"),
			i18n.T("i18n.count.check", failed)+" "+i18n.T("i18n.done.fail"),
			duration)
		os.Exit(1)
	}

	fmt.Printf("%s %s (%s)\n",
		cli.SuccessStyle.Render(cli.SymbolCheck),
		i18n.T("i18n.count.check", passed)+" "+i18n.T("i18n.done.pass"),
		duration)

	return nil
}

func buildChecksForNames(names []string) []QACheck {
	allChecks := map[string]QACheck{
		"fmt": {
			Name:    "format",
			Command: "gofmt",
			Args:    fmtArgs(qaFix),
		},
		"vet": {
			Name:    "vet",
			Command: "go",
			Args:    []string{"vet", "./..."},
		},
		"lint": {
			Name:    "lint",
			Command: "golangci-lint",
			Args:    lintArgs(qaFix),
		},
		"test": {
			Name:    "test",
			Command: "go",
			Args:    []string{"test", "./..."},
		},
		"race": {
			Name:    "test",
			Command: "go",
			Args:    []string{"test", "-race", "./..."},
		},
		"vuln": {
			Name:    "scan",
			Command: "govulncheck",
			Args:    []string{"./..."},
		},
		"sec": {
			Name:    "scan",
			Command: "gosec",
			Args:    []string{"-quiet", "./..."},
		},
	}

	var checks []QACheck
	for _, name := range names {
		if check, ok := allChecks[name]; ok {
			checks = append(checks, check)
		}
	}
	return checks
}

func fmtArgs(fix bool) []string {
	if fix {
		return []string{"-w", "."}
	}
	return []string{"-l", "."}
}

func lintArgs(fix bool) []string {
	args := []string{"run"}
	if fix {
		args = append(args, "--fix")
	}
	args = append(args, "./...")
	return args
}

func runCheck(ctx context.Context, dir string, check QACheck) error {
	// Check if command exists
	if _, err := exec.LookPath(check.Command); err != nil {
		return fmt.Errorf("%s: %s", check.Command, i18n.T("i18n.done.miss"))
	}

	cmd := exec.CommandContext(ctx, check.Command, check.Args...)
	cmd.Dir = dir

	// For gofmt -l, capture output to check if files need formatting
	if check.Name == "format" && len(check.Args) > 0 && check.Args[0] == "-l" {
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		if len(output) > 0 {
			// Show files that need formatting
			fmt.Print(string(output))
			return fmt.Errorf("%s (use --fix)", i18n.T("i18n.fail.format", i18n.T("i18n.count.file", len(output))))
		}
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
