package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/i18n"
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

// ConfirmOption configures Confirm behaviour.
type ConfirmOption func(*confirmConfig)

type confirmConfig struct {
	defaultYes bool
	required   bool
}

// DefaultYes sets the default response to "yes" (pressing Enter confirms).
func DefaultYes() ConfirmOption {
	return func(c *confirmConfig) {
		c.defaultYes = true
	}
}

// Required prevents empty responses; user must explicitly type y/n.
func Required() ConfirmOption {
	return func(c *confirmConfig) {
		c.required = true
	}
}

// Confirm prompts the user for yes/no confirmation.
// Returns true if the user enters "y" or "yes" (case-insensitive).
//
// Basic usage:
//
//	if Confirm("Delete file?") { ... }
//
// With options:
//
//	if Confirm("Save changes?", DefaultYes()) { ... }
//	if Confirm("Dangerous!", Required()) { ... }
func Confirm(prompt string, opts ...ConfirmOption) bool {
	cfg := &confirmConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build the prompt suffix
	var suffix string
	if cfg.required {
		suffix = "[y/n] "
	} else if cfg.defaultYes {
		suffix = "[Y/n] "
	} else {
		suffix = "[y/N] "
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s %s", prompt, suffix)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))

		// Handle empty response
		if response == "" {
			if cfg.required {
				continue // Ask again
			}
			return cfg.defaultYes
		}

		// Check for yes/no responses
		if response == "y" || response == "yes" {
			return true
		}
		if response == "n" || response == "no" {
			return false
		}

		// Invalid response
		if cfg.required {
			fmt.Println("Please enter 'y' or 'n'")
			continue
		}

		// Non-required: treat invalid as default
		return cfg.defaultYes
	}
}

// ConfirmIntent prompts for confirmation using a semantic intent.
// The intent determines the question text, danger level, and default response.
//
//	if ConfirmIntent("core.delete", i18n.S("file", "config.yaml")) { ... }
func ConfirmIntent(intent string, subject *i18n.Subject, opts ...ConfirmOption) bool {
	result := i18n.C(intent, subject)

	// Apply intent metadata to options
	if result.Meta.Dangerous {
		opts = append([]ConfirmOption{Required()}, opts...)
	}
	if result.Meta.Default == "yes" {
		opts = append([]ConfirmOption{DefaultYes()}, opts...)
	}

	return Confirm(result.Question, opts...)
}

// ConfirmDangerous prompts for confirmation of a dangerous action.
// Shows both the question and a confirmation prompt, requiring explicit "yes".
//
//	if ConfirmDangerous("core.delete", i18n.S("file", "config.yaml")) { ... }
func ConfirmDangerous(intent string, subject *i18n.Subject) bool {
	result := i18n.C(intent, subject)

	// Show initial question
	if !Confirm(result.Question, Required()) {
		return false
	}

	// For dangerous actions, show confirmation prompt
	if result.Meta.Dangerous && result.Confirm != "" {
		return Confirm(result.Confirm, Required())
	}

	return true
}

// QuestionOption configures Question behaviour.
type QuestionOption func(*questionConfig)

type questionConfig struct {
	defaultValue string
	required     bool
	validator    func(string) error
}

// WithDefault sets the default value shown in brackets.
func WithDefault(value string) QuestionOption {
	return func(c *questionConfig) {
		c.defaultValue = value
	}
}

// WithValidator adds a validation function for the response.
func WithValidator(fn func(string) error) QuestionOption {
	return func(c *questionConfig) {
		c.validator = fn
	}
}

// RequiredInput prevents empty responses.
func RequiredInput() QuestionOption {
	return func(c *questionConfig) {
		c.required = true
	}
}

// Question prompts the user for text input.
//
//	name := Question("Enter your name:")
//	name := Question("Enter your name:", WithDefault("Anonymous"))
//	name := Question("Enter your name:", RequiredInput())
func Question(prompt string, opts ...QuestionOption) string {
	cfg := &questionConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Build prompt with default
		if cfg.defaultValue != "" {
			fmt.Printf("%s [%s] ", prompt, cfg.defaultValue)
		} else {
			fmt.Printf("%s ", prompt)
		}

		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		// Handle empty response
		if response == "" {
			if cfg.required {
				fmt.Println("Response required")
				continue
			}
			response = cfg.defaultValue
		}

		// Validate if validator provided
		if cfg.validator != nil {
			if err := cfg.validator(response); err != nil {
				fmt.Printf("Invalid: %v\n", err)
				continue
			}
		}

		return response
	}
}

// QuestionIntent prompts for text input using a semantic intent.
//
//	name := QuestionIntent("core.rename", i18n.S("file", "old.txt"))
func QuestionIntent(intent string, subject *i18n.Subject, opts ...QuestionOption) string {
	result := i18n.C(intent, subject)
	return Question(result.Question, opts...)
}

// ChooseOption configures Choose behaviour.
type ChooseOption[T any] func(*chooseConfig[T])

type chooseConfig[T any] struct {
	displayFn func(T) string
	defaultN  int // 0-based index of default selection
}

// WithDisplay sets a custom display function for items.
func WithDisplay[T any](fn func(T) string) ChooseOption[T] {
	return func(c *chooseConfig[T]) {
		c.displayFn = fn
	}
}

// WithDefaultIndex sets the default selection index (0-based).
func WithDefaultIndex[T any](idx int) ChooseOption[T] {
	return func(c *chooseConfig[T]) {
		c.defaultN = idx
	}
}

// Choose prompts the user to select from a list of items.
// Returns the selected item. Uses simple numbered selection for terminal compatibility.
//
//	choice := Choose("Select a file:", files)
//	choice := Choose("Select a file:", files, WithDisplay(func(f File) string { return f.Name }))
func Choose[T any](prompt string, items []T, opts ...ChooseOption[T]) T {
	var zero T
	if len(items) == 0 {
		return zero
	}

	cfg := &chooseConfig[T]{
		displayFn: func(item T) string { return fmt.Sprint(item) },
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Display options
	fmt.Println(prompt)
	for i, item := range items {
		marker := " "
		if i == cfg.defaultN {
			marker = "*"
		}
		fmt.Printf("  %s%d. %s\n", marker, i+1, cfg.displayFn(item))
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Enter number [1-%d]: ", len(items))
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		// Empty response uses default
		if response == "" {
			return items[cfg.defaultN]
		}

		// Parse number
		var n int
		if _, err := fmt.Sscanf(response, "%d", &n); err == nil {
			if n >= 1 && n <= len(items) {
				return items[n-1]
			}
		}

		fmt.Printf("Please enter a number between 1 and %d\n", len(items))
	}
}

// ChooseIntent prompts for selection using a semantic intent.
//
//	file := ChooseIntent("core.select", i18n.S("file", ""), files)
func ChooseIntent[T any](intent string, subject *i18n.Subject, items []T, opts ...ChooseOption[T]) T {
	result := i18n.C(intent, subject)
	return Choose(result.Question, items, opts...)
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
