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
		return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
	}

	version, err := release.DetermineVersion(projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("cmd.ci.error.determine_version"), err)
	}

	fmt.Printf("%s %s\n", i18n.T("common.label.version"), releaseValueStyle.Render(version))
	return nil
}
