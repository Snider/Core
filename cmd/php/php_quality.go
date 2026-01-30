package php

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/spf13/cobra"
)

var (
	testParallel bool
	testCoverage bool
	testFilter   string
	testGroup    string
)

func addPHPTestCommand(parent *cobra.Command) {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run PHP tests (PHPUnit/Pest)",
		Long: "Run PHP tests using PHPUnit or Pest.\n\n" +
			"Auto-detects Pest if tests/Pest.php exists, otherwise uses PHPUnit.\n\n" +
			"Examples:\n" +
			"  core php test                    # Run all tests\n" +
			"  core php test --parallel         # Run tests in parallel\n" +
			"  core php test --coverage         # Run with coverage\n" +
			"  core php test --filter UserTest  # Filter by test name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Detect test runner
			runner := phppkg.DetectTestRunner(cwd)
			fmt.Printf("%s Running tests with %s\n\n", dimStyle.Render("PHP:"), runner)

			ctx := context.Background()

			opts := phppkg.TestOptions{
				Dir:      cwd,
				Filter:   testFilter,
				Parallel: testParallel,
				Coverage: testCoverage,
				Output:   os.Stdout,
			}

			if testGroup != "" {
				opts.Groups = []string{testGroup}
			}

			if err := phppkg.RunTests(ctx, opts); err != nil {
				return fmt.Errorf("tests failed: %w", err)
			}

			return nil
		},
	}

	testCmd.Flags().BoolVar(&testParallel, "parallel", false, "Run tests in parallel")
	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, "Generate code coverage")
	testCmd.Flags().StringVar(&testFilter, "filter", "", "Filter tests by name pattern")
	testCmd.Flags().StringVar(&testGroup, "group", "", "Run only tests in specified group")

	parent.AddCommand(testCmd)
}

var (
	fmtFix  bool
	fmtDiff bool
)

func addPHPFmtCommand(parent *cobra.Command) {
	fmtCmd := &cobra.Command{
		Use:   "fmt [paths...]",
		Short: "Format PHP code with Laravel Pint",
		Long: "Format PHP code using Laravel Pint.\n\n" +
			"Examples:\n" +
			"  core php fmt           # Check formatting (dry-run)\n" +
			"  core php fmt --fix     # Auto-fix formatting issues\n" +
			"  core php fmt --diff    # Show diff of changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Detect formatter
			formatter, found := phppkg.DetectFormatter(cwd)
			if !found {
				return fmt.Errorf("no formatter found (install Laravel Pint: composer require laravel/pint --dev)")
			}

			action := "Checking"
			if fmtFix {
				action = "Formatting"
			}
			fmt.Printf("%s %s code with %s\n\n", dimStyle.Render("PHP:"), action, formatter)

			ctx := context.Background()

			opts := phppkg.FormatOptions{
				Dir:    cwd,
				Fix:    fmtFix,
				Diff:   fmtDiff,
				Output: os.Stdout,
			}

			// Get any additional paths from args
			if len(args) > 0 {
				opts.Paths = args
			}

			if err := phppkg.Format(ctx, opts); err != nil {
				if fmtFix {
					return fmt.Errorf("formatting failed: %w", err)
				}
				return fmt.Errorf("formatting issues found: %w", err)
			}

			if fmtFix {
				fmt.Printf("\n%s Code formatted successfully\n", successStyle.Render("Done:"))
			} else {
				fmt.Printf("\n%s No formatting issues found\n", successStyle.Render("Done:"))
			}

			return nil
		},
	}

	fmtCmd.Flags().BoolVar(&fmtFix, "fix", false, "Auto-fix formatting issues")
	fmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, "Show diff of changes")

	parent.AddCommand(fmtCmd)
}

var (
	analyseLevel  int
	analyseMemory string
)

