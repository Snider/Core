package ci

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/release"
)

// runCIReleaseVersion shows the determined version.
func runCIReleaseVersion() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	version, err := release.DetermineVersion(projectDir)
	if err != nil {
		return fmt.Errorf("failed to determine version: %w", err)
	}

	fmt.Printf("Version: %s\n", releaseValueStyle.Render(version))
	return nil
}
