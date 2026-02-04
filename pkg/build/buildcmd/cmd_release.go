// cmd_release.go implements the release command: build + archive + publish in one step.

package buildcmd

import (
	"context"
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/framework/core"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// Flag variables for release command
var (
	releaseVersion     string
	releaseDraft       bool
	releasePrerelease  bool
	releaseGoForLaunch bool
)

var releaseCmd = &cli.Command{
	Use:   "release",
	Short: i18n.T("cmd.build.release.short"),
	Long:  i18n.T("cmd.build.release.long"),
	RunE: func(cmd *cli.Command, args []string) error {
		return runRelease(cmd.Context(), !releaseGoForLaunch, releaseVersion, releaseDraft, releasePrerelease)
	},
}

func init() {
	releaseCmd.Flags().BoolVar(&releaseGoForLaunch, "we-are-go-for-launch", false, i18n.T("cmd.build.release.flag.go_for_launch"))
	releaseCmd.Flags().StringVar(&releaseVersion, "version", "", i18n.T("cmd.build.release.flag.version"))
	releaseCmd.Flags().BoolVar(&releaseDraft, "draft", false, i18n.T("cmd.build.release.flag.draft"))
	releaseCmd.Flags().BoolVar(&releasePrerelease, "prerelease", false, i18n.T("cmd.build.release.flag.prerelease"))
}

// AddReleaseCommand adds the release subcommand to the build command.
func AddReleaseCommand(buildCmd *cli.Command) {
	buildCmd.AddCommand(releaseCmd)
}

// runRelease executes the full release workflow: build + archive + checksum + publish.
func runRelease(ctx context.Context, dryRun bool, version string, draft, prerelease bool) error {
	// Get current directory
	projectDir, err := os.Getwd()
	if err != nil {
		return core.E("release", "get working directory", err)
	}

	// Check for release config
	if !release.ConfigExists(projectDir) {
		cli.Print("%s %s\n",
			buildErrorStyle.Render(i18n.Label("error")),
			i18n.T("cmd.build.release.error.no_config"),
		)
		cli.Print("  %s\n", buildDimStyle.Render(i18n.T("cmd.build.release.hint.create_config")))
		return core.E("release", "config not found", nil)
	}

	// Load configuration
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return core.E("release", "load config", err)
	}

	// Apply CLI overrides
	if version != "" {
		cfg.SetVersion(version)
	}

	// Apply draft/prerelease overrides to all publishers
	if draft || prerelease {
		for i := range cfg.Publishers {
			if draft {
				cfg.Publishers[i].Draft = true
			}
			if prerelease {
				cfg.Publishers[i].Prerelease = true
			}
		}
	}

	// Print header
	cli.Print("%s %s\n", buildHeaderStyle.Render(i18n.T("cmd.build.release.label.release")), i18n.T("cmd.build.release.building_and_publishing"))
	if dryRun {
		cli.Print("  %s\n", buildDimStyle.Render(i18n.T("cmd.build.release.dry_run_hint")))
	}
	cli.Blank()

	// Run full release (build + archive + checksum + publish)
	rel, err := release.Run(ctx, cfg, dryRun)
	if err != nil {
		return err
	}

	// Print summary
	cli.Blank()
	cli.Print("%s %s\n", buildSuccessStyle.Render(i18n.T("i18n.done.pass")), i18n.T("cmd.build.release.completed"))
	cli.Print("  %s   %s\n", i18n.Label("version"), buildTargetStyle.Render(rel.Version))
	cli.Print("  %s %d\n", i18n.T("cmd.build.release.label.artifacts"), len(rel.Artifacts))

	if !dryRun {
		for _, pub := range cfg.Publishers {
			cli.Print("  %s %s\n", i18n.T("cmd.build.release.label.published"), buildTargetStyle.Render(pub.Type))
		}
	}

	return nil
}
