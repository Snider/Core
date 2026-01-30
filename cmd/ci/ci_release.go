// Package ci provides release publishing commands.
package ci

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	releaseHeaderStyle  = cli.RepoNameStyle
	releaseSuccessStyle = cli.SuccessStyle
	releaseErrorStyle   = cli.ErrorStyle
	releaseDimStyle     = cli.DimStyle
	releaseValueStyle   = cli.ValueStyle
)

// Flag variables for ci command
var (
	ciGoForLaunch bool
	ciVersion     string
	ciDraft       bool
	ciPrerelease  bool
)

// Flag variables for changelog subcommand
var (
	changelogFromRef string
	changelogToRef   string
)

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: i18n.T("cmd.ci.short"),
	Long:  i18n.T("cmd.ci.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun := !ciGoForLaunch
		return runCIPublish(dryRun, ciVersion, ciDraft, ciPrerelease)
	},
}

var ciInitCmd = &cobra.Command{
	Use:   "init",
	Short: i18n.T("cmd.ci.init.short"),
	Long:  i18n.T("cmd.ci.init.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCIReleaseInit()
	},
}

var ciChangelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: i18n.T("cmd.ci.changelog.short"),
	Long:  i18n.T("cmd.ci.changelog.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog(changelogFromRef, changelogToRef)
	},
}

var ciVersionCmd = &cobra.Command{
	Use:   "version",
	Short: i18n.T("cmd.ci.version.short"),
	Long:  i18n.T("cmd.ci.version.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCIReleaseVersion()
	},
}

func init() {
	// Main ci command flags
	ciCmd.Flags().BoolVar(&ciGoForLaunch, "we-are-go-for-launch", false, i18n.T("cmd.ci.flag.go_for_launch"))
	ciCmd.Flags().StringVar(&ciVersion, "version", "", i18n.T("cmd.ci.flag.version"))
	ciCmd.Flags().BoolVar(&ciDraft, "draft", false, i18n.T("cmd.ci.flag.draft"))
	ciCmd.Flags().BoolVar(&ciPrerelease, "prerelease", false, i18n.T("cmd.ci.flag.prerelease"))

	// Changelog subcommand flags
	ciChangelogCmd.Flags().StringVar(&changelogFromRef, "from", "", i18n.T("cmd.ci.changelog.flag.from"))
	ciChangelogCmd.Flags().StringVar(&changelogToRef, "to", "", i18n.T("cmd.ci.changelog.flag.to"))

	// Add subcommands
	ciCmd.AddCommand(ciInitCmd)
	ciCmd.AddCommand(ciChangelogCmd)
	ciCmd.AddCommand(ciVersionCmd)
}
