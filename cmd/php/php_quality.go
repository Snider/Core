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
	"github.com/host-uk/core/pkg/i18n"
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
		Short: i18n.T("cmd.php.test.short"),
		Long:  i18n.T("cmd.php.test.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			// Detect test runner
			runner := phppkg.DetectTestRunner(cwd)
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.test.running", map[string]interface{}{"Runner": runner}))

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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.tests_failed"), err)
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			// Detect formatter
			formatter, found := phppkg.DetectFormatter(cwd)
			if !found {
				return fmt.Errorf(i18n.T("cmd.php.fmt.no_formatter"))
			}

			var msg string
			if fmtFix {
				msg = i18n.T("cmd.php.fmt.formatting", map[string]interface{}{"Formatter": formatter})
			} else {
				msg = i18n.T("cmd.php.fmt.checking", map[string]interface{}{"Formatter": formatter})
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), msg)

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
					return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.fmt_failed"), err)
				}
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.fmt_issues"), err)
			}

			if fmtFix {
				fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.fmt.success"))
			} else {
				fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.fmt.no_issues"))
			}

			return nil
		},
	}

	fmtCmd.Flags().BoolVar(&fmtFix, "fix", false, i18n.T("cmd.php.fmt.flag.fix"))
	fmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, i18n.T("cmd.php.fmt.flag.diff"))

	parent.AddCommand(fmtCmd)
}

var (
	analyseLevel  int
	analyseMemory string
)

