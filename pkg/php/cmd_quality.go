package php

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.php.test.short"),
		Long:  i18n.T("cmd.php.test.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.ProgressSubject("run", "tests"))

			ctx := context.Background()

			opts := TestOptions{
				Dir:      cwd,
				Filter:   testFilter,
				Parallel: testParallel,
				Coverage: testCoverage,
				Output:   os.Stdout,
			}

			if testGroup != "" {
				opts.Groups = []string{testGroup}
			}

			if err := RunTests(ctx, opts); err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.run", "tests"), err)
			}

			return nil
		},
	}

	testCmd.Flags().BoolVar(&testParallel, "parallel", false, i18n.T("cmd.php.test.flag.parallel"))
	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, i18n.T("cmd.php.test.flag.coverage"))
	testCmd.Flags().StringVar(&testFilter, "filter", "", i18n.T("cmd.php.test.flag.filter"))
	testCmd.Flags().StringVar(&testGroup, "group", "", i18n.T("cmd.php.test.flag.group"))

	parent.AddCommand(testCmd)
}

var (
	fmtFix  bool
	fmtDiff bool
)

func addPHPFmtCommand(parent *cobra.Command) {
	fmtCmd := &cobra.Command{
		Use:   "fmt [paths...]",
		Short: i18n.T("cmd.php.fmt.short"),
		Long:  i18n.T("cmd.php.fmt.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			// Detect formatter
			formatter, found := DetectFormatter(cwd)
			if !found {
				return errors.New(i18n.T("cmd.php.fmt.no_formatter"))
			}

			var msg string
			if fmtFix {
				msg = i18n.T("cmd.php.fmt.formatting", map[string]interface{}{"Formatter": formatter})
			} else {
				msg = i18n.ProgressSubject("check", "code style")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), msg)

			ctx := context.Background()

			opts := FormatOptions{
				Dir:    cwd,
				Fix:    fmtFix,
				Diff:   fmtDiff,
				Output: os.Stdout,
			}

			// Get any additional paths from args
			if len(args) > 0 {
				opts.Paths = args
			}

			if err := Format(ctx, opts); err != nil {
				if fmtFix {
					return cli.Err("%s: %w", i18n.T("cmd.php.error.fmt_failed"), err)
				}
				return cli.Err("%s: %w", i18n.T("cmd.php.error.fmt_issues"), err)
			}

			if fmtFix {
				cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("common.success.completed", map[string]any{"Action": "Code formatted"}))
			} else {
				cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("cmd.php.fmt.no_issues"))
			}

			return nil
		},
	}

	fmtCmd.Flags().BoolVar(&fmtFix, "fix", false, i18n.T("cmd.php.fmt.flag.fix"))
	fmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, i18n.T("common.flag.diff"))

	parent.AddCommand(fmtCmd)
}

var (
	stanLevel  int
	stanMemory string
)

