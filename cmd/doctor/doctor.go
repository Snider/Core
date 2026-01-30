// Package doctor provides environment check commands.
package doctor

import (
	"fmt"

	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Style aliases from shared
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
)

// Flag variable for doctor command
var doctorVerbose bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check development environment",
	Long: `Checks that all required tools are installed and configured.
Run this before 'core setup' to ensure your environment is ready.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctor(doctorVerbose)
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false, "Show detailed version information")
}

func runDoctor(verbose bool) error {
	fmt.Println("Checking development environment...")
	fmt.Println()

	var passed, failed, optional int

	// Check required tools
	fmt.Println("Required:")
	for _, c := range requiredChecks {
		ok, version := runCheck(c)
		if ok {
			if verbose {
				fmt.Println(shared.CheckResult(true, c.name, version))
			} else {
				fmt.Println(shared.CheckResult(true, c.name, ""))
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", errorStyle.Render(shared.SymbolCross), c.name, c.description)
			failed++
		}
	}

	// Check optional tools
	fmt.Println("\nOptional:")
	for _, c := range optionalChecks {
		ok, version := runCheck(c)
		if ok {
			if verbose {
				fmt.Println(shared.CheckResult(true, c.name, version))
			} else {
				fmt.Println(shared.CheckResult(true, c.name, ""))
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", dimStyle.Render(shared.SymbolSkip), c.name, dimStyle.Render(c.description))
			optional++
		}
	}

	// Check GitHub access
	fmt.Println("\nGitHub Access:")
	if checkGitHubSSH() {
		fmt.Println(shared.CheckResult(true, "SSH key found", ""))
	} else {
		fmt.Printf("  %s SSH key missing - run: ssh-keygen && gh ssh-key add\n", errorStyle.Render(shared.SymbolCross))
		failed++
	}

	if checkGitHubCLI() {
		fmt.Println(shared.CheckResult(true, "CLI authenticated", ""))
	} else {
		fmt.Printf("  %s CLI authentication - run: gh auth login\n", errorStyle.Render(shared.SymbolCross))
		failed++
	}

	// Check workspace
	fmt.Println("\nWorkspace:")
	checkWorkspace()

	// Summary
	fmt.Println()
	if failed > 0 {
		fmt.Println(shared.Error(fmt.Sprintf("Doctor: %d issues found", failed)))
		fmt.Println("\nInstall missing tools:")
		printInstallInstructions()
		return fmt.Errorf("%d required tools missing", failed)
	}

	fmt.Println(shared.Success("Doctor: Environment ready"))
	return nil
}
