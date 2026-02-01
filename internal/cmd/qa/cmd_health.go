// cmd_health.go implements the 'qa health' command for aggregate CI health.
//
// Usage:
//   core qa health              # Show CI health summary
//   core qa health --problems   # Show only repos with problems

package qa

import (
	"encoding/json"
	"os/exec"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/errors"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Health command flags
var (
	healthProblems bool
	healthRegistry string
)

// HealthWorkflowRun represents a GitHub Actions workflow run
type HealthWorkflowRun struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
	HeadSha    string `json:"headSha"`
	UpdatedAt  string `json:"updatedAt"`
	URL        string `json:"url"`
}

// RepoHealth represents the CI health of a single repo
type RepoHealth struct {
	Name         string
	Status       string // "passing", "failing", "pending", "no_ci", "disabled"
	Message      string
	URL          string
	FailingSince string
}

// addHealthCommand adds the 'health' subcommand to qa.
func addHealthCommand(parent *cli.Command) {
	healthCmd := &cli.Command{
		Use:   "health",
		Short: i18n.T("cmd.qa.health.short"),
		Long:  i18n.T("cmd.qa.health.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runHealth()
		},
	}

	healthCmd.Flags().BoolVarP(&healthProblems, "problems", "p", false, i18n.T("cmd.qa.health.flag.problems"))
	healthCmd.Flags().StringVar(&healthRegistry, "registry", "", i18n.T("common.flag.registry"))

	parent.AddCommand(healthCmd)
}

func runHealth() error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.E("qa.health", i18n.T("error.gh_not_found"), nil)
	}

	// Load registry
	var reg *repos.Registry
	var err error

	if healthRegistry != "" {
		reg, err = repos.LoadRegistry(healthRegistry)
	} else {
		registryPath, findErr := repos.FindRegistry()
		if findErr != nil {
			return errors.E("qa.health", i18n.T("error.registry_not_found"), nil)
		}
		reg, err = repos.LoadRegistry(registryPath)
	}
	if err != nil {
		return errors.E("qa.health", "failed to load registry", err)
	}

	// Fetch CI status from all repos
	var healthResults []RepoHealth
	repoList := reg.List()

	for i, repo := range repoList {
		cli.Print("\033[2K\r%s %d/%d %s",
			dimStyle.Render(i18n.T("cmd.qa.issues.fetching")),
			i+1, len(repoList), repo.Name)

		health := fetchRepoHealth(reg.Org, repo.Name)
		healthResults = append(healthResults, health)
	}
	cli.Print("\033[2K\r") // Clear progress

	// Sort: problems first, then passing
	sort.Slice(healthResults, func(i, j int) bool {
		return healthPriority(healthResults[i].Status) < healthPriority(healthResults[j].Status)
	})

	// Filter if --problems flag
	if healthProblems {
		var problems []RepoHealth
		for _, h := range healthResults {
			if h.Status != "passing" {
				problems = append(problems, h)
			}
		}
		healthResults = problems
	}

	// Calculate summary
	passing := 0
	for _, h := range healthResults {
		if h.Status == "passing" {
			passing++
		}
	}
	total := len(repoList)
	percentage := 0
	if total > 0 {
		percentage = (passing * 100) / total
	}

	// Print summary
	cli.Print("%s: %d/%d repos healthy (%d%%)\n\n",
		i18n.T("cmd.qa.health.summary"),
		passing, total, percentage)

	if len(healthResults) == 0 {
		cli.Text(i18n.T("cmd.qa.health.all_healthy"))
		return nil
	}

	// Group by status
	grouped := make(map[string][]RepoHealth)
	for _, h := range healthResults {
		grouped[h.Status] = append(grouped[h.Status], h)
	}

	// Print problems first
	printHealthGroup("failing", grouped["failing"], errorStyle)
	printHealthGroup("pending", grouped["pending"], warningStyle)
	printHealthGroup("no_ci", grouped["no_ci"], dimStyle)
	printHealthGroup("disabled", grouped["disabled"], dimStyle)

	if !healthProblems {
		printHealthGroup("passing", grouped["passing"], successStyle)
	}

	return nil
}

func fetchRepoHealth(org, repoName string) RepoHealth {
	repoFullName := cli.Sprintf("%s/%s", org, repoName)

	args := []string{
		"run", "list",
		"--repo", repoFullName,
		"--limit", "1",
		"--json", "status,conclusion,name,headSha,updatedAt,url",
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a 404 (no workflows)
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "no workflows") || strings.Contains(stderr, "not found") {
				return RepoHealth{
					Name:    repoName,
					Status:  "no_ci",
					Message: i18n.T("cmd.qa.health.no_ci_configured"),
				}
			}
		}
		return RepoHealth{
			Name:    repoName,
			Status:  "no_ci",
			Message: i18n.T("cmd.qa.health.fetch_error"),
		}
	}

	var runs []HealthWorkflowRun
	if err := json.Unmarshal(output, &runs); err != nil {
		return RepoHealth{
			Name:    repoName,
			Status:  "no_ci",
			Message: i18n.T("cmd.qa.health.parse_error"),
		}
	}

	if len(runs) == 0 {
		return RepoHealth{
			Name:    repoName,
			Status:  "no_ci",
			Message: i18n.T("cmd.qa.health.no_ci_configured"),
		}
	}

	run := runs[0]
	health := RepoHealth{
		Name: repoName,
		URL:  run.URL,
	}

	switch run.Status {
	case "completed":
		switch run.Conclusion {
		case "success":
			health.Status = "passing"
			health.Message = i18n.T("cmd.qa.health.passing")
		case "failure":
			health.Status = "failing"
			health.Message = i18n.T("cmd.qa.health.tests_failing")
		case "cancelled":
			health.Status = "pending"
			health.Message = i18n.T("cmd.qa.health.cancelled")
		case "skipped":
			health.Status = "passing"
			health.Message = i18n.T("cmd.qa.health.skipped")
		default:
			health.Status = "failing"
			health.Message = run.Conclusion
		}
	case "in_progress", "queued", "waiting":
		health.Status = "pending"
		health.Message = i18n.T("cmd.qa.health.running")
	default:
		health.Status = "no_ci"
		health.Message = run.Status
	}

	return health
}

func healthPriority(status string) int {
	switch status {
	case "failing":
		return 0
	case "pending":
		return 1
	case "no_ci":
		return 2
	case "disabled":
		return 3
	case "passing":
		return 4
	default:
		return 5
	}
}

func printHealthGroup(status string, repos []RepoHealth, style *cli.AnsiStyle) {
	if len(repos) == 0 {
		return
	}

	var label string
	switch status {
	case "failing":
		label = i18n.T("cmd.qa.health.count_failing")
	case "pending":
		label = i18n.T("cmd.qa.health.count_pending")
	case "no_ci":
		label = i18n.T("cmd.qa.health.count_no_ci")
	case "disabled":
		label = i18n.T("cmd.qa.health.count_disabled")
	case "passing":
		label = i18n.T("cmd.qa.health.count_passing")
	}

	cli.Print("%s (%d):\n", style.Render(label), len(repos))
	for _, repo := range repos {
		cli.Print("  %s %s\n",
			cli.RepoStyle.Render(repo.Name),
			dimStyle.Render(repo.Message))
		if repo.URL != "" && status == "failing" {
			cli.Print("      -> %s\n", dimStyle.Render(repo.URL))
		}
	}
	cli.Blank()
}
