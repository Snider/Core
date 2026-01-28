package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Snider/Core/pkg/cache"
	"github.com/Snider/Core/pkg/repos"
	"github.com/leaanthony/clir"
)

// AddSearchCommand adds the 'search' command to the given parent command.
func AddSearchCommand(parent *clir.Cli) {
	var org string
	var pattern string
	var repoType string
	var limit int
	var refresh bool

	searchCmd := parent.NewSubCommand("search", "Search GitHub for repos by pattern")
	searchCmd.LongDescription("Searches GitHub for repositories matching a pattern.\n" +
		"Uses gh CLI for authenticated search. Results are cached for 1 hour.\n\n" +
		"Examples:\n" +
		"  core search --org host-uk --pattern 'core-*'\n" +
		"  core search --org mycompany --pattern '*-mod-*'\n" +
		"  core search --org letheanvpn --refresh")

	searchCmd.StringFlag("org", "GitHub organization to search (required)", &org)
	searchCmd.StringFlag("pattern", "Repo name pattern (* for wildcard)", &pattern)
	searchCmd.StringFlag("type", "Filter by type in name (mod, services, plug, website)", &repoType)
	searchCmd.IntFlag("limit", "Max results (default 50)", &limit)
	searchCmd.BoolFlag("refresh", "Bypass cache and fetch fresh data", &refresh)

	searchCmd.Action(func() error {
		if org == "" {
			return fmt.Errorf("--org is required")
		}
		if pattern == "" {
			pattern = "*"
		}
		if limit == 0 {
			limit = 50
		}
		return runSearch(org, pattern, repoType, limit, refresh)
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

func runSearch(org, pattern, repoType string, limit int, refresh bool) error {
	// Initialize cache in workspace .core/ directory
	var cacheDir string
	if regPath, err := repos.FindRegistry(); err == nil {
		cacheDir = filepath.Join(filepath.Dir(regPath), ".core", "cache")
	}

	c, err := cache.New(cacheDir, 0)
	if err != nil {
		// Cache init failed, continue without cache
		c = nil
	}

	cacheKey := cache.GitHubReposKey(org)
	var repos []ghRepo
	var fromCache bool

	// Try cache first (unless refresh requested)
	if c != nil && !refresh {
		if found, err := c.Get(cacheKey, &repos); found && err == nil {
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

		// Check for bad GH_TOKEN which can override keyring auth
		if os.Getenv("GH_TOKEN") != "" {
			fmt.Printf("%s GH_TOKEN env var is set - this may cause auth issues\n", dimStyle.Render("Note:"))
			fmt.Printf("%s Unset it with: unset GH_TOKEN\n\n", dimStyle.Render(""))
		}

		fmt.Printf("%s %s... ", dimStyle.Render("Fetching:"), org)

		// Always use gh repo list (more reliable than gh search repos)
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

		if err := json.Unmarshal(output, &repos); err != nil {
			return fmt.Errorf("failed to parse results: %w", err)
		}

		// Cache the results
		if c != nil {
			_ = c.Set(cacheKey, repos)
		}

		fmt.Printf("%s\n", successStyle.Render("✓"))
	}

	// Filter by glob pattern and type
	var filtered []ghRepo
	for _, r := range repos {
		// Check glob pattern
		if !matchGlob(pattern, r.Name) {
			continue
		}
		// Check type filter (e.g., "mod" matches "*-mod-*" or "*-mod")
		if repoType != "" && !strings.Contains(r.Name, repoType) {
			continue
		}
		filtered = append(filtered, r)
	}

	if len(filtered) == 0 {
		fmt.Println("No repositories found matching pattern.")
		return nil
	}

	// Sort by name
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	// Display results
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
	fmt.Printf("Install with: %s\n", dimStyle.Render(fmt.Sprintf("core install %s/<repo-name>", org)))

	return nil
}

// matchGlob does simple glob matching with * wildcards
func matchGlob(pattern, name string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}

	// Simple glob: split by * and check if all parts exist in order
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

		// First part must be at start if pattern doesn't start with *
		if i == 0 && !strings.HasPrefix(pattern, "*") && idx != 0 {
			return false
		}

		pos += idx + len(part)
	}

	// Last part must be at end if pattern doesn't end with *
	if !strings.HasSuffix(pattern, "*") && pos != len(name) {
		return false
	}

	return true
}
