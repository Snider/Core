package ci

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

// runChangelog generates and prints a changelog.
func runChangelog(fromRef, toRef string) error {
	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
	}

	// Load config for changelog settings
	cfg, err := release.LoadConfig(projectDir)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "config"), err)
	}

	// Generate changelog
	changelog, err := release.GenerateWithConfig(projectDir, fromRef, toRef, &cfg.Changelog)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.generate", "changelog"), err)
	}

	fmt.Println(changelog)
	return nil
}
