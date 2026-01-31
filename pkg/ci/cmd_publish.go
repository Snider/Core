package ci

import (
	"context"
	"errors"
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// runCIPublish publishes pre-built artifacts from dist/.
// It does NOT build - use `core build` first.
func runCIPublish(dryRun bool, version string, draft, prerelease bool) error {
	ctx := context.Background()

	// Get current directory
	projectDir, err := os.Getwd()
	if err != nil {
		return cli.WrapVerb(err, "get", "working directory")
	}

	// Load configuration
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return cli.WrapVerb(err, "load", "config")
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
	cli.Print("%s %s\n", releaseHeaderStyle.Render(i18n.T("cmd.ci.label.ci")), i18n.T("cmd.ci.publishing"))
	if dryRun {
		cli.Print("  %s\n", releaseDimStyle.Render(i18n.T("cmd.ci.dry_run_hint")))
	} else {
		cli.Print("  %s\n", releaseSuccessStyle.Render(i18n.T("cmd.ci.go_for_launch")))
	}
	cli.Line("")

	// Check for publishers
	if len(cfg.Publishers) == 0 {
		return errors.New(i18n.T("cmd.ci.error.no_publishers"))
	}

	// Publish pre-built artifacts
	rel, err := release.Publish(ctx, cfg, dryRun)
	if err != nil {
		cli.Print("%s %v\n", releaseErrorStyle.Render(i18n.Label("error")), err)
		return err
	}

	// Print summary
	cli.Line("")
	cli.Print("%s %s\n", releaseSuccessStyle.Render(i18n.T("i18n.done.pass")), i18n.T("cmd.ci.publish_completed"))
	cli.Print("  %s   %s\n", i18n.Label("version"), releaseValueStyle.Render(rel.Version))
	cli.Print("  %s %d\n", i18n.T("cmd.ci.label.artifacts"), len(rel.Artifacts))

	if !dryRun {
		for _, pub := range cfg.Publishers {
			cli.Print("  %s %s\n", i18n.T("cmd.ci.label.published"), releaseValueStyle.Render(pub.Type))
		}
	}

	return nil
}
