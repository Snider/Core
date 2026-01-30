package ci

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/release"
)

// runChangelog generates and prints a changelog.
func runChangelog(fromRef, toRef string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config for changelog settings
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Generate changelog
	changelog, err := release.GenerateWithConfig(projectDir, fromRef, toRef, &cfg.Changelog)
	if err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	fmt.Println(changelog)
	return nil
}
