package doctor

import (
	"os/exec"
	"strings"
)

// check represents a tool check configuration
type check struct {
	name        string
	description string
	command     string
	args        []string
	versionFlag string
}

// requiredChecks are tools that must be installed
var requiredChecks = []check{
	{
		name:        "Git",
		description: "Version control",
		command:     "git",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
	{
		name:        "GitHub CLI",
		description: "GitHub integration (issues, PRs, CI)",
		command:     "gh",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
	{
		name:        "PHP",
		description: "Laravel packages",
		command:     "php",
		args:        []string{"-v"},
		versionFlag: "-v",
	},
	{
		name:        "Composer",
		description: "PHP dependencies",
		command:     "composer",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
	{
		name:        "Node.js",
		description: "Frontend builds",
		command:     "node",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
}

// optionalChecks are tools that are nice to have
var optionalChecks = []check{
	{
		name:        "pnpm",
		description: "Fast package manager",
		command:     "pnpm",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
	{
		name:        "Claude Code",
		description: "AI-assisted development",
		command:     "claude",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
	{
		name:        "Docker",
		description: "Container runtime",
		command:     "docker",
		args:        []string{"--version"},
		versionFlag: "--version",
	},
}

// runCheck executes a tool check and returns success status and version info
func runCheck(c check) (bool, string) {
	cmd := exec.Command(c.command, c.args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	// Extract first line as version
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		return true, strings.TrimSpace(lines[0])
	}
	return true, ""
}
