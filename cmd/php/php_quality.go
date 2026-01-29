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
	"github.com/leaanthony/clir"
)

func addPHPTestCommand(parent *clir.Command) {
	var (
		parallel bool
		coverage bool
		filter   string
		group    string
	)

	testCmd := parent.NewSubCommand("test", "Run PHP tests (PHPUnit/Pest)")
	testCmd.LongDescription("Run PHP tests using PHPUnit or Pest.\n\n" +
		"Auto-detects Pest if tests/Pest.php exists, otherwise uses PHPUnit.\n\n" +
		"Examples:\n" +
		"  core php test                    # Run all tests\n" +
		"  core php test --parallel         # Run tests in parallel\n" +
		"  core php test --coverage         # Run with coverage\n" +
		"  core php test --filter UserTest  # Filter by test name")

	testCmd.BoolFlag("parallel", "Run tests in parallel", &parallel)
	testCmd.BoolFlag("coverage", "Generate code coverage", &coverage)
	testCmd.StringFlag("filter", "Filter tests by name pattern", &filter)
	testCmd.StringFlag("group", "Run only tests in specified group", &group)

	testCmd.Action(func() error {
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
			Filter:   filter,
			Parallel: parallel,
			Coverage: coverage,
			Output:   os.Stdout,
		}

		if group != "" {
			opts.Groups = []string{group}
		}

		if err := phppkg.RunTests(ctx, opts); err != nil {
			return fmt.Errorf("tests failed: %w", err)
		}

		return nil
	})
}

func addPHPFmtCommand(parent *clir.Command) {
	var (
		fix  bool
		diff bool
	)

	fmtCmd := parent.NewSubCommand("fmt", "Format PHP code with Laravel Pint")
	fmtCmd.LongDescription("Format PHP code using Laravel Pint.\n\n" +
		"Examples:\n" +
		"  core php fmt           # Check formatting (dry-run)\n" +
		"  core php fmt --fix     # Auto-fix formatting issues\n" +
		"  core php fmt --diff    # Show diff of changes")

	fmtCmd.BoolFlag("fix", "Auto-fix formatting issues", &fix)
	fmtCmd.BoolFlag("diff", "Show diff of changes", &diff)

	fmtCmd.Action(func() error {
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
		if fix {
			action = "Formatting"
		}
		fmt.Printf("%s %s code with %s\n\n", dimStyle.Render("PHP:"), action, formatter)

		ctx := context.Background()

		opts := phppkg.FormatOptions{
			Dir:    cwd,
			Fix:    fix,
			Diff:   diff,
			Output: os.Stdout,
		}

		// Get any additional paths from args
		if args := fmtCmd.OtherArgs(); len(args) > 0 {
			opts.Paths = args
		}

		if err := phppkg.Format(ctx, opts); err != nil {
			if fix {
				return fmt.Errorf("formatting failed: %w", err)
			}
			return fmt.Errorf("formatting issues found: %w", err)
		}

		if fix {
			fmt.Printf("\n%s Code formatted successfully\n", successStyle.Render("Done:"))
		} else {
			fmt.Printf("\n%s No formatting issues found\n", successStyle.Render("Done:"))
		}

		return nil
	})
}

func addPHPAnalyseCommand(parent *clir.Command) {
	var (
		level  int
		memory string
	)

	analyseCmd := parent.NewSubCommand("analyse", "Run PHPStan static analysis")
	analyseCmd.LongDescription("Run PHPStan or Larastan static analysis.\n\n" +
		"Auto-detects Larastan if installed, otherwise uses PHPStan.\n\n" +
		"Examples:\n" +
		"  core php analyse              # Run analysis\n" +
		"  core php analyse --level 9    # Run at max strictness\n" +
		"  core php analyse --memory 2G  # Increase memory limit")

	analyseCmd.IntFlag("level", "PHPStan analysis level (0-9)", &level)
	analyseCmd.StringFlag("memory", "Memory limit (e.g., 2G)", &memory)

	analyseCmd.Action(func() error {
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
			Level:  level,
			Memory: memory,
			Output: os.Stdout,
		}

		// Get any additional paths from args
		if args := analyseCmd.OtherArgs(); len(args) > 0 {
			opts.Paths = args
		}

		if err := phppkg.Analyse(ctx, opts); err != nil {
			return fmt.Errorf("analysis found issues: %w", err)
		}

		fmt.Printf("\n%s No issues found\n", successStyle.Render("Done:"))
		return nil
	})
}

// =============================================================================
// New QA Commands
// =============================================================================

