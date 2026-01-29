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

import "github.com/leaanthony/clir"

// AddCommands registers the 'ai' command and all subcommands.
func AddCommands(app *clir.Cli) {
	aiCmd := app.NewSubCommand("ai", "AI agent task management")
	aiCmd.LongDescription("Manage tasks from the core-agentic service for AI-assisted development.\n\n" +
		"Commands:\n" +
		"  tasks          List tasks (filterable by status, priority, labels)\n" +
		"  task           View task details or auto-select highest priority\n" +
		"  task:update    Update task status or progress\n" +
		"  task:complete  Mark task as completed or failed\n" +
		"  task:commit    Create git commit with task reference\n" +
		"  task:pr        Create GitHub PR linked to task\n" +
		"  claude         Claude Code integration\n\n" +
		"Workflow:\n" +
		"  core ai tasks                      # List pending tasks\n" +
		"  core ai task --auto --claim        # Auto-select and claim a task\n" +
		"  core ai task:commit <id> -m 'msg'  # Commit with task reference\n" +
		"  core ai task:complete <id>         # Mark task done")

	// Add Claude command
	addClaudeCommand(aiCmd)

	// Add agentic task commands
	AddAgenticCommands(aiCmd)
}

// addClaudeCommand adds the 'claude' subcommand for Claude Code integration.
func addClaudeCommand(parent *clir.Command) {
	claudeCmd := parent.NewSubCommand("claude", "Claude Code integration")
	claudeCmd.LongDescription("Tools for working with Claude Code.\n\n" +
		"Commands:\n" +
		"  run       Run Claude in the current directory\n" +
		"  config    Manage Claude configuration")

	// core ai claude run
	runCmd := claudeCmd.NewSubCommand("run", "Run Claude Code in the current directory")
	runCmd.Action(func() error {
		return runClaudeCode()
	})

	// core ai claude config
	configCmd := claudeCmd.NewSubCommand("config", "Manage Claude configuration")
	configCmd.Action(func() error {
		return showClaudeConfig()
	})
}

func runClaudeCode() error {
	// Placeholder - will integrate with claude CLI
	return nil
}

func showClaudeConfig() error {
	// Placeholder - will show claude configuration
	return nil
}
