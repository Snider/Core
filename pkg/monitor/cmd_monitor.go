// cmd_monitor.go implements the 'monitor' command for aggregating security findings.
//
// Usage:
//   core monitor                    # Monitor current repo
//   core monitor --repo X           # Monitor specific repo
//   core monitor --all              # Monitor all repos in registry
//   core monitor --severity high    # Filter by severity
//   core monitor --json             # Output as JSON

package monitor

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/errors"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

// Command flags
var (
	monitorRepo     string
	monitorSeverity []string
	monitorJSON     bool
	monitorAll      bool
)

// Finding represents a security finding from any source
type Finding struct {
	Source      string `json:"source"`      // semgrep, trivy, dependabot, secret-scanning, etc.
	Severity    string `json:"severity"`    // critical, high, medium, low
	Rule        string `json:"rule"`        // Rule ID or CVE
	File        string `json:"file"`        // Affected file path
	Line        int    `json:"line"`        // Line number (0 if N/A)
	Message     string `json:"message"`     // Description
	URL         string `json:"url"`         // Link to finding
	State       string `json:"state"`       // open, dismissed, fixed
	RepoName    string `json:"repo"`        // Repository name
	CreatedAt   string `json:"created_at"`  // When found
	Labels      []string `json:"suggested_labels,omitempty"`
}

