package ci

import (
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// runCIReleaseVersion shows the determined version.
func runCIReleaseVersion() error {
	projectDir, err := os.Getwd()
	if err != nil {
		return cli.WrapVerb(err, "get", "working directory")
	}

	version, err := release.DetermineVersion(projectDir)
	if err != nil {
		return cli.WrapVerb(err, "determine", "version")
	}

	cli.Print("%s %s\n", i18n.Label("version"), releaseValueStyle.Render(version))
	return nil
}
