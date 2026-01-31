package gocmd

import (
	"os"
	"os/exec"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	fmtFix   bool
	fmtDiff  bool
	fmtCheck bool
)

func addGoFmtCommand(parent *cli.Command) {
	fmtCmd := &cli.Command{
		Use:   "fmt",
		Short: "Format Go code",
		Long:  "Format Go code using goimports or gofmt",
		RunE: func(cmd *cli.Command, args []string) error {
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
			fmtArgs = append(fmtArgs, ".")

			// Try goimports first, fall back to gofmt
			var execCmd *exec.Cmd
			if _, err := exec.LookPath("goimports"); err == nil {
				execCmd = exec.Command("goimports", fmtArgs...)
			} else {
				execCmd = exec.Command("gofmt", fmtArgs...)
			}

			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	fmtCmd.Flags().BoolVar(&fmtFix, "fix", false, i18n.T("common.flag.fix"))
	fmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, "Show diff of changes")
	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false, "Check if formatted (exit 1 if not)")

	parent.AddCommand(fmtCmd)
}

var lintFix bool

func addGoLintCommand(parent *cli.Command) {
	lintCmd := &cli.Command{
		Use:   "lint",
		Short: "Run golangci-lint",
		Long:  "Run golangci-lint for comprehensive static analysis",
		RunE: func(cmd *cli.Command, args []string) error {
			lintArgs := []string{"run"}
			if lintFix {
				lintArgs = append(lintArgs, "--fix")
			}

			execCmd := exec.Command("golangci-lint", lintArgs...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			return execCmd.Run()
		},
	}

	lintCmd.Flags().BoolVar(&lintFix, "fix", false, i18n.T("common.flag.fix"))

	parent.AddCommand(lintCmd)
}
