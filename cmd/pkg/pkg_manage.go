package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/repos"
	"github.com/leaanthony/clir"
)

// addPkgListCommand adds the 'pkg list' command.
func addPkgListCommand(parent *clir.Command) {
	listCmd := parent.NewSubCommand("list", "List installed packages")
	listCmd.LongDescription("Lists all packages in the current workspace.\n\n" +
		"Reads from repos.yaml or scans for git repositories.\n\n" +
		"Examples:\n" +
		"  core pkg list")

	listCmd.Action(func() error {
		return runPkgList()
	})
}

func runPkgList() error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return fmt.Errorf("no repos.yaml found - run from workspace directory")
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
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
		fmt.Println("No packages in registry.")
		return nil
	}

	fmt.Printf("%s\n\n", repoNameStyle.Render("Installed Packages"))

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
			desc = dimStyle.Render("(no description)")
		}

		fmt.Printf("  %s %s\n", status, repoNameStyle.Render(r.Name))
		fmt.Printf("      %s\n", desc)
	}

	fmt.Println()
	fmt.Printf("%s %d installed, %d missing\n", dimStyle.Render("Total:"), installed, missing)

	if missing > 0 {
		fmt.Printf("\nInstall missing: %s\n", dimStyle.Render("core setup"))
	}

	return nil
}

// addPkgUpdateCommand adds the 'pkg update' command.
func addPkgUpdateCommand(parent *clir.Command) {
	var all bool

	updateCmd := parent.NewSubCommand("update", "Update installed packages")
	updateCmd.LongDescription("Pulls latest changes for installed packages.\n\n" +
		"Examples:\n" +
		"  core pkg update core-php       # Update specific package\n" +
		"  core pkg update --all          # Update all packages")

	updateCmd.BoolFlag("all", "Update all packages", &all)

	updateCmd.Action(func() error {
		args := updateCmd.OtherArgs()
		if !all && len(args) == 0 {
			return fmt.Errorf("specify package name or use --all")
		}
		return runPkgUpdate(args, all)
	})
}

func runPkgUpdate(packages []string, all bool) error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return fmt.Errorf("no repos.yaml found")
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
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

	fmt.Printf("%s Updating %d package(s)\n\n", dimStyle.Render("Update:"), len(toUpdate))

	var updated, skipped, failed int
	for _, name := range toUpdate {
		repoPath := filepath.Join(basePath, name)

		if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
			fmt.Printf("  %s %s (not installed)\n", dimStyle.Render("○"), name)
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
			fmt.Printf("%s\n", dimStyle.Render("up to date"))
		} else {
			fmt.Printf("%s\n", successStyle.Render("✓"))
		}
		updated++
	}

	fmt.Println()
	fmt.Printf("%s %d updated, %d skipped, %d failed\n",
		dimStyle.Render("Done:"), updated, skipped, failed)

	return nil
}

// addPkgOutdatedCommand adds the 'pkg outdated' command.
func addPkgOutdatedCommand(parent *clir.Command) {
	outdatedCmd := parent.NewSubCommand("outdated", "Check for outdated packages")
	outdatedCmd.LongDescription("Checks which packages have unpulled commits.\n\n" +
		"Examples:\n" +
		"  core pkg outdated")

	outdatedCmd.Action(func() error {
		return runPkgOutdated()
	})
}

func runPkgOutdated() error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return fmt.Errorf("no repos.yaml found")
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	basePath := reg.BasePath
	if basePath == "" {
		basePath = "."
	}
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(filepath.Dir(regPath), basePath)
	}

	fmt.Printf("%s Checking for updates...\n\n", dimStyle.Render("Outdated:"))

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
			fmt.Printf("  %s %s (%s commits behind)\n",
				errorStyle.Render("↓"), repoNameStyle.Render(r.Name), count)
			outdated++
		} else {
			upToDate++
		}
	}

	fmt.Println()
	if outdated == 0 {
		fmt.Printf("%s All packages up to date\n", successStyle.Render("Done:"))
	} else {
		fmt.Printf("%s %d outdated, %d up to date\n",
			dimStyle.Render("Summary:"), outdated, upToDate)
		fmt.Printf("\nUpdate with: %s\n", dimStyle.Render("core pkg update --all"))
	}

	return nil
}
