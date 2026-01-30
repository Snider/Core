package ci

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// runCIReleaseVersion shows the determined version.
func runCIReleaseVersion() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
	}

	version, err := release.DetermineVersion(projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.determine", "version"), err)
	}

	fmt.Printf("%s %s\n", i18n.Label("version"), releaseValueStyle.Render(version))
	return nil
}
