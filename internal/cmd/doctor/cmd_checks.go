package doctor

import (
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/i18n"
)

// check represents a tool check configuration
type check struct {
	name        string
	description string
	command     string
	args        []string
	versionFlag string
}

// requiredChecks returns tools that must be installed
func requiredChecks() []check {
	return []check{
		{
			name:        i18n.T("cmd.doctor.check.git.name"),
			description: i18n.T("cmd.doctor.check.git.description"),
			command:     "git",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
		{
			name:        i18n.T("cmd.doctor.check.gh.name"),
			description: i18n.T("cmd.doctor.check.gh.description"),
			command:     "gh",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
		{
			name:        i18n.T("cmd.doctor.check.php.name"),
			description: i18n.T("cmd.doctor.check.php.description"),
			command:     "php",
			args:        []string{"-v"},
			versionFlag: "-v",
		},
		{
			name:        i18n.T("cmd.doctor.check.composer.name"),
			description: i18n.T("cmd.doctor.check.composer.description"),
			command:     "composer",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
		{
			name:        i18n.T("cmd.doctor.check.node.name"),
			description: i18n.T("cmd.doctor.check.node.description"),
			command:     "node",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
	}
}

// optionalChecks returns tools that are nice to have
func optionalChecks() []check {
	return []check{
		{
			name:        i18n.T("cmd.doctor.check.pnpm.name"),
			description: i18n.T("cmd.doctor.check.pnpm.description"),
			command:     "pnpm",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
		{
			name:        i18n.T("cmd.doctor.check.claude.name"),
			description: i18n.T("cmd.doctor.check.claude.description"),
			command:     "claude",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
		{
			name:        i18n.T("cmd.doctor.check.docker.name"),
			description: i18n.T("cmd.doctor.check.docker.description"),
			command:     "docker",
			args:        []string{"--version"},
			versionFlag: "--version",
		},
	}
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
