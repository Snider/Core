// cmd_ci.go implements the 'php ci' command for CI/CD pipeline integration.
//
// Usage:
//   core php ci                 # Run full CI pipeline
//   core php ci --json          # Output combined JSON report
//   core php ci --summary       # Output markdown summary
//   core php ci --sarif         # Generate SARIF files
//   core php ci --upload-sarif  # Upload SARIF to GitHub Security
//   core php ci --fail-on=high  # Only fail on high+ severity

package php

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// CI command flags
var (
	ciJSON        bool
	ciSummary     bool
	ciSARIF       bool
	ciUploadSARIF bool
	ciFailOn      string
)

// CIResult represents the overall CI pipeline result
type CIResult struct {
	Passed    bool            `json:"passed"`
	ExitCode  int             `json:"exit_code"`
	Duration  string          `json:"duration"`
	StartedAt time.Time       `json:"started_at"`
	Checks    []CICheckResult `json:"checks"`
	Summary   CISummary       `json:"summary"`
	Artifacts []string        `json:"artifacts,omitempty"`
}

// CICheckResult represents an individual check result
type CICheckResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // passed, failed, warning, skipped
	Duration string `json:"duration"`
	Details  string `json:"details,omitempty"`
	Issues   int    `json:"issues,omitempty"`
	Errors   int    `json:"errors,omitempty"`
	Warnings int    `json:"warnings,omitempty"`
}

// CISummary contains aggregate statistics
type CISummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Warnings int `json:"warnings"`
	Skipped  int `json:"skipped"`
}

func addPHPCICommand(parent *cobra.Command) {
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: i18n.T("cmd.php.ci.short"),
		Long:  i18n.T("cmd.php.ci.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPCI()
		},
	}

	ciCmd.Flags().BoolVar(&ciJSON, "json", false, i18n.T("cmd.php.ci.flag.json"))
	ciCmd.Flags().BoolVar(&ciSummary, "summary", false, i18n.T("cmd.php.ci.flag.summary"))
	ciCmd.Flags().BoolVar(&ciSARIF, "sarif", false, i18n.T("cmd.php.ci.flag.sarif"))
	ciCmd.Flags().BoolVar(&ciUploadSARIF, "upload-sarif", false, i18n.T("cmd.php.ci.flag.upload_sarif"))
	ciCmd.Flags().StringVar(&ciFailOn, "fail-on", "error", i18n.T("cmd.php.ci.flag.fail_on"))

	parent.AddCommand(ciCmd)
}

func runPHPCI() error {
	cwd, err := os.Getwd()
	if err != nil {
		return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
	}

	if !IsPHPProject(cwd) {
		return errors.New(i18n.T("cmd.php.error.not_php"))
	}

	startTime := time.Now()
	ctx := context.Background()

	// Define checks to run in order
	checks := []struct {
		name  string
		run   func(context.Context, string) (CICheckResult, error)
		sarif bool // Whether this check can generate SARIF
	}{
		{"test", runCITest, false},
		{"stan", runCIStan, true},
		{"psalm", runCIPsalm, true},
		{"fmt", runCIFmt, false},
		{"audit", runCIAudit, false},
		{"security", runCISecurity, false},
	}

	result := CIResult{
		StartedAt: startTime,
		Passed:    true,
		Checks:    make([]CICheckResult, 0, len(checks)),
	}

	var artifacts []string

	// Print header unless JSON output
	if !ciJSON {
		cli.Print("\n%s\n", cli.BoldStyle.Render("core php ci - QA Pipeline"))
		cli.Print("%s\n\n", strings.Repeat("─", 40))
	}

	// Run each check
	for _, check := range checks {
		if !ciJSON {
			cli.Print("  %s %s...", dimStyle.Render("→"), check.name)
		}

		checkResult, err := check.run(ctx, cwd)
		if err != nil {
			checkResult = CICheckResult{
				Name:    check.name,
				Status:  "failed",
				Details: err.Error(),
			}
		}

		result.Checks = append(result.Checks, checkResult)

		// Update summary
		result.Summary.Total++
		switch checkResult.Status {
		case "passed":
			result.Summary.Passed++
		case "failed":
			result.Summary.Failed++
			if shouldFailOn(checkResult, ciFailOn) {
				result.Passed = false
			}
		case "warning":
			result.Summary.Warnings++
		case "skipped":
			result.Summary.Skipped++
		}

		// Print result
		if !ciJSON {
			cli.Print("\r  %s %s %s\n", getStatusIcon(checkResult.Status), check.name, dimStyle.Render(checkResult.Details))
		}

		// Generate SARIF if requested
		if (ciSARIF || ciUploadSARIF) && check.sarif {
			sarifFile := filepath.Join(cwd, check.name+".sarif")
			if generateSARIF(ctx, cwd, check.name, sarifFile) == nil {
				artifacts = append(artifacts, sarifFile)
			}
		}
	}

	result.Duration = time.Since(startTime).Round(time.Millisecond).String()
	result.Artifacts = artifacts

	// Set exit code
	if result.Passed {
		result.ExitCode = 0
	} else {
		result.ExitCode = 1
	}

	// Output based on flags
	if ciJSON {
		if err := outputCIJSON(result); err != nil {
			return err
		}
		if !result.Passed {
			return cli.Exit(result.ExitCode, cli.Err("CI pipeline failed"))
		}
		return nil
	}

	if ciSummary {
		if err := outputCISummary(result); err != nil {
			return err
		}
		if !result.Passed {
			return cli.Err("CI pipeline failed")
		}
		return nil
	}

	// Default table output
	cli.Print("\n%s\n", strings.Repeat("─", 40))

	if result.Passed {
		cli.Print("%s %s\n", successStyle.Render("✓ CI PASSED"), dimStyle.Render(result.Duration))
	} else {
		cli.Print("%s %s\n", errorStyle.Render("✗ CI FAILED"), dimStyle.Render(result.Duration))
	}

	if len(artifacts) > 0 {
		cli.Print("\n%s\n", dimStyle.Render("Artifacts:"))
		for _, a := range artifacts {
			cli.Print("  → %s\n", filepath.Base(a))
		}
	}

	// Upload SARIF if requested
	if ciUploadSARIF && len(artifacts) > 0 {
		cli.Blank()
		for _, sarifFile := range artifacts {
			if err := uploadSARIFToGitHub(ctx, sarifFile); err != nil {
				cli.Print("  %s %s: %s\n", errorStyle.Render("✗"), filepath.Base(sarifFile), err)
			} else {
				cli.Print("  %s %s uploaded\n", successStyle.Render("✓"), filepath.Base(sarifFile))
			}
		}
	}

	if !result.Passed {
		return cli.Err("CI pipeline failed")
	}
	return nil
}

