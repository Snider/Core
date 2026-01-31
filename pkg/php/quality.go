package php

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/cli"
)

// FormatOptions configures PHP code formatting.
type FormatOptions struct {
	// Dir is the project directory (defaults to current working directory).
	Dir string

	// Fix automatically fixes formatting issues.
	Fix bool

	// Diff shows a diff of changes instead of modifying files.
	Diff bool

	// Paths limits formatting to specific paths.
	Paths []string

	// Output is the writer for output (defaults to os.Stdout).
	Output io.Writer
}

// AnalyseOptions configures PHP static analysis.
type AnalyseOptions struct {
	// Dir is the project directory (defaults to current working directory).
	Dir string

	// Level is the PHPStan analysis level (0-9).
	Level int

	// Paths limits analysis to specific paths.
	Paths []string

	// Memory is the memory limit for analysis (e.g., "2G").
	Memory string

	// Output is the writer for output (defaults to os.Stdout).
	Output io.Writer
}

// FormatterType represents the detected formatter.
type FormatterType string

const (
	FormatterPint FormatterType = "pint"
)

// AnalyserType represents the detected static analyser.
type AnalyserType string

const (
	AnalyserPHPStan  AnalyserType = "phpstan"
	AnalyserLarastan AnalyserType = "larastan"
)

// DetectFormatter detects which formatter is available in the project.
func DetectFormatter(dir string) (FormatterType, bool) {
	// Check for Pint config
	pintConfig := filepath.Join(dir, "pint.json")
	if _, err := os.Stat(pintConfig); err == nil {
		return FormatterPint, true
	}

	// Check for vendor binary
	pintBin := filepath.Join(dir, "vendor", "bin", "pint")
	if _, err := os.Stat(pintBin); err == nil {
		return FormatterPint, true
	}

	return "", false
}

// DetectAnalyser detects which static analyser is available in the project.
func DetectAnalyser(dir string) (AnalyserType, bool) {
	// Check for PHPStan config
	phpstanConfig := filepath.Join(dir, "phpstan.neon")
	phpstanDistConfig := filepath.Join(dir, "phpstan.neon.dist")

	hasConfig := false
	if _, err := os.Stat(phpstanConfig); err == nil {
		hasConfig = true
	}
	if _, err := os.Stat(phpstanDistConfig); err == nil {
		hasConfig = true
	}

	// Check for vendor binary
	phpstanBin := filepath.Join(dir, "vendor", "bin", "phpstan")
	hasBin := false
	if _, err := os.Stat(phpstanBin); err == nil {
		hasBin = true
	}

	if hasConfig || hasBin {
		// Check if it's Larastan (Laravel-specific PHPStan)
		larastanPath := filepath.Join(dir, "vendor", "larastan", "larastan")
		if _, err := os.Stat(larastanPath); err == nil {
			return AnalyserLarastan, true
		}
		// Also check nunomaduro/larastan
		larastanPath2 := filepath.Join(dir, "vendor", "nunomaduro", "larastan")
		if _, err := os.Stat(larastanPath2); err == nil {
			return AnalyserLarastan, true
		}
		return AnalyserPHPStan, true
	}

	return "", false
}

