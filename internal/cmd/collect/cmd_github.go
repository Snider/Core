package collect

import (
	"context"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// GitHub command flags
var (
	githubOrg        bool
	githubIssuesOnly bool
	githubPRsOnly    bool
)

// addGitHubCommand adds the 'github' subcommand to the collect parent.
func addGitHubCommand(parent *cli.Command) {
	githubCmd := &cli.Command{
		Use:   "github <org/repo>",
		Short: i18n.T("cmd.collect.github.short"),
		Long:  i18n.T("cmd.collect.github.long"),
		Args:  cli.MinimumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runGitHub(args[0])
		},
	}

	cli.BoolFlag(githubCmd, &githubOrg, "org", "", false, i18n.T("cmd.collect.github.flag.org"))
	cli.BoolFlag(githubCmd, &githubIssuesOnly, "issues-only", "", false, i18n.T("cmd.collect.github.flag.issues_only"))
	cli.BoolFlag(githubCmd, &githubPRsOnly, "prs-only", "", false, i18n.T("cmd.collect.github.flag.prs_only"))

	parent.AddCommand(githubCmd)
}

func runGitHub(target string) error {
	if githubIssuesOnly && githubPRsOnly {
		return cli.Err("--issues-only and --prs-only are mutually exclusive")
	}

	// Parse org/repo argument
	var org, repo string
	if strings.Contains(target, "/") {
		parts := strings.SplitN(target, "/", 2)
		org = parts[0]
		repo = parts[1]
	} else if githubOrg {
		org = target
	} else {
		return cli.Err("argument must be in org/repo format, or use --org for organisation-wide collection")
	}

	cfg := newConfig()
	setupVerboseLogging(cfg)

	collector := &collect.GitHubCollector{
		Org:        org,
		Repo:       repo,
		IssuesOnly: githubIssuesOnly,
		PRsOnly:    githubPRsOnly,
	}

	if cfg.DryRun {
		cli.Info("Dry run: would collect from GitHub " + target)
		return nil
	}

	ctx := context.Background()
	result, err := collector.Collect(ctx, cfg)
	if err != nil {
		return cli.Wrap(err, "github collection failed")
	}

	printResult(result)
	return nil
}
