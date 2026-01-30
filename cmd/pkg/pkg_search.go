package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cache"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"github.com/spf13/cobra"
)

var (
	searchOrg     string
	searchPattern string
	searchType    string
	searchLimit   int
	searchRefresh bool
)

// addPkgSearchCommand adds the 'pkg search' command.
func addPkgSearchCommand(parent *cobra.Command) {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: i18n.T("cmd.pkg.search.short"),
		Long:  i18n.T("cmd.pkg.search.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			org := searchOrg
			pattern := searchPattern
			limit := searchLimit
			if org == "" {
				org = "host-uk"
			}
			if pattern == "" {
				pattern = "*"
			}
			if limit == 0 {
				limit = 50
			}
			return runPkgSearch(org, pattern, searchType, limit, searchRefresh)
		},
	}

	searchCmd.Flags().StringVar(&searchOrg, "org", "", i18n.T("cmd.pkg.search.flag.org"))
	searchCmd.Flags().StringVar(&searchPattern, "pattern", "", i18n.T("cmd.pkg.search.flag.pattern"))
	searchCmd.Flags().StringVar(&searchType, "type", "", i18n.T("cmd.pkg.search.flag.type"))
	searchCmd.Flags().IntVar(&searchLimit, "limit", 0, i18n.T("cmd.pkg.search.flag.limit"))
	searchCmd.Flags().BoolVar(&searchRefresh, "refresh", false, i18n.T("cmd.pkg.search.flag.refresh"))

	parent.AddCommand(searchCmd)
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
			fmt.Printf("%s %s %s\n", dimStyle.Render(i18n.T("cmd.pkg.search.cache_label")), org, dimStyle.Render(fmt.Sprintf("(%s ago)", age.Round(time.Second))))
		}
	}

	// Fetch from GitHub if not cached
	if !fromCache {
		if !ghAuthenticated() {
			return fmt.Errorf(i18n.T("cmd.pkg.error.gh_not_authenticated"))
		}

		if os.Getenv("GH_TOKEN") != "" {
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("common.label.note")), i18n.T("cmd.pkg.search.gh_token_warning"))
			fmt.Printf("%s %s\n\n", dimStyle.Render(""), i18n.T("cmd.pkg.search.gh_token_unset"))
		}

		fmt.Printf("%s %s... ", dimStyle.Render(i18n.T("cmd.pkg.search.fetching_label")), org)

		cmd := exec.Command("gh", "repo", "list", org,
			"--json", "name,description,visibility,updatedAt,primaryLanguage",
			"--limit", fmt.Sprintf("%d", limit))
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Println()
			errStr := strings.TrimSpace(string(output))
			if strings.Contains(errStr, "401") || strings.Contains(errStr, "Bad credentials") {
				return fmt.Errorf(i18n.T("cmd.pkg.error.auth_failed"))
			}
			return fmt.Errorf(i18n.T("cmd.pkg.error.search_failed"), errStr)
		}

		if err := json.Unmarshal(output, &ghRepos); err != nil {
			return fmt.Errorf(i18n.T("common.error.failed", map[string]any{"Action": "parse results"}), err)
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
		fmt.Println(i18n.T("cmd.pkg.search.no_repos_found"))
		return nil
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	fmt.Printf(i18n.T("cmd.pkg.search.found_repos", map[string]int{"Count": len(filtered)}) + "\n\n")

	for _, r := range filtered {
		visibility := ""
		if r.Visibility == "private" {
			visibility = dimStyle.Render(" " + i18n.T("cmd.pkg.search.private_label"))
		}

		desc := r.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		if desc == "" {
			desc = dimStyle.Render(i18n.T("cmd.pkg.no_description"))
		}

		fmt.Printf("  %s%s\n", repoNameStyle.Render(r.Name), visibility)
		fmt.Printf("    %s\n", desc)
	}

	fmt.Println()
	fmt.Printf("%s %s\n", i18n.T("common.hint.install_with"), dimStyle.Render(fmt.Sprintf("core pkg install %s/<repo-name>", org)))

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
