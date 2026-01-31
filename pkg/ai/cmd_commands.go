// Package ai provides AI agent task management and Claude Code integration.
//
// Commands:
//   - tasks: List tasks from the agentic service
//   - task: View, claim, or auto-select tasks
//   - task:update: Update task status and progress
//   - task:complete: Mark tasks as completed or failed
//   - task:commit: Create commits with task references
//   - task:pr: Create pull requests linked to tasks
//   - claude: Claude Code CLI integration (planned)
package ai

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

func init() {
	cli.RegisterCommands(AddAICommands)
}

var aiCmd = &cli.Command{
	Use:   "ai",
	Short: i18n.T("cmd.ai.short"),
	Long:  i18n.T("cmd.ai.long"),
}

var claudeCmd = &cli.Command{
	Use:   "claude",
	Short: i18n.T("cmd.ai.claude.short"),
	Long:  i18n.T("cmd.ai.claude.long"),
}

var claudeRunCmd = &cli.Command{
	Use:   "run",
	Short: i18n.T("cmd.ai.claude.run.short"),
	RunE: func(cmd *cli.Command, args []string) error {
		return runClaudeCode()
	},
}

var claudeConfigCmd = &cli.Command{
	Use:   "config",
	Short: i18n.T("cmd.ai.claude.config.short"),
	RunE: func(cmd *cli.Command, args []string) error {
		return showClaudeConfig()
	},
}

func initCommands() {
	// Add Claude subcommands
	claudeCmd.AddCommand(claudeRunCmd)
	claudeCmd.AddCommand(claudeConfigCmd)

	// Add Claude command to ai
	aiCmd.AddCommand(claudeCmd)

	// Add agentic task commands
	AddAgenticCommands(aiCmd)
}

// AddAICommands registers the 'ai' command and all subcommands.
func AddAICommands(root *cli.Command) {
	initCommands()
	root.AddCommand(aiCmd)
}

func runClaudeCode() error {
	// Placeholder - will integrate with claude CLI
	return nil
}

func showClaudeConfig() error {
	// Placeholder - will show claude configuration
	return nil
}