// CodeScanningAlert represents a GitHub code scanning alert
type CodeScanningAlert struct {
	Number    int    `json:"number"`
	State     string `json:"state"` // open, dismissed, fixed
	Rule      struct {
		ID          string `json:"id"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
	} `json:"rule"`
	Tool struct {
		Name string `json:"name"`
	} `json:"tool"`
	MostRecentInstance struct {
		Location struct {
			Path      string `json:"path"`
			StartLine int    `json:"start_line"`
		} `json:"location"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	} `json:"most_recent_instance"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
}

// DependabotAlert represents a GitHub Dependabot alert
type DependabotAlert struct {
	Number    int    `json:"number"`
	State     string `json:"state"` // open, dismissed, fixed
	SecurityVulnerability struct {
		Severity string `json:"severity"`
		Package  struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
	} `json:"security_vulnerability"`
	SecurityAdvisory struct {
		CVEID       string `json:"cve_id"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
	} `json:"security_advisory"`
	Dependency struct {
		ManifestPath string `json:"manifest_path"`
	} `json:"dependency"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
}

// SecretScanningAlert represents a GitHub secret scanning alert
type SecretScanningAlert struct {
	Number       int    `json:"number"`
	State        string `json:"state"` // open, resolved
	SecretType   string `json:"secret_type"`
	Secret       string `json:"secret"` // Partial, redacted
	HTMLURL      string `json:"html_url"`
	LocationType string `json:"location_type"`
	CreatedAt    string `json:"created_at"`
}

func runMonitor() error {
	// Check gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.E("monitor", i18n.T("error.gh_not_found"), err)
	}

	// Determine repos to scan
	repoList, err := resolveRepos()
	if err != nil {
		return err
	}

	if len(repoList) == 0 {
		return errors.E("monitor", i18n.T("cmd.monitor.error.no_repos"), nil)
	}

	// Collect all findings and errors
	var allFindings []Finding
	var fetchErrors []string
	for _, repo := range repoList {
		if !monitorJSON {
			cli.Print("\033[2K\r%s %s...", dimStyle.Render(i18n.T("cmd.monitor.scanning")), repo)
		}

		findings, errs := fetchRepoFindings(repo)
		allFindings = append(allFindings, findings...)
		fetchErrors = append(fetchErrors, errs...)
	}

	// Filter by severity if specified
	if len(monitorSeverity) > 0 {
		allFindings = filterBySeverity(allFindings, monitorSeverity)
	}

	// Sort by severity (critical first)
	sortBySeverity(allFindings)

	// Output
	if monitorJSON {
		return outputJSON(allFindings)
	}

	cli.Print("\033[2K\r") // Clear scanning line

	// Show any fetch errors as warnings
	if len(fetchErrors) > 0 {
		for _, e := range fetchErrors {
			cli.Print("%s %s\n", warningStyle.Render("!"), e)
		}
		cli.Blank()
	}

	return outputTable(allFindings)
}

// resolveRepos determines which repos to scan
func resolveRepos() ([]string, error) {
	if monitorRepo != "" {
		// Specific repo - if fully qualified (org/repo), use as-is
		if strings.Contains(monitorRepo, "/") {
			return []string{monitorRepo}, nil
		}
		// Otherwise, try to detect org from git remote, fallback to host-uk
		// Note: Users outside host-uk org should use fully qualified names
		org := detectOrgFromGit()
		if org == "" {
			org = "host-uk"
		}
		return []string{org + "/" + monitorRepo}, nil
	}

	if monitorAll {
		// All repos from registry
		registry, err := repos.FindRegistry()
		if err != nil {
			return nil, errors.E("monitor", "failed to find registry", err)
		}

		loaded, err := repos.LoadRegistry(registry)
		if err != nil {
			return nil, errors.E("monitor", "failed to load registry", err)
		}

		var repoList []string
		for _, r := range loaded.Repos {
			repoList = append(repoList, loaded.Org+"/"+r.Name)
		}
		return repoList, nil
	}

	// Default to current repo
	repo, err := detectRepoFromGit()
	if err != nil {
		return nil, err
	}
	return []string{repo}, nil
}

// fetchRepoFindings fetches all security findings for a repo
// Returns findings and any errors encountered (errors don't stop other fetches)
func fetchRepoFindings(repoFullName string) ([]Finding, []string) {
	var findings []Finding
	var errs []string
	repoName := strings.Split(repoFullName, "/")[1]

	// Fetch code scanning alerts
	codeFindings, err := fetchCodeScanningAlerts(repoFullName)
	if err != nil {
		errs = append(errs, fmt.Sprintf("%s: code-scanning: %s", repoName, err))
	}
	findings = append(findings, codeFindings...)

	// Fetch Dependabot alerts
	depFindings, err := fetchDependabotAlerts(repoFullName)
	if err != nil {
		errs = append(errs, fmt.Sprintf("%s: dependabot: %s", repoName, err))
	}
	findings = append(findings, depFindings...)

	// Fetch secret scanning alerts
	secretFindings, err := fetchSecretScanningAlerts(repoFullName)
	if err != nil {
		errs = append(errs, fmt.Sprintf("%s: secret-scanning: %s", repoName, err))
	}
	findings = append(findings, secretFindings...)

	return findings, errs
}

// fetchCodeScanningAlerts fetches code scanning alerts
func fetchCodeScanningAlerts(repoFullName string) ([]Finding, error) {
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/code-scanning/alerts", repoFullName),
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		// Check for expected "not enabled" responses vs actual errors
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			// These are expected conditions, not errors
			if strings.Contains(stderr, "Advanced Security must be enabled") ||
				strings.Contains(stderr, "no analysis found") ||
				strings.Contains(stderr, "Not Found") {
				return nil, nil
			}
		}
		return nil, errors.E("monitor.fetchCodeScanning", "API request failed", err)
	}

	var alerts []CodeScanningAlert
	if err := json.Unmarshal(output, &alerts); err != nil {
		return nil, errors.E("monitor.fetchCodeScanning", "failed to parse response", err)
	}

	repoName := strings.Split(repoFullName, "/")[1]
	var findings []Finding
	for _, alert := range alerts {
		if alert.State != "open" {
			continue
		}
		f := Finding{
			Source:    alert.Tool.Name,
			Severity:  normalizeSeverity(alert.Rule.Severity),
			Rule:      alert.Rule.ID,
			File:      alert.MostRecentInstance.Location.Path,
			Line:      alert.MostRecentInstance.Location.StartLine,
			Message:   alert.MostRecentInstance.Message.Text,
			URL:       alert.HTMLURL,
			State:     alert.State,
			RepoName:  repoName,
			CreatedAt: alert.CreatedAt,
			Labels:    []string{"type:security"},
		}
		if f.Message == "" {
			f.Message = alert.Rule.Description
		}
		findings = append(findings, f)
	}

	return findings, nil
}

// fetchDependabotAlerts fetches Dependabot alerts
func fetchDependabotAlerts(repoFullName string) ([]Finding, error) {
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/dependabot/alerts", repoFullName),
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			// Dependabot not enabled is expected
			if strings.Contains(stderr, "Dependabot alerts are not enabled") ||
				strings.Contains(stderr, "Not Found") {
				return nil, nil
			}
		}
		return nil, errors.E("monitor.fetchDependabot", "API request failed", err)
	}

	var alerts []DependabotAlert
	if err := json.Unmarshal(output, &alerts); err != nil {
		return nil, errors.E("monitor.fetchDependabot", "failed to parse response", err)
	}

	repoName := strings.Split(repoFullName, "/")[1]
	var findings []Finding
	for _, alert := range alerts {
		if alert.State != "open" {
			continue
		}
		f := Finding{
			Source:    "dependabot",
			Severity:  normalizeSeverity(alert.SecurityVulnerability.Severity),
			Rule:      alert.SecurityAdvisory.CVEID,
			File:      alert.Dependency.ManifestPath,
			Line:      0,
			Message:   fmt.Sprintf("%s: %s", alert.SecurityVulnerability.Package.Name, alert.SecurityAdvisory.Summary),
			URL:       alert.HTMLURL,
			State:     alert.State,
			RepoName:  repoName,
			CreatedAt: alert.CreatedAt,
			Labels:    []string{"type:security", "dependencies"},
		}
		findings = append(findings, f)
	}

	return findings, nil
}

// fetchSecretScanningAlerts fetches secret scanning alerts
func fetchSecretScanningAlerts(repoFullName string) ([]Finding, error) {
	args := []string{
		"api",
		fmt.Sprintf("repos/%s/secret-scanning/alerts", repoFullName),
	}

	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			// Secret scanning not enabled is expected
			if strings.Contains(stderr, "Secret scanning is disabled") ||
				strings.Contains(stderr, "Not Found") {
				return nil, nil
			}
		}
		return nil, errors.E("monitor.fetchSecretScanning", "API request failed", err)
	}

	var alerts []SecretScanningAlert
	if err := json.Unmarshal(output, &alerts); err != nil {
		return nil, errors.E("monitor.fetchSecretScanning", "failed to parse response", err)
	}

	repoName := strings.Split(repoFullName, "/")[1]
	var findings []Finding
	for _, alert := range alerts {
		if alert.State != "open" {
			continue
		}
		f := Finding{
			Source:   "secret-scanning",
			Severity: "critical", // Secrets are always critical
			Rule:     alert.SecretType,
			File:     alert.LocationType,
			Line:     0,
			Message:  fmt.Sprintf("Exposed %s detected", alert.SecretType),
			URL:      alert.HTMLURL,
			State:    alert.State,
			RepoName: repoName,
			CreatedAt: alert.CreatedAt,
			Labels:   []string{"type:security", "secrets"},
		}
		findings = append(findings, f)
	}

	return findings, nil
}

// normalizeSeverity normalizes severity strings to standard values
func normalizeSeverity(s string) string {
	s = strings.ToLower(s)
	switch s {
	case "critical", "crit":
		return "critical"
	case "high", "error":
		return "high"
	case "medium", "moderate", "warning":
		return "medium"
	case "low", "info", "note":
		return "low"
	default:
		return "medium"
	}
}

// filterBySeverity filters findings by severity
func filterBySeverity(findings []Finding, severities []string) []Finding {
	sevSet := make(map[string]bool)
	for _, s := range severities {
		sevSet[strings.ToLower(s)] = true
	}

	var filtered []Finding
	for _, f := range findings {
		if sevSet[f.Severity] {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// sortBySeverity sorts findings by severity (critical first)
func sortBySeverity(findings []Finding) {
	severityOrder := map[string]int{
		"critical": 0,
		"high":     1,
		"medium":   2,
		"low":      3,
	}

	sort.Slice(findings, func(i, j int) bool {
		oi := severityOrder[findings[i].Severity]
		oj := severityOrder[findings[j].Severity]
		if oi != oj {
			return oi < oj
		}
		return findings[i].RepoName < findings[j].RepoName
	})
}

// outputJSON outputs findings as JSON
func outputJSON(findings []Finding) error {
	data, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return errors.E("monitor", "failed to marshal findings", err)
	}
	cli.Print("%s\n", string(data))
	return nil
}

// outputTable outputs findings as a formatted table
func outputTable(findings []Finding) error {
	if len(findings) == 0 {
		cli.Print("%s\n", successStyle.Render(i18n.T("cmd.monitor.no_findings")))
		return nil
	}

	// Count by severity
	counts := make(map[string]int)
	for _, f := range findings {
		counts[f.Severity]++
	}

	// Header summary
	var parts []string
	if counts["critical"] > 0 {
		parts = append(parts, errorStyle.Render(fmt.Sprintf("%d critical", counts["critical"])))
	}
	if counts["high"] > 0 {
		parts = append(parts, errorStyle.Render(fmt.Sprintf("%d high", counts["high"])))
	}
	if counts["medium"] > 0 {
		parts = append(parts, warningStyle.Render(fmt.Sprintf("%d medium", counts["medium"])))
	}
	if counts["low"] > 0 {
		parts = append(parts, dimStyle.Render(fmt.Sprintf("%d low", counts["low"])))
	}
	cli.Print("%s: %s\n", i18n.T("cmd.monitor.found"), strings.Join(parts, ", "))
	cli.Blank()

	// Group by repo
	byRepo := make(map[string][]Finding)
	for _, f := range findings {
		byRepo[f.RepoName] = append(byRepo[f.RepoName], f)
	}

	// Sort repos for consistent output
	repoNames := make([]string, 0, len(byRepo))
	for repo := range byRepo {
		repoNames = append(repoNames, repo)
	}
	sort.Strings(repoNames)

	// Print by repo
	for _, repo := range repoNames {
		repoFindings := byRepo[repo]
		cli.Print("%s\n", cli.BoldStyle.Render(repo))
		for _, f := range repoFindings {
			sevStyle := dimStyle
			switch f.Severity {
			case "critical", "high":
				sevStyle = errorStyle
			case "medium":
				sevStyle = warningStyle
			}

			// Format: [severity] source: message (file:line)
			location := ""
			if f.File != "" {
				location = f.File
				if f.Line > 0 {
					location = fmt.Sprintf("%s:%d", f.File, f.Line)
				}
			}

			cli.Print("  %s %s: %s",
				sevStyle.Render(fmt.Sprintf("[%s]", f.Severity)),
				dimStyle.Render(f.Source),
				truncate(f.Message, 60))
			if location != "" {
				cli.Print(" %s", dimStyle.Render("("+location+")"))
			}
			cli.Blank()
		}
		cli.Blank()
	}

	return nil
}

// truncate truncates a string to max runes (Unicode-safe)
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// detectRepoFromGit detects the repo from git remote
func detectRepoFromGit() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.E("monitor", i18n.T("cmd.monitor.error.not_git_repo"), err)
	}

	url := strings.TrimSpace(string(output))
	return parseGitHubRepo(url)
}

// detectOrgFromGit tries to detect the org from git remote
func detectOrgFromGit() string {
	repo, err := detectRepoFromGit()
	if err != nil {
		return ""
	}
	parts := strings.Split(repo, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// parseGitHubRepo extracts org/repo from a git URL
func parseGitHubRepo(url string) (string, error) {
	// Handle SSH URLs: git@github.com:org/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		path := strings.TrimPrefix(url, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		return path, nil
	}

	// Handle HTTPS URLs: https://github.com/org/repo.git
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) >= 2 {
			path := strings.TrimSuffix(parts[1], ".git")
			return path, nil
		}
	}

	return "", fmt.Errorf("could not parse GitHub repo from URL: %s", url)
}
