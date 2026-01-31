package ci

import (
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/release"
)

// runChangelog generates and prints a changelog.
func runChangelog(fromRef, toRef string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return cli.WrapVerb(err, "get", "working directory")
	}

	// Load config for changelog settings
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return cli.WrapVerb(err, "load", "config")
	}

	// Generate changelog
	changelog, err := release.GenerateWithConfig(projectDir, fromRef, toRef, &cfg.Changelog)
	if err != nil {
		return cli.WrapVerb(err, "generate", "changelog")
	}

	cli.Text(changelog)
	return nil
}
