package gocmd

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	fmtFix   bool
	fmtDiff  bool
	fmtCheck bool
	fmtAll   bool
)

func addGoFmtCommand(parent *cli.Command) {
	fmtCmd := &cli.Command{
		Use:   "fmt",
		Short: "Format Go code",
		Long:  "Format Go code using goimports or gofmt. By default only checks changed files.",
		RunE: func(cmd *cli.Command, args []string) error {
			// Get list of files to check
			var files []string
			if fmtAll {
				// Check all Go files
				files = []string{"."}
			} else {
				// Only check changed Go files (git-aware)
				files = getChangedGoFiles()
				if len(files) == 0 {
					cli.Print("%s\n", i18n.T("cmd.go.fmt.no_changes"))
					return nil
				}
			}

			// Validate flag combinations
			if fmtCheck && fmtFix {
				return cli.Err("--check and --fix are mutually exclusive")
			}

			fmtArgs := []string{}
			if fmtFix {
				fmtArgs = append(fmtArgs, "-w")
			}
			if fmtDiff {
				fmtArgs = append(fmtArgs, "-d")
			}
			if !fmtFix && !fmtDiff {
				fmtArgs = append(fmtArgs, "-l")
			}
			fmtArgs = append(fmtArgs, files...)

			// Try goimports first, fall back to gofmt
			var execCmd *exec.Cmd
			if _, err := exec.LookPath("goimports"); err == nil {
				execCmd = exec.Command("goimports", fmtArgs...)
			} else {
				execCmd = exec.Command("gofmt", fmtArgs...)
			}

			// For --check mode, capture output to detect unformatted files
			if fmtCheck {
				output, err := execCmd.CombinedOutput()
				if err != nil {
					os.Stderr.Write(output)
					return err
				}
				if len(output) > 0 {
					os.Stdout.Write(output)
					return cli.Err("files need formatting (use --fix)")
				}
				return nil
			}

			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	fmtCmd.Flags().BoolVar(&fmtFix, "fix", false, i18n.T("common.flag.fix"))
	fmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, i18n.T("common.flag.diff"))
	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false, i18n.T("cmd.go.fmt.flag.check"))
	fmtCmd.Flags().BoolVar(&fmtAll, "all", false, i18n.T("cmd.go.fmt.flag.all"))

	parent.AddCommand(fmtCmd)
}

// getChangedGoFiles returns Go files that have been modified, staged, or are untracked.
func getChangedGoFiles() []string {
	var files []string

	// Get modified and staged files
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		files = append(files, filterGoFiles(string(output))...)
	}

	// Get untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err = cmd.Output()
	if err == nil {
		files = append(files, filterGoFiles(string(output))...)
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			// Verify file exists (might have been deleted)
			if _, err := os.Stat(f); err == nil {
				unique = append(unique, f)
			}
		}
	}

	return unique
}

// filterGoFiles filters a newline-separated list of files to only include .go files.
func filterGoFiles(output string) []string {
	var goFiles []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" && filepath.Ext(file) == ".go" {
			goFiles = append(goFiles, file)
		}
	}
	return goFiles
}

var (
	lintFix bool
	lintAll bool
)

func addGoLintCommand(parent *cli.Command) {
	lintCmd := &cli.Command{
		Use:   "lint",
		Short: "Run golangci-lint",
		Long:  "Run golangci-lint for comprehensive static analysis. By default only lints changed files.",
		RunE: func(cmd *cli.Command, args []string) error {
			lintArgs := []string{"run"}
			if lintFix {
				lintArgs = append(lintArgs, "--fix")
			}

			if !lintAll {
				// Use --new-from-rev=HEAD to only report issues in uncommitted changes
				// This is golangci-lint's native way to handle incremental linting
				lintArgs = append(lintArgs, "--new-from-rev=HEAD")
			}

			// Always lint all packages
			lintArgs = append(lintArgs, "./...")

			execCmd := exec.Command("golangci-lint", lintArgs...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	lintCmd.Flags().BoolVar(&lintFix, "fix", false, i18n.T("common.flag.fix"))
	lintCmd.Flags().BoolVar(&lintAll, "all", false, i18n.T("cmd.go.lint.flag.all"))

	parent.AddCommand(lintCmd)
}
