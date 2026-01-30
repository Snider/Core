// Package testcmd provides test running commands.
//
// Note: Package named testcmd to avoid conflict with Go's test package.
package testcmd

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases from shared
var (
	testHeaderStyle = shared.RepoNameStyle
	testPassStyle   = shared.SuccessStyle
	testFailStyle   = shared.ErrorStyle
	testSkipStyle   = shared.WarningStyle
	testDimStyle    = shared.DimStyle
)

// Coverage-specific styles
var (
	testCovHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22c55e")) // green-500

	testCovMedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f59e0b")) // amber-500

	testCovLowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ef4444")) // red-500
)

// AddTestCommand adds the 'test' command to the given parent command.
func AddTestCommand(parent *clir.Cli) {
	var verbose bool
	var coverage bool
	var short bool
	var pkg string
	var run string
	var race bool
	var json bool

	testCmd := parent.NewSubCommand("test", "Run tests with coverage")
	testCmd.LongDescription("Runs Go tests with coverage reporting.\n\n" +
		"Sets MACOSX_DEPLOYMENT_TARGET=26.0 to suppress linker warnings on macOS.\n\n" +
		"Examples:\n" +
		"  core test                     # Run all tests with coverage summary\n" +
		"  core test --verbose           # Show test output as it runs\n" +
		"  core test --coverage          # Show detailed per-package coverage\n" +
		"  core test --pkg ./pkg/...     # Test specific packages\n" +
		"  core test --run TestName      # Run specific test by name\n" +
		"  core test --short             # Skip long-running tests\n" +
		"  core test --race              # Enable race detector\n" +
		"  core test --json              # Output JSON for CI/agents")

	testCmd.BoolFlag("verbose", "Show test output as it runs (-v)", &verbose)
	testCmd.BoolFlag("coverage", "Show detailed per-package coverage", &coverage)
	testCmd.BoolFlag("short", "Skip long-running tests (-short)", &short)
	testCmd.StringFlag("pkg", "Package pattern to test (default: ./...)", &pkg)
	testCmd.StringFlag("run", "Run only tests matching this regex (-run)", &run)
	testCmd.BoolFlag("race", "Enable race detector (-race)", &race)
	testCmd.BoolFlag("json", "Output JSON for CI/agents", &json)

	testCmd.Action(func() error {
		return runTest(verbose, coverage, short, pkg, run, race, json)
	})
}
