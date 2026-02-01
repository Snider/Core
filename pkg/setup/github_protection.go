// github_protection.go implements GitHub branch protection synchronization.
//
// Uses the gh api command for branch protection operations:
//   - gh api repos/{owner}/{repo}/branches/{branch}/protection --method GET
//   - gh api repos/{owner}/{repo}/branches/{branch}/protection --method PUT

package setup

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/cli"
)

// GitHubBranchProtection represents branch protection rules from the GitHub API.
type GitHubBranchProtection struct {
	RequiredStatusChecks           *RequiredStatusChecks           `json:"required_status_checks"`
	RequiredPullRequestReviews     *RequiredPullRequestReviews     `json:"required_pull_request_reviews"`
	EnforceAdmins                  *EnforceAdmins                  `json:"enforce_admins"`
	RequiredLinearHistory          *RequiredLinearHistory          `json:"required_linear_history"`
	AllowForcePushes               *AllowForcePushes               `json:"allow_force_pushes"`
	AllowDeletions                 *AllowDeletions                 `json:"allow_deletions"`
	RequiredConversationResolution *RequiredConversationResolution `json:"required_conversation_resolution"`
}

// RequiredStatusChecks defines required CI checks.
type RequiredStatusChecks struct {
	Strict   bool     `json:"strict"`
	Contexts []string `json:"contexts"`
}

// RequiredPullRequestReviews defines review requirements.
type RequiredPullRequestReviews struct {
	DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews      bool `json:"require_code_owner_reviews"`
	RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
}

// EnforceAdmins indicates if admins are subject to rules.
type EnforceAdmins struct {
	Enabled bool `json:"enabled"`
}

// RequiredLinearHistory indicates if linear history is required.
type RequiredLinearHistory struct {
	Enabled bool `json:"enabled"`
}

// AllowForcePushes indicates if force pushes are allowed.
type AllowForcePushes struct {
	Enabled bool `json:"enabled"`
}

// AllowDeletions indicates if branch deletion is allowed.
type AllowDeletions struct {
	Enabled bool `json:"enabled"`
}

// RequiredConversationResolution indicates if conversation resolution is required.
type RequiredConversationResolution struct {
	Enabled bool `json:"enabled"`
}

// GetBranchProtection fetches branch protection rules for a branch.
func GetBranchProtection(repoFullName, branch string) (*GitHubBranchProtection, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format: %s", repoFullName)
	}

	endpoint := fmt.Sprintf("repos/%s/%s/branches/%s/protection", parts[0], parts[1], branch)
	cmd := exec.Command("gh", "api", endpoint)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			// Branch protection not enabled returns 404
			if strings.Contains(stderr, "404") || strings.Contains(stderr, "Branch not protected") {
				return nil, nil // No protection set
			}
			if strings.Contains(stderr, "403") {
				return nil, cli.Err("insufficient permissions to manage branch protection (requires admin)")
			}
			return nil, cli.Err("%s", stderr)
		}
		return nil, err
	}

	var protection GitHubBranchProtection
	if err := json.Unmarshal(output, &protection); err != nil {
		return nil, err
	}

	return &protection, nil
}

