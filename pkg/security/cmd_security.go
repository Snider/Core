package security

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
)

var (
	// Command flags
	securityRegistryPath string
	securityRepo         string
	securitySeverity     string
	securityJSON         bool
)

// AddSecurityCommands adds the 'security' command to the root.
func AddSecurityCommands(root *cli.Command) {
	secCmd := &cli.Command{
		Use:   "security",
		Short: i18n.T("cmd.security.short"),
		Long:  i18n.T("cmd.security.long"),
	}

	addAlertsCommand(secCmd)
	addDepsCommand(secCmd)
	addScanCommand(secCmd)
	addSecretsCommand(secCmd)

	root.AddCommand(secCmd)
}

// DependabotAlert represents a Dependabot vulnerability alert.
type DependabotAlert struct {
	Number   int    `json:"number"`
	State    string `json:"state"`
	Advisory struct {
		Severity    string `json:"severity"`
		CVEID       string `json:"cve_id"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
	} `json:"security_advisory"`
	Dependency struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		ManifestPath string `json:"manifest_path"`
	} `json:"dependency"`
	SecurityVulnerability struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		FirstPatchedVersion struct {
			Identifier string `json:"identifier"`
		} `json:"first_patched_version"`
		VulnerableVersionRange string `json:"vulnerable_version_range"`
	} `json:"security_vulnerability"`
}

// CodeScanningAlert represents a code scanning alert.
type CodeScanningAlert struct {
	Number          int    `json:"number"`
	State           string `json:"state"`
	DismissedReason string `json:"dismissed_reason"`
	Rule            struct {
		ID          string `json:"id"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
		Tags        []string `json:"tags"`
	} `json:"rule"`
	Tool struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"tool"`
	MostRecentInstance struct {
		Location struct {
			Path      string `json:"path"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		} `json:"location"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	} `json:"most_recent_instance"`
}

// SecretScanningAlert represents a secret scanning alert.
type SecretScanningAlert struct {
	Number       int    `json:"number"`
	State        string `json:"state"`
	SecretType   string `json:"secret_type"`
	Secret       string `json:"secret"`
	PushProtection bool `json:"push_protection_bypassed"`
	Resolution   string `json:"resolution"`
}

// loadRegistry loads the repository registry.
func loadRegistry(registryPath string) (*repos.Registry, error) {
	if registryPath != "" {
		reg, err := repos.LoadRegistry(registryPath)
		if err != nil {
			return nil, cli.Wrap(err, "load registry")
		}
		return reg, nil
	}

	path, err := repos.FindRegistry()
	if err != nil {
		return nil, cli.Wrap(err, "find registry")
	}
	reg, err := repos.LoadRegistry(path)
	if err != nil {
		return nil, cli.Wrap(err, "load registry")
	}
	return reg, nil
}

// checkGH verifies gh CLI is available.
func checkGH() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return errors.New(i18n.T("error.gh_not_found"))
	}
	return nil
}

// runGHAPI runs a gh api command and returns the output.
func runGHAPI(endpoint string) ([]byte, error) {
	cmd := exec.Command("gh", "api", endpoint, "--paginate")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			// Handle common errors gracefully
			if strings.Contains(stderr, "404") || strings.Contains(stderr, "Not Found") {
				return []byte("[]"), nil // Return empty array for not found
			}
			if strings.Contains(stderr, "403") {
				return nil, fmt.Errorf("access denied (check token permissions)")
			}
		}
		return nil, cli.Wrap(err, "run gh api")
	}
	return output, nil
}

// severityRank returns a numeric rank for severity (higher = more severe).
func severityRank(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// severityStyle returns the appropriate style for a severity level.
func severityStyle(severity string) *cli.AnsiStyle {
	switch strings.ToLower(severity) {
	case "critical":
		return cli.ErrorStyle
	case "high":
		return cli.WarningStyle
	case "medium":
		return cli.ValueStyle
	default:
		return cli.DimStyle
	}
}

// filterBySeverity checks if the severity matches the filter.
func filterBySeverity(severity, filter string) bool {
	if filter == "" {
		return true
	}

	severities := strings.Split(strings.ToLower(filter), ",")
	sev := strings.ToLower(severity)

	for _, s := range severities {
		if strings.TrimSpace(s) == sev {
			return true
		}
	}
	return false
}

// getReposToCheck returns the list of repos to check based on flags.
func getReposToCheck(reg *repos.Registry, repoFilter string) []*repos.Repo {
	if repoFilter != "" {
		if repo, ok := reg.Get(repoFilter); ok {
			return []*repos.Repo{repo}
		}
		return nil
	}
	return reg.List()
}

// AlertSummary holds aggregated alert counts.
type AlertSummary struct {
	Critical int
	High     int
	Medium   int
	Low      int
	Unknown  int
	Total    int
}

// Add increments summary counters for the provided severity.
func (s *AlertSummary) Add(severity string) {
	s.Total++
	switch strings.ToLower(severity) {
	case "critical":
		s.Critical++
	case "high":
		s.High++
	case "medium":
		s.Medium++
	case "low":
		s.Low++
	default:
		s.Unknown++
	}
}

// String renders a human-readable summary of alert counts.
func (s *AlertSummary) String() string {
	parts := []string{}
	if s.Critical > 0 {
		parts = append(parts, cli.ErrorStyle.Render(fmt.Sprintf("%d critical", s.Critical)))
	}
	if s.High > 0 {
		parts = append(parts, cli.WarningStyle.Render(fmt.Sprintf("%d high", s.High)))
	}
	if s.Medium > 0 {
		parts = append(parts, cli.ValueStyle.Render(fmt.Sprintf("%d medium", s.Medium)))
	}
	if s.Low > 0 {
		parts = append(parts, cli.DimStyle.Render(fmt.Sprintf("%d low", s.Low)))
	}
	if s.Unknown > 0 {
		parts = append(parts, cli.DimStyle.Render(fmt.Sprintf("%d unknown", s.Unknown)))
	}
	if len(parts) == 0 {
		return cli.SuccessStyle.Render("No alerts")
	}
	return strings.Join(parts, " | ")
}
