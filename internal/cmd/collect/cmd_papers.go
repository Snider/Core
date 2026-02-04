package collect

import (
	"context"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// Papers command flags
var (
	papersSource   string
	papersCategory string
	papersQuery    string
)

// addPapersCommand adds the 'papers' subcommand to the collect parent.
func addPapersCommand(parent *cli.Command) {
	papersCmd := &cli.Command{
		Use:   "papers",
		Short: i18n.T("cmd.collect.papers.short"),
		Long:  i18n.T("cmd.collect.papers.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runPapers()
		},
	}

	cli.StringFlag(papersCmd, &papersSource, "source", "s", "all", i18n.T("cmd.collect.papers.flag.source"))
	cli.StringFlag(papersCmd, &papersCategory, "category", "c", "", i18n.T("cmd.collect.papers.flag.category"))
	cli.StringFlag(papersCmd, &papersQuery, "query", "q", "", i18n.T("cmd.collect.papers.flag.query"))

	parent.AddCommand(papersCmd)
}

func runPapers() error {
	if papersQuery == "" {
		return cli.Err("--query (-q) is required")
	}

	cfg := newConfig()
	setupVerboseLogging(cfg)

	collector := &collect.PapersCollector{
		Source:   papersSource,
		Category: papersCategory,
		Query:    papersQuery,
	}

	if cfg.DryRun {
		cli.Info("Dry run: would collect papers from " + papersSource)
		return nil
	}

	ctx := context.Background()
	result, err := collector.Collect(ctx, cfg)
	if err != nil {
		return cli.Wrap(err, "papers collection failed")
	}

	printResult(result)
	return nil
}