// runCITest runs Pest/PHPUnit tests
func runCITest(ctx context.Context, dir string) (CICheckResult, error) {
	start := time.Now()
	result := CICheckResult{Name: "test", Status: "passed"}

	opts := TestOptions{
		Dir:    dir,
		Output: nil, // Suppress output
	}

	if err := RunTests(ctx, opts); err != nil {
		result.Status = "failed"
		result.Details = err.Error()
	} else {
		result.Details = "all tests passed"
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// runCIStan runs PHPStan
func runCIStan(ctx context.Context, dir string) (CICheckResult, error) {
	start := time.Now()
	result := CICheckResult{Name: "stan", Status: "passed"}

	_, found := DetectAnalyser(dir)
	if !found {
		result.Status = "skipped"
		result.Details = "PHPStan not configured"
		return result, nil
	}

	opts := AnalyseOptions{
		Dir:    dir,
		Output: nil,
	}

	if err := Analyse(ctx, opts); err != nil {
		result.Status = "failed"
		result.Details = "errors found"
	} else {
		result.Details = "0 errors"
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// runCIPsalm runs Psalm
func runCIPsalm(ctx context.Context, dir string) (CICheckResult, error) {
	start := time.Now()
	result := CICheckResult{Name: "psalm", Status: "passed"}

	_, found := DetectPsalm(dir)
	if !found {
		result.Status = "skipped"
		result.Details = "Psalm not configured"
		return result, nil
	}

	opts := PsalmOptions{
		Dir:    dir,
		Output: nil,
	}

	if err := RunPsalm(ctx, opts); err != nil {
		result.Status = "failed"
		result.Details = "errors found"
	} else {
		result.Details = "0 errors"
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// runCIFmt checks code formatting
func runCIFmt(ctx context.Context, dir string) (CICheckResult, error) {
	start := time.Now()
	result := CICheckResult{Name: "fmt", Status: "passed"}

	_, found := DetectFormatter(dir)
	if !found {
		result.Status = "skipped"
		result.Details = "no formatter configured"
		return result, nil
	}

	opts := FormatOptions{
		Dir:    dir,
		Fix:    false, // Check only
		Output: nil,
	}

	if err := Format(ctx, opts); err != nil {
		result.Status = "warning"
		result.Details = "formatting issues"
	} else {
		result.Details = "code style OK"
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// runCIAudit runs composer audit
func runCIAudit(ctx context.Context, dir string) (CICheckResult, error) {
	start := time.Now()
	result := CICheckResult{Name: "audit", Status: "passed"}

	results, err := RunAudit(ctx, AuditOptions{
		Dir:    dir,
		Output: nil,
	})
	if err != nil {
		result.Status = "failed"
		result.Details = err.Error()
		result.Duration = time.Since(start).Round(time.Millisecond).String()
		return result, nil
	}

	totalVulns := 0
	for _, r := range results {
		totalVulns += r.Vulnerabilities
	}

	if totalVulns > 0 {
		result.Status = "failed"
		result.Details = fmt.Sprintf("%d vulnerabilities", totalVulns)
		result.Issues = totalVulns
	} else {
		result.Details = "no vulnerabilities"
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// runCISecurity runs security checks
func runCISecurity(ctx context.Context, dir string) (CICheckResult, error) {
	start := time.Now()
	result := CICheckResult{Name: "security", Status: "passed"}

	secResult, err := RunSecurityChecks(ctx, SecurityOptions{
		Dir:    dir,
		Output: nil,
	})
	if err != nil {
		result.Status = "failed"
		result.Details = err.Error()
		result.Duration = time.Since(start).Round(time.Millisecond).String()
		return result, nil
	}

	if secResult.Summary.Critical > 0 || secResult.Summary.High > 0 {
		result.Status = "failed"
		result.Details = fmt.Sprintf("%d critical, %d high", secResult.Summary.Critical, secResult.Summary.High)
		result.Issues = secResult.Summary.Critical + secResult.Summary.High
	} else if secResult.Summary.Medium > 0 {
		result.Status = "warning"
		result.Details = fmt.Sprintf("%d medium issues", secResult.Summary.Medium)
		result.Warnings = secResult.Summary.Medium
	} else {
		result.Details = "no issues"
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// shouldFailOn determines if a check should cause CI failure based on --fail-on
func shouldFailOn(check CICheckResult, level string) bool {
	switch level {
	case "critical":
		return check.Status == "failed" && check.Issues > 0
	case "high", "error":
		return check.Status == "failed"
	case "warning":
		return check.Status == "failed" || check.Status == "warning"
	default:
		return check.Status == "failed"
	}
}

// getStatusIcon returns the icon for a check status
func getStatusIcon(status string) string {
	switch status {
	case "passed":
		return successStyle.Render("✓")
	case "failed":
		return errorStyle.Render("✗")
	case "warning":
		return phpQAWarningStyle.Render("⚠")
	case "skipped":
		return dimStyle.Render("-")
	default:
		return dimStyle.Render("?")
	}
}

// outputCIJSON outputs the result as JSON
func outputCIJSON(result CIResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// outputCISummary outputs a markdown summary
func outputCISummary(result CIResult) error {
	var sb strings.Builder

	sb.WriteString("## CI Pipeline Results\n\n")

	if result.Passed {
		sb.WriteString("**Status:** ✅ Passed\n\n")
	} else {
		sb.WriteString("**Status:** ❌ Failed\n\n")
	}

	sb.WriteString("| Check | Status | Details |\n")
	sb.WriteString("|-------|--------|----------|\n")

	for _, check := range result.Checks {
		icon := "✅"
		switch check.Status {
		case "failed":
			icon = "❌"
		case "warning":
			icon = "⚠️"
		case "skipped":
			icon = "⏭️"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", check.Name, icon, check.Details))
	}

	sb.WriteString(fmt.Sprintf("\n**Duration:** %s\n", result.Duration))

	fmt.Print(sb.String())
	return nil
}

// generateSARIF generates a SARIF file for a specific check
func generateSARIF(ctx context.Context, dir, checkName, outputFile string) error {
	var args []string

	switch checkName {
	case "stan":
		args = []string{"vendor/bin/phpstan", "analyse", "--error-format=sarif", "--no-progress"}
	case "psalm":
		args = []string{"vendor/bin/psalm", "--output-format=sarif"}
	default:
		return fmt.Errorf("SARIF not supported for %s", checkName)
	}

	cmd := exec.CommandContext(ctx, "php", args...)
	cmd.Dir = dir

	// Capture output - command may exit non-zero when issues are found
	// but still produce valid SARIF output
	output, err := cmd.CombinedOutput()
	if len(output) == 0 {
		if err != nil {
			return fmt.Errorf("failed to generate SARIF: %w", err)
		}
		return fmt.Errorf("no SARIF output generated")
	}

	// Validate output is valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(output, &js); err != nil {
		return fmt.Errorf("invalid SARIF output: %w", err)
	}

	return os.WriteFile(outputFile, output, 0644)
}

// uploadSARIFToGitHub uploads a SARIF file to GitHub Security tab
func uploadSARIFToGitHub(ctx context.Context, sarifFile string) error {
	// Validate commit SHA before calling API
	sha := getGitSHA()
	if sha == "" {
		return errors.New("cannot upload SARIF: git commit SHA not available (ensure you're in a git repository)")
	}

	// Use gh CLI to upload
	cmd := exec.CommandContext(ctx, "gh", "api",
		"repos/{owner}/{repo}/code-scanning/sarifs",
		"-X", "POST",
		"-F", "sarif=@"+sarifFile,
		"-F", "ref="+getGitRef(),
		"-F", "commit_sha="+sha,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// getGitRef returns the current git ref
func getGitRef() string {
	cmd := exec.Command("git", "symbolic-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "refs/heads/main"
	}
	return strings.TrimSpace(string(output))
}

// getGitSHA returns the current git commit SHA
func getGitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