func addPHPStanCommand(parent *cobra.Command) {
	stanCmd := &cobra.Command{
		Use:   "stan [paths...]",
		Short: i18n.T("cmd.php.analyse.short"),
		Long:  i18n.T("cmd.php.analyse.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			// Detect analyser
			_, found := DetectAnalyser(cwd)
			if !found {
				return errors.New(i18n.T("cmd.php.analyse.no_analyser"))
			}

			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.ProgressSubject("run", "static analysis"))

			ctx := context.Background()

			opts := AnalyseOptions{
				Dir:    cwd,
				Level:  stanLevel,
				Memory: stanMemory,
				Output: os.Stdout,
			}

			// Get any additional paths from args
			if len(args) > 0 {
				opts.Paths = args
			}

			if err := Analyse(ctx, opts); err != nil {
				return cli.Err("%s: %w", i18n.T("cmd.php.error.analysis_issues"), err)
			}

			cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("common.result.no_issues"))
			return nil
		},
	}

	stanCmd.Flags().IntVar(&stanLevel, "level", 0, i18n.T("cmd.php.analyse.flag.level"))
	stanCmd.Flags().StringVar(&stanMemory, "memory", "", i18n.T("cmd.php.analyse.flag.memory"))

	parent.AddCommand(stanCmd)
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
		Short: i18n.T("cmd.php.psalm.short"),
		Long:  i18n.T("cmd.php.psalm.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			// Check if Psalm is available
			_, found := DetectPsalm(cwd)
			if !found {
				cli.Print("%s %s\n\n", errorStyle.Render(i18n.Label("error")), i18n.T("cmd.php.psalm.not_found"))
				cli.Print("%s %s\n", dimStyle.Render(i18n.Label("install")), i18n.T("cmd.php.psalm.install"))
				cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.setup")), i18n.T("cmd.php.psalm.setup"))
				return errors.New(i18n.T("cmd.php.error.psalm_not_installed"))
			}

			var msg string
			if psalmFix {
				msg = i18n.T("cmd.php.psalm.analysing_fixing")
			} else {
				msg = i18n.T("cmd.php.psalm.analysing")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.psalm")), msg)

			ctx := context.Background()

			opts := PsalmOptions{
				Dir:      cwd,
				Level:    psalmLevel,
				Fix:      psalmFix,
				Baseline: psalmBaseline,
				ShowInfo: psalmShowInfo,
				Output:   os.Stdout,
			}

			if err := RunPsalm(ctx, opts); err != nil {
				return cli.Err("%s: %w", i18n.T("cmd.php.error.psalm_issues"), err)
			}

			cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("common.result.no_issues"))
			return nil
		},
	}

	psalmCmd.Flags().IntVar(&psalmLevel, "level", 0, i18n.T("cmd.php.psalm.flag.level"))
	psalmCmd.Flags().BoolVar(&psalmFix, "fix", false, i18n.T("common.flag.fix"))
	psalmCmd.Flags().BoolVar(&psalmBaseline, "baseline", false, i18n.T("cmd.php.psalm.flag.baseline"))
	psalmCmd.Flags().BoolVar(&psalmShowInfo, "show-info", false, i18n.T("cmd.php.psalm.flag.show_info"))

	parent.AddCommand(psalmCmd)
}

var (
	auditJSONOutput bool
	auditFix        bool
)

