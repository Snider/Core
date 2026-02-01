// github_webhooks.go implements GitHub webhook synchronization.
//
// Uses the gh api command for webhook operations:
//   - gh api repos/{owner}/{repo}/hooks --method GET
//   - gh api repos/{owner}/{repo}/hooks --method POST

package setup

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/cli"
)

// GitHubWebhook represents a webhook as returned by the GitHub API.
type GitHubWebhook struct {
	ID     int                 `json:"id"`
	Name   string              `json:"name"`
	Active bool                `json:"active"`
	Events []string            `json:"events"`
	Config GitHubWebhookConfig `json:"config"`
}

// GitHubWebhookConfig contains webhook configuration details.
type GitHubWebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSSL string `json:"insecure_ssl"`
}

// ListWebhooks fetches all webhooks for a repository.
func ListWebhooks(repoFullName string) ([]GitHubWebhook, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: %s", repoFullName)
	}

	endpoint := fmt.Sprintf("repos/%s/%s/hooks", parts[0], parts[1])
	cmd := exec.Command("gh", "api", endpoint)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			// Check for permission error
			if strings.Contains(stderr, "Must have admin rights") || strings.Contains(stderr, "403") {
				return nil, cli.Err("insufficient permissions to manage webhooks (requires admin)")
			}
			return nil, cli.Err("%s", stderr)
		}
		return nil, err
	}

	var hooks []GitHubWebhook
	if err := json.Unmarshal(output, &hooks); err != nil {
		return nil, err
	}

	return hooks, nil
}

// CreateWebhook creates a new webhook in a repository.
func CreateWebhook(repoFullName string, name string, config WebhookConfig) error {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", repoFullName)
	}

	// Build the webhook payload
	payload := map[string]interface{}{
		"name":   "web",
		"active": true,
		"events": config.Events,
		"config": map[string]interface{}{
			"url":          config.URL,
			"content_type": config.ContentType,
			"insecure_ssl": "0",
		},
	}

	if config.Active != nil {
		payload["active"] = *config.Active
	}

	if config.Secret != "" {
		configMap := payload["config"].(map[string]interface{})
		configMap["secret"] = config.Secret
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("repos/%s/%s/hooks", parts[0], parts[1])
	cmd := exec.Command("gh", "api", endpoint, "--method", "POST", "--input", "-")
	cmd.Stdin = strings.NewReader(string(payloadJSON))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// UpdateWebhook updates an existing webhook.
func UpdateWebhook(repoFullName string, hookID int, config WebhookConfig) error {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", repoFullName)
	}

	payload := map[string]interface{}{
		"active": true,
		"events": config.Events,
		"config": map[string]interface{}{
			"url":          config.URL,
			"content_type": config.ContentType,
			"insecure_ssl": "0",
		},
	}

	if config.Active != nil {
		payload["active"] = *config.Active
	}

	if config.Secret != "" {
		configMap := payload["config"].(map[string]interface{})
		configMap["secret"] = config.Secret
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("repos/%s/%s/hooks/%d", parts[0], parts[1], hookID)
	cmd := exec.Command("gh", "api", endpoint, "--method", "PATCH", "--input", "-")
	cmd.Stdin = strings.NewReader(string(payloadJSON))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// SyncWebhooks synchronizes webhooks for a repository.
// Webhooks are matched by URL - if a webhook with the same URL exists, it's updated.
// Otherwise, a new webhook is created.
func SyncWebhooks(repoFullName string, config *GitHubConfig, dryRun bool) (*ChangeSet, error) {
	changes := NewChangeSet(repoFullName)

	// Skip if no webhooks configured
	if len(config.Webhooks) == 0 {
		return changes, nil
	}

	// Get existing webhooks
	existing, err := ListWebhooks(repoFullName)
	if err != nil {
		// If permission denied, note it but don't fail entirely
		if strings.Contains(err.Error(), "insufficient permissions") {
			changes.Add(CategoryWebhook, ChangeSkip, "all", "insufficient permissions")
			return changes, nil
		}
		return nil, cli.Wrap(err, "failed to list webhooks")
	}

	// Build lookup map by URL
	existingByURL := make(map[string]GitHubWebhook)
	for _, hook := range existing {
		existingByURL[hook.Config.URL] = hook
	}

	// Process each configured webhook
	for name, wantHook := range config.Webhooks {
		// Skip webhooks with empty URLs (env var not set)
		if wantHook.URL == "" {
			changes.Add(CategoryWebhook, ChangeSkip, name, "URL not configured")
			continue
		}

		existingHook, exists := existingByURL[wantHook.URL]

		if !exists {
			// Create new webhook
			changes.Add(CategoryWebhook, ChangeCreate, name, wantHook.URL)
			if !dryRun {
				if err := CreateWebhook(repoFullName, name, wantHook); err != nil {
					return changes, cli.Wrap(err, "failed to create webhook "+name)
				}
			}
			continue
		}

		// Check if update is needed
		needsUpdate := false
		details := make(map[string]string)

		// Check events
		if !stringSliceEqual(existingHook.Events, wantHook.Events) {
			needsUpdate = true
			details["events"] = fmt.Sprintf("%v -> %v", existingHook.Events, wantHook.Events)
		}

		// Check content type
		if existingHook.Config.ContentType != wantHook.ContentType {
			needsUpdate = true
			details["content_type"] = fmt.Sprintf("%s -> %s", existingHook.Config.ContentType, wantHook.ContentType)
		}

		// Check active state
		wantActive := true
		if wantHook.Active != nil {
			wantActive = *wantHook.Active
		}
		if existingHook.Active != wantActive {
			needsUpdate = true
			details["active"] = fmt.Sprintf("%v -> %v", existingHook.Active, wantActive)
		}

		if needsUpdate {
			changes.AddWithDetails(CategoryWebhook, ChangeUpdate, name, "", details)
			if !dryRun {
				if err := UpdateWebhook(repoFullName, existingHook.ID, wantHook); err != nil {
					return changes, cli.Wrap(err, "failed to update webhook "+name)
				}
			}
		} else {
			changes.Add(CategoryWebhook, ChangeSkip, name, "up to date")
		}
	}

	return changes, nil
}

// stringSliceEqual compares two string slices for equality (order-independent).
// Uses frequency counting to properly handle duplicates.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	// Count frequencies in slice a
	counts := make(map[string]int)
	for _, s := range a {
		counts[s]++
	}
	// Decrement for each element in slice b
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	// All counts should be zero if slices are equal
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}