func addPHPAnalyseCommand(parent *cobra.Command) {
	analyseCmd := &cobra.Command{
		Use:   "analyse [paths...]",
		Short: "Run PHPStan static analysis",
		Long: "Run PHPStan or Larastan static analysis.\n\n" +
			"Auto-detects Larastan if installed, otherwise uses PHPStan.\n\n" +
			"Examples:\n" +
			"  core php analyse              # Run analysis\n" +
			"  core php analyse --level 9    # Run at max strictness\n" +
			"  core php analyse --memory 2G  # Increase memory limit",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Detect analyser
			analyser, found := phppkg.DetectAnalyser(cwd)
			if !found {
				return fmt.Errorf("no static analyser found (install PHPStan: composer require phpstan/phpstan --dev)")
			}

			fmt.Printf("%s Running static analysis with %s\n\n", dimStyle.Render("PHP:"), analyser)

			ctx := context.Background()

			opts := phppkg.AnalyseOptions{
				Dir:    cwd,
				Level:  analyseLevel,
				Memory: analyseMemory,
				Output: os.Stdout,
			}

			// Get any additional paths from args
			if len(args) > 0 {
				opts.Paths = args
			}

			if err := phppkg.Analyse(ctx, opts); err != nil {
				return fmt.Errorf("analysis found issues: %w", err)
			}

			fmt.Printf("\n%s No issues found\n", successStyle.Render("Done:"))
			return nil
		},
	}

	analyseCmd.Flags().IntVar(&analyseLevel, "level", 0, "PHPStan analysis level (0-9)")
	analyseCmd.Flags().StringVar(&analyseMemory, "memory", "", "Memory limit (e.g., 2G)")

	parent.AddCommand(analyseCmd)
}

// =============================================================================
// New QA Commands
// =============================================================================

var (
	psalmLevel    int
	psalmFix      bool
	psalmBaseline bool
	psalmShowInfo bool
)

func addPHPPsalmCommand(parent *cobra.Command) {
	psalmCmd := &cobra.Command{
		Use:   "psalm",
		Short: "Run Psalm static analysis",
		Long: "Run Psalm deep static analysis with Laravel plugin support.\n\n" +
			"Psalm provides deeper type inference than PHPStan and catches\n" +
			"different classes of bugs. Both should be run for best coverage.\n\n" +
			"Examples:\n" +
			"  core php psalm              # Run analysis\n" +
			"  core php psalm --fix        # Auto-fix issues where possible\n" +
			"  core php psalm --level 3    # Run at specific level (1-8)\n" +
			"  core php psalm --baseline   # Generate baseline file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Check if Psalm is available
			_, found := phppkg.DetectPsalm(cwd)
			if !found {
				fmt.Printf("%s Psalm not found\n\n", errorStyle.Render("Error:"))
				fmt.Printf("%s composer require --dev vimeo/psalm\n", dimStyle.Render("Install:"))
				fmt.Printf("%s ./vendor/bin/psalm --init\n", dimStyle.Render("Setup:"))
				return fmt.Errorf("psalm not installed")
			}

			action := "Analysing"
			if psalmFix {
				action = "Analysing and fixing"
			}
			fmt.Printf("%s %s code with Psalm\n\n", dimStyle.Render("Psalm:"), action)

			ctx := context.Background()

			opts := phppkg.PsalmOptions{
				Dir:      cwd,
				Level:    psalmLevel,
				Fix:      psalmFix,
				Baseline: psalmBaseline,
				ShowInfo: psalmShowInfo,
				Output:   os.Stdout,
			}

			if err := phppkg.RunPsalm(ctx, opts); err != nil {
				return fmt.Errorf("psalm found issues: %w", err)
			}

			fmt.Printf("\n%s No issues found\n", successStyle.Render("Done:"))
			return nil
		},
	}

	psalmCmd.Flags().IntVar(&psalmLevel, "level", 0, "Error level (1=strictest, 8=most lenient)")
	psalmCmd.Flags().BoolVar(&psalmFix, "fix", false, "Auto-fix issues where possible")
	psalmCmd.Flags().BoolVar(&psalmBaseline, "baseline", false, "Generate/update baseline file")
	psalmCmd.Flags().BoolVar(&psalmShowInfo, "show-info", false, "Show info-level issues")

	parent.AddCommand(psalmCmd)
}

var (
	auditJSONOutput bool
	auditFix        bool
)

