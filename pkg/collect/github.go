package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
)

// ghIssue represents a GitHub issue or pull request as returned by the gh CLI.
type ghIssue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Author    ghAuthor  `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	Labels    []ghLabel `json:"labels"`
	URL       string    `json:"url"`
}

type ghAuthor struct {
	Login string `json:"login"`
}

type ghLabel struct {
	Name string `json:"name"`
}

// ghRepo represents a GitHub repository as returned by the gh CLI.
type ghRepo struct {
	Name string `json:"name"`
}

// GitHubCollector collects issues and PRs from GitHub repositories.
type GitHubCollector struct {
	// Org is the GitHub organisation.
	Org string

	// Repo is the repository name. If empty and Org is set, all repos are collected.
	Repo string

	// IssuesOnly limits collection to issues (excludes PRs).
	IssuesOnly bool

	// PRsOnly limits collection to PRs (excludes issues).
	PRsOnly bool
}

// Name returns the collector name.
func (g *GitHubCollector) Name() string {
	if g.Repo != "" {
		return fmt.Sprintf("github:%s/%s", g.Org, g.Repo)
	}
	return fmt.Sprintf("github:%s", g.Org)
}

// Collect gathers issues and/or PRs from GitHub repositories.
func (g *GitHubCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: g.Name()}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(g.Name(), "Starting GitHub collection")
	}

	// If no specific repo, list all repos in the org
	repos := []string{g.Repo}
	if g.Repo == "" {
		var err error
		repos, err = g.listOrgRepos(ctx)
		if err != nil {
			return result, err
		}
	}

	for _, repo := range repos {
		if ctx.Err() != nil {
			return result, core.E("collect.GitHub.Collect", "context cancelled", ctx.Err())
		}

		if !g.PRsOnly {
			issueResult, err := g.collectIssues(ctx, cfg, repo)
			if err != nil {
				result.Errors++
				if cfg.Dispatcher != nil {
					cfg.Dispatcher.EmitError(g.Name(), fmt.Sprintf("Error collecting issues for %s: %v", repo, err), nil)
				}
			} else {
				result.Items += issueResult.Items
				result.Skipped += issueResult.Skipped
				result.Files = append(result.Files, issueResult.Files...)
			}
		}

		if !g.IssuesOnly {
			prResult, err := g.collectPRs(ctx, cfg, repo)
			if err != nil {
				result.Errors++
				if cfg.Dispatcher != nil {
					cfg.Dispatcher.EmitError(g.Name(), fmt.Sprintf("Error collecting PRs for %s: %v", repo, err), nil)
				}
			} else {
				result.Items += prResult.Items
				result.Skipped += prResult.Skipped
				result.Files = append(result.Files, prResult.Files...)
			}
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(g.Name(), fmt.Sprintf("Collected %d items", result.Items), result)
	}

	return result, nil
}

// listOrgRepos returns all repository names for the configured org.
func (g *GitHubCollector) listOrgRepos(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "gh", "repo", "list", g.Org,
		"--json", "name",
		"--limit", "1000",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, core.E("collect.GitHub.listOrgRepos", "failed to list repos", err)
	}

	var repos []ghRepo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, core.E("collect.GitHub.listOrgRepos", "failed to parse repo list", err)
	}

	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name
	}
	return names, nil
}

// collectIssues collects issues for a single repository.
func (g *GitHubCollector) collectIssues(ctx context.Context, cfg *Config, repo string) (*Result, error) {
	result := &Result{Source: fmt.Sprintf("github:%s/%s/issues", g.Org, repo)}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(g.Name(), fmt.Sprintf("[dry-run] Would collect issues for %s/%s", g.Org, repo), nil)
		}
		return result, nil
	}

	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "github"); err != nil {
			return result, err
		}
	}

	repoRef := fmt.Sprintf("%s/%s", g.Org, repo)
	cmd := exec.CommandContext(ctx, "gh", "issue", "list",
		"--repo", repoRef,
		"--json", "number,title,state,author,body,createdAt,labels,url",
		"--limit", "100",
		"--state", "all",
	)
	out, err := cmd.Output()
	if err != nil {
		return result, core.E("collect.GitHub.collectIssues", "gh issue list failed for "+repoRef, err)
	}

	var issues []ghIssue
	if err := json.Unmarshal(out, &issues); err != nil {
		return result, core.E("collect.GitHub.collectIssues", "failed to parse issues", err)
	}

	baseDir := filepath.Join(cfg.OutputDir, "github", g.Org, repo, "issues")
	if err := cfg.Output.EnsureDir(baseDir); err != nil {
		return result, core.E("collect.GitHub.collectIssues", "failed to create output directory", err)
	}

	for _, issue := range issues {
		filePath := filepath.Join(baseDir, fmt.Sprintf("%d.md", issue.Number))
		content := formatIssueMarkdown(issue)

		if err := cfg.Output.Write(filePath, content); err != nil {
			result.Errors++
			continue
		}

		result.Items++
		result.Files = append(result.Files, filePath)

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitItem(g.Name(), fmt.Sprintf("Issue #%d: %s", issue.Number, issue.Title), nil)
		}
	}

	return result, nil
}

// collectPRs collects pull requests for a single repository.
func (g *GitHubCollector) collectPRs(ctx context.Context, cfg *Config, repo string) (*Result, error) {
	result := &Result{Source: fmt.Sprintf("github:%s/%s/pulls", g.Org, repo)}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(g.Name(), fmt.Sprintf("[dry-run] Would collect PRs for %s/%s", g.Org, repo), nil)
		}
		return result, nil
	}

	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "github"); err != nil {
			return result, err
		}
	}

	repoRef := fmt.Sprintf("%s/%s", g.Org, repo)
	cmd := exec.CommandContext(ctx, "gh", "pr", "list",
		"--repo", repoRef,
		"--json", "number,title,state,author,body,createdAt,labels,url",
		"--limit", "100",
		"--state", "all",
	)
	out, err := cmd.Output()
	if err != nil {
		return result, core.E("collect.GitHub.collectPRs", "gh pr list failed for "+repoRef, err)
	}

	var prs []ghIssue
	if err := json.Unmarshal(out, &prs); err != nil {
		return result, core.E("collect.GitHub.collectPRs", "failed to parse pull requests", err)
	}

	baseDir := filepath.Join(cfg.OutputDir, "github", g.Org, repo, "pulls")
	if err := cfg.Output.EnsureDir(baseDir); err != nil {
		return result, core.E("collect.GitHub.collectPRs", "failed to create output directory", err)
	}

	for _, pr := range prs {
		filePath := filepath.Join(baseDir, fmt.Sprintf("%d.md", pr.Number))
		content := formatIssueMarkdown(pr)

		if err := cfg.Output.Write(filePath, content); err != nil {
			result.Errors++
			continue
		}

		result.Items++
		result.Files = append(result.Files, filePath)

		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitItem(g.Name(), fmt.Sprintf("PR #%d: %s", pr.Number, pr.Title), nil)
		}
	}

	return result, nil
}

// formatIssueMarkdown formats a GitHub issue or PR as markdown.
func formatIssueMarkdown(issue ghIssue) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", issue.Title)
	fmt.Fprintf(&b, "- **Number:** #%d\n", issue.Number)
	fmt.Fprintf(&b, "- **State:** %s\n", issue.State)
	fmt.Fprintf(&b, "- **Author:** %s\n", issue.Author.Login)
	fmt.Fprintf(&b, "- **Created:** %s\n", issue.CreatedAt.Format(time.RFC3339))

	if len(issue.Labels) > 0 {
		labels := make([]string, len(issue.Labels))
		for i, l := range issue.Labels {
			labels[i] = l.Name
		}
		fmt.Fprintf(&b, "- **Labels:** %s\n", strings.Join(labels, ", "))
	}

	if issue.URL != "" {
		fmt.Fprintf(&b, "- **URL:** %s\n", issue.URL)
	}

	if issue.Body != "" {
		fmt.Fprintf(&b, "\n%s\n", issue.Body)
	}

	return b.String()
}