// SetBranchProtection sets branch protection rules for a branch.
func SetBranchProtection(repoFullName, branch string, config BranchProtectionConfig) error {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repo format: %s", repoFullName)
	}

	// Build the protection payload
	payload := map[string]interface{}{
		"enforce_admins":                   config.EnforceAdmins,
		"required_linear_history":          config.RequireLinearHistory,
		"allow_force_pushes":               config.AllowForcePushes,
		"allow_deletions":                  config.AllowDeletions,
		"required_conversation_resolution": config.RequireConversationResolution,
	}

	// Required pull request reviews
	if config.RequiredReviews > 0 {
		payload["required_pull_request_reviews"] = map[string]interface{}{
			"dismiss_stale_reviews":           config.DismissStale,
			"require_code_owner_reviews":      config.RequireCodeOwnerReviews,
			"required_approving_review_count": config.RequiredReviews,
		}
	} else {
		payload["required_pull_request_reviews"] = nil
	}

	// Required status checks
	if len(config.RequiredStatusChecks) > 0 {
		payload["required_status_checks"] = map[string]interface{}{
			"strict":   true,
			"contexts": config.RequiredStatusChecks,
		}
	} else {
		payload["required_status_checks"] = nil
	}

	// Restrictions (required but can be empty for non-org repos)
	payload["restrictions"] = nil

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("repos/%s/%s/branches/%s/protection", parts[0], parts[1], branch)
	cmd := exec.Command("gh", "api", endpoint, "--method", "PUT", "--input", "-")
	cmd.Stdin = strings.NewReader(string(payloadJSON))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cli.Err("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// SyncBranchProtection synchronizes branch protection for a repository.
func SyncBranchProtection(repoFullName string, config *GitHubConfig, dryRun bool) (*ChangeSet, error) {
	changes := NewChangeSet(repoFullName)

	// Skip if no branch protection configured
	if len(config.BranchProtection) == 0 {
		return changes, nil
	}

	// Process each configured branch
	for _, wantProtection := range config.BranchProtection {
		branch := wantProtection.Branch

		// Get existing protection
		existing, err := GetBranchProtection(repoFullName, branch)
		if err != nil {
			// If permission denied, note it but don't fail
			if strings.Contains(err.Error(), "insufficient permissions") {
				changes.Add(CategoryProtection, ChangeSkip, branch, "insufficient permissions")
				continue
			}
			return nil, cli.Wrap(err, "failed to get protection for "+branch)
		}

		// Check if protection needs to be created or updated
		if existing == nil {
			// Create new protection
			changes.Add(CategoryProtection, ChangeCreate, branch, describeProtection(wantProtection))
			if !dryRun {
				if err := SetBranchProtection(repoFullName, branch, wantProtection); err != nil {
					return changes, cli.Wrap(err, "failed to set protection for "+branch)
				}
			}
			continue
		}

		// Compare and check if update is needed
		needsUpdate := false
		details := make(map[string]string)

		// Check required reviews
		existingReviews := 0
		existingDismissStale := false
		existingCodeOwner := false
		if existing.RequiredPullRequestReviews != nil {
			existingReviews = existing.RequiredPullRequestReviews.RequiredApprovingReviewCount
			existingDismissStale = existing.RequiredPullRequestReviews.DismissStaleReviews
			existingCodeOwner = existing.RequiredPullRequestReviews.RequireCodeOwnerReviews
		}

		if existingReviews != wantProtection.RequiredReviews {
			needsUpdate = true
			details["required_reviews"] = fmt.Sprintf("%d -> %d", existingReviews, wantProtection.RequiredReviews)
		}
		if existingDismissStale != wantProtection.DismissStale {
			needsUpdate = true
			details["dismiss_stale"] = fmt.Sprintf("%v -> %v", existingDismissStale, wantProtection.DismissStale)
		}
		if existingCodeOwner != wantProtection.RequireCodeOwnerReviews {
			needsUpdate = true
			details["code_owner_reviews"] = fmt.Sprintf("%v -> %v", existingCodeOwner, wantProtection.RequireCodeOwnerReviews)
		}

		// Check enforce admins
		existingEnforceAdmins := false
		if existing.EnforceAdmins != nil {
			existingEnforceAdmins = existing.EnforceAdmins.Enabled
		}
		if existingEnforceAdmins != wantProtection.EnforceAdmins {
			needsUpdate = true
			details["enforce_admins"] = fmt.Sprintf("%v -> %v", existingEnforceAdmins, wantProtection.EnforceAdmins)
		}

		// Check linear history
		existingLinear := false
		if existing.RequiredLinearHistory != nil {
			existingLinear = existing.RequiredLinearHistory.Enabled
		}
		if existingLinear != wantProtection.RequireLinearHistory {
			needsUpdate = true
			details["linear_history"] = fmt.Sprintf("%v -> %v", existingLinear, wantProtection.RequireLinearHistory)
		}

		// Check force pushes
		existingForcePush := false
		if existing.AllowForcePushes != nil {
			existingForcePush = existing.AllowForcePushes.Enabled
		}
		if existingForcePush != wantProtection.AllowForcePushes {
			needsUpdate = true
			details["allow_force_pushes"] = fmt.Sprintf("%v -> %v", existingForcePush, wantProtection.AllowForcePushes)
		}

		// Check deletions
		existingDeletions := false
		if existing.AllowDeletions != nil {
			existingDeletions = existing.AllowDeletions.Enabled
		}
		if existingDeletions != wantProtection.AllowDeletions {
			needsUpdate = true
			details["allow_deletions"] = fmt.Sprintf("%v -> %v", existingDeletions, wantProtection.AllowDeletions)
		}

		// Check required status checks
		var existingStatusChecks []string
		if existing.RequiredStatusChecks != nil {
			existingStatusChecks = existing.RequiredStatusChecks.Contexts
		}
		if !stringSliceEqual(existingStatusChecks, wantProtection.RequiredStatusChecks) {
			needsUpdate = true
			details["status_checks"] = fmt.Sprintf("%v -> %v", existingStatusChecks, wantProtection.RequiredStatusChecks)
		}

		if needsUpdate {
			changes.AddWithDetails(CategoryProtection, ChangeUpdate, branch, "", details)
			if !dryRun {
				if err := SetBranchProtection(repoFullName, branch, wantProtection); err != nil {
					return changes, cli.Wrap(err, "failed to update protection for "+branch)
				}
			}
		} else {
			changes.Add(CategoryProtection, ChangeSkip, branch, "up to date")
		}
	}

	return changes, nil
}

// describeProtection returns a human-readable description of protection rules.
func describeProtection(p BranchProtectionConfig) string {
	var parts []string
	if p.RequiredReviews > 0 {
		parts = append(parts, fmt.Sprintf("%d review(s)", p.RequiredReviews))
	}
	if p.DismissStale {
		parts = append(parts, "dismiss stale")
	}
	if p.EnforceAdmins {
		parts = append(parts, "enforce admins")
	}
	if len(parts) == 0 {
		return "basic protection"
	}
	return strings.Join(parts, ", ")
}