func addPHPPsalmCommand(parent *clir.Command) {
	var (
		level    int
		fix      bool
		baseline bool
		showInfo bool
	)

	psalmCmd := parent.NewSubCommand("psalm", "Run Psalm static analysis")
	psalmCmd.LongDescription("Run Psalm deep static analysis with Laravel plugin support.\n\n" +
		"Psalm provides deeper type inference than PHPStan and catches\n" +
		"different classes of bugs. Both should be run for best coverage.\n\n" +
		"Examples:\n" +
		"  core php psalm              # Run analysis\n" +
		"  core php psalm --fix        # Auto-fix issues where possible\n" +
		"  core php psalm --level 3    # Run at specific level (1-8)\n" +
		"  core php psalm --baseline   # Generate baseline file")

	psalmCmd.IntFlag("level", "Error level (1=strictest, 8=most lenient)", &level)
	psalmCmd.BoolFlag("fix", "Auto-fix issues where possible", &fix)
	psalmCmd.BoolFlag("baseline", "Generate/update baseline file", &baseline)
	psalmCmd.BoolFlag("show-info", "Show info-level issues", &showInfo)

	psalmCmd.Action(func() error {
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
		if fix {
			action = "Analysing and fixing"
		}
		fmt.Printf("%s %s code with Psalm\n\n", dimStyle.Render("Psalm:"), action)

		ctx := context.Background()

		opts := phppkg.PsalmOptions{
			Dir:      cwd,
			Level:    level,
			Fix:      fix,
			Baseline: baseline,
			ShowInfo: showInfo,
			Output:   os.Stdout,
		}

		if err := phppkg.RunPsalm(ctx, opts); err != nil {
			return fmt.Errorf("psalm found issues: %w", err)
		}

		fmt.Printf("\n%s No issues found\n", successStyle.Render("Done:"))
		return nil
	})
}

func addPHPAuditCommand(parent *clir.Command) {
	var (
		jsonOutput bool
		fix        bool
	)

	auditCmd := parent.NewSubCommand("audit", "Security audit for dependencies")
	auditCmd.LongDescription("Check PHP and JavaScript dependencies for known vulnerabilities.\n\n" +
		"Runs composer audit and npm audit (if package.json exists).\n\n" +
		"Examples:\n" +
		"  core php audit           # Check all dependencies\n" +
		"  core php audit --json    # Output as JSON\n" +
		"  core php audit --fix     # Auto-fix where possible (npm only)")

	auditCmd.BoolFlag("json", "Output in JSON format", &jsonOutput)
	auditCmd.BoolFlag("fix", "Auto-fix vulnerabilities (npm only)", &fix)

	auditCmd.Action(func() error {
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
			JSON:   jsonOutput,
			Fix:    fix,
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
	})
}

func addPHPSecurityCommand(parent *clir.Command) {
	var (
		severity   string
		jsonOutput bool
		sarif      bool
		url        string
	)

	securityCmd := parent.NewSubCommand("security", "Security vulnerability scanning")
	securityCmd.LongDescription("Scan for security vulnerabilities in configuration and code.\n\n" +
		"Checks environment config, file permissions, code patterns,\n" +
		"and runs security-focused static analysis.\n\n" +
		"Examples:\n" +
		"  core php security                    # Run all checks\n" +
		"  core php security --severity=high   # Only high+ severity\n" +
		"  core php security --json            # JSON output")

	securityCmd.StringFlag("severity", "Minimum severity (critical, high, medium, low)", &severity)
	securityCmd.BoolFlag("json", "Output in JSON format", &jsonOutput)
	securityCmd.BoolFlag("sarif", "Output in SARIF format (for GitHub Security)", &sarif)
	securityCmd.StringFlag("url", "URL to check HTTP headers (optional)", &url)

	securityCmd.Action(func() error {
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
			Severity: severity,
			JSON:     jsonOutput,
			SARIF:    sarif,
			URL:      url,
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
	})
}

