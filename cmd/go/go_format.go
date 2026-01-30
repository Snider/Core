package gocmd

import (
	"os"
	"os/exec"

	"github.com/leaanthony/clir"
)

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