func addPHPAnalyseCommand(parent *cobra.Command) {
	analyseCmd := &cobra.Command{
		Use:   "analyse [paths...]",
		Short: i18n.T("cmd.php.analyse.short"),
		Long:  i18n.T("cmd.php.analyse.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			// Detect analyser
			analyser, found := phppkg.DetectAnalyser(cwd)
			if !found {
				return fmt.Errorf(i18n.T("cmd.php.analyse.no_analyser"))
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.analyse.running", map[string]interface{}{"Analyser": analyser}))

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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.analysis_issues"), err)
			}

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.analyse.no_issues"))
			return nil
		},
	}

	analyseCmd.Flags().IntVar(&analyseLevel, "level", 0, i18n.T("cmd.php.analyse.flag.level"))
	analyseCmd.Flags().StringVar(&analyseMemory, "memory", "", i18n.T("cmd.php.analyse.flag.memory"))

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
		Short: i18n.T("cmd.php.psalm.short"),
		Long:  i18n.T("cmd.php.psalm.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			// Check if Psalm is available
			_, found := phppkg.DetectPsalm(cwd)
			if !found {
				fmt.Printf("%s %s\n\n", errorStyle.Render(i18n.T("cmd.php.label.error")), i18n.T("cmd.php.psalm.not_found"))
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.install")), i18n.T("cmd.php.psalm.install"))
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.setup")), i18n.T("cmd.php.psalm.setup"))
				return fmt.Errorf(i18n.T("cmd.php.error.psalm_not_installed"))
			}

			var msg string
			if psalmFix {
				msg = i18n.T("cmd.php.psalm.analysing_fixing")
			} else {
				msg = i18n.T("cmd.php.psalm.analysing")
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.psalm")), msg)

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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.psalm_issues"), err)
			}

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.psalm.no_issues"))
			return nil
		},
	}

	psalmCmd.Flags().IntVar(&psalmLevel, "level", 0, i18n.T("cmd.php.psalm.flag.level"))
	psalmCmd.Flags().BoolVar(&psalmFix, "fix", false, i18n.T("cmd.php.psalm.flag.fix"))
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.audit")), i18n.T("cmd.php.audit.scanning"))

			ctx := context.Background()

			results, err := phppkg.RunAudit(ctx, phppkg.AuditOptions{
				Dir:    cwd,
				JSON:   auditJSONOutput,
				Fix:    auditFix,
				Output: os.Stdout,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.audit_failed"), err)
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
				fmt.Printf("%s %s\n", errorStyle.Render(i18n.T("cmd.php.label.warning")), i18n.T("cmd.php.audit.found_vulns", map[string]interface{}{"Count": totalVulns}))
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.fix")), i18n.T("cmd.php.audit.fix_hint"))
				return fmt.Errorf(i18n.T("cmd.php.error.vulns_found"))
			}

			if hasErrors {
				return fmt.Errorf(i18n.T("cmd.php.audit.completed_errors"))
			}

			fmt.Printf("%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.audit.all_secure"))
			return nil
		},
	}

	auditCmd.Flags().BoolVar(&auditJSONOutput, "json", false, i18n.T("cmd.php.audit.flag.json"))
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.security")), i18n.T("cmd.php.security.running"))

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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.security_failed"), err)
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
					fmt.Printf("  %s\n", dimStyle.Render(strings.ToUpper(category)+i18n.T("cmd.php.security.checks_suffix")))
				}

				icon := successStyle.Render("✓")
				if !check.Passed {
					icon = getSeverityStyle(check.Severity).Render("✗")
				}

				fmt.Printf("    %s %s\n", icon, check.Name)
				if !check.Passed && check.Message != "" {
					fmt.Printf("        %s\n", dimStyle.Render(check.Message))
					if check.Fix != "" {
						fmt.Printf("        %s %s\n", dimStyle.Render(i18n.T("cmd.php.security.fix_label")), check.Fix)
					}
				}
			}

			fmt.Println()

			// Print summary
			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.summary")), i18n.T("cmd.php.security.summary"))
			fmt.Printf("  %s %d/%d\n", dimStyle.Render(i18n.T("cmd.php.security.passed")), result.Summary.Passed, result.Summary.Total)

			if result.Summary.Critical > 0 {
				fmt.Printf("  %s %d\n", phpSecurityCriticalStyle.Render(i18n.T("cmd.php.security.critical")), result.Summary.Critical)
			}
			if result.Summary.High > 0 {
				fmt.Printf("  %s %d\n", phpSecurityHighStyle.Render(i18n.T("cmd.php.security.high")), result.Summary.High)
			}
			if result.Summary.Medium > 0 {
				fmt.Printf("  %s %d\n", phpSecurityMediumStyle.Render(i18n.T("cmd.php.security.medium")), result.Summary.Medium)
			}
			if result.Summary.Low > 0 {
				fmt.Printf("  %s %d\n", phpSecurityLowStyle.Render(i18n.T("cmd.php.security.low")), result.Summary.Low)
			}

			if result.Summary.Critical > 0 || result.Summary.High > 0 {
				return fmt.Errorf(i18n.T("cmd.php.error.critical_high_issues"))
			}

			return nil
		},
	}

	securityCmd.Flags().StringVar(&securitySeverity, "severity", "", i18n.T("cmd.php.security.flag.severity"))
	securityCmd.Flags().BoolVar(&securityJSONOutput, "json", false, i18n.T("cmd.php.security.flag.json"))
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
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
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.qa")), i18n.T("cmd.php.qa.running", map[string]interface{}{"Stages": strings.Join(stageNames, " -> ")}))

			ctx := context.Background()
			var allPassed = true
			var results []phppkg.QACheckResult

			for _, stage := range stages {
				fmt.Printf("%s\n", phpQAStageStyle.Render(i18n.T("cmd.php.qa.stage_prefix")+strings.ToUpper(string(stage))+i18n.T("cmd.php.qa.stage_suffix")))

				checks := phppkg.GetQAChecks(cwd, stage)
				if len(checks) == 0 {
					fmt.Printf("  %s\n\n", dimStyle.Render(i18n.T("cmd.php.qa.no_checks")))
					continue
				}

				for _, checkName := range checks {
					result := runQACheck(ctx, cwd, checkName, qaFix)
					result.Stage = stage
					results = append(results, result)

					icon := phpQAPassedStyle.Render("✓")
					status := phpQAPassedStyle.Render(i18n.T("cmd.php.qa.passed"))
					if !result.Passed {
						icon = phpQAFailedStyle.Render("✗")
						status = phpQAFailedStyle.Render(i18n.T("cmd.php.qa.failed"))
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
				fmt.Printf("%s %s\n", phpQAPassedStyle.Render("QA PASSED:"), i18n.T("cmd.php.qa.all_passed", map[string]interface{}{"Passed": passedCount, "Total": len(results)}))
				return nil
			}

			fmt.Printf("%s %s\n\n", phpQAFailedStyle.Render("QA FAILED:"), i18n.T("cmd.php.qa.some_failed", map[string]interface{}{"Passed": passedCount, "Total": len(results)}))

			// Show what needs fixing
			fmt.Printf("%s\n", dimStyle.Render(i18n.T("cmd.php.qa.to_fix")))
			for _, check := range failedChecks {
				fixCmd := getQAFixCommand(check.Name, qaFix)
				issue := check.Output
				if issue == "" {
					issue = "issues found"
				}
				fmt.Printf("  %s %s\n", phpQAFailedStyle.Render("*"), check.Name+": "+issue)
				if fixCmd != "" {
					fmt.Printf("    %s %s\n", dimStyle.Render("->"), fixCmd)
				}
			}

			return fmt.Errorf(i18n.T("cmd.php.qa.pipeline_failed"))
		},
	}

	qaCmd.Flags().BoolVar(&qaQuick, "quick", false, i18n.T("cmd.php.qa.flag.quick"))
	qaCmd.Flags().BoolVar(&qaFull, "full", false, i18n.T("cmd.php.qa.flag.full"))
	qaCmd.Flags().BoolVar(&qaFix, "fix", false, i18n.T("cmd.php.qa.flag.fix"))

	parent.AddCommand(qaCmd)
}

