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

import "github.com/spf13/cobra"

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI agent task management",
	Long: `Manage tasks from the core-agentic service for AI-assisted development.

Commands:
  tasks          List tasks (filterable by status, priority, labels)
  task           View task details or auto-select highest priority
  task:update    Update task status or progress
  task:complete  Mark task as completed or failed
  task:commit    Create git commit with task reference
  task:pr        Create GitHub PR linked to task
  claude         Claude Code integration

Workflow:
  core ai tasks                      # List pending tasks
  core ai task --auto --claim        # Auto-select and claim a task
  core ai task:commit <id> -m 'msg'  # Commit with task reference
  core ai task:complete <id>         # Mark task done`,
}

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Claude Code integration",
	Long: `Tools for working with Claude Code.

Commands:
  run       Run Claude in the current directory
  config    Manage Claude configuration`,
}

var claudeRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run Claude Code in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClaudeCode()
	},
}

var claudeConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Claude configuration",
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
