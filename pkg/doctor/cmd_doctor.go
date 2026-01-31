// Package doctor provides environment check commands.
package doctor

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	dimStyle     = cli.DimStyle
)

// Flag variable for doctor command
var doctorVerbose bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: i18n.T("cmd.doctor.short"),
	Long:  i18n.T("cmd.doctor.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor(doctorVerbose)
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false, i18n.T("cmd.doctor.verbose_flag"))
}

func runDoctor(verbose bool) error {
	fmt.Println(i18n.T("common.progress.checking", map[string]any{"Item": "development environment"}))
	fmt.Println()

	var passed, failed, optional int

	// Check required tools
	fmt.Println(i18n.T("cmd.doctor.required"))
	for _, c := range requiredChecks() {
		ok, version := runCheck(c)
		if ok {
			if verbose {
				fmt.Println(formatCheckResult(true, c.name, version))
			} else {
				fmt.Println(formatCheckResult(true, c.name, ""))
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", errorStyle.Render(cli.Glyph(":cross:")), c.name, c.description)
			failed++
		}
	}

	// Check optional tools
	fmt.Printf("\n%s\n", i18n.T("cmd.doctor.optional"))
	for _, c := range optionalChecks() {
		ok, version := runCheck(c)
		if ok {
			if verbose {
				fmt.Println(formatCheckResult(true, c.name, version))
			} else {
				fmt.Println(formatCheckResult(true, c.name, ""))
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", dimStyle.Render(cli.Glyph(":skip:")), c.name, dimStyle.Render(c.description))
			optional++
		}
	}

	// Check GitHub access
	fmt.Printf("\n%s\n", i18n.T("cmd.doctor.github"))
	if checkGitHubSSH() {
		fmt.Println(formatCheckResult(true, i18n.T("cmd.doctor.ssh_found"), ""))
	} else {
		fmt.Printf("  %s %s\n", errorStyle.Render(cli.Glyph(":cross:")), i18n.T("cmd.doctor.ssh_missing"))
		failed++
	}

	if checkGitHubCLI() {
		fmt.Println(formatCheckResult(true, i18n.T("cmd.doctor.cli_auth"), ""))
	} else {
		fmt.Printf("  %s %s\n", errorStyle.Render(cli.Glyph(":cross:")), i18n.T("cmd.doctor.cli_auth_missing"))
		failed++
	}

	// Check workspace
	fmt.Printf("\n%s\n", i18n.T("cmd.doctor.workspace"))
	checkWorkspace()

	// Summary
	fmt.Println()
	if failed > 0 {
		cli.Error(i18n.T("cmd.doctor.issues", map[string]interface{}{"Count": failed}))
		fmt.Printf("\n%s\n", i18n.T("cmd.doctor.install_missing"))
		printInstallInstructions()
		return fmt.Errorf("%s", i18n.T("cmd.doctor.issues_error", map[string]interface{}{"Count": failed}))
	}

	cli.Success(i18n.T("cmd.doctor.ready"))
	return nil
}

func formatCheckResult(ok bool, name, detail string) string {
	check := cli.Check(name)
	if ok {
		check.Pass()
	} else {
		check.Fail()
	}
	if detail != "" {
		check.Message(detail)
	} else {
		check.Message("")
	}
	return check.String()
}