func addPHPAuditCommand(parent *cobra.Command) {
	auditCmd := &cobra.Command{
		Use:   "audit",
		Short: "Security audit for dependencies",
		Long: "Check PHP and JavaScript dependencies for known vulnerabilities.\n\n" +
			"Runs composer audit and npm audit (if package.json exists).\n\n" +
			"Examples:\n" +
			"  core php audit           # Check all dependencies\n" +
			"  core php audit --json    # Output as JSON\n" +
			"  core php audit --fix     # Auto-fix where possible (npm only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			fmt.Printf("%s Scanning dependencies for vulnerabilities\n\n", dimStyle.Render("Audit:"))

			ctx := context.Background()

			results, err := phppkg.RunAudit(ctx, phppkg.AuditOptions{
				Dir:    cwd,
				JSON:   auditJSONOutput,
				Fix:    auditFix,
				Output: os.Stdout,
			})
			if err != nil {
				return fmt.Errorf("audit failed: %w", err)
			}

			// Print results
			totalVulns := 0
			hasErrors := false

			for _, result := range results {
				icon := successStyle.Render("✓")
				status := successStyle.Render("secure")

				if result.Error != nil {
					icon = errorStyle.Render("✗")
					status = errorStyle.Render("error")
					hasErrors = true
				} else if result.Vulnerabilities > 0 {
					icon = errorStyle.Render("✗")
					status = errorStyle.Render(fmt.Sprintf("%d vulnerabilities", result.Vulnerabilities))
					totalVulns += result.Vulnerabilities
				}

				fmt.Printf("  %s %s %s\n", icon, dimStyle.Render(result.Tool+":"), status)

				// Show advisories
				for _, adv := range result.Advisories {
					severity := adv.Severity
					if severity == "" {
						severity = "unknown"
					}
					sevStyle := getSeverityStyle(severity)
					fmt.Printf("      %s %s\n", sevStyle.Render("["+severity+"]"), adv.Package)
					if adv.Title != "" {
						fmt.Printf("               %s\n", dimStyle.Render(adv.Title))
					}
				}
			}

			fmt.Println()

			if totalVulns > 0 {
				fmt.Printf("%s Found %d vulnerabilities across dependencies\n", errorStyle.Render("Warning:"), totalVulns)
				fmt.Printf("%s composer update && npm update\n", dimStyle.Render("Fix:"))
				return fmt.Errorf("vulnerabilities found")
			}

			if hasErrors {
				return fmt.Errorf("audit completed with errors")
			}

			fmt.Printf("%s All dependencies are secure\n", successStyle.Render("Done:"))
			return nil
		},
	}

	auditCmd.Flags().BoolVar(&auditJSONOutput, "json", false, "Output in JSON format")
	auditCmd.Flags().BoolVar(&auditFix, "fix", false, "Auto-fix vulnerabilities (npm only)")

	parent.AddCommand(auditCmd)
}

var (
	securitySeverity   string
	securityJSONOutput bool
	securitySarif      bool
	securityURL        string
)

func addPHPSecurityCommand(parent *cobra.Command) {
	securityCmd := &cobra.Command{
		Use:   "security",
		Short: "Security vulnerability scanning",
		Long: "Scan for security vulnerabilities in configuration and code.\n\n" +
			"Checks environment config, file permissions, code patterns,\n" +
			"and runs security-focused static analysis.\n\n" +
			"Examples:\n" +
			"  core php security                    # Run all checks\n" +
			"  core php security --severity=high   # Only high+ severity\n" +
			"  core php security --json            # JSON output",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			fmt.Printf("%s Running security checks\n\n", dimStyle.Render("Security:"))

			ctx := context.Background()

			result, err := phppkg.RunSecurityChecks(ctx, phppkg.SecurityOptions{
				Dir:      cwd,
				Severity: securitySeverity,
				JSON:     securityJSONOutput,
				SARIF:    securitySarif,
				URL:      securityURL,
				Output:   os.Stdout,
			})
			if err != nil {
				return fmt.Errorf("security check failed: %w", err)
			}

			// Print results by category
			currentCategory := ""
			for _, check := range result.Checks {
				category := strings.Split(check.ID, "_")[0]
				if category != currentCategory {
					if currentCategory != "" {
						fmt.Println()
					}
					currentCategory = category
					fmt.Printf("  %s\n", dimStyle.Render(strings.ToUpper(category)+" CHECKS:"))
				}

				icon := successStyle.Render("✓")
				if !check.Passed {
					icon = getSeverityStyle(check.Severity).Render("✗")
				}

				fmt.Printf("    %s %s\n", icon, check.Name)
				if !check.Passed && check.Message != "" {
					fmt.Printf("        %s\n", dimStyle.Render(check.Message))
					if check.Fix != "" {
						fmt.Printf("        %s %s\n", dimStyle.Render("Fix:"), check.Fix)
					}
				}
			}

			fmt.Println()

			// Print summary
			fmt.Printf("%s Security scan complete\n", dimStyle.Render("Summary:"))
			fmt.Printf("  %s %d/%d\n", dimStyle.Render("Passed:"), result.Summary.Passed, result.Summary.Total)

			if result.Summary.Critical > 0 {
				fmt.Printf("  %s %d\n", phpSecurityCriticalStyle.Render("Critical:"), result.Summary.Critical)
			}
			if result.Summary.High > 0 {
				fmt.Printf("  %s %d\n", phpSecurityHighStyle.Render("High:"), result.Summary.High)
			}
			if result.Summary.Medium > 0 {
				fmt.Printf("  %s %d\n", phpSecurityMediumStyle.Render("Medium:"), result.Summary.Medium)
			}
			if result.Summary.Low > 0 {
				fmt.Printf("  %s %d\n", phpSecurityLowStyle.Render("Low:"), result.Summary.Low)
			}

			if result.Summary.Critical > 0 || result.Summary.High > 0 {
				return fmt.Errorf("critical or high severity issues found")
			}

			return nil
		},
	}

	securityCmd.Flags().StringVar(&securitySeverity, "severity", "", "Minimum severity (critical, high, medium, low)")
	securityCmd.Flags().BoolVar(&securityJSONOutput, "json", false, "Output in JSON format")
	securityCmd.Flags().BoolVar(&securitySarif, "sarif", false, "Output in SARIF format (for GitHub Security)")
	securityCmd.Flags().StringVar(&securityURL, "url", "", "URL to check HTTP headers (optional)")

	parent.AddCommand(securityCmd)
}

