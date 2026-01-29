// Package pkg provides package management commands for core-* repos.
package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/cmd/core/cmd/shared"
	"github.com/host-uk/core/pkg/cache"
	"github.com/host-uk/core/pkg/repos"
	"github.com/leaanthony/clir"
)

// Style and utility aliases
var (
	repoNameStyle   = shared.RepoNameStyle
	successStyle    = shared.SuccessStyle
	errorStyle      = shared.ErrorStyle
	dimStyle        = shared.DimStyle
	ghAuthenticated = shared.GhAuthenticated
	gitClone        = shared.GitClone
)

// AddPkgCommands adds the 'pkg' command and subcommands for package management.
func AddPkgCommands(parent *clir.Cli) {
	pkgCmd := parent.NewSubCommand("pkg", "Package management for core-* repos")
	pkgCmd.LongDescription("Manage host-uk/core-* packages and repositories.\n\n" +
		"Commands:\n" +
		"  search    Search GitHub for packages\n" +
		"  install   Clone a package from GitHub\n" +
		"  list      List installed packages\n" +
		"  update    Update installed packages\n" +
		"  outdated  Check for outdated packages")

	addPkgSearchCommand(pkgCmd)
	addPkgInstallCommand(pkgCmd)
	addPkgListCommand(pkgCmd)
	addPkgUpdateCommand(pkgCmd)
	addPkgOutdatedCommand(pkgCmd)
}

// addPkgSearchCommand adds the 'pkg search' command.
func addPkgSearchCommand(parent *clir.Command) {
	var org string
	var pattern string
	var repoType string
	var limit int
	var refresh bool

	searchCmd := parent.NewSubCommand("search", "Search GitHub for packages")
	searchCmd.LongDescription("Searches GitHub for repositories matching a pattern.\n" +
		"Uses gh CLI for authenticated search. Results are cached for 1 hour.\n\n" +
		"Examples:\n" +
		"  core pkg search                           # List all host-uk repos\n" +
		"  core pkg search --pattern 'core-*'        # Search for core-* repos\n" +
		"  core pkg search --org mycompany           # Search different org\n" +
		"  core pkg search --refresh                 # Bypass cache")

	searchCmd.StringFlag("org", "GitHub organization (default: host-uk)", &org)
	searchCmd.StringFlag("pattern", "Repo name pattern (* for wildcard)", &pattern)
	searchCmd.StringFlag("type", "Filter by type in name (mod, services, plug, website)", &repoType)
	searchCmd.IntFlag("limit", "Max results (default 50)", &limit)
	searchCmd.BoolFlag("refresh", "Bypass cache and fetch fresh data", &refresh)

	searchCmd.Action(func() error {
		if org == "" {
			org = "host-uk"
		}
		if pattern == "" {
			pattern = "*"
		}
		if limit == 0 {
			limit = 50
		}
		return runPkgSearch(org, pattern, repoType, limit, refresh)
	})
}

type ghRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	UpdatedAt   string `json:"updated_at"`
	Language    string `json:"language"`
}

func runPkgSearch(org, pattern, repoType string, limit int, refresh bool) error {
	// Initialize cache in workspace .core/ directory
	var cacheDir string
	if regPath, err := repos.FindRegistry(); err == nil {
		cacheDir = filepath.Join(filepath.Dir(regPath), ".core", "cache")
	}

	c, err := cache.New(cacheDir, 0)
	if err != nil {
		c = nil
	}

	cacheKey := cache.GitHubReposKey(org)
	var ghRepos []ghRepo
	var fromCache bool

	// Try cache first (unless refresh requested)
	if c != nil && !refresh {
		if found, err := c.Get(cacheKey, &ghRepos); found && err == nil {
			fromCache = true
			age := c.Age(cacheKey)
			fmt.Printf("%s %s %s\n", dimStyle.Render("Cache:"), org, dimStyle.Render(fmt.Sprintf("(%s ago)", age.Round(time.Second))))
		}
	}

	// Fetch from GitHub if not cached
	if !fromCache {
		if !ghAuthenticated() {
			return fmt.Errorf("gh CLI not authenticated. Run: gh auth login")
		}

		if os.Getenv("GH_TOKEN") != "" {
			fmt.Printf("%s GH_TOKEN env var is set - this may cause auth issues\n", dimStyle.Render("Note:"))
			fmt.Printf("%s Unset it with: unset GH_TOKEN\n\n", dimStyle.Render(""))
		}

		fmt.Printf("%s %s... ", dimStyle.Render("Fetching:"), org)

		cmd := exec.Command("gh", "repo", "list", org,
			"--json", "name,description,visibility,updatedAt,primaryLanguage",
			"--limit", fmt.Sprintf("%d", limit))
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Println()
			errStr := strings.TrimSpace(string(output))
			if strings.Contains(errStr, "401") || strings.Contains(errStr, "Bad credentials") {
				return fmt.Errorf("authentication failed - try: unset GH_TOKEN && gh auth login")
			}
			return fmt.Errorf("search failed: %s", errStr)
		}

		if err := json.Unmarshal(output, &ghRepos); err != nil {
			return fmt.Errorf("failed to parse results: %w", err)
		}

		if c != nil {
			_ = c.Set(cacheKey, ghRepos)
		}

		fmt.Printf("%s\n", successStyle.Render("✓"))
	}

	// Filter by glob pattern and type
	var filtered []ghRepo
	for _, r := range ghRepos {
		if !matchGlob(pattern, r.Name) {
			continue
		}
		if repoType != "" && !strings.Contains(r.Name, repoType) {
			continue
		}
		filtered = append(filtered, r)
	}

	if len(filtered) == 0 {
		fmt.Println("No repositories found matching pattern.")
		return nil
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	fmt.Printf("Found %d repositories:\n\n", len(filtered))

	for _, r := range filtered {
		visibility := ""
		if r.Visibility == "private" {
			visibility = dimStyle.Render(" [private]")
		}

		desc := r.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		if desc == "" {
			desc = dimStyle.Render("(no description)")
		}

		fmt.Printf("  %s%s\n", repoNameStyle.Render(r.Name), visibility)
		fmt.Printf("    %s\n", desc)
	}

	fmt.Println()
	fmt.Printf("Install with: %s\n", dimStyle.Render(fmt.Sprintf("core pkg install %s/<repo-name>", org)))

	return nil
}