func addPHPAuditCommand(parent *cobra.Command) {
	auditCmd := &cobra.Command{
		Use:   "audit",
		Short: i18n.T("cmd.php.audit.short"),
		Long:  i18n.T("cmd.php.audit.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.audit")), i18n.T("cmd.php.audit.scanning"))

			ctx := context.Background()

			results, err := RunAudit(ctx, AuditOptions{
				Dir:    cwd,
				JSON:   auditJSONOutput,
				Fix:    auditFix,
				Output: os.Stdout,
			})
			if err != nil {
				return cli.Err("%s: %w", i18n.T("cmd.php.error.audit_failed"), err)
			}

			// Print results
			totalVulns := 0
			hasErrors := false

			for _, result := range results {
				icon := successStyle.Render("✓")
				status := successStyle.Render(i18n.T("cmd.php.audit.secure"))

				if result.Error != nil {
					icon = errorStyle.Render("✗")
					status = errorStyle.Render(i18n.T("cmd.php.audit.error"))
					hasErrors = true
				} else if result.Vulnerabilities > 0 {
					icon = errorStyle.Render("✗")
					status = errorStyle.Render(i18n.T("cmd.php.audit.vulnerabilities", map[string]interface{}{"Count": result.Vulnerabilities}))
					totalVulns += result.Vulnerabilities
				}

				cli.Print("  %s %s %s\n", icon, dimStyle.Render(result.Tool+":"), status)

				// Show advisories
				for _, adv := range result.Advisories {
					severity := adv.Severity
					if severity == "" {
						severity = "unknown"
					}
					sevStyle := getSeverityStyle(severity)
					cli.Print("      %s %s\n", sevStyle.Render("["+severity+"]"), adv.Package)
					if adv.Title != "" {
						cli.Print("               %s\n", dimStyle.Render(adv.Title))
					}
				}
			}

			cli.Line("")

			if totalVulns > 0 {
				cli.Print("%s %s\n", errorStyle.Render(i18n.Label("warning")), i18n.T("cmd.php.audit.found_vulns", map[string]interface{}{"Count": totalVulns}))
				cli.Print("%s %s\n", dimStyle.Render(i18n.Label("fix")), i18n.T("common.hint.fix_deps"))
				return errors.New(i18n.T("cmd.php.error.vulns_found"))
			}

			if hasErrors {
				return errors.New(i18n.T("cmd.php.audit.completed_errors"))
			}

			cli.Print("%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("cmd.php.audit.all_secure"))
			return nil
		},
	}

	auditCmd.Flags().BoolVar(&auditJSONOutput, "json", false, i18n.T("common.flag.json"))
	auditCmd.Flags().BoolVar(&auditFix, "fix", false, i18n.T("cmd.php.audit.flag.fix"))

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
		Short: i18n.T("cmd.php.security.short"),
		Long:  i18n.T("cmd.php.security.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.security")), i18n.ProgressSubject("run", "security checks"))

			ctx := context.Background()

			result, err := RunSecurityChecks(ctx, SecurityOptions{
				Dir:      cwd,
				Severity: securitySeverity,
				JSON:     securityJSONOutput,
				SARIF:    securitySarif,
				URL:      securityURL,
				Output:   os.Stdout,
			})
			if err != nil {
				return cli.Err("%s: %w", i18n.T("cmd.php.error.security_failed"), err)
			}

			// Print results by category
			currentCategory := ""
			for _, check := range result.Checks {
				category := strings.Split(check.ID, "_")[0]
				if category != currentCategory {
					if currentCategory != "" {
						cli.Line("")
					}
					currentCategory = category
					cli.Print("  %s\n", dimStyle.Render(strings.ToUpper(category)+i18n.T("cmd.php.security.checks_suffix")))
				}

				icon := successStyle.Render("✓")
				if !check.Passed {
					icon = getSeverityStyle(check.Severity).Render("✗")
				}

				cli.Print("    %s %s\n", icon, check.Name)
				if !check.Passed && check.Message != "" {
					cli.Print("        %s\n", dimStyle.Render(check.Message))
					if check.Fix != "" {
						cli.Print("        %s %s\n", dimStyle.Render(i18n.Label("fix")), check.Fix)
					}
				}
			}

			cli.Line("")

			// Print summary
			cli.Print("%s %s\n", dimStyle.Render(i18n.Label("summary")), i18n.T("cmd.php.security.summary"))
			cli.Print("  %s %d/%d\n", dimStyle.Render(i18n.T("cmd.php.security.passed")), result.Summary.Passed, result.Summary.Total)

			if result.Summary.Critical > 0 {
				cli.Print("  %s %d\n", phpSecurityCriticalStyle.Render(i18n.T("cmd.php.security.critical")), result.Summary.Critical)
			}
			if result.Summary.High > 0 {
				cli.Print("  %s %d\n", phpSecurityHighStyle.Render(i18n.T("cmd.php.security.high")), result.Summary.High)
			}
			if result.Summary.Medium > 0 {
				cli.Print("  %s %d\n", phpSecurityMediumStyle.Render(i18n.T("cmd.php.security.medium")), result.Summary.Medium)
			}
			if result.Summary.Low > 0 {
				cli.Print("  %s %d\n", phpSecurityLowStyle.Render(i18n.T("cmd.php.security.low")), result.Summary.Low)
			}

			if result.Summary.Critical > 0 || result.Summary.High > 0 {
				return errors.New(i18n.T("cmd.php.error.critical_high_issues"))
			}

			return nil
		},
	}

	securityCmd.Flags().StringVar(&securitySeverity, "severity", "", i18n.T("cmd.php.security.flag.severity"))
	securityCmd.Flags().BoolVar(&securityJSONOutput, "json", false, i18n.T("common.flag.json"))
	securityCmd.Flags().BoolVar(&securitySarif, "sarif", false, i18n.T("cmd.php.security.flag.sarif"))
	securityCmd.Flags().StringVar(&securityURL, "url", "", i18n.T("cmd.php.security.flag.url"))

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
		Short: i18n.T("cmd.php.qa.short"),
		Long:  i18n.T("cmd.php.qa.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			// Determine stages
			opts := QAOptions{
				Dir:   cwd,
				Quick: qaQuick,
				Full:  qaFull,
				Fix:   qaFix,
			}
			stages := GetQAStages(opts)

			// Print header
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("qa")), i18n.ProgressSubject("run", "QA pipeline"))

			ctx := context.Background()

			// Create QA runner using pkg/process
			runner, err := NewQARunner(cwd, qaFix)
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.create", "QA runner"), err)
			}

			// Run all checks with dependency ordering
			result, err := runner.Run(ctx, stages)
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.run", "QA checks"), err)
			}

			// Display results by stage
			currentStage := ""
			for _, checkResult := range result.Results {
				// Determine stage for this check
				stage := getCheckStage(checkResult.Name, stages, cwd)
				if stage != currentStage {
					if currentStage != "" {
						cli.Line("")
					}
					currentStage = stage
					cli.Print("%s\n", phpQAStageStyle.Render("── "+strings.ToUpper(stage)+" ──"))
				}

				icon := phpQAPassedStyle.Render("✓")
				status := phpQAPassedStyle.Render(i18n.T("i18n.done.pass"))
				if checkResult.Skipped {
					icon = dimStyle.Render("-")
					status = dimStyle.Render(i18n.T("i18n.done.skip"))
				} else if !checkResult.Passed {
					icon = phpQAFailedStyle.Render("✗")
					status = phpQAFailedStyle.Render(i18n.T("i18n.done.fail"))
				}

				cli.Print("  %s %s %s %s\n", icon, checkResult.Name, status, dimStyle.Render(checkResult.Duration))
			}
			cli.Line("")

			// Print summary
			if result.Passed {
				cli.Print("%s %s\n", phpQAPassedStyle.Render("QA PASSED:"), i18n.T("i18n.count.check", result.PassedCount)+" "+i18n.T("i18n.done.pass"))
				cli.Print("%s %s\n", dimStyle.Render(i18n.T("i18n.label.duration")), result.Duration)
				return nil
			}

			cli.Print("%s %s\n\n", phpQAFailedStyle.Render("QA FAILED:"), i18n.T("i18n.count.check", result.PassedCount)+"/"+cli.Sprint(len(result.Results))+" "+i18n.T("i18n.done.pass"))

			// Show what needs fixing
			cli.Print("%s\n", dimStyle.Render(i18n.T("i18n.label.fix")))
			for _, checkResult := range result.Results {
				if checkResult.Passed || checkResult.Skipped {
					continue
				}
				fixCmd := getQAFixCommand(checkResult.Name, qaFix)
				issue := checkResult.GetIssueMessage()
				if issue == "" {
					issue = "issues found"
				}
				cli.Print("  %s %s\n", phpQAFailedStyle.Render("*"), checkResult.Name+": "+issue)
				if fixCmd != "" {
					cli.Print("    %s %s\n", dimStyle.Render("->"), fixCmd)
				}
			}

			return cli.Err("%s", i18n.T("i18n.fail.run", "QA pipeline"))
		},
	}

	qaCmd.Flags().BoolVar(&qaQuick, "quick", false, "Run quick checks only (audit, fmt, stan)")
	qaCmd.Flags().BoolVar(&qaFull, "full", false, "Run all stages including slow checks")
	qaCmd.Flags().BoolVar(&qaFix, "fix", false, "Auto-fix issues where possible")

	parent.AddCommand(qaCmd)
}

