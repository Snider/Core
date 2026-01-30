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
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: i18n.T("cmd.ai.short"),
	Long:  i18n.T("cmd.ai.long"),
}

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: i18n.T("cmd.ai.claude.short"),
	Long:  i18n.T("cmd.ai.claude.long"),
}

var claudeRunCmd = &cobra.Command{
	Use:   "run",
	Short: i18n.T("cmd.ai.claude.run.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClaudeCode()
	},
}

var claudeConfigCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("cmd.ai.claude.config.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showClaudeConfig()
	},
}

func init() {
	// Add Claude subcommands
	claudeCmd.AddCommand(claudeRunCmd)
	claudeCmd.AddCommand(claudeConfigCmd)

	// Add Claude command to ai
	aiCmd.AddCommand(claudeCmd)

	// Add agentic task commands
	AddAgenticCommands(aiCmd)
}

// AddCommands registers the 'ai' command and all subcommands.
func AddCommands(root *cobra.Command) {
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
