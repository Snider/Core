package shared

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GhAuthenticated checks if the GitHub CLI is authenticated.
// Returns true if 'gh auth status' indicates a logged-in user.
func GhAuthenticated() bool {
	cmd := exec.Command("gh", "auth", "status")
	output, _ := cmd.CombinedOutput()
	return strings.Contains(string(output), "Logged in")
}

// Truncate shortens a string to max characters, adding "..." if truncated.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Confirm prompts the user for yes/no confirmation.
// Returns true if the user enters "y" or "yes" (case-insensitive).
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// FormatAge formats a time as a human-readable age string.
// Examples: "5m ago", "2h ago", "3d ago", "1w ago", "2mo ago"
func FormatAge(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	}
	return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
}

// GitClone clones a GitHub repository to the specified path.
// Prefers 'gh repo clone' if authenticated, falls back to SSH.
func GitClone(ctx context.Context, org, repo, path string) error {
	if GhAuthenticated() {
		httpsURL := fmt.Sprintf("https://github.com/%s/%s.git", org, repo)
		cmd := exec.CommandContext(ctx, "gh", "repo", "clone", httpsURL, path)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		errStr := strings.TrimSpace(string(output))
		if strings.Contains(errStr, "already exists") {
			return fmt.Errorf("%s", errStr)
		}
	}
	// Fall back to SSH clone
	cmd := exec.CommandContext(ctx, "git", "clone", fmt.Sprintf("git@github.com:%s/%s.git", org, repo), path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}
