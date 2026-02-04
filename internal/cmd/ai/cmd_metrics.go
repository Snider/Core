// cmd_metrics.go implements the metrics viewing command.

package ai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/ai"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

var (
	metricsSince string
	metricsJSON  bool
)

var metricsCmd = &cli.Command{
	Use:   "metrics",
	Short: i18n.T("cmd.ai.metrics.short"),
	Long:  i18n.T("cmd.ai.metrics.long"),
	RunE: func(cmd *cli.Command, args []string) error {
		return runMetrics()
	},
}

func initMetricsFlags() {
	metricsCmd.Flags().StringVar(&metricsSince, "since", "7d", i18n.T("cmd.ai.metrics.flag.since"))
	metricsCmd.Flags().BoolVar(&metricsJSON, "json", false, i18n.T("common.flag.json"))
}

func addMetricsCommand(parent *cli.Command) {
	initMetricsFlags()
	parent.AddCommand(metricsCmd)
}

func runMetrics() error {
	since, err := parseDuration(metricsSince)
	if err != nil {
		return cli.Err("invalid --since value %q: %v", metricsSince, err)
	}

	sinceTime := time.Now().Add(-since)
	events, err := ai.ReadEvents(sinceTime)
	if err != nil {
		return cli.WrapVerb(err, "read", "metrics")
	}

	if metricsJSON {
		summary := ai.Summary(events)
		output, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return cli.Wrap(err, "marshal JSON output")
		}
		cli.Text(string(output))
		return nil
	}

	summary := ai.Summary(events)

	cli.Blank()
	cli.Print("%s %s\n", dimStyle.Render("Period:"), metricsSince)
	total, _ := summary["total"].(int)
	cli.Print("%s %d\n", dimStyle.Render("Total events:"), total)
	cli.Blank()

	// By type
	if byType, ok := summary["by_type"].([]map[string]any); ok && len(byType) > 0 {
		cli.Print("%s\n", dimStyle.Render("By type:"))
		for _, entry := range byType {
			cli.Print("  %-30s %v\n", entry["key"], entry["count"])
		}
		cli.Blank()
	}

	// By repo
	if byRepo, ok := summary["by_repo"].([]map[string]any); ok && len(byRepo) > 0 {
		cli.Print("%s\n", dimStyle.Render("By repo:"))
		for _, entry := range byRepo {
			cli.Print("  %-30s %v\n", entry["key"], entry["count"])
		}
		cli.Blank()
	}

	// By agent
	if byAgent, ok := summary["by_agent"].([]map[string]any); ok && len(byAgent) > 0 {
		cli.Print("%s\n", dimStyle.Render("By contributor:"))
		for _, entry := range byAgent {
			cli.Print("  %-30s %v\n", entry["key"], entry["count"])
		}
		cli.Blank()
	}

	if len(events) == 0 {
		cli.Text(i18n.T("cmd.ai.metrics.none_found"))
	}

	return nil
}

// parseDuration parses a human-friendly duration like "7d", "24h", "30d".
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	unit := s[len(s)-1]
	value := s[:len(s)-1]

	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	if n <= 0 {
		return 0, fmt.Errorf("duration must be positive: %s", s)
	}

	switch unit {
	case 'd':
		return time.Duration(n) * 24 * time.Hour, nil
	case 'h':
		return time.Duration(n) * time.Hour, nil
	case 'm':
		return time.Duration(n) * time.Minute, nil
	default:
		return 0, fmt.Errorf("unknown unit %c in duration: %s", unit, s)
	}
}
