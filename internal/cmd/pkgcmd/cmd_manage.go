package pkgcmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

// addPkgListCommand adds the 'pkg list' command.
func addPkgListCommand(parent *cobra.Command) {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("cmd.pkg.list.short"),
		Long:  i18n.T("cmd.pkg.list.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPkgList()
		},
	}

	parent.AddCommand(listCmd)
}

func runPkgList() error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return errors.New(i18n.T("cmd.pkg.error.no_repos_yaml_workspace"))
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "registry"), err)
	}

	basePath := reg.BasePath
	if basePath == "" {
		basePath = "."
	}
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(filepath.Dir(regPath), basePath)
	}

	allRepos := reg.List()
	if len(allRepos) == 0 {
		fmt.Println(i18n.T("cmd.pkg.list.no_packages"))
		return nil
	}

	fmt.Printf("%s\n\n", repoNameStyle.Render(i18n.T("cmd.pkg.list.title")))

	var installed, missing int
	for _, r := range allRepos {
		repoPath := filepath.Join(basePath, r.Name)
		exists := false
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
			exists = true
			installed++
		} else {
			missing++
		}

		status := successStyle.Render("✓")
		if !exists {
			status = dimStyle.Render("○")
		}

		desc := r.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		if desc == "" {
			desc = dimStyle.Render(i18n.T("cmd.pkg.no_description"))
		}

		fmt.Printf("  %s %s\n", status, repoNameStyle.Render(r.Name))
		fmt.Printf("      %s\n", desc)
	}

	fmt.Println()
	fmt.Printf("%s %s\n", dimStyle.Render(i18n.Label("total")), i18n.T("cmd.pkg.list.summary", map[string]int{"Installed": installed, "Missing": missing}))

	if missing > 0 {
		fmt.Printf("\n%s %s\n", i18n.T("cmd.pkg.list.install_missing"), dimStyle.Render("core setup"))
	}

	return nil
}

var updateAll bool

// addPkgUpdateCommand adds the 'pkg update' command.
func addPkgUpdateCommand(parent *cobra.Command) {
	updateCmd := &cobra.Command{
		Use:   "update [packages...]",
		Short: i18n.T("cmd.pkg.update.short"),
		Long:  i18n.T("cmd.pkg.update.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !updateAll && len(args) == 0 {
				return errors.New(i18n.T("cmd.pkg.error.specify_package"))
			}
			return runPkgUpdate(args, updateAll)
		},
	}

	updateCmd.Flags().BoolVar(&updateAll, "all", false, i18n.T("cmd.pkg.update.flag.all"))

	parent.AddCommand(updateCmd)
}

func runPkgUpdate(packages []string, all bool) error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return errors.New(i18n.T("cmd.pkg.error.no_repos_yaml"))
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "registry"), err)
	}

	basePath := reg.BasePath
	if basePath == "" {
		basePath = "."
	}
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(filepath.Dir(regPath), basePath)
	}

	var toUpdate []string
	if all {
		for _, r := range reg.List() {
			toUpdate = append(toUpdate, r.Name)
		}
	} else {
		toUpdate = packages
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.pkg.update.update_label")), i18n.T("cmd.pkg.update.updating", map[string]int{"Count": len(toUpdate)}))

	var updated, skipped, failed int
	for _, name := range toUpdate {
		repoPath := filepath.Join(basePath, name)

		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("  %s %s (%s)\n", dimStyle.Render("○"), name, i18n.T("cmd.pkg.update.not_installed"))
			skipped++
			continue
		}

		fmt.Printf("  %s %s... ", dimStyle.Render("↓"), name)

		cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("✗"))
			fmt.Printf("      %s\n", strings.TrimSpace(string(output)))
			failed++
			continue
		}

		if strings.Contains(string(output), "Already up to date") {
			fmt.Printf("%s\n", dimStyle.Render(i18n.T("common.status.up_to_date")))
		} else {
			fmt.Printf("%s\n", successStyle.Render("✓"))
		}
		updated++
	}

	fmt.Println()
	fmt.Printf("%s %s\n",
		dimStyle.Render(i18n.T("i18n.done.update")), i18n.T("cmd.pkg.update.summary", map[string]int{"Updated": updated, "Skipped": skipped, "Failed": failed}))

	return nil
}

// addPkgOutdatedCommand adds the 'pkg outdated' command.
func addPkgOutdatedCommand(parent *cobra.Command) {
	outdatedCmd := &cobra.Command{
		Use:   "outdated",
		Short: i18n.T("cmd.pkg.outdated.short"),
		Long:  i18n.T("cmd.pkg.outdated.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPkgOutdated()
		},
	}

	parent.AddCommand(outdatedCmd)
}

func runPkgOutdated() error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return errors.New(i18n.T("cmd.pkg.error.no_repos_yaml"))
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("i18n.fail.load", "registry"), err)
	}

	basePath := reg.BasePath
	if basePath == "" {
		basePath = "."
	}
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(filepath.Dir(regPath), basePath)
	}

	fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.pkg.outdated.outdated_label")), i18n.T("common.progress.checking_updates"))

	var outdated, upToDate, notInstalled int

	for _, r := range reg.List() {
		repoPath := filepath.Join(basePath, r.Name)

		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			notInstalled++
			continue
		}

		// Fetch updates
		exec.Command("git", "-C", repoPath, "fetch", "--quiet").Run()

		// Check if behind
		cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD..@{u}")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		count := strings.TrimSpace(string(output))
		if count != "0" {
			fmt.Printf("  %s %s (%s)\n",
				errorStyle.Render("↓"), repoNameStyle.Render(r.Name), i18n.T("cmd.pkg.outdated.commits_behind", map[string]string{"Count": count}))
			outdated++
		} else {
			upToDate++
		}
	}

	fmt.Println()
	if outdated == 0 {
		fmt.Printf("%s %s\n", successStyle.Render(i18n.T("i18n.done.update")), i18n.T("cmd.pkg.outdated.all_up_to_date"))
	} else {
		fmt.Printf("%s %s\n",
			dimStyle.Render(i18n.Label("summary")), i18n.T("cmd.pkg.outdated.summary", map[string]int{"Outdated": outdated, "UpToDate": upToDate}))
		fmt.Printf("\n%s %s\n", i18n.T("cmd.pkg.outdated.update_with"), dimStyle.Render("core pkg update --all"))
	}

	return nil
}