var (
	qaQuick bool
	qaFull  bool
	qaFix   bool
)

func addPHPQACommand(parent *cobra.Command) {
	qaCmd := &cobra.Command{
		Use:   "qa",
		Short: "Run full QA pipeline",
		Long: "Run the complete quality assurance pipeline.\n\n" +
			"Stages:\n" +
			"  quick:    Security audit, code style, PHPStan\n" +
			"  standard: Psalm, tests\n" +
			"  full:     Rector dry-run, mutation testing (slow)\n\n" +
			"Examples:\n" +
			"  core php qa              # Run quick + standard stages\n" +
			"  core php qa --quick      # Only quick checks\n" +
			"  core php qa --full       # All stages including slow ones\n" +
			"  core php qa --fix        # Auto-fix where possible",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Determine stages
			opts := phppkg.QAOptions{
				Dir:   cwd,
				Quick: qaQuick,
				Full:  qaFull,
				Fix:   qaFix,
			}
			stages := phppkg.GetQAStages(opts)

			// Print header
			stageNames := make([]string, len(stages))
			for i, s := range stages {
				stageNames[i] = string(s)
			}
			fmt.Printf("%s Running QA pipeline (%s)\n\n", dimStyle.Render("QA:"), strings.Join(stageNames, " → "))

			ctx := context.Background()
			var allPassed = true
			var results []phppkg.QACheckResult

			for _, stage := range stages {
				fmt.Printf("%s\n", phpQAStageStyle.Render("═══ "+strings.ToUpper(string(stage))+" STAGE ═══"))

				checks := phppkg.GetQAChecks(cwd, stage)
				if len(checks) == 0 {
					fmt.Printf("  %s\n\n", dimStyle.Render("No checks available"))
					continue
				}

				for _, checkName := range checks {
					result := runQACheck(ctx, cwd, checkName, qaFix)
					result.Stage = stage
					results = append(results, result)

					icon := phpQAPassedStyle.Render("✓")
					status := phpQAPassedStyle.Render("passed")
					if !result.Passed {
						icon = phpQAFailedStyle.Render("✗")
						status = phpQAFailedStyle.Render("failed")
						allPassed = false
					}

					fmt.Printf("  %s %s %s %s\n", icon, result.Name, status, dimStyle.Render(result.Duration))
				}
				fmt.Println()
			}

			// Print summary
			passedCount := 0
			var failedChecks []phppkg.QACheckResult
			for _, r := range results {
				if r.Passed {
					passedCount++
				} else {
					failedChecks = append(failedChecks, r)
				}
			}

			if allPassed {
				fmt.Printf("%s All checks passed (%d/%d)\n", phpQAPassedStyle.Render("QA PASSED:"), passedCount, len(results))
				return nil
			}

			fmt.Printf("%s Some checks failed (%d/%d passed)\n\n", phpQAFailedStyle.Render("QA FAILED:"), passedCount, len(results))

			// Show what needs fixing
			fmt.Printf("%s\n", dimStyle.Render("To fix:"))
			for _, check := range failedChecks {
				fixCmd := getQAFixCommand(check.Name, qaFix)
				issue := check.Output
				if issue == "" {
					issue = "issues found"
				}
				fmt.Printf("  %s %s\n", phpQAFailedStyle.Render("•"), check.Name+": "+issue)
				if fixCmd != "" {
					fmt.Printf("    %s %s\n", dimStyle.Render("→"), fixCmd)
				}
			}

			return fmt.Errorf("QA pipeline failed")
		},
	}

	qaCmd.Flags().BoolVar(&qaQuick, "quick", false, "Only run quick checks")
	qaCmd.Flags().BoolVar(&qaFull, "full", false, "Run all stages including slow checks")
	qaCmd.Flags().BoolVar(&qaFix, "fix", false, "Auto-fix issues where possible")

	parent.AddCommand(qaCmd)
}

