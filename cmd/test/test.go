// Package testcmd provides test running commands.
//
// Note: Package named testcmd to avoid conflict with Go's test package.
package testcmd

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	testHeaderStyle  = shared.RepoNameStyle
	testPassStyle    = shared.SuccessStyle
	testFailStyle    = shared.ErrorStyle
	testSkipStyle    = shared.WarningStyle
	testDimStyle     = shared.DimStyle
	testCovHighStyle = shared.CoverageHighStyle
	testCovMedStyle  = shared.CoverageMedStyle
	testCovLowStyle  = shared.CoverageLowStyle
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
	Short: i18n.T("cmd.test.short"),
	Long:  i18n.T("cmd.test.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTest(testVerbose, testCoverage, testShort, testPkg, testRun, testRace, testJSON)
	},
}

func init() {
	testCmd.Flags().BoolVar(&testVerbose, "verbose", false, i18n.T("cmd.test.flag.verbose"))
	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, i18n.T("cmd.test.flag.coverage"))
	testCmd.Flags().BoolVar(&testShort, "short", false, i18n.T("cmd.test.flag.short"))
	testCmd.Flags().StringVar(&testPkg, "pkg", "", i18n.T("cmd.test.flag.pkg"))
	testCmd.Flags().StringVar(&testRun, "run", "", i18n.T("cmd.test.flag.run"))
	testCmd.Flags().BoolVar(&testRace, "race", false, i18n.T("cmd.test.flag.race"))
	testCmd.Flags().BoolVar(&testJSON, "json", false, i18n.T("cmd.test.flag.json"))
}
