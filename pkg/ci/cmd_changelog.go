package ci

import (
	"os"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/release"
)

func runChangelog(fromRef, toRef string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
	}

	// Auto-detect refs if not provided
	if fromRef == "" || toRef == "" {
		tag, err := latestTag(cwd)
		if err == nil {
			if fromRef == "" {
				fromRef = tag
			}
			if toRef == "" {
				toRef = "HEAD"
			}
		} else {
			// No tags, use initial commit? Or just HEAD?
			cli.Text(i18n.T("cmd.ci.changelog.no_tags"))
			return nil
		}
	}

	cli.Print("%s %s..%s\n\n", releaseDimStyle.Render(i18n.T("cmd.ci.changelog.generating")), fromRef, toRef)

	// Generate changelog
	changelog, err := release.Generate(cwd, fromRef, toRef)
	if err != nil {
		return cli.Err("%s: %w", i18n.T("i18n.fail.generate", "changelog"), err)
	}

	cli.Text(changelog)

	return nil
}

func latestTag(dir string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}