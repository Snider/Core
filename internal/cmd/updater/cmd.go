package updater

import (
	"context"
	"fmt"
	"runtime"

	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

// Repository configuration for updates
const (
	repoOwner = "host-uk"
	repoName  = "core"
)

// Command flags
var (
	updateChannel  string
	updateForce    bool
	updateCheck    bool
	updateWatchPID int
)

func init() {
	cli.RegisterCommands(AddUpdateCommands)
}

// AddUpdateCommands registers the update command and subcommands.
func AddUpdateCommands(root *cobra.Command) {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update core CLI to the latest version",
		Long: `Update the core CLI to the latest version from GitHub releases.

By default, checks the 'stable' channel for tagged releases (v*.*.*)
Use --channel=dev for the latest development build.

Examples:
  core update              # Update to latest stable release
  core update --check      # Check for updates without applying
  core update --channel=dev   # Update to latest dev build
  core update --force      # Force update even if already on latest`,
		RunE: runUpdate,
	}

	updateCmd.PersistentFlags().StringVar(&updateChannel, "channel", "stable", "Release channel: stable, beta, alpha, or dev")
	updateCmd.PersistentFlags().BoolVar(&updateForce, "force", false, "Force update even if already on latest version")
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check for updates, don't apply")
	updateCmd.Flags().IntVar(&updateWatchPID, "watch-pid", 0, "Internal: watch for parent PID to die then restart")
	_ = updateCmd.Flags().MarkHidden("watch-pid")

	updateCmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Check for available updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			updateCheck = true
			return runUpdate(cmd, args)
		},
	})

	root.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// If we're in watch mode, wait for parent to die then restart
	if updateWatchPID > 0 {
		return watchAndRestart(updateWatchPID)
	}

	currentVersion := cli.AppVersion

	cli.Print("%s %s\n", cli.DimStyle.Render("Current version:"), cli.ValueStyle.Render(currentVersion))
	cli.Print("%s %s/%s\n", cli.DimStyle.Render("Platform:"), runtime.GOOS, runtime.GOARCH)
	cli.Print("%s %s\n\n", cli.DimStyle.Render("Channel:"), updateChannel)

	// Handle dev channel specially - it's a prerelease tag, not a semver channel
	if updateChannel == "dev" {
		return handleDevUpdate(currentVersion)
	}

	// Check for newer version
	release, updateAvailable, err := CheckForNewerVersion(repoOwner, repoName, updateChannel, true)
	if err != nil {
		return cli.Wrap(err, "failed to check for updates")
	}

	if release == nil {
		cli.Print("%s No releases found in %s channel\n", cli.WarningStyle.Render("!"), updateChannel)
		return nil
	}

	if !updateAvailable && !updateForce {
		cli.Print("%s Already on latest version (%s)\n",
			cli.SuccessStyle.Render(cli.Glyph(":check:")),
			release.TagName)
		return nil
	}

	cli.Print("%s %s\n", cli.DimStyle.Render("Latest version:"), cli.SuccessStyle.Render(release.TagName))

	if updateCheck {
		if updateAvailable {
			cli.Print("\n%s Update available: %s → %s\n",
				cli.WarningStyle.Render("!"),
				currentVersion,
				release.TagName)
			cli.Print("Run %s to update\n", cli.ValueStyle.Render("core update"))
		}
		return nil
	}

	// Spawn watcher before applying update
	if err := spawnWatcher(); err != nil {
		// If watcher fails, continue anyway - update will still work
		cli.Print("%s Could not spawn restart watcher: %v\n", cli.DimStyle.Render("!"), err)
	}

	// Apply update
	cli.Print("\n%s Downloading update...\n", cli.DimStyle.Render("→"))

	downloadURL, err := GetDownloadURL(release, "")
	if err != nil {
		return cli.Wrap(err, "failed to get download URL")
	}

	if err := DoUpdate(downloadURL); err != nil {
		return cli.Wrap(err, "failed to apply update")
	}

	cli.Print("%s Updated to %s\n", cli.SuccessStyle.Render(cli.Glyph(":check:")), release.TagName)
	cli.Print("%s Restarting...\n", cli.DimStyle.Render("→"))

	return nil
}

// handleDevUpdate handles updates from the dev release (rolling prerelease)
func handleDevUpdate(currentVersion string) error {
	client := NewGithubClient()

	// Fetch the dev release directly by tag
	release, err := client.GetLatestRelease(context.TODO(), repoOwner, repoName, "beta")
	if err != nil {
		// Try fetching the "dev" tag directly
		return handleDevTagUpdate(currentVersion)
	}

	if release == nil {
		return handleDevTagUpdate(currentVersion)
	}

	cli.Print("%s %s\n", cli.DimStyle.Render("Latest dev:"), cli.ValueStyle.Render(release.TagName))

	if updateCheck {
		cli.Print("\nRun %s to update\n", cli.ValueStyle.Render("core update --channel=dev"))
		return nil
	}

	// Spawn watcher before applying update
	if err := spawnWatcher(); err != nil {
		cli.Print("%s Could not spawn restart watcher: %v\n", cli.DimStyle.Render("!"), err)
	}

	cli.Print("\n%s Downloading update...\n", cli.DimStyle.Render("→"))

	downloadURL, err := GetDownloadURL(release, "")
	if err != nil {
		return cli.Wrap(err, "failed to get download URL")
	}

	if err := DoUpdate(downloadURL); err != nil {
		return cli.Wrap(err, "failed to apply update")
	}

	cli.Print("%s Updated to %s\n", cli.SuccessStyle.Render(cli.Glyph(":check:")), release.TagName)
	cli.Print("%s Restarting...\n", cli.DimStyle.Render("→"))

	return nil
}

// handleDevTagUpdate fetches the dev release using the direct tag
func handleDevTagUpdate(currentVersion string) error {
	// Construct download URL directly for dev release
	downloadURL := fmt.Sprintf(
		"https://github.com/%s/%s/releases/download/dev/core-%s-%s",
		repoOwner, repoName, runtime.GOOS, runtime.GOARCH,
	)

	if runtime.GOOS == "windows" {
		downloadURL += ".exe"
	}

	cli.Print("%s dev (rolling)\n", cli.DimStyle.Render("Latest:"))

	if updateCheck {
		cli.Print("\nRun %s to update\n", cli.ValueStyle.Render("core update --channel=dev"))
		return nil
	}

	// Spawn watcher before applying update
	if err := spawnWatcher(); err != nil {
		cli.Print("%s Could not spawn restart watcher: %v\n", cli.DimStyle.Render("!"), err)
	}

	cli.Print("\n%s Downloading from dev release...\n", cli.DimStyle.Render("→"))

	if err := DoUpdate(downloadURL); err != nil {
		return cli.Wrap(err, "failed to apply update")
	}

	cli.Print("%s Updated to latest dev build\n", cli.SuccessStyle.Render(cli.Glyph(":check:")))
	cli.Print("%s Restarting...\n", cli.DimStyle.Render("→"))

	return nil
}
