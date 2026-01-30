package ci

import (
	"context"
	"fmt"
	"os"

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
		return fmt.Errorf("%s: %w", i18n.T("cmd.ci.error.working_dir"), err)
	}

	// Load configuration
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.ci.error.load_config"), err)
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
	fmt.Printf("%s %s\n", releaseHeaderStyle.Render(i18n.T("cmd.ci.label.ci")), i18n.T("cmd.ci.publishing"))
	if dryRun {
		fmt.Printf("  %s\n", releaseDimStyle.Render(i18n.T("cmd.ci.dry_run_hint")))
	} else {
		fmt.Printf("  %s\n", releaseSuccessStyle.Render(i18n.T("cmd.ci.go_for_launch")))
	}
	fmt.Println()

	// Check for publishers
	if len(cfg.Publishers) == 0 {
		return fmt.Errorf(i18n.T("cmd.ci.error.no_publishers"))
	}

	// Publish pre-built artifacts
	rel, err := release.Publish(ctx, cfg, dryRun)
	if err != nil {
		fmt.Printf("%s %v\n", releaseErrorStyle.Render(i18n.T("cmd.ci.label.error")), err)
		return err
	}

	// Print summary
	fmt.Println()
	fmt.Printf("%s %s\n", releaseSuccessStyle.Render(i18n.T("cmd.ci.label.success")), i18n.T("cmd.ci.publish_completed"))
	fmt.Printf("  %s   %s\n", i18n.T("cmd.ci.label.version"), releaseValueStyle.Render(rel.Version))
	fmt.Printf("  %s %d\n", i18n.T("cmd.ci.label.artifacts"), len(rel.Artifacts))

	if !dryRun {
		for _, pub := range cfg.Publishers {
			fmt.Printf("  %s %s\n", i18n.T("cmd.ci.label.published"), releaseValueStyle.Render(pub.Type))
		}
	}

	return nil
}
