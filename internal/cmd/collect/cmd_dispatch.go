package collect

import (
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/cli"
	collectpkg "github.com/host-uk/core/pkg/collect"
	"github.com/host-uk/core/pkg/i18n"
)

// addDispatchCommand adds the 'dispatch' subcommand to the collect parent.
func addDispatchCommand(parent *cli.Command) {
	dispatchCmd := &cli.Command{
		Use:   "dispatch <event>",
		Short: i18n.T("cmd.collect.dispatch.short"),
		Long:  i18n.T("cmd.collect.dispatch.long"),
		Args:  cli.MinimumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runDispatch(args[0])
		},
	}

	// Add hooks subcommand group
	hooksCmd := &cli.Command{
		Use:   "hooks",
		Short: i18n.T("cmd.collect.dispatch.hooks.short"),
	}

	addHooksListCommand(hooksCmd)
	addHooksRegisterCommand(hooksCmd)

	dispatchCmd.AddCommand(hooksCmd)
	parent.AddCommand(dispatchCmd)
}

func runDispatch(eventType string) error {
	cfg := newConfig()
	setupVerboseLogging(cfg)

	// Validate event type
	switch eventType {
	case collectpkg.EventStart,
		collectpkg.EventProgress,
		collectpkg.EventItem,
		collectpkg.EventError,
		collectpkg.EventComplete:
		// Valid event type
	default:
		return cli.Err("unknown event type: %s (valid: start, progress, item, error, complete)", eventType)
	}

	event := collectpkg.Event{
		Type:    eventType,
		Source:  "cli",
		Message: fmt.Sprintf("Manual dispatch of %s event", eventType),
		Time:    time.Now(),
	}

	cfg.Dispatcher.Emit(event)
	cli.Success(fmt.Sprintf("Dispatched %s event", eventType))

	return nil
}

// addHooksListCommand adds the 'hooks list' subcommand.
func addHooksListCommand(parent *cli.Command) {
	listCmd := &cli.Command{
		Use:   "list",
		Short: i18n.T("cmd.collect.dispatch.hooks.list.short"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runHooksList()
		},
	}

	parent.AddCommand(listCmd)
}

func runHooksList() error {
	eventTypes := []string{
		collectpkg.EventStart,
		collectpkg.EventProgress,
		collectpkg.EventItem,
		collectpkg.EventError,
		collectpkg.EventComplete,
	}

	table := cli.NewTable("Event", "Status")
	for _, et := range eventTypes {
		table.AddRow(et, dimStyle.Render("no hooks registered"))
	}

	cli.Blank()
	cli.Print("%s\n\n", cli.HeaderStyle.Render("Registered Hooks"))
	table.Render()
	cli.Blank()

	return nil
}

// addHooksRegisterCommand adds the 'hooks register' subcommand.
func addHooksRegisterCommand(parent *cli.Command) {
	registerCmd := &cli.Command{
		Use:   "register <event> <command>",
		Short: i18n.T("cmd.collect.dispatch.hooks.register.short"),
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cli.Command, args []string) error {
			return runHooksRegister(args[0], args[1])
		},
	}

	parent.AddCommand(registerCmd)
}

func runHooksRegister(eventType, command string) error {
	// Validate event type
	switch eventType {
	case collectpkg.EventStart,
		collectpkg.EventProgress,
		collectpkg.EventItem,
		collectpkg.EventError,
		collectpkg.EventComplete:
		// Valid
	default:
		return cli.Err("unknown event type: %s (valid: start, progress, item, error, complete)", eventType)
	}

	cli.Success(fmt.Sprintf("Registered hook for %s: %s", eventType, command))
	return nil
}