func getQAFixCommand(checkName string, fixEnabled bool) string {
	switch checkName {
	case "audit":
		return "composer update && npm update"
	case "fmt":
		if fixEnabled {
			return ""
		}
		return "core php fmt --fix"
	case "analyse":
		return "Fix PHPStan errors shown above"
	case "psalm":
		return "Fix Psalm errors shown above"
	case "test":
		return "Fix failing tests shown above"
	case "rector":
		if fixEnabled {
			return ""
		}
		return "core php rector --fix"
	case "infection":
		return "Improve test coverage for mutated code"
	}
	return ""
}

func runQACheck(ctx context.Context, dir string, checkName string, fix bool) phppkg.QACheckResult {
	start := time.Now()
	result := phppkg.QACheckResult{Name: checkName, Passed: true}

	// Capture output to prevent noise in QA pipeline
	var buf bytes.Buffer

	switch checkName {
	case "audit":
		auditResults, _ := phppkg.RunAudit(ctx, phppkg.AuditOptions{Dir: dir, Output: io.Discard})
		var issues []string
		for _, r := range auditResults {
			if r.Vulnerabilities > 0 {
				issues = append(issues, fmt.Sprintf("%s: %d vulnerabilities", r.Tool, r.Vulnerabilities))
				result.Passed = false
			} else if r.Error != nil {
				issues = append(issues, fmt.Sprintf("%s: %v", r.Tool, r.Error))
				result.Passed = false
			}
		}
		if len(issues) > 0 {
			result.Output = strings.Join(issues, ", ")
		}

	case "fmt":
		err := phppkg.Format(ctx, phppkg.FormatOptions{Dir: dir, Fix: fix, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = "Code style issues found"
		}

	case "analyse":
		err := phppkg.Analyse(ctx, phppkg.AnalyseOptions{Dir: dir, Output: &buf})
		result.Passed = err == nil
		if err != nil {
			result.Output = "Static analysis errors"
		}

	case "psalm":
		err := phppkg.RunPsalm(ctx, phppkg.PsalmOptions{Dir: dir, Fix: fix, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = "Type errors found"
		}

	case "test":
		err := phppkg.RunTests(ctx, phppkg.TestOptions{Dir: dir, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = "Test failures"
		}

	case "rector":
		err := phppkg.RunRector(ctx, phppkg.RectorOptions{Dir: dir, Fix: fix, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = "Code improvements available"
		}

	case "infection":
		err := phppkg.RunInfection(ctx, phppkg.InfectionOptions{Dir: dir, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = "Mutation score below threshold"
		}
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result
}

var (
	rectorFix        bool
	rectorDiff       bool
	rectorClearCache bool
)

func addPHPRectorCommand(parent *cobra.Command) {
	rectorCmd := &cobra.Command{
		Use:   "rector",
		Short: "Automated code refactoring",
		Long: "Run Rector for automated code improvements and PHP upgrades.\n\n" +
			"Rector can automatically upgrade PHP syntax, improve code quality,\n" +
			"and apply framework-specific refactorings.\n\n" +
			"Examples:\n" +
			"  core php rector              # Dry-run (show changes)\n" +
			"  core php rector --fix        # Apply changes\n" +
			"  core php rector --diff       # Show detailed diff",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Check if Rector is available
			if !phppkg.DetectRector(cwd) {
				fmt.Printf("%s Rector not found\n\n", errorStyle.Render("Error:"))
				fmt.Printf("%s composer require --dev rector/rector\n", dimStyle.Render("Install:"))
				fmt.Printf("%s ./vendor/bin/rector init\n", dimStyle.Render("Setup:"))
				return fmt.Errorf("rector not installed")
			}

			action := "Analysing"
			if rectorFix {
				action = "Refactoring"
			}
			fmt.Printf("%s %s code with Rector\n\n", dimStyle.Render("Rector:"), action)

			ctx := context.Background()

			opts := phppkg.RectorOptions{
				Dir:        cwd,
				Fix:        rectorFix,
				Diff:       rectorDiff,
				ClearCache: rectorClearCache,
				Output:     os.Stdout,
			}

			if err := phppkg.RunRector(ctx, opts); err != nil {
				if rectorFix {
					return fmt.Errorf("rector failed: %w", err)
				}
				// Dry-run returns non-zero if changes would be made
				fmt.Printf("\n%s Changes suggested (use --fix to apply)\n", phpQAWarningStyle.Render("Info:"))
				return nil
			}

			if rectorFix {
				fmt.Printf("\n%s Code refactored successfully\n", successStyle.Render("Done:"))
			} else {
				fmt.Printf("\n%s No changes needed\n", successStyle.Render("Done:"))
			}
			return nil
		},
	}

	rectorCmd.Flags().BoolVar(&rectorFix, "fix", false, "Apply changes (default is dry-run)")
	rectorCmd.Flags().BoolVar(&rectorDiff, "diff", false, "Show detailed diff of changes")
	rectorCmd.Flags().BoolVar(&rectorClearCache, "clear-cache", false, "Clear Rector cache before running")

	parent.AddCommand(rectorCmd)
}

var (
	infectionMinMSI        int
	infectionMinCoveredMSI int
	infectionThreads       int
	infectionFilter        string
	infectionOnlyCovered   bool
)

func addPHPInfectionCommand(parent *cobra.Command) {
	infectionCmd := &cobra.Command{
		Use:   "infection",
		Short: "Mutation testing for test quality",
		Long: "Run Infection mutation testing to measure test suite quality.\n\n" +
			"Mutation testing modifies your code and checks if tests catch\n" +
			"the changes. High mutation score = high quality tests.\n\n" +
			"Warning: This can be slow on large codebases.\n\n" +
			"Examples:\n" +
			"  core php infection                    # Run mutation testing\n" +
			"  core php infection --min-msi=70      # Require 70% mutation score\n" +
			"  core php infection --filter=User     # Only test User* files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf("not a PHP project (missing composer.json)")
			}

			// Check if Infection is available
			if !phppkg.DetectInfection(cwd) {
				fmt.Printf("%s Infection not found\n\n", errorStyle.Render("Error:"))
				fmt.Printf("%s composer require --dev infection/infection\n", dimStyle.Render("Install:"))
				return fmt.Errorf("infection not installed")
			}

			fmt.Printf("%s Running mutation testing\n", dimStyle.Render("Infection:"))
			fmt.Printf("%s This may take a while...\n\n", dimStyle.Render("Note:"))

			ctx := context.Background()

			opts := phppkg.InfectionOptions{
				Dir:           cwd,
				MinMSI:        infectionMinMSI,
				MinCoveredMSI: infectionMinCoveredMSI,
				Threads:       infectionThreads,
				Filter:        infectionFilter,
				OnlyCovered:   infectionOnlyCovered,
				Output:        os.Stdout,
			}

			if err := phppkg.RunInfection(ctx, opts); err != nil {
				return fmt.Errorf("mutation testing failed: %w", err)
			}

			fmt.Printf("\n%s Mutation testing complete\n", successStyle.Render("Done:"))
			return nil
		},
	}

	infectionCmd.Flags().IntVar(&infectionMinMSI, "min-msi", 0, "Minimum mutation score indicator (0-100, default: 50)")
	infectionCmd.Flags().IntVar(&infectionMinCoveredMSI, "min-covered-msi", 0, "Minimum covered mutation score (0-100, default: 70)")
	infectionCmd.Flags().IntVar(&infectionThreads, "threads", 0, "Number of parallel threads (default: 4)")
	infectionCmd.Flags().StringVar(&infectionFilter, "filter", "", "Filter files by pattern")
	infectionCmd.Flags().BoolVar(&infectionOnlyCovered, "only-covered", false, "Only mutate covered code")

	parent.AddCommand(infectionCmd)
}

func getSeverityStyle(severity string) lipgloss.Style {
	switch strings.ToLower(severity) {
	case "critical":
		return phpSecurityCriticalStyle
	case "high":
		return phpSecurityHighStyle
	case "medium":
		return phpSecurityMediumStyle
	case "low":
		return phpSecurityLowStyle
	default:
		return dimStyle
	}
}