// matchGlob does simple glob matching with * wildcards
func matchGlob(pattern, name string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}

	parts := strings.Split(pattern, "*")
	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(name[pos:], part)
		if idx == -1 {
			return false
		}
		if i == 0 && !strings.HasPrefix(pattern, "*") && idx != 0 {
			return false
		}
		pos += idx + len(part)
	}
	if !strings.HasSuffix(pattern, "*") && pos != len(name) {
		return false
	}
	return true
}

// addPkgInstallCommand adds the 'pkg install' command.
func addPkgInstallCommand(parent *clir.Command) {
	var targetDir string
	var addToRegistry bool

	installCmd := parent.NewSubCommand("install", "Clone a package from GitHub")
	installCmd.LongDescription("Clones a repository from GitHub.\n\n" +
		"Examples:\n" +
		"  core pkg install host-uk/core-php\n" +
		"  core pkg install host-uk/core-tenant --dir ./packages\n" +
		"  core pkg install host-uk/core-admin --add")

	installCmd.StringFlag("dir", "Target directory (default: ./packages or current dir)", &targetDir)
	installCmd.BoolFlag("add", "Add to repos.yaml registry", &addToRegistry)

	installCmd.Action(func() error {
		args := installCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("repository is required (e.g., core pkg install host-uk/core-php)")
		}
		return runPkgInstall(args[0], targetDir, addToRegistry)
	})
}

func runPkgInstall(repoArg, targetDir string, addToRegistry bool) error {
	ctx := context.Background()

	// Parse org/repo
	parts := strings.Split(repoArg, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: use org/repo (e.g., host-uk/core-php)")
	}
	org, repoName := parts[0], parts[1]

	// Determine target directory
	if targetDir == "" {
		if regPath, err := repos.FindRegistry(); err == nil {
			if reg, err := repos.LoadRegistry(regPath); err == nil {
				targetDir = reg.BasePath
				if targetDir == "" {
					targetDir = "./packages"
				}
				if !filepath.IsAbs(targetDir) {
					targetDir = filepath.Join(filepath.Dir(regPath), targetDir)
				}
			}
		}
		if targetDir == "" {
			targetDir = "."
		}
	}

	if strings.HasPrefix(targetDir, "~/") {
		home, _ := os.UserHomeDir()
		targetDir = filepath.Join(home, targetDir[2:])
	}

	repoPath := filepath.Join(targetDir, repoName)

	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		fmt.Printf("%s %s already exists at %s\n", dimStyle.Render("Skip:"), repoName, repoPath)
		return nil
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("%s %s/%s\n", dimStyle.Render("Installing:"), org, repoName)
	fmt.Printf("%s %s\n", dimStyle.Render("Target:"), repoPath)
	fmt.Println()

	fmt.Printf("  %s... ", dimStyle.Render("Cloning"))
	err := gitClone(ctx, org, repoName, repoPath)
	if err != nil {
		fmt.Printf("%s\n", errorStyle.Render("✗ "+err.Error()))
		return err
	}
	fmt.Printf("%s\n", successStyle.Render("✓"))

	if addToRegistry {
		if err := addToRegistryFile(org, repoName); err != nil {
			fmt.Printf("  %s add to registry: %s\n", errorStyle.Render("✗"), err)
		} else {
			fmt.Printf("  %s added to repos.yaml\n", successStyle.Render("✓"))
		}
	}

	fmt.Println()
	fmt.Printf("%s Installed %s\n", successStyle.Render("Done:"), repoName)

	return nil
}

func addToRegistryFile(org, repoName string) error {
	regPath, err := repos.FindRegistry()
	if err != nil {
		return fmt.Errorf("no repos.yaml found")
	}

	reg, err := repos.LoadRegistry(regPath)
	if err != nil {
		return err
	}

	if _, exists := reg.Get(repoName); exists {
		return nil
	}

	f, err := os.OpenFile(regPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	repoType := detectRepoType(repoName)
	entry := fmt.Sprintf("\n  %s:\n    type: %s\n    description: (installed via core pkg install)\n",
		repoName, repoType)

	_, err = f.WriteString(entry)
	return err
}

func detectRepoType(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "-mod-") || strings.HasSuffix(lower, "-mod") {
		return "module"
	}
	if strings.Contains(lower, "-plug-") || strings.HasSuffix(lower, "-plug") {
		return "plugin"
	}
	if strings.Contains(lower, "-services-") || strings.HasSuffix(lower, "-services") {
		return "service"
	}
	if strings.Contains(lower, "-website-") || strings.HasSuffix(lower, "-website") {
		return "website"
	}
	if strings.HasPrefix(lower, "core-") {
		return "package"
	}
	return "package"
}

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
	var outdatedList []string

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
			outdatedList = append(outdatedList, r.Name)
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