func getQAFixCommand(checkName string, fixEnabled bool) string {
	switch checkName {
	case "audit":
		return i18n.T("cmd.php.qa.fix_audit")
	case "fmt":
		if fixEnabled {
			return ""
		}
		return "core php fmt --fix"
	case "analyse":
		return i18n.T("cmd.php.qa.fix_phpstan")
	case "psalm":
		return i18n.T("cmd.php.qa.fix_psalm")
	case "test":
		return i18n.T("cmd.php.qa.fix_tests")
	case "rector":
		if fixEnabled {
			return ""
		}
		return "core php rector --fix"
	case "infection":
		return i18n.T("cmd.php.qa.fix_infection")
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
			result.Output = i18n.T("cmd.php.qa.issue_style")
		}

	case "analyse":
		err := phppkg.Analyse(ctx, phppkg.AnalyseOptions{Dir: dir, Output: &buf})
		result.Passed = err == nil
		if err != nil {
			result.Output = i18n.T("cmd.php.qa.issue_analysis")
		}

	case "psalm":
		err := phppkg.RunPsalm(ctx, phppkg.PsalmOptions{Dir: dir, Fix: fix, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = i18n.T("cmd.php.qa.issue_types")
		}

	case "test":
		err := phppkg.RunTests(ctx, phppkg.TestOptions{Dir: dir, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = i18n.T("cmd.php.qa.issue_tests")
		}

	case "rector":
		err := phppkg.RunRector(ctx, phppkg.RectorOptions{Dir: dir, Fix: fix, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = i18n.T("cmd.php.qa.issue_rector")
		}

	case "infection":
		err := phppkg.RunInfection(ctx, phppkg.InfectionOptions{Dir: dir, Output: io.Discard})
		result.Passed = err == nil
		if err != nil {
			result.Output = i18n.T("cmd.php.qa.issue_mutation")
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
		Short: i18n.T("cmd.php.rector.short"),
		Long:  i18n.T("cmd.php.rector.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			// Check if Rector is available
			if !phppkg.DetectRector(cwd) {
				fmt.Printf("%s %s\n\n", errorStyle.Render(i18n.T("cmd.php.label.error")), i18n.T("cmd.php.rector.not_found"))
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.install")), i18n.T("cmd.php.rector.install"))
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.setup")), i18n.T("cmd.php.rector.setup"))
				return fmt.Errorf(i18n.T("cmd.php.error.rector_not_installed"))
			}

			var msg string
			if rectorFix {
				msg = i18n.T("cmd.php.rector.refactoring")
			} else {
				msg = i18n.T("cmd.php.rector.analysing")
			}
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.rector")), msg)

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
					return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.rector_failed"), err)
				}
				// Dry-run returns non-zero if changes would be made
				fmt.Printf("\n%s %s\n", phpQAWarningStyle.Render(i18n.T("cmd.php.label.info")), i18n.T("cmd.php.rector.changes_suggested"))
				return nil
			}

			if rectorFix {
				fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.rector.refactored"))
			} else {
				fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.rector.no_changes"))
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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.working_dir"), err)
			}

			if !phppkg.IsPHPProject(cwd) {
				return fmt.Errorf(i18n.T("cmd.php.error.not_php"))
			}

			// Check if Infection is available
			if !phppkg.DetectInfection(cwd) {
				fmt.Printf("%s %s\n\n", errorStyle.Render(i18n.T("cmd.php.label.error")), i18n.T("cmd.php.infection.not_found"))
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.install")), i18n.T("cmd.php.infection.install"))
				return fmt.Errorf(i18n.T("cmd.php.error.infection_not_installed"))
			}

			fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.infection")), i18n.T("cmd.php.infection.running"))
			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.info")), i18n.T("cmd.php.infection.note"))

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
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.infection_failed"), err)
			}

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("cmd.php.label.done")), i18n.T("cmd.php.infection.complete"))
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
