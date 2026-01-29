// Package ai provides AI agent tools and task management commands.
package ai

import "github.com/leaanthony/clir"

// AddCommands registers the 'ai' command and all subcommands.
func AddCommands(app *clir.Cli) {
	aiCmd := app.NewSubCommand("ai", "AI agent tools")
	aiCmd.LongDescription("AI and agent-related tools for development.\n\n" +
		"Commands:\n" +
		"  claude    Claude Code integration\n" +
		"  tasks     List available tasks\n" +
		"  task      Show/claim a specific task\n\n" +
		"Task workflow:\n" +
		"  core ai tasks              # List pending tasks\n" +
		"  core ai task <id>          # View and claim a task\n" +
		"  core ai task:complete <id> # Mark task complete")

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