func addPHPQACommand(parent *clir.Command) {
	var (
		quick bool
		full  bool
		fix   bool
	)

	qaCmd := parent.NewSubCommand("qa", "Run full QA pipeline")
	qaCmd.LongDescription("Run the complete quality assurance pipeline.\n\n" +
		"Stages:\n" +
		"  quick:    Security audit, code style, PHPStan\n" +
		"  standard: Psalm, tests\n" +
		"  full:     Rector dry-run, mutation testing (slow)\n\n" +
		"Examples:\n" +
		"  core php qa              # Run quick + standard stages\n" +
		"  core php qa --quick      # Only quick checks\n" +
		"  core php qa --full       # All stages including slow ones\n" +
		"  core php qa --fix        # Auto-fix where possible")

	qaCmd.BoolFlag("quick", "Only run quick checks", &quick)
	qaCmd.BoolFlag("full", "Run all stages including slow checks", &full)
	qaCmd.BoolFlag("fix", "Auto-fix issues where possible", &fix)

	qaCmd.Action(func() error {
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
			Quick: quick,
			Full:  full,
			Fix:   fix,
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
				result := runQACheck(ctx, cwd, checkName, fix)
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
			fixCmd := getQAFixCommand(check.Name, fix)
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
	})
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

func addPHPRectorCommand(parent *clir.Command) {
	var (
		fix        bool
		diff       bool
		clearCache bool
	)

	rectorCmd := parent.NewSubCommand("rector", "Automated code refactoring")
	rectorCmd.LongDescription("Run Rector for automated code improvements and PHP upgrades.\n\n" +
		"Rector can automatically upgrade PHP syntax, improve code quality,\n" +
		"and apply framework-specific refactorings.\n\n" +
		"Examples:\n" +
		"  core php rector              # Dry-run (show changes)\n" +
		"  core php rector --fix        # Apply changes\n" +
		"  core php rector --diff       # Show detailed diff")

	rectorCmd.BoolFlag("fix", "Apply changes (default is dry-run)", &fix)
	rectorCmd.BoolFlag("diff", "Show detailed diff of changes", &diff)
	rectorCmd.BoolFlag("clear-cache", "Clear Rector cache before running", &clearCache)

	rectorCmd.Action(func() error {
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
		if fix {
			action = "Refactoring"
		}
		fmt.Printf("%s %s code with Rector\n\n", dimStyle.Render("Rector:"), action)

		ctx := context.Background()

		opts := phppkg.RectorOptions{
			Dir:        cwd,
			Fix:        fix,
			Diff:       diff,
			ClearCache: clearCache,
			Output:     os.Stdout,
		}

		if err := phppkg.RunRector(ctx, opts); err != nil {
			if fix {
				return fmt.Errorf("rector failed: %w", err)
			}
			// Dry-run returns non-zero if changes would be made
			fmt.Printf("\n%s Changes suggested (use --fix to apply)\n", phpQAWarningStyle.Render("Info:"))
			return nil
		}

		if fix {
			fmt.Printf("\n%s Code refactored successfully\n", successStyle.Render("Done:"))
		} else {
			fmt.Printf("\n%s No changes needed\n", successStyle.Render("Done:"))
		}
		return nil
	})
}

func addPHPInfectionCommand(parent *clir.Command) {
	var (
		minMSI        int
		minCoveredMSI int
		threads       int
		filter        string
		onlyCovered   bool
	)

	infectionCmd := parent.NewSubCommand("infection", "Mutation testing for test quality")
	infectionCmd.LongDescription("Run Infection mutation testing to measure test suite quality.\n\n" +
		"Mutation testing modifies your code and checks if tests catch\n" +
		"the changes. High mutation score = high quality tests.\n\n" +
		"Warning: This can be slow on large codebases.\n\n" +
		"Examples:\n" +
		"  core php infection                    # Run mutation testing\n" +
		"  core php infection --min-msi=70      # Require 70% mutation score\n" +
		"  core php infection --filter=User     # Only test User* files")

	infectionCmd.IntFlag("min-msi", "Minimum mutation score indicator (0-100, default: 50)", &minMSI)
	infectionCmd.IntFlag("min-covered-msi", "Minimum covered mutation score (0-100, default: 70)", &minCoveredMSI)
	infectionCmd.IntFlag("threads", "Number of parallel threads (default: 4)", &threads)
	infectionCmd.StringFlag("filter", "Filter files by pattern", &filter)
	infectionCmd.BoolFlag("only-covered", "Only mutate covered code", &onlyCovered)

	infectionCmd.Action(func() error {
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
			MinMSI:        minMSI,
			MinCoveredMSI: minCoveredMSI,
			Threads:       threads,
			Filter:        filter,
			OnlyCovered:   onlyCovered,
			Output:        os.Stdout,
		}

		if err := phppkg.RunInfection(ctx, opts); err != nil {
			return fmt.Errorf("mutation testing failed: %w", err)
		}

		fmt.Printf("\n%s Mutation testing complete\n", successStyle.Render("Done:"))
		return nil
	})
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