// Format runs Laravel Pint to format PHP code.
func Format(ctx context.Context, opts FormatOptions) error {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Check if formatter is available
	formatter, found := DetectFormatter(opts.Dir)
	if !found {
		return cli.Err("no formatter found (install Laravel Pint: composer require laravel/pint --dev)")
	}

	var cmdName string
	var args []string

	switch formatter {
	case FormatterPint:
		cmdName, args = buildPintCommand(opts)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// Analyse runs PHPStan or Larastan for static analysis.
func Analyse(ctx context.Context, opts AnalyseOptions) error {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Check if analyser is available
	analyser, found := DetectAnalyser(opts.Dir)
	if !found {
		return cli.Err("no static analyser found (install PHPStan: composer require phpstan/phpstan --dev)")
	}

	var cmdName string
	var args []string

	switch analyser {
	case AnalyserPHPStan, AnalyserLarastan:
		cmdName, args = buildPHPStanCommand(opts)
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// buildPintCommand builds the command for running Laravel Pint.
func buildPintCommand(opts FormatOptions) (string, []string) {
	// Check for vendor binary first
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "pint")
	cmdName := "pint"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	var args []string

	if !opts.Fix {
		args = append(args, "--test")
	}

	if opts.Diff {
		args = append(args, "--diff")
	}

	// Add specific paths if provided
	args = append(args, opts.Paths...)

	return cmdName, args
}

// buildPHPStanCommand builds the command for running PHPStan.
func buildPHPStanCommand(opts AnalyseOptions) (string, []string) {
	// Check for vendor binary first
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "phpstan")
	cmdName := "phpstan"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	args := []string{"analyse"}

	if opts.Level > 0 {
		args = append(args, "--level", cli.Sprintf("%d", opts.Level))
	}

	if opts.Memory != "" {
		args = append(args, "--memory-limit", opts.Memory)
	}

	// Add specific paths if provided
	args = append(args, opts.Paths...)

	return cmdName, args
}

// =============================================================================
// Psalm Static Analysis
// =============================================================================

// PsalmOptions configures Psalm static analysis.
type PsalmOptions struct {
	Dir      string
	Level    int  // Error level (1=strictest, 8=most lenient)
	Fix      bool // Auto-fix issues where possible
	Baseline bool // Generate/update baseline file
	ShowInfo bool // Show info-level issues
	Output   io.Writer
}

// PsalmType represents the detected Psalm configuration.
type PsalmType string

const (
	PsalmStandard PsalmType = "psalm"
)

// DetectPsalm checks if Psalm is available in the project.
func DetectPsalm(dir string) (PsalmType, bool) {
	// Check for psalm.xml config
	psalmConfig := filepath.Join(dir, "psalm.xml")
	psalmDistConfig := filepath.Join(dir, "psalm.xml.dist")

	hasConfig := false
	if _, err := os.Stat(psalmConfig); err == nil {
		hasConfig = true
	}
	if _, err := os.Stat(psalmDistConfig); err == nil {
		hasConfig = true
	}

	// Check for vendor binary
	psalmBin := filepath.Join(dir, "vendor", "bin", "psalm")
	if _, err := os.Stat(psalmBin); err == nil {
		return PsalmStandard, true
	}

	if hasConfig {
		return PsalmStandard, true
	}

	return "", false
}

// RunPsalm runs Psalm static analysis.
func RunPsalm(ctx context.Context, opts PsalmOptions) error {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Build command
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "psalm")
	cmdName := "psalm"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	args := []string{"--no-progress"}

	if opts.Level > 0 && opts.Level <= 8 {
		args = append(args, cli.Sprintf("--error-level=%d", opts.Level))
	}

	if opts.Fix {
		args = append(args, "--alter", "--issues=all")
	}

	if opts.Baseline {
		args = append(args, "--set-baseline=psalm-baseline.xml")
	}

	if opts.ShowInfo {
		args = append(args, "--show-info=true")
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// =============================================================================
// Security Audit
// =============================================================================

// AuditOptions configures dependency security auditing.
type AuditOptions struct {
	Dir    string
	JSON   bool // Output in JSON format
	Fix    bool // Auto-fix vulnerabilities (npm only)
	Output io.Writer
}

// AuditResult holds the results of a security audit.
type AuditResult struct {
	Tool            string
	Vulnerabilities int
	Advisories      []AuditAdvisory
	Error           error
}

// AuditAdvisory represents a single security advisory.
type AuditAdvisory struct {
	Package     string
	Severity    string
	Title       string
	URL         string
	Identifiers []string
}

// RunAudit runs security audits on dependencies.
func RunAudit(ctx context.Context, opts AuditOptions) ([]AuditResult, error) {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	var results []AuditResult

	// Run composer audit
	composerResult := runComposerAudit(ctx, opts)
	results = append(results, composerResult)

	// Run npm audit if package.json exists
	if _, err := os.Stat(filepath.Join(opts.Dir, "package.json")); err == nil {
		npmResult := runNpmAudit(ctx, opts)
		results = append(results, npmResult)
	}

	return results, nil
}

func runComposerAudit(ctx context.Context, opts AuditOptions) AuditResult {
	result := AuditResult{Tool: "composer"}

	args := []string{"audit", "--format=json"}

	cmd := exec.CommandContext(ctx, "composer", args...)
	cmd.Dir = opts.Dir

	output, err := cmd.Output()
	if err != nil {
		// composer audit returns non-zero if vulnerabilities found
		if exitErr, ok := err.(*exec.ExitError); ok {
			output = append(output, exitErr.Stderr...)
		}
	}

	// Parse JSON output
	var auditData struct {
		Advisories map[string][]struct {
			Title          string `json:"title"`
			Link           string `json:"link"`
			CVE            string `json:"cve"`
			AffectedRanges string `json:"affectedVersions"`
		} `json:"advisories"`
	}

	if jsonErr := json.Unmarshal(output, &auditData); jsonErr == nil {
		for pkg, advisories := range auditData.Advisories {
			for _, adv := range advisories {
				result.Advisories = append(result.Advisories, AuditAdvisory{
					Package:     pkg,
					Title:       adv.Title,
					URL:         adv.Link,
					Identifiers: []string{adv.CVE},
				})
			}
		}
		result.Vulnerabilities = len(result.Advisories)
	} else if err != nil {
		result.Error = err
	}

	return result
}

func runNpmAudit(ctx context.Context, opts AuditOptions) AuditResult {
	result := AuditResult{Tool: "npm"}

	args := []string{"audit", "--json"}
	if opts.Fix {
		args = []string{"audit", "fix"}
	}

	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = opts.Dir

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			output = append(output, exitErr.Stderr...)
		}
	}

	if !opts.Fix {
		// Parse JSON output
		var auditData struct {
			Metadata struct {
				Vulnerabilities struct {
					Total int `json:"total"`
				} `json:"vulnerabilities"`
			} `json:"metadata"`
			Vulnerabilities map[string]struct {
				Severity string `json:"severity"`
				Via      []any  `json:"via"`
			} `json:"vulnerabilities"`
		}

		if jsonErr := json.Unmarshal(output, &auditData); jsonErr == nil {
			result.Vulnerabilities = auditData.Metadata.Vulnerabilities.Total
			for pkg, vuln := range auditData.Vulnerabilities {
				result.Advisories = append(result.Advisories, AuditAdvisory{
					Package:  pkg,
					Severity: vuln.Severity,
				})
			}
		} else if err != nil {
			result.Error = err
		}
	}

	return result
}

// =============================================================================
// Rector Automated Refactoring
// =============================================================================

// RectorOptions configures Rector code refactoring.
type RectorOptions struct {
	Dir        string
	Fix        bool // Apply changes (default is dry-run)
	Diff       bool // Show detailed diff
	ClearCache bool // Clear cache before running
	Output     io.Writer
}

// DetectRector checks if Rector is available in the project.
func DetectRector(dir string) bool {
	// Check for rector.php config
	rectorConfig := filepath.Join(dir, "rector.php")
	if _, err := os.Stat(rectorConfig); err == nil {
		return true
	}

	// Check for vendor binary
	rectorBin := filepath.Join(dir, "vendor", "bin", "rector")
	if _, err := os.Stat(rectorBin); err == nil {
		return true
	}

	return false
}

// RunRector runs Rector for automated code refactoring.
func RunRector(ctx context.Context, opts RectorOptions) error {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Build command
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "rector")
	cmdName := "rector"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	args := []string{"process"}

	if !opts.Fix {
		args = append(args, "--dry-run")
	}

	if opts.Diff {
		args = append(args, "--output-format", "diff")
	}

	if opts.ClearCache {
		args = append(args, "--clear-cache")
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// =============================================================================
// Infection Mutation Testing
// =============================================================================

// InfectionOptions configures Infection mutation testing.
type InfectionOptions struct {
	Dir           string
	MinMSI        int    // Minimum mutation score indicator (0-100)
	MinCoveredMSI int    // Minimum covered mutation score (0-100)
	Threads       int    // Number of parallel threads
	Filter        string // Filter files by pattern
	OnlyCovered   bool   // Only mutate covered code
	Output        io.Writer
}

// DetectInfection checks if Infection is available in the project.
func DetectInfection(dir string) bool {
	// Check for infection config files
	configs := []string{"infection.json", "infection.json5", "infection.json.dist"}
	for _, config := range configs {
		if _, err := os.Stat(filepath.Join(dir, config)); err == nil {
			return true
		}
	}

	// Check for vendor binary
	infectionBin := filepath.Join(dir, "vendor", "bin", "infection")
	if _, err := os.Stat(infectionBin); err == nil {
		return true
	}

	return false
}

// RunInfection runs Infection mutation testing.
func RunInfection(ctx context.Context, opts InfectionOptions) error {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	// Build command
	vendorBin := filepath.Join(opts.Dir, "vendor", "bin", "infection")
	cmdName := "infection"
	if _, err := os.Stat(vendorBin); err == nil {
		cmdName = vendorBin
	}

	var args []string

	// Set defaults
	minMSI := opts.MinMSI
	if minMSI == 0 {
		minMSI = 50
	}
	minCoveredMSI := opts.MinCoveredMSI
	if minCoveredMSI == 0 {
		minCoveredMSI = 70
	}
	threads := opts.Threads
	if threads == 0 {
		threads = 4
	}

	args = append(args, cli.Sprintf("--min-msi=%d", minMSI))
	args = append(args, cli.Sprintf("--min-covered-msi=%d", minCoveredMSI))
	args = append(args, cli.Sprintf("--threads=%d", threads))

	if opts.Filter != "" {
		args = append(args, "--filter="+opts.Filter)
	}

	if opts.OnlyCovered {
		args = append(args, "--only-covered")
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Dir = opts.Dir
	cmd.Stdout = opts.Output
	cmd.Stderr = opts.Output

	return cmd.Run()
}

// =============================================================================
// QA Pipeline
// =============================================================================

// QAOptions configures the full QA pipeline.
type QAOptions struct {
	Dir   string
	Quick bool // Only run quick checks
	Full  bool // Run all stages including slow checks
	Fix   bool // Auto-fix issues where possible
	JSON  bool // Output results as JSON
}

// QAStage represents a stage in the QA pipeline.
type QAStage string

const (
	QAStageQuick    QAStage = "quick"
	QAStageStandard QAStage = "standard"
	QAStageFull     QAStage = "full"
)

// QACheckResult holds the result of a single QA check.
type QACheckResult struct {
	Name     string
	Stage    QAStage
	Passed   bool
	Duration string
	Error    error
	Output   string
}

// QAResult holds the results of the full QA pipeline.
type QAResult struct {
	Stages  []QAStage
	Checks  []QACheckResult
	Passed  bool
	Summary string
}

// GetQAStages returns the stages to run based on options.
func GetQAStages(opts QAOptions) []QAStage {
	if opts.Quick {
		return []QAStage{QAStageQuick}
	}
	if opts.Full {
		return []QAStage{QAStageQuick, QAStageStandard, QAStageFull}
	}
	// Default: quick + standard
	return []QAStage{QAStageQuick, QAStageStandard}
}

// GetQAChecks returns the checks for a given stage.
func GetQAChecks(dir string, stage QAStage) []string {
	switch stage {
	case QAStageQuick:
		checks := []string{"audit", "fmt", "stan"}
		return checks
	case QAStageStandard:
		checks := []string{}
		if _, found := DetectPsalm(dir); found {
			checks = append(checks, "psalm")
		}
		checks = append(checks, "test")
		return checks
	case QAStageFull:
		checks := []string{}
		if DetectRector(dir) {
			checks = append(checks, "rector")
		}
		if DetectInfection(dir) {
			checks = append(checks, "infection")
		}
		return checks
	}
	return nil
}

// =============================================================================
// Security Checks
// =============================================================================

// SecurityOptions configures security scanning.
type SecurityOptions struct {
	Dir      string
	Severity string // Minimum severity (critical, high, medium, low)
	JSON     bool   // Output in JSON format
	SARIF    bool   // Output in SARIF format
	URL      string // URL to check HTTP headers (optional)
	Output   io.Writer
}

// SecurityResult holds the results of security scanning.
type SecurityResult struct {
	Checks  []SecurityCheck
	Summary SecuritySummary
}

// SecurityCheck represents a single security check result.
type SecurityCheck struct {
	ID          string
	Name        string
	Description string
	Severity    string
	Passed      bool
	Message     string
	Fix         string
	CWE         string
}

// SecuritySummary summarizes security check results.
type SecuritySummary struct {
	Total    int
	Passed   int
	Critical int
	High     int
	Medium   int
	Low      int
}

// RunSecurityChecks runs security checks on the project.
func RunSecurityChecks(ctx context.Context, opts SecurityOptions) (*SecurityResult, error) {
	if opts.Dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, cli.WrapVerb(err, "get", "working directory")
		}
		opts.Dir = cwd
	}

	result := &SecurityResult{}

	// Run composer audit
	auditResults, _ := RunAudit(ctx, AuditOptions{Dir: opts.Dir})
	for _, audit := range auditResults {
		check := SecurityCheck{
			ID:          audit.Tool + "_audit",
			Name:        strings.Title(audit.Tool) + " Security Audit",
			Description: "Check " + audit.Tool + " dependencies for vulnerabilities",
			Severity:    "critical",
			Passed:      audit.Vulnerabilities == 0 && audit.Error == nil,
			CWE:         "CWE-1395",
		}
		if !check.Passed {
			check.Message = cli.Sprintf("Found %d vulnerabilities", audit.Vulnerabilities)
		}
		result.Checks = append(result.Checks, check)
	}

	// Check .env file for security issues
	envChecks := runEnvSecurityChecks(opts.Dir)
	result.Checks = append(result.Checks, envChecks...)

	// Check filesystem security
	fsChecks := runFilesystemSecurityChecks(opts.Dir)
	result.Checks = append(result.Checks, fsChecks...)

	// Calculate summary
	for _, check := range result.Checks {
		result.Summary.Total++
		if check.Passed {
			result.Summary.Passed++
		} else {
			switch check.Severity {
			case "critical":
				result.Summary.Critical++
			case "high":
				result.Summary.High++
			case "medium":
				result.Summary.Medium++
			case "low":
				result.Summary.Low++
			}
		}
	}

	return result, nil
}

func runEnvSecurityChecks(dir string) []SecurityCheck {
	var checks []SecurityCheck

	envPath := filepath.Join(dir, ".env")
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		return checks
	}

	envLines := strings.Split(string(envContent), "\n")
	envMap := make(map[string]string)
	for _, line := range envLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Check APP_DEBUG
	if debug, ok := envMap["APP_DEBUG"]; ok {
		check := SecurityCheck{
			ID:          "debug_mode",
			Name:        "Debug Mode Disabled",
			Description: "APP_DEBUG should be false in production",
			Severity:    "critical",
			Passed:      strings.ToLower(debug) != "true",
			CWE:         "CWE-215",
		}
		if !check.Passed {
			check.Message = "Debug mode exposes sensitive information"
			check.Fix = "Set APP_DEBUG=false in .env"
		}
		checks = append(checks, check)
	}

	// Check APP_KEY
	if key, ok := envMap["APP_KEY"]; ok {
		check := SecurityCheck{
			ID:          "app_key_set",
			Name:        "Application Key Set",
			Description: "APP_KEY must be set and valid",
			Severity:    "critical",
			Passed:      len(key) >= 32,
			CWE:         "CWE-321",
		}
		if !check.Passed {
			check.Message = "Missing or weak encryption key"
			check.Fix = "Run: php artisan key:generate"
		}
		checks = append(checks, check)
	}

	// Check APP_URL for HTTPS
	if url, ok := envMap["APP_URL"]; ok {
		check := SecurityCheck{
			ID:          "https_enforced",
			Name:        "HTTPS Enforced",
			Description: "APP_URL should use HTTPS in production",
			Severity:    "high",
			Passed:      strings.HasPrefix(url, "https://"),
			CWE:         "CWE-319",
		}
		if !check.Passed {
			check.Message = "Application not using HTTPS"
			check.Fix = "Update APP_URL to use https://"
		}
		checks = append(checks, check)
	}

	return checks
}

func runFilesystemSecurityChecks(dir string) []SecurityCheck {
	var checks []SecurityCheck

	// Check .env not in public
	publicEnvPaths := []string{"public/.env", "public_html/.env"}
	for _, path := range publicEnvPaths {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			checks = append(checks, SecurityCheck{
				ID:          "env_not_public",
				Name:        ".env Not Publicly Accessible",
				Description: ".env file should not be in public directory",
				Severity:    "critical",
				Passed:      false,
				Message:     "Environment file exposed to web at " + path,
				CWE:         "CWE-538",
			})
		}
	}

	// Check .git not in public
	publicGitPaths := []string{"public/.git", "public_html/.git"}
	for _, path := range publicGitPaths {
		fullPath := filepath.Join(dir, path)
		if _, err := os.Stat(fullPath); err == nil {
			checks = append(checks, SecurityCheck{
				ID:          "git_not_public",
				Name:        ".git Not Publicly Accessible",
				Description: ".git directory should not be in public",
				Severity:    "critical",
				Passed:      false,
				Message:     "Git repository exposed to web (source code leak)",
				CWE:         "CWE-538",
			})
		}
	}

	return checks
}