// getCheckStage determines which stage a check belongs to.
func getCheckStage(checkName string, stages []QAStage, dir string) string {
	for _, stage := range stages {
		checks := GetQAChecks(dir, stage)
		for _, c := range checks {
			if c == checkName {
				return string(stage)
			}
		}
	}
	return "unknown"
}

func getQAFixCommand(checkName string, fixEnabled bool) string {
	switch checkName {
	case "audit":
		return i18n.T("i18n.progress.update", "dependencies")
	case "fmt":
		if fixEnabled {
			return ""
		}
		return "core php fmt --fix"
	case "stan":
		return i18n.T("i18n.progress.fix", "PHPStan errors")
	case "psalm":
		return i18n.T("i18n.progress.fix", "Psalm errors")
	case "test":
		return i18n.T("i18n.progress.fix", i18n.T("i18n.done.fail")+" tests")
	case "rector":
		if fixEnabled {
			return ""
		}
		return "core php rector --fix"
	case "infection":
		return i18n.T("i18n.progress.improve", "test coverage")
	}
	return ""
}

var (
	rectorFix        bool
	rectorDiff       bool
	rectorClearCache bool
)

func addPHPRectorCommand(parent *cobra.Command) {
	rectorCmd := &cobra.Command{
		Use:   "rector",
		Short: i18n.T("cmd.php.rector.short"),
		Long:  i18n.T("cmd.php.rector.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			// Check if Rector is available
			if !DetectRector(cwd) {
				cli.Print("%s %s\n\n", errorStyle.Render(i18n.Label("error")), i18n.T("cmd.php.rector.not_found"))
				cli.Print("%s %s\n", dimStyle.Render(i18n.Label("install")), i18n.T("cmd.php.rector.install"))
				cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.setup")), i18n.T("cmd.php.rector.setup"))
				return errors.New(i18n.T("cmd.php.error.rector_not_installed"))
			}

			var msg string
			if rectorFix {
				msg = i18n.T("cmd.php.rector.refactoring")
			} else {
				msg = i18n.T("cmd.php.rector.analysing")
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.rector")), msg)

			ctx := context.Background()

			opts := RectorOptions{
				Dir:        cwd,
				Fix:        rectorFix,
				Diff:       rectorDiff,
				ClearCache: rectorClearCache,
				Output:     os.Stdout,
			}

			if err := RunRector(ctx, opts); err != nil {
				if rectorFix {
					return cli.Err("%s: %w", i18n.T("cmd.php.error.rector_failed"), err)
				}
				// Dry-run returns non-zero if changes would be made
				cli.Print("\n%s %s\n", phpQAWarningStyle.Render(i18n.T("cmd.php.label.info")), i18n.T("cmd.php.rector.changes_suggested"))
				return nil
			}

			if rectorFix {
				cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("common.success.completed", map[string]any{"Action": "Code refactored"}))
			} else {
				cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("cmd.php.rector.no_changes"))
			}
			return nil
		},
	}

	rectorCmd.Flags().BoolVar(&rectorFix, "fix", false, i18n.T("cmd.php.rector.flag.fix"))
	rectorCmd.Flags().BoolVar(&rectorDiff, "diff", false, i18n.T("cmd.php.rector.flag.diff"))
	rectorCmd.Flags().BoolVar(&rectorClearCache, "clear-cache", false, i18n.T("cmd.php.rector.flag.clear_cache"))

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
		Short: i18n.T("cmd.php.infection.short"),
		Long:  i18n.T("cmd.php.infection.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return cli.Err("%s: %w", i18n.T("i18n.fail.get", "working directory"), err)
			}

			if !IsPHPProject(cwd) {
				return errors.New(i18n.T("cmd.php.error.not_php"))
			}

			// Check if Infection is available
			if !DetectInfection(cwd) {
				cli.Print("%s %s\n\n", errorStyle.Render(i18n.Label("error")), i18n.T("cmd.php.infection.not_found"))
				cli.Print("%s %s\n", dimStyle.Render(i18n.Label("install")), i18n.T("cmd.php.infection.install"))
				return errors.New(i18n.T("cmd.php.error.infection_not_installed"))
			}

			cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.infection")), i18n.ProgressSubject("run", "mutation testing"))
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.info")), i18n.T("cmd.php.infection.note"))

			ctx := context.Background()

			opts := InfectionOptions{
				Dir:           cwd,
				MinMSI:        infectionMinMSI,
				MinCoveredMSI: infectionMinCoveredMSI,
				Threads:       infectionThreads,
				Filter:        infectionFilter,
				OnlyCovered:   infectionOnlyCovered,
				Output:        os.Stdout,
			}

			if err := RunInfection(ctx, opts); err != nil {
				return cli.Err("%s: %w", i18n.T("cmd.php.error.infection_failed"), err)
			}

			cli.Print("\n%s %s\n", successStyle.Render(i18n.Label("done")), i18n.T("cmd.php.infection.complete"))
			return nil
		},
	}

	infectionCmd.Flags().IntVar(&infectionMinMSI, "min-msi", 0, i18n.T("cmd.php.infection.flag.min_msi"))
	infectionCmd.Flags().IntVar(&infectionMinCoveredMSI, "min-covered-msi", 0, i18n.T("cmd.php.infection.flag.min_covered_msi"))
	infectionCmd.Flags().IntVar(&infectionThreads, "threads", 0, i18n.T("cmd.php.infection.flag.threads"))
	infectionCmd.Flags().StringVar(&infectionFilter, "filter", "", i18n.T("cmd.php.infection.flag.filter"))
	infectionCmd.Flags().BoolVar(&infectionOnlyCovered, "only-covered", false, i18n.T("cmd.php.infection.flag.only_covered"))

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
