// Package ci provides release publishing commands.
package ci

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	releaseHeaderStyle  = shared.RepoNameStyle
	releaseSuccessStyle = shared.SuccessStyle
	releaseErrorStyle   = shared.ErrorStyle
	releaseDimStyle     = shared.DimStyle
	releaseValueStyle   = shared.ValueStyle
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
	Short: "Publish releases (dry-run by default)",
	Long: `Publishes pre-built artifacts from dist/ to configured targets.
Run 'core build' first to create artifacts.

SAFE BY DEFAULT: Runs in dry-run mode unless --we-are-go-for-launch is specified.

Configuration: .core/release.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun := !ciGoForLaunch
		return runCIPublish(dryRun, ciVersion, ciDraft, ciPrerelease)
	},
}

var ciInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize release configuration",
	Long:  "Creates a .core/release.yaml configuration file interactively.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCIReleaseInit()
	},
}

var ciChangelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate changelog",
	Long:  "Generates a changelog from conventional commits.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChangelog(changelogFromRef, changelogToRef)
	},
}

var ciVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show or set version",
	Long:  "Shows the determined version or validates a version string.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCIReleaseVersion()
	},
}

func init() {
	// Main ci command flags
	ciCmd.Flags().BoolVar(&ciGoForLaunch, "we-are-go-for-launch", false, "Actually publish (default is dry-run for safety)")
	ciCmd.Flags().StringVar(&ciVersion, "version", "", "Version to release (e.g., v1.2.3)")
	ciCmd.Flags().BoolVar(&ciDraft, "draft", false, "Create release as a draft")
	ciCmd.Flags().BoolVar(&ciPrerelease, "prerelease", false, "Mark release as a prerelease")

	// Changelog subcommand flags
	ciChangelogCmd.Flags().StringVar(&changelogFromRef, "from", "", "Starting ref (default: previous tag)")
	ciChangelogCmd.Flags().StringVar(&changelogToRef, "to", "", "Ending ref (default: HEAD)")

	// Add subcommands
	ciCmd.AddCommand(ciInitCmd)
	ciCmd.AddCommand(ciChangelogCmd)
	ciCmd.AddCommand(ciVersionCmd)
}
