// Package testcmd provides test running commands.
//
// Note: Package named testcmd to avoid conflict with Go's test package.
package testcmd

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
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

// Flag variables for test command
var (
	testVerbose  bool
	testCoverage bool
	testShort    bool
	testPkg      string
	testRun      string
	testRace     bool
	testJSON     bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests with coverage",
	Long: `Runs Go tests with coverage reporting.

Sets MACOSX_DEPLOYMENT_TARGET=26.0 to suppress linker warnings on macOS.

Examples:
  core test                     # Run all tests with coverage summary
  core test --verbose           # Show test output as it runs
  core test --coverage          # Show detailed per-package coverage
  core test --pkg ./pkg/...     # Test specific packages
  core test --run TestName      # Run specific test by name
  core test --short             # Skip long-running tests
  core test --race              # Enable race detector
  core test --json              # Output JSON for CI/agents`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTest(testVerbose, testCoverage, testShort, testPkg, testRun, testRace, testJSON)
	},
}

func init() {
	testCmd.Flags().BoolVar(&testVerbose, "verbose", false, "Show test output as it runs (-v)")
	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, "Show detailed per-package coverage")
	testCmd.Flags().BoolVar(&testShort, "short", false, "Skip long-running tests (-short)")
	testCmd.Flags().StringVar(&testPkg, "pkg", "", "Package pattern to test (default: ./...)")
	testCmd.Flags().StringVar(&testRun, "run", "", "Run only tests matching this regex (-run)")
	testCmd.Flags().BoolVar(&testRace, "race", false, "Enable race detector (-race)")
	testCmd.Flags().BoolVar(&testJSON, "json", false, "Output JSON for CI/agents")
}
