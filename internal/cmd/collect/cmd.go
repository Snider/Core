package collect

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/io"
)

func init() {
	cli.RegisterCommands(AddCollectCommands)
}

// Style aliases from shared package
var (
	dimStyle     = cli.DimStyle
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
)

// Shared flags across all collect subcommands
var (
	collectOutputDir string
	collectVerbose   bool
	collectDryRun    bool
)

// AddCollectCommands registers the 'collect' command and all subcommands.
func AddCollectCommands(root *cli.Command) {
	collectCmd := &cli.Command{
		Use:   "collect",
		Short: i18n.T("cmd.collect.short"),
		Long:  i18n.T("cmd.collect.long"),
	}

	// Persistent flags shared across subcommands
	cli.PersistentStringFlag(collectCmd, &collectOutputDir, "output", "o", "./collect", i18n.T("cmd.collect.flag.output"))
	cli.PersistentBoolFlag(collectCmd, &collectVerbose, "verbose", "v", false, i18n.T("common.flag.verbose"))
	cli.PersistentBoolFlag(collectCmd, &collectDryRun, "dry-run", "", false, i18n.T("cmd.collect.flag.dry_run"))

	root.AddCommand(collectCmd)

	addGitHubCommand(collectCmd)
	addBitcoinTalkCommand(collectCmd)
	addMarketCommand(collectCmd)
	addPapersCommand(collectCmd)
	addExcavateCommand(collectCmd)
	addProcessCommand(collectCmd)
	addDispatchCommand(collectCmd)
}

// newConfig creates a collection Config using the shared persistent flags.
// It uses io.Local for real filesystem access rather than the mock medium.
func newConfig() *collect.Config {
	cfg := collect.NewConfigWithMedium(io.Local, collectOutputDir)
	cfg.Verbose = collectVerbose
	cfg.DryRun = collectDryRun
	return cfg
}

// setupVerboseLogging registers event handlers on the dispatcher for verbose output.
func setupVerboseLogging(cfg *collect.Config) {
	if !cfg.Verbose {
		return
	}

	cfg.Dispatcher.On(collect.EventStart, func(e collect.Event) {
		cli.Print("%s %s\n", dimStyle.Render("[start]"), e.Message)
	})
	cfg.Dispatcher.On(collect.EventProgress, func(e collect.Event) {
		cli.Print("%s %s\n", dimStyle.Render("[progress]"), e.Message)
	})
	cfg.Dispatcher.On(collect.EventItem, func(e collect.Event) {
		cli.Print("%s %s\n", dimStyle.Render("[item]"), e.Message)
	})
	cfg.Dispatcher.On(collect.EventError, func(e collect.Event) {
		cli.Print("%s %s\n", errorStyle.Render("[error]"), e.Message)
	})
	cfg.Dispatcher.On(collect.EventComplete, func(e collect.Event) {
		cli.Print("%s %s\n", successStyle.Render("[complete]"), e.Message)
	})
}

// printResult prints a formatted summary of a collection result.
func printResult(result *collect.Result) {
	if result == nil {
		return
	}

	if result.Items > 0 {
		cli.Success(fmt.Sprintf("Collected %d items from %s", result.Items, result.Source))
	} else {
		cli.Dim(fmt.Sprintf("No items collected from %s", result.Source))
	}

	if result.Skipped > 0 {
		cli.Dim(fmt.Sprintf("  Skipped: %d", result.Skipped))
	}

	if result.Errors > 0 {
		cli.Warn(fmt.Sprintf("  Errors: %d", result.Errors))
	}

	if collectVerbose && len(result.Files) > 0 {
		cli.Dim(fmt.Sprintf("  Files: %d", len(result.Files)))
		for _, f := range result.Files {
			cli.Print("    %s\n", dimStyle.Render(f))
		}
	}
}
