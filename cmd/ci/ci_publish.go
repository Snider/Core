package ci

import (
	"context"
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/release"
)

// runCIPublish publishes pre-built artifacts from dist/.
// It does NOT build - use `core build` first.
func runCIPublish(dryRun bool, version string, draft, prerelease bool) error {
	ctx := context.Background()

	// Get current directory
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load configuration
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
	fmt.Printf("%s Publishing release\n", releaseHeaderStyle.Render("CI:"))
	if dryRun {
		fmt.Printf("  %s\n", releaseDimStyle.Render("(dry-run) use --we-are-go-for-launch to publish"))
	} else {
		fmt.Printf("  %s\n", releaseSuccessStyle.Render("GO FOR LAUNCH"))
	}
	fmt.Println()

	// Check for publishers
	if len(cfg.Publishers) == 0 {
		return fmt.Errorf("no publishers configured in .core/release.yaml")
	}

	// Publish pre-built artifacts
	rel, err := release.Publish(ctx, cfg, dryRun)
	if err != nil {
		fmt.Printf("%s %v\n", releaseErrorStyle.Render("Error:"), err)
		return err
	}

	// Print summary
	fmt.Println()
	fmt.Printf("%s Publish completed!\n", releaseSuccessStyle.Render("Success:"))
	fmt.Printf("  Version:   %s\n", releaseValueStyle.Render(rel.Version))
	fmt.Printf("  Artifacts: %d\n", len(rel.Artifacts))

	if !dryRun {
		for _, pub := range cfg.Publishers {
			fmt.Printf("  Published: %s\n", releaseValueStyle.Render(pub.Type))
		}
	}

	return nil
}
