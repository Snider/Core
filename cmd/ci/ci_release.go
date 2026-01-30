// Package ci provides release publishing commands.
package ci

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases from shared
var (
	releaseHeaderStyle  = shared.RepoNameStyle
	releaseSuccessStyle = shared.SuccessStyle
	releaseErrorStyle   = shared.ErrorStyle
	releaseDimStyle     = shared.DimStyle
	releaseValueStyle   = shared.ValueStyle
)

// AddCIReleaseCommand adds the release command and its subcommands.
func AddCIReleaseCommand(app *clir.Cli) {
	releaseCmd := app.NewSubCommand("ci", "Publish releases (dry-run by default)")
	releaseCmd.LongDescription("Publishes pre-built artifacts from dist/ to configured targets.\n" +
		"Run 'core build' first to create artifacts.\n\n" +
		"SAFE BY DEFAULT: Runs in dry-run mode unless --we-are-go-for-launch is specified.\n\n" +
		"Configuration: .core/release.yaml")

	// Flags for the main release command
	var goForLaunch bool
	var version string
	var draft bool
	var prerelease bool

	releaseCmd.BoolFlag("we-are-go-for-launch", "Actually publish (default is dry-run for safety)", &goForLaunch)
	releaseCmd.StringFlag("version", "Version to release (e.g., v1.2.3)", &version)
	releaseCmd.BoolFlag("draft", "Create release as a draft", &draft)
	releaseCmd.BoolFlag("prerelease", "Mark release as a prerelease", &prerelease)

	// Default action for `core ci` - dry-run by default for safety
	releaseCmd.Action(func() error {
		dryRun := !goForLaunch
		return runCIPublish(dryRun, version, draft, prerelease)
	})

	// `release init` subcommand
	initCmd := releaseCmd.NewSubCommand("init", "Initialize release configuration")
	initCmd.LongDescription("Creates a .core/release.yaml configuration file interactively.")
	initCmd.Action(func() error {
		return runCIReleaseInit()
	})

	// `release changelog` subcommand
	changelogCmd := releaseCmd.NewSubCommand("changelog", "Generate changelog")
	changelogCmd.LongDescription("Generates a changelog from conventional commits.")
	var fromRef, toRef string
	changelogCmd.StringFlag("from", "Starting ref (default: previous tag)", &fromRef)
	changelogCmd.StringFlag("to", "Ending ref (default: HEAD)", &toRef)
	changelogCmd.Action(func() error {
		return runChangelog(fromRef, toRef)
	})

	// `release version` subcommand
	versionCmd := releaseCmd.NewSubCommand("version", "Show or set version")
	versionCmd.LongDescription("Shows the determined version or validates a version string.")
	versionCmd.Action(func() error {
		return runCIReleaseVersion()
	})
}
