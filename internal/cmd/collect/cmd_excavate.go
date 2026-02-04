package collect

import (
	"context"
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// Excavate command flags
var (
	excavateScanOnly bool
	excavateResume   bool
)

// addExcavateCommand adds the 'excavate' subcommand to the collect parent.
func addExcavateCommand(parent *cli.Command) {
	excavateCmd := &cli.Command{
		Use:   "excavate <project>",
		Short: i18n.T("cmd.collect.excavate.short"),
		Long:  i18n.T("cmd.collect.excavate.long"),
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runExcavate(args[0])
		},
	}

	cli.BoolFlag(excavateCmd, &excavateScanOnly, "scan-only", "", false, i18n.T("cmd.collect.excavate.flag.scan_only"))
	cli.BoolFlag(excavateCmd, &excavateResume, "resume", "r", false, i18n.T("cmd.collect.excavate.flag.resume"))

	parent.AddCommand(excavateCmd)
}

func runExcavate(project string) error {
	cfg := newConfig()
	setupVerboseLogging(cfg)

	// Load state for resume
	if excavateResume {
		if err := cfg.State.Load(); err != nil {
			return cli.Wrap(err, "failed to load collection state")
		}
	}

	// Build collectors for the project
	collectors := buildProjectCollectors(project)
	if len(collectors) == 0 {
		return cli.Err("no collectors configured for project: %s", project)
	}

	excavator := &collect.Excavator{
		Collectors: collectors,
		ScanOnly:   excavateScanOnly,
		Resume:     excavateResume,
	}

	if cfg.DryRun {
		cli.Info(fmt.Sprintf("Dry run: would excavate project %s with %d collectors", project, len(collectors)))
		for _, c := range collectors {
			cli.Dim(fmt.Sprintf("  - %s", c.Name()))
		}
		return nil
	}

	ctx := context.Background()
	result, err := excavator.Run(ctx, cfg)
	if err != nil {
		return cli.Wrap(err, "excavation failed")
	}

	// Save state for future resume
	if err := cfg.State.Save(); err != nil {
		cli.Warnf("Failed to save state: %v", err)
	}

	printResult(result)
	return nil
}

// buildProjectCollectors creates collectors based on the project name.
// This maps known project names to their collector configurations.
func buildProjectCollectors(project string) []collect.Collector {
	switch project {
	case "bitcoin":
		return []collect.Collector{
			&collect.GitHubCollector{Org: "bitcoin", Repo: "bitcoin"},
			&collect.MarketCollector{CoinID: "bitcoin", Historical: true},
		}
	case "ethereum":
		return []collect.Collector{
			&collect.GitHubCollector{Org: "ethereum", Repo: "go-ethereum"},
			&collect.MarketCollector{CoinID: "ethereum", Historical: true},
			&collect.PapersCollector{Source: "all", Query: "ethereum"},
		}
	default:
		// Treat unknown projects as GitHub org/repo
		return []collect.Collector{
			&collect.GitHubCollector{Org: project},
		}
	}
}
