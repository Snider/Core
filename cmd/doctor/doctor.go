// Package doctor provides environment check commands.
package doctor

import (
	"fmt"

	"github.com/host-uk/core/cmd/shared"
	"github.com/leaanthony/clir"
)

// Style aliases from shared
var (
	successStyle = shared.SuccessStyle
	errorStyle   = shared.ErrorStyle
	dimStyle     = shared.DimStyle
)

// AddDoctorCommand adds the 'doctor' command to the given parent command.
func AddDoctorCommand(parent *clir.Cli) {
	var verbose bool

	doctorCmd := parent.NewSubCommand("doctor", "Check development environment")
	doctorCmd.LongDescription("Checks that all required tools are installed and configured.\n" +
		"Run this before `core setup` to ensure your environment is ready.")

	doctorCmd.BoolFlag("verbose", "Show detailed version information", &verbose)

	doctorCmd.Action(func() error {
		return runDoctor(verbose)
	})
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
			if verbose && version != "" {
				fmt.Printf("  %s %s %s\n", successStyle.Render("✓"), c.name, dimStyle.Render(version))
			} else {
				fmt.Printf("  %s %s\n", successStyle.Render("✓"), c.name)
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", errorStyle.Render("✗"), c.name, c.description)
			failed++
		}
	}

	// Check optional tools
	fmt.Println("\nOptional:")
	for _, c := range optionalChecks {
		ok, version := runCheck(c)
		if ok {
			if verbose && version != "" {
				fmt.Printf("  %s %s %s\n", successStyle.Render("✓"), c.name, dimStyle.Render(version))
			} else {
				fmt.Printf("  %s %s\n", successStyle.Render("✓"), c.name)
			}
			passed++
		} else {
			fmt.Printf("  %s %s - %s\n", dimStyle.Render("○"), c.name, dimStyle.Render(c.description))
			optional++
		}
	}

	// Check GitHub access
	fmt.Println("\nGitHub Access:")
	if checkGitHubSSH() {
		fmt.Printf("  %s SSH key found\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s SSH key missing - run: ssh-keygen && gh ssh-key add\n", errorStyle.Render("✗"))
		failed++
	}

	if checkGitHubCLI() {
		fmt.Printf("  %s CLI authenticated\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s CLI authentication - run: gh auth login\n", errorStyle.Render("✗"))
		failed++
	}

	// Check workspace
	fmt.Println("\nWorkspace:")
	checkWorkspace()

	// Summary
	fmt.Println()
	if failed > 0 {
		fmt.Printf("%s %d issues found\n", errorStyle.Render("Doctor:"), failed)
		fmt.Println("\nInstall missing tools:")
		printInstallInstructions()
		return fmt.Errorf("%d required tools missing", failed)
	}

	fmt.Printf("%s Environment ready\n", successStyle.Render("Doctor:"))
	return nil
}
