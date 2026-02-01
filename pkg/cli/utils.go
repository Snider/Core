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

// ConfirmOption configures Confirm behaviour.
type ConfirmOption func(*confirmConfig)

type confirmConfig struct {
	defaultYes bool
	required   bool
	timeout    time.Duration
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

// Timeout sets a timeout after which the default response is auto-selected.
// If no default is set (not Required and not DefaultYes), defaults to "no".
//
//	Confirm("Continue?", Timeout(30*time.Second))  // Auto-no after 30s
//	Confirm("Continue?", DefaultYes(), Timeout(10*time.Second))  // Auto-yes after 10s
func Timeout(d time.Duration) ConfirmOption {
	return func(c *confirmConfig) {
		c.timeout = d
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
//	if Confirm("Auto-continue?", Timeout(30*time.Second)) { ... }
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

	// Add timeout indicator if set
	if cfg.timeout > 0 {
		suffix = fmt.Sprintf("%s(auto in %s) ", suffix, cfg.timeout.Round(time.Second))
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s %s", prompt, suffix)

		var response string

		if cfg.timeout > 0 {
			// Use timeout-based reading
			resultChan := make(chan string, 1)
			go func() {
				line, _ := reader.ReadString('\n')
				resultChan <- line
			}()

			select {
			case response = <-resultChan:
				response = strings.ToLower(strings.TrimSpace(response))
			case <-time.After(cfg.timeout):
				fmt.Println() // New line after timeout
				return cfg.defaultYes
			}
		} else {
			response, _ = reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))
		}

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

// ConfirmAction prompts for confirmation of an action using grammar composition.
//
//	if ConfirmAction("delete", "config.yaml") { ... }
//	if ConfirmAction("save", "changes", DefaultYes()) { ... }
func ConfirmAction(verb, subject string, opts ...ConfirmOption) bool {
	question := i18n.Title(verb) + " " + subject + "?"
	return Confirm(question, opts...)
}

// ConfirmDangerousAction prompts for double confirmation of a dangerous action.
// Shows initial question, then a "Really verb subject?" confirmation.
//
//	if ConfirmDangerousAction("delete", "config.yaml") { ... }
func ConfirmDangerousAction(verb, subject string) bool {
	question := i18n.Title(verb) + " " + subject + "?"
	if !Confirm(question, Required()) {
		return false
	}

	confirm := "Really " + verb + " " + subject + "?"
	return Confirm(confirm, Required())
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

// QuestionAction prompts for text input using grammar composition.
//
//	name := QuestionAction("rename", "old.txt")
func QuestionAction(verb, subject string, opts ...QuestionOption) string {
	question := i18n.Title(verb) + " " + subject + "?"
	return Question(question, opts...)
}

// ChooseOption configures Choose behaviour.
type ChooseOption[T any] func(*chooseConfig[T])

type chooseConfig[T any] struct {
	displayFn func(T) string
	defaultN  int  // 0-based index of default selection
	filter    bool // Enable fuzzy filtering
	multi     bool // Allow multiple selection
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

// Filter enables type-to-filter functionality.
// Users can type to narrow down the list of options.
// Note: This is a hint for interactive UIs; the basic CLI Choose
// implementation uses numbered selection which doesn't support filtering.
func Filter[T any]() ChooseOption[T] {
	return func(c *chooseConfig[T]) {
		c.filter = true
	}
}

// Multi allows multiple selections.
// Use ChooseMulti instead of Choose when this option is needed.
func Multi[T any]() ChooseOption[T] {
	return func(c *chooseConfig[T]) {
		c.multi = true
	}
}

// Display sets a custom display function for items.
// Alias for WithDisplay for shorter syntax.
//
//	Choose("Select:", items, Display(func(f File) string { return f.Name }))
func Display[T any](fn func(T) string) ChooseOption[T] {
	return WithDisplay[T](fn)
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

// ChooseAction prompts for selection using grammar composition.
//
//	file := ChooseAction("select", "file", files)
func ChooseAction[T any](verb, subject string, items []T, opts ...ChooseOption[T]) T {
	question := i18n.Title(verb) + " " + subject + ":"
	return Choose(question, items, opts...)
}

// ChooseMulti prompts the user to select multiple items from a list.
// Returns the selected items. Uses space-separated numbers or ranges.
//
//	choices := ChooseMulti("Select files:", files)
//	choices := ChooseMulti("Select files:", files, WithDisplay(func(f File) string { return f.Name }))
//
// Input format:
//   - "1 3 5" - select items 1, 3, and 5
//   - "1-3" - select items 1, 2, and 3
//   - "1 3-5" - select items 1, 3, 4, and 5
//   - "" (empty) - select none
func ChooseMulti[T any](prompt string, items []T, opts ...ChooseOption[T]) []T {
	if len(items) == 0 {
		return nil
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
		fmt.Printf("  %d. %s\n", i+1, cfg.displayFn(item))
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Enter numbers (e.g., 1 3 5 or 1-3) or empty for none: ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		// Empty response returns no selections
		if response == "" {
			return nil
		}

		// Parse the selection
		selected, err := parseMultiSelection(response, len(items))
		if err != nil {
			fmt.Printf("Invalid selection: %v\n", err)
			continue
		}

		// Build result
		result := make([]T, 0, len(selected))
		for _, idx := range selected {
			result = append(result, items[idx])
		}
		return result
	}
}

// parseMultiSelection parses a multi-selection string like "1 3 5" or "1-3 5".
// Returns 0-based indices.
func parseMultiSelection(input string, maxItems int) ([]int, error) {
	selected := make(map[int]bool)
	parts := strings.Fields(input)

	for _, part := range parts {
		// Check for range (e.g., "1-3")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			var start, end int
			if _, err := fmt.Sscanf(rangeParts[0], "%d", &start); err != nil {
				return nil, fmt.Errorf("invalid range start: %s", rangeParts[0])
			}
			if _, err := fmt.Sscanf(rangeParts[1], "%d", &end); err != nil {
				return nil, fmt.Errorf("invalid range end: %s", rangeParts[1])
			}
			if start < 1 || start > maxItems || end < 1 || end > maxItems || start > end {
				return nil, fmt.Errorf("range out of bounds: %s", part)
			}
			for i := start; i <= end; i++ {
				selected[i-1] = true // Convert to 0-based
			}
		} else {
			// Single number
			var n int
			if _, err := fmt.Sscanf(part, "%d", &n); err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			if n < 1 || n > maxItems {
				return nil, fmt.Errorf("number out of range: %d", n)
			}
			selected[n-1] = true // Convert to 0-based
		}
	}

	// Convert map to sorted slice
	result := make([]int, 0, len(selected))
	for i := 0; i < maxItems; i++ {
		if selected[i] {
			result = append(result, i)
		}
	}
	return result, nil
}

// ChooseMultiAction prompts for multiple selections using grammar composition.
//
//	files := ChooseMultiAction("select", "files", files)
func ChooseMultiAction[T any](verb, subject string, items []T, opts ...ChooseOption[T]) []T {
	question := i18n.Title(verb) + " " + subject + ":"
	return ChooseMulti(question, items, opts...)
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
